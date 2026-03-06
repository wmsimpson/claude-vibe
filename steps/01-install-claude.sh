#!/usr/bin/env bash
# Step 1: Install Claude Code

step_install_claude() {
  print_header "Step 1 of 8 — Install Claude Code"

  if command -v claude &>/dev/null; then
    local ver=$(claude --version 2>&1)
    print_success "Claude Code already installed: ${ver}"
    if ask_yes_no "Skip this step?" "y"; then
      mark_step_complete "install_claude"
      return 0
    fi
  fi

  # Detect AVX support
  print_step "Checking CPU features..."
  local has_avx=false
  if sysctl -a 2>/dev/null | grep machdep.cpu.features | grep -qi avx; then
    has_avx=true
    print_success "CPU supports AVX"
  else
    print_warn "CPU does not support AVX"
  fi

  # Choose install method
  if $has_avx; then
    echo ""
    print_info "Your CPU supports AVX — both install methods will work."
    local method
    method=$(ask_select "Choose install method:" "Homebrew Cask (recommended)" "npm (universal)")
    if [[ $? -eq 0 ]]; then
      print_step "Installing via Homebrew..."
      if run_with_spinner "Installing claude-code cask..." brew install --cask claude-code; then
        print_success "Claude Code installed via Homebrew"
      else
        print_error "Homebrew install failed — falling back to npm"
        _install_claude_npm
      fi
    else
      _install_claude_npm
    fi
  else
    echo ""
    print_info "The Homebrew cask uses Bun which requires AVX."
    print_info "Installing via npm (uses Node.js — works on all CPUs)."
    echo ""
    _install_claude_npm
  fi

  # Verify
  if command -v claude &>/dev/null; then
    print_success "Verified: $(claude --version 2>&1)"
    mark_step_complete "install_claude"
  else
    print_error "claude not found in PATH after install"
    print_info "You may need to restart your terminal or add it to your PATH"
    return 1
  fi
}

_install_claude_npm() {
  # Ensure node is installed
  if ! command -v node &>/dev/null; then
    print_step "Installing Node.js..."
    run_with_spinner "Installing node..." brew install node
  fi

  print_step "Installing Claude Code via npm..."
  if run_with_spinner "npm install -g @anthropic-ai/claude-code..." npm install -g @anthropic-ai/claude-code; then
    # Symlink to /usr/local/bin if not already in PATH
    local npm_bin
    npm_bin="$(npm prefix -g 2>/dev/null)/bin/claude"
    if [[ -f "$npm_bin" ]] && ! command -v claude &>/dev/null; then
      ln -sf "$npm_bin" /usr/local/bin/claude 2>/dev/null
      print_info "Symlinked claude to /usr/local/bin/claude"
    fi
    print_success "Claude Code installed via npm"
  else
    print_error "npm install failed"
    return 1
  fi
}
