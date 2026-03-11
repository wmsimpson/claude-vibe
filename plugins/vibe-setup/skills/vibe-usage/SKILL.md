---
name: vibe-usage
description: >
  Show Claude Code token usage, spending, and budget for the current month.
  Set a monthly spending limit. Use when the user asks how much they have spent,
  how many tokens they have used, what their quota is, or wants to change their budget.
  Triggers on: "how much have I spent", "usage", "quota", "budget", "tokens used",
  "set my budget", "change my limit", "how much left".
---

# Vibe Usage Tracker

Shows monthly Claude Code spending vs. your configured budget, with token breakdown.

## Quick Commands

```bash
# Show this month's usage
python3 ~/.local/share/vibe/track_usage.py show

# Set monthly budget
python3 ~/.local/share/vibe/track_usage.py set-budget 75

# Reset this month's counter (e.g. if you want a fresh start)
python3 ~/.local/share/vibe/track_usage.py reset
```

## How It Works

**Automatic tracking:** A Stop hook runs `track_usage.py track` at the end of every
Claude Code session, reading the session cost from `~/.claude.json` and appending it
to `~/.vibe/usage.json`.

**Budget:** Soft limit stored in `~/.vibe/usage.json`. No hard enforcement — it's a
personal gauge. For a hard server-side limit, use an Anthropic API key and set a
monthly spend cap at https://console.anthropic.com/settings/limits.

**Data format:** `~/.vibe/usage.json` stores per-session records with date, cost,
and token counts. All data is local — never sent anywhere.

## Setup (first-time)

`configure-vibe` sets this up automatically. To set it up manually:

### 1 — Install the tracking script

```bash
mkdir -p ~/.local/share/vibe
cp ~/.claude/plugins/cache/claude-vibe/vibe-setup/*/skills/vibe-usage/resources/track_usage.py \
   ~/.local/share/vibe/track_usage.py
```

### 2 — Add the Stop hook to ~/.claude/settings.json

Add this to the `hooks.Stop` array in your `~/.claude/settings.json`:

```json
{
  "matcher": "",
  "hooks": [{
    "type": "command",
    "command": "python3 ~/.local/share/vibe/track_usage.py track 2>/dev/null || true",
    "timeout": 10
  }]
}
```

### 3 — Set your monthly budget

```bash
python3 ~/.local/share/vibe/track_usage.py set-budget 50
```

## Example Output

```
════════════════════════════════════════════════════
  Claude Usage  ·  February 2026
════════════════════════════════════════════════════
  🟡  [████████████████████░░░░░░░░]  72.4%

  Spent this month:     $72.40
  Monthly budget:      $100.00
  Remaining:            $27.60
  Projected (full mo):  $98.40
  Days remaining:            3
────────────────────────────────────────────────────
  Sessions logged:          14
  Input tokens:        623,041
  Output tokens:       187,204
  Cache reads:       4,821,903
  Cache writes:      1,203,441
────────────────────────────────────────────────────
  Budget: set with 'vibe usage set-budget <amount>'
  Data:   ~/.vibe/usage.json  ·  47 sessions total
════════════════════════════════════════════════════
```

## Choosing a Budget

| Usage pattern | Suggested budget |
|---------------|-----------------|
| Light — occasional sessions | $20–40/month |
| Regular — daily use, moderate sessions | $50–80/month |
| Heavy — long multi-skill workflows | $100–150/month |
| Power — parallel agents, all-day sessions | $200+/month |

## Anthropic API Key vs. Subscription

For **hard** quota enforcement (prevents spending over your limit):

1. Get an API key: https://console.anthropic.com/settings/api-keys
2. Set a monthly limit: https://console.anthropic.com/settings/limits
3. Configure Claude Code to use it:
   ```bash
   echo 'export ANTHROPIC_API_KEY=sk-ant-...' >> ~/.zprofile
   # Restart Claude Code — it will use the API key instead of your subscription
   ```

With an API key, once you hit your console limit Claude Code will stop responding
until the next billing cycle or you raise the limit.

## Instructions for Claude

When this skill is invoked:

1. Run the show command to display current usage:
   ```bash
   python3 ~/.local/share/vibe/track_usage.py show
   ```

2. If the user wants to change their budget, run set-budget:
   ```bash
   python3 ~/.local/share/vibe/track_usage.py set-budget <amount>
   ```
   Then run show again to confirm the new gauge.

3. If the tracking script is not installed at `~/.local/share/vibe/track_usage.py`,
   install it first from the plugin cache:
   ```bash
   mkdir -p ~/.local/share/vibe
   CACHE=$(ls -d ~/.claude/plugins/cache/claude-vibe/vibe-setup/*/skills/vibe-usage/resources/ 2>/dev/null | head -1)
   cp "$CACHE/track_usage.py" ~/.local/share/vibe/track_usage.py
   ```

4. If the Stop hook is not configured, show the user the setup instructions above and
   ask if they want to add it to `~/.claude/settings.json` now.
