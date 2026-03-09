# Databricks Lakebase

Create, configure, and query Lakebase Postgres databases using the Databricks CLI. Lakebase provides fully-managed PostgreSQL with automatic scaling, branching, and Unity Catalog integration.

## How to Invoke

### Slash Command

```
/databricks-lakebase
```

### Example Prompts

```
"Create a Lakebase Postgres instance for my app"
"Connect to my Lakebase database and create tables"
"Set up Lakebase with a REST Data API for my frontend"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| FE-VM Workspace | Serverless workspace via `/databricks-fe-vm-workspace-deployment` |
| Databricks CLI | Version 0.229.0+ authenticated to the workspace |
| psql (optional) | PostgreSQL client for direct connections (`brew install postgresql@16`) |

## What This Skill Does

1. Creates a Lakebase instance with configurable capacity (CU_1 through CU_8)
2. Provides multiple connection methods: CLI psql, direct psql with OAuth, Python (psycopg2/SQLAlchemy)
3. Manages databases, tables, scaling, read replicas, and stop/start
4. Registers Lakebase databases in Unity Catalog for SQL warehouse access
5. Configures the PostgREST-compatible Data API for HTTP access
6. Supports synced tables for reverse ETL from Delta Lake

## Related Skills

- `/databricks-fe-vm-workspace-deployment` - Provision a serverless workspace with Lakebase support
- `/databricks-apps` - Build apps with Lakebase as the backend database
- `/databricks-resource-deployment` - Deploy Lakebase instances via Databricks bundles
