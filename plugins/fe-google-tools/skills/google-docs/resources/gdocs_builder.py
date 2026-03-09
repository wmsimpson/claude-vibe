#!/usr/bin/env python3
"""
Google Docs Builder - Build complex documents with proper index management

This script helps create well-formatted Google Docs by:
1. Reading current document structure
2. Calculating correct indices
3. Generating batchUpdate requests in correct order

Usage:
    # Create a new document
    python3 gdocs_builder.py create --title "My Document"

    # Read document structure
    python3 gdocs_builder.py read --doc-id "DOC_ID"

    # Add content with heading
    python3 gdocs_builder.py add-section --doc-id "DOC_ID" --heading "Introduction" --text "This is the intro."

    # Add a table
    python3 gdocs_builder.py add-table --doc-id "DOC_ID" --rows 3 --cols 3 --data '[["A","B","C"],...]'
"""

import argparse
import json
import sys
from typing import Dict, List, Optional, Tuple

from google_api_utils import api_call_with_retry


def api_call(method: str, url: str, data: Optional[Dict] = None) -> Dict:
    """Make an API call using curl with retry logic."""
    return api_call_with_retry(method, url, data=data)


def create_document(title: str) -> str:
    """Create a new Google Doc and return its ID."""
    response = api_call("POST", "https://docs.googleapis.com/v1/documents", {"title": title})

    if "error" in response:
        raise RuntimeError(f"Failed to create document: {response['error']['message']}")

    return response["documentId"]


def read_document(doc_id: str) -> Dict:
    """Read a document and return its full structure."""
    return api_call("GET", f"https://docs.googleapis.com/v1/documents/{doc_id}")


def get_document_end_index(doc_id: str) -> int:
    """Get the index at the end of the document."""
    doc = read_document(doc_id)
    content = doc.get("body", {}).get("content", [])

    if not content:
        return 1

    # Find the last element's end index
    last_element = content[-1]
    return last_element.get("endIndex", 1) - 1


def get_document_structure(doc_id: str) -> List[Dict]:
    """Get simplified document structure with indices and content."""
    doc = read_document(doc_id)
    content = doc.get("body", {}).get("content", [])

    structure = []
    for element in content:
        if "paragraph" in element:
            para = element["paragraph"]
            text = ""
            for elem in para.get("elements", []):
                if "textRun" in elem:
                    text += elem["textRun"]["content"]

            structure.append({
                "type": "paragraph",
                "startIndex": element["startIndex"],
                "endIndex": element["endIndex"],
                "style": para.get("paragraphStyle", {}).get("namedStyleType", "NORMAL_TEXT"),
                "text": text
            })
        elif "table" in element:
            table = element["table"]
            structure.append({
                "type": "table",
                "startIndex": element["startIndex"],
                "endIndex": element["endIndex"],
                "rows": table["rows"],
                "columns": table["columns"]
            })

    return structure


def batch_update(doc_id: str, requests: List[Dict]) -> Dict:
    """Execute a batchUpdate on the document."""
    return api_call(
        "POST",
        f"https://docs.googleapis.com/v1/documents/{doc_id}:batchUpdate",
        {"requests": requests}
    )


def build_text_insert_requests(
    index: int,
    text: str,
    style: Optional[str] = None,
    link_url: Optional[str] = None,
    bold: bool = False,
    italic: bool = False,
    strikethrough: bool = False
) -> List[Dict]:
    """
    Build requests to insert text with optional styling.

    IMPORTANT: Returns requests in the order they should be executed.
    The caller is responsible for combining multiple inserts in reverse index order.
    """
    requests = []

    # Insert text
    requests.append({
        "insertText": {
            "location": {"index": index},
            "text": text
        }
    })

    end_index = index + len(text)

    # Apply paragraph style if specified (for headings)
    if style and style != "NORMAL_TEXT":
        # Find the paragraph boundaries (from index to next newline or end)
        para_end = index + text.find("\n") + 1 if "\n" in text else end_index
        requests.append({
            "updateParagraphStyle": {
                "range": {"startIndex": index, "endIndex": para_end},
                "paragraphStyle": {"namedStyleType": style},
                "fields": "namedStyleType"
            }
        })

    # Apply text style if needed
    text_style = {}
    fields = []

    if link_url:
        text_style["link"] = {"url": link_url}
        fields.append("link")

    if bold:
        text_style["bold"] = True
        fields.append("bold")

    if italic:
        text_style["italic"] = True
        fields.append("italic")

    if strikethrough:
        text_style["strikethrough"] = True
        fields.append("strikethrough")

    if text_style:
        # Don't include trailing newline in text style range
        style_end = end_index
        if text.endswith("\n"):
            style_end -= 1

        if style_end > index:
            requests.append({
                "updateTextStyle": {
                    "range": {"startIndex": index, "endIndex": style_end},
                    "textStyle": text_style,
                    "fields": ",".join(fields)
                }
            })

    return requests


def calculate_table_cell_index(table_start: int, row: int, col: int, num_cols: int) -> int:
    """Calculate the content index for a table cell."""
    return table_start + 3 + row * (num_cols * 2 + 1) + col * 2


def build_table_insert_requests(
    table_start_index: int,
    rows: int,
    cols: int,
    data: List[List[str]],
    links: Optional[Dict[str, str]] = None,
    header_style: bool = True
) -> List[Dict]:
    """
    Build requests to fill table content.

    Args:
        table_start_index: Index where table starts
        rows: Number of rows
        cols: Number of columns
        data: 2D array of cell values
        links: Dict mapping "row,col" to URL for hyperlinks
        header_style: Apply bold to first row

    Returns requests in REVERSE order (highest index first) to prevent drift.
    """
    links = links or {}
    requests = []

    # Collect all cell operations
    cells = []
    for r in range(rows):
        for c in range(cols):
            if r < len(data) and c < len(data[r]):
                text = str(data[r][c])
                if text:
                    index = calculate_table_cell_index(table_start_index, r, c, cols)
                    link_key = f"{r},{c}"
                    link_url = links.get(link_key)
                    is_header = header_style and r == 0
                    cells.append((index, text, link_url, is_header))

    # Sort by index descending (highest first)
    cells.sort(key=lambda x: x[0], reverse=True)

    # Build requests
    for index, text, link_url, is_header in cells:
        # Insert text
        requests.append({
            "insertText": {
                "location": {"index": index},
                "text": text
            }
        })

        # Apply hyperlink if provided
        if link_url:
            requests.append({
                "updateTextStyle": {
                    "range": {"startIndex": index, "endIndex": index + len(text)},
                    "textStyle": {"link": {"url": link_url}},
                    "fields": "link"
                }
            })

        # Apply bold for header
        if is_header:
            requests.append({
                "updateTextStyle": {
                    "range": {"startIndex": index, "endIndex": index + len(text)},
                    "textStyle": {"bold": True},
                    "fields": "bold"
                }
            })

    return requests


def add_section(doc_id: str, heading: str, text: str, heading_level: int = 1) -> Dict:
    """Add a new section with heading and content at the end of the document."""
    end_index = get_document_end_index(doc_id)

    heading_style = f"HEADING_{heading_level}" if heading_level <= 6 else "HEADING_1"

    # Build the content
    full_text = f"{heading}\n{text}\n"

    requests = build_text_insert_requests(
        index=end_index,
        text=full_text,
        style=heading_style
    )

    # The heading style only applies to the heading line
    # We need to reset the rest to normal
    heading_end = end_index + len(heading) + 1
    text_start = heading_end
    text_end = end_index + len(full_text)

    if text_end > text_start:
        requests.append({
            "updateParagraphStyle": {
                "range": {"startIndex": text_start, "endIndex": text_end},
                "paragraphStyle": {"namedStyleType": "NORMAL_TEXT"},
                "fields": "namedStyleType"
            }
        })

    return batch_update(doc_id, requests)


def add_table(
    doc_id: str,
    rows: int,
    cols: int,
    data: List[List[str]],
    links: Optional[Dict[str, str]] = None
) -> Dict:
    """Add a table at the end of the document and fill it with data."""
    end_index = get_document_end_index(doc_id)

    # First, insert the table
    table_request = {
        "insertTable": {
            "rows": rows,
            "columns": cols,
            "location": {"index": end_index}
        }
    }

    result = batch_update(doc_id, [table_request])

    if "error" in result:
        return result

    # Read document to get the new table's start index
    doc = read_document(doc_id)
    content = doc.get("body", {}).get("content", [])

    # Find the table we just inserted
    table_start = None
    for element in content:
        if "table" in element and element["startIndex"] >= end_index:
            table_start = element["startIndex"]
            break

    if table_start is None:
        raise RuntimeError("Could not find inserted table")

    # Fill the table
    fill_requests = build_table_insert_requests(
        table_start, rows, cols, data, links
    )

    return batch_update(doc_id, fill_requests)


def add_person(doc_id: str, email: str, index: Optional[int] = None) -> Dict:
    """
    Insert a person chip (smart chip) for the given email at the specified index.
    If index is None, appends at end of document.

    Note: Person chips are interactive elements that show profile info on hover.
    """
    if index is None:
        index = get_document_end_index(doc_id)

    request = {
        "insertPerson": {
            "personProperties": {
                "email": email
            },
            "location": {"index": index}
        }
    }

    return batch_update(doc_id, [request])


def add_checklist(doc_id: str, items: List[str], checked_indices: Optional[List[int]] = None) -> Dict:
    """
    Add a checklist at the end of the document.

    Args:
        doc_id: Document ID
        items: List of checklist item texts
        checked_indices: List of indices (0-based) of items to strikethrough as "done"
                        Note: Google Docs API cannot programmatically check boxes,
                        but we can apply strikethrough to indicate completion.

    Returns:
        API response
    """
    checked_indices = checked_indices or []
    end_index = get_document_end_index(doc_id)

    # Build content string
    content = "\n".join(items) + "\n"

    # Insert the text
    insert_request = {
        "insertText": {
            "location": {"index": end_index},
            "text": content
        }
    }

    result = batch_update(doc_id, [insert_request])
    if "error" in result:
        return result

    # Get updated indices
    doc = read_document(doc_id)
    body_content = doc.get("body", {}).get("content", [])

    # Find the range of our inserted content
    content_start = end_index
    content_end = end_index + len(content)

    # Apply checkbox bullets
    bullet_request = {
        "createParagraphBullets": {
            "range": {"startIndex": content_start, "endIndex": content_end - 1},
            "bulletPreset": "BULLET_CHECKBOX"
        }
    }

    result = batch_update(doc_id, [bullet_request])
    if "error" in result:
        return result

    # Apply strikethrough to "checked" items
    if checked_indices:
        strikethrough_requests = []
        current_idx = content_start

        for i, item in enumerate(items):
            if i in checked_indices:
                strikethrough_requests.append({
                    "updateTextStyle": {
                        "range": {"startIndex": current_idx, "endIndex": current_idx + len(item)},
                        "textStyle": {"strikethrough": True},
                        "fields": "strikethrough"
                    }
                })
            current_idx += len(item) + 1  # +1 for newline

        if strikethrough_requests:
            result = batch_update(doc_id, strikethrough_requests)

    return result


def add_bullet_list(doc_id: str, items: List[str], preset: str = "BULLET_DISC_CIRCLE_SQUARE") -> Dict:
    """
    Add a bulleted list at the end of the document.

    Args:
        doc_id: Document ID
        items: List of bullet point texts
        preset: Bullet style preset. Options:
                - BULLET_DISC_CIRCLE_SQUARE (default)
                - BULLET_DIAMONDX_ARROW3D_SQUARE
                - BULLET_CHECKBOX (creates checkboxes)
                - BULLET_ARROW_DIAMOND_DISC
                - NUMBERED_DECIMAL_ALPHA_ROMAN
                - NUMBERED_DECIMAL_ALPHA_ROMAN_PARENS
                - NUMBERED_DECIMAL_NESTED
                - NUMBERED_UPPERALPHA_ALPHA_ROMAN
                - NUMBERED_UPPERROMAN_UPPERALPHA_DECIMAL
                - NUMBERED_ZERODECIMAL_ALPHA_ROMAN

    Returns:
        API response
    """
    end_index = get_document_end_index(doc_id)

    # Build content string
    content = "\n".join(items) + "\n"

    # Insert the text
    insert_request = {
        "insertText": {
            "location": {"index": end_index},
            "text": content
        }
    }

    result = batch_update(doc_id, [insert_request])
    if "error" in result:
        return result

    # Apply bullet formatting
    content_start = end_index
    content_end = end_index + len(content)

    bullet_request = {
        "createParagraphBullets": {
            "range": {"startIndex": content_start, "endIndex": content_end - 1},
            "bulletPreset": preset
        }
    }

    return batch_update(doc_id, [bullet_request])


def apply_strikethrough(doc_id: str, start_index: int, end_index: int) -> Dict:
    """Apply strikethrough formatting to a range of text."""
    request = {
        "updateTextStyle": {
            "range": {"startIndex": start_index, "endIndex": end_index},
            "textStyle": {"strikethrough": True},
            "fields": "strikethrough"
        }
    }
    return batch_update(doc_id, [request])


def add_link(doc_id: str, start_index: int, end_index: int, url: str) -> Dict:
    """Apply hyperlink to a range of text.

    Args:
        doc_id: Document ID
        start_index: Start index of text to link
        end_index: End index of text to link
        url: URL for the hyperlink
    """
    request = {
        "updateTextStyle": {
            "range": {"startIndex": start_index, "endIndex": end_index},
            "textStyle": {"link": {"url": url}},
            "fields": "link"
        }
    }
    return batch_update(doc_id, [request])


def get_document_text(doc_id: str) -> str:
    """Get full text content of a document."""
    doc = read_document(doc_id)
    full_text = ""

    for element in doc.get('body', {}).get('content', []):
        if 'paragraph' in element:
            for elem in element['paragraph'].get('elements', []):
                if 'textRun' in elem:
                    full_text += elem['textRun'].get('content', '')

    return full_text


def find_text_positions(doc_id: str, search_text: str) -> List[Tuple[int, int]]:
    """Find all positions of text in document.

    Args:
        doc_id: Document ID
        search_text: Text to search for

    Returns:
        List of (start_index, end_index) tuples (1-indexed for Docs API)
    """
    full_text = get_document_text(doc_id)

    positions = []
    start = 0
    while True:
        pos = full_text.find(search_text, start)
        if pos == -1:
            break
        # +1 because docs API is 1-indexed
        positions.append((pos + 1, pos + 1 + len(search_text)))
        start = pos + 1

    return positions


def add_links_by_text(doc_id: str, links: Dict[str, str]) -> int:
    """Add hyperlinks to all occurrences of specified text.

    Args:
        doc_id: Document ID
        links: Dict mapping text to URL, e.g. {"#channel-name": "https://slack.com/..."}

    Returns:
        Number of links added
    """
    link_requests = []

    for text, url in links.items():
        positions = find_text_positions(doc_id, text)
        for start, end in positions:
            link_requests.append({
                "updateTextStyle": {
                    "range": {"startIndex": start, "endIndex": end},
                    "textStyle": {"link": {"url": url}},
                    "fields": "link"
                }
            })

    # Apply in batches of 50
    if link_requests:
        for i in range(0, len(link_requests), 50):
            batch_update(doc_id, link_requests[i:i+50])

    return len(link_requests)


def main():
    parser = argparse.ArgumentParser(
        description="Build and manage Google Docs",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    subparsers = parser.add_subparsers(dest="command", required=True)

    # Create command
    create_parser = subparsers.add_parser("create", help="Create a new document")
    create_parser.add_argument("--title", required=True, help="Document title")

    # Read command
    read_parser = subparsers.add_parser("read", help="Read document structure")
    read_parser.add_argument("--doc-id", required=True, help="Document ID")
    read_parser.add_argument("--full", action="store_true", help="Show full JSON")

    # Add section command
    section_parser = subparsers.add_parser("add-section", help="Add a section with heading")
    section_parser.add_argument("--doc-id", required=True, help="Document ID")
    section_parser.add_argument("--heading", required=True, help="Section heading")
    section_parser.add_argument("--text", required=True, help="Section content")
    section_parser.add_argument("--level", type=int, default=1, help="Heading level (1-6)")

    # Add table command
    table_parser = subparsers.add_parser("add-table", help="Add a table")
    table_parser.add_argument("--doc-id", required=True, help="Document ID")
    table_parser.add_argument("--rows", type=int, required=True, help="Number of rows")
    table_parser.add_argument("--cols", type=int, required=True, help="Number of columns")
    table_parser.add_argument("--data", required=True, help="JSON 2D array of cell values")
    table_parser.add_argument("--links", default="{}", help="JSON object mapping 'row,col' to URL")

    # Get end index command
    end_parser = subparsers.add_parser("end-index", help="Get document end index")
    end_parser.add_argument("--doc-id", required=True, help="Document ID")

    # Add person chip command
    person_parser = subparsers.add_parser("add-person", help="Add a person chip (smart chip)")
    person_parser.add_argument("--doc-id", required=True, help="Document ID")
    person_parser.add_argument("--email", required=True, help="Email address of person")
    person_parser.add_argument("--index", type=int, help="Insert at index (default: end)")

    # Add checklist command
    checklist_parser = subparsers.add_parser("add-checklist", help="Add a checklist")
    checklist_parser.add_argument("--doc-id", required=True, help="Document ID")
    checklist_parser.add_argument("--items", required=True, help="JSON array of checklist items")
    checklist_parser.add_argument("--checked", default="[]", help="JSON array of indices to mark as done (strikethrough)")

    # Add bullet list command
    bullets_parser = subparsers.add_parser("add-bullets", help="Add a bulleted list")
    bullets_parser.add_argument("--doc-id", required=True, help="Document ID")
    bullets_parser.add_argument("--items", required=True, help="JSON array of bullet items")
    bullets_parser.add_argument("--preset", default="BULLET_DISC_CIRCLE_SQUARE", help="Bullet preset style")

    # Apply strikethrough command
    strike_parser = subparsers.add_parser("strikethrough", help="Apply strikethrough to text range")
    strike_parser.add_argument("--doc-id", required=True, help="Document ID")
    strike_parser.add_argument("--start", type=int, required=True, help="Start index")
    strike_parser.add_argument("--end", type=int, required=True, help="End index")

    # Add link command
    link_parser = subparsers.add_parser("add-link", help="Add hyperlink to text range")
    link_parser.add_argument("--doc-id", required=True, help="Document ID")
    link_parser.add_argument("--start", type=int, required=True, help="Start index")
    link_parser.add_argument("--end", type=int, required=True, help="End index")
    link_parser.add_argument("--url", required=True, help="URL for the hyperlink")

    # Add links by text command
    links_by_text_parser = subparsers.add_parser("add-links-by-text", help="Add links to all occurrences of text")
    links_by_text_parser.add_argument("--doc-id", required=True, help="Document ID")
    links_by_text_parser.add_argument("--links", required=True, help="JSON object mapping text to URL")

    # Find text command
    find_parser = subparsers.add_parser("find-text", help="Find positions of text in document")
    find_parser.add_argument("--doc-id", required=True, help="Document ID")
    find_parser.add_argument("--text", required=True, help="Text to search for")

    args = parser.parse_args()

    try:
        if args.command == "create":
            doc_id = create_document(args.title)
            print(json.dumps({
                "documentId": doc_id,
                "url": f"https://docs.google.com/document/d/{doc_id}/edit"
            }, indent=2))

        elif args.command == "read":
            if args.full:
                doc = read_document(args.doc_id)
                print(json.dumps(doc, indent=2))
            else:
                structure = get_document_structure(args.doc_id)
                print(json.dumps(structure, indent=2))

        elif args.command == "add-section":
            result = add_section(args.doc_id, args.heading, args.text, args.level)
            print(json.dumps(result, indent=2))

        elif args.command == "add-table":
            data = json.loads(args.data)
            links = json.loads(args.links)
            result = add_table(args.doc_id, args.rows, args.cols, data, links)
            print(json.dumps(result, indent=2))

        elif args.command == "end-index":
            end_index = get_document_end_index(args.doc_id)
            print(json.dumps({"endIndex": end_index}))

        elif args.command == "add-person":
            result = add_person(args.doc_id, args.email, args.index)
            print(json.dumps(result, indent=2))

        elif args.command == "add-checklist":
            items = json.loads(args.items)
            checked = json.loads(args.checked)
            result = add_checklist(args.doc_id, items, checked)
            print(json.dumps(result, indent=2))

        elif args.command == "add-bullets":
            items = json.loads(args.items)
            result = add_bullet_list(args.doc_id, items, args.preset)
            print(json.dumps(result, indent=2))

        elif args.command == "strikethrough":
            result = apply_strikethrough(args.doc_id, args.start, args.end)
            print(json.dumps(result, indent=2))

        elif args.command == "add-link":
            result = add_link(args.doc_id, args.start, args.end, args.url)
            print(json.dumps(result, indent=2))

        elif args.command == "add-links-by-text":
            links = json.loads(args.links)
            count = add_links_by_text(args.doc_id, links)
            print(json.dumps({"links_added": count}))

        elif args.command == "find-text":
            positions = find_text_positions(args.doc_id, args.text)
            print(json.dumps({"positions": positions}))

    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
