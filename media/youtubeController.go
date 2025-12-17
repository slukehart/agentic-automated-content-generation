package media

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"content-generation-automation/metadata"

	"google.golang.org/api/youtube/v3"
)

// YouTubeUploadResult contains the result of a YouTube video upload
type YouTubeUploadResult struct {
	VideoID    string
	VideoURL   string
	UploadedAt time.Time
}

// UploadVideoToYouTube uploads a video to YouTube with metadata from ContentItem
// This function authenticates with YouTube API, uploads the video file,
// and sets all metadata (title, description, tags, etc.) from the ContentItem.
func UploadVideoToYouTube(contentItem *metadata.ContentItem) (*YouTubeUploadResult, error) {
	ctx := context.Background()

	// Get authenticated YouTube client
	service, err := GetYouTubeClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube client: %w", err)
	}

	// Verify video file exists
	videoPath := contentItem.Media.VideoPath
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("video file not found: %s", videoPath)
	}

	// Open video file
	file, err := os.Open(videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open video file: %w", err)
	}
	defer file.Close()

	// Format description with timestamps
	description := formatYouTubeDescription(contentItem)

	// Add Shorts-specific tags to help YouTube identify this as a Short
	tags := contentItem.Platforms.YouTube.Tags
	shortsTagFound := false
	for _, tag := range tags {
		if strings.ToLower(tag) == "shorts" || strings.ToLower(tag) == "short" {
			shortsTagFound = true
			break
		}
	}
	if !shortsTagFound {
		tags = append([]string{"Shorts"}, tags...) // Prepend "Shorts" tag
	}

	// Create YouTube video resource
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:           truncateTitle(contentItem.Platforms.YouTube.Title, 100),
			Description:     truncateDescription(description, 5000),
			Tags:            tags,
			CategoryId:      contentItem.Platforms.YouTube.CategoryID,
			DefaultLanguage: contentItem.Platforms.YouTube.DefaultLanguage,
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus:           contentItem.Platforms.YouTube.PrivacyStatus,
			SelfDeclaredMadeForKids: false, // Set to false for news content
		},
	}

	// Upload video
	call := service.Videos.Insert([]string{"snippet", "status"}, video)

	fmt.Printf("📤 Uploading video to YouTube...\n")
	fmt.Printf("   Title: %s\n", video.Snippet.Title)
	fmt.Printf("   File: %s\n", videoPath)

	response, err := call.Media(file).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	// Create result
	result := &YouTubeUploadResult{
		VideoID:    response.Id,
		VideoURL:   fmt.Sprintf("https://www.youtube.com/watch?v=%s", response.Id),
		UploadedAt: time.Now(),
	}

	fmt.Printf("✅ Video uploaded successfully as YouTube Short!\n")
	fmt.Printf("   Video ID: %s\n", result.VideoID)
	fmt.Printf("   URL: %s\n", result.VideoURL)
	fmt.Printf("   📱 Shorts URL: https://youtube.com/shorts/%s\n", response.Id)
	fmt.Printf("   ⏳ Note: Video is processing and may take 1-5 minutes to be fully available\n")
	fmt.Printf("   💡 Tip: YouTube Shorts appear in the Shorts feed within a few hours\n")

	return result, nil
}

// formatYouTubeDescription creates a formatted description with timestamps and source info
func formatYouTubeDescription(item *metadata.ContentItem) string {
	var sb strings.Builder

	// Add #Shorts tag at the beginning for YouTube Shorts algorithm
	sb.WriteString("#Shorts\n\n")

	// Main description
	sb.WriteString(item.Platforms.YouTube.Description)
	sb.WriteString("\n\n")

	// Add timestamps if available
	if len(item.Platforms.YouTube.Timestamps) > 0 {
		sb.WriteString("⏱️ Timestamps:\n")
		for _, ts := range item.Platforms.YouTube.Timestamps {
			sb.WriteString(fmt.Sprintf("%s - %s\n", ts.Time, ts.Label))
		}
		sb.WriteString("\n")
	}

	// Source attribution
	sb.WriteString(fmt.Sprintf("📰 Source: %s\n", item.Source.SourceName))
	if item.Source.URL != "" {
		sb.WriteString(fmt.Sprintf("🔗 Full Article: %s\n", item.Source.URL))
	}
	sb.WriteString("\n")

	// AI disclosure
	sb.WriteString("---\n")
	sb.WriteString("This video was automatically generated using AI technology:\n")
	sb.WriteString("• Grok AI for news summarization\n")
	sb.WriteString("• HeyGen for AI avatar video generation\n")
	sb.WriteString("• NewsAPI for news sourcing\n\n")
	sb.WriteString("For in-depth coverage and analysis, please refer to the original source article linked above.\n")

	return sb.String()
}

// truncateTitle ensures the title doesn't exceed YouTube's limit
func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}

// truncateDescription ensures the description doesn't exceed YouTube's limit
func truncateDescription(description string, maxLen int) string {
	if len(description) <= maxLen {
		return description
	}
	return description[:maxLen-50] + "\n\n[Description truncated due to length]"
}
