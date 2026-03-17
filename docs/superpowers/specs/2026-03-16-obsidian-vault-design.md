# Obsidian Vault Knowledge Base — Design Spec

**Date:** 2026-03-16
**Status:** Approved
**Inspiration:** [@nyk_builderz memory stack](https://x.com/nyk_builderz/status/2030904887186514336), [@Atenov_D second brain](https://x.com/Atenov_D/status/2032528386745315553)

## Problem

AI agents working on this project lose context between sessions. There is no persistent, queryable knowledge base where agents can learn how the system works, what decisions were made and why, or what gotchas exist. The only state is `content_manifest.json`, which is a production log — not a knowledge store.

## Solution

Add an Obsidian vault at `vault/` inside the project that serves as an agent-readable knowledge base. Agents query it at runtime via MCP servers (semantic search + structured queries) for token-efficient context retrieval. The vault is organized into human-owned and agent-owned zones with a draft-based approval workflow.

## Goals

1. Give AI agents persistent, queryable memory about how the system works
2. Build institutional knowledge that compounds over time
3. Document external service integrations (6 social platforms + AWS GPU + NewsAPI + Grok)
4. Provide token-efficient context retrieval via MCP servers
5. Maintain human oversight — agents draft, user approves

## Non-Goals

- Replacing `content_manifest.json` — manifest stays as the pipeline's working store
- Content/editorial knowledge (topic tracking, audience analytics)
- Cross-project knowledge base — vault is scoped to this project only

---

## 1. Project Cleanup

Before adding the vault, clean up legacy files from the HeyGen-era codebase.

### Files to Remove

| Path | Reason |
|------|--------|
| `video/` | HeyGen integration code, being replaced by custom model |
| `WEBAPP_VERIFICATION/` | Empty directory |
| `poetry.lock` | Already in .gitignore |
| `AUDIO_SETUP.md` | Legacy TTS setup docs |
| `AUDIO_VIDEO_SETUP.md` | Legacy combined setup docs |
| `VIDEO_SETUP.md` | HeyGen setup instructions |
| `TTS_VIDEO_GUIDE.md` | Legacy TTS guide |
| `QUICKSTART.md` | References HeyGen workflow, needs full rewrite |

> **Note:** `audio/`, `content-gen`, and `test_video.sh` were already removed in prior cleanup.

### Files to Reorganize

| From | To | Reason |
|------|----|--------|
| `TIKTOK_WORKFLOW_DIAGRAM.md` | `docs/guides/` | Consolidate docs |
| `TIKTOK_USAGE_GUIDE.md` | `docs/guides/` | Consolidate docs |
| `METADATA_GUIDE.md` | `docs/guides/` | Consolidate docs |
| `WORKFLOW.md` | `docs/guides/` | Consolidate docs |
| `GITHUB_PAGES_SETUP.md` | `docs/setup/` | Consolidate docs |
| `LEGAL_README.md` | `docs/legal/` | Consolidate legal |
| `PRIVACY_POLICY.md` + `PRIVACY_POLICY/` | `docs/legal/` | Consolidate legal |
| `TERMS_OF_SERVICE.md` + `TERMS_OF_SERVICE/` | `docs/legal/` | Consolidate legal |
| `privacy_policy.html` | `docs/legal/` | Consolidate legal |
| `terms_of_service.html` | `docs/legal/` | Consolidate legal |
| `news_*_final.mp4` (12 files, ~100MB) | `output/videos/` | Clean up root |

---

## 2. Vault Structure

```
vault/
├── vision/                          # Human writes, agents read-only
│   ├── goals.md
│   ├── roadmap.md
│   └── feature-ideas/
│
├── system/                          # Agents draft, human approves
│   ├── architecture/
│   │   ├── pipeline-overview.md
│   │   ├── news-fetching.md
│   │   ├── metadata-generation.md
│   │   ├── video-generation.md      # Custom model on AWS GPU
│   │   ├── upload-system.md
│   │   └── manifest-system.md
│   ├── decisions/
│   │   └── 2026-03-16-custom-model-over-heygen.md
│   └── lessons/
│       └── (component-named, e.g., tiktok-api-quirks.md)
│
├── integrations/                    # Both write
│   ├── social-platforms/
│   │   ├── youtube-api.md           ✓ implemented
│   │   ├── tiktok-api.md            ✓ implemented
│   │   ├── instagram-api.md         ○ planned
│   │   ├── twitter-api.md           ○ planned
│   │   ├── facebook-api.md          ○ planned
│   │   └── linkedin-api.md          ○ planned
│   └── infrastructure/
│       ├── aws-gpu.md               ○ planned
│       ├── newsapi.md               ✓ implemented
│       └── grok-xai.md              ✓ implemented
│
├── prompts/                         # Both write
│   ├── metadata-generation.md
│   └── summary-tuning.md
│
├── templates/                       # Shared
│   ├── architecture-note.md
│   ├── decision-note.md
│   ├── lesson-note.md
│   └── integration-note.md
│
└── .drafts/                         # Agent staging area (gitignored)
```

### Ownership Rules

- **`vision/`** — Human-only writes. Agents read for context on goals and direction.
- **`system/`** — Agents draft to `.drafts/`, human approves and moves to final location.
- **`integrations/`** — Both write. Agents document API details; human adds strategic context.
- **`prompts/`** — Both write. Human tunes prompts; agents document what configurations work.
- **`templates/`** — Shared. Defines frontmatter schema and note structure for consistency.

---

## 3. Frontmatter Schema

Consistent frontmatter enables structured queries via the qmd MCP server.

### Architecture Notes (`system/architecture/*.md`)

```yaml
---
type: architecture
component: pipeline | news-fetching | metadata | video | upload | manifest
status: current | outdated | proposed
created: YYYY-MM-DD
updated: YYYY-MM-DD
tags: []
---
```

### Decision Notes (`system/decisions/*.md`)

```yaml
---
type: decision
component: video | metadata | upload | ...
status: accepted | superseded | proposed
created: YYYY-MM-DD
context: "short description of why this decision was made"
tags: []
---
```

### Lesson Notes (`system/lessons/*.md`)

```yaml
---
type: lesson
relates_to: tiktok | youtube | instagram | twitter | facebook | linkedin | aws-gpu | newsapi | grok-xai | pipeline | metadata | video | upload | manifest
severity: high | medium | low
created: YYYY-MM-DD
tags: []
---
```

> **Note:** Lessons use `relates_to` instead of `component` because a lesson can relate to either a system component (pipeline, upload) or an external service (tiktok, youtube). This avoids conflating the two dimensions.

### Integration Notes (`integrations/**/*.md`)

```yaml
---
type: integration
service: youtube | tiktok | instagram | twitter | facebook | linkedin | aws-gpu | newsapi | grok-xai
category: social-platform | infrastructure
status: implemented | in-progress | planned
auth_method: oauth2 | api-key | iam
created: YYYY-MM-DD
updated: YYYY-MM-DD
tags: []
---
```

### Integration Note Sections

Each integration note covers:

1. **Overview** — what this service does in the pipeline
2. **Authentication** — OAuth flow, token management, scopes
3. **API Endpoints** — key endpoints, request/response formats
4. **Content Requirements** — video specs, caption limits, hashtag rules
5. **Rate Limits & Quotas** — API limits, retry strategies
6. **Gotchas** — platform-specific quirks agents should know
7. **Implementation Status** — what's built, pending, blocked

---

## 4. MCP Server Configuration

Two MCP servers provide token-efficient retrieval for agents.

### smart-connections — Semantic Search

Finds notes by meaning. Agent asks "what do we know about upload failures?" and gets the most relevant notes regardless of exact keyword match.

### qmd — Structured Queries

Queries frontmatter fields. Agent asks for `type:integration AND status:planned` and gets exactly those notes.

### Configuration (`.mcp.json` at project root)

```json
{
  "mcpServers": {
    "smart-connections": {
      "command": "npx",
      "args": ["-y", "@gogogadgetbytes/smart-connections-mcp"],
      "env": {
        "VAULT_PATH": "./vault"
      }
    },
    "qmd": {
      "command": "npx",
      "args": ["-y", "@tobilu/qmd"],
      "env": {
        "VAULT_PATH": "./vault"
      }
    }
  }
}
```

> **Note:** Verify env variable names against each package's README before implementation. The `VAULT_PATH` variable name shown here may differ — check `@gogogadgetbytes/smart-connections-mcp` (v0.2.0) and `@tobilu/qmd` (v2.0.1) docs.

### Token Efficiency

Without MCP: Agent runs Glob → Read on ~20 files → thousands of tokens consumed scanning for relevance.

With MCP: Agent makes one search call → gets back 2-3 relevant notes with frontmatter → hundreds of tokens.

---

## 5. Agent Write-Back Workflow

### Draft Location

Agents write to `vault/.drafts/` — never directly to vault folders.

### Draft Lifecycle

1. Agent finishes a task (e.g., implements Instagram upload)
2. Agent drafts relevant notes to `vault/.drafts/` using the appropriate template
3. Agent notifies user: "Drafted `integration-instagram-api.md` and `lesson-instagram-oauth-scopes.md` in `vault/.drafts/`"
4. User reviews in Obsidian or editor
5. User moves approved files to correct vault folder or deletes/edits
6. If `.drafts/` already has pending notes, agents draft anyway using date-prefixed filenames (e.g., `2026-03-16-integration-instagram-api.md`) and warn the user that older drafts are still pending review

### When to Draft

Agents should draft when:

- A new component is built or significantly changed → architecture note
- A non-obvious decision was made → decision note
- Something broke or was surprising → lesson note
- A new external service was integrated → integration note

Agents should NOT draft for: small bug fixes, config tweaks, formatting changes.

### Draft Format

Every draft includes a `draft_reason` field in frontmatter:

```yaml
---
type: lesson
relates_to: tiktok
severity: high
created: 2026-03-16
draft_reason: "TikTok API returns 200 OK on failed uploads — agents need to check publish_status field"
tags: [tiktok, api, upload]
---
```

---

## 6. .gitignore Updates

Add to `.gitignore`:

```
# Obsidian
vault/.obsidian/
vault/.drafts/
vault/.smart-connections/

# Superpowers brainstorm sessions
.superpowers/
```

- `vault/.obsidian/` — Obsidian's local settings (themes, plugins, workspace state)
- `vault/.drafts/` — Agent staging area, not committed
- `vault/.smart-connections/` — Embedding index generated by smart-connections MCP server for semantic search. Rebuilt automatically on first query.

The vault notes themselves (`vault/system/`, `vault/integrations/`, etc.) ARE committed — they're part of the project's institutional knowledge.
