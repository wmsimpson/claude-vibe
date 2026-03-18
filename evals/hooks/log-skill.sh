#!/bin/bash
# Receives JSON via stdin from Claude Code hook
# Logs skill invocation to eval log file

INPUT=$(cat)
SKILL=$(echo "$INPUT" | jq -r '.tool_input.skill // "unknown"')
ARGS=$(echo "$INPUT" | jq -r '.tool_input.args // ""')
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Append to log file (path set via environment variable)
echo "{\"timestamp\":\"$TIMESTAMP\",\"skill\":\"$SKILL\",\"args\":\"$ARGS\"}" >> "${EVAL_LOG_FILE:-/tmp/claude-eval-skills.jsonl}"

exit 0
