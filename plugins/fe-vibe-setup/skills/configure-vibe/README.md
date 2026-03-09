# Configure Vibe

Comprehensive environment setup and validation for the vibe Field Engineering toolkit, including CLI tools, authentication, and user profile.

## How to Invoke

### Slash Command

```
/configure-vibe
```

### Example Prompts

```
"Set up my vibe environment"
"I need to configure vibe from scratch"
"Run the vibe setup and validate my tools"
```

## What This Skill Does

1. Installs Homebrew (if missing) and required CLI tools (jq, yq, ripgrep, awscli, uv, pipx)
2. Installs and configures the Databricks CLI
3. Installs the Salesforce CLI and authenticates via browser SSO
4. Installs gcloud CLI and authenticates with Google Workspace (Drive, Docs, Sheets, Gmail, Calendar)
5. Configures AWS profiles by downloading the SSO config
6. Sets up directories (`~/.vibe`, `~/code`)
7. Checks for an existing vibe profile and optionally builds one via the `vibe-profile` agent
8. Summarizes setup status and provides next steps

## Key Resources

| File | Description |
|------|-------------|
| `resources/permissions.yaml` | Master permissions configuration |
| `resources/mcp-servers.yaml` | External MCP server configurations |
| `resources/hooks.yaml` | Hook configurations for event handling |
| `resources/vibe_sync.sh` | Sync script for vibe updates |
| `resources/VIBE_PROFILE.md` | Documentation for the vibe profile format |

## Related Skills

- `/vibe-update` - Update vibe and plugins to the latest version
- `/validate-mcp-access` - Validate MCP connection credentials after setup
- `/aws-authentication` - Standalone AWS authentication if setup partially fails
