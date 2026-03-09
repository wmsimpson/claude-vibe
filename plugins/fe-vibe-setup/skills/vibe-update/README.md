# Vibe Update

Updates vibe marketplace, plugins, permissions, and MCP server configurations to the latest version.

## How to Invoke

### Slash Command

```
/vibe-update
```

### Example Prompts

```
"Update vibe to the latest version"
"Pull the latest vibe plugins"
"Sync my vibe installation"
```

## What This Skill Does

1. Runs `vibe update` to download the latest release from GitHub
2. Updates the marketplace at `~/.vibe/marketplace`
3. Syncs permissions to `~/.claude/settings.json`
4. Syncs MCP server configurations
5. Reinstalls all currently installed plugins to pick up changes
6. Updates the vibe CLI itself (may require sudo)
7. Advises restarting Claude Code to apply changes

## Related Skills

- `/configure-vibe` - Full environment setup if update reveals missing dependencies
- `/vibe-publish-plugin` - Publish new skills before updating to include them
