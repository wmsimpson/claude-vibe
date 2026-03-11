---
name: validate-mcp-access
description: Run a live self-diagnostic of all configured tools, MCP servers, and external integrations. Uses Chrome DevTools to visually debug any issues. Run this after configure-vibe or when something isn't working. Triggers on "validate setup", "check connections", "something isn't working", "debug my setup", "test integrations".
user-invocable: true
---

# Validate Setup & Connections

Active self-diagnostic — verifies chrome-devtools, local access, Google APIs, and all configured tools. Uses Chrome DevTools for visual debugging when needed. Designed to work without external IT support.

---

## Phase 1 — chrome-devtools (Visual Debugging)

**Run first.** If chrome-devtools works, you can use it to debug any other failures visually.

### Check 1a: MCP connection active

Try to use `mcp__chrome-devtools__new_page` to open a new Chrome tab. If the tool is available in this session, chrome-devtools is working.

**If chrome-devtools MCP tools ARE available:**
```
Use mcp__chrome-devtools__new_page to open a new page
Navigate to: https://www.google.com
Take a screenshot with mcp__chrome-devtools__take_screenshot
Report: "Chrome DevTools MCP ✅ — visual debugging available"
```

**If chrome-devtools MCP tools are NOT available:**

Check the config:
```bash
python3 -c "
import json, os
with open(os.path.expanduser('~/.claude.json')) as f:
    cfg = json.load(f)
servers = cfg.get('mcpServers', {})
if 'chrome-devtools' in servers:
    print('Config found:', json.dumps(servers['chrome-devtools'], indent=2))
else:
    print('NOT configured — chrome-devtools missing from mcpServers')
"
```

Check npx is available:
```bash
which npx && npx --version
```

If config is present but tools not available: **Restart Claude Code** — MCP servers load at startup.

If config is missing, add it:
```bash
python3 << 'PYEOF'
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
print("Added chrome-devtools to ~/.claude.json. Restart Claude Code to activate.")
PYEOF
```

---

## Phase 2 — Local Directory Access

Verify Claude can read and write to key local directories.

### Check 2a: Home directory read access
```bash
ls ~/ | head -5 && echo "✅ Home directory readable"
```

### Check 2b: Code directory
```bash
ls ~/code/ 2>/dev/null && echo "✅ ~/code exists" || (mkdir -p ~/code && echo "Created ~/code")
```

### Check 2c: Vibe directory
```bash
ls ~/.vibe/ 2>/dev/null && echo "✅ ~/.vibe exists" || (mkdir -p ~/.vibe && echo "Created ~/.vibe")
```

### Check 2d: Permissions in Claude settings
```bash
python3 -c "
import json, os
with open(os.path.expanduser('~/.claude/settings.json')) as f:
    s = json.load(f)
perms = s.get('permissions', {}).get('allow', [])
file_perms = [p for p in perms if p.startswith('Read(') or p.startswith('Edit(') or p.startswith('Write(')]
skill_count = len([p for p in perms if p.startswith('Skill(')])
print(f'File permissions: {len(file_perms)} entries')
print(f'Skill permissions: {skill_count} entries')
for p in sorted(file_perms):
    print(' ', p)
"
```

Expected: `Read(~/**`, `Read(//tmp/**)`, `Edit(~/code/**)`, `Write(~/code/**)`, `Write(~/.vibe/**)` present.

If missing, sync permissions from source:
```bash
cd ~/code/claude-vibe
NEW_SKILLS=$(grep -E '^\s*- "Skill\(' permissions.yaml | sed 's/.*"\(Skill([^)]*)\)".*/"\1"/' | tr '\n' ',' | sed 's/,$//')
python3 -c "
import json, os
with open(os.path.expanduser('~/.claude/settings.json')) as f:
    s = json.load(f)
new = json.loads('[$NEW_SKILLS]'.replace(\"'\", '\"'))
existing = s.get('permissions', {}).get('allow', [])
non_skill = [x for x in existing if not x.startswith('Skill(')]
s.setdefault('permissions', {})['allow'] = sorted(set(non_skill + new))
with open(os.path.expanduser('~/.claude/settings.json'), 'w') as f:
    json.dump(s, f, indent=2)
print('Permissions synced.')
"
```

---

## Phase 3 — Google Account & API Access

### Check 3a: Auth credentials exist
```bash
python3 -c "
import json, os
adc = os.path.expanduser('~/.config/gcloud/application_default_credentials.json')
if not os.path.exists(adc):
    print('❌ No ADC credentials — run configure-vibe Step 5 to sign in')
else:
    with open(adc) as f:
        d = json.load(f)
    ctype = d.get('type', 'unknown')
    client = d.get('client_id', '')[:40]
    quota = d.get('quota_project_id', 'NOT SET')
    print(f'✅ Credentials found: type={ctype}')
    print(f'   Quota project: {quota}')
    if quota == 'NOT SET':
        print('   ⚠️  Sheets/Slides/Forms will fail until quota project is set')
        print('   Fix: gcloud auth application-default set-quota-project YOUR_PROJECT_ID')
"
```

### Check 3b: Token is valid
```bash
TOKEN=$(gcloud auth application-default print-access-token 2>/dev/null)
if [ -z "$TOKEN" ]; then
    echo "❌ Token invalid — re-run: gcloud auth application-default login"
else
    echo "✅ Access token valid (${#TOKEN} chars)"
    # Get account email from token
    EMAIL=$(curl -s "https://oauth2.googleapis.com/tokeninfo?access_token=$TOKEN" | \
            python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('email', 'unknown'))" 2>/dev/null)
    echo "   Signed in as: $EMAIL"
fi
```

### Check 3c: Test each Google API

```bash
TOKEN=$(gcloud auth application-default print-access-token 2>/dev/null)
QUOTA="${GCP_QUOTA_PROJECT:-}"

check_api() {
    local name="$1" url="$2"
    local args=("-s" "-o" "/dev/null" "-w" "%{http_code}" "$url" "-H" "Authorization: Bearer $TOKEN")
    [ -n "$QUOTA" ] && args+=("-H" "x-goog-user-project: $QUOTA")
    local code=$(curl "${args[@]}" 2>/dev/null)
    case "$code" in
        200|404) echo "✅ $name (HTTP $code)" ;;
        403)     echo "❌ $name — 403 Forbidden (quota project needed — see Step 6)" ;;
        401)     echo "❌ $name — 401 Unauthorized (token expired — re-run login)" ;;
        000)     echo "⚠️  $name — network unreachable" ;;
        *)       echo "⚠️  $name — unexpected HTTP $code" ;;
    esac
}

check_api "Google Docs"     "https://docs.googleapis.com/v1/documents/test-invalid"
check_api "Google Drive"    "https://www.googleapis.com/drive/v3/files?pageSize=1"
check_api "Gmail"           "https://gmail.googleapis.com/gmail/v1/users/me/profile"
check_api "Google Calendar" "https://www.googleapis.com/calendar/v3/calendars/primary"
check_api "Google Sheets"   "https://sheets.googleapis.com/v4/spreadsheets/test-invalid"
check_api "Google Slides"   "https://slides.googleapis.com/v1/presentations/test-invalid"
check_api "Google Forms"    "https://forms.googleapis.com/v1/forms/test-invalid"
```

### If Sheets/Slides/Forms fail (403):
Use chrome-devtools to walk through the fix visually:
```
Open https://console.cloud.google.com/ in Chrome (use mcp__chrome-devtools__new_page)
Take a screenshot to confirm you're signed into the right Google account
If not signed in, navigate to the sign-in page
Once in the console:
  - Check the project selector at the top — create a new project if needed
  - Go to APIs & Services → Library
  - Search for "Sheets API" and verify it's enabled
  - Repeat for Slides API, Forms API, Drive API
Then run in terminal:
  gcloud auth application-default set-quota-project YOUR_PROJECT_ID
  echo 'export GCP_QUOTA_PROJECT=YOUR_PROJECT_ID' >> ~/.vibe/env
  source ~/.vibe/env
Re-run Check 3c to confirm ✅
```

---

## Phase 4 — Optional MCP Servers

Check which optional MCP servers are configured in `~/.claude.json`:

```python
import json, os
with open(os.path.expanduser("~/.claude.json")) as f:
    cfg = json.load(f)
servers = cfg.get("mcpServers", {})
known = ["chrome-devtools", "github", "slack", "jira", "confluence",
         "notion", "linear", "brave-search", "supabase", "memory", "fetch"]
print("=== MCP Servers in ~/.claude.json ===")
for s in known:
    if s in servers:
        print(f"  ✅ {s} — configured")
    else:
        print(f"  — {s} — not configured (run /setup-integrations to add)")
extra = [s for s in servers if s not in known]
for s in extra:
    print(f"  ✅ {s} — configured (custom)")
```

For each configured optional MCP, verify the required env var is set:
```bash
echo "=== MCP env vars ==="
[ -n "$GITHUB_PERSONAL_ACCESS_TOKEN" ] && echo "  ✅ GITHUB_PERSONAL_ACCESS_TOKEN set" || echo "  — GITHUB_PERSONAL_ACCESS_TOKEN not set"
[ -n "$SLACK_BOT_TOKEN" ] && echo "  ✅ SLACK_BOT_TOKEN set" || echo "  — SLACK_BOT_TOKEN not set"
[ -n "$JIRA_API_TOKEN" ] && echo "  ✅ JIRA_API_TOKEN set" || echo "  — JIRA_API_TOKEN not set"
[ -n "$NOTION_API_TOKEN" ] && echo "  ✅ NOTION_API_TOKEN set" || echo "  — NOTION_API_TOKEN not set"
[ -n "$LINEAR_API_KEY" ] && echo "  ✅ LINEAR_API_KEY set" || echo "  — LINEAR_API_KEY not set"
[ -n "$BRAVE_API_KEY" ] && echo "  ✅ BRAVE_API_KEY set" || echo "  — BRAVE_API_KEY not set"
[ -n "$SUPABASE_ACCESS_TOKEN" ] && echo "  ✅ SUPABASE_ACCESS_TOKEN set" || echo "  — SUPABASE_ACCESS_TOKEN not set"
```

**If a server is configured but its env var is missing:** the MCP will fail to start. Fix: run `/setup-integrations` and re-configure that integration to write the token to `~/.vibe/env`.

**If no optional MCPs are configured:** that's fine — run `/setup-integrations` to add services as needed.

---

## Phase 5 — Installed Plugins

```bash
claude plugin list 2>/dev/null | grep -E "(›|Status)" | head -20
```

Expected: all plugins show `✔ enabled`.

If a plugin shows `✘ disabled` or is missing:
```bash
# Re-install a specific plugin
claude plugin install PLUGIN_NAME@claude-vibe

# Re-install all
for p in app-dev databricks-tools google-tools specialized-agents vibe-setup jira-tools workflows lean-sigma-tools macos-scheduler; do
    claude plugin install "${p}@claude-vibe" 2>&1 | tail -1
done
```

---

## Phase 5 — Optional Tools Check

```bash
echo "=== Development Tools ==="
which go && go version || echo "❌ Go not installed"
which node && node --version || echo "❌ Node not installed"
which gh && gh --version | head -1 || echo "❌ GitHub CLI not installed"
which mmdc && mmdc --version || echo "⚠️  mermaid-cli not installed (npm install -g @mermaid-js/mermaid-cli)"
which dot && dot -V 2>&1 || echo "⚠️  graphviz not installed (brew install graphviz)"

echo ""
echo "=== Mobile Dev Tools (optional) ==="
which expo && expo --version || echo "⚠️  Expo CLI not installed"
which flutter && flutter --version | head -1 || echo "⚠️  Flutter not installed"
xcode-select -p 2>/dev/null && echo "✅ Xcode CLI tools" || echo "⚠️  Xcode CLI tools: run xcode-select --install"
which pod && pod --version || echo "⚠️  CocoaPods not installed"

echo ""
echo "=== Databricks (optional) ==="
which databricks && databricks --version || echo "⚠️  Databricks CLI not installed"
databricks auth profiles 2>/dev/null || echo "⚠️  No Databricks profiles configured"
```

---

## Final Report

Summarize results in a table:

| Component | Status | Action Needed |
|-----------|--------|---------------|
| chrome-devtools MCP | ✅/❌ | — or restart Claude Code |
| Local directory access | ✅/❌ | — or sync permissions |
| Google sign-in | ✅/❌ | — or re-run configure-vibe Step 5 |
| Google Docs/Gmail | ✅/❌ | — |
| Google Sheets/Slides/Forms | ✅/❌ | — or set GCP_QUOTA_PROJECT (Step 6) |
| Plugins installed | ✅/❌ | — or reinstall |
| GitHub CLI | ✅/❌ | — or `gh auth login` |
| Optional MCPs (GitHub/Slack/etc.) | configured/not configured | — or run /setup-integrations |

For any ❌ items: provide the exact command to fix it. If the issue is visual (wrong account, wrong project), use chrome-devtools to navigate and screenshot to show the user exactly what to do.
