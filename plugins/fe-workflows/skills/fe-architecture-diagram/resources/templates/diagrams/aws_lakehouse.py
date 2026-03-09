"""
AWS Lakehouse Template - Databricks on AWS with native AWS services

Usage:
    ~/.vibe/diagrams/.venv/bin/python aws_lakehouse.py

Output:
    aws_lakehouse.png
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.aws.storage import S3
from diagrams.aws.analytics import Kinesis, KinesisDataFirehose, Glue, Athena, Redshift, EMR
from diagrams.aws.database import RDS, DynamoDB
from diagrams.aws.integration import SQS, SNS, Eventbridge
from diagrams.aws.security import IAM
from diagrams.aws.network import VPC, PrivateSubnet
from diagrams.aws.compute import Lambda
from diagrams.onprem.client import Users
from diagrams.custom import Custom
import os
from glob import glob

def find_icons_dir():
    patterns = [
        os.path.expanduser("~/.claude/plugins/cache/fe-vibe/fe-workflows/*/skills/fe-architecture-diagram/resources/icons"),
        os.path.expanduser("~/code/vibe/plugins/fe-workflows/skills/fe-architecture-diagram/resources/icons"),
    ]
    for pattern in patterns:
        matches = glob(pattern)
        if matches:
            return matches[0]
    raise FileNotFoundError("Could not find icons directory")

ICONS = find_icons_dir()

with Diagram(
    "Databricks Lakehouse on AWS",
    show=False,
    filename="aws_lakehouse",
    outformat="png",
    direction="TB",
    graph_attr={
        "splines": "ortho",
        "nodesep": "0.8",
        "ranksep": "1.2",
        "pad": "0.5",
        "fontsize": "14",
        "bgcolor": "white",
        "dpi": "150"
    }
):
    # Data Sources
    with Cluster("Data Sources"):
        with Cluster("Streaming"):
            kinesis = Kinesis("Kinesis\nData Streams")
            msk = Kinesis("MSK\n(Kafka)")

        with Cluster("Databases"):
            rds = RDS("RDS\n(CDC)")
            dynamo = DynamoDB("DynamoDB")

        with Cluster("Events"):
            eventbridge = Eventbridge("EventBridge")
            sqs = SQS("SQS")

    # AWS Infrastructure
    with Cluster("AWS Infrastructure"):
        vpc = VPC("VPC")
        iam = IAM("IAM Roles\n& Instance Profiles")

    # Databricks Platform
    with Cluster("Databricks on AWS"):
        workspace = Custom("Databricks\nWorkspace", f"{ICONS}/databricks/workspace.png")
        unity = Custom("Unity\nCatalog", f"{ICONS}/databricks/unity_catalog.png")

        with Cluster("Ingestion"):
            autoloader = Custom("Auto Loader", f"{ICONS}/databricks/delta_lake.png")
            lakeflow = Custom("Lakeflow\nConnect", f"{ICONS}/databricks/workspace.png")

        with Cluster("Lakehouse Storage (S3)"):
            bronze = Custom("Bronze", f"{ICONS}/databricks/delta_lake.png")
            silver = Custom("Silver", f"{ICONS}/databricks/delta_lake.png")
            gold = Custom("Gold", f"{ICONS}/databricks/delta_lake.png")
            s3 = S3("S3 Bucket")

        with Cluster("Compute"):
            warehouse = Custom("SQL\nWarehouse", f"{ICONS}/databricks/sql_warehouse.png")
            serving = Custom("Model\nServing", f"{ICONS}/databricks/model_serving.png")

    # Consumption
    with Cluster("Consumption"):
        users = Users("Analysts")
        with Cluster("AWS Analytics (Optional)"):
            redshift = Redshift("Redshift\nSpectrum")
            athena = Athena("Athena")

    # Data Flow
    kinesis >> autoloader
    msk >> autoloader
    rds >> lakeflow
    dynamo >> lakeflow
    eventbridge >> autoloader
    sqs >> autoloader

    autoloader >> bronze
    lakeflow >> bronze
    bronze >> silver >> gold

    [bronze, silver, gold] >> s3

    unity >> workspace

    gold >> warehouse >> users
    gold >> serving
    gold >> Edge(style="dashed", label="Delta Sharing") >> redshift
    gold >> Edge(style="dashed", label="manifest") >> athena

print(f"Generated: aws_lakehouse.png")
