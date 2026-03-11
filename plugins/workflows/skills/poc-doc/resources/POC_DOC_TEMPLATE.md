# POC Document Template

Use this template structure when creating POC Google Docs. Write this content as markdown, then convert using:
```
python3 ~/.claude/plugins/cache/claude-vibe/google-tools/*/skills/google-docs/resources/markdown_to_gdocs.py --input /tmp/poc_doc.md --title "[External] [Customer]-Databricks POC DOC"
```

## Document Title Format

```
[External] [Customer Name]-Databricks POC DOC
```

Example: `[External] Meta-Databricks POC DOC`

---

## Document Structure

### Header

```markdown
# [Customer Name] // Databricks
## [Use Case / Product Area] POC
```

Example:
```markdown
# CrowdStrike // Databricks
## Unity Catalog and Managed Iceberg POC
```

---

### Key Contacts

**Databricks Key Contacts:**

| Name | Role |
|------|------|
| @FirstName LastName | Account Executive |
| @FirstName LastName | Solutions Architect (Primary SA) |
| @FirstName LastName | SA Manager |
| @FirstName LastName | Professional Services Director |
| @FirstName LastName | Specialist Solutions Architect |

**[Customer Name] Key Stakeholders:**

| Name | Title / Role |
|------|-------------|
| Name | Director, Data Engineering |
| Name | VP of Engineering |
| Name | Sr. Software Engineer |

NOTE: Include all key stakeholders from both sides. For Databricks, include executive sponsors, SA leadership, engineering overlays, and PS resources where applicable. For the customer, include decision makers, technical leads, and hands-on engineers.

---

### Executive Alignment

List executive meetings chronologically:

```markdown
- **[Date]:** [Customer Exec] ([Title]) and [Databricks Exec] ([Title]) - [Meeting Type/Purpose]
- **[Date]:** [Customer Exec] ([Title]) and [Databricks Exec] ([Title]) - [Meeting Type/Purpose]

**Next Alignment:**
- **[Date]:** [Customer Exec] ([Title]) and [Databricks Exec] ([Title])
```

Example:
```markdown
- **7/10/25:** JP Loose (Meta Director of Analytics) and Reynold Xin (Databricks Co-Founder) Alignment Call Completed
- **8/21/25:** JP Loose and Srikanth Sakhamuri (DE Leader) and Ron Gabrisko (Databricks CRO)

**Next Alignment:**
- **12/1/25:** Anil Wilson (VP of Enterprise Engineering) and Arsalan Tavakoli (Databricks Co-Founder)
```

---

### Summary & Scope

[1-2 paragraphs describing the business context and why the customer is evaluating Databricks]

Example:
> The Enterprise Engineering team at Meta is evaluating cloud data platforms to serve as their single platform for all things enterprise related to power Supply Chain, Finance, HR, and other key internal applications at Meta.

---

### Business Use Cases

For each use case:

```markdown
**[Use Case Category] - [Use Case Name]**

- **Overview:** [2-3 sentences describing the business problem and current state]
- **Goals:**
  - [Measurable goal 1]
  - [Measurable goal 2]
```

Example:
```markdown
**Supply Chain - Supply Chain Command Center Use Case**

- **Overview:** Meta will ingest inventory, order data, and other supply chain data in near real time to enable analytics that react faster to their dynamic product fulfillment needs. They currently are only able to provide daily and sometimes hourly analytics to their users which has caused inventory shortages and other fulfillment issues in the past.
- **Goals:**
  - AI-driven productivity enhancement to perform data analysis with no code.
  - Real-time data availability for planning inventory and decision making on manufacturing and inventory.
```

---

### Positive Business Outcomes

```markdown
**Outcome #1: [Outcome Title]**

- **Description:** [How Databricks addresses the need and what value it delivers. Be specific about the technical and business impact.]
```

Example:
```markdown
**Outcome #1: Decoupled Cloud Data Platform to Increase Developer Velocity and Feature Availability**

- **Description:** Having a cloud data platform that is not tightly coupled to the core monolithic data platform enables the Enterprise Engineering team to develop new features faster, deploy more robust functionality with all the capabilities & interoperability of Databricks, and manage the environment easily in a single platform for any data need.

**Outcome #2: Single Platform for Agentic AI system development in the cloud**

- **Description:** Implementing Unity Catalog & Databricks will give Meta a single platform to develop AI-native systems that range from simple OOTB solutions to totally bespoke needs. This will serve as a large productivity catalyst to business users at scale via users time saved as well as platform engineering time saved.

**Outcome #3: Near real-time data availability**

- **Description:** With a cloud native streaming platform, Meta will be able to keep their data up to date as often as each use case required to get access to data within minutes as it arrives.
```

---

### Project Summary

```markdown
As a next step, [Customer Name] & Databricks will engage in a Proof of Concept Evaluation, which will run from [Start Date] and [End Date]. The goal of this Evaluation is to demonstrate the value Databricks can provide as measured by the evaluation criteria below.
```

Include a note about timeline chart if applicable:
```markdown
The timeline for all phases will look as follows:
[Image placeholder - link to Gantt chart or timeline visualization]
([Link to live Chart])
```

---

### POC Scope & Success Criteria

| OKR | Business Value | Scope | Success Criteria | Status |
|-----|---------------|-------|-----------------|--------|
| **Phase 1 - [Phase Name]** | | | | |
| **OKR #1 - [OKR Title]** | [Why this matters to the business] | **Data Sources/Tables:** [List]; **Framework:** [Technology] | [Quantitative metric 1]; [Quantitative metric 2] | |
| **OKR #2 - [OKR Title]** | [Why this matters to the business] | **Data Sources/Tables:** [List]; **Framework:** [Technology] | [Quantitative metric] | |
| **Phase 2 - [Phase Name]** | | | | |
| **OKR #3 - [OKR Title]** | [Why this matters to the business] | **Tables/Data Sources to Query:** [List]; **Framework:** [Technology] | [Quantitative metric 1]; [Quantitative metric 2] | |
| **Phase 3 - [Phase Name]** | | | | |
| **OKR #4 - [OKR Title]** | [Why this matters to the business] | **Data Sources/Tables:** [List]; **Framework:** [Technology] | [Quantitative metric 1]; [Quantitative metric 2] | |

Example success criteria formats:
- `<5 minute data delivery from raw to gold layer`
- `Ingestion of 60GB per day for scoped datasets`
- `P95 SLA of 3-4 seconds`
- `Up to 600 concurrent users at peak time`
- `50 pre-prepared data analysis questions as benchmark`
- `Sustained throughput without degradation`
- `Enable persona-based row-filtering via formal policies (Yes/No)`

---

### Tasks and Timeline

**Abbreviation Dictionary** (include if relevant):
- POC - Proof of Concept
- SA - Solutions Architect
- PS - Professional Services
- [Add domain-specific abbreviations]

| Task | [Customer] Owner | Databricks Owner | Comments | Target Due Date | Status | Notes |
|------|-----------------|-----------------|----------|----------------|--------|-------|
| **Kickoff & Environment Setup** | | | | | | |
| Kickoff call - full team alignment | [Name] | @FirstName LastName | | [Date] | | |
| Workspace creation and configuration | | @FirstName LastName | | [Date] | | |
| Access and connectivity validation | [Name] | @FirstName LastName | | [Date] | | |
| **Phase 1 - [Phase Name]** | | | | | | |
| [Data delivery / integration task] | [Name] | @FirstName LastName | | [Date] | | |
| [Testing / validation task] | [Name] | @FirstName LastName | | [Date] | | |
| **Phase 2 - [Phase Name]** | | | | | | |
| [Task description] | [Name] | @FirstName LastName | | [Date] | | |
| **Readout & Conclusions** | | | | | | |
| Final readout meeting | [Name] | @FirstName LastName | Full team | [Date] | | |

---

### POC Staffing Plan

```markdown
Databricks & [Customer Name] will jointly collaborate on this POC with the following resources:
```

**Databricks Resources:**

| Name | Role | Focus Area | Engagement Model |
|------|------|-----------|-----------------|
| @FirstName LastName | Primary Lead SA | All | End to End Ownership, Project lead for POC |
| @FirstName LastName | SA | [Specific area] | Focused Overlay |
| @FirstName LastName | Professional Services | [Specific area] | Full time dedicated |
| @FirstName LastName | SA Manager | Overlay, PM interface | Full time overlay |
| @FirstName LastName | Specialist SA, [SME area] | [Specific expertise] | Part time overlay |
| @FirstName LastName | [Product/Engineering role] | [Feature area] | As needed overlay |

**[Customer Name] Resources:**

| Name | Role | Focus Area | Engagement Model |
|------|------|-----------|-----------------|
| Name | [Title] | [Area] | [Model] |

**Meeting Schedule:**
- Kickoff Call Day 1 with full team on both sides
- Weekly Progress Report on [Day]
- Daily [AM/PM] Standup ([Duration] goal) for [Customer] Eng / Databricks SA
- Final readout when complete with full team on both sides

---

### Current State & Architecture

[Image placeholder - architecture diagram of customer's current state]

**Key Components & Challenges:**

- **[Component 1]** - [Description and challenges]
- **[Component 2]** - [Description and challenges]
- **[Component 3]** - [Description and challenges]

Example:
```markdown
**Key Components & Challenges:**

- **On-Prem Architecture** - Currently all infrastructure is hosted in the customer's data centers. They are looking to add a cloud-based data platform to supplement this infrastructure.
- **Push based data extraction** - All data is pushed out of the data center into the cloud. Pull-based federation not currently being considered.
- **Monolithic Architecture** - All downstream business users are customers of the core data platform team. This creates a single bottleneck that makes it hard to iterate independently.
- **Not set up for real-time analytics** - Most workloads are refreshed daily/hourly at most. Some data sets need to be delivered in under 1 minute.
```

---

### Mutual Success Plan

| Objective | Actions | Owner | Timeline | Status |
|-----------|---------|-------|----------|--------|
| Funding / Budget Approval | [Specific actions] | [Name] | [Date range] | |
| Executive Alignment Meeting | Schedule [Exec] / [Exec] call | @FirstName LastName | [Date] | |
| Workspace & Environment | Create and configure workspace | @FirstName LastName | [Date range] | |
| Phase 1 Completion | [Phase 1 deliverables] | Joint | [Date range] | |
| Phase 2 Completion | [Phase 2 deliverables] | Joint | [Date range] | |
| Final Readout | Present results and recommendations | @FirstName LastName | [Date] | |
| Post-POC Survey | Send CSAT/feedback survey | @FirstName LastName | [Date] | |

---

### Evaluation Results

*To be filled out during and post Proof of Concept*

| OKR | Target | Result | Status |
|-----|--------|--------|--------|
| OKR #1 - [Title] | [Target metric] | | |
| OKR #2 - [Title] | [Target metric] | | |
| OKR #3 - [Title] | [Target metric] | | |
| OKR #4 - [Title] | [Target metric] | | |

---

### Resources

- [Product Documentation](https://docs.databricks.com)
- [Unity Catalog Documentation](https://docs.databricks.com/en/data-governance/unity-catalog/index.html)
- [Demo Overview Deck](link)
- [Additional relevant documentation]

---

### Notes

[Running notes section - to be filled during POC]

---

## Formatting Reminders

1. **Company names:** Always use actual company name, never "customer" or "the customer"
2. **People:** Use `@FirstName LastName` format for Databricks employees; include roles
3. **Links:** Embed all links - don't show raw URLs
4. **Success criteria:** Must be quantitative and measurable with units/thresholds
5. **Status columns:** Leave empty for new POC docs; fill in during execution
6. **Phase headers:** Bold within tables to visually separate phases
7. **Dates:** Use MM/DD or Month Day format consistently throughout
8. **Engagement models:** Use "Full time dedicated", "Full time overlay", or "Part time overlay"
