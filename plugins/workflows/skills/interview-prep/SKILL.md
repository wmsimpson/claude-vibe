---
name: interview-prep
description: Prepare a structured interview plan for the Databricks SA/DSA Design & Architecture interview. Analyzes a candidate's resume and generates a tailored scenario, discovery questions, probing questions by phase, spike-specific deep dives, and a scoring rubric. Use this when preparing to interview a candidate.
user-invocable: true
---

# Interview Prep Skill

Generate a comprehensive, structured interview plan for the Databricks SA/DSA Design & Architecture interview. This skill reads a candidate's resume, applies the official interviewer playbook framework, and produces a detailed session guide tailored to the candidate's background and declared spike.

## Step 1: Gather Inputs

Before generating the plan, you MUST collect three pieces of information from the user. Use the AskUserQuestion tool to ask all three at once:

1. **Resume location**: Ask the user to provide the file path to the candidate's resume (PDF, DOCX, or text file), OR to paste the resume content directly.
2. **Technical Spike**: Ask which spike the candidate has declared. Options:
   - Data Engineering / ELT
   - Data Warehousing / Analytics
   - AI/ML
3. **Target Level**: Ask what level the candidate is being evaluated for. Options:
   - L4 (Associate SA/DSA)
   - L5 (SA/DSA)
   - L6 (Senior SA/DSA)
4. **Output location**: Ask where the interviewer wants the plan saved. Example: `~/Documents/WorkOS/interviews/` or any directory of their choice.

If the user already provided any of these in their initial message, do not re-ask for those.

## Step 2: Analyze the Resume

Read the resume file (or parse the pasted content). Extract:

- **Current/recent role and company** -- What domain are they in?
- **Core technical skills** -- What tools, languages, and platforms do they list?
- **Data platform experience** -- Any Databricks, Spark, cloud platform (AWS/Azure/GCP) experience?
- **Domain expertise** -- What industries have they worked in? (manufacturing, retail, healthcare, fintech, etc.)
- **ML/AI experience** -- Any model building, MLOps, feature engineering?
- **Scale indicators** -- Have they worked with large-scale data? What volumes?
- **Claimed achievements** -- What specific accomplishments could you probe on?
- **Databricks-specific experience** -- Unity Catalog, Delta Lake, MLflow, Feature Store, Workflows, etc.

Use this analysis to customize the scenario and identify areas where you can probe depth vs. surface knowledge.

## Step 3: Generate the Interview Plan

Create a markdown file at the user's specified output location as `<candidate-name>-plan.md`. If the directory doesn't exist, create it. Use the following structure. Every section is REQUIRED.

---

### Output Template

```markdown
# <Candidate Name> -- Design & Architecture Interview Plan

**Target Level:** L4 / L5 / L6
**Declared Spike:** <Data Engineering | Data Warehousing | AI/ML>
**Interview Duration:** 60 minutes
**Your Role:** Customer (Technical Lead / Stakeholder)
**Their Role:** Lead Architect / Advisor
**Date Prepared:** <today's date>

---

## Resume Analysis

### Background Summary
<2-3 sentence summary of the candidate's experience, domain, and claimed strengths>

### Key Claims to Validate
<Bulleted list of 3-5 resume claims that this interview should verify. For each, note WHY it matters and what depth looks like.>

| Resume Claim | Why It Matters | How to Probe |
|-------------|---------------|--------------|
| <claim from resume> | <what it signals about their level> | <specific question to test depth> |

### Domain Mapping
<How does the candidate's background map to the interview scenario? What domain knowledge should they naturally bring?>

---

## Custom Scenario

<IMPORTANT: Do NOT use the standard retail POS prompt verbatim. Instead, adapt the core architectural challenge (REST API ingestion at scale, medallion architecture, dual BI + ML serving) into the candidate's domain. The scenario must test all three evaluation dimensions identically but in a context where the candidate should naturally go deeper.>

### The Prompt (read this verbatim to the candidate)

> "<Custom scenario prompt tailored to the candidate's industry/domain. Must include:
> - A specific company/org type relevant to their experience
> - A data source that exposes data via REST API (or similar)
> - A volume target (~100M records/hour or equivalent scale)
> - A dual requirement: high-performance BI reporting AND downstream ML models
> - Enough ambiguity that they MUST ask discovery questions>"

### Why This Scenario Works
<Explain how this scenario maps to the standard retail prompt while leveraging the candidate's domain expertise. Note what the candidate should naturally know vs. what will stretch them.>

---

## Pre-Scripted "Customer" Answers

When the candidate asks discovery questions, use these answers. If they skip a category, that is a signal -- note it.

| # | Category | What to Listen For (their question) | Your Answer | Signal If They Skip |
|---|----------|-------------------------------------|-------------|-------------------|
| 1 | **Business Value** | Why are we doing this? What's the business goal? | <Specific business pain point with dollar impact or SLA target> | They may jump to solution mode without understanding the "why" |
| 2 | **Consumers** | Who uses this data? | <3 distinct consumer groups with different latency/format needs> | They may design a one-size-fits-all solution |
| 3 | **Source Systems** | How does the source API work? | <API specifics: rate limits, auth, schema, quirks, edge cases> | They may make dangerous assumptions about the source |
| 4 | **Data Volume** | How big is the data? | <Specific volume: records/hr, record size, total GB/hr, peak multiplier> | They may under-engineer the solution |
| 5 | **Data Quality** | Is the data clean? | <Specific quality issues: duplicates, late-arriving, NULLs, drift> | They may skip data quality entirely |
| 6 | **Compliance** | Any regulatory/security constraints? | <Specific compliance requirement: GDPR, data residency, encryption, PII> | They may miss a critical non-functional requirement |
| 7 | **Latency** | Batch or streaming? | <Mixed requirement: near-real-time for BI + batch for ML> | They may anchor on only batch or only streaming |
| 8 | **Budget/Ops** | Cost and team constraints? | <Small team, prefer managed services, reliability > cost> | They may over-engineer for a team that can't maintain it |

---

## The "Suboptimal Anchor"

Drop this early in discovery to test whether the candidate pushes back on bad technical decisions.

> "<A specific, plausible-sounding but architecturally flawed suggestion from 'the customer.' For example: a centralized gateway, a single-threaded approach, storing everything in one table, using a tool that doesn't fit the scale, etc.>"

**What you're testing:** <Explain the specific architectural flaw and why a strong candidate should challenge it.>

**If they accept it without pushback, probe:**
> "<Follow-up question that forces them to confront the flaw. Example: 'If that gateway goes down during peak hours, what happens to our data?'>"

**What good pushback looks like:**
- **L4:** Expresses concern, suggests it might be a bottleneck, offers a simple alternative
- **L5:** Clearly articulates why it fails at scale, proposes a distributed alternative, connects to the business SLA
- **L6:** Reframes the problem entirely, challenges the premise, proposes a platform-level pattern that prevents this class of issue

---

## Phase-by-Phase Breakdown

---

### Phase 1: Setup & Context (0-5 min)

**What to say:**

> "Thanks for joining, <Name>. Here's how this will work. Think of this as a discovery session with a customer. I'm your stakeholder -- I'm the <role, e.g., Director of Data Engineering> at <company type>. You are the lead architect we've brought in to design our data platform. I want you to treat me like a customer: ask me questions, challenge my assumptions, and walk me through your thinking out loud. You can use Google Slides, a whiteboard, or whatever tool you're comfortable with -- please share your screen. The goal isn't a perfect diagram, it's a visual aid for our discussion. Ready?"

Then read the **Custom Scenario Prompt** above.

**Watch for:** Do they immediately start drawing, or do they pause and ask questions first? First instinct matters.

---

### Phase 2: Discovery & Problem Framing (5-20 min)

**Time guard:** If discovery exceeds 20 minutes, say: *"This is really thorough -- can you start sketching out what you're thinking so far? We can fill in gaps as we go."*

#### What Good Looks Like by Level

| Dimension | L4 | L5 | L6 |
|-----------|----|----|-----|
| **Questions asked** | Asks about data volume, API basics, who consumes the data | Proactively surfaces constraints (rate limits, schema inconsistency, edge cases). Connects requirements to latency/SLA implications | Challenges the entire premise. Asks about organizational readiness, existing systems, and whether the stated approach is the right one |
| **Problem framing** | Frames basic assumptions before designing | Aligns the design to the three distinct consumer groups with different needs | Defines the problem scope, identifies what's out of scope, and establishes success criteria |
| **Suboptimal anchor** | May accept it or weakly question it | Pushes back with specific technical reasoning | Reframes the problem to prevent the class of issue entirely |

#### Probing Questions (if they're not going deep enough)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "What other questions would you want to ask before you start designing?" | Completeness of discovery thinking |
| 2 | "You haven't asked about [compliance / data quality / the edge-case source] -- is that not relevant?" | Whether they consider non-functional requirements |
| 3 | "What assumptions are you making right now?" | Self-awareness about gaps in their understanding |
| 4 | "If we stick with my original plan [the suboptimal anchor], what's the biggest risk you see?" | First-principles thinking and pushback ability |
| 5 | "Who else should we be thinking about as consumers of this data?" | Whether they think beyond the immediate ask |

---

### Phase 3: Core Architecture Design (20-40 min)

**What you MUST see on the whiteboard:**

1. **Ingestion layer** -- How does data get from the source system(s) into the Lakehouse?
2. **Medallion architecture** -- Bronze (raw), Silver (cleaned/conformed), Gold (aggregated/served) -- or equivalent layered approach
3. **Separation of concerns** -- Ingestion logic separate from transformation logic
4. **Dual serving pattern** -- Low-latency path for BI vs. batch path for ML

**The Guardrail:** If the candidate goes too wide (designing UI, building ML models), pull them back:
> "Let's focus on the data platform and pipeline layers. Assume the dashboard and ML teams will consume from your serving layer."

#### Core Architecture Probing Questions

| # | Question | What You're Testing | L4 Answer | L5 Answer | L6 Answer |
|---|----------|-------------------|-----------|-----------|-----------|
| 1 | "Walk me through what happens to a record from the moment it leaves the source system to the moment a user sees it on a dashboard." | End-to-end flow, separation of concerns | Describes a linear flow: API -> storage -> transform -> dashboard. May miss intermediate layers. | Describes a layered flow with clear Bronze/Silver/Gold stages. Explains why raw data is preserved. Addresses the latency requirement for the BI path specifically. | Describes a platform-level flow with pluggable components. Discusses how the pattern extends to new sources. Addresses operational concerns (monitoring, alerting, SLA guarantees). |
| 2 | "How do you handle schema differences across sources?" | Schema evolution, data conformance | "We normalize during ingestion." | "Ingest raw into Bronze as-is, apply source-specific mapping in Bronze-to-Silver. New sources don't break existing pipelines." | "Establish a schema registry and contract-based ingestion pattern. Source teams own their schemas; the platform validates and conforms automatically." |
| 3 | "What happens when the source goes offline for 4 hours and then comes back with a backlog?" | Fault tolerance, late-arriving data | Mentions retry logic or queuing. | Discusses idempotent ingestion, watermarking for late-arriving data, merge/upsert patterns. | Designs a backpressure system with dead-letter queues, replay capabilities, and SLA-aware prioritization. |
| 4 | "Why did you choose [Component X] over [Alternative Y]?" | First-principles thinking | Gives a reasonable but shallow answer. | Articulates specific trade-offs (cost, latency, complexity, team size) and connects to the customer's constraints. | Frames the decision as a platform choice with long-term implications. Discusses migration paths if requirements change. |
| 5 | "If data volumes triple overnight, where does this break first?" | Scalability awareness | Identifies one bottleneck (e.g., API rate limits). | Identifies multiple bottlenecks and proposes autoscaling or backpressure mechanisms. | Designs the architecture to be elastic from the start. Discusses cost-proportional scaling and capacity planning. |
| 6 | "How do you ensure data quality between layers?" | Data quality thinking | Mentions basic validation checks. | Designs quality gates between Bronze→Silver→Gold with specific checks (schema validation, null checks, dedup). | Implements a data quality framework with automated testing, quarantine patterns, and quality SLAs per layer. |
| 7 | "How would you orchestrate this pipeline?" | Operational thinking | Mentions a scheduler (cron, Airflow). | Designs event-driven orchestration with dependency management, retry policies, and alerting. | Establishes an orchestration framework with SLA monitoring, dynamic resource allocation, and self-healing pipelines. |

---

### Phase 4: Technical Spike -- Deep Dive (40-55 min)

This is the primary evaluation zone. Use the **three-layer probing method** below.

**Transition into the spike:**
> "<Tailored transition statement that connects the candidate's architecture to their declared spike. For example: 'Great, now let's go deeper on the [Data Engineering / Data Warehousing / ML] side...' followed by a specific, challenging sub-problem.>"

---

#### IF SPIKE = Data Engineering / ELT

##### Layer 1: Implementation (5 min)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "How specifically are you parallelizing these API calls to handle the volume? Walk me through the mechanics." | Practical implementation knowledge -- Spark, async patterns, connection pooling |
| 2 | "How do you handle incremental loads vs. full refreshes? What's your change detection strategy?" | CDC patterns, merge/upsert, watermarks |
| 3 | "Show me how your Bronze-to-Silver transformation handles the schema differences across sources." | Data conformance, mapping logic, testability |
| 4 | "What does your error handling look like? What happens when a single record fails?" | Dead-letter patterns, circuit breakers, observability |

**L4 bar:** Can describe a working ingestion pipeline with basic parallelism and error handling.
**L5 bar:** Designs modular, testable pipelines with parameterization, idempotency, and monitoring.
**L6 bar:** Establishes engineering standards: CI/CD for pipelines, automated testing, observability dashboards, and operational runbooks.

##### Layer 2: Stress Test (5 min)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "Volume triples overnight. Where does your pipeline break first, and how do you fix it?" | Scalability thinking under pressure |
| 2 | "A schema change is deployed on the source system without warning. What happens to your pipeline?" | Schema evolution, defensive programming |
| 3 | "Your Silver-to-Gold job has been running for 6 hours. How do you debug it?" | Operational maturity, observability |

##### Layer 3: Trade-offs & First Principles (5 min)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "Why choose a distributed approach here instead of a monolithic one? What are the cost vs. latency trade-offs?" | Architectural reasoning |
| 2 | "Batch vs. micro-batch vs. streaming -- where do you draw the line for this use case?" | Pragmatic decision-making |
| 3 | "If you had to rebuild this pipeline from scratch with half the team, what would you cut?" | Prioritization and pragmatism |

---

#### IF SPIKE = Data Warehousing / Analytics

##### Layer 1: Implementation (5 min)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "Walk me through your Gold layer design. How are you modeling data for the BI consumers?" | Dimensional modeling, star/snowflake schemas, denormalization strategy |
| 2 | "How do you optimize query performance for 5,000+ concurrent BI users?" | Partitioning, clustering, caching, materialized views |
| 3 | "How does the Gold layer differ for BI consumers vs. ML consumers? Can one design serve both?" | Understanding that aggregate BI tables and record-level ML features have different needs |
| 4 | "How do you handle slowly changing dimensions in your data model?" | SCD patterns (Type 1/2/3), temporal modeling |

**L4 bar:** Models data for specific use cases (simple aggregate tables for BI reporting). Understands star schema basics.
**L5 bar:** Balances flexibility and performance, optimizes for multiple workload patterns and consumers. Designs for query performance at scale.
**L6 bar:** Defines enterprise-wide modeling patterns and serving strategies. Designs for self-service analytics.

##### Layer 2: Stress Test (5 min)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "A VP says dashboards are slow during month-end close. 10x more users, same data. How do you fix it?" | Concurrency scaling, query optimization |
| 2 | "A new business unit wants to add 50 new dimensions to the Gold layer. What's your approach?" | Schema evolution at the warehouse level |
| 3 | "Data scientists complain the Gold layer aggregations lose the granularity they need. How do you serve both?" | Multi-resolution serving, balancing aggregation levels |

##### Layer 3: Trade-offs & First Principles (5 min)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "Normalized vs. denormalized for your serving layer -- what did you choose and why?" | Data modeling trade-offs |
| 2 | "Materialized views vs. pre-aggregated tables -- when do you use each?" | Performance optimization reasoning |
| 3 | "How do you govern data access when 50 teams want different slices of the same Gold table?" | Governance and access patterns at scale |

---

#### IF SPIKE = AI/ML

##### Layer 1: Implementation (5 min)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "What does the feature engineering pipeline look like? Where do features come from, and where do they live?" | Feature stores/tables, separation of feature engineering from model training |
| 2 | "The model needs features from the last N hours of data. How do you compute and serve those features?" | Windowed aggregations, point-in-time correctness, feature freshness |
| 3 | "How do you ensure training data doesn't leak future information?" | Point-in-time correctness -- critical and often missed |
| 4 | "Walk me through model training to deployment. What does the lifecycle look like?" | MLOps maturity: experiment tracking, model registry, deployment patterns |

**L4 bar:** Creates clean datasets for training. Understands feature/label alignment. Knows MLflow basics.
**L5 bar:** Proposes a feature store with point-in-time correctness. Designs automated retraining with drift monitoring. Explains online vs. offline feature serving.
**L6 bar:** Architects ML platform with integrated governance, scalable model serving, and enterprise-wide feature sharing.

##### Layer 2: Stress Test (5 min)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "We're scaling from 10 models to 10,000 models (one per entity). How does your architecture handle that?" | Scale thinking for ML systems |
| 2 | "Retraining is daily. Product wants hourly. What breaks?" | Compute costs, feature freshness, retraining pipeline dependencies |
| 3 | "A data source starts drifting -- still sending data but it's subtly off. How does your system detect that before it corrupts the model?" | Data drift detection, monitoring, quality gates |

##### Layer 3: Trade-offs & First Principles (5 min)

| # | Question | What You're Testing |
|---|----------|-------------------|
| 1 | "For this prediction task, why would you choose [model type A] over [model type B]?" | Trade-off reasoning, not just defaulting to what they know |
| 2 | "Should the feature store serve both real-time inference and batch training? Or separate stores?" | Online vs. offline feature store trade-offs |
| 3 | "This model's prediction drives a high-stakes business decision. How do you handle model confidence and human-in-the-loop?" | Business-aware ML design |

---

### Phase 5: The "Hard Question" & Close (55-60 min)

**The cross-domain probe:** Ask something outside their declared spike to test breadth.

> "<If their spike is Data Engineering, ask an ML question. If ML, ask a Data Engineering question. If DW, ask about either. The question should connect to their architecture -- e.g., 'What happens if corrupted data makes it through your pipeline and an ML model retrains on it?' or 'How would you add a real-time feature serving layer to this architecture?'>"

**What you're testing:** Comfort with adjacent domains, intellectual humility, problem-solving approach for unfamiliar territory.

**What good looks like:**
- **L4:** Gives a reasonable attempt. May say "I'm not sure" but shows willingness to reason through it.
- **L5:** Provides a directionally correct answer, acknowledges gaps, explains how they'd find the answer.
- **L6:** Gives a strong answer even outside their spike. Connects it back to the architecture. May proactively identify this as a risk area that needs a specialist.

**Close:** "Any questions for me about the role or the team?"
- Listen for quality of questions. Good candidates ask about the team, tech stack, or business challenges. Great candidates ask questions that show they've been thinking about the problem beyond what was asked.

---

## Evaluation Scorecard

### Pass/Fail Signals Checklist

#### Pass Signals
- [ ] Asked 5+ meaningful discovery questions before whiteboarding
- [ ] Pushed back on the suboptimal anchor with technical reasoning
- [ ] Designed a clear layered architecture (medallion or equivalent)
- [ ] Separated ingestion from transformation from serving
- [ ] Addressed both the BI and ML serving patterns
- [ ] Communicated clearly -- could explain decisions to a non-technical stakeholder
- [ ] Articulated trade-offs, not just solutions
- [ ] Demonstrated depth in their declared spike area
- [ ] Showed comfort with ambiguity -- asked for clarification rather than assuming

#### Red Flags (potential fail signals)
- [ ] Jumped straight to whiteboarding without asking questions
- [ ] Accepted the suboptimal anchor without questioning it
- [ ] Designed a monolithic "black box" solution with no separation of concerns
- [ ] Proposed flat data model with no layering
- [ ] No mention of error handling, late-arriving data, or fault tolerance
- [ ] Used heavy jargon without being able to explain simply
- [ ] Could not articulate why they chose one approach over another
- [ ] Rigid thinking -- anchored on batch-only or streaming-only
- [ ] Treated their non-spike area as completely unknown (no breadth)
- [ ] Could not handle the stress-test questions in their spike area

### Level Assessment Matrix

| Dimension | L4 | L5 | L6 |
|-----------|----|----|-----|
| **Discovery** | Frames the problem with relevant assumptions. Asks about volume, API basics, data consumers. | Proactively surfaces constraints (rate limits, schema inconsistency). Connects requirements to latency/SLA implications. Challenges the suboptimal anchor. | Shapes the problem definition. Challenges the REST approach if better alternatives exist. Establishes success criteria before designing. |
| **Core Arch** | Designs a coherent source → ingest → transform → serve flow. Uses layered storage (Bronze/Silver/Gold or equivalent). | Designs a distributed architecture that scales despite workload diversity. Separates real-time and batch paths. Considers operational constraints (small team, managed services). | Architects a platform-level ingestion framework. Anticipates risks and evolution. Designs for extensibility and multi-tenancy. |
| **Spike Depth** | Can build a working solution using standard patterns. May need prompting. | Builds production-ready systems that survive real-world constraints. Good bi-directional conversation. | Shapes the problem and defines standards for the organization. Leads the conversation. |
| **Communication** | Answers questions clearly. Can walk through the diagram. | Leads a bi-directional conversation. Proactively explains trade-offs. | Translates technical decisions into business terms. Drives the discussion proactively. |

### Overall Recommendation

| Rating | Criteria |
|--------|----------|
| **Strong Hire at L<N>** | Exceeded expectations in all three areas. Demonstrated depth beyond target level in spike area. |
| **Hire at L<N>** | Met expectations in all three areas. Solid performance with no red flags. |
| **Lean Hire at L<N>** | Met expectations in 2 of 3 areas. Minor concerns but overall positive signal. |
| **Lean No Hire** | Met expectations in only 1 area. Significant gaps in discovery or core architecture. |
| **No Hire** | Did not meet expectations. Multiple red flags. |

---

## Notes Section

<Leave space for:>
- Resume validation notes (did their architecture confirm or contradict their resume claims?)
- Specific quotes or moments to include in the debrief
- Databricks product knowledge observations (nice-to-have, not required)
- Candidate's questions at the end and what they signal
```

## Important Reminders

- **Do NOT use the retail POS prompt verbatim.** Always customize the scenario to the candidate's domain. The scenario must test the same architectural patterns (REST API ingestion, scale, medallion, dual BI+ML) but in the candidate's territory.
- **The suboptimal anchor is critical.** Every plan must include one. It tests first-principles thinking and pushback ability -- two of the most important signals.
- **Tailor the spike section.** Only include the deep-dive questions for the candidate's declared spike. Do not include all three spike sections in the output.
- **Include L4/L5/L6 expected answers** for the target level AND one level above/below so the interviewer can calibrate.
- **Pre-scripted customer answers must be specific.** Use concrete numbers, dollar amounts, team sizes, and SLA targets. Vague answers make it harder to evaluate the candidate.
- **The output file must be a complete, standalone document** that the interviewer can print and use without any other reference material.
