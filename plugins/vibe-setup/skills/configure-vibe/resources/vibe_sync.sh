#!/bin/bash
#
# vibe_sync.sh - Sync vibe config files to ~/.claude/settings.json
#
# This script reads YAML config files from the same directory as this script
# and merges them into ~/.claude/settings.json without overwriting existing values.
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR"
SETTINGS_FILE="$HOME/.claude/settings.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check for required tools
check_dependencies() {
    local missing=()

    if ! command -v jq &> /dev/null; then
        missing+=("jq")
    fi

    if ! command -v yq &> /dev/null; then
        missing+=("yq")
    fi

    if [ ${#missing[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing[*]}"
        echo "Install with: brew install ${missing[*]}"
        exit 1
    fi
}

# Ensure settings file exists
ensure_settings_file() {
    if [ ! -f "$SETTINGS_FILE" ]; then
        log_info "Creating $SETTINGS_FILE"
        mkdir -p "$(dirname "$SETTINGS_FILE")"
        echo '{}' > "$SETTINGS_FILE"
    fi
}

# Merge MCP servers from mcp-servers.yaml
merge_mcp_servers() {
    local config_file="$CONFIG_DIR/mcp-servers.yaml"

    if [ ! -f "$config_file" ]; then
        log_warn "No mcp-servers.yaml found, skipping"
        return
    fi

    log_info "Merging MCP servers..."

    # Convert YAML servers to JSON and merge with existing mcpServers
    local servers_json
    servers_json=$(yq -o=json '.servers // {}' "$config_file")

    # Read current settings and merge
    local current_settings
    current_settings=$(cat "$SETTINGS_FILE")

    # Merge: existing mcpServers + new servers (new servers take precedence for same keys)
    local merged
    merged=$(echo "$current_settings" | jq --argjson new_servers "$servers_json" '
        .mcpServers = ((.mcpServers // {}) * $new_servers)
    ')

    echo "$merged" > "$SETTINGS_FILE"
    log_info "  Merged $(echo "$servers_json" | jq 'keys | length') MCP server(s)"
}

# Merge MCP servers to ~/.config/mcp/config.json
merge_mcp_servers_to_config() {
    local config_file="$CONFIG_DIR/mcp-servers.yaml"
    local mcp_config="$HOME/.config/mcp/config.json"

    if [ ! -f "$config_file" ]; then
        log_warn "No mcp-servers.yaml found, skipping MCP config merge"
        return
    fi

    log_info "Merging MCP servers to .config/mcp/config.json..."

    # Convert YAML servers to JSON, expanding ~ to $HOME, adding enabled and name fields, removing type
    local servers_json
    servers_json=$(yq -o=json '.servers // {}' "$config_file" | jq '
        walk(
            if type == "string" then
                if startswith("~/") then
                    sub("^~/"; env.HOME + "/")
                elif contains("=~/") then
                    gsub("=~/"; "=" + env.HOME + "/")
                else
                    .
                end
            else
                .
            end
        ) |
        to_entries | map({
            key: .key,
            value: ((.value + {enabled: true, name: .key}) | del(.type))
        }) | from_entries
    ')

    # Ensure config file exists
    mkdir -p "$HOME/.config/mcp"
    if [ ! -f "$mcp_config" ]; then
        echo '{"claude-code":{}}' > "$mcp_config"
    fi

    # Merge into claude-code section
    local current_config
    current_config=$(cat "$mcp_config")

    local merged
    merged=$(echo "$current_config" | jq --argjson new_servers "$servers_json" '
        .["claude-code"] = ((.["claude-code"] // {}) * $new_servers)
    ')

    echo "$merged" > "$mcp_config"
    log_info "  Merged $(echo "$servers_json" | jq 'keys | length') MCP server(s) to .config/mcp/config.json"
}

# Merge permissions from permissions.yaml
merge_permissions() {
    local config_file="$CONFIG_DIR/permissions.yaml"

    if [ ! -f "$config_file" ]; then
        log_warn "No permissions.yaml found, skipping"
        return
    fi

    log_info "Merging permissions..."

    # Get allow and deny arrays from YAML
    local allow_json deny_json
    allow_json=$(yq -o=json '.allow // []' "$config_file")
    deny_json=$(yq -o=json '.deny // []' "$config_file")

    # Read current settings
    local current_settings
    current_settings=$(cat "$SETTINGS_FILE")

    # Merge: combine arrays and remove duplicates
    local merged
    merged=$(echo "$current_settings" | jq --argjson new_allow "$allow_json" --argjson new_deny "$deny_json" '
        .permissions.allow = ((.permissions.allow // []) + $new_allow | unique) |
        .permissions.deny = ((.permissions.deny // []) + $new_deny | unique)
    ')

    echo "$merged" > "$SETTINGS_FILE"

    local allow_count deny_count
    allow_count=$(echo "$allow_json" | jq 'length')
    deny_count=$(echo "$deny_json" | jq 'length')
    log_info "  Merged $allow_count allow permission(s), $deny_count deny permission(s)"
}

# Merge hooks from hooks.yaml
merge_hooks() {
    local config_file="$CONFIG_DIR/hooks.yaml"

    if [ ! -f "$config_file" ]; then
        log_warn "No hooks.yaml found, skipping"
        return
    fi

    log_info "Merging hooks..."

    # Get hooks object from YAML
    local hooks_json
    hooks_json=$(yq -o=json '.hooks // {}' "$config_file")

    # Skip if hooks is empty
    if [ "$hooks_json" = "{}" ] || [ "$hooks_json" = "null" ]; then
        log_info "  No hooks defined, skipping"
        return
    fi

    # Read current settings
    local current_settings
    current_settings=$(cat "$SETTINGS_FILE")

    # Merge hooks: for each hook type, combine arrays
    local merged
    merged=$(echo "$current_settings" | jq --argjson new_hooks "$hooks_json" '
        .hooks = (
            (.hooks // {}) as $existing |
            $new_hooks | to_entries | reduce .[] as $entry (
                $existing;
                .[$entry.key] = ((.[$entry.key] // []) + $entry.value | unique)
            )
        )
    ')

    echo "$merged" > "$SETTINGS_FILE"

    local hook_count
    hook_count=$(echo "$hooks_json" | jq 'keys | length')
    log_info "  Merged $hook_count hook type(s)"
}

# Format the settings file
format_settings() {
    log_info "Formatting settings file..."
    local formatted
    formatted=$(jq '.' "$SETTINGS_FILE")
    echo "$formatted" > "$SETTINGS_FILE"
}

# Main
main() {
    echo "========================================="
    echo "  Vibe Config Sync"
    echo "========================================="
    echo ""

    check_dependencies
    ensure_settings_file

    log_info "Config directory: $CONFIG_DIR"
    log_info "Settings file: $SETTINGS_FILE"
    echo ""

    # Create backup
    if [ -f "$SETTINGS_FILE" ]; then
        cp "$SETTINGS_FILE" "$SETTINGS_FILE.backup"
        log_info "Created backup at $SETTINGS_FILE.backup"
    fi

    merge_mcp_servers
    merge_mcp_servers_to_config
    merge_permissions
    merge_hooks

    format_settings

    echo ""
    log_info "Sync complete!"
}

main "$@"
