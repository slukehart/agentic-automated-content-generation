package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"content-generation-automation/video"

	"github.com/joho/godotenv"
)

// Local test harness for LongCat drift experiments. It trims an existing
// narration WAV and renders a short clip, so each experiment costs about $1
// instead of a full run. It does not run news, Grok metadata or Kokoro.
func main() {
	image := flag.String("image", "backgrounds/newsroom_20260930_202332.jpg", "avatar image with background baked in")
	audio := flag.String("audio", "audio/news_20260930_202332.wav", "narration WAV")
	seconds := flag.Float64("seconds", 26, "trim narration to this many seconds (26s is 8 segments)")
	prompt := flag.String("prompt", "", "scene prompt (default: video.DefaultPrompt)")
	label := flag.String("label", "test", "name for the output file in output/")
	refImgIndex := flag.Int("ref-img-index", -1, "LongCat ref_img_index, -1 keeps the default")
	maskFrameRange := flag.Int("mask-frame-range", -1, "LongCat mask_frame_range, -1 keeps the default")
	flag.Parse()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	if err := os.MkdirAll("output", 0755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}
	trimmed := filepath.Join("output", fmt.Sprintf("test_%s.wav", *label))
	trim := exec.Command("ffmpeg", "-v", "error", "-y", "-i", *audio, "-t", fmt.Sprint(*seconds), trimmed)
	trim.Stderr = os.Stderr
	if err := trim.Run(); err != nil {
		log.Fatalf("trim audio: %v", err)
	}

	opts := video.Options{Prompt: *prompt}
	if *refImgIndex >= 0 {
		opts.RefImgIndex = refImgIndex
	}
	if *maskFrameRange >= 0 {
		opts.MaskFrameRange = maskFrameRange
	}

	out := filepath.Join("output", fmt.Sprintf("test_%s.mp4", *label))
	fmt.Printf("rendering %s (%.0fs of audio)\n", out, *seconds)
	resp, err := video.GenerateAvatarVideoWithOptions(*image, trimmed, out, opts)
	if err != nil {
		if resp != nil && resp.Message != "" {
			log.Printf("details: %s", resp.Message)
		}
		log.Fatalf("render failed: %v", err)
	}

	fmt.Printf("saved to: %s\n", out)
	if resp.Applied == nil {
		fmt.Println("handler reported no applied args (older image without the pass-through)")
	} else {
		fmt.Printf("handler applied: %v\n", resp.Applied)
	}
}
