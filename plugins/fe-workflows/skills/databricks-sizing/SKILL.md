---
name: databricks-sizing
description: Get Databricks cost estimates using the Quicksizer agent. Sizes ETL, Data Warehousing, ML, Interactive, Lakebase, Lakeflow Connect, and Apps workloads.
user-invocable: true
---

# Databricks Sizing (Quicksizer)

This skill provides automated Databricks cost estimation using the Quicksizer agent. It calculates monthly DBU costs for customer workloads across all major use case types.

**Announce at start:** "I'm using the databricks-sizing skill to get a Databricks cost estimate."

## Supported Use Cases

| Use Case | Description |
|----------|-------------|
| **Ingestion/ETL** | Batch and streaming data processing pipelines |
| **Data Warehousing** | SQL analytics, BI dashboards, Genie, AI/BI |
| **ML** | Model training, batch inference, real-time serving |
| **Interactive** | Notebook development and testing |
| **Lakebase** | Managed Postgres OLTP databases |
| **Lakeflow Connect** | Managed ingestion from SaaS and database sources |
| **Apps** | Databricks Apps serverless compute |

## Prerequisites

> **NOTE**: This skill calls a Databricks AI agent endpoint. By default it points
> to a Databricks-hosted quicksizer instance. If you have your own Databricks
> workspace, you can deploy a sizing agent and update `ENDPOINT_URL` in
> `resources/quicksizer_api.py`. Requires Databricks CLI authentication.

1. **Databricks authentication** - Run `/databricks-authentication` and set up your workspace profile
2. **Python 3.10+** - For running the API helper script

## Workflow

### Step 1: Authenticate

Verify Databricks authentication:

```bash
databricks auth describe -p <your-profile>
```

If not authenticated, run `/databricks-authentication` first.

### Step 2: Call the API

#### Single-Shot Request

If the user provides enough sizing details, call the API directly:

```bash
python3 resources/quicksizer_api.py "Size an ETL workload on AWS Premium: 500GB daily, 10 batch jobs, standard processing"
```

#### Get Discovery Questions First

To get discovery questions for a specific use case type:

```bash
python3 resources/quicksizer_api.py --discovery "etl"
```

### Step 3: Handle Multi-Turn Conversations

**IMPORTANT:** The Quicksizer agent often asks clarifying questions before providing a cost estimate. Use `--messages-file` to maintain conversation state across turns.

**How it works:**
- Pass `--messages-file /tmp/qs_chat.json` on every call
- The script creates the file on the first call, then reads/appends on follow-ups
- The full conversation history is sent to the API each turn
- The file is updated with both the user message and assistant response after each call

**Multi-turn flow:**

```bash
# Turn 1: Initial request (creates /tmp/qs_chat.json)
python3 resources/quicksizer_api.py --messages-file /tmp/qs_chat.json \
  "Size ETL: 500GB daily, 10 batch jobs, AWS Premium"

# Agent responds with clarifying questions...

# Turn 2: Answer the clarifying questions (appends to /tmp/qs_chat.json)
python3 resources/quicksizer_api.py --messages-file /tmp/qs_chat.json \
  "Yes, 50GB per job run is correct. Standard processing complexity."

# Turn 3: Continue until you get a cost estimate with "Monthly Cost Breakdown"
python3 resources/quicksizer_api.py --messages-file /tmp/qs_chat.json \
  "No ESC add-on needed. US East region."
```

**Conversation guidelines:**
1. Send the user's sizing request to the API
2. If the agent asks clarifying questions:
   - Check if the user already provided that info (rephrase it)
   - OR make reasonable assumptions based on context (e.g., "50GB per job" if they said "500GB across 10 jobs")
   - OR ask the user for the missing information
3. Send a follow-up with the answers using the same `--messages-file`
4. Repeat until you get a cost estimate with "Monthly Cost Breakdown"

**Starting a new conversation:** Delete the messages file or use a different path:

```bash
rm -f /tmp/qs_chat.json
```

### Step 4: Gather Information (if needed upfront)

If the user's request is vague, collect these before calling the API:

**Global Requirements (all use cases):**
- Cloud provider: AWS, Azure, GCP, or SAP
- Platform tier: Premium (default), Enterprise, or NA (SAP only)
- Region (optional): Improves estimate accuracy
- ESC add-on (optional): Required for HIPAA/PCI compliance

**Use-Case Specific:** See [Discovery Questions](#discovery-questions) below.

## Discovery Questions

### Ingestion/ETL

| Parameter | Required | Discovery Question |
|-----------|----------|-------------------|
| Data size GB per run | Yes | How much data processed per job run? |
| Frequency | Yes | Batch or streaming? |
| Number of jobs | Yes | How many pipelines? |
| Processing type | No | Light (minimal transforms) or standard? |
| Batch runs per day | No | How many times per day? (default: 1) |
| Streaming hours daily | No | Hours per day for streaming? (default: 24) |
| SKU type | No | Serverless (default), Classic, or DLT? |

### Data Warehousing

| Parameter | Required | Discovery Question |
|-----------|----------|-------------------|
| Total data size TB | Yes | Total data in scope for analysts/dashboards? |
| Number of dashboards | Yes | How many dashboards or reports? |
| Number of analysts | Yes | How many users query data directly? |
| Usage hours daily | Yes | Hours per day warehouse is active? |
| SKU type | No | SQL Serverless (default), Classic, or Pro? |

### ML

| Parameter | Required | Discovery Question |
|-----------|----------|-------------------|
| Total models | Yes | How many ML models? |
| Models requiring GPU | No | How many need GPU for training? (default: 0) |
| Models requiring serving | Yes | How many need real-time inference? |
| Serving models requiring GPU | No | How many serving models need GPU? (default: 0) |
| Serving uptime hours | No | Hours per day endpoints active? |
| Average QPS served | No | Peak requests per second per endpoint? |
| Training data size GB | Yes | Size of training data? |
| Training runs per month | No | Retraining frequency? (default: 2) |
| Batch inference data size GB | No | Data scored per batch run? |
| Batch inference runs per day | No | Daily batch inference runs? (default: 1) |

### Interactive

| Parameter | Required | Discovery Question |
|-----------|----------|-------------------|
| Number of developers | Yes | How many developers using notebooks? |
| Use case data size GB | Yes | Total data for dev/test? |
| Daily dev hours | No | Hours per day per developer? (default: 5) |

### Lakebase

| Parameter | Required | Discovery Question |
|-----------|----------|-------------------|
| Total data volume GB | Yes | Total database storage? (max 2TB) |
| Estimated QPS | Yes | Peak queries per second? |
| Query type | Yes | Simple (point lookups) or complex (joins/aggregations)? |
| Sync mode | Yes | Batch, continuous, or no sync from Delta? |
| Daily uptime hours | No | Hours per day database runs? (default: 24) |
| Data retention days | No | Days for rollback retention? (default: 7) |
| Batch sync frequency | No | Hourly, daily, or weekly? (if batch sync) |
| Batch sync size GB | No | Data synced per batch? (if batch sync) |

### Lakeflow Connect

| Parameter | Required | Discovery Question |
|-----------|----------|-------------------|
| Is database connection | Yes | Database source (vs SaaS like Salesforce)? |
| Is snapshot | No | Full load or CDC/incremental? (if database) |
| Ingestion data volume per run GB | Yes | Data volume per pipeline run? |
| Pipeline runs per day | No | Runs per day? (default: 1) |

### Apps

| Parameter | Required | Discovery Question |
|-----------|----------|-------------------|
| App compute size | No | Medium (default) or large? |

**Note:** Apps often use other Databricks services (DBSQL, Lakebase, Jobs). Size those separately.

## Example Requests

### Simple ETL Sizing

```
I need to size an ETL workload on AWS Premium:
- 500GB data processed daily
- 10 batch jobs running once per day
- Standard processing complexity
```

### Multi-Use-Case Sizing

```
Customer needs sizing for Azure Premium:
1. ETL: 2TB daily, 20 streaming jobs
2. Data Warehousing: 10TB total, 50 analysts, 20 dashboards, 8 hours/day
3. ML: 5 models, 500GB training data, 2 models for real-time serving 24/7
```

### From Meeting Notes

```
Customer discussion notes:
- Migrating from Snowflake
- Processing 2TB of data daily
- 100 analysts running queries
- Azure cloud
- Need ML for fraud detection with 500GB training data
```

## CLI Reference

```bash
# Single request
python3 resources/quicksizer_api.py "Size an ETL workload: 500GB daily on AWS Premium"

# Discovery questions for a use case
python3 resources/quicksizer_api.py --discovery "data warehousing"

# Multi-turn with conversation file
python3 resources/quicksizer_api.py --messages-file /tmp/qs_chat.json "Your message here"

# Override email
python3 resources/quicksizer_api.py --email user@databricks.com "Size ETL workload..."

# Raw JSON output (for debugging)
python3 resources/quicksizer_api.py --raw "Size ETL workload..."
```

## Important Notes

1. **Mosaic AI not supported** - Use [go/genaicalculator](https://go/genaicalculator) for Mosaic AI use cases
2. **Moderate confidence estimates** - Tool makes assumptions; communicate these to customers
3. **ETL + consumption** - Always ask how ingested data will be consumed (warehousing, ML, AI)
4. **ESC add-on** - Adds 15% cost multiplier; requires Enterprise tier on AWS/GCP
5. **Region matters** - Providing region improves accuracy, especially for serverless

## Support

- **Slack:** #sizing
- **Team:** @sizingassistant
- **Feedback:** https://forms.gle/ee4yT2k12JQCrEFr9
- **Chatbot UI:** https://quicksizer-chatbot-1199445695448328.aws.databricksapps.com/
