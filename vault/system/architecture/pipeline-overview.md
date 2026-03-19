---
type: architecture
component: pipeline
status: current
created: 2026-03-17
updated: 2026-03-17
tags: [pipeline, orchestration]
---

## Overview

`main.go` is the single entry point that orchestrates the full end-to-end content generation pipeline: from fetching a live news article through generating an AI avatar video and optionally uploading to YouTube or TikTok.

## How It Works

The pipeline runs sequentially in one process:

1. **Parse CLI flags** — `flag.Parse()` reads `-upload-youtube`, `-upload-tiktok`, and `-sandbox`
2. **Load environment** — `godotenv.Load()` reads `.env` for local dev; production uses system env or cloud secrets
3. **Initialize manifest** — `metadata.NewManifestManager("content_manifest.json")`
4. **Fetch news** — `news.ParseNewsArticles()` → returns `AiArticleParameters{ArticleUrl, ArticleTitle}`
5. **Generate metadata** — `news.GenerateEnrichedNewsContent(article)` → single Grok API call produces summary + all platform metadata
6. **Generate background** — `news.GenerateNewsroomBackground("")` → X.AI image API generates newsroom JPEG; falls back gracefully to default if this fails
7. **Generate video** — `video.GenerateNewsVideoWithBackgroundImage(...)` or `video.GenerateNewsVideoFromText(...)` (HeyGen, currently being replaced)
8. **Build ContentItem** — `news.ConvertToContentItem(...)` assembles the full manifest record
9. **Save to manifest** — `manifestManager.AddItem(contentItem)`
10. **Optional upload to YouTube** — `media.UploadVideoToYouTube(&contentItem)` if `-upload-youtube` flag is set
11. **Optional upload to TikTok** — `media.UploadVideoToTikTok(&contentItem, sandbox)` if `-upload-tiktok` flag is set
12. **Update manifest with upload status** — `manifestManager.UpdateItem(contentItem)` after each upload attempt

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-upload-youtube` | false | Upload generated video to YouTube after creation |
| `-upload-tiktok` | false | Upload generated video to TikTok after creation |
| `-sandbox` | false | Use TikTok sandbox environment for testing |

## Key Files

- `main.go` — pipeline orchestration, all stages wired together here
- `news/parseNewsArticles.go` — stage 1: NewsAPI fetch
- `news/metadata_generation.go` — stages 2 and 3: Grok LLM + image generation
- `media/youtubeController.go` — stage 5: YouTube upload
- `media/tiktokController.go` — stage 5: TikTok upload

## Dependencies

```
main.go
  ├── news package (ParseNewsArticles, GenerateEnrichedNewsContent, GenerateNewsroomBackground, ConvertToContentItem)
  ├── video package (GenerateNewsVideoFromText, GenerateNewsVideoWithBackgroundImage)
  ├── metadata package (NewManifestManager, ContentItem)
  └── media package (UploadVideoToYouTube, UploadVideoToTikTok)
```

## Configuration

| Env Var | Required | Description |
|---------|----------|-------------|
| `NEWS_API_KEY` | Yes | NewsAPI authentication |
| `X_AI_KEY` | Yes | Grok LLM + image generation |
| `HEYGEN_API_KEY` | Yes (currently) | HeyGen video generation (being replaced) |
| `TIKTOK_CLIENT_KEY` | If using TikTok | TikTok OAuth client key |
| `TIKTOK_CLIENT_SECRET` | If using TikTok | TikTok OAuth client secret |
| `TIKTOK_CLIENT_KEY_SANDBOX` | If using sandbox | TikTok sandbox credentials |
| `TIKTOK_CLIENT_SECRET_SANDBOX` | If using sandbox | TikTok sandbox credentials |

Content ID format: `news_YYYYMMDD_HHMMSS` (e.g., `news_20260317_143022`)
Output video path: `{contentID}_final.mp4`

## Related

- [[news-fetching]] — Stage 1: article retrieval
- [[metadata-generation]] — Stage 2: summary + platform metadata
- [[video-generation]] — Stage 3: AI avatar video
- [[manifest-system]] — Stage 4: content manifest storage
- [[upload-system]] — Stage 5: platform uploads
