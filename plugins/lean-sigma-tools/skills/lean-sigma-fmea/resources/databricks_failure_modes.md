# Databricks Failure Mode Library

Pre-built failure modes organized by **Databricks Well-Architected Framework pillar**. Use as a starting point when brainstorming failure modes for a FMEA. Always review with the user — not all will apply.

Each entry includes:
- **Failure Mode**: How the step fails
- **Typical Effect**: Impact on downstream consumers
- **Common Causes**: Why it typically happens
- **Current Control Options**: What can mitigate detection or occurrence
- **Recommended Action**: Databricks-specific fix or prevention

---

## Data & AI Governance (GOV) Failures

| Failure Mode | Typical Effect | Common Causes | Control Options | Recommended Action |
|---|---|---|---|---|
| Data lineage broken between pipeline stages | Compliance audit fails; impact analysis impossible | Manual data movement outside Unity Catalog; raw SQL CTAS without lineage hooks | Unity Catalog lineage UI; manual lineage docs | Migrate all pipelines to Unity Catalog; use DLT for automatic lineage capture |
| PII data written to non-compliant table without masking | Regulatory violation (GDPR/CCPA/HIPAA); potential breach | Missing masking step; schema drift introduced PII column; access control misconfiguration | Column tags in Unity Catalog; data classification scans | Tag PII columns in UC schema; enforce dynamic views or column masks before consumption |
| Data product published without business description or owner | Data not discoverable; consumers use wrong data | No catalog documentation process; no ownership enforcement | Unity Catalog table properties; documentation automation | Mandate table owner + description tags as part of CI/CD gate |
| External data shared without Delta Sharing governance | Unauthorized data access; compliance violation | Direct S3/ADLS link shared instead of Delta Sharing | Unity Catalog shares; access logs | Replace direct storage links with Delta Sharing; audit existing external shares |
| Audit log gap for sensitive data access | Cannot prove compliance; forensics impossible | System tables not enabled; audit logging disabled | system.access.audit table; Databricks audit logs | Enable workspace audit logging; alert on sensitive table access via system tables |
| Data retention policy not enforced | Data kept beyond legal requirement; GDPR violation | No automated purge process; VACUUM not run | VACUUM with retention period; TTL policies; scheduled delete jobs | Implement retention automation using Delta VACUUM and scheduled purge jobs |

---

## Operational Excellence (OPS) Failures

| Failure Mode | Typical Effect | Common Causes | Control Options | Recommended Action |
|---|---|---|---|---|
| Pipeline failure not detected until customer complains | SLA miss; damaged trust | No alerting; job failure silently retried; alert sent to wrong channel | Databricks job failure alerts; system.lakeflow.job_run_timeline | Configure webhook + Slack/PagerDuty alerts; monitor system tables for job health |
| No runbook for common failure modes | Long MTTR during incidents; inconsistent response | Documentation not created or outdated | Playbooks in Confluence; on-call documentation | Create runbooks for top 5 failure modes identified in FMEA |
| CI/CD not enforced for notebook/pipeline changes | Untested code in production; regressions | Direct workspace editing; no PR review process | Databricks Asset Bundles (DAB); GitHub Actions | Implement DAB with CI/CD pipeline; require PR approval before deploying to prod |
| No data freshness monitoring | Stale data not noticed; consumer decisions based on old data | No SLA tracking; no monitoring alerts on table update time | Databricks Lakehouse Monitoring; system.operational_data tables | Alert when table's last modified timestamp exceeds SLA threshold |
| Cluster/job configuration not version-controlled | Config drift; impossible to reproduce past state | Manually created jobs; no IaC | Databricks Asset Bundles; Terraform | Export all job/cluster configs to git; manage via IaC |
| Ownership of pipeline undefined | No one responds to alerts; incidents escalate | Org changes; lack of explicit assignment | Table properties "owner" field; on-call rotation docs | Assign explicit owner to every production job and table in Unity Catalog |

---

## Security (SEC) Failures

| Failure Mode | Typical Effect | Common Causes | Control Options | Recommended Action |
|---|---|---|---|---|
| Hardcoded credentials in notebook | Credential exposure; unauthorized access | Developer convenience; no secret manager enforcement | Databricks Secrets (dbutils.secrets); secret detection in CI/CD | Scan all notebooks for hardcoded secrets; enforce dbutils.secrets usage |
| Overly permissive Unity Catalog grants | Unauthorized data access; data exfiltration risk | Convenience grants (ALL PRIVILEGES to everyone); group misconfiguration | Unity Catalog privilege audits; least-privilege review | Audit UC grants; implement role-based access with principle of least privilege |
| Network traffic not using private endpoints | Data in transit exposed; compliance risk | Default public endpoint configuration | PrivateLink (AWS); Private Endpoints (Azure/GCP); VNet injection | Enable private connectivity for all production workspaces |
| MFA not enforced for workspace access | Account takeover risk; unauthorized access | SSO misconfiguration; service account with password auth | Okta SSO + MFA; Databricks IP access lists | Enforce SSO with MFA; add IP allowlisting for sensitive workspaces |
| Service principal credentials not rotated | Stale credentials increase breach window | Manual rotation process; no expiry enforcement | Managed identities; Databricks OAuth; Azure Managed Identity | Use managed identities or short-lived OAuth tokens instead of static credentials |
| Row/column-level security not applied to sensitive tables | Unauthorized access to sensitive data subset | Unity Catalog dynamic views not implemented; row filters missing | Unity Catalog row filters; column masks | Apply row-level and column-level security via Unity Catalog for all sensitive tables |
| Compute not isolated for sensitive workloads | Noisy neighbor; potential data cross-contamination | Shared all-purpose cluster used for all workloads | Cluster policies; dedicated clusters for sensitive workloads | Isolate sensitive workloads to dedicated clusters with strict init scripts |

---

## Reliability (REL) Failures

| Failure Mode | Typical Effect | Common Causes | Control Options | Recommended Action |
|---|---|---|---|---|
| Data quality expectation failures unhandled | Bad data reaches Gold layer; downstream corruption | DLT expectations in WARN mode instead of FAIL; no quarantine | DLT expectations with quarantine; Great Expectations | Set DLT expectations to FAIL or quarantine; never WARN in production |
| Job fails with no retry and no idempotency | Data gap; manual intervention required | Default no-retry config; non-idempotent transforms | Databricks job retry policies; idempotent writes (MERGE/REPLACE WHERE) | Configure 3 retries with backoff; ensure all writes are idempotent |
| Source system delivers duplicate records | Duplicate data in downstream tables; wrong aggregates | No deduplication step; source CDC issues | Deduplication logic; MERGE statements; DLT expectations | Add deduplication step; use MERGE for CDC upserts; validate uniqueness |
| Late-arriving data causes incomplete aggregates | Wrong business metrics; incorrect reporting | Watermark too tight; batch window too small | Databricks structured streaming watermarks; late data handling in DLT | Implement appropriate watermark strategy; consider reprocessing window |
| Schema change in source breaks pipeline | Pipeline fails; data gap until manually fixed | No schema evolution handling; strict schema enforcement without alerts | Schema evolution in Auto Loader (cloudFiles.schemaEvolutionMode); DLT schema hints | Enable Auto Loader schema evolution; alert on schema changes; use DLT schema inference |
| Delta table VACUUM removes files needed by downstream | Query fails on historical data; time travel broken | VACUUM run too aggressively (< 7-day default); no coordination | Set VACUUM retention ≥ 7 days; check Delta history before vacuuming | Configure VACUUM retention aligned with downstream time-travel requirements |
| Cluster spot interruption causes job failure | Data loss or gap; reprocessing required | Spot/preemptible instances used without fallback | On-demand fallback in cluster config; job retry with checkpointing | Configure spot + on-demand fallback; use structured streaming checkpoints |
| Dependencies not tracked (implicit ordering) | Downstream jobs run on stale data | No workflow dependency graph; jobs run on schedule regardless of upstream status | Databricks Workflows task dependencies; Lakeflow orchestration | Model all dependencies explicitly in Databricks Workflows |

---

## Performance Efficiency (PERF) Failures

| Failure Mode | Typical Effect | Common Causes | Control Options | Recommended Action |
|---|---|---|---|---|
| Query timeout due to large unoptimized table scan | BI dashboard fails; users abandon tool | No Z-ordering or liquid clustering; no partition pruning | OPTIMIZE + ZORDER; Liquid Clustering; Delta stats | Run OPTIMIZE with ZORDER on high-cardinality filter columns; enable liquid clustering |
| Data skew causes shuffle bottleneck | Job runs 10× longer than expected; cluster underutilized | Non-uniform join key distribution; large partition files | Spark adaptive query execution (AQE); salting; skew hints | Enable AQE (spark.sql.adaptive.enabled=true); identify skewed keys; apply salting |
| OOM on driver or executor | Job fails; cluster restarts; data loss | Collect() on large dataset; no broadcast hint; wrong cluster size | Spark UI monitoring; driver OOM alerts; broadcast hints | Avoid collect() on large datasets; use broadcast for small dimension tables; right-size cluster |
| Cold cluster startup latency violates SLA | Job starts late; downstream SLA cascade miss | Job cluster creation time (3–5 min) included in SLA window | Cluster pre-warming; serverless compute; SQL warehouses | Use Databricks serverless for latency-sensitive workloads; pre-warm job clusters |
| Small file problem causing slow reads | Query performance degrades over time | Streaming micro-batches; no periodic OPTIMIZE | Delta auto-optimize; OPTIMIZE schedule; compaction jobs | Enable Auto Optimize (autoOptimize.optimizeWrite + autoCompact) on streaming tables |
| Photon not enabled for SQL-heavy workloads | 2–5× slower query performance; higher DBU cost | Default cluster config; Photon not selected | Photon-enabled cluster type selection | Enable Photon for all SQL warehouses and DBR compute running heavy SQL |
| ML model inference latency exceeds SLA | Poor user experience; timeout errors in apps | Under-provisioned serving endpoint; cold start | Model Serving autoscaling; GPU endpoints; warm pool | Configure autoscaling with minimum endpoints > 0; use GPU for deep learning models |

---

## Cost Optimization (COST) Failures

| Failure Mode | Typical Effect | Common Causes | Control Options | Recommended Action |
|---|---|---|---|---|
| Cluster running 24/7 with no workload | Wasted spend; budget overrun | All-purpose cluster without auto-terminate; interactive dev left running | Auto-terminate policies; cluster policies; budget alerts | Enforce auto-terminate (60 min default); use cluster policies to enforce limits |
| Cross-cloud/cross-region data egress | Unexpected large cost; monthly bill surprise | Processing data in region A, storing in region B; no cost visibility | Cloud provider cost tools; Databricks cluster tagging | Colocate compute and storage in same region; tag clusters for cost attribution |
| Inefficient query doing full table scan | High DBU cost per query; warehouse overloaded | No partition filter; wrong join order; missing statistics | EXPLAIN plan review; Delta stats; ANALYZE TABLE | Run ANALYZE TABLE to collect statistics; add partition filters; review query plans |
| Over-provisioned SQL warehouse | Paying for capacity never used | Default warehouse size; no autoscaling configured | Serverless SQL; warehouse autoscaling; size-to-workload | Use serverless SQL warehouses; right-size based on system table query metrics |
| Unrestricted external API calls causing cost spike | Budget overrun; rate limiting from provider | No throttling; no cost alerts on API usage | API cost monitoring; budget alerts; caching layer | Cache API responses in Delta table; add cost alerts; implement request throttling |
| Delta tables never VACUUM'd | Storage costs grow unbounded | Default 7-day retention; VACUUM never scheduled | Scheduled VACUUM jobs; table maintenance workflows | Schedule weekly VACUUM jobs; set appropriate retention based on requirements |

---

## Interoperability & Usability (INT) Failures

| Failure Mode | Typical Effect | Common Causes | Control Options | Recommended Action |
|---|---|---|---|---|
| Data not accessible in preferred BI tool | Analyst workaround; data silos; shadow IT | JDBC/ODBC not configured; no semantic layer; SQL permissions not granted | Databricks Partner Connect; BI tool connectors; ODBC setup | Configure Databricks SQL endpoint with BI tool connector; grant appropriate SQL permissions |
| Genie Room query returns wrong answer | Executive loses trust in AI; reverts to manual | Genie not configured with correct context; ambiguous table names | Genie room configuration; table descriptions; certified answers | Add business descriptions to all Gold tables; configure Genie with certified Q&A pairs |
| Data product not discoverable in catalog | Consumers don't find data; duplicate datasets created | No catalog description; no tags; not in correct catalog | Unity Catalog search; table tags; certified data products | Add searchable tags and business descriptions; certify quality data products in UC |
| External partner cannot access shared data | Business process blocked; partnership delayed | Delta Sharing not set up; credentials not exchanged | Delta Sharing provider/recipient setup; Marketplace listing | Set up Delta Sharing with proper access controls; document the sharing process |
| Output format change breaks downstream consumer | Pipeline failures; emergency fixes required | No data contract; no semantic versioning; ad-hoc schema changes | Schema Registry; data contracts; column deprecation process | Implement schema evolution with backward compatibility; notify consumers of changes |
| Non-technical users cannot query Gold layer | Analytics work bottlenecked on engineers | No Genie access; no BI layer; SQL skills required | Genie Rooms; Databricks Apps; BI dashboards | Set up Genie Room for natural language access; create certified dashboards for common queries |
