# TAP Map - SFDC Schema and Writeback Reference

## CRITICAL: User Consent Required

**NEVER write data to Salesforce without EXPLICIT user approval.** This is a production CRM system. You MUST:

1. **STOP** after generating the Google Doc
2. **SHOW** the user the Google Doc URL and summary
3. **ASK** explicitly: "Would you like me to write this TAP Map data to Salesforce Account Plan?"
4. **WAIT** for affirmative consent ("yes", "approve", "proceed")
5. **CHECK** for existing data in SFDC (see Diff Before Write below)
6. **SHOW** the diff if existing data is found — get a SECOND confirmation before overwriting
7. **ONLY THEN** proceed

### Diff Before Write (HARD REQUIREMENT)

Before writing ANY data to SFDC, you MUST check for existing records and show the user a comparison if data already exists. **This is not optional.**

**Step 1:** Query all existing `AccountPlanRelated__c` records for the Account Plan (EBChampion record type).

**Step 2:** If existing records are found, display the diff as a **compact ASCII table** showing **ONLY rows that changed**. Omit unchanged fields entirely — they are noise. The table MUST use box-drawing characters and look like this:

```
SFDC Writeback Diff - [Account Name]
Account Plan: [Plan Name] (ID: [Plan ID])
URL: https://databricks.lightning.force.com/lightning/r/AccountPlan/[ID]/view

Showing CHANGES ONLY (N of 32 fields):

┌──────────────┬─────────┬───────────────────┬───────────────────┬───────────────┐
│ Workload     │ Field   │ Current           │ New               │ Change        │
├──────────────┼─────────┼───────────────────┼───────────────────┼───────────────┤
│ Ingest       │ TAM     │ $1,200,000        │ $1,500,000        │ ⬆ UPDATE      │
│              │ DBRX %  │ 40%               │ 45%               │ ⬆ UPDATE      │
│              │ Tools   │ Snowflake         │ ADF;Fabric        │ UPDATE        │
├──────────────┼─────────┼───────────────────┼───────────────────┼───────────────┤
│ ML Stack     │ EB      │ [none]            │ Oscar Maldonado   │ ADD           │
│              │ TAM     │ [none]            │ $225,000          │ ADD           │
├──────────────┼─────────┼───────────────────┼───────────────────┼───────────────┤
│ Data Sharing │ EB      │ Tom Smith         │ [GAP]             │ ⚠️ WOULD CLEAR │
└──────────────┴─────────┴───────────────────┴───────────────────┴───────────────┘

Unchanged: I&T Champion, DW (all), GenAI (all), Lakebase (all), Gov (all), Format (all)
```

**Formatting rules:**
- Box-drawing characters (┌ ─ ┬ ┐ │ ├ ┼ ┤ └ ┴ ┘) — NOT markdown tables or flat lists
- **Only show rows where Current != New** — omit NO CHANGE rows
- Use compact labels: `EB` (not Champion), `Tools` (not Competitors), `Ingest` (not Ingest & Transform)
- Group by workload, workload name on first row only, no row separator within a group
- Change values: `⬆ UPDATE`, `⬇ UPDATE`, `ADD`, `⚠️ WOULD CLEAR`
- End with a one-line "Unchanged:" summary listing workloads/fields with no changes

**Step 3:** Below the table, include a brief **Summary**:
```
N fields updated across M workloads. No fields cleared.
(or: ⚠️ 1 field would be CLEARED — Data Sharing EB.)

Proceed with SFDC writeback?
```

**Step 4:** Ask the user to confirm: "Proceed with SFDC writeback? This will update N existing records with the changes above."

**Step 5:** Only proceed after the user confirms. If the user says to skip certain workloads or fields, respect that.

**If no existing records are found**, skip the diff and proceed after the initial consent from step 4 above.

---

## SFDC Data Model

### Parent Object: `AccountPlan`

This is a **standard object** (no `__c` suffix). Key field: `AccountId` (not `Account__c`).

```bash
# Find the current Account Plan
sf data query --query "SELECT Id, Name, AccountId FROM AccountPlan WHERE AccountId = '${ACCOUNT_ID}' ORDER BY CreatedDate DESC LIMIT 1" --json
```

**URL format:** `https://databricks.lightning.force.com/lightning/r/AccountPlan/{ACCOUNT_PLAN_ID}/view`

### Child Object: `AccountPlanRelated__c`

TAP Map data is stored as **one record per workload area** in `AccountPlanRelated__c`, linked to the parent `AccountPlan`.

**Record Type:** `EBChampion` (ID: `012Vp000002oeCPIAY`)

### Fields

| Field API Name | Type | Description |
|---|---|---|
| `AccountPlan__c` | Reference | Parent AccountPlan ID |
| `RecordTypeId` | ID | Must be `012Vp000002oeCPIAY` (EBChampion) |
| `Category__c` | Picklist | Workload area name (see values below) |
| `RelatedContact__c` | Reference | Contact ID for EB/Champion |
| `EstimatedTAM__c` | Currency | TAM estimate in dollars (e.g., 1500000) |
| `EstimatedDatabricks__c` | Percent | DBRX % of TAM as whole number (e.g., 45 for 45%) |
| `CompetitorProducts__c` | Multi-select picklist | Semicolon-separated values from standard picklist |

---

## Picklist Values

### Category__c (exact match required)

These are the 8 workload area values for the TAP Map:
- `Ingest & Transform`
- `ML Stack`
- `Data Warehouse`
- `GenAI Stack`
- `Data Sharing`
- `Data Governance`
- `Data Format`
- `Lakebase / OLTP`

### CompetitorProducts__c (multi-select, semicolon-separated)

**MUST use exact picklist values.** Common ones:

| Our Reference | SFDC Picklist Value |
|---|---|
| Azure Data Factory | `Azure Data Factory` |
| Azure Data Explorer | `Azure Data Explorer` |
| Azure Synapse | `Azure Synapse` |
| Azure AI Studio | `Azure AI Studio` |
| Azure Machine Learning | `Azure Machine Learning` |
| Microsoft Fabric | `Microsoft Fabric` |
| PowerBI | `Microsoft Power BI` |
| Microsoft Purview | `Microsoft Purview` |
| SQL Server | `Microsoft SQL Server` |
| Snowflake | `Snowflake` |
| Redshift | `AWS Redshift` |
| SageMaker | `AWS SageMaker AI / Bedrock` |
| BigQuery | `Google BigQuery` |
| Vertex AI | `Google Vertex AI` |
| Looker | `Google Looker` |
| dbt | `dbt` |
| Fivetran | `Fivetran` |
| Informatica | `Informatica` |
| Airflow | `Apache Airflow` |
| Confluent | `Confluent Cloud` |
| Collibra | `Collibra` |
| DataRobot | `DataRobot` |
| Dataiku | `Dataiku` |
| Hadoop | `Hadoop` |
| OpenAI | `OpenAI` |
| Anthropic | `Anthropic` |
| MongoDB | `Other` (not in picklist) |

Use `Other` for any tool not in the picklist.

---

## Writeback Process

### Step 1: Check for Existing Records and Build Diff

Query ALL fields for existing records (needed for the diff comparison):
```bash
sf data query --query "SELECT Id, Category__c, RelatedContact__c, RelatedContact__r.Name, EstimatedTAM__c, EstimatedDatabricks__c, CompetitorProducts__c FROM AccountPlanRelated__c WHERE AccountPlan__c = '${ACCOUNT_PLAN_ID}' AND RecordTypeId = '012Vp000002oeCPIAY'" --json
```

**If records are returned:** Build and display the diff table (see "Diff Before Write" above). Wait for user confirmation before proceeding to Step 3.

**If no records are returned:** Proceed directly — the initial consent is sufficient.

### Step 2: Look Up Contact IDs for Champions

```bash
sf data query --query "SELECT Id, Name, Title FROM Contact WHERE AccountId = '${ACCOUNT_ID}' AND (Name = 'Champion1' OR Name = 'Champion2')" --json
```

### Step 3: Create/Update Records (One per Workload)

**Create new record:**
```bash
sf data create record --sobject AccountPlanRelated__c \
  --values "AccountPlan__c='${ACCOUNT_PLAN_ID}' RecordTypeId='012Vp000002oeCPIAY' Category__c='Ingest & Transform' RelatedContact__c='${CONTACT_ID}' EstimatedTAM__c=1500000 EstimatedDatabricks__c=45 CompetitorProducts__c='Azure Data Factory;Microsoft Fabric'" \
  --json
```

**Update existing record:**
```bash
sf data update record --sobject AccountPlanRelated__c \
  --record-id '${EXISTING_RECORD_ID}' \
  --values "RelatedContact__c='${CONTACT_ID}' EstimatedTAM__c=1500000 EstimatedDatabricks__c=45 CompetitorProducts__c='Azure Data Factory;Microsoft Fabric'" \
  --json
```

Repeat for all 8 workload areas.

### Step 4: Verify Writeback

```bash
sf data query --query "SELECT Id, Category__c, RelatedContact__r.Name, EstimatedTAM__c, EstimatedDatabricks__c, CompetitorProducts__c FROM AccountPlanRelated__c WHERE AccountPlan__c = '${ACCOUNT_PLAN_ID}' AND RecordTypeId = '012Vp000002oeCPIAY' ORDER BY Category__c" --json
```

---

## Data Transformation Rules

### EB/Champion → RelatedContact__c
- If value is a name (e.g., "Brandon Bates, Principal Architect"), look up Contact ID by name
- If value is "[CHAMPION GAP]", SKIP the `RelatedContact__c` field (create record without it)
- If value contains "OR" (multiple options), look up the first name's Contact ID

### Estimated TAM → EstimatedTAM__c
- Convert range (e.g., "$1.2M - $1.8M") to midpoint: 1500000
- Format as plain number (no $ or commas)
- For overlay metrics (Governance, Data Format), use 0 if no dollar estimate

### DBRX % of TAM → EstimatedDatabricks__c
- Write as whole number: 45 (not 0.45 or "45%")
- If 0%, write as 0
- If N/A, write as 0

### Incumbent Tools → CompetitorProducts__c
- Map tool names to exact picklist values (see table above)
- Separate multiple values with semicolons: `Microsoft Power BI;Microsoft Fabric`
- If tool not in picklist, use `Other`
- If "None identified", omit the field

---

## Safety Checks

- If a Contact ID is not found, SKIP `RelatedContact__c` (create record without it)
- If a competitor tool is not in the picklist, use `Other` or skip it
- If a record already exists for a category, UPDATE instead of creating a duplicate
- If a currency value is invalid, SKIP that field and log an error

## Error Handling

| Issue | Resolution |
|---|---|
| User does not consent | STOP. Do not write. Google Doc is the deliverable. |
| No Account Plan found | STOP. Notify user to create Account Plan in Salesforce first. |
| Field does not exist | SKIP that field. Log warning. Continue with others. |
| Invalid data type | SKIP that field. Log error. Continue with others. |
| SFDC API error (permissions) | STOP. Notify user. Suggest manual writeback from Google Doc. |
| Multiple Account Plans found | Ask user to specify which one (by Name or ID). |
