# Databricks Iceberg Table Access from Snowflake

This guide covers configuring Snowflake to access Databricks Unity Catalog tables using vended credentials. This enables Snowflake to query Iceberg tables (including Delta tables with UniForm) stored in Databricks.

## Prerequisites

Before starting, ensure:

1. **Valid Snowflake account exists** (check `~/.vibe/snowflake/environment`)
   ```bash
   source ~/.vibe/snowflake/environment
   snow connection test
   ```

2. **Snowflake CLI is authenticated** (run `snow connection test`)

3. **Valid FE-VM Databricks workspace** - either existing or newly created

> **IMPORTANT: Always use an FE-VM workspace for Snowflake demos.** FE-VM workspaces are specifically provisioned for demos and can be safely configured for external connectivity. Never use customer workspaces or production workspaces for Snowflake integration demos.

### Getting an FE-VM Workspace

Use the `databricks-fe-vm-workspace-deployment` skill to get a workspace:

```bash
# Check for existing workspaces
python3 $VIBE_HOME/plugins/databricks-tools/skills/databricks-fe-vm-workspace-deployment/resources/environment_manager.py list

# Find a suitable workspace (serverless, at least 7 days remaining)
python3 $VIBE_HOME/plugins/databricks-tools/skills/databricks-fe-vm-workspace-deployment/resources/environment_manager.py find --type serverless --min-days 7

# Or deploy a new one if needed
python3 $VIBE_HOME/plugins/databricks-tools/skills/databricks-fe-vm-workspace-deployment/resources/fe_vm_client.py deploy-serverless --name snowflake-demo --region us-east-1 --lifetime 30
```

Once you have a workspace URL, authenticate the Databricks CLI:

```bash
databricks auth login <workspace_url> --profile=fe-vm-snowflake
```

### Disable IP Access Lists (Required for Snowflake Connectivity)

For Snowflake to communicate with the Databricks workspace using vended credentials, you must disable IP access lists (IP ACLs) on the workspace.

> **WARNING: Only disable IP ACLs on FE-VM workspaces.** Never disable IP ACLs on customer workspaces, production workspaces, or any workspace that contains sensitive data. FE-VM workspaces are ephemeral demo environments specifically designed for this purpose.

```bash
# Disable IP access lists
databricks workspace-conf set-status --json '{"enableIpAccessLists": "false"}' --profile=fe-vm-snowflake
```

Verify the setting was applied:

```bash
# Confirm IP ACLs are disabled
databricks workspace-conf get-status enableIpAccessLists --profile=fe-vm-snowflake
```

Expected output:
```json
{
  "enableIpAccessLists":"false"
}
```

---

## Step 1: Create Catalog, Schema, and Tables in Databricks

First, create a SQL warehouse to execute commands:

```bash
# List existing warehouses
databricks warehouses list --profile=fe-vm-snowflake

# Or create a new serverless warehouse
databricks api post /api/2.0/sql/warehouses --profile=fe-vm-snowflake --json='{
  "name": "snowflake_demo_wh",
  "cluster_size": "2X-Small",
  "warehouse_type": "PRO",
  "enable_serverless_compute": true,
  "auto_stop_mins": 10
}'
```

Save the warehouse ID for subsequent queries.

### Create Catalog and Schema

```sql
-- Create a dedicated catalog for the demo
CREATE CATALOG IF NOT EXISTS snowflake_demo_catalog;

-- Create a schema within the catalog
CREATE SCHEMA IF NOT EXISTS snowflake_demo_catalog.demo_schema;

-- Set as default context
USE CATALOG snowflake_demo_catalog;
USE SCHEMA demo_schema;
```

Execute via CLI:

```bash
WAREHOUSE_ID="<your-warehouse-id>"

databricks api post /api/2.0/sql/statements --profile=fe-vm-snowflake --json='{
  "statement": "CREATE CATALOG IF NOT EXISTS snowflake_demo_catalog",
  "warehouse_id": "'$WAREHOUSE_ID'",
  "wait_timeout": "30s"
}'

databricks api post /api/2.0/sql/statements --profile=fe-vm-snowflake --json='{
  "statement": "CREATE SCHEMA IF NOT EXISTS snowflake_demo_catalog.demo_schema",
  "warehouse_id": "'$WAREHOUSE_ID'",
  "wait_timeout": "30s"
}'
```

### Create Managed Iceberg Table

Create a native Iceberg table:

```sql
CREATE TABLE snowflake_demo_catalog.demo_schema.iceberg_sales (
  sale_id BIGINT,
  product_name STRING,
  quantity INT,
  unit_price DECIMAL(10,2),
  sale_date DATE,
  region STRING
)
USING iceberg;
```

Insert sample data:

```sql
INSERT INTO snowflake_demo_catalog.demo_schema.iceberg_sales VALUES
  (1, 'Widget A', 100, 29.99, '2026-01-01', 'North'),
  (2, 'Widget B', 75, 49.99, '2026-01-02', 'South'),
  (3, 'Gadget X', 200, 19.99, '2026-01-03', 'East'),
  (4, 'Gadget Y', 150, 39.99, '2026-01-04', 'West'),
  (5, 'Widget A', 80, 29.99, '2026-01-05', 'North'),
  (6, 'Widget B', 120, 49.99, '2026-01-06', 'South'),
  (7, 'Gadget X', 90, 19.99, '2026-01-07', 'East'),
  (8, 'Gadget Y', 110, 39.99, '2026-01-08', 'West'),
  (9, 'Premium Z', 50, 99.99, '2026-01-09', 'North'),
  (10, 'Premium Z', 45, 99.99, '2026-01-10', 'South');
```

### Create Delta Table with UniForm (Iceberg v3)

Create a Delta table with UniForm enabled for Iceberg compatibility:

```sql
CREATE TABLE snowflake_demo_catalog.demo_schema.uniform_inventory (
  inventory_id BIGINT,
  product_name STRING,
  warehouse_location STRING,
  quantity_on_hand INT,
  reorder_point INT,
  last_updated TIMESTAMP
)
TBLPROPERTIES (
  'delta.columnMapping.mode' = 'name',
  'delta.enableIcebergCompatV2' = 'true',
  'delta.universalFormat.enabledFormats' = 'iceberg'
);
```

Insert sample data:

```sql
INSERT INTO snowflake_demo_catalog.demo_schema.uniform_inventory VALUES
  (1, 'Widget A', 'Warehouse-01', 500, 100, '2026-01-07 10:00:00'),
  (2, 'Widget B', 'Warehouse-01', 300, 75, '2026-01-07 10:00:00'),
  (3, 'Gadget X', 'Warehouse-02', 1000, 200, '2026-01-07 10:00:00'),
  (4, 'Gadget Y', 'Warehouse-02', 750, 150, '2026-01-07 10:00:00'),
  (5, 'Premium Z', 'Warehouse-01', 200, 50, '2026-01-07 10:00:00'),
  (6, 'Widget A', 'Warehouse-03', 400, 100, '2026-01-07 11:00:00'),
  (7, 'Widget B', 'Warehouse-03', 250, 75, '2026-01-07 11:00:00'),
  (8, 'Gadget X', 'Warehouse-01', 800, 200, '2026-01-07 11:00:00');
```

---

## Step 2: Grant External Use Schema Permission

Grant the `EXTERNAL USE SCHEMA` privilege to allow external systems (Snowflake) to access the schema via the Iceberg REST Catalog:

```sql
GRANT EXTERNAL USE SCHEMA ON SCHEMA snowflake_demo_catalog.demo_schema TO `<your-email>`;
```

Execute via CLI:

```bash
# Get current user email
USER_EMAIL=$(databricks current-user me --profile=fe-vm-snowflake | jq -r '.userName')

databricks api post /api/2.0/sql/statements --profile=fe-vm-snowflake --json='{
  "statement": "GRANT EXTERNAL USE SCHEMA ON SCHEMA snowflake_demo_catalog.demo_schema TO `'$USER_EMAIL'`",
  "warehouse_id": "'$WAREHOUSE_ID'",
  "wait_timeout": "30s"
}'
```

---

## Step 3: Create Databricks Personal Access Token

Create a PAT that Snowflake will use to authenticate to Databricks:

```bash
# Create a PAT valid for 30 days
databricks tokens create --comment "Snowflake integration token" --lifetime-seconds 2592000 --profile=fe-vm-snowflake
```

**Save the token value securely.** The token will look like: `dapi...`

Store the token for later use:

```bash
# Save to environment file
DATABRICKS_PAT="<your-token-value>"
DATABRICKS_WORKSPACE_URL="<your-workspace-url>"  # e.g., https://fe-vm-snowflake-demo.cloud.databricks.com

# Append to snowflake environment
cat >> ~/.vibe/snowflake/environment << EOF

# Databricks Integration
DATABRICKS_WORKSPACE_URL=$DATABRICKS_WORKSPACE_URL
DATABRICKS_PAT=$DATABRICKS_PAT
DATABRICKS_CATALOG=snowflake_demo_catalog
DATABRICKS_SCHEMA=demo_schema
EOF
```

---

## Step 4: Create Catalog Integration in Snowflake

Connect to Snowflake and create the catalog integration:

```sql
CREATE OR REPLACE CATALOG INTEGRATION databricks_unity_catalog
  CATALOG_SOURCE = ICEBERG_REST
  TABLE_FORMAT = ICEBERG
  CATALOG_NAMESPACE = 'demo_schema'
  REST_CONFIG = (
    CATALOG_URI = '<workspace-url>/api/2.1/unity-catalog/iceberg-rest',
    WAREHOUSE = 'snowflake_demo_catalog',
    ACCESS_DELEGATION_MODE = VENDED_CREDENTIALS
  )
  REST_AUTHENTICATION = (
    TYPE = BEARER,
    BEARER_TOKEN = '<databricks-pat>'
  )
  ENABLED = TRUE;
```

Execute via Snowflake CLI:

```bash
source ~/.vibe/snowflake/environment

snow sql -q "
CREATE OR REPLACE CATALOG INTEGRATION databricks_unity_catalog
  CATALOG_SOURCE = ICEBERG_REST
  TABLE_FORMAT = ICEBERG
  CATALOG_NAMESPACE = 'demo_schema'
  REST_CONFIG = (
    CATALOG_URI = '$DATABRICKS_WORKSPACE_URL/api/2.1/unity-catalog/iceberg-rest',
    WAREHOUSE = 'snowflake_demo_catalog',
    ACCESS_DELEGATION_MODE = VENDED_CREDENTIALS
  )
  REST_AUTHENTICATION = (
    TYPE = BEARER,
    BEARER_TOKEN = '$DATABRICKS_PAT'
  )
  ENABLED = TRUE;
"
```

Verify the integration:

```bash
snow sql -q "DESCRIBE CATALOG INTEGRATION databricks_unity_catalog"
```

---

## Step 5: Create Catalog-Linked Database in Snowflake

Create a database that links to the Databricks Unity Catalog:

```sql
CREATE OR REPLACE DATABASE databricks_linked_db
  FROM CATALOG INTEGRATION databricks_unity_catalog;
```

Execute via CLI:

```bash
snow sql -q "CREATE OR REPLACE DATABASE databricks_linked_db FROM CATALOG INTEGRATION databricks_unity_catalog"
```

---

## Step 6: Check Catalog Sync Status

Snowflake will automatically discover and sync tables from the linked catalog. Check the sync status:

```bash
snow sql -q "SELECT SYSTEM\$CATALOG_LINK_STATUS('databricks_linked_db')"
```

You can also list the discovered tables:

```bash
# List schemas (namespaces from Databricks)
snow sql -q "SHOW SCHEMAS IN DATABASE databricks_linked_db"

# List Iceberg tables
snow sql -q "SHOW ICEBERG TABLES IN DATABASE databricks_linked_db"
```

Wait for tables to sync (typically 30-60 seconds for initial discovery).

---

## Step 7: Query Databricks Tables from Snowflake

Once synced, you can query the Databricks tables directly from Snowflake:

```bash
# Query the native Iceberg table
snow sql -q "SELECT * FROM databricks_linked_db.demo_schema.iceberg_sales ORDER BY sale_id"

# Query the UniForm Delta table (appears as Iceberg to Snowflake)
snow sql -q "SELECT * FROM databricks_linked_db.demo_schema.uniform_inventory ORDER BY inventory_id"

# Run analytics across both tables
snow sql -q "
SELECT
    s.region,
    s.product_name,
    SUM(s.quantity) as total_sold,
    SUM(s.quantity * s.unit_price) as total_revenue,
    i.quantity_on_hand as current_inventory
FROM databricks_linked_db.demo_schema.iceberg_sales s
JOIN databricks_linked_db.demo_schema.uniform_inventory i
    ON s.product_name = i.product_name
GROUP BY s.region, s.product_name, i.quantity_on_hand
ORDER BY total_revenue DESC
"
```

---

## Troubleshooting

### Catalog Integration Fails

```bash
# Check integration status
snow sql -q "SHOW CATALOG INTEGRATIONS"
snow sql -q "DESCRIBE CATALOG INTEGRATION databricks_unity_catalog"
```

Common issues:
- **Invalid token**: Ensure the PAT hasn't expired
- **Wrong workspace URL**: Must include `https://` prefix
- **Missing permissions**: Verify `EXTERNAL USE SCHEMA` was granted

### Tables Not Syncing

```bash
# Check sync status for details
snow sql -q "SELECT SYSTEM\$CATALOG_LINK_STATUS('databricks_linked_db')"

# Check for specific table issues
snow sql -q "SHOW ICEBERG TABLES IN DATABASE databricks_linked_db"
```

Common issues:
- **Case sensitivity**: Unity Catalog uses lowercase. Use double quotes for identifiers if needed
- **Schema not exposed**: Verify `CATALOG_NAMESPACE` matches your schema name
- **Tables not Iceberg-readable**: Ensure UniForm is properly enabled for Delta tables

### Permission Denied Errors

Ensure the user who created the PAT has:
- `SELECT` on the tables
- `USE CATALOG` on the catalog
- `USE SCHEMA` on the schema
- `EXTERNAL USE SCHEMA` grant

```sql
-- In Databricks, verify permissions
SHOW GRANTS ON SCHEMA snowflake_demo_catalog.demo_schema;
```

### Connection Blocked by IP Access Lists

If you see connection timeout errors or access denied errors when Snowflake tries to connect to the Databricks workspace, IP ACLs may be blocking the connection.

Check if IP ACLs are enabled:

```bash
databricks workspace-conf get-status enableIpAccessLists --profile=fe-vm-snowflake
```

If `enableIpAccessLists` is `true`, disable it (FE-VM workspaces only):

```bash
databricks workspace-conf set-status --json '{"enableIpAccessLists": "false"}' --profile=fe-vm-snowflake
```

After disabling, retry creating the catalog integration or linked database in Snowflake.

---

## Cleanup

To remove the integration:

```bash
# In Snowflake
snow sql -q "DROP DATABASE IF EXISTS databricks_linked_db"
snow sql -q "DROP CATALOG INTEGRATION IF EXISTS databricks_unity_catalog"

# In Databricks (optional - keep if reusing)
databricks api post /api/2.0/sql/statements --profile=fe-vm-snowflake --json='{
  "statement": "DROP SCHEMA IF EXISTS snowflake_demo_catalog.demo_schema CASCADE",
  "warehouse_id": "'$WAREHOUSE_ID'",
  "wait_timeout": "30s"
}'

databricks api post /api/2.0/sql/statements --profile=fe-vm-snowflake --json='{
  "statement": "DROP CATALOG IF EXISTS snowflake_demo_catalog CASCADE",
  "warehouse_id": "'$WAREHOUSE_ID'",
  "wait_timeout": "30s"
}'
```

---

## Reference Links

- [Databricks Managed Tables](https://docs.databricks.com/aws/en/tables/managed)
- [Delta UniForm for Iceberg](https://docs.databricks.com/aws/en/delta/uniform)
- [Databricks External Access with Snowflake](https://docs.databricks.com/aws/en/external-access/iceberg)
- [Snowflake CREATE CATALOG INTEGRATION (REST)](https://docs.snowflake.com/en/sql-reference/sql/create-catalog-integration-rest)
- [Snowflake CREATE DATABASE (catalog-linked)](https://docs.snowflake.com/en/sql-reference/sql/create-database-catalog-linked)
- [Snowflake Catalog-Linked Database Guide](https://docs.snowflake.com/user-guide/tables-iceberg-catalog-linked-database)
