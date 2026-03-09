---
name: configure-vibe
description: Configure, setup, and validate the vibe environment including dependencies, directories, and permissions. Run this on first install or when setting up on a new machine.
---

# Configure Vibe

Complete environment setup for claude-vibe on a personal macOS machine. Designed to be fully self-sufficient — no IT support required.

## CRITICAL: Do Not Give Up Early

If one step fails, **continue with the rest**. Track failures and report them all at the end. A partial setup is better than stopping early.

---

## Step 1 — Chrome + chrome-devtools MCP (DO THIS FIRST)

**Why first:** chrome-devtools is your primary debugging tool. If any external integration setup fails, you'll use Chrome to diagnose it visually. Set this up before anything else.

### 1a. Install Google Chrome
Check: `ls /Applications/Google\ Chrome.app 2>/dev/null && echo "Chrome installed" || echo "NOT installed"`

If not installed:
```bash
brew install --cask google-chrome </dev/null
```
Or download from https://www.google.com/chrome/

### 1b. Verify Node.js is available (needed for chrome-devtools MCP)
```bash
node --version && npx --version
```
If not installed, jump to Step 3 (Node.js) first, then return here.

### 1c. Verify chrome-devtools MCP is configured
```bash
cat ~/.claude.json | python3 -m json.tool | grep -A 5 "chrome-devtools"
```
Expected output: a block with `"command": "npx"` and `chrome-devtools-mcp@latest`.

If missing, add it manually:
```bash
python3 << 'EOF'
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f:
    cfg = json.load(f)
cfg.setdefault("mcpServers", {})["chrome-devtools"] = {
    "type": "stdio",
    "command": "npx",
    "args": ["chrome-devtools-mcp@latest", "--userDataDir=" + os.path.expanduser("~/.vibe/chrome/profile")],
    "env": {}
}
with open(path, "w") as f:
    json.dump(cfg, f, indent=2)
print("chrome-devtools added to ~/.claude.json")
print("RESTART Claude Code for this to take effect.")
EOF
```

### 1d. Create Chrome profile directory
```bash
mkdir -p ~/.vibe/chrome/profile
```

### 1e. Test chrome-devtools is working
After restarting Claude Code (if config was just added), use the chrome-devtools MCP tools to open a test page. If `mcp__chrome-devtools__new_page` is available, that means chrome-devtools is active.

If the MCP tools are not available after restart:
```bash
# Verify npx can find the package
npx chrome-devtools-mcp@latest --help 2>&1 | head -5
```

---

## Step 2 — Homebrew

Check: `brew --version`

Install if not present:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

After installation on Apple Silicon, add to PATH:
```bash
eval "$(/opt/homebrew/bin/brew shellenv)"
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
```

---

## Step 3 — Node.js

Check: `node --version`

If not installed (recommended: use nvm for version management):
```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
nvm install --lts && nvm use --lts
```

Or simpler: `brew install node </dev/null`

Install global packages:
```bash
npm install -g @mermaid-js/mermaid-cli
```

Verify: `node --version && npm --version && mmdc --version`

---

## Step 4 — Core CLI Tools

```bash
brew install jq </dev/null
brew install yq </dev/null
brew install ripgrep </dev/null
brew install terminal-notifier </dev/null
brew install graphviz </dev/null
brew install uv </dev/null
```

Verify: `jq --version && yq --version && rg --version && dot -V`

---

## Step 5 — Google Account Sign-In (Personal Gmail)

**This signs your personal Google account into the AI tool.** It opens a browser window where you log in with your Gmail/Google Workspace email. No GCP billing account or IT admin required for this step.

### 5a. Install gcloud CLI
Check: `gcloud version`

Install if not present:
```bash
brew install --cask google-cloud-sdk </dev/null
```

Source the PATH (Apple Silicon):
```bash
source "/opt/homebrew/share/google-cloud-sdk/path.bash.inc"
echo 'source "/opt/homebrew/share/google-cloud-sdk/path.bash.inc"' >> ~/.zprofile
```

Intel Mac path: `/usr/local/share/google-cloud-sdk/path.bash.inc`

### 5b. Sign in with your personal Google account
Run this command — it will open your browser:
```bash
gcloud auth application-default login \
  --scopes="https://www.googleapis.com/auth/drive,https://www.googleapis.com/auth/cloud-platform,https://www.googleapis.com/auth/documents,https://www.googleapis.com/auth/presentations,https://www.googleapis.com/auth/spreadsheets,https://www.googleapis.com/auth/calendar,https://www.googleapis.com/auth/gmail.modify,https://www.googleapis.com/auth/forms.body,https://www.googleapis.com/auth/forms.responses.readonly,https://www.googleapis.com/auth/tasks"
```

**In the browser that opens:**
1. Select your personal Google account (Gmail address)
2. Click "Continue" on the "Google Auth Library wants to access your Google Account" screen
3. Check the boxes to allow access (or click "Select All") — these grant the AI access to Docs, Sheets, Gmail, Calendar, etc.
4. Click "Continue" / "Allow"
5. The page will show "You are now authenticated" — the terminal will continue

**CRITICAL:** You MUST complete the browser flow. If the browser doesn't open, run:
```bash
gcloud auth application-default login --no-browser
```
This prints a URL you can paste into any browser.

### 5c. Verify sign-in worked
```bash
gcloud auth application-default print-access-token
```
If this prints a long token string starting with `ya29.`, sign-in succeeded.

If it prints an error, re-run the login command in step 5b.

---

## Step 6 — Enable Google Sheets, Slides, and Forms (Free GCP Project)

**Why this is needed:** Google Docs and Gmail work immediately after sign-in. However, Sheets, Slides, and Forms APIs require a "quota project" — a free Google Cloud project linked to your account that tracks API usage.

**This takes about 5 minutes and is completely free.**

### 6a. Create a free GCP project
1. Open https://console.cloud.google.com/ in your browser
2. Sign in with the SAME Google account you used in Step 5
3. Click the project dropdown at the top (next to "Google Cloud")
4. Click **"New Project"**
5. Name it anything (e.g., `personal-ai-tools`)
6. Click **"Create"** — wait ~30 seconds
7. Copy the **Project ID** (shown below the project name — looks like `personal-ai-tools-123456`)

### 6b. Enable the required APIs
Still in Google Cloud Console with your new project selected:
1. Click the hamburger menu → **"APIs & Services"** → **"Library"**
2. Search for and **Enable** each of these:
   - Google Sheets API
   - Google Slides API
   - Google Drive API
   - Google Docs API
   - Google Forms API
   *(Calendar and Gmail are enabled automatically via OAuth scopes)*

### 6c. Set the quota project
```bash
# Replace YOUR_PROJECT_ID with the ID from step 6a (e.g., personal-ai-tools-123456)
gcloud auth application-default set-quota-project YOUR_PROJECT_ID

# Add to your environment so it persists across sessions
echo 'export GCP_QUOTA_PROJECT=YOUR_PROJECT_ID' >> ~/.vibe/env
source ~/.vibe/env
```

### 6d. Verify Sheets works
```bash
TOKEN=$(gcloud auth application-default print-access-token)
RESULT=$(curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: $GCP_QUOTA_PROJECT" \
  -H "Content-Type: application/json" \
  -d '{"properties":{"title":"API Test - delete me"}}')
echo "$RESULT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
if 'spreadsheetId' in d:
    print('✅ Sheets API working! Sheet ID:', d['spreadsheetId'])
else:
    print('❌ Error:', d.get('error', {}).get('message', 'unknown'))
    print('   Check that APIs are enabled and GCP_QUOTA_PROJECT is set correctly.')
"
```

### Troubleshooting Sheets/Slides/Forms setup
If the test fails, use chrome-devtools to debug:
```
Use the chrome-devtools MCP to open https://console.cloud.google.com/
Take a screenshot to see if you're signed in to the right project
Navigate to APIs & Services → Library and verify the APIs are enabled
```

---

## Step 7 — Go (for vibe CLI)

Check: `go version`

Install if not present: `brew install go </dev/null`

Install the vibe CLI:
```bash
cd ~/code/claude-vibe
./setup.sh
```

This symlinks `vibe` to `~/.local/bin/` and adds it to your PATH.

Verify: `vibe version`

---

## Step 8 — GitHub CLI

Check: `gh --version`

Install: `brew install gh </dev/null`

Authenticate with your personal GitHub account:
```bash
gh auth login
# Select: GitHub.com → HTTPS → Authenticate with your browser
```

Complete the browser flow — it will ask you to paste a code into github.com/login/device.

Verify: `gh auth status`

---

## Step 9 — Directories and Environment File

```bash
mkdir -p ~/.vibe ~/.vibe/chrome/profile ~/.local/bin ~/code
```

Create `~/.vibe/env` if it doesn't exist:
```bash
[ -f ~/.vibe/env ] || cat > ~/.vibe/env << 'EOF'
# claude-vibe environment
export PATH="$HOME/.local/bin:$PATH"

# Google Sheets/Slides/Forms quota project
# export GCP_QUOTA_PROJECT=your-project-id  # Set during configure-vibe Step 6

# Your email address (used in telemetry + app dev scaffolding)
# export VIBE_USER_EMAIL=you@example.com

# Optional integrations — set by /setup-integrations
# export GITHUB_PERSONAL_ACCESS_TOKEN=ghp_...
# export SLACK_BOT_TOKEN=xoxb-...
# export SLACK_TEAM_ID=T0...
# export JIRA_URL=https://your-org.atlassian.net
# export JIRA_USERNAME=you@example.com
# export JIRA_API_TOKEN=...
# export NOTION_API_TOKEN=secret_...
# export LINEAR_API_KEY=lin_api_...
# export BRAVE_API_KEY=BSA-...
# export SUPABASE_ACCESS_TOKEN=sbp_...

# App dev defaults (optional)
# export MOBILE_PLATFORM=expo          # expo, react-native, flutter, swift
# export WEB_DEPLOY_TARGET=vercel      # vercel, netlify, firebase
EOF
```

Auto-source the env file from your shell profile:
```bash
grep -q 'source ~/.vibe/env' ~/.zprofile || \
  echo 'source ~/.vibe/env 2>/dev/null || true' >> ~/.zprofile
source ~/.vibe/env
```

---

## Step 10 — Databricks CLI (optional)

Only needed for `fe-databricks-tools` skills.

Check: `databricks --version`

Install:
```bash
brew tap databricks/tap
brew install databricks </dev/null
```

Configure a workspace profile:
```bash
databricks configure --profile myworkspace
# Prompts for: workspace URL + personal access token
# Workspace URL example: https://your-workspace.cloud.databricks.com
# Access token: create at workspace → Settings → Developer → Access tokens
```

Verify: `databricks auth profiles`

---

## Step 11 — Mobile / Web Dev Tools (optional)

Ask which tools are needed and install only what applies:

**React Native / Expo:**
```bash
npm install -g expo-cli
xcode-select --install   # Xcode CLI tools
brew install cocoapods </dev/null
```

**Swift / iOS** — install Xcode from Mac App Store (free), then:
```bash
sudo xcode-select --switch /Applications/Xcode.app/Contents/Developer
```

**Flutter:**
```bash
brew install --cask flutter </dev/null
flutter doctor
```

**Web deployment CLIs:**
```bash
npm install -g vercel netlify-cli firebase-tools
```

---

## Step 12 — User Profile

Check for existing profile:
```bash
cat ~/.vibe/profile 2>/dev/null | head -20
```

If no profile exists, ask: *"Would you like to build your vibe profile now?"*

If yes: spawn the `vibe-profile` agent. It gathers your name, GitHub username, active projects, and preferences from what's available locally.

---

## Step 13 — Usage Tracking and Budget

Install the usage tracker and configure a monthly budget. This logs every session's
cost to `~/.vibe/usage.json` and warns you when approaching your limit.

```bash
# Install tracking script
mkdir -p ~/.local/share/vibe
CACHE=$(ls -d ~/.claude/plugins/cache/claude-vibe/fe-vibe-setup/*/skills/vibe-usage/resources/ 2>/dev/null | head -1)
cp "$CACHE/track_usage.py" ~/.local/share/vibe/track_usage.py
echo "✓ Usage tracker installed at ~/.local/share/vibe/track_usage.py"
```

Ask the user: *"What monthly budget would you like for Claude usage?"*
- Suggest: `$50` light use · `$100` daily use · `$200` heavy/power use

```bash
python3 ~/.local/share/vibe/track_usage.py set-budget <AMOUNT>
```

Add the Stop hook to `~/.claude/settings.json` so every session is tracked automatically.
Read the existing file with the Read tool, then use Edit to add to `hooks.Stop`:

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

Show the initial dashboard to confirm setup:
```bash
python3 ~/.local/share/vibe/track_usage.py show
```

Tell the user:
> *"Run `vibe usage` anytime to check spending and tokens. Change your budget with `vibe usage set-budget <amount>`. For a hard server-side limit, set a monthly cap at https://console.anthropic.com/settings/limits (requires API key)."*

---

## Step 14 — Optional Integrations

Ask the user: *"Would you like to connect any additional services? All options are free-tier or open source."*

If yes, spawn the `setup-integrations` skill. Available integrations:
- **GitHub** — repos, PRs, issues, Actions (free, needs PAT)
- **Slack** — messages and channels (free tier, needs bot token)
- **JIRA + Confluence** — issue tracking + wiki (free tier, needs Atlassian API token)
- **Notion** — notes and databases (free tier, needs integration token)
- **Linear** — engineering issue tracking (free tier, needs API key)
- **Brave Search** — real-time web search, 1000 queries/month free
- **Supabase** — Postgres + auth backend for apps (free tier)
- **Memory** — local AI knowledge graph (no account needed)
- **Fetch** — web page fetcher/parser (no account needed)

The skill will guide through account creation and token setup for each selected integration.

---

## Step 15 — Final Verification

Run the `validate-mcp-access` skill to do a live check of all configured tools and connections. It will:
- Open Chrome via chrome-devtools MCP and take a screenshot to confirm visual debugging works
- Test Google API access for each service
- Check local directory permissions and installed MCPs
- Report anything that needs attention with specific remediation steps

---

## Troubleshooting Reference

| Problem | Fix |
|---------|-----|
| `vibe` not found | `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zprofile && source ~/.zprofile` |
| gcloud not found | `source "/opt/homebrew/share/google-cloud-sdk/path.bash.inc"` |
| Google sign-in: browser doesn't open | `gcloud auth application-default login --no-browser` (paste URL manually) |
| Sheets/Forms/Slides return 403 | Complete Step 6 — create free GCP project, set quota project |
| chrome-devtools MCP tools not available | Restart Claude Code after adding to `~/.claude.json` |
| Databricks CLI won't connect | Verify URL format: `https://host.databricks.com` (no trailing slash) |
| `npm install -g` fails with permissions | Use nvm instead of system Node.js |
