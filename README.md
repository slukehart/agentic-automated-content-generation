# Content Generation Automation

> Automated news-to-video pipeline with AI narration using Go, Python, Google Cloud, and FAL AI

## Overview

This project automates the complete workflow of fetching news articles, generating AI summaries, creating professional audio narration, and producing cinematic videos with synchronized voiceovers.

**Pipeline:**
```
NewsAPI → AI Summarization → Audio (Google Cloud TTS)
                           ↓
                         Video (FAL AI)
                           ↓
                      ffmpeg Merge
                           ↓
                 Narrated News Video ✨
```

**Technologies:** Go • Python • Google Cloud TTS • FAL AI • ffmpeg

## Quick Start

```bash
# 1. Install dependencies
brew install ffmpeg                    # Required for audio/video merging
poetry install
go mod download

# 2. Authenticate with Google Cloud
gcloud auth application-default login

# 3. Set environment variables in .env file
FAL_KEY=your-fal-api-key              # Get from https://fal.ai
NEWS_API_KEY=your-newsapi-key         # Get from https://newsapi.org
GROK_API_KEY=your-grok-key            # Your AI service key

# 4. Run the complete workflow
go run main.go
```

## What It Does

1. **Fetches News** - Retrieves latest articles from NewsAPI
2. **Generates Summaries** - Creates broadcast-quality summaries using AI
3. **Creates Audio Narration** - Professional voiceover using Google Cloud Text-to-Speech
4. **Generates Videos** - Cinematic visuals using FAL AI
5. **Merges Audio + Video** - Combines narration with video using ffmpeg

### Output

```
news_1_final.mp4  ← Fully narrated video
news_2_final.mp4
news_3_final.mp4
...
```

Each video is a professional news clip with synchronized audio narration and cinematic visuals.

## Project Structure

```
content-generation-automation/
├── main.go                     # Main orchestration
├── news/                       # News processing (Go)
│   └── parseNewsArticles.go
├── audio/                      # Audio generation (TTS)
│   ├── tts.go                  # Go interface
│   └── tts_generation.py       # Google Cloud TTS integration
├── video/                      # Video generation
│   ├── video.go                # Go interface
│   └── video_generation.py     # FAL AI integration
└── [documentation files]
```

## Architecture

### Complete Workflow

```
                    ┌──────────────┐
                    │   main.go    │
                    │ (Orchestrator)│
                    └───────┬──────┘
                            │
            ┌───────────────┼───────────────┐
            │               │               │
            ▼               ▼               ▼
    ┌──────────┐    ┌──────────┐   ┌──────────┐
    │  News    │    │  Audio   │   │  Video   │
    │  Fetch   │    │   TTS    │   │   Gen    │
    └──────────┘    └─────┬────┘   └─────┬────┘
                          │              │
                          ▼              ▼
                  ┌───────────┐  ┌──────────┐
                  │  Google   │  │  FAL AI  │
                  │ Cloud TTS │  │   API    │
                  └─────┬─────┘  └─────┬────┘
                        │              │
                        └───────┬──────┘
                                ▼
                          ┌──────────┐
                          │  ffmpeg  │
                          │  Merge   │
                          └─────┬────┘
                                ▼
                        news_N_final.mp4
```

### Go ↔ Python Integration

```
┌─────────┐      JSON       ┌────────────┐      API      ┌──────────────┐
│   Go    │ ────stdin──────→│   Python   │ ───────────→ │ Google Cloud │
│ (main)  │                  │ TTS Script │               │     TTS      │
└─────────┘      JSON       └────────────┘               └──────────────┘
     │          stdout              │
     │                              │
     └──────── Audio ───────────────┘

┌─────────┐      JSON       ┌────────────┐      API      ┌──────────────┐
│   Go    │ ────stdin──────→│   Python   │ ───────────→ │   FAL AI     │
│ (main)  │                  │ Video Script│               │  (LTX Video) │
└─────────┘      JSON       └────────────┘               └──────────────┘
     │          stdout              │
     │                              │
     └──────── Video ───────────────┘
```

**Why this architecture?**
- **Go**: Fast orchestration, efficient subprocess management, production-ready
- **Python**: Access to Google Cloud TTS and FAL AI SDKs
- **Clean separation**: Each language handles what it does best
- **Cloud-native**: Ready for Google Cloud deployment

## Features

✅ **Professional Audio Narration** - Google Cloud Neural2 voices for broadcast-quality speech
✅ **Cinematic Videos** - AI-generated visuals using FAL AI's LTX Video model
✅ **Automated Merging** - Seamless audio/video synchronization with ffmpeg
✅ **Token-Efficient** - Batch processing reduces API calls and costs
✅ **Error Handling** - Failed items don't stop the pipeline
✅ **Progress Tracking** - Real-time console output with status updates
✅ **Cloud-Native** - Built for Google Cloud deployment
✅ **Flexible** - Easy to customize voices, prompts, video parameters
✅ **Well-Documented** - Comprehensive guides for every component

## Documentation

- **[QUICKSTART.md](QUICKSTART.md)** - Setup and basic usage
- **[AUDIO_SETUP.md](AUDIO_SETUP.md)** - Google Cloud TTS setup and voice customization
- **[VIDEO_SETUP.md](VIDEO_SETUP.md)** - FAL AI video generation setup
- **[WORKFLOW.md](WORKFLOW.md)** - Detailed architecture and data flow
- **[TOKEN_OPTIMIZATION_GUIDE.md](TOKEN_OPTIMIZATION_GUIDE.md)** - Cost optimization strategies

## Requirements

### System Requirements
- Go 1.21+
- Python 3.8+
- Poetry (Python package manager)
- ffmpeg (for audio/video merging)
- gcloud CLI (for Google Cloud authentication)

### API Keys & Authentication
- **Google Cloud** - Text-to-Speech API ([Setup guide](AUDIO_SETUP.md))
  - Authenticate with: `gcloud auth application-default login`
- **FAL AI** - For video generation ([Get key](https://fal.ai/dashboard/keys))
- **NewsAPI** - For fetching articles ([Get key](https://newsapi.org/account))
- **Grok/AI Service** - For news summarization

## Usage Examples

### Basic Usage

```bash
# Run complete workflow
go run main.go
```

### Using Audio + Video Packages in Your Code

```go
import (
    "content-generation-automation/audio"
    "content-generation-automation/video"
)

// Generate audio narration
audioResp, err := audio.GenerateNewsAudio(
    "Breaking news: Major announcement today",
    "news_audio.mp3",
)

// Generate video
videoResp, err := video.GenerateNewsVideo(
    "Breaking news: Major announcement today",
    "news_video.mp4",
)

// Merge audio and video
err = audio.MergeAudioVideo(
    "news_video.mp4",
    "news_audio.mp3",
    "news_final.mp4",
)
```

### Direct Python Usage

```bash
# Generate audio with Google Cloud TTS
poetry run python audio/tts_generation.py "your text" output.mp3

# Generate video with FAL AI
poetry run python video/video_generation.py text "your prompt" output.mp4

# Merge with ffmpeg
ffmpeg -i video.mp4 -i audio.mp3 -c:v copy -c:a aac -shortest final.mp4
```

## Configuration

### Audio/Voice Settings

Edit `audio/tts.go` (line 58) to customize:
- Voice selection (default: `en-US-Neural2-J` - male professional)
- Speaking speed (default: 1.0)
- See [AUDIO_SETUP.md](AUDIO_SETUP.md) for all available voices

Popular options:
- `en-US-Neural2-J` - Male, authoritative
- `en-US-Neural2-F` - Female, professional
- `en-US-Studio-M` - Premium male voice
- `en-GB-Neural2-B` - British male

### Video Parameters

Edit `video/video.go` to customize:
- Duration (default: 5 seconds)
- FPS (default: 24)
- FAL model selection
- Prompt engineering for visual style

### News Sources

Edit `news/parseNewsArticles.go` to customize:
- News topics and categories
- Article count
- Source filters
- Language preferences

### AI Summary Prompts

Edit `main.go` `systemPrompt` to customize:
- Summary style and tone
- Length requirements (match to video duration)
- Format and structure
- Fact verification rules

## Testing

```bash
# Test the complete setup
./test_video.sh

# Test individual components
go test ./news/...
go test ./video/...

# Test Python script directly
poetry run python video/video_generation.py
```

## Cost Estimation

**Per run with 10 news articles:**

| Component | Cost per Run | Details |
|-----------|-------------|---------|
| News Fetching | Free | NewsAPI free tier (100 requests/day) |
| AI Summarization | ~$0.01-0.05 | Depends on your AI service |
| Audio (Google TTS) | ~$0.16 | 10K characters @ $16/1M (Neural2 voices) |
| Video (FAL AI) | ~$0.50-1.00 | 10 videos @ $0.05-0.10 each |
| **Total** | **~$0.67-1.21** | Per complete run |

**Monthly estimate (daily runs):**
- 30 runs × $0.94 average = **~$28/month**

**Cost optimization tips:**
- Use Google Cloud's free tier (1M characters/month TTS)
- Batch process to reduce API overhead
- Adjust video duration (shorter = cheaper)
- See [TOKEN_OPTIMIZATION_GUIDE.md](TOKEN_OPTIMIZATION_GUIDE.md) for more strategies

## Troubleshooting

### "FAL_KEY not set"
```bash
# Add to your .env file
echo "FAL_KEY=your-api-key" >> .env
```

### "Your default credentials were not found" (Google Cloud TTS)
```bash
# Authenticate with Google Cloud
gcloud auth application-default login
```

### "ffmpeg not found"
```bash
# Install ffmpeg
brew install ffmpeg
```

### "poetry: command not found"
```bash
curl -sSL https://install.python-poetry.org | python3 -
```

### Audio/Video length mismatch
If audio is longer than video, increase video duration in `video/video.go`:
```go
Duration: 10,  // Change from 5 to 10 seconds
```

### Video generation fails
- Verify FAL_KEY is set in `.env`
- Check API credits at https://fal.ai/dashboard
- Review stderr output for detailed errors
- Ensure poetry dependencies are installed: `poetry install`

### Google Cloud TTS fails
- Verify Text-to-Speech API is enabled
- Check authentication: `gcloud auth list`
- Verify project: `gcloud config get-value project`
- See [AUDIO_SETUP.md](AUDIO_SETUP.md) for detailed troubleshooting

### Build errors
```bash
go mod tidy
go mod download
```

## Roadmap

- [x] Audio narration with Google Cloud TTS
- [x] Audio/video synchronization with ffmpeg
- [ ] Background music and sound effects
- [ ] Text overlays and captions (subtitles)
- [ ] Multi-clip video assembly (full news programs)
- [ ] Social media auto-upload (YouTube, Twitter, etc.)
- [ ] Scheduled automation (cron jobs)
- [ ] Web dashboard for monitoring
- [ ] Multiple video styles/templates
- [ ] Google Cloud deployment (Cloud Run/Functions)

## Contributing

Feel free to:
- Report issues
- Suggest improvements
- Submit pull requests
- Share your use cases

## License

[Your License Here]

## Acknowledgments

- **Google Cloud Text-to-Speech** - Professional AI narration
- **FAL AI** - Advanced video generation (LTX Video model)
- **NewsAPI** - Real-time news data
- **Go** - Fast, efficient orchestration
- **Python** - ML/AI ecosystem integration
- **ffmpeg** - Audio/video processing

---

**Built with ❤️ for automated news content creation**

For detailed setup instructions, see [QUICKSTART.md](QUICKSTART.md) and [AUDIO_SETUP.md](AUDIO_SETUP.md)

