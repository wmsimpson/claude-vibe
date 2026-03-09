# Vendor Icon Reference

Complete reference for available icons when creating architecture diagrams.

## Icon Coverage Summary

| Provider | Built-in Icons | Custom Icons Needed |
|----------|---------------|---------------------|
| **AWS** | ~150 (EC2, S3, Lambda, RDS, Kinesis, Glue, etc.) | None |
| **GCP** | ~80 (GCE, GCS, BigQuery, Pub/Sub, Dataflow, etc.) | None |
| **Azure** | ~80 (VMs, Blob, Synapse, Event Hubs, Databricks!) | None |
| **Kubernetes** | ~30 (Pods, Services, Deployments, etc.) | None |
| **On-Prem** | ~50 (Kafka, Spark, Airflow, PostgreSQL, etc.) | None |
| **Databricks** | 1 (Azure Databricks only) | 6 (Workspace, Unity Catalog, Delta Lake, etc.) |
| **Other** | - | 4 (Snowflake, Confluent, alt Kafka/Airflow) |

**Bottom line**: The library has 300+ icons. Custom icons are only needed for Databricks-specific products and Snowflake.

---

## Built-in Icons (mingrammer/diagrams) - 300+ Available

### AWS Icons

```python
from diagrams.aws.compute import EC2, Lambda, ECS, EKS, Fargate, Batch
from diagrams.aws.database import RDS, DynamoDB, Redshift, Aurora, ElastiCache, Neptune
from diagrams.aws.storage import S3, EFS, FSx, Glacier
from diagrams.aws.analytics import Glue, Athena, EMR, Kinesis, KinesisDataStreams, KinesisDataFirehose, QuickSight, LakeFormation
from diagrams.aws.integration import SQS, SNS, EventBridge, StepFunctions, MQ
from diagrams.aws.network import VPC, CloudFront, Route53, ELB, ALB, NLB, APIGateway
from diagrams.aws.security import IAM, KMS, SecretsManager, Cognito
from diagrams.aws.ml import Sagemaker, SagemakerModel, SagemakerNotebook
```

### GCP Icons

```python
from diagrams.gcp.compute import GCE, GKE, Functions as CloudFunctions, Run as CloudRun, AppEngine
from diagrams.gcp.database import BigQuery, Spanner, SQL as CloudSQL, Firestore, Bigtable
from diagrams.gcp.storage import GCS
from diagrams.gcp.analytics import Dataflow, Dataproc, PubSub, Composer, DataFusion
from diagrams.gcp.network import LoadBalancing, CDN, DNS, VPC as GCPVPC
from diagrams.gcp.ml import AIPlatform, AutoML, NaturalLanguageAPI, VisionAPI
```

### Azure Icons

```python
from diagrams.azure.compute import VM, AKS, FunctionApps, ContainerInstances, AppServices
from diagrams.azure.database import SQLDatabase, CosmosDB, SynapseAnalytics, SQLManagedInstances
from diagrams.azure.storage import BlobStorage, DataLakeStorage, StorageAccounts
from diagrams.azure.analytics import Databricks, DataFactory, EventHubs, StreamAnalytics, HDInsight
from diagrams.azure.network import LoadBalancers, ApplicationGateway, VirtualNetworks
from diagrams.azure.ml import MachineLearningServiceWorkspaces
```

### Kubernetes Icons

```python
from diagrams.k8s.compute import Pod, Deployment, StatefulSet, DaemonSet, Job, CronJob, ReplicaSet
from diagrams.k8s.network import Service, Ingress, NetworkPolicy
from diagrams.k8s.storage import PV, PVC, StorageClass
from diagrams.k8s.group import Namespace
from diagrams.k8s.controlplane import APIServer, Scheduler, ControllerManager
```

### On-Premises / Open Source

```python
# Messaging & Streaming
from diagrams.onprem.queue import Kafka, RabbitMQ, ActiveMQ, Celery

# Analytics & Processing
from diagrams.onprem.analytics import Spark, Flink, Presto, Trino, Hive, Beam

# Workflow & Orchestration
from diagrams.onprem.workflow import Airflow, NiFi, Kubeflow

# Databases
from diagrams.onprem.database import PostgreSQL, MySQL, MongoDB, Cassandra, Redis, Elasticsearch, InfluxDB, Neo4J

# CI/CD
from diagrams.onprem.ci import Jenkins, GitlabCI, GithubActions, CircleCI, TravisCI

# Monitoring
from diagrams.onprem.monitoring import Prometheus, Grafana, Datadog, Splunk

# Containers
from diagrams.onprem.container import Docker

# Clients / Users
from diagrams.onprem.client import Users, Client

# Network
from diagrams.onprem.network import Nginx, HAProxy, Traefik
```

### Programming & Generic

```python
from diagrams.programming.language import Python, Java, Go, Rust, Javascript, Typescript
from diagrams.programming.framework import React, Vue, Angular, FastAPI, Flask, Django
from diagrams.generic.storage import Storage
from diagrams.generic.database import SQL as GenericSQL
from diagrams.generic.compute import Rack
```

---

## Custom Bundled Icons

Located in `${CLAUDE_PLUGIN_ROOT}/skills/fe-architecture-diagram/resources/icons/`

### Databricks Icons (`icons/databricks/`)

| File | Description | Use Case |
|------|-------------|----------|
| `workspace.png` | Databricks Workspace | Main platform entry point |
| `unity_catalog.png` | Unity Catalog | Data governance layer |
| `delta_lake.png` | Delta Lake | Storage tables (Bronze/Silver/Gold) |
| `lakehouse.png` | Lakehouse | Generic lakehouse representation |
| `sql_warehouse.png` | SQL Warehouse | Query compute for analytics |
| `model_serving.png` | Model Serving | ML inference endpoints |

### Cloud Services Icons (`icons/cloud/`)

| File | Description | Use Case |
|------|-------------|----------|
| `snowflake.png` | Snowflake | External data warehouse |
| `kafka.png` | Apache Kafka | Alternative Kafka icon |
| `confluent.png` | Confluent Cloud | Managed Kafka |
| `airflow.png` | Apache Airflow | Alternative Airflow icon |

---

## Using Custom Icons

```python
from diagrams.custom import Custom
import os

# Define icon path (use glob to handle version wildcards)
from glob import glob

def find_icons():
    patterns = [
        os.path.expanduser("~/.claude/plugins/cache/fe-vibe/fe-workflows/*/skills/fe-architecture-diagram/resources/icons"),
        os.path.expanduser("~/code/vibe/plugins/fe-workflows/skills/fe-architecture-diagram/resources/icons"),
    ]
    for pattern in patterns:
        matches = glob(pattern)
        if matches:
            return matches[0]
    raise FileNotFoundError("Icons directory not found")

ICONS = find_icons()

# Use custom icon
databricks = Custom("Databricks", f"{ICONS}/databricks/workspace.png")
snowflake = Custom("Snowflake", f"{ICONS}/cloud/snowflake.png")
```

---

## Creating New Custom Icons

Requirements for custom icons:
1. **Format**: PNG with transparent background
2. **Size**: 256x256 pixels recommended
3. **Style**: Clean, professional, consistent with existing icons
4. **Placement**: Add to appropriate subdirectory under `icons/`

Sources for vendor icons:
- Official brand asset pages
- SVG→PNG conversion from official logos
- Icon libraries (with proper licensing)

Always document icon sources in `icons/README.md`.
