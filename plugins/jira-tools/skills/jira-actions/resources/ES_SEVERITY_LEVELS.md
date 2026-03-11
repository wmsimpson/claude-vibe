# ES Ticket Severity Levels

Severity levels determine response priority and SLA expectations. Choose the appropriate level based on customer impact, not urgency.

## SEV0 - Critical

**Response:** 24/7 immediate response, interrupts all other work.

**Criteria - Must meet ONE of:**

### Service Outage
- Customer with Production or Mission Critical Support experiencing total outage
- More than 2 customers experiencing workload failures with same root cause
- Multiple production workloads failing with severe business impact

### Data Integrity (see go/data-integrity-sop)
- Data correctness issues (incomplete, inaccurate, truncated results)
- Data loss (accidental deletion)
- Data corruption (unrecoverable)

### Financial/Billing Impact
- Impact >$10k (or unknown) on customer costs
- Incorrect billing, leaked VMs, excessive egress/storage costs
- Logs data corruption or loss

### Data Security and Privacy
- Sensitive data (passwords, PII) exposed to unauthorized users
- Irrecoverable data loss due to security incident
- Confirmed unauthorized access to platform

### Other SEV0 Criteria
- Any SEV1 affecting a Mission Critical Customer
- Executive Leadership discretion

**Notes:**
- Only Engineering or Customer Support should file SEV0
- Cannot be filed through FE Portal (downgraded to SEV1)
- For SEV0, use go/fe-break-glass for emergency access

## SEV1 - High

**Response:** Highest priority during business hours (9am-5pm team timezone), no after-hours support.

**Criteria:**

### Service Outage
- Total outage for single customer without Production Support tier
- Partial service outage preventing some workloads (degraded operation)
- Performance regression with major production impact
- SLA misses requiring manual intervention
- Egregiously visible UI issues affecting all customers

### Financial/Billing Impact
- Impact <$10k on customer costs
- Staging billing issues
- Logs data loss out of SLA

**Examples:**
- Job SLA misses due to recent release changes
- Specific category of jobs failing
- Serverless endpoints not starting
- High risk to go-live timelines

## SEV2 - Standard Non-Critical

**Response:** Standard engineering queue priority.

**Criteria:**
- Customer impacting but doesn't meet SEV0/SEV1 criteria
- Minor functionality impacted
- Development environment issues
- UI issues on low-traffic surfaces

**Notes:**
- Advanced Support tickets default to SEV2 and cannot be changed
- Most common severity level for FE-filed tickets

## SEV3 - Low

**Response:** No committed SLA from Engineering.

**Criteria:**
- Trivial or low urgency requests
- Minimal or no customer impact
- Nice-to-have improvements
- Non-blocking issues

## Severity Selection Guidelines

| Impact | Urgency | Recommended Severity |
|--------|---------|---------------------|
| Total outage, Production/Mission Critical customer | Immediate | SEV0 |
| Total outage, Standard customer | High | SEV1 |
| Partial outage, degraded service | High | SEV1 |
| Blocking issue, has workaround | Medium | SEV2 |
| Non-blocking, minor impact | Low | SEV3 |
| Feature question, guidance needed | N/A | SEV2 (Advanced Support) |

## Changing Severity

- Severity should reflect **actual impact**, not urgency
- Engineering triages and may adjust severity based on assessment
- Don't change severity after triage unless impact assessment was incorrect
- Downgrade if impact was less than initially feared
- Never downgrade just because issue was mitigated quickly

## Salesforce to ES Severity Mapping

| SF Case Priority | ES Severity |
|-----------------|-------------|
| Critical | SEV0 |
| Urgent | SEV1 |
| High | SEV1 |
| Normal | SEV2 |
| Low | SEV3 |
| No SFDC Priority | SEV2 |

## Key Definitions

**Workload:** Any customer operation critical to their use - notebooks, jobs, pipelines, SQL queries, model serving, Lakebase, etc.

**Production Support Tier:** Customer has paid support contract with SLA guarantees.

**Mission Critical:** Highest support tier with most stringent SLAs.
