# Content Generation Automation

> Automated news-to-video pipeline with AI avatars using Go, Python, Google Cloud TTS, and HeyGen

## Overview

This project automates the complete workflow of fetching news articles, generating AI summaries, creating professional audio narration, and producing photorealistic AI avatar videos with perfect lip-sync.

**Pipeline:**
```
NewsAPI → AI Summarization → Grok Image Gen → HeyGen (TTS + AI Avatar + Dynamic Background)
                                     ↓
                          Talking Head News Video with Custom Newsroom ✨
```

**Technologies:** Go • Python • HeyGen AI Avatars & TTS • JSON-based IPC

## Quick Start

```bash
# 1. Install dependencies
poetry install
go mod download

# 2. Set environment variables in .env file
HEYGEN_API_KEY=your-heygen-api-key    # Get from https://app.heygen.com/settings/api
NEWS_API_KEY=your-newsapi-key         # Get from https://newsapi.org
X_AI_KEY=your-x-ai-key                # Your Grok AI key

# 3. Run the complete workflow
go run main.go
```

## What It Does

1. **Fetches News** - Retrieves latest articles from NewsAPI
2. **Generates Summaries** - Creates broadcast-quality summaries with platform-specific metadata using AI
3. **Generates Custom Backgrounds** - Creates unique newsroom backgrounds for each video using Grok image generation
4. **Generates AI Avatar Videos** - Photorealistic news anchors with built-in text-to-speech, perfect lip-sync, and dynamic backgrounds using HeyGen

### Output

```
news_1_final.mp4  ← AI avatar speaking your news
news_2_final.mp4
news_3_final.mp4
...
```

Each video features a professional AI news anchor with perfect lip-sync, ready to publish!

## Project Structure

```
content-generation-automation/
├── main.go                     # Main orchestration
├── news/                       # News processing (Go)
│   ├── parseNewsArticles.go    # NewsAPI fetching
│   └── metadata_generation.go  # Grok AI summaries + image generation
├── audio/                      # Audio generation (LEGACY - optional)
│   ├── tts.go                  # Go interface
│   └── tts_generation.py       # Google Cloud TTS (use HeyGen TTS instead)
├── video/                      # Video generation with TTS
│   ├── video.go                # Go interface
│   ├── constants.go            # Configuration constants (avatar, voice, dimensions, etc.)
│   └── video_generation.py     # HeyGen AI avatar + TTS integration
├── metadata/                   # Content metadata & manifest
│   ├── types.go                # Data structures for all platforms
│   ├── manifest.go             # Manifest CRUD operations
│   └── prompts.go              # LLM prompts for metadata generation
├── backgrounds/                # Generated newsroom backgrounds
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
                  │  Google   │  │  HeyGen  │
                  │ Cloud TTS │  │ AI Avatar│
                  └─────┬─────┘  └─────┬────┘
                        │              │
                        │  Audio .mp3  │
                        └──────┬───────┘
                               ▼
                        ┌──────────┐
                        │  HeyGen  │
                        │Lip-Sync  │
                        └─────┬────┘
                              ▼
                      news_N_final.mp4
                   (avatar + audio merged)
```

### Go ↔ Python Integration

```
AUDIO GENERATION:
┌─────────┐      JSON       ┌────────────┐      API      ┌──────────────┐
│   Go    │ ────stdin──────→│   Python   │ ───────────→ │ Google Cloud │
│ (main)  │                  │ TTS Script │               │     TTS      │
└─────────┘      JSON       └────────────┘               └──────────────┘
     │          stdout              │
     │                              │
     └────────Audio.mp3─────────────┘

VIDEO GENERATION (AI Avatar):
┌─────────┐      JSON       ┌────────────┐                ┌──────────────┐
│   Go    │ ────stdin──────→│   Python   │   Upload      │   HeyGen     │
│ (main)  │  {audio_path}    │   Video    │────audio─────→│  API Upload  │
└─────────┘                  │   Script   │               └──────┬───────┘
     │                       └─────┬──────┘                      │
     │                             │                             │
     │                             │         Create Video        │
     │                             │←────────with avatar─────────┘
     │                             │
     │                             ▼
     │                      ┌──────────────┐
     │                      │   HeyGen     │
     │                      │  Generates   │
     │                      │Avatar Video  │
     │                      └──────┬───────┘
     │                             │
     │         JSON Response       │
     │        (video_url)          │
     └──────────────┬──────────────┘
                    ▼
            Download & Save
            news_N_final.mp4
```

**Why this architecture?**
- **Go**: Fast orchestration, efficient subprocess management, production-ready
- **Python**: Native access to Google Cloud TTS and HeyGen API SDKs
- **JSON IPC**: Type-safe communication via stdin/stdout between Go and Python
- **HeyGen Integration**: Handles audio upload, avatar rendering, and lip-sync automatically
- **Clean separation**: Each language handles what it does best
- **Cloud-native**: Ready for Google Cloud deployment

## Features

✅ **Photorealistic AI Avatars** - Professional news anchors with HeyGen
✅ **Perfect Lip-Sync** - HeyGen automatically syncs avatar mouth movements to voice
✅ **Built-in Text-to-Speech** - Professional voices directly from HeyGen (no separate TTS needed!)
✅ **Dynamic Backgrounds** - AI-generated unique newsroom backgrounds for each video using Grok
✅ **Portrait Mode Support** - Generate videos in landscape (16:9), portrait (9:16), or square (1:1) aspect ratios
✅ **Speech Speed Control** - Adjust narration speed from 0.5x to 1.5x
✅ **Webhook Support** - Optional webhook integration for long video generation without timeouts
✅ **Centralized Configuration** - Global constants for avatar, voice, dimensions, and styling
✅ **Automated Workflow** - Text → Background Gen → HeyGen TTS + Avatar rendering in one pipeline
✅ **Fast Generation** - 5-20 minutes per video (varies by aspect ratio and length)
✅ **Platform-Optimized Metadata** - Generate YouTube, TikTok, Instagram, Twitter, Facebook, LinkedIn metadata in one LLM call
✅ **Token-Efficient** - Batch processing reduces API calls and costs
✅ **Error Handling** - Failed items don't stop the pipeline
✅ **Progress Tracking** - Real-time console output with status updates
✅ **Cloud-Native** - Built for Google Cloud deployment
✅ **Flexible** - Easy to customize avatars, voices, and styling
✅ **Well-Documented** - Comprehensive guides for every component

## Documentation

- **[QUICKSTART.md](QUICKSTART.md)** - Setup and basic usage
- **[HEYGEN_SETUP.md](HEYGEN_SETUP.md)** - HeyGen AI avatar setup and customization
- **[AUDIO_SETUP.md](AUDIO_SETUP.md)** - Google Cloud TTS setup and voice customization
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
- **HeyGen** - AI avatar videos ([Get key](https://app.heygen.com/settings/api))
- **Google Cloud** - Text-to-Speech API ([Setup guide](AUDIO_SETUP.md))
  - Authenticate with: `gcloud auth application-default login`
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

// Step 1: Generate audio narration with Google Cloud TTS
audioResp, err := audio.GenerateNewsAudio(
    "Breaking news: Major announcement today",
    "news_audio.mp3",
)

// Step 2: Generate AI avatar video with HeyGen (includes lip-sync)
videoResp, err := video.GenerateNewsVideo(
    "news_audio.mp3",  // Pass audio file to HeyGen
    "news_final.mp4",  // HeyGen returns complete video with audio
)

// Note: No manual merging needed! HeyGen handles audio + video integration
```

### Direct Python Usage

```bash
# Generate audio with Google Cloud TTS
poetry run python audio/tts_generation.py "your text" output.mp3

# Generate AI avatar video with HeyGen (using pre-generated audio)
poetry run python video/video_generation.py output.mp3 news_final.mp4

# Or with JSON input for full control
echo '{"audio_path":"output.mp3","output_path":"final.mp4","avatar_id":"Kristin_public_3_20240108"}' | \
  poetry run python video/video_generation.py
```

## Configuration

### Global Video Constants

All default video settings are centralized in `video/constants.go`:

```go
const (
    DefaultAvatarID          = "Angela-inblackskirt-20220820"  // Professional female avatar
    DefaultVoiceID           = "1bd001e7e50f421d891986aad5158bc8"  // Professional female voice
    DefaultBackground        = "newsroom"  // Will generate AI newsroom image
    DefaultBackgroundColor   = "#0e1118"  // Fallback dark blue
    DefaultSpeechSpeed       = 1.0  // Normal speed (0.5 to 1.5)
    DefaultVideoWidth        = 1080  // Portrait mode width
    DefaultVideoHeight       = 1920  // Portrait mode height
    DefaultAspectRatio       = "9:16"  // Portrait mode for social media
)
```

**Popular Configurations:**

**Portrait Mode (TikTok/Instagram Reels):**
- Width: 1080, Height: 1920, Aspect Ratio: "9:16"

**Landscape Mode (YouTube):**
- Width: 1920, Height: 1080, Aspect Ratio: "16:9"

**Square Mode (Instagram Feed):**
- Width: 1080, Height: 1080, Aspect Ratio: "1:1"

### Audio/Voice Settings (LEGACY - Use HeyGen TTS)

Edit `audio/tts.go` (line 58) to customize Google Cloud TTS:
- Voice selection (default: `en-US-Neural2-J` - male professional)
- Speaking speed (default: 1.0)
- See [AUDIO_SETUP.md](AUDIO_SETUP.md) for all available voices

**Note:** We recommend using HeyGen's built-in TTS (text input) instead of separate audio generation for better lip-sync quality.

### Background Generation

Backgrounds are automatically generated using Grok image generation API with prompts like:
```
"A professional modern newsroom background with a blue and white color scheme,
soft lighting, bokeh effect, clean minimalist design, TV studio aesthetic"
```

To customize, edit the prompt in `news/metadata_generation.go` → `GenerateNewsroomBackground`

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
| Video (HeyGen) | ~$6.00-10.00 | 10 videos @ $0.60-1.00 per video credit |
| **Total** | **~$6.17-10.21** | Per complete run |

**Monthly estimate (daily runs):**
- 30 runs × $8.19 average = **~$246/month**
- Or use HeyGen subscription plans:
  - **Creator**: $29/month (15 video credits)
  - **Business**: $89/month (90 video credits - enough for daily 10-clip runs)
  - **Enterprise**: Custom pricing

**Cost optimization tips:**
- Use Google Cloud's free tier (1M characters/month TTS = ~625 news clips)
- Use shorter audio clips to reduce HeyGen processing time
- HeyGen charges by video credit, not duration
- See [TOKEN_OPTIMIZATION_GUIDE.md](TOKEN_OPTIMIZATION_GUIDE.md) for more strategies

## Troubleshooting

### "HEYGEN_API_KEY not set"
```bash
# Add to your .env file
echo "HEYGEN_API_KEY=your-api-key" >> .env
# Get your key from: https://app.heygen.com/settings/api
```

### "Your default credentials were not found" (Google Cloud TTS)
```bash
# Authenticate with Google Cloud
gcloud auth application-default login
```

### "poetry: command not found"
```bash
curl -sSL https://install.python-poetry.org | python3 -
```

### Video generation fails (HeyGen)
- Verify HEYGEN_API_KEY is set in `.env`
- Check API credits/subscription at https://app.heygen.com/
- Review stderr output for detailed errors (Python script logs progress)
- Ensure poetry dependencies are installed: `poetry install`
- Check avatar ID is valid: https://app.heygen.com/avatars
- **Portrait videos take longer!** Landscape: 5-10 min, Portrait: 10-20 min
- For timeouts, consider using webhook support (see `video/video_generation.py`)
- Check video status manually: https://app.heygen.com/videos
- Verify background image was generated in `backgrounds/` folder

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
- [x] AI avatar video generation with HeyGen
- [x] Automated lip-sync (HeyGen handles this)
- [x] Dynamic background generation with Grok AI
- [x] Portrait/landscape/square aspect ratio support
- [x] Speech speed control
- [x] Webhook support for long video generation
- [x] Centralized configuration constants
- [x] Built-in captions (HeyGen native support)
- [ ] Multiple avatar selection/rotation per video
- [ ] Background music and sound effects
- [ ] Multi-clip video assembly (full news programs)
- [ ] Social media auto-upload (YouTube, TikTok, Instagram, etc.)
- [ ] Scheduled automation (cron jobs)
- [ ] Web dashboard for monitoring
- [ ] Multiple video styles/templates
- [ ] Custom avatar creation (HeyGen Photo Avatar)
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

- **HeyGen** - Photorealistic AI avatars with perfect lip-sync
- **Google Cloud Text-to-Speech** - Professional AI narration
- **NewsAPI** - Real-time news data
- **Go** - Fast, efficient orchestration
- **Python** - ML/AI ecosystem integration

---

**Built with ❤️ for automated AI news anchors**

For detailed setup instructions, see [QUICKSTART.md](QUICKSTART.md), [HEYGEN_SETUP.md](HEYGEN_SETUP.md), and [AUDIO_SETUP.md](AUDIO_SETUP.md)

