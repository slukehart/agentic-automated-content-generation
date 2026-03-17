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
