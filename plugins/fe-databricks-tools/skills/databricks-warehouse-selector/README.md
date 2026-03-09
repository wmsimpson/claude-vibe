# Databricks Warehouse Selector

Select the best SQL warehouse for executing Databricks SQL queries, or create a new serverless warehouse if none are available.

## How to Invoke

### Slash Command

```
/databricks-warehouse-selector
```

### Example Prompts

```
"Select a warehouse for my SQL query"
"Which warehouse should I use in this workspace?"
"Find or create a SQL warehouse for running queries"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks CLI | Authenticated via `/databricks-authentication` |

## What This Skill Does

1. Lists all available warehouses using the Databricks CLI
2. Ranks candidates using the `select_warehouses.py` helper script
3. Returns the best warehouse for query execution
4. Creates a new serverless Small warehouse if none exist

## Key Resources

| File | Description |
|------|-------------|
| `resources/select_warehouses.py` | Ranks warehouses from `databricks warehouses list` output |

## Related Skills

- `/databricks-authentication` - Authenticate before listing warehouses
- `/databricks-query` - Execute queries using the selected warehouse
- `/databricks-lakeview-dashboard` - Dashboards require a warehouse for their queries
