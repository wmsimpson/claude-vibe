# Databricks Feature Tester

End-to-end testing of Databricks features, feature interactions, and external integrations. Handles the full workflow from research to infrastructure provisioning to test execution to documentation in a Google Doc.

## How to Invoke

### Slash Command

```
/fe-databricks-feature-tester
```

### Example Prompts

```
"Test if materialized views support incremental refresh on Delta tables"
"Test Lakeflow Connect streaming into Delta tables with UniForm enabled"
"Test DynamoDB integration with Databricks using boto3"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks Auth | Run `/databricks-authentication` for workspace access |
| Google Auth | Run `/google-auth` for creating the test report doc |
| AWS Auth (optional) | Run `/aws-authentication` for tests involving AWS resources |

## What This Skill Does

1. **Research & Planning** - Researches the feature via Glean, determines infrastructure requirements, and selects the appropriate workspace (e2-demo, FE-VM, or One-Env)
2. **Parallel Setup** - Provisions workspace infrastructure while developing test code
3. **Test Execution** - Validates prerequisites, runs tests with clear success/failure criteria, and iterates on setup failures vs actual feature failures
4. **Documentation** - Compiles findings into a structured Google Doc with environment details, test results, confidence assessment, and recommendations

## Related Skills

- `/product-question-research` - For deep research on features before testing
- `/databricks-fe-vm-workspace-deployment` - For provisioning FE-VM workspaces
- `/databricks-oneenv-workspace-deployment` - For provisioning One-Env workspaces (required for AWS resource integrations)
- `/aws-authentication` - For tests involving AWS services (RDS, DynamoDB, Kinesis, etc.)
- `/fe-snowflake` - For tests involving Snowflake integration
