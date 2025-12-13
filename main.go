package main

import (
	"fmt"
	"log"
	"os"

	"content-generation-automation/audio"
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

	// Define your custom prompt for Grok
	// This prompt is sent ONCE for all articles (token efficient)
	systemPrompt := `You are a professional news writer. Visit each article URL provided, read the article title and the article content, and create a comprehensive news report summarized from the article content suitable for broadcast:
REQUIREMENTS:
- Start with an engaging hook
- Use neutral, factual, authoritative tone
- Each summary must be 150-200 words minimum (for ~60 seconds of speech)
- Include all key points, context, background, and verified facts
- Provide sufficient detail to fully explain the story
- Cite the original source
- No speculation, opinions, or emotional language
- Summary provides a meaningful conclusion to convey the result of the article or what the article is about

VERIFICATION:
If an article is inaccessible or unverifiable, find the story from a reputable outlet (AP, Reuters, BBC, Guardian, Al Jazeera). If no reliable source confirms it, silently exclude it.

OUTPUT:
Provide only completed news reports for verified articles. Each summary should be detailed enough to speak for a maximum of 1 minute.`

	// Fetch news articles from NewsAPI
	fmt.Println("Fetching news articles...")
	articles := news.ParseNewsArticles()

	fmt.Printf("Found %d articles\n\n", articles.ArticleTitle)

	fmt.Println("=== Batch Processing All Articles (Token Efficient) ===")
	summaries, err := news.GenerateBatchNewsReportSummaries(articles, systemPrompt)
	if err != nil {
		log.Fatalf("Error in batch summarization: %v", err)
	}

	fmt.Println("\n=== News Summaries ===")
	for i, summary := range summaries {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(summaries), truncateString(summary, 80))
	}

	// Generate audio + video for each news summary
	fmt.Println("\n=== Generating Audio + Video from Summaries ===")

	successCount := 0
	failedCount := 0

	for i, summary := range summaries {
		baseFilename := fmt.Sprintf("news_%d", i+1)
		audioPath := fmt.Sprintf("%s_audio.mp3", baseFilename)
		finalPath := fmt.Sprintf("%s_final.mp4", baseFilename)

		fmt.Printf("\n[%d/%d] Processing: %s\n", i+1, len(summaries), baseFilename)
		fmt.Printf("    Summary: %s...\n", truncateString(summary, 60))

		// Step 1: Generate audio narration
		fmt.Printf("    🎤 Generating audio narration...\n")
		audioResp, err := audio.GenerateNewsAudio(summary, audioPath)
		if err != nil {
			log.Printf("    ❌ Audio failed: %v", err)
			failedCount++
			continue
		}
		if audioResp.Status != "success" {
			log.Printf("    ❌ Audio error: %s", audioResp.Message)
			failedCount++
			continue
		}
		fmt.Printf("    ✅ Audio: %s\n", audioResp.AudioPath)

		// Step 2: Generate AI avatar video with the audio
		fmt.Printf("    🎬 Generating AI avatar video (this may take 2-3 minutes)...\n")
		videoResp, err := video.GenerateNewsVideo(audioPath, finalPath)
		if err != nil {
			log.Printf("    ❌ Video failed: %v", err)
			if videoResp != nil && videoResp.Message != "" {
				log.Printf("    Details: %s", videoResp.Message)
			}
			// Keep audio file for manual use
			log.Printf("    Note: Audio file saved at %s", audioPath)
			failedCount++
			continue
		}
		if videoResp.Status != "success" {
			log.Printf("    ❌ Video error: %s", videoResp.Message)
			failedCount++
			continue
		}

		// Clean up intermediate audio file (video already includes it)
		os.Remove(audioPath)

		fmt.Printf("    ✅ Final narrated video: %s\n", finalPath)
		if videoResp.VideoURL != "" {
			fmt.Printf("    🔗 HeyGen URL: %s\n", videoResp.VideoURL)
		}
		successCount++
	}

	// Summary
	fmt.Println("\n=== Workflow Complete ===")
	fmt.Printf("📰 News summaries: %d\n", len(summaries))
	fmt.Printf("🎬 Videos with audio: %d\n", successCount)
	if failedCount > 0 {
		fmt.Printf("❌ Failed: %d\n", failedCount)
	}
	fmt.Println("\n✅ Done!")
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
