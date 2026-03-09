# POC Documentation

Create and maintain comprehensive Proof of Concept documentation for customer engagements with Databricks. POC docs serve as the single source of truth for scope, success criteria, timelines, staffing, and evaluation results.

## How to Invoke

### Slash Command

```
/fe-poc-doc
```

### Example Prompts

```
"Create a POC doc for the Meta evaluation, here's the Salesforce opp: 006ABC123"
"Build a POC document for CrowdStrike evaluating Unity Catalog and Managed Iceberg"
"Update the POC doc for Discord with the latest evaluation results"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Salesforce Auth | Run `/salesforce-authentication` for opportunity and account lookups |
| Google Auth | Run `/google-auth` for Google Doc creation |
| Slack MCP (optional) | Active Slack MCP if using Slack channels as input sources |
| Glean MCP | Active Glean MCP for searching additional account context |

## What This Skill Does

1. Parses sources (Salesforce opportunity, Google Docs, Slack channels, customer name)
2. Gathers context from Salesforce, JIRA, Slack, and Glean
3. Collaborates with the user to define scope, success criteria, and timeline
4. Generates a customer-facing Google Doc with key contacts, executive alignment, business use cases, PBOs, scope and success criteria, task timeline, staffing plan, and architecture context

## Key Resources

| File | Description |
|------|-------------|
| `resources/POC_DOC_TEMPLATE.md` | Full POC document template with all sections |
| `resources/SECTION_GUIDANCE.md` | Detailed guidance for completing each section |

## Related Skills

- `/fe-poc-postmortem` - For generating post-mortem retrospectives after a POC concludes
- `/salesforce-authentication` - Required for Salesforce data access
- `/google-docs` - Used for reading existing Google Docs referenced as sources
