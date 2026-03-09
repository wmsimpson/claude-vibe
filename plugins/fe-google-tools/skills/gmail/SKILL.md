---
name: gmail
description: Search, read, compose, organize Gmail emails, and manage filters (label rules) using gcloud CLI + curl. Use this skill for ANY email-related request - sending emails, reading emails, drafting emails, forwarding, replying, checking inbox, email attachments, email labels, or any operation involving e-mail.
---

# Gmail Skill

Manage Gmail emails using gcloud CLI + curl. This skill provides patterns and utilities for searching emails, reading messages, composing rich HTML emails with formatting, sending/forwarding messages, and organizing with labels.

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

## Core Concepts

### Message Format

Gmail API uses RFC 2822 format for messages. For rich HTML emails, you construct a MIME message with:
- Headers (From, To, Subject, Content-Type)
- Body (plain text and/or HTML)
- Attachments (base64-encoded)

### Message IDs and Thread IDs

- `id` - Unique identifier for a message
- `threadId` - Groups related messages (replies share the same threadId)
- Both are required for operations like replying or forwarding

### Labels

Gmail uses labels instead of folders:
- `INBOX` - Messages in inbox
- `SENT` - Sent messages
- `DRAFT` - Draft messages
- `TRASH` - Deleted messages
- `SPAM` - Spam messages
- `UNREAD` - Unread messages
- `STARRED` - Starred messages
- `IMPORTANT` - Important messages
- Custom labels can be created

## API Reference

### Base URL

```
https://gmail.googleapis.com/gmail/v1/users/me
```

All endpoints use `me` to refer to the authenticated user.

## Reading and Searching Emails

### List Messages

```bash
TOKEN=$(gcloud auth application-default print-access-token)

# List recent messages (default 100)
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages?maxResults=10" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Search Messages with Query

Gmail uses the same search syntax as the web interface:

```bash
# Search by sender
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages?q=from:user@example.com" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Search by subject
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages?q=subject:meeting" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Search unread messages
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages?q=is:unread" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Search with date range
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages?q=after:2024/01/01+before:2024/12/31" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Combined search
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages?q=from:boss@company.com+is:unread+has:attachment" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

**Common Search Operators:**
- `from:` - Sender address
- `to:` - Recipient address
- `subject:` - Subject line contains
- `is:unread` / `is:read` - Read status
- `is:starred` - Starred messages
- `is:important` - Important messages
- `has:attachment` - Has attachments
- `after:YYYY/MM/DD` - After date
- `before:YYYY/MM/DD` - Before date
- `newer_than:7d` - Within last 7 days
- `older_than:1m` - Older than 1 month
- `in:inbox` / `in:sent` / `in:trash` - Label/folder
- `label:` - Custom label

### Get Full Message

```bash
# Get message with full content
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}?format=full" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Get message metadata only (faster)
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}?format=metadata" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Get raw RFC 2822 message
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}?format=raw" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Extract Message Body (Python Helper)

The message body is base64url encoded. Use the helper script:

```bash
python3 resources/gmail_builder.py \
  read-message --message-id "MESSAGE_ID"
```

Or decode manually:

```bash
# Get message body (look in payload.body.data or payload.parts[].body.data)
MESSAGE_ID="your_message_id"
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}?format=full" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" | \
  jq -r '.payload.body.data // .payload.parts[0].body.data' | \
  base64 -d
```

### List Threads

```bash
# List threads
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/threads?maxResults=10" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Get full thread with all messages
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/threads/${THREAD_ID}?format=full" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

## Creating and Sending Emails

### Create Email Message (RFC 2822 Format)

Emails must be base64url encoded RFC 2822 messages:

```bash
# Simple text email
EMAIL_RAW=$(cat <<EOF | base64 | tr '+/' '-_' | tr -d '='
From: me
To: recipient@example.com
Subject: Test Email
Content-Type: text/plain; charset="UTF-8"

This is the email body.
EOF
)
```

### Send Email

```bash
# Send the email
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/send" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d "{\"raw\": \"$EMAIL_RAW\"}"
```

### Send HTML Email with Rich Formatting

```bash
# Create MIME message with HTML
EMAIL_RAW=$(cat <<'EOF' | base64 | tr '+/' '-_' | tr -d '='
MIME-Version: 1.0
From: me
To: colleague@example.com
Subject: Test HTML Email
Content-Type: multipart/alternative; boundary="boundary123"

--boundary123
Content-Type: text/plain; charset="UTF-8"

This is the plain text version.

--boundary123
Content-Type: text/html; charset="UTF-8"

<!DOCTYPE html>
<html>
<body>
  <h1>Hello!</h1>
  <p>This is a <strong>bold</strong> and <em>italic</em> message.</p>
  <ul>
    <li>Bullet point 1</li>
    <li>Bullet point 2</li>
    <li>Bullet point 3</li>
  </ul>
  <p>Visit <a href="https://databricks.com">Databricks</a> for more info.</p>
  <blockquote style="border-left: 3px solid #ccc; padding-left: 10px; color: #666;">
    This is a quoted text block.
  </blockquote>
</body>
</html>
--boundary123--
EOF
)

curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/send" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d "{\"raw\": \"$EMAIL_RAW\"}"
```

### Using the Helper Script for Rich Emails

```bash
# Send plain text email
python3 resources/gmail_builder.py \
  send --to "colleague@example.com" \
  --subject "Test Email" \
  --body "This is a test message."

# Send HTML email with formatting
python3 resources/gmail_builder.py \
  send --to "colleague@example.com" \
  --subject "Formatted Email" \
  --html '<h1>Hello!</h1><p>This is <strong>bold</strong> text.</p>'

# Send with CC and BCC
python3 resources/gmail_builder.py \
  send --to "primary@example.com" \
  --cc "copy@example.com" \
  --bcc "hidden@example.com" \
  --subject "Multi-recipient Email" \
  --body "Hello everyone!"
```

## HTML Formatting Reference

### Basic Formatting

```html
<!-- Bold -->
<strong>Bold text</strong>
<b>Also bold</b>

<!-- Italic -->
<em>Italic text</em>
<i>Also italic</i>

<!-- Underline -->
<u>Underlined text</u>

<!-- Strikethrough -->
<s>Strikethrough text</s>

<!-- Colored text -->
<span style="color: #FF0000;">Red text</span>
<span style="color: rgb(0, 128, 0);">Green text</span>

<!-- Background highlight -->
<span style="background-color: yellow;">Highlighted text</span>

<!-- Font size -->
<span style="font-size: 18px;">Larger text</span>
```

### Links

```html
<!-- Simple link -->
<a href="https://databricks.com">Click here</a>

<!-- Link with title -->
<a href="https://databricks.com" title="Visit Databricks">Databricks Website</a>

<!-- Email link -->
<a href="mailto:support@example.com">Email support</a>
```

### Lists

```html
<!-- Bulleted list -->
<ul>
  <li>First item</li>
  <li>Second item</li>
  <li>Third item</li>
</ul>

<!-- Numbered list -->
<ol>
  <li>Step one</li>
  <li>Step two</li>
  <li>Step three</li>
</ol>

<!-- Nested list -->
<ul>
  <li>Main item
    <ul>
      <li>Sub-item 1</li>
      <li>Sub-item 2</li>
    </ul>
  </li>
</ul>
```

### Quoted Text / Blockquotes

```html
<!-- Simple blockquote -->
<blockquote>
  This is quoted text from a previous email.
</blockquote>

<!-- Styled blockquote (Gmail style) -->
<blockquote style="margin: 0 0 0 0.8ex; border-left: 1px solid #ccc; padding-left: 1ex;">
  On Jan 1, 2024, John wrote:<br>
  <br>
  Original message content here.
</blockquote>
```

### Images

```html
<!-- Inline image from URL -->
<img src="https://example.com/image.png" alt="Description" width="400">

<!-- Centered image -->
<div style="text-align: center;">
  <img src="https://example.com/logo.png" alt="Logo" width="200">
</div>
```

**Note:** For inline images (embedded in email), you need to use CID references and multipart/related MIME. Use the helper script for this:

```bash
python3 resources/gmail_builder.py \
  send --to "colleague@example.com" \
  --subject "Email with Image" \
  --html '<h1>See image below</h1><img src="cid:image1">' \
  --attach-inline "/path/to/image.png:image1"
```

### Tables

```html
<table border="1" cellpadding="8" cellspacing="0" style="border-collapse: collapse;">
  <tr style="background-color: #f0f0f0;">
    <th>Name</th>
    <th>Email</th>
    <th>Status</th>
  </tr>
  <tr>
    <td>John Doe</td>
    <td>john@example.com</td>
    <td style="color: green;">Active</td>
  </tr>
  <tr>
    <td>Jane Smith</td>
    <td>jane@example.com</td>
    <td style="color: red;">Inactive</td>
  </tr>
</table>
```

## Drafts

### Create Draft

```bash
# Create draft email
EMAIL_RAW=$(cat <<EOF | base64 | tr '+/' '-_' | tr -d '='
From: me
To: recipient@example.com
Subject: Draft Email
Content-Type: text/plain; charset="UTF-8"

This is a draft that will be saved but not sent.
EOF
)

curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/drafts" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d "{\"message\": {\"raw\": \"$EMAIL_RAW\"}}"
```

### Create HTML Draft with Helper

```bash
python3 resources/gmail_builder.py \
  create-draft --to "colleague@example.com" \
  --subject "Draft with Formatting" \
  --html '<h1>Draft Email</h1><p>This has <strong>formatting</strong>.</p><ul><li>Item 1</li><li>Item 2</li></ul>'
```

### Create Reply Draft

**IMPORTANT:** To create a draft reply that appears in the correct thread, use the `create-reply-draft` command or pass `--reply-to-id` and `--thread-id` to `create-draft`.

```bash
# Create a draft reply to a message (easiest method)
python3 resources/gmail_builder.py \
  create-reply-draft --message-id "MESSAGE_ID" \
  --body "Thanks for your email! I'll review this and get back to you."

# Create HTML draft reply
python3 resources/gmail_builder.py \
  create-reply-draft --message-id "MESSAGE_ID" \
  --html '<p>Thanks for your email!</p><p>Here are my thoughts:</p><ul><li>Point 1</li><li>Point 2</li></ul>'

# By default, replies include all original recipients (To + CC)
# To reply only to the sender, use --no-reply-all (not yet implemented in CLI)

# Advanced: Create draft with explicit threading (for custom scenarios)
python3 resources/gmail_builder.py \
  create-draft --to "recipient@example.com" \
  --subject "Re: Original Subject" \
  --body "Reply content" \
  --reply-to-id "ORIGINAL_MESSAGE_ID" \
  --thread-id "THREAD_ID"
```

The `create-reply-draft` command automatically:
- Fetches the original message
- Sets proper threading headers (In-Reply-To, References)
- Builds the reply subject with "Re:" prefix
- Quotes the original message
- Associates the draft with the correct thread
- **Includes all original recipients by default** (To + CC) - acts as "Reply All"

**Without proper threading**, drafts may not appear in the correct conversation on Gmail desktop web interface.

### List Drafts

```bash
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/drafts" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Get Draft

```bash
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/drafts/${DRAFT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Update Draft

```bash
EMAIL_RAW=$(cat <<EOF | base64 | tr '+/' '-_' | tr -d '='
From: me
To: recipient@example.com
Subject: Updated Draft Subject
Content-Type: text/plain; charset="UTF-8"

Updated draft content.
EOF
)

curl -s -X PUT "https://gmail.googleapis.com/gmail/v1/users/me/drafts/${DRAFT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d "{\"message\": {\"raw\": \"$EMAIL_RAW\"}}"
```

### Delete Draft

```bash
curl -s -X DELETE "https://gmail.googleapis.com/gmail/v1/users/me/drafts/${DRAFT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Send Draft

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/drafts/send" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d "{\"id\": \"$DRAFT_ID\"}"
```

## Forwarding Emails

### Forward Message

To forward an email, you need to:
1. Get the original message
2. Create a new message with the original content quoted
3. Send the new message

```bash
python3 resources/gmail_builder.py \
  forward --message-id "ORIGINAL_MESSAGE_ID" \
  --to "forward-to@example.com" \
  --comment "FYI - see below"
```

### Reply to Message

```bash
python3 resources/gmail_builder.py \
  reply --message-id "ORIGINAL_MESSAGE_ID" \
  --body "Thanks for your message!"

# Reply with HTML
python3 resources/gmail_builder.py \
  reply --message-id "ORIGINAL_MESSAGE_ID" \
  --html "<p>Thanks for your message!</p><p>Here are the details:</p><ul><li>Point 1</li><li>Point 2</li></ul>"
```

### Reply All

```bash
python3 resources/gmail_builder.py \
  reply-all --message-id "ORIGINAL_MESSAGE_ID" \
  --body "Replying to everyone on this thread."
```

## Organizing Emails

### List Labels

```bash
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/labels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Create Label

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/labels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Custom Label",
    "labelListVisibility": "labelShow",
    "messageListVisibility": "show"
  }'
```

### Add Labels to Message

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/modify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "addLabelIds": ["STARRED", "IMPORTANT"]
  }'
```

### Remove Labels from Message

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/modify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "removeLabelIds": ["UNREAD"]
  }'
```

### Mark as Read/Unread

```bash
# Mark as read (remove UNREAD label)
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/modify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{"removeLabelIds": ["UNREAD"]}'

# Mark as unread (add UNREAD label)
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/modify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{"addLabelIds": ["UNREAD"]}'
```

### Star/Unstar Message

```bash
# Star message
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/modify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{"addLabelIds": ["STARRED"]}'

# Unstar message
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/modify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{"removeLabelIds": ["STARRED"]}'
```

### Archive Message (Remove from Inbox)

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/modify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{"removeLabelIds": ["INBOX"]}'
```

### Move to Trash

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/trash" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Untrash Message

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/untrash" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Permanently Delete Message

**WARNING: This is irreversible!**

```bash
curl -s -X DELETE "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Mark as Spam

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/modify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "addLabelIds": ["SPAM"],
    "removeLabelIds": ["INBOX"]
  }'
```

### Mark as Not Spam

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/modify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "removeLabelIds": ["SPAM"],
    "addLabelIds": ["INBOX"]
  }'
```

### Empty Trash

```bash
# List messages in trash
TRASH_MESSAGES=$(curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages?q=in:trash&maxResults=100" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" | jq -r '.messages[].id')

# Delete each message permanently
for MSG_ID in $TRASH_MESSAGES; do
  curl -s -X DELETE "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MSG_ID}" \
    -H "Authorization: Bearer $TOKEN" \
    -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
done
```

Or use the helper:

```bash
python3 resources/gmail_builder.py \
  empty-trash --confirm
```

### Empty Spam

```bash
python3 resources/gmail_builder.py \
  empty-spam --confirm
```

## Managing Filters (Label Rules)

Gmail filters automatically apply actions to incoming emails based on criteria.

### List All Filters

```bash
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/settings/filters" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

Or use the helper:

```bash
python3 resources/gmail_builder.py list-filters
```

### Create a Filter

```bash
# Create filter to label emails from boss as Important
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/settings/filters" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "criteria": {
      "from": "boss@company.com"
    },
    "action": {
      "addLabelIds": ["IMPORTANT", "STARRED"]
    }
  }'
```

Using the helper:

```bash
# Label emails from boss as Important and Star
python3 resources/gmail_builder.py create-filter \
  --from "boss@company.com" \
  --mark-important --star

# Archive newsletters automatically
python3 resources/gmail_builder.py create-filter \
  --from "newsletter@example.com" \
  --archive

# Label and mark read emails matching a query
python3 resources/gmail_builder.py create-filter \
  --query "subject:automated report" \
  --add-label "Label_123" \
  --mark-read

# Auto-forward emails from VIP
python3 resources/gmail_builder.py create-filter \
  --from "vip@customer.com" \
  --forward "team@company.com"
```

### Filter Criteria Reference

| Criteria | Description | Example |
|----------|-------------|---------|
| `from` | Sender address | `boss@company.com` |
| `to` | Recipient address | `team@company.com` |
| `subject` | Subject contains | `Meeting` |
| `query` | Any Gmail search query | `has:attachment larger:5M` |
| `hasAttachment` | Has attachments | `true` |
| `size` | Message size in bytes | `5000000` |
| `sizeComparison` | Size comparison | `larger` or `smaller` |

### Filter Actions Reference

| Action | Description | Helper Flag |
|--------|-------------|-------------|
| `addLabelIds` | Add labels | `--add-label LABEL_ID` |
| `removeLabelIds` | Remove labels | `--remove-label LABEL_ID` |
| Archive (skip inbox) | Remove INBOX label | `--archive` |
| Mark as read | Remove UNREAD label | `--mark-read` |
| Star | Add STARRED label | `--star` |
| Mark important | Add IMPORTANT label | `--mark-important` |
| `forward` | Forward to address | `--forward email@example.com` |

### Get a Filter

```bash
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/settings/filters/${FILTER_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Or with helper
python3 resources/gmail_builder.py get-filter --filter-id "FILTER_ID"
```

### Delete a Filter

```bash
curl -s -X DELETE "https://gmail.googleapis.com/gmail/v1/users/me/settings/filters/${FILTER_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Or with helper
python3 resources/gmail_builder.py delete-filter --filter-id "FILTER_ID"
```

**Note:** Gmail API does not support updating filters. To modify a filter, delete it and create a new one.

### Batch Operations

Modify multiple messages at once:

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/batchModify" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "ids": ["MESSAGE_ID_1", "MESSAGE_ID_2", "MESSAGE_ID_3"],
    "addLabelIds": ["STARRED"],
    "removeLabelIds": ["UNREAD"]
  }'
```

Delete multiple messages:

```bash
curl -s -X POST "https://gmail.googleapis.com/gmail/v1/users/me/messages/batchDelete" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "ids": ["MESSAGE_ID_1", "MESSAGE_ID_2", "MESSAGE_ID_3"]
  }'
```

## Attachments

### Send Email with Attachment

```bash
python3 resources/gmail_builder.py \
  send --to "colleague@example.com" \
  --subject "Email with Attachment" \
  --body "Please see the attached file." \
  --attach "/path/to/document.pdf"

# Multiple attachments
python3 resources/gmail_builder.py \
  send --to "colleague@example.com" \
  --subject "Multiple Attachments" \
  --body "Here are the files." \
  --attach "/path/to/file1.pdf" \
  --attach "/path/to/file2.xlsx"
```

### Download Attachment

```bash
# Get message with attachment info
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" | jq '.payload.parts[] | select(.filename != "") | {filename, attachmentId: .body.attachmentId}'

# Download attachment
curl -s "https://gmail.googleapis.com/gmail/v1/users/me/messages/${MESSAGE_ID}/attachments/${ATTACHMENT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" | jq -r '.data' | base64 -d > attachment.pdf
```

Or use the helper:

```bash
python3 resources/gmail_builder.py \
  download-attachment --message-id "MESSAGE_ID" \
  --output-dir "/path/to/downloads"
```

## Helper Scripts

### gmail_builder.py - Complete Gmail Operations

```bash
# Check authentication
python3 gmail_builder.py auth-status

# Search emails
python3 gmail_builder.py search --query "from:boss@company.com is:unread" --max-results 10

# Read message
python3 gmail_builder.py read-message --message-id "MESSAGE_ID"

# Send plain text email
python3 gmail_builder.py send --to "user@example.com" --subject "Subject" --body "Body text"

# Send HTML email
python3 gmail_builder.py send --to "user@example.com" --subject "Subject" \
  --html '<h1>Title</h1><p>Paragraph with <strong>bold</strong></p>'

# Send with attachment
python3 gmail_builder.py send --to "user@example.com" --subject "Files" \
  --body "See attached" --attach "/path/to/file.pdf"

# Create draft
python3 gmail_builder.py create-draft --to "user@example.com" --subject "Draft" --body "Draft content"

# Create reply draft (appears in correct thread)
python3 gmail_builder.py create-reply-draft --message-id "MSG_ID" --body "Thanks for your message!"

# Create reply draft with HTML
python3 gmail_builder.py create-reply-draft --message-id "MSG_ID" \
  --html '<p>Thanks!</p><ul><li>Point 1</li><li>Point 2</li></ul>'

# Create reply-all draft
python3 gmail_builder.py create-reply-draft --message-id "MSG_ID" --body "Thanks everyone!" --reply-all

# List drafts
python3 gmail_builder.py list-drafts

# Send draft
python3 gmail_builder.py send-draft --draft-id "DRAFT_ID"

# Forward message
python3 gmail_builder.py forward --message-id "MSG_ID" --to "forward@example.com" --comment "FYI"

# Reply to message
python3 gmail_builder.py reply --message-id "MSG_ID" --body "Thanks!"

# Reply all
python3 gmail_builder.py reply-all --message-id "MSG_ID" --body "Thanks everyone!"

# Add label
python3 gmail_builder.py add-label --message-id "MSG_ID" --label "STARRED"

# Remove label
python3 gmail_builder.py remove-label --message-id "MSG_ID" --label "UNREAD"

# Move to trash
python3 gmail_builder.py trash --message-id "MSG_ID"

# Mark as spam
python3 gmail_builder.py spam --message-id "MSG_ID"

# Empty trash
python3 gmail_builder.py empty-trash --confirm

# List labels
python3 gmail_builder.py list-labels

# Create custom label
python3 gmail_builder.py create-label --name "My Label"

# Download attachments
python3 gmail_builder.py download-attachment --message-id "MSG_ID" --output-dir "./downloads"
```

### gmail_auth.py - Authentication Management

```bash
python3 gmail_auth.py status    # Check auth status
python3 gmail_auth.py login     # Login with required scopes
python3 gmail_auth.py token     # Get access token
python3 gmail_auth.py validate  # Validate current token
```

## Best Practices

1. **Always check auth first** - Run `gmail_auth.py status` before API calls
2. **Use search operators** - Gmail's search syntax is powerful and efficient
3. **Batch operations** - Use batchModify/batchDelete for multiple messages
4. **HTML for rich emails** - Use multipart/alternative for best compatibility
5. **Include plain text** - Always include a text/plain part for email clients that don't render HTML
6. **Thread awareness** - Use threadId when replying to keep conversations together
7. **Be careful with delete** - Trash is recoverable, delete is permanent
8. **Test with drafts first** - Create drafts before sending to verify formatting
9. **Check quota limits** - Gmail API has rate limits; batch operations help
10. **Use helper scripts** - They handle encoding, MIME structure, and error handling

## Example: Send Formatted Newsletter

```bash
#!/bin/bash
TOKEN=$(gcloud auth application-default print-access-token)

python3 resources/gmail_builder.py \
  send --to "colleague@example.com" \
  --subject "Weekly Update - January 2025" \
  --html '
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; }
    .header { background: #1a73e8; color: white; padding: 20px; text-align: center; }
    .content { padding: 20px; }
    .highlight { background: #fef7e0; padding: 10px; border-left: 4px solid #f9ab00; }
    .footer { background: #f5f5f5; padding: 10px; text-align: center; font-size: 12px; }
  </style>
</head>
<body>
  <div class="header">
    <h1>Weekly Update</h1>
    <p>January 6, 2025</p>
  </div>
  <div class="content">
    <h2>Highlights</h2>
    <div class="highlight">
      <strong>Big Announcement:</strong> We launched the new Gmail skill!
    </div>

    <h2>This Week</h2>
    <ul>
      <li>Completed Gmail API integration</li>
      <li>Added HTML email support</li>
      <li>Implemented label management</li>
    </ul>

    <h2>Next Steps</h2>
    <ol>
      <li>Add attachment support</li>
      <li>Implement threading</li>
      <li>Add batch operations</li>
    </ol>

    <p>Questions? <a href="mailto:support@example.com">Contact us</a></p>
  </div>
  <div class="footer">
    <p>You received this because you are awesome.</p>
  </div>
</body>
</html>'
```

## Sources

- [Gmail API Reference](https://developers.google.com/gmail/api/reference/rest)
- [Gmail API Guides](https://developers.google.com/gmail/api/guides)
- [Search Operators](https://support.google.com/mail/answer/7190)
- [MIME Types](https://developers.google.com/gmail/api/guides/sending#creating_messages)
