package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"content-generation-automation/media"
	"content-generation-automation/metadata"
	"content-generation-automation/news"
	"content-generation-automation/video"

	"github.com/joho/godotenv"
)

func main() {
	// Parse command-line flags
	uploadToYouTube := flag.Bool("upload-youtube", false, "Upload generated video to YouTube after creation")
	uploadToTikTok := flag.Bool("upload-tiktok", false, "Upload generated video to TikTok after creation")
	useSandbox := flag.Bool("sandbox", false, "Use TikTok sandbox environment (for demo/testing)")
	flag.Parse()

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
	if os.Getenv("RUNPOD_API_KEY") == "" || os.Getenv("RUNPOD_ENDPOINT_ID") == "" {
		log.Fatal("❌ RUNPOD_API_KEY and RUNPOD_ENDPOINT_ID must be set. Please add them to your .env file or export them.")
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
		// LongCat animates whatever is in cond_image, so there is no fallback without it
		log.Fatalf("❌ Error generating avatar image: %v", err)
	}
	fmt.Printf("✅ Avatar image saved: %s\n", backgroundResult.ImagePath)

	// Generate narration (Kokoro) then animate the avatar image with it (LongCat on RunPod)
	fmt.Println("\n=== Generating AI Avatar Video ===")

	// Generate unique content ID
	contentID := fmt.Sprintf("news_%s", time.Now().Format("20060102_150405"))
	finalPath := fmt.Sprintf("%s_final.mp4", contentID)

	fmt.Printf("\n📝 Processing: %s\n", contentID)
	fmt.Printf("    Summary: %s...\n", truncateString(enrichedContent.Summary, 60))

	audioPath := fmt.Sprintf("audio/%s.wav", contentID)
	fmt.Printf("    🎙️ Generating narration with Kokoro...\n")
	if err := video.GenerateNarration(enrichedContent.Summary, audioPath); err != nil {
		log.Fatalf("    ❌ Narration failed: %v", err)
	}

	// A ~60s video takes ~40 minutes of A100 time plus queue wait
	fmt.Printf("    🎬 Generating avatar video on RunPod (this can take 30+ minutes)...\n")
	videoResp, err := video.GenerateAvatarVideo(backgroundResult.ImagePath, audioPath, "", finalPath)
	if err != nil {
		log.Printf("    ❌ Video failed: %v", err)
		if videoResp != nil && videoResp.Message != "" {
			log.Printf("    Details: %s", videoResp.Message)
		}
		log.Fatalf("Cannot continue without video")
	}

	fmt.Printf("    ✅ Final narrated video: %s (%.1fs)\n", finalPath, videoResp.Duration)

	// Step 3: Create content item with all metadata
	fmt.Println("\n💾 Saving content metadata to manifest...")
	avatarID := video.DefaultAvatarID
	contentItem := news.ConvertToContentItem(
		article,
		enrichedContent,
		contentID,
		audioPath,
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

	// Upload to YouTube if flag is set
	if *uploadToYouTube {
		fmt.Println("\n=== Uploading to YouTube ===")
		result, err := media.UploadVideoToYouTube(&contentItem)
		if err != nil {
			log.Printf("❌ YouTube upload failed: %v", err)
			// Update status with error
			errMsg := err.Error()
			contentItem.Status.YouTube.Error = &errMsg
			contentItem.Status.YouTube.Posted = false
		} else {
			// Update status with success
			contentItem.Status.YouTube.Posted = true
			contentItem.Status.YouTube.URL = &result.VideoURL
			contentItem.Status.YouTube.PostedAt = &result.UploadedAt
			fmt.Printf("\n✅ Successfully uploaded to YouTube!\n")
			fmt.Printf("   🔗 Watch at: %s\n", result.VideoURL)
		}

		// Save updated status to manifest
		if err := manifestManager.UpdateItem(contentItem); err != nil {
			log.Printf("⚠️  Warning: Failed to update manifest with upload status: %v", err)
		} else {
			fmt.Printf("   📋 Manifest updated with posting status\n")
		}
	}

	// Upload to TikTok if flag is set
	if *uploadToTikTok {
		fmt.Println("\n=== Uploading to TikTok ===")
		if *useSandbox {
			fmt.Println("🧪 Using SANDBOX environment (for demo/testing)")
		}

		result, err := media.UploadVideoToTikTok(&contentItem, *useSandbox)
		if err != nil {
			log.Printf("❌ TikTok upload failed: %v", err)
			// Update status with error
			errMsg := err.Error()
			contentItem.Status.TikTok.Error = &errMsg
			contentItem.Status.TikTok.Posted = false
		} else {
			// Update status with "inbox_uploaded" (user must complete in app)
			contentItem.Status.TikTok.Posted = false // Not posted to feed yet
			contentItem.Status.TikTok.PostedAt = &result.UploadedAt

			// Store publish_id in URL field for tracking
			publishInfo := fmt.Sprintf("inbox_uploaded (publish_id: %s)", result.PublishID)
			contentItem.Status.TikTok.URL = &publishInfo

			fmt.Printf("\n✅ Successfully uploaded to TikTok inbox!\n")
			fmt.Printf("   📱 Check your TikTok app to complete posting\n")
		}

		// Save updated status to manifest
		if err := manifestManager.UpdateItem(contentItem); err != nil {
			log.Printf("⚠️  Warning: Failed to update manifest with upload status: %v", err)
		} else {
			fmt.Printf("   📋 Manifest updated with posting status\n")
		}
	}

	// Summary
	fmt.Println("\n=== Workflow Complete ===")
	fmt.Printf("📰 Content ID: %s\n", contentID)
	fmt.Printf("🎬 Video: %s\n", finalPath)
	fmt.Printf("📋 Metadata: content_manifest.json\n")
	if *uploadToYouTube && contentItem.Status.YouTube.Posted {
		fmt.Printf("📺 YouTube: %s\n", *contentItem.Status.YouTube.URL)
	}
	fmt.Println("\n✅ Ready to post to all platforms!")
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
