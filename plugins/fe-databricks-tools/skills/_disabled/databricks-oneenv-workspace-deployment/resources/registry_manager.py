#!/usr/bin/env python3
"""
Registry Manager for One-Env Workspace Deployments

Tracks all created resources in ~/.vibe/oneenv/ to enable:
- Resource reuse (metastores, etc.)
- Health verification
- Cleanup operations
- Recovery from partial failures

Usage:
    python registry_manager.py init                              # Initialize registry
    python registry_manager.py add-workspace --name NAME ...     # Add workspace to registry
    python registry_manager.py get-workspace --name NAME         # Get workspace details
    python registry_manager.py list-workspaces                   # List all workspaces
    python registry_manager.py list-metastores                   # List metastores by region
    python registry_manager.py cleanup --workspace NAME          # Remove workspace and resources
    python registry_manager.py export                            # Export registry as JSON
"""

import argparse
import json
import os
import sys
from datetime import datetime
from pathlib import Path
from typing import Optional

REGISTRY_DIR = Path.home() / ".vibe" / "oneenv"
REGISTRY_FILE = REGISTRY_DIR / "registry.json"
WORKSPACES_DIR = REGISTRY_DIR / "workspaces"

# Constants
AWS_SANDBOX_ACCOUNT_ID = "332745928618"
DATABRICKS_ACCOUNT_ID = "0d26daa6-5e44-4c97-a497-ef015f91254a"
DATABRICKS_CONTROL_PLANE_ACCOUNT = "414351767826"


def init_registry() -> dict:
    """Initialize or load the registry."""
    REGISTRY_DIR.mkdir(parents=True, exist_ok=True)
    WORKSPACES_DIR.mkdir(parents=True, exist_ok=True)

    if REGISTRY_FILE.exists():
        with open(REGISTRY_FILE) as f:
            return json.load(f)

    registry = {
        "version": "1.0",
        "aws_account_id": AWS_SANDBOX_ACCOUNT_ID,
        "databricks_account_id": DATABRICKS_ACCOUNT_ID,
        "created_at": datetime.utcnow().isoformat() + "Z",
        "resources": {
            "metastores": {},
            "workspaces": {},
            "iam_roles": {},
            "s3_buckets": {}
        }
    }
    save_registry(registry)
    return registry


def save_registry(registry: dict) -> None:
    """Save the registry to disk."""
    REGISTRY_DIR.mkdir(parents=True, exist_ok=True)
    with open(REGISTRY_FILE, "w") as f:
        json.dump(registry, f, indent=2)


def load_registry() -> dict:
    """Load the registry from disk."""
    if not REGISTRY_FILE.exists():
        return init_registry()
    with open(REGISTRY_FILE) as f:
        return json.load(f)


def add_metastore(
    metastore_id: str,
    name: str,
    region: str,
    storage_root: str,
    owner: Optional[str] = None
) -> None:
    """Add a metastore to the registry."""
    registry = load_registry()

    registry["resources"]["metastores"][region] = {
        "metastore_id": metastore_id,
        "name": name,
        "region": region,
        "storage_root": storage_root,
        "owner": owner,
        "created_at": datetime.utcnow().isoformat() + "Z",
        "created_by_oneenv": True
    }

    save_registry(registry)
    print(f"Added metastore '{name}' ({metastore_id}) for region {region}")


def get_metastore(region: str) -> Optional[dict]:
    """Get metastore for a region if it exists."""
    registry = load_registry()
    return registry["resources"]["metastores"].get(region)


def list_metastores() -> dict:
    """List all registered metastores."""
    registry = load_registry()
    return registry["resources"]["metastores"]


def add_workspace(
    name: str,
    workspace_id: str,
    url: str,
    region: str,
    workspace_type: str,
    metastore_id: Optional[str] = None,
    iam_role_arn: Optional[str] = None,
    s3_bucket: Optional[str] = None,
    vpc_id: Optional[str] = None,
    purpose: Optional[str] = None,
    ip_acls_enabled: bool = True
) -> None:
    """Add a workspace to the registry."""
    registry = load_registry()

    workspace_data = {
        "workspace_id": workspace_id,
        "name": name,
        "url": url,
        "region": region,
        "type": workspace_type,
        "created_at": datetime.utcnow().isoformat() + "Z",
        "last_verified": datetime.utcnow().isoformat() + "Z",
        "last_verified_status": "healthy",
        "ip_acls_enabled": ip_acls_enabled,
        "purpose": purpose,
        "dependencies": {
            "metastore_id": metastore_id,
            "iam_uc_access_role": iam_role_arn,
            "s3_uc_root_bucket": s3_bucket,
            "vpc_id": vpc_id
        }
    }

    registry["resources"]["workspaces"][name] = workspace_data

    # Also save individual workspace file for easy access
    workspace_file = WORKSPACES_DIR / f"{name}.json"
    with open(workspace_file, "w") as f:
        json.dump(workspace_data, f, indent=2)

    # Track IAM role if provided
    if iam_role_arn:
        role_name = iam_role_arn.split("/")[-1]
        registry["resources"]["iam_roles"][role_name] = {
            "arn": iam_role_arn,
            "workspace": name,
            "purpose": "Unity Catalog access",
            "created_at": datetime.utcnow().isoformat() + "Z"
        }

    # Track S3 bucket if provided
    if s3_bucket:
        registry["resources"]["s3_buckets"][s3_bucket] = {
            "name": s3_bucket,
            "region": region,
            "workspace": name,
            "purpose": "Unity Catalog root storage",
            "created_at": datetime.utcnow().isoformat() + "Z"
        }

    save_registry(registry)
    print(f"Added workspace '{name}' ({workspace_id}) to registry")


def get_workspace(name: str) -> Optional[dict]:
    """Get workspace details from registry."""
    registry = load_registry()
    return registry["resources"]["workspaces"].get(name)


def list_workspaces() -> dict:
    """List all registered workspaces."""
    registry = load_registry()
    return registry["resources"]["workspaces"]


def update_verification_status(name: str, status: str) -> None:
    """Update the verification status of a workspace."""
    registry = load_registry()

    if name not in registry["resources"]["workspaces"]:
        print(f"Workspace '{name}' not found in registry", file=sys.stderr)
        return

    registry["resources"]["workspaces"][name]["last_verified"] = datetime.utcnow().isoformat() + "Z"
    registry["resources"]["workspaces"][name]["last_verified_status"] = status

    save_registry(registry)

    # Also update individual workspace file
    workspace_file = WORKSPACES_DIR / f"{name}.json"
    if workspace_file.exists():
        with open(workspace_file) as f:
            workspace_data = json.load(f)
        workspace_data["last_verified"] = registry["resources"]["workspaces"][name]["last_verified"]
        workspace_data["last_verified_status"] = status
        with open(workspace_file, "w") as f:
            json.dump(workspace_data, f, indent=2)


def remove_workspace(name: str, dry_run: bool = False) -> dict:
    """
    Remove a workspace from the registry.
    Returns dict of resources that should be cleaned up in AWS/Databricks.
    """
    registry = load_registry()

    if name not in registry["resources"]["workspaces"]:
        print(f"Workspace '{name}' not found in registry", file=sys.stderr)
        return {}

    workspace = registry["resources"]["workspaces"][name]
    cleanup_resources = {
        "workspace_id": workspace.get("workspace_id"),
        "workspace_name": name,
        "iam_roles": [],
        "s3_buckets": [],
        "vpc_id": workspace.get("dependencies", {}).get("vpc_id")
    }

    # Find associated IAM roles
    for role_name, role_data in list(registry["resources"]["iam_roles"].items()):
        if role_data.get("workspace") == name:
            cleanup_resources["iam_roles"].append(role_name)

    # Find associated S3 buckets
    for bucket_name, bucket_data in list(registry["resources"]["s3_buckets"].items()):
        if bucket_data.get("workspace") == name:
            cleanup_resources["s3_buckets"].append(bucket_name)

    if dry_run:
        print("Dry run - would remove the following:")
        print(json.dumps(cleanup_resources, indent=2))
        return cleanup_resources

    # Remove from registry
    del registry["resources"]["workspaces"][name]

    for role_name in cleanup_resources["iam_roles"]:
        if role_name in registry["resources"]["iam_roles"]:
            del registry["resources"]["iam_roles"][role_name]

    for bucket_name in cleanup_resources["s3_buckets"]:
        if bucket_name in registry["resources"]["s3_buckets"]:
            del registry["resources"]["s3_buckets"][bucket_name]

    save_registry(registry)

    # Remove individual workspace file
    workspace_file = WORKSPACES_DIR / f"{name}.json"
    if workspace_file.exists():
        workspace_file.unlink()

    print(f"Removed workspace '{name}' from registry")
    print("Resources to clean up in AWS/Databricks:")
    print(json.dumps(cleanup_resources, indent=2))

    return cleanup_resources


def export_registry() -> str:
    """Export the full registry as JSON."""
    registry = load_registry()
    return json.dumps(registry, indent=2)


def main():
    parser = argparse.ArgumentParser(description="One-Env Workspace Registry Manager")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # init
    subparsers.add_parser("init", help="Initialize the registry")

    # add-workspace
    add_ws = subparsers.add_parser("add-workspace", help="Add a workspace to registry")
    add_ws.add_argument("--name", required=True, help="Workspace name")
    add_ws.add_argument("--workspace-id", required=True, help="Databricks workspace ID")
    add_ws.add_argument("--url", required=True, help="Workspace URL")
    add_ws.add_argument("--region", required=True, help="AWS region")
    add_ws.add_argument("--type", required=True, choices=["serverless", "classic"], help="Workspace type")
    add_ws.add_argument("--metastore-id", help="Assigned metastore ID")
    add_ws.add_argument("--iam-role-arn", help="IAM role ARN for UC access")
    add_ws.add_argument("--s3-bucket", help="S3 bucket name for UC storage")
    add_ws.add_argument("--vpc-id", help="VPC ID (classic workspaces only)")
    add_ws.add_argument("--purpose", help="Purpose/description of workspace")
    add_ws.add_argument("--ip-acls-disabled", action="store_true", help="IP ACLs are disabled")

    # add-metastore
    add_ms = subparsers.add_parser("add-metastore", help="Add a metastore to registry")
    add_ms.add_argument("--metastore-id", required=True, help="Metastore ID")
    add_ms.add_argument("--name", required=True, help="Metastore name")
    add_ms.add_argument("--region", required=True, help="AWS region")
    add_ms.add_argument("--storage-root", required=True, help="S3 storage root path")
    add_ms.add_argument("--owner", help="Owner email")

    # get-workspace
    get_ws = subparsers.add_parser("get-workspace", help="Get workspace details")
    get_ws.add_argument("--name", required=True, help="Workspace name")

    # get-metastore
    get_ms = subparsers.add_parser("get-metastore", help="Get metastore for region")
    get_ms.add_argument("--region", required=True, help="AWS region")

    # list-workspaces
    subparsers.add_parser("list-workspaces", help="List all workspaces")

    # list-metastores
    subparsers.add_parser("list-metastores", help="List all metastores")

    # cleanup
    cleanup = subparsers.add_parser("cleanup", help="Remove workspace from registry")
    cleanup.add_argument("--workspace", required=True, help="Workspace name to remove")
    cleanup.add_argument("--dry-run", action="store_true", help="Show what would be removed")
    cleanup.add_argument("--confirm", action="store_true", help="Actually remove (required)")

    # update-status
    update = subparsers.add_parser("update-status", help="Update workspace verification status")
    update.add_argument("--name", required=True, help="Workspace name")
    update.add_argument("--status", required=True, choices=["healthy", "broken", "missing"], help="Status")

    # export
    subparsers.add_parser("export", help="Export registry as JSON")

    args = parser.parse_args()

    if args.command == "init":
        init_registry()
        print(f"Registry initialized at {REGISTRY_FILE}")

    elif args.command == "add-workspace":
        add_workspace(
            name=args.name,
            workspace_id=args.workspace_id,
            url=args.url,
            region=args.region,
            workspace_type=args.type,
            metastore_id=args.metastore_id,
            iam_role_arn=args.iam_role_arn,
            s3_bucket=args.s3_bucket,
            vpc_id=args.vpc_id,
            purpose=args.purpose,
            ip_acls_enabled=not args.ip_acls_disabled
        )

    elif args.command == "add-metastore":
        add_metastore(
            metastore_id=args.metastore_id,
            name=args.name,
            region=args.region,
            storage_root=args.storage_root,
            owner=args.owner
        )

    elif args.command == "get-workspace":
        workspace = get_workspace(args.name)
        if workspace:
            print(json.dumps(workspace, indent=2))
        else:
            print(f"Workspace '{args.name}' not found", file=sys.stderr)
            sys.exit(1)

    elif args.command == "get-metastore":
        metastore = get_metastore(args.region)
        if metastore:
            print(json.dumps(metastore, indent=2))
        else:
            print(f"No metastore found for region '{args.region}'", file=sys.stderr)
            sys.exit(1)

    elif args.command == "list-workspaces":
        workspaces = list_workspaces()
        if workspaces:
            print(json.dumps(workspaces, indent=2))
        else:
            print("No workspaces registered")

    elif args.command == "list-metastores":
        metastores = list_metastores()
        if metastores:
            print(json.dumps(metastores, indent=2))
        else:
            print("No metastores registered")

    elif args.command == "cleanup":
        if not args.dry_run and not args.confirm:
            print("Must specify --dry-run or --confirm", file=sys.stderr)
            sys.exit(1)
        remove_workspace(args.workspace, dry_run=args.dry_run)

    elif args.command == "update-status":
        update_verification_status(args.name, args.status)
        print(f"Updated status for '{args.name}' to '{args.status}'")

    elif args.command == "export":
        print(export_registry())

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
