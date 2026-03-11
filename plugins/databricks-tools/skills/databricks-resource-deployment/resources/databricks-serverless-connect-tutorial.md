# Databricks Connect with Serverless Compute

## Requirements
- Databricks Connect 15.4 LTS or above
- Serverless compute enabled in workspace
- Python 3.12
- IDE (e.g., Visual Studio Code)
- Databricks CLI

## Authentication Setup
OAuth user-to-machine via CLI:
```bash
databricks auth login --host <workspace-url>
```
Use DEFAULT profile recommended

## Environment Setup

### 1. Create Virtual Environment
```bash
python3.12 -m venv .venv
```

### 2. Activate
OS-specific commands

### 3. Install Databricks Connect
```bash
pip install "databricks-connect==16.4.*"
```

## Basic Usage
```python
from databricks.connect import DatabricksSession
spark = DatabricksSession.builder.serverless().profile("<profile-name>").getOrCreate()
df = spark.read.table("samples.nyctaxi.trips")
df.show(5)
```

## Production Best Practice
Avoid hardcoded compute specs. Use configuration file:

**.databrickscfg:**
```
serverless_compute_id = auto
```

**Code:**
```python
spark = DatabricksSession.builder.getOrCreate()
```

Or use environment variable:
```bash
DATABRICKS_SERVERLESS_COMPUTE_ID=auto
```
