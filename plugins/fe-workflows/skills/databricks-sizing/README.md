# Databricks Sizing (Quicksizer)

Automated Databricks cost estimation using the Quicksizer agent. Calculates monthly DBU costs for customer workloads across ETL, Data Warehousing, ML, Interactive, Lakebase, Lakeflow Connect, and Apps use cases.

## How to Invoke

### Slash Command

```
/databricks-sizing
```

### Example Prompts

```
"Size an ETL workload on AWS Premium: 500GB daily, 10 batch jobs"
"Get a Databricks cost estimate for a customer migrating from Snowflake with 2TB daily processing"
"Run Quicksizer for Azure Premium with ML model serving and data warehousing workloads"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks Auth | Run `/databricks-authentication` and authenticate with your workspace profile |
| Python 3.10+ | Required for the Quicksizer API helper script |

## What This Skill Does

1. Authenticates with your configured Databricks profile
2. Calls the Quicksizer API with the user's workload description
3. Handles multi-turn conversations when the agent asks clarifying questions
4. Returns a monthly cost breakdown with DBU estimates per use case

## Key Resources

| File | Description |
|------|-------------|
| `resources/quicksizer_api.py` | Python helper script that calls the Quicksizer agent API |

## Related Skills

- `/databricks-authentication` - Required for workspace authentication before sizing
