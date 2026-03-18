# Vibe User Guide

A practical guide to using vibe - the Field Engineering plugins for Claude Code. This covers Claude Code basics, how skills work, what each skill does, and common workflows.

---

## Table of Contents

- [Claude Code Basics](#claude-code-basics)
- [Getting Started with Vibe](#getting-started-with-vibe)
- [Skills Reference](#skills-reference)
  - [Databricks Platform](#1-databricks-platform-databricks-tools)
  - [Google Workspace](#2-google-workspace-google-tools)
  - [Salesforce & CRM](#3-salesforce--crm-fe-salesforce-tools)
  - [Internal Tools](#4-internal-tools-fe-internal-tools)
  - [Workflows](#5-workflows-workflows)
  - [JIRA](#6-jira-jira-tools)
  - [Expenses](#7-expenses-fe-file-expenses)
  - [Diagrams](#8-diagrams-specialized-agents)
  - [Setup & Maintenance](#9-setup--maintenance-vibe-setup)
- [Common Workflows](#common-workflows)
- [Tips & Troubleshooting](#tips--troubleshooting)

---

## Claude Code Basics

Claude Code is a CLI-based AI coding assistant that runs in your terminal. Vibe extends it with Field Engineering-specific skills and integrations.

### Starting a Session

```bash
vibe agent              # Start a vibe session (recommended)
vibe agent /path/to/dir # Start in a specific directory
```

### Key Concepts

| Concept | What It Is |
|---------|------------|
| **Skill** | A specialized prompt that teaches Claude how to perform a specific task (e.g., querying Salesforce, creating a Google Doc). Skills are the main building blocks of vibe. |
| **Agent** | A sub-process that runs in parallel to handle complex sub-tasks. Agents are spawned by skills when work can be delegated. |
| **Plugin** | A collection of related skills and agents bundled together (e.g., `google-tools` contains all Google Workspace skills). |
| **MCP Server** | An external service connection (Slack, JIRA, Confluence, Glean) that gives Claude access to APIs. |

### How to Invoke Skills

There are two ways to trigger a skill:

1. **Natural language (auto-routing)** - Just describe what you need. Claude automatically picks the right skill.
   ```
   "Create a workspace for my demo"
   "Who reports to Jane Smith?"
   "File my expenses"
   ```

2. **Slash command (direct)** - Type `/skill-name` to explicitly invoke a skill.
   ```
   /databricks-authentication
   /google-docs
   /performance-tuning
   ```

Auto-routing works well for most cases. Use slash commands when you want to be explicit, or when the natural language isn't triggering the skill you expect.

### Useful Built-in Commands

| Command | What It Does |
|---------|-------------|
| `/help` | Show available commands |
| `/clear` | Clear conversation history |
| `/compact` | Compress conversation to free up context window |
| `/status` | Show current session info |
| `/cost` | Show token usage and cost |
| `Escape` | Interrupt Claude mid-response |
| `Ctrl+C` | Cancel current operation |

### Permission Prompts

When Claude needs to run a command or access a resource for the first time, you'll see a permission prompt. Options:
- **Allow once** - Approve this specific action
- **Allow always** - Auto-approve this type of action going forward
- **Deny** - Block the action

Most vibe skills have pre-configured permissions, so you'll rarely see prompts for standard operations.

### Tips for Effective Use

- **Be specific** - "Query logfood for Block's model serving quota" works better than "get me some data"
- **Provide context** - Include account names, URLs, ticket numbers when relevant
- **Use `/compact`** when conversations get long - it frees up context without losing important state
- **Interrupt with Escape** if Claude is going down the wrong path, then redirect
- **Authentication first** - Many skills need auth. Run `/configure-vibe` on first use, and authenticate before using platform-specific skills

---

## Getting Started with Vibe

### Installation

See the [main README](../README.md) for installation steps. In short:

```bash
brew install gh && gh auth login --web --hostname github.com --git-protocol ssh --skip-ssh-key && gh release download latest --repo databricks-field-eng/vibe --pattern 'install_cli.sh' -O - | bash
vibe install
```

### First-Time Setup

Start a session and run the configuration skill:

```bash
vibe agent
```

Then tell Claude:

```
Configure vibe
```

This runs `/configure-vibe`, which installs CLI dependencies (Databricks CLI, Salesforce CLI, gcloud, AWS CLI, etc.) and sets up credential paths.

### Verify MCP Connections

After setup, verify your MCP connections (Slack, Glean, JIRA, Confluence):

```
Validate my MCP access
```

This runs `/validate-mcp-access` and reports which connections are active. Follow the login instructions for any that aren't.

### How Skill Routing Works

When you type a message, Claude evaluates it against all available skill descriptions and decides whether to invoke a skill. The routing is based on:

- **Skill descriptions** - Each skill has a description that acts as a routing hint
- **Keywords** - Certain keywords strongly trigger specific skills (e.g., "ES ticket" triggers `/jira-actions`)
- **Context** - Your conversation history helps Claude pick the right skill

If a skill doesn't auto-trigger, you can always invoke it directly with `/skill-name`.

---

## Skills Reference

### 1. Databricks Platform (databricks-tools)

Skills for authentication, workspace management, querying, apps, dashboards, and demos.

---

#### databricks-authentication

**What it does:** Authenticates the Databricks CLI against a specific workspace environment. This is a prerequisite for most Databricks skills.

**Example prompts:**
- "Authenticate with Databricks"
- "Log in to my FE-VM workspace"
- "Set up my logfood profile"

**Slash command:** `/databricks-authentication`

**Prerequisites:** None

**Also invokes:** None - this is a foundational skill that other skills depend on

---

#### databricks-fe-vm-workspace-deployment

**What it does:** Deploys and manages Databricks workspaces using the FE Vending Machine. Automatically handles authentication, caches environment info, and supports serverless and classic workspace types.

**Example prompts:**
- "Create a Databricks classic workspace"
- "Deploy a serverless workspace for my demo"
- "Which workspace did I create earlier?"

**Slash command:** `/databricks-fe-vm-workspace-deployment`

**Prerequisites:** None (handles its own auth)

**Also invokes:** `databricks-authentication` (for CLI auth after deployment)

---

#### databricks-oneenv-workspace-deployment

**What it does:** Creates Databricks workspaces in the One-Env AWS account for demos requiring custom AWS integrations like custom IAM roles, cross-account S3 access, custom VPCs, or PrivateLink. Note: workspaces are auto-cleaned after 2 weeks.

**Example prompts:**
- "I need to create a workspace that has custom IAM roles for S3 cross-account access"
- "Create a workspace in the one-env account for testing AWS Glue integration"
- "I need a classic Databricks workspace with a custom VPC for PrivateLink demo"

**Slash command:** `/databricks-oneenv-workspace-deployment`

**Prerequisites:** `/aws-authentication`, `/databricks-authentication`

**Also invokes:** AWS CLI, Databricks CLI

---

#### databricks-workspace-files

**What it does:** Explores and retrieves files from Databricks workspaces using the CLI. Lists directories, exports notebooks (.py, .sql, .ipynb), and pulls code into context for review.

**Example prompts:**
- "List the notebooks in my Databricks workspace under /Users/me@company.com"
- "Pull the ETL notebook from my Databricks workspace into context"
- "Find all the SQL notebooks in /Shared"

**Slash command:** `/databricks-workspace-files`

**Prerequisites:** `/databricks-authentication`

**Also invokes:** None

---

#### databricks-lineage

**What it does:** Explores Unity Catalog data lineage at table and column levels. Traces upstream sources, downstream consumers, and column-level dependencies.

**Example prompts:**
- "Show me the upstream lineage for main.sales.orders table"
- "Trace the column lineage for the total_amount column in our orders table"
- "What downstream tables will be affected if I change the customer table schema?"

**Slash command:** `/databricks-lineage`

**Prerequisites:** `/databricks-authentication`, Unity Catalog enabled

**Also invokes:** `databricks-workspace-files` (for pulling related notebook code)

---

#### databricks-query

**What it does:** Executes SQL queries against Databricks datasets. Handles warehouse selection, query execution via the Databricks SQL Statements API, and pretty-prints results as formatted tables.

**Example prompts:**
- "Run a SQL query on my Databricks workspace"
- "Query the orders table in my catalog"

**Slash command:** `/databricks-query`

**Prerequisites:** `/databricks-authentication`, `/databricks-warehouse-selector`

**Also invokes:** `/databricks-warehouse-selector` (to pick a warehouse)

---

#### databricks-apps

**What it does:** Builds and deploys full-stack Databricks Apps with Lakebase database integration, Foundation Model API for AI features, and React/FastAPI architecture. Covers project setup, local testing, and deployment.

**Example prompts:**
- "Build a Databricks App with a React frontend and FastAPI backend"
- "Deploy my app to the workspace"
- "Create an app that uses Lakebase and Foundation Models"

**Slash command:** `/databricks-apps`

**Prerequisites:** `/databricks-fe-vm-workspace-deployment` (serverless workspace), `/databricks-authentication`

**Also invokes:** `databricks-apps-developer` agent, `web-devloop-tester` agent, `/databricks-resource-deployment`

---

#### databricks-lakebase

**What it does:** Creates and manages Lakebase (managed Postgres) databases on Databricks. Covers instance creation, connection methods (psql, OAuth, Python, SQLAlchemy), schema management, Unity Catalog integration, and the Data API.

**Example prompts:**
- "Create a Lakebase Postgres database"
- "Connect to my Lakebase instance with psql"
- "Set up a Lakebase database for my app"

**Slash command:** `/databricks-lakebase`

**Prerequisites:** `/databricks-fe-vm-workspace-deployment` (serverless workspace), `/databricks-authentication`

**Also invokes:** `/databricks-apps` (for app integration), `/databricks-resource-deployment`

---

#### databricks-lakeview-dashboard

**What it does:** Programmatically creates and manages Lakeview (AI/BI) dashboards using the Dashboard API. Handles JSON dashboard definitions, visualization types, layouts, and deployment.

**Example prompts:**
- "Create a Lakeview dashboard for my sales data"
- "Build a dashboard with bar charts and counter tiles"

**Slash command:** `/databricks-lakeview-dashboard`

**Prerequisites:** `/databricks-authentication`, SQL warehouse available

**Also invokes:** None

---

#### databricks-lakeview-dashboard-analyzer

**What it does:** Analyzes existing Lakeview dashboards through browser automation. Navigates dashboards, takes screenshots, extracts data from tables and charts, and interprets visualizations.

**Example prompts:**
- "Can you look at this Lakeview dashboard and summarize what it says?"
- "Analyze the dashboard at this URL and extract the key metrics"

**Slash command:** `/databricks-lakeview-dashboard-analyzer`

**Prerequisites:** Chrome DevTools MCP

**Also invokes:** Chrome DevTools for browser automation

---

#### databricks-demo

**What it does:** Creates, deploys, and runs demos relating to or integrating with Databricks. Orchestrates workspace provisioning, data generation, code writing, and deployment.

**Example prompts:**
- "Create a demo showing Delta Lake change data capture"
- "Build a streaming demo for my customer"

**Slash command:** `/databricks-demo`

**Prerequisites:** `/databricks-authentication`

**Also invokes:** `/databricks-fe-vm-workspace-deployment`, `/databricks-resource-deployment`, `/fe-snowflake`, `databricks-apps-developer` agent

---

#### databricks-resource-deployment

**What it does:** Deploys resources (notebooks, jobs, clusters, warehouses, apps, Lakebase instances, catalogs, schemas) into an existing Databricks workspace.

**Example prompts:**
- "Deploy a job to my workspace"
- "Create a cluster configuration for my demo"

**Slash command:** `/databricks-resource-deployment`

**Prerequisites:** `/databricks-authentication`, an existing workspace

**Also invokes:** `/databricks-fe-vm-workspace-deployment` (for new workspace if needed), `/fe-snowflake` (for Snowflake integrations)

---

#### databricks-warehouse-selector

**What it does:** Selects the best SQL warehouse to execute queries on. Lists available warehouses and picks the optimal one, or creates a new warehouse if none exist.

**Example prompts:**
- "Which warehouse should I use for this query?"
- (Typically invoked by other skills automatically)

**Slash command:** `/databricks-warehouse-selector`

**Prerequisites:** `/databricks-authentication`

**Also invokes:** None

---

#### fe-snowflake

**What it does:** Sets up and manages Snowflake trial accounts for FE demos and testing. Automates the full flow: burner Gmail creation, Snowflake trial signup via browser automation, CLI installation, and credential storage.

**Example prompts:**
- "Set up a Snowflake trial account for my demo"
- "I need a Snowflake instance for testing Iceberg interop"

**Slash command:** `/fe-snowflake`

**Prerequisites:** Chrome DevTools MCP

**Also invokes:** Chrome DevTools for browser automation

---

### 2. Google Workspace (google-tools)

Skills for Google Docs, Sheets, Slides, Calendar, Gmail, Tasks, and Drive.

---

#### google-auth

**What it does:** Unified authentication for all Google Workspace APIs (Docs, Sheets, Slides, Drive, Calendar, Gmail, Tasks). Run this before using any other Google skill.

**Example prompts:**
- "Authenticate with Google"
- (Typically invoked automatically by other Google skills)

**Slash command:** `/google-auth`

**Prerequisites:** None

**Also invokes:** None - this is the foundational auth skill for all Google skills

---

#### google-docs

**What it does:** Opens, reads, creates, edits, and manages Google Docs. Handles document content extraction, markdown-to-Google-Docs conversion, tables, hyperlinks, images, and @mentions. Also manages Drive files and Slides.

**Example prompts:**
- "Read this document and summarize what's in it: https://docs.google.com/document/d/..."
- "Create a Google Doc with my meeting notes"
- "Convert this markdown to a Google Doc"

**Slash command:** `/google-docs`

**Prerequisites:** `/google-auth`

**Also invokes:** `google-drive` agent (for complex operations)

---

#### google-sheets

**What it does:** Creates and manages Google Sheets with cell operations, formulas, charts, colors, conditional formatting, multiple sheets, and batch updates.

**Example prompts:**
- "Create a spreadsheet tracking our account metrics"
- "Add a chart to my Google Sheet"

**Slash command:** `/google-sheets`

**Prerequisites:** `/google-auth`

**Also invokes:** None

---

#### google-slides

**What it does:** Creates and manages Google Slides presentations with Databricks corporate templates, tables, charts, images, and professional layouts.

**Example prompts:**
- "Create a presentation for my customer demo"
- "Build slides using the Databricks template"

**Slash command:** `/google-slides`

**Prerequisites:** `/google-auth`

**Also invokes:** `/google-sheets` (for chart embedding)

---

#### google-calendar

**What it does:** Creates, modifies, and manages Google Calendar events. Finds available meeting times across attendees using FreeBusy queries.

**Example prompts:**
- "Find a time when brandon@databricks.com and sarah@databricks.com are both available next week for a 30-minute meeting"
- "Schedule a team sync for next Tuesday"
- "When are all of these people free: alice@company.com, bob@company.com"

**Slash command:** `/google-calendar`

**Prerequisites:** `/google-auth`

**Also invokes:** None

---

#### gmail

**What it does:** Searches, reads, composes, organizes Gmail emails, and manages filters/label rules. Supports rich HTML formatting, attachments, forwarding, and replying.

**Example prompts:**
- "Create a Gmail filter that labels emails from my boss as important"
- "Show me all my Gmail filters and label rules"
- "Send an email to the team about the meeting"

**Slash command:** `/gmail`

**Prerequisites:** `/google-auth`

**Also invokes:** None

---

#### google-tasks

**What it does:** Manages Google Tasks - create tasks, manage task lists, organize subtasks, set due dates, and track completion.

**Example prompts:**
- "Add a task to remember to buy groceries tomorrow"
- "Show me all my tasks in Google Tasks"
- "Mark my grocery shopping task as done"

**Slash command:** `/google-tasks`

**Prerequisites:** `/google-auth`

**Also invokes:** None

---

### 3. Salesforce & CRM (fe-salesforce-tools)

Skills for Salesforce authentication, data operations, and UCO management.

---

#### salesforce-authentication

**What it does:** Authenticates with Salesforce CRM using the Salesforce CLI. Required before any Salesforce data operations.

**Example prompts:**
- "Log in to Salesforce"
- "Authenticate with Salesforce"

**Slash command:** `/salesforce-authentication`

**Prerequisites:** Salesforce CLI installed (via `/configure-vibe`)

**Also invokes:** None

---

#### salesforce-actions

**What it does:** Read and update Salesforce data including Use Case Objects (UCOs), Accounts, Opportunities, Blockers, ASQs/Specialist Requests, and Preview Feature Requests. Intelligently chooses between Salesforce API (for writes) and Logfood (for reads/aggregations).

**Example prompts:**
- "Which use cases do we have at Instacart in stage 6"
- "Show me the opportunities for Block"
- "Update the UCO status for this account"

**Slash command:** `/salesforce-actions`

**Prerequisites:** `/salesforce-authentication`

**Also invokes:** `/logfood-querier` (for read-heavy queries), `/internal-jargon` (for acronym lookup)

---

#### uco-updates

**What it does:** Updates Use Case Objects (UCOs) with weekly status updates following AMER Emerging guidelines. Covers required fields (Next Steps, Target Dates, Health Status, Implementation Notes) and stage progression (U2-U6).

**Example prompts:**
- "Update my UCOs for this week"
- "Do my weekly UCO updates"
- "What UCOs are assigned to me?"

**Slash command:** `/uco-updates`

**Prerequisites:** `/salesforce-authentication`

**Also invokes:** `field-data-analyst` agent

---

### 4. Internal Tools (fe-internal-tools)

Skills for Databricks internal analytics, org lookup, and knowledge bases.

---

#### logfood-querier

**What it does:** Queries Databricks internal data and metrics (consumption, feature usage, account info) from the Logfood workspace. Critical for data analysis tasks involving GTM data.

**Example prompts:**
- "Query logfood for how much quota Block has used for model serving"
- "What's the consumption breakdown for this account?"

**Slash command:** `/logfood-querier`

**Prerequisites:** `/databricks-authentication` (logfood profile)

**Also invokes:** `/salesforce-actions` (loads proactively), `/databricks-warehouse-selector`, `/internal-jargon`

---

#### databricks-org

**What it does:** Looks up organizational information for Databricks employees - BU, segment, vertical, manager hierarchy, direct/indirect reports, and territory accounts.

**Example prompts:**
- "Who reports to Maneesh Bhide?"
- "Who is Brandon Kvarda's manager?"
- "What business unit is John Pelz in?"
- "Show me the full organization chart for Kyle Pistor"
- "What accounts are in Samir Gupta's territory?"

**Slash command:** `/databricks-org`

**Prerequisites:** `/databricks-authentication` (logfood profile)

**Also invokes:** `field-data-analyst` agent

---

#### genie-rooms

**What it does:** Queries Databricks Genie Rooms using natural language. Supports 12+ pre-configured internal Genie Rooms (Global, Emerging, CME, Retail, HLS, FINS, MFG, LATAM, CAN) or custom room IDs.

**Example prompts:**
- "Query this genie room 01ef336cd40b11f2b4931415636694eb for top accounts by revenue"
- "What are the top accounts in the Emerging segment?"
- "Query the retail genie room for Q4 consumption"

**Slash command:** `/genie-rooms`

**Prerequisites:** `/databricks-authentication`, Genie Room access (go/gtm_genie_access), VPN

**Also invokes:** `/logfood-querier`, `/databricks-query`

---

#### internal-jargon

**What it does:** Looks up Databricks internal jargon, acronyms, and terminology from the company glossary (go/glossary). Context-efficient - only fetches specific terms when needed.

**Example prompts:**
- "What does DNB stand for in Databricks?"
- "What does EDNB mean?"
- "How many UCOs does DNB have?" (may co-trigger with salesforce-actions)

**Slash command:** `/internal-jargon`

**Prerequisites:** Confluence MCP (preferred) or local glossary fallback

**Also invokes:** None

---

#### uco-consumption-analysis

**What it does:** Analyzes UCO portfolios against actual product consumption data. Validates UCO stages, identifies consumption not captured in UCOs, creates missing UCOs, and generates progression plans following the U1-U6 framework.

**Example prompts:**
- "Analyze the UCO portfolio for Block against their consumption data"
- "Are our UCO stages accurate for this account?"
- "Identify missing UCOs backed by consumption"

**Slash command:** `/uco-consumption-analysis`

**Prerequisites:** `/salesforce-authentication`, `/databricks-authentication`

**Also invokes:** `/logfood-querier`, `/internal-jargon`, `/salesforce-actions`, `/uco-updates`, `/google-docs`

---

#### aws-authentication

**What it does:** Authenticates with AWS using the AWS CLI. Required for One-Env workspace deployments and AWS resource integrations.

**Example prompts:**
- "Authenticate with AWS"
- "Log in to my AWS sandbox account"

**Slash command:** `/aws-authentication`

**Prerequisites:** AWS CLI installed (via `/configure-vibe`)

**Also invokes:** None

---

### 5. Workflows (workflows)

Complex multi-step workflows for customer engagements, documentation, research, and testing.

---

#### product-question-research

**What it does:** Researches and answers Databricks product questions for customers. Searches public docs, Glean, and Slack to provide an answer with confidence rating (1-10 scale). Creates a formatted Google Doc with the answer, references, and relevant Slack channels.

**Example prompts:**
- "Do materialized views on foreign iceberg tables support incrementalization?"
- "Research whether Delta Sharing supports streaming tables and create a doc with the answer"
- "I need to answer a product question for a customer about Unity Catalog feature X"

**Slash command:** `/product-question-research`

**Prerequisites:** Glean MCP, Slack MCP, `/google-docs`

**Also invokes:** `product-question-researcher` agent, `/google-docs`

---

#### fe-answer-customer-questions

**What it does:** Drafts responses to customer questions from Slack, email, or Google Docs. Researches answers using public docs, Glean, and Slack, then creates a combined Q&A document with confidence ratings and references.

**Example prompts:**
- "Draft responses to the customer questions in this meeting notes doc: https://docs.google.com/document/d/ABC123"
- "I have a bunch of unanswered customer questions from our sync meeting"
- "Help me answer these customer questions from our Slack channel"

**Slash command:** `/fe-answer-customer-questions`

**Prerequisites:** Glean MCP, Slack MCP, `/google-docs`

**Also invokes:** `customer-question-answerer` agent, `/google-docs`

---

#### fe-customer-courses

**What it does:** Generates customer-specific training course recommendations based on Salesforce use cases, Google Drive research, and the Databricks course catalog. Creates a customer-facing Google Doc with personalized course suggestions.

**Example prompts:**
- "Create training recommendations for Meta"
- "Generate a course plan for Acme Corp focusing on data engineering"
- "Create me a training plan for Block"

**Slash command:** `/fe-customer-courses`

**Prerequisites:** `/salesforce-authentication`, `/google-docs`

**Also invokes:** `customer-courses-generator` agent, `/google-docs`

---

#### security-questionnaire

**What it does:** Guides you through completing security questionnaires and compliance documentation for customer procurement. Uses the SQRC Search Engine (500+ pre-approved answers) as the primary source. Important: Azure Databricks questionnaires are prohibited - route to Azure team.

**Example prompts:**
- "I need to complete a security questionnaire for a customer"
- "Help me answer questions about access control and authentication for a security assessment"

**Slash command:** `/security-questionnaire`

**Prerequisites:** Glean MCP, Slack MCP, `/google-docs`, NDA verified

**Also invokes:** `sqrc-questionnaire` agent, `sqrc-validator` agent, `/google-docs`

---

#### fe-poc-doc

**What it does:** Creates and maintains Proof of Concept documentation for customer engagements. Gathers data from Salesforce, JIRA, Slack, Glean, and existing documents to produce a structured POC document covering scope, success criteria, timelines, and staffing.

**Example prompts:**
- "Create a POC document for the CrowdStrike evaluation of Unity Catalog"
- "Build a POC doc for the Meta engagement. Here's the Salesforce opp: 0061234567890ABCDEF"
- "Update the POC doc for Discord with the latest evaluation results"

**Slash command:** `/fe-poc-doc`

**Prerequisites:** `/salesforce-authentication`, `/validate-mcp-access`, `/google-docs`

**Also invokes:** `fe-poc-doc` agent

---

#### fe-poc-postmortem

**What it does:** Generates comprehensive post-mortem retrospectives for competitive POCs that we didn't win. Analyzes data from multiple sources and creates a narrative document with evaluation history, competition analysis, challenges, and recommendations.

**Example prompts:**
- "Write a post-mortem for the Acme Corp POC we lost"
- "Create a retrospective for our failed evaluation at BigCo"

**Slash command:** `/fe-poc-postmortem`

**Prerequisites:** `/validate-mcp-access`, `/salesforce-authentication`

**Also invokes:** `poc-postmortem` agent, `/google-docs`

---

#### fe-account-transition

**What it does:** Creates comprehensive account transition documents for new AEs and SAs taking over accounts. Gathers data from Salesforce, Logfood, Glean, and JIRA to provide technical depth on production use cases, evaluation pipelines, and historical context.

**Example prompts:**
- "Create an account transition document for Block - Ryan Werth is the new AE"
- "I need to hand off the Instacart account to a new SA"
- "A new team is taking over the Square account, help me create a handoff document"

**Slash command:** `/fe-account-transition`

**Prerequisites:** `/salesforce-authentication`, `/databricks-authentication`, `/google-auth`

**Also invokes:** `/logfood-querier`, `/google-docs`

---

#### performance-tuning

**What it does:** Systematic approach to diagnosing and optimizing Databricks workload performance. Covers Spark jobs (PySpark/Scala), Spark SQL, DBSQL warehouse queries, and Structured Streaming. Uses the "4 S's" framework: Skew, Spill, Shuffle, Small Files.

**Example prompts:**
- "My customer has a SQL query that takes 45 seconds on their SQL warehouse, help me optimize it"
- "Customer's ETL Spark job went from 2 hours to 8 hours after a data volume increase"
- "I think my customer has a data skew problem in their Spark job"
- "Customer's structured streaming job has increasing latency"
- "Help me analyze these Spark UI screenshots"

**Slash command:** `/performance-tuning`

**Prerequisites:** `/databricks-authentication`

**Also invokes:** `/databricks-query`

---

#### databricks-sizing

**What it does:** Gets Databricks cost estimates using the Quicksizer agent. Calculates monthly DBU costs for ETL, Data Warehousing, ML, Interactive, Lakebase, Lakeflow Connect, and Apps workloads.

**Example prompts:**
- "Size an ETL workload: 500GB daily data on AWS Premium with 10 batch jobs"
- "Use quicksizer to estimate Databricks costs for a customer"
- "How much would it cost to run this workload on Databricks?"

**Slash command:** `/databricks-sizing`

**Prerequisites:** `/databricks-authentication` (logfood profile)

**Also invokes:** Quicksizer API

---

#### databricks-troubleshooting

**What it does:** Provides a systematic troubleshooting guide for common Databricks issues. Covers clusters, jobs, authentication, performance, and networking problems with diagnostic steps and escalation paths.

**Example prompts:**
- "My customer's cluster won't start"
- "Help me troubleshoot this Databricks job failure"

**Slash command:** `/databricks-troubleshooting`

**Prerequisites:** None

**Also invokes:** Loads related skills contextually

---

#### fe-databricks-feature-tester

**What it does:** Actually tests whether Databricks features work by running real end-to-end tests. Provisions infrastructure, executes test code, and documents results. Use this when you want to *run* a test, not just research whether something is supported.

**Example prompts:**
- "Test if materialized views support incremental refresh on Delta tables with CDF enabled"
- "I need to test whether Lakeflow Connect streaming tables work with UniForm for Iceberg reads"
- "Verify that Delta Sharing works with streaming tables - test it end-to-end"

**Slash command:** `/fe-databricks-feature-tester`

**Prerequisites:** `/databricks-authentication`, workspace access

**Also invokes:** `product-question-researcher` agent (for research phase), `databricks-feature-test-executor` agent (for execution), `/databricks-fe-vm-workspace-deployment` or `/databricks-oneenv-workspace-deployment`, `/google-docs`

---

#### fe-architecture-diagram

**What it does:** Generates professional architecture diagrams with vendor icons (Databricks, AWS, GCP, Snowflake, Kafka). Supports two engines: mingrammer/diagrams (Python, for icon-heavy diagrams) and Mermaid (for simple flowcharts). Features visual feedback-driven refinement via Chrome DevTools.

**Example prompts:**
- "Create an architecture diagram showing data flow from Kafka to Databricks to Snowflake"
- "Build a diagram of our customer's current architecture"

**Slash command:** `/fe-architecture-diagram`

**Prerequisites:** Auto-installs dependencies on first use

**Also invokes:** Chrome DevTools (for visual feedback loop)

---

#### draft-rca

**What it does:** Drafts Root Cause Analysis documents for incidents using data from Salesforce, JIRA, Slack, and email. Cross-references multiple sources and generates a formatted Google Doc with timeline, root cause, impact, and action items.

**Example prompts:**
- "Draft an RCA for ES-123456"
- "Create a root cause analysis from this Salesforce case and JIRA ticket"

**Slash command:** `/draft-rca`

**Prerequisites:** `/salesforce-authentication`, JIRA MCP, Slack MCP

**Also invokes:** `rca-doc` agent, `/google-docs`

---

### 6. JIRA (jira-tools)

---

#### jira-actions

**What it does:** Searches, views, comments on, and creates ES (Engineering Support) tickets using JIRA MCP tools. Supports JQL search, viewing ticket details, adding comments in ADF format, and creating new tickets (with access restrictions since Dec 2024).

**Example prompts:**
- "Search for ES tickets related to Unity Catalog"
- "Show me the details of ES-123456"
- "Add a comment to this ES ticket"
- "File an ES ticket for this customer issue"

**Slash command:** `/jira-actions`

**Prerequisites:** JIRA MCP (verify with `/validate-mcp-access`)

**Also invokes:** `jira-ticket-assistant` agent (for complex operations)

---

### 7. Expenses (fe-file-expenses)

---

#### file-expenses

**What it does:** Automates expense report filing with Emburse ChromeRiver. Orchestrates specialized sub-agents to: analyze your past expense patterns, scan calendar and downloads for receipts, analyze receipt images, process them into line items, and create draft expense reports. Does NOT auto-submit - always presents for review first.

**Example prompts:**
- "File my expenses"
- "Help me submit my expense report"
- "Process my receipts from last month"

**Slash command:** `/file-expenses`

**Prerequisites:** Chrome DevTools MCP, `/google-auth` (for calendar access)

**Also invokes:** `historical-profile-builder` agent, `expense-identifier` agent, `receipt-analyzer` agent, `expense-line-item-processor` agent, `/emburse-expenses`

---

#### emburse-expenses

**What it does:** Low-level API primitives for Emburse ChromeRiver expense management. Create reports, add expenses, link receipts, submit reports. Typically used by the `file-expenses` skill rather than directly.

**Example prompts:**
- (Typically invoked by `file-expenses`, not directly)
- "Create an expense report in ChromeRiver"

**Slash command:** `/emburse-expenses`

**Prerequisites:** Chrome DevTools MCP, Emburse access

**Also invokes:** None - this is a primitive used by `file-expenses`

---

### 8. Diagrams (specialized-agents)

---

#### lucid-diagram

**What it does:** Generates architecture, data flow, and sequence diagrams as Graphviz DOT files and converts them to PNG images and Lucid Chart-compatible XML for import.

**Example prompts:**
- "Create an architecture diagram I can import into Lucid Chart"
- "Generate a data flow diagram as a DOT file"

**Slash command:** `/lucid-diagram`

**Prerequisites:** Graphviz (auto-installed)

**Also invokes:** None

---

### 9. Setup & Maintenance (vibe-setup)

---

#### configure-vibe

**What it does:** Comprehensive environment setup and validation. Installs CLI dependencies (Databricks CLI, Salesforce CLI, gcloud, AWS CLI, uv, etc.), creates directory structures, and configures credentials. Resilient - continues past individual failures and reports them at the end.

**Example prompts:**
- "Configure vibe"
- "Set up my environment"

**Slash command:** `/configure-vibe`

**Prerequisites:** Homebrew

**Also invokes:** None

---

#### validate-mcp-access

**What it does:** Validates that MCP connections (Slack, Glean, JIRA, Confluence) are active and properly authenticated. Reports status in a table and provides login instructions for any inactive connections.

**Example prompts:**
- "Validate my MCP access"
- "Check if my Slack and JIRA connections are working"

**Slash command:** `/validate-mcp-access`

**Prerequisites:** Databricks CLI (logfood profile)

**Also invokes:** None

---

#### vibe-publish-plugin

**What it does:** Publishes plugins to the vibe marketplace. Handles validation, PR creation, and the publishing workflow.

**Example prompts:**
- "Publish my plugin to the marketplace"
- "Submit my plugin for review"

**Slash command:** `/vibe-publish-plugin`

**Prerequisites:** Git access to the vibe repo

**Also invokes:** None

---

## Common Workflows

### Onboarding to a New Account

```
1. "Authenticate with Salesforce and Databricks"
   → /salesforce-authentication, /databricks-authentication

2. "Show me the use cases and consumption for Acme Corp"
   → /salesforce-actions, /logfood-querier

3. "Create an account transition document for Acme Corp - I'm the new SA"
   → /fe-account-transition → creates Google Doc
```

### Answering Customer Product Questions

```
1. "Do materialized views support incremental refresh on foreign tables?"
   → /product-question-research → researches docs, Glean, Slack → creates Google Doc

-- OR for multiple questions at once --

1. "Draft responses to the customer questions in this doc: [Google Doc URL]"
   → /fe-answer-customer-questions → researches each question → creates combined Q&A doc
```

### Deploying a Customer Demo

```
1. "Create a serverless workspace for my demo"
   → /databricks-fe-vm-workspace-deployment

2. "Create a demo showing streaming ETL with Delta Live Tables"
   → /databricks-demo → provisions resources, generates data, writes code

3. "Build an app with a React dashboard for the demo"
   → /databricks-apps → scaffolds and deploys full-stack app
```

### Filing Expenses

```
1. "File my expenses"
   → /file-expenses → runs profiler → scans calendar/downloads →
   analyzes receipts → creates draft report → presents for review
```

### Creating a POC Document

```
1. "Create a POC document for the CrowdStrike evaluation of Unity Catalog"
   → /fe-poc-doc → gathers data from Salesforce, JIRA, Slack →
   collaborates on scope/criteria → creates Google Doc
```

### Researching and Testing a Feature

```
1. "Does Delta Sharing work with streaming tables?"
   → /product-question-research → creates research doc

2. "Actually test it end-to-end"
   → /fe-databricks-feature-tester → provisions workspace → runs tests → documents results
```

### Weekly UCO Updates

```
1. "Do my weekly UCO updates"
   → /uco-updates → discovers your UCOs → walks through required fields → updates Salesforce
```

### Getting a Cost Estimate

```
1. "Size a data warehousing workload: 50 concurrent users, 2TB data, AWS Premium"
   → /databricks-sizing → calls Quicksizer API → returns monthly DBU estimate
```

---

## Tips & Troubleshooting

### Authentication Order

Many skills need authentication first. The general order is:

1. **`/configure-vibe`** - First time only, installs dependencies
2. **`/validate-mcp-access`** - Verify Slack, Glean, JIRA, Confluence connections
3. **Platform-specific auth** as needed:
   - `/databricks-authentication` - Before any Databricks operations
   - `/salesforce-authentication` - Before any Salesforce operations
   - `/google-auth` - Before any Google Workspace operations (usually auto-triggered)
   - `/aws-authentication` - Before One-Env workspace deployments or AWS integrations

### When Skills Don't Auto-Trigger

If Claude doesn't pick up the right skill from your natural language:
- Try the slash command directly: `/skill-name`
- Be more specific in your prompt - include keywords from the skill description
- Check that the skill's prerequisites are met (auth, MCP connections)

### MCP Connection Issues

If you see errors about Slack, JIRA, Glean, or Confluence:
1. Run "Validate my MCP access" (`/validate-mcp-access`)
2. Follow the login URLs provided for any inactive connections
3. If connections keep failing, try: `echo "" | llm agent configure mcp`

### Context Window Getting Full

Long conversations consume context. When things slow down or Claude seems to lose track:
- Use `/compact` to compress the conversation
- Start a new session for unrelated tasks
- Be concise in your prompts

### Common Errors

| Error | Solution |
|-------|----------|
| "GraphQL: Could not resolve repository" | Log in to GitHub EMU account (ends in `_data`) via Okta |
| "permission denied" with dbexec | Run `vibe install --force-reinstall` |
| "command not found: llm" | Add `alias llm="dbexec repo run llm"` to your `~/.zshrc` |
| "Cost limit exceeded" | Request a seat license at go/ai-devtools/quota |
| Chrome DevTools MCP disconnected | Restart vibe session, ensure Chrome is open |

### Getting Help

- **`/help`** in Claude Code for built-in commands
- **Vibe repo issues** for bug reports and feature requests
- **#vibe-users** Slack channel for questions and discussion
