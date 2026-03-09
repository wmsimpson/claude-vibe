#!/bin/bash
# job_wrapper.sh — Invoked by launchd to run a scheduled Claude prompt via isaac
# Usage: job_wrapper.sh <job_id>

set -euo pipefail

JOB_ID="$1"
SCHEDULER_DIR="$HOME/.vibe/scheduler"
JOBS_FILE="$SCHEDULER_DIR/jobs.json"
LOG_DIR="$SCHEDULER_DIR/logs"

# Ensure log directory exists
mkdir -p "$LOG_DIR"

LOG_FILE="$LOG_DIR/${JOB_ID}.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG_FILE"
}

log "=== Job execution started: $JOB_ID ==="

# Read the prompt, allowed_tools, and allowed_channels from the jobs registry
if [ ! -f "$JOBS_FILE" ]; then
    log "ERROR: Jobs file not found at $JOBS_FILE"
    exit 1
fi

JOB_DATA=$(python3 -c "
import json, sys
with open('$JOBS_FILE') as f:
    jobs = json.load(f)
for job in jobs:
    if job['id'] == '$JOB_ID':
        # Line 1: prompt, Line 2: allowed_tools, Line 3: allowed_channels
        print(job['prompt'])
        tools = job.get('allowed_tools', [])
        print(','.join(tools) if tools else '')
        channels = job.get('allowed_channels', [])
        print(','.join(channels) if channels else '')
        sys.exit(0)
print('JOB_NOT_FOUND', file=sys.stderr)
sys.exit(1)
" 2>>"$LOG_FILE")

# Parse multi-line output: last line = channels, second-to-last = tools, everything else = prompt
TOTAL_LINES=$(echo "$JOB_DATA" | wc -l)
if [ "$TOTAL_LINES" -ge 3 ]; then
    PROMPT_END=$((TOTAL_LINES - 2))
    PROMPT=$(echo "$JOB_DATA" | sed -n "1,${PROMPT_END}p")
    ALLOWED_TOOLS=$(echo "$JOB_DATA" | sed -n "$((TOTAL_LINES - 1))p")
    ALLOWED_CHANNELS=$(echo "$JOB_DATA" | sed -n "${TOTAL_LINES}p")
else
    PROMPT=$(echo "$JOB_DATA" | head -1)
    ALLOWED_TOOLS=$(echo "$JOB_DATA" | sed -n '2p')
    ALLOWED_CHANNELS=""
fi

if [ -z "$PROMPT" ]; then
    log "ERROR: Could not read prompt for job $JOB_ID"
    exit 1
fi

# Defense-in-depth: inject channel restriction at runtime even if jobs.json was edited
if [ -n "$ALLOWED_CHANNELS" ]; then
    PROMPT="CHANNEL RESTRICTION: You may ONLY post to these Slack channels: ${ALLOWED_CHANNELS}. Do not post to any other channel.

${PROMPT}"
    log "Allowed channels: $ALLOWED_CHANNELS"
fi

log "Running prompt: $PROMPT"
if [ -n "$ALLOWED_TOOLS" ]; then
    log "Allowed tools: $ALLOWED_TOOLS"
fi

# Unset CLAUDECODE to avoid nested session errors
unset CLAUDECODE 2>/dev/null || true

# Determine which runner to use: isaac (preferred) or llm agent (fallback)
RUNNER=""
ISAAC_BIN=$(which isaac 2>/dev/null || true)

if [ -n "$ISAAC_BIN" ] && [ -x "$ISAAC_BIN" ]; then
    RUNNER="isaac"
    RUNNER_BIN="$ISAAC_BIN"
else
    # isaac is often an alias — check for dbexec
    DBEXEC_BIN=$(which dbexec 2>/dev/null || true)
    if [ -n "$DBEXEC_BIN" ] && [ -x "$DBEXEC_BIN" ]; then
        RUNNER="dbexec"
        RUNNER_BIN="$DBEXEC_BIN"
    else
        # Fall back to llm agent
        LLM_BIN=$(which llm 2>/dev/null || true)
        if [ -n "$LLM_BIN" ] && [ -x "$LLM_BIN" ]; then
            RUNNER="llm"
            RUNNER_BIN="$LLM_BIN"
        else
            log "ERROR: Neither isaac nor llm binary found"
            exit 1
        fi
        log "WARNING: isaac not found, falling back to llm agent"
    fi
fi

# Execute based on runner
# Prompt is always piped via stdin to avoid arg-parsing issues with special chars.
# --allowedTools uses comma-separated format as a single arg.
if [ "$RUNNER" = "isaac" ] || [ "$RUNNER" = "dbexec" ]; then
    CMD_ARGS=(start --print --no-session-persistence --dangerously-skip-permissions)

    if [ -n "$ALLOWED_TOOLS" ]; then
        CMD_ARGS+=(--allowedTools "$ALLOWED_TOOLS")
    fi

    if [ "$RUNNER" = "dbexec" ]; then
        log "Executing: echo <prompt> | $RUNNER_BIN repo run isaac -- ${CMD_ARGS[*]}"
        echo "$PROMPT" | "$RUNNER_BIN" repo run isaac -- "${CMD_ARGS[@]}" >> "$LOG_FILE" 2>&1
    else
        log "Executing: echo <prompt> | $RUNNER_BIN ${CMD_ARGS[*]}"
        echo "$PROMPT" | "$RUNNER_BIN" "${CMD_ARGS[@]}" >> "$LOG_FILE" 2>&1
    fi
else
    # llm agent fallback — pipe prompt via stdin
    CMD_ARGS=(agent)

    if [ -n "$ALLOWED_TOOLS" ]; then
        CMD_ARGS+=(--allowedTools "$ALLOWED_TOOLS")
    fi

    log "Executing: echo <prompt> | $RUNNER_BIN ${CMD_ARGS[*]}"
    echo "$PROMPT" | "$RUNNER_BIN" "${CMD_ARGS[@]}" >> "$LOG_FILE" 2>&1
fi

EXIT_CODE=$?

log "=== Job execution finished (exit code: $EXIT_CODE) ==="
