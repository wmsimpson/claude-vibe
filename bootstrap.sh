#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Claude Vibe — Bootstrap
# ============================================================================
#
# Checks prerequisites, installs the vibe CLI, optionally builds the Go
# binary, and launches the interactive installer.
#
# Usage:
#   git clone https://github.com/wmsimpson/claude-vibe.git
#   cd claude-vibe
#   ./bootstrap.sh
#
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── Colors ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

ok()   { echo -e "  ${GREEN}✓${NC} $1"; }
fail() { echo -e "  ${RED}✗${NC} $1"; }
warn() { echo -e "  ${YELLOW}!${NC} $1"; }
info() { echo -e "  ${DIM}$1${NC}"; }

# ── Banner ──────────────────────────────────────────────────────────────────
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

# ── Detect platform ────────────────────────────────────────────────────────
OS="$(uname -s)"
case "$OS" in
  Darwin) PLATFORM="macos" ;;
  Linux)  PLATFORM="linux" ;;
  *)
    fail "Unsupported platform: $OS"
    echo ""
    echo "  Claude Vibe supports macOS and Linux (including WSL)."
    echo "  On Windows, install WSL first:"
    echo "    powershell: wsl --install"
    echo "  Then run this script inside the Ubuntu terminal."
    exit 1
    ;;
esac

IS_WSL=false
if [[ "$PLATFORM" == "linux" ]] && grep -qi microsoft /proc/version 2>/dev/null; then
  IS_WSL=true
fi

echo -e "${BOLD}Checking prerequisites...${NC}"
echo ""

ERRORS=0

# ── Check: Homebrew ─────────────────────────────────────────────────────────
if command -v brew &>/dev/null; then
  ok "Homebrew $(brew --version 2>&1 | head -1 | awk '{print $2}')"
else
  fail "Homebrew not found"
  echo ""
  echo "  Install Homebrew first:"
  echo ""
  echo "    /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
  echo ""
  if [[ "$PLATFORM" == "linux" ]]; then
    echo "  After installing, add it to your PATH:"
    echo ""
    echo "    echo 'eval \"\$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)\"' >> ~/.bashrc"
    echo "    eval \"\$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)\""
    echo ""
  fi
  ERRORS=$((ERRORS + 1))
fi

# ── Check: Git ──────────────────────────────────────────────────────────────
if command -v git &>/dev/null; then
  ok "Git $(git --version | awk '{print $3}')"
else
  fail "Git not found"
  echo ""
  if [[ "$PLATFORM" == "macos" ]]; then
    echo "  Install with: xcode-select --install"
    echo "  Or:           brew install git"
  else
    echo "  Install with: brew install git"
  fi
  echo ""
  ERRORS=$((ERRORS + 1))
fi

# ── Check: curl ─────────────────────────────────────────────────────────────
if command -v curl &>/dev/null; then
  ok "curl"
else
  fail "curl not found"
  echo "  Install with: brew install curl"
  ERRORS=$((ERRORS + 1))
fi

# ── Info: Platform ──────────────────────────────────────────────────────────
if [[ "$PLATFORM" == "macos" ]]; then
  ARCH="$(uname -m)"
  ok "macOS ($ARCH)"
elif $IS_WSL; then
  ok "Linux (WSL)"
else
  ok "Linux"
fi

# ── Info: Go (optional, for building the CLI) ───────────────────────────────
HAS_GO=false
if command -v go &>/dev/null; then
  GO_VER="$(go version | awk '{print $3}' | sed 's/go//')"
  ok "Go $GO_VER (will build the Go CLI)"
  HAS_GO=true
else
  warn "Go not installed — will be installed by vibe install (Step 4)"
  info "The Go CLI adds TUI configuration, doctor diagnostics, and self-update"
fi

# ── Info: Node (optional, installed by setup) ───────────────────────────────
if command -v node &>/dev/null; then
  ok "Node.js $(node --version)"
else
  warn "Node.js not installed — will be installed by vibe install (Step 4)"
fi

# ── Bail if prerequisites missing ───────────────────────────────────────────
echo ""
if [[ $ERRORS -gt 0 ]]; then
  fail "Missing $ERRORS required prerequisite(s). Install them and re-run ./bootstrap.sh"
  exit 1
fi

echo -e "${GREEN}${BOLD}All prerequisites met.${NC}"
echo ""

# ── Install the vibe shell command ──────────────────────────────────────────
TARGET_DIR="$HOME/.local/bin"
TARGET="$TARGET_DIR/vibe"
SOURCE="$SCRIPT_DIR/bin/vibe"

mkdir -p "$TARGET_DIR"
chmod +x "$SOURCE"

# ── Build Go CLI if Go is available ─────────────────────────────────────────
if $HAS_GO; then
  echo -e "${BOLD}Building Go CLI...${NC}"
  if (cd "$SCRIPT_DIR/cli" && go build -o "$TARGET" ./cmd/vibe/) 2>/dev/null; then
    ok "Built Go CLI at $TARGET"
  else
    warn "Go build failed — falling back to shell CLI"
    ln -sf "$SOURCE" "$TARGET"
    ok "Installed shell CLI at $TARGET"
  fi
else
  ln -sf "$SOURCE" "$TARGET"
  ok "Installed shell CLI at $TARGET"
  info "Run 'vibe install' to install Go, then re-run ./bootstrap.sh to build the Go CLI"
fi

# ── Ensure PATH includes ~/.local/bin ───────────────────────────────────────
if ! echo "$PATH" | tr ':' '\n' | grep -qx "$TARGET_DIR"; then
  export PATH="$TARGET_DIR:$PATH"

  SHELL_RC=""
  if [[ -n "${ZSH_VERSION:-}" ]] || [[ "$(basename "$SHELL")" == "zsh" ]]; then
    SHELL_RC="$HOME/.zprofile"
  else
    SHELL_RC="$HOME/.bashrc"
  fi

  if [[ -n "$SHELL_RC" ]] && ! grep -q 'HOME/.local/bin' "$SHELL_RC" 2>/dev/null; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_RC"
    ok "Added ~/.local/bin to PATH in $(basename "$SHELL_RC")"
  fi
fi

# ── Done ────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}Ready.${NC} You can now use ${CYAN}vibe${NC} from anywhere:"
echo ""
echo -e "    ${CYAN}vibe install${NC}              Full guided setup"
echo -e "    ${CYAN}vibe doctor${NC}               Diagnose issues"
echo -e "    ${CYAN}vibe profile list${NC}          Manage profiles"
echo -e "    ${CYAN}vibe agent${NC}                Launch Claude Code"
echo -e "    ${CYAN}vibe help${NC}                 Show all commands"
echo ""

read -rp "Run vibe install now? [Y/n] " response
response="${response:-y}"
if [[ "$response" =~ ^[Yy] ]]; then
  exec "$TARGET" install "$@"
else
  echo ""
  info "Run 'vibe install' when you're ready."
fi
