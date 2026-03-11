---
name: google-tasks
description: Use this skill for Google Tasks operations - add tasks to your todo list, show all tasks, list tasks, mark tasks as done or complete, create task lists, manage subtasks, and set due dates. Handles any request about Google Tasks, todo lists, reminders, or task management.
---

# Google Tasks Skill

Manage Google Tasks using gcloud CLI + curl. This skill provides patterns and utilities for creating tasks, managing task lists, organizing subtasks, and tracking completion status.

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

### Task Lists

- `@default` - The user's primary/default task list (use this when no specific list is needed)
- Each task list has a unique ID
- Users can have multiple task lists for different projects or categories

### Task Status

Tasks have two possible status values:
- `needsAction` - Task is pending/incomplete
- `completed` - Task has been marked as done

### Due Dates

Due dates use RFC 3339 format with date-only precision:
- `2025-01-15T00:00:00.000Z` (full format)
- The helper script accepts `2025-01-15` and converts it automatically

**IMPORTANT: Google Tasks API only supports dates, NOT specific times.** Any time component is ignored and normalized to midnight UTC (`T00:00:00.000Z`). If a user requests a task "at 9am", the time will NOT be stored - only the date.

**IMPORTANT: Claude must convert natural language dates to RFC 3339 format before calling the helper script.**

#### Date Conversion Rules

When the user specifies a date in natural language, convert it to `YYYY-MM-DD` format before passing to the `--due` parameter.

**Date Conversion Examples:**
| User Input | Converted Output |
|------------|------------------|
| "Feb 2nd" | `2026-02-02` |
| "February 2" | `2026-02-02` |
| "2/2" | `2026-02-02` |
| "tomorrow" | (calculate tomorrow's date as YYYY-MM-DD) |
| "today" | (calculate today's date as YYYY-MM-DD) |
| "next Monday" | (calculate the date as YYYY-MM-DD) |
| "Jan 15, 2027" | `2027-01-15` |

**Conversion Rules:**
1. If no year is specified, use the current year (or next year if the date has already passed)
2. Use today's date from the system to calculate relative dates like "today", "tomorrow", "next week"
3. If the user specifies a time (e.g., "at 9am"), acknowledge it but inform them that Google Tasks only stores dates, not times

**Example command with converted date:**
```bash
# User says: "Add a task for Feb 2nd"
# Claude converts to YYYY-MM-DD and runs:
python3 resources/gtasks_builder.py create-task \
  --title "User's task title" \
  --due "2026-02-02"
```

**When user specifies a time:**
```
User: "Add a task for tomorrow at 9am to call John"

Claude: I'll create that task for tomorrow (February 2, 2026). Note that Google Tasks
only supports due dates, not specific times, so the 9am time won't be stored - but
I can add it to the task notes if you'd like.
```

### Subtasks

Tasks can have parent-child relationships:
- Use the `parent` field to specify the parent task ID
- Subtasks are indented under their parent in the Tasks UI
- Completing a parent task does NOT automatically complete subtasks

### Task IDs

Each task has a unique ID used for updates, completion, deletion, and moving.

## API Reference

### Base URL

```
https://tasks.googleapis.com/tasks/v1
```

## Task List Operations

### List All Task Lists

```bash
TOKEN=$(python3 ../google-auth/resources/google_auth.py token)

curl -s "https://tasks.googleapis.com/tasks/v1/users/@me/lists" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Create a Task List

```bash
curl -s -X POST "https://tasks.googleapis.com/tasks/v1/users/@me/lists" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Work Projects"
  }'
```

### Delete a Task List

```bash
curl -s -X DELETE "https://tasks.googleapis.com/tasks/v1/users/@me/lists/${TASKLIST_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Using the Helper Script

```bash
# List all task lists
python3 resources/gtasks_builder.py list-tasklists

# Create a new task list
python3 resources/gtasks_builder.py create-tasklist --title "Work Projects"

# Get task list details
python3 resources/gtasks_builder.py get-tasklist --id TASKLIST_ID

# Update task list title
python3 resources/gtasks_builder.py update-tasklist --id TASKLIST_ID --title "New Title"

# Delete a task list
python3 resources/gtasks_builder.py delete-tasklist --id TASKLIST_ID
```

## Task Operations

### List Tasks

```bash
# List tasks in the default list
curl -s "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Include completed tasks
curl -s "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks?showCompleted=true" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Create a Task

```bash
curl -s -X POST "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Buy groceries",
    "notes": "Milk, eggs, bread",
    "due": "2025-01-15T00:00:00.000Z"
  }'
```

### Create a Subtask

```bash
# First, get the parent task ID
# Then create the subtask with parent parameter
curl -s -X POST "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks?parent=${PARENT_TASK_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Buy milk"
  }'
```

### Update a Task

```bash
curl -s -X PATCH "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks/${TASK_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated task title",
    "notes": "Updated notes"
  }'
```

### Complete a Task

```bash
curl -s -X PATCH "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks/${TASK_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed"
  }'
```

### Delete a Task

```bash
curl -s -X DELETE "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks/${TASK_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Clear Completed Tasks

Remove all completed tasks from a task list:

```bash
curl -s -X POST "https://tasks.googleapis.com/tasks/v1/lists/@default/clear" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Move a Task

Reorder a task or change its parent:

```bash
# Move task to be after another task
curl -s -X POST "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks/${TASK_ID}/move?previous=${PREVIOUS_TASK_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Make task a subtask of another task
curl -s -X POST "https://tasks.googleapis.com/tasks/v1/lists/@default/tasks/${TASK_ID}/move?parent=${PARENT_TASK_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Using the Helper Script

```bash
# List tasks from ALL task lists (default behavior)
python3 resources/gtasks_builder.py list-all-tasks

# Include completed tasks from all lists
python3 resources/gtasks_builder.py list-all-tasks --show-completed

# List tasks from a specific list only
python3 resources/gtasks_builder.py list-tasks --tasklist TASKLIST_ID

# List tasks from the default list only
python3 resources/gtasks_builder.py list-tasks

# Get task details
python3 resources/gtasks_builder.py get-task --id TASK_ID

# Create a simple task
python3 resources/gtasks_builder.py create-task --title "Buy groceries"

# Create a task with notes and due date
python3 resources/gtasks_builder.py create-task \
  --title "Prepare presentation" \
  --notes "Include Q4 metrics and roadmap" \
  --due "2025-01-20"

# Create a subtask
python3 resources/gtasks_builder.py create-task \
  --title "Create slides" \
  --parent PARENT_TASK_ID

# Update a task
python3 resources/gtasks_builder.py update-task \
  --id TASK_ID \
  --title "New title" \
  --notes "Updated notes"

# Mark a task as completed
python3 resources/gtasks_builder.py complete-task --id TASK_ID

# Mark a task as not completed
python3 resources/gtasks_builder.py uncomplete-task --id TASK_ID

# Delete a task
python3 resources/gtasks_builder.py delete-task --id TASK_ID

# Clear all completed tasks
python3 resources/gtasks_builder.py clear-completed

# Move a task (change order or parent)
python3 resources/gtasks_builder.py move-task --id TASK_ID --previous OTHER_TASK_ID
python3 resources/gtasks_builder.py move-task --id TASK_ID --parent PARENT_TASK_ID
```

## Working with Multiple Task Lists

Google Tasks supports multiple task lists for organizing different projects or areas. **By default, always query ALL task lists** to ensure you don't miss tasks in other lists.

### List All Task Lists

```bash
# See all available task lists
python3 resources/gtasks_builder.py list-tasklists
```

### List Tasks from ALL Lists (Default)

```bash
# Get tasks from every task list at once (ALWAYS USE THIS)
python3 resources/gtasks_builder.py list-all-tasks

# Include completed tasks from all lists
python3 resources/gtasks_builder.py list-all-tasks --show-completed
```

The output includes `tasklist_id` and `tasklist_title` for each task so you know which list it belongs to.

### List Tasks from a Specific List

```bash
# First, find the task list ID
python3 resources/gtasks_builder.py list-tasklists

# Then query that specific list
python3 resources/gtasks_builder.py list-tasks --tasklist MDI5MTg3ODU4ODIyNjczNzQyNDI6MDow
```

### Filter Tasks by Due Date (All Lists)

To find tasks due on a specific date across all lists:

```bash
# Get all tasks due on a specific date (filter null due dates first)
python3 resources/gtasks_builder.py list-all-tasks | jq '[.[] | select(.due) | select(.due | startswith("2025-01-15"))]'

# Get tasks due today
TODAY=$(date +%Y-%m-%d)
python3 resources/gtasks_builder.py list-all-tasks | jq --arg d "$TODAY" '[.[] | select(.due) | select(.due | startswith($d))]'

# Get tasks due this week
python3 resources/gtasks_builder.py list-all-tasks | jq '[.[] | select(.due) | select(.due >= "2025-01-13" and .due <= "2025-01-19")]'
```

### Create Task in a Specific List

```bash
# Create task in a non-default list
python3 resources/gtasks_builder.py create-task \
  --tasklist MDI5MTg3ODU4ODIyNjczNzQyNDI6MDow \
  --title "Project milestone" \
  --due "2025-01-20"
```

## Working with Subtasks

Subtasks allow you to break down complex tasks into smaller items:

### Creating a Task Hierarchy

```bash
# 1. Create the parent task
PARENT=$(python3 resources/gtasks_builder.py create-task --title "Plan vacation")
PARENT_ID=$(echo $PARENT | jq -r '.id')

# 2. Create subtasks
python3 resources/gtasks_builder.py create-task \
  --title "Book flights" \
  --parent $PARENT_ID

python3 resources/gtasks_builder.py create-task \
  --title "Reserve hotel" \
  --parent $PARENT_ID

python3 resources/gtasks_builder.py create-task \
  --title "Plan activities" \
  --parent $PARENT_ID
```

### Moving a Task to Become a Subtask

```bash
python3 resources/gtasks_builder.py move-task \
  --id TASK_ID \
  --parent NEW_PARENT_ID
```

### Moving a Subtask to Top Level

```bash
# Use empty parent to move to top level
python3 resources/gtasks_builder.py move-task \
  --id SUBTASK_ID \
  --parent ""
```

## Helper Script Reference

### gtasks_builder.py Commands

| Command | Description |
|---------|-------------|
| `list-tasklists` | List all task lists |
| `get-tasklist --id ID` | Get task list details |
| `create-tasklist --title TITLE` | Create a new task list |
| `update-tasklist --id ID --title TITLE` | Update task list title |
| `delete-tasklist --id ID` | Delete a task list |
| `list-all-tasks` | **List tasks from ALL task lists (recommended)** |
| `list-tasks [--tasklist ID]` | List tasks from one list (default: @default) |
| `get-task --id ID` | Get task details |
| `create-task --title TITLE [options]` | Create a task |
| `update-task --id ID [options]` | Update a task |
| `complete-task --id ID` | Mark task as completed |
| `uncomplete-task --id ID` | Mark task as not completed |
| `delete-task --id ID` | Delete a task |
| `clear-completed [--tasklist ID]` | Clear completed tasks |
| `move-task --id ID [--parent ID] [--previous ID]` | Move/reorder a task |

### Common Options

- `--tasklist ID` - Task list ID (defaults to `@default`)
- `--title TEXT` - Task or list title
- `--notes TEXT` - Task notes/description
- `--due DATE` - Due date in YYYY-MM-DD format. **Claude converts natural language dates to this format.** Note: Google Tasks only supports dates, not times.
- `--parent ID` - Parent task ID for subtasks
- `--previous ID` - Previous sibling for ordering

## Troubleshooting

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `401 Unauthorized` | Token expired or invalid | Run `/google-auth` to re-authenticate |
| `403 Forbidden` | Missing Tasks scope | Re-run `/google-auth` to get all scopes |
| `404 Not Found` | Invalid task/list ID | Verify the ID exists with list commands |
| `400 Bad Request` | Invalid date format | Use RFC 3339 format or YYYY-MM-DD |

### Checking Authentication

```bash
# Verify you have Tasks scope
python3 ../google-auth/resources/google_auth.py status

# Look for: https://www.googleapis.com/auth/tasks
```

### Task List Not Found

If `@default` doesn't work:

```bash
# List all task lists to find the correct ID
python3 resources/gtasks_builder.py list-tasklists
```

## Best Practices

1. **Always use `list-all-tasks` by default** - Users often have tasks spread across multiple lists; query all lists to get the complete picture
2. **Create separate lists for projects** - Use task lists to organize by project or category
3. **Use subtasks for complex tasks** - Break down big tasks into manageable pieces
4. **Set due dates** - Makes tasks appear in Google Calendar's task panel
5. **Include notes** - Add context and details that help when returning to a task
6. **Clear completed tasks periodically** - Keep lists clean with `clear-completed`
7. **Use move to reorder** - Prioritize by moving important tasks to the top
8. **Check authentication first** - Always verify `/google-auth` before operations

## Complete Example: Project Task List

```bash
#!/bin/bash

# 1. Create a project task list
TASKLIST=$(python3 resources/gtasks_builder.py create-tasklist --title "Q1 Product Launch")
TASKLIST_ID=$(echo $TASKLIST | jq -r '.id')
echo "Created task list: $TASKLIST_ID"

# 2. Create main tasks
TASK1=$(python3 resources/gtasks_builder.py create-task \
  --tasklist $TASKLIST_ID \
  --title "Finalize product specs" \
  --due "2025-01-15")
TASK1_ID=$(echo $TASK1 | jq -r '.id')

TASK2=$(python3 resources/gtasks_builder.py create-task \
  --tasklist $TASKLIST_ID \
  --title "Create marketing materials" \
  --due "2025-01-20")
TASK2_ID=$(echo $TASK2 | jq -r '.id')

TASK3=$(python3 resources/gtasks_builder.py create-task \
  --tasklist $TASKLIST_ID \
  --title "Launch preparation" \
  --due "2025-01-25")
TASK3_ID=$(echo $TASK3 | jq -r '.id')

# 3. Add subtasks to marketing task
python3 resources/gtasks_builder.py create-task \
  --tasklist $TASKLIST_ID \
  --title "Design landing page" \
  --parent $TASK2_ID

python3 resources/gtasks_builder.py create-task \
  --tasklist $TASKLIST_ID \
  --title "Write blog post" \
  --parent $TASK2_ID

python3 resources/gtasks_builder.py create-task \
  --tasklist $TASKLIST_ID \
  --title "Create social media posts" \
  --parent $TASK2_ID

# 4. List all tasks
echo "Tasks in project:"
python3 resources/gtasks_builder.py list-tasks --tasklist $TASKLIST_ID
```

## Sources

- [Google Tasks API Reference](https://developers.google.com/tasks/reference/rest)
- [Tasks Resource](https://developers.google.com/tasks/reference/rest/v1/tasks)
- [TaskLists Resource](https://developers.google.com/tasks/reference/rest/v1/tasklists)
