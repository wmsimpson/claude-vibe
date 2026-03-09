# Support Escalation

Create and manage support escalations for critical customer support cases. Provides a workflow for assessing severity, gathering information, creating the escalation, and monitoring progress.

## How to Invoke

### Slash Command

```
/support-escalation
```

### Example Prompts

```
"Create a support escalation for a P1 customer issue with cluster failures"
"Escalate this critical production outage for Acme Corp"
"Help me file a support escalation for a data loss incident"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| JIRA Access | Access to create and manage ES tickets |

## What This Skill Does

1. Assesses the severity of the customer issue
2. Gathers required information for the escalation
3. Creates the escalation through the appropriate channels
4. Provides guidance on monitoring and follow-up

## Related Skills

- `/jira-actions` - For general JIRA ticket operations (lookups, creation) outside of escalation workflows
- `/draft-rca` - For creating Root Cause Analysis documents after incident resolution
- `/databricks-troubleshooting` - For diagnosing the underlying technical issue
