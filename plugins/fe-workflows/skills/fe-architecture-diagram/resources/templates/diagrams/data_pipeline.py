"""
Data Pipeline Template - Kafka → Bronze → Silver → Gold → Snowflake

Usage:
    ~/.vibe/diagrams/.venv/bin/python data_pipeline.py

Output:
    data_pipeline.png
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.onprem.queue import Kafka
from diagrams.aws.storage import S3
from diagrams.custom import Custom
import os
from glob import glob

# Find the icon directory (handles version wildcards in plugin cache path)
def find_icons_dir():
    """Find the icons directory in the plugin cache."""
    patterns = [
        os.path.expanduser("~/.claude/plugins/cache/fe-vibe/fe-workflows/*/skills/fe-architecture-diagram/resources/icons"),
        os.path.expanduser("~/code/vibe/plugins/fe-workflows/skills/fe-architecture-diagram/resources/icons"),
    ]
    for pattern in patterns:
        matches = glob(pattern)
        if matches:
            return matches[0]
    raise FileNotFoundError("Could not find icons directory. Ensure the skill is installed.")

ICONS = find_icons_dir()

with Diagram(
    "Data Pipeline - Medallion Architecture",
    show=False,
    filename="data_pipeline",
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
    # Data Sources
    with Cluster("Data Sources"):
        kafka = Kafka("Kafka\nStreams")
        s3_raw = S3("S3 Raw\nFiles")

    # Databricks Lakehouse - Medallion Architecture
    with Cluster("Databricks Lakehouse"):
        with Cluster("Bronze Layer"):
            bronze = Custom("Raw\nIngestion", f"{ICONS}/databricks/delta_lake.png")

        with Cluster("Silver Layer"):
            silver = Custom("Cleaned &\nValidated", f"{ICONS}/databricks/delta_lake.png")

        with Cluster("Gold Layer"):
            gold = Custom("Business\nAggregates", f"{ICONS}/databricks/delta_lake.png")

    # Destinations
    with Cluster("Consumption"):
        snowflake = Custom("Snowflake\nAnalytics", f"{ICONS}/cloud/snowflake.png")
        warehouse = Custom("SQL\nWarehouse", f"{ICONS}/databricks/sql_warehouse.png")

    # Data flow with labels
    kafka >> Edge(label="streaming") >> bronze
    s3_raw >> Edge(label="batch") >> bronze
    bronze >> Edge(label="cleanse") >> silver
    silver >> Edge(label="aggregate") >> gold
    gold >> Edge(label="share") >> snowflake
    gold >> Edge(label="query") >> warehouse

print(f"Generated: data_pipeline.png")
