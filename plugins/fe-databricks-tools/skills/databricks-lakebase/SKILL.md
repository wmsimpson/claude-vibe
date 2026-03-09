---
name: databricks-lakebase
description: Create, configure, and query Lakebase Postgres databases using CLI and code
---

# Databricks Lakebase Skill

Create and manage Lakebase Postgres databases on Databricks. Lakebase provides fully-managed PostgreSQL with automatic scaling, branching, and Unity Catalog integration.

## Tier Selection

Lakebase has two tiers. **Always default to Autoscaling** unless the user specifically requests Provisioned. If unclear, ask the user which tier they want and recommend Autoscaling.

| Feature | Autoscaling Tier (Preferred) | Provisioned Tier (Legacy) |
|---------|------------------------------|--------------------------|
| CLI Command | `databricks postgres` | `databricks database` |
| CLI Version Required | **v0.285.0+** | v0.229.0+ |
| Capacity | 0.5-32 CU (auto) | Fixed CU_1, CU_2, CU_4, CU_8 |
| Scale-to-Zero | Yes | No |
| Branching | First-class branches | Via PITR only |
| Resource Model | Hierarchical (Project > Branch > Endpoint) | Flat (Instance) |
| `databricks psql` | Not supported (use direct psql) | Supported |
| Read Replicas | Yes (read-only endpoints) | Yes |

For **Provisioned Tier** instructions, see `references/provisioned.md` in this skill directory.

## Prerequisites

1. **FE-VM Workspace** - Required for Lakebase
   - Use `/databricks-fe-vm-workspace-deployment` skill to get a workspace
   - Need "serverless" workspace type for Lakebase support

2. **Databricks CLI** - **Version 0.285.0+** required for Autoscaling Tier
   - Check version: `databricks --version`
   - If below 0.285.0, upgrade: see [CLI install docs](https://docs.databricks.com/aws/en/dev-tools/cli/install)
   - Authenticate: `databricks auth login --host <workspace-url> --profile <profile-name>`

3. **psql client** (optional but recommended)
   - macOS: `brew install postgresql@16`
   - Linux: `apt install postgresql-client`

## Autoscaling Resource Model

Autoscaling Lakebase uses a hierarchical resource model:

```
Project (e.g., projects/my-app)
  └── Branch (e.g., projects/my-app/branches/production)
       └── Endpoint (e.g., projects/my-app/branches/production/endpoints/primary)
```

- **Project**: Top-level container. Creating a project auto-creates a `production` branch and `primary` read-write endpoint.
- **Branch**: Database branch (like git). Can be created from a source branch. Has its own data.
- **Endpoint**: Compute endpoint attached to a branch. Has a host for connections. Types: `ENDPOINT_TYPE_READ_WRITE` or `ENDPOINT_TYPE_READ_ONLY`.

Resource IDs must be 3-63 characters, start with a lowercase letter, and contain only lowercase letters, numbers, and hyphens.

## Quick Reference: CLI Commands

```bash
# List all projects
databricks postgres list-projects -p PROFILE

# Get project details
databricks postgres get-project projects/PROJECT_ID -p PROFILE

# Create project (auto-creates production branch + primary endpoint)
databricks postgres create-project PROJECT_ID \
  --json '{"spec": {"display_name": "My Project"}}' \
  -p PROFILE

# List branches
databricks postgres list-branches projects/PROJECT_ID -p PROFILE

# Create branch from source
databricks postgres create-branch projects/PROJECT_ID BRANCH_ID \
  --json '{"spec": {"source_branch": "projects/PROJECT_ID/branches/production", "no_expiry": true}}' \
  -p PROFILE

# List endpoints
databricks postgres list-endpoints projects/PROJECT_ID/branches/BRANCH_ID -p PROFILE

# Create endpoint on branch
databricks postgres create-endpoint projects/PROJECT_ID/branches/BRANCH_ID ENDPOINT_ID \
  --json '{"spec": {"endpoint_type": "ENDPOINT_TYPE_READ_WRITE", "autoscaling_limit_min_cu": 0.5, "autoscaling_limit_max_cu": 2.0}}' \
  -p PROFILE

# Generate OAuth credential for connection
databricks postgres generate-database-credential \
  projects/PROJECT_ID/branches/BRANCH_ID/endpoints/ENDPOINT_ID \
  -p PROFILE

# Scale endpoint
databricks postgres update-endpoint \
  projects/PROJECT_ID/branches/BRANCH_ID/endpoints/ENDPOINT_ID \
  "spec.autoscaling_limit_min_cu,spec.autoscaling_limit_max_cu" \
  --json '{"spec": {"autoscaling_limit_min_cu": 1.0, "autoscaling_limit_max_cu": 8.0}}' \
  -p PROFILE

# Delete branch (cascades to its endpoints)
databricks postgres delete-branch projects/PROJECT_ID/branches/BRANCH_ID -p PROFILE

# Delete project (cascades to all branches and endpoints)
databricks postgres delete-project projects/PROJECT_ID -p PROFILE
```

## Creating a Lakebase Project

### Step 1: Create the Project

```bash
databricks postgres create-project my-app \
  --json '{"spec": {"display_name": "My Application"}}' \
  --no-wait \
  -p PROFILE
```

This automatically creates:
- A `production` branch
- A `primary` read-write endpoint (default: 1 CU min/max)

### Step 2: Verify Project is Ready

```bash
# Check project
databricks postgres get-project projects/my-app -p PROFILE -o json

# Check branch state (should be READY)
databricks postgres list-branches projects/my-app -p PROFILE -o json | jq '.[].status.current_state'

# Check endpoint state (should be ACTIVE)
databricks postgres list-endpoints projects/my-app/branches/production -p PROFILE -o json | jq '.[].status.current_state'
```

### Step 3: Get Connection Host

```bash
# Get endpoint host for connections
databricks postgres list-endpoints projects/my-app/branches/production \
  -p PROFILE -o json | jq -r '.[0].status.hosts.host'
```

## Connecting to Lakebase (Autoscaling)

> **Important:** `databricks psql` does NOT work with Autoscaling projects. Use direct psql with an OAuth token.

### Option 1: Direct psql with OAuth Token (Recommended)

```bash
# Get endpoint host
HOST=$(databricks postgres list-endpoints projects/my-app/branches/production \
  -p PROFILE -o json | jq -r '.[0].status.hosts.host')

# Generate OAuth token
TOKEN=$(databricks postgres generate-database-credential \
  projects/my-app/branches/production/endpoints/primary \
  -p PROFILE -o json | jq -r '.token')

# Get your email
EMAIL=$(databricks current-user me -p PROFILE -o json | jq -r '.userName')

# Connect
PGPASSWORD=$TOKEN psql "host=$HOST port=5432 dbname=postgres user=$EMAIL sslmode=require"

# Run a single command
PGPASSWORD=$TOKEN psql "host=$HOST port=5432 dbname=postgres user=$EMAIL sslmode=require" \
  -c "SELECT version();"

# Connect to a specific database
PGPASSWORD=$TOKEN psql "host=$HOST port=5432 dbname=mydb user=$EMAIL sslmode=require" \
  -c "SELECT * FROM users;"
```

### Option 2: Python with psycopg2

```python
import subprocess
import json
import psycopg2

def get_autoscaling_connection(project_id: str, branch_id: str, endpoint_id: str,
                                profile: str, database: str = "postgres"):
    """Get a connection to Autoscaling Lakebase using OAuth."""
    endpoint_path = f"projects/{project_id}/branches/{branch_id}/endpoints/{endpoint_id}"
    branch_path = f"projects/{project_id}/branches/{branch_id}"

    # Get endpoint host
    result = subprocess.run([
        'databricks', 'postgres', 'list-endpoints', branch_path,
        '--profile', profile, '--output', 'json'
    ], capture_output=True, text=True)
    endpoints = json.loads(result.stdout)
    host = endpoints[0]['status']['hosts']['host']

    # Generate credentials
    result = subprocess.run([
        'databricks', 'postgres', 'generate-database-credential', endpoint_path,
        '--profile', profile, '--output', 'json'
    ], capture_output=True, text=True)
    token = json.loads(result.stdout)['token']

    # Get user email
    result = subprocess.run([
        'databricks', 'current-user', 'me',
        '--profile', profile, '--output', 'json'
    ], capture_output=True, text=True)
    email = json.loads(result.stdout)['userName']

    return psycopg2.connect(
        host=host, port=5432, database=database,
        user=email, password=token, sslmode='require'
    )

# Usage
conn = get_autoscaling_connection("my-app", "production", "primary", "my-profile", "mydb")
cur = conn.cursor()
cur.execute("SELECT * FROM users")
print(cur.fetchall())
conn.close()
```

### Option 3: SQLAlchemy

```python
from sqlalchemy import create_engine, text
from urllib.parse import quote_plus

# Using the same helper pattern as above to get host, token, email...
url = f"postgresql://{email}:{quote_plus(token)}@{host}:5432/{database}?sslmode=require"
engine = create_engine(url)

with engine.connect() as conn:
    result = conn.execute(text("SELECT * FROM users"))
    for row in result:
        print(row)
```

## Creating Databases and Tables

```bash
# Set connection variables (reuse for multiple commands)
HOST=$(databricks postgres list-endpoints projects/my-app/branches/production \
  -p PROFILE -o json | jq -r '.[0].status.hosts.host')
TOKEN=$(databricks postgres generate-database-credential \
  projects/my-app/branches/production/endpoints/primary \
  -p PROFILE -o json | jq -r '.token')
EMAIL=$(databricks current-user me -p PROFILE -o json | jq -r '.userName')

# Create a database (required before creating tables)
PGPASSWORD=$TOKEN psql "host=$HOST port=5432 dbname=postgres user=$EMAIL sslmode=require" \
  -c "CREATE DATABASE myapp;"

# Create tables in the new database
PGPASSWORD=$TOKEN psql "host=$HOST port=5432 dbname=myapp user=$EMAIL sslmode=require" -c "
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    total DECIMAL(10,2),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);
"
```

> **Note:** You must `CREATE DATABASE` first. The default `postgres` database has a restricted `public` schema.

## Branching

Branches let you create isolated copies of your database (like git branches).

### Create a Branch

```bash
databricks postgres create-branch projects/my-app dev \
  --json '{"spec": {"source_branch": "projects/my-app/branches/production", "no_expiry": true}}' \
  -p PROFILE
```

The branch is created from a point-in-time snapshot of the source branch.

### Create an Endpoint on a Branch

New branches don't have endpoints by default. Create one to connect:

```bash
databricks postgres create-endpoint projects/my-app/branches/dev read-write \
  --json '{"spec": {"endpoint_type": "ENDPOINT_TYPE_READ_WRITE", "autoscaling_limit_min_cu": 0.5, "autoscaling_limit_max_cu": 2.0}}' \
  -p PROFILE
```

### Protect a Branch

```bash
databricks postgres update-branch projects/my-app/branches/production \
  "spec.is_protected" \
  --json '{"spec": {"is_protected": true}}' \
  -p PROFILE
```

### Delete a Branch

```bash
# Unprotect first if protected
databricks postgres update-branch projects/my-app/branches/dev \
  "spec.is_protected" \
  --json '{"spec": {"is_protected": false}}' \
  -p PROFILE

# Delete (cascades to all endpoints on the branch)
databricks postgres delete-branch projects/my-app/branches/dev -p PROFILE
```

> **Note:** Read-write endpoints cannot be deleted individually. They are deleted when the branch is deleted.

## Scaling Endpoints

```bash
# Scale up the primary endpoint
databricks postgres update-endpoint \
  projects/my-app/branches/production/endpoints/primary \
  "spec.autoscaling_limit_min_cu,spec.autoscaling_limit_max_cu" \
  --json '{"spec": {"autoscaling_limit_min_cu": 1.0, "autoscaling_limit_max_cu": 8.0}}' \
  -p PROFILE
```

Autoscaling range: 0.5 to 32 CU. The endpoint scales automatically within the configured min/max.

### Add Read Replicas

```bash
databricks postgres create-endpoint projects/my-app/branches/production read-replica \
  --json '{"spec": {"endpoint_type": "ENDPOINT_TYPE_READ_ONLY", "autoscaling_limit_min_cu": 0.5, "autoscaling_limit_max_cu": 2.0}}' \
  -p PROFILE
```

## Unity Catalog Integration

### Register Lakebase Database in Unity Catalog

```bash
databricks database create-database-catalog my_catalog my-lakebase myapp \
  --create-database-if-not-exists \
  -p PROFILE
```

This allows querying Lakebase tables from Databricks SQL and notebooks:

```sql
-- In Databricks SQL
SELECT * FROM my_catalog.public.users;
```

### Create Synced Tables (Reverse ETL)

Sync data from Delta Lake to Lakebase for low-latency serving:

```bash
databricks database create-synced-database-table main.default.user_profiles \
  --database-instance-name my-lakebase \
  --logical-database-name myapp \
  -p PROFILE
```

## Data API (REST)

Lakebase provides a PostgREST-compatible REST API for direct HTTP access.

### Enable Data API

```sql
-- Connect to your database first, then run:

-- Create authenticator role
CREATE ROLE authenticator LOGIN NOINHERIT;

-- Create API role
CREATE ROLE api_user NOLOGIN;
GRANT api_user TO authenticator;
GRANT USAGE ON SCHEMA public TO api_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO api_user;
```

### Query via REST

```bash
# Get OAuth token
TOKEN=$(databricks postgres generate-database-credential \
  projects/my-app/branches/production/endpoints/primary \
  -p PROFILE -o json | jq -r '.token')

# Get endpoint host
HOST=$(databricks postgres list-endpoints projects/my-app/branches/production \
  -p PROFILE -o json | jq -r '.[0].status.hosts.host')

# Get workspace ID
WORKSPACE_ID=$(databricks current-user me -p PROFILE | jq -r '.id')

# Query users table
curl -H "Authorization: Bearer $TOKEN" \
  "https://$HOST/api/2.0/workspace/$WORKSPACE_ID/rest/myapp/public/users"

# Insert data
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}' \
  "https://$HOST/api/2.0/workspace/$WORKSPACE_ID/rest/myapp/public/users"
```

## Deleting a Project

```bash
# Delete project (PERMANENT - cascades to ALL branches, endpoints, and data!)
databricks postgres delete-project projects/my-app -p PROFILE
```

## Troubleshooting

### "unknown command postgres"
- Your Databricks CLI is below v0.285.0. Check with `databricks --version` and upgrade.

### Endpoint state not ACTIVE
- Project creation takes 1-2 minutes. Check: `databricks postgres list-endpoints projects/PROJECT/branches/BRANCH -p PROFILE -o json | jq '.[].status.current_state'`

### Connection refused
- Ensure endpoint state is ACTIVE
- Verify you're using the correct host from `status.hosts.host`

### "permission denied for schema public"
- You need to `CREATE DATABASE` first and work in that database. The default `postgres` database has a restricted `public` schema.

### Authentication failed
- OAuth tokens expire after 1 hour
- Regenerate: `databricks postgres generate-database-credential <endpoint-path> -p PROFILE`

### "Could not find psql"
- Install PostgreSQL client: `brew install postgresql@16`
- Add to PATH: `export PATH="/opt/homebrew/opt/postgresql@16/bin:$PATH"`

### Endpoint ID validation error
- Endpoint IDs must be 3-63 characters, start with a lowercase letter, and contain only lowercase letters, numbers, and hyphens.

### Cannot delete read-write endpoint
- Read-write endpoints cannot be deleted individually. Delete the branch instead, which cascades to all its endpoints.

## Pricing (AWS)

| Resource | Cost |
|----------|------|
| Compute | $0.111 per CU-hour |
| Storage | $0.35 per GB-month |
| PITR Storage | $0.20 per GB-month |

## Related Skills & Agents

- `/databricks-fe-vm-workspace-deployment` - Create FE-VM workspace with Lakebase support
- `/databricks-apps` - Build apps with Lakebase backend
- `/databricks-resource-deployment` - Deploy Lakebase via bundles
- `databricks-apps-developer` agent - Full-stack app development with Lakebase

## Full Example: E-commerce Backend (Autoscaling)

```bash
PROFILE=my-profile

# 1. Create project
databricks postgres create-project ecommerce \
  --json '{"spec": {"display_name": "E-commerce Backend"}}' \
  -p $PROFILE

# 2. Get connection details
HOST=$(databricks postgres list-endpoints projects/ecommerce/branches/production \
  -p $PROFILE -o json | jq -r '.[0].status.hosts.host')
TOKEN=$(databricks postgres generate-database-credential \
  projects/ecommerce/branches/production/endpoints/primary \
  -p $PROFILE -o json | jq -r '.token')
EMAIL=$(databricks current-user me -p $PROFILE -o json | jq -r '.userName')

# 3. Create database
PGPASSWORD=$TOKEN psql "host=$HOST port=5432 dbname=postgres user=$EMAIL sslmode=require" \
  -c "CREATE DATABASE shop;"

# 4. Create schema
PGPASSWORD=$TOKEN psql "host=$HOST port=5432 dbname=shop user=$EMAIL sslmode=require" -c "
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock INT DEFAULT 0
);

CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100)
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(id),
    total DECIMAL(10,2),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id),
    product_id INT REFERENCES products(id),
    quantity INT,
    price DECIMAL(10,2)
);

INSERT INTO products (name, price, stock) VALUES
    ('Widget', 9.99, 100),
    ('Gadget', 24.99, 50),
    ('Gizmo', 14.99, 75);
"

# 5. Verify
PGPASSWORD=$TOKEN psql "host=$HOST port=5432 dbname=shop user=$EMAIL sslmode=require" \
  -c "SELECT * FROM products;"

# 6. Scale up for production traffic
databricks postgres update-endpoint \
  projects/ecommerce/branches/production/endpoints/primary \
  "spec.autoscaling_limit_min_cu,spec.autoscaling_limit_max_cu" \
  --json '{"spec": {"autoscaling_limit_min_cu": 1.0, "autoscaling_limit_max_cu": 4.0}}' \
  -p $PROFILE
```
