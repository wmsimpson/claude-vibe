"""
GCP Lakehouse Template - Databricks on GCP with native GCP services

Usage:
    ~/.vibe/diagrams/.venv/bin/python gcp_lakehouse.py

Output:
    gcp_lakehouse.png
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.gcp.storage import GCS
from diagrams.gcp.analytics import PubSub, Dataflow, BigQuery, Dataproc, Composer
from diagrams.gcp.database import SQL as CloudSQL, Spanner, Firestore
from diagrams.gcp.compute import Functions
from diagrams.gcp.security import IAM as GCPIAM, KMS
from diagrams.gcp.network import VPC as GCPVPC
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
    "Databricks Lakehouse on GCP",
    show=False,
    filename="gcp_lakehouse",
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
            pubsub = PubSub("Pub/Sub")
            dataflow_src = Dataflow("Dataflow")

        with Cluster("Databases"):
            cloudsql = CloudSQL("Cloud SQL\n(CDC)")
            spanner = Spanner("Spanner")
            firestore = Firestore("Firestore")

        with Cluster("Orchestration"):
            composer = Composer("Cloud\nComposer")
            functions = Functions("Cloud\nFunctions")

    # GCP Infrastructure
    with Cluster("GCP Infrastructure"):
        vpc = GCPVPC("VPC")
        iam = GCPIAM("IAM &\nService Accounts")
        kms = KMS("Cloud KMS")

    # Databricks Platform
    with Cluster("Databricks on GCP"):
        workspace = Custom("Databricks\nWorkspace", f"{ICONS}/databricks/workspace.png")
        unity = Custom("Unity\nCatalog", f"{ICONS}/databricks/unity_catalog.png")

        with Cluster("Ingestion"):
            autoloader = Custom("Auto Loader", f"{ICONS}/databricks/delta_lake.png")
            lakeflow = Custom("Lakeflow\nConnect", f"{ICONS}/databricks/workspace.png")

        with Cluster("Lakehouse Storage (GCS)"):
            bronze = Custom("Bronze", f"{ICONS}/databricks/delta_lake.png")
            silver = Custom("Silver", f"{ICONS}/databricks/delta_lake.png")
            gold = Custom("Gold", f"{ICONS}/databricks/delta_lake.png")
            gcs = GCS("GCS Bucket")

        with Cluster("Compute"):
            warehouse = Custom("SQL\nWarehouse", f"{ICONS}/databricks/sql_warehouse.png")
            serving = Custom("Model\nServing", f"{ICONS}/databricks/model_serving.png")

    # Consumption
    with Cluster("Consumption"):
        users = Users("Analysts")
        with Cluster("GCP Analytics (Optional)"):
            bigquery = BigQuery("BigQuery")

    # Data Flow
    pubsub >> autoloader
    dataflow_src >> autoloader
    cloudsql >> lakeflow
    spanner >> lakeflow
    firestore >> lakeflow
    composer >> autoloader
    functions >> autoloader

    autoloader >> bronze
    lakeflow >> bronze
    bronze >> silver >> gold

    [bronze, silver, gold] >> gcs

    unity >> workspace
    iam >> workspace
    kms >> workspace

    gold >> warehouse >> users
    gold >> serving
    gold >> Edge(style="dashed", label="BigLake/Delta Sharing") >> bigquery

print(f"Generated: gcp_lakehouse.png")
