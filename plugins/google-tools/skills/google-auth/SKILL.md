---
name: google-auth
description: Unified authentication for all Google Workspace APIs (Docs, Sheets, Slides, Drive, Calendar, Gmail, Tasks, Forms)
---

# Google Authentication Skill

Unified authentication for all Google Workspace APIs. This skill is the single source of truth for Google authentication - all other Google skills (gmail, google-calendar, google-docs, google-forms, google-sheets, google-slides, google-tasks) use this for auth.

**Run this skill FIRST before using any other Google skill.**

## Quick Start

```bash
# Check if you're authenticated
python3 resources/google_auth.py status

# Login (will open browser if needed)
python3 resources/google_auth.py login
```

## CRITICAL: Authentication Flow Requirements

**IMPORTANT FOR AGENTS:** When running the login command:

1. **The user MUST complete the OAuth flow** in the browser window that opens
2. **If authentication fails**, it means the user did NOT complete the OAuth flow in time
3. **DO NOT try alternative authentication methods** such as:
   - Creating OAuth client IDs manually
   - Creating service account credentials
   - Using different gcloud auth commands
   - Setting up application credentials manually
4. **The ONLY solution** is to re-run the login command and have the user complete the browser OAuth flow
5. **If the user doesn't complete OAuth**: Stop, explain what happened, and ask if they want to retry

The `python3 resources/google_auth.py login` command includes automatic retry logic with clear instructions.

## Commands

### Check Authentication Status

Shows gcloud installation, credentials, token validity, scopes, and API access for all Google services:

```bash
python3 resources/google_auth.py status
```

### Login

Authenticates with all required scopes for Google Workspace:

```bash
# Normal login (only re-auths if needed)
python3 resources/google_auth.py login

# Force re-authentication
python3 resources/google_auth.py login --force
```

### Get Access Token

For use in scripts and API calls:

```bash
TOKEN=$(python3 resources/google_auth.py token)
```

### Validate Token

Check if current token has all required scopes:

```bash
python3 resources/google_auth.py validate
```

### Show Login Command

Displays the gcloud command if you want to run it manually:

```bash
python3 resources/google_auth.py show-login-command
```

## gcloud Installation

The script automatically finds gcloud using `which gcloud`. If gcloud is not installed, it will attempt to install it via Homebrew:

```bash
brew install --cask google-cloud-sdk
```

If Homebrew is not available, you'll be directed to install gcloud manually from:
https://cloud.google.com/sdk/docs/install

## Required Scopes

One authentication covers ALL Google Workspace APIs:

| Scope | Service |
|-------|---------|
| `https://www.googleapis.com/auth/drive` | Google Drive |
| `https://www.googleapis.com/auth/documents` | Google Docs |
| `https://www.googleapis.com/auth/presentations` | Google Slides |
| `https://www.googleapis.com/auth/spreadsheets` | Google Sheets |
| `https://www.googleapis.com/auth/calendar` | Google Calendar |
| `https://www.googleapis.com/auth/gmail.modify` | Gmail |
| `https://www.googleapis.com/auth/tasks` | Google Tasks |
| `https://www.googleapis.com/auth/forms.body` | Google Forms (read/write) |
| `https://www.googleapis.com/auth/forms.responses.readonly` | Google Forms (responses) |
| `https://www.googleapis.com/auth/cloud-platform` | GCP Access |

## Quota Project

All API calls require a quota project header:

```bash
-H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

This is automatically used by all Google skills.

## Credentials Location

Credentials are stored in the standard Application Default Credentials (ADC) location:

```
~/.config/gcloud/application_default_credentials.json
```

## Troubleshooting

### "Authentication failed" - OAuth not completed

**Symptom:** The login command fails and shows "Authentication Failed"

**Cause:** The user did not complete the OAuth flow in the browser window

**Solution:**
1. Run the login command again: `python3 resources/google_auth.py login`
2. When the browser opens, click through the OAuth prompts
3. Click "Allow" to grant permissions
4. The command will automatically offer to retry if you miss the window

**DO NOT:**
- Try to create OAuth credentials manually
- Attempt alternative authentication methods
- Create service accounts or application credentials

### "gcloud not found"

The script will try to install gcloud via Homebrew. If that fails:

1. Install manually: https://cloud.google.com/sdk/docs/install
2. Restart your terminal
3. Run `which gcloud` to verify installation

### "Missing scopes"

Run login with force flag to get all required scopes:

```bash
python3 resources/google_auth.py login --force
```

### "API access failed"

1. Check your GCP quota project env var is set: `echo $GCP_QUOTA_PROJECT` (see configure-vibe step 6)
2. Verify token is valid: `python3 resources/google_auth.py validate`
3. Check specific API is enabled in GCP console

### Token expired

Tokens auto-refresh, but if issues persist:

```bash
python3 resources/google_auth.py login --force
```

## Usage with Other Google Skills

After running `/google-auth`, you can use any Google skill:

- `/gmail` - Email operations
- `/google-calendar` - Calendar events
- `/google-docs` - Document creation
- `/google-sheets` - Spreadsheet operations
- `/google-slides` - Presentation creation
- `/google-tasks` - Task management
- `/google-forms` - Form creation and response collection

All these skills share the same authentication token.

## Manual API Calls

After authentication, you can make direct API calls:

```bash
TOKEN=$(python3 resources/google_auth.py token)

# Example: List Drive files
curl -s "https://www.googleapis.com/drive/v3/files?pageSize=10" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Example: Get Calendar events
curl -s "https://www.googleapis.com/calendar/v3/calendars/primary/events?maxResults=10" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```
