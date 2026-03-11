# Databricks Products & Features for Architecture Diagrams

This reference helps you understand where each Databricks product fits in an architecture and what it competes with.

---

## Data Engineering

### Delta Lake
**What it is**: Open-source storage layer with ACID transactions, schema enforcement, and time travel
**Where it fits**: Storage layer in any lakehouse architecture (Bronze/Silver/Gold)
**Competes with**: Apache Iceberg, Apache Hudi, raw Parquet files
**Icon**: `databricks/delta_lake.png`

### Delta Live Tables (DLT)
**What it is**: Declarative ETL framework for building data pipelines
**Where it fits**: Data transformation layer between Bronze → Silver → Gold
**Competes with**: dbt, Apache Airflow (for orchestration), AWS Glue ETL, Azure Data Factory
**Use in diagram**: Show as transformation nodes between medallion layers

### Auto Loader
**What it is**: Incrementally ingests new data files as they arrive
**Where it fits**: Ingestion layer, typically feeds Bronze tables
**Competes with**: Fivetran, Airbyte, AWS Glue Crawlers, Kafka Connect
**Use in diagram**: Arrow from source (S3/ADLS/GCS) to Bronze

### Lakeflow Connect
**What it is**: Managed connectors for ingesting data from SaaS apps and databases
**Where it fits**: Ingestion layer for external sources (Salesforce, SAP, databases)
**Competes with**: Fivetran, Airbyte, Stitch, HVR, Qlik Replicate
**Use in diagram**: Arrow from external systems to Bronze

### Workflows (Jobs)
**What it is**: Orchestration for running notebooks, JARs, and pipelines
**Where it fits**: Orchestration layer coordinating multiple data processing steps
**Competes with**: Apache Airflow, Prefect, Dagster, AWS Step Functions, Azure Data Factory
**Use in diagram**: Often implicit (controls flow), or show as orchestration box

---

## Data Warehousing & SQL Analytics

### Databricks SQL (DBSQL)
**What it is**: SQL analytics service optimized for BI workloads
**Where it fits**: Query layer for analysts and BI tools
**Competes with**: Snowflake, BigQuery, Redshift, Azure Synapse Dedicated Pools
**Use in diagram**: Show connecting Gold layer to BI tools/analysts

### SQL Warehouse
**What it is**: Serverless or classic compute for running SQL queries
**Where it fits**: Compute layer under Databricks SQL
**Competes with**: Snowflake virtual warehouses, BigQuery slots, Redshift clusters
**Icon**: `databricks/sql_warehouse.png`

### Lakeview Dashboards
**What it is**: Native BI/visualization tool in Databricks
**Where it fits**: Visualization layer for business users
**Competes with**: Tableau, Power BI, Looker, Sigma, Hex
**Use in diagram**: Show as dashboard/visualization endpoint

### Genie
**What it is**: Natural language interface to query data
**Where it fits**: Self-service analytics layer for business users
**Competes with**: ThoughtSpot, Microsoft Copilot, natural language features in BI tools
**Use in diagram**: Show as user interface between users and data

---

## Data Governance

### Unity Catalog
**What it is**: Unified governance for data and AI assets across clouds
**Where it fits**: Governance layer spanning the entire platform
**Competes with**: Collibra, Alation, Atlan, AWS Lake Formation, Snowflake Governance
**Icon**: `databricks/unity_catalog.png`
**Use in diagram**: Show as overarching governance layer or separate box

### Delta Sharing
**What it is**: Open protocol for secure data sharing across organizations
**Where it fits**: Data sharing layer for external partners/customers
**Competes with**: Snowflake Data Sharing, AWS Data Exchange, data marketplace solutions
**Use in diagram**: Arrow from internal data to external consumers

---

## AI/BI (Databricks Intelligence)

AI/BI is the unified intelligence layer combining natural language access with visualization.

### Genie
**What it is**: Natural language interface to query data via "Genie Rooms"
**Where it fits**: Self-service analytics layer - business users ask questions in plain English
**Key features**:
- Genie Rooms: Curated spaces with trusted data and instructions
- Auto-generates SQL from natural language
- Learns from feedback to improve over time
- Certified answers for trusted responses
**Competes with**: ThoughtSpot, Tableau Ask Data, Power BI Q&A, Qlik Insight Advisor
**Use in diagram**: Show as user-facing interface above SQL Warehouse

### Lakeview Dashboards
**What it is**: Native BI/visualization tool built into Databricks
**Where it fits**: Visualization layer for business users and embedded analytics
**Key features**:
- Drag-and-drop dashboard builder
- Live connection to lakehouse data
- Scheduled refreshes and alerts
- Embeddable in external apps
**Competes with**: Tableau, Power BI, Looker, Sigma, Mode, Hex
**Use in diagram**: Show as dashboard/visualization endpoint

### AI/BI Alerts
**What it is**: Intelligent alerting based on data conditions
**Where it fits**: Operational monitoring and anomaly detection
**Competes with**: PagerDuty, Datadog alerts, custom monitoring solutions
**Use in diagram**: Show connected to dashboards with notification arrows

---

## Generative AI & Agents

### Mosaic AI Agent Framework
**What it is**: End-to-end framework for building, evaluating, and deploying AI agents
**Where it fits**: Application layer for compound AI systems
**Key components**:
- **Agent Builder**: Visual tool for designing agent workflows
- **Tool Calling**: Connect agents to SQL, APIs, Vector Search, Python
- **Guardrails**: Safety controls and content filtering
- **Agent Evaluation**: Systematic testing of agent behavior
**Competes with**: LangChain, LlamaIndex, AWS Bedrock Agents, Semantic Kernel, AutoGen
**Use in diagram**: Show as orchestration layer between user and tools/LLMs

### Mosaic AI Gateway
**What it is**: Unified API gateway for LLM access with governance
**Where it fits**: API layer managing access to multiple LLM providers
**Key features**:
- Route requests to different models (OpenAI, Anthropic, DBRX, Llama)
- Rate limiting and cost controls
- Usage tracking and audit logs
- Fallback routing
**Competes with**: LiteLLM, Portkey, Helicone, direct API access
**Use in diagram**: Show between applications and Foundation Model APIs

### Foundation Model APIs
**What it is**: Pay-per-token access to LLMs hosted on Databricks
**Where it fits**: LLM inference layer for GenAI applications
**Available models**:
- DBRX (Databricks' own model)
- Llama 3.x (Meta)
- Mixtral (Mistral)
- BGE embeddings
**Competes with**: OpenAI API, Azure OpenAI, Anthropic API, AWS Bedrock, Google Vertex AI
**Use in diagram**: Show as API endpoint; can show specific model names

### Model Serving
**What it is**: Unified endpoint for serving ML models and LLMs
**Where it fits**: Inference layer serving predictions/completions to applications
**Key features**:
- Real-time and batch inference
- GPU and CPU endpoints
- Auto-scaling
- A/B testing and traffic splitting
**Competes with**: SageMaker endpoints, Vertex AI endpoints, Azure ML, Seldon, KServe, vLLM
**Icon**: `databricks/model_serving.png`

### Vector Search
**What it is**: Managed vector database for similarity search
**Where it fits**: Retrieval layer for RAG applications
**Key features**:
- Delta Sync: Auto-sync from Delta tables
- Hybrid search (vector + keyword)
- Filtered search with metadata
- Direct integration with Unity Catalog
**Competes with**: Pinecone, Weaviate, Milvus, Chroma, pgvector, OpenSearch
**Use in diagram**: Show between document embeddings and LLM context

---

## MLflow (GenAI Features)

### MLflow Tracing
**What it is**: Observability for LLM applications - traces every step of agent execution
**Where it fits**: Debugging and monitoring layer for GenAI apps
**Key features**:
- Automatic tracing of LangChain, LlamaIndex, OpenAI calls
- Trace visualization in UI
- Latency and token usage tracking
- Production trace collection
**Competes with**: LangSmith, Arize Phoenix, Weights & Biases Prompts, Helicone
**Use in diagram**: Show as observability layer with dashed lines from agents

### MLflow Evaluation
**What it is**: Systematic evaluation of LLM outputs and agent behavior
**Where it fits**: Quality assurance layer for GenAI development
**Key features**:
- Built-in judges (answer correctness, relevance, toxicity)
- Custom evaluation metrics
- Evaluation datasets management
- Comparison across model versions
**Competes with**: Ragas, DeepEval, TruLens, custom eval frameworks
**Use in diagram**: Show in CI/CD or development workflow

### MLflow AI Gateway (Legacy)
**What it is**: Earlier name for Mosaic AI Gateway functionality
**Note**: Being consolidated into Mosaic AI Gateway

### Prompt Engineering UI
**What it is**: Interactive playground for developing and testing prompts
**Where it fits**: Development tool for prompt iteration
**Competes with**: OpenAI Playground, Anthropic Console, PromptLayer
**Use in diagram**: Usually not shown (dev tool)

---

## Traditional ML

### MLflow (Classic)
**What it is**: Open-source platform for ML lifecycle management
**Where it fits**: ML experimentation, tracking, and model registry
**Key features**:
- Experiment tracking
- Model registry with stages
- Model versioning
- Artifact storage
**Competes with**: Weights & Biases, Neptune, Comet, SageMaker Model Registry
**Use in diagram**: Show in ML/AI section as experiment tracking

### Feature Store
**What it is**: Centralized repository for ML features with online/offline serving
**Where it fits**: Feature engineering layer between data and ML training
**Key features**:
- Feature tables in Unity Catalog
- Point-in-time lookups
- Online feature serving
- Feature lineage
**Competes with**: Feast, Tecton, SageMaker Feature Store, Vertex AI Feature Store
**Use in diagram**: Show between Gold layer and ML training

---

## Real-Time & Streaming

### Structured Streaming
**What it is**: Stream processing engine built on Spark
**Where it fits**: Stream processing layer for real-time data
**Competes with**: Apache Flink, Kafka Streams, Apache Beam, AWS Kinesis Analytics
**Use in diagram**: Processing node between Kafka and Delta tables

### Delta Lake Streaming
**What it is**: Streaming reads/writes to Delta tables
**Where it fits**: Streaming sink/source for real-time pipelines
**Competes with**: Kafka + raw storage, Iceberg streaming, Hudi streaming
**Use in diagram**: Show streaming arrows into/out of Delta tables

---

## Compute Options

### All-Purpose Clusters
**What it is**: Long-running clusters for interactive development
**Where it fits**: Development and exploration workloads
**Competes with**: Jupyter servers, EMR notebooks, Vertex AI Workbench
**Use in diagram**: Usually implicit (underlying compute)

### Job Clusters
**What it is**: Ephemeral clusters for scheduled/triggered jobs
**Where it fits**: Production batch processing
**Competes with**: EMR transient clusters, Dataproc ephemeral clusters
**Use in diagram**: Usually implicit

### Serverless Compute
**What it is**: On-demand compute without cluster management
**Where it fits**: SQL, notebooks, and jobs with auto-scaling
**Competes with**: BigQuery on-demand, Snowflake, Athena
**Use in diagram**: Often implicit or noted as "serverless"

---

## Applications & Platform

### Databricks Apps
**What it is**: Framework for deploying full-stack applications
**Where it fits**: Application layer for custom data apps
**Competes with**: Streamlit Cloud, Hex, custom deployments, internal app platforms
**Use in diagram**: Show as application tier above data layer

### Lakebase (Postgres)
**What it is**: Managed PostgreSQL databases in Databricks
**Where it fits**: Application backend database for Apps
**Competes with**: RDS, Aurora, Cloud SQL, Neon, Supabase
**Use in diagram**: Show as app database connected to Databricks Apps

### Workspace
**What it is**: Web-based environment for notebooks, SQL, and collaboration
**Where it fits**: User interface layer for data teams
**Competes with**: JupyterHub, Google Colab, Deepnote, Hex
**Icon**: `databricks/workspace.png`

---

## Common Architecture Patterns

### Pattern 1: Modern Data Stack Replacement
```
Before: Fivetran → Snowflake → dbt → Looker
After:  Lakeflow Connect → Delta Lake → DLT → Databricks SQL → Lakeview
```

### Pattern 2: Streaming + Batch Unified
```
Kafka → Structured Streaming → Delta Lake (real-time)
       +→ Batch Jobs → Delta Lake (historical)
All queries via SQL Warehouse
```

### Pattern 3: MLOps Pipeline
```
Gold Tables → Feature Store → ML Training → Model Registry → Model Serving
              ↑                                    ↓
              Feature Serving ←←←←←←←←←←←←←←←←←←←↓
```

### Pattern 4: RAG Application
```
Documents → Chunking → Embeddings → Vector Search
                                         ↓
User Query → Embedding → Vector Search → Context → LLM → Response
```

### Pattern 5: Multi-Cloud Lakehouse
```
AWS S3 ─────┐
GCS ────────┼→ Unity Catalog → Shared Metastore → Workspaces (any cloud)
Azure ADLS ─┘
```

### Pattern 6: AI Agent Architecture
```
User → Databricks App → Agent Framework
                              ↓
                    ┌─────────┼─────────┐
                    ↓         ↓         ↓
               SQL Tool  Vector Search  API Tool
                    ↓         ↓         ↓
               Gold Data  Embeddings  External API
                              ↓
                    Foundation Model API
                              ↓
                    MLflow Tracing (observability)
```

### Pattern 7: AI/BI Self-Service
```
Business Users
      ↓
┌─────┴─────┐
↓           ↓
Genie    Dashboards
Rooms    (Lakeview)
↓           ↓
└─────┬─────┘
      ↓
SQL Warehouse
      ↓
Unity Catalog (governance)
      ↓
Gold Tables + Semantic Models
```

### Pattern 8: Full-Stack Data App
```
End Users → Databricks App (React/FastAPI)
                    ↓
            Lakebase (app state)
                    ↓
            ┌───────┴───────┐
            ↓               ↓
      Model Serving    SQL Warehouse
            ↓               ↓
      ML Models        Gold Tables
```

---

## Cloud-Specific Considerations

### AWS Lakehouse
Native AWS services that integrate with Databricks:
- **Ingestion**: Kinesis, MSK (Kafka), DMS, AppFlow
- **Storage**: S3 (primary), EFS
- **Governance**: Lake Formation (limited - Unity Catalog preferred)
- **Competing analytics**: Redshift, Athena, EMR
- **IAM**: Instance profiles, cross-account roles

### Azure Lakehouse
Native Azure services that integrate with Databricks:
- **Ingestion**: Event Hubs, IoT Hub, Data Factory
- **Storage**: ADLS Gen2 (primary), Blob Storage
- **Governance**: Purview (integrates with Unity Catalog)
- **Competing analytics**: Synapse, Fabric
- **IAM**: Managed identities, service principals

### GCP Lakehouse
Native GCP services that integrate with Databricks:
- **Ingestion**: Pub/Sub, Dataflow, Data Fusion
- **Storage**: GCS (primary)
- **Governance**: Data Catalog, Dataplex
- **Competing analytics**: BigQuery, Dataproc
- **IAM**: Service accounts, Workload Identity

### Cross-Cloud Pattern
When customers use multiple clouds:
```
AWS Region          Azure Region         GCP Region
    ↓                   ↓                    ↓
   S3                 ADLS                  GCS
    ↓                   ↓                    ↓
    └───────────────────┴────────────────────┘
                        ↓
              Unity Catalog (multi-cloud)
                        ↓
              Primary Databricks Workspace
                        ↓
              Cross-cloud Delta Sharing
```
