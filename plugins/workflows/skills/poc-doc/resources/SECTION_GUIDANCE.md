# POC Document Section Guidance

Detailed guidance for each section of the POC document. Use this to ensure quality and completeness.

---

## Key Contacts

### What to Include
- **Databricks side:** Account Executive, Primary SA, SA Manager, PS Director/Engagement Manager, Specialist SAs, Engineering overlays (product managers, staff engineers involved), executive sponsors
- **Customer side:** Decision makers (VP/Director level), technical leads, hands-on engineers, program managers

### Quality Criteria
- Every person listed has a clear role/title
- Include enough context for someone unfamiliar with the account to understand the org structure
- For large teams (10+ people on either side), consider grouping by function
- Executive sponsors should be listed even if they aren't hands-on

### Common Mistakes
- Listing only the SA and AE from Databricks (include all overlay resources)
- Missing the customer's executive decision maker
- Not including engineering/product overlays who will be critical during the POC

---

## Executive Alignment

### What to Include
- Chronological list of completed executive meetings with dates
- Attendees by name and title on both sides
- Meeting type or outcome (e.g., "Alignment Call Completed", "Dinner in Menlo Park")
- Upcoming scheduled meetings under "Next Alignment"

### Quality Criteria
- Shows consistent executive engagement cadence (monthly or bi-weekly)
- Pairs customer executives with appropriately senior Databricks leaders
- Demonstrates momentum and relationship building over time

### Why This Section Matters
- Signals to the customer that Databricks leadership is invested
- Creates accountability for follow-through
- Provides a record of executive commitments made

---

## Summary & Scope

### What to Include
- Business context: What does this team/department do at the company?
- Current pain points: Why are they looking for a new solution?
- Evaluation driver: What triggered this POC? (Contract renewal, new initiative, competitive pressure)
- Teams involved: Which business units or departments will use the platform?

### Quality Criteria
- Written from the customer's perspective (their needs, their pain points)
- Avoids Databricks marketing language
- Specific enough to differentiate this POC from a generic evaluation
- 1-2 paragraphs maximum

### Common Mistakes
- Writing a Databricks sales pitch instead of describing the customer's situation
- Being too generic ("Company X wants a modern data platform")
- Not mentioning the competitive landscape if relevant

---

## Business Use Cases

### What to Include
For each use case:
- **Category and name** that maps to the customer's internal terminology
- **Overview** describing the current state and desired future state (2-3 sentences)
- **Goals** that are specific and measurable

### Quality Criteria
- Use cases are described in the customer's language, not Databricks terminology
- Goals connect to measurable business outcomes
- Each use case is distinct and well-scoped
- 2-5 use cases is typical; more than 5 suggests scope may be too broad

### Mapping to Success Criteria
Every business use case should map to at least one OKR in the POC Scope & Success Criteria table. If a use case doesn't have a corresponding OKR, either:
1. Add an OKR for it, or
2. Explicitly note it's out of POC scope but important for the broader relationship

---

## Positive Business Outcomes (PBOs)

### What to Include
- 3-4 high-level outcomes that articulate the business value
- Each outcome should have a descriptive title and 2-3 sentence description
- Focus on what changes for the customer, not what Databricks does

### Quality Criteria
- Outcomes are written from the customer's perspective
- Each outcome connects technical capabilities to business impact
- Outcomes are differentiated (not just restating the same value in different words)
- Language matches what a VP/C-level would care about (velocity, cost, risk, scale)

### PBO Categories (common patterns)
1. **Developer velocity / reduced time to market** - Decoupled architecture, modern tooling
2. **AI/ML capabilities** - Unified platform for analytics and AI
3. **Real-time data** - Streaming, low-latency data delivery
4. **Governance and security** - Unity Catalog, fine-grained access control
5. **Cost optimization** - Better price/performance, reduced operational overhead
6. **Operational simplicity** - Single platform, reduced tool sprawl

---

## POC Scope & Success Criteria

### What to Include
- OKRs organized by phase
- Quantitative success criteria for every OKR
- Data sources and frameworks in scope
- Business value statement for each OKR

### Quality Criteria
- **Every success criterion is measurable** - includes numbers, thresholds, or yes/no
- **Scope is specific** - lists exact tables, data volumes, frameworks
- **Phases are logical** - foundational work first, advanced capabilities later
- **Business value is customer-centric** - explains why this metric matters to them

### Phase Organization (common patterns)
- **Phase 1:** Data delivery, ingestion, ETL, catalog setup (foundational)
- **Phase 2:** Query serving, analytics, governance (value delivery)
- **Phase 3:** AI/ML, Genie, advanced analytics (differentiation)

### Success Criteria Formats
Good examples:
- `<5 minute data delivery from raw to gold layer for Inventory dataset`
- `Ingestion of 60GB per day for scoped datasets`
- `P95 SLA of 3-4 seconds`
- `Up to 600 concurrent users at peak time`
- `50 pre-prepared data analysis questions as benchmark`
- `Sustained throughput >1PB/day without degradation`
- `Enable persona-based row-filtering via formal policies (Yes/No)`

Bad examples:
- `Performs well` (not measurable)
- `Meets expectations` (whose expectations?)
- `Fast enough` (no threshold)
- `Scalable` (to what level?)

---

## Tasks and Timeline

### What to Include
- Every task needed to complete the POC, organized by phase
- Named owners on both sides for each task
- Target due dates
- Comments and notes columns for context

### Quality Criteria
- No task without an owner
- Dates are realistic and account for dependencies
- Include buffer time for unexpected issues
- Kickoff and readout tasks are always included
- Tasks are specific enough to be actionable (not "do Phase 1")

### Standard Task Categories
1. **Kickoff & Setup:** Kickoff call, workspace creation, access provisioning, connectivity validation
2. **Data Delivery:** Source connectivity, ingestion pipelines, data quality validation
3. **Testing & Validation:** Performance benchmarks, functional testing, security validation
4. **Analytics & AI:** Dashboard creation, Genie rooms, model serving
5. **Readout:** Results compilation, final presentation, next steps discussion

---

## POC Staffing Plan

### What to Include
- Named resources on both sides with roles, focus areas, and engagement models
- Meeting cadence (kickoff, standups, weekly syncs, final readout)
- Clear distinction between full-time dedicated, full-time overlay, and part-time overlay

### Engagement Model Definitions
- **Full time dedicated:** This person is working on this POC as their primary focus
- **Full time overlay:** Available full-time during the POC but has other responsibilities
- **Part time overlay:** Available for specific topics as needed, not day-to-day
- **As needed:** Engineering/product resources available for escalations

### Quality Criteria
- Primary SA is clearly identified as the project lead
- Specialist resources are mapped to specific OKRs/phases
- Customer-side resources include both technical and business stakeholders
- Meeting schedule is realistic (daily standups should be 15 min)

---

## Current State & Architecture

### What to Include
- Description of the customer's current data architecture
- Key technologies and components
- Known challenges and limitations
- Architecture diagram (or placeholder for one)
- Competitors being evaluated (if applicable)

### Quality Criteria
- Written to help someone unfamiliar with the account understand the landscape
- Highlights specific pain points that the POC addresses
- Includes enough technical detail for the SA team to understand integration points
- Mentions competitors neutrally if relevant

### Common Components to Document
- Data sources (on-prem, cloud, SaaS)
- Current data platform (Hadoop, Spark, Presto, Snowflake, etc.)
- Data delivery method (push vs pull, batch vs streaming)
- Analytics tools (Tableau, internal tools, etc.)
- AI/ML infrastructure (if any)
- Governance and security model

---

## Mutual Success Plan

### What to Include
- Execution milestones beyond just the technical POC
- Business/commercial milestones (funding, procurement)
- Executive alignment milestones
- Post-POC actions (surveys, decision timeline)

### Quality Criteria
- Covers the full lifecycle from POC kickoff to contract close
- Includes actions on both sides
- Has realistic timelines
- Tracks commercial and technical milestones together

---

## Evaluation Results

### What to Include
- Results against each OKR from the Success Criteria table
- Quantitative metrics achieved
- Comparison to success criteria thresholds
- Pass/fail or status for each OKR

### Quality Criteria
- Filled in during the POC, not just at the end
- Results use the same units as the success criteria
- Includes context for any misses (why, what was tried, what's needed)
- Honest assessment - don't inflate results

### When to Fill This Section
- **During POC:** Update as each phase completes
- **Post-POC:** Final compilation with all results
- **If incomplete:** Note which OKRs weren't tested and why

---

## Information Gathering Strategy

When building a POC document, gather information in this order:

### 1. Salesforce (Account and Opportunity context)
- Account name, industry, size
- Opportunity details, competitors, timeline
- Related use cases and contacts

### 2. Existing Documents (Google Docs, Slides)
- Prior POC proposals or scoping documents
- Technical design documents
- Meeting notes from customer calls

### 3. Slack (Internal context)
- Account-specific channels for recent discussions
- Technical challenges and solutions discussed
- Stakeholder feedback and concerns

### 4. JIRA (Technical support context)
- Open ES tickets for the account
- Historical technical issues
- Engineering overlay requests

### 5. Glean (Broad search)
- Internal documents about the customer
- Similar POC documents for reference
- Competitive intelligence

### 6. User Input (Fill gaps)
- Success criteria details
- Timeline and staffing specifics
- Customer-specific requirements not found in existing sources
