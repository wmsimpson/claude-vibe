#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Claude Vibe — Bootstrap
# ============================================================================
#
# This script installs the `vibe` command and runs `vibe install`.
#
# Quick start:
#   git clone https://github.com/wmsimpson/claude-vibe.git
#   cd claude-vibe
#   ./setup.sh
#
# After setup, use `vibe` directly:
#   vibe install, vibe validate, vibe status, vibe help
#
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source "$SCRIPT_DIR/lib/tty.sh"

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

# Install the vibe command
TARGET_DIR="$HOME/.local/bin"
TARGET="$TARGET_DIR/vibe"
SOURCE="$SCRIPT_DIR/bin/vibe"

mkdir -p "$TARGET_DIR"
chmod +x "$SOURCE"
ln -sf "$SOURCE" "$TARGET"

# Ensure PATH includes ~/.local/bin
if ! echo "$PATH" | tr ':' '\n' | grep -qx "$TARGET_DIR"; then
  export PATH="$TARGET_DIR:$PATH"
  if ! grep -q 'HOME/.local/bin' "$HOME/.zprofile" 2>/dev/null; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zprofile"
  fi
fi

print_success "Installed ${BOLD}vibe${NC} command at $TARGET"
print_blank
print_info "You can now use ${BOLD}vibe${NC} from anywhere:"
echo ""
echo -e "    ${CYAN}vibe install${NC}              Full guided setup"
echo -e "    ${CYAN}vibe install --step N${NC}     Run a specific step"
echo -e "    ${CYAN}vibe install --resume${NC}     Resume from last step"
echo -e "    ${CYAN}vibe validate${NC}             Run validation suite"
echo -e "    ${CYAN}vibe status${NC}               Show progress"
echo -e "    ${CYAN}vibe doctor${NC}               Diagnose issues"
echo -e "    ${CYAN}vibe help${NC}                 Show all commands"
echo ""

if ask_yes_no "Run ${BOLD}vibe install${NC} now?" "y"; then
  exec "$TARGET" install "$@"
else
  print_info "Run ${BOLD}vibe install${NC} when you're ready"
fi
