# Databricks One-Env Workspace Deployment

Create and manage Databricks workspaces in the One-Env AWS account for demos requiring custom AWS integrations such as custom IAM roles, cross-account S3 access, PrivateLink, or AWS service integrations.

## How to Invoke

### Slash Command

```
/databricks-oneenv-workspace-deployment
```

### Example Prompts

```
"Create a Databricks workspace with custom VPC networking in One-Env"
"Deploy a classic workspace for a PrivateLink demo"
"Set up a workspace with Glue and Kinesis integration for a streaming demo"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| AWS Sandbox Auth | `aws-sandbox-field-eng_databricks-sandbox-admin` profile via `/aws-authentication` |
| Databricks Account Auth | `one-env-admin-aws` profile via `/databricks-authentication` |

## What This Skill Does

1. Verifies both AWS and Databricks account authentication
2. Creates AWS infrastructure (S3 buckets, IAM roles, VPCs for classic workspaces)
3. Deploys Serverless or Classic workspaces in the One-Env Databricks account
4. Assigns shared metastores and creates storage credentials
5. Sets up Unity Catalog with catalogs, external locations, and integration IAM roles
6. Configures IP ACLs and registers instance profiles
7. Tracks all resources in a local registry (`~/.vibe/oneenv/`)
8. Supports health checks, repair, and full cleanup workflows

## Key Resources

| File | Description |
|------|-------------|
| `resources/registry_manager.py` | Track and manage deployed workspace resources |
| `resources/health_checker.py` | Verify resource health and offer repair options |
| `resources/full_cleanup.py` | Delete a workspace and all associated AWS/Databricks resources |
| `resources/integration_resources.py` | Create test AWS resources (DynamoDB, Kinesis, RDS, MSK) |
| `resources/templates/` | IAM policy and VPC configuration templates |

## Related Skills

- `/databricks-fe-vm-workspace-deployment` - Simpler/faster workspace deployment (use when custom AWS control is not needed)
- `/databricks-authentication` - Authenticate to the deployed workspace
- `/aws-authentication` - Authenticate to the AWS sandbox account
