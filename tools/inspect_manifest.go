package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"content-generation-automation/metadata"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tools/inspect_manifest.go [command]")
		fmt.Println("\nCommands:")
		fmt.Println("  list              - List all content items")
		fmt.Println("  show <id>         - Show details for a specific item")
		fmt.Println("  stats             - Show manifest statistics")
		fmt.Println("  youtube <id>      - Show YouTube metadata for item")
		fmt.Println("  tiktok <id>       - Show TikTok metadata for item")
		fmt.Println("  instagram <id>    - Show Instagram metadata for item")
		fmt.Println("  twitter <id>      - Show Twitter metadata for item")
		fmt.Println("  facebook <id>     - Show Facebook metadata for item")
		fmt.Println("  linkedin <id>     - Show LinkedIn metadata for item")
		fmt.Println("  unposted          - List items not posted to any platform")
		return
	}

	manager := metadata.NewManifestManager("content_manifest.json")
	command := os.Args[1]

	switch command {
	case "list":
		listItems(manager)
	case "show":
		if len(os.Args) < 3 {
			log.Fatal("Please provide item ID")
		}
		showItem(manager, os.Args[2])
	case "stats":
		showStats(manager)
	case "youtube", "tiktok", "instagram", "twitter", "facebook", "linkedin":
		if len(os.Args) < 3 {
			log.Fatal("Please provide item ID")
		}
		showPlatformMetadata(manager, os.Args[2], command)
	case "unposted":
		showUnposted(manager)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func listItems(manager *metadata.ManifestManager) {
	items, err := manager.GetAllItems()
	if err != nil {
		log.Fatalf("Error loading items: %v", err)
	}

	if len(items) == 0 {
		fmt.Println("No items in manifest")
		return
	}

	fmt.Printf("Found %d items:\n\n", len(items))
	for i, item := range items {
		fmt.Printf("[%d] %s\n", i+1, item.ID)
		fmt.Printf("    Created: %s\n", item.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("    Source: %s\n", item.Source.Title)
		fmt.Printf("    Video: %s\n", item.Media.VideoPath)
		fmt.Printf("    Posted: YT=%v TT=%v IG=%v TW=%v FB=%v LI=%v\n\n",
			item.Status.YouTube.Posted,
			item.Status.TikTok.Posted,
			item.Status.Instagram.Posted,
			item.Status.Twitter.Posted,
			item.Status.Facebook.Posted,
			item.Status.LinkedIn.Posted,
		)
	}
}

func showItem(manager *metadata.ManifestManager, id string) {
	item, err := manager.GetItem(id)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	data, _ := json.MarshalIndent(item, "", "  ")
	fmt.Println(string(data))
}

func showStats(manager *metadata.ManifestManager) {
	stats, err := manager.GetStats()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	data, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(data))
}

func showPlatformMetadata(manager *metadata.ManifestManager, id, platform string) {
	item, err := manager.GetItem(id)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("=== %s Metadata for %s ===\n\n", platform, id)

	switch platform {
	case "youtube":
		fmt.Printf("Title: %s\n\n", item.Platforms.YouTube.Title)
		fmt.Printf("Description:\n%s\n\n", item.Platforms.YouTube.Description)
		fmt.Printf("Tags: %v\n\n", item.Platforms.YouTube.Tags)
		if len(item.Platforms.YouTube.Timestamps) > 0 {
			fmt.Println("Timestamps:")
			for _, ts := range item.Platforms.YouTube.Timestamps {
				fmt.Printf("  %s - %s\n", ts.Time, ts.Label)
			}
		}
	case "tiktok":
		fmt.Printf("Caption:\n%s\n\n", item.Platforms.TikTok.Caption)
		fmt.Printf("Hashtags: %v\n", item.Platforms.TikTok.Hashtags)
	case "instagram":
		fmt.Printf("Caption:\n%s\n\n", item.Platforms.Instagram.Caption)
		fmt.Printf("Hashtags: %v\n", item.Platforms.Instagram.Hashtags)
	case "twitter":
		fmt.Printf("Tweet:\n%s\n\n", item.Platforms.Twitter.Tweet)
		fmt.Printf("Hashtags: %v\n", item.Platforms.Twitter.Hashtags)
	case "facebook":
		fmt.Printf("Message:\n%s\n\n", item.Platforms.Facebook.Message)
		fmt.Printf("Link Description: %s\n", item.Platforms.Facebook.LinkDescription)
	case "linkedin":
		fmt.Printf("Post:\n%s\n\n", item.Platforms.LinkedIn.PostText)
		fmt.Printf("Hashtags: %v\n", item.Platforms.LinkedIn.Hashtags)
	}

	// Show posting status
	var status metadata.PlatformStatus
	switch platform {
	case "youtube":
		status = item.Status.YouTube
	case "tiktok":
		status = item.Status.TikTok
	case "instagram":
		status = item.Status.Instagram
	case "twitter":
		status = item.Status.Twitter
	case "facebook":
		status = item.Status.Facebook
	case "linkedin":
		status = item.Status.LinkedIn
	}

	fmt.Printf("\nPosting Status: Posted=%v\n", status.Posted)
	if status.URL != nil {
		fmt.Printf("URL: %s\n", *status.URL)
	}
	if status.Error != nil {
		fmt.Printf("Error: %s\n", *status.Error)
	}
}

func showUnposted(manager *metadata.ManifestManager) {
	items, err := manager.GetAllItems()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Items not posted to any platform:\n")
	count := 0
	for _, item := range items {
		if !item.Status.YouTube.Posted &&
			!item.Status.TikTok.Posted &&
			!item.Status.Instagram.Posted &&
			!item.Status.Twitter.Posted &&
			!item.Status.Facebook.Posted &&
			!item.Status.LinkedIn.Posted {
			count++
			fmt.Printf("[%d] %s\n", count, item.ID)
			fmt.Printf("    Video: %s\n", item.Media.VideoPath)
			fmt.Printf("    YouTube Title: %s\n\n", item.Platforms.YouTube.Title)
		}
	}

	if count == 0 {
		fmt.Println("All items have been posted to at least one platform")
	}
}

