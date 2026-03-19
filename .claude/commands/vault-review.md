# Vault Review

Review the Obsidian knowledge vault for accuracy and completeness.

## Mode

Check if the user provided an argument:
- No argument or empty → **Session-focused review**
- `full` → **Full audit**

---

## Session-Focused Review (Default)

Analyze what changed recently and check if vault documentation needs updating.

### Steps

1. **Identify recent changes:**
   - Run `git diff HEAD` to see uncommitted changes
   - Run `git log --oneline -10` to see recent commits
   - Use judgment to determine the relevant scope of changes

2. **Find related vault notes:**
   - For each changed file/package, search the vault using `qmd` MCP: `search "<package-name>"` or `search "<filename>"`
   - Also try `smart-connections` MCP for semantic matches if keyword search misses relevant notes

3. **Compare vault notes against code:**
   - For each related vault note, read the note and the current code
   - Check: Are file paths still correct? Are function signatures accurate? Are env vars still used? Is the described behavior still true?

4. **Check index.md currency:**
   - For any vault folder where files were added or removed, verify the `index.md` lists all current files

5. **Check wiki-link opportunities:**
   - If new vault notes were created, do they have `## Related` sections?
   - Are there new cross-references that should be added?

6. **Output report:**

```
## Vault Review Report (Session)

### Notes Needing Updates
- `vault/system/architecture/<note>.md` — <specific stale section and what changed>

### Missing Coverage
- `<package>/` was modified but has no vault documentation → suggest drafting an architecture note

### Index Updates Needed
- `vault/<folder>/index.md` — missing entry for `<new-file>.md`

### Wiki-Link Suggestions
- `<note-a>` should link to `[[<note-b>]]` — <reason>

### No Issues Found
- (if everything is current)
```

7. **Ask for approval** before making any changes.

---

## Full Audit

Comprehensive review of the entire vault against the current codebase.

### Steps

1. **Read all vault notes:**
   - Use `find vault -name "*.md" -not -path "*/templates/*" -not -name "index.md" -not -name "README.md"` to list all content notes
   - Read each note

2. **Verify claims against codebase:**
   - For each architecture note: verify file paths exist, function names are correct, env vars are still used, behavior descriptions match code
   - For each integration note: verify API endpoints, auth flows, credential paths
   - For each decision note: verify the decision is still current (not superseded)
   - For each prompt note: verify prompt text matches what the code actually sends

3. **Check frontmatter accuracy:**
   - `status` field: is `current` still current? Is `planned` now implemented? Is `implemented` now outdated?
   - `updated` field: does it reflect the last meaningful edit?

4. **Check all index.md files:**
   - Does each index.md list every content file in its folder?
   - Are descriptions still accurate?

5. **Check wiki-links:**
   - Does every content note have a `## Related` section?
   - Are all `[[links]]` resolving to real files?
   - Are there missing cross-references based on content overlap?

6. **Output prioritized report:**

```
## Vault Full Audit Report

### Critical (claims contradicted by code)
- ...

### Stale (outdated but not wrong)
- ...

### Missing Coverage
- ...

### Frontmatter Issues
- ...

### Index Issues
- ...

### Wiki-Link Issues
- ...
```

7. **Ask for approval** before making any changes.

---

## Governance

- This skill produces a **report only** — never auto-fix
- All changes require user approval
- Fixes follow the vault's draft-review-promote model
- For new notes, draft to `vault/.drafts/` using templates
