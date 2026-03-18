# Vibe CLI Specification

## Overview

Vibe CLI is a Go-based command-line tool for managing the Field Engineering Claude Code plugin ecosystem. It provides a rich terminal user interface (TUI) using the Bubble Tea framework for interactive configuration, updates, and diagnostics.

## Goals

1. **Portability**: Single static binary with no runtime dependencies
2. **Rich UX**: Interactive TUI with tables, tabs, spinners, and progress animations
3. **Agent Agnostic**: Abstract agent interactions to support Claude Code now, with Cursor and others in the future
4. **Robustness**: Comprehensive error handling, diagnostics, and self-repair capabilities
5. **Test-Driven**: Full test coverage with tests written before implementation

## Technology Stack

- **Language**: Go 1.21+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Components**: [Bubbles](https://github.com/charmbracelet/bubbles) (tables, lists, spinners, progress bars, text inputs)
- **GitHub API**: [go-gh](https://github.com/cli/go-gh) for GitHub authentication (shares auth with `gh` CLI)

## Directory Structure

```
cli/
├── cmd/
│   └── vibe/
│       └── main.go              # Entry point
├── internal/
│   ├── agent/                   # Agent abstraction layer
│   │   ├── agent.go             # Agent interface definition
│   │   ├── claude.go            # Claude Code implementation
│   │   └── cursor.go            # Cursor implementation (future)
│   ├── config/                  # Configuration management
│   │   ├── config.go            # Core config types and loading
│   │   ├── mcp.go               # MCP server configuration
│   │   ├── plugins.go           # Plugin configuration
│   │   ├── permissions.go       # Permissions management
│   │   └── profiles.go          # Profile management
│   ├── doctor/                  # Health checks and diagnostics
│   │   ├── checks.go            # Individual health checks
│   │   ├── collector.go         # Log/env collection for sharing
│   │   └── repair.go            # Auto-repair functionality
│   ├── marketplace/             # Marketplace operations
│   │   ├── marketplace.go       # Marketplace management
│   │   ├── download.go          # Release downloading
│   │   └── sync.go              # Permission/MCP syncing
│   ├── tui/                     # Terminal UI components
│   │   ├── app.go               # Main application model
│   │   ├── styles.go            # Lip Gloss styles/themes
│   │   ├── components/          # Reusable TUI components
│   │   │   ├── table.go         # Selectable table
│   │   │   ├── tabs.go          # Tab navigation
│   │   │   ├── progress.go      # Progress bar/spinner
│   │   │   ├── list.go          # Selectable list with toggle
│   │   │   └── confirm.go       # Y/N confirmation dialog
│   │   └── views/               # Command-specific views
│   │       ├── update.go        # Update command view
│   │       ├── doctor.go        # Doctor command view
│   │       ├── configure.go     # Configure command view
│   │       ├── plugins.go       # Plugins tab/view
│   │       ├── mcp.go           # MCP tab/view
│   │       ├── settings.go      # Settings tab/view
│   │       └── profiles.go      # Profiles tab/view
│   └── util/                    # Shared utilities
│       ├── paths.go             # Path constants and helpers
│       ├── shell.go             # Shell rc file management
│       ├── version.go           # Semantic version comparison
│       └── exec.go              # Command execution helpers
├── pkg/                         # Public API (if needed)
├── testdata/                    # Test fixtures
├── go.mod
├── go.sum
├── Makefile
└── SPEC.md                      # This file
```

## Command Groups

### 1. `vibe update`

Self-update the vibe binary and refresh plugins/configuration.

#### Behavior

1. Display a selectable table of available versions (fetched from GitHub releases)
2. Latest version shown at top, highlighted if newer than installed
3. User selects version with arrow keys, Enter to confirm
4. Download and extract selected release with package-manager style animations:
   - Spinner while downloading
   - Progress bar for extraction
   - Checklist of steps (marketplace, permissions, MCP, plugins)
5. Self-replace the binary in place
6. Sync permissions and MCP servers
7. Reinstall/update default plugins

#### UI Reference

- Table: Based on `examples/table` from Bubble Tea
- Animations: Based on `examples/package-manager`

#### Flags

- `--version <version>`: Skip selection, install specific version
- `--skip-plugins`: Don't reinstall plugins
- `--skip-sync`: Don't sync permissions/MCP

### 2. `vibe doctor`

Run health checks and diagnose issues.

#### Behavior (Default)

1. Run all health checks in sequence with progress indicator
2. Display results as checklist (✓ pass, ✗ fail, ⚠ warning)
3. If issues found, prompt "Would you like to repair? [y/N]"
4. If user confirms, run repair for each failed check

#### Health Checks

Based on the legacy installer and CLI doctor command:

| Check | Description | Repair Action |
|-------|-------------|---------------|
| `prereqs` | gh, jq, yq, claude installed | Print install instructions |
| `marketplace` | ~/.vibe/marketplace exists | Run `vibe update` |
| `marketplace_registered` | vibe registered with claude | `claude plugin marketplace add` |
| `settings_json` | ~/.claude/settings.json valid | Create default or fix JSON |
| `permissions` | Required permissions present | Sync from marketplace |
| `mcp_config` | ~/.config/mcp/config.json valid | Create/fix |
| `mcp_servers` | MCP servers configured | Sync from marketplace |
| `plugins_installed` | Default plugins installed | Install missing |
| `plugins_outdated` | No outdated plugins | Reinstall outdated |
| `python3` | Python 3.10+ available | Print install instructions |
| `env_vars` | Required env vars in shell rc | Add missing |
| `local_ownership` | ~/.local not root-owned | `chown` fix |

#### Flags

- `--repair`: Automatically repair without prompting
- `--collect`: Generate diagnostic tar.gz for sharing
- `--check <name>`: Run specific check only

#### Collect Output

When `--collect` is specified:
1. Run all checks, capture output
2. Collect environment info:
   - OS version, shell, PATH
   - Installed tool versions (gh, jq, yq, claude, python3)
   - Config file contents (redacted secrets)
   - Plugin list and versions
3. Package into `vibe-diagnostics-<timestamp>.tar.gz`
4. Print path for user to share

### 3. `vibe configure`

Interactive configuration UI with tabbed interface.

#### Tabs

##### Settings Tab

User preferences and global settings.

| Setting | Type | Description |
|---------|------|-------------|
| Theme | Select | Color theme (default, dark, light, high-contrast) |
| Default Agent | Select | claude, cursor (when supported) |
| Auto-update Check | Toggle | Check for updates on launch |
| Telemetry | Toggle | Anonymous usage stats (future) |

Settings stored in `~/.vibe/config.yaml`.

##### MCP Tab

Manage MCP servers with enable/disable toggles.

**Data Sources**:
- `~/.claude.json` (mcpServers)
- `~/.config/mcp/config.json` (claude-code section)

**UI**:
- List of MCP servers with:
  - Name
  - Enabled/Disabled status (toggle with Space)
  - Command (dimmed)
- Changes reflected in both config files on save

**Behavior**:
- Load current state from config files
- Space toggles enabled state
- Tab/Shift+Tab or j/k to navigate
- Enter or Ctrl+S to save
- Esc to cancel changes

##### Plugins Tab

Manage plugins from the claude-vibe marketplace.

**Data Sources**:
- `~/.vibe/marketplace/.claude-plugin/marketplace.json` (available)
- `~/.claude/plugins/installed_plugins.json` (installed)

**UI**:
- List of all available plugins:
  - Installed plugins: normal text, enable/disable toggle
  - Not installed: dimmed/grayed out
- Space on installed: toggle enabled/disabled
- Space on not installed: install plugin
- Delete/Backspace on installed: uninstall plugin

**Scopes**:
- User scope (default): affects all projects
- Project scope: only when using profiles

##### Profiles Tab

Named collections of plugins, MCP servers, and permissions.

**Concept**:
- `default` profile: matches user-scope settings (read-only in this UI)
- Custom profiles: project-level overrides activated with `--profile`

**UI**:
- List of profiles with New Profile option
- Selecting a profile shows its configuration:
  - Plugins (can add from uninstalled, not from default)
  - MCP servers (can add, not modify default ones)
  - Permissions (can add custom permissions)
- Items from default profile shown with "(global)" indicator
- Attempting to modify global items shows message

**Storage**:
- Profiles stored in `~/.vibe/profiles/<name>.yaml`
- When `--profile` is used, writes to CWD's `.claude/` directory

#### Flags

- `--tab <name>`: Jump directly to specific tab (settings, mcp, plugins, profiles)

### 4. `vibe plugins`

Shortcut to the Plugins tab in configure.

#### Behavior

- Equivalent to `vibe configure --tab plugins`
- If `--profile <name>` specified, shows profile's plugins tab

#### Flags

- `--profile <name>`: Manage plugins for specific profile
- `--list`: Non-interactive list of plugins (for scripting)
- `--install <name>`: Install plugin non-interactively
- `--uninstall <name>`: Uninstall plugin non-interactively

### 5. `vibe local [plugin-name...]`

Test local plugin changes without manual marketplace management steps.

#### Behavior

1. Walk up from the current directory to find the vibe repository root (looks for `.claude-plugin/marketplace.json`)
2. Remove the existing `claude-vibe` marketplace registration
3. Re-register the local repo root as the `claude-vibe` marketplace source
4. For each plugin name provided, run `claude plugin install <name>@claude-vibe`
5. Launch an agent session (unless `--no-agent` is set)

#### Flags

- `--no-agent`: Skip launching the agent after installation

#### Example

```bash
cd ~/code/vibe
vibe local databricks-tools           # Install one plugin and launch agent
vibe local databricks-tools google-tools  # Install multiple plugins
vibe local databricks-tools --no-agent       # Install without launching agent
```

### 6. `vibe agent [name]`

Launch an agent session with optional project isolation.

#### Behavior

1. If `name` provided:
   - Check if `./name/` exists in CWD
   - If not, create `./name/` directory
   - Change to `./name/`
2. If `--profile <name>` specified (and not "default"):
   - Check for existing `.claude/` directory in CWD
   - If exists, prompt: "Project settings exist. Overwrite? [y/N]"
   - If confirmed or no existing settings:
     - Copy profile's plugins to `.claude/plugins/` (project scope)
     - Copy profile's MCP config to `.claude.json`
     - Copy profile's permissions to `.claude/settings.json`
3. Launch agent session: `claude`

#### Flags

- `--profile <name>`: Apply profile settings before launching
- `--force`: Don't prompt for overwrite, just do it
- `--dry-run`: Show what would be done without doing it

## Agent Abstraction

### Interface

```go
// Agent represents a coding agent (Claude Code, Cursor, etc.)
type Agent interface {
    // Name returns the agent identifier
    Name() string

    // IsInstalled checks if the agent is available
    IsInstalled() bool

    // Version returns the installed version
    Version() (string, error)

    // LaunchSession starts an interactive session
    LaunchSession(opts SessionOptions) error

    // Plugins returns the plugin manager for this agent
    Plugins() PluginManager

    // MCP returns the MCP configuration manager
    MCP() MCPManager

    // Permissions returns the permissions manager
    Permissions() PermissionsManager

    // ConfigPaths returns paths to agent config files
    ConfigPaths() AgentPaths
}

type SessionOptions struct {
    WorkDir string
    Profile string
}

type AgentPaths struct {
    Settings     string // e.g., ~/.claude/settings.json
    Plugins      string // e.g., ~/.claude/plugins/installed_plugins.json
    MCPConfig    string // e.g., ~/.config/mcp/config.json
    GlobalConfig string // e.g., ~/.claude.json
}
```

### Implementations

#### Claude Code (`internal/agent/claude.go`)

- Uses `claude` binary
- Config at `~/.claude/` and `~/.config/mcp/`
- Plugins via `claude plugin` commands

#### Cursor (`internal/agent/cursor.go`) - Future

- Placeholder implementation
- Different config locations
- Different plugin system

## Configuration Files

### Vibe Config (`~/.vibe/config.yaml`)

```yaml
version: 1
settings:
  theme: default
  default_agent: claude
  auto_update_check: true
```

### Profile (`~/.vibe/profiles/<name>.yaml`)

```yaml
version: 1
name: my-profile
description: "Custom profile for ML projects"
plugins:
  - databricks-tools
  - internal-tools
mcp_servers:
  chrome-devtools:
    enabled: true
  slack:
    enabled: false
permissions:
  allow:
    - "Bash(python:*)"
    - "Read(~/ml-projects/**)"
```

## UI Themes

### Default Theme

```go
var DefaultTheme = Theme{
    Primary:    lipgloss.Color("#7C3AED"),  // Purple
    Secondary:  lipgloss.Color("#10B981"),  // Green
    Warning:    lipgloss.Color("#F59E0B"),  // Amber
    Error:      lipgloss.Color("#EF4444"),  // Red
    Muted:      lipgloss.Color("#6B7280"),  // Gray
    Background: lipgloss.Color("#1F2937"),  // Dark gray
}
```

### Styling Guidelines

- Headers: Bold, primary color
- Success indicators: Green checkmark (✓)
- Warnings: Yellow warning (⚠)
- Errors: Red cross (✗)
- Disabled/dim items: Muted color
- Selected items: Inverse colors or highlight

## Testing Strategy

### Unit Tests

Every package must have corresponding `*_test.go` files:

- `internal/config/config_test.go`
- `internal/doctor/checks_test.go`
- `internal/marketplace/marketplace_test.go`
- etc.

### Integration Tests

Test full command flows:

- `test/integration/update_test.go`
- `test/integration/doctor_test.go`
- `test/integration/configure_test.go`

### Test Fixtures

`testdata/` directory contains:

- Sample config files
- Mock marketplace data
- Expected output snapshots

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/doctor/...
```

## Build & Release

### Makefile Targets

```makefile
.PHONY: build test lint clean release

build:
    go build -o bin/vibe ./cmd/vibe

test:
    go test ./...

lint:
    golangci-lint run

clean:
    rm -rf bin/

release:
    goreleaser release --clean
```

### Cross-Compilation

Build for multiple platforms:

- `darwin/amd64` (macOS Intel)
- `darwin/arm64` (macOS Apple Silicon)
- `linux/amd64`
- `linux/arm64`
- `windows/amd64`

### Release Artifacts

- `vibe-darwin-amd64.tar.gz`
- `vibe-darwin-arm64.tar.gz`
- `vibe-linux-amd64.tar.gz`
- `vibe-linux-arm64.tar.gz`
- `vibe-windows-amd64.zip`

## Migration Path

### Phase 1: Core CLI (This PR)

- [ ] Project scaffolding with Go modules
- [ ] Agent abstraction layer (Claude implementation)
- [ ] Config loading/saving
- [ ] `vibe version` command
- [ ] `vibe doctor` command (basic checks)

### Phase 2: TUI Commands

- [ ] Bubble Tea app structure
- [ ] `vibe update` with version table
- [ ] `vibe doctor` with interactive repair
- [ ] Progress animations

### Phase 3: Configure UI

- [ ] Tabbed interface
- [ ] Settings tab
- [ ] MCP tab
- [ ] Plugins tab

### Phase 4: Profiles

- [ ] Profile management
- [ ] Profiles tab
- [ ] `vibe agent` with profile support

### Phase 5: Polish

- [ ] Themes
- [ ] `--collect` diagnostics
- [ ] Cross-platform testing
- [ ] Documentation

## Open Questions

1. **Installation location**: Should the binary go to `/usr/local/bin/vibe` or `~/.local/bin/vibe`?
   - Current: `/usr/local/bin/vibe` (requires sudo)
   - Proposal: `~/.local/bin/vibe` (no sudo needed)

2. **Config file format**: YAML vs TOML vs JSON?
   - Proposal: YAML (human-readable, matches existing config)

3. **Backwards compatibility**: Support old shell script features during transition?
   - Proposal: Yes, maintain command compatibility

4. **Profile storage**: Store profiles globally or per-project?
   - Proposal: Global definitions in `~/.vibe/profiles/`, applied to projects on use

## References

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Bubble Tea Examples](https://github.com/charmbracelet/bubbletea/tree/main/examples)
- [Lip Gloss Styling](https://github.com/charmbracelet/lipgloss)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [go-gh Library](https://github.com/cli/go-gh)
