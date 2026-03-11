"""
Azure Lakehouse Template - Databricks on Azure with native Azure services

Usage:
    ~/.vibe/diagrams/.venv/bin/python azure_lakehouse.py

Output:
    azure_lakehouse.png
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.azure.storage import DataLakeStorage, BlobStorage
from diagrams.azure.analytics import Databricks as AzureDatabricks, EventHubs, StreamAnalytics, SynapseAnalytics, DataFactory
from diagrams.azure.database import SQLDatabase, CosmosDB
from diagrams.azure.integration import ServiceBus, EventGrid
from diagrams.azure.security import KeyVault
from diagrams.azure.identity import ManagedIdentities
from diagrams.azure.network import VirtualNetworks
from diagrams.azure.general import Resourcegroups
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
    "Databricks Lakehouse on Azure",
    show=False,
    filename="azure_lakehouse",
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
            eventhubs = EventHubs("Event Hubs")
            servicebus = ServiceBus("Service Bus")

        with Cluster("Databases"):
            sqldb = SQLDatabase("Azure SQL\n(CDC)")
            cosmos = CosmosDB("Cosmos DB")

        with Cluster("Orchestration"):
            adf = DataFactory("Data Factory")
            eventgrid = EventGrid("Event Grid")

    # Azure Infrastructure
    with Cluster("Azure Infrastructure"):
        vnet = VirtualNetworks("VNet")
        identity = ManagedIdentities("Managed\nIdentity")
        keyvault = KeyVault("Key Vault")

    # Databricks Platform
    with Cluster("Azure Databricks"):
        workspace = Custom("Databricks\nWorkspace", f"{ICONS}/databricks/workspace.png")
        unity = Custom("Unity\nCatalog", f"{ICONS}/databricks/unity_catalog.png")

        with Cluster("Ingestion"):
            autoloader = Custom("Auto Loader", f"{ICONS}/databricks/delta_lake.png")
            lakeflow = Custom("Lakeflow\nConnect", f"{ICONS}/databricks/workspace.png")

        with Cluster("Lakehouse Storage (ADLS Gen2)"):
            bronze = Custom("Bronze", f"{ICONS}/databricks/delta_lake.png")
            silver = Custom("Silver", f"{ICONS}/databricks/delta_lake.png")
            gold = Custom("Gold", f"{ICONS}/databricks/delta_lake.png")
            adls = DataLakeStorage("ADLS Gen2")

        with Cluster("Compute"):
            warehouse = Custom("SQL\nWarehouse", f"{ICONS}/databricks/sql_warehouse.png")
            serving = Custom("Model\nServing", f"{ICONS}/databricks/model_serving.png")

    # Consumption
    with Cluster("Consumption"):
        users = Users("Analysts")
        with Cluster("Azure Analytics (Optional)"):
            synapse = SynapseAnalytics("Synapse\nServerless")

    # Data Flow
    eventhubs >> autoloader
    servicebus >> autoloader
    sqldb >> lakeflow
    cosmos >> lakeflow
    adf >> autoloader
    eventgrid >> autoloader

    autoloader >> bronze
    lakeflow >> bronze
    bronze >> silver >> gold

    [bronze, silver, gold] >> adls

    unity >> workspace
    identity >> workspace
    keyvault >> workspace

    gold >> warehouse >> users
    gold >> serving
    gold >> Edge(style="dashed", label="Delta Sharing") >> synapse

print(f"Generated: azure_lakehouse.png")
