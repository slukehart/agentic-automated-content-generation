# Video Generation Setup Guide

## Overview
This setup enables Go-to-Python integration for AI video generation using FAL AI's models.

## Architecture
```
Go (main.go) → Python (video_generation.py) → FAL AI API → Generated Video
```

## Prerequisites

1. **Python Environment**
   ```bash
   # Install dependencies
   poetry install
   ```

2. **FAL AI API Key**
   ```bash
   # Get your API key from https://fal.ai/
   export FAL_KEY=your-fal-api-key-here

   # Add to your shell profile (~/.zshrc or ~/.bashrc)
   echo 'export FAL_KEY=your-fal-api-key-here' >> ~/.zshrc
   ```

## Available Models

FAL AI supports several video generation models:

- `fal-ai/ltx-video` - Fast, high-quality text/image-to-video (recommended)
- `fal-ai/fast-animatediff/text-to-video` - Alternative text-to-video
- `fal-ai/fast-svd` - Stable Video Diffusion

## Usage

### 1. From Go Code

```go
import "content-generation-automation/video"

// Text to video
response, err := video.GenerateTextToVideo(
    "A serene mountain landscape at sunrise",
    "output.mp4",
)

// Image to video
response, err := video.GenerateImageToVideo(
    "input.png",
    "The camera zooms in slowly",
    "output.mp4",
)

// News summary to video
response, err := video.GenerateNewsVideo(
    newsSummary,
    "news_video.mp4",
)
```

### 2. Direct Python Script

```bash
# Text to video
poetry run python video/video_generation.py text "your prompt here" output.mp4

# Image to video
poetry run python video/video_generation.py image input.png "motion prompt" output.mp4

# JSON stdin (for Go integration)
echo '{"mode":"text_to_video","prompt":"...","output_path":"output.mp4"}' | \
    poetry run python video/video_generation.py
```

### 3. Test the Integration

```bash
# Run the complete workflow
go run main.go
```

## API Response Format

```json
{
  "status": "success",
  "video_path": "output.mp4",
  "video_url": "https://...",
  "duration": 5,
  "fps": 24
}
```

Error response:
```json
{
  "status": "error",
  "message": "error description"
}
```

## Integration in Your News Workflow

```go
// 1. Parse news articles
articles := news.ParseNewsArticles()

// 2. Generate summaries
summaries, _ := news.GenerateBatchNewsReportSummaries(articles, systemPrompt)

// 3. Generate videos for each summary
for i, summary := range summaries {
    outputPath := fmt.Sprintf("news_%d.mp4", i)
    response, err := video.GenerateNewsVideo(summary, outputPath)
    // Handle response...
}
```

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

### "Python script failed"
- Check that FAL_KEY is set
- Ensure poetry install completed successfully
- Check stderr output for detailed error messages

### Rate Limits
- FAL AI has rate limits on free tier
- Add delays between requests if needed
- Consider upgrading your FAL plan for production use

## Cost Considerations

- Text-to-video: ~$0.05-0.10 per video
- Image-to-video: ~$0.03-0.08 per video
- Check current pricing at https://fal.ai/pricing

## Advanced Configuration

Customize video parameters in Go:

```go
request := video.VideoRequest{
    Mode:       "text_to_video",
    Prompt:     "your prompt",
    OutputPath: "output.mp4",
    Model:      "fal-ai/ltx-video",
    Duration:   10,  // 10 seconds
    FPS:        30,  // 30 fps
}
response, err := video.executeVideoGeneration(request)
```

## Next Steps

1. Test basic video generation
2. Integrate with your news workflow
3. Add audio/narration (see TTS_VIDEO_GUIDE.md)
4. Combine multiple clips into final video

