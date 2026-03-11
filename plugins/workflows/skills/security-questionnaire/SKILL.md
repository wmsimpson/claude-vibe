---
name: security-questionnaire
description: Complete security questionnaires, assessments, and compliance documentation for customer procurement
user-invocable: true
---

# Security Questionnaire Workflow

This skill guides you through completing security questionnaires and compliance documentation for customer procurement processes. It uses available documentation, web research, and your organization's internal resources.

> **NOTE:** Answers should be verified against your organization's specific security policies before sharing with customers. This skill provides a structured approach to researching and drafting responses, not pre-approved answers.

## Critical Policies

> **⚠️ Verify Before Sharing:** Do not share answers you cannot verify. Use your organization's official security documentation, compliance reports (SOC 2, ISO certs), and public Trust Center as primary sources.

> **⚠️ Avoid Unverified GenAI Answers:** Be cautious using AI-generated answers for security questions — they may contain inaccuracies. Always cross-reference with authoritative documentation.

> **⚠️ Avoid Custom Answers for Standard Topics:** Use pre-built content from your organization's security documentation whenever possible. Custom answers risk errors and may not have been reviewed by security/legal.

## Prerequisites

Before starting:

1. **Verify NDA status** - Confirm the customer has a signed NDA before sharing sensitive security documentation
2. **Authenticate** - Ensure you have access to:
   - Glean MCP (internal documentation search)
   - Slack MCP (for escalation and expert consultation)
   - Google Docs skill (for reading/creating questionnaire documents)

## Strategy 1: Avoid the Questionnaire

Before filling out a custom questionnaire, **always try to redirect the customer to pre-built content first**. This saves time and ensures accuracy.

**Resources to offer instead of filling custom questionnaires:**

| Resource | Description | How to Access |
|----------|-------------|---------------|
| Security & Compliance Package | Comprehensive security documentation bundle | Request from your security/legal team |
| Pre-filled SIG Questionnaire | Industry-standard pre-answered questionnaire | Request from your security team (if available) |
| Trust Center | Public security and compliance information | Your organization's public trust/security page |
| Standard Security FAQ | Commonly asked questions with approved answers | Your internal security knowledge base |

**Suggested language to redirect customers:**

> "We have a comprehensive pre-filled security questionnaire that covers most standard security questions. Would you be willing to review that instead? It's been reviewed by our security and legal teams."

> "Rather than filling out a custom questionnaire, we can provide our Security & Compliance Package which includes our SOC 2 report, ISO certificates, and detailed security documentation. This typically addresses 90%+ of customer security questions."

**When to proceed with custom questionnaire:**

- Customer explicitly rejects pre-built options
- Questionnaire contains organization-specific questions
- Regulatory requirements mandate specific format

## Phase 1: Triage Flowchart

Follow the decision tree in `resources/TRIAGE_FLOWCHART.md` to determine the appropriate handling path.

```
Read the triage flowchart resource:
/read plugins/workflows/skills/security-questionnaire/resources/TRIAGE_FLOWCHART.md
```

### Quick Decision Summary

1. **Does the customer have a signed NDA?**
   - Some security documents (SOC 2 reports, penetration test summaries) require a signed NDA before sharing
   - Confirm NDA status before sharing any NDA-protected content

2. **Can your organization's pre-built content answer it?**
   - Check your internal security knowledge base / approved Q&A library first
   - Copy-paste approved answers - do not paraphrase or rewrite them

3. **Is this ESG/Privacy focused?**
   - ESG (Environmental, Social, Governance) → Route to your ESG/sustainability team
   - Privacy-specific → Route to your Privacy team

4. **Does the deal/engagement require security team review?**
   - Large deals or complex questionnaires → Submit for security team review per your internal process
   - Smaller deals → Self-service using available documentation + this skill

## Phase 2: SQRC Completion Workflow

### Step 1: Parse and Categorize Questions

Analyze the questionnaire and categorize each question into one of 9 categories:

| # | Category | Common Topics |
|---|----------|---------------|
| 1 | Policy & People | Training, NDAs, background checks, AUP |
| 2 | Risk Management | Assessments, SOC reports, certifications |
| 3 | Incident Management | SIRT, breach notification, forensics |
| 4 | Physical Security | CSP-inherited, visitor management, media destruction |
| 5 | Infrastructure | Encryption, MFA, patching, network security |
| 6 | Application Security | SDLC, environment segregation, code review |
| 7 | Access Control | SCIM, password policies, least privilege |
| 8 | Business Continuity | DR, RTO/RPO, backups, geographic redundancy |
| 9 | Data Security | PCI, PII, subprocessors, data classification |

See `resources/CATEGORY_GUIDANCE.md` for detailed guidance on each category.

### Step 2: Research Using Internal Knowledge Base (Primary) and Web Research (Secondary)

> **⚠️ WARNING:** Always use authoritative sources. Cross-reference AI-generated answers with official documentation before including them.

**Research Priority Order:**

1. **PRIMARY - Internal knowledge base / approved Q&A library:** Search your organization's security documentation library (e.g., Glean, Confluence, shared Drive folder) for pre-approved answers. Copy-paste the detailed answers directly.
2. **SECONDARY - Official public documentation:** Search your organization's Trust Center, public docs, and compliance pages for supporting evidence.
3. **TERTIARY - Web research:** Use WebSearch for publicly available compliance information (e.g., certifications, regulatory frameworks). Verify before using.
4. **CAUTION with GenAI:** AI-generated answers may contain inaccuracies for security topics. Always verify against authoritative sources.

If you have Glean MCP configured:

```bash
# Example: Search internal knowledge base for security training questions
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "security training awareness program", "page_size": 15}}'

# Example: Search for penetration testing documentation
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "penetration test vulnerability assessment", "page_size": 15}}'

# Example: Search for compliance certifications
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "SOC 2 Type II ISO 27001 certification", "page_size": 15}}'
```

See `resources/EVIDENCE_SOURCES.md` for example search queries by category.

### Step 3: Apply Maturity Scoring

Score each response using the 1-5 maturity scale:

| Score | Level | Description |
|-------|-------|-------------|
| 5 | Optimized | Continuous improvement, regularly audited, industry-leading |
| 4 | Managed | Measured and controlled, consistently followed (Databricks typical) |
| 3 | Defined | Documented process, may not be consistently applied |
| 2 | Developing | Ad-hoc processes, partial coverage |
| 1 | Initial | No formal process |

See `resources/SCORING_GUIDE.md` for detailed scoring criteria with examples.

### Step 4: Draft Responses

Use the standard response format from `resources/RESPONSE_TEMPLATE.md`:

```markdown
### Q[#]: [Question text]

**Suggested Score:** [1-5]

**Response:** [Detailed answer addressing the specific question]

**Evidence:**
- [Document/certification 1]
- [Document/certification 2]
```

### Step 5: Distinguish Your Organization's Controls vs Vendor/CSP Controls

**Critical for accurate responses:** Clearly distinguish what your organization directly controls vs what is inherited from cloud providers or sub-processors.

**Typically Organization-Owned Controls:**
- Security certifications and compliance attestations (ISO 27001, SOC 2, etc.)
- Security policies and procedures
- Incident response process
- Penetration testing program
- SDLC and vulnerability management
- Access control and authentication for your platform
- Encryption implementation
- Audit logging

**Typically Inherited from Cloud Service Providers (AWS, Azure, GCP):**
- Physical data center security
- Physical media destruction
- Network infrastructure redundancy
- Geographic availability zones

**How to frame responses:**
- Lead with controls your organization directly owns
- Reference CSP controls as "inherited" where applicable
- Note that your organization validates CSP controls through their compliance reports (SOC 2, ISO certs)

> **NOTE:** Answers should be verified against your organization's specific security policies before sharing with customers.

## Resources

### Internal Resources

> Configure these for your organization. Examples below.

| Resource | How to Access | Description |
|----------|---------------|-------------|
| Security Q&A Library | Internal knowledge base (Confluence, Glean, Drive) | Pre-answered common questions |
| Submit for Security Review | Your security team's intake form or Slack channel | Submit complex questionnaires |
| Security FAQ | Internal wiki / knowledge base | Security and compliance FAQ |
| Security Policies | Internal document repository | Comprehensive security policy documents |

### Public Resources (Databricks-Specific)

If you work with Databricks and are completing questionnaires about Databricks products:

| Resource | URL |
|----------|-----|
| Trust Center | https://databricks.com/trust |
| Compliance | https://databricks.com/trust/compliance |
| Security Features | https://databricks.com/trust/security-features |
| Privacy Notice | https://databricks.com/legal/privacynotice |
| Acceptable Use Policy | https://databricks.com/legal/aup |
| Subprocessors | https://databricks.com/legal/databricks-subprocessors |
| Security Addendum | https://databricks.com/legal/security-addendum |

### Documents to Provide to Customers

For customer completion, these documents may be requested:

1. **SOC 2 Type II Report** (NDA required)
2. **SOC 1 Type II Report** (NDA required)
3. **ISO 27001 Certificate**
4. **PCI-DSS Attestation of Compliance** (if applicable)
5. **Enterprise Security Guide** (NDA required, if available)
6. **Penetration Test Executive Summary** (NDA required)
7. **Data Processing Agreement (DPA)**
8. **Subprocessors List**
9. **Shared Responsibility Model** (cloud-specific)
10. **Security Addendum**

## Escalation Guidelines

### When to Engage Your Security Team

Submit for security team review when:

- Deal value exceeds your organization's threshold for self-service
- Questions are outside standard coverage in your knowledge base
- Customer requests custom security terms or contractual changes
- Questions about unreleased features or non-standard configurations
- Regulatory-specific requirements (FedRAMP, StateRAMP, HIPAA BAA, etc.)

### When to Engage Your Security Channel (Slack or Equivalent)

Use your team's security help channel for:

- Time-sensitive requests
- Clarification on existing answers
- Questions about recent security incidents or vulnerabilities
- Guidance on complex multi-cloud scenarios

### When to Escalate to Trust / Legal Team

Contact the Trust or Legal team directly for:

- Customer-specific security architecture reviews
- Custom compliance attestations
- Regulatory-specific requirements requiring custom documentation

## Complex Questionnaires

For questionnaires with 20+ questions or complex security assessments, use the `sqrc-questionnaire` agent (if available):

```
This agent will:
1. Parse the entire questionnaire
2. Categorize all questions
3. Research using internal knowledge base as primary source
4. Draft responses with evidence (copy-paste from approved sources)
5. Target 90-95% completion before flagging for review
6. Flag items needing escalation to the security team
```

## Post-Completion Validation

Before sharing questionnaire responses with the customer, use the `sqrc-validator` agent for a final review (if available):

```
This agent will validate:
1. Source verification - Verify answers came from approved documentation
2. Custom answer flagging - Flag any unverified answers for review
3. Completeness check - Verify 90-95% target met
4. Accuracy spot-check - Cross-reference key claims against documentation
5. Control ownership - Verify your organization vs CSP attribution
6. NDA content check - Flag NDA-protected content
```

> **NOTE:** Answers should be verified against your organization's specific security policies before sharing with customers.

## Workflow Summary

```
1. Receive security questionnaire from customer
   ↓
2. TRIAGE
   - Check NDA status (required for some documents)
   - Check deal value / complexity threshold
   - Determine if security team review is needed
   ↓
3. Try to AVOID the custom questionnaire
   - Offer Security Package, pre-filled SIG, Trust Center
   - Customer insists on custom? → Continue
   ↓
4. Follow Triage Flowchart (resources/TRIAGE_FLOWCHART.md)
   - Check NDA status
   - Check deal value threshold
   - Determine if review needed
   ↓
5. Categorize questions (9 categories)
   ↓
6. Research using internal knowledge base FIRST
   - Search internal docs / Glean for approved answers
   - Copy-paste approved answers (verify before using AI-generated content)
   - Target 90-95% from approved sources
   ↓
7. Draft responses with:
   - Maturity score (1-5)
   - Detailed response (from approved sources)
   - Supporting evidence
   ↓
8. Distinguish your organization vs. CSP controls
   ↓
9. Run VALIDATION (sqrc-validator agent, if available)
   - Source verification
   - Completeness check
   - Control ownership check
   - NDA content check
   ↓
10. Review with Trust/Security team if needed
    ↓
11. Deliver completed questionnaire
```

> **NOTE:** Answers should be verified against your organization's specific security policies before sharing with customers.
