# Databricks Lineage

Explore Databricks Unity Catalog data lineage to trace data flow between tables, understand how assets are connected, find upstream sources or downstream consumers, and investigate column-level dependencies.

## How to Invoke

### Slash Command

```
/databricks-lineage
```

### Example Prompts

```
"Show me the upstream and downstream lineage for the orders table"
"Trace which columns feed into the total_amount field"
"What notebooks and jobs depend on the customers table?"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks CLI | Authenticated via `/databricks-authentication` |
| Unity Catalog | Workspace must have UC enabled with lineage tracking active |

## What This Skill Does

1. Retrieves table-level lineage (upstream sources and downstream consumers)
2. Traces column-level lineage for impact analysis and debugging
3. Searches for tables by pattern and explores their lineage graph to a configurable depth
4. Identifies connected notebooks, jobs, pipelines, and dashboards
5. Supports direct API access for advanced lineage queries

## Key Resources

| File | Description |
|------|-------------|
| `scripts/get_table_lineage.py` | Retrieve upstream/downstream table dependencies |
| `scripts/get_column_lineage.py` | Trace lineage for a specific column |
| `scripts/search_lineage.py` | Search for tables and explore their lineage graph |

## Related Skills

- `/databricks-workspace-files` - Pull notebook/script code into context after discovering lineage connections
- `/databricks-authentication` - Authenticate before exploring lineage
