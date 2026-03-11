# Claude Vibe

A guided interactive setup for a fully-equipped local AI development environment using Claude Code on macOS.

One command gets you:
- **Claude Code** with AVX-aware install (npm or Homebrew)
- **Google Workspace** access (Gmail, Docs, Sheets, Slides, Calendar, Forms, Tasks)
- **Databricks AI Dev Kit** (50+ MCP tools, 34 skills)
- **11 plugin collections** — Google tools, app dev, workflows, JIRA, Lean Six Sigma, and more (bundled in this repo)
- **MCP Integrations** (GitHub, Chrome DevTools, Slack, Brave Search)
- **Profile management** — switch between identities (personal/work) instantly
- **Full validation** to confirm everything works

## Quick Start

Make sure [Homebrew](https://brew.sh) is installed first, then:

```bash
git clone https://github.com/wmsimpson/claude-vibe.git
cd claude-vibe
./setup.sh
```

This installs the `vibe` command and launches the interactive installer. You'll be prompted about optional integrations (Databricks, Slack, etc.) so only the tools you need get installed. From then on, use `vibe` directly from anywhere.

## Usage

```bash
vibe install              # Full guided setup
vibe install --step 3     # Run a specific step
vibe install --resume     # Resume from where you left off
vibe validate             # Run full validation suite
vibe status               # Show setup progress
vibe doctor               # Diagnose common issues
vibe reset                # Reset install state
vibe agent                # Launch Claude Code with active profile
vibe version              # Show version
vibe help                 # Show all commands

# Profile management
vibe profile              # Show active profile and list all
vibe profile save personal    # Save current config as "personal"
vibe profile create work      # Create empty "work" profile
vibe profile switch work      # Switch to "work" profile
vibe profile delete work      # Delete a profile
```

## What It Does

| Step | What | Time |
|------|------|------|
| 1 | Install Claude Code (AVX detection, npm/Homebrew) | 1 min |
| 2 | Google OAuth client setup (browser-guided) | 3 min |
| 3 | Enable Google APIs and set quota project | 1 min |
| 4 | Install dev tools (go, jq, yq, rg, node, gh, mmdc, graphviz, gcloud) | 2 min |
| 5 | Install Databricks AI Dev Kit *(optional — skipped if Databricks not enabled)* | 3 min |
| 6 | Install Claude Code plugins from bundled collection | 1 min |
| 7 | Configure MCP integrations (GitHub, Chrome DevTools, optional others) | 2 min |
| 8 | Run full validation suite | 30 sec |

## Bundled Plugins

All plugins ship with this repo — no external repos or git clones required.

| Plugin | Description |
|--------|-------------|
| `google-tools` | Gmail, Docs, Sheets, Slides, Calendar, Forms, Tasks |
| `app-dev` | React Native, Expo, Next.js, Swift, Flutter — scaffold, deploy, debug |
| `workflows` | Architecture diagrams, RCA, POC docs, security questionnaires, sizing |
| `vibe-setup` | Environment setup, validation, usage stats, integrations |
| `specialized-agents` | Lucid Chart diagrams, Graphviz, web dev testing |
| `databricks-tools` | Databricks queries, deployments, workspace management, demos |
| `jira-tools` | Search, create, view, comment on JIRA tickets |
| `macos-scheduler` | Schedule recurring launchd tasks on macOS |
| `lean-sigma-tools` | FMEA risk tables, SIPOC diagrams, swimlane process maps |
| `mcp-servers` | MCP server framework (future-ready) |
| `shared-resources` | Shared Python utilities and configs across plugins |

During setup, you can install all plugins or select specific ones interactively.

## Prerequisites

Before running `./setup.sh`, make sure you have the following:

### Required

| Prerequisite | Why | Install |
|---|---|---|
| **macOS** (Intel or Apple Silicon) | This tool is macOS-only (uses Homebrew, launchd, zsh) | — |
| **Homebrew** | Installs all CLI tools (node, go, jq, gcloud, etc.) | `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"` |
| **Git** | Clone the repo; also used by GitHub MCP integration | `brew install git` (or Xcode CLT: `xcode-select --install`) |
| **Claude subscription** | Claude Code requires an active Anthropic plan | [claude.ai/pricing](https://claude.ai/pricing) — Pro ($20/mo) or Max ($100/mo recommended) |

### Recommended (installed automatically during setup if missing)

These are installed by Step 4, but having them ahead of time speeds things up:

| Tool | Purpose |
|---|---|
| **Node.js** | Required for Claude Code (npm install) and mermaid-cli |
| **Python 3** | Used by setup scripts for JSON/YAML processing and plugin permissions sync |
| **Google Cloud SDK (`gcloud`)** | Google OAuth and API access — `brew install --cask google-cloud-sdk` |

### Optional accounts (prompted during setup)

| Account | What it enables |
|---|---|
| **Google account** (free) | Gmail, Docs, Sheets, Slides, Calendar, Forms, Tasks APIs |
| **GitHub account** | GitHub MCP integration — repo management, PRs, issues |
| **Databricks workspace** | Databricks CLI, AI Dev Kit, query/deploy tools (prompted — skipped if not needed) |
| **JIRA instance** | Ticket search, creation, and management |
| **Slack workspace** | Slack MCP integration (optional in Step 7) |

### Verify prerequisites

```bash
# Check that Homebrew is installed
brew --version

# Check git
git --version

# (Optional) Check Python 3
python3 --version
```

If Homebrew is not installed, run:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

## After Setup

Restart Claude Code, then try:

```
"Create a Google Doc titled Hello World"
"List my GitHub repos"
"Search my Gmail for unread messages"
"Create an architecture diagram for a React app with a Python backend"
```

## Profiles

Profiles let you maintain separate identity contexts (different Google accounts, API keys, MCP configs) and switch between them instantly.

Each profile stores:
- `claude.json` — MCP server configs (tokens, endpoints)
- `env` — Environment variables (API keys, tokens)
- `gcloud-adc.json` — Google Application Default Credentials
- `gcloud-oauth.json` — Google OAuth client credentials
- `gcp-project-id` — GCP quota project

Shared across all profiles (not swapped): plugins, skills, installed CLI tools.

```bash
# Save your current setup
vibe profile save personal

# Create and set up a second identity
vibe profile create work
vibe profile switch work
vibe install    # Run setup with new Google account, keys, etc.
vibe profile save work

# Switch back
vibe profile switch personal
vibe agent    # Launch Claude with personal profile loaded
```

## Project Structure

```
claude-vibe/
  setup.sh                Bootstrap — installs vibe command and runs installer
  bin/vibe                CLI entry point (symlinked to ~/.local/bin/vibe)
  lib/
    tty.sh                TTY UI helpers (colors, prompts, spinners)
    profiles.sh           Profile management (save, switch, create, delete)
  steps/
    01-install-claude.sh
    02-google-oauth.sh
    03-enable-apis.sh
    04-install-tools.sh
    05-ai-dev-kit.sh
    06-install-plugins.sh
    07-configure-mcp.sh
    08-validate.sh
  plugins/                Bundled plugin collections
    google-tools/      Google Workspace skills + agents
    app-dev/           App development skills
    workflows/         Workflow automation skills + agents
    vibe-setup/        Setup and diagnostics skills + agents
    specialized-agents/  Diagram and testing agents
    databricks-tools/  Databricks integration skills + agents
    jira-tools/        JIRA skills + agents
    macos-scheduler/   macOS scheduler skills
    lean-sigma-tools/     Lean Six Sigma skills
    mcp-servers/       MCP server framework
    shared-resources/  Shared utilities
  .claude-plugin/         Plugin manifests (plugin.json, marketplace.json)
  permissions.yaml        Master skill permissions config
  mcp-servers.yaml        MCP server configs
  config/
    env.template          Template for ~/.vibe/env
```

## What Gets Installed Where

| What | Where |
|------|-------|
| vibe CLI | `~/.local/bin/vibe` (symlink to repo) |
| Claude Code binary | `/usr/local/bin/claude` |
| OAuth credentials | `~/.config/gcloud/credentials/<profile>-google-auth.json` |
| Google ADC | `~/.config/gcloud/application_default_credentials.json` |
| AI Dev Kit | `~/.ai-dev-kit/` |
| Claude skills | `~/.claude/skills/` |
| Plugin cache | `~/.claude/plugins/cache/claude-vibe/` |
| Plugin settings | `~/.claude/settings.json` |
| MCP server config | `~/.claude.json` |
| Environment vars | `~/.vibe/env` |
| Install state | `~/.claude-vibe/install-state.json` |
| Profiles | `~/.claude-vibe/profiles/<name>/` |

## Troubleshooting

Run `vibe doctor` for automated diagnostics.

**"This app is blocked" during Google OAuth**
- You need your own OAuth client ID (Step 2 walks you through creating one)

**Claude shows "CPU lacks AVX support" warning**
- The installer detects this and uses npm instead of Homebrew (no Bun runtime)

**Google APIs return 403**
- Run `vibe install --step 3` to enable APIs and set the quota project

**Plugins not loading after install**
- Restart Claude Code — plugins load at startup

**Plugin cache shows `individual-vibe-tool`**
- Old installs cached under the previous name. Run `vibe install --step 6` to reinstall under `claude-vibe`

## License

MIT
