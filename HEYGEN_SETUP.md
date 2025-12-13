# HeyGen AI Avatar Setup Guide

## Overview

HeyGen provides photorealistic AI avatars that lip-sync perfectly with your audio narration. This guide shows you how to set up HeyGen for creating professional talking head news videos.

## Why HeyGen for News Videos?

✅ **Photorealistic avatars** - Professional news anchor quality
✅ **Perfect lip-sync** - Works with any audio (Google TTS)
✅ **1-2 minute videos** - Perfect for news clips
✅ **Fast generation** - 2-3 minutes per video
✅ **Cost-effective** - ~$0.60-1.00 per video
✅ **API integration** - Easy automation

## Setup Steps

### 1. Create HeyGen Account

1. Go to https://app.heygen.com/
2. Sign up for an account
3. Choose a plan:
   - **Free Trial:** Test with limited credits
   - **Creator Plan:** $29/month (15 minutes video/month)
   - **Business Plan:** $89/month (90 minutes video/month)
   - **Enterprise:** Custom pricing

### 2. Get Your API Key

1. Go to https://app.heygen.com/settings/api
2. Click "Generate API Key"
3. Copy your API key
4. Add to your `.env` file:

```bash
HEYGEN_API_KEY=your-api-key-here
```

### 3. Choose Your Avatar

1. Browse avatars: https://app.heygen.com/avatars
2. Find avatar IDs in the HeyGen dashboard or use these defaults:

**Professional News Anchors:**
- `Kristin_public_3_20240108` - Female, professional (current default)
- `Wayne_20240711` - Male, professional
- `Anna_public_3_20240108` - Female, friendly
- `Josh_lite3_20230714` - Male, casual

**Or create your own custom avatar!**

### 4. Test the Integration

```bash
# Generate a test audio file
poetry run python audio/tts_generation.py "This is a test news report" test_audio.mp3

# Generate avatar video
poetry run python video/video_generation.py test_audio.mp3 test_video.mp4

# Should create a talking head video!
```

## Usage

### From Go (Integrated Workflow)

```go
import "content-generation-automation/video"

// Generate avatar video with audio
response, err := video.GenerateNewsVideo(
    "path/to/audio.mp3",
    "output.mp4",
)
```

### From Python (Direct)

```bash
# Basic usage
poetry run python video/video_generation.py audio.mp3 output.mp4

# With custom avatar
poetry run python video/video_generation.py audio.mp3 output.mp4 Wayne_20240711

# Via JSON stdin (for Go integration)
echo '{"audio_path":"audio.mp3","output_path":"output.mp4","avatar_id":"Kristin_public_3_20240108"}' | \
    poetry run python video/video_generation.py
```

### Run Complete Workflow

```bash
go run main.go
```

This will:
1. Fetch news articles
2. Generate AI summaries
3. Create audio narration (Google TTS)
4. Generate avatar videos (HeyGen)
5. Output professional news videos!

## Customization

### Change Avatar

Edit `video/video.go` (line 55):

```go
func GenerateNewsVideo(audioPath string, outputPath string) (*VideoResponse, error) {
    return GenerateAvatarVideoWithOptions(
        audioPath,
        outputPath,
        "Wayne_20240711",  // Change to male anchor
        "#0e1118",         // Dark studio background
    )
}
```

### Change Background

```go
// Solid color
"#0e1118"  // Dark blue (news studio)

// Or use an image URL
"https://example.com/newsroom-background.jpg"
```

### Custom Avatar Settings

For more control, edit `video/video_generation.py`:

```python
payload = {
    "video_inputs": [{
        "character": {
            "type": "avatar",
            "avatar_id": avatar_id,
            "avatar_style": "normal"  // or "circle"
        },
        "voice": {
            "type": "audio",
            "audio_url": audio_url
        },
        "background": {
            "type": "color",      // or "image"
            "value": background   // color code or image URL
        }
    }],
    "dimension": {
        "width": 1920,   // 1080p
        "height": 1080
    },
    "aspect_ratio": "16:9"
}
```

## Available Avatars

To see all available avatars:

1. Visit: https://app.heygen.com/avatars
2. Or use HeyGen API to list avatars:

```bash
curl -X GET "https://api.heygen.com/v2/avatars" \
  -H "X-Api-Key: YOUR_API_KEY"
```

## Cost & Limits

### Pricing
- **Free Trial:** Limited credits to test
- **Creator Plan ($29/month):**
  - 15 minutes of video/month
  - ~15 news clips (1 min each)
  - Good for testing/small projects

- **Business Plan ($89/month):**
  - 90 minutes of video/month
  - ~90 news clips
  - For regular production

- **Enterprise:** Custom pricing for high volume

### Generation Time
- **1 minute video:** ~2-3 minutes to generate
- **Concurrent:** 1 video at a time (sequential processing)

### File Limits
- **Audio:** Up to 5 minutes
- **Video output:** Up to 5 minutes per clip
- **Format:** MP4, 1080p

## Workflow Details

### Complete Process:

```
News Summary (text)
        ↓
Google Cloud TTS → audio.mp3
        ↓
Upload to HeyGen
        ↓
Create Avatar Video Request
        ↓
HeyGen Processing (2-3 min)
        ↓
Poll for Completion
        ↓
Download Final Video
        ↓
news_1_final.mp4 ✨
```

### Timing Breakdown (10 news articles):

1. Fetch news: ~2 seconds
2. Generate summaries: ~10-30 seconds
3. Generate audio (10 clips): ~10-20 seconds
4. Generate avatars (10 clips): ~20-30 minutes (2-3 min each)
5. **Total:** ~21-31 minutes for 10 complete news videos

## Troubleshooting

### "HEYGEN_API_KEY not set"

```bash
# Add to .env file
echo "HEYGEN_API_KEY=your-key-here" >> .env
```

### "Failed to upload audio"

- Check audio file exists and is readable
- Ensure audio file is valid MP3
- Check file size (< 10MB recommended)

### "Video generation timed out"

- Default timeout is 5 minutes
- For longer videos, increase timeout in `video_generation.py`:
```python
max_attempts = 120  # 10 minutes
```

### "Avatar not found"

- Verify avatar ID is correct
- Check available avatars in your HeyGen dashboard
- Use default: `Kristin_public_3_20240108`

### Poor lip-sync quality

- Ensure audio is clear (Google TTS usually perfect)
- Try different avatar (some sync better than others)
- Check audio isn't too fast (< 2x speed)

## Advanced Features

### Custom Avatars

Create your own avatar:
1. Go to https://app.heygen.com/avatar/custom
2. Upload your photo or record video
3. Wait for processing (~1-2 hours)
4. Use your custom avatar ID in the code

### Multiple Avatars

Rotate between different anchors:

```go
avatars := []string{
    "Kristin_public_3_20240108",
    "Wayne_20240711",
    "Anna_public_3_20240108",
}

avatarID := avatars[i % len(avatars)]
video.GenerateAvatarVideoWithOptions(audio, output, avatarID, background)
```

### Adding Branding

Add your logo or lower third:
- Use background image with your branding
- Or post-process with ffmpeg to overlay graphics

## Best Practices

1. **Audio Quality:** Use Google TTS for consistent, clear audio
2. **Video Length:** Keep clips 1-2 minutes for best results
3. **Avatar Selection:** Choose professional news-style avatars
4. **Background:** Use dark, professional backgrounds for news
5. **Batch Processing:** Process one at a time (HeyGen limitation)
6. **Error Handling:** Keep audio files if video fails for retry

## Next Steps

1. ✅ Test with a single news clip
2. Experiment with different avatars
3. Customize backgrounds and styling
4. Set up automated daily news production
5. Add post-production effects (graphics, music)

---

**Questions?** Check [HeyGen Documentation](https://docs.heygen.com/) or contact support@heygen.com

