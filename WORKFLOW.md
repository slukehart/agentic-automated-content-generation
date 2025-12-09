# Content Generation Automation Workflow

## Complete Pipeline Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         MAIN.GO                                 │
│                    (Orchestration Layer)                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │        STEP 1: FETCH NEWS               │
        │   news.ParseNewsArticles()              │
        │                                         │
        │  • Fetches from NewsAPI                │
        │  • Returns article URLs & titles        │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │    STEP 2: GENERATE SUMMARIES           │
        │  news.GenerateBatchNewsReportSummaries()│
        │                                         │
        │  • Sends to Grok AI                    │
        │  • Token-efficient batch processing    │
        │  • Returns 150-200 word summaries      │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │     STEP 3: GENERATE VIDEOS             │
        │   video.GenerateNewsVideo()             │
        │                                         │
        │  For each summary:                      │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │      Go → Python Bridge                 │
        │                                         │
        │  • Marshal summary to JSON              │
        │  • Execute: poetry run python ...       │
        │  • Pass JSON via stdin                  │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │    PYTHON: video_generation.py          │
        │                                         │
        │  • Parse JSON input                     │
        │  • Create cinematic prompt              │
        │  • Submit to FAL AI API                 │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │         FAL AI Processing               │
        │                                         │
        │  • LTX Video Model                     │
        │  • Text-to-Video Generation            │
        │  • Returns video URL                   │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │      Download & Save Video              │
        │                                         │
        │  • Downloads from FAL URL              │
        │  • Saves as news_video_N.mp4           │
        │  • Returns JSON response to Go         │
        └─────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │         OUTPUT FILES                    │
        │                                         │
        │  news_video_1.mp4                      │
        │  news_video_2.mp4                      │
        │  news_video_N.mp4                      │
        └─────────────────────────────────────────┘
```

## Data Flow

### 1. News Article Data
```json
{
  "articles": [
    {
      "title": "Breaking: Major Tech Announcement",
      "url": "https://news.example.com/article1"
    }
  ]
}
```

### 2. AI-Generated Summary
```
"Breaking news from Silicon Valley: A major technology company
has announced a groundbreaking advancement in quantum computing.
The development, unveiled at their annual conference, represents
a significant leap forward in processing capabilities. Industry
experts suggest this could revolutionize data encryption and
artificial intelligence applications within the next decade.
The announcement has already impacted global tech markets,
with shares rising 15% in after-hours trading."
```

### 3. Go → Python Request
```json
{
  "mode": "text_to_video",
  "prompt": "Cinematic news footage: Breaking news from Silicon Valley...",
  "output_path": "news_video_1.mp4",
  "duration": 5,
  "fps": 24
}
```

### 4. Python → Go Response
```json
{
  "status": "success",
  "video_path": "news_video_1.mp4",
  "video_url": "https://fal.ai/files/abc123/video.mp4",
  "duration": 5,
  "fps": 24
}
```

## File Structure

```
content-generation-automation/
├── main.go                      # Main orchestration
├── go.mod                       # Go dependencies
├── pyproject.toml              # Python dependencies
│
├── news/                       # News processing (Go)
│   └── parseNewsArticles.go
│
├── video/                      # Video generation
│   ├── video.go               # Go interface
│   └── video_generation.py    # Python FAL integration
│
├── Generated Output/
│   ├── news_video_1.mp4
│   ├── news_video_2.mp4
│   └── news_video_N.mp4
│
└── Documentation/
    ├── QUICKSTART.md
    ├── VIDEO_SETUP.md
    └── WORKFLOW.md (this file)
```

## Component Communication

### Go Video Package → Python Script

**Method:** Subprocess execution with JSON stdin/stdout

**Go Side:**
```go
cmd := exec.Command("poetry", "run", "python", "video/video_generation.py")
cmd.Stdin = bytes.NewReader(jsonData)
output, err := cmd.Output()
```

**Python Side:**
```python
input_data = json.loads(sys.stdin.read())
# Process...
print(json.dumps(result))
```

**Benefits:**
- Clean separation of concerns
- Language-specific strengths (Go for orchestration, Python for ML APIs)
- Easy to test components independently
- No network overhead

## Error Handling Flow

```
┌─────────────────┐
│  Error Occurs   │
└────────┬────────┘
         │
         ▼
┌────────────────────────────┐
│ Python catches exception   │
│ Returns JSON with error    │
└────────┬───────────────────┘
         │
         ▼
┌────────────────────────────┐
│ Go receives error response │
│ Logs error details         │
│ Increments failedCount     │
└────────┬───────────────────┘
         │
         ▼
┌────────────────────────────┐
│ Continue with next summary │
│ Don't halt entire pipeline │
└────────────────────────────┘
```

## Performance Considerations

### Token Efficiency
- **Batch Processing:** All articles summarized in one API call
- **Reduces:** API overhead and costs
- **System Prompt:** Sent once, not repeated per article

### Video Generation
- **Sequential Processing:** Videos generated one at a time
- **Why:** FAL API rate limits and memory management
- **Future:** Could parallelize with goroutines + rate limiter

### Typical Execution Time
```
10 Articles:
├── Fetch News:        ~2 seconds
├── Generate Summaries: ~10-30 seconds (batch)
├── Generate Videos:    ~60-120 seconds (10 × 6-12s each)
└── Total:             ~72-152 seconds
```

## Extending the Workflow

### Add Audio Narration (Next Step)
```go
// After video generation
audioPath := tts.GenerateNarration(summary)
finalVideo := video.MergeAudioVideo(videoPath, audioPath)
```

### Add Image Generation
```go
// Before video generation
thumbnailPath := image.GenerateFromPrompt(summary)
video := video.GenerateImageToVideo(thumbnailPath, motionPrompt)
```

### Add Social Media Upload
```go
// After video generation
youtube.Upload(videoPath, title, description)
twitter.Post(videoPath, caption)
```

## Configuration Options

### Environment Variables
```bash
FAL_KEY=xxx           # Required: FAL AI API key
NEWS_API_KEY=xxx      # Required: NewsAPI key
GROK_API_KEY=xxx      # Required: Grok AI key
```

### Runtime Options (Future)
```go
type Config struct {
    MaxArticles    int    // Limit articles to process
    VideoDuration  int    // Seconds per video
    VideoQuality   string // "low", "medium", "high"
    SkipVideos     bool   // Generate summaries only
}
```

## Monitoring & Logging

Current logging points:
1. Article fetch count
2. Summary generation progress
3. Each video generation attempt
4. Success/failure counts
5. Final statistics

Output format:
```
[1/10] Generating video: news_video_1.mp4
    Prompt: Breaking news...
    ✅ Saved: news_video_1.mp4
    🔗 URL: https://fal.ai/...
```

## Security Considerations

1. **API Keys:** Stored in environment variables, not code
2. **File Paths:** Sanitized to prevent directory traversal
3. **Input Validation:** JSON parsed safely with error handling
4. **Subprocess Security:** Using exec.Command with explicit args
5. **Dependencies:** Managed through go.mod and poetry.lock

## Testing Strategy

### Unit Tests
```bash
# Test individual components
go test ./news/...
go test ./video/...
```

### Integration Tests
```bash
# Test full pipeline with mock data
go test -tags=integration
```

### Manual Testing
```bash
# Test Python script directly
poetry run python video/video_generation.py text "test" test.mp4

# Test Go package
go run main.go
```

---

**Last Updated:** December 2025
**Version:** 1.0.0

