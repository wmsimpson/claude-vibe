#!/usr/bin/env python3
"""
Track Claude Code usage per session for budget monitoring.

Usage:
  python3 track_usage.py show              # Show monthly usage and budget
  python3 track_usage.py track             # Record current session cost (called by Stop hook)
  python3 track_usage.py set-budget <amt>  # Set monthly budget in USD
  python3 track_usage.py reset             # Clear this month's session log
"""

import json
import os
import sys
from datetime import datetime, timezone
from pathlib import Path


USAGE_FILE = Path.home() / ".vibe" / "usage.json"
CLAUDE_JSON = Path.home() / ".claude.json"


# ─── Data helpers ────────────────────────────────────────────────────────────

def load_usage():
    if not USAGE_FILE.exists():
        return {"sessions": [], "budget": {"monthly_usd": 100.0}}
    with open(USAGE_FILE) as f:
        return json.load(f)


def save_usage(data):
    USAGE_FILE.parent.mkdir(parents=True, exist_ok=True)
    with open(USAGE_FILE, "w") as f:
        json.dump(data, f, indent=2)


def read_claude_session():
    """Read last session cost from ~/.claude.json."""
    if not CLAUDE_JSON.exists():
        return None

    with open(CLAUDE_JSON) as f:
        data = json.load(f)

    projects = data.get("projects", {})
    if not projects:
        return None

    cwd = os.getcwd()

    # Prefer current directory, fall back to home, then most expensive project
    project = (
        projects.get(cwd)
        or projects.get(str(Path.home()))
        or max(projects.values(), key=lambda p: p.get("lastCost", 0), default=None)
    )

    if not project or not project.get("lastCost"):
        return None

    return {
        "cost_usd": project.get("lastCost", 0),
        "input_tokens": project.get("lastTotalInputTokens", 0),
        "output_tokens": project.get("lastTotalOutputTokens", 0),
        "cache_read_tokens": project.get("lastTotalCacheReadInputTokens", 0),
        "cache_write_tokens": project.get("lastTotalCacheCreationInputTokens", 0),
        "session_id": project.get("lastSessionId", ""),
        "model_usage": project.get("lastModelUsage", {}),
        "project": cwd,
    }


# ─── Commands ────────────────────────────────────────────────────────────────

def cmd_track():
    """Record the completed session's cost (called by Stop hook)."""
    session = read_claude_session()
    if not session or session["cost_usd"] <= 0:
        return

    log = load_usage()
    sessions = log.setdefault("sessions", [])

    # Dedup: skip if we already logged this session_id
    sid = session["session_id"]
    if sid and any(s.get("session_id") == sid for s in sessions):
        return

    now = datetime.now(timezone.utc).isoformat()
    sessions.append({
        "date": now,
        "session_id": sid,
        "project": session["project"],
        "cost_usd": round(session["cost_usd"], 6),
        "input_tokens": session["input_tokens"],
        "output_tokens": session["output_tokens"],
        "cache_read_tokens": session["cache_read_tokens"],
        "cache_write_tokens": session["cache_write_tokens"],
    })

    save_usage(log)


def cmd_show():
    """Display monthly usage summary with budget gauge."""
    log = load_usage()
    sessions = log.get("sessions", [])
    budget = log.get("budget", {}).get("monthly_usd", 100.0)

    now = datetime.now(timezone.utc)
    month_key = now.strftime("%Y-%m")
    month_label = now.strftime("%B %Y")

    month_sessions = [s for s in sessions if s.get("date", "").startswith(month_key)]

    cost = sum(s.get("cost_usd", 0) for s in month_sessions)
    input_tok = sum(s.get("input_tokens", 0) for s in month_sessions)
    output_tok = sum(s.get("output_tokens", 0) for s in month_sessions)
    cache_read = sum(s.get("cache_read_tokens", 0) for s in month_sessions)
    cache_write = sum(s.get("cache_write_tokens", 0) for s in month_sessions)
    remaining = budget - cost
    pct = (cost / budget * 100) if budget > 0 else 0

    # Progress bar
    bar_w = 28
    filled = int(bar_w * min(pct / 100, 1.0))
    bar = "█" * filled + "░" * (bar_w - filled)
    icon = "🟢" if pct < 70 else "🟡" if pct < 90 else "🔴"

    # Days remaining in month
    import calendar
    days_in_month = calendar.monthrange(now.year, now.month)[1]
    days_remaining = days_in_month - now.day
    daily_avg = cost / now.day if now.day > 0 else 0
    projected = daily_avg * days_in_month

    print(f"\n{'═'*52}")
    print(f"  Claude Usage  ·  {month_label}")
    print(f"{'═'*52}")
    print(f"  {icon}  [{bar}]  {pct:.1f}%")
    print(f"")
    print(f"  Spent this month:   ${cost:>8.2f}")
    print(f"  Monthly budget:     ${budget:>8.2f}")
    print(f"  Remaining:          ${remaining:>8.2f}")
    print(f"  Projected (full mo): ${projected:>7.2f}")
    print(f"  Days remaining:     {days_remaining:>8}")
    print(f"{'─'*52}")
    print(f"  Sessions logged:    {len(month_sessions):>8}")
    print(f"  Input tokens:       {input_tok:>8,}")
    print(f"  Output tokens:      {output_tok:>8,}")
    print(f"  Cache reads:        {cache_read:>8,}")
    print(f"  Cache writes:       {cache_write:>8,}")
    print(f"{'─'*52}")
    print(f"  Budget: set with 'vibe usage set-budget <amount>'")
    print(f"  Data:   ~/.vibe/usage.json  ·  {len(sessions)} sessions total")
    print(f"{'═'*52}\n")

    if pct >= 90:
        print(f"  ⚠️  {pct:.0f}% of monthly budget used! ({days_remaining} days remaining)")
        print(f"  Run: vibe usage set-budget <new_amount>  to adjust\n")
    elif pct >= 70:
        print(f"  ⚡ {pct:.0f}% used — on track for ${projected:.0f} this month\n")


def cmd_set_budget(amount_str):
    """Set monthly budget in USD."""
    try:
        amount = float(amount_str)
    except ValueError:
        print(f"Error: '{amount_str}' is not a valid dollar amount", file=sys.stderr)
        sys.exit(1)

    log = load_usage()
    log.setdefault("budget", {})["monthly_usd"] = amount
    save_usage(log)

    current_month = datetime.now(timezone.utc).strftime("%Y-%m")
    month_sessions = [s for s in log.get("sessions", []) if s.get("date", "").startswith(current_month)]
    current_cost = sum(s.get("cost_usd", 0) for s in month_sessions)
    pct = (current_cost / amount * 100) if amount > 0 else 0

    print(f"✓ Monthly budget set to ${amount:.2f}")
    print(f"  Current spend: ${current_cost:.2f} ({pct:.1f}% of new budget)")


def cmd_reset():
    """Clear this month's session log."""
    log = load_usage()
    now = datetime.now(timezone.utc)
    month_key = now.strftime("%Y-%m")
    before = len(log.get("sessions", []))
    log["sessions"] = [s for s in log.get("sessions", []) if not s.get("date", "").startswith(month_key)]
    after = len(log["sessions"])
    save_usage(log)
    print(f"✓ Cleared {before - after} sessions from {now.strftime('%B %Y')}")


# ─── Main ────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    cmd = sys.argv[1] if len(sys.argv) > 1 else "show"

    if cmd == "track":
        cmd_track()
    elif cmd == "show":
        cmd_show()
    elif cmd == "set-budget":
        if len(sys.argv) < 3:
            print("Usage: track_usage.py set-budget <amount>", file=sys.stderr)
            sys.exit(1)
        cmd_set_budget(sys.argv[2])
    elif cmd == "reset":
        cmd_reset()
    else:
        print(__doc__)
        sys.exit(1)
