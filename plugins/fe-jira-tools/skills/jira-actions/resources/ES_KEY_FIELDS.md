# ES Ticket Key Fields Reference

This document maps JIRA custom fields to their names and describes when each is required or important.

## Field ID to Name Mapping

| Field ID | Field Name | Description |
|----------|-----------|-------------|
| customfield_11500 | Support Severity Level | SEV0/SEV1/SEV2/SEV3 |
| customfield_11600 | Workspace ID | Primary workspace ID (single) |
| customfield_12301 | Outage Start Time | When issue began |
| customfield_12302 | Outage End Time | When issue resolved |
| customfield_14000 | Category | Usually "Other" |
| customfield_14100 | Salesforce Case | SF case number if linked |
| customfield_14207 | Cloud | AWS, Azure, or GCP |
| customfield_14262 | Customer/Account | Customer name |
| customfield_14401 | Eng-Team Owner | Auto-populated from component |
| customfield_14623 | Feature/Service Area | E.g., Photon, Spark, FMAPI |
| customfield_17025 | Workspace IDs | Multiple workspace IDs (comma-separated) |
| customfield_17026 | Environment Types | AWS Multi-tenant, Azure, etc. |
| customfield_18107 | Product Area | Engine, Lakeflow, AI Platform, etc. |
| customfield_18150 | ES Component | Routes to engineering team |
| customfield_19594 | Preview Status | Private Preview indicator |
| customfield_19660 | Severity (Display) | Mirror of Support Severity Level |
| customfield_20800 | Environment | Production, Dev, Staging |
| customfield_21301 | Customer Name | Customer identifier |
| customfield_21308 | Customer Timezone | E.g., America/Los_Angeles |
| customfield_27295 | Component (Alt) | Alternative component field |
| versions | Affects Version | DBR version, e.g., dbr-14.3.40 |

## CRITICAL: Field Format for API Calls

**When creating tickets via the JIRA API, select/dropdown fields MUST use the `{"id": "XXXXX"}` format, NOT `{"value": "..."}` format.**

Using `{"value": "..."}` will cause the API request to fail silently. Always use the option ID.

### Example - Correct vs Incorrect

```
# CORRECT - Use ID
"customfield_11500": {"id": "10602"}

# INCORRECT - Will fail
"customfield_11500": {"value": "SEV2 Standard-Non-Critical"}
```

## Field Options Reference

Many fields are select/dropdown fields with predefined options. Below are the available values discovered from real tickets.

### Support Severity Level (customfield_11500)

| Value | ID | Description |
|-------|-----|-------------|
| `SEV0 Critical` | `10600` | Production down, major financial impact |
| `SEV1 High` | `10601` | Production degraded, significant impact |
| `SEV2 Standard-Non-Critical` | `10602` | Important but not urgent (default for Advanced Support) |
| `SEV3 Low` | `10603` | Minor issues, questions |

### External Customer Facing (customfield_14677) - REQUIRED

| Value | ID | Description |
|-------|-----|-------------|
| `Yes` | `13344` | Customer-facing issue (default) |
| `No` | `13345` | Internal issue |

### Cloud (customfield_14207)

| Value | Description |
|-------|-------------|
| `AWS` | Amazon Web Services |
| `Azure` | Microsoft Azure |
| `GCP` | Google Cloud Platform |

### Product Area (customfield_18107)

| Value | Description |
|-------|-------------|
| `AI Platform` | Model Serving, FMAPI, MLflow, Vector Search |
| `Engine` | Spark, Photon, Query Processing |
| `Lakeflow` | Pipelines, DLT, Connectors |
| `Data Engineering` | Jobs, Clusters, Notebooks |
| `SQL Analytics` | DBSQL, Warehouses |
| `Unity Catalog` | UC, Sharing, Governance |
| `Platform` | Core platform services |

### Feature/Service Area (customfield_14623)

Common values observed:
- `Foundation Model APIs`
- `Lakeflow Pipelines`
- `Lakeflow Connect`
- `Photon`
- `Spark Execution`
- `Delta`
- `Jobs/Workflows`
- `Clusters`
- `Unity Catalog`
- `DBSQL`

### ES Component (customfield_18150) - REQUIRED

Components follow a hierarchical naming pattern. **You must use the component ID when creating tickets via API.**

#### Common Component IDs (from production tickets)

| Component | ID | Team/Area |
|-----------|-----|-----------|
| `MLflow.ModelRegistry` | `20802` | MLflow Model Registry |
| `MLflow.ExperimentRunTracking` | TBD | MLflow Experiment/Run Tracking |
| `QueryProcessing.SparkExecution` | TBD | Spark Execution |
| `Languages.SparkSQL` | TBD | SQL Language |
| `Storage.Delta` | TBD | Delta Lake |
| `FoundationModelServing.FMAPI` | TBD | Foundation Model APIs |
| `Workflows.Stability.Limits` | TBD | Jobs/Workflows Limits |

#### Finding Component IDs

Since there are hundreds of components, the best way to find the correct ID is:

1. **Search for a similar ticket** and extract the component ID:
   ```
   jira_read_api_call("issues.get", {"issue_id": "ES-XXXXXX"})
   # Look for customfield_18150.id in the response
   ```

2. **Search the create metadata** for component name:
   ```
   jira_read_api_call("issues.get_create_metadata", {
       "project": "ES",
       "issuetype": "Incident",
       "bash_query": "grep -i 'mlflow'"
   })
   ```

3. **Use go/whoownsit** to find the component name, then search for its ID.

#### Component Categories

Components follow this hierarchical pattern:
- `MLflow.*` - MLflow services (ModelRegistry, ExperimentRunTracking, etc.)
- `QueryProcessing.*` - Spark query processing
- `Storage.*` - Delta, storage services
- `FoundationModelServing.*` - FMAPI, model serving
- `Workflows.*` - Jobs, workflows, triggers
- `Connectors.*` - Lakeflow Connect connectors
- `Cluster.*` - Cluster management
- `AI.*` - AI services (VectorSearch, ModelServing)
- `Unity.*` - Unity Catalog services

### Environment (customfield_20800)

| Value | Description |
|-------|-------------|
| `Production` | Customer production workloads |
| `Development` | Customer dev/test environments |
| `Staging` | Staging environments |

### Environment Types (customfield_17026)

| Value | Description |
|-------|-------------|
| `AWS Multi-tenant` | Standard AWS deployment |
| `Azure Multi-tenant` | Standard Azure deployment |
| `GCP Multi-tenant` | Standard GCP deployment |
| `AWS Single-tenant` | Dedicated AWS deployment |
| `Azure Single-tenant` | Dedicated Azure deployment |

### Boolean Fields

Several fields use Yes/No values:

| Field | Values |
|-------|--------|
| Customer Critical (customfield_14303) | `Yes`, `No` |
| Data Loss (customfield_18801) | `Yes`, `No` |
| Regression (customfield_18802) | `Yes`, `No` |
| Workaround Available (customfield_14304) | `Yes`, `No` |

### Preview Status (customfield_19594)

| Value | Description |
|-------|-------------|
| `Private Preview` | Feature in private preview |
| `Public Preview` | Feature in public preview |
| `GA` | Generally available |

## Required Fields by Issue Type

### Incident

| Field | Required | Notes |
|-------|----------|-------|
| Summary | Yes | Format: `[CustomerName] Brief description` |
| Description | Yes | Problem details, impact, timeline |
| Support Severity Level | Yes | SEV0-SEV3 based on impact |
| Workspace ID | Yes | Primary affected workspace |
| Cloud | Yes | AWS, Azure, or GCP |
| ES Component | Yes | Determines routing team |
| Outage Start Time | Recommended | When issue began |
| Affects Version | Recommended | DBR version if applicable |
| Environment | Recommended | Production, Dev, etc. |

### Advanced Support

| Field | Required | Notes |
|-------|----------|-------|
| Summary | Yes | Format: `[CustomerName] Request description` |
| Description | Yes | Detailed question/request |
| Support Severity Level | Fixed | Always SEV2 (cannot change) |
| Workspace ID | Yes | Customer workspace |
| Cloud | Yes | AWS, Azure, or GCP |
| ES Component | Yes | Best match component |

### Customization/Service Request

| Field | Required | Notes |
|-------|----------|-------|
| Summary | Yes | Format: `[CustomerName] Request type` |
| Description | Yes | Current state, requested state, justification |
| Support Severity Level | Yes | Usually SEV2 or SEV3 |
| Workspace ID | Yes | Affected workspace(s) |
| Cloud | Yes | AWS, Azure, or GCP |
| ES Component | Yes | Service area handling request |
| Workspace IDs | If multiple | Comma-separated list |

### Private Preview Bug

| Field | Required | Notes |
|-------|----------|-------|
| Summary | Yes | Format: `[Private Preview] FeatureName - Bug description` |
| Description | Yes | Preview name, repro steps, expected/actual |
| Workspace ID | Yes | Where bug was found |
| Cloud | Yes | AWS, Azure, or GCP |
| ES Component | Yes | Preview feature area |
| Preview Status | Yes | Indicates Private Preview |
| Affects Version | If applicable | DBR or N/A |

## ES Component Examples

The ES Component field is critical for routing. Common patterns:

| Component | Routes To |
|-----------|----------|
| `QueryProcessing.SparkExecution` | Spark Execution team |
| `Languages.SparkSQL` | SQL team |
| `Storage.Delta` | Delta team |
| `FoundationModelServing.FMAPI` | FMAPI team |
| `Workflows.Stability.Limits` | Jobs/Workflows team |
| `Connectors.D365.F&O` | Lakeflow Connect team |
| `Cluster.Management` | Cluster team |

Use go/whoownsit or the WhoOwnsIt Databricks App to find the correct component.

## Product Area Examples

| Product Area | Description |
|--------------|-------------|
| Engine | Spark, SQL, Photon |
| Lakeflow | Pipelines, DLT, Connectors |
| AI Platform | Model Serving, FMAPI, MLflow |
| Data Engineering | Jobs, Clusters, Notebooks |
| SQL Analytics | DBSQL, Warehouses |
| Unity Catalog | UC, Sharing, Governance |

## Description Best Practices

**IMPORTANT: Always include full error details directly in the ticket.**

Do NOT rely solely on Slack links for error information - not everyone has access to Slack channels, and messages can be deleted or hard to find later. Copy the actual error messages, stack traces, and relevant context directly into the ticket description.

### For Incidents
```
**Customer:** [Customer name and team if applicable]
**Workspace ID:** [ID]
**Region:** [AWS/Azure/GCP region, e.g., AWS US Oregon]

---

**Problem Description:**
[What is happening - be specific]

**Error Message (FULL):**
```
[Paste the complete error message and stack trace here]
[Do not truncate - include the full error]
```

**Timeline:**
- Issue started: [datetime]
- Frequency: [constant, intermittent, etc.]
- Customer impact: [describe business impact]

**Customer Context:**
- [Any relevant customer quotes or observations]
- [Recent changes on their side, if any]
- [Scale of operation - e.g., "~5 metrics per run"]

**Workspace Information:**
- Workspace ID(s): [IDs]
- Workspace URL: [e.g., https://customer.cloud.databricks.com]
- DBR Version: [version if applicable]
- Cluster type: [type if applicable]

**Steps to Reproduce (if known):**
1. [step 1]
2. [step 2]

**Workaround (if any):**
- [workaround description]

**Similar Previous Tickets:**
- [ES-XXXXXX: Brief description of similar issue and resolution]

**Slack Discussion (for additional context):**
[Link - but remember to include key details above]
```

### For Advanced Support
```
**Request Summary:**
[Brief description of what guidance is needed]

**Context:**
- Workspace ID: [ID]
- Use case: [what customer is trying to achieve]

**Current Approach:**
[What they're currently doing]

**Questions:**
1. [specific question 1]
2. [specific question 2]

**Additional Context:**
[Any notebooks, error logs, etc.]
```

### For Customization/Service Request
```
**Request Type:**
[E.g., Quota Increase, Feature Flag Enable]

**Business Justification:**
[Why this is needed]

**Current State:**
- Current value/setting: [value]
- Current limit: [limit]

**Requested State:**
- Requested value/setting: [value]
- Requested limit: [limit]

**Workspace Information:**
- Workspace ID(s): [IDs]
- Account: [account name]

**Timeline:**
- When is this needed by: [date]
```

### For Private Preview Bug
```
**Private Preview Information:**
- Preview Name: [name]
- Preview Status: Private Preview

**Problem Statement:**
[Description of the bug]

**Steps to Reproduce:**
1. [step 1]
2. [step 2]
3. [step 3]

**Expected Behavior:**
[What should happen]

**Actual Behavior:**
[What actually happens]

**Workspace Information:**
- Workspace ID: [ID]
- DBR Version: [version if applicable]

**Screenshots/Logs:**
[Attach or link]
```

## Viewing All Fields on a Ticket

Use the JIRA MCP tools:

```
# View ticket with all details
jira_read_api_call("issues.get", {"issue_id": "ES-XXXXXX"})
```

## Finding Similar Tickets

When creating a new ticket, always search for similar ones first:

```
# Find similar quota requests
jira_read_api_call("issues.search", {
    "jql": "project = ES AND summary ~ 'quota increase' AND issuetype = 'Customization/Service Request'",
    "max_results": 10
})

# Find similar incidents for a component
jira_read_api_call("issues.search", {
    "jql": "project = ES AND issuetype = Incident AND cf[18150] = 'Storage.Delta'",
    "max_results": 10
})

# Find similar customer tickets
jira_read_api_call("issues.search", {
    "jql": "project = ES AND summary ~ 'CustomerName'",
    "max_results": 10
})
```

Then view a good example to understand the field structure:
```
jira_read_api_call("issues.get", {"issue_id": "ES-EXAMPLE"})
```
