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

```bash
git clone https://github.com/wmsimpson/claude-vibe.git
cd claude-vibe
./setup.sh
```

This installs the `vibe` command and launches the interactive installer. From then on, use `vibe` directly from anywhere.

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
| 4 | Install dev tools (go, jq, yq, rg, node, gh, mmdc, graphviz) | 2 min |
| 5 | Install Databricks AI Dev Kit (50+ MCP tools, 34 skills) | 3 min |
| 6 | Install Claude Code plugins from bundled collection | 1 min |
| 7 | Configure MCP integrations (GitHub, Chrome DevTools, optional others) | 2 min |
| 8 | Run full validation suite | 30 sec |

## Bundled Plugins

All plugins ship with this repo — no external repos or git clones required.

| Plugin | Description |
|--------|-------------|
| `fe-google-tools` | Gmail, Docs, Sheets, Slides, Calendar, Forms, Tasks |
| `fe-app-dev` | React Native, Expo, Next.js, Swift, Flutter — scaffold, deploy, debug |
| `fe-workflows` | Architecture diagrams, RCA, POC docs, security questionnaires, sizing |
| `fe-vibe-setup` | Environment setup, validation, usage stats, integrations |
| `fe-specialized-agents` | Lucid Chart diagrams, Graphviz, web dev testing |
| `fe-databricks-tools` | Databricks queries, deployments, workspace management, demos |
| `fe-jira-tools` | Search, create, view, comment on JIRA tickets |
| `fe-macos-scheduler` | Schedule recurring launchd tasks on macOS |
| `lean-sigma-tools` | FMEA risk tables, SIPOC diagrams, swimlane process maps |
| `fe-mcp-servers` | MCP server framework (future-ready) |
| `fe-shared-resources` | Shared Python utilities and configs across plugins |

During setup, you can install all plugins or select specific ones interactively.

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
    fe-google-tools/      Google Workspace skills + agents
    fe-app-dev/           App development skills
    fe-workflows/         Workflow automation skills + agents
    fe-vibe-setup/        Setup and diagnostics skills + agents
    fe-specialized-agents/  Diagram and testing agents
    fe-databricks-tools/  Databricks integration skills + agents
    fe-jira-tools/        JIRA skills + agents
    fe-macos-scheduler/   macOS scheduler skills
    lean-sigma-tools/     Lean Six Sigma skills
    fe-mcp-servers/       MCP server framework
    fe-shared-resources/  Shared utilities
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
