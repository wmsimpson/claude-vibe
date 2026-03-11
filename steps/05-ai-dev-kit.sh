#!/usr/bin/env bash
# Step 5: Install Databricks AI Dev Kit

step_ai_dev_kit() {
  print_header "Step 5 of 8 — Databricks AI Dev Kit"

  if ! is_databricks_enabled; then
    print_info "Databricks integration is not enabled — skipping AI Dev Kit."
    print_info "To enable, run: ${CYAN}vibe install --step 5${NC}"
    mark_step_complete "ai_dev_kit"
    return 0
  fi

  local skills_dir="$HOME/.claude/skills"
  local aidk_dir="$HOME/.ai-dev-kit"

  # Check if already installed
  if [[ -d "$aidk_dir" ]] && [[ -d "$skills_dir" ]]; then
    local skill_count
    skill_count=$(ls "$skills_dir" 2>/dev/null | wc -l | tr -d ' ')
    if [[ $skill_count -gt 20 ]]; then
      print_success "AI Dev Kit already installed ($skill_count skills)"
      if ask_yes_no "Skip this step?" "y"; then
        _merge_aidk_mcp
        mark_step_complete "ai_dev_kit"
        return 0
      fi
    fi
  fi

  print_info "The Databricks AI Dev Kit provides 50+ MCP tools and 34 skills."
  print_info "Source: github.com/databricks-solutions/ai-dev-kit"
  print_blank

  if ! ask_yes_no "Install Databricks AI Dev Kit?" "y"; then
    print_info "Skipping — you can install later with:"
    print_info "  bash <(curl -sL https://raw.githubusercontent.com/databricks-solutions/ai-dev-kit/main/install.sh)"
    mark_step_complete "ai_dev_kit"
    return 0
  fi

  print_blank
  print_step "Launching AI Dev Kit installer..."
  print_info "Select: Claude Code, your Databricks profile, and global scope"
  print_blank

  # Run the interactive installer
  bash <(curl -sL https://raw.githubusercontent.com/databricks-solutions/ai-dev-kit/main/install.sh) </dev/tty

  # Merge MCP config to global scope
  _merge_aidk_mcp

  # Verify
  local skill_count
  skill_count=$(ls "$skills_dir" 2>/dev/null | wc -l | tr -d ' ')
  if [[ $skill_count -gt 0 ]]; then
    print_success "$skill_count skills installed in ~/.claude/skills/"
    mark_step_complete "ai_dev_kit"
  else
    print_warn "No skills found — installer may not have completed"
  fi
}

_merge_aidk_mcp() {
  # Move project-scope MCP config to global ~/.claude.json
  if [[ -f "$HOME/.mcp.json" ]]; then
    python3 -c "
import json, os
mcp_path = os.path.expanduser('~/.mcp.json')
claude_path = os.path.expanduser('~/.claude.json')

with open(mcp_path) as f:
    mcp = json.load(f)

if os.path.exists(claude_path):
    with open(claude_path) as f:
        cfg = json.load(f)
else:
    cfg = {}

servers = mcp.get('mcpServers', {})
if servers:
    cfg.setdefault('mcpServers', {}).update(servers)
    with open(claude_path, 'w') as f:
        json.dump(cfg, f, indent=2)
    print('Merged:', list(servers.keys()))
" 2>/dev/null && print_success "MCP server config merged to global scope"
  fi
}
