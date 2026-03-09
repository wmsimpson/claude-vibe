"""
AI/BI Template - Genie + Lakeview Dashboards

Usage:
    ~/.vibe/diagrams/.venv/bin/python ai_bi.py

Output:
    ai_bi.png
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.onprem.client import Users
from diagrams.aws.storage import S3
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
    "AI/BI - Self-Service Analytics",
    show=False,
    filename="ai_bi",
    outformat="png",
    direction="TB",
    graph_attr={
        "splines": "ortho",
        "nodesep": "1.2",
        "ranksep": "1.5",
        "pad": "0.5",
        "fontsize": "14",
        "bgcolor": "white",
        "dpi": "150"
    }
):
    # Users
    with Cluster("Business Users"):
        analysts = Users("Analysts")
        executives = Users("Executives")
        ops = Users("Operations")

    # AI/BI Layer
    with Cluster("AI/BI - Databricks Intelligence"):
        with Cluster("Natural Language"):
            genie = Custom("Genie\nRooms", f"{ICONS}/databricks/workspace.png")

        with Cluster("Visualization"):
            dashboards = Custom("Lakeview\nDashboards", f"{ICONS}/databricks/sql_warehouse.png")

        with Cluster("Alerts"):
            alerts = Custom("Alerts &\nSchedules", f"{ICONS}/databricks/workspace.png")

    # Query Layer
    with Cluster("Databricks SQL"):
        warehouse = Custom("SQL\nWarehouse", f"{ICONS}/databricks/sql_warehouse.png")

    # Governance
    with Cluster("Unity Catalog"):
        unity = Custom("Data\nGovernance", f"{ICONS}/databricks/unity_catalog.png")
        lineage = Custom("Lineage &\nAudit", f"{ICONS}/databricks/unity_catalog.png")

    # Data Layer
    with Cluster("Lakehouse"):
        with Cluster("Curated Data"):
            gold = Custom("Gold\nTables", f"{ICONS}/databricks/delta_lake.png")
            semantic = Custom("Semantic\nModels", f"{ICONS}/databricks/delta_lake.png")

    # User flows
    analysts >> genie
    analysts >> dashboards
    executives >> dashboards
    ops >> alerts

    # Data flow
    genie >> warehouse
    dashboards >> warehouse
    alerts >> warehouse

    warehouse >> unity >> gold
    warehouse >> unity >> semantic

    unity >> lineage

print(f"Generated: ai_bi.png")
