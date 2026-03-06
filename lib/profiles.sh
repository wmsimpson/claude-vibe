#!/usr/bin/env bash
# Profile management for Claude Vibe
#
# Each profile stores per-identity configs:
#   ~/.claude-vibe/profiles/<name>/
#     claude.json         - MCP server configs (tokens, endpoints)
#     env                 - Environment variables (API keys, tokens)
#     gcloud-adc.json     - Google Application Default Credentials
#     gcloud-oauth.json   - Google OAuth client credentials
#     gcp-project-id      - GCP quota project
#     profile.json        - Profile metadata (email, description)
#
# Shared across all profiles (not swapped):
#   ~/.claude/settings.json   - Plugins and permissions
#   ~/.claude/skills/         - Skills
#   ~/.claude/plugins/        - Plugin cache
#   Installed CLI tools

PROFILES_DIR="$HOME/.claude-vibe/profiles"
ACTIVE_PROFILE_FILE="$HOME/.claude-vibe/active-profile"

# Files that are per-profile
PROFILE_FILES=(
  "$HOME/.claude.json|claude.json"
  "$HOME/.vibe/env|env"
  "$HOME/.config/gcloud/application_default_credentials.json|gcloud-adc.json"
  "$HOME/.config/gcloud/credentials/claude-google-auth.json|gcloud-oauth.json"
  "$HOME/.claude-vibe/gcp-project-id|gcp-project-id"
)

# ── Helpers ─────────────────────────────────────────────────────────────────

_get_active_profile() {
  if [[ -f "$ACTIVE_PROFILE_FILE" ]]; then
    cat "$ACTIVE_PROFILE_FILE"
  else
    echo ""
  fi
}

_set_active_profile() {
  echo "$1" > "$ACTIVE_PROFILE_FILE"
}

_profile_exists() {
  [[ -d "$PROFILES_DIR/$1" ]]
}

_list_profiles() {
  if [[ -d "$PROFILES_DIR" ]]; then
    ls "$PROFILES_DIR" 2>/dev/null
  fi
}

# Save current live config files into a profile directory
_save_to_profile() {
  local name="$1"
  local profile_dir="$PROFILES_DIR/$name"
  mkdir -p "$profile_dir"

  for entry in "${PROFILE_FILES[@]}"; do
    local src="${entry%%|*}"
    local dest="${entry##*|}"
    if [[ -f "$src" ]]; then
      cp "$src" "$profile_dir/$dest"
    fi
  done
}

# Restore a profile's config files to the live locations
_restore_from_profile() {
  local name="$1"
  local profile_dir="$PROFILES_DIR/$name"

  if [[ ! -d "$profile_dir" ]]; then
    print_error "Profile '$name' not found"
    return 1
  fi

  for entry in "${PROFILE_FILES[@]}"; do
    local dest="${entry%%|*}"
    local src="${entry##*|}"
    local src_path="$profile_dir/$src"
    if [[ -f "$src_path" ]]; then
      mkdir -p "$(dirname "$dest")"
      cp "$src_path" "$dest"
    fi
  done
}

# ── Commands ────────────────────────────────────────────────────────────────

profile_list() {
  print_header "Profiles"

  local active
  active=$(_get_active_profile)
  local profiles
  profiles=$(_list_profiles)

  if [[ -z "$profiles" ]]; then
    print_info "No profiles configured yet"
    print_info "Save your current setup: ${BOLD}vibe profile save personal${NC}"
    return 0
  fi

  while IFS= read -r name; do
    local marker="  "
    local meta=""
    if [[ "$name" == "$active" ]]; then
      marker="${GREEN}▸${NC}"
    else
      marker=" "
    fi

    # Read profile metadata
    local meta_file="$PROFILES_DIR/$name/profile.json"
    if [[ -f "$meta_file" ]]; then
      local email desc
      email=$(python3 -c "import json; d=json.load(open('$meta_file')); print(d.get('email',''))" 2>/dev/null)
      desc=$(python3 -c "import json; d=json.load(open('$meta_file')); print(d.get('description',''))" 2>/dev/null)
      meta="${DIM}${email}${NC}"
      [[ -n "$desc" ]] && meta="$meta ${DIM}— ${desc}${NC}"
    fi

    if [[ "$name" == "$active" ]]; then
      echo -e "  ${marker} ${BOLD}${CYAN}${name}${NC} ${meta} ${DIM}(active)${NC}"
    else
      echo -e "  ${marker} ${name} ${meta}"
    fi
  done <<< "$profiles"

  print_blank
}

profile_current() {
  local active
  active=$(_get_active_profile)
  if [[ -n "$active" ]]; then
    echo -e "  Active profile: ${BOLD}${CYAN}${active}${NC}"
    local meta_file="$PROFILES_DIR/$active/profile.json"
    if [[ -f "$meta_file" ]]; then
      local email
      email=$(python3 -c "import json; d=json.load(open('$meta_file')); print(d.get('email',''))" 2>/dev/null)
      [[ -n "$email" ]] && echo -e "  Email: ${email}"
    fi
  else
    echo -e "  ${DIM}No active profile — run 'vibe profile save <name>' to create one${NC}"
  fi
}

profile_save() {
  local name="${1:-}"
  if [[ -z "$name" ]]; then
    name=$(ask_input "Profile name" "personal")
  fi

  # Validate name (alphanumeric, dashes, underscores)
  if [[ ! "$name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
    print_error "Profile name must be alphanumeric (dashes and underscores OK)"
    return 1
  fi

  if _profile_exists "$name"; then
    if ! ask_yes_no "Profile '$name' exists. Overwrite?" "n"; then
      return 0
    fi
  fi

  # Collect metadata
  local email desc
  email=$(ask_input "Email for this profile" "")
  desc=$(ask_input "Description (optional)" "")

  print_step "Saving current config to profile '${BOLD}$name${NC}'..."

  _save_to_profile "$name"

  # Write metadata
  python3 -c "
import json
meta = {
    'name': '$name',
    'email': '$email',
    'description': '$desc'
}
with open('$PROFILES_DIR/$name/profile.json', 'w') as f:
    json.dump(meta, f, indent=2)
" 2>/dev/null

  _set_active_profile "$name"

  print_success "Profile '${BOLD}$name${NC}' saved and set as active"

  # Show what was saved
  local saved_files=0
  for entry in "${PROFILE_FILES[@]}"; do
    local dest="${entry##*|}"
    [[ -f "$PROFILES_DIR/$name/$dest" ]] && ((saved_files++))
  done
  print_info "$saved_files config files saved"
}

profile_create() {
  local name="${1:-}"
  if [[ -z "$name" ]]; then
    name=$(ask_input "Profile name")
  fi

  if [[ ! "$name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
    print_error "Profile name must be alphanumeric (dashes and underscores OK)"
    return 1
  fi

  if _profile_exists "$name"; then
    print_error "Profile '$name' already exists"
    print_info "Use 'vibe profile save $name' to overwrite"
    return 1
  fi

  local email desc
  email=$(ask_input "Email for this profile" "")
  desc=$(ask_input "Description (optional)" "")

  local profile_dir="$PROFILES_DIR/$name"
  mkdir -p "$profile_dir"

  # Write metadata
  python3 -c "
import json
meta = {
    'name': '$name',
    'email': '$email',
    'description': '$desc'
}
with open('$profile_dir/profile.json', 'w') as f:
    json.dump(meta, f, indent=2)
" 2>/dev/null

  # Create empty env file
  cat > "$profile_dir/env" << ENVEOF
# ~/.vibe/env for profile: $name ($email)
export GCP_QUOTA_PROJECT=your-gcp-project-id

# Add your tokens for this profile below
# export GITHUB_PERSONAL_ACCESS_TOKEN=ghp_...
# export SLACK_BOT_TOKEN=xoxb-...
ENVEOF

  # Create empty claude.json
  echo '{"mcpServers":{}}' > "$profile_dir/claude.json"

  print_success "Profile '${BOLD}$name${NC}' created"
  print_blank
  print_info "Next steps:"
  echo -e "    ${CYAN}vibe profile switch $name${NC}     Switch to this profile"
  echo -e "    ${CYAN}vibe install${NC}                   Run setup for this profile"
  echo -e "    ${CYAN}vibe profile save $name${NC}        Save after setup"
}

profile_switch() {
  local name="${1:-}"

  if [[ -z "$name" ]]; then
    # Interactive selection
    local profiles
    profiles=$(_list_profiles)
    if [[ -z "$profiles" ]]; then
      print_error "No profiles found — create one with 'vibe profile save <name>'"
      return 1
    fi

    local profile_array=()
    while IFS= read -r line; do
      profile_array+=("$line")
    done <<< "$profiles"

    local selected
    selected=$(ask_select "Switch to profile:" "${profile_array[@]}")
    name="${profile_array[$?]}"
  fi

  if ! _profile_exists "$name"; then
    print_error "Profile '$name' not found"
    print_info "Available profiles:"
    _list_profiles | while read -r p; do echo "  $p"; done
    return 1
  fi

  local current
  current=$(_get_active_profile)

  # Auto-save current profile before switching
  if [[ -n "$current" ]] && _profile_exists "$current"; then
    print_step "Saving current profile '${BOLD}$current${NC}'..."
    _save_to_profile "$current"
    print_success "Saved '$current'"
  fi

  # Restore target profile
  print_step "Switching to profile '${BOLD}$name${NC}'..."
  _restore_from_profile "$name"
  _set_active_profile "$name"

  print_success "Active profile: ${BOLD}${CYAN}$name${NC}"

  local meta_file="$PROFILES_DIR/$name/profile.json"
  if [[ -f "$meta_file" ]]; then
    local email
    email=$(python3 -c "import json; d=json.load(open('$meta_file')); print(d.get('email',''))" 2>/dev/null)
    [[ -n "$email" ]] && print_info "Email: $email"
  fi

  print_blank
  print_warn "Restart Claude Code to apply the new profile's MCP configs"
}

profile_delete() {
  local name="${1:-}"
  if [[ -z "$name" ]]; then
    print_error "Usage: vibe profile delete <name>"
    return 1
  fi

  if ! _profile_exists "$name"; then
    print_error "Profile '$name' not found"
    return 1
  fi

  local active
  active=$(_get_active_profile)
  if [[ "$name" == "$active" ]]; then
    print_error "Cannot delete the active profile"
    print_info "Switch to another profile first: vibe profile switch <other>"
    return 1
  fi

  if ask_yes_no "Delete profile '$name'? This cannot be undone" "n"; then
    rm -rf "$PROFILES_DIR/$name"
    print_success "Profile '$name' deleted"
  fi
}

# ── Profile command router ──────────────────────────────────────────────────

cmd_profile() {
  local subcmd="${1:-}"
  shift 2>/dev/null || true

  case "$subcmd" in
    list|ls)    profile_list ;;
    current)    profile_current ;;
    save)       profile_save "$@" ;;
    create|new) profile_create "$@" ;;
    switch|use) profile_switch "$@" ;;
    delete|rm)  profile_delete "$@" ;;
    "")
      profile_current
      echo ""
      profile_list
      ;;
    *)
      print_error "Unknown profile command: $subcmd"
      echo ""
      echo -e "  ${BOLD}Usage:${NC} vibe profile <command>"
      echo ""
      echo "  list              List all profiles"
      echo "  current           Show active profile"
      echo "  save [name]       Save current config as a profile"
      echo "  create <name>     Create a new empty profile"
      echo "  switch [name]     Switch to a different profile"
      echo "  delete <name>     Delete a profile"
      exit 1
      ;;
  esac
}
