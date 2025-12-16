package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"content-generation-automation/metadata"
	"content-generation-automation/news"
	"content-generation-automation/video"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (for local development)
	// In production (Google Cloud), use Secret Manager instead
	if err := godotenv.Load(); err != nil {
		// .env file not found is OK (might be using system env vars or cloud secrets)
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	} else {
		log.Println("✅ Loaded environment variables from .env file")
	}

	// Verify critical API keys are set
	if os.Getenv("HEYGEN_API_KEY") == "" {
		log.Fatal("❌ HEYGEN_API_KEY environment variable is not set. Please add it to your .env file or export it.")
	}

	// Initialize manifest manager
	manifestManager := metadata.NewManifestManager("content_manifest.json")
	fmt.Println("📋 Initialized content manifest manager")

	// Fetch news articles from NewsAPI
	fmt.Println("\n🔍 Fetching news articles...")
	article := news.ParseNewsArticles()
	fmt.Printf("✅ Found article: %s\n", truncateString(article.ArticleTitle, 60))

	// Generate enriched content (summary + platform metadata) in single LLM call
	fmt.Println("\n🤖 Generating news summary and platform metadata...")
	enrichedContent, err := news.GenerateEnrichedNewsContent(article)
	if err != nil {
		log.Fatalf("❌ Error generating enriched content: %v", err)
	}
	fmt.Printf("✅ Generated summary (%d words) and metadata for all platforms\n", len(enrichedContent.Summary)/5)

	// Generate custom newsroom background
	fmt.Println("\n🎨 Generating custom newsroom background with Grok AI...")
	backgroundResult, err := news.GenerateNewsroomBackground("")
	if err != nil {
		log.Printf("⚠️  Warning: Failed to generate background image: %v", err)
		log.Printf("    Falling back to default newsroom background")
		backgroundResult = nil
	} else {
		fmt.Printf("✅ Background image saved: %s\n", backgroundResult.ImagePath)
	}

	// Generate video with HeyGen's built-in text-to-speech
	fmt.Println("\n=== Generating AI Avatar Video with Text-to-Speech ===")

	// Generate unique content ID
	contentID := fmt.Sprintf("news_%s", time.Now().Format("20060102_150405"))
	finalPath := fmt.Sprintf("%s_final.mp4", contentID)

	fmt.Printf("\n📝 Processing: %s\n", contentID)
	fmt.Printf("    Summary: %s...\n", truncateString(enrichedContent.Summary, 60))

	// Generate AI avatar video directly from text (no separate audio step!)
	fmt.Printf("    🎬 Generating AI avatar video with TTS (this may take 5-10 minutes)...\n")
	fmt.Printf("    📺 Avatar: Professional female news anchor\n")
	fmt.Printf("    🎙️ Voice: HeyGen professional female (US)\n")
	if backgroundResult != nil {
		fmt.Printf("    🏢 Background: Custom AI-generated newsroom (%s)\n", backgroundResult.ImagePath)
	} else {
		fmt.Printf("    🏢 Background: Default professional newsroom\n")
	}

	var videoResp *video.VideoResponse
	if backgroundResult != nil {
		// Use custom background image
		videoResp, err = video.GenerateNewsVideoWithBackgroundImage(enrichedContent.Summary, finalPath, backgroundResult.ImagePath)
	} else {
		// Fall back to default background
		videoResp, err = video.GenerateNewsVideoFromText(enrichedContent.Summary, finalPath)
	}
	if err != nil {
		log.Printf("    ❌ Video failed: %v", err)
		if videoResp != nil && videoResp.Message != "" {
			log.Printf("    Details: %s", videoResp.Message)
		}
		log.Fatalf("Cannot continue without video")
	}
	if videoResp.Status != "success" {
		log.Fatalf("    ❌ Video error: %s", videoResp.Message)
	}

	fmt.Printf("    ✅ Final narrated video: %s\n", finalPath)
	if videoResp.VideoURL != "" {
		fmt.Printf("    🔗 HeyGen URL: %s\n", videoResp.VideoURL)
	}

	fmt.Printf("    💡 Note: Video generated with HeyGen's TTS (no Google Cloud TTS needed!)\n")

	// Step 3: Create content item with all metadata
	fmt.Println("\n💾 Saving content metadata to manifest...")
	avatarID := video.DefaultAvatarID
	contentItem := news.ConvertToContentItem(
		article,
		enrichedContent,
		contentID,
		"", // Audio was deleted (already in video)
		finalPath,
		avatarID,
		videoResp.Duration,
	)

	// Add to manifest
	if err := manifestManager.AddItem(contentItem); err != nil {
		log.Printf("⚠️  Warning: Failed to save to manifest: %v", err)
	} else {
		fmt.Printf("✅ Saved metadata to content_manifest.json (ID: %s)\n", contentID)
	}

	// Display platform metadata preview
	fmt.Println("\n=== Platform Metadata Preview ===")
	fmt.Printf("📺 YouTube Title: %s\n", contentItem.Platforms.YouTube.Title)
	fmt.Printf("📱 TikTok Caption: %s\n", truncateString(contentItem.Platforms.TikTok.Caption, 60))
	fmt.Printf("📷 Instagram: %d hashtags\n", len(contentItem.Platforms.Instagram.Hashtags))
	fmt.Printf("🐦 Twitter: %s\n", truncateString(contentItem.Platforms.Twitter.Tweet, 60))

	// Summary
	fmt.Println("\n=== Workflow Complete ===")
	fmt.Printf("📰 Content ID: %s\n", contentID)
	fmt.Printf("🎬 Video: %s\n", finalPath)
	fmt.Printf("📋 Metadata: content_manifest.json\n")
	fmt.Println("\n✅ Ready to post to all platforms!")
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
