#!/usr/bin/env bash
# Step 4: Install Core Tools & Configure Environment

step_install_tools() {
  print_header "Step 4 of 8 — Install Core Tools"

  # ── Homebrew ────────────────────────────────────────────────────────────
  if ! command -v brew &>/dev/null; then
    print_step "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" </dev/tty
  else
    print_success "Homebrew installed"
  fi

  # ── Brew packages ───────────────────────────────────────────────────────
  local brew_tools=(go jq yq ripgrep terminal-notifier graphviz uv node git gh)
  local missing=()

  for tool in "${brew_tools[@]}"; do
    if brew list "$tool" &>/dev/null; then
      print_success "$tool"
    else
      missing+=("$tool")
    fi
  done

  if [[ ${#missing[@]} -gt 0 ]]; then
    print_blank
    print_step "Installing missing tools: ${missing[*]}"
    if run_with_spinner "brew install ${missing[*]}..." brew install "${missing[@]}"; then
      print_success "All Homebrew tools installed"
    else
      print_error "Some tools failed to install"
    fi
  fi

  # ── npm globals ─────────────────────────────────────────────────────────
  print_blank
  print_step "Checking npm global packages..."

  if command -v mmdc &>/dev/null; then
    print_success "mermaid-cli (mmdc)"
  else
    print_step "Installing mermaid-cli..."
    if run_with_spinner "npm install -g @mermaid-js/mermaid-cli..." npm install -g @mermaid-js/mermaid-cli; then
      # Symlink if needed
      local npm_prefix
      npm_prefix="$(npm prefix -g 2>/dev/null)"
      if [[ -f "$npm_prefix/bin/mmdc" ]] && ! command -v mmdc &>/dev/null; then
        ln -sf "$npm_prefix/bin/mmdc" /usr/local/bin/mmdc 2>/dev/null
      fi
      print_success "mermaid-cli installed"
    fi
  fi

  # ── gcloud SDK ──────────────────────────────────────────────────────────
  if command -v gcloud &>/dev/null; then
    print_success "Google Cloud SDK"
  else
    print_step "Installing Google Cloud SDK..."
    if run_with_spinner "brew install google-cloud-sdk..." brew install --cask google-cloud-sdk; then
      print_success "Google Cloud SDK installed"
    else
      print_error "gcloud install failed — install manually from cloud.google.com/sdk"
    fi
  fi

  # ── Databricks CLI (optional) ──────────────────────────────────────────
  if is_databricks_enabled; then
    if command -v databricks &>/dev/null; then
      print_success "Databricks CLI"
    else
      print_step "Installing Databricks CLI..."
      if run_with_spinner "brew install databricks..." brew install databricks; then
        print_success "Databricks CLI installed"
      else
        print_warn "Databricks CLI not installed — Databricks features will be limited"
      fi
    fi
  else
    print_info "Databricks CLI — ${DIM}skipped (not enabled)${NC}"
  fi

  # ── Configure ~/.zprofile ──────────────────────────────────────────────
  print_blank
  print_step "Configuring shell environment..."

  local zprofile="$HOME/.zprofile"
  local needs_update=false

  # Ensure PATH
  if ! grep -q 'HOME/.local/bin' "$zprofile" 2>/dev/null; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$zprofile"
    needs_update=true
  fi

  # Ensure vibe env sourcing
  if ! grep -q '.vibe/env' "$zprofile" 2>/dev/null; then
    echo '[ -f ~/.vibe/env ] && source ~/.vibe/env' >> "$zprofile"
    needs_update=true
  fi

  if $needs_update; then
    print_success "Updated ~/.zprofile"
  else
    print_success "~/.zprofile already configured"
  fi

  # ── Create ~/.vibe/env ─────────────────────────────────────────────────
  mkdir -p "$HOME/.vibe"
  if [[ ! -f "$HOME/.vibe/env" ]]; then
    local gcp_project=""
    [[ -f "$HOME/.claude-vibe/gcp-project-id" ]] && gcp_project=$(cat "$HOME/.claude-vibe/gcp-project-id")

    cat > "$HOME/.vibe/env" << ENVEOF
# ~/.vibe/env — sourced automatically by ~/.zprofile
# Tokens and environment variables for Claude Vibe

export GCP_QUOTA_PROJECT=${gcp_project:-your-gcp-project-id}

# Optional tokens (uncomment and fill in as needed)
# export GITHUB_PERSONAL_ACCESS_TOKEN=ghp_...
# export SLACK_BOT_TOKEN=xoxb-...
# export JIRA_API_TOKEN=...
# export ANTHROPIC_API_KEY=sk-ant-...
ENVEOF
    print_success "Created ~/.vibe/env"
  else
    print_success "~/.vibe/env exists"

    # Update GCP project if we know it
    if [[ -f "$HOME/.claude-vibe/gcp-project-id" ]]; then
      local proj
      proj=$(cat "$HOME/.claude-vibe/gcp-project-id")
      if grep -q "your-gcp-project-id" "$HOME/.vibe/env"; then
        sed -i '' "s/your-gcp-project-id/$proj/" "$HOME/.vibe/env"
        print_info "Updated GCP_QUOTA_PROJECT to $proj"
      fi
    fi
  fi

  source "$HOME/.zprofile" 2>/dev/null
  mark_step_complete "install_tools"
}
