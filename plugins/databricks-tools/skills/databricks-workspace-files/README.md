# Databricks Workspace Files

Explore and retrieve code from Databricks workspaces using the Databricks CLI. List, browse, and pull notebooks and scripts into context for code review, debugging, or understanding existing code.

## How to Invoke

### Slash Command

```
/databricks-workspace-files
```

### Example Prompts

```
"List the notebooks in my Databricks workspace"
"Pull the ETL notebook from /Users/me@company.com/projects into context"
"Show me the files under /Repos in the Databricks workspace"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks CLI | Authenticated via `/databricks-authentication` |

## What This Skill Does

1. Lists files and directories at any workspace path (supports /Repos, /Users, /Shared, /Workspace)
2. Provides recursive tree-style directory listing with type indicators
3. Exports notebooks in SOURCE format (.py, .sql, .r, .scala) or JUPYTER format
4. Saves files locally for detailed analysis with the Read tool

## Key Resources

| File | Description |
|------|-------------|
| `scripts/list_workspace.py` | Recursive tree-style workspace directory listing |

## Related Skills

- `/databricks-authentication` - Authenticate before browsing workspace files
- `/databricks-lineage` - Discover related notebooks and then pull their code with this skill
