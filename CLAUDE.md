# Content Generation Automation

Go-based pipeline that fetches news, generates AI metadata + video, and uploads to social platforms.

## Codebase

- **Language:** Go
- **Entry point:** `main.go`
- **Pipeline:** news fetch → metadata generation (Grok) → video generation (AWS GPU, WIP) → upload (YouTube, TikTok)
- **Key packages:** `news/`, `metadata/`, `video/`, `media/`
- **Config:** Environment variables — `NEWS_API_KEY`, `X_AI_KEY`, `TIKTOK_CLIENT_KEY`, `TIKTOK_CLIENT_SECRET`

## Knowledge Vault (`vault/`)

Obsidian-based knowledge base documenting this project. Use it as primary context before exploring code.

### Searching the vault

- **Semantic search:** `smart-connections` MCP — use for conceptual queries ("how does upload work?")
- **Keyword/hybrid search:** `qmd` MCP — use for specific lookups (e.g., `search "manifest"`)
- **Direct read:** `vault/` folder tree when you know the exact file
- **Orientation:** Each vault subfolder has an `index.md` listing its contents — read these first

### Writing to the vault

- **Never** write directly to vault folders
- Draft to `vault/.drafts/` using templates from `vault/templates/`
- Include `draft_reason` in frontmatter explaining why this note is worth keeping
- Notify the user so they can review and promote to the correct folder
- If `.drafts/` has pending notes, use date-prefixed filenames and warn user

### When to draft a vault note

- New component built or significantly changed → architecture note
- Non-obvious decision made → decision note
- Something broke or was surprising → lesson note
- New external service integrated → integration note
- Do NOT draft for: small bug fixes, config tweaks, formatting changes

### Index maintenance

- Every vault subfolder has an `index.md` listing its contents
- **When you create or delete a vault file, update the `index.md` in that folder**
- Keep index entries as one-line descriptions, not full summaries

## Conventions

- Commit messages: short, imperative, lowercase
- Vault notes use YAML frontmatter (see `vault/templates/`)
- Use `[[wiki-links]]` between vault notes for Obsidian graph connectivity
