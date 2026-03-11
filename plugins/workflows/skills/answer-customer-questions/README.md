# Answer Customer Questions

Draft responses to multiple customer questions from Slack, email, or Google Docs. Researches answers using public documentation, Glean, and Slack, then creates a combined Q&A Google Doc with confidence ratings and references.

## How to Invoke

### Slash Command

```
/answer-customer-questions
```

### Example Prompts

```
"Answer the customer questions in this doc: https://docs.google.com/document/d/ABC123/edit"
"Draft responses to these questions from our sync meeting: [pasted text]"
"Help me respond to these Slack questions: https://databricks.slack.com/archives/C123/p456"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` for reading source docs and creating output doc |
| Glean MCP | Active Glean MCP for internal documentation searches |
| Slack MCP | Active Slack MCP for searching recent discussions |

## What This Skill Does

1. Extracts and categorizes questions from the provided source (Google Doc, Slack thread, or raw text)
2. Researches each question using public Databricks docs, Glean, and Slack
3. Assigns confidence ratings (1-10) based on source quality and agreement
4. Compiles all findings into a single formatted Google Doc organized by category
5. Flags questions requiring further PM/engineering confirmation

## Related Skills

- `/product-question-research` - For deep-dive research on a single product question
- `/google-docs` - Used for reading Google Doc source materials
