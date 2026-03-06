# Claude Vibe

A guided interactive setup for a fully-equipped local AI development environment using Claude Code on macOS.

One command gets you:
- **Claude Code** with AVX-aware install (npm or Homebrew)
- **Google Workspace** access (Gmail, Docs, Sheets, Slides, Calendar, Forms, Tasks)
- **Databricks AI Dev Kit** (50+ MCP tools, 34 skills)
- **Claude Code Plugins** (Google tools, app dev, workflows, Lean Six Sigma, and more)
- **MCP Integrations** (GitHub, Chrome DevTools, Slack, JIRA, Notion, Brave Search)
- **Full validation** to confirm everything works

## Quick Start

```bash
git clone https://github.com/wsimpsonjr/claude-vibe.git
cd claude-vibe
./setup.sh
```

The installer walks you through each step interactively. No prior setup needed beyond a Mac and a Claude subscription.

## What It Does

| Step | What | Time |
|------|------|------|
| 1 | Install Claude Code (AVX detection, npm/Homebrew) | 1 min |
| 2 | Google OAuth client setup (browser-guided) | 3 min |
| 3 | Enable Google APIs and set quota project | 1 min |
| 4 | Install dev tools (go, jq, yq, rg, node, gh, mmdc, graphviz) | 2 min |
| 5 | Install Databricks AI Dev Kit (50+ MCP tools, 34 skills) | 3 min |
| 6 | Install Claude Code plugins (Google, app dev, workflows) | 1 min |
| 7 | Configure MCP integrations (GitHub, Chrome DevTools, optional others) | 2 min |
| 8 | Run full validation suite | 30 sec |

## Usage

```bash
./setup.sh                # Full guided setup
./setup.sh --step 3       # Run a specific step
./setup.sh --resume       # Resume from where you left off
./setup.sh --validate     # Run validation only
./setup.sh --reset        # Start over
```

## Prerequisites

- **macOS** (Intel or Apple Silicon)
- **Claude subscription** (Pro $20/mo, Max $100/mo recommended)
- **Google account** (for Workspace APIs — free)
- **Homebrew** (installed automatically if missing)

## After Setup

Restart Claude Code, then try:

```
"Create a Google Doc titled Hello World"
"List my GitHub repos"
"Search my Gmail for unread messages"
"Create an architecture diagram for a React app with a Python backend"
```

## Project Structure

```
claude-vibe/
  setup.sh              Main interactive installer
  lib/tty.sh            TTY UI helpers (colors, prompts, spinners)
  steps/
    01-install-claude.sh
    02-google-oauth.sh
    03-enable-apis.sh
    04-install-tools.sh
    05-ai-dev-kit.sh
    06-install-plugins.sh
    07-configure-mcp.sh
    08-validate.sh
  config/
    env.template         Template for ~/.vibe/env
  README.md
```

## What Gets Installed Where

| What | Where |
|------|-------|
| Claude Code binary | `/usr/local/bin/claude` |
| OAuth credentials | `~/.config/gcloud/credentials/claude-google-auth.json` |
| Google ADC | `~/.config/gcloud/application_default_credentials.json` |
| AI Dev Kit | `~/.ai-dev-kit/` |
| Claude skills | `~/.claude/skills/` |
| Plugin cache | `~/.claude/plugins/cache/` |
| Plugin settings | `~/.claude/settings.json` |
| MCP server config | `~/.claude.json` |
| Environment vars | `~/.vibe/env` |
| Install state | `~/.claude-vibe/install-state.json` |

## Troubleshooting

**"This app is blocked" during Google OAuth**
- You need your own OAuth client ID (Step 2 walks you through creating one)

**Claude shows "CPU lacks AVX support" warning**
- The installer detects this and uses npm instead of Homebrew (no Bun runtime)

**Google APIs return 403**
- Run `./setup.sh --step 3` to enable APIs and set the quota project

**Plugins not loading after install**
- Restart Claude Code — plugins load at startup

## License

MIT
