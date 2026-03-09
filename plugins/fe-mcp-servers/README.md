# FE MCP Servers Plugin

This plugin is ready for custom Field Engineering MCP servers. Currently empty.

## Adding a New MCP Server

### 1. Create Server Directory

```bash
mkdir -p servers/my-server/src
cd servers/my-server
```

### 2. Create pyproject.toml

```toml
[project]
name = "my-server"
version = "1.0.0"
requires-python = ">=3.10"
dependencies = [
    "mcp>=0.1.0",
    # add your dependencies
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

### 3. Write Server Code

Create `src/my_server.py` with your MCP server implementation.

### 4. Register in .mcp.json

Add to root `.mcp.json`:

```json
{
  "my-server": {
    "command": "${CLAUDE_PLUGIN_ROOT}/scripts/run-server.sh",
    "args": ["my-server"]
  }
}
```

### 5. Install Dependencies

```bash
cd servers/my-server
uv sync
```

Or run the setup script from the plugin root:

```bash
./scripts/setup-mcp-servers.sh
```

### 6. Test

```bash
claude
/mcp  # Should show your server
```

## Distribution Options

### Development (Recommended)
Use uv venv for fast iteration:

```bash
cd servers/my-server
uv sync
```

### Production (Optional)
Build PEX for guaranteed isolation:

```bash
cd servers/my-server
uv build
pip install pex
pex . -o dist/my-server.pex -c my-server
```

The `run-server.sh` script will automatically prefer venv over PEX if both exist.

## Architecture

- **servers/**: Each subdirectory is an independent MCP server
- **scripts/setup-mcp-servers.sh**: Installs all servers with `uv sync`
- **scripts/run-server.sh**: Smart launcher (venv → PEX fallback)
- **.mcp.json**: Server registration for Claude Code

## Dependencies

- **uv**: Python package manager (required)
- **Python 3.10+**: For MCP server runtime
