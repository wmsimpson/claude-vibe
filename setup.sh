#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Claude Vibe — Local AI Development Environment Setup
# ============================================================================
#
# Usage:
#   ./setup.sh             Full guided setup
#   ./setup.sh --step N    Run a specific step (1-8)
#   ./setup.sh --resume    Resume from last completed step
#   ./setup.sh --validate  Run validation only (Step 8)
#   ./setup.sh --help      Show this help
#
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source helpers and step scripts
source "$SCRIPT_DIR/lib/tty.sh"
source "$SCRIPT_DIR/steps/01-install-claude.sh"
source "$SCRIPT_DIR/steps/02-google-oauth.sh"
source "$SCRIPT_DIR/steps/03-enable-apis.sh"
source "$SCRIPT_DIR/steps/04-install-tools.sh"
source "$SCRIPT_DIR/steps/05-ai-dev-kit.sh"
source "$SCRIPT_DIR/steps/06-install-plugins.sh"
source "$SCRIPT_DIR/steps/07-configure-mcp.sh"
source "$SCRIPT_DIR/steps/08-validate.sh"

# Step registry
STEPS=(
  "install_claude|Step 1: Install Claude Code"
  "google_oauth|Step 2: Google OAuth Setup"
  "enable_apis|Step 3: Enable Google APIs"
  "install_tools|Step 4: Install Core Tools"
  "ai_dev_kit|Step 5: Databricks AI Dev Kit"
  "install_plugins|Step 6: Claude Code Plugins"
  "configure_mcp|Step 7: MCP Integrations"
  "validate|Step 8: Full Validation"
)

STEP_FUNCTIONS=(
  step_install_claude
  step_google_oauth
  step_enable_apis
  step_install_tools
  step_ai_dev_kit
  step_install_plugins
  step_configure_mcp
  step_validate
)

# ── Banner ──────────────────────────────────────────────────────────────────

show_banner() {
  echo ""
  echo -e "${CYAN}${BOLD}"
  cat << 'BANNER'
     _____ _                 _      __     _____ _
    / ____| |               | |     \ \   / /_ _| |__   ___
   | |    | | __ _ _   _  __| | ___  \ \ / / | || '_ \ / _ \
   | |    | |/ _` | | | |/ _` |/ _ \  \ V /  | || |_) |  __/
   | |____| | (_| | |_| | (_| |  __/   | |  |___|_.__/ \___|
    \_____|_|\__,_|\__,_|\__,_|\___|   |_|
BANNER
  echo -e "${NC}"
  echo -e "  ${DIM}Local AI Development Environment${NC}"
  echo -e "  ${DIM}────────────────────────────────────────${NC}"
  echo ""
}

# ── Step overview ───────────────────────────────────────────────────────────

show_overview() {
  echo -e "  ${BOLD}Setup Overview${NC}"
  echo ""
  for i in "${!STEPS[@]}"; do
    local step_id="${STEPS[$i]%%|*}"
    local step_name="${STEPS[$i]##*|}"
    local num=$((i + 1))
    if is_step_complete "$step_id"; then
      echo -e "    ${CHECK} ${DIM}${step_name}${NC}"
    else
      echo -e "    ${BULLET} ${step_name}"
    fi
  done
  echo ""
}

# ── Run steps ───────────────────────────────────────────────────────────────

run_step() {
  local step_num=$1
  local step_idx=$((step_num - 1))
  local step_id="${STEPS[$step_idx]%%|*}"

  show_progress "$step_num"

  if is_step_complete "$step_id"; then
    local step_name="${STEPS[$step_idx]##*|}"
    print_success "${step_name} — already complete"
    if ! ask_yes_no "Run again?" "n"; then
      return 0
    fi
  fi

  ${STEP_FUNCTIONS[$step_idx]}
}

run_all() {
  local start=${1:-1}
  for i in $(seq "$start" ${#STEPS[@]}); do
    run_step "$i"
    local step_idx=$((i - 1))
    local step_id="${STEPS[$step_idx]%%|*}"
    if ! is_step_complete "$step_id" && [[ "$step_id" != "validate" ]]; then
      print_blank
      if ! ask_yes_no "Step did not complete. Continue to next step?" "y"; then
        print_info "Run './setup.sh --resume' to continue later"
        return 1
      fi
    fi
  done
}

find_resume_point() {
  for i in "${!STEPS[@]}"; do
    local step_id="${STEPS[$i]%%|*}"
    if ! is_step_complete "$step_id"; then
      echo $((i + 1))
      return
    fi
  done
  echo 1
}

# ── Main ────────────────────────────────────────────────────────────────────

main() {
  require_macos
  init_state

  case "${1:-}" in
    --help|-h)
      echo "Usage: ./setup.sh [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --step N      Run a specific step (1-8)"
      echo "  --resume      Resume from last completed step"
      echo "  --validate    Run validation only (Step 8)"
      echo "  --reset       Reset install state and start over"
      echo "  --help        Show this help"
      echo ""
      echo "Steps:"
      for i in "${!STEPS[@]}"; do
        echo "  $((i + 1)). ${STEPS[$i]##*|}"
      done
      ;;
    --step)
      local step="${2:-}"
      if [[ -z "$step" || "$step" -lt 1 || "$step" -gt 8 ]]; then
        echo "Usage: ./setup.sh --step N (1-8)"
        exit 1
      fi
      show_banner
      run_step "$step"
      ;;
    --resume)
      show_banner
      show_overview
      local resume_point
      resume_point=$(find_resume_point)
      print_info "Resuming from step $resume_point"
      print_blank
      run_all "$resume_point"
      ;;
    --validate)
      show_banner
      step_validate
      ;;
    --reset)
      rm -f "$HOME/.claude-vibe/install-state.json"
      echo "Install state reset."
      ;;
    *)
      show_banner
      show_overview
      if ask_yes_no "Start full setup?" "y"; then
        run_all
      else
        print_info "Run './setup.sh --help' for options"
      fi
      ;;
  esac
}

main "$@"
