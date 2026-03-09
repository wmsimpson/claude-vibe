# Databricks Well-Architected Framework — SIPOC Alignment Reference

The Databricks Well-Architected Lakehouse Framework has **7 pillars**. Use this reference to tag SIPOC process steps and elements with the pillars they most directly impact, turning the SIPOC into a strategic architecture planning tool.

---

## The 7 Pillars

| Pillar | Short Code | Core Question |
|--------|-----------|---------------|
| Data & AI Governance | **GOV** | Is data access, lineage, quality, and policy enforced at every step? |
| Operational Excellence | **OPS** | Is the process observable, automated, and continuously improving? |
| Security | **SEC** | Is confidentiality, integrity, and availability guaranteed? |
| Reliability | **REL** | Does the process recover from failures and meet its SLA? |
| Performance Efficiency | **PERF** | Are compute and storage resources used optimally? |
| Cost Optimization | **COST** | Is spend justified, visible, and minimized where possible? |
| Interoperability & Usability | **INT** | Are outputs accessible to consumers using their preferred tools and formats? |

---

## Pillar-to-SIPOC Mapping

### Data & AI Governance (GOV)

**SIPOC elements most affected:** All — but especially Inputs, Process, Outputs

Governance checkpoints to look for in a SIPOC:
- Is every Input registered in Unity Catalog with schema, owner, and tags?
- Are PII/PHI inputs identified and masked/tokenized before processing?
- Is lineage captured at column level through every Process step?
- Are Outputs governed with row-level and column-level security?
- Do Customers have a documented data contract?
- Is there an audit log of who accessed what?

**Red flags in SIPOC that need GOV attention:**
- Process step "Copy data to S3" without Unity Catalog registration
- Input with PII classification and no masking step
- Output shared with external partner without Delta Sharing governance
- No lineage-tracking step in the process

---

### Operational Excellence (OPS)

**SIPOC elements most affected:** Process steps

OPS checkpoints:
- Is each Process step monitored with alerts for failure?
- Is there a CI/CD pipeline for Process step deployments?
- Are runbooks documented for common failure scenarios?
- Is there an on-call rotation that owns this pipeline?
- Are job SLAs tracked in Databricks system tables?

**Red flags:**
- Manual Process steps with no automation
- No observability/monitoring step in the process
- No defined owner for process steps
- Process depends on a human action (e.g., "send email to trigger next step")

---

### Security (SEC)

**SIPOC elements most affected:** Suppliers, Inputs, Customers

SEC checkpoints:
- Are Supplier integration methods using secure auth (OAuth, managed identity)?
- Are secrets stored in Databricks Secret Manager, not hardcoded?
- Is data encrypted in transit and at rest?
- Are network paths through private endpoints (PrivateLink, VNet injection)?
- Do Customers access via Unity Catalog (not direct cloud storage)?
- Is there IP allowlisting for external Customer access?

**Red flags:**
- Supplier uses API key stored in plaintext
- Input delivered over HTTP (not HTTPS)
- Customer accessing data directly via S3/ADLS (bypasses governance)
- No authentication mechanism listed for any Supplier

---

### Reliability (REL)

**SIPOC elements most affected:** Process steps, Inputs, Outputs

REL checkpoints:
- Are SLA requirements defined for Inputs (delivery latency, uptime)?
- Do Process steps have retry logic and idempotency guarantees?
- Is there a failure handling path (dead-letter queues, alerting)?
- Are Outputs validated before being surfaced to Customers?
- Is there a DR/failover plan for critical Process steps?
- Are data quality expectations (DLT/Great Expectations) enforced?

**Red flags:**
- No retry logic mentioned for any Process step
- Single supplier with no fallback
- Output freshness SLA not defined
- No data quality gate between Bronze → Silver → Gold

---

### Performance Efficiency (PERF)

**SIPOC elements most affected:** Inputs, Process steps, Outputs

PERF checkpoints:
- Are Input volumes documented to right-size compute?
- Are Process steps using optimized compute (photon, GPU, spot instances)?
- Is caching used where outputs are read frequently?
- Are Output tables optimized (Z-ordering, liquid clustering, file compaction)?
- Is streaming latency appropriate for the use case (real-time vs. micro-batch)?

**Red flags:**
- Input volume unknown — cannot right-size clusters
- All jobs run on all-purpose clusters (should use job clusters)
- Gold tables not OPTIMIZE'd/ZORDER'd
- Streaming job using batch compute type

---

### Cost Optimization (COST)

**SIPOC elements most affected:** Process steps, Outputs

COST checkpoints:
- Are cross-cloud/cross-region data transfers minimized?
- Are compute resources auto-scaling or on-demand?
- Is cold/warm/hot storage tiering applied to Outputs?
- Is there cost attribution (tagging) per Process step or team?
- Are idle clusters automatically terminated?
- Are Photon and serverless SQL being used where cost-effective?

**Red flags:**
- Supplier in a different cloud region than Databricks workspace (egress costs)
- All-purpose clusters running 24/7 for batch workloads
- No cost tagging on clusters or jobs
- Large Output tables with full scans (missing Z-ordering or clustering)

---

### Interoperability & Usability (INT)

**SIPOC elements most affected:** Outputs, Customers

INT checkpoints:
- Do Outputs use open formats (Delta, Iceberg) to prevent lock-in?
- Are Outputs discoverable in Unity Catalog with business descriptions?
- Can non-technical Customers use Genie for natural language queries?
- Is there a BI-layer abstraction (semantic model) for analyst Customers?
- Are APIs documented (OpenAPI/Swagger) for programmatic Customers?
- Can outputs be shared externally via Delta Sharing without data movement?

**Red flags:**
- Outputs in proprietary format not readable outside Databricks
- No catalog descriptions — data is technically accessible but not discoverable
- Executive Customer expected to write SQL queries directly
- External partner receiving data via FTP instead of Delta Sharing

---

## Tagging Guide for SIPOC Process Steps

When tagging each process step in the SIPOC, use 1–3 pillar codes:

| Process Step Example | Suggested WA Tags |
|---------------------|-------------------|
| Ingest raw events from Kafka | REL, GOV, PERF |
| Apply PII masking rules | SEC, GOV |
| Validate schema with DLT expectations | REL, GOV |
| Transform Bronze → Silver | PERF, OPS |
| Aggregate Silver → Gold | PERF, COST |
| Register outputs in Unity Catalog | GOV, INT |
| Serve predictions via Model Serving API | INT, REL, PERF |
| Monitor pipeline health via system tables | OPS, REL |
| Share Gold data via Delta Sharing | INT, SEC, GOV |

---

## Gap Analysis Quick Reference

Use this to quickly identify architectural gaps from the SIPOC:

| If this is missing... | WA Gap | Recommended next step |
|----------------------|--------|-----------------------|
| No Unity Catalog registration for any Input or Output | GOV | Add as a Process step; flag in FMEA |
| No observability/monitoring step | OPS | Add alerting step; create runbook |
| Supplier uses insecure auth | SEC | Security FMEA item; prioritize fix |
| No SLA defined for any Output | REL | Define freshness SLA; add monitoring |
| Input volume unknown | PERF | Add capacity planning to next steps |
| No cost attribution | COST | Tag clusters and jobs; review in system tables |
| Outputs not in open format | INT | Evaluate Delta/Iceberg migration path |
