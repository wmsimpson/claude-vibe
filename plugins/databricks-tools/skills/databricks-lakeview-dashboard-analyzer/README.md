# Databricks Lakeview Dashboard Analyzer

Analyze existing Databricks AI/BI (Lakeview) dashboards using the Lakeview REST API (default) or Chrome DevTools for visual inspection. Extracts datasets, queries, widgets, and page structure from dashboard definitions.

## How to Invoke

### Slash Command

```
/databricks-lakeview-dashboard-analyzer
```

### Example Prompts

```
"Analyze the dashboard at this Databricks URL and summarize the key metrics"
"Extract the data from the tables in my Lakeview dashboard"
"What queries does this dashboard use?"
"Take a screenshot of my dashboard and explain the trends"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks CLI | Authenticated CLI profile for the target workspace |
| Dashboard URL or ID | Link to a Lakeview dashboard or the dashboard UUID |
| Read Access | Permission to view the dashboard |
| Chrome DevTools MCP | *Optional* — only needed for visual analysis (screenshots, UI interaction) |

## What This Skill Does

1. Parses the dashboard URL to extract the workspace host and dashboard ID
2. Determines the correct Databricks CLI profile for that workspace
3. Fetches the full dashboard definition via the Lakeview REST API
4. Double-parses the `serialized_dashboard` field (JSON string within JSON)
5. Summarizes datasets, pages, widgets, queries, and filters
6. Optionally uses Chrome DevTools for visual analysis when explicitly requested

## Key Resources

| File | Description |
|------|-------------|
| `references/ELEMENT_INTERACTION.md` | Guide for clicking, filling, dropdowns, and tabs |
| `references/DATA_EXTRACTION.md` | Techniques for extracting table data and downloading CSVs |
| `references/WORKFLOWS.md` | Complete analysis workflows (summary, filtering, Q&A) |
| `references/CHART_INTERPRETATION_GUIDE.md` | How to interpret different chart types |
| `references/TROUBLESHOOTING.md` | Fixing MCP disconnects, auth issues, session expiration |

## Related Skills

- `/databricks-lakeview-dashboard` - Create dashboards programmatically (this skill only analyzes existing ones)
- `/databricks-query` - Run SQL queries directly against the underlying data
- `/databricks-warehouse-selector` - Select a warehouse for direct queries
- `/databricks-authentication` - CLI-based authentication
