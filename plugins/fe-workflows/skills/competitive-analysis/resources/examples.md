# Competitive Analysis Examples

This file contains example outputs for the `competitive-analysis` skill in both fast mode (default) and comprehensive mode.

---

## Quick Chat Response (Fast Mode - Default)

**User prompt:** "Why Databricks over Snowflake for streaming?"

**Output format:**
```markdown
## Quick Answer

Snowflake focuses on data warehousing with batch-oriented micro-batches (Snowpipe), while Databricks provides a unified data + AI platform with true streaming capabilities.

**Key advantage:** Databricks offers native streaming with sub-second latency and stateful operations—Snowflake uses micro-batches that can't handle true stream processing.

**When to choose Databricks:** When you need real-time streaming (< 1 second latency), stream-to-stream joins, or unified data engineering + ML workflows.

## Why Databricks

- **True Streaming Engine**: Structured Streaming provides sub-second latency with exactly-once semantics vs Snowflake's micro-batch approach (Snowpipe)
- **Unified Platform**: Data engineering, analytics, and ML in one platform vs Snowflake's warehousing focus
- **Open Formats**: Delta Lake (open format) vs Snowflake's proprietary storage format

## Snowflake Limitations

- **Batch-Oriented Streaming**: Snowpipe uses micro-batches (minutes latency) vs Databricks' true streaming (sub-second)
- **No Stateful Operations**: Can't do stream-to-stream joins or complex windowing—requires landing tables first
- **Warehousing Focus**: Optimized for SQL analytics, not data engineering or ML workflows

## When to Choose Each

**Choose Databricks when:**
- You need real-time streaming (< 1 second latency)
- Stream-to-stream joins or complex stateful operations required
- Unified data + AI platform needed

**Snowflake may be suitable when:**
- Simple batch ETL with micro-batch ingestion is sufficient
- SQL-only analytics team with no ML needs

## Sources

- Battlecard registry: Snowflake differentiators
- Databricks Structured Streaming documentation
```

---

## Governance Comparison (Fast Mode - Default)

**User prompt:** "Quick compare Unity Catalog to Polaris"

**Output format:**
```markdown
## Quick Answer

Polaris is Snowflake's data catalog (Snowflake-locked), while Unity Catalog provides an open, vendor-neutral catalog that works with any compute engine.

**Key advantage:** Unity Catalog is open and works with Spark, Presto, and other engines—Polaris only works with Snowflake.

**When to choose Databricks:** When you need multi-engine support, multi-cloud governance, or want to avoid vendor lock-in.

## Why Databricks

- **Open, Vendor-Neutral Catalog**: Unity Catalog works with any compute engine (Spark, Presto, etc.) vs Polaris (Snowflake-only)
- **Multi-Cloud Support**: Unity Catalog works across AWS, Azure, GCP vs Polaris (Snowflake ecosystem only)
- **Unified Governance**: Unity Catalog governs data and AI assets together vs Polaris (data-focused)

## Polaris Limitations

- **Snowflake-Locked**: Only works with Snowflake compute—can't use with other engines
- **Limited Ecosystem**: Tied to Snowflake's ecosystem vs Unity Catalog's open approach
- **No Multi-Engine Support**: Can't govern data accessed by Spark, Presto, or other engines

## When to Choose Each

**Choose Databricks when:**
- You need multi-engine support (Spark, Presto, etc.)
- Multi-cloud governance required
- Want to avoid vendor lock-in
- Need unified governance for data + AI assets

**Polaris may be suitable when:**
- You're 100% committed to Snowflake ecosystem
- Only need Snowflake-specific governance

## Sources

- Battlecard registry: Polaris differentiators
```

---

## Comprehensive Analysis (Comprehensive Mode - When Requested)

**User prompt:** "Create a comprehensive battlecard comparing Databricks to Snowflake for real-time streaming"

**Workflow execution:**
1. Identify context: Competitor (Snowflake), Product area (Streaming), Use case (Real-time analytics)
2. Find battlecards: Databricks vs Snowflake battlecard (streaming, governance, ML tags)
3. Read battlecard: Extract key differentiators, competitor gaps
4. Search Glean: Internal competitive intelligence, win/loss stories, pricing comparisons
5. Research Databricks: Lakeflow streaming capabilities, DLT, exactly-once semantics
6. Research Snowflake: Snowpipe Streaming docs, micro-batch ingestion, limitations
7. Generate Google Doc using markdown converter

**Output message:**
```
✅ Competitive analysis complete!

📄 Document: Databricks vs Snowflake: Real-Time Streaming
🔗 https://docs.google.com/document/d/xyz789

Key points:
- Databricks: Native streaming with sub-second latency
- Snowflake: Batch-oriented (Snowpipe = micro-batches)
- Top differentiator: Stateful stream processing (windowing, joins)

Sources:
- Internal battlecard (updated 2024-12-01)
- Glean: Customer win story, Q4 2024 pricing comparison
- Databricks DLT documentation
- Snowflake Snowpipe Streaming docs
```

---

## Additional Example Questions

### Amazon Redshift
- "Why Databricks over Redshift for data warehousing?"
- "Redshift vs Databricks for SQL analytics performance"
- "Customer says Redshift is cheaper. How do we respond?"

### Microsoft Fabric
- "Why Databricks over Microsoft Fabric for unified analytics?"
- "How does Delta Lake compare to Fabric's OneLake?"
- "Customer is Microsoft-heavy. Should they choose Fabric or Databricks?"

### Governance Competitors
- "Unity Catalog vs Microsoft Purview"
- "Unity Catalog vs AWS Glue Catalog"
- "Why is Unity Catalog better than Polaris?"

### OLTP Competitors
- "Should we use Databricks or Lakebase for this use case?"
- "Databricks vs Supabase for transactional workloads"
