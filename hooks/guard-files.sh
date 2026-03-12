#!/bin/bash
# guard-files.sh — Block writes/edits to sensitive files
# Exit 0 = allow, Exit 2 = block (stderr shown to Claude)

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

if [[ -z "$FILE_PATH" ]]; then
  exit 0
fi

# ── Secrets and credentials ──────────────────────────────────────
case "$FILE_PATH" in
  *.env|*.env.local|*.env.production)
    echo "BLOCKED: Cannot write to env file — $FILE_PATH" >&2
    exit 2
    ;;
  *credentials.json|*client_secret*.json|*service_account*.json)
    echo "BLOCKED: Cannot write to credential file — $FILE_PATH" >&2
    exit 2
    ;;
  *.pem|*.key|*.p12|*.pfx|*id_rsa|*id_ed25519)
    echo "BLOCKED: Cannot write to private key — $FILE_PATH" >&2
    exit 2
    ;;
esac

# ── SSH config ───────────────────────────────────────────────────
if [[ "$FILE_PATH" == *"/.ssh/"* ]]; then
  echo "BLOCKED: Cannot modify SSH directory — $FILE_PATH" >&2
  exit 2
fi

# ── Git internals ────────────────────────────────────────────────
if [[ "$FILE_PATH" == *"/.git/"* ]]; then
  echo "BLOCKED: Cannot modify git internals — $FILE_PATH" >&2
  exit 2
fi

# ── System files ─────────────────────────────────────────────────
if [[ "$FILE_PATH" == /etc/* || "$FILE_PATH" == /System/* || "$FILE_PATH" == /Library/* ]]; then
  echo "BLOCKED: Cannot modify system files — $FILE_PATH" >&2
  exit 2
fi

# ── Shell profiles (prevent persistent backdoors) ────────────────
case "$FILE_PATH" in
  */.zshrc|*/.bashrc|*/.bash_profile|*/.zprofile|*/.profile|*/.zshenv)
    echo "BLOCKED: Cannot modify shell profile — $FILE_PATH — do this manually" >&2
    exit 2
    ;;
esac

exit 0
