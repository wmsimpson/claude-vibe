#!/usr/bin/env python3
"""
Google Calendar Builder - Complete Calendar Operations Helper

Provides high-level operations for Google Calendar:
- Create events with Google Meet links
- Manage attendees
- Search and list events
- Attach documents
- Handle recurring events
- Update event details

Usage:
    python3 gcal_builder.py create --summary "Meeting" --start "2025-01-15T10:00:00" --end "2025-01-15T11:00:00"
    python3 gcal_builder.py list --max-results 10
    python3 gcal_builder.py search --query "team meeting"
"""

import argparse
import json
import subprocess
import sys
import time
from datetime import datetime, timedelta, timezone
from typing import Dict, List, Optional
from zoneinfo import ZoneInfo

from google_api_utils import get_access_token, api_call_with_retry, QUOTA_PROJECT

CALENDAR_API_BASE = "https://www.googleapis.com/calendar/v3"
DEFAULT_TIMEZONE = "America/Los_Angeles"

# Cache for calendar timezone to avoid repeated API calls
_calendar_timezone_cache: Optional[str] = None

# Cache for user's email to avoid repeated API calls
_user_email_cache: Optional[str] = None


def api_request(method: str, endpoint: str, data: Optional[Dict] = None,
                params: Optional[Dict] = None) -> Dict:
    """Make an authenticated API request to Calendar API with retry logic."""
    url = f"{CALENDAR_API_BASE}/{endpoint}"
    try:
        return api_call_with_retry(method, url, data=data, params=params)
    except RuntimeError:
        return {}


def get_user_email() -> Optional[str]:
    """
    Get the authenticated user's email address from their calendar settings.

    This is used to automatically add the organizer as an attendee in created events.

    Returns the user's email address or None if it cannot be determined.
    """
    global _user_email_cache

    if _user_email_cache:
        return _user_email_cache

    # Try to get from calendar list (primary calendar has the user's email)
    result = api_request("GET", "users/me/calendarList/primary")

    if result and "id" in result:
        _user_email_cache = result["id"]
        return result["id"]

    return None


def get_calendar_timezone() -> str:
    """
    Get the user's calendar timezone from Google Calendar settings.

    This provides the authoritative timezone that the user has configured
    in their Google Calendar, which is what they expect to see when
    scheduling events.

    Returns the IANA timezone name (e.g., "America/Los_Angeles").
    """
    global _calendar_timezone_cache

    if _calendar_timezone_cache:
        return _calendar_timezone_cache

    result = api_request("GET", "users/me/settings/timezone")

    if result and "value" in result:
        _calendar_timezone_cache = result["value"]
        return result["value"]

    # Fallback to system timezone if API call fails
    try:
        system_tz = datetime.now().astimezone().tzname()
        # Try to convert common abbreviations to IANA names
        tz_map = {
            "PST": "America/Los_Angeles",
            "PDT": "America/Los_Angeles",
            "EST": "America/New_York",
            "EDT": "America/New_York",
            "CST": "America/Chicago",
            "CDT": "America/Chicago",
            "MST": "America/Denver",
            "MDT": "America/Denver",
        }
        return tz_map.get(system_tz, DEFAULT_TIMEZONE)
    except:
        return DEFAULT_TIMEZONE


def get_context() -> Dict:
    """
    Get comprehensive calendar context including timezone, current date/time, and year.

    This function automatically infers context from:
    1. User's Google Calendar timezone setting (primary source)
    2. System timezone (fallback)
    3. Current date/time in that timezone

    Returns all context needed to avoid scheduling errors (wrong year, wrong timezone, etc.)
    """
    # Get timezone from calendar settings
    tz_name = get_calendar_timezone()
    tz = ZoneInfo(tz_name)

    # Get current time in user's timezone
    now = datetime.now(tz)

    # Get upcoming events to show context of what's already scheduled
    upcoming = list_events(max_results=3)

    return {
        "timezone": {
            "name": tz_name,
            "abbreviation": now.strftime("%Z"),
            "offset": now.strftime("%z"),
            "offset_hours": now.strftime("%z")[:3] + ":" + now.strftime("%z")[3:]
        },
        "current": {
            "datetime": now.isoformat(),
            "date": now.strftime("%Y-%m-%d"),
            "year": now.year,
            "month": now.month,
            "day": now.day,
            "day_of_week": now.strftime("%A"),
            "time": now.strftime("%I:%M %p"),
            "time_24h": now.strftime("%H:%M")
        },
        "upcoming_events": [{
            "summary": e.get("summary"),
            "start": e.get("start"),
            "day": datetime.fromisoformat(e["start"].replace("Z", "+00:00")).astimezone(tz).strftime("%A, %B %d, %Y") if e.get("start") else None
        } for e in upcoming[:3]],
        "context_summary": f"Today is {now.strftime('%A, %B %d, %Y')} at {now.strftime('%I:%M %p %Z')} (timezone: {tz_name})"
    }


def parse_datetime(dt_str: str, tz_name: Optional[str] = None) -> str:
    """
    Parse datetime string and return RFC 3339 format with proper timezone.

    Args:
        dt_str: Datetime string (with or without timezone)
        tz_name: Timezone name to use if dt_str has no timezone info.
                 If None, uses the user's calendar timezone.

    Returns:
        RFC 3339 formatted datetime string with timezone
    """
    # If already has timezone, return as is
    if '+' in dt_str or 'Z' in dt_str:
        return dt_str

    # Get timezone to use
    if tz_name is None:
        tz_name = get_calendar_timezone()

    # Parse and add timezone
    try:
        dt = datetime.fromisoformat(dt_str)
        # Add timezone info using ZoneInfo for proper DST handling
        tz = ZoneInfo(tz_name)
        dt_with_tz = dt.replace(tzinfo=tz)
        return dt_with_tz.isoformat()
    except ValueError:
        return dt_str


def build_recurrence_rule(freq: str, interval: int = 1, days: Optional[str] = None,
                          count: Optional[int] = None, until: Optional[str] = None) -> str:
    """Build RRULE string from parameters."""
    parts = [f"RRULE:FREQ={freq.upper()}"]

    if interval > 1:
        parts.append(f"INTERVAL={interval}")

    if days:
        parts.append(f"BYDAY={days.upper()}")

    if count:
        parts.append(f"COUNT={count}")
    elif until:
        parts.append(f"UNTIL={until.replace('-', '').replace(':', '')}Z")

    return ";".join(parts)


def list_events(max_results: int = 10, start_date: Optional[str] = None,
                end_date: Optional[str] = None) -> List[Dict]:
    """List calendar events."""
    params = {
        "maxResults": max_results,
        "singleEvents": "true",
        "orderBy": "startTime"
    }

    if start_date:
        params["timeMin"] = parse_datetime(start_date + "T00:00:00")
    else:
        params["timeMin"] = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    if end_date:
        params["timeMax"] = parse_datetime(end_date + "T23:59:59")

    result = api_request("GET", "calendars/primary/events", params=params)
    events = result.get("items", [])

    return [{
        "id": e.get("id"),
        "summary": e.get("summary", "(No title)"),
        "start": e.get("start", {}).get("dateTime", e.get("start", {}).get("date")),
        "end": e.get("end", {}).get("dateTime", e.get("end", {}).get("date")),
        "attendees": [a.get("email") for a in e.get("attendees", [])],
        "meetLink": e.get("conferenceData", {}).get("entryPoints", [{}])[0].get("uri") if e.get("conferenceData") else None,
        "htmlLink": e.get("htmlLink"),
        "recurrence": e.get("recurrence"),
        "attachments": e.get("attachments", [])
    } for e in events]


def search_events(query: str, max_results: int = 10) -> List[Dict]:
    """Search for events matching query."""
    params = {
        "q": query,
        "maxResults": max_results,
        "singleEvents": "true",
        "orderBy": "startTime",
        "timeMin": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    }

    result = api_request("GET", "calendars/primary/events", params=params)
    events = result.get("items", [])

    return [{
        "id": e.get("id"),
        "summary": e.get("summary", "(No title)"),
        "start": e.get("start", {}).get("dateTime", e.get("start", {}).get("date")),
        "end": e.get("end", {}).get("dateTime", e.get("end", {}).get("date")),
        "description": e.get("description", "")[:100],
        "attendees": [a.get("email") for a in e.get("attendees", [])],
        "htmlLink": e.get("htmlLink")
    } for e in events]


def get_event(event_id: str, show_attachments: bool = False) -> Dict:
    """Get a single event by ID."""
    result = api_request("GET", f"calendars/primary/events/{event_id}")

    event = {
        "id": result.get("id"),
        "summary": result.get("summary"),
        "description": result.get("description"),
        "start": result.get("start"),
        "end": result.get("end"),
        "location": result.get("location"),
        "attendees": result.get("attendees", []),
        "recurrence": result.get("recurrence"),
        "meetLink": None,
        "htmlLink": result.get("htmlLink"),
        "attachments": result.get("attachments", [])
    }

    if result.get("conferenceData"):
        entry_points = result["conferenceData"].get("entryPoints", [])
        for ep in entry_points:
            if ep.get("entryPointType") == "video":
                event["meetLink"] = ep.get("uri")
                break

    return event


def create_event(summary: str, start: str, end: str,
                 description: Optional[str] = None,
                 attendees: Optional[List[str]] = None,
                 location: Optional[str] = None,
                 add_meet: bool = True,
                 recurrence: Optional[str] = None,
                 interval: int = 1,
                 days: Optional[str] = None,
                 count: Optional[int] = None,
                 attach_doc_id: Optional[str] = None,
                 attach_title: Optional[str] = None,
                 timezone: Optional[str] = None,
                 include_organizer: bool = True,
                 check_availability: bool = True,
                 check_all_attendees: bool = True) -> Dict:
    """
    Create a calendar event.

    Args:
        timezone: Optional timezone to use. If None, uses user's calendar timezone.
        include_organizer: If True (default), automatically adds the organizer as an attendee.
                          This ensures the organizer appears in the attendee list, not just as organizer.
        check_availability: If True (default), checks availability before creating event.
                           If conflicts are found, returns error without creating event.
        check_all_attendees: If True (default), checks all attendees. If False, only checks organizer.

    Returns:
        Dict with event details or error if conflicts found.
        If conflicts found:
        {
            "error": "Availability conflicts found",
            "availability": {...},  # Full availability check results
            "conflicts": [...]      # List of conflicts
        }
    """
    # Get timezone from calendar if not specified
    if timezone is None:
        timezone = get_calendar_timezone()

    # Check availability before creating event
    if check_availability:
        # Determine which attendees to check
        attendees_to_check = []
        if attendees:
            attendees_to_check = [email.strip() for email in attendees]

        # Add organizer if include_organizer is True
        if include_organizer:
            organizer_email = get_user_email()
            if organizer_email:
                attendees_to_check.insert(0, organizer_email)

        # Perform availability check
        availability = check_availability(
            attendees_to_check,
            start,
            end,
            timezone,
            check_all=check_all_attendees
        )

        # If conflicts found, return error
        if not availability["available"]:
            return {
                "error": "Availability conflicts found",
                "message": availability["message"],
                "availability": availability,
                "conflicts": availability["conflicts"],
                "proposed_time": {
                    "start": start,
                    "end": end,
                    "timezone": timezone
                }
            }

    event_data = {
        "summary": summary,
        "start": {
            "dateTime": parse_datetime(start, timezone),
            "timeZone": timezone
        },
        "end": {
            "dateTime": parse_datetime(end, timezone),
            "timeZone": timezone
        }
    }

    if description:
        event_data["description"] = description

    if location:
        event_data["location"] = location

    if attendees:
        attendee_list = [{"email": email.strip()} for email in attendees]

        # Add organizer as attendee if requested and not already in list
        if include_organizer:
            organizer_email = get_user_email()
            if organizer_email:
                # Check if organizer is already in attendee list
                attendee_emails = {a["email"].lower() for a in attendee_list}
                if organizer_email.lower() not in attendee_emails:
                    # Add organizer at the beginning of the list
                    attendee_list.insert(0, {"email": organizer_email})

        event_data["attendees"] = attendee_list

    if add_meet:
        event_data["conferenceData"] = {
            "createRequest": {
                "requestId": f"meet-{int(time.time())}",
                "conferenceSolutionKey": {"type": "hangoutsMeet"}
            }
        }

    if recurrence:
        rrule = build_recurrence_rule(recurrence, interval, days, count)
        event_data["recurrence"] = [rrule]

    if attach_doc_id:
        event_data["attachments"] = [{
            "fileUrl": f"https://docs.google.com/document/d/{attach_doc_id}/edit",
            "title": attach_title or "Document"
        }]

    params = {}
    if add_meet:
        params["conferenceDataVersion"] = "1"
    if attach_doc_id:
        params["supportsAttachments"] = "true"

    result = api_request("POST", "calendars/primary/events", event_data, params)

    return {
        "id": result.get("id"),
        "summary": result.get("summary"),
        "htmlLink": result.get("htmlLink"),
        "meetLink": result.get("conferenceData", {}).get("entryPoints", [{}])[0].get("uri") if result.get("conferenceData") else None,
        "start": result.get("start"),
        "end": result.get("end")
    }


def update_event(event_id: str, summary: Optional[str] = None,
                 description: Optional[str] = None,
                 start: Optional[str] = None,
                 end: Optional[str] = None,
                 location: Optional[str] = None,
                 send_updates: str = "all",
                 timezone: Optional[str] = None) -> Dict:
    """
    Update an existing event.

    Args:
        timezone: Optional timezone to use. If None, uses user's calendar timezone.
    """
    # Get timezone from calendar if not specified
    if timezone is None:
        timezone = get_calendar_timezone()

    update_data = {}

    if summary:
        update_data["summary"] = summary
    if description:
        update_data["description"] = description
    if location:
        update_data["location"] = location
    if start:
        update_data["start"] = {
            "dateTime": parse_datetime(start, timezone),
            "timeZone": timezone
        }
    if end:
        update_data["end"] = {
            "dateTime": parse_datetime(end, timezone),
            "timeZone": timezone
        }

    params = {"sendUpdates": send_updates}
    result = api_request("PATCH", f"calendars/primary/events/{event_id}",
                        update_data, params)

    return {
        "id": result.get("id"),
        "summary": result.get("summary"),
        "htmlLink": result.get("htmlLink"),
        "updated": True
    }


def add_attendees(event_id: str, attendees: List[str], send_updates: str = "all") -> Dict:
    """Add attendees to an existing event."""
    # Get current event
    current = api_request("GET", f"calendars/primary/events/{event_id}")
    current_attendees = current.get("attendees", [])

    # Add new attendees
    existing_emails = {a.get("email") for a in current_attendees}
    for email in attendees:
        email = email.strip()
        if email not in existing_emails:
            current_attendees.append({"email": email})

    # Update event
    params = {"sendUpdates": send_updates}
    result = api_request("PATCH", f"calendars/primary/events/{event_id}",
                        {"attendees": current_attendees}, params)

    return {
        "id": result.get("id"),
        "attendees": [a.get("email") for a in result.get("attendees", [])],
        "updated": True
    }


def remove_attendees(event_id: str, attendees: List[str], send_updates: str = "all") -> Dict:
    """Remove attendees from an existing event."""
    # Get current event
    current = api_request("GET", f"calendars/primary/events/{event_id}")
    current_attendees = current.get("attendees", [])

    # Filter out removed attendees
    remove_set = {email.strip().lower() for email in attendees}
    new_attendees = [a for a in current_attendees
                     if a.get("email", "").lower() not in remove_set]

    # Update event
    params = {"sendUpdates": send_updates}
    result = api_request("PATCH", f"calendars/primary/events/{event_id}",
                        {"attendees": new_attendees}, params)

    return {
        "id": result.get("id"),
        "attendees": [a.get("email") for a in result.get("attendees", [])],
        "updated": True
    }


def set_recurrence(event_id: str, recurrence: Optional[str] = None,
                   interval: int = 1, days: Optional[str] = None,
                   count: Optional[int] = None, remove: bool = False) -> Dict:
    """Set or remove recurrence on an event."""
    if remove:
        update_data = {"recurrence": None}
    else:
        rrule = build_recurrence_rule(recurrence, interval, days, count)
        update_data = {"recurrence": [rrule]}

    result = api_request("PATCH", f"calendars/primary/events/{event_id}", update_data)

    return {
        "id": result.get("id"),
        "recurrence": result.get("recurrence"),
        "updated": True
    }


def attach_document(event_id: str, doc_id: str, title: str = "Document") -> Dict:
    """Attach a Google Doc to an event."""
    # Get current event to preserve existing attachments
    current = api_request("GET", f"calendars/primary/events/{event_id}")
    attachments = current.get("attachments", [])

    # Add new attachment
    attachments.append({
        "fileUrl": f"https://docs.google.com/document/d/{doc_id}/edit",
        "title": title
    })

    params = {"supportsAttachments": "true"}
    result = api_request("PATCH", f"calendars/primary/events/{event_id}",
                        {"attachments": attachments}, params)

    return {
        "id": result.get("id"),
        "attachments": result.get("attachments", []),
        "updated": True
    }


def delete_event(event_id: str, send_updates: str = "all") -> Dict:
    """Delete an event."""
    token = get_access_token()
    url = f"{CALENDAR_API_BASE}/calendars/primary/events/{event_id}?sendUpdates={send_updates}"

    cmd = ["curl", "-s", "-X", "DELETE", url,
           "-H", f"Authorization: Bearer {token}",
           "-H", f"x-goog-user-project: {QUOTA_PROJECT}"]

    subprocess.run(cmd, capture_output=True)
    return {"deleted": True, "event_id": event_id}


def query_freebusy(emails: List[str], time_min: str, time_max: str,
                   timezone: Optional[str] = None) -> Dict:
    """
    Query free/busy information for a list of calendars.

    Args:
        timezone: Optional timezone to use. If None, uses user's calendar timezone.
    """
    if timezone is None:
        timezone = get_calendar_timezone()

    data = {
        "timeMin": parse_datetime(time_min, timezone),
        "timeMax": parse_datetime(time_max, timezone),
        "timeZone": timezone,
        "items": [{"id": email.strip()} for email in emails]
    }

    token = get_access_token()
    url = f"{CALENDAR_API_BASE}/freeBusy"

    cmd = ["curl", "-s", "-X", "POST", url,
           "-H", f"Authorization: Bearer {token}",
           "-H", f"x-goog-user-project: {QUOTA_PROJECT}",
           "-H", "Content-Type: application/json",
           "-d", json.dumps(data)]

    result = subprocess.run(cmd, capture_output=True, text=True)

    try:
        return json.loads(result.stdout) if result.stdout else {}
    except json.JSONDecodeError:
        print(f"Failed to parse response: {result.stdout}", file=sys.stderr)
        return {}


def analyze_conflict_reschedulability(email: str, busy_start: str, busy_end: str,
                                      timezone: Optional[str] = None) -> Dict:
    """
    Analyze a conflicting event to determine if it can be rescheduled.

    Checks:
    1. Is it an internal-only meeting (no external attendees)?
    2. Is it a low-priority meeting (INVESTech, internal team calls, 1:1s)?
    3. Is the user the organizer (can move it)?

    Args:
        email: Email of the attendee with the conflict
        busy_start: Start time of the conflicting event
        busy_end: End time of the conflicting event
        timezone: Timezone to use

    Returns:
        Dict with reschedulability information:
        {
            "can_reschedule": bool,
            "reason": str,
            "event": {...},  # Full event details
            "is_organizer": bool,
            "has_external_attendees": bool,
            "meeting_type": str  # "investech", "1:1", "team-call", "external", "other"
        }
    """
    if timezone is None:
        timezone = get_calendar_timezone()

    organizer_email = get_user_email()

    # Get events in the busy period to find details
    events = list_events(max_results=50)

    # Find the event that matches this busy period
    matching_event = None
    for event in events:
        event_start = event.get("start")
        if event_start == busy_start:
            # Get full event details
            event_id = event.get("id")
            if event_id:
                matching_event = get_event(event_id)
                break

    if not matching_event:
        return {
            "can_reschedule": False,
            "reason": "Could not find event details",
            "event": None,
            "is_organizer": False,
            "has_external_attendees": False,
            "meeting_type": "unknown"
        }

    # Determine if user is organizer
    organizer = matching_event.get("organizer", {})
    is_organizer = organizer.get("email", "").lower() == (organizer_email or "").lower()

    # Check attendees for external domains
    attendees = matching_event.get("attendees", [])
    internal_domain = os.environ.get("INTERNAL_EMAIL_DOMAIN", "")  # Set via INTERNAL_EMAIL_DOMAIN env var
    has_external_attendees = False

    for attendee in attendees:
        attendee_email = attendee.get("email", "")
        if internal_domain not in attendee_email.lower():
            has_external_attendees = True
            break

    # Categorize meeting type based on title
    summary = (matching_event.get("summary") or "").lower()
    meeting_type = "other"

    if "investech" in summary:
        meeting_type = "investech"
    elif "1:1" in summary or "1-1" in summary or "one-on-one" in summary:
        meeting_type = "1:1"
    elif any(term in summary for term in ["team call", "team sync", "team meeting", "standup", "stand-up"]):
        meeting_type = "team-call"
    elif has_external_attendees:
        meeting_type = "external"

    # Determine if can reschedule
    can_reschedule = False
    reason = ""

    if has_external_attendees:
        can_reschedule = False
        reason = "Has external attendees - should not reschedule"
    elif meeting_type in ["investech", "team-call"]:
        can_reschedule = True
        reason = f"Internal {meeting_type} meeting - can potentially reschedule with user approval"
    elif meeting_type == "1:1" and is_organizer:
        can_reschedule = True
        reason = "User is organizer of 1:1 - can move with user approval"
    elif meeting_type == "1:1" and not is_organizer:
        can_reschedule = True
        reason = "Internal 1:1 (not organizer) - can potentially reschedule with user approval"
    else:
        can_reschedule = False
        reason = "Unknown meeting type or external attendees"

    return {
        "can_reschedule": can_reschedule,
        "reason": reason,
        "event": matching_event,
        "is_organizer": is_organizer,
        "has_external_attendees": has_external_attendees,
        "meeting_type": meeting_type,
        "summary": matching_event.get("summary"),
        "event_id": matching_event.get("id")
    }


def check_availability(emails: List[str], start: str, end: str,
                       timezone: Optional[str] = None,
                       check_all: bool = True,
                       analyze_reschedulability: bool = True) -> Dict:
    """
    Check if attendees are available for a proposed meeting time.

    Args:
        emails: List of email addresses to check
        start: Proposed start time (ISO format)
        end: Proposed end time (ISO format)
        timezone: Optional timezone to use. If None, uses user's calendar timezone.
        check_all: If True, check all attendees. If False, only check organizer.

    Returns:
        Dict with availability status and conflicts:
        {
            "available": bool,  # True if no conflicts found
            "conflicts": [      # List of conflicts for each attendee
                {
                    "email": "person@example.com",
                    "busy_periods": [{"start": "...", "end": "..."}]
                }
            ],
            "message": "Human-readable summary"
        }
    """
    if timezone is None:
        timezone = get_calendar_timezone()

    # If check_all is False, only check organizer
    if not check_all:
        organizer_email = get_user_email()
        if organizer_email:
            emails = [organizer_email]
        else:
            return {
                "available": True,
                "conflicts": [],
                "message": "Could not determine organizer email, skipping availability check"
            }

    if not emails:
        return {
            "available": True,
            "conflicts": [],
            "message": "No attendees to check"
        }

    # Query freebusy for the proposed time
    freebusy_result = query_freebusy(emails, start, end, timezone)

    if not freebusy_result or "calendars" not in freebusy_result:
        return {
            "available": False,
            "conflicts": [],
            "message": "Failed to check availability - could not query free/busy information",
            "error": True
        }

    # Parse the start/end times for conflict detection
    start_dt = parse_datetime_with_tz(parse_datetime(start, timezone), timezone)
    end_dt = parse_datetime_with_tz(parse_datetime(end, timezone), timezone)

    conflicts = []
    for email in emails:
        email = email.strip()
        cal_info = freebusy_result.get("calendars", {}).get(email, {})

        # Check for API errors
        if cal_info.get("errors"):
            conflicts.append({
                "email": email,
                "busy_periods": [],
                "error": cal_info["errors"][0].get("reason", "unknown")
            })
            continue

        # Check for busy periods that overlap with proposed time
        busy_periods = cal_info.get("busy", [])
        conflicting_periods = []

        for period in busy_periods:
            busy_start = parse_datetime_with_tz(period["start"], timezone)
            busy_end = parse_datetime_with_tz(period["end"], timezone)

            # Check if there's overlap
            if not (end_dt <= busy_start or start_dt >= busy_end):
                conflict_info = {
                    "start": period["start"],
                    "end": period["end"]
                }

                # Analyze if this conflict can be rescheduled
                if analyze_reschedulability:
                    reschedule_analysis = analyze_conflict_reschedulability(
                        email,
                        period["start"],
                        period["end"],
                        timezone
                    )
                    conflict_info["reschedule_analysis"] = reschedule_analysis

                conflicting_periods.append(conflict_info)

        if conflicting_periods:
            conflicts.append({
                "email": email,
                "busy_periods": conflicting_periods
            })

    # Categorize conflicts
    reschedulable_conflicts = []
    hard_conflicts = []

    for conflict in conflicts:
        for period in conflict["busy_periods"]:
            analysis = period.get("reschedule_analysis")
            if analysis and analysis.get("can_reschedule"):
                reschedulable_conflicts.append({
                    "email": conflict["email"],
                    "period": period,
                    "analysis": analysis
                })
            elif analysis:
                hard_conflicts.append({
                    "email": conflict["email"],
                    "period": period,
                    "analysis": analysis
                })

    # Build summary message
    if not conflicts:
        message = f"All {len(emails)} attendee(s) are available"
    else:
        conflict_count = len(conflicts)
        total_count = len(emails)
        if conflict_count == total_count:
            message = f"All {total_count} attendee(s) have conflicts"
        else:
            message = f"{conflict_count}/{total_count} attendee(s) have conflicts"

        if reschedulable_conflicts:
            message += f" ({len(reschedulable_conflicts)} potentially reschedulable)"

    return {
        "available": len(conflicts) == 0,
        "conflicts": conflicts,
        "reschedulable_conflicts": reschedulable_conflicts,
        "hard_conflicts": hard_conflicts,
        "message": message,
        "checked_attendees": emails,
        "requires_user_approval": len(reschedulable_conflicts) > 0
    }


def parse_datetime_with_tz(dt_str: str, tz_name: str = DEFAULT_TIMEZONE) -> datetime:
    """
    Parse a datetime string and ensure it has timezone info.

    Handles various formats:
    - ISO format with timezone offset: 2025-01-20T09:00:00-08:00
    - ISO format with Z: 2025-01-20T17:00:00Z
    - ISO format without timezone: 2025-01-20T09:00:00
    """
    # Handle Z suffix (UTC)
    if dt_str.endswith('Z'):
        dt_str = dt_str[:-1] + '+00:00'

    try:
        dt = datetime.fromisoformat(dt_str)
    except ValueError:
        # Try parsing without timezone and add default
        dt = datetime.fromisoformat(dt_str.split('+')[0].split('-08:00')[0])
        dt = dt.replace(tzinfo=ZoneInfo(tz_name))
        return dt

    # If no timezone info, add the default timezone
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=ZoneInfo(tz_name))

    return dt


def validate_date_range(start_date: str, end_date: str) -> Dict:
    """
    Validate a date range and return information about the days included.

    This helps prevent errors where users specify incorrect date ranges
    (e.g., thinking Friday is Jan 31 when it's actually Jan 30).

    Returns a dict with:
    - is_valid: bool
    - days: list of {date, day_of_week, is_weekend}
    - warnings: any issues detected
    """
    try:
        # Parse dates (handle both date-only and datetime formats)
        start_str = start_date.split('T')[0]
        end_str = end_date.split('T')[0]

        start = datetime.strptime(start_str, "%Y-%m-%d")
        end = datetime.strptime(end_str, "%Y-%m-%d")

        if end < start:
            return {
                "is_valid": False,
                "error": "End date is before start date",
                "start": start_str,
                "end": end_str
            }

        days = []
        current = start
        weekend_days = []

        while current <= end:
            day_name = current.strftime("%A")
            is_weekend = current.weekday() >= 5  # Saturday=5, Sunday=6

            day_info = {
                "date": current.strftime("%Y-%m-%d"),
                "day_of_week": day_name,
                "is_weekend": is_weekend
            }
            days.append(day_info)

            if is_weekend:
                weekend_days.append(f"{day_name} {current.strftime('%m/%d')}")

            current += timedelta(days=1)

        warnings = []
        if weekend_days:
            warnings.append(f"Date range includes weekend days: {', '.join(weekend_days)}")

        return {
            "is_valid": True,
            "start": start_str,
            "end": end_str,
            "total_days": len(days),
            "days": days,
            "warnings": warnings if warnings else None
        }

    except ValueError as e:
        return {
            "is_valid": False,
            "error": f"Invalid date format: {str(e)}"
        }


def find_available_slots(emails: List[str], time_min: str, time_max: str,
                         duration_minutes: int = 30,
                         working_hours_start: int = 9,
                         working_hours_end: int = 17,
                         tz_name: Optional[str] = None) -> Dict:
    """
    Find time slots when all (or most) attendees are available.

    Args:
        tz_name: Optional timezone to use. If None, uses user's calendar timezone.

    Returns slots sorted by number of available attendees (most available first).
    """
    # Get timezone from calendar if not specified
    if tz_name is None:
        tz_name = get_calendar_timezone()

    # First, validate the date range and show what days we're searching
    date_validation = validate_date_range(time_min, time_max)

    freebusy = query_freebusy(emails, time_min, time_max, tz_name)

    if not freebusy or "calendars" not in freebusy:
        return {"error": "Failed to query free/busy information", "raw": freebusy}

    # Collect all busy periods per person
    busy_by_person = {}
    errors = {}
    for email in emails:
        email = email.strip()
        cal_info = freebusy.get("calendars", {}).get(email, {})
        if cal_info.get("errors"):
            errors[email] = cal_info["errors"]
            continue
        busy_by_person[email] = cal_info.get("busy", [])

    # Get timezone for consistent handling
    tz = ZoneInfo(tz_name)

    # Parse time range with proper timezone handling
    start_dt = parse_datetime_with_tz(time_min, tz_name)
    end_dt = parse_datetime_with_tz(time_max, tz_name)

    # Adjust to working hours on the start day
    start_dt = start_dt.replace(hour=working_hours_start, minute=0, second=0, microsecond=0)

    slots = []
    current = start_dt
    slot_duration = timedelta(minutes=duration_minutes)

    while current + slot_duration <= end_dt:
        # Skip weekends
        if current.weekday() >= 5:  # Saturday=5, Sunday=6
            # Move to next day at working_hours_start
            current = (current + timedelta(days=1)).replace(
                hour=working_hours_start, minute=0, second=0, microsecond=0
            )
            continue

        # Check if within working hours
        slot_end_hour = (current + slot_duration).hour
        slot_end_minute = (current + slot_duration).minute

        # If slot would end after working hours, move to next day
        if current.hour >= working_hours_end or (slot_end_hour > working_hours_end) or \
           (slot_end_hour == working_hours_end and slot_end_minute > 0):
            # Move to next day at working_hours_start
            current = (current + timedelta(days=1)).replace(
                hour=working_hours_start, minute=0, second=0, microsecond=0
            )
            continue

        if current.hour >= working_hours_start:
            slot_start_str = current.isoformat()
            slot_end_str = (current + slot_duration).isoformat()

            # Format for display: "Mon 1/20 9:00-10:00"
            day_label = current.strftime("%a %-m/%-d")
            time_label = f"{current.strftime('%-I:%M%p').lower()}-{(current + slot_duration).strftime('%-I:%M%p').lower()}"

            # Check availability for each person
            available = []
            busy = []

            for email, busy_periods in busy_by_person.items():
                is_free = True
                for period in busy_periods:
                    busy_start = parse_datetime_with_tz(period["start"], tz_name)
                    busy_end = parse_datetime_with_tz(period["end"], tz_name)

                    # Check for overlap (both datetimes are now timezone-aware)
                    if not (current >= busy_end or current + slot_duration <= busy_start):
                        is_free = False
                        break

                if is_free:
                    available.append(email)
                else:
                    busy.append(email)

            slots.append({
                "start": slot_start_str,
                "end": slot_end_str,
                "day": day_label,
                "time": time_label,
                "available": available,
                "busy": busy,
                "available_count": len(available),
                "total_count": len(busy_by_person)
            })

        # Move to next slot (30-minute increments for scanning)
        current += timedelta(minutes=30)

    # Sort by most available first, then by date
    slots.sort(key=lambda x: (-x["available_count"], x["start"]))

    # Filter to only slots where at least someone is available
    slots = [s for s in slots if s["available_count"] > 0]

    # Group slots by availability count for easier reading
    all_available = [s for s in slots if s["available_count"] == len(busy_by_person)]
    some_available = [s for s in slots if 0 < s["available_count"] < len(busy_by_person)]

    return {
        "query": {
            "attendees": emails,
            "time_range": {"start": time_min, "end": time_max},
            "duration_minutes": duration_minutes,
            "working_hours": f"{working_hours_start}:00-{working_hours_end}:00",
            "timezone": tz_name
        },
        "date_validation": date_validation,
        "all_available": all_available[:10],  # Top 10 slots where everyone is free
        "some_available": some_available[:10],  # Top 10 partial availability
        "summary": {
            "total_attendees": len(busy_by_person),
            "slots_all_available": len(all_available),
            "slots_some_available": len(some_available)
        },
        "errors": errors if errors else None
    }


def main():
    parser = argparse.ArgumentParser(
        description="Google Calendar operations helper",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    subparsers = parser.add_subparsers(dest="command")

    # List events
    list_parser = subparsers.add_parser("list", help="List upcoming events")
    list_parser.add_argument("--max-results", "-n", type=int, default=10)
    list_parser.add_argument("--start", help="Start date (YYYY-MM-DD)")
    list_parser.add_argument("--end", help="End date (YYYY-MM-DD)")

    # Search events
    search_parser = subparsers.add_parser("search", help="Search events")
    search_parser.add_argument("--query", "-q", required=True, help="Search query")
    search_parser.add_argument("--max-results", "-n", type=int, default=10)

    # Get event
    get_parser = subparsers.add_parser("get", help="Get event details")
    get_parser.add_argument("--event-id", required=True, help="Event ID")
    get_parser.add_argument("--show-attachments", action="store_true")

    # Create event
    create_parser = subparsers.add_parser("create", help="Create event")
    create_parser.add_argument("--summary", required=True, help="Event title")
    create_parser.add_argument("--start", required=True, help="Start time (ISO format)")
    create_parser.add_argument("--end", required=True, help="End time (ISO format)")
    create_parser.add_argument("--description", help="Event description")
    create_parser.add_argument("--attendees", help="Comma-separated emails")
    create_parser.add_argument("--location", help="Event location")
    create_parser.add_argument("--no-meet", action="store_true", help="Don't add Meet link")
    create_parser.add_argument("--recurrence", help="Recurrence: DAILY, WEEKLY, MONTHLY, YEARLY")
    create_parser.add_argument("--interval", type=int, default=1, help="Recurrence interval")
    create_parser.add_argument("--days", help="Days for WEEKLY (e.g., MO,WE,FR)")
    create_parser.add_argument("--count", type=int, help="Number of occurrences")
    create_parser.add_argument("--attach-doc", help="Google Doc ID to attach")
    create_parser.add_argument("--attach-title", help="Title for attachment")

    # Update event
    update_parser = subparsers.add_parser("update", help="Update event")
    update_parser.add_argument("--event-id", required=True, help="Event ID")
    update_parser.add_argument("--summary", help="New title")
    update_parser.add_argument("--description", help="New description")
    update_parser.add_argument("--start", help="New start time")
    update_parser.add_argument("--end", help="New end time")
    update_parser.add_argument("--location", help="New location")
    update_parser.add_argument("--no-notify", action="store_true", help="Don't notify attendees")

    # Add attendees
    add_att_parser = subparsers.add_parser("add-attendees", help="Add attendees")
    add_att_parser.add_argument("--event-id", required=True)
    add_att_parser.add_argument("--attendees", required=True, help="Comma-separated emails")
    add_att_parser.add_argument("--no-notify", action="store_true")

    # Remove attendees
    rm_att_parser = subparsers.add_parser("remove-attendees", help="Remove attendees")
    rm_att_parser.add_argument("--event-id", required=True)
    rm_att_parser.add_argument("--attendees", required=True, help="Comma-separated emails")
    rm_att_parser.add_argument("--no-notify", action="store_true")

    # Set recurrence
    recur_parser = subparsers.add_parser("set-recurrence", help="Set/change recurrence")
    recur_parser.add_argument("--event-id", required=True)
    recur_parser.add_argument("--recurrence", help="DAILY, WEEKLY, MONTHLY, YEARLY")
    recur_parser.add_argument("--interval", type=int, default=1)
    recur_parser.add_argument("--days", help="Days for WEEKLY")
    recur_parser.add_argument("--count", type=int)
    recur_parser.add_argument("--remove", action="store_true", help="Remove recurrence")

    # Attach document
    attach_parser = subparsers.add_parser("attach", help="Attach document")
    attach_parser.add_argument("--event-id", required=True)
    attach_parser.add_argument("--doc-id", required=True, help="Google Doc ID")
    attach_parser.add_argument("--title", default="Document", help="Attachment title")

    # Delete event
    delete_parser = subparsers.add_parser("delete", help="Delete event")
    delete_parser.add_argument("--event-id", required=True)
    delete_parser.add_argument("--notify", action="store_true", help="Send cancellation emails")

    # FreeBusy query
    freebusy_parser = subparsers.add_parser("freebusy", help="Query free/busy for calendars")
    freebusy_parser.add_argument("--attendees", required=True, help="Comma-separated emails")
    freebusy_parser.add_argument("--start", required=True, help="Start time (ISO format)")
    freebusy_parser.add_argument("--end", required=True, help="End time (ISO format)")

    # Find availability
    avail_parser = subparsers.add_parser("find-availability", help="Find times when attendees are available")
    avail_parser.add_argument("--attendees", required=True, help="Comma-separated emails")
    avail_parser.add_argument("--start", required=True, help="Start date/time (ISO format)")
    avail_parser.add_argument("--end", required=True, help="End date/time (ISO format)")
    avail_parser.add_argument("--duration", type=int, default=30, help="Meeting duration in minutes (default: 30)")
    avail_parser.add_argument("--working-hours-start", type=int, default=9, help="Working hours start (default: 9)")
    avail_parser.add_argument("--working-hours-end", type=int, default=17, help="Working hours end (default: 17)")

    # Auth status
    subparsers.add_parser("auth-status", help="Check authentication status")

    # Get context (timezone, current date/time, year)
    subparsers.add_parser("get-context", help="Get calendar timezone and current date/time context")

    # Validate dates
    validate_parser = subparsers.add_parser("validate-dates", help="Validate date range and show days of week")
    validate_parser.add_argument("--start", required=True, help="Start date (YYYY-MM-DD or ISO format)")
    validate_parser.add_argument("--end", required=True, help="End date (YYYY-MM-DD or ISO format)")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    result = None

    if args.command == "list":
        result = list_events(args.max_results, args.start, args.end)

    elif args.command == "search":
        result = search_events(args.query, args.max_results)

    elif args.command == "get":
        result = get_event(args.event_id, args.show_attachments)

    elif args.command == "create":
        attendees = args.attendees.split(",") if args.attendees else None
        result = create_event(
            args.summary, args.start, args.end,
            description=args.description,
            attendees=attendees,
            location=args.location,
            add_meet=not args.no_meet,
            recurrence=args.recurrence,
            interval=args.interval,
            days=args.days,
            count=args.count,
            attach_doc_id=args.attach_doc,
            attach_title=args.attach_title
        )

    elif args.command == "update":
        send_updates = "none" if args.no_notify else "all"
        result = update_event(
            args.event_id,
            summary=args.summary,
            description=args.description,
            start=args.start,
            end=args.end,
            location=args.location,
            send_updates=send_updates
        )

    elif args.command == "add-attendees":
        send_updates = "none" if args.no_notify else "all"
        attendees = [e.strip() for e in args.attendees.split(",")]
        result = add_attendees(args.event_id, attendees, send_updates)

    elif args.command == "remove-attendees":
        send_updates = "none" if args.no_notify else "all"
        attendees = [e.strip() for e in args.attendees.split(",")]
        result = remove_attendees(args.event_id, attendees, send_updates)

    elif args.command == "set-recurrence":
        result = set_recurrence(
            args.event_id,
            recurrence=args.recurrence,
            interval=args.interval,
            days=args.days,
            count=args.count,
            remove=args.remove
        )

    elif args.command == "attach":
        result = attach_document(args.event_id, args.doc_id, args.title)

    elif args.command == "delete":
        send_updates = "all" if args.notify else "none"
        result = delete_event(args.event_id, send_updates)

    elif args.command == "freebusy":
        attendees = [e.strip() for e in args.attendees.split(",")]
        result = query_freebusy(attendees, args.start, args.end)

    elif args.command == "find-availability":
        attendees = [e.strip() for e in args.attendees.split(",")]
        result = find_available_slots(
            attendees, args.start, args.end,
            duration_minutes=args.duration,
            working_hours_start=args.working_hours_start,
            working_hours_end=args.working_hours_end
        )

    elif args.command == "auth-status":
        import gcal_auth
        gcal_auth.print_status()
        sys.exit(0)

    elif args.command == "get-context":
        result = get_context()

    elif args.command == "validate-dates":
        result = validate_date_range(args.start, args.end)

    if result is not None:
        print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
