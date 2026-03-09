# Draft RCA

Draft Root Cause Analysis documents for customer incidents by gathering information from Salesforce, JIRA, Slack, and email, then generating a formatted Google Doc.

## How to Invoke

### Slash Command

```
/draft-rca
```

### Example Prompts

```
"Create an RCA for ES-1667009 and slack channel C0A4EHNKHLH"
"Draft an RCA for Salesforce case 5001234567"
"Generate a root cause analysis for the Block Iceberg incident"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Salesforce Auth | Run `/salesforce-authentication` for case lookups |
| Google Auth | Run `/google-auth` for Google Doc creation |
| Slack MCP | Active Slack MCP credentials for thread searches |
| Glean MCP | Active Glean MCP for searching internal docs |

## What This Skill Does

1. Parses source identifiers (JIRA tickets, Salesforce cases, Slack threads) from the request
2. Gathers data from each source using the appropriate MCP or CLI tool
3. Cross-references linked artifacts (JIRA to Salesforce, Slack links in comments)
4. Validates data sufficiency (95% confidence threshold) before proceeding
5. Generates a formatted Google Doc following the RCA template with incident summary, root cause, timeline, and action items

## Key Resources

| File | Description |
|------|-------------|
| `resources/RCA_TEMPLATE.md` | Full RCA document template with section structure |

## Related Skills

- `/support-escalation` - For creating new escalation tickets related to the incident
- `/product-question-research` - For researching product-level root causes
