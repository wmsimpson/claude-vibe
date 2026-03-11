#!/usr/bin/env python3
"""
Gmail Builder - Complete Gmail Operations Helper

Provides high-level operations for Gmail:
- Search and read emails
- Send emails with HTML formatting
- Create and manage drafts
- Forward and reply to emails
- Manage labels and organize emails
- Handle attachments

Usage:
    python3 gmail_builder.py send --to "user@example.com" --subject "Test" --body "Hello"
    python3 gmail_builder.py search --query "is:unread"
    python3 gmail_builder.py read-message --message-id "MSG_ID"
"""

import argparse
import base64
import json
import mimetypes
import os
import subprocess
import sys
from email.mime.base import MIMEBase
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from email.mime.image import MIMEImage
from email import encoders
from typing import Dict, List, Optional

from google_api_utils import get_access_token, api_call_with_retry, QUOTA_PROJECT

GMAIL_API_BASE = "https://gmail.googleapis.com/gmail/v1/users/me"

# Cache for authenticated user's email
_my_email_cache: Optional[str] = None


def api_request(method: str, endpoint: str, data: Optional[Dict] = None) -> Dict:
    """Make an authenticated API request to Gmail with retry logic."""
    url = f"{GMAIL_API_BASE}/{endpoint}"
    try:
        return api_call_with_retry(method, url, data=data)
    except RuntimeError:
        return {}


def get_my_email() -> str:
    """Get the authenticated user's email address from the Gmail API profile."""
    global _my_email_cache
    if _my_email_cache:
        return _my_email_cache
    profile = api_request("GET", "profile")
    _my_email_cache = profile.get('emailAddress', '')
    return _my_email_cache


def extract_email(addr: str) -> str:
    """Extract bare email address from a header value like '"Name" <email@example.com>'."""
    import re
    match = re.search(r'<([^>]+)>', addr)
    if match:
        return match.group(1).lower()
    return addr.strip().lower()


def is_my_email(addr: str) -> bool:
    """Check if an address string matches the authenticated user's email."""
    my_email = get_my_email()
    if not my_email:
        return False
    return extract_email(addr) == my_email.lower()


def base64url_encode(data: bytes) -> str:
    """Encode bytes to base64url format (URL-safe, no padding)."""
    return base64.urlsafe_b64encode(data).decode('utf-8').rstrip('=')


def base64url_decode(data: str) -> bytes:
    """Decode base64url format to bytes."""
    # Add padding if needed
    padding = 4 - len(data) % 4
    if padding != 4:
        data += '=' * padding
    return base64.urlsafe_b64decode(data)


def create_message(to: str, subject: str, body: Optional[str] = None,
                   html: Optional[str] = None, cc: Optional[str] = None,
                   bcc: Optional[str] = None, attachments: Optional[List[str]] = None,
                   inline_images: Optional[List[str]] = None,
                   reply_to_id: Optional[str] = None,
                   thread_id: Optional[str] = None) -> Dict:
    """
    Create an email message.

    Args:
        to: Recipient email address(es), comma-separated
        subject: Email subject
        body: Plain text body
        html: HTML body
        cc: CC recipients
        bcc: BCC recipients
        attachments: List of file paths to attach
        inline_images: List of "path:cid" for inline images
        reply_to_id: Message ID to reply to (for threading)
        thread_id: Thread ID for reply

    Returns:
        Dict with 'raw' field containing base64url encoded message
    """
    # Determine message type based on content
    if attachments or inline_images:
        if html and inline_images:
            msg = MIMEMultipart('related')
            alt = MIMEMultipart('alternative')
            if body:
                alt.attach(MIMEText(body, 'plain', 'utf-8'))
            alt.attach(MIMEText(html, 'html', 'utf-8'))
            msg.attach(alt)
        else:
            msg = MIMEMultipart('mixed')
            if html:
                alt = MIMEMultipart('alternative')
                if body:
                    alt.attach(MIMEText(body, 'plain', 'utf-8'))
                alt.attach(MIMEText(html, 'html', 'utf-8'))
                msg.attach(alt)
            elif body:
                msg.attach(MIMEText(body, 'plain', 'utf-8'))
    elif html:
        msg = MIMEMultipart('alternative')
        if body:
            msg.attach(MIMEText(body, 'plain', 'utf-8'))
        else:
            # Generate plain text from HTML (basic)
            import re
            plain = re.sub('<[^<]+?>', '', html)
            msg.attach(MIMEText(plain, 'plain', 'utf-8'))
        msg.attach(MIMEText(html, 'html', 'utf-8'))
    else:
        msg = MIMEText(body or '', 'plain', 'utf-8')

    # Set headers
    msg['To'] = to
    msg['Subject'] = subject
    if cc:
        msg['Cc'] = cc
    if bcc:
        msg['Bcc'] = bcc

    # Add reply headers if replying
    if reply_to_id:
        # Get original message headers for proper threading
        original = api_request("GET", f"messages/{reply_to_id}?format=metadata&metadataHeaders=Message-ID&metadataHeaders=References&metadataHeaders=In-Reply-To")
        headers = {h['name']: h['value'] for h in original.get('payload', {}).get('headers', [])}

        original_msg_id = headers.get('Message-ID', '')
        references = headers.get('References', '')

        if original_msg_id:
            msg['In-Reply-To'] = original_msg_id
            if references:
                msg['References'] = f"{references} {original_msg_id}"
            else:
                msg['References'] = original_msg_id

    # Add attachments
    if attachments:
        for filepath in attachments:
            if not os.path.exists(filepath):
                print(f"Warning: Attachment not found: {filepath}", file=sys.stderr)
                continue

            filename = os.path.basename(filepath)
            mime_type, _ = mimetypes.guess_type(filepath)
            if mime_type is None:
                mime_type = 'application/octet-stream'

            main_type, sub_type = mime_type.split('/', 1)

            with open(filepath, 'rb') as f:
                attachment = MIMEBase(main_type, sub_type)
                attachment.set_payload(f.read())
                encoders.encode_base64(attachment)
                attachment.add_header('Content-Disposition', 'attachment', filename=filename)
                msg.attach(attachment)

    # Add inline images
    if inline_images:
        for img_spec in inline_images:
            parts = img_spec.split(':')
            if len(parts) != 2:
                print(f"Warning: Invalid inline image spec: {img_spec}", file=sys.stderr)
                continue

            filepath, cid = parts
            if not os.path.exists(filepath):
                print(f"Warning: Inline image not found: {filepath}", file=sys.stderr)
                continue

            with open(filepath, 'rb') as f:
                img = MIMEImage(f.read())
                img.add_header('Content-ID', f'<{cid}>')
                img.add_header('Content-Disposition', 'inline', filename=os.path.basename(filepath))
                msg.attach(img)

    raw = base64url_encode(msg.as_bytes())
    result = {'raw': raw}
    if thread_id:
        result['threadId'] = thread_id

    return result


def search_messages(query: str, max_results: int = 10) -> List[Dict]:
    """Search for messages matching query."""
    # URL encode the query
    import urllib.parse
    encoded_query = urllib.parse.quote(query)

    result = api_request("GET", f"messages?q={encoded_query}&maxResults={max_results}")
    messages = result.get('messages', [])

    detailed_messages = []
    for msg in messages:
        details = api_request("GET", f"messages/{msg['id']}?format=metadata&metadataHeaders=From&metadataHeaders=To&metadataHeaders=Subject&metadataHeaders=Date")
        headers = {h['name']: h['value'] for h in details.get('payload', {}).get('headers', [])}
        detailed_messages.append({
            'id': msg['id'],
            'threadId': msg['threadId'],
            'snippet': details.get('snippet', ''),
            'from': headers.get('From', ''),
            'to': headers.get('To', ''),
            'subject': headers.get('Subject', ''),
            'date': headers.get('Date', ''),
            'labelIds': details.get('labelIds', [])
        })

    return detailed_messages


def read_message(message_id: str, format: str = 'full') -> Dict:
    """Read a full message."""
    result = api_request("GET", f"messages/{message_id}?format={format}")

    if format == 'full':
        # Extract and decode body
        payload = result.get('payload', {})
        headers = {h['name']: h['value'] for h in payload.get('headers', [])}

        body_text = ''
        body_html = ''

        def extract_body(part):
            nonlocal body_text, body_html
            mime_type = part.get('mimeType', '')
            if 'body' in part and 'data' in part['body']:
                decoded = base64url_decode(part['body']['data']).decode('utf-8', errors='replace')
                if mime_type == 'text/plain':
                    body_text = decoded
                elif mime_type == 'text/html':
                    body_html = decoded
            if 'parts' in part:
                for p in part['parts']:
                    extract_body(p)

        extract_body(payload)

        return {
            'id': result.get('id'),
            'threadId': result.get('threadId'),
            'labelIds': result.get('labelIds', []),
            'snippet': result.get('snippet', ''),
            'headers': headers,
            'body_text': body_text,
            'body_html': body_html
        }

    return result


def send_message(to: str, subject: str, body: Optional[str] = None,
                 html: Optional[str] = None, cc: Optional[str] = None,
                 bcc: Optional[str] = None, attachments: Optional[List[str]] = None,
                 inline_images: Optional[List[str]] = None) -> Dict:
    """Send an email."""
    msg = create_message(to, subject, body, html, cc, bcc, attachments, inline_images)
    return api_request("POST", "messages/send", msg)


def create_draft(to: str, subject: str, body: Optional[str] = None,
                 html: Optional[str] = None, cc: Optional[str] = None,
                 bcc: Optional[str] = None, attachments: Optional[List[str]] = None,
                 reply_to_id: Optional[str] = None, thread_id: Optional[str] = None) -> Dict:
    """Create a draft email.

    Args:
        to: Recipient email address(es)
        subject: Email subject
        body: Plain text body
        html: HTML body
        cc: CC recipients
        bcc: BCC recipients
        attachments: List of file paths to attach
        reply_to_id: Message ID to reply to (for threading)
        thread_id: Thread ID for reply drafts
    """
    msg = create_message(to, subject, body, html, cc, bcc, attachments,
                        reply_to_id=reply_to_id, thread_id=thread_id)
    return api_request("POST", "drafts", {"message": msg})


def list_drafts() -> List[Dict]:
    """List all drafts."""
    result = api_request("GET", "drafts")
    drafts = result.get('drafts', [])

    detailed_drafts = []
    for draft in drafts:
        details = api_request("GET", f"drafts/{draft['id']}")
        msg = details.get('message', {})
        payload = msg.get('payload', {})
        headers = {h['name']: h['value'] for h in payload.get('headers', [])}

        detailed_drafts.append({
            'id': draft['id'],
            'messageId': msg.get('id'),
            'snippet': msg.get('snippet', ''),
            'to': headers.get('To', ''),
            'subject': headers.get('Subject', '')
        })

    return detailed_drafts


def send_draft(draft_id: str) -> Dict:
    """Send a draft."""
    return api_request("POST", "drafts/send", {"id": draft_id})


def delete_draft(draft_id: str) -> Dict:
    """Delete a draft."""
    token = get_access_token()
    cmd = ["curl", "-s", "-X", "DELETE",
           f"{GMAIL_API_BASE}/drafts/{draft_id}",
           "-H", f"Authorization: Bearer {token}",
           "-H", f"x-goog-user-project: {QUOTA_PROJECT}"]
    subprocess.run(cmd, capture_output=True)
    return {"status": "deleted"}


def forward_message(message_id: str, to: str, comment: Optional[str] = None) -> Dict:
    """Forward a message."""
    # Get original message
    original = read_message(message_id)

    # Build forward subject
    orig_subject = original['headers'].get('Subject', '')
    if not orig_subject.lower().startswith('fwd:'):
        subject = f"Fwd: {orig_subject}"
    else:
        subject = orig_subject

    # Build forward body
    orig_from = original['headers'].get('From', '')
    orig_date = original['headers'].get('Date', '')
    orig_to = original['headers'].get('To', '')

    forward_header = f"""
---------- Forwarded message ---------
From: {orig_from}
Date: {orig_date}
Subject: {orig_subject}
To: {orig_to}

"""

    if original['body_html']:
        forward_html = f"""
<div>{comment or ''}</div>
<br><br>
<div style="border-left: 1px solid #ccc; padding-left: 10px;">
<b>---------- Forwarded message ---------</b><br>
<b>From:</b> {orig_from}<br>
<b>Date:</b> {orig_date}<br>
<b>Subject:</b> {orig_subject}<br>
<b>To:</b> {orig_to}<br>
<br>
{original['body_html']}
</div>
"""
        body = (comment or '') + forward_header + original['body_text']
        return send_message(to, subject, body=body, html=forward_html)
    else:
        body = (comment + "\n\n" if comment else "") + forward_header + original['body_text']
        return send_message(to, subject, body=body)


def reply_message(message_id: str, body: Optional[str] = None,
                  html: Optional[str] = None, reply_all: bool = False) -> Dict:
    """Reply to a message."""
    # Get original message
    original = read_message(message_id)

    # Build reply subject
    orig_subject = original['headers'].get('Subject', '')
    if not orig_subject.lower().startswith('re:'):
        subject = f"Re: {orig_subject}"
    else:
        subject = orig_subject

    # Determine recipients
    orig_from = original['headers'].get('From', '')
    orig_to = original['headers'].get('To', '')
    orig_cc = original['headers'].get('Cc', '')

    if reply_all:
        # To: original sender + original To recipients (minus self and sender)
        to_recipients = [orig_from]
        for addr in orig_to.split(','):
            addr = addr.strip()
            if addr and not is_my_email(addr) and extract_email(addr) != extract_email(orig_from):
                to_recipients.append(addr)
        to = ', '.join(to_recipients)

        # CC: original CC recipients (minus self)
        cc_recipients = []
        if orig_cc:
            for addr in orig_cc.split(','):
                addr = addr.strip()
                if addr and not is_my_email(addr):
                    cc_recipients.append(addr)
        cc = ', '.join(cc_recipients) if cc_recipients else None
    else:
        to = orig_from
        cc = None

    # Build quoted original
    orig_date = original['headers'].get('Date', '')
    quote_header = f"\n\nOn {orig_date}, {orig_from} wrote:\n"

    if html:
        quoted_html = f"""
{html}
<br><br>
<div style="border-left: 1px solid #ccc; padding-left: 10px; color: #666;">
On {orig_date}, {orig_from} wrote:<br>
<br>
{original['body_html'] or original['body_text']}
</div>
"""
        full_body = (body or '') + quote_header + "\n> " + "\n> ".join(original['body_text'].split('\n'))
        msg = create_message(to, subject, body=full_body, html=quoted_html, cc=cc,
                            reply_to_id=message_id, thread_id=original['threadId'])
    else:
        quoted_text = "\n> ".join(original['body_text'].split('\n'))
        full_body = (body or '') + quote_header + "\n> " + quoted_text
        msg = create_message(to, subject, body=full_body, cc=cc,
                            reply_to_id=message_id, thread_id=original['threadId'])

    return api_request("POST", "messages/send", msg)


def create_reply_draft(message_id: str, body: Optional[str] = None,
                       html: Optional[str] = None, reply_all: bool = True) -> Dict:
    """Create a draft reply to a message.

    Args:
        message_id: ID of the message to reply to
        body: Plain text reply content
        html: HTML reply content
        reply_all: If True (default), include all original recipients (To + CC).
                  Set to False to reply only to the original sender.

    Returns:
        Dict with draft information including draft ID
    """
    # Get original message
    original = read_message(message_id)

    # Build reply subject
    orig_subject = original['headers'].get('Subject', '')
    if not orig_subject.lower().startswith('re:'):
        subject = f"Re: {orig_subject}"
    else:
        subject = orig_subject

    # Determine recipients
    orig_from = original['headers'].get('From', '')
    orig_to = original['headers'].get('To', '')
    orig_cc = original['headers'].get('Cc', '')

    if reply_all:
        # Reply-all: preserve original To/CC structure, but replace self with sender
        # To: original To recipients (minus self) + original sender
        to_recipients = []
        # Always include the original sender
        to_recipients.append(orig_from)
        # Include original To recipients (except self and the original sender)
        for addr in orig_to.split(','):
            addr = addr.strip()
            if addr and not is_my_email(addr) and extract_email(addr) != extract_email(orig_from):
                to_recipients.append(addr)
        to = ', '.join(to_recipients)

        # CC: original CC recipients (minus self)
        cc_recipients = []
        if orig_cc:
            for addr in orig_cc.split(','):
                addr = addr.strip()
                if addr and not is_my_email(addr):
                    cc_recipients.append(addr)
        cc = ', '.join(cc_recipients) if cc_recipients else None
    else:
        # Reply to sender only
        to = orig_from
        cc = None

    # Build quoted original
    orig_date = original['headers'].get('Date', '')
    quote_header = f"\n\nOn {orig_date}, {orig_from} wrote:\n"

    if html:
        quoted_html = f"""
{html}
<br><br>
<div style="border-left: 1px solid #ccc; padding-left: 10px; color: #666;">
On {orig_date}, {orig_from} wrote:<br>
<br>
{original['body_html'] or original['body_text']}
</div>
"""
        full_body = (body or '') + quote_header + "\n> " + "\n> ".join(original['body_text'].split('\n'))
        return create_draft(to, subject, body=full_body, html=quoted_html, cc=cc,
                           reply_to_id=message_id, thread_id=original['threadId'])
    else:
        quoted_text = "\n> ".join(original['body_text'].split('\n'))
        full_body = (body or '') + quote_header + "\n> " + quoted_text
        return create_draft(to, subject, body=full_body, cc=cc,
                           reply_to_id=message_id, thread_id=original['threadId'])


def modify_labels(message_id: str, add_labels: Optional[List[str]] = None,
                  remove_labels: Optional[List[str]] = None) -> Dict:
    """Add or remove labels from a message."""
    data = {}
    if add_labels:
        data['addLabelIds'] = add_labels
    if remove_labels:
        data['removeLabelIds'] = remove_labels

    return api_request("POST", f"messages/{message_id}/modify", data)


def trash_message(message_id: str) -> Dict:
    """Move message to trash."""
    return api_request("POST", f"messages/{message_id}/trash", {})


def untrash_message(message_id: str) -> Dict:
    """Remove message from trash."""
    return api_request("POST", f"messages/{message_id}/untrash", {})


def spam_message(message_id: str) -> Dict:
    """Mark message as spam."""
    return modify_labels(message_id, add_labels=['SPAM'], remove_labels=['INBOX'])


def unspam_message(message_id: str) -> Dict:
    """Mark message as not spam."""
    return modify_labels(message_id, add_labels=['INBOX'], remove_labels=['SPAM'])


def delete_message(message_id: str) -> Dict:
    """Permanently delete a message."""
    token = get_access_token()
    cmd = ["curl", "-s", "-X", "DELETE",
           f"{GMAIL_API_BASE}/messages/{message_id}",
           "-H", f"Authorization: Bearer {token}",
           "-H", f"x-goog-user-project: {QUOTA_PROJECT}"]
    subprocess.run(cmd, capture_output=True)
    return {"status": "deleted"}


def empty_folder(folder: str, confirm: bool = False) -> Dict:
    """Empty trash or spam folder."""
    if not confirm:
        return {"error": "Use --confirm to empty folder"}

    if folder not in ['trash', 'spam']:
        return {"error": "Can only empty trash or spam"}

    query = f"in:{folder}"
    messages = search_messages(query, max_results=500)

    deleted_count = 0
    for msg in messages:
        delete_message(msg['id'])
        deleted_count += 1

    return {"deleted": deleted_count, "folder": folder}


def list_labels() -> List[Dict]:
    """List all labels."""
    result = api_request("GET", "labels")
    return result.get('labels', [])


def create_label(name: str) -> Dict:
    """Create a new label."""
    return api_request("POST", "labels", {
        "name": name,
        "labelListVisibility": "labelShow",
        "messageListVisibility": "show"
    })


def delete_label(label_id: str) -> Dict:
    """Delete a label."""
    token = get_access_token()
    cmd = ["curl", "-s", "-X", "DELETE",
           f"{GMAIL_API_BASE}/labels/{label_id}",
           "-H", f"Authorization: Bearer {token}",
           "-H", f"x-goog-user-project: {QUOTA_PROJECT}"]
    subprocess.run(cmd, capture_output=True)
    return {"status": "deleted"}


# --- Filter Management ---

def list_filters() -> List[Dict]:
    """List all Gmail filters."""
    result = api_request("GET", "settings/filters")
    return result.get('filter', [])


def get_filter(filter_id: str) -> Dict:
    """Get a specific filter by ID."""
    return api_request("GET", f"settings/filters/{filter_id}")


def create_filter(criteria: Dict, action: Dict) -> Dict:
    """
    Create a new Gmail filter.

    Criteria options:
        from, to, subject, query, negatedQuery, hasAttachment,
        excludeChats, size, sizeComparison (larger/smaller)

    Action options:
        addLabelIds, removeLabelIds, forward,
        archive (removeLabelIds: ["INBOX"]), markRead (removeLabelIds: ["UNREAD"]),
        star (addLabelIds: ["STARRED"]), markImportant (addLabelIds: ["IMPORTANT"])
    """
    return api_request("POST", "settings/filters", {
        "criteria": criteria,
        "action": action
    })


def delete_filter(filter_id: str) -> Dict:
    """Delete a filter."""
    token = get_access_token()
    cmd = ["curl", "-s", "-X", "DELETE",
           f"{GMAIL_API_BASE}/settings/filters/{filter_id}",
           "-H", f"Authorization: Bearer {token}",
           "-H", f"x-goog-user-project: {QUOTA_PROJECT}"]
    subprocess.run(cmd, capture_output=True)
    return {"status": "deleted", "id": filter_id}


def download_attachments(message_id: str, output_dir: str) -> List[str]:
    """Download all attachments from a message."""
    os.makedirs(output_dir, exist_ok=True)

    result = api_request("GET", f"messages/{message_id}?format=full")
    payload = result.get('payload', {})

    downloaded = []

    def process_parts(parts):
        for part in parts:
            filename = part.get('filename', '')
            if filename and 'body' in part:
                body = part['body']
                if 'attachmentId' in body:
                    # Fetch attachment data
                    att = api_request("GET", f"messages/{message_id}/attachments/{body['attachmentId']}")
                    if 'data' in att:
                        data = base64url_decode(att['data'])
                        filepath = os.path.join(output_dir, filename)
                        with open(filepath, 'wb') as f:
                            f.write(data)
                        downloaded.append(filepath)

            if 'parts' in part:
                process_parts(part['parts'])

    if 'parts' in payload:
        process_parts(payload['parts'])

    return downloaded


def main():
    parser = argparse.ArgumentParser(
        description="Gmail operations helper",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    subparsers = parser.add_subparsers(dest="command")

    # Search
    search_parser = subparsers.add_parser("search", help="Search for messages")
    search_parser.add_argument("--query", "-q", required=True, help="Search query")
    search_parser.add_argument("--max-results", "-n", type=int, default=10, help="Max results")

    # Read message
    read_parser = subparsers.add_parser("read-message", help="Read a message")
    read_parser.add_argument("--message-id", required=True, help="Message ID")

    # Send
    send_parser = subparsers.add_parser("send", help="Send an email")
    send_parser.add_argument("--to", required=True, help="Recipient(s)")
    send_parser.add_argument("--subject", required=True, help="Subject")
    send_parser.add_argument("--body", help="Plain text body")
    send_parser.add_argument("--html", help="HTML body")
    send_parser.add_argument("--cc", help="CC recipients")
    send_parser.add_argument("--bcc", help="BCC recipients")
    send_parser.add_argument("--attach", action="append", help="File to attach")
    send_parser.add_argument("--attach-inline", action="append", help="Inline image (path:cid)")

    # Create draft
    draft_parser = subparsers.add_parser("create-draft", help="Create a draft")
    draft_parser.add_argument("--to", required=True, help="Recipient(s)")
    draft_parser.add_argument("--subject", required=True, help="Subject")
    draft_parser.add_argument("--body", help="Plain text body")
    draft_parser.add_argument("--html", help="HTML body")
    draft_parser.add_argument("--cc", help="CC recipients")
    draft_parser.add_argument("--bcc", help="BCC recipients")
    draft_parser.add_argument("--attach", action="append", help="File to attach")
    draft_parser.add_argument("--reply-to-id", help="Message ID to reply to (for threading)")
    draft_parser.add_argument("--thread-id", help="Thread ID for reply drafts")

    # List drafts
    subparsers.add_parser("list-drafts", help="List all drafts")

    # Send draft
    send_draft_parser = subparsers.add_parser("send-draft", help="Send a draft")
    send_draft_parser.add_argument("--draft-id", required=True, help="Draft ID")

    # Delete draft
    del_draft_parser = subparsers.add_parser("delete-draft", help="Delete a draft")
    del_draft_parser.add_argument("--draft-id", required=True, help="Draft ID")

    # Create reply draft
    reply_draft_parser = subparsers.add_parser("create-reply-draft", help="Create a draft reply to a message")
    reply_draft_parser.add_argument("--message-id", required=True, help="Message ID to reply to")
    reply_draft_parser.add_argument("--body", help="Plain text reply")
    reply_draft_parser.add_argument("--html", help="HTML reply")
    reply_draft_parser.add_argument("--reply-all", action="store_true", default=True, help="Reply to all recipients (default: True)")
    reply_draft_parser.add_argument("--reply-sender-only", action="store_true", help="Reply only to sender (overrides --reply-all)")

    # Forward
    forward_parser = subparsers.add_parser("forward", help="Forward a message")
    forward_parser.add_argument("--message-id", required=True, help="Message ID")
    forward_parser.add_argument("--to", required=True, help="Forward to")
    forward_parser.add_argument("--comment", help="Optional comment")

    # Reply
    reply_parser = subparsers.add_parser("reply", help="Reply to a message")
    reply_parser.add_argument("--message-id", required=True, help="Message ID")
    reply_parser.add_argument("--body", help="Plain text reply")
    reply_parser.add_argument("--html", help="HTML reply")

    # Reply all
    reply_all_parser = subparsers.add_parser("reply-all", help="Reply to all")
    reply_all_parser.add_argument("--message-id", required=True, help="Message ID")
    reply_all_parser.add_argument("--body", help="Plain text reply")
    reply_all_parser.add_argument("--html", help="HTML reply")

    # Add label
    add_label_parser = subparsers.add_parser("add-label", help="Add label to message")
    add_label_parser.add_argument("--message-id", required=True, help="Message ID")
    add_label_parser.add_argument("--label", required=True, help="Label to add")

    # Remove label
    remove_label_parser = subparsers.add_parser("remove-label", help="Remove label from message")
    remove_label_parser.add_argument("--message-id", required=True, help="Message ID")
    remove_label_parser.add_argument("--label", required=True, help="Label to remove")

    # Trash
    trash_parser = subparsers.add_parser("trash", help="Move message to trash")
    trash_parser.add_argument("--message-id", required=True, help="Message ID")

    # Untrash
    untrash_parser = subparsers.add_parser("untrash", help="Remove message from trash")
    untrash_parser.add_argument("--message-id", required=True, help="Message ID")

    # Spam
    spam_parser = subparsers.add_parser("spam", help="Mark message as spam")
    spam_parser.add_argument("--message-id", required=True, help="Message ID")

    # Unspam
    unspam_parser = subparsers.add_parser("unspam", help="Mark message as not spam")
    unspam_parser.add_argument("--message-id", required=True, help="Message ID")

    # Delete
    delete_parser = subparsers.add_parser("delete", help="Permanently delete message")
    delete_parser.add_argument("--message-id", required=True, help="Message ID")

    # Empty trash
    empty_trash_parser = subparsers.add_parser("empty-trash", help="Empty trash")
    empty_trash_parser.add_argument("--confirm", action="store_true", help="Confirm deletion")

    # Empty spam
    empty_spam_parser = subparsers.add_parser("empty-spam", help="Empty spam")
    empty_spam_parser.add_argument("--confirm", action="store_true", help="Confirm deletion")

    # List labels
    subparsers.add_parser("list-labels", help="List all labels")

    # Create label
    create_label_parser = subparsers.add_parser("create-label", help="Create a label")
    create_label_parser.add_argument("--name", required=True, help="Label name")

    # Delete label
    delete_label_parser = subparsers.add_parser("delete-label", help="Delete a label")
    delete_label_parser.add_argument("--label-id", required=True, help="Label ID")

    # List filters
    subparsers.add_parser("list-filters", help="List all Gmail filters")

    # Get filter
    get_filter_parser = subparsers.add_parser("get-filter", help="Get a filter")
    get_filter_parser.add_argument("--filter-id", required=True, help="Filter ID")

    # Create filter
    create_filter_parser = subparsers.add_parser("create-filter", help="Create a filter")
    create_filter_parser.add_argument("--from", dest="from_addr", help="Match sender")
    create_filter_parser.add_argument("--to", help="Match recipient")
    create_filter_parser.add_argument("--subject", help="Match subject")
    create_filter_parser.add_argument("--query", help="Gmail search query")
    create_filter_parser.add_argument("--has-attachment", action="store_true", help="Has attachment")
    create_filter_parser.add_argument("--add-label", action="append", help="Label ID to add")
    create_filter_parser.add_argument("--remove-label", action="append", help="Label ID to remove")
    create_filter_parser.add_argument("--archive", action="store_true", help="Skip inbox")
    create_filter_parser.add_argument("--mark-read", action="store_true", help="Mark as read")
    create_filter_parser.add_argument("--star", action="store_true", help="Star message")
    create_filter_parser.add_argument("--mark-important", action="store_true", help="Mark important")
    create_filter_parser.add_argument("--forward", help="Forward to email address")

    # Delete filter
    delete_filter_parser = subparsers.add_parser("delete-filter", help="Delete a filter")
    delete_filter_parser.add_argument("--filter-id", required=True, help="Filter ID")

    # Download attachments
    download_parser = subparsers.add_parser("download-attachment", help="Download attachments")
    download_parser.add_argument("--message-id", required=True, help="Message ID")
    download_parser.add_argument("--output-dir", required=True, help="Output directory")

    # Auth status
    subparsers.add_parser("auth-status", help="Check authentication status")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    result = None

    if args.command == "search":
        result = search_messages(args.query, args.max_results)

    elif args.command == "read-message":
        result = read_message(args.message_id)

    elif args.command == "send":
        result = send_message(args.to, args.subject, args.body, args.html,
                             args.cc, args.bcc, args.attach, args.attach_inline)

    elif args.command == "create-draft":
        result = create_draft(args.to, args.subject, args.body, args.html,
                             args.cc, args.bcc, args.attach,
                             getattr(args, 'reply_to_id', None),
                             getattr(args, 'thread_id', None))

    elif args.command == "list-drafts":
        result = list_drafts()

    elif args.command == "send-draft":
        result = send_draft(args.draft_id)

    elif args.command == "delete-draft":
        result = delete_draft(args.draft_id)

    elif args.command == "create-reply-draft":
        reply_all = not args.reply_sender_only if hasattr(args, 'reply_sender_only') else args.reply_all
        result = create_reply_draft(args.message_id, args.body, args.html, reply_all)

    elif args.command == "forward":
        result = forward_message(args.message_id, args.to, args.comment)

    elif args.command == "reply":
        result = reply_message(args.message_id, args.body, args.html, reply_all=False)

    elif args.command == "reply-all":
        result = reply_message(args.message_id, args.body, args.html, reply_all=True)

    elif args.command == "add-label":
        result = modify_labels(args.message_id, add_labels=[args.label])

    elif args.command == "remove-label":
        result = modify_labels(args.message_id, remove_labels=[args.label])

    elif args.command == "trash":
        result = trash_message(args.message_id)

    elif args.command == "untrash":
        result = untrash_message(args.message_id)

    elif args.command == "spam":
        result = spam_message(args.message_id)

    elif args.command == "unspam":
        result = unspam_message(args.message_id)

    elif args.command == "delete":
        result = delete_message(args.message_id)

    elif args.command == "empty-trash":
        result = empty_folder("trash", args.confirm)

    elif args.command == "empty-spam":
        result = empty_folder("spam", args.confirm)

    elif args.command == "list-labels":
        result = list_labels()

    elif args.command == "create-label":
        result = create_label(args.name)

    elif args.command == "delete-label":
        result = delete_label(args.label_id)

    elif args.command == "list-filters":
        result = list_filters()

    elif args.command == "get-filter":
        result = get_filter(args.filter_id)

    elif args.command == "create-filter":
        # Build criteria
        criteria = {}
        if args.from_addr:
            criteria['from'] = args.from_addr
        if args.to:
            criteria['to'] = args.to
        if args.subject:
            criteria['subject'] = args.subject
        if args.query:
            criteria['query'] = args.query
        if args.has_attachment:
            criteria['hasAttachment'] = True

        # Build action
        action = {}
        add_labels = args.add_label or []
        remove_labels = args.remove_label or []

        if args.archive:
            remove_labels.append('INBOX')
        if args.mark_read:
            remove_labels.append('UNREAD')
        if args.star:
            add_labels.append('STARRED')
        if args.mark_important:
            add_labels.append('IMPORTANT')
        if args.forward:
            action['forward'] = args.forward

        if add_labels:
            action['addLabelIds'] = list(set(add_labels))
        if remove_labels:
            action['removeLabelIds'] = list(set(remove_labels))

        result = create_filter(criteria, action)

    elif args.command == "delete-filter":
        result = delete_filter(args.filter_id)

    elif args.command == "download-attachment":
        result = download_attachments(args.message_id, args.output_dir)

    elif args.command == "auth-status":
        import gmail_auth
        gmail_auth.print_status()
        sys.exit(0)

    if result is not None:
        print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
