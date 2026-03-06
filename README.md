# Content Generation Automation

> Automated, local-first news-to-short video pipeline with AI avatars, multi-platform metadata, and safe TikTok/YouTube upload helpers.

This repository powers a **CLI-only, local workflow** that:

- Fetches breaking news from trusted publishers
- Generates broadcast-style summaries and **platform-optimized metadata** in a single LLM call
- Creates AI newsroom backgrounds
- Renders short, portrait AI-avatar news clips with **HeyGen TTS + avatars**
- Optionally uploads those clips to **YouTube** and **TikTok (inbox only)** to help creators post faster

---

## For TikTok / Platform Reviewers

This section is written specifically for reviewers evaluating API usage, data handling, and user impact.

### What the app does

- **Use case**: Help a creator turn **public news articles** into short, factual news explainer videos.
- **Form factor**: Command-line tool; there is **no hosted backend** or mobile/desktop app.
- **Output**: 50–70s vertical MP4 clips with an AI news anchor, plus per-platform metadata stored locally.
- **User workflow**:
  1. Run `go run main.go [-upload-youtube] [-upload-tiktok]`
  2. The tool generates one short news video and saves it locally
  3. (Optional) It uploads that video to YouTube and/or to the user’s **TikTok inbox** for final review

### TikTok products & APIs used

- **Product**: TikTok **Login Kit** (OAuth 2.0)
  - Used once to obtain user consent and OAuth tokens
  - Scopes:
    - `user.info.basic` – show the logged-in TikTok account in the CLI for confirmation
    - `video.upload` – upload the MP4 file to TikTok
    - `video.publish` – publish the uploaded file **to the user’s inbox only**
- **Product**: TikTok **Content Posting API – Upload to TikTok**
  - Docs: `https://developers.tiktok.com/products/content-posting-api/`
  - Our usage:
    - Initialize upload (`/v2/post/publish/inbox/video/init/`)
    - Upload MP4 via the returned `upload_url`
    - Mark as **inbox upload** so the creator reviews and posts (or deletes) in the TikTok app
  - We **do not** post directly to the public feed; the creator always makes the final decision in-app.

For a detailed diagram, see **[TIKTOK_WORKFLOW_DIAGRAM.md](TIKTOK_WORKFLOW_DIAGRAM.md)**.

### Data sources and AI usage

- **News content**:
  - Fetched from **NewsAPI** and original publisher URLs (e.g., BBC, Politico, Washington Post, etc.).
  - Only **public, published news articles** are used as input.
- **LLM usage (summaries & metadata)**:
  - Takes only the article headline, URL, and text content.
  - Generates:
    - A 150–200 word neutral news summary
    - SEO metadata
    - Platform-specific metadata (YouTube, TikTok, Instagram, Twitter/X, Facebook, LinkedIn) **in a single call**.
- **Image & video generation**:
  - **Grok / image model**: creates a static newsroom background from a generic prompt (no user data).
  - **HeyGen**: takes the generated summary text and optional background and returns:
    - A narrated video (AI voice + avatar)
    - A hosted URL for that render

**No TikTok user data, viewing history, or analytics are ever sent to any AI or third‑party model.** Only public news text is used as input to AI systems.

### Storage and data handling

All state is **local to the user’s machine**:

- **Video files**: `news_YYYYMMDD_HHMMSS_final.mp4` stored alongside this repo.
- **Metadata manifest**: `content_manifest.json` in the project root:
  - Stores source article URL/title
  - Stores generated summary + SEO + per-platform metadata
  - Tracks whether each item has been uploaded to YouTube/TikTok (and, for TikTok, the `publish_id`)
- **TikTok OAuth tokens**:
  - Stored in `~/.credentials/tiktok-oauth.json` on the user’s machine
  - Not checked into git; never uploaded anywhere
  - User can revoke by:
    - Deleting the file locally, and/or
    - Revoking access in TikTok app settings


See **[LEGAL_README.md](LEGAL_README.md)** for how this maps to the published Privacy Policy and Terms of Service.

### User control, safety, and policy alignment

- **Inbox-only TikTok uploads**:
  - Uploads go to the creator’s **TikTok inbox**, never directly to the public feed.
  - Creators can:
    - Preview the video
    - Edit the caption/hashtags/music
    - Adjust privacy / duet / stitch / comments
    - Either **Post** or **Delete**
- **No hidden automation**:
  - The CLI prints each step and the TikTok `publish_id`.
  - Users must run commands manually; there is no always-on scheduler or background posting.
- **Factual, cited news**:
  - Summaries are neutral and fact-focused.
  - Each generated description or caption includes a **clear source attribution** (e.g., “Source: BBC News”).
  - The YouTube description generator also includes an **AI disclosure** block.
- **Rate limits & anti-spam**:
  - Typical usage: 1–3 videos/hour, well under TikTok’s standard rate limits.
  - Upload tools support **dry-run** mode for safe testing.

### Legal documents

- **Privacy Policy**: see **[PRIVACY_POLICY.md](PRIVACY_POLICY.md)** and the GitHub Pages version used in production.
- **Terms of Service**: see **[TERMS_OF_SERVICE.md](TERMS_OF_SERVICE.md)**.
- **TikTok integration rationale, scopes, and diagrams**: **[TIKTOK_WORKFLOW_DIAGRAM.md](TIKTOK_WORKFLOW_DIAGRAM.md)** and **[TIKTOK_USAGE_GUIDE.md](TIKTOK_USAGE_GUIDE.md)**.

---

## High-Level Architecture

### End-to-end workflow

At a high level, a single run of `main.go` performs:

```text
NewsAPI → AI Summarization & Metadata → Grok Image Gen → HeyGen (Text → Video)
                                                         ↓
                                              AI Avatar News Short (MP4)
                                               + Multi-platform metadata
                                                   ↓
                           Optional upload to YouTube + TikTok (inbox only)
```

In code, this corresponds to:

- `news/parseNewsArticles.go` – select one news article from NewsAPI
- `news/metadata_generation.go` – call the LLM once to get:
  - A 150–200 word news summary
  - SEO metadata
  - Per-platform metadata (YouTube, TikTok, Instagram, Twitter, Facebook, LinkedIn)
- `news/metadata_generation.go` – generate an **AI newsroom background** via Grok
- `video/video.go` + `video/video_generation.py` – call HeyGen to:
  - Render an AI avatar reading the summary
  - Use either the AI-generated background or a default newsroom
  - Return a finished MP4 plus a hosted URL
- `metadata/manifest.go` – persist everything into `content_manifest.json`
- `media/youtubeController.go` – (optional) upload the MP4 to YouTube Shorts
- `media/tiktokController.go` – (optional) upload the MP4 to TikTok inbox

### Code structure (current)

```text
content-generation-automation/
├── main.go                     # Orchestrates a single news → video → manifest run
├── news/
│   ├── parseNewsArticles.go    # NewsAPI fetching and article selection
│   └── metadata_generation.go  # LLM summaries, metadata, and background prompt/image
├── video/
│   ├── video.go                # Go interface for kicking off video generation
│   ├── constants.go            # Default avatar, aspect ratio, durations, etc.
│   └── video_generation.py     # HeyGen API integration and polling
├── metadata/
│   ├── types.go                # Strongly-typed manifest + platform metadata structs
│   ├── manifest.go             # Read/write/query `content_manifest.json`
│   └── prompts.go              # LLM prompt for summary + metadata
├── media/
│   ├── youtubeController.go    # YouTube Shorts upload helper
│   ├── tiktokController.go     # TikTok Content Posting API integration
│   └── auth.go, tiktok_auth.go # OAuth helpers
├── tools/
│   ├── inspect_manifest.go     # CLI for exploring the manifest
│   └── upload_tiktok/main.go   # Batch / dry-run TikTok uploader
├── AUDIO_VIDEO_SETUP.md        # Historical setup; audio is now optional
├── METADATA_GUIDE.md           # Deep-dive into manifest + metadata fields
├── TIKTOK_USAGE_GUIDE.md       # How to run TikTok upload flows
└── TIKTOK_WORKFLOW_DIAGRAM.md  # TikTok OAuth + upload data-flow diagram
```

For the older, more detailed pipeline doc (pre-HeyGen-TTS), see **[WORKFLOW.md](WORKFLOW.md)**.

---

## Content Manifest & Multi-Platform Metadata

Every generated piece of content is stored in **`content_manifest.json`** as a single, structured record. This file:

- Links the **source article** to:
  - A normalized summary
  - SEO metadata
  - Per-platform metadata (YouTube/TikTok/Instagram/Twitter/Facebook/LinkedIn)
  - Local media paths (MP4, background, avatar)
  - Posting status for each platform
- Enables:
  - Quick inspection via `tools/inspect_manifest.go`
  - Batch uploads via TikTok/YouTube tools
  - Future analytics or dashboards without scraping platforms

See **[METADATA_GUIDE.md](METADATA_GUIDE.md)** for examples and field-by-field documentation.

---

## TikTok Upload Flow (Inbox Only)

TikTok upload can be triggered in two ways:

1. **Inline with generation** (single item)
   - `go run main.go -upload-tiktok` (production)
   - `go run main.go -upload-tiktok -sandbox` (sandbox for app review/demo)
2. **From existing manifest entries** (batch or single)
   - `go run tools/upload_tiktok/main.go <content_id>`
   - `go run tools/upload_tiktok/main.go --sandbox <content_id>`
   - `go run tools/upload_tiktok/main.go --all-unposted`
   - `go run tools/upload_tiktok/main.go --dry-run --all-unposted`

Key behavior:

- First-time use triggers a **Login Kit OAuth browser flow**:
  - User logs into TikTok and grants scopes
  - A short-lived authorization code is pasted back into the CLI
  - Tokens are stored **locally** at `~/.credentials/tiktok-oauth.json`
- Upload API specifics:
  - Initializes an upload session with **file size and chunk info**
  - Uploads the MP4 file (single-chunk for typical video sizes)
  - Publishes the video to the **creator’s inbox**, not directly to the feed
- Manifest tracking:
  - Each upload stores:
    - `publish_id`
    - Upload timestamp
    - An `inbox_uploaded (...)` status string

See **[TIKTOK_USAGE_GUIDE.md](TIKTOK_USAGE_GUIDE.md)** and **[TIKTOK_WORKFLOW_DIAGRAM.md](TIKTOK_WORKFLOW_DIAGRAM.md)** for an API-by-API walkthrough with diagrams.

---

## YouTube Shorts Upload Flow

YouTube upload is optional and uses the official **YouTube Data API v3**:

- Triggered from `main.go` with:

  ```bash
  go run main.go -upload-youtube
  go run main.go -upload-youtube -upload-tiktok   # both platforms
  ```

- The uploader:
  - Uploads the MP4 with a **Shorts-oriented title, tags, and description**
  - Prepends `#Shorts` and includes:
    - Timestamps
    - Source attribution
    - AI disclosure (Grok + HeyGen + NewsAPI)
  - Updates `content_manifest.json` with the resulting YouTube URL and posting timestamp

YouTube upload uses OAuth and local token caching similar to TikTok, but is separate and independent.

---

## Quick Start (Developers / Maintainers)

While the primary audience for this README section is reviewers, these steps show how the system is run in practice on a local machine.

### 1. Install dependencies

```bash
go mod download
poetry install     # For Python-side HeyGen integration
```

### 2. Configure environment

Create a `.env` file in the repo root:

```env
HEYGEN_API_KEY=your-heygen-api-key      # https://app.heygen.com/settings/api
NEWS_API_KEY=your-newsapi-key           # https://newsapi.org
X_AI_KEY=your-grok-api-key              # Grok / xAI for summaries + background prompts
TIKTOK_CLIENT_KEY=your_production_client_key
TIKTOK_CLIENT_SECRET=your_production_client_secret
TIKTOK_CLIENT_KEY_SANDBOX=your_sandbox_client_key       # optional
TIKTOK_CLIENT_SECRET_SANDBOX=your_sandbox_client_secret # optional
```

For YouTube upload, configure Google OAuth credentials separately (see `AUDIO_VIDEO_SETUP.md` and standard YouTube API docs).

### 3. Run a full generation (no uploads)

```bash
go run main.go
```

This will:

- Fetch one news article
- Generate the summary + metadata
- Render an AI avatar video via HeyGen
- Save `news_YYYYMMDD_HHMMSS_final.mp4`
- Append a new item to `content_manifest.json`

### 4. Run with YouTube + TikTok upload

```bash
# YouTube only
go run main.go -upload-youtube

# TikTok inbox only (production)
go run main.go -upload-tiktok

# TikTok inbox (sandbox) for app review video
go run main.go -upload-tiktok -sandbox

# Both platforms
go run main.go -upload-youtube -upload-tiktok
```

On first TikTok upload, you’ll walk through the browser-based OAuth flow described in **[TIKTOK_USAGE_GUIDE.md](TIKTOK_USAGE_GUIDE.md)**.

---

## Additional Documentation

- **[QUICKSTART.md](QUICKSTART.md)** – more detailed setup and end-to-end usage.
- **[AUDIO_VIDEO_SETUP.md](AUDIO_VIDEO_SETUP.md)** – historical separation of TTS + video; audio is now optional because HeyGen handles TTS internally.
- **[TTS_VIDEO_GUIDE.md](TTS_VIDEO_GUIDE.md)** – how summary length maps to 1-minute TTS videos.
- **[METADATA_GUIDE.md](METADATA_GUIDE.md)** – manifest and metadata schema in depth.
- **[TOKEN_OPTIMIZATION_GUIDE.md](TOKEN_OPTIMIZATION_GUIDE.md)** – API token and cost optimization strategies.
- **[GITHUB_PAGES_SETUP.md](GITHUB_PAGES_SETUP.md)** – how the legal/privacy pages are published for review.

---

## Contributing & Contact

This project is primarily a reference implementation and integration demo for AI-assisted news content automation.

- Issues and PRs are welcome for:
  - Bug fixes
  - Documentation clarifications
  - Additional platform integrations
- For questions specifically about TikTok integration, privacy, or review,
  - See **[LEGAL_README.md](LEGAL_README.md)** and
  - Reach out via the contact information listed there.

---

**Built to demonstrate safe, transparent AI-assisted news workflows across social platforms.**


