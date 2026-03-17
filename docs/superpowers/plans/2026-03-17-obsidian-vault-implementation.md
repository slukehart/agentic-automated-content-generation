# Obsidian Vault Knowledge Base — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an Obsidian vault knowledge base to the content-generation-automation project with MCP server integration for token-efficient agent retrieval.

**Architecture:** Obsidian vault at `vault/` with layered folders (vision, system, integrations, prompts, templates). Two MCP servers — smart-connections for semantic search and qmd for structured/hybrid search. Agent draft workflow via `vault/.drafts/`.

**Tech Stack:** Obsidian, @gogogadgetbytes/smart-connections-mcp (Node.js), @tobilu/qmd (Node.js + node-llama-cpp), Markdown with YAML frontmatter.

**Spec:** `docs/superpowers/specs/2026-03-16-obsidian-vault-design.md`

---

## Chunk 1: Project Cleanup

### Task 1: Remove Legacy HeyGen/TTS Files

**Files:**
- Remove: `video/` (entire directory — HeyGen integration; skip if already removed)
- Remove: `WEBAPP_VERIFICATION/` (empty directory)
- Remove: `poetry.lock`
- Remove: `AUDIO_SETUP.md`
- Remove: `AUDIO_VIDEO_SETUP.md`
- Remove: `VIDEO_SETUP.md`
- Remove: `TTS_VIDEO_GUIDE.md`
- Remove: `QUICKSTART.md`

- [ ] **Step 1: Remove legacy files and directories**

```bash
cd /Users/slukehart/Documents/Github/content-generation-automation
rm -rf video/
rm -rf WEBAPP_VERIFICATION/
rm -f poetry.lock
rm -f AUDIO_SETUP.md AUDIO_VIDEO_SETUP.md VIDEO_SETUP.md TTS_VIDEO_GUIDE.md QUICKSTART.md
```

- [ ] **Step 2: Verify removal**

```bash
ls video/ WEBAPP_VERIFICATION/ poetry.lock AUDIO_SETUP.md AUDIO_VIDEO_SETUP.md VIDEO_SETUP.md TTS_VIDEO_GUIDE.md QUICKSTART.md 2>&1
```

Expected: All files report "No such file or directory"

- [ ] **Step 3: Commit cleanup**

```bash
git add video/ WEBAPP_VERIFICATION/ poetry.lock AUDIO_SETUP.md AUDIO_VIDEO_SETUP.md VIDEO_SETUP.md TTS_VIDEO_GUIDE.md QUICKSTART.md --ignore-unmatch
git commit -m "chore: remove legacy HeyGen/TTS files and empty directories

Remove video/ (HeyGen integration), legacy TTS docs, empty
WEBAPP_VERIFICATION/, and poetry.lock. Project is moving to
custom video model on AWS GPU."
```

### Task 2: Reorganize Documentation

**Files:**
- Create: `docs/guides/` directory
- Create: `docs/setup/` directory
- Create: `docs/legal/` directory
- Move: `TIKTOK_WORKFLOW_DIAGRAM.md` → `docs/guides/`
- Move: `TIKTOK_USAGE_GUIDE.md` → `docs/guides/`
- Move: `METADATA_GUIDE.md` → `docs/guides/`
- Move: `WORKFLOW.md` → `docs/guides/`
- Move: `GITHUB_PAGES_SETUP.md` → `docs/setup/`
- Move: `LEGAL_README.md` → `docs/legal/`
- Move: `PRIVACY_POLICY.md` + `PRIVACY_POLICY/` → `docs/legal/`
- Move: `TERMS_OF_SERVICE.md` + `TERMS_OF_SERVICE/` → `docs/legal/`
- Move: `privacy_policy.html` → `docs/legal/`
- Move: `terms_of_service.html` → `docs/legal/`

- [ ] **Step 1: Create documentation directories**

```bash
mkdir -p docs/guides docs/setup docs/legal
```

- [ ] **Step 2: Move guide docs**

```bash
mv TIKTOK_WORKFLOW_DIAGRAM.md docs/guides/
mv TIKTOK_USAGE_GUIDE.md docs/guides/
mv METADATA_GUIDE.md docs/guides/
mv WORKFLOW.md docs/guides/
```

- [ ] **Step 3: Move setup docs**

```bash
mv GITHUB_PAGES_SETUP.md docs/setup/
```

- [ ] **Step 4: Move legal docs**

```bash
mv LEGAL_README.md docs/legal/
mv PRIVACY_POLICY.md docs/legal/
mv PRIVACY_POLICY/ docs/legal/
mv TERMS_OF_SERVICE.md docs/legal/
mv TERMS_OF_SERVICE/ docs/legal/
mv privacy_policy.html docs/legal/
mv terms_of_service.html docs/legal/
```

- [ ] **Step 5: Verify moves**

```bash
ls docs/guides/ docs/setup/ docs/legal/
```

Expected: All files present in new locations.

- [ ] **Step 6: Commit reorganization**

```bash
git add -A
git commit -m "chore: consolidate documentation into docs/ subdirectories

Move guides, setup docs, and legal files from project root into
docs/guides/, docs/setup/, and docs/legal/ respectively."
```

### Task 3: Move Video Files to output/videos/

**Files:**
- Create: `output/videos/` directory
- Move: All `news_*_final.mp4` files from project root

- [ ] **Step 1: Create output directory**

```bash
mkdir -p output/videos
```

- [ ] **Step 2: Move MP4 files**

```bash
mv news_*_final.mp4 output/videos/
```

- [ ] **Step 3: Verify**

```bash
ls output/videos/
ls news_*_final.mp4 2>&1  # Should show "No such file"
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: move generated videos to output/videos/

Clean up project root by moving MP4 files to dedicated output directory."
```

> **Note:** MP4 files are already in `.gitignore` (`*.mp4`), so this is a filesystem-only change. The commit will just track the directory creation.

### Task 4: Update .gitignore

**Files:**
- Modify: `.gitignore`

- [ ] **Step 1: Update .gitignore**

The current `.gitignore` has `docs/**` on line 45 which ignores all documentation including specs and plans. Remove that line and add vault + superpowers entries.

Replace line 45 (`docs/**`) with:

```
# Obsidian vault internals
vault/.obsidian/
vault/.drafts/
vault/.smart-connections/

# Superpowers brainstorm sessions
.superpowers/

# Output
output/
```

The final `.gitignore` should look like this at the end (after the Go coverage section):

```
# Dependency directories (remove the comment below to include it)
# vendor/

# End of https://mrkandreev.name/snippets/gitignore-generator/#Go

# Obsidian vault internals
vault/.obsidian/
vault/.drafts/
vault/.smart-connections/

# Superpowers brainstorm sessions
.superpowers/

# Output
output/
```

> **Note:** `poetry.lock` remains in `.gitignore` under the Python section as protection even though the file was deleted — it prevents accidental future commits if Poetry is used again.

- [ ] **Step 2: Verify docs are now tracked**

```bash
git status
```

Expected: Files in `docs/` should now appear as trackable.

- [ ] **Step 3: Commit**

```bash
git add .gitignore
git add docs/
git commit -m "chore: update .gitignore for vault and track docs/

Remove blanket docs/** exclusion so specs, plans, guides, and legal
docs are version controlled. Add vault/.obsidian/, vault/.drafts/,
vault/.smart-connections/, .superpowers/, and output/ to gitignore."
```

---

## Chunk 2: Vault Structure & Templates

### Task 5: Create Vault Directory Structure

**Files:**
- Create: `vault/vision/feature-ideas/.gitkeep`
- Create: `vault/system/architecture/.gitkeep`
- Create: `vault/system/decisions/.gitkeep`
- Create: `vault/system/lessons/.gitkeep`
- Create: `vault/integrations/social-platforms/.gitkeep`
- Create: `vault/integrations/infrastructure/.gitkeep`
- Create: `vault/prompts/.gitkeep`
- Create: `vault/templates/`
- Create: `vault/.drafts/.gitkeep`

- [ ] **Step 1: Create all vault directories**

```bash
mkdir -p vault/vision/feature-ideas
mkdir -p vault/system/architecture
mkdir -p vault/system/decisions
mkdir -p vault/system/lessons
mkdir -p vault/integrations/social-platforms
mkdir -p vault/integrations/infrastructure
mkdir -p vault/prompts
mkdir -p vault/templates
mkdir -p vault/.drafts
```

- [ ] **Step 2: Add .gitkeep files for empty directories**

```bash
touch vault/vision/feature-ideas/.gitkeep
touch vault/system/lessons/.gitkeep
touch vault/.drafts/.gitkeep
```

- [ ] **Step 3: Verify structure**

```bash
find vault -type d | sort
```

Expected output:
```
vault
vault/.drafts
vault/integrations
vault/integrations/infrastructure
vault/integrations/social-platforms
vault/prompts
vault/system
vault/system/architecture
vault/system/decisions
vault/system/lessons
vault/templates
vault/vision
vault/vision/feature-ideas
```

- [ ] **Step 4: Commit vault skeleton**

```bash
git add vault/
git commit -m "feat: create Obsidian vault directory structure

Layered vault with vision/ (human), system/ (agent-drafted),
integrations/ (shared), prompts/ (shared), and templates/.
Includes .drafts/ staging area for agent write-back workflow."
```

### Task 6: Create Note Templates

**Files:**
- Create: `vault/templates/architecture-note.md`
- Create: `vault/templates/decision-note.md`
- Create: `vault/templates/lesson-note.md`
- Create: `vault/templates/integration-note.md`

- [ ] **Step 1: Create architecture note template**

Write to `vault/templates/architecture-note.md`:

```markdown
---
type: architecture
component: # pipeline | news-fetching | metadata | video | upload | manifest
status: current # current | outdated | proposed
created: {{date}}
updated: {{date}}
tags: []
---

## Overview

What this component does and its role in the pipeline.

## How It Works

Technical details of the implementation.

## Key Files

- `path/to/file.go` — description

## Dependencies

What this component depends on and what depends on it.

## Configuration

Environment variables, settings, or constants.
```

- [ ] **Step 2: Create decision note template**

Write to `vault/templates/decision-note.md`:

```markdown
---
type: decision
component: # pipeline | news-fetching | metadata | video | upload | manifest
status: accepted # accepted | superseded | proposed
created: {{date}}
context: ""
tags: []
---

## Decision

What was decided.

## Context

Why this decision was needed.

## Options Considered

1. **Option A** — description
2. **Option B** — description

## Rationale

Why this option was chosen over alternatives.

## Consequences

What this decision means going forward.
```

- [ ] **Step 3: Create lesson note template**

Write to `vault/templates/lesson-note.md`:

```markdown
---
type: lesson
relates_to: # tiktok | youtube | instagram | twitter | facebook | linkedin | aws-gpu | newsapi | grok-xai | pipeline | metadata | video | upload | manifest
severity: medium # high | medium | low
created: {{date}}
tags: []
---

## What Happened

Brief description of the issue or surprise.

## Root Cause

Why it happened.

## Fix / Workaround

What was done to resolve it.

## Prevention

How to avoid this in the future.
```

- [ ] **Step 4: Create integration note template**

Write to `vault/templates/integration-note.md`:

```markdown
---
type: integration
service: # youtube | tiktok | instagram | twitter | facebook | linkedin | aws-gpu | newsapi | grok-xai
category: # social-platform | infrastructure
status: planned # implemented | in-progress | planned
auth_method: # oauth2 | api-key | iam
created: {{date}}
updated: {{date}}
tags: []
---

## Overview

What this service does in our pipeline.

## Authentication

OAuth flow, token management, scopes needed.

## API Endpoints

Key endpoints, request/response formats.

## Content Requirements

Video specs, caption limits, hashtag rules, aspect ratios.

## Rate Limits & Quotas

API limits, daily quotas, retry strategies.

## Gotchas

Platform-specific quirks agents should know.

## Implementation Status

What's built, what's pending, what's blocked.
```

- [ ] **Step 5: Commit templates**

```bash
git add vault/templates/
git commit -m "feat: add vault note templates with frontmatter schemas

Four templates: architecture, decision, lesson, and integration.
Each defines the frontmatter fields for qmd structured queries."
```

### Task 7: Create Vault README

**Files:**
- Create: `vault/README.md`

- [ ] **Step 1: Write vault README**

Write to `vault/README.md`:

```markdown
# Knowledge Vault

Agent-readable knowledge base for the content-generation-automation project.

## Structure

| Folder | Owner | Purpose |
|--------|-------|---------|
| `vision/` | Human | Goals, roadmap, feature ideas |
| `system/architecture/` | Agent (draft) | How components work |
| `system/decisions/` | Agent (draft) | Why things were built this way |
| `system/lessons/` | Agent (draft) | Gotchas, failures, tips |
| `integrations/` | Both | External service API knowledge |
| `prompts/` | Both | Prompt engineering notes |
| `templates/` | Shared | Note templates with frontmatter schemas |
| `.drafts/` | Agent | Staging area — human approves before promotion |

## For Agents

### Reading
- Use MCP servers for token-efficient retrieval:
  - **smart-connections**: Semantic search — "what do we know about upload failures?"
  - **qmd**: Structured/hybrid search — `qmd search "upload" -c vault`

### Writing
- **Never** write directly to vault folders.
- Draft to `vault/.drafts/` using templates from `vault/templates/`.
- Include `draft_reason` in frontmatter explaining why this note is worth keeping.
- Notify the user so they can review and promote to the correct folder.
- If `.drafts/` has pending notes, use date-prefixed filenames and warn user.

### When to Draft
- New component built or significantly changed → architecture note
- Non-obvious decision made → decision note
- Something broke or was surprising → lesson note
- New external service integrated → integration note
- **Do NOT draft** for: small bug fixes, config tweaks, formatting changes.
```

- [ ] **Step 2: Commit README**

```bash
git add vault/README.md
git commit -m "feat: add vault README with agent usage instructions"
```

---

## Chunk 3: Seed Content — System & Integration Notes

### Task 8: Write Architecture Notes from Existing Code

These notes document the current system based on what's already implemented. They are the initial knowledge base that future agents will query.

**Files:**
- Create: `vault/system/architecture/pipeline-overview.md`
- Create: `vault/system/architecture/news-fetching.md`
- Create: `vault/system/architecture/metadata-generation.md`
- Create: `vault/system/architecture/video-generation.md`
- Create: `vault/system/architecture/upload-system.md`
- Create: `vault/system/architecture/manifest-system.md`

**Reference files to read first:**
- `main.go` — pipeline orchestration
- `news/parseNewsArticles.go` — NewsAPI fetching
- `news/metadata_generation.go` — Grok LLM calls + image generation
- `metadata/types.go` — manifest data model
- `metadata/manifest.go` — manifest CRUD
- `metadata/prompts.go` — LLM prompt templates
- `media/youtubeController.go` — YouTube upload
- `media/tiktokController.go` — TikTok upload
- `media/auth.go` — YouTube OAuth
- `media/tiktok_auth.go` — TikTok OAuth

- [ ] **Step 1: Read all reference files**

Read each file listed above to understand current implementations before writing notes.

- [ ] **Step 2: Write pipeline-overview.md**

Write to `vault/system/architecture/pipeline-overview.md`. Must include:
- Frontmatter: `type: architecture`, `component: pipeline`, `status: current`
- End-to-end flow: NewsAPI → Grok summarization → Grok image gen → Video generation → Optional uploads
- CLI flags: `-upload-youtube`, `-upload-tiktok`, `-sandbox`
- Entry point: `main.go`
- Key dependencies between components

- [ ] **Step 3: Write news-fetching.md**

Write to `vault/system/architecture/news-fetching.md`. Must include:
- Frontmatter: `type: architecture`, `component: news-fetching`, `status: current`
- NewsAPI endpoint: `GET /v2/top-headlines?country=us&sortBy=popularity`
- Auth: `NEWS_API_KEY` env var
- Article selection: picks first article from results
- Output: `AiArticleParameters{ArticleUrl, ArticleTitle}`

- [ ] **Step 4: Write metadata-generation.md**

Write to `vault/system/architecture/metadata-generation.md`. Must include:
- Frontmatter: `type: architecture`, `component: metadata`, `status: current`
- Single Grok API call generates summary + all platform metadata (most token-efficient approach)
- Model: `grok-3` via `github.com/SimonMorphy/grok-go`
- Also generates newsroom background images via X.AI image API (`grok-2-image`)
- System prompt from `metadata/prompts.go`, produces JSON parsed into `LLMMetadataResponse`

- [ ] **Step 5: Write video-generation.md**

Write to `vault/system/architecture/video-generation.md`. Must include:
- Frontmatter: `type: architecture`, `component: video`, `status: proposed`
- Current state: HeyGen code removed, migrating to custom model on AWS GPU
- Target: portrait video (720x1280), 50-70 seconds, AI avatar narrating news summary
- This is the active development area

- [ ] **Step 6: Write upload-system.md**

Write to `vault/system/architecture/upload-system.md`. Must include:
- Frontmatter: `type: architecture`, `component: upload`, `status: current`
- YouTube: YouTube Data API v3, OAuth 2.0, uploads as Shorts (public)
- TikTok: Content Posting API + Login Kit OAuth 2.0, uploads to inbox only (user manually posts)
- Planned platforms: Instagram, Twitter, Facebook, LinkedIn (metadata generated but no upload code yet)

- [ ] **Step 7: Write manifest-system.md**

Write to `vault/system/architecture/manifest-system.md`. Must include:
- Frontmatter: `type: architecture`, `component: manifest`, `status: current`
- `content_manifest.json` — JSON file, append-only with updates
- Thread-safe via `sync.RWMutex` in `ManifestManager`
- Schema: `ContentManifest{version, generated_at, items[]}`
- Each `ContentItem` tracks: source, content, media, SEO, platforms (6), posting status, analytics

- [ ] **Step 8: Commit architecture notes**

```bash
git add vault/system/architecture/
git commit -m "feat: seed vault with architecture notes from existing codebase

Document pipeline overview, news fetching, metadata generation,
video generation (proposed), upload system, and manifest system."
```

### Task 9: Write the HeyGen Migration Decision Note

**Files:**
- Create: `vault/system/decisions/2026-03-16-custom-model-over-heygen.md`

- [ ] **Step 1: Write decision note**

Write to `vault/system/decisions/2026-03-16-custom-model-over-heygen.md`:

```markdown
---
type: decision
component: video
status: accepted
created: 2026-03-16
context: "HeyGen too expensive for target video volume"
tags: [video, cost, aws, heygen]
---

## Decision

Replace HeyGen with a custom video generation model trained and hosted on AWS GPU.

## Context

The pipeline generates AI avatar videos narrating news summaries. HeyGen was the initial provider but costs don't scale for the volume of videos we want to post across 6 social platforms.

## Options Considered

1. **Keep HeyGen** — Proven quality, easy integration, but per-video pricing doesn't scale.
2. **Custom model on AWS GPU** — Higher upfront investment in training, but marginal cost per video drops dramatically at volume.
3. **Alternative SaaS (D-ID, Synthesia)** — Similar pricing models to HeyGen; same scaling problem.

## Rationale

Volume target requires cost-per-video to be near zero after infrastructure costs. A custom model on AWS GPU achieves this. The team has the capability to train and deploy the model.

## Consequences

- `video/` directory (HeyGen code) has been removed
- `audio/` directory (legacy TTS) was already removed — HeyGen handled TTS internally
- New video generation architecture needs to be built from scratch
- Output format must remain: portrait 720x1280, 50-70 seconds, AI avatar with TTS
```

- [ ] **Step 2: Commit decision note**

```bash
git add vault/system/decisions/
git commit -m "feat: add decision note for HeyGen to custom model migration"
```

### Task 10: Write Integration Notes for Implemented Services

**Files:**
- Create: `vault/integrations/social-platforms/youtube-api.md`
- Create: `vault/integrations/social-platforms/tiktok-api.md`
- Create: `vault/integrations/infrastructure/newsapi.md`
- Create: `vault/integrations/infrastructure/grok-xai.md`

**Reference files:** Same as Task 8.

- [ ] **Step 1: Write youtube-api.md**

Write to `vault/integrations/social-platforms/youtube-api.md`. Frontmatter: `type: integration`, `service: youtube`, `category: social-platform`, `status: implemented`, `auth_method: oauth2`. Include:
- YouTube Data API v3, `google.golang.org/api/youtube/v3`
- OAuth 2.0 via `client_secret.json` → browser flow → token stored locally
- Upload endpoint: `Videos.Insert` with `snippet,status` parts
- Shorts: prepend "Shorts" tag, portrait video, title max 100 chars, description max 5000 chars
- Category ID 25 (News & Politics)
- Privacy: public by default

- [ ] **Step 2: Write tiktok-api.md**

Write to `vault/integrations/social-platforms/tiktok-api.md`. Frontmatter: `type: integration`, `service: tiktok`, `category: social-platform`, `status: implemented`, `auth_method: oauth2`. Include:
- Content Posting API + Login Kit (OAuth 2.0)
- Base URL: `https://open.tiktokapis.com`
- Init upload: `POST /v2/post/publish/inbox/video/init/`
- File upload: `PUT {upload_url}` with Content-Range header
- Uploads to inbox ONLY — user must manually post from TikTok app
- Video size max: 301 MB
- Chunking: single chunk if < 64MB, multi-chunk otherwise (5-64MB each)
- Sandbox mode available for testing (separate client key/secret)
- Token stored at `~/.credentials/tiktok-oauth.json`

- [ ] **Step 3: Write newsapi.md**

Write to `vault/integrations/infrastructure/newsapi.md`. Frontmatter: `type: integration`, `service: newsapi`, `category: infrastructure`, `status: implemented`, `auth_method: api-key`. Include:
- `GET https://newsapi.org/v2/top-headlines?country=us&sortBy=popularity`
- Auth: `NEWS_API_KEY` query parameter
- Returns: array of articles with title, URL, source, description
- Current usage: fetch headlines, select first article
- Output: `AiArticleParameters{ArticleUrl, ArticleTitle}`

- [ ] **Step 4: Write grok-xai.md**

Write to `vault/integrations/infrastructure/grok-xai.md`. Frontmatter: `type: integration`, `service: grok-xai`, `category: infrastructure`, `status: implemented`, `auth_method: api-key`. Include:
- Two APIs used:
  1. Chat completions (`grok-3` model) — news summarization + metadata generation
  2. Image generation (`grok-2-image` model) — newsroom backgrounds
- Chat: `github.com/SimonMorphy/grok-go` Go client, 5-minute timeout, temp 0.7, max 4000 tokens
- Images: `POST https://api.x.ai/v1/images/generations`, response_format `b64_json`
- Auth: `X_AI_KEY` env var (Bearer token)
- Single API call generates summary + metadata for all 6 platforms (most token-efficient approach)

- [ ] **Step 5: Commit integration notes**

```bash
git add vault/integrations/
git commit -m "feat: seed vault with integration notes for implemented services

Document YouTube, TikTok, NewsAPI, and Grok/X.AI integrations
with auth details, endpoints, and platform-specific quirks."
```

### Task 11: Create Stub Integration Notes for Planned Services

**Files:**
- Create: `vault/integrations/social-platforms/instagram-api.md`
- Create: `vault/integrations/social-platforms/twitter-api.md`
- Create: `vault/integrations/social-platforms/facebook-api.md`
- Create: `vault/integrations/social-platforms/linkedin-api.md`
- Create: `vault/integrations/infrastructure/aws-gpu.md`

- [ ] **Step 1: Write planned integration stubs**

Each file gets the integration template frontmatter with `status: planned` and a brief Overview section describing the intended use. All other sections contain `TODO: Research and document when implementation begins.`

- [ ] **Step 2: Commit planned stubs**

```bash
git add vault/integrations/
git commit -m "feat: add stub integration notes for planned services

Instagram, Twitter, Facebook, LinkedIn (social platforms) and
AWS GPU (custom video model infrastructure). To be filled when
implementation begins."
```

---

## Chunk 4: MCP Server Setup

> **Spec deviation note:** The spec's `.mcp.json` shows `npx` invocations with `VAULT_PATH` env vars. After researching actual package docs, the correct configurations differ: qmd uses `qmd mcp` (global install with collection registration), and smart-connections requires clone+build with `node` command. The spec's own note acknowledges this: "Verify env variable names against each package's README before implementation."

### Task 12: Install and Configure qmd

qmd is the simpler of the two — it uses its own indexing, no Obsidian plugin dependency.

**Files:**
- Modify: `.mcp.json` (create if doesn't exist)

- [ ] **Step 1: Install qmd globally**

```bash
npm install -g @tobilu/qmd
```

- [ ] **Step 2: Create a qmd collection for the vault**

```bash
cd /Users/slukehart/Documents/Github/content-generation-automation
qmd collection add ./vault --name vault
```

- [ ] **Step 3: Add context descriptions for collections**

```bash
qmd context add qmd://vault "Knowledge base for content-generation-automation project. Contains architecture docs, integration guides, decision records, lessons learned, and prompt engineering notes."
```

- [ ] **Step 4: Generate embeddings**

```bash
qmd embed
```

This downloads a small GGUF model on first run and indexes all markdown files in the vault.

- [ ] **Step 5: Test search**

```bash
qmd search "pipeline" -c vault
qmd vsearch "how does video generation work" -c vault
```

Expected: Returns matching notes from the vault.

- [ ] **Step 6: Configure MCP server**

Create or update `.mcp.json` at project root:

```json
{
  "mcpServers": {
    "qmd": {
      "command": "qmd",
      "args": ["mcp"]
    }
  }
}
```

- [ ] **Step 7: Commit MCP config**

```bash
git add .mcp.json
git commit -m "feat: add qmd MCP server configuration

qmd provides hybrid search (BM25 + vector + LLM reranking) over
the vault for token-efficient agent retrieval."
```

### Task 13: Install and Configure smart-connections

smart-connections requires the Obsidian Smart Connections plugin to build embeddings.

**Files:**
- Modify: `.mcp.json`

- [ ] **Step 1: Open vault in Obsidian**

Open Obsidian → "Open folder as vault" → select `vault/` directory.

- [ ] **Step 2: Install Smart Connections plugin**

In Obsidian: Settings → Community Plugins → Browse → search "Smart Connections" → Install → Enable.

Wait for it to build embeddings (indexes all notes in the vault). This creates a `.smart-connections/` directory inside the vault.

- [ ] **Step 3: Clone and build smart-connections-mcp**

```bash
cd /Users/slukehart/Documents/Github
git clone https://github.com/gogogadgetbytes/smart-connections-mcp.git
cd smart-connections-mcp
npm install
npm run build
```

- [ ] **Step 4: Verify build**

```bash
ls dist/index.js
```

Expected: File exists.

- [ ] **Step 5: Update .mcp.json with smart-connections**

Update `.mcp.json` to add smart-connections server:

```json
{
  "mcpServers": {
    "qmd": {
      "command": "qmd",
      "args": ["mcp"]
    },
    "smart-connections": {
      "command": "node",
      "args": ["/Users/slukehart/Documents/Github/smart-connections-mcp/dist/index.js"],
      "env": {
        "VAULT_PATH": "/Users/slukehart/Documents/Github/content-generation-automation/vault"
      }
    }
  }
}
```

- [ ] **Step 6: Test smart-connections**

Restart Claude Code, then ask it to search the vault using the smart-connections tool.

- [ ] **Step 7: Commit updated MCP config**

```bash
git add .mcp.json
git commit -m "feat: add smart-connections MCP server for semantic search

Pairs with qmd for hybrid retrieval. smart-connections uses
Obsidian Smart Connections plugin embeddings for semantic search."
```

---

## Chunk 5: Prompts & Vision Seed Content

### Task 14: Seed Prompt Notes

**Files:**
- Create: `vault/prompts/metadata-generation.md`
- Create: `vault/prompts/summary-tuning.md`

**Reference:** `metadata/prompts.go`

- [ ] **Step 1: Read metadata/prompts.go**

Read the current prompt templates to understand what's being sent to Grok.

- [ ] **Step 2: Write metadata-generation.md**

Document the current metadata generation prompt: what it asks for, expected JSON output schema, platform-specific formatting rules. Note what works well and what could be improved.

- [ ] **Step 3: Write summary-tuning.md**

Document the summary generation parameters: 150-200 word target, neutral broadcast tone, temperature 0.7, max 4000 tokens. Note observations about quality.

- [ ] **Step 4: Commit prompt notes**

```bash
git add vault/prompts/
git commit -m "feat: add prompt engineering notes for metadata and summary generation"
```

### Task 15: Create Vision Placeholder Files

**Files:**
- Create: `vault/vision/goals.md`
- Create: `vault/vision/roadmap.md`

- [ ] **Step 1: Write goals.md placeholder**

Write to `vault/vision/goals.md`:

```markdown
---
type: vision
author: human
created: 2026-03-17
tags: [goals]
---

# Goals

<!-- This file is yours to fill in. Write about what you're building toward. -->
<!-- Agents will read this for context but will never modify it. -->
```

- [ ] **Step 2: Write roadmap.md placeholder**

Write to `vault/vision/roadmap.md`:

```markdown
---
type: vision
author: human
created: 2026-03-17
tags: [roadmap]
---

# Roadmap

<!-- This file is yours to fill in. Outline phases and priorities. -->
<!-- Agents will read this for context but will never modify it. -->
```

- [ ] **Step 3: Commit vision placeholders**

```bash
git add vault/vision/
git commit -m "feat: add vision placeholder files for human-authored goals and roadmap"
```

---

## Chunk 6: Verification & Final Commit

### Task 16: Verify Complete Setup

- [ ] **Step 1: Verify vault structure**

```bash
find vault -type f -name "*.md" | sort
```

Expected: All notes present across vision/, system/, integrations/, prompts/, templates/.

- [ ] **Step 2: Verify .gitignore**

```bash
cat .gitignore | grep -E "vault|superpowers|output"
```

Expected: `vault/.obsidian/`, `vault/.drafts/`, `vault/.smart-connections/`, `.superpowers/`, `output/`

- [ ] **Step 3: Verify MCP config**

```bash
cat .mcp.json
```

Expected: Both qmd and smart-connections servers configured.

- [ ] **Step 4: Verify qmd search works**

```bash
qmd search "youtube" -c vault
```

Expected: Returns youtube-api.md and upload-system.md (or similar relevant results).

- [ ] **Step 5: Verify project root is clean**

```bash
ls *.md
```

Expected: Only `README.md` remains at project root (all others moved to docs/).

- [ ] **Step 6: Verify no legacy files remain**

```bash
ls video/ WEBAPP_VERIFICATION/ AUDIO_SETUP.md 2>&1
```

Expected: All report "No such file or directory".

- [ ] **Step 7: Re-index qmd after all content is added**

```bash
qmd embed
```

- [ ] **Step 8: Final verification commit**

```bash
git status
git add -A
git commit -m "feat: complete Obsidian vault knowledge base setup

Vault structure with 4 templates, 6 architecture notes,
4 implemented integration docs, 5 planned integration stubs,
1 decision record, 2 prompt notes, and MCP server configuration
for smart-connections (semantic search) and qmd (hybrid search)."
```
