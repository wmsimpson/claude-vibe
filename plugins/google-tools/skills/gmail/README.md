# Gmail

Search, read, compose, organize Gmail emails, and manage filters using gcloud CLI + curl. Handles any email-related request including sending, reading, drafting, forwarding, replying, attachments, labels, and filter management.

## How to Invoke

### Slash Command

```
/gmail
```

### Example Prompts

```
"Send an email to colleague@example.com with the Q4 summary"
"Check my inbox for unread emails from the last week"
"Create a draft reply to the last email from my manager"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` first to authenticate with Google Workspace |
| gcloud CLI | Must be installed (`brew install --cask google-cloud-sdk`) |
| Quota Project | Uses your GCP quota project (`$GCP_QUOTA_PROJECT`) for API billing |

## What This Skill Does

1. Authenticates via the shared `/google-auth` module
2. Searches and reads emails using Gmail API with full search operator support
3. Composes and sends plain text or rich HTML emails with formatting, attachments, and inline images
4. Manages drafts (create, update, send) with proper threading for reply drafts
5. Forwards and replies to messages (including reply-all)
6. Organizes emails with labels, starring, archiving, and trash management
7. Creates and manages Gmail filters for automatic email routing

## Key Resources

| File | Description |
|------|-------------|
| `resources/gmail_builder.py` | Complete Gmail operations helper (send, search, draft, forward, reply, labels, filters) |
| `resources/gmail_cli.sh` | Shell-based CLI helper for Gmail operations |

## Related Skills

- `/google-auth` - Required authentication before using Gmail
- `/google-docs` - Create documents that can be referenced in emails
- `/google-calendar` - Schedule meetings related to email discussions
