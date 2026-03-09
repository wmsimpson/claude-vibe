#!/usr/bin/env python3
"""
Markdown to Google Docs Converter

Converts markdown files to well-formatted Google Docs, preserving:
- Headings (# to ######)
- Bold and italic text
- Hyperlinks [text](url)
- Bullet lists (- or *)
- Numbered lists (1. 2. 3.)
- Code blocks (inline and fenced)
- Tables (pipe syntax)
- Blockquotes (>)
- Horizontal rules (---)

Usage:
    python3 markdown_to_gdocs.py --input README.md --title "My Document"
    python3 markdown_to_gdocs.py --input README.md --doc-id "existing_doc_id"
"""

import argparse
import json
import os
import re
import sys
from dataclasses import dataclass
from typing import Dict, List, Optional, Tuple, Any

# google_api_utils lives in the google-auth skill. Scripts outside that directory
# must resolve the path explicitly — this is the established pattern (see evals/tests/test_google_api_utils.py).
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "../../google-auth/resources"))
from google_api_utils import api_call_with_retry


@dataclass
class TextSpan:
    """A span of text with styling."""
    text: str
    bold: bool = False
    italic: bool = False
    code: bool = False
    link_url: Optional[str] = None
    strikethrough: bool = False


@dataclass
class Paragraph:
    """A paragraph with style and content."""
    spans: List[TextSpan]
    style: str = "NORMAL_TEXT"  # HEADING_1 through HEADING_6, NORMAL_TEXT
    bullet: bool = False
    numbered: bool = False
    list_level: int = 0
    blockquote: bool = False


@dataclass
class TableCell:
    """A table cell with content and optional link."""
    text: str
    link_url: Optional[str] = None
    bold: bool = False


@dataclass
class Table:
    """A table with rows and columns."""
    rows: List[List[TableCell]]


def api_call(method: str, url: str, data: Optional[Dict] = None) -> Dict:
    """Make an API call using curl with retry logic."""
    return api_call_with_retry(method, url, data=data)


def parse_inline_formatting(text: str) -> List[TextSpan]:
    """Parse inline markdown formatting (bold, italic, links, code)."""
    spans = []

    # Regex patterns for inline formatting
    # Order matters: more specific patterns first
    patterns = [
        # Links: [text](url)
        (r'\[([^\]]+)\]\(([^)]+)\)', lambda m: TextSpan(m.group(1), link_url=m.group(2))),
        # Bold+Italic: ***text*** or ___text___
        (r'\*\*\*(.+?)\*\*\*', lambda m: TextSpan(m.group(1), bold=True, italic=True)),
        (r'___(.+?)___', lambda m: TextSpan(m.group(1), bold=True, italic=True)),
        # Bold: **text** or __text__
        (r'\*\*(.+?)\*\*', lambda m: TextSpan(m.group(1), bold=True)),
        (r'__(.+?)__', lambda m: TextSpan(m.group(1), bold=True)),
        # Italic: *text* or _text_
        (r'\*(.+?)\*', lambda m: TextSpan(m.group(1), italic=True)),
        (r'_(.+?)_', lambda m: TextSpan(m.group(1), italic=True)),
        # Strikethrough: ~~text~~
        (r'~~(.+?)~~', lambda m: TextSpan(m.group(1), strikethrough=True)),
        # Inline code: `code`
        (r'`([^`]+)`', lambda m: TextSpan(m.group(1), code=True)),
    ]

    # Process text character by character, finding formatted spans
    pos = 0
    while pos < len(text):
        found = False
        for pattern, handler in patterns:
            match = re.match(pattern, text[pos:])
            if match:
                # Add any plain text before this match
                if match.start() > 0:
                    spans.append(TextSpan(text[pos:pos + match.start()]))

                # Add the formatted span
                spans.append(handler(match))
                pos += match.end()
                found = True
                break

        if not found:
            # Check if any pattern starts later in the remaining text
            next_match_pos = len(text)
            for pattern, _ in patterns:
                search = re.search(pattern, text[pos:])
                if search and pos + search.start() < next_match_pos:
                    next_match_pos = pos + search.start()

            if next_match_pos > pos:
                spans.append(TextSpan(text[pos:next_match_pos]))
                pos = next_match_pos
            else:
                # No more patterns found, add rest as plain text
                spans.append(TextSpan(text[pos:]))
                break

    # Combine adjacent spans with same formatting
    if not spans:
        spans = [TextSpan(text)]

    return spans


def parse_markdown(content: str) -> List[Any]:
    """Parse markdown content into structured elements."""
    lines = content.split('\n')
    elements = []
    i = 0

    while i < len(lines):
        line = lines[i]

        # Skip empty lines
        if not line.strip():
            i += 1
            continue

        # Horizontal rule
        if re.match(r'^(-{3,}|\*{3,}|_{3,})$', line.strip()):
            elements.append({"type": "hr"})
            i += 1
            continue

        # Headings: # through ######
        heading_match = re.match(r'^(#{1,6})\s+(.+)$', line)
        if heading_match:
            level = len(heading_match.group(1))
            text = heading_match.group(2)
            spans = parse_inline_formatting(text)
            elements.append(Paragraph(spans=spans, style=f"HEADING_{level}"))
            i += 1
            continue

        # Fenced code block
        if line.strip().startswith('```'):
            code_lines = []
            language = line.strip()[3:].strip()
            i += 1
            while i < len(lines) and not lines[i].strip().startswith('```'):
                code_lines.append(lines[i])
                i += 1
            i += 1  # Skip closing ```

            code_text = '\n'.join(code_lines)
            elements.append({
                "type": "code_block",
                "language": language,
                "code": code_text
            })
            continue

        # Blockquote
        if line.startswith('>'):
            quote_text = line[1:].strip()
            spans = parse_inline_formatting(quote_text)
            elements.append(Paragraph(spans=spans, blockquote=True))
            i += 1
            continue

        # Table
        if '|' in line and i + 1 < len(lines) and re.match(r'^[\s|:-]+$', lines[i + 1]):
            table_rows = []

            # Parse header row
            header_cells = [cell.strip() for cell in line.split('|')[1:-1]]
            table_rows.append([TableCell(text=cell, bold=True) for cell in header_cells])

            i += 2  # Skip header and separator

            # Parse data rows
            while i < len(lines) and '|' in lines[i]:
                cells = [cell.strip() for cell in lines[i].split('|')[1:-1]]
                row = []
                for cell in cells:
                    # Check for bold text: **text** or __text__
                    bold_match = re.match(r'^\*\*(.+?)\*\*$', cell) or re.match(r'^__(.+?)__$', cell)
                    # Check for links in cells: [text](url)
                    link_match = re.match(r'\[([^\]]+)\]\(([^)]+)\)', cell)
                    # Check for bold link: **[text](url)**
                    bold_link_match = re.match(r'^\*\*\[([^\]]+)\]\(([^)]+)\)\*\*$', cell)

                    if bold_link_match:
                        row.append(TableCell(text=bold_link_match.group(1), link_url=bold_link_match.group(2), bold=True))
                    elif link_match:
                        row.append(TableCell(text=link_match.group(1), link_url=link_match.group(2)))
                    elif bold_match:
                        row.append(TableCell(text=bold_match.group(1), bold=True))
                    else:
                        row.append(TableCell(text=cell))
                table_rows.append(row)
                i += 1

            elements.append(Table(rows=table_rows))
            continue

        # Bullet list
        bullet_match = re.match(r'^(\s*)[-*+]\s+(.+)$', line)
        if bullet_match:
            indent = len(bullet_match.group(1))
            level = indent // 2
            text = bullet_match.group(2)
            spans = parse_inline_formatting(text)
            elements.append(Paragraph(spans=spans, bullet=True, list_level=level))
            i += 1
            continue

        # Numbered list
        numbered_match = re.match(r'^(\s*)\d+\.\s+(.+)$', line)
        if numbered_match:
            indent = len(numbered_match.group(1))
            level = indent // 2
            text = numbered_match.group(2)
            spans = parse_inline_formatting(text)
            elements.append(Paragraph(spans=spans, numbered=True, list_level=level))
            i += 1
            continue

        # Regular paragraph
        spans = parse_inline_formatting(line)
        elements.append(Paragraph(spans=spans))
        i += 1

    return elements


def build_paragraph_requests(
    para: Paragraph,
    start_index: int
) -> Tuple[List[Dict], int]:
    """
    Build requests for a paragraph, return (requests, new_index).
    """
    requests = []

    # Build the full text
    full_text = ""
    for span in para.spans:
        full_text += span.text
    full_text += "\n"

    # Insert text
    requests.append({
        "insertText": {
            "location": {"index": start_index},
            "text": full_text
        }
    })

    # Track position for styling
    pos = start_index
    for span in para.spans:
        end_pos = pos + len(span.text)

        style_fields = []
        text_style = {}

        if span.bold:
            text_style["bold"] = True
            style_fields.append("bold")

        if span.italic:
            text_style["italic"] = True
            style_fields.append("italic")

        if span.strikethrough:
            text_style["strikethrough"] = True
            style_fields.append("strikethrough")

        if span.link_url:
            text_style["link"] = {"url": span.link_url}
            style_fields.append("link")

        if span.code:
            text_style["weightedFontFamily"] = {"fontFamily": "Roboto Mono", "weight": 400}
            text_style["backgroundColor"] = {"color": {"rgbColor": {"red": 0.95, "green": 0.95, "blue": 0.95}}}
            style_fields.extend(["weightedFontFamily", "backgroundColor"])

        if style_fields and end_pos > pos:
            requests.append({
                "updateTextStyle": {
                    "range": {"startIndex": pos, "endIndex": end_pos},
                    "textStyle": text_style,
                    "fields": ",".join(style_fields)
                }
            })

        pos = end_pos

    # Apply paragraph style
    para_end = start_index + len(full_text)

    if para.style != "NORMAL_TEXT":
        requests.append({
            "updateParagraphStyle": {
                "range": {"startIndex": start_index, "endIndex": para_end},
                "paragraphStyle": {"namedStyleType": para.style},
                "fields": "namedStyleType"
            }
        })

    if para.bullet:
        requests.append({
            "createParagraphBullets": {
                "range": {"startIndex": start_index, "endIndex": para_end},
                "bulletPreset": "BULLET_DISC_CIRCLE_SQUARE"
            }
        })

    if para.numbered:
        requests.append({
            "createParagraphBullets": {
                "range": {"startIndex": start_index, "endIndex": para_end},
                "bulletPreset": "NUMBERED_DECIMAL_NESTED"
            }
        })

    if para.blockquote:
        requests.append({
            "updateParagraphStyle": {
                "range": {"startIndex": start_index, "endIndex": para_end},
                "paragraphStyle": {
                    "indentStart": {"magnitude": 36, "unit": "PT"},
                    "borderLeft": {
                        "color": {"color": {"rgbColor": {"red": 0.8, "green": 0.8, "blue": 0.8}}},
                        "width": {"magnitude": 2, "unit": "PT"},
                        "padding": {"magnitude": 8, "unit": "PT"},
                        "dashStyle": "SOLID"
                    }
                },
                "fields": "indentStart,borderLeft"
            }
        })

    return requests, para_end


def get_document_end_index(doc_id: str) -> int:
    """Get the current end index of a document."""
    doc = api_call("GET", f"https://docs.googleapis.com/v1/documents/{doc_id}")
    content = doc.get("body", {}).get("content", [])
    if not content:
        return 1
    return content[-1].get("endIndex", 1) - 1


def calculate_table_cell_index(table_start: int, row: int, col: int, num_cols: int) -> int:
    """Calculate the content index for a table cell."""
    return table_start + 3 + row * (num_cols * 2 + 1) + col * 2


def convert_markdown_to_gdocs(
    markdown_content: str,
    doc_id: Optional[str] = None,
    title: Optional[str] = None
) -> Dict:
    """
    Convert markdown to Google Docs.

    Args:
        markdown_content: The markdown text to convert
        doc_id: Existing document ID (optional)
        title: Title for new document (required if doc_id not provided)

    Returns:
        Dict with documentId and url
    """
    # Create document if needed
    if not doc_id:
        if not title:
            title = "Converted Document"
        response = api_call("POST", "https://docs.googleapis.com/v1/documents", {"title": title})
        if "error" in response:
            raise RuntimeError(f"Failed to create document: {response['error']['message']}")
        doc_id = response["documentId"]

    # Parse markdown
    elements = parse_markdown(markdown_content)

    # Group elements into batches, splitting at tables
    # Tables need special handling - we must commit previous work before inserting
    batches = []
    current_batch = []

    for element in elements:
        if isinstance(element, Table):
            # Commit current batch before table
            if current_batch:
                batches.append(("text", current_batch))
                current_batch = []
            # Add table as its own batch
            batches.append(("table", element))
        else:
            current_batch.append(element)

    # Don't forget the last batch
    if current_batch:
        batches.append(("text", current_batch))

    # Process each batch
    for batch_type, batch_content in batches:
        # Get current end index before each batch
        current_index = get_document_end_index(doc_id)

        if batch_type == "text":
            all_requests = []

            for element in batch_content:
                if isinstance(element, Paragraph):
                    requests, current_index = build_paragraph_requests(element, current_index)
                    all_requests.extend(requests)

                elif isinstance(element, dict):
                    if element.get("type") == "hr":
                        hr_text = "———\n"
                        all_requests.append({
                            "insertText": {
                                "location": {"index": current_index},
                                "text": hr_text
                            }
                        })
                        current_index += len(hr_text)

                    elif element.get("type") == "code_block":
                        code = element["code"] + "\n\n"  # Add extra newline for spacing
                        all_requests.append({
                            "insertText": {
                                "location": {"index": current_index},
                                "text": code
                            }
                        })

                        # Style as code (excluding trailing newlines)
                        code_end = current_index + len(code) - 2
                        if code_end > current_index:
                            all_requests.append({
                                "updateTextStyle": {
                                    "range": {"startIndex": current_index, "endIndex": code_end},
                                    "textStyle": {
                                        "weightedFontFamily": {"fontFamily": "Roboto Mono", "weight": 400},
                                        "fontSize": {"magnitude": 10, "unit": "PT"},
                                        "backgroundColor": {"color": {"rgbColor": {"red": 0.95, "green": 0.95, "blue": 0.95}}}
                                    },
                                    "fields": "weightedFontFamily,fontSize,backgroundColor"
                                }
                            })

                        current_index += len(code)

            # Execute text batch
            if all_requests:
                result = api_call(
                    "POST",
                    f"https://docs.googleapis.com/v1/documents/{doc_id}:batchUpdate",
                    {"requests": all_requests}
                )
                if "error" in result:
                    print(f"Warning: Text batch error: {result['error']['message']}", file=sys.stderr)

        elif batch_type == "table":
            table = batch_content
            rows = len(table.rows)
            cols = len(table.rows[0]) if table.rows else 1

            # Insert the empty table
            result = api_call(
                "POST",
                f"https://docs.googleapis.com/v1/documents/{doc_id}:batchUpdate",
                {"requests": [{
                    "insertTable": {
                        "rows": rows,
                        "columns": cols,
                        "location": {"index": current_index}
                    }
                }]}
            )

            if "error" in result:
                print(f"Warning: Table insert error: {result['error']['message']}", file=sys.stderr)
                continue

            # Read document to find the table's actual start index
            doc = api_call("GET", f"https://docs.googleapis.com/v1/documents/{doc_id}")
            content = doc.get("body", {}).get("content", [])

            # Find the table we just inserted
            table_start = None
            for elem in content:
                if "table" in elem and elem.get("startIndex", 0) >= current_index:
                    table_start = elem["startIndex"]
                    break

            if table_start is None:
                print("Warning: Could not find inserted table", file=sys.stderr)
                continue

            # Build cell content requests in REVERSE order
            cell_requests = []
            cells_to_insert = []

            for r in range(rows):
                for c in range(cols):
                    if r < len(table.rows) and c < len(table.rows[r]):
                        cell = table.rows[r][c]
                        if cell.text:
                            idx = calculate_table_cell_index(table_start, r, c, cols)
                            cells_to_insert.append((idx, cell))

            # Sort by index DESCENDING
            cells_to_insert.sort(key=lambda x: x[0], reverse=True)

            for idx, cell in cells_to_insert:
                cell_requests.append({
                    "insertText": {
                        "location": {"index": idx},
                        "text": cell.text
                    }
                })

                # Apply hyperlink if present
                if cell.link_url:
                    cell_requests.append({
                        "updateTextStyle": {
                            "range": {"startIndex": idx, "endIndex": idx + len(cell.text)},
                            "textStyle": {"link": {"url": cell.link_url}},
                            "fields": "link"
                        }
                    })

                # Apply bold for headers
                if cell.bold:
                    cell_requests.append({
                        "updateTextStyle": {
                            "range": {"startIndex": idx, "endIndex": idx + len(cell.text)},
                            "textStyle": {"bold": True},
                            "fields": "bold"
                        }
                    })

            # Execute cell content
            if cell_requests:
                result = api_call(
                    "POST",
                    f"https://docs.googleapis.com/v1/documents/{doc_id}:batchUpdate",
                    {"requests": cell_requests}
                )
                if "error" in result:
                    print(f"Warning: Table content error: {result['error']['message']}", file=sys.stderr)

    return {
        "documentId": doc_id,
        "url": f"https://docs.google.com/document/d/{doc_id}/edit"
    }


def main():
    parser = argparse.ArgumentParser(
        description="Convert markdown files to Google Docs",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    parser.add_argument("--input", "-i", required=True, help="Input markdown file")
    parser.add_argument("--title", "-t", help="Document title (for new documents)")
    parser.add_argument("--doc-id", "-d", help="Existing document ID to append to")
    parser.add_argument("--output", "-o", help="Output file for result JSON")

    args = parser.parse_args()

    try:
        # Read markdown file
        with open(args.input, 'r') as f:
            markdown_content = f.read()

        # Use filename as title if not provided
        title = args.title
        if not title and not args.doc_id:
            title = os.path.splitext(os.path.basename(args.input))[0]

        # Convert
        result = convert_markdown_to_gdocs(
            markdown_content,
            doc_id=args.doc_id,
            title=title
        )

        output = json.dumps(result, indent=2)

        if args.output:
            with open(args.output, 'w') as f:
                f.write(output)
            print(f"Result written to {args.output}", file=sys.stderr)
        else:
            print(output)

    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
