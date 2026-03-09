# Security Questionnaire Scoring Guide

Use this guide to apply consistent maturity scores (1-5) when completing security questionnaires.

## Maturity Scale Overview

| Score | Level | Description |
|-------|-------|-------------|
| 5 | Optimized | Continuous improvement, industry-leading, externally validated |
| 4 | Managed | Measured, controlled, consistently followed (Databricks typical) |
| 3 | Defined | Documented process, may not be consistently applied |
| 2 | Developing | Ad-hoc processes, partial coverage |
| 1 | Initial | No formal process |

---

## Score 5: Optimized

### Criteria
- Formal program with documented policies and procedures
- Regularly audited by independent third parties
- Continuous improvement based on audit findings and industry trends
- Industry-leading practices exceeding common requirements
- Automated controls where possible
- Metrics tracked and reported to leadership

### When Databricks Scores 5
- **Encryption**: TLS 1.2+ enforced, AES-256 at rest, CMK support, continuous monitoring
- **MFA**: Required for all access, hardware keys for privileged access
- **Penetration Testing**: Annual third-party testing, continuous vulnerability scanning
- **SDLC**: SAST/DAST integrated, mandatory code review, security design reviews

### Example Questions & Responses

**Q: Is data encrypted in transit using TLS 1.2 or higher?**
- Score: **5**
- Rationale: TLS 1.2 minimum enforced, TLS 1.3 supported, weak ciphers disabled, continuously monitored

**Q: Is multi-factor authentication required for all system access?**
- Score: **5**
- Rationale: MFA required for all access, hardware security keys for privileged access, integrated with SSO

---

## Score 4: Managed

### Criteria
- Documented policies and procedures
- Process is consistently followed across the organization
- Regular internal reviews and updates
- Controls are measured and monitored
- Exceptions are tracked and approved
- External audit coverage (SOC 2, ISO)

### When Databricks Scores 4
Most Databricks controls score at Level 4. This includes:
- **Background Checks**: Comprehensive pre-employment screening
- **Security Training**: Annual mandatory training with tracking
- **Access Reviews**: Quarterly access reviews with documented process
- **Incident Response**: Documented SIRT procedures, regular testing
- **Business Continuity**: Annual DR testing, documented RTO/RPO
- **Audit Logging**: Comprehensive logging with defined retention

### Example Questions & Responses

**Q: Does your organization conduct background checks on employees?**
- Score: **4**
- Rationale: Comprehensive background checks for all employees, documented in HR procedures, audited in SOC 2

**Q: Is there a documented incident response procedure?**
- Score: **4**
- Rationale: SIRT procedures documented, team trained, regular tabletop exercises, included in SOC 2 scope

**Q: Are access rights reviewed periodically?**
- Score: **4**
- Rationale: Quarterly access reviews, documented process, exceptions tracked, managers certify access

---

## Score 3: Defined

### Criteria
- Process exists and is documented
- May not be consistently applied across all areas
- Limited measurement or monitoring
- Occasional reviews but not systematic
- Some manual processes that could be automated

### When Databricks Might Score 3
Databricks rarely scores at Level 3, but might for:
- Emerging processes in early implementation
- Customer-specific configurations not yet mature
- New acquisitions during integration

### Example Questions & Responses

**Q: Is there a formal data classification policy?**
- Score: **3** (if customer hasn't implemented Unity Catalog governance)
- Rationale: Databricks provides tools (Unity Catalog) but implementation depends on customer configuration

---

## Score 2: Developing

### Criteria
- Ad-hoc processes
- Partial coverage across the organization
- Limited documentation
- Reactive rather than proactive
- Dependent on individual knowledge

### When to Use for Databricks
Score 2 is rarely appropriate for Databricks core controls. May apply to:
- Customer-specific requests outside standard offerings
- Features in early preview without formal process

---

## Score 1: Initial

### Criteria
- No formal process exists
- Completely ad-hoc approach
- No documentation
- No measurement or oversight

### When to Use for Databricks
Score 1 is almost never appropriate for Databricks. If a question addresses something Databricks doesn't do:
- Explain why it's not applicable (e.g., "Databricks does not process payment card data on behalf of customers")
- Or escalate if the question reveals a gap

---

## Scoring Decision Tree

```
Does Databricks have this control?
├─ No → Score N/A or explain why not applicable
└─ Yes → Is it documented?
         ├─ No → Score 2
         └─ Yes → Is it consistently followed?
                  ├─ No → Score 3
                  └─ Yes → Is it externally audited?
                           ├─ No → Score 4
                           └─ Yes → Is there continuous improvement?
                                    ├─ No → Score 4
                                    └─ Yes → Score 5
```

## Category-Specific Scoring Guidance

### Policy & People (Typical: 4)
| Topic | Score | Rationale |
|-------|-------|-----------|
| Security policies | 4 | Documented, reviewed annually, SOC 2 audited |
| Background checks | 4 | Comprehensive, consistent, audited |
| Security training | 4 | Annual, mandatory, tracked |
| NDAs | 4 | Required for all employees/contractors |

### Risk Management (Typical: 4-5)
| Topic | Score | Rationale |
|-------|-------|-----------|
| Risk assessments | 4 | Annual enterprise risk assessment |
| SOC 2 | 5 | Annual Type II, continuous monitoring |
| Penetration testing | 5 | Annual third-party, continuous scanning |
| Vulnerability management | 5 | Automated scanning, defined SLAs |

### Incident Management (Typical: 4)
| Topic | Score | Rationale |
|-------|-------|-----------|
| Incident response | 4 | Documented SIRT, regular testing |
| Breach notification | 4 | 72-hour notification per DPA |
| Forensics | 4 | Capability exists, procedures documented |

### Physical Security (Typical: 4)
| Topic | Score | Rationale |
|-------|-------|-----------|
| Data center access | 4 | CSP-inherited, validated via CSP SOC 2 |
| Media destruction | 4 | CSP-inherited, contractually required |
| Corporate offices | 4 | Badge access, visitor management |

### Infrastructure (Typical: 4-5)
| Topic | Score | Rationale |
|-------|-------|-----------|
| Encryption in transit | 5 | TLS 1.2+ enforced, monitored |
| Encryption at rest | 5 | AES-256, CMK supported |
| Network security | 4 | Segmented, monitored, audited |
| Patch management | 4 | Defined SLAs, tracked |
| MFA | 5 | Required, hardware keys for privileged |

### Application Security (Typical: 4)
| Topic | Score | Rationale |
|-------|-------|-----------|
| SDLC | 4 | Documented, security reviews integrated |
| Code review | 4 | Mandatory for all changes |
| SAST/DAST | 4 | Integrated in CI/CD |
| Environment segregation | 4 | Separate prod/dev/test |

### Access Control (Typical: 4)
| Topic | Score | Rationale |
|-------|-------|-----------|
| SCIM provisioning | 4 | Supported, documented |
| Access reviews | 4 | Quarterly, documented |
| Password policy | 4 | Meets industry standards |
| Privileged access | 4 | PAM implemented |

### Business Continuity (Typical: 4)
| Topic | Score | Rationale |
|-------|-------|-----------|
| DR plan | 4 | Documented, annually tested |
| Backups | 4 | Regular, restoration tested |
| Geographic redundancy | 4 | Multi-region available |

### Data Security (Typical: 4)
| Topic | Score | Rationale |
|-------|-------|-----------|
| Data isolation | 4 | Workspace isolation, Unity Catalog |
| PCI compliance | 4 | Attestation available |
| Subprocessors | 4 | Published list, DPA coverage |

## Red Flags: When Not to Score

Do not assign a score if:
1. The question is about a capability Databricks doesn't offer - explain N/A
2. The question requires customer-specific implementation details
3. The question asks about future capabilities - escalate
4. You're uncertain about the current state - research first

## Consistency Tips

1. **Use your knowledge base as baseline** - If your internal Q&A library has a score for a similar question, use it
2. **Document your rationale** - Include why you chose that score
3. **Be conservative** - When uncertain between 4 and 5, choose 4
4. **Cite evidence** - Scores should be backed by verifiable controls
