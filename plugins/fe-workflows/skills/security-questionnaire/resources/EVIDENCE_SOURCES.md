# Security Questionnaire Evidence Sources

Pre-built Glean queries and reference links for researching security questionnaire answers.

## Glean MCP Queries by Category

### Category 1: Policy & People

```bash
# Information Security Program
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks information security program ISO 27001 ISMS", "page_size": 15}}'

# Security Training
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks security awareness training employee program", "page_size": 15}}'

# Background Checks
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks employee background check pre-employment screening", "page_size": 15}}'

# Acceptable Use Policy
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks acceptable use policy AUP security policy", "page_size": 15}}'

# Third-Party Security
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks third party vendor security assessment subcontractor", "page_size": 15}}'
```

### Category 2: Risk Management

```bash
# Risk Assessment
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks risk assessment methodology enterprise risk management", "page_size": 15}}'

# SOC Reports
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks SOC 1 SOC 2 Type II audit report", "page_size": 15}}'

# Certifications
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks ISO 27001 27017 27018 27701 certification compliance", "page_size": 15}}'

# Vulnerability Management
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks vulnerability management scanning penetration testing", "page_size": 15}}'
```

### Category 3: Incident Management

```bash
# Incident Response
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks incident response SIRT security incident procedure", "page_size": 15}}'

# Breach Notification
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks breach notification timeline DPA data processing", "page_size": 15}}'

# Security Monitoring
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks security operations center SOC monitoring alerting", "page_size": 15}}'
```

### Category 4: Physical Security

```bash
# Data Center Security
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks physical security data center CSP cloud provider inherited", "page_size": 15}}'

# Shared Responsibility
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks shared responsibility model AWS Azure GCP physical", "page_size": 15}}'

# Media Destruction
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks media destruction sanitization disposal CSP", "page_size": 15}}'
```

### Category 5: Infrastructure

```bash
# Encryption
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks encryption TLS AES customer managed keys CMK BYOK", "page_size": 15}}'

# Network Security
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks network security firewall segmentation isolation VPC", "page_size": 15}}'

# MFA
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks MFA multi-factor authentication SSO SCIM", "page_size": 15}}'

# Patch Management
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks patch management vulnerability remediation SLA", "page_size": 15}}'

# Audit Logging
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks audit logging monitoring retention SIEM", "page_size": 15}}'
```

### Category 6: Application Security

```bash
# SDLC
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks SDLC secure development lifecycle code review", "page_size": 15}}'

# Security Testing
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks SAST DAST application security testing CI CD", "page_size": 15}}'

# Environment Segregation
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks environment segregation production development test", "page_size": 15}}'
```

### Category 7: Access Control

```bash
# User Provisioning
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks SCIM user provisioning deprovisioning identity", "page_size": 15}}'

# Access Reviews
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks access review certification RBAC role based", "page_size": 15}}'

# Password Policy
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks password policy complexity length expiration", "page_size": 15}}'

# Privileged Access
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks privileged access management PAM admin", "page_size": 15}}'
```

### Category 8: Business Continuity

```bash
# DR/BC Plans
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks disaster recovery business continuity DR BC plan", "page_size": 15}}'

# RTO/RPO
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks RTO RPO recovery time objective point", "page_size": 15}}'

# Backups
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks backup restoration testing data recovery", "page_size": 15}}'
```

### Category 9: Data Security

```bash
# Data Classification
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks data classification Unity Catalog governance", "page_size": 15}}'

# PCI/PII
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks PCI DSS PII HIPAA data privacy compliance", "page_size": 15}}'

# Data Retention
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks data retention deletion policy GDPR CCPA", "page_size": 15}}'

# Subprocessors
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks subprocessor list DPA data processing agreement", "page_size": 15}}'
```

---

## Public Documentation Links

### Trust Center

| Resource | URL |
|----------|-----|
| Trust Center Home | https://databricks.com/trust |
| Compliance | https://databricks.com/trust/compliance |
| Security Features | https://databricks.com/trust/security-features |
| Privacy | https://databricks.com/trust/privacy |

### Legal Documents

| Resource | URL |
|----------|-----|
| Security Addendum | https://databricks.com/legal/security-addendum |
| Privacy Notice | https://databricks.com/legal/privacynotice |
| Acceptable Use Policy | https://databricks.com/legal/aup |
| Subprocessors | https://databricks.com/legal/databricks-subprocessors |
| DPA | https://databricks.com/legal/data-processing-agreement |

### Product Documentation

| Resource | URL |
|----------|-----|
| Security Guide | https://docs.databricks.com/security/index.html |
| Authentication | https://docs.databricks.com/administration-guide/users-groups/single-sign-on/ |
| Unity Catalog | https://docs.databricks.com/data-governance/unity-catalog/ |
| Encryption | https://docs.databricks.com/security/keys/index.html |
| Audit Logs | https://docs.databricks.com/administration-guide/account-settings/audit-logs.html |

---

## Internal Resources

> Configure these for your organization. The locations below are examples — replace with your actual internal knowledge base, Confluence spaces, or Drive folders.

### Internal Knowledge Base / Q&A Library

| Resource | Description |
|----------|-------------|
| Security Q&A Database | Pre-answered common questions (search your internal knowledge base) |
| Security Team Intake | Submit complex questionnaires for security team review |
| Security FAQ | Security and compliance FAQ (internal wiki) |
| Enterprise Security Guide | Cloud-specific security guides (request from security team) |
| Shared Responsibility Model | Your organization's shared responsibility documentation |

### Document Repositories (Internal)

| Resource | Location |
|----------|----------|
| Security Policies | Your internal document management system (Confluence, Drive, etc.) |
| Penetration Testing | Request from your security team |
| Incident Response | Your internal runbook / wiki |
| Compliance Reports | Request from your compliance/security team |

---

## Documents Available to Customers

### Under NDA

These documents require a signed NDA before sharing:

| Document | How to Obtain |
|----------|---------------|
| SOC 2 Type II Report | Request via Trust Center or account team |
| SOC 1 Type II Report | Request via account team |
| Enterprise Security Guide | Request via account team |
| Penetration Test Executive Summary | Request via account team |

### Publicly Available

| Document | Location |
|----------|----------|
| ISO 27001 Certificate | Trust Center |
| ISO 27017 Certificate | Trust Center |
| ISO 27018 Certificate | Trust Center |
| ISO 27701 Certificate | Trust Center |
| PCI-DSS AOC | Trust Center (upon request) |
| Security Addendum | databricks.com/legal |
| DPA | databricks.com/legal |
| Subprocessors List | databricks.com/legal |

---

## Escalation Channels

Configure these for your organization. Examples:

| Channel / Contact | Purpose |
|-------------------|---------|
| #security (or equivalent) | General security questions, time-sensitive requests |
| #trust / #compliance (or equivalent) | Trust team questions, compliance attestations |
| #privacy (or equivalent) | Privacy-specific questions, GDPR/CCPA |
| Security team intake form | Submit questionnaires for security team review |
| #legal (or equivalent) | Contractual security terms |

---

## Quick Reference: Evidence by Question Type

| Question About | Primary Evidence | Secondary Evidence |
|----------------|------------------|-------------------|
| Certifications | ISO certificates, SOC reports | Trust Center |
| Encryption | Public docs, Security Addendum | Enterprise Security Guide |
| Access Control | Public docs, SOC 2 | Internal knowledge base responses |
| Incident Response | Security Addendum, DPA | SOC 2 report |
| Physical Security | CSP SOC 2 reports | Shared Responsibility Model |
| Compliance | Trust Center, specific attestations | Internal knowledge base, Confluence |
| Data Handling | DPA, Privacy Notice | Subprocessors list |
| Testing/Scanning | Pen test summary (NDA) | SOC 2 report |
