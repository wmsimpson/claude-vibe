#!/usr/bin/env python3
"""
Health Checker for One-Env Workspace Deployments

Verifies that all resources for a workspace actually exist and are functional.
Handles the common case where the 2-week cleanup has deleted some resources.

Usage:
    python health_checker.py check-all --workspace NAME       # Full health check
    python health_checker.py check-workspace --workspace NAME # Check workspace only
    python health_checker.py check-aws --workspace NAME       # Check AWS resources only
    python health_checker.py repair --workspace NAME          # Repair missing resources
"""

import argparse
import json
import subprocess
import sys
from dataclasses import dataclass
from enum import Enum
from pathlib import Path
from typing import Optional

# Import registry manager
sys.path.insert(0, str(Path(__file__).parent))
from registry_manager import load_registry, get_workspace, update_verification_status

# Profiles
AWS_PROFILE = "aws-sandbox-field-eng_databricks-sandbox-admin"
DATABRICKS_ACCOUNT_PROFILE = "one-env-admin-aws"


class ResourceStatus(Enum):
    HEALTHY = "HEALTHY"
    MISSING = "MISSING"
    BROKEN = "BROKEN"
    UNKNOWN = "UNKNOWN"


@dataclass
class HealthCheckResult:
    resource_type: str
    resource_name: str
    status: ResourceStatus
    details: str = ""


def run_command(cmd: list[str], capture_output: bool = True) -> tuple[int, str, str]:
    """Run a command and return (returncode, stdout, stderr)."""
    try:
        result = subprocess.run(
            cmd,
            capture_output=capture_output,
            text=True,
            timeout=60
        )
        return result.returncode, result.stdout, result.stderr
    except subprocess.TimeoutExpired:
        return -1, "", "Command timed out"
    except Exception as e:
        return -1, "", str(e)


def check_databricks_workspace(workspace_id: str) -> HealthCheckResult:
    """Check if workspace exists and is running."""
    cmd = [
        "databricks", "account", "workspaces", "get",
        "--workspace-id", workspace_id,
        "--profile", DATABRICKS_ACCOUNT_PROFILE,
        "--output", "json"
    ]

    returncode, stdout, stderr = run_command(cmd)

    if returncode != 0:
        if "NOT_FOUND" in stderr or "404" in stderr:
            return HealthCheckResult(
                resource_type="Workspace",
                resource_name=workspace_id,
                status=ResourceStatus.MISSING,
                details="Workspace not found in Databricks account"
            )
        return HealthCheckResult(
            resource_type="Workspace",
            resource_name=workspace_id,
            status=ResourceStatus.UNKNOWN,
            details=f"Error checking workspace: {stderr}"
        )

    try:
        data = json.loads(stdout)
        state = data.get("workspace_status", "UNKNOWN")
        if state == "RUNNING":
            return HealthCheckResult(
                resource_type="Workspace",
                resource_name=workspace_id,
                status=ResourceStatus.HEALTHY,
                details=f"Workspace is running at {data.get('deployment_name', 'unknown')}"
            )
        else:
            return HealthCheckResult(
                resource_type="Workspace",
                resource_name=workspace_id,
                status=ResourceStatus.BROKEN,
                details=f"Workspace status is {state}"
            )
    except json.JSONDecodeError:
        return HealthCheckResult(
            resource_type="Workspace",
            resource_name=workspace_id,
            status=ResourceStatus.UNKNOWN,
            details="Could not parse workspace response"
        )


def check_metastore_assignment(workspace_id: str) -> HealthCheckResult:
    """Check if workspace has a metastore assigned."""
    cmd = [
        "databricks", "account", "metastore-assignments", "get",
        "--workspace-id", workspace_id,
        "--profile", DATABRICKS_ACCOUNT_PROFILE,
        "--output", "json"
    ]

    returncode, stdout, stderr = run_command(cmd)

    if returncode != 0:
        if "NOT_FOUND" in stderr or "404" in stderr:
            return HealthCheckResult(
                resource_type="Metastore Assignment",
                resource_name=f"workspace-{workspace_id}",
                status=ResourceStatus.MISSING,
                details="No metastore assigned to workspace"
            )
        return HealthCheckResult(
            resource_type="Metastore Assignment",
            resource_name=f"workspace-{workspace_id}",
            status=ResourceStatus.UNKNOWN,
            details=f"Error checking assignment: {stderr}"
        )

    try:
        data = json.loads(stdout)
        metastore_id = data.get("metastore_id", "unknown")
        return HealthCheckResult(
            resource_type="Metastore Assignment",
            resource_name=f"workspace-{workspace_id}",
            status=ResourceStatus.HEALTHY,
            details=f"Metastore {metastore_id} assigned"
        )
    except json.JSONDecodeError:
        return HealthCheckResult(
            resource_type="Metastore Assignment",
            resource_name=f"workspace-{workspace_id}",
            status=ResourceStatus.UNKNOWN,
            details="Could not parse response"
        )


def check_iam_role(role_name: str) -> HealthCheckResult:
    """Check if IAM role exists."""
    cmd = [
        "aws", "iam", "get-role",
        "--role-name", role_name,
        "--profile", AWS_PROFILE,
        "--output", "json"
    ]

    returncode, stdout, stderr = run_command(cmd)

    if returncode != 0:
        if "NoSuchEntity" in stderr:
            return HealthCheckResult(
                resource_type="IAM Role",
                resource_name=role_name,
                status=ResourceStatus.MISSING,
                details="IAM role not found"
            )
        return HealthCheckResult(
            resource_type="IAM Role",
            resource_name=role_name,
            status=ResourceStatus.UNKNOWN,
            details=f"Error checking role: {stderr}"
        )

    return HealthCheckResult(
        resource_type="IAM Role",
        resource_name=role_name,
        status=ResourceStatus.HEALTHY,
        details="IAM role exists"
    )


def check_s3_bucket(bucket_name: str) -> HealthCheckResult:
    """Check if S3 bucket exists and is accessible."""
    cmd = [
        "aws", "s3api", "head-bucket",
        "--bucket", bucket_name,
        "--profile", AWS_PROFILE
    ]

    returncode, stdout, stderr = run_command(cmd)

    if returncode != 0:
        if "404" in stderr or "Not Found" in stderr:
            return HealthCheckResult(
                resource_type="S3 Bucket",
                resource_name=bucket_name,
                status=ResourceStatus.MISSING,
                details="S3 bucket not found"
            )
        if "403" in stderr or "Forbidden" in stderr:
            return HealthCheckResult(
                resource_type="S3 Bucket",
                resource_name=bucket_name,
                status=ResourceStatus.BROKEN,
                details="S3 bucket exists but access denied"
            )
        return HealthCheckResult(
            resource_type="S3 Bucket",
            resource_name=bucket_name,
            status=ResourceStatus.UNKNOWN,
            details=f"Error checking bucket: {stderr}"
        )

    return HealthCheckResult(
        resource_type="S3 Bucket",
        resource_name=bucket_name,
        status=ResourceStatus.HEALTHY,
        details="S3 bucket exists and accessible"
    )


def check_vpc(vpc_id: str) -> HealthCheckResult:
    """Check if VPC exists (for classic workspaces)."""
    cmd = [
        "aws", "ec2", "describe-vpcs",
        "--vpc-ids", vpc_id,
        "--profile", AWS_PROFILE,
        "--output", "json"
    ]

    returncode, stdout, stderr = run_command(cmd)

    if returncode != 0:
        if "InvalidVpcID.NotFound" in stderr:
            return HealthCheckResult(
                resource_type="VPC",
                resource_name=vpc_id,
                status=ResourceStatus.MISSING,
                details="VPC not found - workspace likely non-functional"
            )
        return HealthCheckResult(
            resource_type="VPC",
            resource_name=vpc_id,
            status=ResourceStatus.UNKNOWN,
            details=f"Error checking VPC: {stderr}"
        )

    try:
        data = json.loads(stdout)
        vpcs = data.get("Vpcs", [])
        if not vpcs:
            return HealthCheckResult(
                resource_type="VPC",
                resource_name=vpc_id,
                status=ResourceStatus.MISSING,
                details="VPC not found"
            )

        vpc_state = vpcs[0].get("State", "unknown")
        if vpc_state == "available":
            return HealthCheckResult(
                resource_type="VPC",
                resource_name=vpc_id,
                status=ResourceStatus.HEALTHY,
                details=f"VPC is {vpc_state}"
            )
        return HealthCheckResult(
            resource_type="VPC",
            resource_name=vpc_id,
            status=ResourceStatus.BROKEN,
            details=f"VPC state is {vpc_state}"
        )
    except json.JSONDecodeError:
        return HealthCheckResult(
            resource_type="VPC",
            resource_name=vpc_id,
            status=ResourceStatus.UNKNOWN,
            details="Could not parse VPC response"
        )


def check_storage_credential(workspace_profile: str, cred_name: str) -> HealthCheckResult:
    """Check if storage credential exists in workspace."""
    cmd = [
        "databricks", "storage-credentials", "get",
        cred_name,
        "--profile", workspace_profile,
        "--output", "json"
    ]

    returncode, stdout, stderr = run_command(cmd)

    if returncode != 0:
        if "NOT_FOUND" in stderr or "RESOURCE_DOES_NOT_EXIST" in stderr:
            return HealthCheckResult(
                resource_type="Storage Credential",
                resource_name=cred_name,
                status=ResourceStatus.MISSING,
                details="Storage credential not found in workspace"
            )
        if "UNAUTHENTICATED" in stderr or "401" in stderr:
            return HealthCheckResult(
                resource_type="Storage Credential",
                resource_name=cred_name,
                status=ResourceStatus.UNKNOWN,
                details="Not authenticated to workspace - run databricks auth login"
            )
        return HealthCheckResult(
            resource_type="Storage Credential",
            resource_name=cred_name,
            status=ResourceStatus.UNKNOWN,
            details=f"Error checking credential: {stderr}"
        )

    return HealthCheckResult(
        resource_type="Storage Credential",
        resource_name=cred_name,
        status=ResourceStatus.HEALTHY,
        details="Storage credential exists"
    )


def run_full_health_check(workspace_name: str) -> list[HealthCheckResult]:
    """Run all health checks for a workspace."""
    results = []

    # Get workspace from registry
    workspace = get_workspace(workspace_name)
    if not workspace:
        print(f"Workspace '{workspace_name}' not found in registry", file=sys.stderr)
        print("Use 'python registry_manager.py list-workspaces' to see registered workspaces")
        return []

    workspace_id = workspace.get("workspace_id")
    dependencies = workspace.get("dependencies", {})
    workspace_type = workspace.get("type", "serverless")

    # Check workspace
    results.append(check_databricks_workspace(workspace_id))

    # Check metastore assignment
    results.append(check_metastore_assignment(workspace_id))

    # Check IAM role
    iam_role_arn = dependencies.get("iam_uc_access_role")
    if iam_role_arn:
        role_name = iam_role_arn.split("/")[-1]
        results.append(check_iam_role(role_name))

    # Check S3 bucket
    s3_bucket = dependencies.get("s3_uc_root_bucket")
    if s3_bucket:
        results.append(check_s3_bucket(s3_bucket))

    # Check VPC (classic only)
    vpc_id = dependencies.get("vpc_id")
    if vpc_id and workspace_type == "classic":
        results.append(check_vpc(vpc_id))

    # Check storage credential (requires workspace auth)
    workspace_profile = f"oneenv-{workspace_name}"
    cred_name = f"oneenv-{workspace_name}-cred"
    results.append(check_storage_credential(workspace_profile, cred_name))

    return results


def print_health_report(workspace_name: str, results: list[HealthCheckResult]) -> None:
    """Print a formatted health report."""
    print(f"\nResource Health Check for: {workspace_name}")
    print("=" * 60)
    print(f"{'Resource':<25} {'Status':<12} {'Notes'}")
    print("-" * 60)

    issues = 0
    for r in results:
        status_str = r.status.value
        if r.status == ResourceStatus.HEALTHY:
            status_display = status_str
        elif r.status == ResourceStatus.MISSING:
            status_display = status_str
            issues += 1
        elif r.status == ResourceStatus.BROKEN:
            status_display = status_str
            issues += 1
        else:
            status_display = status_str

        print(f"{r.resource_type:<25} {status_display:<12} {r.details}")

    print("-" * 60)
    if issues > 0:
        print(f"\nIssues Found: {issues}")
        print("\nOptions:")
        print("  1. Repair: python health_checker.py repair --workspace <name>")
        print("  2. Start Fresh: python registry_manager.py cleanup --workspace <name>")
    else:
        print("\nAll resources healthy!")

    # Update registry with overall status
    overall_status = "healthy" if issues == 0 else "broken"
    update_verification_status(workspace_name, overall_status)


def suggest_repairs(results: list[HealthCheckResult]) -> list[str]:
    """Suggest repair commands based on health check results."""
    repairs = []

    for r in results:
        if r.status == ResourceStatus.MISSING:
            if r.resource_type == "S3 Bucket":
                repairs.append(f"# Recreate S3 bucket: {r.resource_name}")
                repairs.append(f"aws s3api create-bucket --bucket {r.resource_name} --profile {AWS_PROFILE}")
                repairs.append(f"aws s3api put-public-access-block --bucket {r.resource_name} --public-access-block-configuration 'BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true' --profile {AWS_PROFILE}")

            elif r.resource_type == "IAM Role":
                repairs.append(f"# Recreate IAM role: {r.resource_name}")
                repairs.append(f"# See SKILL.md Step 4b for full IAM role creation commands")

            elif r.resource_type == "Storage Credential":
                repairs.append(f"# Recreate storage credential after IAM role is restored")
                repairs.append(f"# databricks storage-credentials create --name {r.resource_name} --aws-iam-role-arn <role-arn>")

            elif r.resource_type == "VPC":
                repairs.append(f"# VPC was deleted - recommend starting fresh")
                repairs.append(f"# Classic workspaces with missing VPCs are non-functional")

            elif r.resource_type == "Workspace":
                repairs.append(f"# Workspace was deleted - must create new workspace")

        elif r.status == ResourceStatus.BROKEN:
            repairs.append(f"# {r.resource_type} is broken: {r.details}")

    return repairs


def main():
    parser = argparse.ArgumentParser(description="One-Env Workspace Health Checker")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # check-all
    check_all = subparsers.add_parser("check-all", help="Full health check")
    check_all.add_argument("--workspace", required=True, help="Workspace name")

    # check-workspace
    check_ws = subparsers.add_parser("check-workspace", help="Check workspace only")
    check_ws.add_argument("--workspace", required=True, help="Workspace name")

    # check-aws
    check_aws = subparsers.add_parser("check-aws", help="Check AWS resources only")
    check_aws.add_argument("--workspace", required=True, help="Workspace name")

    # repair
    repair = subparsers.add_parser("repair", help="Show repair commands")
    repair.add_argument("--workspace", required=True, help="Workspace name")
    repair.add_argument("--confirm", action="store_true", help="Actually run repairs (not implemented - manual for safety)")

    args = parser.parse_args()

    if args.command == "check-all":
        results = run_full_health_check(args.workspace)
        if results:
            print_health_report(args.workspace, results)

    elif args.command == "check-workspace":
        workspace = get_workspace(args.workspace)
        if not workspace:
            print(f"Workspace '{args.workspace}' not found in registry", file=sys.stderr)
            sys.exit(1)
        result = check_databricks_workspace(workspace["workspace_id"])
        print(f"{result.resource_type}: {result.status.value} - {result.details}")

    elif args.command == "check-aws":
        workspace = get_workspace(args.workspace)
        if not workspace:
            print(f"Workspace '{args.workspace}' not found in registry", file=sys.stderr)
            sys.exit(1)

        deps = workspace.get("dependencies", {})
        results = []

        if deps.get("iam_uc_access_role"):
            role_name = deps["iam_uc_access_role"].split("/")[-1]
            results.append(check_iam_role(role_name))

        if deps.get("s3_uc_root_bucket"):
            results.append(check_s3_bucket(deps["s3_uc_root_bucket"]))

        if deps.get("vpc_id"):
            results.append(check_vpc(deps["vpc_id"]))

        for r in results:
            print(f"{r.resource_type}: {r.status.value} - {r.details}")

    elif args.command == "repair":
        results = run_full_health_check(args.workspace)
        if not results:
            sys.exit(1)

        repairs = suggest_repairs(results)
        if not repairs:
            print("No repairs needed - all resources healthy!")
        else:
            print("\nSuggested repair commands:")
            print("=" * 60)
            for cmd in repairs:
                print(cmd)
            print("\nNote: Review and run these commands manually for safety.")

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
