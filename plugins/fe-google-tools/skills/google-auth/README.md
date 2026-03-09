# Google Authentication

Unified authentication for all Google Workspace APIs (Docs, Sheets, Slides, Drive, Calendar, Gmail, Tasks). This is the single source of truth for Google authentication -- run this first before using any other Google skill.

## How to Invoke

### Slash Command

```
/google-auth
```

### Example Prompts

```
"Authenticate with Google Workspace"
"Log in to Google so I can use Gmail and Calendar"
"Check my Google auth status"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| gcloud CLI | Must be installed (`brew install --cask google-cloud-sdk`) |
| Browser Access | OAuth flow requires completing consent in a browser |
| GCP Project | A free GCP project set as `GCP_QUOTA_PROJECT` env var |

## What This Skill Does

1. Checks if gcloud is installed (auto-installs via Homebrew if missing)
2. Authenticates with all required Google Workspace scopes in a single OAuth flow
3. Stores credentials in the standard Application Default Credentials location
4. Provides token retrieval for use by all other Google skills
5. Validates token scopes and API access across all Google services

## Key Resources

| File | Description |
|------|-------------|
| `resources/google_auth.py` | Auth management script (status, login, token, validate, show-login-command) |

## Related Skills

- `/gmail` - Email operations (requires this auth)
- `/google-calendar` - Calendar management (requires this auth)
- `/google-docs` - Document operations (requires this auth)
- `/google-sheets` - Spreadsheet operations (requires this auth)
- `/google-slides` - Presentation creation (requires this auth)
- `/google-tasks` - Task management (requires this auth)
