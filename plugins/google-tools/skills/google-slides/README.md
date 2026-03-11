# Google Slides

Create and manage professional Google Slides presentations with Databricks corporate templates, tables, charts, images, and multiple layout options using gcloud CLI + curl.

## How to Invoke

### Slash Command

```
/google-slides
```

### Example Prompts

```
"Create a Databricks-branded presentation for the customer success story"
"Add a slide with a comparison table showing before and after metrics"
"Build a 5-slide deck from this outline using the dark theme template"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` first to authenticate with Google Workspace |
| gcloud CLI | Must be installed (`brew install --cask google-cloud-sdk`) |
| Quota Project | Uses your GCP quota project (`$GCP_QUOTA_PROJECT`) for API billing |

## What This Skill Does

1. Creates presentations from the Databricks corporate template (light and dark themes)
2. Adds slides using 20+ template layouts (title, content, 2-column, 3-column, section breaks, industry-specific)
3. Inserts text boxes, images, tables, and shapes with precise positioning
4. Embeds linked charts from Google Sheets that auto-update with data changes
5. Supports batch creation from JSON specs for programmatic deck generation
6. Manages placeholders, text replacement, and slide duplication for template-driven workflows

## Key Resources

| File | Description |
|------|-------------|
| `resources/gslides_builder.py` | Complete presentation operations (create-from-template, add-template-slide, create-from-spec, add-table, add-chart) |

## Related Skills

- `/google-auth` - Required authentication before using Slides
- `/google-sheets` - Create charts in Sheets and embed them in presentations
- `/google-docs` - Create supporting documents alongside presentations
