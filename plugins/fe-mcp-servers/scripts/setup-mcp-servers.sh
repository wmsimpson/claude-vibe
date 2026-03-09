#!/bin/bash
# Setup script for MCP servers using uv
# Runs `uv sync` for each server directory with a pyproject.toml

set -e

PLUGIN_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVERS_DIR="$PLUGIN_ROOT/servers"

echo "Setting up FE MCP servers..."

if [ ! -d "$SERVERS_DIR" ]; then
  echo "No servers directory found. Nothing to install."
  exit 0
fi

# Check if uv is installed
if ! command -v uv &> /dev/null; then
  echo "⚠️  uv is not installed. Please install it first:"
  echo "   curl -LsSf https://astral.sh/uv/install.sh | sh"
  exit 1
fi

# Find and setup each server
server_count=0
for server_dir in "$SERVERS_DIR"/*; do
  if [ -d "$server_dir" ] && [ -f "$server_dir/pyproject.toml" ]; then
    server_name=$(basename "$server_dir")
    echo "  Setting up $server_name..."
    cd "$server_dir"
    uv sync --frozen
    server_count=$((server_count + 1))
  fi
done

if [ $server_count -eq 0 ]; then
  echo "No MCP servers found. Add servers to the servers/ directory."
else
  echo "✅ $server_count MCP server(s) installed successfully"
fi
