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
      ((installed++))
    else
      # Fallback: install from local path directly
      if claude plugin install "${plugin}@${VIBE_HOME}" &>/dev/null; then
        print_success "$plugin (local)"
        ((installed++))
      else
        print_error "$plugin — install failed"
      fi
    fi
  done

  print_blank
  print_success "$installed/${#selected_plugins[@]} plugins installed"

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

  # Install security guard hooks
  print_blank
  print_step "Installing security guard hooks..."

  local hooks_dir="$HOME/.claude/hooks"
  mkdir -p "$hooks_dir"

  local hooks_src="$VIBE_HOME/hooks"
  local hooks_installed=0

  for hook_file in "$hooks_src"/guard-*.sh; do
    if [[ -f "$hook_file" ]]; then
      local hook_name
      hook_name=$(basename "$hook_file")
      cp "$hook_file" "$hooks_dir/$hook_name"
      chmod +x "$hooks_dir/$hook_name"
      ((hooks_installed++))
    fi
  done

  if [[ $hooks_installed -gt 0 ]]; then
    # Register hooks in settings.json
    python3 -c "
import json, os

settings_path = os.path.expanduser('~/.claude/settings.json')
if os.path.exists(settings_path):
    with open(settings_path) as f:
        settings = json.load(f)
else:
    settings = {}

hooks = settings.setdefault('hooks', {})
pre_tool_use = hooks.get('PreToolUse', [])

# Define guard hooks
guard_hooks = [
    {
        'matcher': 'Bash',
        'hooks': [{
            'type': 'command',
            'command': '~/.claude/hooks/guard-bash.sh',
            'timeout': 5
        }]
    },
    {
        'matcher': 'Edit|Write',
        'hooks': [{
            'type': 'command',
            'command': '~/.claude/hooks/guard-files.sh',
            'timeout': 5
        }]
    }
]

# Add guards if not already present (check by command path)
existing_commands = set()
for entry in pre_tool_use:
    for h in entry.get('hooks', []):
        existing_commands.add(h.get('command', ''))

for guard in guard_hooks:
    cmd = guard['hooks'][0]['command']
    if cmd not in existing_commands:
        pre_tool_use.append(guard)

hooks['PreToolUse'] = pre_tool_use
settings['hooks'] = hooks

with open(settings_path, 'w') as f:
    json.dump(settings, f, indent=2)

print(f'{len(pre_tool_use)} PreToolUse hooks active')
" 2>/dev/null

    if [[ $? -eq 0 ]]; then
      print_success "$hooks_installed security hooks installed"
    else
      print_error "Hook registration failed"
    fi
  else
    print_warn "No guard hooks found in $hooks_src"
  fi

  mark_step_complete "install_plugins"
}
