# Data Generation with dbldatagen and Databricks Serverless Jobs

This guide provides the complete, correct steps and configurations for generating data using dbldatagen and deploying to Databricks serverless jobs.

## Critical Success Factors

### Python and Serverless Environment Version Compatibility

**MOST IMPORTANT:** The serverless client version MUST match your local Python version.

| Serverless Client | Python Version | Databricks Connect |
|-------------------|----------------|-------------------|
| "1" | 3.10.12 | 14.3.7 |
| "2" | 3.11.9 | 15.4.5 |
| "3" | 3.12.3 | 16.4.2 |
| "4" | 3.12.3 | 17.0.1 |

**For Python 3.12 local development:**
- Use `client: "4"` in databricks.yml (serverless environment version 4)
- Use `databricks-connect==17.0.1` in dependencies
- Set `.python-version` to `3.12`
- Set `requires-python = ">=3.12"` in pyproject.toml

## Prerequisites

- Databricks workspace with access to serverless compute
- `uv` package manager installed
- `databricks` CLI installed and authenticated
- Access to Unity Catalog

## Project Structure

```
.
├── databricks.yml                      # Bundle configuration
├── pyproject.toml                      # Python project config
├── .python-version                     # Python version pinning (3.12)
├── generate_data.py                    # Python data generation script
├── notebooks/
│   └── generate_data_notebook.ipynb   # Notebook data generation
├── DEMO.md                            # Demo overview
├── TASKS.md                           # Task tracking
└── DATA_GENERATION.md                 # This file
```

## Step 1: Initialize Project with uv

```bash
# Initialize uv project
uv init --no-readme

# Set Python version to 3.12 (matches serverless environment version 4)
echo "3.12" > .python-version

# Add required dependencies with EXACT versions for client 4 compatibility
uv add "databricks-connect>=17.0.0,<=17.0.1" dbldatagen jmespath pyparsing
```

### Key Configuration Files

**pyproject.toml:**
```toml
[project]
name = "data-generation-demo"
version = "0.1.0"
requires-python = ">=3.12"
dependencies = [
    "databricks-connect>=17.0.0,<=17.0.1",
    "dbldatagen>=0.4.0.post1",
    "jmespath>=1.0.1",
    "pyparsing>=3.2.5",
]
```

**.python-version:**
```
3.12
```

## Step 2: Create Python Data Generation Script

Create `generate_data.py` with the following key requirements:

### Critical Requirements for Serverless Compatibility

1. **Use DatabricksSession with environment dependencies for local execution:**
   ```python
   import os
   if os.environ.get('DATABRICKS_RUNTIME_VERSION'):
       # Running in Databricks
       spark = SparkSession.builder.getOrCreate()
   else:
       # Running locally with Databricks Connect
       from databricks.connect import DatabricksSession, DatabricksEnv

       # Create environment with UDF dependencies for local execution
       env = (DatabricksEnv()
              .withDependencies("dbldatagen==0.4.0.post1")
              .withDependencies("jmespath==1.0.1")
              .withDependencies("pyparsing==3.2.5"))

       spark = DatabricksSession.builder.serverless(True).withEnvironment(env).getOrCreate()
   ```

   **Why this is needed:** dbldatagen uses UDFs internally. UDFs execute on remote Databricks compute, so dependencies must be specified via `DatabricksEnv()` to be available during UDF execution. This feature requires Databricks Connect 16.4+.

2. **Avoid sparkContext (not supported in serverless):**
   ```python
   # DO NOT USE:
   # spark.sparkContext.appName

   # USE INSTEAD:
   print(f"Spark version: {spark.version}")
   ```

3. **Avoid dbldatagen parameters not supported:**
   ```python
   # DO NOT USE: unique=True parameter
   .withColumn("customer_id", "long", minValue=100000, maxValue=999999, unique=True)

   # USE INSTEAD:
   .withColumn("customer_id", "long", minValue=100000, maxValue=999999)
   ```

## Step 3: Create Notebook

Create `notebooks/generate_data_notebook.ipynb` with the same requirements:

### Key Points for Notebooks

1. **Same SparkSession pattern as Python script**
2. **No `spark.sparkContext` references**
3. **No `unique=True` in dbldatagen column specs**
4. **Use `display()` for showing data in notebooks (works fine in remote execution)**

## Step 4: Create Databricks Bundle Configuration

Create `databricks.yml`:

### Critical Bundle Configuration Requirements

1. **Use client version "4" for Python 3.12:**
   ```yaml
   environments:
     - environment_key: serverless_env
       spec:
         client: "4"  # CRITICAL: Must be "4" for Python 3.12
         dependencies:
           - "dbldatagen==0.4.0.post1"
           - "jmespath==1.0.1"
           - "pyparsing==3.2.5"
   ```

2. **Use serverless environments, not clusters:**
   ```yaml
   tasks:
     - task_key: generate_data_python
       environment_key: serverless_env  # Use environment, not cluster
   ```

3. **DO NOT specify libraries at task level for serverless:**
   ```yaml
   # WRONG - Do not do this:
   tasks:
     - task_key: my_task
       libraries:
         - pypi:
             package: dbldatagen

   # CORRECT - Specify in environment:
   tasks:
     - task_key: my_task
       environment_key: serverless_env
   environments:
     - environment_key: serverless_env
       spec:
         client: "4"
         dependencies:
           - "dbldatagen==0.4.0.post1"
   ```

### Complete Bundle Example

```yaml
bundle:
  name: data_generation_demo

workspace:
  host: https://e2-demo-field-eng.cloud.databricks.com 

resources:
  jobs:
    generate_data_python_job:
      name: "[${bundle.environment}] Data Generation - Python Script"

      tasks:
        - task_key: generate_data_python
          spark_python_task:
            python_file: ${workspace.file_path}/generate_data.py
            source: WORKSPACE
          environment_key: serverless_env

      environments:
        - environment_key: serverless_env
          spec:
            client: "4"
            dependencies:
              - "dbldatagen==0.4.0.post1"
              - "jmespath==1.0.1"
              - "pyparsing==3.2.5"

    generate_data_notebook_job:
      name: "[${bundle.environment}] Data Generation - Notebook"

      tasks:
        - task_key: generate_data_notebook
          notebook_task:
            notebook_path: ${workspace.file_path}/notebooks/generate_data_notebook
            source: WORKSPACE
          environment_key: serverless_env

      environments:
        - environment_key: serverless_env
          spec:
            client: "4"
            dependencies:
              - "dbldatagen==0.4.0.post1"
              - "jmespath==1.0.1"
              - "pyparsing==3.2.5"

targets:
  development:
    mode: development
    default: true
    workspace:
      root_path: /Users/${workspace.current_user.userName}/.bundle/${bundle.name}/${bundle.environment}
```

## Step 5: Local Testing with Databricks Connect

```bash
# Sync dependencies
uv sync

# Run locally using Databricks Connect
uv run python generate_data.py
```

**Expected Behavior:**
- ✅ Databricks Connect successfully connects to serverless
- ✅ Spark session is created with UDF dependencies
- ✅ Data generation completes successfully
- ✅ Data is written to Unity Catalog table
- ✅ Verification shows 10,000 rows written

**How this works:**
- dbldatagen uses UDFs internally that execute on remote Databricks compute
- `DatabricksEnv().withDependencies()` specifies packages needed for UDF execution
- Dependencies are installed in the remote serverless environment automatically
- This feature requires Databricks Connect 16.4+ (we're using 17.0.1)

**Local execution output:**
```
Initializing Spark session...
Spark version: 4.0.0
Generating data for table: main.default.generated_data_python

Writing data to main.default.generated_data_python...

Verifying data was written...
Total rows in main.default.generated_data_python: 10000

Data generation completed successfully!
```

## Step 6: Deploy and Run

```bash
# Validate bundle
databricks bundle validate

# Deploy bundle
databricks bundle deploy

# Run Python script job
databricks bundle run generate_data_python_job

# Run notebook job
databricks bundle run generate_data_notebook_job
```

### Expected Output

Both jobs should complete successfully with output similar to:

```
Initializing Spark session...
Spark version: 4.0.0
Generating data for table: main.default.generated_data_python

Writing data to main.default.generated_data_python...

Verifying data was written...
Total rows in main.default.generated_data_python: 10000

Data generation completed successfully!
```

## Step 7: Verify Data

Query the tables to verify data was generated:

```sql
-- Verify Python script job
SELECT COUNT(*) FROM main.default.generated_data_python;
-- Expected: 10000 rows

-- Verify notebook job
SELECT COUNT(*) FROM main.default.generated_data_notebook;
-- Expected: 10000 rows

-- View sample data
SELECT * FROM main.default.generated_data_python LIMIT 10;
```

## Common Issues and Solutions

### Issue 1: Python Version Mismatch

**Error:** `Python versions in the Spark Connect client and server are different`

**Solution:**
- Use Python 3.12 locally (matches serverless environment version 4)
- Set `.python-version` file to `3.12`
- Update `pyproject.toml` to `requires-python = ">=3.12"`
- Use `client: "4"` in databricks.yml

### Issue 2: Wrong Serverless Client Version

**Error:** Jobs fail or incompatibility errors

**Solution:**
- ALWAYS use `client: "4"` for Python 3.12
- NEVER use `client: "1"` with Python 3.12 (it uses Python 3.10)
- Check the compatibility table at the top of this document

### Issue 3: Databricks Connect Version Incompatibility

**Error:** Connection errors or version mismatch

**Solution:**
- Use `databricks-connect==17.0.1` for client version 4
- Match Databricks Connect version to serverless client version (see table above)

### Issue 4: Libraries Not Found During Local Execution

**Error:** `ModuleNotFoundError: No module named 'dbldatagen'` (when running locally)

**Solution:**
- Use `DatabricksEnv().withDependencies()` when creating the Spark session locally
- This specifies packages needed for UDF execution on remote compute
- For serverless jobs, specify dependencies in `environments.spec.dependencies`
- Libraries field is NOT supported in `tasks.libraries` for serverless

### Issue 5: Missing node_type_id

**Error:** `Missing required field 'node_type_id'`

**Solution:**
- Don't use `job_clusters` for serverless
- Use `environment_key` at task level instead
- Let serverless handle cluster provisioning automatically

### Issue 6: sparkContext Not Supported

**Error:** `[JVM_ATTRIBUTE_NOT_SUPPORTED] Attribute 'sparkContext' is not supported`

**Solution:**
- Remove all references to `spark.sparkContext`
- Serverless doesn't support JVM attribute access
- Use alternatives like `spark.version` instead

### Issue 7: dbldatagen unique Parameter

**Error:** `DataGenError(msg='invalid column option unique')`

**Solution:**
- Remove `unique=True` parameter from column specifications
- This parameter is not supported in the version of dbldatagen we're using

## Summary: Key Success Factors

1. ✅ **Python 3.12** - Matches serverless environment version 4
2. ✅ **Client version "4"** - Critical for Python 3.12 compatibility
3. ✅ **Databricks Connect 17.0.1** - Compatible with client version 4
4. ✅ **DatabricksEnv with dependencies** - Enables local execution with UDFs
5. ✅ **Environment-based dependencies** - Not task-level libraries for jobs
6. ✅ **No sparkContext references** - Serverless limitation
7. ✅ **No unique parameter** - dbldatagen compatibility
8. ✅ **Proper SparkSession creation** - Different for local vs remote

## Version Compatibility Quick Reference

**Using Python 3.12 locally? (RECOMMENDED)**
```toml
# pyproject.toml
requires-python = ">=3.12"
dependencies = ["databricks-connect>=17.0.0,<=17.0.1"]
```

```python
# generate_data.py - Local execution with UDF dependencies
from databricks.connect import DatabricksSession, DatabricksEnv

env = (DatabricksEnv()
       .withDependencies("dbldatagen==0.4.0.post1")
       .withDependencies("jmespath==1.0.1")
       .withDependencies("pyparsing==3.2.5"))

spark = DatabricksSession.builder.serverless(True).withEnvironment(env).getOrCreate()
```

```yaml
# databricks.yml - Job deployment
environments:
  - environment_key: serverless_env
    spec:
      client: "4"  # MUST be "4" for Python 3.12
      dependencies:
        - "dbldatagen==0.4.0.post1"
        - "jmespath==1.0.1"
        - "pyparsing==3.2.5"
```

**Using Python 3.11 locally?**
```toml
# pyproject.toml
requires-python = ">=3.11"
dependencies = ["databricks-connect==15.4.5"]
```

```yaml
# databricks.yml
environments:
  - environment_key: serverless_env
    spec:
      client: "2"  # MUST be "2" for Python 3.11
```

**Using Python 3.10 locally?**
```toml
# pyproject.toml
requires-python = ">=3.10"
dependencies = ["databricks-connect==14.3.7"]
```

```yaml
# databricks.yml
environments:
  - environment_key: serverless_env
    spec:
      client: "1"  # MUST be "1" for Python 3.10
```

## Data Generation Best Practices

1. **Test locally first** - Use Databricks Connect with `DatabricksEnv()` for rapid iteration
2. **Make code idempotent** - Use `.write.mode("overwrite")` to ensure reruns work
3. **Write to Unity Catalog** - Use `catalog.schema.table` naming
4. **Use partitioning** - Specify `partitions` in DataGenerator for performance
5. **Test with small datasets** - Use smaller `rows` value for testing (e.g., 1000 rows)
6. **Verify data** - Always check row counts and schema after generation
7. **Match dependencies** - Keep local `DatabricksEnv()` and job `environments.spec.dependencies` in sync

## Next Steps

After successful deployment:

1. Schedule jobs for regular execution
2. Set up alerts for job failures
3. Monitor Unity Catalog table sizes
4. Consider adding data quality checks
5. Implement incremental data generation if needed

## References

- [Databricks Serverless Environment Versions](https://docs.databricks.com/aws/en/release-notes/serverless/environment-version/)
- [Databricks Connect Documentation](https://docs.databricks.com/dev-tools/databricks-connect.html)
- [Databricks Connect UDFs with Dependencies](https://docs.databricks.com/aws/en/dev-tools/databricks-connect/python/udf)
- [dbldatagen Documentation](https://github.com/databrickslabs/dbldatagen)
- [Databricks Asset Bundles](https://docs.databricks.com/dev-tools/bundles/index.html)
