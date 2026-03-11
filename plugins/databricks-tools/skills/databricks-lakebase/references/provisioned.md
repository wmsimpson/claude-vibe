# Lakebase Provisioned Tier

> **Note:** The Provisioned Tier is the legacy approach. For new projects, use the **Autoscaling Tier** (see main SKILL.md). The Provisioned Tier uses fixed compute capacity and does not support scale-to-zero or first-class branching.

## Prerequisites

- **Databricks CLI** - Version 0.229.0+
- **psql client** (optional): `brew install postgresql@16`

## Quick Reference: CLI Commands

```bash
# List all instances
databricks database list-database-instances -p PROFILE

# Get instance details
databricks database get-database-instance INSTANCE_NAME -p PROFILE

# Create instance
databricks database create-database-instance INSTANCE_NAME \
  --capacity=CU_1 \
  --enable-pg-native-login \
  -p PROFILE

# Update instance
databricks database update-database-instance INSTANCE_NAME "capacity" \
  --capacity=CU_2 \
  -p PROFILE

# Delete instance
databricks database delete-database-instance INSTANCE_NAME -p PROFILE

# Connect with psql (works with provisioned tier only)
databricks psql INSTANCE_NAME -p PROFILE
```

## Creating an Instance

### Step 1: Create the Instance

```bash
databricks database create-database-instance my-lakebase \
  --capacity=CU_1 \
  --enable-pg-native-login \
  --no-wait \
  -p PROFILE
```

**Capacity Options:**
- `CU_1` - 1 Compute Unit (~2GB RAM) - Development/Testing
- `CU_2` - 2 Compute Units (~4GB RAM) - Light production
- `CU_4` - 4 Compute Units (~8GB RAM) - Production
- `CU_8` - 8 Compute Units (~16GB RAM) - Heavy production

**Additional Options:**
- `--enable-pg-native-login` - Allow password-based authentication
- `--retention-window-in-days INT` - PITR retention (default: 7 days)
- `--node-count INT` - Number of nodes (1 primary + N-1 secondaries)
- `--enable-readable-secondaries` - Enable read replicas
- `--no-wait` - Don't wait for instance to be available

### Step 2: Wait for Instance to be Available

```bash
# Check status
databricks database get-database-instance my-lakebase -p PROFILE | jq '.state'

# Wait for AVAILABLE state (takes 2-5 minutes)
while [ "$(databricks database get-database-instance my-lakebase -p PROFILE | jq -r '.state')" != "AVAILABLE" ]; do
  echo "Waiting..."
  sleep 30
done
echo "Instance ready!"
```

### Step 3: Get Connection Details

```bash
databricks database get-database-instance my-lakebase -p PROFILE
```

Output includes:
- `read_write_dns` - Primary endpoint (read-write)
- `read_only_dns` - Read replica endpoint
- `pg_version` - PostgreSQL version

## Connecting

### Option 1: Databricks CLI psql (Recommended for Provisioned)

```bash
# Interactive session
databricks psql my-lakebase -p PROFILE

# Run single command
databricks psql my-lakebase -p PROFILE -- -c "SELECT version();"

# Connect to specific database
databricks psql my-lakebase -p PROFILE -- -d mydb -c "SELECT * FROM users;"
```

The CLI automatically handles OAuth authentication.

> **Note:** `databricks psql` does NOT work with Autoscaling Tier projects. Use direct psql with OAuth token instead (see main SKILL.md).

### Option 2: Direct psql with OAuth Token

```bash
# Generate OAuth token
TOKEN=$(databricks database generate-database-credential \
  --json '{"request_id": "cli", "instance_names": ["my-lakebase"]}' \
  -p PROFILE | jq -r '.token')

# Get host
HOST=$(databricks database get-database-instance my-lakebase -p PROFILE | jq -r '.read_write_dns')

# Connect
PGPASSWORD=$TOKEN psql \
  "host=$HOST port=5432 dbname=postgres user=you@example.com sslmode=require"
```

## Managing Instance Configuration

### Scale Capacity

```bash
databricks database update-database-instance my-lakebase "capacity" \
  --capacity=CU_4 \
  -p PROFILE
```

### Add Read Replicas

```bash
databricks database update-database-instance my-lakebase "node_count,enable_readable_secondaries" \
  --node-count=2 \
  --enable-readable-secondaries \
  -p PROFILE
```

Connections to `read_only_dns` will route to the replica.

### Stop/Start Instance

```bash
# Stop instance (saves compute cost, keeps data)
databricks database update-database-instance my-lakebase "stopped" \
  --stopped \
  -p PROFILE

# Start instance
databricks database update-database-instance my-lakebase "stopped" \
  --json '{"stopped": false}' \
  -p PROFILE
```

### Change Retention Window

```bash
databricks database update-database-instance my-lakebase "retention_window_in_days" \
  --retention-window-in-days=14 \
  -p PROFILE
```

## Deleting an Instance

```bash
# Delete instance (PERMANENT - deletes all data!)
databricks database delete-database-instance my-lakebase -p PROFILE

# Force delete if has PITR descendants
databricks database delete-database-instance my-lakebase --force -p PROFILE
```
