#!/usr/bin/env python3
"""
Full Cleanup for One-Env Workspace Deployments

Deletes all resources associated with a workspace in the correct order.
This handles both resources tracked in the registry AND resources that
follow our naming convention.

Deletion order (reverse of creation):
1. Integration test resources (DynamoDB, Kinesis, RDS, MSK)
2. Databricks workspace resources (catalog, schemas, external locations, storage credentials)
3. Instance profile registration in Databricks
4. AWS IAM (instance profiles, roles)
5. AWS S3 buckets
6. Databricks workspace
7. Databricks account resources (network config, credentials config, storage config)
8. AWS VPC resources (if classic workspace)

Usage:
    uv run full_cleanup.py --workspace NAME --region REGION [--dry-run] [--confirm]
    uv run full_cleanup.py --workspace NAME --region REGION --skip-workspace  # Keep workspace, delete supporting resources
"""

import argparse
import json
import subprocess
import sys
import time
from pathlib import Path

# Constants
AWS_PROFILE = "aws-sandbox-field-eng_databricks-sandbox-admin"
DATABRICKS_ACCOUNT_PROFILE = "one-env-admin-aws"
AWS_ACCOUNT_ID = "332745928618"


def run_command(cmd: list[str], description: str = "", ignore_errors: bool = False) -> tuple[bool, str]:
    """Run a command and return (success, output)."""
    if description:
        print(f"  → {description}")

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)
        if result.returncode != 0:
            if not ignore_errors:
                print(f"    ⚠ Failed: {result.stderr.strip()[:200]}")
            return False, result.stderr
        return True, result.stdout
    except subprocess.TimeoutExpired:
        print(f"    ⚠ Timeout")
        return False, "Command timed out"
    except Exception as e:
        print(f"    ⚠ Error: {e}")
        return False, str(e)


def get_workspace_profile(workspace_name: str) -> str:
    """Get the Databricks CLI profile name for a workspace."""
    return f"oneenv-{workspace_name}"


def discover_resources(workspace_name: str, region: str) -> dict:
    """
    Discover all resources associated with a workspace.
    Uses both registry data and naming convention scanning.
    """
    print(f"\n🔍 Discovering resources for workspace: {workspace_name}")

    resources = {
        "workspace": None,
        "integration_resources": {"dynamodb": [], "kinesis": [], "rds": [], "msk": []},
        "databricks_workspace": {
            "catalogs": [],
            "external_locations": [],
            "storage_credentials": [],
            "instance_profiles": []
        },
        "aws_iam": {"roles": [], "instance_profiles": []},
        "aws_s3": [],
        "databricks_account": {"network_id": None, "credentials_id": None, "storage_config_id": None},
        "aws_vpc": {"vpc_id": None, "subnets": [], "security_groups": [], "nat_gateways": [], "internet_gateways": []}
    }

    # Load workspace file if exists
    workspace_file = Path.home() / ".vibe" / "oneenv" / "workspaces" / f"{workspace_name}.json"
    if workspace_file.exists():
        with open(workspace_file) as f:
            ws_data = json.load(f)
            resources["workspace"] = ws_data

            # Extract tracked resources
            aws_res = ws_data.get("aws_resources", {})
            db_res = ws_data.get("databricks_resources", {})

            resources["databricks_account"]["network_id"] = db_res.get("network_id")
            resources["databricks_account"]["credentials_id"] = db_res.get("credentials_id")
            resources["databricks_account"]["storage_config_id"] = db_res.get("storage_configuration_id")
            resources["aws_vpc"]["vpc_id"] = aws_res.get("vpc_id")

    # Scan for S3 buckets matching our naming pattern
    print("  Scanning S3 buckets...")
    success, output = run_command(
        ["aws", "s3", "ls", "--profile", AWS_PROFILE],
        ignore_errors=True
    )
    if success:
        for line in output.strip().split("\n"):
            if workspace_name in line:
                bucket = line.split()[-1]
                resources["aws_s3"].append(bucket)
                print(f"    Found bucket: {bucket}")

    # Scan for IAM roles matching our naming pattern
    print("  Scanning IAM roles...")
    success, output = run_command(
        ["aws", "iam", "list-roles", "--profile", AWS_PROFILE, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            roles = json.loads(output)
            for role in roles.get("Roles", []):
                if workspace_name in role["RoleName"]:
                    resources["aws_iam"]["roles"].append(role["RoleName"])
                    print(f"    Found role: {role['RoleName']}")
        except json.JSONDecodeError:
            pass

    # Scan for IAM instance profiles
    print("  Scanning IAM instance profiles...")
    success, output = run_command(
        ["aws", "iam", "list-instance-profiles", "--profile", AWS_PROFILE, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            profiles = json.loads(output)
            for profile in profiles.get("InstanceProfiles", []):
                if workspace_name in profile["InstanceProfileName"]:
                    resources["aws_iam"]["instance_profiles"].append({
                        "name": profile["InstanceProfileName"],
                        "roles": [r["RoleName"] for r in profile.get("Roles", [])]
                    })
                    print(f"    Found instance profile: {profile['InstanceProfileName']}")
        except json.JSONDecodeError:
            pass

    # Scan for DynamoDB tables
    print("  Scanning DynamoDB tables...")
    success, output = run_command(
        ["aws", "dynamodb", "list-tables", "--region", region, "--profile", AWS_PROFILE, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            tables = json.loads(output)
            for table in tables.get("TableNames", []):
                if workspace_name in table:
                    resources["integration_resources"]["dynamodb"].append(table)
                    print(f"    Found DynamoDB table: {table}")
        except json.JSONDecodeError:
            pass

    # Scan for Kinesis streams
    print("  Scanning Kinesis streams...")
    success, output = run_command(
        ["aws", "kinesis", "list-streams", "--region", region, "--profile", AWS_PROFILE, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            streams = json.loads(output)
            for stream in streams.get("StreamNames", []):
                if workspace_name in stream:
                    resources["integration_resources"]["kinesis"].append(stream)
                    print(f"    Found Kinesis stream: {stream}")
        except json.JSONDecodeError:
            pass

    # Scan for RDS instances
    print("  Scanning RDS instances...")
    success, output = run_command(
        ["aws", "rds", "describe-db-instances", "--region", region, "--profile", AWS_PROFILE, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            instances = json.loads(output)
            for instance in instances.get("DBInstances", []):
                # RDS identifiers use hyphens, workspace names might have underscores
                if workspace_name.replace("_", "-") in instance["DBInstanceIdentifier"]:
                    resources["integration_resources"]["rds"].append(instance["DBInstanceIdentifier"])
                    print(f"    Found RDS instance: {instance['DBInstanceIdentifier']}")
        except json.JSONDecodeError:
            pass

    # Scan for MSK clusters
    print("  Scanning MSK clusters...")
    success, output = run_command(
        ["aws", "kafka", "list-clusters-v2", "--region", region, "--profile", AWS_PROFILE, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            clusters = json.loads(output)
            for cluster in clusters.get("ClusterInfoList", []):
                if workspace_name in cluster["ClusterName"]:
                    resources["integration_resources"]["msk"].append({
                        "name": cluster["ClusterName"],
                        "arn": cluster["ClusterArn"]
                    })
                    print(f"    Found MSK cluster: {cluster['ClusterName']}")
        except json.JSONDecodeError:
            pass

    # Scan for Secrets Manager secrets
    print("  Scanning Secrets Manager...")
    success, output = run_command(
        ["aws", "secretsmanager", "list-secrets", "--region", region, "--profile", AWS_PROFILE, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            secrets = json.loads(output)
            for secret in secrets.get("SecretList", []):
                if workspace_name in secret["Name"]:
                    if "secrets" not in resources["integration_resources"]:
                        resources["integration_resources"]["secrets"] = []
                    resources["integration_resources"]["secrets"].append(secret["Name"])
                    print(f"    Found secret: {secret['Name']}")
        except json.JSONDecodeError:
            pass

    # Scan for VPC resources if we have a VPC ID
    vpc_id = resources["aws_vpc"]["vpc_id"]
    if vpc_id:
        print(f"  Scanning VPC resources for {vpc_id}...")

        # Subnets
        success, output = run_command(
            ["aws", "ec2", "describe-subnets", "--filters", f"Name=vpc-id,Values={vpc_id}",
             "--region", region, "--profile", AWS_PROFILE, "--output", "json"],
            ignore_errors=True
        )
        if success:
            try:
                subnets = json.loads(output)
                resources["aws_vpc"]["subnets"] = [s["SubnetId"] for s in subnets.get("Subnets", [])]
            except json.JSONDecodeError:
                pass

        # Security groups
        success, output = run_command(
            ["aws", "ec2", "describe-security-groups", "--filters", f"Name=vpc-id,Values={vpc_id}",
             "--region", region, "--profile", AWS_PROFILE, "--output", "json"],
            ignore_errors=True
        )
        if success:
            try:
                sgs = json.loads(output)
                resources["aws_vpc"]["security_groups"] = [
                    sg["GroupId"] for sg in sgs.get("SecurityGroups", [])
                    if sg["GroupName"] != "default"  # Can't delete default SG
                ]
            except json.JSONDecodeError:
                pass

        # NAT Gateways
        success, output = run_command(
            ["aws", "ec2", "describe-nat-gateways", "--filter", f"Name=vpc-id,Values={vpc_id}",
             "--region", region, "--profile", AWS_PROFILE, "--output", "json"],
            ignore_errors=True
        )
        if success:
            try:
                nats = json.loads(output)
                resources["aws_vpc"]["nat_gateways"] = [
                    n["NatGatewayId"] for n in nats.get("NatGateways", [])
                    if n["State"] != "deleted"
                ]
            except json.JSONDecodeError:
                pass

        # Internet Gateways
        success, output = run_command(
            ["aws", "ec2", "describe-internet-gateways",
             "--filters", f"Name=attachment.vpc-id,Values={vpc_id}",
             "--region", region, "--profile", AWS_PROFILE, "--output", "json"],
            ignore_errors=True
        )
        if success:
            try:
                igws = json.loads(output)
                resources["aws_vpc"]["internet_gateways"] = [
                    igw["InternetGatewayId"] for igw in igws.get("InternetGateways", [])
                ]
            except json.JSONDecodeError:
                pass

    # Scan Databricks workspace for UC resources
    workspace_profile = get_workspace_profile(workspace_name)

    # Try to list storage credentials
    print("  Scanning Databricks storage credentials...")
    success, output = run_command(
        ["databricks", "storage-credentials", "list", "--profile", workspace_profile, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            creds = json.loads(output)
            for cred in creds if isinstance(creds, list) else creds.get("storage_credentials", []):
                cred_name = cred.get("name", "")
                if workspace_name in cred_name or "oneenv" in cred_name:
                    resources["databricks_workspace"]["storage_credentials"].append(cred_name)
                    print(f"    Found storage credential: {cred_name}")
        except json.JSONDecodeError:
            pass

    # Try to list external locations
    print("  Scanning Databricks external locations...")
    success, output = run_command(
        ["databricks", "external-locations", "list", "--profile", workspace_profile, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            locs = json.loads(output)
            for loc in locs if isinstance(locs, list) else locs.get("external_locations", []):
                loc_name = loc.get("name", "")
                if workspace_name in loc_name or "oneenv" in loc_name:
                    resources["databricks_workspace"]["external_locations"].append(loc_name)
                    print(f"    Found external location: {loc_name}")
        except json.JSONDecodeError:
            pass

    # Try to list catalogs
    print("  Scanning Databricks catalogs...")
    success, output = run_command(
        ["databricks", "catalogs", "list", "--profile", workspace_profile, "--output", "json"],
        ignore_errors=True
    )
    if success:
        try:
            cats = json.loads(output)
            for cat in cats if isinstance(cats, list) else cats.get("catalogs", []):
                cat_name = cat.get("name", "")
                # Match our naming convention (workspace name with underscores)
                ws_underscore = workspace_name.replace("-", "_")
                if ws_underscore in cat_name or "oneenv" in cat_name:
                    resources["databricks_workspace"]["catalogs"].append(cat_name)
                    print(f"    Found catalog: {cat_name}")
        except json.JSONDecodeError:
            pass

    return resources


def print_resources_summary(resources: dict):
    """Print a summary of discovered resources."""
    print("\n" + "=" * 60)
    print("RESOURCE SUMMARY")
    print("=" * 60)

    # Integration resources
    int_res = resources["integration_resources"]
    if any([int_res["dynamodb"], int_res["kinesis"], int_res["rds"], int_res["msk"], int_res.get("secrets", [])]):
        print("\n📦 Integration Test Resources:")
        for table in int_res["dynamodb"]:
            print(f"  • DynamoDB: {table}")
        for stream in int_res["kinesis"]:
            print(f"  • Kinesis: {stream}")
        for db in int_res["rds"]:
            print(f"  • RDS: {db}")
        for cluster in int_res["msk"]:
            print(f"  • MSK: {cluster['name']}")
        for secret in int_res.get("secrets", []):
            print(f"  • Secret: {secret}")

    # Databricks workspace resources
    db_ws = resources["databricks_workspace"]
    if any([db_ws["catalogs"], db_ws["external_locations"], db_ws["storage_credentials"]]):
        print("\n🔷 Databricks Workspace Resources:")
        for cat in db_ws["catalogs"]:
            print(f"  • Catalog: {cat}")
        for loc in db_ws["external_locations"]:
            print(f"  • External Location: {loc}")
        for cred in db_ws["storage_credentials"]:
            print(f"  • Storage Credential: {cred}")

    # AWS IAM
    aws_iam = resources["aws_iam"]
    if any([aws_iam["roles"], aws_iam["instance_profiles"]]):
        print("\n🔐 AWS IAM:")
        for profile in aws_iam["instance_profiles"]:
            print(f"  • Instance Profile: {profile['name']}")
        for role in aws_iam["roles"]:
            print(f"  • Role: {role}")

    # S3 buckets
    if resources["aws_s3"]:
        print("\n🪣 AWS S3 Buckets:")
        for bucket in resources["aws_s3"]:
            print(f"  • {bucket}")

    # Databricks account resources
    db_acc = resources["databricks_account"]
    if any([db_acc["network_id"], db_acc["credentials_id"], db_acc["storage_config_id"]]):
        print("\n🏢 Databricks Account Resources:")
        if db_acc["network_id"]:
            print(f"  • Network Config: {db_acc['network_id']}")
        if db_acc["credentials_id"]:
            print(f"  • Credentials Config: {db_acc['credentials_id']}")
        if db_acc["storage_config_id"]:
            print(f"  • Storage Config: {db_acc['storage_config_id']}")

    # VPC resources
    vpc = resources["aws_vpc"]
    if vpc["vpc_id"]:
        print("\n🌐 AWS VPC Resources:")
        print(f"  • VPC: {vpc['vpc_id']}")
        for subnet in vpc["subnets"]:
            print(f"  • Subnet: {subnet}")
        for sg in vpc["security_groups"]:
            print(f"  • Security Group: {sg}")
        for nat in vpc["nat_gateways"]:
            print(f"  • NAT Gateway: {nat}")
        for igw in vpc["internet_gateways"]:
            print(f"  • Internet Gateway: {igw}")

    # Workspace
    if resources["workspace"]:
        print("\n🏠 Workspace:")
        ws = resources["workspace"]
        print(f"  • Name: {ws.get('workspace_name')}")
        print(f"  • ID: {ws.get('workspace_id')}")
        print(f"  • URL: {ws.get('workspace_url')}")

    print("\n" + "=" * 60)


def delete_resources(workspace_name: str, region: str, resources: dict,
                     skip_workspace: bool = False, dry_run: bool = False):
    """Delete resources in the correct order."""

    workspace_profile = get_workspace_profile(workspace_name)

    if dry_run:
        print("\n🔍 DRY RUN - No resources will be deleted\n")
    else:
        print("\n🗑️  DELETING RESOURCES\n")

    # 1. Delete integration test resources
    print("Step 1: Integration Test Resources")

    for table in resources["integration_resources"]["dynamodb"]:
        cmd = ["aws", "dynamodb", "delete-table", "--table-name", table,
               "--region", region, "--profile", AWS_PROFILE]
        if not dry_run:
            run_command(cmd, f"Deleting DynamoDB table: {table}", ignore_errors=True)
        else:
            print(f"  Would delete DynamoDB table: {table}")

    for stream in resources["integration_resources"]["kinesis"]:
        cmd = ["aws", "kinesis", "delete-stream", "--stream-name", stream,
               "--region", region, "--profile", AWS_PROFILE]
        if not dry_run:
            run_command(cmd, f"Deleting Kinesis stream: {stream}", ignore_errors=True)
        else:
            print(f"  Would delete Kinesis stream: {stream}")

    for db in resources["integration_resources"]["rds"]:
        cmd = ["aws", "rds", "delete-db-instance", "--db-instance-identifier", db,
               "--skip-final-snapshot", "--delete-automated-backups",
               "--region", region, "--profile", AWS_PROFILE]
        if not dry_run:
            run_command(cmd, f"Deleting RDS instance: {db}", ignore_errors=True)
        else:
            print(f"  Would delete RDS instance: {db}")

    for cluster in resources["integration_resources"]["msk"]:
        cmd = ["aws", "kafka", "delete-cluster", "--cluster-arn", cluster["arn"],
               "--region", region, "--profile", AWS_PROFILE]
        if not dry_run:
            run_command(cmd, f"Deleting MSK cluster: {cluster['name']}", ignore_errors=True)
        else:
            print(f"  Would delete MSK cluster: {cluster['name']}")

    for secret in resources["integration_resources"].get("secrets", []):
        cmd = ["aws", "secretsmanager", "delete-secret", "--secret-id", secret,
               "--force-delete-without-recovery", "--region", region, "--profile", AWS_PROFILE]
        if not dry_run:
            run_command(cmd, f"Deleting secret: {secret}", ignore_errors=True)
        else:
            print(f"  Would delete secret: {secret}")

    # 2. Delete Databricks workspace resources
    print("\nStep 2: Databricks Workspace Resources")

    for cat in resources["databricks_workspace"]["catalogs"]:
        cmd = ["databricks", "catalogs", "delete", cat, "--force", "--profile", workspace_profile]
        if not dry_run:
            run_command(cmd, f"Deleting catalog: {cat}", ignore_errors=True)
        else:
            print(f"  Would delete catalog: {cat}")

    for loc in resources["databricks_workspace"]["external_locations"]:
        cmd = ["databricks", "external-locations", "delete", loc, "--force", "--profile", workspace_profile]
        if not dry_run:
            run_command(cmd, f"Deleting external location: {loc}", ignore_errors=True)
        else:
            print(f"  Would delete external location: {loc}")

    for cred in resources["databricks_workspace"]["storage_credentials"]:
        cmd = ["databricks", "storage-credentials", "delete", cred, "--force", "--profile", workspace_profile]
        if not dry_run:
            run_command(cmd, f"Deleting storage credential: {cred}", ignore_errors=True)
        else:
            print(f"  Would delete storage credential: {cred}")

    # 3. Delete AWS IAM resources
    print("\nStep 3: AWS IAM Resources")

    for profile in resources["aws_iam"]["instance_profiles"]:
        # First remove roles from instance profile
        for role in profile["roles"]:
            cmd = ["aws", "iam", "remove-role-from-instance-profile",
                   "--instance-profile-name", profile["name"], "--role-name", role,
                   "--profile", AWS_PROFILE]
            if not dry_run:
                run_command(cmd, f"Removing role {role} from instance profile", ignore_errors=True)

        # Then delete instance profile
        cmd = ["aws", "iam", "delete-instance-profile",
               "--instance-profile-name", profile["name"], "--profile", AWS_PROFILE]
        if not dry_run:
            run_command(cmd, f"Deleting instance profile: {profile['name']}", ignore_errors=True)
        else:
            print(f"  Would delete instance profile: {profile['name']}")

    for role in resources["aws_iam"]["roles"]:
        # First delete inline policies
        success, output = run_command(
            ["aws", "iam", "list-role-policies", "--role-name", role, "--profile", AWS_PROFILE, "--output", "json"],
            ignore_errors=True
        )
        if success:
            try:
                policies = json.loads(output)
                for policy in policies.get("PolicyNames", []):
                    cmd = ["aws", "iam", "delete-role-policy", "--role-name", role,
                           "--policy-name", policy, "--profile", AWS_PROFILE]
                    if not dry_run:
                        run_command(cmd, f"Deleting role policy: {policy}", ignore_errors=True)
            except json.JSONDecodeError:
                pass

        # Then delete the role
        cmd = ["aws", "iam", "delete-role", "--role-name", role, "--profile", AWS_PROFILE]
        if not dry_run:
            run_command(cmd, f"Deleting role: {role}", ignore_errors=True)
        else:
            print(f"  Would delete role: {role}")

    # 4. Delete S3 buckets
    print("\nStep 4: AWS S3 Buckets")

    for bucket in resources["aws_s3"]:
        # First empty the bucket
        cmd = ["aws", "s3", "rm", f"s3://{bucket}", "--recursive", "--profile", AWS_PROFILE]
        if not dry_run:
            run_command(cmd, f"Emptying bucket: {bucket}", ignore_errors=True)

        # Then delete the bucket
        cmd = ["aws", "s3api", "delete-bucket", "--bucket", bucket, "--profile", AWS_PROFILE]
        if not dry_run:
            run_command(cmd, f"Deleting bucket: {bucket}", ignore_errors=True)
        else:
            print(f"  Would delete bucket: {bucket}")

    if not skip_workspace:
        # 5. Delete Databricks workspace
        print("\nStep 5: Databricks Workspace")

        ws = resources.get("workspace", {})
        workspace_id = ws.get("workspace_id")
        if workspace_id:
            cmd = ["databricks", "account", "workspaces", "delete",
                   "--workspace-id", workspace_id, "--profile", DATABRICKS_ACCOUNT_PROFILE]
            if not dry_run:
                run_command(cmd, f"Deleting workspace: {workspace_name}", ignore_errors=True)
                print("  ⏳ Waiting for workspace deletion (30 seconds)...")
                if not dry_run:
                    time.sleep(30)
            else:
                print(f"  Would delete workspace: {workspace_name} (ID: {workspace_id})")

        # 6. Delete Databricks account resources
        print("\nStep 6: Databricks Account Resources")

        db_acc = resources["databricks_account"]

        if db_acc["storage_config_id"]:
            cmd = ["databricks", "account", "storage-configurations", "delete",
                   "--storage-configuration-id", db_acc["storage_config_id"],
                   "--profile", DATABRICKS_ACCOUNT_PROFILE]
            if not dry_run:
                run_command(cmd, f"Deleting storage config: {db_acc['storage_config_id']}", ignore_errors=True)
            else:
                print(f"  Would delete storage config: {db_acc['storage_config_id']}")

        if db_acc["credentials_id"]:
            cmd = ["databricks", "account", "credentials", "delete",
                   "--credentials-id", db_acc["credentials_id"],
                   "--profile", DATABRICKS_ACCOUNT_PROFILE]
            if not dry_run:
                run_command(cmd, f"Deleting credentials config: {db_acc['credentials_id']}", ignore_errors=True)
            else:
                print(f"  Would delete credentials config: {db_acc['credentials_id']}")

        if db_acc["network_id"]:
            cmd = ["databricks", "account", "networks", "delete",
                   "--network-id", db_acc["network_id"],
                   "--profile", DATABRICKS_ACCOUNT_PROFILE]
            if not dry_run:
                run_command(cmd, f"Deleting network config: {db_acc['network_id']}", ignore_errors=True)
            else:
                print(f"  Would delete network config: {db_acc['network_id']}")

        # 7. Delete VPC resources
        vpc = resources["aws_vpc"]
        if vpc["vpc_id"]:
            print("\nStep 7: AWS VPC Resources")

            # Delete NAT Gateways first (they take time)
            for nat in vpc["nat_gateways"]:
                cmd = ["aws", "ec2", "delete-nat-gateway", "--nat-gateway-id", nat,
                       "--region", region, "--profile", AWS_PROFILE]
                if not dry_run:
                    run_command(cmd, f"Deleting NAT Gateway: {nat}", ignore_errors=True)
                else:
                    print(f"  Would delete NAT Gateway: {nat}")

            if vpc["nat_gateways"] and not dry_run:
                print("  ⏳ Waiting for NAT Gateway deletion (60 seconds)...")
                time.sleep(60)

            # Detach and delete Internet Gateways
            for igw in vpc["internet_gateways"]:
                cmd = ["aws", "ec2", "detach-internet-gateway", "--internet-gateway-id", igw,
                       "--vpc-id", vpc["vpc_id"], "--region", region, "--profile", AWS_PROFILE]
                if not dry_run:
                    run_command(cmd, f"Detaching Internet Gateway: {igw}", ignore_errors=True)

                cmd = ["aws", "ec2", "delete-internet-gateway", "--internet-gateway-id", igw,
                       "--region", region, "--profile", AWS_PROFILE]
                if not dry_run:
                    run_command(cmd, f"Deleting Internet Gateway: {igw}", ignore_errors=True)
                else:
                    print(f"  Would delete Internet Gateway: {igw}")

            # Delete subnets
            for subnet in vpc["subnets"]:
                cmd = ["aws", "ec2", "delete-subnet", "--subnet-id", subnet,
                       "--region", region, "--profile", AWS_PROFILE]
                if not dry_run:
                    run_command(cmd, f"Deleting subnet: {subnet}", ignore_errors=True)
                else:
                    print(f"  Would delete subnet: {subnet}")

            # Delete security groups
            for sg in vpc["security_groups"]:
                cmd = ["aws", "ec2", "delete-security-group", "--group-id", sg,
                       "--region", region, "--profile", AWS_PROFILE]
                if not dry_run:
                    run_command(cmd, f"Deleting security group: {sg}", ignore_errors=True)
                else:
                    print(f"  Would delete security group: {sg}")

            # Finally delete VPC
            cmd = ["aws", "ec2", "delete-vpc", "--vpc-id", vpc["vpc_id"],
                   "--region", region, "--profile", AWS_PROFILE]
            if not dry_run:
                run_command(cmd, f"Deleting VPC: {vpc['vpc_id']}", ignore_errors=True)
            else:
                print(f"  Would delete VPC: {vpc['vpc_id']}")

    # 8. Clean up registry
    print("\nStep 8: Registry Cleanup")

    workspace_file = Path.home() / ".vibe" / "oneenv" / "workspaces" / f"{workspace_name}.json"
    if workspace_file.exists():
        if not dry_run:
            workspace_file.unlink()
            print(f"  → Deleted registry file: {workspace_file}")
        else:
            print(f"  Would delete registry file: {workspace_file}")

    if dry_run:
        print("\n✅ Dry run complete. Run with --confirm to actually delete.")
    else:
        print("\n✅ Cleanup complete!")


def main():
    parser = argparse.ArgumentParser(description="Full cleanup for One-Env workspace")
    parser.add_argument("--workspace", required=True, help="Workspace name")
    parser.add_argument("--region", required=True, help="AWS region")
    parser.add_argument("--dry-run", action="store_true", help="Show what would be deleted without deleting")
    parser.add_argument("--confirm", action="store_true", help="Actually delete resources")
    parser.add_argument("--skip-workspace", action="store_true",
                        help="Keep workspace, only delete supporting resources")

    args = parser.parse_args()

    if not args.dry_run and not args.confirm:
        print("Must specify --dry-run or --confirm")
        sys.exit(1)

    # Discover resources
    resources = discover_resources(args.workspace, args.region)

    # Print summary
    print_resources_summary(resources)

    # Delete resources
    delete_resources(
        args.workspace,
        args.region,
        resources,
        skip_workspace=args.skip_workspace,
        dry_run=args.dry_run
    )


if __name__ == "__main__":
    main()
