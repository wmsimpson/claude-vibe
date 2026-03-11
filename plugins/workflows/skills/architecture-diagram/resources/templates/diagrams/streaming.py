"""
Streaming Template - Real-time Streaming Architecture

Usage:
    ~/.vibe/diagrams/.venv/bin/python streaming.py

Output:
    streaming.png
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.onprem.queue import Kafka
from diagrams.aws.analytics import Kinesis
from diagrams.gcp.analytics import Pubsub
from diagrams.onprem.database import PostgreSQL
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
    "Real-Time Streaming Architecture",
    show=False,
    filename="streaming",
    outformat="png",
    direction="LR",
    graph_attr={
        "splines": "ortho",
        "nodesep": "1.0",
        "ranksep": "2.0",
        "pad": "0.5",
        "fontsize": "14",
        "bgcolor": "white",
        "dpi": "150"
    }
):
    # Event Sources
    with Cluster("Event Sources"):
        iot = Users("IoT Devices")
        apps = Users("Applications")
        cdc = PostgreSQL("CDC Source")

    # Streaming Ingestion
    with Cluster("Message Brokers"):
        kafka = Kafka("Kafka")
        confluent = Custom("Confluent\nCloud", f"{ICONS}/cloud/confluent.png")
        kinesis = Kinesis("Kinesis")

    # Stream Processing
    with Cluster("Databricks Streaming"):
        workspace = Custom("Databricks\nWorkspace", f"{ICONS}/databricks/workspace.png")

        with Cluster("Structured Streaming"):
            ingest = Custom("Auto\nLoader", f"{ICONS}/databricks/delta_lake.png")
            transform = Custom("Stream\nProcessing", f"{ICONS}/databricks/delta_lake.png")
            sink = Custom("Delta\nLive Tables", f"{ICONS}/databricks/delta_lake.png")

    # Real-time Serving
    with Cluster("Real-time Serving"):
        model = Custom("Model\nServing", f"{ICONS}/databricks/model_serving.png")
        warehouse = Custom("SQL\nWarehouse", f"{ICONS}/databricks/sql_warehouse.png")
        dashboard = Users("Live\nDashboard")

    # Event flow
    [iot, apps] >> kafka
    cdc >> confluent

    kafka >> ingest
    confluent >> ingest
    kinesis >> ingest

    ingest >> transform >> sink

    sink >> model >> dashboard
    sink >> warehouse >> dashboard

print(f"Generated: streaming.png")
