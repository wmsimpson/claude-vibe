# Validate MCP Access

Validates that the user has active credentials for MCP connections (Slack, Glean, JIRA, Confluence) in the Logfood Databricks workspace.

## How to Invoke

### Slash Command

```
/validate-mcp-access
```

### Example Prompts

```
"Check if my MCP connections are working"
"Validate my Slack and Confluence MCP access"
"Are my MCP credentials still active?"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Databricks Auth | (optional) CLI profile must be configured if using Databricks |

## What This Skill Does

1. Retrieves the user's Databricks ID from the Logfood workspace
2. Checks credential status for each MCP connection (Slack, Glean, JIRA, Confluence)
3. Reports status in a table format (ACTIVE or NOT CONFIGURED)
4. Provides direct login URLs for any connections that need authentication
5. Confirms all MCP tools are ready once all connections show ACTIVE

## Related Skills

- `/configure-vibe` - Full environment setup including MCP prerequisites
