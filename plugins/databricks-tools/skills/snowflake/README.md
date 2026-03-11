# Snowflake Trial Account Setup

Automate the setup of Snowflake trial accounts for Field Engineering demos and testing, from burner Gmail creation through CLI configuration and Databricks integration.

## How to Invoke

### Slash Command

```
/snowflake
```

### Example Prompts

```
"Set up a Snowflake trial account for a demo"
"I need a Snowflake environment to test Iceberg table access from Databricks"
"Configure the Snowflake CLI with my trial account"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Chrome DevTools MCP | For browser automation during signup and activation |
| Homebrew | For installing the Snowflake CLI |

## What This Skill Does

1. Checks for an existing valid Snowflake account in `~/.vibe/snowflake/environment`
2. If needed, guides creation of a burner Gmail and Snowflake trial signup via browser
3. Generates a fake identity for privacy during signup
4. Activates the account and generates secure credentials
5. Stores all credentials securely in `~/.vibe/snowflake/environment`
6. Installs and configures the Snowflake CLI (`snow`) with a default connection
7. Verifies connectivity with `snow connection test`

## Key Resources

| File | Description |
|------|-------------|
| `resources/DATABRICKS_INTEGRATION.md` | Guide for querying Databricks Iceberg tables from Snowflake |

## Related Skills

- `/databricks-fe-vm-workspace-deployment` - Provision a Databricks workspace for Iceberg integration
- `/databricks-authentication` - Authenticate to Databricks for cross-platform demos
- `/databricks-demo` - Orchestrate demos that combine Snowflake and Databricks
