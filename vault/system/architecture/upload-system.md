---
type: architecture
component: upload
status: current
created: 2026-03-17
updated: 2026-03-17
tags: [upload, youtube, tiktok]
---

## Overview

The upload system posts generated videos to social platforms. YouTube and TikTok are fully implemented. Instagram, Twitter, Facebook, and LinkedIn have platform metadata generated (and stored in the manifest) but no upload code exists yet.

## YouTube Upload

**File:** `media/youtubeController.go`
**Auth file:** `media/auth.go`

### Authentication

- Uses YouTube Data API v3 with OAuth 2.0 (`google.golang.org/api/youtube/v3`)
- OAuth client credentials read from `client_secret.json` in project root
- Scope: `youtube.YoutubeUploadScope`
- Token cached at `~/.credentials/youtube-oauth.json`
- First run prompts user to open a browser URL and paste the auth code; subsequent runs use the cached token automatically

### Upload Flow (`UploadVideoToYouTube`)

1. Calls `GetYouTubeClient(ctx)` to get an authenticated `*youtube.Service`
2. Opens the video file from `contentItem.Media.VideoPath`
3. Prepends `"Shorts"` tag if not already present
4. Creates a `youtube.Video` resource with snippet (title, description, tags, category, language) and status (privacy, not-made-for-kids)
5. Calls `service.Videos.Insert(["snippet", "status"], video).Media(file).Do()`
6. Returns `YouTubeUploadResult{VideoID, VideoURL, UploadedAt}`

Videos are uploaded as **YouTube Shorts** — the description is prefixed with `#Shorts\n\n` automatically. Category ID is hardcoded to `"25"` (News & Politics). Privacy status comes from `contentItem.Platforms.YouTube.PrivacyStatus` (set to `"public"` by `ConvertToContentItem`).

## TikTok Upload

**File:** `media/tiktokController.go`
**Auth file:** `media/tiktok_auth.go`

### Authentication

- Uses TikTok Login Kit with OAuth 2.0 + PKCE (S256 code challenge)
- Credentials read from env: `TIKTOK_CLIENT_KEY` / `TIKTOK_CLIENT_SECRET`
- Sandbox mode uses: `TIKTOK_CLIENT_KEY_SANDBOX` / `TIKTOK_CLIENT_SECRET_SANDBOX`
- Token cached at `~/.credentials/tiktok-oauth.json`
- Redirect URI: `https://slukehart.github.io/agentic-automated-content-generation/callback`
- Scopes requested: `user.info.basic,video.upload,video.publish`
- First run prompts user to open a browser URL and paste the auth code

### Upload Flow (`UploadVideoToTikTok`)

Uses TikTok Content Posting API v2:

1. Calls `GetTikTokAccessToken(ctx, sandbox)` to get access token
2. Checks video file size (max 301 MB)
3. Determines chunk strategy: single chunk if ≤ 64 MB; multiple 64 MB chunks if larger
4. POST to `https://open.tiktokapis.com/v2/post/publish/inbox/video/init/` with `TikTokInitUploadRequest` — gets back `publish_id` and `upload_url`
5. PUT the video file to `upload_url` with `Content-Range` header
6. Returns `TikTokUploadResult{PublishID, Status: "inbox_uploaded", UploadedAt}`

**Important:** TikTok API uploads to the creator's **inbox only**. The video is not posted to the feed automatically. The user must open the TikTok app, review the video, add caption/hashtags, and tap Post manually. The `publish_id` is stored in the manifest's `Status.TikTok.URL` field for tracking.

### Sandbox Mode

Pass `-sandbox` CLI flag to use the TikTok sandbox environment (separate client credentials). Sandbox uploads go to a test environment and do not affect the real account.

## Planned Platforms

Metadata is generated for all 6 platforms on every run but upload code exists for only 2:

| Platform | Metadata Generated | Upload Implemented |
|----------|-------------------|-------------------|
| YouTube | Yes | Yes |
| TikTok | Yes | Yes |
| Instagram | Yes | No |
| Twitter/X | Yes | No |
| Facebook | Yes | No |
| LinkedIn | Yes | No |

## Key Files

- `media/youtubeController.go` — `UploadVideoToYouTube`, description formatting helpers
- `media/auth.go` — `GetYouTubeClient`, OAuth token cache logic
- `media/tiktokController.go` — `UploadVideoToTikTok`, chunked upload logic
- `media/tiktok_auth.go` — `GetTikTokAccessToken`, PKCE flow, token cache

## Dependencies

- Depends on: `metadata.ContentItem` (video path, platform metadata, privacy settings)
- Depends on: `client_secret.json` (YouTube), `TIKTOK_CLIENT_KEY`/`SECRET` env vars (TikTok)
- Used by: `main.go` (conditionally, based on CLI flags)
- Updates: `ContentItem.Status.YouTube` and `ContentItem.Status.TikTok` with posted/error state, then calls `manifestManager.UpdateItem`

## Configuration

| Env Var | Description |
|---------|-------------|
| `TIKTOK_CLIENT_KEY` | TikTok production OAuth client key |
| `TIKTOK_CLIENT_SECRET` | TikTok production OAuth client secret |
| `TIKTOK_CLIENT_KEY_SANDBOX` | TikTok sandbox OAuth client key |
| `TIKTOK_CLIENT_SECRET_SANDBOX` | TikTok sandbox OAuth client secret |

| File | Description |
|------|-------------|
| `client_secret.json` | YouTube OAuth2 client credentials (from Google Cloud Console) |
| `~/.credentials/youtube-oauth.json` | Cached YouTube OAuth token |
| `~/.credentials/tiktok-oauth.json` | Cached TikTok OAuth token |
