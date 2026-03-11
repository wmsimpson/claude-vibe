# Databricks Apps

Build and deploy full-stack Databricks Apps with Lakebase database integration, Foundation Model API for AI features, and React/FastAPI frontends.

## How to Invoke

### Slash Command

```
/databricks-apps
```

### Example Prompts

```
"Build a Databricks App with a React frontend and FastAPI backend"
"Create a full-stack app that uses Lakebase for data storage"
"Deploy my app to Databricks Apps with Foundation Model API integration"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| FE-VM Workspace | Serverless workspace via `/databricks-fe-vm-workspace-deployment` |
| Databricks CLI | Version 0.229.0+ authenticated to the workspace |
| uv | Python package manager for backend dependencies |

## What This Skill Does

1. Guides you through choosing an architecture (Node.js + FastAPI or pure Node.js)
2. Sets up project scaffolding with app.yaml, backend, and React frontend
3. Implements dual-mode authentication (local dev vs. deployed app)
4. Configures Lakebase connection with OAuth token refresh
5. Integrates Foundation Model API for AI-powered features
6. Walks through local testing with Chrome DevTools MCP
7. Deploys the app to Databricks Apps with resource binding

## Related Skills

- `/databricks-fe-vm-workspace-deployment` - Provision a workspace with Lakebase and Apps support
- `/databricks-lakebase` - Create and manage Lakebase Postgres databases
- `/databricks-resource-deployment` - Deploy Lakebase instances and serving endpoints via bundles
- `/databricks-demo` - End-to-end demo orchestration that may include an app
