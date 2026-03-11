# Databricks Connect Serverless Configuration

**Status**: Public Preview (Python only)

## Configuration Methods

### 1. Environment Variable
```bash
export DATABRICKS_SERVERLESS_COMPUTE_ID=auto
```
Ignores any cluster_id setting

### 2. Configuration Profile
Add to `.databrickscfg`:
```
[DEFAULT]
host = https://my-workspace.cloud.databricks.com/
serverless_compute_id = auto
token = dapi123...
```

### 3. Programmatic Setup
**Option A:**
```python
spark = DatabricksSession.builder.serverless(True).getOrCreate()
```

**Option B:**
```python
spark = DatabricksSession.builder.remote(serverless=True).getOrCreate()
```

## Important Limitations
- **Python only** support
- **No additional dependencies** beyond serverless environment defaults
- **Custom module UDFs unsupported**

## Session Timeout
- Sessions expire after **10 minutes of inactivity**
- Must create new session using `getOrCreate()`

## Version Compatibility
Specific version requirements between Python, Databricks Connect, and serverless compute
