#!/bin/bash
# guard-bash.sh — Block destructive bash commands before execution
# Exit 0 = allow, Exit 2 = block (stderr shown to Claude)

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if [[ -z "$COMMAND" ]]; then
  exit 0
fi

# ── Destructive file operations ──────────────────────────────────
if echo "$COMMAND" | grep -qE 'rm\s+-(r|f|rf|fr)\s+(/|~|\$HOME|\.\./)'; then
  echo "BLOCKED: Recursive/forced delete targeting root, home, or parent directory" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'rm\s+-(r|f|rf|fr)\s+\*'; then
  echo "BLOCKED: Recursive/forced delete with wildcard glob" >&2
  exit 2
fi

# ── Git destructive operations ───────────────────────────────────
if echo "$COMMAND" | grep -qE 'git\s+push\s+.*--force\s.*(main|master)'; then
  echo "BLOCKED: Force push to main/master" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'git\s+push\s+-f\s.*(main|master)'; then
  echo "BLOCKED: Force push to main/master" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'git\s+reset\s+--hard\s+(origin/)?(main|master)'; then
  echo "BLOCKED: Hard reset to main/master — use a branch instead" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'git\s+clean\s+-fd'; then
  echo "BLOCKED: git clean -fd removes untracked files permanently" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'git\s+branch\s+-D\s+(main|master)'; then
  echo "BLOCKED: Deleting main/master branch" >&2
  exit 2
fi

# ── System-level dangers ─────────────────────────────────────────
if echo "$COMMAND" | grep -qE '^\s*sudo\s'; then
  echo "BLOCKED: sudo commands require explicit user approval" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'chmod\s+-R\s+777'; then
  echo "BLOCKED: chmod -R 777 is a security risk" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'mkfs|fdisk|dd\s+if=.*of=/dev'; then
  echo "BLOCKED: Disk/partition operations not allowed" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE ':()\s*\{\s*:\|:&\s*\}\s*;'; then
  echo "BLOCKED: Fork bomb detected" >&2
  exit 2
fi

# ── Credential / secret exposure ─────────────────────────────────
if echo "$COMMAND" | grep -qE 'cat\s+.*\.(env|pem|key|secret|credentials)'; then
  echo "BLOCKED: Printing secrets/credentials to stdout" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'curl\s+.*(-d|--data).*\b(token|password|secret|key)\b.*https?://'; then
  echo "BLOCKED: Sending credentials via curl — verify the endpoint first" >&2
  exit 2
fi

# ── Network exfiltration patterns ────────────────────────────────
if echo "$COMMAND" | grep -qE 'curl\s+.*-X\s*POST.*\|\s*base64'; then
  echo "BLOCKED: Suspicious data exfiltration pattern" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'nc\s+-.*\d+\.\d+\.\d+\.\d+'; then
  echo "BLOCKED: Netcat to external IP" >&2
  exit 2
fi

# ── Process killing ──────────────────────────────────────────────
if echo "$COMMAND" | grep -qE 'kill\s+-9\s+-1'; then
  echo "BLOCKED: kill -9 -1 would kill all user processes" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'pkill\s+-(9|KILL)\s+(Finder|loginwindow|WindowServer|dock)'; then
  echo "BLOCKED: Killing system-critical macOS processes" >&2
  exit 2
fi

# ── Package/dependency attacks ───────────────────────────────────
if echo "$COMMAND" | grep -qE '(npm|pip|brew)\s+install\s+.*\|'; then
  echo "BLOCKED: Piping package install output is suspicious" >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'curl\s+.*\|\s*(bash|sh|zsh)'; then
  echo "BLOCKED: Piping remote script to shell — download and review first" >&2
  exit 2
fi

exit 0
