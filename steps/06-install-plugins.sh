#!/usr/bin/env bash
# Step 6: Install Claude Code Plugins

step_install_plugins() {
  print_header "Step 6 of 8 — Claude Code Plugins"

  print_info "Installing plugin collections from claude-vibe."
  print_blank

  # Core plugins (always available)
  local all_plugins=(
    "google-tools"
    "app-dev"
    "workflows"
    "vibe-setup"
    "specialized-agents"
    "community-skills"
    "macos-scheduler"
    "lean-sigma-tools"
    "jira-tools"
    "mcp-servers"
  )

  local descriptions=(
    "Google Workspace (Gmail, Docs, Sheets, Slides, Calendar, Forms, Tasks)"
    "App Development (React Native, Expo, Next.js, Swift, Flutter)"
    "Workflows (Architecture diagrams, RCA, POC docs, Security questionnaires)"
    "Setup & Diagnostics (Validate, usage stats, integrations)"
    "Diagram Agents (Lucid Chart, Graphviz)"
    "Community Skills (Humanizer, PPTX, Find Skills, Web Design, Spark Data Source)"
    "macOS Scheduler (launchd background tasks)"
    "Lean Six Sigma (FMEA, SIPOC, Process Maps)"
    "JIRA (Search, create, update tickets)"
    "MCP Server Configs"
  )

  # Add Databricks plugin if enabled
  if is_databricks_enabled; then
    all_plugins+=("databricks-tools")
    descriptions+=("Databricks Tools (Query, deploy, manage workspaces)")
  fi

  print_info "Available plugins:"
  print_blank
  for i in "${!all_plugins[@]}"; do
    echo -e "    ${CYAN}${all_plugins[$i]}${NC} — ${DIM}${descriptions[$i]}${NC}"
  done
  print_blank

  local selected_plugins=()
  if ask_yes_no "Install all plugins?" "y"; then
    selected_plugins=("${all_plugins[@]}")
  else
    ask_multi_select "Select plugins to install:" "${all_plugins[@]}"
    for idx in "${SELECTED_INDICES[@]}"; do
      selected_plugins+=("${all_plugins[$idx]}")
    done
  fi

  # Install selected plugins from this repo
  print_blank
  local installed=0
  for plugin in "${selected_plugins[@]}"; do
    print_step "Installing $plugin..."
    if claude plugin install "${plugin}@claude-vibe" &>/dev/null; then
      print_success "$plugin"
      installed=$((installed + 1))
    else
      # Fallback: install from local path directly
      if claude plugin install "${plugin}@${VIBE_HOME}" &>/dev/null; then
        print_success "$plugin (local)"
        installed=$((installed + 1))
      else
        print_error "$plugin — install failed"
      fi
    fi
  done

  print_blank
  print_success "$installed/${#selected_plugins[@]} plugins installed"

  # Install superpowers (external plugin — development workflow skills)
  print_blank
  print_step "Installing superpowers (brainstorming, TDD, code review, planning)..."
  if claude plugin install superpowers@claude-plugins-official &>/dev/null; then
    print_success "superpowers"
  else
    print_warn "superpowers install failed — install manually: claude plugin install superpowers@claude-plugins-official"
  fi

  # Sync permissions
  print_blank
  print_step "Syncing skill permissions..."

  local perms_file="$VIBE_HOME/permissions.yaml"
  if [[ -f "$perms_file" ]]; then
    python3 -c "
import json, os, subprocess

result = subprocess.run(
    ['yq', '-o=json', '$perms_file'],
    capture_output=True, text=True
)
new_perms = json.loads(result.stdout).get('allow', [])

settings_path = os.path.expanduser('~/.claude/settings.json')
if os.path.exists(settings_path):
    with open(settings_path) as f:
        settings = json.load(f)
else:
    settings = {}

existing = settings.get('permissions', {}).get('allow', [])
merged = sorted(set(existing + new_perms))
settings.setdefault('permissions', {})['allow'] = merged

with open(settings_path, 'w') as f:
    json.dump(settings, f, indent=2)

skill_count = len([p for p in merged if p.startswith('Skill(')])
print(f'{len(merged)} total, {skill_count} skills')
" 2>/dev/null

    local result=$?
    if [[ $result -eq 0 ]]; then
      print_success "Permissions synced"
    else
      print_error "Permission sync failed — run manually after install"
    fi
  else
    print_warn "permissions.yaml not found at $perms_file"
  fi

  mark_step_complete "install_plugins"
}
