# TAP Map - Taxonomy and Methodology Reference

## Standard Workload Taxonomy

Always map TAP data across these 8 workload areas:

| Workload Area | SKU Mapping | Type |
|---|---|---|
| **Ingest & Transform** | Jobs Compute, Serverless Jobs, DLT/Lakeflow Pipelines, CDC, streaming | Direct spend |
| **Data Warehouse** | SQL Warehouse, Serverless SQL, Photon, DBSQL, Lakeview dashboards | Direct spend |
| **ML Stack** | ML Runtime, model training, MLflow, Feature Store, notebooks (ML-focused) | Direct spend |
| **Gen AI Stack** | Model Serving, Foundation Model APIs, Vector Search, RT Inference, RAG | Direct spend |
| **Data Sharing** | Delta Sharing, Clean Rooms, marketplace | Direct spend |
| **Lakebase / OLTP** | Database Serverless Compute, operational database workloads | Direct spend |
| **Data Governance** | Unity Catalog penetration (% of compute tagged), data quality, lineage | Overlay (% penetration) |
| **Data Format** | Delta format penetration (% of compute tagged), table optimization | Overlay (% penetration) |

**Note:** Governance and Data Format are "overlay" metrics, not direct spend lines.

### SKU-to-Workload Mapping (Logfood)

| SKU Pattern | Workload Area |
|---|---|
| `premium_jobs_compute`, `premium_jobs_serverless_compute*` | Ingest & Transform |
| `premium_sql_compute`, `premium_serverless_sql_compute*`, `photon*` | Data Warehouse |
| `premium_all_purpose_compute`, `ml_runtime*` | ML Stack |
| `model_serving*`, `foundation_model*`, `vector_search*` | Gen AI Stack |
| `delta_sharing*`, `clean_rooms*` | Data Sharing |
| `database_serverless_compute*` | Lakebase |

---

## TAM Sizing Methodology

### 5-Step Initial TAM Estimation

For each workload area:

1. **Baseline (High Confidence):** Annualized consumption from Logfood, mapped by SKU
2. **Pipeline Uplift:** Active UCOs in U4-U5 (onboarding) and their expected contribution
3. **Recovery Potential:** Lost/stalled UCOs that failed for non-technical reasons (leadership changes, budget cycles, resource constraints - NOT technical failures). These are the best expansion signals.
4. **Adoption Gaps:** Features available but not yet adopted (e.g., Photon at 0%, UC at <100%, no Model Serving)
5. **Growth Trajectory:** Observed ramp rates extrapolated (e.g., if a feature grew 5x in the last year)

### Confidence Ratings

| Rating | Meaning | Evidence Required |
|---|---|---|
| **High** | Strong consumption data + active pipeline + clear technical path | Logfood data + U5/U6 UCOs |
| **Medium-High** | Good consumption base + identified expansion levers | Logfood data + adoption gap signals |
| **Medium** | Some consumption data + qualitative signals | Some Logfood + UCO pipeline |
| **Medium-Low** | Qualitative signals only | Executive interest, closed-lost opps |
| **Low-Medium** | Very early adoption, minimal data points | <$5K spend, no UCOs |

### DBRX % of TAM Calculation

```
DBRX % = (Current Annual Spend in Workload / Est. TAM Midpoint) * 100
```

**Interpretation:**
- **<10%** = Early adoption, massive whitespace
- **10-30%** = Growing, significant expansion opportunity
- **30-60%** = Mature, incremental growth
- **>60%** = Deep penetration, defend and expand to adjacent workloads

---

## Peer Benchmarking

### Peer Selection

Identify 4-8 companies based on the account's industry vertical:
- Direct industry competitors (same vertical)
- Similar-sized companies in adjacent verticals
- Aspirational peers (larger companies showing what's possible)

### Peer Queries

```sql
-- Find peer accounts
SELECT account_id, account_name, SUM(dbu_dollars) as total_spend
FROM main.gtm_data.c360_consumption_monthly
WHERE account_name LIKE '%PEER_NAME%'
AND usage_date >= DATE_ADD(CURRENT_DATE(), -12)
GROUP BY account_id, account_name
ORDER BY total_spend DESC

-- Peer SKU breakdown
SELECT sku, SUM(dbu_dollars) as total_spend
FROM main.gtm_data.c360_consumption_monthly
WHERE account_id = 'PEER_ACCOUNT_ID'
AND usage_date >= DATE_ADD(CURRENT_DATE(), -12)
GROUP BY sku ORDER BY total_spend DESC
```

### Peer Filtering Rules
- Exclude if annual spend < $50K (too small)
- Exclude if annual spend > 10x target account (too large)
- Target 4-8 peers; minimum 2 required
- If no industry peers found, broaden to adjacent verticals

### Peer-Adjusted TAM Logic

- **Over-indexed workloads (account > peer avg):** Narrow TAM range. Growth comes from specific levers, not broad mix expansion. These are strengths to DEFEND.
- **Under-indexed workloads (account < peer avg):** Widen upside. Peers demonstrate achievability. These are primary EXPANSION opportunities.
- **At-parity workloads:** Keep initial estimates. Growth is incremental.

---

## Champion Assignment Logic

**IMPORTANT: Be conservative. Only assign a single champion when you have strong evidence from UCO ownership.**

### CRITICAL: Databricks Employees Are NOT Champions

**NEVER assign a Databricks employee as an EB/Champion.** Champions are CUSTOMER-SIDE contacts only. Databricks employees (AEs, SAs, CSMs, etc.) appear in SFDC as UCO Owners, Opportunity Owners, and Account Team members — these are internal roles, not customer champions.

**How to identify Databricks employees:**
- Email ends in `@databricks.com`
- Listed as UCO `OwnerId` or Opportunity `OwnerId` (these are always Databricks employees)
- Title contains "Account Executive", "Solution Architect", "Customer Success", "Sales Engineer", or similar Databricks sales roles
- Appear in the `User` object (Databricks internal users), not the `Contact` object (customer contacts)

**Rule: Only people from the `Contact` object (customer contacts) can be champions.** If a name appears as a UCO Owner but is NOT in the Contact object for this account, they are a Databricks employee — skip them.

### Data Sources (in priority order)
1. **UCO Champion/Contact fields** — the customer-side contact linked to UCOs (STRONGEST signal). These are fields like `Champion__c`, `Contact__c`, or `Technical_Contact__c` on the UCO, NOT the `OwnerId` field.
2. **UCO workload tags** (which workload area each UCO belongs to)
3. **Account Contacts with roles and titles** — from the `Contact` object for this account (WEAKEST signal - use only when no UCO contact data)

**WARNING:** The UCO `OwnerId` field is the DATABRICKS employee who manages the UCO (usually the AE or SA). Do NOT use this field for champion assignment. Always use the customer contact fields instead.

### Decision Tree

**0. FILTER STEP (always do this first)**

Before assigning any champion, verify:
- The person is in the `Contact` object for this account (customer contact)
- The person does NOT have a `@databricks.com` email
- The person's title is NOT a Databricks sales role (AE, SA, CSM, etc.)

If any of these checks fail, SKIP that person entirely.

**1. HIGH CONFIDENCE - Assign Single Champion**

Assign a single name if ANY of these conditions are met:
- Customer contact is linked to 2+ active UCOs (U3-U6) in this workload area, OR
- Customer contact is linked to 1 active UCO + holds VP/C-level title matching the workload, OR
- Customer contact is explicitly listed as "Champion" or "Executive Sponsor" in UCO fields

Output: `"FirstName LastName, Title"`

**2. MEDIUM CONFIDENCE - List Multiple Options**

If 2-3 customer contacts have UCO linkage but none meet the high confidence bar:

Output: `"Name1, Title OR Name2, Title (Multiple UCO contacts - confirm primary)"`

**3. LOW CONFIDENCE - Title Match Only**

If NO UCO contact data exists for this workload, but account contacts have relevant titles:

Conservative title-to-workload mapping:
- "Data Engineering", "Head of Data Engineering" → Ingest & Transform ONLY
- "Analytics", "BI", "Director of Analytics" → Data Warehouse ONLY
- "Data Science", "Machine Learning", "ML" → ML Stack ONLY
- "AI", "GenAI", "LLM", "Applied AI" → Gen AI Stack ONLY
- "Data Sharing", "Partnerships", "Data Strategy" → Data Sharing ONLY
- "Database", "DBA", "Platform Engineer" → Lakebase ONLY
- "Governance", "Data Architect", "Compliance" → Data Governance ONLY
- "Platform", "Infrastructure", "Principal Engineer" → Data Format ONLY

**DO NOT** use generic titles: "VP of Data", "Chief Data Officer", "Director of Data" (too broad)

Output: `"FirstName LastName, Title (Title match only - verify with UCO data)"`

**4. NO CHAMPION - Mark as Gap**

If none of the above conditions are met:

Output: `"[CHAMPION GAP] - No UCO ownership or clear title match"`

### Special Cases

**Executive Sponsor:** If someone has CDO/VP-level title, list separately as "Executive Sponsor" in the stakeholder map. Do NOT assign as champion for individual workloads unless they are linked to UCOs as a customer contact.

**Platform/Infrastructure roles:** Often span multiple workloads. Only assign if they own UCOs in that specific workload.

### UCO-to-Workload Keyword Mapping

| Workload | Keywords in UCO Name/Description |
|---|---|
| Ingest & Transform | "ETL", "Pipeline", "Ingestion", "DLT", "Streaming", "CDC" |
| Data Warehouse | "SQL", "Analytics", "BI", "Reporting", "Dashboard", "DBSQL" |
| ML Stack | "ML", "Machine Learning", "Model Training", "MLflow", "Feature Store" |
| Gen AI Stack | "Gen AI", "LLM", "RAG", "Vector Search", "Model Serving", "Foundation Model" |
| Data Sharing | "Data Sharing", "Delta Sharing", "Clean Rooms", "Marketplace" |
| Lakebase | "Lakebase", "OLTP", "Operational Database", "Real-time" |
| Data Governance | "Unity Catalog", "UC", "Governance", "Compliance", "Lineage" |
| Data Format | "Delta", "Migration", "Table Format", "Optimization" |

---

## Competitive Tool-to-Workload Mapping

| Competitive Tool | Primary Workload Area(s) |
|---|---|
| **Snowflake** | Data Warehouse, Data Sharing |
| **Redshift** | Data Warehouse |
| **BigQuery** | Data Warehouse |
| **dbt** | Ingest & Transform, Data Warehouse |
| **Airflow / Dagster / Prefect** | Ingest & Transform |
| **Fivetran / Stitch** | Ingest & Transform |
| **SageMaker** | ML Stack |
| **Vertex AI** | ML Stack, Gen AI Stack |
| **Kubeflow** | ML Stack |
| **OpenAI API / Anthropic API** | Gen AI Stack |
| **Pinecone / Weaviate** | Gen AI Stack (Vector Search) |
| **MongoDB / PostgreSQL / MySQL** | Lakebase / OLTP |
| **Collibra / Alation** | Data Governance |
| **Great Expectations / Monte Carlo** | Data Governance |
| **Parquet (non-Delta)** | Data Format |
| **Microsoft Fabric** | Ingest & Transform, Data Warehouse |
| **Azure Data Factory** | Ingest & Transform |
| **Azure Data Explorer** | Data Warehouse |

### Incumbent Tool Identification Rules

**Sources (check all of these, in priority order):**
1. **Competitors__c** field from Opportunities (strongest signal)
2. **Lost UCOs** - tools the customer chose instead of Databricks
3. **UCO names and descriptions** - mentions of tools being migrated from or replaced (e.g., "Migrate from Airflow to DLT", "Replace Snowflake warehouse")
4. **UCO implementation notes** - references to existing tools in the customer's stack
5. **Contact titles** - role names that imply specific tooling (e.g., "Snowflake Admin" → Snowflake, "Azure Data Engineer" → Azure Data Factory/Synapse)
6. **ASQ/Specialist Request notes** - competitive context mentioned in specialist engagements
7. **Blocker descriptions** - blockers that reference competing tools or migration challenges

**Every tool listed must trace back to a specific SFDC record.** Do not infer tools from industry norms, company type, or general assumptions. If none of the above sources mention a tool for a workload, leave it blank.

**Priority order when multiple tools found:**
1. Tools mentioned in open/active opportunities
2. Tools from recent lost UCOs (last 12 months)
3. Tools referenced in UCO names/descriptions
4. Tools mentioned multiple times across any source

**Heuristics:**
- If a tool maps to multiple workloads, include it in all relevant ones
- Include version/details if available (e.g., "Snowflake Enterprise" vs "Snowflake Standard")
- A single mention in any SFDC record is sufficient to list a tool - don't require multiple mentions
