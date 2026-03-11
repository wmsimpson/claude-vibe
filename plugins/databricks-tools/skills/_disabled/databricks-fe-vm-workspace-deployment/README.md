# FE Vending Machine Workspace Deployment

Deploy and manage Databricks workspaces using the FE Vending Machine (FEVM) with automatic authentication and environment caching. Also retrieves information about previously deployed workspaces.

## How to Invoke

### Slash Command

```
/databricks-fe-vm-workspace-deployment
```

### Example Prompts

```
"Create a new serverless Databricks workspace for my demo"
"What workspace did I deploy earlier?"
"Find an existing FE-VM workspace with at least 7 days remaining"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Chrome DevTools MCP | For browser-based FEVM authentication |

## What This Skill Does

1. Checks cached environments in `~/.vibe/fe-vm/` for existing workspaces
2. Refreshes the cache from FEVM if needed (handles SSO auth via browser)
3. Finds a suitable existing workspace or deploys a new one (Serverless or Classic)
4. Monitors deployment status until the workspace is ready
5. Guides CLI authentication to the new workspace

## Key Resources

| File | Description |
|------|-------------|
| `resources/environment_manager.py` | List, find, and manage cached workspace environments |
| `resources/fe_vm_client.py` | Deploy workspaces, refresh cache, check deployment status |
| `resources/browser_auth.py` | Handle browser-based FEVM authentication |

## Related Skills

- `/databricks-authentication` - Set up CLI profile for the deployed workspace
- `/databricks-apps` - Build apps on the provisioned workspace
- `/databricks-lakebase` - Create Lakebase instances (requires serverless workspace)
- `/databricks-resource-deployment` - Deploy resources into the workspace
