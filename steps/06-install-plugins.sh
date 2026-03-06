#!/usr/bin/env bash
# Step 6: Install Claude Code Plugins

PLUGIN_REPO_URL="https://github.com/wsimpsonjr/individual-vibe-tool"
PLUGIN_REPO_DIR="$HOME/individual-vibe-tool"

step_install_plugins() {
  print_header "Step 6 of 8 — Claude Code Plugins"

  print_info "Installing plugin collections for Google Workspace, app dev, workflows, and more."
  print_info "Source: $PLUGIN_REPO_URL"
  print_blank

  # Clone or update plugin repo
  if [[ -d "$PLUGIN_REPO_DIR/plugins" ]]; then
    print_success "Plugin repo found at $PLUGIN_REPO_DIR"
  else
    print_step "Cloning plugin repo..."
    if run_with_spinner "Cloning individual-vibe-tool..." git clone "$PLUGIN_REPO_URL" "$PLUGIN_REPO_DIR"; then
      print_success "Plugin repo cloned"
    else
      print_error "Failed to clone plugin repo"
      print_info "Clone manually: git clone $PLUGIN_REPO_URL $PLUGIN_REPO_DIR"
      return 1
    fi
  fi

  # Select which plugins to install
  local all_plugins=(
    "fe-google-tools"
    "fe-app-dev"
    "fe-workflows"
    "fe-vibe-setup"
    "fe-specialized-agents"
    "fe-macos-scheduler"
    "lean-sigma-tools"
    "fe-jira-tools"
    "fe-databricks-tools"
    "fe-mcp-servers"
  )

  local descriptions=(
    "Google Workspace (Gmail, Docs, Sheets, Slides, Calendar, Forms, Tasks)"
    "App Development (React Native, Expo, Next.js, Swift, Flutter)"
    "Workflows (Architecture diagrams, RCA, POC docs, Security questionnaires)"
    "Setup & Diagnostics (Validate, usage stats, integrations)"
    "Diagram Agents (Lucid Chart, Graphviz)"
    "macOS Scheduler (launchd background tasks)"
    "Lean Six Sigma (FMEA, SIPOC, Process Maps)"
    "JIRA (Search, create, update tickets)"
    "Databricks Tools (Query, deploy, manage workspaces)"
    "MCP Server Configs"
  )

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

  # Install selected plugins
  print_blank
  local installed=0
  for plugin in "${selected_plugins[@]}"; do
    print_step "Installing $plugin..."
    if claude plugin install "${plugin}@individual-vibe-tool" &>/dev/null; then
      print_success "$plugin"
      ((installed++))
    else
      print_error "$plugin — install failed"
    fi
  done

  print_blank
  print_success "$installed/${#selected_plugins[@]} plugins installed"

  # Sync permissions
  print_blank
  print_step "Syncing skill permissions..."

  if [[ -f "$PLUGIN_REPO_DIR/permissions.yaml" ]]; then
    python3 -c "
import json, os, subprocess

result = subprocess.run(
    ['yq', '-o=json', os.path.expanduser('$PLUGIN_REPO_DIR/permissions.yaml')],
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
    print_warn "permissions.yaml not found — skipping permission sync"
  fi

  mark_step_complete "install_plugins"
}
