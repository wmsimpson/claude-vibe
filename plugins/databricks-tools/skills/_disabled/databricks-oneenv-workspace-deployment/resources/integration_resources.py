#!/usr/bin/env python3
"""
Integration Resources Manager for One-Env Workspace Deployments

Creates and manages AWS integration test resources (DynamoDB, Kinesis, RDS, MSK)
for testing Databricks integrations.

Usage:
    uv run integration_resources.py create-dynamodb --workspace NAME --table-name NAME
    uv run integration_resources.py create-kinesis --workspace NAME --stream-name NAME
    uv run integration_resources.py create-rds --workspace NAME --db-name NAME
    uv run integration_resources.py create-msk --workspace NAME --cluster-name NAME
    uv run integration_resources.py list --workspace NAME
    uv run integration_resources.py cleanup --workspace NAME [--resource-type TYPE]
"""

import argparse
import json
import subprocess
import sys
import time
from pathlib import Path

# Constants
AWS_PROFILE = "aws-sandbox-field-eng_databricks-sandbox-admin"
AWS_ACCOUNT_ID = "332745928618"

# Import registry manager for tracking
sys.path.insert(0, str(Path(__file__).parent))
from registry_manager import load_registry, save_registry, get_workspace


def run_aws_command(cmd: list[str], parse_json: bool = True) -> tuple[bool, any]:
    """Run AWS CLI command and return (success, result)."""
    full_cmd = ["aws"] + cmd + ["--profile", AWS_PROFILE]
    if parse_json:
        full_cmd.extend(["--output", "json"])

    try:
        result = subprocess.run(full_cmd, capture_output=True, text=True, timeout=300)
        if result.returncode != 0:
            return False, result.stderr
        if parse_json and result.stdout.strip():
            return True, json.loads(result.stdout)
        return True, result.stdout
    except subprocess.TimeoutExpired:
        return False, "Command timed out"
    except json.JSONDecodeError:
        return True, result.stdout
    except Exception as e:
        return False, str(e)


def get_workspace_prefix(workspace_name: str) -> str:
    """Get the resource prefix for a workspace."""
    return workspace_name


def create_dynamodb_table(workspace_name: str, table_name: str, region: str) -> dict:
    """
    Create a minimal DynamoDB table for integration testing.
    Uses on-demand billing (PAY_PER_REQUEST) for cost efficiency.
    """
    full_table_name = f"{get_workspace_prefix(workspace_name)}-{table_name}"

    print(f"Creating DynamoDB table: {full_table_name}")

    # Create table with simple schema
    success, result = run_aws_command([
        "dynamodb", "create-table",
        "--table-name", full_table_name,
        "--attribute-definitions",
        "AttributeName=pk,AttributeType=S",
        "AttributeName=sk,AttributeType=S",
        "--key-schema",
        "AttributeName=pk,KeyType=HASH",
        "AttributeName=sk,KeyType=RANGE",
        "--billing-mode", "PAY_PER_REQUEST",
        "--region", region,
        "--tags", f"Key=workspace,Value={workspace_name}", "Key=oneenv,Value=true"
    ])

    if not success:
        if "ResourceInUseException" in str(result):
            print(f"Table {full_table_name} already exists")
            return {"table_name": full_table_name, "status": "already_exists"}
        print(f"Failed to create table: {result}")
        return {"error": result}

    # Wait for table to be active
    print("Waiting for table to become active...")
    for _ in range(30):
        success, desc = run_aws_command([
            "dynamodb", "describe-table",
            "--table-name", full_table_name,
            "--region", region
        ])
        if success and desc.get("Table", {}).get("TableStatus") == "ACTIVE":
            print(f"Table {full_table_name} is active")
            return {
                "table_name": full_table_name,
                "arn": desc["Table"]["TableArn"],
                "status": "created"
            }
        time.sleep(2)

    return {"table_name": full_table_name, "status": "creating"}


def create_kinesis_stream(workspace_name: str, stream_name: str, region: str, shard_count: int = 1) -> dict:
    """
    Create a minimal Kinesis stream for integration testing.
    Uses 1 shard by default for minimal cost.
    """
    full_stream_name = f"{get_workspace_prefix(workspace_name)}-{stream_name}"

    print(f"Creating Kinesis stream: {full_stream_name}")

    success, result = run_aws_command([
        "kinesis", "create-stream",
        "--stream-name", full_stream_name,
        "--shard-count", str(shard_count),
        "--region", region,
        "--tags", f"workspace={workspace_name},oneenv=true"
    ], parse_json=False)

    if not success:
        if "ResourceInUseException" in str(result):
            print(f"Stream {full_stream_name} already exists")
            # Get stream details
            success, desc = run_aws_command([
                "kinesis", "describe-stream",
                "--stream-name", full_stream_name,
                "--region", region
            ])
            if success:
                return {
                    "stream_name": full_stream_name,
                    "arn": desc["StreamDescription"]["StreamARN"],
                    "status": "already_exists"
                }
            return {"stream_name": full_stream_name, "status": "already_exists"}
        print(f"Failed to create stream: {result}")
        return {"error": result}

    # Wait for stream to be active
    print("Waiting for stream to become active...")
    for _ in range(30):
        success, desc = run_aws_command([
            "kinesis", "describe-stream",
            "--stream-name", full_stream_name,
            "--region", region
        ])
        if success and desc.get("StreamDescription", {}).get("StreamStatus") == "ACTIVE":
            print(f"Stream {full_stream_name} is active")
            return {
                "stream_name": full_stream_name,
                "arn": desc["StreamDescription"]["StreamARN"],
                "status": "created"
            }
        time.sleep(2)

    return {"stream_name": full_stream_name, "status": "creating"}


def create_rds_instance(workspace_name: str, db_name: str, region: str,
                        vpc_security_group_id: str = None, subnet_group: str = None) -> dict:
    """
    Create a minimal RDS PostgreSQL instance for integration testing.
    Uses db.t3.micro for minimal cost.
    """
    full_db_name = f"{get_workspace_prefix(workspace_name)}-{db_name}"
    # RDS identifiers can't have underscores, replace with hyphens
    full_db_identifier = full_db_name.replace("_", "-")

    print(f"Creating RDS instance: {full_db_identifier}")

    # Generate a random password
    import secrets
    import string
    password = ''.join(secrets.choice(string.ascii_letters + string.digits) for _ in range(16))

    cmd = [
        "rds", "create-db-instance",
        "--db-instance-identifier", full_db_identifier,
        "--db-instance-class", "db.t3.micro",
        "--engine", "postgres",
        "--engine-version", "15",
        "--master-username", "admin",
        "--master-user-password", password,
        "--allocated-storage", "20",
        "--storage-type", "gp2",
        "--no-multi-az",
        "--publicly-accessible",
        "--region", region,
        "--tags", f"Key=workspace,Value={workspace_name}", "Key=oneenv,Value=true"
    ]

    if vpc_security_group_id:
        cmd.extend(["--vpc-security-group-ids", vpc_security_group_id])

    if subnet_group:
        cmd.extend(["--db-subnet-group-name", subnet_group])

    success, result = run_aws_command(cmd)

    if not success:
        if "DBInstanceAlreadyExists" in str(result):
            print(f"RDS instance {full_db_identifier} already exists")
            return {"db_identifier": full_db_identifier, "status": "already_exists"}
        print(f"Failed to create RDS instance: {result}")
        return {"error": result}

    print(f"RDS instance {full_db_identifier} is being created (this takes 5-10 minutes)")
    print(f"Master username: admin")
    print(f"Master password: {password}")
    print("IMPORTANT: Save this password - it won't be shown again!")

    # Store password in Secrets Manager
    secret_name = f"{workspace_name}-rds-{db_name}-credentials"
    secret_value = json.dumps({"username": "admin", "password": password, "dbname": "postgres"})

    run_aws_command([
        "secretsmanager", "create-secret",
        "--name", secret_name,
        "--secret-string", secret_value,
        "--region", region,
        "--tags", f"Key=workspace,Value={workspace_name}", "Key=oneenv,Value=true"
    ])

    return {
        "db_identifier": full_db_identifier,
        "master_username": "admin",
        "master_password": password,
        "secret_name": secret_name,
        "status": "creating",
        "note": "RDS takes 5-10 minutes to become available"
    }


def create_msk_cluster(workspace_name: str, cluster_name: str, region: str,
                       vpc_id: str = None, subnet_ids: list = None, security_group_id: str = None) -> dict:
    """
    Create a minimal MSK Serverless cluster for integration testing.
    MSK Serverless is the most cost-effective option for testing.
    """
    full_cluster_name = f"{get_workspace_prefix(workspace_name)}-{cluster_name}"

    print(f"Creating MSK Serverless cluster: {full_cluster_name}")

    if not subnet_ids or not security_group_id:
        print("Error: MSK requires subnet_ids and security_group_id from the workspace VPC")
        return {"error": "Missing VPC configuration. MSK requires subnet_ids and security_group_id."}

    # Create MSK Serverless cluster
    vpc_config = {
        "SubnetIds": subnet_ids,
        "SecurityGroupIds": [security_group_id]
    }

    cmd = [
        "kafka", "create-cluster-v2",
        "--cluster-name", full_cluster_name,
        "--serverless", json.dumps({
            "VpcConfigs": [vpc_config],
            "ClientAuthentication": {
                "Sasl": {
                    "Iam": {"Enabled": True}
                }
            }
        }),
        "--region", region,
        "--tags", f"workspace={workspace_name},oneenv=true"
    ]

    success, result = run_aws_command(cmd)

    if not success:
        if "ConflictException" in str(result):
            print(f"MSK cluster {full_cluster_name} already exists")
            return {"cluster_name": full_cluster_name, "status": "already_exists"}
        print(f"Failed to create MSK cluster: {result}")
        return {"error": result}

    print(f"MSK Serverless cluster {full_cluster_name} is being created (this takes 10-15 minutes)")

    return {
        "cluster_name": full_cluster_name,
        "cluster_arn": result.get("ClusterArn"),
        "status": "creating",
        "note": "MSK takes 10-15 minutes to become available"
    }


def list_integration_resources(workspace_name: str, region: str) -> dict:
    """List all integration resources for a workspace."""
    prefix = get_workspace_prefix(workspace_name)
    resources = {
        "dynamodb_tables": [],
        "kinesis_streams": [],
        "rds_instances": [],
        "msk_clusters": []
    }

    # List DynamoDB tables
    success, tables = run_aws_command([
        "dynamodb", "list-tables",
        "--region", region
    ])
    if success:
        for table in tables.get("TableNames", []):
            if table.startswith(prefix):
                resources["dynamodb_tables"].append(table)

    # List Kinesis streams
    success, streams = run_aws_command([
        "kinesis", "list-streams",
        "--region", region
    ])
    if success:
        for stream in streams.get("StreamNames", []):
            if stream.startswith(prefix):
                resources["kinesis_streams"].append(stream)

    # List RDS instances
    success, instances = run_aws_command([
        "rds", "describe-db-instances",
        "--region", region
    ])
    if success:
        for instance in instances.get("DBInstances", []):
            if instance["DBInstanceIdentifier"].startswith(prefix.replace("_", "-")):
                resources["rds_instances"].append({
                    "identifier": instance["DBInstanceIdentifier"],
                    "status": instance["DBInstanceStatus"],
                    "endpoint": instance.get("Endpoint", {}).get("Address")
                })

    # List MSK clusters
    success, clusters = run_aws_command([
        "kafka", "list-clusters-v2",
        "--region", region
    ])
    if success:
        for cluster in clusters.get("ClusterInfoList", []):
            if cluster["ClusterName"].startswith(prefix):
                resources["msk_clusters"].append({
                    "name": cluster["ClusterName"],
                    "state": cluster["State"],
                    "arn": cluster["ClusterArn"]
                })

    return resources


def cleanup_integration_resources(workspace_name: str, region: str, resource_type: str = None) -> dict:
    """Delete integration resources for a workspace."""
    prefix = get_workspace_prefix(workspace_name)
    deleted = {
        "dynamodb_tables": [],
        "kinesis_streams": [],
        "rds_instances": [],
        "msk_clusters": [],
        "secrets": []
    }

    # Delete DynamoDB tables
    if not resource_type or resource_type == "dynamodb":
        success, tables = run_aws_command([
            "dynamodb", "list-tables",
            "--region", region
        ])
        if success:
            for table in tables.get("TableNames", []):
                if table.startswith(prefix):
                    print(f"Deleting DynamoDB table: {table}")
                    run_aws_command([
                        "dynamodb", "delete-table",
                        "--table-name", table,
                        "--region", region
                    ], parse_json=False)
                    deleted["dynamodb_tables"].append(table)

    # Delete Kinesis streams
    if not resource_type or resource_type == "kinesis":
        success, streams = run_aws_command([
            "kinesis", "list-streams",
            "--region", region
        ])
        if success:
            for stream in streams.get("StreamNames", []):
                if stream.startswith(prefix):
                    print(f"Deleting Kinesis stream: {stream}")
                    run_aws_command([
                        "kinesis", "delete-stream",
                        "--stream-name", stream,
                        "--region", region
                    ], parse_json=False)
                    deleted["kinesis_streams"].append(stream)

    # Delete RDS instances
    if not resource_type or resource_type == "rds":
        success, instances = run_aws_command([
            "rds", "describe-db-instances",
            "--region", region
        ])
        if success:
            for instance in instances.get("DBInstances", []):
                if instance["DBInstanceIdentifier"].startswith(prefix.replace("_", "-")):
                    print(f"Deleting RDS instance: {instance['DBInstanceIdentifier']}")
                    run_aws_command([
                        "rds", "delete-db-instance",
                        "--db-instance-identifier", instance["DBInstanceIdentifier"],
                        "--skip-final-snapshot",
                        "--delete-automated-backups",
                        "--region", region
                    ], parse_json=False)
                    deleted["rds_instances"].append(instance["DBInstanceIdentifier"])

    # Delete MSK clusters
    if not resource_type or resource_type == "msk":
        success, clusters = run_aws_command([
            "kafka", "list-clusters-v2",
            "--region", region
        ])
        if success:
            for cluster in clusters.get("ClusterInfoList", []):
                if cluster["ClusterName"].startswith(prefix):
                    print(f"Deleting MSK cluster: {cluster['ClusterName']}")
                    run_aws_command([
                        "kafka", "delete-cluster",
                        "--cluster-arn", cluster["ClusterArn"],
                        "--region", region
                    ], parse_json=False)
                    deleted["msk_clusters"].append(cluster["ClusterName"])

    # Delete secrets
    success, secrets = run_aws_command([
        "secretsmanager", "list-secrets",
        "--region", region
    ])
    if success:
        for secret in secrets.get("SecretList", []):
            if secret["Name"].startswith(workspace_name):
                print(f"Deleting secret: {secret['Name']}")
                run_aws_command([
                    "secretsmanager", "delete-secret",
                    "--secret-id", secret["Name"],
                    "--force-delete-without-recovery",
                    "--region", region
                ], parse_json=False)
                deleted["secrets"].append(secret["Name"])

    return deleted


def main():
    parser = argparse.ArgumentParser(description="Integration Resources Manager")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # create-dynamodb
    ddb = subparsers.add_parser("create-dynamodb", help="Create DynamoDB table")
    ddb.add_argument("--workspace", required=True, help="Workspace name")
    ddb.add_argument("--table-name", required=True, help="Table name (will be prefixed)")
    ddb.add_argument("--region", default="us-west-2", help="AWS region")

    # create-kinesis
    kinesis = subparsers.add_parser("create-kinesis", help="Create Kinesis stream")
    kinesis.add_argument("--workspace", required=True, help="Workspace name")
    kinesis.add_argument("--stream-name", required=True, help="Stream name (will be prefixed)")
    kinesis.add_argument("--region", default="us-west-2", help="AWS region")
    kinesis.add_argument("--shards", type=int, default=1, help="Number of shards")

    # create-rds
    rds = subparsers.add_parser("create-rds", help="Create RDS PostgreSQL instance")
    rds.add_argument("--workspace", required=True, help="Workspace name")
    rds.add_argument("--db-name", required=True, help="Database name (will be prefixed)")
    rds.add_argument("--region", default="us-west-2", help="AWS region")
    rds.add_argument("--security-group", help="VPC security group ID")
    rds.add_argument("--subnet-group", help="DB subnet group name")

    # create-msk
    msk = subparsers.add_parser("create-msk", help="Create MSK Serverless cluster")
    msk.add_argument("--workspace", required=True, help="Workspace name")
    msk.add_argument("--cluster-name", required=True, help="Cluster name (will be prefixed)")
    msk.add_argument("--region", default="us-west-2", help="AWS region")
    msk.add_argument("--subnet-ids", help="Comma-separated subnet IDs")
    msk.add_argument("--security-group", help="Security group ID")

    # list
    list_cmd = subparsers.add_parser("list", help="List integration resources")
    list_cmd.add_argument("--workspace", required=True, help="Workspace name")
    list_cmd.add_argument("--region", default="us-west-2", help="AWS region")

    # cleanup
    cleanup = subparsers.add_parser("cleanup", help="Delete integration resources")
    cleanup.add_argument("--workspace", required=True, help="Workspace name")
    cleanup.add_argument("--region", default="us-west-2", help="AWS region")
    cleanup.add_argument("--resource-type", choices=["dynamodb", "kinesis", "rds", "msk"],
                         help="Only delete specific resource type")
    cleanup.add_argument("--confirm", action="store_true", help="Required to actually delete")

    args = parser.parse_args()

    if args.command == "create-dynamodb":
        result = create_dynamodb_table(args.workspace, args.table_name, args.region)
        print(json.dumps(result, indent=2))

    elif args.command == "create-kinesis":
        result = create_kinesis_stream(args.workspace, args.stream_name, args.region, args.shards)
        print(json.dumps(result, indent=2))

    elif args.command == "create-rds":
        result = create_rds_instance(
            args.workspace, args.db_name, args.region,
            args.security_group, args.subnet_group
        )
        print(json.dumps(result, indent=2))

    elif args.command == "create-msk":
        subnet_ids = args.subnet_ids.split(",") if args.subnet_ids else None
        result = create_msk_cluster(
            args.workspace, args.cluster_name, args.region,
            subnet_ids=subnet_ids, security_group_id=args.security_group
        )
        print(json.dumps(result, indent=2))

    elif args.command == "list":
        resources = list_integration_resources(args.workspace, args.region)
        print(json.dumps(resources, indent=2))

    elif args.command == "cleanup":
        if not args.confirm:
            print("Dry run - would delete these resources:")
            resources = list_integration_resources(args.workspace, args.region)
            print(json.dumps(resources, indent=2))
            print("\nRun with --confirm to actually delete")
        else:
            deleted = cleanup_integration_resources(args.workspace, args.region, args.resource_type)
            print("Deleted resources:")
            print(json.dumps(deleted, indent=2))

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
