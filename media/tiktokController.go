package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"content-generation-automation/metadata"
)

const (
	// TikTok Content Posting API endpoints
	tiktokAPIBase       = "https://open.tiktokapis.com"
	tiktokInitUploadURL = "/v2/post/publish/inbox/video/init/"
	tiktokPublishURL    = "/v2/post/publish/video/init/"
)

// TikTokUploadResult contains the result of a TikTok video upload
type TikTokUploadResult struct {
	PublishID  string
	Status     string // "inbox_uploaded"
	UploadedAt time.Time
}

// TikTokInitUploadRequest is the request body for initializing video upload
// See: https://developers.tiktok.com/doc/content-posting-api-reference-upload-video
type TikTokInitUploadRequest struct {
	SourceInfo struct {
		Source          string `json:"source"`            // "FILE_UPLOAD" or "PULL_FROM_URL"
		VideoSize       int64  `json:"video_size"`        // Total video size in bytes
		ChunkSize       int64  `json:"chunk_size"`        // Size of each chunk in bytes
		TotalChunkCount int    `json:"total_chunk_count"` // Number of chunks
	} `json:"source_info"`
	PostInfo struct {
		Title           string `json:"title,omitempty"`
		PrivacyLevel    string `json:"privacy_level"`                   // "SELF_ONLY", "MUTUAL_FOLLOW_FRIENDS", "FOLLOWER_OF_CREATOR", "PUBLIC_TO_EVERYONE"
		DisableComment  bool   `json:"disable_comment"`
		DisableDuet     bool   `json:"disable_duet"`
		DisableStitch   bool   `json:"disable_stitch"`
		VideoCoverTs    int    `json:"video_cover_timestamp_ms,omitempty"`
	} `json:"post_info"`
}

// TikTokInitUploadResponse is the response from init upload
type TikTokInitUploadResponse struct {
	Data struct {
		PublishID string `json:"publish_id"`
		UploadURL string `json:"upload_url"`
	} `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// UploadVideoToTikTok uploads a video to TikTok with metadata from ContentItem
// This function authenticates with TikTok API, uploads the video file to the creator's inbox
func UploadVideoToTikTok(contentItem *metadata.ContentItem, sandbox bool) (*TikTokUploadResult, error) {
	ctx := context.Background()

	// Get access token
	accessToken, err := GetTikTokAccessToken(ctx, sandbox)
	if err != nil {
		return nil, fmt.Errorf("failed to get TikTok access token: %w", err)
	}

	// Verify video file exists
	videoPath := contentItem.Media.VideoPath
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("video file not found: %s", videoPath)
	}

	// Get video file size
	fileInfo, err := os.Stat(videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video file info: %w", err)
	}
	videoSize := fileInfo.Size()

	// Check video size limit (287.6 MB ≈ 301 MB)
	maxSize := int64(301 * 1024 * 1024)
	if videoSize > maxSize {
		return nil, fmt.Errorf("video file too large: %d bytes (max: %d bytes)", videoSize, maxSize)
	}

	fmt.Printf("📤 Uploading video to TikTok inbox...\n")
	if sandbox {
		fmt.Printf("   🧪 SANDBOX MODE - For demo/testing only\n")
	}
	fmt.Printf("   Title: %s\n", truncateString(contentItem.Platforms.TikTok.Caption, 50))
	fmt.Printf("   File: %s\n", videoPath)
	fmt.Printf("   Size: %.2f MB\n", float64(videoSize)/(1024*1024))

	// Step 1: Initialize upload
	// Per TikTok docs: videos < 5MB must be uploaded as one chunk.
	// Videos > 64MB must use multiple chunks (each 5-64MB).
	chunkSize := videoSize // Default: upload as single chunk
	totalChunks := 1
	const minChunk int64 = 5 * 1024 * 1024  // 5 MB
	const maxChunk int64 = 64 * 1024 * 1024  // 64 MB
	if videoSize > maxChunk {
		chunkSize = maxChunk
		totalChunks = int(videoSize / chunkSize)
		if videoSize%chunkSize != 0 {
			// Remaining bytes merge into the last chunk (TikTok allows last chunk up to 128MB)
		}
	}

	initReq := TikTokInitUploadRequest{}
	initReq.SourceInfo.Source = "FILE_UPLOAD"
	initReq.SourceInfo.VideoSize = videoSize
	initReq.SourceInfo.ChunkSize = chunkSize
	initReq.SourceInfo.TotalChunkCount = totalChunks
	initReq.PostInfo.PrivacyLevel = contentItem.Platforms.TikTok.PrivacyLevel
	initReq.PostInfo.DisableComment = false
	initReq.PostInfo.DisableDuet = !contentItem.Platforms.TikTok.DuetEnabled
	initReq.PostInfo.DisableStitch = !contentItem.Platforms.TikTok.StitchEnabled

	initResp, err := initializeTikTokUpload(accessToken, &initReq)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize upload: %w", err)
	}

	fmt.Printf("   ✅ Upload initialized (publish_id: %s)\n", initResp.Data.PublishID)

	// Step 2: Upload video file
	err = uploadTikTokVideoFile(initResp.Data.UploadURL, videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video file: %w", err)
	}

	fmt.Printf("   ✅ Video file uploaded\n")

	// Create result
	result := &TikTokUploadResult{
		PublishID:  initResp.Data.PublishID,
		Status:     "inbox_uploaded",
		UploadedAt: time.Now(),
	}

	fmt.Printf("\n✅ Video uploaded to TikTok inbox!\n")
	fmt.Printf("   Publish ID: %s\n", result.PublishID)
	fmt.Printf("\n📱 NEXT STEPS:\n")
	fmt.Printf("   1. Open TikTok app on your phone\n")
	fmt.Printf("   2. Check notifications/inbox\n")
	fmt.Printf("   3. Review the video\n")
	fmt.Printf("   4. Add caption: %s\n", truncateString(contentItem.Platforms.TikTok.Caption, 60))
	fmt.Printf("   5. Add hashtags: %v\n", contentItem.Platforms.TikTok.Hashtags)
	fmt.Printf("   6. Click 'Post' to publish (or delete if you don't want to post)\n")
	fmt.Printf("\n💡 Note: TikTok API uploads to inbox only. You must complete posting in the app.\n")

	return result, nil
}

// initializeTikTokUpload initializes the upload session
func initializeTikTokUpload(accessToken string, req *TikTokInitUploadRequest) (*TikTokInitUploadResponse, error) {
	url := tiktokAPIBase + tiktokInitUploadURL

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TikTok API error (status %d): %s", resp.StatusCode, string(body))
	}

	var initResp TikTokInitUploadResponse
	if err := json.Unmarshal(body, &initResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if initResp.Error.Code != "" && initResp.Error.Code != "ok" {
		return nil, fmt.Errorf("TikTok API error: %s - %s", initResp.Error.Code, initResp.Error.Message)
	}

	if initResp.Data.UploadURL == "" || initResp.Data.PublishID == "" {
		return nil, fmt.Errorf("invalid response: missing upload_url or publish_id")
	}

	return &initResp, nil
}

// uploadTikTokVideoFile uploads the video file to the provided upload URL
// See: https://developers.tiktok.com/doc/content-posting-api-media-transfer-guide
func uploadTikTokVideoFile(uploadURL, videoPath string) error {
	file, err := os.Open(videoPath)
	if err != nil {
		return fmt.Errorf("failed to open video file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	totalSize := fileInfo.Size()

	httpReq, err := http.NewRequest("PUT", uploadURL, file)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "video/mp4")
	httpReq.Header.Set("Content-Length", fmt.Sprintf("%d", totalSize))
	httpReq.Header.Set("Content-Range", fmt.Sprintf("bytes 0-%d/%d", totalSize-1, totalSize))
	httpReq.ContentLength = totalSize

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to upload video: %w", err)
	}
	defer resp.Body.Close()

	// 201 = all parts uploaded, 206 = partial (more chunks needed)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusPartialContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

