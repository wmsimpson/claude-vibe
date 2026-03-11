#!/bin/bash
# Gmail CLI Helper
# Quick commands for Gmail API operations

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GCLOUD_PATH="${GCLOUD:-$(which gcloud 2>/dev/null || echo "$HOME/google-cloud-sdk/bin/gcloud")}"
QUOTA_PROJECT="${GCP_QUOTA_PROJECT:-}"
GMAIL_API="https://gmail.googleapis.com/gmail/v1/users/me"

# Get access token
get_token() {
    $GCLOUD_PATH auth application-default print-access-token 2>/dev/null
}

# Check authentication status
auth_status() {
    python3 "$SCRIPT_DIR/gmail_auth.py" status
}

# List recent messages
list_messages() {
    local max_results="${1:-10}"
    TOKEN=$(get_token)
    curl -s "$GMAIL_API/messages?maxResults=$max_results" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.'
}

# Search messages
search_messages() {
    local query="$1"
    local max_results="${2:-10}"
    TOKEN=$(get_token)
    local encoded_query=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$query'))")
    curl -s "$GMAIL_API/messages?q=$encoded_query&maxResults=$max_results" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.'
}

# Get message by ID
get_message() {
    local message_id="$1"
    local format="${2:-full}"
    TOKEN=$(get_token)
    curl -s "$GMAIL_API/messages/$message_id?format=$format" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.'
}

# List labels
list_labels() {
    TOKEN=$(get_token)
    curl -s "$GMAIL_API/labels" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.labels[] | {id, name, type}'
}

# List drafts
list_drafts() {
    TOKEN=$(get_token)
    curl -s "$GMAIL_API/drafts" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.'
}

# Trash a message
trash_message() {
    local message_id="$1"
    TOKEN=$(get_token)
    curl -s -X POST "$GMAIL_API/messages/$message_id/trash" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.'
}

# Star a message
star_message() {
    local message_id="$1"
    TOKEN=$(get_token)
    curl -s -X POST "$GMAIL_API/messages/$message_id/modify" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d '{"addLabelIds": ["STARRED"]}' | jq '.'
}

# Mark as read
mark_read() {
    local message_id="$1"
    TOKEN=$(get_token)
    curl -s -X POST "$GMAIL_API/messages/$message_id/modify" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d '{"removeLabelIds": ["UNREAD"]}' | jq '.'
}

# Mark as unread
mark_unread() {
    local message_id="$1"
    TOKEN=$(get_token)
    curl -s -X POST "$GMAIL_API/messages/$message_id/modify" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d '{"addLabelIds": ["UNREAD"]}' | jq '.'
}

# Get unread count
unread_count() {
    TOKEN=$(get_token)
    curl -s "$GMAIL_API/messages?q=is:unread&maxResults=500" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.resultSizeEstimate'
}

# Help
show_help() {
    cat << EOF
Gmail CLI Helper

Usage: gmail_cli.sh <command> [args]

Commands:
    auth-status                Check authentication status
    list [max]                 List recent messages (default: 10)
    search <query> [max]       Search messages
    get <message_id> [format]  Get message (format: full|metadata|minimal|raw)
    labels                     List all labels
    drafts                     List all drafts
    trash <message_id>         Move message to trash
    star <message_id>          Star a message
    read <message_id>          Mark message as read
    unread <message_id>        Mark message as unread
    unread-count               Get unread message count
    help                       Show this help

Examples:
    gmail_cli.sh auth-status
    gmail_cli.sh list 5
    gmail_cli.sh search "from:boss@company.com is:unread"
    gmail_cli.sh get MESSAGE_ID
    gmail_cli.sh trash MESSAGE_ID
EOF
}

# Main
case "$1" in
    auth-status)
        auth_status
        ;;
    list)
        list_messages "$2"
        ;;
    search)
        search_messages "$2" "$3"
        ;;
    get)
        get_message "$2" "$3"
        ;;
    labels)
        list_labels
        ;;
    drafts)
        list_drafts
        ;;
    trash)
        trash_message "$2"
        ;;
    star)
        star_message "$2"
        ;;
    read)
        mark_read "$2"
        ;;
    unread)
        mark_unread "$2"
        ;;
    unread-count)
        unread_count
        ;;
    help|--help|-h|"")
        show_help
        ;;
    *)
        echo "Unknown command: $1"
        echo "Use 'gmail_cli.sh help' for usage"
        exit 1
        ;;
esac
