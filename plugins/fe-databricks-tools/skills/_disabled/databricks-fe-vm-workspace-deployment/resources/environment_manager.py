#!/usr/bin/env python3
"""
FE Vending Machine Environment Manager

Manages cached environment information in ~/.vibe/fe-vm/ to enable
workspace reuse and minimize browser interactions.

Usage:
    python3 environment_manager.py list
    python3 environment_manager.py find --type serverless --min-days 7
    python3 environment_manager.py get <workspace_name>
    python3 environment_manager.py remove <workspace_name>
    python3 environment_manager.py clear
"""

import os
import sys
import json
import argparse
from pathlib import Path
from datetime import datetime, timezone, timedelta
from typing import Optional, Dict, List, Any


# Base directory for all FE VM data
FEVM_DIR = Path.home() / ".vibe" / "fe-vm"
ENVIRONMENTS_FILE = FEVM_DIR / "environments.json"
SESSION_FILE = FEVM_DIR / "session.json"
DEPLOYMENTS_DIR = FEVM_DIR / "deployments"


def ensure_directories():
    """Ensure all required directories exist."""
    FEVM_DIR.mkdir(parents=True, exist_ok=True)
    DEPLOYMENTS_DIR.mkdir(parents=True, exist_ok=True)


def load_environments() -> Dict[str, Any]:
    """Load cached environments from file."""
    if not ENVIRONMENTS_FILE.exists():
        return {"workspaces": [], "last_updated": None}

    try:
        with open(ENVIRONMENTS_FILE, "r") as f:
            return json.load(f)
    except (json.JSONDecodeError, IOError):
        return {"workspaces": [], "last_updated": None}


def save_environments(data: Dict[str, Any]):
    """Save environments to cache file."""
    ensure_directories()
    data["last_updated"] = datetime.now(timezone.utc).isoformat()
    with open(ENVIRONMENTS_FILE, "w") as f:
        json.dump(data, f, indent=2)


def _calculate_days_remaining(expected_deletion: Optional[str]) -> float:
    """Calculate days remaining from an ISO datetime string."""
    if not expected_deletion:
        return 0.0
    try:
        exp_dt = datetime.fromisoformat(expected_deletion.replace("Z", "+00:00"))
        delta = exp_dt - datetime.now(timezone.utc)
        return max(0.0, delta.total_seconds() / 86400)
    except (ValueError, TypeError):
        return 0.0


def update_environments_from_api(deployments: List[Dict[str, Any]]):
    """
    Update the environments cache from API response.

    Args:
        deployments: List of deployment dicts from /api/deployments (v2 format)
    """
    workspaces = []

    for d in deployments:
        # v2 API provides workspace_url as a top-level field
        workspace_url = d.get("workspace_url")

        # Determine workspace type from template id or name
        template_id = d.get("resource_template_id", "")
        template_name = d.get("resource_template_name", "")
        template_key = (template_id + " " + template_name).lower()
        if "serverless" in template_key:
            workspace_type = "serverless"
        elif "classic" in template_key:
            workspace_type = "classic"
        else:
            workspace_type = "other"

        # Calculate days remaining from expected_deletion ISO string
        expected_deletion = d.get("expected_deletion")
        days_remaining = _calculate_days_remaining(expected_deletion)

        workspace = {
            "deployment_id": d.get("resource_id"),
            "workspace_name": d.get("resource_name"),
            "workspace_url": workspace_url,
            "workspace_id": None,
            "state": d.get("resource_state"),
            "cloud_provider": d.get("resource_cloud_provider"),
            "region": d.get("resource_cloud_regions"),
            "template": template_name,
            "template_id": template_id,
            "workspace_type": workspace_type,
            "created_at": d.get("created_at"),
            "expected_deletion_date": expected_deletion,
            "days_remaining": days_remaining,
            "has_extension": d.get("has_extension", False),
            "extension_status": d.get("extension_status"),
        }
        workspaces.append(workspace)

        # Also save individual deployment file
        if workspace["workspace_name"]:
            deployment_file = DEPLOYMENTS_DIR / f"{workspace['workspace_name']}.json"
            with open(deployment_file, "w") as f:
                json.dump(workspace, f, indent=2)

    save_environments({"workspaces": workspaces})
    return workspaces


def list_environments(show_all: bool = False) -> List[Dict[str, Any]]:
    """
    List all cached environments.

    Args:
        show_all: If True, show all including deleted. Otherwise only Active.

    Returns:
        List of workspace dicts
    """
    data = load_environments()
    workspaces = data.get("workspaces", [])

    if not show_all:
        workspaces = [w for w in workspaces if w.get("state") == "Active"]

    return workspaces


def find_workspace(
    workspace_type: Optional[str] = None,
    min_days_remaining: int = 1,
    cloud_provider: str = "aws",
    region: Optional[str] = None
) -> Optional[Dict[str, Any]]:
    """
    Find a suitable workspace matching criteria.

    Args:
        workspace_type: "serverless" or "classic" (None for any)
        min_days_remaining: Minimum days until expiration
        cloud_provider: Cloud provider (default: aws)
        region: Specific region (None for any)

    Returns:
        Best matching workspace or None
    """
    workspaces = list_environments(show_all=False)

    candidates = []
    for w in workspaces:
        # Filter by state
        if w.get("state") != "Active":
            continue

        # Filter by type
        if workspace_type and w.get("workspace_type") != workspace_type:
            continue

        # Filter by cloud
        if w.get("cloud_provider") != cloud_provider:
            continue

        # Filter by region
        if region and w.get("region") != region:
            continue

        # Filter by remaining days
        days = w.get("days_remaining", 0)
        if days < min_days_remaining:
            continue

        candidates.append(w)

    if not candidates:
        return None

    # Sort by days remaining (prefer workspaces with more time left)
    candidates.sort(key=lambda x: x.get("days_remaining", 0), reverse=True)
    return candidates[0]


def get_workspace(workspace_name: str) -> Optional[Dict[str, Any]]:
    """Get details for a specific workspace by name."""
    workspaces = list_environments(show_all=True)

    for w in workspaces:
        if w.get("workspace_name") == workspace_name:
            return w

    # Also check individual deployment files
    deployment_file = DEPLOYMENTS_DIR / f"{workspace_name}.json"
    if deployment_file.exists():
        with open(deployment_file, "r") as f:
            return json.load(f)

    return None


def remove_workspace(workspace_name: str) -> bool:
    """Remove a workspace from the cache (does not delete from FEVM)."""
    data = load_environments()
    workspaces = data.get("workspaces", [])

    original_count = len(workspaces)
    workspaces = [w for w in workspaces if w.get("workspace_name") != workspace_name]

    if len(workspaces) < original_count:
        data["workspaces"] = workspaces
        save_environments(data)

        # Also remove individual file
        deployment_file = DEPLOYMENTS_DIR / f"{workspace_name}.json"
        if deployment_file.exists():
            deployment_file.unlink()

        return True

    return False


def clear_cache():
    """Clear all cached data."""
    if ENVIRONMENTS_FILE.exists():
        ENVIRONMENTS_FILE.unlink()

    # Clear deployment files
    for f in DEPLOYMENTS_DIR.glob("*.json"):
        f.unlink()

    print("Cache cleared.")


def get_session() -> Optional[Dict[str, Any]]:
    """Get the current session info."""
    if not SESSION_FILE.exists():
        return None

    try:
        with open(SESSION_FILE, "r") as f:
            session = json.load(f)

        # Check if expired
        expires_at = session.get("expires_at")
        if expires_at:
            exp_time = datetime.fromisoformat(expires_at.replace("Z", "+00:00"))
            if exp_time < datetime.now(timezone.utc):
                return None

        return session
    except (json.JSONDecodeError, IOError):
        return None


def save_session(cookie: str, expires_hours: int = 24):
    """Save session cookie with expiry."""
    ensure_directories()

    expires_at = datetime.now(timezone.utc) + timedelta(hours=expires_hours)

    session = {
        "cookie": cookie,
        "created_at": datetime.now(timezone.utc).isoformat(),
        "expires_at": expires_at.isoformat(),
    }

    with open(SESSION_FILE, "w") as f:
        json.dump(session, f, indent=2)


def format_workspace(w: Dict[str, Any]) -> str:
    """Format workspace for display."""
    lines = [
        f"  Name: {w.get('workspace_name', 'N/A')}",
        f"  URL: {w.get('workspace_url', 'N/A')}",
        f"  State: {w.get('state', 'N/A')}",
        f"  Type: {w.get('workspace_type', 'N/A')}",
        f"  Region: {w.get('region', 'N/A')}",
        f"  Days Remaining: {w.get('days_remaining', 0):.1f}",
    ]

    # Format dates (ISO string format from v2 API)
    created = w.get("created_at")
    if created and isinstance(created, str):
        try:
            created_dt = datetime.fromisoformat(created.replace("Z", "+00:00"))
            lines.append(f"  Created: {created_dt.strftime('%Y-%m-%d %H:%M UTC')}")
        except ValueError:
            lines.append(f"  Created: {created}")

    expires = w.get("expected_deletion_date")
    if expires and isinstance(expires, str):
        try:
            expires_dt = datetime.fromisoformat(expires.replace("Z", "+00:00"))
            lines.append(f"  Expires: {expires_dt.strftime('%Y-%m-%d %H:%M UTC')}")
        except ValueError:
            lines.append(f"  Expires: {expires}")

    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(
        description="FE Vending Machine Environment Manager"
    )

    subparsers = parser.add_subparsers(dest="command", help="Commands")

    # list command
    list_parser = subparsers.add_parser("list", help="List cached environments")
    list_parser.add_argument("--all", action="store_true", help="Include deleted workspaces")
    list_parser.add_argument("--json", action="store_true", help="Output as JSON")

    # find command
    find_parser = subparsers.add_parser("find", help="Find suitable workspace")
    find_parser.add_argument("--type", choices=["serverless", "classic"], help="Workspace type")
    find_parser.add_argument("--min-days", type=int, default=1, help="Minimum days remaining")
    find_parser.add_argument("--region", help="Specific region")
    find_parser.add_argument("--cloud", default="aws", help="Cloud provider")
    find_parser.add_argument("--json", action="store_true", help="Output as JSON")

    # get command
    get_parser = subparsers.add_parser("get", help="Get workspace details")
    get_parser.add_argument("workspace_name", help="Workspace name")
    get_parser.add_argument("--json", action="store_true", help="Output as JSON")

    # remove command
    remove_parser = subparsers.add_parser("remove", help="Remove workspace from cache")
    remove_parser.add_argument("workspace_name", help="Workspace name")

    # clear command
    subparsers.add_parser("clear", help="Clear all cached data")

    # session command
    session_parser = subparsers.add_parser("session", help="Check session status")
    session_parser.add_argument("--json", action="store_true", help="Output as JSON")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    ensure_directories()

    if args.command == "list":
        workspaces = list_environments(show_all=args.all)

        if args.json:
            print(json.dumps(workspaces, indent=2))
        elif not workspaces:
            print("No cached environments found.")
            print("\nRun 'python3 fe_vm_client.py refresh-cache' to fetch from FEVM.")
        else:
            data = load_environments()
            last_updated = data.get("last_updated", "Never")
            print(f"Cached Environments (last updated: {last_updated})")
            print("=" * 60)

            for w in workspaces:
                print(f"\n{w.get('workspace_name', 'Unknown')}:")
                print(format_workspace(w))

    elif args.command == "find":
        workspace = find_workspace(
            workspace_type=args.type,
            min_days_remaining=args.min_days,
            cloud_provider=args.cloud,
            region=args.region
        )

        if args.json:
            print(json.dumps(workspace, indent=2) if workspace else "null")
        elif workspace:
            print("Found suitable workspace:")
            print(format_workspace(workspace))
            print(f"\n  CLI Profile: fe-vm-{workspace.get('workspace_name')}")
        else:
            print("No suitable workspace found matching criteria.")
            sys.exit(1)

    elif args.command == "get":
        workspace = get_workspace(args.workspace_name)

        if args.json:
            print(json.dumps(workspace, indent=2) if workspace else "null")
        elif workspace:
            print(f"Workspace: {args.workspace_name}")
            print(format_workspace(workspace))
        else:
            print(f"Workspace '{args.workspace_name}' not found in cache.")
            sys.exit(1)

    elif args.command == "remove":
        if remove_workspace(args.workspace_name):
            print(f"Removed '{args.workspace_name}' from cache.")
        else:
            print(f"Workspace '{args.workspace_name}' not found in cache.")
            sys.exit(1)

    elif args.command == "clear":
        clear_cache()

    elif args.command == "session":
        session = get_session()

        if args.json:
            print(json.dumps(session, indent=2) if session else "null")
        elif session:
            print("Session Status: Valid")
            print(f"  Created: {session.get('created_at', 'N/A')}")
            print(f"  Expires: {session.get('expires_at', 'N/A')}")
        else:
            print("Session Status: No valid session")
            print("\nRun 'python3 fe_vm_client.py refresh-cache' to authenticate.")
            sys.exit(1)


if __name__ == "__main__":
    main()
