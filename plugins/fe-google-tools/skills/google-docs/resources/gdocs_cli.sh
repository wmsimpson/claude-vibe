#!/bin/bash
#
# Google Docs CLI - Quick command-line access to Google Docs/Drive/Slides APIs
#
# Usage:
#   ./gdocs_cli.sh create-doc "My Document"
#   ./gdocs_cli.sh read-doc DOC_ID
#   ./gdocs_cli.sh list-files
#   ./gdocs_cli.sh create-presentation "My Slides"
#

set -e

GCLOUD_PATH="${GCLOUD_PATH:-$(which gcloud 2>/dev/null || echo "$HOME/google-cloud-sdk/bin/gcloud")}"
# Set GCP_QUOTA_PROJECT env var if Sheets/Slides APIs require a quota project
QUOTA_PROJECT="${GCP_QUOTA_PROJECT:-}"

get_token() {
    $GCLOUD_PATH auth application-default print-access-token
}

# Check if authenticated
check_auth() {
    if ! $GCLOUD_PATH auth application-default print-access-token &>/dev/null; then
        echo "Not authenticated. Run:"
        echo "$GCLOUD_PATH auth application-default login --scopes=\"https://www.googleapis.com/auth/drive,https://www.googleapis.com/auth/cloud-platform,https://www.googleapis.com/auth/documents,https://www.googleapis.com/auth/presentations\""
        exit 1
    fi
}

create_doc() {
    local title="$1"
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://docs.googleapis.com/v1/documents" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"title\": \"$title\"}" | jq '{documentId, title, url: "https://docs.google.com/document/d/\(.documentId)/edit"}'
}

read_doc() {
    local doc_id="$1"
    check_auth
    TOKEN=$(get_token)

    curl -s "https://docs.googleapis.com/v1/documents/$doc_id" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT"
}

read_doc_structure() {
    local doc_id="$1"
    check_auth
    TOKEN=$(get_token)

    curl -s "https://docs.googleapis.com/v1/documents/$doc_id" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | \
        jq '.body.content[] | select(.paragraph) | {startIndex, endIndex, text: .paragraph.elements[0].textRun.content, style: .paragraph.paragraphStyle.namedStyleType}'
}

update_doc() {
    local doc_id="$1"
    local requests="$2"
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://docs.googleapis.com/v1/documents/$doc_id:batchUpdate" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"requests\": $requests}"
}

insert_text() {
    local doc_id="$1"
    local index="$2"
    local text="$3"

    update_doc "$doc_id" "[{\"insertText\": {\"location\": {\"index\": $index}, \"text\": \"$text\"}}]"
}

list_files() {
    local page_size="${1:-10}"
    check_auth
    TOKEN=$(get_token)

    curl -s "https://www.googleapis.com/drive/v3/files?pageSize=$page_size&fields=files(id,name,mimeType,modifiedTime)" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.files'
}

search_files() {
    local query="$1"
    check_auth
    TOKEN=$(get_token)

    # URL encode the query
    local encoded_query=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$query'))")

    curl -s "https://www.googleapis.com/drive/v3/files?q=name%20contains%20%27$encoded_query%27&fields=files(id,name,mimeType)" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.files'
}

create_folder() {
    local name="$1"
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://www.googleapis.com/drive/v3/files" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"$name\", \"mimeType\": \"application/vnd.google-apps.folder\"}" | jq '{id, name}'
}

share_file() {
    local file_id="$1"
    local email="$2"
    local role="${3:-writer}"  # reader, commenter, writer
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://www.googleapis.com/drive/v3/files/$file_id/permissions" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"type\": \"user\", \"role\": \"$role\", \"emailAddress\": \"$email\"}"
}

create_presentation() {
    local title="$1"
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://slides.googleapis.com/v1/presentations" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"title\": \"$title\"}" | jq '{presentationId, title, url: "https://docs.google.com/presentation/d/\(.presentationId)/edit"}'
}

add_slide() {
    local presentation_id="$1"
    local layout="${2:-TITLE_AND_BODY}"  # BLANK, TITLE, TITLE_AND_BODY, TITLE_AND_TWO_COLUMNS, etc.
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://slides.googleapis.com/v1/presentations/$presentation_id:batchUpdate" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"requests\": [{\"createSlide\": {\"slideLayoutReference\": {\"predefinedLayout\": \"$layout\"}}}]}"
}

# Main command dispatcher
case "$1" in
    create-doc)
        create_doc "$2"
        ;;
    read-doc)
        read_doc "$2"
        ;;
    read-structure)
        read_doc_structure "$2"
        ;;
    update-doc)
        update_doc "$2" "$3"
        ;;
    insert-text)
        insert_text "$2" "$3" "$4"
        ;;
    list-files)
        list_files "$2"
        ;;
    search-files)
        search_files "$2"
        ;;
    create-folder)
        create_folder "$2"
        ;;
    share)
        share_file "$2" "$3" "$4"
        ;;
    create-presentation)
        create_presentation "$2"
        ;;
    add-slide)
        add_slide "$2" "$3"
        ;;
    auth-status)
        check_auth
        echo "Authenticated successfully"
        ;;
    help|--help|-h)
        cat << 'EOF'
Google Docs CLI - Quick command-line access to Google APIs

COMMANDS:
  Document Operations:
    create-doc TITLE              Create a new Google Doc
    read-doc DOC_ID               Read full document JSON
    read-structure DOC_ID         Read document structure with indices
    update-doc DOC_ID REQUESTS    Execute batchUpdate with JSON requests
    insert-text DOC_ID INDEX TEXT Insert text at index

  Drive Operations:
    list-files [PAGE_SIZE]        List recent files (default: 10)
    search-files QUERY            Search files by name
    create-folder NAME            Create a new folder
    share FILE_ID EMAIL [ROLE]    Share file (role: reader/commenter/writer)

  Slides Operations:
    create-presentation TITLE     Create a new presentation
    add-slide PRES_ID [LAYOUT]    Add slide (layout: BLANK/TITLE/TITLE_AND_BODY)

  Utility:
    auth-status                   Check authentication status
    help                          Show this help

EXAMPLES:
  ./gdocs_cli.sh create-doc "Meeting Notes"
  ./gdocs_cli.sh list-files 5
  ./gdocs_cli.sh share abc123 user@example.com writer
  ./gdocs_cli.sh read-structure 1asOEQDp0biMqrd6NGMc7J-DGLBSTUV
EOF
        ;;
    *)
        echo "Unknown command: $1"
        echo "Run './gdocs_cli.sh help' for usage"
        exit 1
        ;;
esac
