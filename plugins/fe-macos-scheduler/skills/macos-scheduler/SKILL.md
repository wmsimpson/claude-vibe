---
name: macos-scheduler
description: Schedule and manage recurring LOCAL macOS launchd tasks on this computer. NOT for Databricks jobs or cloud workloads. Create, list, run, view logs, and remove launchd-based scheduled tasks that execute Claude prompts locally — on time intervals (every N minutes), calendar schedules (weekdays at 9am, 1st of month), or file-change watchers. Use this for local automation like periodic Slack messages, Google Docs updates, file monitoring, and any recurring task on the user's Mac. Keywords — scheduled task, background task, launchd, cron, recurring, automate, timer, file watcher.
user-invocable: true
---

# macOS Scheduler

Schedule recurring background tasks that run Claude prompts using macOS launchd. Supports interval scheduling, calendar schedules (cron-like), and file-watch triggers.

## Setup

```bash
SCHEDULER_PY="$(find ~/.claude/plugins/cache -path '*/fe-macos-scheduler/*/skills/macos-scheduler/resources/scheduler_manager.py' | head -1)"
```

Make the task wrapper executable (one-time):

```bash
WRAPPER_SH="$(find ~/.claude/plugins/cache -path '*/fe-macos-scheduler/*/skills/macos-scheduler/resources/job_wrapper.sh' | head -1)"
chmod +x "$WRAPPER_SH"
```

## Creating a Task — Plan Mode (REQUIRED)

When the user wants to create a scheduled task, you MUST follow the gather → plan → approve → execute flow. Never skip the plan.

### Step 1: Gather Requirements

Ask the user targeted questions to determine:

1. **What should this task do?** — The Claude prompt to execute each run
2. **What schedule type?**
   - **Interval** — run every N minutes/hours (e.g., "every 30 minutes")
   - **Calendar** — cron-like schedule (e.g., "weekdays at 9am", "1st of every month")
   - **File watcher** — trigger when specific file paths change
3. **What tools/skills will this task need?** — Identify which tools must be pre-authorized so the task can run non-interactively. Common examples:
   - Slack messaging: `mcp__slack__slack_write_api_call`
   - Google Docs: `mcp__clasp-enhanced__*` or `Bash(gcloud:*)` `Bash(curl:*)`
   - Databricks: `Bash(databricks:*)`
   - File operations: `Read` `Write` `Edit` `Bash`
   - Skills: `Skill(google-docs)` `Skill(salesforce-actions)` etc.
4. **Should it run immediately when created?** — `RunAtLoad` (default: no; auto-enabled for Slack tasks)
5. **Any advanced options?** — working directory, process priority, environment variables, throttle interval
6. **Which Slack channel(s) will this target?** (if Slack tools are needed) — Ask explicitly which channels the task will post to. Use `--allowed-channels` to restrict the task to only those channels.

> **Slack Safety Warning:** Never schedule tasks to post to channels with 50+ members (e.g. #general, #random, #all-company) without explicit user confirmation. Scheduled tasks run unattended — a misconfigured prompt can spam hundreds of people.

### Step 2: Present a Task Plan

Always present this plan and wait for explicit approval before creating anything:

```
## Scheduled Task Plan

**Name:** <human-readable name>
**Prompt:** "<the Claude prompt>"
**Schedule:** <type> — <human-readable description>
**Run at load:** Yes/No
**Process type:** Background (default) | Standard | Adaptive | Interactive

**Allowed tools:** <list of tools that will be pre-authorized>
- These tools will run without interactive permission prompts
- Tasks without the correct tools will fail silently or ask for approval (which hangs in non-interactive mode)

**Technical details:**
- Label: com.vibe.scheduler.<id>
- Plist: ~/Library/LaunchAgents/com.vibe.scheduler.<id>.plist
- Logs: ~/.vibe/scheduler/logs/<id>.log
- Runner: isaac (falls back to llm agent if unavailable)

<If Slack tools are detected, add:>
**Slack Safety:**
- Target channels: <list of channels from user>
- Channel verification: User confirmed channels are private/small
- Dry-run: REQUIRED (will test before scheduling)
- Run at load: Yes (auto-enabled for Slack tasks)
- Allowed channels flag: --allowed-channels <channels>

<If advanced options are set:>
**Advanced options:**
- Working directory: <path>
- Nice value: <n>
- Throttle interval: <n> seconds
- Environment: KEY=VALUE, ...
```

Present options to the user:
- **"looks good"** → create the task (for Slack tasks, a dry-run test runs automatically first)
- **"test first"** → run the prompt once with `isaac start --print "<prompt>"` to verify it works
- **"skip dry-run"** → (advanced) skip the automatic dry-run with `--no-dry-run`
- Or request changes → update plan and re-present

> **Note:** For tasks with Slack tools, dry-run is **required by default**. The scheduler will execute the prompt once and show the output before creating the schedule. Pass `--no-dry-run` only if you're confident the prompt is safe.

### Step 3: Execute

Only after user approval, run the create command.

### Schedule Type Reference

**Interval** — run every N minutes:
```bash
python3 "$SCHEDULER_PY" create \
  --name "Task Name" \
  --prompt "Your Claude prompt" \
  --schedule-type interval \
  --interval 15 \
  --allowed-tools Bash Read Write
```

**Calendar** — cron-like schedules using launchd `StartCalendarInterval` keys:

| Key | Range | Description |
|-----|-------|-------------|
| `--minute` | 0-59 | Minute of the hour |
| `--hour` | 0-23 | Hour of the day |
| `--day` | 1-31 | Day of the month |
| `--weekday` | 0-7 | Day of week (0 and 7 = Sunday) |
| `--month` | 1-12 | Month of the year |

Missing keys act as wildcards. Common patterns:

- **Every day at 9am:** `--schedule-type calendar --hour 9 --minute 0`
- **Weekdays at 9am:** `--schedule-type calendar --hour 9 --minute 0 --weekday 1,2,3,4,5`
- **1st of each month:** `--schedule-type calendar --day 1 --hour 0 --minute 0`
- **Every hour on the half:** `--schedule-type calendar --minute 30`

```bash
python3 "$SCHEDULER_PY" create \
  --name "Daily standup reminder" \
  --prompt "Send a message to #team-standup saying 'Time for standup!'" \
  --schedule-type calendar \
  --hour 9 --minute 0 --weekday 1,2,3,4,5 \
  --allowed-tools mcp__slack__slack_write_api_call mcp__slack__slack_read_api_call
```

Note: For weekday schedules, launchd creates one `StartCalendarInterval` entry per weekday. Unlike cron, launchd fires on wake if intervals were missed during sleep.

**File watcher** — trigger when paths change:
```bash
python3 "$SCHEDULER_PY" create \
  --name "Config change handler" \
  --prompt "Read /tmp/config.yaml and validate it" \
  --schedule-type watch-paths \
  --watch-paths /tmp/config.yaml /tmp/settings.json
```

### Advanced Options (all schedule types)

```bash
python3 "$SCHEDULER_PY" create \
  --name "Task Name" \
  --prompt "Your prompt" \
  --schedule-type interval \
  --interval 30 \
  --run-at-load \
  --process-type Background \
  --working-dir /some/path \
  --nice 10 \
  --env KEY1=VALUE1 --env KEY2=VALUE2 \
  --throttle-interval 30
```

| Option | Description |
|--------|-------------|
| `--run-at-load` | Run immediately when the task is loaded (default: off; auto-enabled for Slack tasks) |
| `--process-type` | One of: `Background` (default), `Standard`, `Adaptive`, `Interactive` |
| `--working-dir` | Set the working directory for the task |
| `--nice` | Process priority adjustment (higher = lower priority) |
| `--env KEY=VALUE` | Extra environment variables (repeatable) |
| `--throttle-interval` | Minimum seconds between consecutive runs |
| `--allowed-tools` | Tools to pre-authorize so the task runs without permission prompts (space-separated) |
| `--allowed-channels` | Slack channels this task may post to (space-separated, e.g. `#team-standup #my-alerts`). Enforced at runtime. |
| `--dry-run` | Test the prompt once before creating the schedule |
| `--no-dry-run` | Skip the automatic dry-run for Slack tasks (advanced) |

## Other Operations

These execute directly without a plan.

### List All Tasks

```bash
python3 "$SCHEDULER_PY" list
```

### Get Task Status

```bash
python3 "$SCHEDULER_PY" status --id <task_id>
```

### View Task Logs

```bash
python3 "$SCHEDULER_PY" logs --id <task_id> --lines 50
```

### Run a Task Immediately

Trigger a task right now via `launchctl kickstart`:

```bash
python3 "$SCHEDULER_PY" run-now --id <task_id>
```

### Print launchd Service Info

Show full launchd service details for a task:

```bash
python3 "$SCHEDULER_PY" print --id <task_id>
```

### Remove a Task

Completely remove a task — unloads from launchd, deletes the plist, archives logs:

```bash
python3 "$SCHEDULER_PY" remove --id <task_id>
```

### Disable / Enable a Task

```bash
python3 "$SCHEDULER_PY" disable --id <task_id>
python3 "$SCHEDULER_PY" enable --id <task_id>
```

## How It Works

1. Each task is a **launchd user agent** with a plist in `~/Library/LaunchAgents/`
2. Plists are generated programmatically using Python's `plistlib` (no XML templates)
3. launchd invokes `job_wrapper.sh` which runs `isaac start --print "<prompt>"` (falls back to `llm agent` if isaac is unavailable)
4. Pre-authorized tools are passed via `--allowedTools` so tasks can use MCP tools and skills non-interactively
5. Task registry is stored in `~/.vibe/scheduler/jobs.json`
6. All output is logged to `~/.vibe/scheduler/logs/<task_id>.log`

## Important Notes

- **macOS only** — uses launchd, which is not available on Linux
- **Minimum interval is 5 minutes** — to prevent excessive resource usage
- **Tasks run as your user** — they have the same permissions as your login session
- **Tasks persist across restarts** — launchd automatically manages agent lifecycle
- **No credentials stored** — tasks rely on your existing auth sessions (Slack MCP, Google OAuth, etc.)
- **Calendar schedules fire on wake** — if your Mac was asleep when a scheduled time passed, launchd fires the task when you wake up

## Troubleshooting

If a task isn't running:

1. Check if it's loaded: `launchctl list | grep com.vibe.scheduler`
2. Check logs: `python3 "$SCHEDULER_PY" logs --id <task_id>`
3. Print launchd info: `python3 "$SCHEDULER_PY" print --id <task_id>`
4. Verify claude is accessible: `which claude`
5. Try disabling and re-enabling: `python3 "$SCHEDULER_PY" disable --id <task_id>` then `enable`
6. Force-run to test: `python3 "$SCHEDULER_PY" run-now --id <task_id>`
