#!/usr/bin/env python3
"""
scheduler_manager.py — Manage scheduled Claude Code jobs via macOS launchd.

Supports interval, calendar, and file-watch scheduling with full launchd plist options.

Usage:
    python3 scheduler_manager.py create --name "Job Name" --prompt "Your prompt" --schedule-type interval --interval 15
    python3 scheduler_manager.py create --name "Job Name" --prompt "Your prompt" --schedule-type calendar --hour 9 --minute 0 --weekday 1,2,3,4,5
    python3 scheduler_manager.py create --name "Job Name" --prompt "Your prompt" --schedule-type watch-paths --watch-paths /path/one /path/two
    python3 scheduler_manager.py list
    python3 scheduler_manager.py status --id <job_id>
    python3 scheduler_manager.py logs --id <job_id> [--lines 50]
    python3 scheduler_manager.py run-now --id <job_id>
    python3 scheduler_manager.py print --id <job_id>
    python3 scheduler_manager.py remove --id <job_id>
    python3 scheduler_manager.py enable --id <job_id>
    python3 scheduler_manager.py disable --id <job_id>
"""

import argparse
import json
import os
import platform
import plistlib
import shutil
import subprocess
import sys
import uuid
from datetime import datetime
from pathlib import Path

# Constants
SCHEDULER_DIR = Path.home() / ".vibe" / "scheduler"
JOBS_FILE = SCHEDULER_DIR / "jobs.json"
LOG_DIR = SCHEDULER_DIR / "logs"
ARCHIVE_DIR = SCHEDULER_DIR / "archive"
LAUNCH_AGENTS_DIR = Path.home() / "Library" / "LaunchAgents"
LABEL_PREFIX = "com.vibe.scheduler"
MIN_INTERVAL_MINUTES = 1
VALID_PROCESS_TYPES = ("Background", "Standard", "Adaptive", "Interactive")

# Paths to resources (same directory as this script)
SCRIPT_DIR = Path(__file__).parent
JOB_WRAPPER = SCRIPT_DIR / "job_wrapper.sh"


def ensure_dirs():
    """Create required directories."""
    SCHEDULER_DIR.mkdir(parents=True, exist_ok=True)
    LOG_DIR.mkdir(parents=True, exist_ok=True)
    ARCHIVE_DIR.mkdir(parents=True, exist_ok=True)
    LAUNCH_AGENTS_DIR.mkdir(parents=True, exist_ok=True)


def load_jobs():
    """Load the job registry."""
    if not JOBS_FILE.exists():
        return []
    with open(JOBS_FILE) as f:
        return json.load(f)


def save_jobs(jobs):
    """Save the job registry."""
    ensure_dirs()
    with open(JOBS_FILE, "w") as f:
        json.dump(jobs, f, indent=2)


def get_label(job_id):
    """Get the launchd label for a job."""
    return f"{LABEL_PREFIX}.{job_id}"


def get_plist_path(job_id):
    """Get the plist file path for a job."""
    return LAUNCH_AGENTS_DIR / f"{get_label(job_id)}.plist"


def get_log_path(job_id):
    """Get the log file path for a job."""
    return LOG_DIR / f"{job_id}.log"


def generate_plist(job_id, config):
    """Generate a launchd plist programmatically using plistlib."""
    env_vars = {
        "PATH": f"{Path.home()}/.local/bin:/usr/local/bin:/usr/bin:/bin:/opt/homebrew/bin",
        "HOME": str(Path.home()),
    }
    env_vars.update(config.get("env", {}))

    plist = {
        "Label": get_label(job_id),
        "ProgramArguments": ["/bin/bash", str(JOB_WRAPPER), job_id],
        "RunAtLoad": config.get("run_at_load", False),
        "KeepAlive": False,
        "EnvironmentVariables": env_vars,
        "StandardOutPath": str(get_log_path(job_id)),
        "StandardErrorPath": str(get_log_path(job_id)),
    }

    # Schedule type
    schedule_type = config["schedule_type"]
    if schedule_type == "interval":
        plist["StartInterval"] = config["interval_minutes"] * 60
    elif schedule_type == "calendar":
        schedule_config = config.get("schedule_config", {})
        # Weekday can be a list — create one entry per weekday
        weekdays = schedule_config.get("weekday")
        if isinstance(weekdays, list) and len(weekdays) > 1:
            entries = []
            for wd in weekdays:
                entry = {}
                if schedule_config.get("minute") is not None:
                    entry["Minute"] = schedule_config["minute"]
                if schedule_config.get("hour") is not None:
                    entry["Hour"] = schedule_config["hour"]
                if schedule_config.get("day") is not None:
                    entry["Day"] = schedule_config["day"]
                if schedule_config.get("month") is not None:
                    entry["Month"] = schedule_config["month"]
                entry["Weekday"] = wd
                entries.append(entry)
            plist["StartCalendarInterval"] = entries
        else:
            cal = {}
            if schedule_config.get("minute") is not None:
                cal["Minute"] = schedule_config["minute"]
            if schedule_config.get("hour") is not None:
                cal["Hour"] = schedule_config["hour"]
            if schedule_config.get("day") is not None:
                cal["Day"] = schedule_config["day"]
            if schedule_config.get("month") is not None:
                cal["Month"] = schedule_config["month"]
            if weekdays is not None:
                # Single weekday (int or list of one)
                cal["Weekday"] = weekdays[0] if isinstance(weekdays, list) else weekdays
            plist["StartCalendarInterval"] = cal
    elif schedule_type == "watch-paths":
        plist["WatchPaths"] = config["watch_paths"]

    # Optional keys
    if config.get("process_type"):
        plist["ProcessType"] = config["process_type"]
    if config.get("working_dir"):
        plist["WorkingDirectory"] = config["working_dir"]
    if config.get("nice") is not None:
        plist["Nice"] = config["nice"]
    if config.get("throttle_interval"):
        plist["ThrottleInterval"] = config["throttle_interval"]

    return plistlib.dumps(plist, fmt=plistlib.FMT_XML)


def launchctl_bootstrap(plist_path):
    """Load a plist into launchd."""
    uid = os.getuid()
    domain = f"gui/{uid}"
    result = subprocess.run(
        ["launchctl", "bootstrap", domain, str(plist_path)],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        # Fall back to legacy load
        result = subprocess.run(
            ["launchctl", "load", str(plist_path)],
            capture_output=True, text=True
        )
    return result


def launchctl_bootout(job_id):
    """Remove a job from launchd."""
    uid = os.getuid()
    domain = f"gui/{uid}"
    label = get_label(job_id)
    result = subprocess.run(
        ["launchctl", "bootout", f"{domain}/{label}"],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        # Fall back to legacy unload
        plist_path = get_plist_path(job_id)
        if plist_path.exists():
            result = subprocess.run(
                ["launchctl", "unload", str(plist_path)],
                capture_output=True, text=True
            )
    return result


def launchctl_list_job(job_id):
    """Query launchd for job status."""
    label = get_label(job_id)
    result = subprocess.run(
        ["launchctl", "list", label],
        capture_output=True, text=True
    )
    if result.returncode == 0:
        return result.stdout.strip()
    return None


def launchctl_kickstart(job_id):
    """Immediately run a job via launchctl kickstart."""
    uid = os.getuid()
    label = get_label(job_id)
    return subprocess.run(
        ["launchctl", "kickstart", "-p", f"gui/{uid}/{label}"],
        capture_output=True, text=True
    )


def launchctl_print(job_id):
    """Get full service info from launchd."""
    uid = os.getuid()
    label = get_label(job_id)
    return subprocess.run(
        ["launchctl", "print", f"gui/{uid}/{label}"],
        capture_output=True, text=True
    )


def parse_weekday_arg(value):
    """Parse weekday argument: '1,2,3,4,5' or '1-5' or '3' into a list of ints."""
    if "-" in value and "," not in value:
        parts = value.split("-")
        start, end = int(parts[0]), int(parts[1])
        return list(range(start, end + 1))
    return [int(x) for x in value.split(",")]


def parse_env_arg(values):
    """Parse --env KEY=VALUE arguments into a dict."""
    env = {}
    if values:
        for item in values:
            if "=" not in item:
                print(f"ERROR: Invalid env format '{item}'. Use KEY=VALUE.")
                sys.exit(1)
            key, val = item.split("=", 1)
            env[key] = val
    return env


def format_schedule_description(config):
    """Return a human-readable description of the schedule."""
    schedule_type = config["schedule_type"]
    if schedule_type == "interval":
        minutes = config["interval_minutes"]
        if minutes >= 60 and minutes % 60 == 0:
            return f"every {minutes // 60} hour(s)"
        return f"every {minutes} minutes"
    elif schedule_type == "calendar":
        sc = config.get("schedule_config", {})
        parts = []
        if sc.get("weekday"):
            day_names = {0: "Sun", 1: "Mon", 2: "Tue", 3: "Wed", 4: "Thu", 5: "Fri", 6: "Sat", 7: "Sun"}
            wds = sc["weekday"] if isinstance(sc["weekday"], list) else [sc["weekday"]]
            parts.append(", ".join(day_names.get(d, str(d)) for d in wds))
        if sc.get("day"):
            parts.append(f"day {sc['day']} of month")
        if sc.get("month"):
            parts.append(f"month {sc['month']}")
        time_str = ""
        if sc.get("hour") is not None and sc.get("minute") is not None:
            time_str = f" at {sc['hour']:02d}:{sc['minute']:02d}"
        elif sc.get("hour") is not None:
            time_str = f" at hour {sc['hour']}"
        elif sc.get("minute") is not None:
            time_str = f" at minute {sc['minute']}"
        desc = " ".join(parts) + time_str if parts else f"calendar{time_str}"
        return desc
    elif schedule_type == "watch-paths":
        paths = config.get("watch_paths", [])
        return f"when {', '.join(paths)} change(s)"
    return schedule_type


def has_slack_tools(allowed_tools):
    """Check if any of the allowed tools are Slack-related."""
    return any('slack' in t.lower() for t in (allowed_tools or []))


SLACK_SAFETY_PREFIX = """IMPORTANT SAFETY RULE: Before sending any Slack message, verify the target channel.
NEVER post to #general, #random, #all-company, or any channel with more than 50 members
unless the prompt below explicitly names that channel. If unsure about channel size,
use conversations.info to check member_count first.

USER PROMPT:
"""


def create_job(args):
    """Create and register a new scheduled job."""
    if platform.system() != "Darwin":
        print("ERROR: This scheduler only works on macOS (uses launchd).")
        sys.exit(1)

    ensure_dirs()

    # Build config from args
    config = {
        "schedule_type": args.schedule_type,
        "run_at_load": args.run_at_load,
        "process_type": args.process_type,
        "working_dir": args.working_dir,
        "nice": args.nice,
        "env": parse_env_arg(args.env),
        "throttle_interval": args.throttle_interval,
    }

    # Schedule-type-specific validation and config
    if args.schedule_type == "interval":
        if not args.interval:
            print("ERROR: --interval is required for schedule-type 'interval'.")
            sys.exit(1)
        if args.interval < MIN_INTERVAL_MINUTES:
            print(f"ERROR: Minimum interval is {MIN_INTERVAL_MINUTES} minutes.")
            sys.exit(1)
        config["interval_minutes"] = args.interval
    elif args.schedule_type == "calendar":
        schedule_config = {}
        if args.minute is not None:
            schedule_config["minute"] = args.minute
        if args.hour is not None:
            schedule_config["hour"] = args.hour
        if args.day is not None:
            schedule_config["day"] = args.day
        if args.weekday is not None:
            schedule_config["weekday"] = parse_weekday_arg(args.weekday)
        if args.month is not None:
            schedule_config["month"] = args.month
        if not schedule_config:
            print("ERROR: Calendar schedule requires at least one of: --minute, --hour, --day, --weekday, --month.")
            sys.exit(1)
        config["schedule_config"] = schedule_config
    elif args.schedule_type == "watch-paths":
        if not args.watch_paths:
            print("ERROR: --watch-paths is required for schedule-type 'watch-paths'.")
            sys.exit(1)
        config["watch_paths"] = args.watch_paths

    # Validate process type
    if args.process_type and args.process_type not in VALID_PROCESS_TYPES:
        print(f"ERROR: Invalid process type '{args.process_type}'. Must be one of: {', '.join(VALID_PROCESS_TYPES)}")
        sys.exit(1)

    job_id = str(uuid.uuid4())[:8]

    # Slack safety guardrails
    slack_detected = has_slack_tools(args.allowed_tools)

    if slack_detected:
        # Safety prefix injection — prepend to prompt so the LLM sees it first
        args.prompt = SLACK_SAFETY_PREFIX + args.prompt

        # If allowed-channels specified, add channel restriction to prompt
        if args.allowed_channels:
            channels_str = ", ".join(args.allowed_channels)
            args.prompt = f"CHANNEL RESTRICTION: You may ONLY post to these Slack channels: {channels_str}. Do not post to any other channel.\n\n" + args.prompt

        # Auto-enable run-at-load for Slack tasks so user can verify first run
        if not args.run_at_load:
            args.run_at_load = True
            config["run_at_load"] = True
            print("INFO: Auto-enabled --run-at-load for Slack task (so you can verify the first run immediately).")

    # Dry-run: test the prompt before creating the schedule
    should_dry_run = args.dry_run or (slack_detected and not args.no_dry_run)
    if should_dry_run:
        print("\n--- DRY RUN: Testing prompt before scheduling ---")
        print(f"Prompt: {args.prompt[:200]}{'...' if len(args.prompt) > 200 else ''}\n")

        # Find a runner to execute the dry run
        dry_run_bin = shutil.which("isaac") or shutil.which("dbexec") or shutil.which("llm")
        if dry_run_bin:
            runner_name = Path(dry_run_bin).name
            if runner_name in ("isaac", "dbexec"):
                dry_cmd = [dry_run_bin]
                if runner_name == "dbexec":
                    dry_cmd += ["repo", "run", "isaac", "--"]
                dry_cmd += ["start", "--print", "--no-session-persistence", "--dangerously-skip-permissions"]
                if args.allowed_tools:
                    dry_cmd += ["--allowedTools", ",".join(args.allowed_tools)]
            else:
                dry_cmd = [dry_run_bin, "agent"]
                if args.allowed_tools:
                    dry_cmd += ["--allowedTools", ",".join(args.allowed_tools)]

            try:
                result = subprocess.run(
                    dry_cmd, input=args.prompt, capture_output=True, text=True, timeout=120
                )
                print("--- DRY RUN OUTPUT ---")
                if result.stdout:
                    print(result.stdout[:2000])
                if result.stderr:
                    print(f"STDERR: {result.stderr[:500]}")
                print(f"--- END DRY RUN (exit code: {result.returncode}) ---\n")
            except subprocess.TimeoutExpired:
                print("--- DRY RUN TIMED OUT (120s) ---\n")
            except Exception as e:
                print(f"--- DRY RUN FAILED: {e} ---\n")
        else:
            print("WARNING: No runner (isaac/llm) found — skipping dry run execution.")

        # Prompt for confirmation
        try:
            answer = input("Proceed with creating the scheduled task? [y/N] ").strip().lower()
        except EOFError:
            answer = "n"
        if answer not in ("y", "yes"):
            print("Aborted. Task was NOT created.")
            sys.exit(0)

    # Generate and write plist
    plist_content = generate_plist(job_id, config)
    plist_path = get_plist_path(job_id)
    with open(plist_path, "wb") as f:
        f.write(plist_content)

    # Register in jobs file
    jobs = load_jobs()
    job_entry = {
        "id": job_id,
        "name": args.name,
        "prompt": args.prompt,
        "schedule_type": args.schedule_type,
        "schedule_config": config.get("schedule_config", {}),
        "interval_minutes": config.get("interval_minutes"),
        "watch_paths": config.get("watch_paths"),
        "run_at_load": config.get("run_at_load", False),
        "process_type": config.get("process_type"),
        "working_dir": config.get("working_dir"),
        "nice": config.get("nice"),
        "env": config.get("env", {}),
        "throttle_interval": config.get("throttle_interval"),
        "allowed_tools": args.allowed_tools or [],
        "allowed_channels": args.allowed_channels or [],
        "enabled": True,
        "created_at": datetime.now().isoformat(),
        "label": get_label(job_id),
        "plist_path": str(plist_path),
    }
    jobs.append(job_entry)
    save_jobs(jobs)

    # Load into launchd
    result = launchctl_bootstrap(plist_path)
    if result.returncode != 0 and result.stderr:
        print(f"WARNING: launchctl returned: {result.stderr.strip()}")

    schedule_desc = format_schedule_description(config)
    print(f"Job created successfully!")
    print(f"  ID:       {job_id}")
    print(f"  Name:     {args.name}")
    print(f"  Schedule: {args.schedule_type} — {schedule_desc}")
    if args.allowed_tools:
        print(f"  Tools:    {', '.join(args.allowed_tools)}")
    if args.allowed_channels:
        print(f"  Channels: {', '.join(args.allowed_channels)}")
    print(f"  Label:    {get_label(job_id)}")
    print(f"  Plist:    {plist_path}")
    print(f"  Logs:     {get_log_path(job_id)}")


def list_jobs():
    """List all registered jobs with runtime status."""
    jobs = load_jobs()
    if not jobs:
        print("No scheduled jobs found.")
        return

    print(f"{'ID':<10} {'Name':<30} {'Type':<12} {'Schedule':<25} {'Enabled':<9} {'Loaded':<8}")
    print("-" * 100)

    for job in jobs:
        loaded = "yes" if launchctl_list_job(job["id"]) else "no"
        enabled = "yes" if job.get("enabled", True) else "no"
        stype = job.get("schedule_type", "interval")
        schedule_desc = format_schedule_description(job)
        # Truncate long descriptions
        if len(schedule_desc) > 23:
            schedule_desc = schedule_desc[:20] + "..."
        print(f"{job['id']:<10} {job['name']:<30} {stype:<12} {schedule_desc:<25} {enabled:<9} {loaded:<8}")


def get_status(job_id):
    """Get detailed status for a specific job."""
    jobs = load_jobs()
    job = next((j for j in jobs if j["id"] == job_id), None)
    if not job:
        print(f"ERROR: Job '{job_id}' not found.")
        sys.exit(1)

    launchd_info = launchctl_list_job(job_id)
    log_path = get_log_path(job_id)
    log_size = log_path.stat().st_size if log_path.exists() else 0

    schedule_desc = format_schedule_description(job)

    print(f"Job: {job['name']} ({job_id})")
    print(f"  Prompt:     {job['prompt']}")
    print(f"  Schedule:   {job.get('schedule_type', 'interval')} — {schedule_desc}")
    print(f"  Enabled:    {'yes' if job.get('enabled', True) else 'no'}")
    print(f"  Run at load: {'yes' if job.get('run_at_load', False) else 'no'}")
    if job.get("process_type"):
        print(f"  Process type: {job['process_type']}")
    if job.get("working_dir"):
        print(f"  Working dir: {job['working_dir']}")
    if job.get("nice") is not None:
        print(f"  Nice:       {job['nice']}")
    if job.get("throttle_interval"):
        print(f"  Throttle:   {job['throttle_interval']}s")
    if job.get("env"):
        print(f"  Env vars:   {job['env']}")
    if job.get("watch_paths"):
        print(f"  Watch paths: {', '.join(job['watch_paths'])}")
    if job.get("allowed_tools"):
        print(f"  Allowed tools: {', '.join(job['allowed_tools'])}")
    if job.get("allowed_channels"):
        print(f"  Allowed channels: {', '.join(job['allowed_channels'])}")
    print(f"  Created:    {job['created_at']}")
    print(f"  Label:      {job['label']}")
    print(f"  Plist:      {job['plist_path']}")
    print(f"  Log file:   {log_path} ({log_size} bytes)")
    print(f"  Loaded:     {'yes' if launchd_info else 'no'}")
    if launchd_info:
        print(f"  launchctl:  {launchd_info}")


def get_logs(job_id, lines=50):
    """Display recent log output for a job."""
    log_path = get_log_path(job_id)
    if not log_path.exists():
        print(f"No logs found for job '{job_id}'.")
        return

    with open(log_path) as f:
        all_lines = f.readlines()

    recent = all_lines[-lines:]
    for line in recent:
        print(line, end="")


def run_now(job_id):
    """Immediately trigger a job via launchctl kickstart."""
    jobs = load_jobs()
    job = next((j for j in jobs if j["id"] == job_id), None)
    if not job:
        print(f"ERROR: Job '{job_id}' not found.")
        sys.exit(1)

    if not job.get("enabled", True):
        print(f"ERROR: Job '{job_id}' is disabled. Enable it first.")
        sys.exit(1)

    # Check if loaded
    if not launchctl_list_job(job_id):
        print(f"ERROR: Job '{job_id}' is not loaded in launchd. Try enabling it first.")
        sys.exit(1)

    result = launchctl_kickstart(job_id)
    if result.returncode == 0:
        print(f"Job '{job_id}' ({job['name']}) kicked off successfully.")
        print(f"  Check logs: python3 scheduler_manager.py logs --id {job_id}")
    else:
        print(f"ERROR: Failed to kickstart job '{job_id}'.")
        if result.stderr:
            print(f"  {result.stderr.strip()}")


def print_service(job_id):
    """Show full launchd service info for a job."""
    jobs = load_jobs()
    job = next((j for j in jobs if j["id"] == job_id), None)
    if not job:
        print(f"ERROR: Job '{job_id}' not found.")
        sys.exit(1)

    result = launchctl_print(job_id)
    if result.returncode == 0:
        print(result.stdout)
    else:
        print(f"ERROR: Could not get service info for '{job_id}'.")
        if result.stderr:
            print(f"  {result.stderr.strip()}")
        print(f"  (Job may not be loaded. Check with: launchctl list | grep {get_label(job_id)})")


def remove_job(job_id):
    """Remove a job: bootout from launchd, delete plist, remove from registry."""
    jobs = load_jobs()
    job = next((j for j in jobs if j["id"] == job_id), None)
    if not job:
        print(f"ERROR: Job '{job_id}' not found.")
        sys.exit(1)

    # Bootout from launchd
    launchctl_bootout(job_id)

    # Delete plist
    plist_path = get_plist_path(job_id)
    if plist_path.exists():
        plist_path.unlink()

    # Archive log
    log_path = get_log_path(job_id)
    if log_path.exists():
        archive_path = ARCHIVE_DIR / f"{job_id}_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log"
        shutil.move(str(log_path), str(archive_path))
        print(f"  Log archived to: {archive_path}")

    # Remove from registry
    jobs = [j for j in jobs if j["id"] != job_id]
    save_jobs(jobs)

    print(f"Job '{job_id}' ({job['name']}) removed successfully.")


def enable_job(job_id):
    """Enable a disabled job."""
    jobs = load_jobs()
    job = next((j for j in jobs if j["id"] == job_id), None)
    if not job:
        print(f"ERROR: Job '{job_id}' not found.")
        sys.exit(1)

    if job.get("enabled", True):
        print(f"Job '{job_id}' is already enabled.")
        return

    # Load plist
    plist_path = get_plist_path(job_id)
    if not plist_path.exists():
        print(f"ERROR: Plist not found at {plist_path}. Try recreating the job.")
        sys.exit(1)

    result = launchctl_bootstrap(plist_path)
    if result.returncode != 0 and result.stderr:
        print(f"WARNING: launchctl returned: {result.stderr.strip()}")

    job["enabled"] = True
    save_jobs(jobs)
    print(f"Job '{job_id}' ({job['name']}) enabled.")


def disable_job(job_id):
    """Disable a job without removing it."""
    jobs = load_jobs()
    job = next((j for j in jobs if j["id"] == job_id), None)
    if not job:
        print(f"ERROR: Job '{job_id}' not found.")
        sys.exit(1)

    if not job.get("enabled", True):
        print(f"Job '{job_id}' is already disabled.")
        return

    launchctl_bootout(job_id)

    job["enabled"] = False
    save_jobs(jobs)
    print(f"Job '{job_id}' ({job['name']}) disabled.")


def main():
    parser = argparse.ArgumentParser(description="Manage scheduled Claude Code jobs via macOS launchd")
    subparsers = parser.add_subparsers(dest="command", required=True)

    # create
    p_create = subparsers.add_parser("create", help="Create a new scheduled job")
    p_create.add_argument("--name", required=True, help="Human-readable job name")
    p_create.add_argument("--prompt", required=True, help="Claude prompt to execute")
    p_create.add_argument("--schedule-type", required=True, choices=["interval", "calendar", "watch-paths"],
                          help="Schedule type: interval, calendar, or watch-paths")
    # Interval options
    p_create.add_argument("--interval", type=int, help="Interval in minutes (for schedule-type 'interval', minimum 5)")
    # Calendar options
    p_create.add_argument("--minute", type=int, help="Minute (0-59) for calendar schedule")
    p_create.add_argument("--hour", type=int, help="Hour (0-23) for calendar schedule")
    p_create.add_argument("--day", type=int, help="Day of month (1-31) for calendar schedule")
    p_create.add_argument("--weekday", type=str, help="Day(s) of week (0-7, comma-separated or range) for calendar schedule")
    p_create.add_argument("--month", type=int, help="Month (1-12) for calendar schedule")
    # Watch-paths options
    p_create.add_argument("--watch-paths", nargs="+", help="File paths to watch (for schedule-type 'watch-paths')")
    # Advanced options
    p_create.add_argument("--run-at-load", action="store_true", default=False,
                          help="Run the job immediately when loaded")
    p_create.add_argument("--process-type", type=str, default=None,
                          help="Process type: Background, Standard, Adaptive, or Interactive")
    p_create.add_argument("--working-dir", type=str, default=None, help="Working directory for the job")
    p_create.add_argument("--nice", type=int, default=None, help="Nice value (process priority adjustment)")
    p_create.add_argument("--env", action="append", default=None, help="Environment variable as KEY=VALUE (repeatable)")
    p_create.add_argument("--throttle-interval", type=int, default=None,
                          help="Minimum seconds between consecutive runs")
    # Tool permissions
    p_create.add_argument("--allowed-tools", nargs="+", default=None,
                          help="Tools to pre-authorize for this job (e.g. 'Bash(git:*)' 'mcp__slack__*' 'Skill(google-docs)')")
    # Slack safety options
    p_create.add_argument("--allowed-channels", nargs="+", default=None,
                          help="Slack channels this task may post to (e.g. '#team-standup' '#my-alerts')")
    p_create.add_argument("--dry-run", action="store_true", default=False,
                          help="Test the prompt once before creating the schedule")
    p_create.add_argument("--no-dry-run", action="store_true", default=False,
                          help="Skip automatic dry-run for Slack tasks")

    # list
    subparsers.add_parser("list", help="List all scheduled jobs")

    # status
    p_status = subparsers.add_parser("status", help="Get detailed status of a job")
    p_status.add_argument("--id", required=True, help="Job ID")

    # logs
    p_logs = subparsers.add_parser("logs", help="View job logs")
    p_logs.add_argument("--id", required=True, help="Job ID")
    p_logs.add_argument("--lines", type=int, default=50, help="Number of lines to show (default: 50)")

    # run-now
    p_run_now = subparsers.add_parser("run-now", help="Immediately trigger a job")
    p_run_now.add_argument("--id", required=True, help="Job ID")

    # print
    p_print = subparsers.add_parser("print", help="Show launchd service info for a job")
    p_print.add_argument("--id", required=True, help="Job ID")

    # remove
    p_remove = subparsers.add_parser("remove", help="Remove a job")
    p_remove.add_argument("--id", required=True, help="Job ID")

    # enable
    p_enable = subparsers.add_parser("enable", help="Enable a disabled job")
    p_enable.add_argument("--id", required=True, help="Job ID")

    # disable
    p_disable = subparsers.add_parser("disable", help="Disable a job without removing it")
    p_disable.add_argument("--id", required=True, help="Job ID")

    args = parser.parse_args()

    if args.command == "create":
        create_job(args)
    elif args.command == "list":
        list_jobs()
    elif args.command == "status":
        get_status(args.id)
    elif args.command == "logs":
        get_logs(args.id, args.lines)
    elif args.command == "run-now":
        run_now(args.id)
    elif args.command == "print":
        print_service(args.id)
    elif args.command == "remove":
        remove_job(args.id)
    elif args.command == "enable":
        enable_job(args.id)
    elif args.command == "disable":
        disable_job(args.id)


if __name__ == "__main__":
    main()
