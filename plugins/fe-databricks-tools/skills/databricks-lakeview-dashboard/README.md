# Databricks Lakeview Dashboard

Programmatically create and manage Lakeview (AI/BI) dashboards in Databricks using the Dashboard API and a Python builder helper.

## How to Invoke

### Slash Command

```
/databricks-lakeview-dashboard
```

### Example Prompts

```
"Create a Lakeview dashboard showing sales metrics by category"
"Build a dashboard with KPI counters, a line chart, and filters"
"Add a bar chart and pie chart to my Databricks dashboard"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks CLI | Authenticated via `/databricks-authentication` |
| SQL Warehouse | Available for powering dashboard queries |
| Unity Catalog Tables | Data sources for the dashboard visualizations |

## What This Skill Does

1. Provides the full JSON schema for the Lakeview `serialized_dashboard` format
2. Documents all 19 visualization types (bar, line, pie, counter, scatter, heatmap, etc.)
3. Includes widget configuration for charts, tables, and filter widgets
4. Supports dashboard creation, updating, and publishing via the Lakeview API
5. Offers a Python builder class for simplified dashboard construction

## Key Resources

| File | Description |
|------|-------------|
| `resources/lakeview_builder.py` | Python helper class for building dashboards programmatically |
| `resources/example_dashboard.json` | Complete working dashboard JSON example |

## Related Skills

- `/databricks-lakeview-dashboard-analyzer` - Analyze existing dashboards via browser automation
- `/databricks-query` - Test SQL queries before adding them to dashboard datasets
- `/databricks-warehouse-selector` - Choose a warehouse for the dashboard
