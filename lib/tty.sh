#!/usr/bin/env bash
# TTY UI helpers for Claude Vibe installer

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m' # No Color

# Symbols
CHECK="${GREEN}✓${NC}"
CROSS="${RED}✗${NC}"
ARROW="${CYAN}→${NC}"
WARN="${YELLOW}!${NC}"
BULLET="${DIM}•${NC}"

# State file for tracking progress
STATE_FILE="$HOME/.claude-vibe/install-state.json"

# ── Output helpers ──────────────────────────────────────────────────────────

print_header() {
  local width=60
  echo ""
  echo -e "${BOLD}${CYAN}"
  printf '  %.0s─' $(seq 1 $width)
  echo ""
  echo "  $1"
  printf '  %.0s─' $(seq 1 $width)
  echo -e "${NC}"
  echo ""
}

print_step() {
  echo -e "  ${ARROW} $1"
}

print_success() {
  echo -e "  ${CHECK} $1"
}

print_error() {
  echo -e "  ${CROSS} $1"
}

print_warn() {
  echo -e "  ${WARN} $1"
}

print_info() {
  echo -e "  ${BULLET} $1"
}

print_blank() {
  echo ""
}

# ── Progress bar ────────────────────────────────────────────────────────────

TOTAL_STEPS=8
show_progress() {
  local current=$1
  local total=${2:-$TOTAL_STEPS}
  local width=40
  local filled=$((current * width / total))
  local empty=$((width - filled))
  local pct=$((current * 100 / total))

  printf "\r  ${DIM}[${NC}"
  printf "${GREEN}%0.s█${NC}" $(seq 1 $filled 2>/dev/null)
  printf "${DIM}%0.s░${NC}" $(seq 1 $empty 2>/dev/null)
  printf "${DIM}]${NC} ${BOLD}%3d%%${NC} " "$pct"
  echo ""
}

# ── User prompts ────────────────────────────────────────────────────────────

ask_yes_no() {
  local prompt="$1"
  local default="${2:-y}"
  local hint="[Y/n]"
  [[ "$default" == "n" ]] && hint="[y/N]"

  while true; do
    echo -ne "  ${BOLD}?${NC} ${prompt} ${DIM}${hint}${NC}: "
    read -r answer </dev/tty
    answer="${answer:-$default}"
    case "$answer" in
      [Yy]*) return 0 ;;
      [Nn]*) return 1 ;;
      *) echo -e "  ${DIM}Please answer y or n${NC}" ;;
    esac
  done
}

ask_input() {
  local prompt="$1"
  local default="$2"
  local hint=""
  [[ -n "$default" ]] && hint=" ${DIM}(${default})${NC}"

  echo -ne "  ${BOLD}?${NC} ${prompt}${hint}: "
  read -r answer </dev/tty
  echo "${answer:-$default}"
}

ask_secret() {
  local prompt="$1"
  echo -ne "  ${BOLD}?${NC} ${prompt}: "
  read -rs answer </dev/tty
  echo ""
  echo "$answer"
}

# Arrow-key selector
ask_select() {
  local prompt="$1"
  shift
  local options=("$@")
  local selected=0
  local count=${#options[@]}

  echo -e "  ${BOLD}?${NC} ${prompt} ${DIM}(↑↓ to select, enter to confirm)${NC}"

  # Hide cursor
  tput civis 2>/dev/null

  while true; do
    # Draw options
    for i in "${!options[@]}"; do
      if [[ $i -eq $selected ]]; then
        echo -e "    ${CYAN}❯ ${options[$i]}${NC}"
      else
        echo -e "    ${DIM}  ${options[$i]}${NC}"
      fi
    done

    # Read keypress
    IFS= read -rsn1 key </dev/tty
    if [[ "$key" == $'\x1b' ]]; then
      read -rsn2 key </dev/tty
      case "$key" in
        '[A') ((selected > 0)) && ((selected--)) ;;       # Up
        '[B') ((selected < count - 1)) && ((selected++)) ;; # Down
      esac
    elif [[ "$key" == "" ]]; then
      break
    fi

    # Move cursor up to redraw
    tput cuu "$count" 2>/dev/null
  done

  # Show cursor
  tput cnorm 2>/dev/null

  echo "${options[$selected]}"
  return $selected
}

# Multi-select with checkboxes
ask_multi_select() {
  local prompt="$1"
  shift
  local options=("$@")
  local count=${#options[@]}
  local selected=0
  declare -a checked
  for i in "${!options[@]}"; do checked[$i]=1; done  # All checked by default

  echo -e "  ${BOLD}?${NC} ${prompt} ${DIM}(↑↓ move, space toggle, enter confirm)${NC}"

  tput civis 2>/dev/null

  while true; do
    for i in "${!options[@]}"; do
      local marker="${GREEN}◉${NC}"
      [[ ${checked[$i]} -eq 0 ]] && marker="${DIM}○${NC}"
      if [[ $i -eq $selected ]]; then
        echo -e "    ${CYAN}❯${NC} ${marker} ${options[$i]}"
      else
        echo -e "      ${marker} ${DIM}${options[$i]}${NC}"
      fi
    done

    IFS= read -rsn1 key </dev/tty
    if [[ "$key" == $'\x1b' ]]; then
      read -rsn2 key </dev/tty
      case "$key" in
        '[A') ((selected > 0)) && ((selected--)) ;;
        '[B') ((selected < count - 1)) && ((selected++)) ;;
      esac
    elif [[ "$key" == " " ]]; then
      checked[$selected]=$(( 1 - ${checked[$selected]} ))
    elif [[ "$key" == "" ]]; then
      break
    fi

    tput cuu "$count" 2>/dev/null
  done

  tput cnorm 2>/dev/null

  # Return selected indices
  SELECTED_INDICES=()
  for i in "${!options[@]}"; do
    [[ ${checked[$i]} -eq 1 ]] && SELECTED_INDICES+=("$i")
  done
}

# ── Spinner ─────────────────────────────────────────────────────────────────

spin() {
  local pid=$1
  local msg="$2"
  local spinchars='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
  local i=0

  tput civis 2>/dev/null
  while kill -0 "$pid" 2>/dev/null; do
    printf "\r  ${CYAN}%s${NC} %s" "${spinchars:i++%${#spinchars}:1}" "$msg"
    sleep 0.1
  done
  tput cnorm 2>/dev/null
  printf "\r"
}

run_with_spinner() {
  local msg="$1"
  shift
  "$@" &>/dev/null &
  local pid=$!
  spin $pid "$msg"
  wait $pid
  return $?
}

# ── State management ───────────────────────────────────────────────────────

init_state() {
  mkdir -p "$(dirname "$STATE_FILE")"
  if [[ ! -f "$STATE_FILE" ]]; then
    echo '{"completed_steps":[],"started_at":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}' > "$STATE_FILE"
  fi
}

mark_step_complete() {
  local step="$1"
  python3 -c "
import json, os
p = os.path.expanduser('$STATE_FILE')
with open(p) as f: s = json.load(f)
if '$step' not in s['completed_steps']:
    s['completed_steps'].append('$step')
with open(p, 'w') as f: json.dump(s, f, indent=2)
" 2>/dev/null
}

is_step_complete() {
  local step="$1"
  python3 -c "
import json, os
p = os.path.expanduser('$STATE_FILE')
with open(p) as f: s = json.load(f)
exit(0 if '$step' in s['completed_steps'] else 1)
" 2>/dev/null
}

# ── Requirement checks ──────────────────────────────────────────────────────

require_command() {
  local cmd="$1"
  local install_hint="$2"
  if ! command -v "$cmd" &>/dev/null; then
    print_error "$cmd not found"
    [[ -n "$install_hint" ]] && print_info "Install: $install_hint"
    return 1
  fi
  return 0
}

require_macos() {
  if [[ "$(uname)" != "Darwin" ]]; then
    print_error "This installer requires macOS"
    exit 1
  fi
}
