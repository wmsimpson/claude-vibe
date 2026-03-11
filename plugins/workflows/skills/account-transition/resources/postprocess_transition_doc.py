#!/usr/bin/env python3
"""
Post-process account transition documents for proper formatting.

This script:
1. Applies color formatting to status indicators (Green/Yellow/Red)
2. Converts @mentions to Google Docs person chips

Usage:
    python3 postprocess_transition_doc.py --doc-id DOC_ID --mentions '{"@Name": "email@databricks.com"}'
"""

import argparse
import json
import os
import subprocess
import urllib.request
import urllib.error
import time
import sys

QUOTA_PROJECT = os.environ.get("GCP_QUOTA_PROJECT", "")
if not QUOTA_PROJECT:
    print("Error: GCP_QUOTA_PROJECT environment variable is not set.", file=sys.stderr)
    print("Run: export GCP_QUOTA_PROJECT=<your-gcp-project>", file=sys.stderr)
    sys.exit(1)

# Status colors for visual indicators
STATUS_COLORS = {
    "**Green**": {"red": 0.0, "green": 0.5, "blue": 0.0},
    "**Yellow**": {"red": 0.8, "green": 0.6, "blue": 0.0},
    "**Red**": {"red": 0.8, "green": 0.0, "blue": 0.0},
    "Green": {"red": 0.0, "green": 0.5, "blue": 0.0},
    "Yellow": {"red": 0.8, "green": 0.6, "blue": 0.0},
    "Red": {"red": 0.8, "green": 0.0, "blue": 0.0},
}


def get_token():
    """Get Google OAuth token via gcloud."""
    result = subprocess.run(
        ["gcloud", "auth", "application-default", "print-access-token"],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        print(f"Error getting token: {result.stderr}", file=sys.stderr)
        sys.exit(1)
    return result.stdout.strip()


def get_document(doc_id):
    """Fetch document structure from Google Docs API."""
    token = get_token()
    req = urllib.request.Request(
        f"https://docs.googleapis.com/v1/documents/{doc_id}",
        headers={
            "Authorization": f"Bearer {token}",
            "x-goog-user-project": QUOTA_PROJECT
        }
    )
    try:
        with urllib.request.urlopen(req) as response:
            return json.loads(response.read())
    except urllib.error.HTTPError as e:
        print(f"Error fetching document: {e.read().decode()}", file=sys.stderr)
        sys.exit(1)


def batch_update(doc_id, requests):
    """Send batch update to Google Docs API."""
    if not requests:
        return None

    token = get_token()
    data = json.dumps({"requests": requests}).encode('utf-8')
    req = urllib.request.Request(
        f"https://docs.googleapis.com/v1/documents/{doc_id}:batchUpdate",
        data=data,
        headers={
            "Authorization": f"Bearer {token}",
            "x-goog-user-project": QUOTA_PROJECT,
            "Content-Type": "application/json"
        }
    )
    try:
        with urllib.request.urlopen(req) as response:
            return json.loads(response.read())
    except urllib.error.HTTPError as e:
        error_body = e.read().decode()
        print(f"Batch update error: {error_body}", file=sys.stderr)
        return None


def find_text_in_document(doc, search_text):
    """Find all occurrences of text in document with their positions."""
    occurrences = []

    def search_in_elements(elements, search_text):
        """Search for text in a list of paragraph elements."""
        results = []
        for text_elem in elements:
            if 'textRun' in text_elem:
                content = text_elem['textRun'].get('content', '')
                start_idx = text_elem['startIndex']

                pos = 0
                while True:
                    pos = content.find(search_text, pos)
                    if pos == -1:
                        break
                    results.append({
                        'start': start_idx + pos,
                        'end': start_idx + pos + len(search_text),
                        'text': search_text
                    })
                    pos += 1
        return results

    for element in doc.get('body', {}).get('content', []):
        # Search in paragraphs
        if 'paragraph' in element:
            occurrences.extend(
                search_in_elements(element['paragraph'].get('elements', []), search_text)
            )

        # Search in table cells
        if 'table' in element:
            for row in element['table'].get('tableRows', []):
                for cell in row.get('tableCells', []):
                    for cell_content in cell.get('content', []):
                        if 'paragraph' in cell_content:
                            occurrences.extend(
                                search_in_elements(
                                    cell_content['paragraph'].get('elements', []),
                                    search_text
                                )
                            )

    return occurrences


def apply_status_colors(doc_id, doc):
    """Apply color formatting to status indicators."""
    print("Applying status colors...")

    for status_text, color in STATUS_COLORS.items():
        occurrences = find_text_in_document(doc, status_text)
        if occurrences:
            requests = []
            for occ in occurrences:
                requests.append({
                    "updateTextStyle": {
                        "range": {"startIndex": occ['start'], "endIndex": occ['end']},
                        "textStyle": {
                            "foregroundColor": {"color": {"rgbColor": color}},
                            "bold": True
                        },
                        "fields": "foregroundColor,bold"
                    }
                })

            # Apply in batches of 50
            for i in range(0, len(requests), 50):
                batch_update(doc_id, requests[i:i+50])
            print(f"  Applied color to {len(occurrences)} instances of '{status_text}'")


def replace_mentions_with_person_chips(doc_id, doc, mention_to_email):
    """Replace @mentions with Google Docs person chips."""
    if not mention_to_email:
        print("No mentions to process.")
        return

    print("\nReplacing @mentions with person chips...")

    # Collect all mentions with their positions
    all_mentions = []
    for mention_text, email in mention_to_email.items():
        occurrences = find_text_in_document(doc, mention_text)
        for occ in occurrences:
            occ['email'] = email
            all_mentions.append(occ)

    if not all_mentions:
        print("  No @mentions found in document.")
        return

    # Sort by start index descending (MUST process in reverse order)
    all_mentions.sort(key=lambda x: x['start'], reverse=True)

    for mention in all_mentions:
        requests = [
            # Delete the @mention text
            {
                "deleteContentRange": {
                    "range": {
                        "startIndex": mention['start'],
                        "endIndex": mention['end']
                    }
                }
            },
            # Insert person chip at the same location
            {
                "insertPerson": {
                    "personProperties": {
                        "email": mention['email']
                    },
                    "location": {"index": mention['start']}
                }
            }
        ]

        result = batch_update(doc_id, requests)
        if result:
            print(f"  Replaced '{mention['text']}' with person chip for {mention['email']}")

        # Small delay between updates
        time.sleep(0.2)

        # Re-fetch document after each change since indices shift
        doc = get_document(doc_id)


def main():
    parser = argparse.ArgumentParser(
        description='Post-process account transition documents'
    )
    parser.add_argument('--doc-id', required=True, help='Google Doc ID')
    parser.add_argument(
        '--mentions',
        help='JSON mapping of @mentions to emails, e.g., \'{"@John Doe": "john.doe@databricks.com"}\''
    )
    parser.add_argument(
        '--skip-colors',
        action='store_true',
        help='Skip applying status colors'
    )
    parser.add_argument(
        '--skip-mentions',
        action='store_true',
        help='Skip replacing @mentions'
    )

    args = parser.parse_args()

    # Parse mentions JSON
    mention_to_email = {}
    if args.mentions:
        try:
            mention_to_email = json.loads(args.mentions)
        except json.JSONDecodeError as e:
            print(f"Error parsing mentions JSON: {e}", file=sys.stderr)
            sys.exit(1)

    print(f"Processing document: {args.doc_id}")

    # Fetch document
    doc = get_document(args.doc_id)

    # Apply status colors
    if not args.skip_colors:
        apply_status_colors(args.doc_id, doc)
        time.sleep(0.5)
        doc = get_document(args.doc_id)

    # Replace mentions with person chips
    if not args.skip_mentions and mention_to_email:
        replace_mentions_with_person_chips(args.doc_id, doc, mention_to_email)

    print(f"\nDocument updated: https://docs.google.com/document/d/{args.doc_id}/edit")


if __name__ == "__main__":
    main()
