---
name: jira-actions
description: "Use this skill for ANY JIRA ticket operations: search, view, comment, create, or update tickets. Triggers on: 'JIRA ticket', 'create a ticket', 'search JIRA', 'file a bug', any ticket key like 'PROJ-123'."
---

# JIRA Actions Skill

Read, search, comment, and create tickets using the JIRA MCP tools.

## Prerequisites

### MCP Server Required

The JIRA MCP server must be configured in `mcp-servers.yaml` and enabled.
See the commented-out `jira` section in that file for setup instructions using
[mcp-atlassian](https://github.com/sooperset/mcp-atlassian).

If you encounter connection errors, run the `validate-mcp-access` skill.

### Configuration

Set these environment variables (or use the mcp-servers.yaml env block):
- `JIRA_URL` — your JIRA instance (e.g., `https://your-org.atlassian.net`)
- `JIRA_USERNAME` — your email address
- `JIRA_API_TOKEN` — generate at https://id.atlassian.com/manage-profile/security/api-tokens

## CRITICAL: Use the JIRA Ticket Assistant Agent

**For any non-trivial JIRA operation, IMMEDIATELY delegate to the `jira-tools:jira-ticket-assistant` agent using the Task tool.** This ensures proper handling of:
- Creating tickets with all required fields
- Looking up required field values and IDs
- Searching and analyzing tickets
- Adding properly formatted comments

```
Task(subagent_type="jira-tools:jira-ticket-assistant", prompt="<user's request>")
```

Only handle simple read operations (like viewing a single ticket by ID) directly.

## Operations

### Search Tickets

Use `jira_read_api_call` with the `issues.search` endpoint and JQL (JIRA Query Language):

```
# Search by keyword in any project
jira_read_api_call("issues.search", {"jql": "summary ~ 'keyword'", "max_results": 20})

# Search within a specific project
jira_read_api_call("issues.search", {"jql": "project = PROJ AND summary ~ 'keyword'", "max_results": 10})

# Search by assignee
jira_read_api_call("issues.search", {"jql": "assignee = currentUser() ORDER BY created DESC", "max_results": 20})

# Search by status
jira_read_api_call("issues.search", {"jql": "project = PROJ AND status = 'In Progress'", "max_results": 10})

# Search by issue type
jira_read_api_call("issues.search", {"jql": "project = PROJ AND issuetype = Bug", "max_results": 10})

# Search recent tickets
jira_read_api_call("issues.search", {"jql": "project = PROJ ORDER BY created DESC", "max_results": 20})
```

**JQL Syntax Notes:**
- Use single quotes for string values with spaces
- Common operators: `=`, `!=`, `~` (contains), `IN`, `NOT IN`, `ORDER BY`
- Custom fields use syntax: `cf[FIELDID]` (get field IDs from `issues.get_create_metadata`)

### View Ticket Details

Use `jira_read_api_call` with the `issues.get` endpoint:

```
# View ticket details
jira_read_api_call("issues.get", {"issue_id": "PROJ-123"})

# View with specific fields
jira_read_api_call("issues.get", {"issue_id": "PROJ-123", "fields": "key,summary,status,description,comment"})

# Analyze ticket content (for large tickets)
jira_read_api_call("issues.get", {"issue_id": "PROJ-123", "analysis_prompt": "What errors are reported?"})
```

### Add Comments

Use `jira_write_api_call` with the `issues.add_comment` endpoint:

```
# Add a comment (ADF format)
jira_write_api_call("issues.add_comment", {
    "issue_id": "PROJ-123",
    "comment": {
        "type": "doc",
        "version": 1,
        "content": [{
            "type": "paragraph",
            "content": [{"type": "text", "text": "Your comment here"}]
        }]
    }
})
```

### Update Existing Tickets

Use `jira_write_api_call` with the `issues.update` endpoint:

```
# Update summary
jira_write_api_call("issues.update", {
    "issue_id": "PROJ-123",
    "updates": {"summary": "Updated Summary"}
})

# Update status (uses workflow transitions automatically)
jira_write_api_call("issues.update", {
    "issue_id": "PROJ-123",
    "updates": {"status": "In Progress"}
})
```

### Create Tickets

#### STEP 1: Always look up required fields first

**Before creating ANY ticket, use `issues.get_create_metadata` to discover required fields:**

```
jira_read_api_call("issues.get_create_metadata", {
    "project": "PROJ",
    "issuetype": "Bug"
})
```

This returns all available fields including which are required and their allowed values.

#### STEP 2: Create the ticket

```
jira_write_api_call("issues.create", {
    "project": "PROJ",
    "issuetype": "Bug",
    "summary": "Brief description of the issue",
    "description": "**Problem:**\nDescribe the issue in detail.\n\n**Steps to Reproduce:**\n1. Step one\n2. Step two\n\n**Expected:**\nWhat should happen.\n\n**Actual:**\nWhat is happening.\n\n**Error:**\n```\nPaste full error message here\n```",
    "additional_fields": {
        "priority": {"name": "Medium"}
    }
})
```

**CRITICAL: Field Format for Select Fields**

Select/radio button fields must use the `{"id": "XXXXX"}` format, NOT `{"value": "..."}`.
Use the metadata lookup in Step 1 to find the correct IDs.

## Best Practices

1. **Always search first** — look for similar existing tickets before creating new ones
2. **Use descriptive summaries** — include context about the affected area
3. **Include full error messages** — do not truncate stack traces
4. **Set correct priority** — follow your team's priority guidelines
5. **Link related tickets** — reference related issues

---

## Reference: Original Databricks ES Ticket Configuration

The following was the original configuration for Databricks Engineering Support (ES) tickets.
Kept here for reference if working with Databricks JIRA (`databricks.atlassian.net`).

<details>
<summary>Databricks ES Ticket Fields (click to expand)</summary>

### Databricks ES-Specific Access Notes

Most Field Engineering members cannot create ES tickets directly. Use:
- `go/FEfileaticket` — JIRA Portal for filing tickets (preferred)
- `go/fe-break-glass` — Temporary 1-hour access via Opal for emergencies

### Required Fields for ES Incident

| Field | Key | Format |
|-------|-----|--------|
| Project | `project` | `"ES"` |
| Issue Type | `issuetype` | `"Incident"` |
| Summary | `summary` | `[CustomerName] Brief description` |
| Affects Versions | `versions` | `[{"name": "N/A"}]` |
| Support Severity | `customfield_11500` | `{"id": "XXXXX"}` |
| External Customer? | `customfield_14677` | Yes=`13344`, No=`13345` |
| ES Component | `customfield_18150` | `{"id": "XXXXX"}` |

### Severity IDs

| Severity | ID |
|----------|----|
| SEV0 Critical | `10600` |
| SEV1 High | `10601` |
| SEV2 Standard | `10602` |
| SEV3 Low | `10603` |

</details>
