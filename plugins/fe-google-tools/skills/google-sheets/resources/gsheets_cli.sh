#!/bin/bash
#
# Google Sheets CLI - Quick command-line access to Google Sheets/Drive APIs
#
# Usage:
#   ./gsheets_cli.sh create-sheet "My Spreadsheet"
#   ./gsheets_cli.sh read-sheet SHEET_ID
#   ./gsheets_cli.sh update-cells SHEET_ID "Sheet1!A1:B2" '[[" Name","Score"],["Alice","95"]]'
#   ./gsheets_cli.sh list-files
#

set -e

GCLOUD_PATH="${GCLOUD:-$(which gcloud 2>/dev/null || echo "$HOME/google-cloud-sdk/bin/gcloud")}"
QUOTA_PROJECT="${GCP_QUOTA_PROJECT:-}"

get_token() {
    $GCLOUD_PATH auth application-default print-access-token
}

# Check if authenticated
check_auth() {
    if ! $GCLOUD_PATH auth application-default print-access-token &>/dev/null; then
        echo "Not authenticated. Run:"
        echo "$GCLOUD_PATH auth application-default login --scopes=\"https://www.googleapis.com/auth/drive,https://www.googleapis.com/auth/cloud-platform,https://www.googleapis.com/auth/spreadsheets\""
        exit 1
    fi
}

create_sheet() {
    local title="$1"
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"properties\": {\"title\": \"$title\"}}" | jq '{spreadsheetId, title: .properties.title, url: "https://docs.google.com/spreadsheets/d/\(.spreadsheetId)/edit"}'
}

read_sheet() {
    local sheet_id="$1"
    check_auth
    TOKEN=$(get_token)

    curl -s "https://sheets.googleapis.com/v4/spreadsheets/$sheet_id" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT"
}

list_sheets() {
    local sheet_id="$1"
    check_auth
    TOKEN=$(get_token)

    curl -s "https://sheets.googleapis.com/v4/spreadsheets/$sheet_id?fields=sheets.properties" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.sheets[] | .properties | {sheetId, title, index}'
}

get_values() {
    local sheet_id="$1"
    local range="$2"
    check_auth
    TOKEN=$(get_token)

    curl -s "https://sheets.googleapis.com/v4/spreadsheets/$sheet_id/values/$range" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" | jq '.values'
}

update_cells() {
    local sheet_id="$1"
    local range="$2"
    local values="$3"
    local value_input_option="${4:-USER_ENTERED}"
    check_auth
    TOKEN=$(get_token)

    curl -s -X PUT "https://sheets.googleapis.com/v4/spreadsheets/$sheet_id/values/$range?valueInputOption=$value_input_option" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"range\": \"$range\", \"values\": $values}"
}

append_rows() {
    local sheet_id="$1"
    local range="$2"
    local values="$3"
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets/$sheet_id/values/$range:append?valueInputOption=USER_ENTERED" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"values\": $values}"
}

clear_range() {
    local sheet_id="$1"
    local range="$2"
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets/$sheet_id/values/$range:clear" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json"
}

batch_update() {
    local sheet_id="$1"
    local requests="$2"
    check_auth
    TOKEN=$(get_token)

    curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets/$sheet_id:batchUpdate" \
        -H "Authorization: Bearer $TOKEN" \
        -H "x-goog-user-project: $QUOTA_PROJECT" \
        -H "Content-Type: application/json" \
        -d "{\"requests\": $requests}"
}

add_sheet_tab() {
    local sheet_id="$1"
    local title="$2"

    local request="[{\"addSheet\": {\"properties\": {\"title\": \"$title\"}}}]"
    batch_update "$sheet_id" "$request"
}

find_replace() {
    local sheet_id="$1"
    local find_text="$2"
    local replace_text="$3"

    local request="[{\"findReplace\": {\"find\": \"$find_text\", \"replacement\": \"$replace_text\", \"allSheets\": true}}]"
    batch_update "$sheet_id" "$request"
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

# Main command dispatcher
case "$1" in
    create-sheet)
        create_sheet "$2"
        ;;
    read-sheet)
        read_sheet "$2"
        ;;
    list-sheets)
        list_sheets "$2"
        ;;
    get-values)
        get_values "$2" "$3"
        ;;
    update-cells)
        update_cells "$2" "$3" "$4"
        ;;
    append-rows)
        append_rows "$2" "$3" "$4"
        ;;
    clear-range)
        clear_range "$2" "$3"
        ;;
    batch-update)
        batch_update "$2" "$3"
        ;;
    add-sheet-tab)
        add_sheet_tab "$2" "$3"
        ;;
    find-replace)
        find_replace "$2" "$3" "$4"
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
    auth-status)
        check_auth
        echo "Authenticated successfully"
        ;;
    help|--help|-h)
        cat << 'EOF'
Google Sheets CLI - Quick command-line access to Google Sheets API

COMMANDS:
  Spreadsheet Operations:
    create-sheet TITLE                      Create a new Google Sheet
    read-sheet SHEET_ID                     Read full spreadsheet JSON
    list-sheets SHEET_ID                    List all sheets in spreadsheet
    get-values SHEET_ID RANGE               Get cell values (e.g., "Sheet1!A1:B10")
    update-cells SHEET_ID RANGE VALUES      Update cells with JSON array
    append-rows SHEET_ID RANGE VALUES       Append rows to end of sheet
    clear-range SHEET_ID RANGE              Clear cell values in range
    batch-update SHEET_ID REQUESTS          Execute batchUpdate with JSON requests
    add-sheet-tab SHEET_ID TITLE            Add a new sheet tab
    find-replace SHEET_ID FIND REPLACE      Find and replace text

  Drive Operations:
    list-files [PAGE_SIZE]                  List recent files (default: 10)
    search-files QUERY                      Search files by name
    create-folder NAME                      Create a new folder
    share FILE_ID EMAIL [ROLE]              Share file (role: reader/commenter/writer)

  Utility:
    auth-status                             Check authentication status
    help                                    Show this help

EXAMPLES:
  # Create a new spreadsheet
  ./gsheets_cli.sh create-sheet "Sales Data 2025"

  # Update cells with data
  ./gsheets_cli.sh update-cells SHEET_ID "Sheet1!A1:B2" '[["Name","Score"],["Alice","95"]]'

  # Append rows
  ./gsheets_cli.sh append-rows SHEET_ID "Sheet1!A:B" '[["Bob","87"],["Charlie","92"]]'

  # Get cell values
  ./gsheets_cli.sh get-values SHEET_ID "Sheet1!A1:B10"

  # Find and replace
  ./gsheets_cli.sh find-replace SHEET_ID "TODO" "DONE"

  # Add a new sheet tab
  ./gsheets_cli.sh add-sheet-tab SHEET_ID "Q2 Data"

  # Share with collaborator
  ./gsheets_cli.sh share SHEET_ID user@example.com writer

  # List recent files
  ./gsheets_cli.sh list-files 5
EOF
        ;;
    *)
        echo "Unknown command: $1"
        echo "Run './gsheets_cli.sh help' for usage"
        exit 1
        ;;
esac
