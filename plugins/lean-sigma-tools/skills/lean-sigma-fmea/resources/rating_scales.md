# FMEA Rating Scales — Lean Six Sigma Standard (Databricks Adapted)

All three scales run **1 (best) to 10 (worst)**. The descriptions below are adapted from AIAG FMEA standards with Databricks data platform context added.

RPN = Severity × Occurrence × Detection (Range: 1–1,000)

---

## Severity (S) — How bad is the impact?

Rates the **seriousness of the effect** on the customer, business, or compliance if the failure occurs. Score the worst plausible effect.

**Special rule: S ≥ 9 requires attention REGARDLESS of RPN.**

| Score | Level | Business Impact | Databricks Data Platform Context |
|-------|-------|-----------------|----------------------------------|
| **10** | Catastrophic | Safety hazard, regulatory violation, or business-ending event. No warning. | PII/PHI data breach exposed externally; regulatory fine triggered; complete data loss with no backup; financial fraud due to bad data |
| **9** | Critical | Regulatory non-compliance or major customer harm. Warning may exist. | Data corruption reaching production consumers; GDPR/CCPA violation; audit failure; financial reporting error affecting filings |
| **8** | Very High | Loss of primary function; major customer complaint; SLA severely missed | Gold layer unavailable >4 hrs; ML model serving down; BI dashboards blank for business-critical decision |
| **7** | High | Reduced primary function; customer notices significant degradation | Stale data >2× SLA threshold; query performance degraded >50%; key pipeline delayed >2 hrs |
| **6** | Moderate | Loss of secondary function; customer notices, complaints likely | Non-critical dashboard stale; ad-hoc query failures; some data missing from report |
| **5** | Low-Moderate | Customer notices minor degradation; workaround available | Slow queries (2–5×); minor data delays; some columns null unexpectedly |
| **4** | Low | Minor defect; most customers don't notice | Cosmetic report formatting errors; marginal performance degradation; log noise |
| **3** | Very Low | Trivial defect; no customer impact | Job warning logs; minor metadata inconsistencies; non-functional feature unavailable |
| **2** | Minor | Barely perceptible; no operational impact | Cosmetic UI issue; deprecated API call still working; documentation mismatch |
| **1** | None | No discernible effect | |

---

## Occurrence (O) — How likely is it to happen?

Rates the **likelihood the failure mode will occur** in the current process environment. Based on historical data where available; otherwise use expert judgment.

| Score | Level | Frequency | Databricks Data Platform Context |
|-------|-------|-----------|----------------------------------|
| **10** | Almost Certain | Failure occurs > once per day | Schema changes from source system happen daily; no schema enforcement; raw API with no contract |
| **9** | Very High | Failure occurs > once per week | Late-arriving data from known unreliable source; cluster startup failures in shared workspace |
| **8** | High | Failure occurs ~once per week | Occasional OOM on large joins; source system sends partial loads on Mondays |
| **7** | Moderately High | Failure occurs ~2–3 per month | CDC offset resets after source DB maintenance; cluster spot interruptions during peak |
| **6** | Moderate | Failure occurs ~once per month | Schema drift from vendor upgrades; occasional API rate limiting from source |
| **5** | Low-Moderate | Failure occurs a few times per year | Seasonal volume spikes causing timeouts; quarterly source system upgrades |
| **4** | Low | Failure occurs once or twice per year | Major source system migration; annual credential rotation issues |
| **3** | Very Low | Failure occurs rarely (~once in 1–3 years) | Cloud provider outage; infrastructure failure with DR in place |
| **2** | Remote | Failure is unlikely but conceivable | Security breach with MFA + IP allowlisting in place |
| **1** | Extremely Remote | Failure almost impossible | Full data loss with multi-region backup and tested DR |

---

## Detection (D) — Would we catch it in time?

Rates the **likelihood that existing controls will NOT detect the failure** before it reaches the customer. High Detection score = hard to detect = high risk.

**Key insight:** 1 = almost certain to detect; 10 = almost impossible to detect.

| Score | Level | Detection Capability | Databricks Data Platform Context |
|-------|-------|---------------------|----------------------------------|
| **10** | Absolutely Uncertain | No controls exist; no way to detect | No monitoring, no alerts, no data quality checks; failure only discovered by angry customer |
| **9** | Very Remote | Controls are unreliable; very unlikely to detect | Manual spot-checks only; no automated validation; informal "someone usually notices" |
| **8** | Remote | Controls are unlikely to detect | Basic job success/failure alerts only; no data quality validation; schema changes go unnoticed |
| **7** | Very Low | Controls may detect later in the process | Row count check only (misses quality issues); alerting with high latency (>1hr to page) |
| **6** | Low | Controls detect sometimes | DLT expectations on some columns; DBSQL query-level alerting; manual review of dashboards |
| **5** | Moderate | Controls will likely detect with human review | Automated anomaly detection (Databricks Lakehouse Monitoring); data observability tooling |
| **4** | Moderately High | Controls will detect before reaching most customers | DLT quarantine with expectation enforcement; alerting <15min SLA; automated rollback |
| **3** | High | Controls reliably detect before customers are impacted | Great Expectations full suite + alerting; Unity Catalog lineage audit; automated CI/CD validation |
| **2** | Very High | Defect is obvious; almost certain to detect | Schema enforcement at source + destination; DLT fail-on-violation; automated smoke tests |
| **1** | Almost Certain | Cannot reach customers; error-proof mechanism | Poka-yoke controls; write-protected Gold layer; schema registry enforcement; zero-trust access |

---

## RPN Interpretation and Action Thresholds

| RPN Range | Risk Level | Action Required |
|-----------|-----------|-----------------|
| **≥ 200** | Critical 🔴 | **Immediate action required.** Stop and fix before next production run. Escalate to leadership. |
| **100–199** | High 🟠 | **Action required within current sprint.** Define owner, date, and remediation plan. |
| **50–99** | Medium 🟡 | **Action required within quarter.** Add to backlog with priority. Monitor closely. |
| **< 50** | Low 🟢 | **Accept risk or improve opportunistically.** Document and review annually. |

**Regardless of RPN:** Any item with **Severity ≥ 9** must have a mitigation plan, even if Occurrence and Detection scores are low.

---

## Scoring Tips

### Do's
- Base scores on current controls (as-is state), not planned improvements
- Score Severity based on the worst realistic effect, not the worst imaginable
- When uncertain, err toward higher Occurrence and lower Detection (more conservative)
- Use the "Revised" columns after implementing actions — never retroactively change original scores

### Don'ts
- Don't score Occurrence = 1 just because something hasn't happened yet
- Don't score Detection = 1 unless there is a genuine error-proofing mechanism
- Don't average multiple opinions — discuss to reach consensus or use the more conservative score
- Don't let a low Occurrence artificially suppress action on a high-Severity item

---

## Quick Reference for Common Databricks Patterns

| Failure Mode | Typical S | Typical O | Typical D | Notes |
|---|---|---|---|---|
| Schema drift from source | 7 | 7 | 7 | O lower if source has schema contract; D improves with DLT schema enforcement |
| Cluster OOM on large join | 6 | 5 | 5 | Improve D with Spark UI alerting; O with query optimization |
| PII data unmasked in Gold | 9 | 4 | 7 | S is always high for PII; D improves with column-level security in UC |
| Job timeout (SLA miss) | 7 | 5 | 4 | D improves with job SLA alerts in system tables |
| Credential/secret expiry | 8 | 4 | 6 | O reduces with automatic rotation; D improves with dbutils.secrets monitoring |
| Runaway cluster (cost) | 5 | 5 | 6 | D improves with budget alerts and cluster policies |
| Unity Catalog permission misconfiguration | 8 | 5 | 7 | D improves with policy-as-code and automated access reviews |
| Late-arriving data | 6 | 6 | 5 | Improve O with source SLA; D with watermarking in streaming |
| Delta table ACID violation | 8 | 2 | 3 | Delta guarantees help; but cluster crashes mid-write can still occur |
| ML model serving latency spike | 7 | 5 | 4 | D improves with Databricks Model Monitoring |
