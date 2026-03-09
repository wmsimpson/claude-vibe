"""
Multi-Cloud Template - AWS + GCP + Azure Integration

Usage:
    ~/.vibe/diagrams/.venv/bin/python multi_cloud.py

Output:
    multi_cloud.png
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.aws.compute import EC2
from diagrams.aws.storage import S3
from diagrams.aws.database import Redshift
from diagrams.aws.analytics import Glue, Kinesis
from diagrams.gcp.compute import GCE
from diagrams.gcp.storage import GCS
from diagrams.gcp.database import BigQuery
from diagrams.gcp.analytics import Pubsub, Dataflow
from diagrams.azure.compute import VM
from diagrams.azure.storage import BlobStorage
from diagrams.azure.database import SynapseAnalytics
from diagrams.azure.analytics import EventHubs
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
    "Multi-Cloud Data Architecture",
    show=False,
    filename="multi_cloud",
    outformat="png",
    direction="TB",
    graph_attr={
        "splines": "polyline",
        "nodesep": "1.0",
        "ranksep": "1.5",
        "pad": "0.5",
        "fontsize": "14",
        "bgcolor": "white",
        "dpi": "150"
    }
):
    # AWS Region
    with Cluster("AWS"):
        s3 = S3("S3 Data Lake")
        kinesis = Kinesis("Kinesis")
        glue = Glue("Glue ETL")
        redshift = Redshift("Redshift")

        kinesis >> glue >> s3

    # GCP Region
    with Cluster("Google Cloud"):
        gcs = GCS("Cloud Storage")
        pubsub = Pubsub("Pub/Sub")
        dataflow = Dataflow("Dataflow")
        bigquery = BigQuery("BigQuery")

        pubsub >> dataflow >> gcs

    # Azure Region
    with Cluster("Azure"):
        blob = BlobStorage("Blob Storage")
        eventhubs = EventHubs("Event Hubs")
        synapse = SynapseAnalytics("Synapse")

        eventhubs >> synapse >> blob

    # Central Databricks
    with Cluster("Databricks - Unified Analytics"):
        workspace = Custom("Databricks\nWorkspace", f"{ICONS}/databricks/workspace.png")
        unity = Custom("Unity\nCatalog", f"{ICONS}/databricks/unity_catalog.png")
        delta = Custom("Delta\nLake", f"{ICONS}/databricks/delta_lake.png")

    # Connections from each cloud to Databricks
    s3 >> Edge(label="AWS") >> delta
    gcs >> Edge(label="GCP") >> delta
    blob >> Edge(label="Azure") >> delta

    unity >> workspace
    delta >> workspace

    # Outbound to cloud warehouses
    workspace >> redshift
    workspace >> bigquery
    workspace >> synapse

print(f"Generated: multi_cloud.png")
