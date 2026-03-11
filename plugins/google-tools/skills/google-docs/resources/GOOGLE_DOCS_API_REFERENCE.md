# Google Docs / Slides / Drive API Reference

> **This is a post-processing and fix-up reference.** Use it AFTER `markdown_to_gdocs.py`
> creates your document. Do NOT use these API calls as a substitute for the markdown converter.
>
> **Use this for:** person chips (@mentions), comments, status coloring, checklists, Drive operations, Slides operations
>
> **Do NOT use this for:** creating documents, headings, bold, tables, links, bullets -- `markdown_to_gdocs.py` handles all of these.

## Core Concepts

### Document Indices

Google Docs uses a 1-based index system. Every character, paragraph break, and structural element has an index.

**Critical Rules:**
1. **Always read before writing** - Get the current document structure to know exact indices
2. **Insert in REVERSE order** - When making multiple insertions, start from the highest index to avoid drift
3. **Hyperlinks must be applied at insert time** - Apply updateTextStyle in the SAME batchUpdate as insertText

### Index Structure

- Document starts at index 1 (after sectionBreak at index 0)
- Each character takes 1 index
- Newlines (`\n`) take 1 index
- Tables have complex index structures (see Table Operations below)

## API Reference

### Create a Document

```bash
TOKEN=$(gcloud auth application-default print-access-token)
curl -s -X POST "https://docs.googleapis.com/v1/documents" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{"title": "My Document"}'
```

### Read a Document

```bash
curl -s "https://docs.googleapis.com/v1/documents/${DOC_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Get Document Structure (indices only)

```bash
curl -s "https://docs.googleapis.com/v1/documents/${DOC_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" | \
  jq '.body.content[] | select(.paragraph) | {startIndex, endIndex, text: .paragraph.elements[0].textRun.content}'
```

### Batch Update

```bash
curl -s -X POST "https://docs.googleapis.com/v1/documents/${DOC_ID}:batchUpdate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{"requests": [...]}'
```

## Text Operations

### Insert Text

```json
{
  "insertText": {
    "location": {"index": 1},
    "text": "Hello World\n"
  }
}
```

### Apply Heading Style

```json
{
  "updateParagraphStyle": {
    "range": {"startIndex": 1, "endIndex": 12},
    "paragraphStyle": {"namedStyleType": "HEADING_1"},
    "fields": "namedStyleType"
  }
}
```

Available styles: `NORMAL_TEXT`, `TITLE`, `SUBTITLE`, `HEADING_1` through `HEADING_6`

### Insert Text with Hyperlink (MUST be in same batchUpdate)

```json
{
  "requests": [
    {
      "insertText": {
        "location": {"index": 1},
        "text": "Click here"
      }
    },
    {
      "updateTextStyle": {
        "range": {"startIndex": 1, "endIndex": 11},
        "textStyle": {"link": {"url": "https://example.com"}},
        "fields": "link"
      }
    }
  ]
}
```

### Apply Text Formatting

```json
{
  "updateTextStyle": {
    "range": {"startIndex": 1, "endIndex": 10},
    "textStyle": {
      "bold": true,
      "italic": false,
      "underline": false,
      "strikethrough": false,
      "foregroundColor": {"color": {"rgbColor": {"red": 0.2, "green": 0.2, "blue": 0.8}}},
      "fontSize": {"magnitude": 14, "unit": "PT"},
      "weightedFontFamily": {"fontFamily": "Roboto", "weight": 400}
    },
    "fields": "bold,italic,underline,strikethrough,foregroundColor,fontSize,weightedFontFamily"
  }
}
```

### Strikethrough (for crossing items off lists)

```json
{
  "updateTextStyle": {
    "range": {"startIndex": 5, "endIndex": 20},
    "textStyle": {"strikethrough": true},
    "fields": "strikethrough"
  }
}
```

## Table Operations

### Insert a Table

```json
{
  "insertTable": {
    "rows": 3,
    "columns": 3,
    "location": {"index": 1}
  }
}
```

### Table Index Formula

For a table starting at index T with C columns:
- Cell at row R, column C has content index: `T + 3 + R*(C*2+1) + c*2`

Example for 3x3 table starting at index 104:
- Row 0: indices 107, 109, 111
- Row 1: indices 114, 116, 118
- Row 2: indices 121, 123, 125

### Fill Table Cells (insert in REVERSE order)

```json
{
  "requests": [
    {"insertText": {"location": {"index": 125}, "text": "Cell 2,2"}},
    {"insertText": {"location": {"index": 123}, "text": "Cell 2,1"}},
    {"insertText": {"location": {"index": 121}, "text": "Cell 2,0"}},
    {"insertText": {"location": {"index": 118}, "text": "Cell 1,2"}},
    {"insertText": {"location": {"index": 116}, "text": "Cell 1,1"}},
    {"insertText": {"location": {"index": 114}, "text": "Cell 1,0"}},
    {"insertText": {"location": {"index": 111}, "text": "Header 3"}},
    {"insertText": {"location": {"index": 109}, "text": "Header 2"}},
    {"insertText": {"location": {"index": 107}, "text": "Header 1"}}
  ]
}
```

### Style Table Header Row

```json
{
  "updateTableCellStyle": {
    "tableRange": {
      "tableCellLocation": {
        "tableStartLocation": {"index": 104},
        "rowIndex": 0,
        "columnIndex": 0
      },
      "rowSpan": 1,
      "columnSpan": 3
    },
    "tableCellStyle": {
      "backgroundColor": {"color": {"rgbColor": {"red": 0.9, "green": 0.9, "blue": 0.9}}}
    },
    "fields": "backgroundColor"
  }
}
```

## Bullet Lists

### Create Bullet List

```json
{
  "requests": [
    {
      "insertText": {
        "location": {"index": 1},
        "text": "Item 1\nItem 2\nItem 3\n"
      }
    },
    {
      "createParagraphBullets": {
        "range": {"startIndex": 1, "endIndex": 21},
        "bulletPreset": "BULLET_DISC_CIRCLE_SQUARE"
      }
    }
  ]
}
```

Bullet presets: `BULLET_DISC_CIRCLE_SQUARE`, `BULLET_DIAMONDX_ARROW3D_SQUARE`, `NUMBERED_DECIMAL_ALPHA_ROMAN`, `NUMBERED_DECIMAL_NESTED`, etc.

## Images

### Insert Image from URL

```json
{
  "insertInlineImage": {
    "location": {"index": 1},
    "uri": "https://example.com/image.png",
    "objectSize": {
      "width": {"magnitude": 300, "unit": "PT"},
      "height": {"magnitude": 200, "unit": "PT"}
    }
  }
}
```

## @Mentions (Person Chips / Smart Chips)

Google Docs supports true person chips (smart chips) via the `insertPerson` request:

```json
{
  "insertPerson": {
    "personProperties": {
      "email": "user@example.com"
    },
    "location": {"index": 1}
  }
}
```

This creates an interactive person chip that shows profile info on hover. You can also use the builder script:

```bash
python3 ~/.claude/plugins/cache/claude-vibe/google-tools/*/skills/google-docs/resources/gdocs_builder.py \
  add-person --doc-id "DOC_ID" --email "user@example.com"
```

## Validating and Fixing @ Mentions

After creating or editing a document, **always read it back to verify @ mentions are proper person chips**, not just plain text like "@John Smith". Plain text mentions don't notify users and aren't clickable.

### Step 1: Read Document and Identify Text-Only Mentions

```bash
TOKEN=$(gcloud auth application-default print-access-token)
QUOTA_PROJECT="${GCP_QUOTA_PROJECT}"

# Get full document content
curl -s "https://docs.googleapis.com/v1/documents/${DOC_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: $QUOTA_PROJECT" | \
  jq -r '.body.content[] | select(.paragraph) | .paragraph.elements[]? | select(.textRun) | .textRun.content' | \
  grep -oE '@[A-Za-z]+ [A-Za-z]+' | sort -u
```

This extracts any text patterns like "@First Last" that are NOT proper person chips.

**How to identify text-only mentions vs proper person chips:**
- Text-only: Shows as `textRun` element with content like "@John Smith"
- Proper chip: Shows as `person` element with `personProperties.email`

```bash
# Check if document has proper person chips
curl -s "https://docs.googleapis.com/v1/documents/${DOC_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: $QUOTA_PROJECT" | \
  jq '.body.content[] | select(.paragraph) | .paragraph.elements[]? | select(.person)'
```

### Step 2: Look Up Names Using Glean MCP

When you find text-only mentions like "@John Smith", use the Glean MCP to find their email addresses:

```bash
# First check the schema
mcp-cli info glean/glean_read_api_call

# Search for the person in Glean
mcp-cli call glean/glean_read_api_call '{
  "endpoint": "/search",
  "params": {
    "query": "John Smith",
    "datasource": "people"
  }
}'
```

The Glean people search returns user profiles with email addresses. Extract the email from the results.

**Alternative: Use directory/people endpoint if available:**

```bash
mcp-cli call glean/glean_read_api_call '{
  "endpoint": "/people/search",
  "params": {
    "query": "John Smith"
  }
}'
```

### Step 3: Replace Text Mentions with Person Chips

Once you have the email address, replace the text mention with a proper person chip:

```python
#!/usr/bin/env python3
"""Replace text @mentions with proper person chips."""

import json
import subprocess
import urllib.request

DOC_ID = "YOUR_DOC_ID"
QUOTA_PROJECT = "${GCP_QUOTA_PROJECT}"

# Map of text mentions to email addresses (populated from Glean lookup)
MENTION_TO_EMAIL = {
    "@John Smith": "john.smith@example.com",
    "@Jane Doe": "jane.doe@example.com",
}

def get_token():
    result = subprocess.run(
        ["gcloud", "auth", "application-default", "print-access-token"],
        capture_output=True, text=True
    )
    return result.stdout.strip()

def get_document():
    token = get_token()
    req = urllib.request.Request(
        f"https://docs.googleapis.com/v1/documents/{DOC_ID}",
        headers={"Authorization": f"Bearer {token}", "x-goog-user-project": QUOTA_PROJECT}
    )
    with urllib.request.urlopen(req) as response:
        return json.loads(response.read())

def batch_update(requests):
    token = get_token()
    data = json.dumps({"requests": requests}).encode('utf-8')
    req = urllib.request.Request(
        f"https://docs.googleapis.com/v1/documents/{DOC_ID}:batchUpdate",
        data=data,
        headers={
            "Authorization": f"Bearer {token}",
            "x-goog-user-project": QUOTA_PROJECT,
            "Content-Type": "application/json"
        }
    )
    with urllib.request.urlopen(req) as response:
        return json.loads(response.read())

def find_and_replace_mentions():
    doc = get_document()
    requests = []

    # Find all text mentions and their positions (process in REVERSE order)
    mentions_found = []

    for element in doc.get('body', {}).get('content', []):
        if 'paragraph' in element:
            for text_elem in element['paragraph'].get('elements', []):
                if 'textRun' in text_elem:
                    content = text_elem['textRun'].get('content', '')
                    start_idx = text_elem['startIndex']

                    for mention_text, email in MENTION_TO_EMAIL.items():
                        pos = 0
                        while True:
                            pos = content.find(mention_text, pos)
                            if pos == -1:
                                break
                            mentions_found.append({
                                'start': start_idx + pos,
                                'end': start_idx + pos + len(mention_text),
                                'email': email
                            })
                            pos += 1

    # Sort by start index descending (MUST process in reverse order!)
    mentions_found.sort(key=lambda x: x['start'], reverse=True)

    for mention in mentions_found:
        # Delete the text mention
        requests.append({
            "deleteContentRange": {
                "range": {
                    "startIndex": mention['start'],
                    "endIndex": mention['end']
                }
            }
        })
        # Insert person chip at the same location
        requests.append({
            "insertPerson": {
                "personProperties": {
                    "email": mention['email']
                },
                "location": {"index": mention['start']}
            }
        })

    if requests:
        # Process in batches of 50
        for i in range(0, len(requests), 50):
            batch_update(requests[i:i+50])
            print(f"Processed batch {i//50 + 1}")
        print(f"Replaced {len(mentions_found)} text mentions with person chips")
    else:
        print("No text mentions found to replace")

if __name__ == "__main__":
    find_and_replace_mentions()
```

### Automated Validation Workflow

**Always follow this workflow after creating documents with @ mentions:**

1. **Create the document** with your content
2. **Read the document back** to inspect the structure
3. **Search for text patterns** like `@First Last` that indicate failed mentions
4. **For each text mention found:**
   - Use Glean MCP to search for the person by name
   - Extract their email address from the results
   - Replace the text with a proper `insertPerson` request
5. **Verify the fix** by reading the document again and checking for `person` elements

### Common Issues

**Problem: "@Name" shows as plain text, not a chip**
- The `insertPerson` API requires an email address, not a name
- Always look up the email first using Glean before creating the mention

**Problem: Person chip shows "Unknown user"**
- The email address doesn't exist in the organization
- Verify the email using Glean search before inserting

**Problem: Can't find person in Glean**
- Try variations: "John Smith", "Smith, John", "jsmith"
- Check if person is in a different datasource (employees vs contractors)

## Checklists

### Create Interactive Checkboxes

```json
{
  "requests": [
    {
      "insertText": {
        "location": {"index": 1},
        "text": "Task 1\nTask 2\nTask 3\n"
      }
    },
    {
      "createParagraphBullets": {
        "range": {"startIndex": 1, "endIndex": 21},
        "bulletPreset": "BULLET_CHECKBOX"
      }
    }
  ]
}
```

**Note:** The Google Docs API can CREATE checkboxes but cannot programmatically CHECK/UNCHECK them. Users must click checkboxes manually. To indicate completed items programmatically, use strikethrough styling.

### Checklist with Strikethrough for Completed Items

Use the builder script:

```bash
# Create checklist with items 0 and 1 marked as done (strikethrough)
python3 ~/.claude/plugins/cache/claude-vibe/google-tools/*/skills/google-docs/resources/gdocs_builder.py \
  add-checklist --doc-id "DOC_ID" \
  --items '["Visit the zoo", "Check feeding times", "Take photos"]' \
  --checked '[0, 1]'
```

## Comments

### Add a Comment

```bash
curl -s -X POST "https://www.googleapis.com/drive/v3/files/${DOC_ID}/comments" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "This needs review",
    "anchor": "{\"type\":\"text\",\"start\":{\"index\":10},\"end\":{\"index\":20}}"
  }'
```

## Drive Operations

### List Files

```bash
curl -s "https://www.googleapis.com/drive/v3/files?pageSize=10&fields=files(id,name,mimeType)" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Create Folder

```bash
curl -s -X POST "https://www.googleapis.com/drive/v3/files" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Folder",
    "mimeType": "application/vnd.google-apps.folder"
  }'
```

### Move File to Folder

```bash
curl -s -X PATCH "https://www.googleapis.com/drive/v3/files/${FILE_ID}?addParents=${FOLDER_ID}&removeParents=${OLD_PARENT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Share Document

```bash
curl -s -X POST "https://www.googleapis.com/drive/v3/files/${DOC_ID}/permissions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "user",
    "role": "writer",
    "emailAddress": "user@example.com"
  }'
```

Roles: `reader`, `commenter`, `writer`, `owner`

## Slides Operations

### Create Presentation

```bash
curl -s -X POST "https://slides.googleapis.com/v1/presentations" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{"title": "My Presentation"}'
```

### Add a Slide

```bash
curl -s -X POST "https://slides.googleapis.com/v1/presentations/${PRESENTATION_ID}:batchUpdate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "requests": [{
      "createSlide": {
        "slideLayoutReference": {"predefinedLayout": "TITLE_AND_BODY"}
      }
    }]
  }'
```

Layouts: `BLANK`, `TITLE`, `TITLE_AND_BODY`, `TITLE_AND_TWO_COLUMNS`, `TITLE_ONLY`, `SECTION_HEADER`, etc.

### Create Shape (Text Box, Rectangle, etc.)

```json
{
  "createShape": {
    "objectId": "unique_id",
    "shapeType": "TEXT_BOX",
    "elementProperties": {
      "pageObjectId": "slide_id",
      "size": {
        "width": {"magnitude": 3000000, "unit": "EMU"},
        "height": {"magnitude": 1000000, "unit": "EMU"}
      },
      "transform": {
        "scaleX": 1,
        "scaleY": 1,
        "translateX": 500000,
        "translateY": 500000,
        "unit": "EMU"
      }
    }
  }
}
```

Shape types: `TEXT_BOX`, `RECTANGLE`, `ELLIPSE`, `ARROW_NORTH`, `ARROW_EAST`, `ARROW_SOUTH`, `ARROW_WEST`, etc.

### Insert Image

```json
{
  "createImage": {
    "objectId": "unique_id",
    "url": "https://example.com/image.jpg",
    "elementProperties": {
      "pageObjectId": "slide_id",
      "size": {
        "width": {"magnitude": 3000000, "unit": "EMU"},
        "height": {"magnitude": 2000000, "unit": "EMU"}
      },
      "transform": {
        "scaleX": 1, "scaleY": 1,
        "translateX": 500000, "translateY": 1000000,
        "unit": "EMU"
      }
    }
  }
}
```

### Create Table

```json
{
  "createTable": {
    "objectId": "table_id",
    "rows": 4,
    "columns": 3,
    "elementProperties": {
      "pageObjectId": "slide_id",
      "size": {
        "width": {"magnitude": 8000000, "unit": "EMU"},
        "height": {"magnitude": 2500000, "unit": "EMU"}
      },
      "transform": {
        "scaleX": 1, "scaleY": 1,
        "translateX": 500000, "translateY": 1500000,
        "unit": "EMU"
      }
    }
  }
}
```

### Insert Text into Table Cell

```json
{
  "insertText": {
    "objectId": "table_id",
    "cellLocation": {"rowIndex": 0, "columnIndex": 0},
    "text": "Header",
    "insertionIndex": 0
  }
}
```

### Embed Chart from Google Sheets

```json
{
  "createSheetsChart": {
    "objectId": "chart_id",
    "spreadsheetId": "SHEETS_ID",
    "chartId": 123456789,
    "linkingMode": "LINKED",
    "elementProperties": {
      "pageObjectId": "slide_id",
      "size": {
        "width": {"magnitude": 6000000, "unit": "EMU"},
        "height": {"magnitude": 4000000, "unit": "EMU"}
      },
      "transform": {
        "scaleX": 1, "scaleY": 1,
        "translateX": 1500000, "translateY": 1500000,
        "unit": "EMU"
      }
    }
  }
}
```

Linking modes:
- `LINKED` - Chart updates when Google Sheets data changes
- `NOT_LINKED_IMAGE` - Static snapshot

### Refresh Linked Chart

```json
{"refreshSheetsChart": {"objectId": "chart_id"}}
```

### Duplicate Slide

```json
{
  "duplicateObject": {
    "objectId": "slide_id_to_copy",
    "objectIds": {"slide_id_to_copy": "new_slide_id"}
  }
}
```

### Set Slide Background

```json
{
  "updatePageProperties": {
    "objectId": "slide_id",
    "pageProperties": {
      "pageBackgroundFill": {
        "solidFill": {"color": {"rgbColor": {"red": 0.1, "green": 0.3, "blue": 0.5}}}
      }
    },
    "fields": "pageBackgroundFill"
  }
}
```

### Insert Text with Styling

```json
{
  "requests": [
    {"insertText": {"objectId": "shape_id", "text": "Hello World", "insertionIndex": 0}},
    {"updateTextStyle": {
      "objectId": "shape_id",
      "textRange": {"type": "ALL"},
      "style": {"bold": true, "fontSize": {"magnitude": 24, "unit": "PT"}},
      "fields": "bold,fontSize"
    }}
  ]
}
```

### Create Bullet Points in Slides

```json
{
  "createParagraphBullets": {
    "objectId": "shape_id",
    "textRange": {"type": "ALL"},
    "bulletPreset": "BULLET_DISC_CIRCLE_SQUARE"
  }
}
```

## Helper Scripts

See `/resources/` directory for Python helper scripts:

### gslides_builder.py - Build presentations with proper element management

```bash
# Create a new presentation
python3 gslides_builder.py create --title "My Presentation"

# Get presentation info
python3 gslides_builder.py info --pres-id "PRES_ID"
python3 gslides_builder.py info --pres-id "PRES_ID" --full

# List all slides
python3 gslides_builder.py list-slides --pres-id "PRES_ID"

# Add a slide with layout
python3 gslides_builder.py add-slide --pres-id "PRES_ID" --layout "TITLE_AND_BODY"

# Duplicate a slide
python3 gslides_builder.py duplicate-slide --pres-id "PRES_ID" --page-id "SLIDE_ID"

# Delete a slide
python3 gslides_builder.py delete-slide --pres-id "PRES_ID" --page-id "SLIDE_ID"

# Set slide background
python3 gslides_builder.py set-background --pres-id "PRES_ID" --page-id "SLIDE_ID" \
  --color '{"red": 0.2, "green": 0.4, "blue": 0.6}'

# Add a text box
python3 gslides_builder.py add-text-box --pres-id "PRES_ID" --page-id "SLIDE_ID" \
  --text "Hello World" --x 1 --y 1 --width 3 --height 1 --font-size 24 --bold

# Add an image
python3 gslides_builder.py add-image --pres-id "PRES_ID" --page-id "SLIDE_ID" \
  --url "https://example.com/image.jpg" --x 1 --y 2 --width 4 --height 3

# Add a table with data
python3 gslides_builder.py add-table --pres-id "PRES_ID" --page-id "SLIDE_ID" \
  --rows 4 --cols 3 \
  --data '[["Header1","Header2","Header3"],["A","B","C"],["D","E","F"],["G","H","I"]]'

# Add a chart from Google Sheets
python3 gslides_builder.py add-chart --pres-id "PRES_ID" --page-id "SLIDE_ID" \
  --spreadsheet-id "SHEETS_ID" --chart-id 123456789 \
  --x 1 --y 1.5 --width 6 --height 4

# Copy entire presentation
python3 gslides_builder.py copy --pres-id "PRES_ID" --title "Copy of Presentation"

# Set placeholder text (TITLE, SUBTITLE, BODY)
python3 gslides_builder.py set-placeholder --pres-id "PRES_ID" --page-id "SLIDE_ID" \
  --type "TITLE" --text "My Slide Title"
```

### EMU (English Metric Units) Reference

Slides API uses EMU for positioning:
- 1 inch = 914400 EMU
- 1 point = 12700 EMU
- Standard slide: 10" x 5.625" (16:9 aspect ratio)

### gdocs_builder.py - Build complex documents with proper index management

```bash
# Create a new document
python3 gdocs_builder.py create --title "My Document"

# Read document structure (shows indices)
python3 gdocs_builder.py read --doc-id "DOC_ID"
python3 gdocs_builder.py read --doc-id "DOC_ID" --full  # Full JSON

# Get end index
python3 gdocs_builder.py end-index --doc-id "DOC_ID"

# Add a section with heading
python3 gdocs_builder.py add-section --doc-id "DOC_ID" \
  --heading "Introduction" --text "Content here." --level 1

# Add a table with hyperlinks
python3 gdocs_builder.py add-table --doc-id "DOC_ID" \
  --rows 3 --cols 3 \
  --data '[["A","B","C"],["D","E","F"],["G","H","I"]]' \
  --links '{"0,1": "https://example.com"}'

# Add a person chip (smart chip)
python3 gdocs_builder.py add-person --doc-id "DOC_ID" \
  --email "user@example.com"

# Add a checklist with completed items (strikethrough)
python3 gdocs_builder.py add-checklist --doc-id "DOC_ID" \
  --items '["Task 1", "Task 2", "Task 3"]' \
  --checked '[0]'

# Add a bulleted list
python3 gdocs_builder.py add-bullets --doc-id "DOC_ID" \
  --items '["Point 1", "Point 2", "Point 3"]' \
  --preset "BULLET_DISC_CIRCLE_SQUARE"

# Apply strikethrough to a text range
python3 gdocs_builder.py strikethrough --doc-id "DOC_ID" \
  --start 10 --end 25
```

### markdown_to_gdocs.py - Convert markdown files to Google Docs

```bash
# Create new doc from markdown
python3 markdown_to_gdocs.py --input /path/to/file.md --title "Doc Title"

# Append to existing doc
python3 markdown_to_gdocs.py --input /path/to/file.md --doc-id "DOC_ID"
```

### Authentication (via google-auth skill)

For authentication, use the shared `google_auth.py` module from the `google-auth` skill:

```bash
# IMPORTANT: Always capture token in a variable, never print directly
TOKEN=$(python3 ../google-auth/resources/google_auth.py token)

# Check auth status
python3 ../google-auth/resources/google_auth.py status

# Login with required scopes
python3 ../google-auth/resources/google_auth.py login

# Validate current token
python3 ../google-auth/resources/google_auth.py validate
```

**CRITICAL:** Always use `TOKEN=$(...)` to capture the token. Never run the token command without capturing its output, as this will print sensitive credentials to the terminal.
