---
name: databricks-config
description: Manage Databricks workspace connections: check which workspace you're connected to, switch workspaces, list available workspaces, or authenticate to a new workspace.
---

# Databricks Config

Switch between Databricks workspace profiles. The active profile controls which workspace the Databricks MCP server and CLI commands target.

Databricks profiles can be linked to vibe profiles (`~/.vibe/profiles/<name>.yaml`) so switching vibe profiles also switches the Databricks workspace.

---

## Step 1 — Show Current State

```bash
# Current active profile
source ~/.vibe/env 2>/dev/null
echo "Active Databricks profile: ${DATABRICKS_CONFIG_PROFILE:-DEFAULT}"
echo ""

# All available Databricks CLI profiles
databricks auth profiles
echo ""

# Vibe profiles with Databricks bindings
for f in ~/.vibe/profiles/*.yaml; do
  [ -f "$f" ] || continue
  name=$(basename "$f" .yaml)
  db=$(grep 'databricks_profile:' "$f" 2>/dev/null | awk '{print $2}')
  [ -n "$db" ] && echo "Vibe profile '$name' → Databricks profile '$db'"
done
```

---

## Step 2 — Switch Profile

If the user wants to switch to a different Databricks profile:

### 2a. Update ~/.vibe/env

```bash
sed -i '' 's/^export DATABRICKS_CONFIG_PROFILE=.*/export DATABRICKS_CONFIG_PROFILE="<PROFILE_NAME>"/' ~/.vibe/env
source ~/.vibe/env
echo "Active profile set to: $DATABRICKS_CONFIG_PROFILE"
```

### 2b. Update ~/.mcp.json

```bash
python3 -c "
import json, os
path = os.path.expanduser('~/.mcp.json')
with open(path) as f:
    cfg = json.load(f)
if 'databricks' in cfg.get('mcpServers', {}):
    cfg['mcpServers']['databricks']['env']['DATABRICKS_CONFIG_PROFILE'] = '<PROFILE_NAME>'
    with open(path, 'w') as f:
        json.dump(cfg, f, indent=2)
    print('MCP config updated.')
else:
    print('No databricks MCP server configured in ~/.mcp.json')
"
```

### 2c. Verify

```bash
databricks workspace list / --profile=<PROFILE_NAME>
```

Tell the user: **Restart Claude Code** for the MCP server to pick up the new profile.

---

## Step 3 — Link to a Vibe Profile

To bind a Databricks profile to a vibe profile so they switch together:

```bash
mkdir -p ~/.vibe/profiles

# Create or update the vibe profile
python3 -c "
import yaml, os

profile_name = '<VIBE_PROFILE_NAME>'
db_profile = '<DATABRICKS_PROFILE_NAME>'
path = os.path.expanduser(f'~/.vibe/profiles/{profile_name}.yaml')

# Load existing or create new
profile = {}
if os.path.exists(path):
    with open(path) as f:
        profile = yaml.safe_load(f) or {}

profile['databricks_profile'] = db_profile
profile.setdefault('name', profile_name)
profile.setdefault('version', 1)

with open(path, 'w') as f:
    yaml.dump(profile, f, default_flow_style=False)

print(f'Vibe profile \"{profile_name}\" now uses Databricks profile \"{db_profile}\"')
"
```

When `vibe profile apply <name>` is run, the Databricks profile will automatically be set in `~/.mcp.json` and `~/.vibe/env`.

---

## Step 4 — Add a New Databricks Profile

If the user wants to connect to a workspace that doesn't have a profile yet, invoke the `databricks-authentication` skill.

---

## Quick Reference

| Action | Command |
|--------|---------|
| Check active profile | `source ~/.vibe/env && echo $DATABRICKS_CONFIG_PROFILE` |
| List profiles | `databricks auth profiles` |
| Switch profile | `/databricks-config` |
| Add new workspace | `/databricks-authentication` |
| Link to vibe profile | See Step 3 above |
