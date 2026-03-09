---
name: google-calendar
description: Create, modify, and manage Google Calendar events. Find meeting times when attendees are available using rich calendar queries and FreeBusy.
---

# Google Calendar Skill

Manage Google Calendar events using gcloud CLI + curl. This skill provides patterns and utilities for creating meetings with Google Meet links, managing attendees, searching events, attaching documents, handling recurring meetings, and **finding available meeting times** when multiple attendees need to meet.

## CRITICAL: Always Get Context First

**Before ANY calendar operation (scheduling, searching, or availability checks), ALWAYS run:**

```bash
python3 resources/gcal_builder.py get-context
```

This automatically infers and displays:
- **User's timezone** from their Google Calendar settings (authoritative source)
- **Current date, time, and year** in their timezone
- **Day of week** to prevent scheduling errors
- **Upcoming events** for context

**Why this is critical:**
- Prevents scheduling meetings in the wrong year (e.g., 2025 vs 2026)
- Prevents timezone errors (user expects PST but gets UTC)
- Prevents day-of-week errors (thinking Friday is Jan 31 when it's actually Saturday)
- Shows what the user's calendar already has scheduled

**Example output:**

```json
{
  "timezone": {
    "name": "America/Los_Angeles",
    "abbreviation": "PST",
    "offset": "-0800",
    "offset_hours": "-08:00"
  },
  "current": {
    "datetime": "2026-01-27T14:30:00-08:00",
    "date": "2026-01-27",
    "year": 2026,
    "month": 1,
    "day": 27,
    "day_of_week": "Tuesday",
    "time": "02:30 PM",
    "time_24h": "14:30"
  },
  "upcoming_events": [
    {
      "summary": "Team Sync",
      "start": "2026-01-28T10:00:00-08:00",
      "day": "Wednesday, January 28, 2026"
    }
  ],
  "context_summary": "Today is Tuesday, January 27, 2026 at 02:30 PM PST (timezone: America/Los_Angeles)"
}
```

**Use this context for ALL subsequent operations** - the timezone and dates are automatically detected from the user's calendar settings.

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

### Calendar IDs

- `primary` - The user's primary calendar
- Email addresses can be used as calendar IDs for shared calendars
- Calendar IDs are unique identifiers for each calendar

### Event Times

All times use RFC 3339 format with timezone:
- `2025-01-15T10:00:00-08:00` (Pacific time)
- `2025-01-15T18:00:00Z` (UTC)

For all-day events, use date format:
- `2025-01-15` (all-day event)

### Event IDs

Each event has a unique ID used for updates, deletions, and queries.

## API Reference

### Base URL

```
https://www.googleapis.com/calendar/v3
```

## Creating Events

**CRITICAL: Availability is automatically checked before booking**

By default, before creating any event, the skill automatically checks availability:
- **All attendees** are checked for conflicts by default
- If any conflicts are found, the event is **NOT created**
- Returns detailed conflict information with intelligent reschedulability analysis
- At minimum, the **organizer's availability is always checked**

**Intelligent Conflict Resolution**

The skill analyzes each conflict to determine if it can be rescheduled:

**Can reschedule** (with user approval):
- **INVESTech meetings** - Internal technical meetings
- **Internal-only team calls** - Team syncs, standups, etc.
- **Internal-only 1:1s** - One-on-one meetings with no external attendees
- If you're the organizer of a 1:1, it can potentially be moved

**Cannot reschedule**:
- **Meetings with external attendees** - Never schedule over these
- **Unknown meeting types** - Require user validation

**IMPORTANT:** All rescheduling requires explicit user approval before moving or booking over any existing meeting.

**Example conflict response with reschedulability:**
```json
{
  "error": "Availability conflicts found",
  "message": "2/3 attendee(s) have conflicts (1 potentially reschedulable)",
  "reschedulable_conflicts": [
    {
      "email": "person@example.com",
      "period": {"start": "2026-01-28T10:00:00-08:00", "end": "2026-01-28T11:00:00-08:00"},
      "analysis": {
        "can_reschedule": true,
        "reason": "Internal investech meeting - can potentially reschedule with user approval",
        "meeting_type": "investech",
        "summary": "INVESTech [Weekly]",
        "is_organizer": false,
        "has_external_attendees": false
      }
    }
  ],
  "hard_conflicts": [
    {
      "email": "person@example.com",
      "period": {"start": "2026-01-28T14:00:00-08:00", "end": "2026-01-28T15:00:00-08:00"},
      "analysis": {
        "can_reschedule": false,
        "reason": "Has external attendees - should not reschedule",
        "meeting_type": "external",
        "has_external_attendees": true
      }
    }
  ],
  "requires_user_approval": true
}
```

**IMPORTANT: Organizer is automatically added as an attendee**

By default, when you create an event with attendees, the organizer (authenticated user) is automatically added to the attendee list. This ensures they appear in the attendee list, not just as the organizer.

This behavior can be disabled by passing `include_organizer=False` to the `create_event()` function if needed.

### Create Simple Event

```bash
TOKEN=$(gcloud auth application-default print-access-token)

curl -s -X POST "https://www.googleapis.com/calendar/v3/calendars/primary/events" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Team Meeting",
    "description": "Weekly team sync",
    "start": {
      "dateTime": "2025-01-15T10:00:00-08:00",
      "timeZone": "America/Los_Angeles"
    },
    "end": {
      "dateTime": "2025-01-15T11:00:00-08:00",
      "timeZone": "America/Los_Angeles"
    }
  }'
```

### Create Event with Google Meet Link (Default for New Meetings)

```bash
curl -s -X POST "https://www.googleapis.com/calendar/v3/calendars/primary/events?conferenceDataVersion=1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Team Meeting",
    "description": "Weekly team sync",
    "start": {
      "dateTime": "2025-01-15T10:00:00-08:00",
      "timeZone": "America/Los_Angeles"
    },
    "end": {
      "dateTime": "2025-01-15T11:00:00-08:00",
      "timeZone": "America/Los_Angeles"
    },
    "attendees": [
      {"email": "colleague@example.com"},
      {"email": "bkvarda@squareup.com"}
    ],
    "conferenceData": {
      "createRequest": {
        "requestId": "meet-'$(date +%s)'",
        "conferenceSolutionKey": {
          "type": "hangoutsMeet"
        }
      }
    }
  }'
```

**Important:** The `conferenceDataVersion=1` query parameter is required to create/modify conference data.

### Create All-Day Event

```bash
curl -s -X POST "https://www.googleapis.com/calendar/v3/calendars/primary/events" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Company Holiday",
    "start": {
      "date": "2025-01-20"
    },
    "end": {
      "date": "2025-01-21"
    }
  }'
```

### Using the Helper Script

```bash
# Create event with Meet link (default)
python3 resources/gcal_builder.py \
  create --summary "Team Sync" \
  --start "2025-01-15T10:00:00" \
  --end "2025-01-15T11:00:00" \
  --attendees "colleague@example.com,bkvarda@squareup.com" \
  --description "Weekly team synchronization meeting"

# Create without Meet link
python3 resources/gcal_builder.py \
  create --summary "Lunch" \
  --start "2025-01-15T12:00:00" \
  --end "2025-01-15T13:00:00" \
  --no-meet
```

## Managing Attendees

### Add Attendees to Existing Event

First get the event, then update with new attendees:

```bash
# Get current event
EVENT=$(curl -s "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}")

# Update with additional attendees
curl -s -X PATCH "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}?sendUpdates=all" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "attendees": [
      {"email": "existing@example.com"},
      {"email": "new-attendee@example.com"}
    ]
  }'
```

### Remove Attendees

Update the event with the attendee list excluding the person to remove:

```bash
curl -s -X PATCH "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}?sendUpdates=all" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "attendees": [
      {"email": "keep-this-person@example.com"}
    ]
  }'
```

### Using Helper Script for Attendees

```bash
# Add attendees
python3 gcal_builder.py add-attendees --event-id "EVENT_ID" \
  --attendees "new1@example.com,new2@example.com"

# Remove attendees
python3 gcal_builder.py remove-attendees --event-id "EVENT_ID" \
  --attendees "remove@example.com"
```

### sendUpdates Parameter Options

- `all` - Send notifications to all attendees
- `externalOnly` - Send notifications only to non-Google Calendar users
- `none` - Don't send notifications

## Searching and Reading Events

### List Upcoming Events

```bash
# List next 10 events
curl -s "https://www.googleapis.com/calendar/v3/calendars/primary/events?maxResults=10&orderBy=startTime&singleEvents=true&timeMin=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Search Events by Query

```bash
# Search by text (searches summary, description, location, attendees)
curl -s "https://www.googleapis.com/calendar/v3/calendars/primary/events?q=team+meeting&maxResults=10&singleEvents=true" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Get Events in Date Range

```bash
# Events between two dates
TIME_MIN="2025-01-01T00:00:00Z"
TIME_MAX="2025-01-31T23:59:59Z"

curl -s "https://www.googleapis.com/calendar/v3/calendars/primary/events?timeMin=${TIME_MIN}&timeMax=${TIME_MAX}&singleEvents=true&orderBy=startTime" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Get Single Event

```bash
curl -s "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Using Helper Script for Search

```bash
# List upcoming events
python3 gcal_builder.py list --max-results 10

# Search events
python3 gcal_builder.py search --query "team meeting"

# Events in date range
python3 gcal_builder.py list --start "2025-01-01" --end "2025-01-31"

# Get specific event
python3 gcal_builder.py get --event-id "EVENT_ID"
```

## Finding Available Meeting Times

There are two approaches for checking availability. **Always prefer Rich Calendar Queries** for same-org (Google Workspace) colleagues, and fall back to FreeBusy only for external contacts or when direct calendar access is unavailable.

### Approach 1 (Preferred): Rich Calendar Queries

For colleagues in the same Google Workspace organization, you can query their calendar directly using their email as the calendar ID. This returns **full event details** - not just busy/free blocks - giving you the information needed to make intelligent scheduling decisions.

**Why this is better than FreeBusy:**

| Data point | Rich Calendar Query | FreeBusy API |
|-----------|-------------------|-------------|
| Busy/free times | Yes | Yes |
| Event name/summary | **Yes** | No |
| Attendee list | **Yes** | No |
| Accept/decline status | **Yes** | No |
| External attendee detection | **Yes** | No |
| Event type (office hours, hold, etc.) | **Yes** | No |

This matters because it lets you distinguish **soft conflicts** (unaccepted office hours, optional meetings, personal holds) from **hard conflicts** (accepted 1:1s, external customer meetings) - which is critical for finding realistic meeting times.

#### Query a Colleague's Calendar Directly

```bash
TOKEN=$(python3 ../google-auth/resources/google_auth.py token)

# Query another person's calendar using their email as the calendar ID
# URL-encode the @ symbol as %40
curl -s "https://www.googleapis.com/calendar/v3/calendars/colleague%40databricks.com/events?timeMin=2026-02-17T00:00:00-08:00&timeMax=2026-02-17T23:59:59-08:00&singleEvents=true&orderBy=startTime&maxResults=50" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

**Key details:**
- The calendar ID is the person's email with `@` URL-encoded as `%40`
- Returns `accessRole: "reader"` for same-org colleagues (read-only, cannot modify their events)
- Returns full event objects including `summary`, `attendees`, `responseStatus`, `description`, `start`, `end`
- Each attendee object includes `responseStatus`: `accepted`, `declined`, `tentative`, or `needsAction`

#### Example: Rich Availability Check for Multiple People

```bash
TOKEN=$(python3 ../google-auth/resources/google_auth.py token)

# Query each person's calendar in parallel and analyze conflicts
# Person 1: your own calendar
curl -s "https://www.googleapis.com/calendar/v3/calendars/primary/events?timeMin=2026-02-17T00:00:00-08:00&timeMax=2026-02-17T23:59:59-08:00&singleEvents=true&orderBy=startTime&maxResults=50" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Person 2: colleague's calendar
curl -s "https://www.googleapis.com/calendar/v3/calendars/person2%40databricks.com/events?timeMin=2026-02-17T00:00:00-08:00&timeMax=2026-02-17T23:59:59-08:00&singleEvents=true&orderBy=startTime&maxResults=50" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Person 3: another colleague's calendar
curl -s "https://www.googleapis.com/calendar/v3/calendars/person3%40databricks.com/events?timeMin=2026-02-17T00:00:00-08:00&timeMax=2026-02-17T23:59:59-08:00&singleEvents=true&orderBy=startTime&maxResults=50" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

Then analyze the results across all calendars to find open slots.

#### Classifying Conflicts as Soft vs Hard

When analyzing events from rich calendar queries, classify each conflict:

**Soft conflicts** (can potentially schedule over, with user approval):
- Events with `responseStatus: "needsAction"` (not yet accepted)
- Events with `responseStatus: "tentative"`
- Events with `responseStatus: "declined"` (already declined - not a real conflict)
- Office hours (summary contains "office hour")
- Personal holds/blocks (summary contains "HOLD", "DoNotBook", "DNS", "block")
- Auto-scheduled blocks (summary contains "Clockwise", "Lunch via Clockwise")
- Large broadcast meetings (many attendees, not individually critical)

**Hard conflicts** (do NOT schedule over without explicit approval):
- Events with `responseStatus: "accepted"` that are 1:1s or small meetings
- Events with external attendees (email domain is not your primary domain)
- Customer-facing meetings

**Events to skip entirely when analyzing busy times:**
- All-day events (no `dateTime` in start, only `date`) - these are typically OOO markers or working locations
- Events with `transparency: "transparent"` - these don't block the calendar
- Events with `eventType: "workingLocation"` - these are "Home"/"Office" markers

#### When Rich Calendar Queries Fail

If querying a colleague's calendar returns an error (403 or `notFound`), fall back to the FreeBusy API. This can happen for:
- External contacts (outside your Google Workspace org)
- Colleagues who have restricted calendar sharing
- Service accounts or room calendars with limited permissions

### Approach 2 (Fallback): FreeBusy API

Use the FreeBusy API when you cannot access a person's calendar directly. It returns only opaque busy/free time blocks with no event details.

### CRITICAL: Always Get Context and Validate Date Ranges First

**Before scheduling meetings, you MUST:**

1. **Get calendar context** (timezone, current date/time/year):
```bash
python3 resources/gcal_builder.py get-context
```

2. **Validate your date range** corresponds to the correct days of the week:
```bash
python3 resources/gcal_builder.py validate-dates --start "2026-01-27" --end "2026-01-31"
```

These steps prevent common scheduling errors:
- Wrong year (2025 vs 2026)
- Wrong timezone (PST vs UTC)
- Day-of-week mismatches (thinking Friday is Jan 31 when it's actually Saturday)

Example output:
```json
{
  "is_valid": true,
  "start": "2025-01-27",
  "end": "2025-01-31",
  "total_days": 5,
  "days": [
    {"date": "2025-01-27", "day_of_week": "Monday", "is_weekend": false},
    {"date": "2025-01-28", "day_of_week": "Tuesday", "is_weekend": false},
    {"date": "2025-01-29", "day_of_week": "Wednesday", "is_weekend": false},
    {"date": "2025-01-30", "day_of_week": "Thursday", "is_weekend": false},
    {"date": "2025-01-31", "day_of_week": "Friday", "is_weekend": false}
  ],
  "warnings": null
}
```

If weekend days are included, you'll see a warning:
```json
{
  "warnings": ["Date range includes weekend days: Saturday 01/31, Sunday 02/01"]
}
```

**The `find-availability` command also includes date validation in its output** to help catch errors before they cause problems.

### Query Free/Busy Information

```bash
TOKEN=$(gcloud auth application-default print-access-token)

# Query free/busy for multiple people over a date range
curl -s -X POST "https://www.googleapis.com/calendar/v3/freeBusy" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "timeMin": "2025-01-20T00:00:00-08:00",
    "timeMax": "2025-01-24T23:59:59-08:00",
    "timeZone": "America/Los_Angeles",
    "items": [
      {"id": "person1@example.com"},
      {"id": "person2@example.com"},
      {"id": "person3@example.com"}
    ]
  }'
```

### FreeBusy Response Format

The response contains busy periods for each calendar:

```json
{
  "kind": "calendar#freeBusy",
  "timeMin": "2025-01-20T08:00:00.000Z",
  "timeMax": "2025-01-25T07:59:59.000Z",
  "calendars": {
    "person1@example.com": {
      "busy": [
        {"start": "2025-01-20T17:00:00Z", "end": "2025-01-20T18:00:00Z"},
        {"start": "2025-01-21T15:00:00Z", "end": "2025-01-21T16:00:00Z"}
      ]
    },
    "person2@example.com": {
      "busy": [
        {"start": "2025-01-20T18:00:00Z", "end": "2025-01-20T19:00:00Z"}
      ]
    }
  }
}
```

### Using Helper Script to Find Availability

The helper script can automatically find optimal meeting slots:

```bash
# Find 30-minute slots where everyone is available
python3 resources/gcal_builder.py find-availability \
  --attendees "person1@example.com,person2@example.com,person3@example.com" \
  --start "2025-01-20T00:00:00" \
  --end "2025-01-24T23:59:59" \
  --duration 30

# Find 1-hour slots during specific working hours
python3 resources/gcal_builder.py find-availability \
  --attendees "team-lead@example.com,engineer@example.com" \
  --start "2025-01-20T00:00:00" \
  --end "2025-01-24T23:59:59" \
  --duration 60 \
  --working-hours-start 9 \
  --working-hours-end 17

# Raw free/busy query
python3 resources/gcal_builder.py freebusy \
  --attendees "person1@example.com,person2@example.com" \
  --start "2025-01-20T00:00:00" \
  --end "2025-01-24T23:59:59"
```

### Find-Availability Response Format

The helper returns organized results with **date validation included**:

```json
{
  "query": {
    "attendees": ["person1@example.com", "person2@example.com"],
    "time_range": {"start": "2025-01-20T00:00:00", "end": "2025-01-24T23:59:59"},
    "duration_minutes": 30,
    "working_hours": "9:00-17:00",
    "timezone": "America/Los_Angeles"
  },
  "date_validation": {
    "is_valid": true,
    "start": "2025-01-20",
    "end": "2025-01-24",
    "total_days": 5,
    "days": [
      {"date": "2025-01-20", "day_of_week": "Monday", "is_weekend": false},
      {"date": "2025-01-21", "day_of_week": "Tuesday", "is_weekend": false},
      {"date": "2025-01-22", "day_of_week": "Wednesday", "is_weekend": false},
      {"date": "2025-01-23", "day_of_week": "Thursday", "is_weekend": false},
      {"date": "2025-01-24", "day_of_week": "Friday", "is_weekend": false}
    ],
    "warnings": null
  },
  "all_available": [
    {
      "start": "2025-01-20T09:00:00-08:00",
      "end": "2025-01-20T09:30:00-08:00",
      "day": "Mon 1/20",
      "time": "9:00am-9:30am",
      "available": ["person1@example.com", "person2@example.com"],
      "busy": [],
      "available_count": 2,
      "total_count": 2
    }
  ],
  "some_available": [
    {
      "start": "2025-01-21T14:00:00-08:00",
      "end": "2025-01-21T14:30:00-08:00",
      "day": "Tue 1/21",
      "time": "2:00pm-2:30pm",
      "available": ["person1@example.com"],
      "busy": ["person2@example.com"],
      "available_count": 1,
      "total_count": 2
    }
  ],
  "summary": {
    "total_attendees": 2,
    "slots_all_available": 15,
    "slots_some_available": 8
  }
}
```

**Key improvements:**
- `date_validation` section shows all days in the range with their day of week
- `day` field shows human-readable day (e.g., "Mon 1/20")
- `time` field shows human-readable time (e.g., "9:00am-9:30am")
- Weekend days are automatically skipped in availability search
- Warnings alert you if weekend days are in your date range

### FreeBusy Limitations

- **Maximum 50 calendars** per query (`calendarExpansionMax`)
- **Maximum 100 members** per group expansion (`groupExpansionMax`)
- Requires read access to calendars (may show errors for external users)
- Only returns busy periods, not event details (privacy protection)

### Common Errors

| Error | Meaning |
|-------|---------|
| `notFound` | Calendar doesn't exist or you don't have access |
| `groupTooBig` | Group has too many members for a single query |
| `tooManyCalendarsRequested` | Exceeded 50 calendar limit |

### Workflow: Find Time and Schedule Meeting

**Preferred workflow using rich calendar queries (same-org colleagues):**

```bash
TOKEN=$(python3 ../google-auth/resources/google_auth.py token)

# 1. Query each attendee's calendar directly for the target date range
# Your calendar:
curl -s "https://www.googleapis.com/calendar/v3/calendars/primary/events?timeMin=2026-01-20T00:00:00-08:00&timeMax=2026-01-24T23:59:59-08:00&singleEvents=true&orderBy=startTime&maxResults=50" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Colleague's calendar:
curl -s "https://www.googleapis.com/calendar/v3/calendars/colleague%40databricks.com/events?timeMin=2026-01-20T00:00:00-08:00&timeMax=2026-01-24T23:59:59-08:00&singleEvents=true&orderBy=startTime&maxResults=50" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# 2. For each time slot, check if events overlap and classify as soft/hard conflicts
# 3. Present slots with conflict details to the user
# 4. Create meeting at the chosen slot
python3 resources/gcal_builder.py create \
  --summary "Team Sync" \
  --start "2026-01-20T09:00:00" \
  --end "2026-01-20T09:30:00" \
  --attendees "colleague2@example.com" \
  --description "Scheduled at a time when everyone is available"
```

**Fallback workflow using FreeBusy (external contacts):**

```bash
# 1. Find available slots using FreeBusy
AVAILABILITY=$(python3 resources/gcal_builder.py find-availability \
  --attendees "external-person@othercorp.com,colleague2@example.com" \
  --start "2026-01-20T00:00:00" \
  --end "2026-01-24T23:59:59" \
  --duration 30)

# 2. Review the all_available slots
echo "$AVAILABILITY" | jq '.all_available[:5]'

# 3. Create meeting at the best slot
python3 resources/gcal_builder.py create \
  --summary "Team Sync" \
  --start "2026-01-20T09:00:00" \
  --end "2026-01-20T09:30:00" \
  --attendees "external-person@othercorp.com,colleague2@example.com" \
  --description "Scheduled at a time when everyone is available"
```

## Attachments and Documents

### Event Response with Attachments

Events can have attachments in the `attachments` field:

```json
{
  "attachments": [
    {
      "fileUrl": "https://drive.google.com/open?id=FILE_ID",
      "title": "Meeting Notes",
      "mimeType": "application/vnd.google-apps.document",
      "iconLink": "https://drive-thirdparty.googleusercontent.com/16/type/application/vnd.google-apps.document",
      "fileId": "FILE_ID"
    }
  ]
}
```

### Create Event with Attachment

```bash
curl -s -X POST "https://www.googleapis.com/calendar/v3/calendars/primary/events?supportsAttachments=true" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Planning Meeting",
    "start": {"dateTime": "2025-01-15T14:00:00-08:00", "timeZone": "America/Los_Angeles"},
    "end": {"dateTime": "2025-01-15T15:00:00-08:00", "timeZone": "America/Los_Angeles"},
    "attachments": [
      {
        "fileUrl": "https://docs.google.com/document/d/DOC_ID/edit",
        "title": "Meeting Agenda"
      }
    ]
  }'
```

### Add Attachment to Existing Event

```bash
curl -s -X PATCH "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}?supportsAttachments=true" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "attachments": [
      {
        "fileUrl": "https://docs.google.com/document/d/DOC_ID/edit",
        "title": "Running Notes"
      }
    ]
  }'
```

### Using Helper for Attachments

```bash
# Create event with attachment
python3 gcal_builder.py create --summary "Planning" \
  --start "2025-01-15T14:00:00" --end "2025-01-15T15:00:00" \
  --attach-doc "DOC_ID" --attach-title "Meeting Notes"

# Add attachment to existing event
python3 gcal_builder.py attach --event-id "EVENT_ID" \
  --doc-id "DOC_ID" --title "Running Notes"

# List attachments on an event
python3 gcal_builder.py get --event-id "EVENT_ID" --show-attachments
```

### Creating a Running Notes Document

```bash
# Create a Google Doc for meeting notes
python3 ../google-docs/resources/gdocs_builder.py \
  create --title "Meeting Notes - Team Sync 2025-01-15"

# Then attach it to the calendar event
python3 gcal_builder.py attach --event-id "EVENT_ID" \
  --doc-id "NEW_DOC_ID" --title "Running Notes"
```

## Recurring Events (Cadence)

### Recurrence Rule Format (RRULE)

Google Calendar uses iCalendar RRULE format:

```
RRULE:FREQ=<frequency>;[INTERVAL=<n>];[BYDAY=<days>];[COUNT=<n>|UNTIL=<date>]
```

**Frequencies:**
- `DAILY` - Every day
- `WEEKLY` - Every week
- `MONTHLY` - Every month
- `YEARLY` - Every year

**Examples:**
- `RRULE:FREQ=WEEKLY;BYDAY=MO,WE,FR` - Every Mon, Wed, Fri
- `RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=TU` - Every 2 weeks on Tuesday
- `RRULE:FREQ=MONTHLY;BYDAY=1MO` - First Monday of every month
- `RRULE:FREQ=DAILY;COUNT=10` - Daily for 10 occurrences
- `RRULE:FREQ=WEEKLY;UNTIL=20251231T235959Z` - Weekly until end of 2025

### Create Recurring Event

```bash
curl -s -X POST "https://www.googleapis.com/calendar/v3/calendars/primary/events?conferenceDataVersion=1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Weekly Team Standup",
    "start": {"dateTime": "2025-01-13T09:00:00-08:00", "timeZone": "America/Los_Angeles"},
    "end": {"dateTime": "2025-01-13T09:30:00-08:00", "timeZone": "America/Los_Angeles"},
    "recurrence": ["RRULE:FREQ=WEEKLY;BYDAY=MO,WE,FR"],
    "attendees": [
      {"email": "colleague@example.com"},
      {"email": "bkvarda@squareup.com"}
    ],
    "conferenceData": {
      "createRequest": {
        "requestId": "standup-'$(date +%s)'",
        "conferenceSolutionKey": {"type": "hangoutsMeet"}
      }
    }
  }'
```

### Change Recurrence (Cadence)

```bash
# Change from weekly to bi-weekly
curl -s -X PATCH "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "recurrence": ["RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO"]
  }'
```

### Using Helper for Recurrence

```bash
# Create weekly meeting
python3 gcal_builder.py create --summary "Weekly Sync" \
  --start "2025-01-13T09:00:00" --end "2025-01-13T09:30:00" \
  --recurrence "WEEKLY" --days "MO,WE,FR" \
  --attendees "colleague@example.com"

# Create bi-weekly meeting
python3 gcal_builder.py create --summary "Bi-weekly 1:1" \
  --start "2025-01-14T14:00:00" --end "2025-01-14T14:30:00" \
  --recurrence "WEEKLY" --interval 2 --days "TU" \
  --attendees "manager@example.com"

# Change cadence of existing event
python3 gcal_builder.py set-recurrence --event-id "EVENT_ID" \
  --recurrence "WEEKLY" --interval 2 --days "MO"

# Remove recurrence (make single event)
python3 gcal_builder.py set-recurrence --event-id "EVENT_ID" --remove
```

### Common Cadence Patterns

| Pattern | RRULE |
|---------|-------|
| Daily | `RRULE:FREQ=DAILY` |
| Weekly on Monday | `RRULE:FREQ=WEEKLY;BYDAY=MO` |
| Every weekday | `RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR` |
| Bi-weekly | `RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO` |
| Monthly on 1st Monday | `RRULE:FREQ=MONTHLY;BYDAY=1MO` |
| Monthly on 15th | `RRULE:FREQ=MONTHLY;BYMONTHDAY=15` |
| Quarterly | `RRULE:FREQ=MONTHLY;INTERVAL=3` |

## Modifying Events

### Update Title and Description

```bash
curl -s -X PATCH "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}?sendUpdates=all" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Updated Meeting Title",
    "description": "Updated description with <a href=\"https://databricks.com\">embedded link</a>"
  }'
```

### Description with Rich Content

The description field supports HTML for links:

```json
{
  "description": "<b>Agenda:</b>\n<ul>\n<li>Review Q4 results</li>\n<li>Plan Q1 initiatives</li>\n</ul>\n\n<b>Resources:</b>\n<a href=\"https://docs.google.com/document/d/DOC_ID\">Meeting Notes</a>\n<a href=\"https://databricks.com\">Company Website</a>"
}
```

### Update Time

```bash
curl -s -X PATCH "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}?sendUpdates=all" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "start": {"dateTime": "2025-01-15T11:00:00-08:00", "timeZone": "America/Los_Angeles"},
    "end": {"dateTime": "2025-01-15T12:00:00-08:00", "timeZone": "America/Los_Angeles"}
  }'
```

### Update Location

```bash
curl -s -X PATCH "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "location": "Conference Room A / Google Meet"
  }'
```

### Using Helper for Updates

```bash
# Update title
python3 gcal_builder.py update --event-id "EVENT_ID" --summary "New Title"

# Update description with links
python3 gcal_builder.py update --event-id "EVENT_ID" \
  --description "Check out <a href='https://databricks.com'>Databricks</a>"

# Update time
python3 gcal_builder.py update --event-id "EVENT_ID" \
  --start "2025-01-15T11:00:00" --end "2025-01-15T12:00:00"

# Update multiple fields
python3 gcal_builder.py update --event-id "EVENT_ID" \
  --summary "Team Planning" \
  --description "Quarterly planning session" \
  --location "Main Conference Room"
```

## Delete Events

### Delete Single Event

```bash
curl -s -X DELETE "https://www.googleapis.com/calendar/v3/calendars/primary/events/${EVENT_ID}?sendUpdates=all" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Delete Single Instance of Recurring Event

For recurring events, each instance has an ID like `eventId_20250115T170000Z`. Delete that specific instance:

```bash
curl -s -X DELETE "https://www.googleapis.com/calendar/v3/calendars/primary/events/${INSTANCE_ID}?sendUpdates=all" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Using Helper

```bash
python3 gcal_builder.py delete --event-id "EVENT_ID"
python3 gcal_builder.py delete --event-id "EVENT_ID" --notify  # Send cancellation emails
```

## Helper Scripts

### gcal_builder.py - Complete Calendar Operations

```bash
# ALWAYS START HERE: Get calendar context (timezone, current date/time/year)
python3 gcal_builder.py get-context

# Authentication
python3 gcal_builder.py auth-status

# List events
python3 gcal_builder.py list --max-results 10
python3 gcal_builder.py list --start "2026-01-01" --end "2026-01-31"

# Search events
python3 gcal_builder.py search --query "team meeting"

# Get event details
python3 gcal_builder.py get --event-id "EVENT_ID"
python3 gcal_builder.py get --event-id "EVENT_ID" --show-attachments

# Create event (with Meet by default, timezone detected automatically)
python3 gcal_builder.py create --summary "Team Sync" \
  --start "2026-01-28T10:00:00" --end "2026-01-28T11:00:00" \
  --attendees "colleague@example.com,colleague2@example.com" \
  --description "Weekly sync meeting"

# Create without Meet
python3 gcal_builder.py create --summary "Lunch" \
  --start "2026-01-28T12:00:00" --end "2026-01-28T13:00:00" --no-meet

# Create recurring event
python3 gcal_builder.py create --summary "Daily Standup" \
  --start "2026-01-28T09:00:00" --end "2026-01-28T09:15:00" \
  --recurrence "WEEKLY" --days "MO,TU,WE,TH,FR" \
  --attendees "team@example.com"

# Update event
python3 gcal_builder.py update --event-id "EVENT_ID" \
  --summary "New Title" --description "New description"

# Add/remove attendees
python3 gcal_builder.py add-attendees --event-id "EVENT_ID" \
  --attendees "new@example.com"
python3 gcal_builder.py remove-attendees --event-id "EVENT_ID" \
  --attendees "remove@example.com"

# Change recurrence
python3 gcal_builder.py set-recurrence --event-id "EVENT_ID" \
  --recurrence "WEEKLY" --interval 2

# Attach document
python3 gcal_builder.py attach --event-id "EVENT_ID" \
  --doc-id "DOC_ID" --title "Meeting Notes"

# Delete event
python3 gcal_builder.py delete --event-id "EVENT_ID"

# Find available meeting times
python3 gcal_builder.py find-availability \
  --attendees "person1@example.com,person2@example.com" \
  --start "2025-01-20T00:00:00" --end "2025-01-24T23:59:59" \
  --duration 30

# Raw free/busy query
python3 gcal_builder.py freebusy \
  --attendees "person1@example.com,person2@example.com" \
  --start "2025-01-20T00:00:00" --end "2025-01-24T23:59:59"
```

### gcal_auth.py - Authentication Management

```bash
python3 gcal_auth.py status    # Check auth status
python3 gcal_auth.py login     # Login with required scopes
python3 gcal_auth.py token     # Get access token
python3 gcal_auth.py validate  # Validate current token
```

### Date Validation (CRITICAL - Use Before Scheduling)

```bash
# Validate date range and see days of week
python3 gcal_builder.py validate-dates --start "2025-01-27" --end "2025-01-31"

# Example output shows each day with its day of week:
# {
#   "days": [
#     {"date": "2025-01-27", "day_of_week": "Monday"},
#     {"date": "2025-01-28", "day_of_week": "Tuesday"},
#     ...
#   ],
#   "warnings": ["Date range includes weekend days: Saturday 02/01"]
# }
```

**Always validate dates before scheduling** to avoid errors like scheduling meetings on weekends or using incorrect date ranges.

## Best Practices

1. **ALWAYS get context first** using `get-context` command at the start of ANY calendar operation - this automatically detects timezone, current date/time/year from the user's calendar
2. **ALWAYS validate date ranges** using `validate-dates` command before scheduling meetings - this prevents day-of-week errors (e.g., thinking Friday is Jan 31 when it's Saturday)
3. **Prefer rich calendar queries over FreeBusy** for same-org colleagues - query their calendar directly using `calendars/{email}/events` to get full event details including names, attendees, and response status. This lets you classify conflicts as soft (office hours, unaccepted invites) vs hard (accepted meetings, external calls). Only fall back to FreeBusy for external contacts.
4. **Availability is automatically checked** - Before creating events, all attendees' availability is checked by default. If conflicts are found, the event is not created. This prevents double-booking (can be disabled with `check_availability=False`)
5. **Timezone is automatic** - All functions now use the user's calendar timezone automatically. You don't need to specify it unless overriding.
6. **Organizer is automatically added as attendee** - When creating events, the organizer is automatically added to the attendee list so they appear as both organizer and attendee (can be disabled with `include_organizer=False`)
7. **Always use `conferenceDataVersion=1`** when creating/updating events with Meet links
8. **Use `sendUpdates=all`** when modifying events with attendees to send notifications
9. **Use `supportsAttachments=true`** when working with file attachments
10. **Use `singleEvents=true`** when listing to expand recurring events into instances
11. **Generate unique `requestId`** for conference creation to avoid duplicates
12. **PATCH instead of PUT** for partial updates to avoid overwriting fields

## Example: Create Complete Meeting with Notes

```bash
#!/bin/bash
TOKEN=$(gcloud auth application-default print-access-token)
QUOTA_PROJECT="${GCP_QUOTA_PROJECT:-your-gcp-project-id}"

# 1. Create meeting notes document
DOC_ID=$(python3 ../google-docs/resources/gdocs_builder.py \
  create --title "Team Sync Notes - $(date +%Y-%m-%d)" | jq -r '.documentId')

echo "Created notes document: $DOC_ID"

# 2. Create calendar event with Meet link and attachment
EVENT=$(curl -s -X POST "https://www.googleapis.com/calendar/v3/calendars/primary/events?conferenceDataVersion=1&supportsAttachments=true" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: $QUOTA_PROJECT" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Team Sync",
    "description": "<b>Agenda:</b>\n<ul>\n<li>Status updates</li>\n<li>Blockers</li>\n<li>Action items</li>\n</ul>",
    "start": {"dateTime": "'$(date -v+1d +%Y-%m-%dT10:00:00)'-08:00", "timeZone": "America/Los_Angeles"},
    "end": {"dateTime": "'$(date -v+1d +%Y-%m-%dT11:00:00)'-08:00", "timeZone": "America/Los_Angeles"},
    "attendees": [
      {"email": "colleague@example.com"},
      {"email": "bkvarda@squareup.com"}
    ],
    "conferenceData": {
      "createRequest": {
        "requestId": "sync-'$(date +%s)'",
        "conferenceSolutionKey": {"type": "hangoutsMeet"}
      }
    },
    "attachments": [
      {
        "fileUrl": "https://docs.google.com/document/d/'$DOC_ID'/edit",
        "title": "Meeting Notes"
      }
    ]
  }')

EVENT_ID=$(echo $EVENT | jq -r '.id')
MEET_LINK=$(echo $EVENT | jq -r '.conferenceData.entryPoints[0].uri')

echo "Created event: $EVENT_ID"
echo "Meet link: $MEET_LINK"
echo "Calendar URL: https://calendar.google.com/calendar/event?eid=$(echo -n "${EVENT_ID} primary" | base64)"
```

## Sources

- [Google Calendar API Reference](https://developers.google.com/calendar/api/v3/reference)
- [Events Resource](https://developers.google.com/calendar/api/v3/reference/events)
- [RRULE Specification](https://datatracker.ietf.org/doc/html/rfc5545#section-3.3.10)
- [Conference Data](https://developers.google.com/calendar/api/guides/create-events#conferencing)
