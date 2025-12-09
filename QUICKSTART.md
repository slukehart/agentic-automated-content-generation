# Quick Start Guide

## Complete News-to-Video Automation Workflow

This project fetches news articles, generates AI summaries, and creates videos from those summaries.

## Setup (One-time)

### 1. Install Dependencies

```bash
# Install Python dependencies
poetry install

# Verify Go modules
go mod download
```

### 2. Set Environment Variables

```bash
# FAL AI API Key (for video generation)
export FAL_KEY=your-fal-api-key

# Add to your shell profile for persistence
echo 'export FAL_KEY=your-fal-api-key' >> ~/.zshrc
source ~/.zshrc
```

Get your FAL API key from: https://fal.ai/dashboard/keys

### 3. Verify Setup

```bash
# Test the installation
./test_video.sh
```

## Running the Complete Workflow

```bash
# Run the full automation pipeline
go run main.go
```

### What It Does:

1. **Fetches News** - Retrieves latest articles from NewsAPI
2. **Generates Summaries** - Creates broadcast-quality summaries using AI
3. **Creates Videos** - Generates cinematic videos for each summary

### Output:

```
Fetching news articles...
Found 10 articles

=== Batch Processing All Articles (Token Efficient) ===
Processing articles with Grok...

=== News Summaries ===
[1/10] Breaking news: Major technological breakthrough announced in quantum...
[2/10] International summit concludes with historic climate agreement that...
...

=== Generating Videos from Summaries ===
[1/10] Generating video: news_video_1.mp4
    Prompt: Breaking news: Major technological breakthrough...
    ✅ Saved: news_video_1.mp4
    🔗 URL: https://fal.ai/files/...

[2/10] Generating video: news_video_2.mp4
    Prompt: International summit concludes with historic...
    ✅ Saved: news_video_2.mp4
    🔗 URL: https://fal.ai/files/...

=== Workflow Complete ===
📰 News summaries: 10
🎥 Videos generated: 10

✅ Done!
```

## Generated Files

After running, you'll have:
- `news_video_1.mp4` through `news_video_N.mp4` - Generated videos

## Configuration

### Customize News Sources

Edit the news fetching in `news/parseNewsArticles.go`:
- Change topics
- Modify article count
- Update sources

### Adjust Video Parameters

Edit `video/video.go` in the `GenerateNewsVideo()` function:
- Duration (default: 5 seconds)
- FPS (default: 24)
- Model selection
- Prompt engineering

### Modify Summary Style

Edit the `systemPrompt` in `main.go`:
- Change tone
- Adjust length
- Modify format

## Troubleshooting

### "FAL_KEY not set"
```bash
export FAL_KEY=your-api-key
```

### "poetry: command not found"
```bash
# Install Poetry
curl -sSL https://install.python-poetry.org | python3 -
```

### Video generation fails
- Check FAL_KEY is set correctly
- Verify you have API credits: https://fal.ai/dashboard
- Check stderr output for detailed errors
- Ensure poetry dependencies are installed

### Rate limiting
- FAL has rate limits on free tier
- Add delays between requests
- Consider upgrading your plan

## Cost Estimation

**Per run with 10 news articles:**
- News summarization: ~$0.01-0.05 (depends on your AI service)
- Video generation: ~$0.50-1.00 (10 videos × $0.05-0.10 each)

**Total:** ~$0.51-1.05 per complete run

Check current pricing:
- FAL: https://fal.ai/pricing
- Your AI summarization service pricing

## Development

### Run individual components:

**Just news summaries (no videos):**
```go
// Comment out the video generation section in main.go
```

**Test video generation directly:**
```bash
# Text to video
poetry run python video/video_generation.py text "your prompt" output.mp4

# Or from Go
go run -tags novideo main.go
```

### Code Structure:
```
.
├── main.go              # Main workflow orchestration
├── news/
│   └── parseNewsArticles.go  # News fetching & summarization
├── video/
│   ├── video.go              # Go video package
│   └── video_generation.py   # Python FAL integration
└── pyproject.toml           # Python dependencies
```

## Next Steps

1. ✅ Run the basic workflow
2. Add audio narration (TTS) to videos
3. Combine multiple clips into one video
4. Add text overlays/captions
5. Schedule automated runs
6. Upload to social media platforms

See `VIDEO_SETUP.md` and `TTS_VIDEO_GUIDE.md` for advanced features.

## Support

- FAL AI Documentation: https://fal.ai/docs
- Issues: Check stderr output for detailed error messages
- API Status: https://status.fal.ai/

