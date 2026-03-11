# Google Calendar

Create, modify, and manage Google Calendar events. Find available meeting times across multiple attendees using FreeBusy queries, manage recurring events, attach documents, and handle attendees with automatic availability checking.

## How to Invoke

### Slash Command

```
/google-calendar
```

### Example Prompts

```
"Schedule a 30-minute meeting with colleague@example.com next Tuesday at 10am"
"Find a time when the whole team is free this week for a 1-hour sync"
"What's on my calendar for tomorrow?"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` first to authenticate with Google Workspace |
| gcloud CLI | Must be installed (`brew install --cask google-cloud-sdk`) |
| Quota Project | Uses your GCP quota project (`$GCP_QUOTA_PROJECT`) for API billing |

## What This Skill Does

1. Gets calendar context (timezone, current date/time) before any operation
2. Creates events with Google Meet links, attendees, and document attachments
3. Automatically checks attendee availability before booking and analyzes conflict reschedulability
4. Finds available meeting slots across multiple attendees with working-hours filters
5. Manages recurring events with full RRULE support (daily, weekly, bi-weekly, monthly)
6. Validates date ranges to prevent day-of-week and timezone errors
7. Updates, reschedules, and deletes events with attendee notifications

## Key Resources

| File | Description |
|------|-------------|
| `resources/gcal_builder.py` | Complete calendar operations (create, search, find-availability, freebusy, validate-dates, get-context) |
| `resources/gcal_cli.sh` | Shell-based CLI helper for calendar operations |

## Related Skills

- `/google-auth` - Required authentication before using Calendar
- `/google-docs` - Create meeting notes documents to attach to events
- `/google-tasks` - Create follow-up tasks from meetings
