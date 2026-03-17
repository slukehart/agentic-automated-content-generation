---
type: architecture
component: video
status: proposed
created: 2026-03-17
updated: 2026-03-17
tags: [video, aws, gpu]
---

## Overview

Video generation is the active development area of this system. The previous implementation used HeyGen's hosted AI avatar API. That code has been removed and the system is migrating to a custom model hosted on AWS GPU infrastructure.

## Current State

HeyGen integration has been removed from the codebase. References to HeyGen in `main.go` and `video/` are stubs or legacy comments pending replacement. The `HEYGEN_API_KEY` environment variable check in `main.go` will become invalid once the migration completes.

The `video` package still exposes:
- `video.GenerateNewsVideoFromText(summary, outputPath)` — generates video from text using default background
- `video.GenerateNewsVideoWithBackgroundImage(summary, outputPath, backgroundImagePath)` — generates video using AI-generated newsroom background
- `video.DefaultAvatarID` — the avatar identifier passed into the content manifest

Both functions return `*video.VideoResponse{Status, Message, VideoURL, Duration}`.

## Target Architecture

| Attribute | Target Spec |
|-----------|-------------|
| Resolution | Portrait 720x1280 (vertical/Shorts format) |
| Duration | 50–70 seconds |
| Content | AI avatar narrating the news summary |
| Hosting | Custom model on AWS GPU instance |
| Input | Text summary (150–200 words from Grok) + newsroom background image |
| Output | MP4 file saved locally as `{contentID}_final.mp4` |

## Integration Point

`main.go` calls video generation after metadata generation and before manifest creation:

```go
videoResp, err = video.GenerateNewsVideoWithBackgroundImage(
    enrichedContent.Summary,
    finalPath,
    backgroundResult.ImagePath,
)
```

The `videoResp.Duration` is stored in `MediaInfo.DurationSeconds` in the manifest.

## Key Files

- `video/` — package directory (implementation pending AWS migration)
- `main.go` lines 89–106 — video generation call site and error handling

## Dependencies

- Depends on: enriched summary from `news.GenerateEnrichedNewsContent`, optional background image from `news.GenerateNewsroomBackground`
- Used by: `main.go`; output path stored in `metadata.MediaInfo.VideoPath`

## Configuration

Once AWS migration is complete, expected configuration:

| Env Var | Description |
|---------|-------------|
| AWS credentials / endpoint | Access to GPU inference endpoint |
| Avatar model config | TBD based on chosen model |

The `HEYGEN_API_KEY` check in `main.go` will be removed as part of the migration.
