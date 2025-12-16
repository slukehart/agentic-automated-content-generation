package news

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"content-generation-automation/metadata"

	grok "github.com/SimonMorphy/grok-go"
	"github.com/joho/godotenv"
)

// GenerateEnrichedNewsContent generates news summary with platform metadata
// This is the MOST TOKEN-EFFICIENT approach: single API call for summary + all metadata
func GenerateEnrichedNewsContent(article AiArticleParameters) (*EnrichedNewsContent, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("X_AI_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("X_AI_KEY environment variable not set")
	}

	// Initialize client with extended timeout
	client, err := grok.NewClientWithOptions(apiKey,
		grok.WithTimeout(5*time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Use the metadata generation prompt
	systemPrompt := metadata.MetadataGenerationPrompt()
	userPrompt := metadata.UserPromptForArticle(article.ArticleTitle, article.ArticleUrl)

	fmt.Printf("Generating enriched content with metadata for: %s\n", article.ArticleTitle)

	request := &grok.ChatCompletionRequest{
		Model: "grok-3",
		Messages: []grok.ChatCompletionMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   4000, // Enough for summary + all platform metadata
	}
	request.StreamOptions.IncludeUsage = true

	// Send request with extended context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response, err := grok.CreateChatCompletion(ctx, client, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from Grok")
	}

	// Extract the JSON response
	rawResponse := response.Choices[0].Message.Content

	// Clean up response (remove markdown code blocks if present)
	rawResponse = strings.TrimSpace(rawResponse)
	rawResponse = strings.TrimPrefix(rawResponse, "```json")
	rawResponse = strings.TrimPrefix(rawResponse, "```")
	rawResponse = strings.TrimSuffix(rawResponse, "```")
	rawResponse = strings.TrimSpace(rawResponse)

	// Debug output
	fmt.Printf("Raw LLM response length: %d chars\n", len(rawResponse))
	if len(rawResponse) < 100 {
		fmt.Printf("Raw response: %s\n", rawResponse)
	}

	// Parse the JSON response
	var metadataResponse metadata.LLMMetadataResponse
	if err := json.Unmarshal([]byte(rawResponse), &metadataResponse); err != nil {
		// Log the error and response for debugging
		fmt.Printf("Failed to parse LLM response as JSON. Error: %v\n", err)
		fmt.Printf("Response (first 500 chars): %s\n", rawResponse[:min(500, len(rawResponse))])
		return nil, fmt.Errorf("failed to parse LLM JSON response: %w", err)
	}

	// Validate that we got content
	if metadataResponse.Summary == "" {
		return nil, fmt.Errorf("LLM returned empty summary")
	}

	return &EnrichedNewsContent{
		Summary:  metadataResponse.Summary,
		Metadata: metadataResponse,
	}, nil
}

// BackgroundImageResult contains the path to the generated background image
type BackgroundImageResult struct {
	ImagePath string
	ImageURL  string // For future use if we want to upload to cloud storage
}

// GenerateNewsroomBackground generates a newsroom background image using X.AI's Grok image generation API
func GenerateNewsroomBackground(prompt string) (*BackgroundImageResult, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("X_AI_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("X_AI_KEY environment variable not set")
	}

	// Use default prompt if none provided
	if prompt == "" {
		prompt = "A professional modern newsroom background with dark blue tones, large screens displaying news graphics, sleek furniture, and ambient lighting. Cinematic, high quality, photorealistic."
	}

	fmt.Printf("Generating newsroom background image...\n")
	fmt.Printf("Prompt: %s\n", prompt)

	// Create image generation request body
	type imageGenRequest struct {
		Model          string `json:"model"`
		Prompt         string `json:"prompt"`
		N              int    `json:"n"`
		ResponseFormat string `json:"response_format"`
	}

	reqBody := imageGenRequest{
		Model:          "grok-2-image",
		Prompt:         prompt,
		N: 				1,
		ResponseFormat: "b64_json",
	}

	requestData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request to X.AI API
	apiURL := "https://api.x.ai/v1/images/generations"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	type imageGenResponse struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}

	var response imageGenResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if we got any images back
	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no images returned from X.AI API")
	}

	// Get the base64 encoded image data
	imageData := response.Data[0].B64JSON
	if imageData == "" {
		return nil, fmt.Errorf("no image data in response")
	}

	// Decode base64 image
	imageBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %w", err)
	}

	// Create backgrounds directory if it doesn't exist
	backgroundsDir := "backgrounds"
	if err := os.MkdirAll(backgroundsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backgrounds directory: %w", err)
	}

	// Generate filename with timestamp
	// Note: Grok returns JPEG images regardless of what we ask for
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("newsroom_%s.jpg", timestamp)
	imagePath := filepath.Join(backgroundsDir, filename)

	// Save image to file
	if err := os.WriteFile(imagePath, imageBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	fmt.Printf("✅ Background image saved to: %s\n", imagePath)

	return &BackgroundImageResult{
		ImagePath: imagePath,
	}, nil
}

// ConvertToContentItem converts enriched news content to a full ContentItem
func ConvertToContentItem(
	article AiArticleParameters,
	enriched *EnrichedNewsContent,
	contentID string,
	audioPath string,
	videoPath string,
	avatarID string,
	videoDuration float64,
) metadata.ContentItem {
	now := time.Now()

	// Parse published date if available
	var publishedAt time.Time
	// Note: article doesn't have PublishedAt in current struct, using now as fallback
	publishedAt = now

	// Count words in summary (rough estimate)
	wordCount := len(strings.Fields(enriched.Summary))

	// Estimate duration from word count (average speaking rate: 150 words/min = 2.5 words/sec)
	estimatedDuration := int(float64(wordCount) / 2.5)

	return metadata.ContentItem{
		ID:        contentID,
		CreatedAt: now,
		Source: metadata.SourceInfo{
			URL:         article.ArticleUrl,
			Title:       article.ArticleTitle,
			Author:      "", // Not available in current article struct
			PublishedAt: publishedAt,
			SourceName:  extractDomain(article.ArticleUrl),
		},
		Content: metadata.ContentInfo{
			Summary:               enriched.Summary,
			WordCount:             wordCount,
			EstimatedDurationSecs: estimatedDuration,
		},
		Media: metadata.MediaInfo{
			AudioPath:       audioPath,
			VideoPath:       videoPath,
			ThumbnailPath:   nil,
			DurationSeconds: videoDuration,
			Resolution:      "1280x720",
			AvatarID:        avatarID,
		},
		SEO: enriched.Metadata.SEO,
		Platforms: metadata.PlatformMetadata{
			YouTube: metadata.YouTubeMetadata{
				Title:           enriched.Metadata.Platforms.YouTube.Title,
				Description:     enriched.Metadata.Platforms.YouTube.Description,
				Tags:            enriched.Metadata.Platforms.YouTube.Tags,
				CategoryID:      "25", // News & Politics
				DefaultLanguage: "en",
				PrivacyStatus:   "public",
				Timestamps:      enriched.Metadata.Platforms.YouTube.Timestamps,
			},
			TikTok: metadata.TikTokMetadata{
				Caption:       enriched.Metadata.Platforms.TikTok.Caption,
				Hashtags:      enriched.Metadata.Platforms.TikTok.Hashtags,
				PrivacyLevel:  "public",
				DuetEnabled:   true,
				StitchEnabled: true,
			},
			Instagram: metadata.InstagramMetadata{
				Caption:       enriched.Metadata.Platforms.Instagram.Caption,
				Hashtags:      enriched.Metadata.Platforms.Instagram.Hashtags,
				Location:      nil,
				Collaborators: []string{},
			},
			Twitter: metadata.TwitterMetadata{
				Tweet:         enriched.Metadata.Platforms.Twitter.Tweet,
				Hashtags:      enriched.Metadata.Platforms.Twitter.Hashtags,
				ReplySettings: "everyone",
			},
			Facebook: metadata.FacebookMetadata{
				Message:         enriched.Metadata.Platforms.Facebook.Message,
				LinkDescription: enriched.Metadata.Platforms.Facebook.LinkDescription,
			},
			LinkedIn: metadata.LinkedInMetadata{
				PostText: enriched.Metadata.Platforms.LinkedIn.PostText,
				Hashtags: enriched.Metadata.Platforms.LinkedIn.Hashtags,
			},
		},
		Status: metadata.PostingStatus{
			YouTube:   metadata.PlatformStatus{Posted: false},
			TikTok:    metadata.PlatformStatus{Posted: false},
			Instagram: metadata.PlatformStatus{Posted: false},
			Twitter:   metadata.PlatformStatus{Posted: false},
			Facebook:  metadata.PlatformStatus{Posted: false},
			LinkedIn:  metadata.PlatformStatus{Posted: false},
		},
		Analytics: metadata.AnalyticsInfo{},
	}
}

// Helper function to extract domain from URL
func extractDomain(url string) string {
	// Simple domain extraction (could be improved)
	parts := strings.Split(url, "/")
	if len(parts) >= 3 {
		domain := parts[2]
		// Remove www. prefix
		domain = strings.TrimPrefix(domain, "www.")
		return domain
	}
	return "unknown"
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

