---
name: setup-integrations
description: Interactively set up optional MCP integrations (GitHub, Slack, JIRA, Notion, Brave Search, Supabase, Linear). Prompts for account creation if needed. All options are free-tier or open source. Run this after configure-vibe to add the integrations you want. Triggers on "set up integrations", "connect GitHub", "connect Slack", "add integrations", "configure MCP".
user-invocable: true
---

# Setup Integrations

Interactive wizard to connect optional services. All integrations use **open source MCP servers** and **free-tier accounts** where available. You only set up what you actually need.

## IMPORTANT: Run configure-vibe first

This skill assumes `configure-vibe` has already been run and chrome-devtools is active. If not, run `/configure-vibe` first.

---

## Step 1 — Show Available Integrations

Display this menu to the user and ask which ones they want to set up:

```
Available integrations (open source / free-tier):

  [1] GitHub        — Repos, PRs, issues, Actions CI/CD   FREE
  [2] Slack         — Messages, channels, DMs              FREE TIER
  [3] JIRA          — Issue tracking, project boards       FREE TIER
  [4] Confluence    — Wiki, documentation, knowledge base  FREE TIER (with JIRA)
  [5] Notion        — Notes, databases, project planning   FREE TIER
  [6] Linear        — Engineering issue tracking           FREE TIER
  [7] Brave Search  — Real-time web search (1000/mo free)  FREE TIER
  [8] Supabase      — Postgres DB + auth backend for apps  FREE TIER
  [9] Memory        — Local persistent AI memory graph     FREE (local only)
 [10] Fetch         — Web page fetcher/parser              FREE (local only)

Enter numbers separated by commas (e.g. "1,3,7") or "all":
```

Wait for the user's response, then set up only the selected integrations.

---

## Integration Setup Procedures

### [1] GitHub

**Create account (if needed):**
Use chrome-devtools to guide the user if they need an account:
```
Open https://github.com/signup in Chrome
Screenshot to confirm the page loaded
Walk through: enter email → password → username → verify email
```

**Get a Personal Access Token:**
1. Navigate to: https://github.com/settings/tokens?type=beta (fine-grained tokens)
2. Click "Generate new token"
3. Name it: `vibe-mcp`
4. Expiration: 1 year
5. Required permissions:
   - Repository access: All repositories (or select specific ones)
   - Repository permissions: Contents (Read), Issues (Read/Write), Pull requests (Read/Write)
   - Account permissions: Email addresses (Read)
6. Click "Generate token" — **copy the token immediately** (shown once)

**Configure:**
```bash
# Write token to env file
echo 'export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_YOUR_TOKEN_HERE"' >> ~/.vibe/env
source ~/.vibe/env
```

**Activate MCP:**
```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["github"] = {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-github"],
    "env": {"GITHUB_PERSONAL_ACCESS_TOKEN": os.environ.get("GITHUB_PERSONAL_ACCESS_TOKEN", "")}
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("GitHub MCP added. Restart Claude Code to activate.")
```

**Verify:** Restart Claude Code → tools `mcp__github__*` should appear.

---

### [2] Slack

**Create workspace (if needed):**
```
Open https://slack.com/get-started in Chrome
Create a free workspace (email + workspace name)
```

**Create a Slack App with Bot Token:**
1. Go to: https://api.slack.com/apps
2. Click "Create New App" → "From scratch"
3. App name: `vibe-assistant`, workspace: your workspace
4. Go to "OAuth & Permissions" → "Bot Token Scopes" → Add:
   - `channels:history`, `channels:read`, `chat:write`
   - `groups:history`, `groups:read`
   - `im:history`, `mpim:history`
   - `reactions:read`, `users:read`
5. Click "Install to Workspace" → Authorize
6. Copy the **Bot User OAuth Token** (starts with `xoxb-`)
7. Find your Team ID: Slack web app → your workspace URL subdomain, or workspace Settings → About

**Configure:**
```bash
echo 'export SLACK_BOT_TOKEN="xoxb-YOUR_TOKEN_HERE"' >> ~/.vibe/env
echo 'export SLACK_TEAM_ID="T0XXXXXXXXX"' >> ~/.vibe/env
source ~/.vibe/env
```

**Activate MCP:**
```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["slack"] = {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-slack"],
    "env": {
        "SLACK_BOT_TOKEN": os.environ.get("SLACK_BOT_TOKEN", ""),
        "SLACK_TEAM_ID": os.environ.get("SLACK_TEAM_ID", "")
    }
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("Slack MCP added. Restart Claude Code to activate.")
```

---

### [3] JIRA + [4] Confluence

Both use the same `mcp-atlassian` package and same Atlassian credentials.

**Create account (if needed):**
```
Open https://www.atlassian.com/software/jira/free in Chrome
Click "Get it free" → sign up with email
Create a new project (e.g., "Personal Projects")
```

**Create API Token:**
1. Go to: https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Label: `vibe-mcp`
4. Copy the token

**Get your workspace URL:**
Format: `https://your-workspace.atlassian.net` (shown in browser URL when in JIRA)

**Configure:**
```bash
echo 'export JIRA_URL="https://your-workspace.atlassian.net"' >> ~/.vibe/env
echo 'export JIRA_USERNAME="your-email@example.com"' >> ~/.vibe/env
echo 'export JIRA_API_TOKEN="your-api-token"' >> ~/.vibe/env
source ~/.vibe/env
```

**Activate JIRA MCP:**
```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["jira"] = {
    "command": "uvx",
    "args": ["mcp-atlassian"],
    "env": {
        "JIRA_URL": os.environ.get("JIRA_URL", ""),
        "JIRA_USERNAME": os.environ.get("JIRA_USERNAME", ""),
        "JIRA_API_TOKEN": os.environ.get("JIRA_API_TOKEN", ""),
        "MCP_PRIVACY_SUMMARIZATION_ENABLED": "false"
    }
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("JIRA MCP added.")
```

**Add Confluence** (optional, same credentials):
```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["confluence"] = {
    "command": "uvx",
    "args": ["mcp-atlassian"],
    "env": {
        "CONFLUENCE_URL": os.environ.get("JIRA_URL", "") + "/wiki",
        "CONFLUENCE_USERNAME": os.environ.get("JIRA_USERNAME", ""),
        "CONFLUENCE_API_TOKEN": os.environ.get("JIRA_API_TOKEN", "")
    }
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("Confluence MCP added.")
```

---

### [5] Notion

**Create account (if needed):**
```
Open https://www.notion.so/signup in Chrome
Sign up with email (free tier, unlimited personal pages)
```

**Create Integration Token:**
1. Go to: https://www.notion.so/my-integrations
2. Click "New integration"
3. Name: `vibe-assistant`
4. Capabilities: Read content, Update content, Insert content
5. Click "Submit" → copy the **Internal Integration Secret** (starts with `secret_`)
6. **Important**: In each Notion page you want to access → click "..." → "Add connections" → select your integration

**Configure:**
```bash
echo 'export NOTION_API_TOKEN="secret_YOUR_TOKEN_HERE"' >> ~/.vibe/env
source ~/.vibe/env
```

**Activate MCP:**
```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["notion"] = {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-notion"],
    "env": {"NOTION_API_TOKEN": os.environ.get("NOTION_API_TOKEN", "")}
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("Notion MCP added. Restart Claude Code to activate.")
```

---

### [6] Linear

**Create account (if needed):**
```
Open https://linear.app in Chrome
Click "Get started" → sign up free
Create a team/workspace
```

**Get API Key:**
1. Go to: Linear → Settings → API → Personal API keys
2. Click "Create key"
3. Label: `vibe-mcp`
4. Copy the key (starts with `lin_api_`)

**Configure:**
```bash
echo 'export LINEAR_API_KEY="lin_api_YOUR_KEY_HERE"' >> ~/.vibe/env
source ~/.vibe/env
```

**Activate MCP:**
```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["linear"] = {
    "command": "npx",
    "args": ["-y", "linear-mcp-server"],
    "env": {"LINEAR_API_KEY": os.environ.get("LINEAR_API_KEY", "")}
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("Linear MCP added. Restart Claude Code to activate.")
```

---

### [7] Brave Search

**Get API Key (free tier — 1000 queries/month):**
1. Go to: https://brave.com/search/api/
2. Click "Get started for free"
3. Sign up → select the "Free" plan
4. Copy your API key (starts with `BSA-`)

**Configure:**
```bash
echo 'export BRAVE_API_KEY="BSA-YOUR_KEY_HERE"' >> ~/.vibe/env
source ~/.vibe/env
```

**Activate MCP:**
```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["brave-search"] = {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-brave-search"],
    "env": {"BRAVE_API_KEY": os.environ.get("BRAVE_API_KEY", "")}
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("Brave Search MCP added. Restart Claude Code to activate.")
```

---

### [8] Supabase

**Create project (if needed):**
```
Open https://app.supabase.com in Chrome
Click "New project"
Organization: Personal, name your project
Wait ~2 minutes for provisioning
```

**Get Access Token:**
1. Go to: https://supabase.com/dashboard/account/tokens
2. Click "Generate new token"
3. Name: `vibe-mcp`
4. Copy the token

**Configure:**
```bash
echo 'export SUPABASE_ACCESS_TOKEN="sbp_YOUR_TOKEN_HERE"' >> ~/.vibe/env
source ~/.vibe/env
```

**Activate MCP:**
```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["supabase"] = {
    "command": "npx",
    "args": ["-y", "@supabase/mcp-server-supabase@latest", "--read-only"],
    "env": {"SUPABASE_ACCESS_TOKEN": os.environ.get("SUPABASE_ACCESS_TOKEN", "")}
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("Supabase MCP added. Restart Claude Code to activate.")
```

---

### [9] Memory (Local — No Account Needed)

Persistent knowledge graph stored locally at `~/.vibe/memory/graph.json`.

```bash
mkdir -p ~/.vibe/memory
```

```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["memory"] = {
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-memory"],
    "env": {"MEMORY_FILE_PATH": os.path.expanduser("~/.vibe/memory/graph.json")}
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("Memory MCP added. Restart Claude Code to activate.")
```

---

### [10] Fetch (Local — No Account Needed)

Fetches any web URL and returns it as clean markdown — useful for reading docs, blog posts, APIs.

```python
import json, os
path = os.path.expanduser("~/.claude.json")
with open(path) as f: cfg = json.load(f)
cfg.setdefault("mcpServers", {})["fetch"] = {
    "command": "uvx",
    "args": ["mcp-server-fetch"]
}
with open(path, "w") as f: json.dump(cfg, f, indent=2)
print("Fetch MCP added. Restart Claude Code to activate.")
```

---

## Step 2 — Verify Installation

After setting up selected integrations, run the chrome-devtools check to verify the browser works, then restart Claude Code:

```
Tell the user: "Restart Claude Code now to activate the new MCP servers."
After they restart: run /validate-mcp-access to confirm all new integrations are live.
```

---

## Step 3 — Final Summary

After setup, display a table of what was configured:

| Integration | Status | Notes |
|-------------|--------|-------|
| GitHub | ✅/❌ | — |
| Slack | ✅/❌ | — |
| JIRA/Confluence | ✅/❌ | — |
| Notion | ✅/❌ | — |
| Linear | ✅/❌ | — |
| Brave Search | ✅/❌ | — |
| Supabase | ✅/❌ | — |
| Memory | ✅/❌ | — |
| Fetch | ✅/❌ | — |

Remind the user that:
1. New MCPs require a Claude Code restart to activate
2. Tokens/keys are stored in `~/.vibe/env` — keep this file private
3. Run `/validate-mcp-access` after restarting to confirm everything works
4. Run `/setup-integrations` again any time to add more integrations
