---
name: support-escalation
description: Create and manage support escalations for critical customer support cases
user-invocable: true
---

# Support Escalation Workflow

> **IMPORTANT:** For looking up or creating JIRA tickets (including ES tickets), use the `jira-actions` skill instead if it is available. This skill is specifically for support escalation workflows, not general JIRA ticket operations.

This skill helps you create and manage support escalations for critical customer issues.

## Overview

Support escalation is the process of elevating a customer issue beyond standard support channels when the issue requires engineering involvement, executive attention, or expedited resolution. This workflow guides you through assessing severity, gathering the right information, filing the escalation ticket, and monitoring progress.

## When to Escalate

Escalate when any of the following are true:

- **Production impact:** The customer's production workloads are blocked or severely degraded
- **Data integrity risk:** There is potential data loss, corruption, or security exposure
- **SLA breach imminent:** The issue has been open beyond your standard SLA thresholds
- **No workaround available:** Standard troubleshooting has not resolved the issue
- **Customer is at risk:** Account health or renewal is impacted by the unresolved issue
- **Recurring issue:** The same class of problem has occurred multiple times

## Escalation Process

### 1. Assess Severity

Classify the issue using your organization's severity framework. Common levels:

| Severity | Description | Example |
|----------|-------------|---------|
| P1 / Sev1 | Complete production outage, data loss | All jobs failing, workspace inaccessible |
| P2 / Sev2 | Major feature degraded, partial outage | Specific pipeline broken, significant slowdown |
| P3 / Sev3 | Non-critical degradation, workaround exists | Single query slow, UI bug with workaround |
| P4 / Sev4 | Informational, cosmetic, low-impact | Documentation gap, minor display issue |

### 2. Gather Information

Before creating the escalation ticket, collect:

- **Customer / account name** - Exact name and any relevant account IDs
- **Impact description** - What is broken, how many users/jobs are affected
- **Timeline** - When did the issue start? Is it ongoing or intermittent?
- **Environment details** - Workspace URL, cluster config, Databricks runtime version, cloud/region
- **Error messages** - Exact error text, stack traces, job run IDs
- **Reproduction steps** - How to reproduce the issue (if possible)
- **Workarounds tried** - What has already been attempted
- **Business impact** - Revenue at risk, SLA commitments, customer sentiment

### 3. Create Escalation Ticket

Use the `jira-actions` skill to create an escalation ticket in your incident management system (e.g., JIRA):

```
/jira-actions
```

Ask the agent to create a new ticket with:
- **Project:** Your escalation project (e.g., ES, SUPPORT, INC)
- **Summary:** `[CUSTOMER] [SEVERITY] - Brief description of issue`
- **Priority:** Match your severity assessment above
- **Description:** Include all gathered information from Step 2
- **Labels/Components:** Tag appropriately for routing (e.g., product area, cloud)
- **Links:** Link to any existing support cases, Slack threads, or related tickets

### 4. Notify Stakeholders

After the ticket is created:

1. **Post in your incident management channel** (e.g., Slack) with a link to the ticket
2. **Notify the customer-facing team** (AE, CSM) so they can communicate proactively
3. **Tag the relevant engineering team** in the ticket if you know the product area
4. **Update the customer** with ticket number and expected next steps

### 5. Monitor and Update

- Check the ticket for engineering responses and update the customer accordingly
- Post status updates at agreed intervals (e.g., every 2 hours for P1)
- Document workarounds in the ticket as they are found
- Once resolved, confirm with the customer before closing

## Severity Levels

| Level | Response Time | Update Cadence | Escalation Path |
|-------|--------------|----------------|-----------------|
| P1 | Immediate | Every 1-2 hours | Engineering + management |
| P2 | Within 4 hours | Every 4-8 hours | Engineering team |
| P3 | Within 1 business day | Daily | Standard support |
| P4 | Within 3 business days | As needed | Standard support |

## Escalation Channels

Configure these for your organization:

- **Incident ticket system:** Use `jira-actions` skill to create/manage tickets
- **Real-time communication:** Your incident management Slack channel or equivalent
- **Customer communication:** Via your CRM or email
- **Internal escalation:** Your management or on-call rotation as appropriate

## Follow-up

After resolution:

1. Confirm the fix with the customer and close the ticket
2. Consider creating an RCA document using the `draft-rca` skill for significant incidents
3. Document any new workarounds or known issues for future reference
4. Conduct a brief retrospective for P1/P2 incidents to identify process improvements
