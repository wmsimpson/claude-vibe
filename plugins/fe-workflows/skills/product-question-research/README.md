# Product Question Research

Research and answer Databricks product questions by searching public documentation, Glean, and Slack. Creates a formatted Google Doc with the answer, confidence rating, inline citations, and references.

## How to Invoke

### Slash Command

```
/product-question-research
```

### Example Prompts

```
"Does Lakeflow Connect support sinking data as Iceberg?"
"Do materialized views on foreign Iceberg tables support incrementalization?"
"What's the status of Auto CDF?"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` for creating the research output doc |
| Glean MCP | Active Glean MCP for internal documentation searches |
| Slack MCP | Active Slack MCP for searching recent product discussions |

## What This Skill Does

1. Maps the feature landscape - identifies related features, dependencies, and alternative approaches
2. Determines feature status (GA, Public Preview, Private Preview, Roadmap)
3. Searches public Databricks documentation via docs.databricks.com
4. Searches Glean for internal docs, FAQs, and preview guides
5. Identifies and searches relevant Slack channels for PM/engineering confirmations
6. Deep-dives on any roadmap or preview features discovered (status, timeline, access, limitations)
7. Validates the answer against the original question's specificity
8. Creates a Google Doc with answer, confidence rating (1-10), uncertainty breakdown, relevant channels and people, and cited references

## Related Skills

- `/fe-answer-customer-questions` - For batch-answering multiple customer questions into a single doc
- `/google-docs` - Used for reading Google Docs discovered during research
