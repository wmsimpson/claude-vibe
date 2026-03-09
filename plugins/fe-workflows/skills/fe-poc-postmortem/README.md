# POC Post-Mortem Retrospective

Generate comprehensive post-mortem retrospective documents for competitive Databricks POCs that were not won. Captures learnings to help the Field Engineering team improve future engagements.

## How to Invoke

### Slash Command

```
/fe-poc-postmortem
```

### Example Prompts

```
"Create a POC postmortem for the Pinterest deal, here's the internal POC doc: [Google Doc URL]"
"Generate a postmortem for the Block POC loss, Slack channel is #block-poc-internal"
"POC postmortem for Acme Corp opportunity 006ABC123"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` for reading source docs and creating the retrospective |
| Slack MCP (optional) | Active Slack MCP if using Slack channels as input sources |
| Glean MCP | Active Glean MCP for searching additional context and researching technical concepts |

## What This Skill Does

1. Parses sources (Google Docs, Salesforce opportunities, JIRA tickets, Slack channels)
2. Validates MCP access for Slack sources if provided
3. Gathers data from all sources and cross-references findings
4. Researches technical jargon and concepts found in source materials for clarity
5. Synthesizes findings into a narrative retrospective covering evaluation history, competition analysis, challenges and learnings, and actionable recommendations
6. Generates a formatted Google Doc following the retrospective template

## Key Resources

| File | Description |
|------|-------------|
| `resources/POC_RETROSPECTIVE_TEMPLATE.md` | Full retrospective document template |

## Related Skills

- `/fe-poc-doc` - For creating POC documentation before or during an engagement
- `/product-question-research` - For researching technical questions that arise during retrospective analysis
