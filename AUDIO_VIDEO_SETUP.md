# Audio + Video Setup Guide

## Overview

This guide shows you how to create **narrated news videos** by combining:
1. **Text-to-Speech (TTS)** - Audio narration from news summaries
2. **Text-to-Video** - Visual content from FAL AI
3. **Audio/Video Merging** - Combined final product with ffmpeg

## Prerequisites

### 1. Install Python Dependencies

```bash
poetry install
```

### 2. Install ffmpeg (Required for merging)

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt update && sudo apt install ffmpeg

# Windows
# Download from: https://ffmpeg.org/download.html
```

Verify installation:
```bash
ffmpeg -version
```

### 3. Set Up API Keys

Add to your `.env` file:

```bash
# Video generation
FAL_KEY=your-fal-api-key

# Audio/TTS (choose one or more)
OPENAI_API_KEY=your-openai-key        # Recommended: Good quality, reasonable cost
ELEVENLABS_API_KEY=your-elevenlabs-key  # Optional: Best quality, higher cost
GOOGLE_APPLICATION_CREDENTIALS=/path/to/google-credentials.json  # Optional: Google Cloud TTS

# News and AI
NEWS_API_KEY=your-newsapi-key
GROK_API_KEY=your-grok-key
```

## TTS Provider Comparison

| Provider | Cost | Quality | Setup | Recommended For |
|----------|------|---------|-------|----------------|
| **OpenAI TTS** | $15/1M chars | Very Good | Easy | General use, good balance |
| **ElevenLabs** | $0.30/1K chars | Excellent | Easy | Premium quality needed |
| **Google Cloud** | $4/1M chars | Good | Moderate | GCP deployments |
| **gTTS** | Free | Basic | Easy | Testing, prototypes |

### Default Configuration

The system uses **OpenAI TTS** by default with the "alloy" voice (professional, neutral).

## Usage

### Run Complete Workflow

```bash
go run main.go
```

This will:
1. Fetch news articles
2. Generate AI summaries
3. Create audio narration for each summary
4. Generate videos for each summary
5. Merge audio + video into final narrated videos
6. Clean up intermediate files

### Output Files

```
news_1_final.mp4  ← Final video with audio
news_2_final.mp4
news_3_final.mp4
...
```

### Customize TTS Provider

Edit `audio/tts.go` line 58-63:

```go
func GenerateNewsAudio(newsSummary string, outputPath string) (*TTSResponse, error) {
    return GenerateAudioWithOptions(
        newsSummary,
        outputPath,
        "openai",    // Change to: "elevenlabs", "google", or "gtts"
        "alloy",     // Voice ID (provider-specific)
        1.0,         // Speed (0.5 = slower, 2.0 = faster)
    )
}
```

### Voice Options

**OpenAI Voices:**
- `alloy` - Neutral, professional (default)
- `echo` - Male, clear
- `fable` - Warm, narrative
- `onyx` - Deep, authoritative
- `nova` - Friendly, energetic
- `shimmer` - Soft, calm

**ElevenLabs:**
- Use voice IDs from your ElevenLabs account
- Can clone custom voices

**Google Cloud:**
- `en-US-Neural2-J` - Male, professional
- `en-US-Neural2-F` - Female, professional
- See full list: https://cloud.google.com/text-to-speech/docs/voices

## Manual Testing

### Test Audio Generation Directly

```bash
# Using Python script
poetry run python audio/tts_generation.py openai "This is a test" test_audio.mp3

# Using Go package (create test file)
go run -e 'package main; import "content-generation-automation/audio"; func main() { audio.GenerateAudio("Test", "test.mp3") }'
```

### Test Video Generation

```bash
poetry run python video/video_generation.py text "A serene landscape" test_video.mp4
```

### Test Merging

```bash
ffmpeg -i test_video.mp4 -i test_audio.mp3 -c:v copy -c:a aac -shortest output.mp4
```

## Workflow Customization

### Option 1: Audio Only (No Video)

Comment out video generation in `main.go`:

```go
// Skip video generation
// videoResp, err := video.GenerateNewsVideo(summary, videoPath)
// Just generate audio
audioResp, err := audio.GenerateNewsAudio(summary, audioPath)
```

### Option 2: Video Only (No Audio)

Comment out audio generation in `main.go`:

```go
// Skip audio
// Just generate silent video
videoResp, err := video.GenerateNewsVideo(summary, videoPath)
```

### Option 3: Custom Audio for Existing Video

```go
// Use your own video file
err := audio.MergeAudioVideo("my_video.mp4", "my_audio.mp3", "final.mp4")
```

## Cost Estimation

### Per 10 News Articles (assuming 200 words each):

**Audio (OpenAI TTS):**
- 200 words × 10 articles = ~2,000 words
- ~10,000 characters
- Cost: $0.15

**Video (FAL AI):**
- 10 videos × 5 seconds
- Cost: ~$0.50-1.00

**Total per run:** ~$0.65-1.15

### Monthly Costs (Daily runs):

- 30 runs/month × $0.85 average = ~$25.50/month
- Add ~20% buffer for retries = ~$30/month

## Troubleshooting

### "ffmpeg not found"

```bash
# Install ffmpeg
brew install ffmpeg  # macOS
sudo apt install ffmpeg  # Linux
```

### "OPENAI_API_KEY not set"

Add to your `.env` file:
```bash
OPENAI_API_KEY=sk-...your-key...
```

### Audio/Video out of sync

The code uses `-shortest` flag in ffmpeg, which ends the video when the shortest stream (audio or video) ends. If your audio is longer than your video:

1. Increase video duration in `video/video.go` (line 40):
```go
Duration:   10,  // Change from 5 to 10 seconds
```

2. Or adjust TTS speed to make audio shorter:
```go
Speed: 1.5,  // 1.5x faster speech
```

### Poor audio quality

Switch to a better TTS provider:

```go
// In audio/tts.go GenerateNewsAudio()
return GenerateAudioWithOptions(
    newsSummary,
    outputPath,
    "elevenlabs",  // Higher quality
    "default-voice",
    1.0,
)
```

### Video generation takes too long

Each video takes 30-60 seconds to generate. For 10 videos, expect 5-10 minutes total.

**Speed improvements:**
1. Reduce video duration (5s → 3s)
2. Process fewer articles
3. Use faster FAL models (if available)
4. Consider batch processing in background

## Production Deployment (Google Cloud)

### Cloud Run Setup

1. **Install dependencies in container:**
```dockerfile
RUN apt-get update && apt-get install -y ffmpeg
```

2. **Use Secret Manager for API keys:**
```go
// Load from Secret Manager instead of .env
apiKey := getSecretFromGCP("OPENAI_API_KEY")
```

3. **Cloud Storage for output:**
```go
// Upload to GCS after generation
uploadToGCS(finalPath, "my-bucket", "videos/")
```

### Cloud Functions

For serverless deployment:
- Split workflow into separate functions
- Use Cloud Tasks for orchestration
- Store intermediate files in Cloud Storage

See deployment guide for full details.

## Advanced Features

### Custom Voice Cloning (ElevenLabs)

1. Clone your voice on ElevenLabs
2. Get the voice ID
3. Use it in the code:

```go
return GenerateAudioWithOptions(
    text,
    output,
    "elevenlabs",
    "your-cloned-voice-id",
    1.0,
)
```

### Background Music

Add background music using ffmpeg:

```bash
ffmpeg -i video_with_narration.mp4 -i background_music.mp3 \
  -filter_complex "[1:a]volume=0.2[bg];[0:a][bg]amix=inputs=2[a]" \
  -map 0:v -map "[a]" -c:v copy -c:a aac output.mp4
```

### Subtitles/Captions

Generate subtitles from the audio:
1. Use Whisper API to transcribe audio
2. Generate SRT/VTT file
3. Burn into video with ffmpeg

## Next Steps

1. ✅ Test the audio + video workflow
2. Try different TTS providers and voices
3. Adjust video duration to match audio length
4. Add custom styling or branding
5. Set up automated deployment

For deployment to Google Cloud, see the deployment guide.

---

**Questions or issues?** Check the main README.md or create an issue.

