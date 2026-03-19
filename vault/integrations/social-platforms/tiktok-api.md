---
type: integration
service: tiktok
category: social-platform
status: implemented
auth_method: oauth2
created: 2026-03-17
updated: 2026-03-17
tags: [tiktok, upload, oauth, inbox]
---

## Overview

Upload AI-generated news summary videos to a TikTok creator's inbox via the TikTok Content Posting API. The video lands in the creator's inbox — the creator manually reviews and posts it from the TikTok app. Direct publishing is not used.

## API Details

- **APIs used**: Content Posting API + Login Kit (OAuth 2.0)
- **Base URL**: `https://open.tiktokapis.com`
- **Init upload endpoint**: `POST /v2/post/publish/inbox/video/init/`
- **Video upload**: `PUT {upload_url}` (URL returned by init)

Source: `media/tiktokController.go`.

## Authentication

OAuth 2.0 with PKCE (required by TikTok):

1. Credentials read from environment: `TIKTOK_CLIENT_KEY` and `TIKTOK_CLIENT_SECRET` (or `_SANDBOX` variants for sandbox mode).
2. On first run, a browser OAuth flow is triggered using PKCE (SHA256 code challenge, S256 method).
3. Scopes requested: `user.info.basic,video.upload,video.publish` (comma-separated — TikTok does not accept space-separated).
4. Token saved to `~/.credentials/tiktok-oauth.json` for subsequent runs.
5. Redirect URI (must match TikTok Developer Portal exactly): `https://slukehart.github.io/agentic-automated-content-generation/callback`

Source: `media/tiktok_auth.go` — `GetTikTokAccessToken()`.

## Upload Flow

Two-step process:

### Step 1 — Initialize upload

POST `https://open.tiktokapis.com/v2/post/publish/inbox/video/init/` with:
```json
{
  "source_info": {
    "source": "FILE_UPLOAD",
    "video_size": <bytes>,
    "chunk_size": <bytes>,
    "total_chunk_count": <n>
  },
  "post_info": {
    "privacy_level": "...",
    "disable_comment": false,
    "disable_duet": false,
    "disable_stitch": false
  }
}
```

Response returns `publish_id` and `upload_url`.

### Step 2 — Upload video file

`PUT {upload_url}` with headers:
- `Content-Type: video/mp4`
- `Content-Range: bytes 0-{last}/{total}`
- `Content-Length: {total}`

HTTP 201 = complete, 206 = partial (more chunks needed).

## Chunking Logic

| Video size | Behavior |
|---|---|
| <= 64 MB | Single chunk — entire file in one PUT |
| > 64 MB | Multiple chunks of 64 MB each; last chunk may be up to 128 MB |

Minimum chunk size: 5 MB. Maximum: 64 MB (last chunk up to 128 MB per TikTok docs).

## Constraints

- **Max video size**: 301 MB (≈287.6 MB) — enforced before upload attempt.
- **Sandbox mode**: Pass `sandbox=true` to use sandbox credentials and test without real posting.
- **Inbox only**: API uploads to the creator's inbox. The creator must open the TikTok app, review the video, and tap "Post" to publish. Caption and hashtags are added manually in the app.

## Result

Returns `TikTokUploadResult` with:
- `PublishID` — TikTok publish ID
- `Status` — `"inbox_uploaded"`
- `UploadedAt` — timestamp

## Notes

- Caption and hashtags from `ContentItem.Platforms.TikTok` are printed to stdout as instructions for the creator; they are not set via API (inbox flow limitation).
- The `video.publish` scope is listed in the OAuth request but the actual upload uses the inbox endpoint, not the direct publish endpoint (`/v2/post/publish/video/init/`).

## Related

- [[upload-system]] — How the pipeline uses this API
