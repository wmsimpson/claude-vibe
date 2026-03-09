# Databricks Authentication

Authenticate to Databricks workspace and account environments using the Databricks CLI. This should be done before any other Databricks operations.

## How to Invoke

### Slash Command

```
/databricks-authentication
```

### Example Prompts

```
"Authenticate me to the Databricks demo workspace"
"Log in to your Databricks workspace"
"Set up my Databricks CLI profile for my FE-VM workspace"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks CLI | Installed and available on PATH |

## What This Skill Does

1. Identifies the correct environment and profile for your use case (any workspace you have configured)
2. Checks if a valid CLI profile already exists with `databricks auth profiles`
3. If needed, runs `databricks auth login` with the appropriate host and profile name
4. Guides SSO-based login for workspace or account-level authentication

## Related Skills

- `/databricks-demo` - Create demos after authenticating
- `/databricks-resource-deployment` - Deploy resources after authenticating
- `/databricks-query` - Run SQL queries after authenticating
- `/databricks-fe-vm-workspace-deployment` - Provision and authenticate to FE-VM workspaces
