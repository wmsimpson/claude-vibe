#!/usr/bin/env bash
# Step 8: Full Validation

step_validate() {
  print_header "Step 8 of 8 — Full Validation"

  local pass=0
  local fail=0
  local warn=0

  # ── Claude Code ─────────────────────────────────────────────────────────
  print_step "Claude Code"
  if command -v claude &>/dev/null; then
    print_success "claude $(claude --version 2>&1)"
    ((pass++))
  else
    print_error "claude not found"
    ((fail++))
  fi

  # ── Google Auth ─────────────────────────────────────────────────────────
  print_blank
  print_step "Google Auth"
  local token
  token=$(gcloud auth application-default print-access-token 2>/dev/null)
  if [[ -n "$token" ]]; then
    print_success "Access token valid (${#token} chars)"
    ((pass++))
  else
    print_error "No valid access token"
    ((fail++))
  fi

  # ── Google APIs ─────────────────────────────────────────────────────────
  print_blank
  print_step "Google APIs"

  local gcp_project=""
  [[ -f "$HOME/.claude-vibe/gcp-project-id" ]] && gcp_project=$(cat "$HOME/.claude-vibe/gcp-project-id")
  if [[ -z "$gcp_project" ]]; then
    gcp_project=$(grep -v '^#' "$HOME/.vibe/env" 2>/dev/null | grep GCP_QUOTA_PROJECT | cut -d= -f2 | tr -d ' "'"'"'')
  fi

  if [[ -n "$token" ]]; then
    for api_check in \
      "Drive|https://www.googleapis.com/drive/v3/about?fields=user" \
      "Gmail|https://gmail.googleapis.com/gmail/v1/users/me/profile" \
      "Calendar|https://www.googleapis.com/calendar/v3/calendars/primary" \
      "Sheets|https://sheets.googleapis.com/v4/spreadsheets/test" \
      "Docs|https://docs.googleapis.com/v1/documents/test" \
      "Slides|https://slides.googleapis.com/v1/presentations/test" \
      "Forms|https://forms.googleapis.com/v1/forms/test" \
      "Tasks|https://tasks.googleapis.com/tasks/v1/users/@me/lists?maxResults=1"; do
      local name="${api_check%%|*}"
      local url="${api_check##*|}"
      local code
      code=$(curl -s -o /dev/null -w "%{http_code}" "$url" \
        -H "Authorization: Bearer $token" \
        -H "x-goog-user-project: $gcp_project" 2>/dev/null)
      case "$code" in
        200|404) print_success "$name"; ((pass++)) ;;
        403)     print_error "$name (403 — check quota project)"; ((fail++)) ;;
        401)     print_error "$name (401 — re-auth needed)"; ((fail++)) ;;
        *)       print_warn "$name (HTTP $code)"; ((warn++)) ;;
      esac
    done
  fi

  # ── Core Tools ──────────────────────────────────────────────────────────
  print_blank
  print_step "Core Tools"
  for tool in go jq yq rg uv node git gh mmdc; do
    if command -v "$tool" &>/dev/null; then
      local ver
      ver=$($tool --version 2>/dev/null | head -1)
      print_success "$tool ${DIM}${ver}${NC}"
      ((pass++))
    else
      print_error "$tool not found"
      ((fail++))
    fi
  done
  # graphviz (dot) — check separately since the binary name differs
  if command -v dot &>/dev/null || [[ -x /usr/local/bin/dot ]]; then
    print_success "graphviz (dot)"
    ((pass++))
  else
    print_warn "graphviz (dot) not found"
    ((warn++))
  fi

  # ── Databricks AI Dev Kit ──────────────────────────────────────────────
  print_blank
  print_step "Databricks AI Dev Kit"
  local skill_count
  skill_count=$(ls "$HOME/.claude/skills/" 2>/dev/null | wc -l | tr -d ' ')
  if [[ $skill_count -gt 0 ]]; then
    print_success "$skill_count skills in ~/.claude/skills/"
    ((pass++))
  else
    print_warn "No skills found in ~/.claude/skills/"
    ((warn++))
  fi

  # ── Plugins ─────────────────────────────────────────────────────────────
  print_blank
  print_step "Claude Code Plugins"
  python3 -c "
import json, os
p = os.path.expanduser('~/.claude/settings.json')
if not os.path.exists(p):
    print('NO_SETTINGS')
    exit(1)
with open(p) as f:
    s = json.load(f)
plugins = s.get('enabledPlugins', {})
enabled = sum(1 for v in plugins.values() if v)
perms = s.get('permissions', {}).get('allow', [])
skills = len([p for p in perms if p.startswith('Skill(')])
print(f'{enabled}|{skills}')
" 2>/dev/null | {
    IFS='|' read -r plugin_count skill_perm_count
    if [[ -n "$plugin_count" && "$plugin_count" != "NO_SETTINGS" ]]; then
      print_success "$plugin_count plugins enabled"
      print_success "$skill_perm_count skill permissions"
      ((pass += 2))
    else
      print_warn "No plugins or settings found"
      ((warn++))
    fi
  }

  # ── MCP Servers ─────────────────────────────────────────────────────────
  print_blank
  print_step "MCP Servers"
  python3 -c "
import json, os
p = os.path.expanduser('~/.claude.json')
if not os.path.exists(p):
    exit(1)
with open(p) as f:
    cfg = json.load(f)
for name in cfg.get('mcpServers', {}):
    print(name)
" 2>/dev/null | while read -r server; do
    print_success "$server"
    ((pass++))
  done

  # ── GitHub MCP live test ────────────────────────────────────────────────
  if _is_mcp_configured "github" 2>/dev/null; then
    print_blank
    print_step "GitHub MCP live test"
    local gh_token
    gh_token=$(gh auth token 2>/dev/null)
    if [[ -n "$gh_token" ]]; then
      local result
      result=$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
        GITHUB_PERSONAL_ACCESS_TOKEN="$gh_token" timeout 15 npx -y @modelcontextprotocol/server-github 2>/dev/null | head -1)
      if echo "$result" | python3 -c "import sys,json; json.loads(sys.stdin.readline())['result']" &>/dev/null; then
        print_success "GitHub MCP server responds"
        ((pass++))
      else
        print_warn "GitHub MCP did not respond (may need npx download)"
        ((warn++))
      fi
    fi
  fi

  # ── Environment ─────────────────────────────────────────────────────────
  print_blank
  print_step "Environment"
  [[ -f "$HOME/.vibe/env" ]] && print_success "~/.vibe/env exists" && ((pass++)) || { print_error "~/.vibe/env missing"; ((fail++)); }
  [[ -f "$HOME/.config/gcloud/credentials/claude-google-auth.json" ]] && print_success "OAuth credentials file exists" && ((pass++)) || print_warn "OAuth credentials not in standard location"

  # Check for clean zprofile
  local dup_count
  dup_count=$(grep -c 'HOME/.local/bin' "$HOME/.zprofile" 2>/dev/null)
  if [[ $dup_count -gt 1 ]]; then
    print_warn "~/.zprofile has duplicate PATH entries ($dup_count)"
    ((warn++))
  else
    print_success "~/.zprofile is clean"
    ((pass++))
  fi

  # ── Summary ─────────────────────────────────────────────────────────────
  print_blank
  echo -e "  ${BOLD}────────────────────────────────────────${NC}"
  echo -e "  ${GREEN}${BOLD}$pass passed${NC}  ${RED}$fail failed${NC}  ${YELLOW}$warn warnings${NC}"
  echo -e "  ${BOLD}────────────────────────────────────────${NC}"
  print_blank

  if [[ $fail -eq 0 ]]; then
    echo -e "  ${GREEN}${BOLD}Setup complete!${NC} Restart Claude Code to activate all changes."
    print_blank
    echo -e "  ${DIM}Try these in Claude Code:${NC}"
    echo -e "    ${CYAN}\"Create a Google Doc titled Hello World\"${NC}"
    echo -e "    ${CYAN}\"List my GitHub repos\"${NC}"
    echo -e "    ${CYAN}\"What SQL warehouses do I have?\"${NC}"
    mark_step_complete "validate"
  else
    echo -e "  ${YELLOW}Some checks failed. Fix the issues above and re-run:${NC}"
    echo -e "    ${CYAN}./setup.sh --validate${NC}"
  fi
  print_blank
}
