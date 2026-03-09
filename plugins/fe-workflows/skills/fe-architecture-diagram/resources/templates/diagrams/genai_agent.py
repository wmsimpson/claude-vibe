"""
GenAI & Agent Architecture Template - RAG, Agents, MLflow Tracing

Usage:
    ~/.vibe/diagrams/.venv/bin/python genai_agent.py

Output:
    genai_agent.png
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.onprem.database import PostgreSQL
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
    "GenAI Agent Architecture",
    show=False,
    filename="genai_agent",
    outformat="png",
    direction="TB",
    graph_attr={
        "splines": "ortho",
        "nodesep": "1.0",
        "ranksep": "1.5",
        "pad": "0.5",
        "fontsize": "14",
        "bgcolor": "white",
        "dpi": "150"
    }
):
    # Users
    users = Users("End Users")

    # Application Layer
    with Cluster("Databricks App"):
        app = Custom("React/FastAPI\nApp", f"{ICONS}/databricks/workspace.png")
        lakebase = PostgreSQL("Lakebase\n(App State)")

    # Agent Layer
    with Cluster("Mosaic AI Agent Framework"):
        agent = Custom("AI Agent", f"{ICONS}/databricks/model_serving.png")

        with Cluster("Agent Tools"):
            tool_sql = Custom("SQL\nTool", f"{ICONS}/databricks/sql_warehouse.png")
            tool_search = Custom("Vector\nSearch", f"{ICONS}/databricks/unity_catalog.png")
            tool_api = Custom("API\nTool", f"{ICONS}/databricks/workspace.png")

    # LLM Layer
    with Cluster("Foundation Models"):
        llm = Custom("Foundation\nModel API", f"{ICONS}/databricks/model_serving.png")
        embeddings = Custom("Embedding\nModel", f"{ICONS}/databricks/model_serving.png")

    # Data Layer
    with Cluster("Knowledge Base"):
        with Cluster("Vector Store"):
            vectors = Custom("Vector\nIndex", f"{ICONS}/databricks/delta_lake.png")

        with Cluster("Structured Data"):
            gold = Custom("Gold\nTables", f"{ICONS}/databricks/delta_lake.png")

        with Cluster("Documents"):
            docs = S3("Unstructured\nDocs")

    # MLflow Observability
    with Cluster("MLflow GenAI"):
        tracing = Custom("Tracing", f"{ICONS}/databricks/unity_catalog.png")
        eval_suite = Custom("Evaluation", f"{ICONS}/databricks/workspace.png")

    # Connections
    users >> app
    app >> lakebase
    app >> agent

    agent >> tool_sql >> gold
    agent >> tool_search >> vectors
    agent >> tool_api

    agent >> llm
    docs >> embeddings >> vectors

    # Observability
    agent >> Edge(style="dashed", label="traces") >> tracing
    tracing >> eval_suite

print(f"Generated: genai_agent.png")
