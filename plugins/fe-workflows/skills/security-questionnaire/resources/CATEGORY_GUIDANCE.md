# Security Questionnaire Category Guidance

Detailed guidance for answering questions in each of the 9 security questionnaire categories.

## Category 1: Policy & People

### Common Topics
- Information Security program and framework (ISO 27001, NIST)
- Employee background checks
- Security policies and Acceptable Use Policy (AUP)
- Third-party/subcontractor security
- Security awareness training
- NDAs and confidentiality agreements
- Data Privacy and Records Retention policies

### Key Databricks Controls
- ISO 27001/27017/27018/27701 certified Information Security Management System (ISMS)
- Annual security awareness training for all employees
- Mandatory background checks for all employees
- Confidentiality agreements signed by all employees and contractors
- Published Acceptable Use Policy (databricks.com/legal/aup)
- Third-party security assessment program

### Glean Search Queries
```
"Databricks security policy information security program ISO 27001"
"Databricks employee background check screening"
"Databricks security training awareness program"
"Databricks NDA confidentiality agreement contractor"
"Databricks acceptable use policy AUP"
```

### Typical Score: 4-5

---

## Category 2: Risk Management

### Common Topics
- Risk assessment methodology
- Annual risk assessments
- Risk treatment and residual risk approval
- SOC 1/2 Type II reports
- Additional certifications (PCI, ISO, HIPAA)
- Vulnerability management program

### Key Databricks Controls
- Formal risk assessment program aligned with ISO 27001
- Annual enterprise risk assessments
- Continuous vulnerability scanning and penetration testing
- SOC 1 Type II and SOC 2 Type II reports (annual)
- ISO 27001, ISO 27017, ISO 27018, ISO 27701 certifications
- PCI-DSS attestation of compliance
- HIPAA Business Associate Agreement support

### Glean Search Queries
```
"Databricks risk assessment vulnerability management program"
"Databricks SOC 1 SOC 2 Type II report"
"Databricks ISO 27001 certification compliance"
"Databricks PCI DSS attestation compliance"
"Databricks penetration test security assessment"
```

### Typical Score: 4-5

---

## Category 3: Incident Management

### Common Topics
- Security monitoring and alerting
- Incident response procedures
- Security Incident Response Team (SIRT)
- Forensic investigation capabilities
- Breach notification timelines
- Historical incident disclosure

### Key Databricks Controls
- 24/7 Security Operations Center (SOC) monitoring
- Dedicated Security Incident Response Team (SIRT)
- Documented incident response procedures
- Breach notification within 72 hours (per DPA)
- Post-incident review and lessons learned process
- Forensic investigation capabilities

### Glean Search Queries
```
"Databricks incident response SIRT security operations"
"Databricks breach notification timeline DPA"
"Databricks security monitoring SOC alerting"
"Databricks forensic investigation incident"
```

### Typical Score: 4

---

## Category 4: Physical Security

### Common Topics
- Visitor management
- Employee identification
- Access logging and retention
- Physical barriers and access controls
- Data center security
- Media destruction procedures

### Key Databricks Controls
**Note:** Physical data center security is **inherited from CSPs** (AWS, Azure, GCP).

- Databricks corporate offices have badge access controls
- Visitor sign-in and escort policies at corporate facilities
- **CSP-Inherited:** Data center physical access controls, video surveillance, biometric access
- **CSP-Inherited:** Media destruction and sanitization

### Glean Search Queries
```
"Databricks physical security data center CSP inherited"
"Databricks visitor management badge access"
"Databricks media destruction sanitization"
"Databricks shared responsibility model physical"
```

### Typical Score: 4 (with CSP inheritance noted)

---

## Category 5: Infrastructure

### Common Topics
- Network security policies
- Change management
- Network segregation
- Encryption (in-transit and at-rest)
- Firewalls and network access controls
- Vulnerability scanning and penetration testing
- Multi-factor authentication (MFA)
- Audit logging
- Anti-malware controls
- Patch management

### Key Databricks Controls
- TLS 1.2+ encryption for all data in transit
- AES-256 encryption for data at rest
- Customer-managed keys (CMK) support
- Network isolation between customer workspaces
- Regular vulnerability scanning (weekly+)
- Annual third-party penetration testing
- Mandatory MFA for all access
- Comprehensive audit logging
- EDR/anti-malware on all endpoints
- Automated patch management with SLAs

### Glean Search Queries
```
"Databricks encryption TLS AES customer managed keys CMK"
"Databricks network security firewall isolation"
"Databricks vulnerability scanning penetration testing"
"Databricks MFA multi-factor authentication"
"Databricks audit logging monitoring"
"Databricks patch management vulnerability remediation"
```

### Typical Score: 4-5

---

## Category 6: Application Security

### Common Topics
- Secure network architecture
- Configuration standards
- Environment segregation (prod/dev/test)
- Test data policies
- Secure SDLC practices
- Code review requirements
- Static/dynamic application security testing

### Key Databricks Controls
- Secure Software Development Lifecycle (SSDLC)
- Mandatory code review for all changes
- Static Application Security Testing (SAST) in CI/CD
- Dynamic Application Security Testing (DAST)
- Separate development, staging, and production environments
- No customer data in non-production environments
- Security design reviews for new features
- Dependency vulnerability scanning

### Glean Search Queries
```
"Databricks SDLC secure development lifecycle code review"
"Databricks SAST DAST application security testing"
"Databricks environment segregation production development"
"Databricks test data policy non-production"
```

### Typical Score: 4

---

## Category 7: Access Control

### Common Topics
- User provisioning/deprovisioning (SCIM)
- Access reviews
- Equipment return procedures
- Password policies (length, complexity, expiration, history)
- Account lockout
- Credential encryption
- Workstation locking
- Least privilege enforcement
- Privileged access management

### Key Databricks Controls
- SCIM integration for automated user provisioning/deprovisioning
- SSO integration with customer identity providers
- Role-based access control (RBAC)
- Quarterly access reviews
- Password policy: 12+ characters, complexity requirements
- Account lockout after failed attempts
- Privileged Access Management (PAM) for admin access
- Least privilege principle enforced
- Equipment return and access revocation upon termination

### Glean Search Queries
```
"Databricks SCIM user provisioning deprovisioning SSO"
"Databricks access review RBAC role based"
"Databricks password policy complexity expiration"
"Databricks privileged access management PAM"
"Databricks least privilege access control"
```

### Typical Score: 4

---

## Category 8: Business Continuity Management

### Common Topics
- BC/DR plans and testing
- Geographic redundancy
- RTO/RPO definitions
- Backup procedures
- Restoration testing
- Succession planning

### Key Databricks Controls
- Documented Business Continuity and Disaster Recovery plans
- Annual BC/DR testing
- Multi-region availability (customer choice)
- **CSP-Inherited:** Geographic redundancy of underlying infrastructure
- Regular data backups with tested restoration
- RTO/RPO defined per service tier
- Succession planning for key roles

### Glean Search Queries
```
"Databricks disaster recovery business continuity DR BC"
"Databricks RTO RPO recovery time objective"
"Databricks backup restoration testing"
"Databricks multi-region availability redundancy"
```

### Typical Score: 4

---

## Category 9: Data Security

### Common Topics
- Cloud hosting details
- Storage isolation
- Data exfiltration controls
- PCI-DSS compliance
- PII handling and access controls
- Data classification
- Subprocessor list
- Data retention and deletion

### Key Databricks Controls
- Customer data isolated per workspace
- Unity Catalog for data governance and access control
- Data Loss Prevention (DLP) capabilities
- PCI-DSS compliant environment available
- HIPAA-eligible environment available
- Data classification support via Unity Catalog
- Published subprocessor list (databricks.com/legal/databricks-subprocessors)
- Customer-controlled data retention and deletion
- GDPR/CCPA compliance support

### Glean Search Queries
```
"Databricks data security isolation customer data"
"Databricks PCI DSS compliance data classification"
"Databricks PII handling GDPR CCPA privacy"
"Databricks subprocessor list data processing"
"Databricks data retention deletion policy"
"Databricks Unity Catalog data governance"
```

### Typical Score: 4

---

## Cross-Category Notes

### Databricks vs. CSP Responsibility

Always clarify control ownership:

| Control Area | Databricks | CSP |
|--------------|------------|-----|
| Application security | ✓ | |
| Access control | ✓ | |
| Encryption implementation | ✓ | |
| Audit logging | ✓ | |
| Physical data centers | | ✓ |
| Network infrastructure | Shared | Shared |
| Media destruction | | ✓ |

### When to Escalate by Category

| Category | Escalate When |
|----------|---------------|
| Policy & People | Custom policy requests |
| Risk Management | Requests for raw audit findings |
| Incident Management | Questions about specific past incidents |
| Physical Security | On-site audit requests |
| Infrastructure | Custom network architecture requests |
| Application Security | Source code access requests |
| Access Control | Custom authentication requirements |
| Business Continuity | RTO/RPO guarantees outside standard |
| Data Security | Custom data residency requirements |
