#!/bin/bash
# Google Calendar CLI Helper
# Quick commands for Calendar API operations

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GCLOUD_PATH="${GCLOUD:-$(which gcloud 2>/dev/null || echo "$HOME/google-cloud-sdk/bin/gcloud")}"
QUOTA_PROJECT="${GCP_QUOTA_PROJECT:-}"
CALENDAR_API="https://www.googleapis.com/calendar/v3"

# Get access token
get_token() {
    $GCLOUD_PATH auth application-default print-access-token 2>/dev/null
}

# Check authentication status
auth_status() {
    python3 "$SCRIPT_DIR/gcal_auth.py" status
}

# List upcoming events
list_events() {
    local max_results="${1:-10}"
    local time_min=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    TOKEN=$(get_token)
    curl -s "$CALENDAR_API/calendars/primary/events?maxResults=$max_results&singleEvents=true&orderBy=startTime&timeMin=$time_min" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | \
        jq '.items[] | {id, summary, start: .start.dateTime, end: .end.dateTime, meetLink: .conferenceData.entryPoints[0].uri}'
}

# Search events
search_events() {
    local query="$1"
    local max_results="${2:-10}"
    local time_min=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    TOKEN=$(get_token)
    local encoded_query=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$query'))")
    curl -s "$CALENDAR_API/calendars/primary/events?q=$encoded_query&maxResults=$max_results&singleEvents=true&orderBy=startTime&timeMin=$time_min" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | \
        jq '.items[] | {id, summary, start: .start.dateTime, attendees: [.attendees[]?.email]}'
}

# Get event by ID
get_event() {
    local event_id="$1"
    TOKEN=$(get_token)
    curl -s "$CALENDAR_API/calendars/primary/events/$event_id" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.'
}

# Quick create event with Meet
quick_create() {
    local summary="$1"
    local start="$2"
    local end="$3"
    local attendees="$4"

    TOKEN=$(get_token)
    local request_id="meet-$(date +%s)"

    local attendees_json=""
    if [ -n "$attendees" ]; then
        attendees_json=$(echo "$attendees" | tr ',' '\n' | jq -R '{email: .}' | jq -s '.')
    else
        attendees_json="[]"
    fi

    curl -s -X POST "$CALENDAR_API/calendars/primary/events?conferenceDataVersion=1" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{
            \"summary\": \"$summary\",
            \"start\": {\"dateTime\": \"$start\", \"timeZone\": \"America/Los_Angeles\"},
            \"end\": {\"dateTime\": \"$end\", \"timeZone\": \"America/Los_Angeles\"},
            \"attendees\": $attendees_json,
            \"conferenceData\": {
                \"createRequest\": {
                    \"requestId\": \"$request_id\",
                    \"conferenceSolutionKey\": {\"type\": \"hangoutsMeet\"}
                }
            }
        }" | jq '{id, summary, htmlLink, meetLink: .conferenceData.entryPoints[0].uri}'
}

# Delete event
delete_event() {
    local event_id="$1"
    TOKEN=$(get_token)
    curl -s -X DELETE "$CALENDAR_API/calendars/primary/events/$event_id?sendUpdates=all" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT"
    echo "Deleted event: $event_id"
}

# Today's events
today() {
    local start=$(date +%Y-%m-%dT00:00:00Z)
    local end=$(date +%Y-%m-%dT23:59:59Z)
    TOKEN=$(get_token)
    curl -s "$CALENDAR_API/calendars/primary/events?singleEvents=true&orderBy=startTime&timeMin=$start&timeMax=$end" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | \
        jq '.items[] | {summary, start: .start.dateTime, end: .end.dateTime, meetLink: .conferenceData.entryPoints[0].uri}'
}

# Help
show_help() {
    cat << EOF
Google Calendar CLI Helper

Usage: gcal_cli.sh <command> [args]

Commands:
    auth-status                         Check authentication status
    list [max]                          List upcoming events (default: 10)
    search <query> [max]                Search events
    get <event_id>                      Get event details
    today                               Show today's events
    quick <summary> <start> <end> [attendees]
                                        Quick create event with Meet link
                                        Times: 2025-01-15T10:00:00-08:00
                                        Attendees: comma-separated emails
    delete <event_id>                   Delete event
    help                                Show this help

Examples:
    gcal_cli.sh auth-status
    gcal_cli.sh list 5
    gcal_cli.sh today
    gcal_cli.sh search "team meeting"
    gcal_cli.sh get EVENT_ID
    gcal_cli.sh quick "Team Sync" "2025-01-15T10:00:00-08:00" "2025-01-15T11:00:00-08:00" "colleague@example.com"
    gcal_cli.sh delete EVENT_ID
EOF
}

# Main
case "$1" in
    auth-status)
        auth_status
        ;;
    list)
        list_events "$2"
        ;;
    search)
        search_events "$2" "$3"
        ;;
    get)
        get_event "$2"
        ;;
    today)
        today
        ;;
    quick)
        quick_create "$2" "$3" "$4" "$5"
        ;;
    delete)
        delete_event "$2"
        ;;
    help|--help|-h|"")
        show_help
        ;;
    *)
        echo "Unknown command: $1"
        echo "Use 'gcal_cli.sh help' for usage"
        exit 1
        ;;
esac
