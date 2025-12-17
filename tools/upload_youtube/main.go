package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"content-generation-automation/media"
	"content-generation-automation/metadata"
)

func main() {
	// Parse flags
	allUnposted := flag.Bool("all-unposted", false, "Upload all videos not yet posted to YouTube")
	dryRun := flag.Bool("dry-run", false, "Show what would be uploaded without actually uploading")
	flag.Parse()

	// Initialize manifest manager
	manifestManager := metadata.NewManifestManager("content_manifest.json")

	if *allUnposted {
		// Upload all unposted videos
		uploadAllUnposted(manifestManager, *dryRun)
	} else {
		// Upload specific video by ID
		if flag.NArg() < 1 {
			fmt.Println("Usage:")
			fmt.Println("  Upload specific video:    go run tools/upload_youtube.go <content_id>")
			fmt.Println("  Upload all unposted:      go run tools/upload_youtube.go --all-unposted")
			fmt.Println("  Dry run (preview):        go run tools/upload_youtube.go --dry-run --all-unposted")
			os.Exit(1)
		}

		contentID := flag.Arg(0)
		uploadSingle(manifestManager, contentID, *dryRun)
	}
}

func uploadSingle(manager *metadata.ManifestManager, contentID string, dryRun bool) {
	fmt.Printf("📋 Loading content item: %s\n", contentID)

	// Get content item
	item, err := manager.GetItem(contentID)
	if err != nil {
		log.Fatalf("❌ Failed to load content item: %v", err)
	}

	// Check if already posted
	if item.Status.YouTube.Posted {
		fmt.Printf("⚠️  Video already posted to YouTube\n")
		fmt.Printf("   URL: %s\n", *item.Status.YouTube.URL)
		fmt.Printf("   Posted at: %s\n", item.Status.YouTube.PostedAt.Format("2006-01-02 15:04:05"))
		return
	}

	// Dry run - just show what would be uploaded
	if dryRun {
		fmt.Printf("\n🔍 DRY RUN - Would upload:\n")
		fmt.Printf("   Title: %s\n", item.Platforms.YouTube.Title)
		fmt.Printf("   Video: %s\n", item.Media.VideoPath)
		fmt.Printf("   Tags: %v\n", item.Platforms.YouTube.Tags)
		return
	}

	// Upload to YouTube
	fmt.Printf("\n📤 Uploading to YouTube...\n")
	result, err := media.UploadVideoToYouTube(item)
	if err != nil {
		log.Fatalf("❌ Upload failed: %v", err)
	}

	// Update manifest
	item.Status.YouTube.Posted = true
	item.Status.YouTube.URL = &result.VideoURL
	item.Status.YouTube.PostedAt = &result.UploadedAt

	if err := manager.UpdateItem(*item); err != nil {
		log.Printf("⚠️  Warning: Failed to update manifest: %v", err)
	}

	fmt.Printf("\n✅ Upload complete!\n")
	fmt.Printf("   🔗 %s\n", result.VideoURL)
}

func uploadAllUnposted(manager *metadata.ManifestManager, dryRun bool) {
	fmt.Printf("📋 Loading unposted videos...\n")

	// Get all unposted items
	items, err := manager.GetItemsByStatus("youtube", false)
	if err != nil {
		log.Fatalf("❌ Failed to load items: %v", err)
	}

	if len(items) == 0 {
		fmt.Println("✅ No unposted videos found. All caught up!")
		return
	}

	fmt.Printf("Found %d video(s) to upload\n\n", len(items))

	// Dry run - just list what would be uploaded
	if dryRun {
		fmt.Println("🔍 DRY RUN - Would upload the following videos:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for i, item := range items {
			fmt.Printf("\n%d. %s\n", i+1, item.ID)
			fmt.Printf("   Title: %s\n", item.Platforms.YouTube.Title)
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
		fmt.Printf("        Title: %s\n", item.Platforms.YouTube.Title)

		result, err := media.UploadVideoToYouTube(&item)
		if err != nil {
			log.Printf("❌ Upload failed: %v", err)
			failCount++

			// Update with error
			errMsg := err.Error()
			item.Status.YouTube.Error = &errMsg
			manager.UpdateItem(item)
			continue
		}

		// Update manifest with success
		item.Status.YouTube.Posted = true
		item.Status.YouTube.URL = &result.VideoURL
		item.Status.YouTube.PostedAt = &result.UploadedAt

		if err := manager.UpdateItem(item); err != nil {
			log.Printf("⚠️  Warning: Failed to update manifest: %v", err)
		}

		fmt.Printf("✅ Uploaded: %s\n", result.VideoURL)
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
}

