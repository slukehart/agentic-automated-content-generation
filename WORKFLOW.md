# Content Generation Automation Workflow

This file briefly documents **how `main.go` executes** and which key functions are called in order. It is intentionally concise and focused on the control flow.

---

## Execution Steps (`go run main.go`)

1. **Load configuration**
   - Load `.env` (if present) using `godotenv`.
   - Verify required environment variables (e.g. `HEYGEN_API_KEY`) are set.

2. **Initialize manifest manager**
   - Call `metadata.NewManifestManager("content_manifest.json")` to work with the local manifest file.

3. **Fetch one news article**
   - Call `news.ParseNewsArticles()`:
     - Uses **NewsAPI** to fetch recent articles.
     - Selects a single article (URL, title, source, published time, etc.).

4. **Generate summary + multi-platform metadata**
   - Call `news.GenerateEnrichedNewsContent(article)`:
     - Uses an LLM (Grok/xAI) to create:
       - A 150–200 word broadcast-style summary.
       - SEO metadata (keywords, topics, sentiment).
       - Platform-specific metadata for YouTube, TikTok, Instagram, Twitter/X, Facebook, LinkedIn.

5. **Generate newsroom background image**
   - Call `news.GenerateNewsroomBackground("")`:
     - Uses Grok/xAI image generation to create a newsroom-style background.
     - Returns `nil` on failure; in that case, the system falls back to defaults in `video/constants.go`.

6. **Generate AI avatar video with HeyGen**
   - Build a unique `contentID` and `finalPath` (e.g. `news_YYYYMMDD_HHMMSS_final.mp4`).
   - Depending on whether a background image exists:
     - Call `video.GenerateNewsVideoWithBackgroundImage(summary, finalPath, backgroundPath)`, **or**
     - Call `video.GenerateNewsVideoFromText(summary, finalPath)`.
   - Internally, these functions:
     - Launch `video/video_generation.py` via `poetry run python`.
     - That script calls the **HeyGen API** (TTS + avatar + optional background), polls until complete, and downloads the final MP4 to `finalPath`.

7. **Save result to manifest**
   - Call `news.ConvertToContentItem(...)` to build a `metadata.ContentItem`:
     - Includes source article info, summary, SEO, per-platform metadata, and media info (video path, duration, avatar ID).
   - Call `manifestManager.AddItem(contentItem)` to append this item to `content_manifest.json`.

8. **Optionally upload to platforms**
   - If `-upload-youtube` flag is set:
     - Call `media.UploadVideoToYouTube(&contentItem)` to upload the MP4 as a YouTube Short and update the manifest with the resulting URL and status.
   - If `-upload-tiktok` flag is set:
     - Call `media.UploadVideoToTikTok(&contentItem, useSandbox)` (where `useSandbox` is controlled by the `-sandbox` flag).
     - This uses TikTok Login Kit tokens and the Content Posting API to upload the video **to the user’s TikTok inbox**, then stores the `publish_id` and timestamp in the manifest.

9. **Print summary**
   - Log the content ID, video file path, manifest path, and any successful upload URLs.

---

## Pipeline Diagram (High-Level)

```text
┌──────────────────────────────────────────────────────────┐
│                      main.go (CLI)                      │
└───────────────┬───────────────────────────────┬─────────┘
                │                               │
                │                               │
                ▼                               ▼
      ┌───────────────────┐             ┌─────────────────────┐
      │  news.ParseNews…  │             │ ManifestManager     │
      │  (NewsAPI)        │             │ (content_manifest)  │
      └─────────┬─────────┘             └─────────┬───────────┘
                │                                 │
                ▼                                 │
      ┌───────────────────┐                       │
      │ GenerateEnriched… │  (summary + metadata) │
      └─────────┬─────────┘                       │
                │                                 │
                ▼                                 │
      ┌───────────────────┐                       │
      │ GenerateNewsroom… │  (background image)   │
      └─────────┬─────────┘                       │
                │                                 │
                ▼                                 │
      ┌─────────────────────────────┐             │
      │  video.GenerateNewsVideo…  │             │
      │  → video_generation.py     │             │
      │  → HeyGen (TTS + avatar)   │             │
      └─────────┬──────────────────┘             │
                │                                 │
                ▼                                 ▼
        ┌────────────────────┐          ┌─────────────────────┐
        │ news_*_final.mp4   │  ───▶    │  ContentItem in     │
        │ (local video file) │          │  content_manifest   │
        └─────────┬──────────┘          └─────────┬───────────┘
                  │                                 │
                  ▼                                 ▼
        ┌────────────────────┐          ┌─────────────────────┐
        │ media.UploadVideo… │  (opt)   │ Updated posting     │
        │  • YouTube Shorts  │          │ status per platform │
        │  • TikTok inbox    │          └─────────────────────┘
        └────────────────────┘
```

# Content Generation Automation Workflow

This document describes the **current production workflow** for this repository:

- How a single `go run main.go` invocation executes end-to-end.
- How Go, Python, and external APIs (NewsAPI, Grok/xAI, HeyGen, YouTube, TikTok) interact.
- How metadata and posting status are stored locally in `content_manifest.json`.

It supersedes the older FAL/text‑to‑video architecture and focuses on the **HeyGen TTS + avatar** pipeline with manifest + upload helpers.

---

## 1. End-to-End Flow (Single Run of `main.go`)

The main entry point is `main.go` in the repo root.

### 1.1 High-level sequence

```text
NewsAPI → LLM (summary + metadata) → Grok image → HeyGen (TTS + avatar) → MP4 + manifest
                                                                                  ↓
                                                             Optional: YouTube + TikTok upload
```

In order:

1. **Environment & config**
   - Load `.env` if present (`godotenv`).
   - Validate `HEYGEN_API_KEY` (fatal if missing).

2. **Manifest manager**
   - Instantiate `metadata.NewManifestManager("content_manifest.json")`.
   - Ensures a manifest file exists or will be created.

3. **Fetch a single news article**
   - Call `news.ParseNewsArticles()`:
     - Talks to **NewsAPI** using `NEWS_API_KEY`.
     - Returns a single selected article: URL, title, publisher, published_at, etc.

4. **Generate enriched content (LLM)**
   - Call `news.GenerateEnrichedNewsContent(article)`:
     - Uses a Grok/xAI LLM to produce, in **one call**:
       - A **150–200 word broadcast-style summary** (target ≈ 60s of speech).
       - SEO metadata (keywords, topics, sentiment, target audience).
       - Per‑platform metadata for:
         - YouTube (title, description, tags, timestamps, etc.)
         - TikTok, Instagram, Twitter/X, Facebook, LinkedIn (captions, hashtags, etc.).
     - Returns a strongly typed `EnrichedContent` object.

5. **Generate newsroom background (image)**
   - Call `news.GenerateNewsroomBackground(promptOverride string)`:
     - If `promptOverride` is empty, uses a default prompt like:
       > “A professional modern newsroom background with soft lighting…”
     - Calls Grok/xAI’s image model.
     - Saves an image (e.g., PNG/JPEG) to `backgrounds/`.
     - Returns a struct with `ImagePath` or `nil` on failure.
   - On failure:
     - Log warning.
     - Fall back to default background settings baked into `video/constants.go`.

6. **Generate AI avatar video with HeyGen**
   - Compute:
     - `contentID := "news_" + timestamp`
     - `finalPath := contentID + "_final.mp4"`
   - Display summary snippet and background info in the terminal.
   - Decide which video function to call:
     - If background image present:
       - `video.GenerateNewsVideoWithBackgroundImage(enrichedContent.Summary, finalPath, backgroundResult.ImagePath)`
     - Otherwise:
       - `video.GenerateNewsVideoFromText(enrichedContent.Summary, finalPath)`
   - These functions:
     - Marshal a JSON request with:
       - Text (the summary).
       - Output file path (`finalPath`).
       - Optional `background_image_path`.
       - Video configuration (`DefaultVideoWidth`, `DefaultVideoHeight`, `DefaultAspectRatio`, etc.).
     - Run:

       ```bash
       poetry run python video/video_generation.py
       ```

     - Read JSON response from stdout into a `VideoResponse` struct:
       - `Status`, `Message`
       - `VideoURL` (HeyGen-hosted)
       - `Duration`
   - On error or non‑success status:
     - Log detailed information.
     - Abort the run (`log.Fatalf`) because the video is required.

7. **Persist to manifest**

   - Use `news.ConvertToContentItem(...)` to map:
     - Source article → `Source` struct.
     - Enriched summary + SEO → `Content` and `SEO`.
     - Per-platform metadata → `Platforms` sub-structs.
     - Video info:
       - `Media.VideoPath = finalPath`
       - `Media.DurationSeconds = videoResp.Duration`
       - `Media.AvatarID = video.DefaultAvatarID` (current default avatar).
   - Call:

   ```go
   if err := manifestManager.AddItem(contentItem); err != nil {
       log.Printf("Warning: Failed to save to manifest: %v", err)
   }
   ```

   - Failures here are **non-fatal**; they don’t invalidate the video.

8. **Console metadata preview**

   - Print a cross-platform snapshot:
     - YouTube title
     - TikTok caption (truncated)
     - Instagram hashtag count
     - Twitter/X tweet seed (truncated)

9. **Optional: YouTube upload**

   - If `-upload-youtube` flag is set:
     - Call `media.UploadVideoToYouTube(&contentItem)`:
       - Uses YouTube Data API v3 with OAuth.
       - Uploads MP4 as a Short (portrait) with:
         - Title, description, tags from `contentItem.Platforms.YouTube`.
         - `#Shorts` tag and timestamps.
         - Source link and AI disclosure block.
       - Receives `VideoID` and constructs a public URL.
     - On success:
       - Update manifest:
         - `Status.YouTube.Posted = true`
         - `Status.YouTube.URL = videoURL`
         - `Status.YouTube.PostedAt = UploadedAt`
     - On failure:
       - Log error and set `Status.YouTube.Error` (where applicable).

10. **Optional: TikTok upload (inbox only)**

    - If `-upload-tiktok` flag is set:
      - Determine environment:
        - `-sandbox` → use sandbox client credentials.
        - No `-sandbox` → production credentials.
      - Call `media.UploadVideoToTikTok(&contentItem, useSandbox)`:
        - Retrieves TikTok OAuth access token from `~/.credentials/tiktok-oauth.json` (Login Kit).
        - Initializes upload via Content Posting API (`/v2/post/publish/inbox/video/init/`).
        - Uploads MP4 via returned `upload_url` (single-chunk for typical size).
        - Returns a `TikTokUploadResult` with `PublishID` and `UploadedAt`.
      - Update manifest:
        - `Status.TikTok.Posted = false` (uploaded to **inbox**, not feed).
        - `Status.TikTok.PostedAt = UploadedAt`.
        - `Status.TikTok.URL = "inbox_uploaded (publish_id: ...)"`.
      - Errors:
        - Logged and stored in `Status.TikTok.Error` when invoked through tools.

11. **End-of-run summary**

    - Print:
      - Content ID.
      - Video file path.
      - Manifest path.
      - Any successfully posted URLs.

---

## 2. Architecture & Packages

### 2.1 Package overview

```text
content-generation-automation/
├── main.go
├── news/
│   ├── parseNewsArticles.go        # NewsAPI integration & article selection
│   └── metadata_generation.go      # LLM + background image generation
├── video/
│   ├── video.go                    # Go ↔ Python bridge for HeyGen
│   ├── constants.go                # Default avatar, aspect ratio, etc.
│   └── video_generation.py         # HeyGen API client (Python)
├── metadata/
│   ├── types.go                    # Manifest + platform metadata structs
│   ├── manifest.go                 # ManifestManager implementation
│   └── prompts.go                  # Prompt for summary + metadata
├── media/
│   ├── youtubeController.go        # YouTube Shorts upload
│   ├── tiktokController.go         # TikTok Content Posting API (inbox)
│   ├── auth.go                     # OAuth helpers (shared)
│   └── tiktok_auth.go              # TikTok Login Kit specifics
├── tools/
│   ├── inspect_manifest.go         # CLI for exploring content_manifest.json
│   └── upload_tiktok/main.go       # CLI for batch/dry-run TikTok upload
└── content_manifest.json           # Generated manifest file
```

### 2.2 External services

- **NewsAPI** – source of public news articles.
- **Grok/xAI (LLM)** – summarization + multi-platform metadata.
- **Grok/xAI (image)** – newsroom background generation.
- **HeyGen** – text-to-speech + avatar + background → rendered MP4.
- **YouTube Data API v3** – optional Shorts upload.
- **TikTok Login Kit + Content Posting API** – optional inbox upload.

---

## 3. Data Flow & Storage

### 3.1 Manifest: `content_manifest.json`

Every generated item is stored as a `ContentItem` in `content_manifest.json`. Conceptually:

```json
{
  "version": "1.0",
  "generated_at": "2026-02-12T22:56:22.881517-05:00",
  "items": [
    {
      "id": "news_20260210_221136",
      "created_at": "2026-02-10T22:19:36.439983-05:00",
      "source": {
        "url": "https://www.example.com/article",
        "title": "Article headline…",
        "published_at": "2026-02-10T21:00:00Z",
        "source_name": "example.com"
      },
      "content": {
        "summary": "150–200 word broadcast summary…",
        "word_count": 178,
        "estimated_duration_seconds": 71
      },
      "media": {
        "video_path": "news_20260210_221136_final.mp4",
        "duration_seconds": 53.1,
        "resolution": "1280x720",
        "avatar_id": "Marcus_expressive_2024120201"
      },
      "seo": {
        "primary_keywords": ["keyword1", "keyword2"],
        "topics": ["Politics", "International"],
        "sentiment": "neutral"
      },
      "platforms": {
        "youtube": {
          "title": "Optimized YouTube title…",
          "description": "Long-form description…",
          "tags": ["tag1", "tag2"],
          "category_id": "25",
          "default_language": "en",
          "privacy_status": "public",
          "timestamps": [
            { "time": "0:00", "label": "Introduction" },
            { "time": "0:15", "label": "Key event" }
          ]
        },
        "tiktok": {
          "caption": "Short caption…",
          "hashtags": ["TagA", "TagB"],
          "privacy_level": "public",
          "duet_enabled": true,
          "stitch_enabled": true
        }
        /* Other platforms omitted */
      },
      "posting_status": {
        "youtube": {
          "posted": true,
          "url": "https://www.youtube.com/watch?v=...",
          "posted_at": "2025-12-16T23:10:22.475164-05:00"
        },
        "tiktok": {
          "posted": false,
          "url": "inbox_uploaded (publish_id: ...)",
          "posted_at": "2026-02-12T22:56:22.88-05:00"
        }
      },
      "analytics": {}
    }
  ]
}
```

For a complete schema and real examples, see **`METADATA_GUIDE.md`**.

### 3.2 TikTok OAuth credentials

- Stored locally at:

  ```text
  ~/.credentials/tiktok-oauth.json
  ```

- Created via Login Kit OAuth flow the first time TikTok upload is used.
- Not committed to git; user can delete it or revoke access via TikTok app settings.

---

## 4. Go ↔ Python Integration (HeyGen)

The Go `video` package delegates actual HeyGen API calls to `video/video_generation.py`.

### 4.1 Request from Go

Simplified:

```go
cmd := exec.Command("poetry", "run", "python", "video/video_generation.py")
cmd.Stdin = bytes.NewReader(requestJSON)   // includes text, output_path, background, etc.
output, err := cmd.Output()
// Parse JSON into VideoResponse
```

### 4.2 Python script

High-level responsibilities:

1. Read JSON from stdin.
2. Build a HeyGen API request:
   - Text to speak.
   - Avatar and voice configuration.
   - Background (optional).
   - Video dimensions, aspect ratio, duration.
3. Poll HeyGen until the job finishes.
4. Download the resulting MP4 to `output_path`.
5. Print a JSON response describing:
   - `status` (`"success"` or `"error"`).
   - Optional `message` for debugging.
   - `video_url` (HeyGen-hosted).
   - `duration`.

### 4.3 Why this split?

- Go is excellent at:
  - Orchestration.
  - File IO.
  - Typed manifest handling.
- Python is convenient for:
  - Using official API SDKs.
  - Rapid iteration of third-party integrations.

Using JSON stdin/stdout avoids the need for a separate HTTP server or RPC framework and keeps boundaries simple.

---

## 5. Error Handling

### 5.1 Generation phase

- **Missing/invalid environment**
  - `HEYGEN_API_KEY` missing → `log.Fatal`.
  - Other env load issues:
    - `.env` missing → log info, continue.
    - Read errors → log warning, continue.

- **LLM failure**
  - If `GenerateEnrichedNewsContent` returns an error:
    - `log.Fatalf("Error generating enriched content: %v", err)`.

- **Background generation**
  - If `GenerateNewsroomBackground` fails:
    - Log warning.
    - Set `backgroundResult = nil`.
    - Use default background configuration.

- **Video generation**
  - If HeyGen returns an error or a non‑success status:
    - Log details (including `videoResp.Message` when present).
    - `log.Fatalf("Cannot continue without video")`.

- **Manifest write**
  - `AddItem` and `UpdateItem`:
    - On failure, log a warning but continue (generation/upload may have succeeded).

### 5.2 Upload phase

- **YouTube**
  - OAuth/HTTP errors bubble up and are logged where upload is triggered.
  - When appropriate, `Status.YouTube.Error` can store a human-readable error.

- **TikTok**
  - `GetTikTokAccessToken` error → upload fails with a descriptive error.
  - `initializeTikTokUpload`:
    - Includes TikTok’s own JSON error payload in logs.
  - `uploadTikTokVideoFile`:
    - On non‑2xx, logs HTTP status and body content.
  - `tools/upload_tiktok`:
    - On errors, writes details to `Status.TikTok.Error` in the manifest.

---

## 6. Performance & Limits

- **LLM + metadata**
  - Single-call approach reduces API overhead vs per‑platform requests.
  - See **`TTS_VIDEO_GUIDE.md`** and **`METADATA_GUIDE.md`** for concrete token/word math.

- **HeyGen video rendering**
  - Portrait 9:16 is the default and usually:
    - ~10–20 minutes per video (depending on length).
  - CLI logs:
    - When rendering starts.
    - Expected duration hints.
    - Final `duration` from HeyGen.

- **TikTok rate limiting**
  - Content Posting API: 6 requests/minute per user.
  - Typical CLI use (1–3 videos/hour) is far below this limit.

- **YouTube quotas**
  - Standard YouTube Data API v3 limits apply.
  - Upload frequency from this tool is low.

---

## 7. Extensibility & Customization

Common ways to extend the workflow:

- **Add new platforms**
  - Add fields to `metadata/types.go` under `Platforms`.
  - Update `metadata/prompts.go` so the LLM produces that platform’s metadata.
  - Add an uploader in `media/` (similar to YouTube/TikTok).

- **Change avatars / brand identity**
  - Edit `video/constants.go`:
    - `DefaultAvatarID`, voice, colors, aspect ratio.
  - Optionally add per‑item overrides carried through the manifest.

- **Change summary tone or length**
  - Adjust the system prompt in `metadata/prompts.go`:
    - Formal vs conversational tone.
    - Precise word-count bands (e.g., 45–50s vs 60–75s).
    - Stronger compliance/neutrality language.

- **Scheduling / automation**
  - External schedulers (cron, GitHub Actions, etc.) can call:
    - `go run main.go [flags]`
    - `go run tools/upload_tiktok/main.go [...]`
  - No scheduler is built-in; this repo intentionally stays CLI-only.

---

## 8. Testing the Workflow

### 8.1 Automated script

```bash
./test_video.sh
```

Typical responsibilities:

- Verify Go and Python dependencies are installed.
- Run a minimal `go run main.go` flow (with or without external API calls, depending on configuration/mocking).
- Confirm a video file and manifest entry are created.

### 8.2 Go package tests

```bash
go test ./news/...
go test ./video/...
go test ./metadata/...
```

You can also introduce integration tags for end‑to‑end tests that hit real external services when needed.

### 8.3 Manual testing

1. Run:

   ```bash
   go run main.go
   ```

2. Confirm that:
   - A new `news_..._final.mp4` file appears.
   - `content_manifest.json` contains a new item with that ID.

3. Inspect the manifest:

   ```bash
   go run tools/inspect_manifest.go list
   go run tools/inspect_manifest.go show <content_id>
   ```

4. Test optional uploads:

   ```bash
   # YouTube only
   go run main.go -upload-youtube

   # TikTok inbox only (production)
   go run main.go -upload-tiktok

   # TikTok inbox (sandbox)
   go run main.go -upload-tiktok -sandbox

   # Batch TikTok from manifest
   go run tools/upload_tiktok/main.go --dry-run --all-unposted
   go run tools/upload_tiktok/main.go --all-unposted
   ```

---

**Last Updated:** March 2026  
**Version:** 2.0.0

