#!/bin/bash
# Smart launcher for MCP servers
# Tries uv venv first, falls back to PEX if available

SERVER_NAME=$1

if [ -z "$SERVER_NAME" ]; then
  echo "Usage: $0 <server-name>" >&2
  exit 1
fi

PLUGIN_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVER_DIR="$PLUGIN_ROOT/servers/$SERVER_NAME"

# Try uv venv first (preferred)
if [ -f "$SERVER_DIR/.venv/bin/python" ]; then
  exec "$SERVER_DIR/.venv/bin/python" "$SERVER_DIR/src/${SERVER_NAME}_server.py"
# Fallback to PEX
elif [ -f "$SERVER_DIR/dist/${SERVER_NAME}.pex" ]; then
  exec "$SERVER_DIR/dist/${SERVER_NAME}.pex"
else
  echo "Error: No installation found for $SERVER_NAME" >&2
  echo "  Checked:" >&2
  echo "    - $SERVER_DIR/.venv/bin/python" >&2
  echo "    - $SERVER_DIR/dist/${SERVER_NAME}.pex" >&2
  exit 1
fi
