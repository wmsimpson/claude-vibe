# Google Docs

Open, read, create, edit, and manage Google Docs. Converts markdown to fully formatted Google Docs with headings, bold, links, tables, and lists. Also handles Slides and Drive file operations.

## How to Invoke

### Slash Command

```
/google-docs
```

### Example Prompts

```
"Create a Google Doc from this markdown file with the title 'Q4 Report'"
"Read the contents of this Google Doc: https://docs.google.com/document/d/..."
"Share this document with colleague@example.com as an editor"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` first to authenticate with Google Workspace |
| gcloud CLI | Must be installed (`brew install --cask google-cloud-sdk`) |
| Quota Project | Uses your GCP quota project (`$GCP_QUOTA_PROJECT`) for API billing |

## What This Skill Does

1. Creates new Google Docs with rich formatting via the markdown-to-Docs converter
2. Reads and extracts content from existing Google Docs URLs
3. Applies formatting (headings, bold, italic, hyperlinks, tables, bullet lists, code blocks)
4. Inserts person chips (@mentions) as interactive smart chips with email lookup
5. Manages Drive files (list, search, share, create folders, move files)
6. Creates and manages Google Slides presentations with text, images, tables, and charts

## Key Resources

| File | Description |
|------|-------------|
| `resources/markdown_to_gdocs.py` | Converts markdown files to fully formatted Google Docs |
| `resources/gdocs_builder.py` | Document operations (create, read, add-section, add-table, add-person, add-checklist) |
| `resources/gslides_builder.py` | Presentation operations (create, add-slide, add-text-box, add-image, add-table) |
| `resources/gdocs_cli.sh` | Shell-based CLI helper for Docs operations |

## Related Skills

- `/google-auth` - Required authentication before using Docs
- `/google-sheets` - Create spreadsheets with data that can be embedded in Docs
- `/google-slides` - Dedicated skill for more advanced presentation creation
- `/google-calendar` - Attach Docs as meeting notes to calendar events
