# Vault Enhancements Design

**Date:** 2026-03-19
**Status:** Approved
**Source:** Review of [Claude Code + Obsidian article](https://www.whytryai.com/p/claude-code-obsidian) against current vault setup

## Context

The existing Obsidian vault (`vault/`) is well-structured with architecture docs, integration guides, decision records, prompt notes, and templates. After comparing against community best practices, five enhancements were identified to improve agent-vault integration, discoverability, and maintenance.

## Implementation Approach

Layered, in dependency order:
1. CLAUDE.md (foundational)
2. index.md files + wiki-links (vault content, parallel)
3. SessionStart hook (depends on inbox folder)
4. /vault-review skill (depends on CLAUDE.md conventions)

---

## 1. CLAUDE.md (Project Root)

**Purpose:** Ensure every Claude Code session automatically knows about the project, vault, MCP servers, and conventions.

**Location:** `/CLAUDE.md` (project root)

**Content covers:**
- Project overview (Go pipeline, entry point, key packages, env vars)
- Vault search instructions (smart-connections for semantic, qmd for keyword/hybrid)
- Vault writing rules (draft to `.drafts/`, use templates, notify user)
- When to draft (architecture/decision/lesson/integration notes)
- Index maintenance rule: "When you create or delete a vault file, update the index.md in that folder"
- Conventions: commit style, YAML frontmatter, `[[wiki-links]]` usage

**No enforcement hook** — relies on Claude following CLAUDE.md instructions. Can add hook later if unreliable.

---

## 2. index.md Files

**Purpose:** Provide instant orientation for each vault subfolder without requiring file reads or glob operations.

**Format:**
```markdown
# {Folder Name}

{One-line description of folder purpose.}

## Contents

- [[note-name]] — One-line description
- [[other-note]] — One-line description
```

**Folders receiving an index.md (13 total):**
- `vault/` (root-level map)
- `vault/vision/`
- `vault/vision/feature-ideas/`
- `vault/system/`
- `vault/system/architecture/`
- `vault/system/decisions/`
- `vault/system/lessons/`
- `vault/integrations/`
- `vault/integrations/infrastructure/`
- `vault/integrations/social-platforms/`
- `vault/prompts/`
- `vault/templates/`
- `vault/inbox/` (new folder, see Section 4)

**Excluded folders:** `.drafts/`, `.obsidian/`, `.smart-env/` (all git-ignored, transient)

**vault/README.md vs vault/index.md:** README.md is kept as-is for GitHub rendering and agent governance rules. The root `vault/index.md` serves as the Obsidian-native content map with wiki-links. Different purposes, no overlap.

**Empty index files on day one:** `system/lessons/`, `vision/feature-ideas/`, and `inbox/` contain only `.gitkeep` files. Their index.md files will have an empty `## Contents` section — this is expected and they will be populated as content is added.

**Maintenance:** CLAUDE.md instruction — agents update index.md on file create/delete.

---

## 3. Wiki-Links

**Purpose:** Build Obsidian's knowledge graph with meaningful cross-references between notes.

**Approach:** Rich cross-linking (~50+ links across 18 notes). Links placed in a `## Related` section at the bottom of each note to keep main body clean.

**Link format:** Obsidian shortest-path — `[[note-name]]` without folder paths.

**Cross-reference map:**

| Note | Links to |
|------|----------|
| pipeline-overview | news-fetching, metadata-generation, video-generation, manifest-system, upload-system |
| news-fetching | pipeline-overview, newsapi, metadata-generation |
| metadata-generation | pipeline-overview, grok-xai, summary-tuning, metadata-generation-prompt, video-generation |
| video-generation | pipeline-overview, 2026-03-16-custom-model-over-heygen, aws-gpu, manifest-system |
| manifest-system | pipeline-overview, upload-system |
| upload-system | pipeline-overview, manifest-system, youtube-api, tiktok-api |
| 2026-03-16-custom-model-over-heygen | video-generation, aws-gpu |
| newsapi | news-fetching |
| grok-xai | metadata-generation, summary-tuning |
| aws-gpu | video-generation, 2026-03-16-custom-model-over-heygen |
| youtube-api | upload-system |
| tiktok-api | upload-system |
| instagram-api | upload-system |
| twitter-api | upload-system |
| facebook-api | upload-system |
| linkedin-api | upload-system |
| summary-tuning | metadata-generation, grok-xai |
| metadata-generation-prompt | metadata-generation, grok-xai, summary-tuning |

**Name collision fix:** Rename `vault/prompts/metadata-generation.md` to `vault/prompts/metadata-generation-prompt.md` to avoid ambiguity with `vault/system/architecture/metadata-generation.md`. Both share the base name `metadata-generation` which causes unpredictable Obsidian wiki-link resolution.

**Excluded from wiki-links:** `vision/goals.md` and `vision/roadmap.md` are human-authored placeholder files (currently empty). They will appear in `vault/vision/index.md` but will not have `## Related` sections until they have content to cross-reference.

---

## 4. SessionStart Hook

**Purpose:** Surface pending vault items at the start of every Claude Code session.

**What it checks:**
1. `vault/inbox/` — human-captured notes (brain dumps, mobile capture)
2. `vault/.drafts/` — agent-drafted notes awaiting review

**Behavior:**
- If either folder has `.md` files, print a summary listing each file
- If both are empty, output nothing (silent on clean sessions)

**Implementation:**
- Create `vault/inbox/` folder with `.gitkeep`
- Gitignore pattern: `vault/inbox/*` with `!vault/inbox/.gitkeep` exception (so the folder is tracked but contents are not)
- Force-add `.gitkeep` before gitignore rule: `git add -f vault/inbox/.gitkeep`
- Create script at `.claude/hooks/session-start-vault-check.sh`
- Register hook in `.claude/settings.local.json`:
  ```json
  "hooks": {
    "SessionStart": [{
      "type": "command",
      "command": "bash .claude/hooks/session-start-vault-check.sh"
    }]
  }
  ```

**Example output:**
```
📥 Vault Inbox (2 notes waiting):
  - inbox/mobile-idea-video-scheduling.md
  - inbox/brain-dump-2026-03-19.md

📝 Vault Drafts (1 note awaiting review):
  - .drafts/2026-03-18-architecture-video-generation.md

Review these before starting new work.
```

---

## 5. /vault-review Skill

**Purpose:** Keep vault documentation accurate and current through structured review.

**Location:** `.claude/commands/vault-review.md`

**Two modes:**

### Session-focused (default: `/vault-review`)
1. Check what changed: `git diff HEAD` (uncommitted) + `git log --oneline -10` (recent commits). The agent uses judgment to determine relevant scope — there is no formal "session boundary" in git.
2. Search vault (via qmd/smart-connections) for notes related to changed files
3. Compare vault notes against current code state
4. Output report:
   - Notes that need updating (with specific stale sections)
   - New notes that should be drafted (changed components with no vault coverage)
   - New `[[wiki-links]]` to add
   - index.md files that need updating

### Full audit (`/vault-review full`)
1. Read every vault note
2. For each architecture/integration note, verify claims against codebase (file paths, function signatures, env vars)
3. Flag: stale content, missing coverage, broken links, outdated frontmatter status fields
4. Suggest new `[[wiki-links]]` based on content overlap
5. Check all index.md files are current
6. Output prioritized action list

**Governance:** Report-only — no auto-fix. Asks for approval before making changes. Consistent with vault's draft-review-promote model.

---

## Files Created/Modified

**New files:**
- `/CLAUDE.md`
- `vault/index.md` (+ 12 more subfolder index.md files)
- `vault/inbox/.gitkeep`
- `.claude/hooks/session-start-vault-check.sh`
- `.claude/commands/vault-review.md`

**Renamed files:**
- `vault/prompts/metadata-generation.md` → `vault/prompts/metadata-generation-prompt.md` (resolve name collision)

**Modified files:**
- `.claude/settings.local.json` (add SessionStart hook)
- `.gitignore` (add `vault/inbox/*` with `.gitkeep` exception)
- 18 existing vault `.md` files (add `## Related` wiki-link sections)
