# Vibe Publish Plugin

Publishes skills, hooks, MCP servers, or entire plugins to the vibe marketplace. Validates structure, creates branches, and handles PR creation.

## How to Invoke

### Slash Command

```
/vibe-publish-plugin
```

### Example Prompts

```
"Publish this skill to the vibe marketplace"
"I built a new hook and want to add it to vibe"
"Add my plugin to the vibe marketplace"
```

## What This Skill Does

1. **Detects** publishable artifacts in the current directory (skills, hooks, MCP servers, plugins)
2. **Validates** structure and metadata (frontmatter, plugin.json, naming conventions)
3. **Determines** destination -- existing plugin, new plugin, or external repository
4. **Clones** the vibe repo (or uses existing checkout) and creates a feature branch
5. **Integrates** the contribution (copies files, updates permissions, adds evals)
6. **Creates** a pull request with the full PR template filled out

## Related Skills

- `/configure-vibe` - Ensure environment is set up before publishing
- `/vibe-update` - Update vibe after a published plugin is merged
