# Databricks Query

Execute SQL queries against Databricks using the SQL Statements API and display results in a formatted table.

## How to Invoke

### Slash Command

```
/databricks-query
```

### Example Prompts

```
"Run a SQL query on Databricks to show the top 10 customers"
"Query the sales table in my Databricks workspace"
"Execute this SQL and show me the results in a table"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks CLI | Authenticated via `/databricks-authentication` |
| SQL Warehouse | Selected via `/databricks-warehouse-selector` |

## What This Skill Does

1. Ensures authentication is set up for the target workspace
2. Selects an appropriate SQL warehouse using the warehouse selector
3. Executes the SQL statement via the Databricks SQL Statements API
4. Parses and pretty-prints the results using the helper script

## Key Resources

| File | Description |
|------|-------------|
| `resources/databricks_query_pretty_print.py` | Formats query output as a readable table |

## Related Skills

- `/databricks-authentication` - Authenticate before running queries
- `/databricks-warehouse-selector` - Choose the best warehouse for query execution
- `/databricks-lakeview-dashboard` - Visualize query results in a dashboard
