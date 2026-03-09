# Template Customization Guide

How to quickly customize templates for specific customer architectures.

---

## Quick Start: Swapping Components

Templates are Python code. To customize:
1. Copy the template to your working directory
2. Change imports and node definitions
3. Run to generate your diagram

### Example: Kinesis → MSK (Kafka)

**Before:**
```python
from diagrams.aws.analytics import Kinesis

kinesis = Kinesis("Kinesis\nData Streams")
kinesis >> autoloader
```

**After:**
```python
from diagrams.onprem.queue import Kafka

msk = Kafka("MSK\n(Kafka)")
msk >> autoloader
```

---

## Common Component Swaps

### Streaming / Messaging

| If customer uses... | Import | Node |
|---------------------|--------|------|
| AWS Kinesis | `from diagrams.aws.analytics import Kinesis` | `Kinesis("Kinesis")` |
| AWS MSK (Kafka) | `from diagrams.onprem.queue import Kafka` | `Kafka("MSK")` |
| Confluent Cloud | `Custom("Confluent", f"{ICONS}/cloud/confluent.png")` | Use custom icon |
| Azure Event Hubs | `from diagrams.azure.analytics import EventHubs` | `EventHubs("Event Hubs")` |
| GCP Pub/Sub | `from diagrams.gcp.analytics import PubSub` | `PubSub("Pub/Sub")` |
| Self-hosted Kafka | `from diagrams.onprem.queue import Kafka` | `Kafka("Kafka")` |

### Databases (CDC Sources)

| If customer uses... | Import | Node |
|---------------------|--------|------|
| AWS RDS | `from diagrams.aws.database import RDS` | `RDS("RDS MySQL")` |
| AWS Aurora | `from diagrams.aws.database import Aurora` | `Aurora("Aurora")` |
| Azure SQL | `from diagrams.azure.database import SQLDatabase` | `SQLDatabase("Azure SQL")` |
| GCP Cloud SQL | `from diagrams.gcp.database import SQL` | `SQL("Cloud SQL")` |
| PostgreSQL (on-prem) | `from diagrams.onprem.database import PostgreSQL` | `PostgreSQL("Postgres")` |
| MySQL (on-prem) | `from diagrams.onprem.database import MySQL` | `MySQL("MySQL")` |
| Oracle | `from diagrams.onprem.database import Oracle` | `Oracle("Oracle")` |
| MongoDB | `from diagrams.onprem.database import MongoDB` | `MongoDB("MongoDB")` |

### Data Warehouses (Destinations)

| If customer uses... | Import | Node |
|---------------------|--------|------|
| Snowflake | `Custom("Snowflake", f"{ICONS}/cloud/snowflake.png")` | Use custom icon |
| AWS Redshift | `from diagrams.aws.analytics import Redshift` | `Redshift("Redshift")` |
| GCP BigQuery | `from diagrams.gcp.database import BigQuery` | `BigQuery("BigQuery")` |
| Azure Synapse | `from diagrams.azure.analytics import SynapseAnalytics` | `SynapseAnalytics("Synapse")` |

### Orchestration

| If customer uses... | Import | Node |
|---------------------|--------|------|
| Databricks Workflows | Part of workspace | Usually implicit |
| Apache Airflow | `from diagrams.onprem.workflow import Airflow` | `Airflow("Airflow")` |
| AWS Step Functions | `from diagrams.aws.integration import StepFunctions` | `StepFunctions("Step Functions")` |
| Azure Data Factory | `from diagrams.azure.analytics import DataFactory` | `DataFactory("Data Factory")` |
| GCP Composer | `from diagrams.gcp.analytics import Composer` | `Composer("Composer")` |
| Prefect | `from diagrams.onprem.workflow import Airflow` | `Airflow("Prefect")` (no icon) |
| Dagster | `from diagrams.onprem.workflow import Airflow` | `Airflow("Dagster")` (no icon) |

### BI Tools

| If customer uses... | Import | Node |
|---------------------|--------|------|
| Lakeview Dashboards | Use custom Databricks icon | `Custom("Lakeview", ...)` |
| Tableau | `from diagrams.onprem.client import Users` | `Users("Tableau")` |
| Power BI | `from diagrams.onprem.client import Users` | `Users("Power BI")` |
| Looker | `from diagrams.onprem.client import Users` | `Users("Looker")` |
| Sigma | `from diagrams.onprem.client import Users` | `Users("Sigma")` |

### ML Platforms

| If customer uses... | Import | Node |
|---------------------|--------|------|
| Databricks Model Serving | Use custom icon | `Custom("Model Serving", ...)` |
| AWS SageMaker | `from diagrams.aws.ml import Sagemaker` | `Sagemaker("SageMaker")` |
| GCP Vertex AI | `from diagrams.gcp.ml import AIPlatform` | `AIPlatform("Vertex AI")` |
| Azure ML | `from diagrams.azure.ml import MachineLearningServiceWorkspaces` | Long name, alias it |

---

## Adding/Removing Components

### Adding a Component

1. Add the import at the top
2. Create the node inside the appropriate Cluster
3. Connect it with `>>` arrows

```python
# Add Airflow orchestration
from diagrams.onprem.workflow import Airflow

with Cluster("Orchestration"):
    airflow = Airflow("Airflow")

# Connect it
airflow >> bronze  # Airflow triggers Bronze ingestion
```

### Removing a Component

1. Delete the node definition
2. Delete any arrows (`>>`) that reference it
3. Remove the import if no longer used

### Renaming a Component

Just change the label string:
```python
# Before
kinesis = Kinesis("Kinesis\nData Streams")

# After - same icon, different label
kinesis = Kinesis("Real-time\nEvents")
```

---

## Changing Layout

### Horizontal vs Vertical

```python
# Vertical (top to bottom) - good for pipelines
direction="TB"

# Horizontal (left to right) - good for data flow
direction="LR"
```

### Adjusting Spacing

```python
graph_attr={
    "nodesep": "1.5",    # More horizontal space
    "ranksep": "2.0",    # More vertical space
}
```

### Different Arrow Styles

```python
# Straight 90-degree angles (cleanest)
graph_attr={"splines": "ortho"}

# Curved arrows
graph_attr={"splines": "curved"}

# Straight lines (can overlap)
graph_attr={"splines": "polyline"}
```

---

## Adding Custom Labels to Arrows

```python
# Simple arrow
kafka >> bronze

# Arrow with label
kafka >> Edge(label="streaming") >> bronze

# Styled arrow
kafka >> Edge(label="real-time", color="blue", style="bold") >> bronze

# Dashed arrow (for optional/future connections)
gold >> Edge(style="dashed", label="planned") >> snowflake
```

---

## Multi-Environment Diagrams

Show dev/staging/prod in one diagram:

```python
with Cluster("Development"):
    dev_bronze = Custom("Bronze", icon)

with Cluster("Production"):
    prod_bronze = Custom("Bronze", icon)

dev_bronze >> Edge(label="promote", style="dashed") >> prod_bronze
```

---

## Hybrid / Migration Diagrams

Show current state vs future state:

```python
with Cluster("Current State (Legacy)"):
    old_warehouse = Redshift("Redshift")

with Cluster("Future State (Lakehouse)"):
    new_lakehouse = Custom("Delta Lake", icon)

old_warehouse >> Edge(label="migrate", style="dashed", color="green") >> new_lakehouse
```

---

## Quick Template Modification Checklist

When customizing a template for a customer:

- [ ] Update diagram title: `"Customer Name - Architecture"`
- [ ] Swap streaming source (Kinesis/Kafka/Event Hubs)
- [ ] Swap database sources (RDS/Aurora/Cloud SQL)
- [ ] Add/remove orchestration tool
- [ ] Add/remove destination warehouses
- [ ] Update cluster names to match customer terminology
- [ ] Adjust spacing if diagram looks cramped
- [ ] Add customer-specific components (custom APIs, internal tools)
