# Databricks Resource Deployment

Deploy resources such as notebooks, jobs, clusters, warehouses, apps, Lakebase instances, catalogs, and schemas into an existing Databricks workspace.

## How to Invoke

### Slash Command

```
/databricks-resource-deployment
```

### Example Prompts

```
"Deploy a serverless job and notebook to my Databricks workspace"
"Set up a Databricks cluster and catalog for my project"
"Upload my local code to the workspace and create a scheduled job"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks CLI | Authenticated via `/databricks-authentication` |
| Workspace | An existing workspace (FE-VM or e2-demo) |

## What This Skill Does

1. Determines if this is a new project or an update to an existing one
2. Deploys resources with a strong preference for serverless configurations
3. Uses `databricks sync` for file uploads (not import/export commands)
4. Follows Unity Catalog best practices (3-layer namespaces, no DBFS, no Hive Metastore)
5. Reuses existing FE-VM catalogs when possible to avoid permission errors

## Key Resources

| File | Description |
|------|-------------|
| `resources/SERVERLESS.md` | Serverless compute configuration guide |
| `resources/databricks-serverless-config.md` | Serverless cluster and job settings |
| `resources/databricks-serverless-dependencies.md` | Managing dependencies in serverless environments |
| `resources/databricks-serverless-limitations.md` | Known serverless limitations |
| `resources/databricks-serverless-connect-tutorial.md` | Serverless Connect tutorial |
| `resources/databricks-app-templates.md` | App configuration templates |

## Related Skills

- `/databricks-authentication` - Authenticate before deploying resources
- `/databricks-fe-vm-workspace-deployment` - Provision a workspace to deploy into
- `/databricks-apps` - Detailed app development and deployment guidance
- `/snowflake` - Set up Snowflake integrations alongside Databricks resources
