# Google Tasks

Manage Google Tasks -- add tasks to your todo list, view all tasks, mark tasks as done, create task lists, manage subtasks, and set due dates. Handles any request about Google Tasks, todo lists, or task management.

## How to Invoke

### Slash Command

```
/google-tasks
```

### Example Prompts

```
"Add a task to call John by Feb 15th"
"Show me all my tasks across all lists"
"Mark the 'prepare presentation' task as done"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` first to authenticate with Google Workspace |
| gcloud CLI | Must be installed (`brew install --cask google-cloud-sdk`) |
| Quota Project | Uses your GCP quota project (`$GCP_QUOTA_PROJECT`) for API billing |

## What This Skill Does

1. Lists tasks across all task lists (or from a specific list)
2. Creates tasks with titles, notes, and due dates (converts natural language dates automatically)
3. Organizes tasks with subtask hierarchies and reordering
4. Marks tasks as completed or uncompleted
5. Manages multiple task lists for different projects or categories
6. Clears completed tasks and filters by due date

## Key Resources

| File | Description |
|------|-------------|
| `resources/gtasks_builder.py` | Complete task operations (list-all-tasks, create-task, complete-task, create-tasklist, move-task) |

## Related Skills

- `/google-auth` - Required authentication before using Tasks
- `/google-calendar` - Tasks with due dates appear in Calendar's task panel
