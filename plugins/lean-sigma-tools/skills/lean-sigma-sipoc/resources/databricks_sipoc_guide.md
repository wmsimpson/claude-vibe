# Databricks SIPOC Deep-Dive Guide

Use these structured question sets when completing the **SIPOC Deep Dive** sheet — Sheet 2 of the SIPOC Google Sheet. Each element has Databricks-specific fields that turn a standard SIPOC into an enterprise architecture assessment tool.

This guide is inspired by Databricks Well-Architected Framework principles and field-proven discovery patterns.

---

## Suppliers

For each supplier, capture:

| Field | Guidance |
|-------|----------|
| **Supplier Type** | Internal (another team/system), External (SaaS vendor, partner), or Regulatory. Impacts governance boundaries and SLA ownership. |
| **Integration Method** | How Databricks connects: Auto Loader, Lakeflow Connect, JDBC/ODBC, REST API, Delta Sharing, custom Spark source, Kafka/Kinesis, Fivetran/Airbyte. |
| **Data Ownership Model** | Who owns the data — and who is the data steward for quality issues and access escalations? |
| **SLA Requirements** | Delivery frequency (batch/streaming), acceptable latency, uptime requirements. Feed into Reliability design. |
| **Authentication Type** | OAuth 2.0, service account/managed identity, API key, mutual TLS, Databricks secrets. Feeds into Security design. |
| **Unity Catalog Registration** | Is this supplier's data registered as an external location or federated catalog in Unity Catalog? Y/N — drives governance posture. |
| **Data Classification** | PII, PHI, financial, public. Determines encryption requirements and access control rigor. |
| **Cloud & Region** | Source cloud (AWS/Azure/GCP) and region — critical for egress cost planning and data residency compliance. |

**WA Pillars most relevant to Suppliers:** Security, Data & AI Governance, Reliability, Cost Optimization

---

## Inputs

For each input, capture:

| Field | Guidance |
|-------|----------|
| **Volume** | Typical daily/monthly GB, TB, or PB. Drives cluster sizing and storage cost estimates. |
| **Format** | Parquet, JSON, CSV, Avro, Delta, Iceberg, ORC. Affects ingestion speed and schema inference complexity. |
| **Ingestion Mode** | Full load, incremental (CDC), micro-batch, real-time streaming, event-driven. Determines pipeline architecture pattern. |
| **Schema Stability** | Is the schema fixed, semi-structured, or schema-on-read? High schema drift = reliability risk. |
| **Data Quality Baseline** | Known issues: nulls, duplicates, late-arriving records, encoding problems. Feeds into FMEA failure modes. |
| **Update Frequency** | Once daily, hourly, near-real-time (<1 min), ad-hoc. |
| **Lineage Tracking** | Is input lineage captured in Unity Catalog? Does the supplier provide column-level lineage? |
| **Compliance Requirements** | GDPR, CCPA, HIPAA, SOC2 — specific to this input. |

**WA Pillars most relevant to Inputs:** Reliability, Data & AI Governance, Performance Efficiency

---

## Process Steps

For each process step, capture:

| Field | Guidance |
|-------|----------|
| **Databricks Component** | DLT (Delta Live Tables), Lakeflow Connect, Databricks Workflows/Jobs, MLflow, Model Serving, Genie, Notebooks, DBSQL. |
| **Transformation Logic** | SQL, PySpark, Scala Spark, Python, dbt, Spark Structured Streaming. |
| **Orchestration Tool** | Databricks Workflows, Apache Airflow, Azure Data Factory, dbt Cloud, custom scheduler. |
| **Data Quality Checks** | Expectations (DLT), Great Expectations, custom checks, schema enforcement, constraint validation. |
| **Lineage Tracking Method** | Unity Catalog lineage UI, manual documentation, OpenLineage/Marquez, column-level lineage. |
| **Run-time Constraints** | Must run in cloud X, requires GPU, proprietary runtime, latency SLA, memory floor. |
| **WA Pillar Tags** | Tag 1–3 of the 7 WA pillars this step is most critical for. See `wa_alignment.md`. |
| **Current State Maturity** | Is this step manual, scripted, automated, or fully observable? |
| **Waste Category** | For Lean analysis: Defects, Overproduction, Waiting, Non-utilized talent, Transportation, Inventory, Motion, Extra-processing (DOWNTIME). |

**WA Pillars most relevant to Process:** Operational Excellence, Reliability, Performance Efficiency, Data & AI Governance

---

## Outputs

For each output, capture:

| Field | Guidance |
|-------|----------|
| **Output Type** | Delta table (Bronze/Silver/Gold), dashboard, API endpoint, ML model, report, data share, alert/notification. |
| **Consumption Volume** | Expected QPS, concurrent users, query frequency. Drives SQL warehouse sizing. |
| **Update Frequency** | Real-time, near-real-time, hourly, daily, weekly. |
| **Destination Platform** | Databricks SQL, Power BI, Tableau, Looker, custom app, external system via Delta Sharing, S3/ADLS export. |
| **Access Control Method** | Unity Catalog grants (table/column/row-level), IP access lists, private endpoints. |
| **SLA for Freshness** | Maximum acceptable data age for consumers (e.g., "Gold tables must be <4 hrs stale"). |
| **Output Format** | Delta, Parquet, Iceberg, JSON API, Avro. |
| **Data Catalog Registration** | Is output registered and documented in Unity Catalog with description, owner, tags? |

**WA Pillars most relevant to Outputs:** Performance Efficiency, Data & AI Governance, Reliability, Interoperability & Usability

---

## Customers

For each customer/consumer, capture:

| Field | Guidance |
|-------|----------|
| **User Role** | Data Analyst, Data Scientist, ML Engineer, Data Engineer, Business Executive, External Partner. |
| **Access Channel** | Databricks Genie (natural language), DBSQL, BI tool (Power BI, Tableau), REST API, Python SDK, Databricks Apps. |
| **Primary Use Case** | Ad-hoc exploration, scheduled reporting, ML model training, real-time inference, regulatory submission. |
| **Technical Proficiency** | SQL-only, Python/Spark, no-code/Genie. Drives UX requirements and training needs. |
| **Feedback Mechanism** | Direct user feedback, usage telemetry (system tables), error tracking, SLA breach alerting. |
| **Training / Enablement** | Has this team completed Databricks Academy training? Is a learning plan needed? |
| **Data Contract Expectations** | Does this consumer require guaranteed schema stability, SLA commitments, or formal data contracts? |
| **Compliance Obligations** | Does this consumer's use of data create additional regulatory obligations (e.g., reporting to regulators)? |

**WA Pillars most relevant to Customers:** Interoperability & Usability, Data & AI Governance, Security

---

## SIPOC Validation Checklist

Before finalizing, verify:

- [ ] Process has 4–7 steps, each an action verb + subject
- [ ] All process steps have at least one output and one input
- [ ] All customers are named (not "users" or "the business")
- [ ] All suppliers are identified (no orphan inputs)
- [ ] WA pillar tags assigned to every process step
- [ ] Data classification noted for any Supplier or Input with PII/PHI/financial data
- [ ] At least one SLA or freshness requirement captured per Output
- [ ] Unity Catalog integration status noted for Inputs and Outputs

---

## Common SIPOC Anti-Patterns to Avoid

| Anti-Pattern | Why It's a Problem | Fix |
|---|---|---|
| > 7 process steps | Loses the high-level view; becomes a detailed flow map | Consolidate into capability groups |
| Implementation detail in Process | "Run Python notebook cell 4" is not a process step | "Apply business rules to Silver layer" |
| Unnamed customers | "Stakeholders" is not actionable | Name the role or team |
| Missing Governance pillar | Unity Catalog and lineage not addressed | Add as a process step or WA tag |
| Single supplier for critical input | Single point of failure (SPOF) | Flag in FMEA risk column |
| No output SLA | Can't hold the process accountable | Define freshness requirement |
