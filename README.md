# Claude Vibe

A guided interactive setup for a fully-equipped local AI development environment using Claude Code.

One command gets you:
- **Claude Code** with AVX-aware install (npm or Homebrew)
- **Google Workspace** access (Gmail, Docs, Sheets, Slides, Calendar, Forms, Tasks)
- **Databricks AI Dev Kit** (50+ MCP tools, 34 skills)
- **11 plugin collections** — Google tools, app dev, workflows, JIRA, Lean Six Sigma, and more (bundled in this repo)
- **MCP Integrations** (GitHub, Chrome DevTools, Slack, Brave Search)
- **Profile management** — switch between identities (personal/work) instantly
- **Full validation** to confirm everything works

## Quick Start

### macOS

Open Terminal and run:

```bash
# 1. Install Homebrew (if you don't have it)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 2. Install git (if you don't have it)
brew install git

# 3. Clone and run
git clone https://github.com/wmsimpson/claude-vibe.git
cd claude-vibe
./bootstrap.sh
```

### Windows

Claude Vibe runs inside WSL (Windows Subsystem for Linux). Open PowerShell as Administrator and run:

```powershell
# 1. Install WSL with Ubuntu
wsl --install

# 2. Restart your computer, then open the Ubuntu terminal and run:
```

Then inside the Ubuntu terminal:

```bash
# 3. Install Homebrew for Linux
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
echo >> ~/.bashrc
echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> ~/.bashrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

# 4. Install git
brew install git

# 5. Clone and run
git clone https://github.com/wmsimpson/claude-vibe.git
cd claude-vibe
./bootstrap.sh
```

### What happens next

`bootstrap.sh` checks that your system has everything it needs (Homebrew, git, a Claude subscription), installs the `vibe` CLI, builds the Go binary if Go is available, and launches the interactive installer. The installer walks you through 8 steps and prompts about optional integrations (Databricks, Slack, etc.) so only the tools you need get installed.

From then on, use `vibe` from anywhere.

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
| `community-skills` | Humanizer, PPTX creator/editor, find-skills, web design guidelines, Spark data sources |
| `databricks-tools` | Databricks queries, deployments, workspace management, demos |
| `jira-tools` | Search, create, view, comment on JIRA tickets |
| `macos-scheduler` | Schedule recurring launchd tasks on macOS |
| `lean-sigma-tools` | FMEA risk tables, SIPOC diagrams, swimlane process maps |
| `mcp-servers` | MCP server framework (future-ready) |
| `shared-resources` | Shared Python utilities and configs across plugins |

Additionally, the **superpowers** plugin (from `claude-plugins-official`) is installed automatically. It provides brainstorming, writing-plans, test-driven-development, systematic-debugging, code review, and other development workflow skills.

During setup, you can install all plugins or select specific ones interactively.

## Prerequisites

`bootstrap.sh` checks all of these for you, but here's what you need:

### Required

| Prerequisite | Why | macOS | Windows (WSL) |
|---|---|---|---|
| **Homebrew** | Installs all CLI tools | [brew.sh](https://brew.sh) | Same, inside WSL Ubuntu |
| **Git** | Clone the repo | `brew install git` or `xcode-select --install` | `brew install git` |
| **Claude subscription** | Claude Code needs an Anthropic plan | [claude.ai/pricing](https://claude.ai/pricing) | Same |

### Installed automatically by `vibe install`

You don't need to install these yourself. Step 4 handles all of them:

| Tool | Purpose |
|---|---|
| **Node.js** | Required for Claude Code and mermaid-cli |
| **Go** | Builds the vibe Go CLI |
| **Python 3** | JSON/YAML processing and plugin sync |
| **Google Cloud SDK** | Google OAuth and API access |
| **jq, yq, ripgrep, gh** | Dev tooling used by skills and agents |

### Optional accounts (prompted during setup)

| Account | What it enables |
|---|---|
| **Google account** (free) | Gmail, Docs, Sheets, Slides, Calendar, Forms, Tasks |
| **GitHub account** | Repo management, PRs, issues via MCP |
| **Databricks workspace** | CLI, AI Dev Kit, query/deploy tools (skipped if not needed) |
| **JIRA instance** | Ticket search, creation, and management |
| **Slack workspace** | Slack MCP integration |

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
    community-skills/  Humanizer, PPTX, find-skills, web design, Spark data sources
    databricks-tools/  Databricks integration skills + agents
    jira-tools/        JIRA skills + agents
    macos-scheduler/   macOS scheduler skills
    lean-sigma-tools/  Lean Six Sigma skills
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

**Plugin cache shows old name**
- Old installs may be cached under a previous name. Run `vibe install --step 6` to reinstall under `claude-vibe`

## License

MIT
