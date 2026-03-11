# Databricks Demo

Create, deploy, and run end-to-end demos relating to or integrating with Databricks, including data generation, app development, and resource deployment.

## How to Invoke

### Slash Command

```
/databricks-demo
```

### Example Prompts

```
"Build a demo for a retail analytics use case on Databricks"
"Create a customer-facing demo with a Databricks App and Lakebase"
"Set up a Databricks demo showcasing real-time streaming"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks CLI | Authenticated via `/databricks-authentication` |
| FE-VM Workspace | For demos requiring Apps or Lakebase |
| uv | Python package manager (always used instead of pip) |

## What This Skill Does

1. Researches the use case or customer to determine data and architecture requirements
2. Writes a DEMO.md (brand guidelines, architecture) and TASKS.md (step-by-step plan)
3. Provisions infrastructure via `/databricks-fe-vm-workspace-deployment` if needed
4. Generates synthetic data using serverless compute
5. Builds and locally tests code before deploying
6. Deploys resources (notebooks, jobs, apps, tables) into the workspace
7. Tracks progress by updating TASKS.md throughout

## Key Resources

| File | Description |
|------|-------------|
| `resources/DATA_GENERATION_SERVERLESS.md` | Instructions for generating synthetic data |
| `resources/DATABRICKS_APPS.md` | Guidelines for the Databricks Apps development loop |
| `resources/LAKEBASE.md` | Lakebase integration reference |

## Related Skills

- `/databricks-fe-vm-workspace-deployment` - Provision demo workspaces
- `/databricks-apps` - Build Databricks Apps for the demo
- `/databricks-resource-deployment` - Deploy clusters, jobs, and other resources
- `/snowflake` - Add Snowflake integration to a demo
