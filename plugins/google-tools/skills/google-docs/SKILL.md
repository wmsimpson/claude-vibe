---
name: google-docs
description: Open, read, create, edit, and manage Google Docs from docs.google.com URLs. Use this skill for ANY Google Docs operation - reading existing documents, creating new ones, updating content, or managing Drive files. Handles Slides and Drive files too. Use this instead of fetch for reading google docs. Also use this skill to convert markdown to Google Docs format.
---

# Google Docs Skill

Comprehensive Google Docs operations using gcloud CLI + curl (no MCP tools). This skill provides patterns and utilities for:
- **Opening/reading docs.google.com URLs** - Extract and view document content
- **Creating new documents** - With proper headings, tables, hyperlinks, images, and @mentions
- **Updating existing documents** - Modify content, add sections, format text
- **Managing Drive files** - List, search, share, and organize files

---

## IMPORTANT: ALWAYS Write Markdown First, Then Convert

**THIS IS THE REQUIRED WORKFLOW FOR CREATING FORMATTED GOOGLE DOCS:**

1. **Write your content as a markdown file first** (to `/tmp/` or similar)
2. **Use `markdown_to_gdocs.py` to convert it to a Google Doc**

**DO NOT** manually construct Google Docs API batch updates with `insertText`, `updateTextStyle`, `createParagraphBullets`, etc. This approach is error-prone (requires calculating character indices), slow (multiple API calls), and hard to maintain.

**DO NOT** use raw API calls to create tables. The `markdown_to_gdocs.py` script handles markdown pipe-syntax tables (`| col1 | col2 |`) correctly, including bold headers, links in cells, and proper index calculation. Manual table creation via `insertTable` + `insertText` is the #1 source of formatting errors.

---

## REQUIRED: Use markdown_to_gdocs.py for Document Creation

**When creating documents with formatted content (bold, tables, links), ALWAYS use the markdown converter script:**

```bash
# Write content to a temp markdown file, then convert
cat > /tmp/doc_content.md << 'EOF'
# Document Title

## Section 1

This has **bold text** and [a link](https://example.com).

| Column 1 | Column 2 |
|----------|----------|
| **Bold** | Normal |
| [Link](url) | Data |

- Bullet item 1
- Bullet item 2
EOF

python3 ~/.claude/plugins/cache/claude-vibe/google-tools/*/skills/google-docs/resources/markdown_to_gdocs.py \
  --input /tmp/doc_content.md \
  --title "My Document"
```

The script handles:
- Headings (# through ######)
- Bold (`**text**`) and italic (`*text*`)
- Hyperlinks (`[text](url)`)
- Tables with bold cells and links
- Bullet and numbered lists
- Code blocks

The script outputs JSON with `documentId` and `url` fields.

**Arguments:**
- `--input` / `-i` (required): Path to the markdown file
- `--title` / `-t` (optional): Document title (for new documents)
- `--doc-id` / `-d` (optional): Existing doc ID to append to
- `--output` / `-o` (optional): File to write result JSON to

## Authentication

**Run `/google-auth` first** to authenticate with Google Workspace, or use the shared auth module:

```bash
# Check authentication status
python3 ../google-auth/resources/google_auth.py status

# Login if needed (includes automatic retry if OAuth times out)
python3 ../google-auth/resources/google_auth.py login

# Get access token for API calls
TOKEN=$(python3 ../google-auth/resources/google_auth.py token)
```

All Google skills share the same authentication. See `/google-auth` for details on scopes and troubleshooting.

### CRITICAL: If Authentication Fails

**If the login command fails**, it means the user did NOT complete the OAuth flow in the browser.

**DO NOT:**
- Try alternative authentication methods
- Create OAuth credentials manually
- Attempt to set up service accounts

**ONLY solution:**
- Re-run `python3 ../google-auth/resources/google_auth.py login`
- The script includes automatic retry logic with clear instructions
- The user MUST click "Allow" in the browser window

### Quota Project

All API calls require a quota project header:

```bash
-H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Troubleshooting

1. **"API not enabled" error**: Ensure quota project is set correctly
2. **"Insufficient scopes" error**: Re-run login with all required scopes
3. **"Permission denied" error**: Check quota project access in GCP console
4. **Token expired**: Run `python3 ../google-auth/resources/google_auth.py login` to refresh
5. **"Authentication failed"**: User didn't complete OAuth - retry the login command

## Quick Start - Creating a Formatted Document

The recommended workflow for creating any formatted Google Doc:

```bash
# Step 1: Write your content as markdown
cat > /tmp/my_document.md << 'EOF'
# Project Status Report

## Summary

This quarter we achieved **strong growth** across all metrics.

| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| Revenue | $10M | $12M | **Exceeded** |
| Users | 1000 | 1150 | **Exceeded** |
| NPS | 50 | 48 | On Track |

## Next Steps

- Expand into new markets
- Launch v2.0 of the platform
- Hire 5 additional engineers
EOF

# Step 2: Convert to Google Doc
python3 ~/.claude/plugins/cache/claude-vibe/google-tools/*/skills/google-docs/resources/markdown_to_gdocs.py \
  --input /tmp/my_document.md \
  --title "Project Status Report"

# Output: {"documentId": "abc123...", "url": "https://docs.google.com/document/d/abc123.../edit"}
```

**Expected timing:** 10-15 seconds for a complete formatted document with tables, headings, bold, and links.

## Reading a Document

```bash
TOKEN=$(gcloud auth application-default print-access-token)
curl -s "https://docs.googleapis.com/v1/documents/${DOC_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

## Post-Processing After Conversion

After `markdown_to_gdocs.py` creates the document, you may need to apply additional formatting that the script doesn't handle. The most common post-processing tasks:

### Person Chips (@Mentions)

The markdown converter writes `@Name` as plain text. To convert these to interactive person chips, use `gdocs_builder.py`:

```bash
python3 ~/.claude/plugins/cache/claude-vibe/google-tools/*/skills/google-docs/resources/gdocs_builder.py \
  add-person --doc-id "DOC_ID" --email "user@example.com"
```

For bulk @mention replacement (find text mentions and replace with chips), see the full @mentions validation workflow in `resources/GOOGLE_DOCS_API_REFERENCE.md`.

### Checklists

To add interactive checkboxes after document creation:

```bash
python3 ~/.claude/plugins/cache/claude-vibe/google-tools/*/skills/google-docs/resources/gdocs_builder.py \
  add-checklist --doc-id "DOC_ID" \
  --items '["Task 1", "Task 2", "Task 3"]' \
  --checked '[0, 1]'
```

### Status Coloring / Custom Text Styling

To apply colored text, custom fonts, or other styling not supported by markdown, use the `updateTextStyle` API. See `resources/GOOGLE_DOCS_API_REFERENCE.md` for the full API reference.

### Drive Operations (Share, Move, List)

To share the document, move it to a folder, or list Drive files, see `resources/GOOGLE_DOCS_API_REFERENCE.md`.

## API Reference for Post-Processing

For operations that `markdown_to_gdocs.py` does not handle (person chips, comments, checklists, Drive operations, Slides operations, custom text styling), see:

**`resources/GOOGLE_DOCS_API_REFERENCE.md`**

This reference covers:
- Document index concepts and batch update patterns
- Text operations (insert, heading, hyperlink, formatting)
- Table operations (insert, fill, style)
- @Mentions / Person Chips (full validation and replacement workflow)
- Checklists and comments
- Drive operations (list, share, move, create folders)
- Slides operations (create presentations, add slides, shapes, tables, charts)
- Helper script usage (`gdocs_builder.py`, `gslides_builder.py`)

**IMPORTANT:** Use these API calls only for post-processing. Do NOT use them as a substitute for `markdown_to_gdocs.py` when creating documents.

## Common Pitfalls and Solutions

### Problem 1: Using Raw API Calls Instead of markdown_to_gdocs.py

**Symptom**: Tables are malformed, formatting is wrong, document creation takes many API calls

**Solution**: Always use the script first. Write markdown, convert with `markdown_to_gdocs.py`, then use raw API only for post-processing (person chips, comments, etc.).

### Problem 2: Markdown Markers Show Up in Document

**Symptom**: Document shows `**bold**`, `## Heading` as plain text

**Solution**: This happens when you insert raw markdown via the API instead of using `markdown_to_gdocs.py`. The script properly converts all markdown syntax to Google Docs formatting.

### Problem 3: Content Duplication in Document

**Symptom**: Sections appear twice

**Solution**: Check that batch updates aren't applied twice. Common causes:
- Running the same batch_update call multiple times
- Reading document after updates and re-applying formatting

### Problem 4: Batch Update Fails with 400 Error

**Error**: `HTTP Error 400: Bad Request`

**Solution**: Reduce batch size to 50 (not 100+):
```python
for i in range(0, len(requests), 50):  # Safe batch size
    batch_update(requests[i:i+50])
```

### Problem 5: Tables Don't Format Well

**Solution**: Use `markdown_to_gdocs.py` with standard markdown table syntax. The script handles table creation, cell population, bold headers, and links in cells automatically. Do NOT manually construct `insertTable` / `insertText` API calls.

### Problem 6: Document Has Typos

**Solution**: Create a `replaceAllText` batch with common typos:
```bash
curl -s -X POST "https://docs.googleapis.com/v1/documents/${DOC_ID}:batchUpdate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "requests": [
      {"replaceAllText": {"containsText": {"text": "Eecutive", "matchCase": true}, "replaceText": "Executive"}}
    ]
  }' > /dev/null
```

## Performance Tips

**For large documents (1000+ lines):**

1. **Use markdown_to_gdocs.py** - Handles batching automatically
2. **Use replaceAllText for cleanup** - 100x faster than individual replacements
3. **Batch size of 50** - Optimal for API rate limits when doing post-processing
4. **Apply post-processing in order**: Person chips → Status coloring → Typo fixes

## Best Practices

1. **Always use markdown_to_gdocs.py for document creation** - Never construct raw API calls for content the script handles
2. **Read document state before post-processing modifications** - Indices change after every update
3. **Batch your updates** - Multiple requests in one batchUpdate are atomic
4. **Insert in reverse order** - Highest index first to prevent drift (when using raw API)
5. **Apply styles at insert time** - Especially hyperlinks (when using raw API)
6. **Use the helper scripts** - `gdocs_builder.py` handles index calculations correctly for post-processing
7. **Limit table batches to 5 rows** - Prevents API timeouts (when using raw API)

## Example: Create a Complete Document

```bash
#!/bin/bash
# Write your markdown content
cat > /tmp/status_report.md << 'EOF'
# Project Status Report

## Overview

This document tracks our Q4 progress across all workstreams.

## Key Metrics

| Metric | Target | Actual | Delta |
|--------|--------|--------|-------|
| **Revenue** | $10M | $12M | +20% |
| **Active Users** | 1,000 | 1,150 | +15% |
| **NPS Score** | 50 | 48 | -4% |

## Highlights

- Exceeded revenue target by **20%**
- User growth driven by [new onboarding flow](https://example.com/docs)
- NPS slightly below target due to Q4 infrastructure issues

## Next Steps

1. Expand into EMEA market
2. Launch platform v2.0
3. Hire 5 additional engineers
EOF

# Convert to Google Doc
python3 ~/.claude/plugins/cache/claude-vibe/google-tools/*/skills/google-docs/resources/markdown_to_gdocs.py \
  --input /tmp/status_report.md \
  --title "Q4 Project Status Report"
```
