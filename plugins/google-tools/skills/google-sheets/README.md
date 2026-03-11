# Google Sheets

Create and manage well-formatted Google Sheets with cell updates, formulas, charts, conditional formatting, colors, borders, multiple sheets, and comments.

## How to Invoke

### Slash Command

```
/google-sheets
```

### Example Prompts

```
"Create a spreadsheet with monthly sales data and a column chart"
"Add a new sheet called 'Q2 Data' to this spreadsheet and format the header row"
"Update cell B5 with the formula =SUM(B2:B4) in my spreadsheet"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` first to authenticate with Google Workspace |
| gcloud CLI | Must be installed (`brew install --cask google-cloud-sdk`) |
| Quota Project | Uses your GCP quota project (`$GCP_QUOTA_PROJECT`) for API billing |

## What This Skill Does

1. Creates spreadsheets with formatted tables, headers, and borders
2. Updates cells using A1 notation or grid coordinates with batch operations
3. Applies rich formatting (colors, fonts, number formats, conditional formatting, gradients)
4. Creates charts (column, line, pie, bar, scatter) linked to data ranges
5. Manages multiple sheets with cross-sheet references and named ranges
6. Supports find-and-replace (including regex), cell merging, and frozen rows/columns
7. Adds comments with @mentions and cell-anchored discussions

## Key Resources

| File | Description |
|------|-------------|
| `resources/gsheets_builder.py` | Spreadsheet operations (create, update-cells, format-header, add-table, add-chart, find-replace) |
| `resources/gsheets_cli.sh` | Shell-based CLI helper for Sheets operations |

## Related Skills

- `/google-auth` - Required authentication before using Sheets
- `/google-slides` - Embed Sheets charts into presentations
- `/google-docs` - Reference spreadsheet data in documents
