# ES Ticket Issue Types

Engineering Support (ES) tickets use specific issue types to categorize requests. Choosing the correct type is important for proper routing and prioritization.

## Incident

**When to Use:** Report problems related to outages, product defects, and performance issues.

**Examples:**
- Customer experiencing total service outage
- Production workloads failing
- Performance degradation after DBR upgrade
- Data integrity issues
- Billing/financial impact issues

**Key Fields:**
- Support Severity Level (SEV0-SEV3)
- Workspace ID
- Cloud (AWS/Azure/GCP)
- Component (from WhoOwnsIt)
- Outage Start Time
- Customer Impact Description

**Notes:**
- Within Support team, Frontline engineers get issues reviewed by Backline before filing
- Straightforward defects that can be reproduced consistently can be filed directly
- SEV0/SEV1 require immediate engineering attention

## Advanced Support

**When to Use:** Third-party integration questions, best-practice recommendations, code assistance, and issues outside the scope of reproducible bugs/defects.

**Examples:**
- Performance optimization assistance
- Application design guidance
- Custom integration support
- Complex troubleshooting that doesn't have a clear bug
- Best practices for specific use cases

**Key Fields:**
- Support Severity Level (defaults to SEV2, cannot be changed)
- Workspace ID
- Component
- Detailed problem description

**Notes:**
- Should be reviewed by a senior technical resource before opening
- SEV2 by default - cannot be escalated to SEV0/SEV1
- Engineering provides guidance, not necessarily code fixes

## Customization/Service Request

**When to Use:** Request to customize an aspect of Databricks for a user account, migrate shards, rotate keys, enable/disable features.

**Examples:**
- Enable specific feature flags for a workspace
- Quota increases (jobs, clusters, FMAPI concurrency, etc.)
- Key rotation requests
- Shard migration
- Enable/disable dbutils.secrets.get
- DBFS migration requests

**Key Fields:**
- Workspace ID(s) affected
- Feature/Configuration to change
- Current value and requested value
- Business justification

**Common Quota Requests:**
- Job quota increase
- Cluster quota increase
- FMAPI concurrency limits
- SQL Warehouse limits
- Storage limits

## Private Preview Bug

**JIRA Issue Type Name:** `Private Preview Bug` (singular)

**When to Use:** Bugs found specifically in private preview features.

**Examples:**
- Bug in a feature that's in private preview
- Issues with beta functionality
- Problems during early access programs
- Lakeflow Connect issues during preview
- Vector search preview bugs

**Real Ticket Examples:**
- ES-1685816: Lakeflow Connect (D365) pipeline failures
- ES-1685186: LakeFlow Connect (MySQL) failure
- ES-1682522: Vector search storage optimization issues

**Key Fields:**
- Private Preview name (select from list)
- Workspace ID
- Steps to reproduce
- Expected vs actual behavior
- Preview Status field set to "Private Preview"

**Notes:**
- If the Private Preview name isn't listed, reach out to Program Management
- Routes directly to the engineering team owning the preview
- Provides centralized tracking of preview issues across teams
- Description should clearly state "Private Preview" and the feature name

## Xteam-Ask

**When to Use:** R&D Eng-to-Eng cross-team asks during quarterly planning.

**Examples:**
- Engineering dependencies between products
- Feature launch dependencies on platform capabilities
- Requests for teams to migrate to shared infrastructure
- Cross-team collaboration requests

**Key Fields:**
- Requesting team
- Target team
- Dependency description
- Timeline/Quarter

**Notes:**
- See go/xteam-asks-jira for more examples
- See go/xteam-asks-sop for process details
- Primarily for engineering planning, not customer issues

## Choosing the Right Type

| Situation | Issue Type |
|-----------|------------|
| Customer outage or incident | Incident |
| Production bug/defect | Incident |
| Performance degradation | Incident |
| Need guidance on best practices | Advanced Support |
| Integration assistance | Advanced Support |
| Performance optimization help | Advanced Support |
| Enable a feature flag | Customization/Service Request |
| Increase quota | Customization/Service Request |
| Key rotation, shard migration | Customization/Service Request |
| Bug in preview feature | Private Preview Bug |
| Cross-team engineering ask | Xteam-Ask |

## Real Ticket Examples by Type

### Incident Examples
- ES-1679710: `[Doordash] Customer has reported that one of their hourly job with dynamic data size timesout occasionally`
- ES-1687429: `OpenAI: AWS Databricks DBT cluster hits huge Spark task backlogs, missing SLAs`
- ES-1665837: `Doordash has reported urgent issue that they are seeing a PyArrow serialization`

### Advanced Support Examples
- ES-1685524: `[openai] performance proof-of-concept for SPJ`
- ES-1658641: `DoorDash query with ST_DistanceSphere runs 50x slower than snowflake`
- ES-1675189: `DoorDash - Error When accessing foreign Iceberg tables through IRC`

### Customization/Service Request Examples
- ES-1688869: `Jobs quota increase request to 15,000 for PROD workspace`
- ES-1668038: `Block/Square - Raise FMAPI Workspace Quota Limits on Workspaces`
- ES-1068010: `increase the limit of pinned cluster in doordash-dev workspace to 300`

### Private Preview Bug Examples
- ES-1685816: `[Private Preview] Lakeflow Connect (D365) – Dev Pipeline Fails`
- ES-1685186: `[Private Preview] Nykaa - LakeFlow Connect (MySQL) failure`

## DO NOT Use ES Project For

- **Feature Requests** - Use the product feedback channels instead
- **Documentation issues** - Use the docs team channels
- **Sales/Commercial questions** - Use appropriate internal channels
