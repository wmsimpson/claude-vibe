# JIRA Ticket Access Workflow

This document describes the workflow for working with JIRA tickets given typical project access restrictions.

## Access Levels

### Standard Access (Most Users)
- **Can:** Search, view, comment, clone tickets
- **Cannot:** Create tickets in restricted projects directly
- **Must use:** Project-specific submission portal or request access from admin

### Admin / Exception Access
- **Can:** All operations including direct ticket creation
- **How to get:** Request from your JIRA admin or project lead

### Temporary / Break-Glass Access
- **Use case:** Emergencies, one-off urgent tickets
- **How:** Request temporary elevation from your JIRA admin

## Workflow Decision Tree

```
Need to work with a JIRA ticket?
│
├─ View/Search existing ticket?
│  └─ YES: Use JIRA MCP jira_read_api_call (no restrictions)
│
├─ Add comment to existing ticket?
│  └─ YES: Use JIRA MCP jira_write_api_call with issues.add_comment
│
├─ Edit existing ticket?
│  └─ YES: Use JIRA MCP jira_write_api_call with issues.update
│
├─ Create NEW ticket?
│  │
│  ├─ Have create permission on the project?
│  │  └─ YES: Use JIRA MCP jira_write_api_call with issues.create
│  │
│  ├─ Is this a critical/SEV0 issue?
│  │  └─ YES: Contact your support team or on-call directly
│  │
│  ├─ Project has submission portal?
│  │  └─ YES: Use the portal (often creates linked CRM case + JIRA ticket)
│  │
│  └─ Similar ticket exists?
│     └─ YES: Reference it; ask admin to create or clone
```

## Ticket Creation Checklist

Before creating any ticket, gather:
1. **Project** — Which JIRA project? (e.g., ENG, BUG, SUPPORT)
2. **Issue type** — Incident, Bug, Task, Story, etc.
3. **Summary** — Clear, concise one-liner
4. **Description** — Steps to reproduce, impact, context
5. **Priority/Severity** — How urgent? (Critical, High, Medium, Low)
6. **Affected components** — Which service or product area?
7. **Required custom fields** — Use `issues.get_create_metadata` to discover

## Discovering Required Fields

Before creating a ticket, always check what fields are required:

```
jira_read_api_call("issues.get_create_metadata", {
    "project": "YOUR_PROJECT",
    "issuetype": "Bug"
})
```

This returns all available fields and which are required. Use `{"id": "XXXXX"}` format for select fields — NOT `{"value": "..."}`.

## Finding Similar Tickets

Always search for similar tickets before creating a new one:

```
# By keyword
jira_read_api_call("issues.search", {
    "jql": "project = YOUR_PROJECT AND summary ~ 'keyword'",
    "max_results": 10
})

# By issue type + status
jira_read_api_call("issues.search", {
    "jql": "project = YOUR_PROJECT AND issuetype = 'Bug' AND status = 'Open'",
    "max_results": 10
})
```

## Severity / Priority Guidelines

| Situation | Priority |
|-----------|----------|
| Complete outage, production down, data loss | Critical / P0 |
| Major feature broken, many users impacted | High / P1 |
| Feature degraded, has workaround | Medium / P2 |
| Minor issue, cosmetic, low impact | Low / P3 |

## Do NOT

- Modify tickets without explicit user confirmation
- Delete or archive tickets
- Change priority/severity on existing tickets without user request
- Create duplicate tickets — search first
- Assign tickets to specific people unless user specifies

---

<!-- NOTE: This file was generalized from a Databricks-internal FE access workflow.
     The original described ES ticket creation restrictions, go/FEfileaticket portal,
     go/fe-break-glass Opal access, and help@databricks.com escalation.
     Update the "Project" references above to match your actual JIRA project key. -->
