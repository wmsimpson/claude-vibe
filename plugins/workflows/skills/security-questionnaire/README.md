# Security Questionnaire Skill

A comprehensive skill for completing security questionnaires and compliance documentation for customer procurement processes.

> **NOTE:** Answers should be verified against your organization's specific security policies before sharing with customers.

## Table of Contents

- [Overview](#overview)
- [How to Invoke](#how-to-invoke)
- [Critical Policies](#critical-policies)
- [Complete Workflow](#complete-workflow)
- [Triage Flowchart](#triage-flowchart)
- [Question Categories](#question-categories)
- [Research Workflow](#research-workflow)
- [Validation Workflow](#validation-workflow)
- [Resources](#resources)
- [Related Agents](#related-agents)

---

## Overview

The Security Questionnaire skill guides you through completing customer security questionnaires by:

1. **Triaging** the request to determine the appropriate handling path
2. **Avoiding** custom questionnaires when pre-built content suffices
3. **Researching** answers using the SQRC database (primary source)
4. **Drafting** responses with proper evidence and scoring
5. **Validating** responses before sharing with customers

### Key Features

- Enforces critical policy guardrails (NDA verification, caution with GenAI)
- Targets 90-95% completion using pre-approved answers from your knowledge base
- Distinguishes your organization's controls vs CSP-inherited controls
- Includes post-completion validation agent

---

## How to Invoke

### Direct Invocation

```
/security-questionnaire
```

### Example Prompts

```
"I need to complete a security questionnaire for a customer"

"Help me answer questions about our SOC 2 and security certifications"

"Customer sent a compliance assessment, can you help fill it out?"
```

### Prerequisites

Before using this skill, ensure you have access to:

| Tool | Purpose |
|------|---------|
| Glean MCP | Search SQRC database and internal docs |
| Slack MCP | Escalation and expert consultation |
| Google Docs skill | Read/create questionnaire documents |

---

## Critical Policies

| Policy | Action | Consequence |
|--------|--------|-------------|
| NDA Verification | **REQUIRED** before sharing NDA-protected docs | Legal exposure if violated |
| GenAI Answers | **VERIFY** - Cross-reference with authoritative docs | May contain inaccuracies |
| Custom Answers | **AVOID** - Use pre-built content when possible | Not reviewed by security/legal |
| Knowledge Base First | **RECOMMENDED** - 90-95% from approved sources | Expert-reviewed answers |

---

## Complete Workflow

![Complete Workflow](diagrams/complete-workflow.svg)

### Workflow Phases

| Phase | Steps | Purpose |
|-------|-------|---------|
| **Initial Checks** | 1-2 | Verify not Azure, check NDA |
| **Avoidance** | 3 | Try pre-built content first |
| **Triage** | 4 | Route appropriately (ESG, Privacy, deal size) |
| **Completion** | 5-8 | Categorize, research, draft, verify controls |
| **Validation** | 9-11 | Validate, review if needed, deliver |

---

## Triage Flowchart

![Triage Decision Tree](diagrams/triage-flowchart.svg)

### Quick Routing Reference

| Condition | Route To |
|-----------|----------|
| NDA-protected docs requested | Confirm NDA before sharing |
| ESG questions | Your ESG team channel |
| Privacy questions | Your Privacy team channel |
| Complex / large deal | Your security team intake process |
| Standard questions | Self-service with this skill |

---

## Question Categories

![9 Security Question Categories](diagrams/question-categories.svg)

### Category Details

See `resources/CATEGORY_GUIDANCE.md` for detailed guidance on answering questions in each category.

---

## Research Workflow

![Research Priority Order](diagrams/research-workflow.svg)

### Glean Search Examples

```bash
# Search SQRC for training questions
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "SQRC security training awareness", "page_size": 15}}'

# Search SQRC for penetration testing
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "SQRC penetration test vulnerability", "page_size": 15}}'

# Search for supporting evidence
mcp-cli call glean/glean_read_api_call '{"endpoint": "search.query", "params": {"query": "Databricks SOC 2 Type II report", "page_size": 15}}'
```

---

## Validation Workflow

![SQRC Validator - 7 Validation Checks](diagrams/validation-workflow.svg)

### Validation Decision Matrix

| Condition | Result |
|-----------|--------|
| Azure detected | **ESCALATE** - Cannot share |
| GenAI suspected | **NEEDS REVIEW** |
| >5 custom answers | **NEEDS REVIEW** |
| <90% completion | **NEEDS REVIEW** |
| Control misattribution | **NEEDS REVIEW** |
| NDA content + unverified NDA | **NEEDS REVIEW** |
| All checks pass | **READY TO SHARE** |

---

## Control Ownership

![Databricks vs CSP Control Ownership](diagrams/control-ownership.svg)

### How to Frame Responses

1. **Lead with your organization's directly-owned controls** - These are what your organization directly manages
2. **Reference CSP controls as "inherited"** - Note the cloud provider handles these
3. **Validate through compliance reviews** - Your organization validates CSP controls through their compliance report reviews (SOC 2, ISO certs)

---

## Resources

### Skill Resources (in `resources/` directory)

| File | Description |
|------|-------------|
| `TRIAGE_FLOWCHART.md` | Decision tree for routing questionnaires |
| `CATEGORY_GUIDANCE.md` | Detailed guidance for each question category |
| `SCORING_GUIDE.md` | Maturity scoring criteria (1-5 scale) |
| `EVIDENCE_SOURCES.md` | Pre-built Glean queries by category |
| `RESPONSE_TEMPLATE.md` | Standard response format template |

### Internal Resources

> Configure these for your organization.

| Resource | How to Access | Description |
|----------|---------------|-------------|
| Security Q&A Database | Your internal knowledge base (Confluence, Glean, Drive) | Pre-answered questions |
| Submit for Review | Your security team's intake form or Slack channel | Complex questionnaire submission |
| Security Package | Request from security team | Pre-built security documentation |
| Pre-filled SIG Questionnaire | Request from security team | Industry-standard pre-filled answers |
| Security FAQ | Internal wiki / knowledge base | Security and compliance FAQ |

### Public Resources (Databricks-Specific)

If you work with Databricks and are completing questionnaires about Databricks products:

| Resource | URL |
|----------|-----|
| Trust Center | https://databricks.com/trust |
| Compliance | https://databricks.com/trust/compliance |
| Security Features | https://databricks.com/trust/security-features |
| Privacy Notice | https://databricks.com/legal/privacynotice |
| Security Addendum | https://databricks.com/legal/security-addendum |
| Subprocessors | https://databricks.com/legal/databricks-subprocessors |

---

## Related Agents

### sqrc-questionnaire Agent

For complex questionnaires with 20+ questions, use the `sqrc-questionnaire` agent:

```
Use when:
- Questionnaire has 20+ questions
- Complex compliance assessment
- Need parallel research across categories
```

The agent will:
1. Parse and categorize all questions
2. Research using your internal knowledge base as primary source
3. Draft responses with evidence
4. Target 90-95% completion from approved sources
5. Flag items for escalation

### sqrc-validator Agent

Always run validation before sharing with customers:

```
Use when:
- After completing questionnaire drafting
- Before submitting for security review
- Before delivering to customer
```

The agent validates:
1. Source verification (answers from approved docs)
2. GenAI content detection
3. Custom answer flagging
4. Completeness check (90-95%)
5. Control ownership verification
6. NDA content check

---

## Quick Reference

### Do's

- **DO** check your internal knowledge base first for every question
- **DO** copy-paste approved answers directly (don't paraphrase)
- **DO** verify NDA status before sharing sensitive docs
- **DO** run sqrc-validator before delivering (if available)
- **DO** target 90-95% completion from approved sources
- **DO** verify AI-generated answers against authoritative documentation

### Don'ts

- **DON'T** rely on unverified GenAI answers for security questions
- **DON'T** paraphrase or rewrite approved answers
- **DON'T** write custom answers without escalation to your security team
- **DON'T** share NDA content without verified NDA

---

## Escalation Contacts

> Configure these for your organization.

| Type | Channel | When to Use |
|------|---------|-------------|
| ESG | Your ESG team channel | Sustainability, governance |
| Privacy | Your Privacy team channel | GDPR, data subject rights |
| Security | Your security team channel | Urgent requests, clarifications |
| Trust / Legal Team | Your trust/legal intake | Custom attestations, regulatory |
