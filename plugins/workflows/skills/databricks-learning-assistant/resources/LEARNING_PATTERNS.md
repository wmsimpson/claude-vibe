# Learning Patterns

Pedagogical patterns for teaching different categories of Databricks features. These are structural teaching methodologies — not feature-specific knowledge. The assistant researches feature details dynamically; these patterns guide *how* to teach once the feature category is identified.

---

## Pattern 1: Create-and-Query Features

**Applies to:** Tables, views, materialized views, volumes, dashboards, schemas, catalogs

**Steps:**
1. **Create** — Create the resource with a simple configuration
2. **Populate** — Insert sample data or connect a data source
3. **Query** — Read from the resource to verify it works
4. **Modify** — Alter the resource (add columns, change properties, update data)
5. **Verify** — Confirm modifications took effect via query

**What to verify at each stage:**
- Step 1: Resource exists (CLI list/describe command)
- Step 2: Data is present (row count, sample query)
- Step 3: Query returns expected results
- Step 4: Modification succeeded (describe shows new state)
- Step 5: Queries reflect the modified state

**Common pitfalls to highlight:**
- Forgetting to set required table properties at creation time (can't always add later)
- Not understanding the difference between managed and external resources
- Missing Unity Catalog requirements

**Transition to assessment:** After step 5, challenge the user to create a variation with different parameters (different columns, different properties, different data types).

**Typical step count:** 4-5 steps

---

## Pattern 2: Deploy-and-Configure Features

**Applies to:** Clusters, jobs, serving endpoints, model serving, warehouses, DLT pipelines

**Steps:**
1. **Deploy** — Create the resource with recommended defaults
2. **Configure** — Adjust settings for the specific use case (size, autoscaling, libraries)
3. **Test** — Run a workload on the resource
4. **Monitor** — Check metrics, logs, and status
5. **Iterate** — Modify configuration based on observations, re-test

**What to verify at each stage:**
- Step 1: Resource is in RUNNING/ACTIVE state
- Step 2: Configuration shows expected values
- Step 3: Workload completes successfully
- Step 4: Metrics are accessible and reasonable
- Step 5: Changed configuration produces different behavior

**Common pitfalls to highlight:**
- Choosing the wrong compute size for the workload
- Not enabling autoscaling when appropriate
- Forgetting to install required libraries
- Not understanding access mode implications (shared vs single-user)

**Transition to assessment:** After step 5, challenge the user to deploy the same resource type with a different configuration optimized for a different use case.

**Typical step count:** 4-5 steps

---

## Pattern 3: Governance Features

**Applies to:** Unity Catalog permissions, lineage, tagging, row/column filters, data masking, audit logs

**Steps:**
1. **Set up** — Enable the governance feature and configure prerequisites
2. **Apply policy** — Create a policy, grant/revoke, add tags, enable masking
3. **Verify enforcement** — Test that the policy is enforced (try accessing as different principal, query masked data)
4. **Test edge cases** — Try boundary conditions (inherited permissions, transitive grants, cross-catalog)
5. **Audit** — Check audit logs or lineage to confirm the governance action was recorded

**What to verify at each stage:**
- Step 1: Feature is enabled, prerequisites met
- Step 2: Policy exists and is configured correctly
- Step 3: Access is correctly allowed/denied based on policy
- Step 4: Edge cases behave as expected (or document where they don't)
- Step 5: Audit trail shows the governance actions

**Common pitfalls to highlight:**
- Confusing GRANT with ownership
- Not understanding inheritance (catalog → schema → table)
- Forgetting that some governance features require specific access modes
- Not testing with the right principal (testing as admin doesn't prove anything)

**Transition to assessment:** Challenge the user to set up a different governance policy and verify it independently.

**Typical step count:** 4-5 steps

---

## Pattern 4: Integration Features

**Applies to:** Lakeflow Connect, external connectors, Partner Connect, federated queries, foreign catalogs

**Steps:**
1. **Provision** — Set up the external system (RDS, Snowflake, Kafka, etc.) or identify an existing one
2. **Configure source** — Set up the connection from Databricks to the external system (credentials, endpoints, catalogs)
3. **Deploy pipeline** — Create the integration pipeline (Lakeflow, federated query, ingestion job)
4. **Validate data flow** — Verify data moves correctly between systems (row counts, data types, latency)
5. **Error handling** — Introduce a failure condition and observe how the system handles it (connection drop, schema change, permission error)

**What to verify at each stage:**
- Step 1: External system is accessible
- Step 2: Connection test succeeds
- Step 3: Pipeline is running without errors
- Step 4: Data in destination matches source
- Step 5: Errors are surfaced and recoverable

**Common pitfalls to highlight:**
- Network/VPC connectivity issues (especially across AWS accounts — see One-Env requirement)
- Credential rotation and secret management
- Schema evolution handling
- Not testing with realistic data volumes

**Transition to assessment:** Challenge the user to set up a different integration source or modify the pipeline configuration.

**Typical step count:** 5 steps

---

## Pattern 5: App Development Features

**Applies to:** Databricks Apps, Lakebase, custom UI development, Streamlit integration

**Steps:**
1. **Scaffold** — Create the app structure using the appropriate template or CLI
2. **Develop** — Write the core logic (backend, frontend, database schema)
3. **Deploy** — Push the app to the Databricks workspace
4. **Test** — Verify the app works end-to-end (UI, API, data access)
5. **Iterate** — Make a change, redeploy, verify the change took effect

**What to verify at each stage:**
- Step 1: Project structure exists with correct files
- Step 2: Code runs locally or passes syntax checks
- Step 3: App is deployed and accessible via URL
- Step 4: All features work as expected
- Step 5: Changes are reflected after redeployment

**Common pitfalls to highlight:**
- Forgetting to configure app permissions for data access
- Not understanding the deployment model (container, serverless, etc.)
- Missing required dependencies in the requirements file
- Port and routing configuration issues

**Transition to assessment:** Challenge the user to add a new feature or page to the app independently.

**Typical step count:** 4-5 steps

---

## Pattern 6: Query and Analytics Features

**Applies to:** Databricks SQL, Genie rooms, alerts, query federation, query optimization

**Steps:**
1. **Connect** — Authenticate and connect to a SQL warehouse or compute
2. **Explore data** — Browse available catalogs, schemas, and tables
3. **Write queries** — Start with simple queries, build to complex ones
4. **Optimize** — Use EXPLAIN, check query profiles, apply optimization techniques
5. **Share** — Create a dashboard, alert, or Genie room to share the results

**What to verify at each stage:**
- Step 1: Connection succeeds, warehouse is running
- Step 2: User can browse and describe tables
- Step 3: Queries return correct results
- Step 4: Query performance improves after optimization
- Step 5: Shared artifact is accessible to others

**Common pitfalls to highlight:**
- Choosing the wrong warehouse size for the query workload
- Not leveraging liquid clustering or Z-ordering for scan optimization
- Forgetting to use CTEs or temp views for readability
- Not understanding the difference between serverless and classic SQL

**Transition to assessment:** Challenge the user to write a query they haven't tried, optimize it, and explain why their optimization worked.

**Typical step count:** 4-5 steps

---

## Selecting the Right Pattern

When you've identified the feature through research, map it to a pattern:

| Feature Category | Pattern | Examples |
|---|---|---|
| Data objects | Create-and-Query | Delta tables, views, MVs, volumes |
| Compute resources | Deploy-and-Configure | Clusters, jobs, warehouses, endpoints |
| Security/compliance | Governance | UC permissions, lineage, masking, tags |
| External connections | Integration | Lakeflow Connect, Partner Connect, federated |
| UI/applications | App Development | Apps, Lakebase, Streamlit |
| SQL/BI | Query & Analytics | DBSQL, Genie, alerts, dashboards |

If a feature spans multiple categories (e.g., "Lakeflow Connect with Delta tables"), combine patterns: use Integration for the connection setup, then Create-and-Query for validating the resulting tables.

## Adapting for Experience Level

- **Start from scratch** — Follow all steps, explain every concept, provide context for every command
- **Fill gaps** — Start at step 2 or 3, skip creation basics, focus on configuration and edge cases
- **Go deep** — Jump to steps 4-5, focus on optimization, advanced patterns, and failure modes
