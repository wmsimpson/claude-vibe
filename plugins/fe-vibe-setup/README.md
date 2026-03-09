# FE Vibe Setup Plugin

Environment setup, configuration, and validation for the vibe Field Engineering toolkit.

## Skills Included

### configure-vibe
Comprehensive environment setup and validation:

**Environment Setup:**
- Install required CLI tools (brew, gh, databricks, sf, terraform, aws, jq, yq, uv, gcloud)
- Configure AWS profiles
- Setup ~/.vibe directory structure
- Download Terraform provider documentation
- Configure vibe profile

**Permission Management:**
- Validate all required permissions are configured
- Identify missing permissions
- Offer to add missing permissions automatically
- Create backup before making changes

**Plugin Validation:**
- Check vibe plugins are installed correctly
- Suggest fixes if issues found

**Usage:** This skill is invoked automatically by Claude when setup/validation is needed, or you can explicitly run it:

```bash
claude
# Then ask: "configure vibe" or "validate my vibe setup"
```

## Resources

- `permissions.yaml` - Master list of required permissions
- `mcp-servers.yaml` - External MCP server configurations
- `hooks.yaml` - Hook configurations (currently empty)

## Manual Permission Management

If you prefer to manage permissions manually, check the required permissions in:
```
~/.vibe/marketplace/permissions.yaml
```

Then add them to your `~/.claude/settings.json`:
```json
{
  "allow": [
    "Bash",
    "Skill(databricks-authentication)",
    "Skill(configure-vibe)",
    // ... add all from permissions.yaml
  ]
}
```

## Troubleshooting

### Permission Errors

If you see permission errors:
1. Run the configure-vibe skill (it includes permission validation)
2. Or re-run the vibe install script
3. Restart Claude Code after changes

```bash
claude
# Ask: "configure vibe"
```

### Outdated Setup

After updating vibe, you can either:

**Option 1: Re-run install** (updates everything)
```bash
vibe update
```

**Option 2: Use configure-vibe skill** (validates and fixes specific issues)
```bash
claude
# Ask: "configure vibe"
```
