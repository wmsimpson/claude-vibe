# Security Questionnaire Response Template

Use this template for formatting responses to security questionnaire questions.

## Standard Response Format

```markdown
### Q[#]: [Full question text]

**Suggested Score:** [1-5]

**Response:**
[Detailed answer addressing the specific question. Include relevant details about Databricks controls, certifications, and processes. Be specific and avoid generic statements.]

**Evidence:**
- [Document/certification/URL 1]
- [Document/certification/URL 2]
- [Document/certification/URL 3]
```

## Example Responses

### Policy & People Example

```markdown
### Q5: Does your organization conduct background checks on employees with access to customer data?

**Suggested Score:** 4

**Response:**
Yes. Databricks conducts comprehensive background checks on all employees prior to their start date as part of our pre-employment screening process. Background checks are conducted through a third-party provider and include criminal history verification, employment verification, and education verification where applicable and permitted by law. Employees with access to customer data must pass these checks before being granted system access. This process is documented in our HR security procedures and is audited annually as part of our SOC 2 Type II examination.

**Evidence:**
- SOC 2 Type II Report (Section: Human Resources Security)
- Databricks Security Policies - HR Security SOP
- ISO 27001 Certificate (Control A.7: Human Resource Security)
```

### Infrastructure Example

```markdown
### Q34: Is all data encrypted in transit using TLS 1.2 or higher?

**Suggested Score:** 5

**Response:**
Yes. All data transmitted to and from Databricks is encrypted using TLS 1.2 or higher. This applies to:
- Customer connections to the Databricks platform (web UI, APIs, JDBC/ODBC)
- Internal service-to-service communication within the Databricks control plane
- Communication between the control plane and customer compute resources
- Data movement within customer workspaces

Databricks enforces TLS 1.2 as the minimum version and supports TLS 1.3 where available. SSL/TLS certificates are managed through automated certificate management and rotated regularly. Weak cipher suites are disabled.

**Evidence:**
- Databricks Trust Center: Security Features (https://databricks.com/trust/security-features)
- Enterprise Security Guide - Network Security section (NDA required)
- SOC 2 Type II Report (Section: Encryption Controls)
```

### Access Control Example

```markdown
### Q48: Does your organization enforce multi-factor authentication (MFA)?

**Suggested Score:** 5

**Response:**
Yes. Databricks requires multi-factor authentication for all access to customer environments. MFA options include:

1. **Native Databricks MFA** - Time-based one-time passwords (TOTP) via authenticator apps
2. **SSO with Customer IdP** - Customers can integrate their identity provider (Okta, Azure AD, etc.) with their existing MFA policies enforced
3. **SCIM Provisioning** - Automated user lifecycle management integrated with customer identity systems

For Databricks employees accessing production systems, hardware security keys (FIDO2/WebAuthn) are required in addition to SSO MFA. Privileged access requires additional authentication factors.

**Evidence:**
- Databricks Trust Center: Security Features
- Authentication documentation (https://docs.databricks.com/administration-guide/users-groups/single-sign-on/)
- SOC 2 Type II Report (Section: Logical Access Controls)
- ISO 27001 Certificate (Control A.9: Access Control)
```

### Physical Security Example (CSP-Inherited)

```markdown
### Q22: Are data centers protected by physical access controls?

**Suggested Score:** 4

**Response:**
Yes. Databricks operates on cloud service providers (AWS, Azure, GCP) and inherits physical security controls from these providers. Databricks does not operate its own data centers.

CSP physical security controls include:
- 24/7 security personnel and video surveillance
- Multi-factor access control (badge + biometric)
- Visitor escort requirements and logging
- Perimeter security (fencing, barriers, bollards)
- Environmental controls (fire suppression, climate control)

Databricks validates CSP physical security controls through:
- Annual review of CSP SOC 2 Type II reports
- Review of CSP ISO 27001 certifications
- Contractual requirements in cloud service agreements

**Evidence:**
- AWS SOC 2 Type II Report / Azure SOC 2 Type II Report / GCP SOC 2 Type II Report
- Databricks Shared Responsibility Model
- Databricks SOC 2 Type II Report (Section: Third-Party Management)
```

## Response Writing Guidelines

### Do:

1. **Be specific** - Reference actual controls, processes, and certifications
2. **Use evidence** - Always cite supporting documentation
3. **Distinguish ownership** - Clarify what Databricks owns vs. inherits from CSPs
4. **Answer the question** - Address exactly what was asked
5. **Provide context** - Explain how controls work, not just that they exist

### Don't:

1. **Don't be vague** - Avoid "industry standard" or "best practices" without specifics
2. **Don't overclaim** - Don't suggest capabilities that don't exist
3. **Don't share NDA content** - Reference but don't quote NDA-protected documents
4. **Don't speculate** - If unsure, escalate rather than guess
5. **Don't copy/paste blindly** - Tailor approved answers from your knowledge base to the specific question context

## Scoring Guidance Quick Reference

| Score | When to Use |
|-------|-------------|
| 5 | Continuous improvement, industry-leading, externally audited |
| 4 | Managed process, consistently followed, documented (Databricks typical) |
| 3 | Process exists but inconsistently applied |
| 2 | Ad-hoc, partial coverage |
| 1 | No formal process |

See `SCORING_GUIDE.md` for detailed scoring criteria.

## Evidence Hierarchy

When citing evidence, prioritize in this order:

1. **Current certifications** - ISO 27001, SOC 2, PCI-DSS (strongest)
2. **Public documentation** - Trust Center, product docs
3. **Contractual documents** - Security Addendum, DPA
4. **Internal policies** - Security policies (reference only, don't share)
5. **NDA-protected reports** - SOC reports, pen test summaries (offer to share under NDA)
