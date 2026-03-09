#!/usr/bin/env python3
"""
Google Tasks Builder - Complete Task Operations Helper

Provides high-level operations for Google Tasks:
- List and manage task lists
- Create, update, and complete tasks
- Manage subtasks with parent relationships
- Clear completed tasks
- Reorder tasks within a list

Usage:
    python3 gtasks_builder.py list-tasklists
    python3 gtasks_builder.py list-tasks --tasklist @default
    python3 gtasks_builder.py create-task --tasklist @default --title "Buy groceries"
    python3 gtasks_builder.py complete-task --tasklist @default --id TASK_ID
"""

import argparse
import json
import subprocess
import sys
from datetime import datetime, timezone
from typing import Dict, List, Optional

from google_api_utils import get_access_token, api_call_with_retry, QUOTA_PROJECT

TASKS_API_BASE = "https://tasks.googleapis.com/tasks/v1"


def api_request(method: str, endpoint: str, data: Optional[Dict] = None,
                params: Optional[Dict] = None) -> Dict:
    """Make an authenticated API request to Tasks API with retry logic."""
    url = f"{TASKS_API_BASE}/{endpoint}"
    try:
        return api_call_with_retry(method, url, data=data, params=params)
    except RuntimeError:
        return {}


# =============================================================================
# Task List Operations
# =============================================================================

def list_tasklists() -> List[Dict]:
    """List all task lists for the authenticated user."""
    result = api_request("GET", "users/@me/lists")
    items = result.get("items", [])

    return [{
        "id": item.get("id"),
        "title": item.get("title"),
        "updated": item.get("updated"),
        "selfLink": item.get("selfLink")
    } for item in items]


def get_tasklist(tasklist_id: str) -> Dict:
    """Get details of a specific task list."""
    result = api_request("GET", f"users/@me/lists/{tasklist_id}")
    return {
        "id": result.get("id"),
        "title": result.get("title"),
        "updated": result.get("updated"),
        "selfLink": result.get("selfLink")
    }


def create_tasklist(title: str) -> Dict:
    """Create a new task list."""
    data = {"title": title}
    result = api_request("POST", "users/@me/lists", data)
    return {
        "id": result.get("id"),
        "title": result.get("title"),
        "updated": result.get("updated"),
        "selfLink": result.get("selfLink")
    }


def update_tasklist(tasklist_id: str, title: str) -> Dict:
    """Update a task list's title."""
    data = {"title": title}
    result = api_request("PATCH", f"users/@me/lists/{tasklist_id}", data)
    return {
        "id": result.get("id"),
        "title": result.get("title"),
        "updated": result.get("updated")
    }


def delete_tasklist(tasklist_id: str) -> Dict:
    """Delete a task list."""
    token = get_access_token()
    url = f"{TASKS_API_BASE}/users/@me/lists/{tasklist_id}"

    cmd = ["curl", "-s", "-X", "DELETE", url,
           "-H", f"Authorization: Bearer {token}",
           "-H", f"x-goog-user-project: {QUOTA_PROJECT}"]

    subprocess.run(cmd, capture_output=True)
    return {"deleted": True, "tasklist_id": tasklist_id}


# =============================================================================
# Task Operations
# =============================================================================

def list_tasks(tasklist_id: str = "@default", show_completed: bool = False,
               show_hidden: bool = False, max_results: int = 100) -> List[Dict]:
    """
    List tasks in a task list.

    Args:
        tasklist_id: Task list ID (use "@default" for primary list)
        show_completed: Include completed tasks
        show_hidden: Include hidden tasks
        max_results: Maximum number of tasks to return
    """
    params = {
        "maxResults": max_results,
        "showCompleted": str(show_completed).lower(),
        "showHidden": str(show_hidden).lower()
    }

    result = api_request("GET", f"lists/{tasklist_id}/tasks", params=params)
    items = result.get("items", [])

    return [{
        "id": item.get("id"),
        "title": item.get("title"),
        "notes": item.get("notes"),
        "status": item.get("status"),  # "needsAction" or "completed"
        "due": item.get("due"),
        "completed": item.get("completed"),
        "parent": item.get("parent"),
        "position": item.get("position"),
        "updated": item.get("updated"),
        "selfLink": item.get("selfLink")
    } for item in items]


def get_task(tasklist_id: str, task_id: str) -> Dict:
    """Get details of a specific task."""
    result = api_request("GET", f"lists/{tasklist_id}/tasks/{task_id}")
    return {
        "id": result.get("id"),
        "title": result.get("title"),
        "notes": result.get("notes"),
        "status": result.get("status"),
        "due": result.get("due"),
        "completed": result.get("completed"),
        "parent": result.get("parent"),
        "position": result.get("position"),
        "updated": result.get("updated"),
        "links": result.get("links", []),
        "selfLink": result.get("selfLink")
    }


def create_task(tasklist_id: str, title: str, notes: Optional[str] = None,
                due: Optional[str] = None, parent: Optional[str] = None,
                previous: Optional[str] = None) -> Dict:
    """
    Create a new task.

    Args:
        tasklist_id: Task list ID (use "@default" for primary list)
        title: Task title (required)
        notes: Task notes/description
        due: Due date in RFC 3339 format (e.g., "2025-01-15T00:00:00.000Z")
        parent: Parent task ID for creating subtasks
        previous: Previous sibling task ID for positioning
    """
    data = {"title": title}

    if notes:
        data["notes"] = notes

    if due:
        # Ensure due date is in RFC 3339 format
        if "T" not in due:
            due = f"{due}T00:00:00.000Z"
        data["due"] = due

    params = {}
    if parent:
        params["parent"] = parent
    if previous:
        params["previous"] = previous

    result = api_request("POST", f"lists/{tasklist_id}/tasks", data,
                         params if params else None)

    return {
        "id": result.get("id"),
        "title": result.get("title"),
        "notes": result.get("notes"),
        "status": result.get("status"),
        "due": result.get("due"),
        "parent": result.get("parent"),
        "position": result.get("position"),
        "selfLink": result.get("selfLink")
    }


def update_task(tasklist_id: str, task_id: str, title: Optional[str] = None,
                notes: Optional[str] = None, due: Optional[str] = None,
                status: Optional[str] = None) -> Dict:
    """
    Update an existing task.

    Args:
        tasklist_id: Task list ID
        task_id: Task ID to update
        title: New title
        notes: New notes
        due: New due date in RFC 3339 format
        status: New status ("needsAction" or "completed")
    """
    data = {}

    if title:
        data["title"] = title
    if notes is not None:  # Allow empty string to clear notes
        data["notes"] = notes
    if due:
        if "T" not in due:
            due = f"{due}T00:00:00.000Z"
        data["due"] = due
    if status:
        data["status"] = status
        if status == "completed":
            data["completed"] = datetime.now(timezone.utc).isoformat()

    result = api_request("PATCH", f"lists/{tasklist_id}/tasks/{task_id}", data)

    return {
        "id": result.get("id"),
        "title": result.get("title"),
        "notes": result.get("notes"),
        "status": result.get("status"),
        "due": result.get("due"),
        "completed": result.get("completed"),
        "updated": True
    }


def complete_task(tasklist_id: str, task_id: str) -> Dict:
    """Mark a task as completed."""
    return update_task(tasklist_id, task_id, status="completed")


def uncomplete_task(tasklist_id: str, task_id: str) -> Dict:
    """Mark a task as not completed (needs action)."""
    data = {
        "status": "needsAction",
        "completed": None
    }
    result = api_request("PATCH", f"lists/{tasklist_id}/tasks/{task_id}", data)

    return {
        "id": result.get("id"),
        "title": result.get("title"),
        "status": result.get("status"),
        "updated": True
    }


def delete_task(tasklist_id: str, task_id: str) -> Dict:
    """Delete a task."""
    token = get_access_token()
    url = f"{TASKS_API_BASE}/lists/{tasklist_id}/tasks/{task_id}"

    cmd = ["curl", "-s", "-X", "DELETE", url,
           "-H", f"Authorization: Bearer {token}",
           "-H", f"x-goog-user-project: {QUOTA_PROJECT}"]

    subprocess.run(cmd, capture_output=True)
    return {"deleted": True, "task_id": task_id}


def clear_completed(tasklist_id: str) -> Dict:
    """Clear all completed tasks from a task list."""
    token = get_access_token()
    url = f"{TASKS_API_BASE}/lists/{tasklist_id}/clear"

    cmd = ["curl", "-s", "-X", "POST", url,
           "-H", f"Authorization: Bearer {token}",
           "-H", f"x-goog-user-project: {QUOTA_PROJECT}"]

    subprocess.run(cmd, capture_output=True)
    return {"cleared": True, "tasklist_id": tasklist_id}


def list_all_tasks(show_completed: bool = False, show_hidden: bool = False,
                   max_results: int = 100) -> List[Dict]:
    """
    List tasks from ALL task lists.

    Args:
        show_completed: Include completed tasks
        show_hidden: Include hidden tasks
        max_results: Maximum number of tasks per list

    Returns:
        List of tasks with tasklist info included
    """
    tasklists = list_tasklists()
    all_tasks = []

    for tasklist in tasklists:
        tasklist_id = tasklist.get("id")
        tasklist_title = tasklist.get("title")

        tasks = list_tasks(
            tasklist_id,
            show_completed=show_completed,
            show_hidden=show_hidden,
            max_results=max_results
        )

        # Add tasklist info to each task
        for task in tasks:
            task["tasklist_id"] = tasklist_id
            task["tasklist_title"] = tasklist_title
            all_tasks.append(task)

    return all_tasks


def move_task(tasklist_id: str, task_id: str, parent: Optional[str] = None,
              previous: Optional[str] = None) -> Dict:
    """
    Move a task to a different position or make it a subtask.

    Args:
        tasklist_id: Task list ID
        task_id: Task ID to move
        parent: New parent task ID (to make it a subtask), or empty to move to top level
        previous: Previous sibling task ID (for ordering)
    """
    params = {}
    if parent is not None:
        params["parent"] = parent
    if previous is not None:
        params["previous"] = previous

    result = api_request("POST", f"lists/{tasklist_id}/tasks/{task_id}/move",
                         params=params if params else None)

    return {
        "id": result.get("id"),
        "title": result.get("title"),
        "parent": result.get("parent"),
        "position": result.get("position"),
        "moved": True
    }


# =============================================================================
# Main CLI
# =============================================================================

def main():
    parser = argparse.ArgumentParser(
        description="Google Tasks operations helper",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    subparsers = parser.add_subparsers(dest="command")

    # Task List commands
    subparsers.add_parser("list-tasklists", help="List all task lists")

    get_tl_parser = subparsers.add_parser("get-tasklist", help="Get task list details")
    get_tl_parser.add_argument("--id", required=True, help="Task list ID")

    create_tl_parser = subparsers.add_parser("create-tasklist", help="Create a new task list")
    create_tl_parser.add_argument("--title", required=True, help="Task list title")

    update_tl_parser = subparsers.add_parser("update-tasklist", help="Update task list title")
    update_tl_parser.add_argument("--id", required=True, help="Task list ID")
    update_tl_parser.add_argument("--title", required=True, help="New title")

    delete_tl_parser = subparsers.add_parser("delete-tasklist", help="Delete a task list")
    delete_tl_parser.add_argument("--id", required=True, help="Task list ID")

    # Task commands
    list_t_parser = subparsers.add_parser("list-tasks", help="List tasks in a task list")
    list_t_parser.add_argument("--tasklist", default="@default", help="Task list ID (default: @default)")
    list_t_parser.add_argument("--show-completed", action="store_true", help="Include completed tasks")
    list_t_parser.add_argument("--show-hidden", action="store_true", help="Include hidden tasks")
    list_t_parser.add_argument("--max-results", type=int, default=100, help="Maximum results")

    list_all_parser = subparsers.add_parser("list-all-tasks", help="List tasks from ALL task lists")
    list_all_parser.add_argument("--show-completed", action="store_true", help="Include completed tasks")
    list_all_parser.add_argument("--show-hidden", action="store_true", help="Include hidden tasks")
    list_all_parser.add_argument("--max-results", type=int, default=100, help="Maximum results per list")

    get_t_parser = subparsers.add_parser("get-task", help="Get task details")
    get_t_parser.add_argument("--tasklist", default="@default", help="Task list ID")
    get_t_parser.add_argument("--id", required=True, help="Task ID")

    create_t_parser = subparsers.add_parser("create-task", help="Create a new task")
    create_t_parser.add_argument("--tasklist", default="@default", help="Task list ID")
    create_t_parser.add_argument("--title", required=True, help="Task title")
    create_t_parser.add_argument("--notes", help="Task notes/description")
    create_t_parser.add_argument("--due", help="Due date (YYYY-MM-DD or RFC 3339)")
    create_t_parser.add_argument("--parent", help="Parent task ID (for subtasks)")
    create_t_parser.add_argument("--previous", help="Previous sibling task ID")

    update_t_parser = subparsers.add_parser("update-task", help="Update a task")
    update_t_parser.add_argument("--tasklist", default="@default", help="Task list ID")
    update_t_parser.add_argument("--id", required=True, help="Task ID")
    update_t_parser.add_argument("--title", help="New title")
    update_t_parser.add_argument("--notes", help="New notes")
    update_t_parser.add_argument("--due", help="New due date")

    complete_parser = subparsers.add_parser("complete-task", help="Mark task as completed")
    complete_parser.add_argument("--tasklist", default="@default", help="Task list ID")
    complete_parser.add_argument("--id", required=True, help="Task ID")

    uncomplete_parser = subparsers.add_parser("uncomplete-task", help="Mark task as not completed")
    uncomplete_parser.add_argument("--tasklist", default="@default", help="Task list ID")
    uncomplete_parser.add_argument("--id", required=True, help="Task ID")

    delete_t_parser = subparsers.add_parser("delete-task", help="Delete a task")
    delete_t_parser.add_argument("--tasklist", default="@default", help="Task list ID")
    delete_t_parser.add_argument("--id", required=True, help="Task ID")

    clear_parser = subparsers.add_parser("clear-completed", help="Clear all completed tasks")
    clear_parser.add_argument("--tasklist", default="@default", help="Task list ID")

    move_parser = subparsers.add_parser("move-task", help="Move a task")
    move_parser.add_argument("--tasklist", default="@default", help="Task list ID")
    move_parser.add_argument("--id", required=True, help="Task ID")
    move_parser.add_argument("--parent", help="New parent task ID (empty for top level)")
    move_parser.add_argument("--previous", help="Previous sibling task ID")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    result = None

    # Task List commands
    if args.command == "list-tasklists":
        result = list_tasklists()

    elif args.command == "get-tasklist":
        result = get_tasklist(args.id)

    elif args.command == "create-tasklist":
        result = create_tasklist(args.title)

    elif args.command == "update-tasklist":
        result = update_tasklist(args.id, args.title)

    elif args.command == "delete-tasklist":
        result = delete_tasklist(args.id)

    # Task commands
    elif args.command == "list-tasks":
        result = list_tasks(
            args.tasklist,
            show_completed=args.show_completed,
            show_hidden=args.show_hidden,
            max_results=args.max_results
        )

    elif args.command == "list-all-tasks":
        result = list_all_tasks(
            show_completed=args.show_completed,
            show_hidden=args.show_hidden,
            max_results=args.max_results
        )

    elif args.command == "get-task":
        result = get_task(args.tasklist, args.id)

    elif args.command == "create-task":
        result = create_task(
            args.tasklist,
            args.title,
            notes=args.notes,
            due=args.due,
            parent=args.parent,
            previous=args.previous
        )

    elif args.command == "update-task":
        result = update_task(
            args.tasklist,
            args.id,
            title=args.title,
            notes=args.notes,
            due=args.due
        )

    elif args.command == "complete-task":
        result = complete_task(args.tasklist, args.id)

    elif args.command == "uncomplete-task":
        result = uncomplete_task(args.tasklist, args.id)

    elif args.command == "delete-task":
        result = delete_task(args.tasklist, args.id)

    elif args.command == "clear-completed":
        result = clear_completed(args.tasklist)

    elif args.command == "move-task":
        result = move_task(
            args.tasklist,
            args.id,
            parent=args.parent,
            previous=args.previous
        )

    if result is not None:
        print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
