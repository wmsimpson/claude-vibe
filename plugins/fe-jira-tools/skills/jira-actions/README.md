# JIRA Actions

Search, view, comment on, and create Engineering Support (ES) tickets using the JIRA MCP tools.

## How to Invoke

### Slash Command

```
/jira-actions
```

### Example Prompts

```
"Search for ES tickets related to MLflow model registry"
"Show me the details of ES-123456"
"Create an ES incident ticket for a Delta table corruption issue"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| JIRA MCP | Configured automatically via vibe setup; run `validate-mcp-access` if issues arise |
| ES Access | Most FE members must use go/FEfileaticket for ticket creation; direct access requires FE.Exception.ES.Access group |

## What This Skill Does

1. **Search** -- Query ES tickets with JQL (by keyword, customer, component, issue type, or date)
2. **View** -- Retrieve full ticket details including status, description, and comments
3. **Comment** -- Add comments to existing tickets in Atlassian Document Format
4. **Update** -- Modify ticket fields like summary, description, priority, and status
5. **Create** -- File new ES incidents or service requests (access-restricted; falls back to go/FEfileaticket portal)

For non-trivial operations, the skill delegates to the `jira-ticket-assistant` agent for proper field lookups and validation.

## Key Resources

| File | Description |
|------|-------------|
| `resources/ES_TICKET_TYPES.md` | Issue type reference (Incident, Advanced Support, etc.) |
| `resources/ES_SEVERITY_LEVELS.md` | SEV0-SEV3 severity level definitions |
| `resources/ES_KEY_FIELDS.md` | Common custom fields and their keys |
| `resources/FE_ACCESS_WORKFLOW.md` | Access restriction details and workarounds |

## Related Skills

- `/validate-mcp-access` - Verify JIRA MCP connection is working
