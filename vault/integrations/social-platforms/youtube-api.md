---
type: integration
service: youtube
category: social-platform
status: implemented
auth_method: oauth2
created: 2026-03-17
updated: 2026-03-17
tags: [youtube, upload, oauth, shorts]
---

## Overview

Upload AI-generated news summary videos as YouTube Shorts via the YouTube Data API v3.

## API Details

- **API**: YouTube Data API v3
- **Go client**: `google.golang.org/api/youtube/v3`
- **Scope**: `youtube.upload` (`https://www.googleapis.com/auth/youtube.upload`)

## Authentication

OAuth 2.0 via Google:

1. Place `client_secret.json` in the project root (downloaded from Google Cloud Console).
2. On first run, a browser OAuth flow is triggered — user authorizes and pastes the code.
3. Token is saved to `~/.credentials/youtube-oauth.json` (mode 0600) for all subsequent runs.
4. Token refresh is handled automatically by the `golang.org/x/oauth2` library.

Source: `media/auth.go` — `GetYouTubeClient()`.

## Upload Flow

1. Read `client_secret.json` → build `oauth2.Config` with `youtube.YoutubeUploadScope`.
2. Load or acquire token → create authenticated `*youtube.Service`.
3. Construct `youtube.Video` with `Snippet` and `Status` parts.
4. Call `service.Videos.Insert([]string{"snippet", "status"}, video).Media(file).Do()`.

Source: `media/youtubeController.go` — `UploadVideoToYouTube()`.

## Metadata Constraints

| Field | Limit / Default |
|---|---|
| Title | Max 100 characters (truncated with `...` if over) |
| Description | Max 5000 characters (truncated with notice if over) |
| Category ID | `"25"` — News & Politics |
| Privacy status | `"public"` by default |
| Language | `"en"` |
| Made for kids | `false` |

## Shorts Tagging

To signal YouTube Shorts eligibility:

- Description is prepended with `#Shorts\n\n`.
- Tags list is checked for `"shorts"` or `"short"` (case-insensitive); if absent, `"Shorts"` is prepended to the tag list.

## Description Format

The formatted description includes:
- `#Shorts` header
- Main description text
- Optional timestamps (`⏱️ Timestamps:` section)
- Source attribution (`📰 Source:` and `🔗 Full Article:`)
- AI disclosure footer (Grok, HeyGen, NewsAPI attribution)

## Result

Returns `YouTubeUploadResult` with:
- `VideoID` — YouTube video ID
- `VideoURL` — `https://www.youtube.com/watch?v={id}`
- `UploadedAt` — timestamp

Shorts URL: `https://youtube.com/shorts/{id}`

## Notes

- Video processing takes 1–5 minutes after upload before it is fully available.
- Shorts appear in the Shorts feed within a few hours of processing.

## Related

- [[upload-system]] — How the pipeline uses this API
