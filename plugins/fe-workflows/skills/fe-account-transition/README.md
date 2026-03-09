# Account Transition Document Generator

Create comprehensive account transition documents for new Account Executives and Solution Architects taking over customer accounts. Gathers data from Salesforce (optional), Glean, and JIRA to produce a Google Doc with full account context. Consumption data can be pulled from your Databricks workspace or CRM if available.

## How to Invoke

### Slash Command

```
/fe-account-transition
```

### Example Prompts

```
"Create an account transition doc for Acme Corp, new AE is Jane Smith"
"Generate a transition document for the Block account handoff"
"Build an account transition brief for Meta with consumption analysis and use case summary"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Salesforce Auth | Run `/salesforce-authentication` for account and use case queries |
| Databricks Auth | Run `/databricks-authentication` with your workspace profile for consumption data (optional) |
| Google Auth | Run `/google-auth` for Google Doc creation |
| Glean MCP | Active Glean MCP for document and Slack channel discovery |

## What This Skill Does

1. Looks up the account and use cases in Salesforce (optional)
2. Queries consumption data from your Databricks workspace or CRM (optional, TODO: configure for your data source)
3. Searches Glean for internal documents, Slack channels, and JIRA tickets
4. Compiles a structured Google Doc covering account overview, team contacts, stakeholder map, use cases (U1-U6), engineering escalations, consumption analysis, and key resources
5. Post-processes the doc with person chips and color-coded status indicators

## Key Resources

| File | Description |
|------|-------------|
| `resources/postprocess_transition_doc.py` | Script to add person chips and status colors to the Google Doc |

## Related Skills

- `/salesforce-authentication` - Required for Salesforce data access
- `/databricks-authentication` - For Databricks workspace consumption queries (optional)
- `/google-auth` - Required for Google Doc creation
- `/google-docs` - Used for reading Google Docs found via Glean
