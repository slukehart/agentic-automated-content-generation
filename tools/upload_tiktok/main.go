package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"content-generation-automation/media"
	"content-generation-automation/metadata"

	"github.com/joho/godotenv"
)

func main() {
	// Parse flags
	allUnposted := flag.Bool("all-unposted", false, "Upload all videos not yet posted to TikTok")
	dryRun := flag.Bool("dry-run", false, "Show what would be uploaded without actually uploading")
	useSandbox := flag.Bool("sandbox", false, "Use TikTok sandbox environment (for demo/testing)")
	flag.Parse()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	// Initialize manifest manager
	manifestManager := metadata.NewManifestManager("content_manifest.json")

	if *allUnposted {
		// Upload all unposted videos
		uploadAllUnposted(manifestManager, *dryRun, *useSandbox)
	} else {
		// Upload specific video by ID
		if flag.NArg() < 1 {
			fmt.Println("Usage:")
			fmt.Println("  Upload specific video:    go run tools/upload_tiktok/main.go <content_id>")
			fmt.Println("  Upload all unposted:      go run tools/upload_tiktok/main.go --all-unposted")
			fmt.Println("  Dry run (preview):        go run tools/upload_tiktok/main.go --dry-run --all-unposted")
			fmt.Println("  Use sandbox:              go run tools/upload_tiktok/main.go --sandbox <content_id>")
			os.Exit(1)
		}

		contentID := flag.Arg(0)
		uploadSingle(manifestManager, contentID, *dryRun, *useSandbox)
	}
}

func uploadSingle(manager *metadata.ManifestManager, contentID string, dryRun bool, useSandbox bool) {
	fmt.Printf("📋 Loading content item: %s\n", contentID)

	// Get content item
	item, err := manager.GetItem(contentID)
	if err != nil {
		log.Fatalf("❌ Failed to load content item: %v", err)
	}

	// Check if already posted (note: TikTok goes to inbox first, so this checks if uploaded)
	if item.Status.TikTok.Posted && item.Status.TikTok.URL != nil {
		fmt.Printf("⚠️  Video already uploaded to TikTok\n")
		fmt.Printf("   Status: %s\n", *item.Status.TikTok.URL)
		if item.Status.TikTok.PostedAt != nil {
			fmt.Printf("   Uploaded at: %s\n", item.Status.TikTok.PostedAt.Format("2006-01-02 15:04:05"))
		}
		return
	}

	// Dry run - just show what would be uploaded
	if dryRun {
		fmt.Printf("\n🔍 DRY RUN - Would upload:\n")
		fmt.Printf("   Caption: %s\n", item.Platforms.TikTok.Caption)
		fmt.Printf("   Video: %s\n", item.Media.VideoPath)
		fmt.Printf("   Hashtags: %v\n", item.Platforms.TikTok.Hashtags)
		if useSandbox {
			fmt.Printf("   Environment: SANDBOX (demo/testing)\n")
		}
		return
	}

	// Upload to TikTok
	fmt.Printf("\n📤 Uploading to TikTok...\n")
	if useSandbox {
		fmt.Printf("🧪 Using SANDBOX environment\n")
	}

	result, err := media.UploadVideoToTikTok(item, useSandbox)
	if err != nil {
		log.Fatalf("❌ Upload failed: %v", err)
	}

	// Update manifest
	item.Status.TikTok.Posted = false // Not posted to feed yet (in inbox)
	item.Status.TikTok.PostedAt = &result.UploadedAt
	publishInfo := fmt.Sprintf("inbox_uploaded (publish_id: %s)", result.PublishID)
	item.Status.TikTok.URL = &publishInfo

	if err := manager.UpdateItem(*item); err != nil {
		log.Printf("⚠️  Warning: Failed to update manifest: %v", err)
	}

	fmt.Printf("\n✅ Upload complete!\n")
	fmt.Printf("   Status: %s\n", publishInfo)
}

func uploadAllUnposted(manager *metadata.ManifestManager, dryRun bool, useSandbox bool) {
	fmt.Printf("📋 Loading unposted videos...\n")

	// Get all unposted items
	items, err := manager.GetItemsByStatus("tiktok", false)
	if err != nil {
		log.Fatalf("❌ Failed to load items: %v", err)
	}

	if len(items) == 0 {
		fmt.Println("✅ No unposted videos found. All caught up!")
		return
	}

	fmt.Printf("Found %d video(s) to upload\n", len(items))
	if useSandbox {
		fmt.Printf("🧪 Using SANDBOX environment\n")
	}
	fmt.Println()

	// Dry run - just list what would be uploaded
	if dryRun {
		fmt.Println("🔍 DRY RUN - Would upload the following videos:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for i, item := range items {
			fmt.Printf("\n%d. %s\n", i+1, item.ID)
			fmt.Printf("   Caption: %s\n", item.Platforms.TikTok.Caption)
			fmt.Printf("   Video: %s\n", item.Media.VideoPath)
		}
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("\nRun without --dry-run to actually upload these videos.\n")
		return
	}

	// Upload each video
	successCount := 0
	failCount := 0

	for i, item := range items {
		fmt.Printf("\n[%d/%d] Uploading: %s\n", i+1, len(items), item.ID)
		fmt.Printf("        Caption: %s\n", truncate(item.Platforms.TikTok.Caption, 60))

		result, err := media.UploadVideoToTikTok(&item, useSandbox)
		if err != nil {
			log.Printf("❌ Upload failed: %v", err)
			failCount++

			// Update with error
			errMsg := err.Error()
			item.Status.TikTok.Error = &errMsg
			manager.UpdateItem(item)
			continue
		}

		// Update manifest with success
		item.Status.TikTok.Posted = false // Not posted to feed yet
		item.Status.TikTok.PostedAt = &result.UploadedAt
		publishInfo := fmt.Sprintf("inbox_uploaded (publish_id: %s)", result.PublishID)
		item.Status.TikTok.URL = &publishInfo

		if err := manager.UpdateItem(item); err != nil {
			log.Printf("⚠️  Warning: Failed to update manifest: %v", err)
		}

		fmt.Printf("✅ Uploaded to inbox: %s\n", publishInfo)
		successCount++
	}

	// Summary
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 Upload Summary\n")
	fmt.Printf("   ✅ Success: %d\n", successCount)
	if failCount > 0 {
		fmt.Printf("   ❌ Failed: %d\n", failCount)
	}
	fmt.Printf("   📋 Total: %d\n", len(items))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n📱 Remember: Check TikTok app to complete posting for each video!")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

