#!/usr/bin/env bash
# SessionStart hook: surface pending vault inbox and draft items

VAULT_DIR="vault"
INBOX_DIR="$VAULT_DIR/inbox"
DRAFTS_DIR="$VAULT_DIR/.drafts"

output=""

# Check inbox
if [ -d "$INBOX_DIR" ]; then
    inbox_files=$(find "$INBOX_DIR" -name "*.md" -not -name "index.md" 2>/dev/null)
    if [ -n "$inbox_files" ]; then
        count=$(echo "$inbox_files" | wc -l | tr -d ' ')
        output+="📥 Vault Inbox ($count note(s) waiting):\n"
        while IFS= read -r f; do
            output+="  - $(basename "$f")\n"
        done <<< "$inbox_files"
        output+="\n"
    fi
fi

# Check drafts
if [ -d "$DRAFTS_DIR" ]; then
    draft_files=$(find "$DRAFTS_DIR" -name "*.md" 2>/dev/null)
    if [ -n "$draft_files" ]; then
        count=$(echo "$draft_files" | wc -l | tr -d ' ')
        output+="📝 Vault Drafts ($count note(s) awaiting review):\n"
        while IFS= read -r f; do
            output+="  - $(basename "$f")\n"
        done <<< "$draft_files"
        output+="\n"
    fi
fi

# Only print if there's something to report
if [ -n "$output" ]; then
    echo -e "$output"
    echo "Review these before starting new work."
fi
