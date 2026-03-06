#!/usr/bin/env bash
# Step 7: Configure MCP Integrations

step_configure_mcp() {
  print_header "Step 7 of 8 — MCP Integrations"

  print_info "MCP servers extend Claude Code with live integrations."
  print_info "Each runs as a local subprocess that Claude communicates with."
  print_blank

  # ── GitHub MCP ──────────────────────────────────────────────────────────
  print_step "GitHub MCP"

  if _is_mcp_configured "github"; then
    print_success "GitHub MCP already configured"
  else
    if command -v gh &>/dev/null && gh auth status &>/dev/null 2>&1; then
      print_success "GitHub CLI authenticated"
      if ask_yes_no "Set up GitHub MCP using your existing gh auth?" "y"; then
        _setup_github_mcp
      fi
    else
      if ask_yes_no "Set up GitHub MCP? (requires GitHub account)" "y"; then
        print_info "You need a GitHub Personal Access Token."
        print_info "Create one at: github.com/settings/tokens"
        print_info "Required scopes: repo, read:org, gist"
        print_blank
        local token
        token=$(ask_secret "GitHub Personal Access Token")
        if [[ -n "$token" ]]; then
          _add_mcp_server "github" "npx" '["-y","@modelcontextprotocol/server-github"]' \
            "{\"GITHUB_PERSONAL_ACCESS_TOKEN\":\"$token\"}"
          _add_env_var "GITHUB_PERSONAL_ACCESS_TOKEN" "$token"
          print_success "GitHub MCP configured"
        fi
      fi
    fi
  fi

  # ── Chrome DevTools ─────────────────────────────────────────────────────
  print_blank
  print_step "Chrome DevTools MCP"

  if _is_mcp_configured "chrome-devtools"; then
    print_success "Chrome DevTools MCP already configured"
  else
    if ask_yes_no "Set up Chrome DevTools MCP? (no account needed)" "y"; then
      mkdir -p "$HOME/.vibe/chrome/profile"
      _add_mcp_server "chrome-devtools" "npx" \
        "[\"chrome-devtools-mcp@latest\",\"--userDataDir=$HOME/.vibe/chrome/profile\"]" "{}"
      print_success "Chrome DevTools MCP configured"
    fi
  fi

  # ── Optional integrations ──────────────────────────────────────────────
  print_blank
  print_step "Optional integrations"
  print_blank

  local optional_mcps=("Slack" "JIRA" "Brave Search" "Notion" "Skip")
  ask_multi_select "Select additional integrations to configure:" "${optional_mcps[@]}"

  for idx in "${SELECTED_INDICES[@]}"; do
    case "${optional_mcps[$idx]}" in
      "Slack")
        print_blank
        print_info "Slack MCP requires a bot token from api.slack.com/apps"
        local slack_token
        slack_token=$(ask_secret "Slack Bot Token (xoxb-...)")
        local slack_team
        slack_team=$(ask_input "Slack Team/Workspace ID (T0...)")
        if [[ -n "$slack_token" && -n "$slack_team" ]]; then
          _add_mcp_server "slack" "npx" '["-y","@modelcontextprotocol/server-slack"]' \
            "{\"SLACK_BOT_TOKEN\":\"$slack_token\",\"SLACK_TEAM_ID\":\"$slack_team\"}"
          _add_env_var "SLACK_BOT_TOKEN" "$slack_token"
          print_success "Slack MCP configured"
        fi
        ;;
      "JIRA")
        print_blank
        print_info "JIRA MCP requires an Atlassian API token"
        print_info "Create one at: id.atlassian.com/manage-profile/security/api-tokens"
        local jira_url
        jira_url=$(ask_input "JIRA URL" "https://your-org.atlassian.net")
        local jira_user
        jira_user=$(ask_input "JIRA email")
        local jira_token
        jira_token=$(ask_secret "JIRA API Token")
        if [[ -n "$jira_url" && -n "$jira_user" && -n "$jira_token" ]]; then
          _add_mcp_server "jira" "uvx" '["mcp-atlassian"]' \
            "{\"JIRA_URL\":\"$jira_url\",\"JIRA_USERNAME\":\"$jira_user\",\"JIRA_API_TOKEN\":\"$jira_token\"}"
          _add_env_var "JIRA_API_TOKEN" "$jira_token"
          print_success "JIRA MCP configured"
        fi
        ;;
      "Brave Search")
        print_blank
        print_info "Free: 1000 queries/month at brave.com/search/api"
        local brave_key
        brave_key=$(ask_secret "Brave Search API Key")
        if [[ -n "$brave_key" ]]; then
          _add_mcp_server "brave-search" "npx" '["-y","@modelcontextprotocol/server-brave-search"]' \
            "{\"BRAVE_API_KEY\":\"$brave_key\"}"
          _add_env_var "BRAVE_API_KEY" "$brave_key"
          print_success "Brave Search MCP configured"
        fi
        ;;
      "Notion")
        print_blank
        print_info "Create an integration at notion.so/my-integrations"
        local notion_token
        notion_token=$(ask_secret "Notion API Token")
        if [[ -n "$notion_token" ]]; then
          _add_mcp_server "notion" "npx" '["-y","@modelcontextprotocol/server-notion"]' \
            "{\"NOTION_API_TOKEN\":\"$notion_token\"}"
          _add_env_var "NOTION_API_TOKEN" "$notion_token"
          print_success "Notion MCP configured"
        fi
        ;;
      "Skip") ;;
    esac
  done

  # Show summary
  print_blank
  print_step "Configured MCP servers:"
  python3 -c "
import json, os
p = os.path.expanduser('~/.claude.json')
if os.path.exists(p):
    with open(p) as f:
        cfg = json.load(f)
    for name in cfg.get('mcpServers', {}):
        print(f'  {name}')
" 2>/dev/null

  mark_step_complete "configure_mcp"
}

_setup_github_mcp() {
  local token
  token=$(gh auth token 2>/dev/null)
  if [[ -n "$token" ]]; then
    _add_mcp_server "github" "npx" '["-y","@modelcontextprotocol/server-github"]' \
      "{\"GITHUB_PERSONAL_ACCESS_TOKEN\":\"$token\"}"
    _add_env_var "GITHUB_PERSONAL_ACCESS_TOKEN" "$token"
    print_success "GitHub MCP configured (using gh auth token)"
  fi
}

_is_mcp_configured() {
  local name="$1"
  python3 -c "
import json, os
p = os.path.expanduser('~/.claude.json')
if not os.path.exists(p): exit(1)
with open(p) as f: cfg = json.load(f)
exit(0 if '$name' in cfg.get('mcpServers', {}) else 1)
" 2>/dev/null
}

_add_mcp_server() {
  local name="$1" cmd="$2" args="$3" env="$4"
  python3 -c "
import json, os
p = os.path.expanduser('~/.claude.json')
cfg = json.load(open(p)) if os.path.exists(p) else {}
cfg.setdefault('mcpServers', {})['$name'] = {
    'type': 'stdio',
    'command': '$cmd',
    'args': $args,
    'env': $env
}
with open(p, 'w') as f:
    json.dump(cfg, f, indent=2)
" 2>/dev/null
}

_add_env_var() {
  local key="$1" value="$2"
  local env_file="$HOME/.vibe/env"
  mkdir -p "$HOME/.vibe"
  if grep -q "^export $key=" "$env_file" 2>/dev/null; then
    sed -i '' "s|^export $key=.*|export $key=$value|" "$env_file"
  elif grep -q "^# export $key=" "$env_file" 2>/dev/null; then
    sed -i '' "s|^# export $key=.*|export $key=$value|" "$env_file"
  else
    echo "export $key=$value" >> "$env_file"
  fi
}
