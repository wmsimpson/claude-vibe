# RCA Document Template

Use this template structure when creating RCA Google Docs. Pass this content to the `google-drive` agent.

## Document Title Format

```
RCA - [Company Name] [Brief Issue Description] ([JIRA Ticket])
```

Example: `RCA - Block Iceberg Table Missing DELETE File Errors (ES-1667009)`

---

## Document Structure

### 1. Incident Summary

| | |
|-------|-------|
| Company | [Company name - NOT "Customer"] |
| JIRA Ticket | [ES-XXXXXX](link) |
| Salesforce Case | [Case XXXXXXXX](link) (if applicable) |
| Severity | [Production Impact / Degraded Performance / etc.] |
| Status | [Mitigated / Resolved / In Progress] |

NOTE: No header row, no markdown bold - Google Docs API doesn't parse markdown formatting.

---

### 2. Problem Statement

[2-4 sentences describing the issue from the company's perspective]

Example:
> Block experienced "missing DELETE file" errors when querying Iceberg tables through Databricks, causing table access failures. The same queries executed successfully in Athena, indicating the issue was specific to Databricks' handling of the Iceberg metadata.

---

### 3. Root Cause

[Describe the technical root cause]

**Triggering Conditions:**
[List the conditions that trigger this issue. If uncertain, preface with "We believe the following conditions may trigger this issue:"]

Example:
> Root Cause: Delta metadata sync issues where external Iceberg operations (compaction/orphan file cleanup via EMR/Glue) remove files without writing explicit DELETE records in Iceberg metadata. Databricks' incremental metadata conversion doesn't detect "just removed" files, leading to stale references in Delta metadata.
>
> Triggering Conditions:
> - Foreign Iceberg tables managed by external catalog (e.g., Glue)
> - External compaction or orphan file cleanup operations (e.g., via EMR)
> - Files physically removed from S3 without corresponding DELETE records in Iceberg metadata

---

### 4. Impact

[Describe the business and technical impact]

Example:
> - Affected Iceberg tables in Block's production environment became inaccessible via Databricks
> - Table queries returning errors while identical queries succeeded in Athena
> - Production data pipelines blocked for affected tables

---

### 5. Timeline

| Date | Event |
|------|-------|
| [Date] | [Company] reported the issue |
| [Date] | Issue validated/reproduced by Databricks |
| [Date] | Workaround provided |
| [Date] | Permanent fix deployed (if applicable) |

**Note:** Only include customer-facing events. Do not include internal JIRA status changes.

---

### 6. Resolution

#### Immediate Mitigation
[Describe the workaround or immediate fix provided]

Example:
> 1. Disable auto-refresh: `spark.databricks.delta.uniform.ingress.autoRefresh.enabled=false`
> 2. Perform dummy commit in EMR to advance Iceberg metadata
> 3. Execute `REFRESH TABLE <table> FORCE` in DBR 17.3

#### Verification Method
[How to verify the fix worked]

#### Long-term Recommendations
[Strategic recommendations to prevent recurrence]

---

### 7. Key Personnel

#### Databricks Team
| Name | Role |
|------|------|
| @FirstName LastName | [Role - e.g., Technical Lead] |
| @FirstName LastName | [Role - e.g., Support Engineer] |

#### [Company Name] Team
| Name | Role |
|------|------|
| [Name] | [Role] |

---

### 8. Action Items

**Include only customer-facing action items.** Exclude internal Databricks tasks (e.g., documenting workarounds in internal KB, internal process improvements).

| Action Item | Owner | Status | Notes |
|-------------|-------|--------|-------|
| [Customer-facing action description] | @Owner Name | [Completed/In Progress/Backlog] | [Additional context] |

Example:
| Action Item | Owner | Status | Notes |
|-------------|-------|--------|-------|
| Provide custom image with enhanced REFRESH FORCE | @Engineering Team | Completed | Delivered to customer |
| Evaluate fix for metadata sync generation to handle missing DELETE records | @Storage Team | Backlog | Would prevent issue for all customers with external Iceberg maintenance |
| Support migration path to UC managed Iceberg | @Customer Success | In Progress | Prevents compaction-related issues long-term |

**Excluded (internal-only):**
- "Document workaround in internal KB" - internal task
- "Update runbook" - internal task
- "Train support team" - internal task

---

### 9. References

- **JIRA:** [ES-XXXXXX](https://your-jira-instance.atlassian.net/browse/ES-XXXXXX)
- **Slack:** [#channel-name](slack-link) or [Incident Thread](slack-link)
- **Salesforce:** [Case XXXXXXXX](salesforce-link)
- **Workspace:** [Workspace ID]

---

## Formatting Reminders

1. **Company names:** Always use actual company name, never "customer" or "the customer"
2. **People:** Use `@FirstName LastName` format for Databricks employees
3. **Links:** Embed all links - don't show raw URLs
4. **Timeline:** Customer-facing events only (reported → validated → resolved)
5. **Action items:** Incorporate lessons learned as actionable items with owners
6. **Root cause conditions:** If uncertain, prefix with "We believe..."
