"""
Lakehouse Template - Full Databricks Lakehouse with Unity Catalog

Usage:
    ~/.vibe/diagrams/.venv/bin/python lakehouse.py

Output:
    lakehouse.png
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.aws.storage import S3
from diagrams.azure.storage import BlobStorage
from diagrams.gcp.storage import GCS
from diagrams.onprem.queue import Kafka
from diagrams.onprem.client import Users
from diagrams.custom import Custom
import os
from glob import glob

def find_icons_dir():
    patterns = [
        os.path.expanduser("~/.claude/plugins/cache/claude-vibe/workflows/*/skills/architecture-diagram/resources/icons"),
        os.path.expanduser("~/code/vibe/plugins/workflows/skills/architecture-diagram/resources/icons"),
    ]
    for pattern in patterns:
        matches = glob(pattern)
        if matches:
            return matches[0]
    raise FileNotFoundError("Could not find icons directory")

ICONS = find_icons_dir()

with Diagram(
    "Databricks Lakehouse Architecture",
    show=False,
    filename="lakehouse",
    outformat="png",
    direction="TB",
    graph_attr={
        "splines": "ortho",
        "nodesep": "0.8",
        "ranksep": "1.5",
        "pad": "0.5",
        "fontsize": "14",
        "bgcolor": "white",
        "dpi": "150"
    }
):
    # Data Sources (top)
    with Cluster("Data Sources"):
        s3 = S3("AWS S3")
        azure_blob = BlobStorage("Azure Blob")
        gcs = GCS("Google GCS")
        kafka = Kafka("Streaming")

    # Unity Catalog (governance layer)
    with Cluster("Unity Catalog - Data Governance"):
        unity = Custom("Unity\nCatalog", f"{ICONS}/databricks/unity_catalog.png")

    # Lakehouse Platform
    with Cluster("Databricks Lakehouse Platform"):
        workspace = Custom("Workspace", f"{ICONS}/databricks/workspace.png")

        with Cluster("Delta Lake Storage"):
            bronze = Custom("Bronze", f"{ICONS}/databricks/delta_lake.png")
            silver = Custom("Silver", f"{ICONS}/databricks/delta_lake.png")
            gold = Custom("Gold", f"{ICONS}/databricks/delta_lake.png")

        with Cluster("Compute"):
            sql_warehouse = Custom("SQL\nWarehouse", f"{ICONS}/databricks/sql_warehouse.png")
            model_serving = Custom("Model\nServing", f"{ICONS}/databricks/model_serving.png")

    # Consumers (bottom)
    with Cluster("Data Consumers"):
        analysts = Users("Analysts")
        apps = Users("Applications")

    # Connections
    [s3, azure_blob, gcs, kafka] >> bronze
    bronze >> silver >> gold

    unity >> workspace

    gold >> sql_warehouse >> analysts
    gold >> model_serving >> apps

print(f"Generated: lakehouse.png")
