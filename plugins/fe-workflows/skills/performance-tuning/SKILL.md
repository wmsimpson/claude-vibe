---
name: performance-tuning
description: Optimize Databricks workload performance through systematic tuning and analysis. Use this skill when a customer has slow Spark jobs, slow SQL queries, high costs, data skew, shuffle/spill problems, streaming latency issues, or needs help reading Spark UI / Query Profile output. Covers Spark code, Spark SQL, DBSQL, and Structured Streaming workloads.
user-invocable: true
---

# Databricks Performance Tuning Workflow

This skill provides a systematic, end-to-end approach to diagnosing and optimizing Databricks workload performance for customer engagements. It covers Spark jobs (PySpark/Scala), Spark SQL queries, DBSQL warehouse queries, and Structured Streaming workloads.

**Announce at start:** "I'm using the performance-tuning skill to help diagnose and optimize your Databricks workload performance."

## Overview: The Performance Tuning Framework

Follow this systematic methodology for every performance engagement:

```
1. INTAKE      → Understand the workload, gather artifacts
2. BASELINE    → Establish current performance metrics
3. DIAGNOSE    → Identify bottlenecks using the "4 S's" framework
4. OPTIMIZE    → Apply targeted fixes for identified bottlenecks
5. VALIDATE    → Measure improvement, iterate if needed
6. DOCUMENT    → Record findings and recommendations
```

The **"4 S's"** framework for bottleneck identification:
- **S**kew — Uneven data distribution across partitions
- **S**pill — Data evicted from memory to disk
- **S**huffle — Excessive data movement between executors
- **S**mall Files — Too many tiny files degrading I/O

**Reference files:** This skill uses progressive disclosure. Detailed examples, queries, and reference tables are in the `resources/` folder:
- `resources/diagnostic-queries.sql` — SQL queries for system.query.history analysis
- `resources/code-examples.md` — Salting, UDF replacement, shuffle, streaming code examples
- `resources/spark-config-reference.md` — Complete Spark configuration quick reference
- `resources/decision-trees.md` — Diagnostic decision trees for slow queries and streaming
- `resources/anti-patterns.md` — Common anti-patterns checklist
- `resources/internal-resources.md` — Internal go-links, Confluence pages, tools, Slack channels

---

## Step 1: Intake — Gather Information

### Identify the Workload Type

Ask the user which type of workload they are tuning. This determines the diagnostic path:

| Workload Type | Compute | Key Diagnostic Tool | Primary Metrics |
|---|---|---|---|
| **Spark Jobs** (PySpark/Scala) | All-Purpose or Job Clusters | Spark UI | Stage duration, task metrics, shuffle/spill |
| **Spark SQL** (notebooks) | All-Purpose or Serverless | Spark UI + EXPLAIN | Query plan, stage metrics |
| **DBSQL Queries** | SQL Warehouses | Query Profile + system.query.history | Execution/compilation duration, scan size |
| **Structured Streaming** | All-Purpose or Job Clusters | Spark UI + Streaming tab | Batch duration, processing rate, state size |

### Required Inputs from the Customer

Depending on what the customer can provide, gather one or more of these:

**Option A: Direct Workspace Access (Preferred)**
If you have access to the customer's Databricks workspace:
- Workspace URL and credentials
- The specific job/query/notebook to analyze
- Access to Spark UI or Query Profile

**Option B: Exports and Artifacts**
If analyzing offline (typical for customer engagements):

| Artifact | How to Get It | What It Tells You |
|---|---|---|
| **Spark UI screenshots** | Cluster → Spark UI → Jobs/Stages/SQL tabs | Task distribution, shuffle, spill, duration |
| **Query Profile export** | DBSQL → Query History → Query Profile | Operator-level timing, scan stats, join strategies |
| **EXPLAIN output** | Run `EXPLAIN EXTENDED <query>` or `EXPLAIN FORMATTED <query>` | Logical and physical query plan |
| **Query text** (SQL or code) | Copy from notebook/job | Code-level anti-patterns, join order, UDFs |
| **Spark event logs** | Cluster → Event Logs → Download | Full replay of Spark execution (can recreate Spark UI) |
| **system.query.history export** | Run diagnostic queries from `resources/diagnostic-queries.sql` | Duration breakdown, scan stats across queries |
| **Cluster/warehouse config** | Cluster → Configuration or Warehouse → Settings | Instance types, autoscaling, Spark configs |
| **Table statistics** | Run `DESCRIBE DETAIL <table>` and `DESCRIBE EXTENDED <table>` | File count, size, partitioning, clustering |
| **Job run history** | Jobs → Run History → Duration trend | Regression detection over time |

**Option C: Description Only**
If no artifacts are available, ask targeted questions:
1. What is the workload doing? (ETL, analytics, ML, streaming)
2. How long does it currently take vs. expected?
3. What changed recently? (data volume, code, cluster config, DBR version)
4. What is the data volume? (rows, GB/TB, number of files)
5. What compute are they using? (cluster type, instance type, node count, warehouse size)
6. Are they using Photon?
7. Are they using Unity Catalog managed tables?

---

## Step 2: Baseline — Establish Current Performance

### For DBSQL Queries: Query system.query.history

Use the diagnostic SQL queries in `resources/diagnostic-queries.sql` to:
1. Find the top 20 slowest queries (last 7 days)
2. Analyze time breakdown for a specific query (compilation vs. execution vs. fetch)
3. Find queries with high spill (memory pressure)
4. Detect performance regressions (compare avg duration across two periods)

### For Spark Jobs: Establish Baseline Metrics

Collect from the Spark UI or job run history:

| Metric | Where to Find | Healthy Range |
|---|---|---|
| Total job duration | Jobs tab | Depends on workload |
| Longest stage duration | Stages tab (sort by duration) | Should be <50% of total |
| Task count per stage | Stage detail page | Ideally 2-4x number of cores |
| Max task duration vs. median | Stage → Summary Metrics | Max should be < 2x median |
| Shuffle read/write | Stages tab | Lower is better; watch for >10GB |
| Spill (memory + disk) | Stage detail → Spill columns | Should be 0; any spill = problem |
| Input size | Stages tab | Baseline for comparison |

### For Streaming: Establish Baseline Metrics

From the Streaming tab or `StreamingQueryProgress`:

| Metric | Where to Find | Healthy Range |
|---|---|---|
| Batch duration | Streaming tab | Should be < trigger interval |
| Processing rate (rows/sec) | Streaming tab | Should keep up with input rate |
| Input rate | Streaming tab | Compare to processing rate |
| State rows | Streaming tab | Watch for unbounded growth |
| Watermark delay | StreamingQueryProgress | Should not grow over time |

---

## Step 3: Diagnose — Identify Bottlenecks

### Reading the Spark UI

Navigate to the Spark UI: **Cluster → Spark UI** (for interactive clusters) or **Job Run → Spark UI** (for job clusters).

**Step 3a: Start at the Jobs tab**
- Sort jobs by duration — the longest job is your target
- Click into the longest job to see its stages

**Step 3b: Identify the bottleneck stage**
- Sort stages by duration — the longest stage is your bottleneck
- Note the stage's task count and shuffle read/write sizes

**Step 3c: Analyze the bottleneck stage in detail**
Click into the stage. Check these in order:

1. **Spill check**: Look at the top of the stage page for spill statistics. If you see "Spill (Memory)" or "Spill (Disk)" values > 0, you have a spill problem.

2. **Skew check**: In the Summary Metrics table, compare the **Max** task duration to the **75th percentile**:
   - If Max > 1.5x the 75th percentile → **Data skew detected**
   - If Max ≈ 75th percentile → No skew (tasks are balanced)

3. **Shuffle check**: Look at Shuffle Read Size and Shuffle Write Size:
   - Large shuffle (>10GB) with many tasks → Consider reducing shuffle
   - Shuffle spill → Memory too low for shuffle data

4. **I/O check**: Look at Input Size and Records Read:
   - Very high input with low output → Missing predicate pushdown or partition pruning
   - Many tasks reading tiny amounts → Small file problem

### Reading the DBSQL Query Profile

For SQL Warehouse queries, use the Query Profile instead of Spark UI:

1. Go to **SQL Editor → Query History** or **SQL Warehouse → Query History**
2. Click on the slow query → **Query Profile**
3. The profile shows a visual DAG of operators with timing

**What to look for in the Query Profile:**
- **Scan operators**: Check rows scanned vs. rows returned. A large ratio means poor pruning.
- **Join operators**: Check join strategy (BroadcastHashJoin vs. SortMergeJoin). SortMergeJoin on skewed data is slow.
- **Aggregate operators**: Check for spill indicators.
- **Exchange (shuffle) operators**: Check data size flowing through shuffles.
- **Time distribution**: The thickest/reddest nodes in the DAG are the bottlenecks.

### Reading EXPLAIN Output

Run `EXPLAIN EXTENDED` or `EXPLAIN FORMATTED` to get the query plan:

```sql
EXPLAIN EXTENDED
SELECT * FROM orders o
JOIN customers c ON o.customer_id = c.id
WHERE o.order_date > '2024-01-01';
```

**What to look for:**
- `BroadcastHashJoin` (good for small tables) vs. `SortMergeJoin` (expensive for large tables)
- `Filter` pushed into `Scan` (good - predicate pushdown) vs. `Filter` above `Scan` (bad)
- `PartitionFilters` present in scan (good - partition pruning working)
- `DataFilters` present in scan (good - data skipping working)
- `Exchange` nodes indicate shuffles — fewer is better
- `WholeStageCodegen` wrapping operators (good - Photon/codegen active)

---

## Step 4: Optimize — Apply Targeted Fixes

For detailed code examples of all optimizations below, see `resources/code-examples.md`.

### 4A. Data Skew Solutions

**Symptoms:** One or few tasks take much longer than others. Max task duration >> 75th percentile in Spark UI.

**Solution 1: Rely on AQE (Default, Automatic)**
AQE handles skew automatically on Databricks. Verify it's enabled:
```sql
SET spark.sql.adaptive.enabled;                    -- Should be true (default)
SET spark.sql.adaptive.skewJoin.enabled;           -- Should be true (default)
```

AQE skew detection thresholds (tune if needed):
```sql
SET spark.sql.adaptive.skewJoin.skewedPartitionThresholdInBytes = 268435456;  -- 256MB
SET spark.sql.adaptive.skewJoin.skewedPartitionFactor = 5;
```

**Solution 2: Force AQE skew handling for stubborn cases**
```sql
SET spark.sql.adaptive.forceOptimizeSkewedJoin = true;
```

**Solution 3: Salting technique (manual, for extreme skew)** — See `resources/code-examples.md` for PySpark and SQL salting examples.

**Solution 4: Broadcast join (if one side is small)**
```sql
SET spark.sql.autoBroadcastJoinThreshold = 1073741824;  -- 1GB
-- Or use a hint:
SELECT /*+ BROADCAST(small_table) */ *
FROM large_table JOIN small_table ON large_table.key = small_table.key;
```

### 4B. Spill Solutions

**Symptoms:** "Spill (Memory)" and "Spill (Disk)" > 0 in Spark UI stage details.

**Solution 1: Increase executor memory**
```sql
SET spark.executor.memory = 16g;
SET spark.executor.memoryOverhead = 4g;
```

**Solution 2: Increase shuffle partitions to reduce per-partition data size**
```sql
SET spark.sql.shuffle.partitions = auto;  -- Let AQE auto-tune (recommended)
-- Or manually: aim for 128MB-256MB per partition after shuffle
```

**Solution 3: Use memory-optimized instances** — Switch to `r5` (AWS) or `E` series (Azure).

**Solution 4: Enable Low Shuffle Merge for MERGE operations**
```sql
SET spark.databricks.delta.merge.enableLowShuffle = true;
```

### 4C. Shuffle Optimization

**Symptoms:** Large Shuffle Read/Write in Spark UI, many Exchange nodes in query plan.

**Solution 1: Let AQE optimize shuffle partitions**
```sql
SET spark.sql.shuffle.partitions = auto;
```

**Solution 2: Use broadcast joins to eliminate shuffles**
```sql
SET spark.sql.autoBroadcastJoinThreshold = 104857600;  -- 100MB
-- Or per-query: SELECT /*+ BROADCAST(dim_table) */ ...
```

**Solution 3: Optimize join order** — Place the largest table first; filter the fact table early.

**Solution 4: Avoid unnecessary shuffles in code** — See `resources/code-examples.md` for repartition vs. coalesce vs. optimized writes.

### 4D. Small File Solutions

**Symptoms:** Thousands of tiny files in a Delta table, slow scan times.

**Check file statistics:**
```sql
DESCRIBE DETAIL my_catalog.my_schema.my_table;
```

**Solution 1: Run OPTIMIZE**
```sql
OPTIMIZE my_catalog.my_schema.my_table;
```

**Solution 2: Enable Auto Compaction**
```sql
ALTER TABLE my_table SET TBLPROPERTIES ('delta.autoOptimize.autoCompact' = 'true');
```

**Solution 3: Enable Optimized Writes**
```sql
ALTER TABLE my_table SET TBLPROPERTIES ('delta.autoOptimize.optimizeWrite' = 'true');
```

**Solution 4: Tune target file size**
```sql
ALTER TABLE my_table SET TBLPROPERTIES ('delta.tuneFileSizesForRewrites' = 'true');
```

Auto-tuned defaults: Tables < 2.56 TB → 256 MB; 2.56-10 TB → 256 MB to 1 GB; > 10 TB → 1 GB.

**Solution 5: Enable Predictive Optimization (UC managed tables)**
```sql
ALTER CATALOG my_catalog ENABLE PREDICTIVE OPTIMIZATION;
```
Note: Predictive Optimization does NOT support ZORDER. Use Liquid Clustering instead.

### 4E. Delta Table / Storage Optimization

**Liquid Clustering (recommended for all new tables)**
```sql
CREATE TABLE my_table (...) CLUSTER BY (date_col, region_col);
-- Convert existing: ALTER TABLE my_table CLUSTER BY (date_col, region_col);
-- Change keys: ALTER TABLE my_table CLUSTER BY (new_col1, new_col2);
-- Trigger: OPTIMIZE my_table;
```

| Criteria | Liquid Clustering | Partition + Z-ORDER |
|---|---|---|
| New tables | Always recommended | Legacy approach |
| Evolving query patterns | Easy key changes | Painful (requires rewrite) |
| Write-heavy tables | Recommended (incremental) | High write amplification |
| Streaming tables | Required | Not supported well |

**Deletion Vectors:** `ALTER TABLE my_table SET TBLPROPERTIES ('delta.enableDeletionVectors' = 'true');`

**VACUUM:** `VACUUM my_table RETAIN 168 HOURS;`

**ANALYZE TABLE:** `ANALYZE TABLE my_table COMPUTE STATISTICS FOR ALL COLUMNS;`

### 4F. Join Optimization

| Join Type | When Used | Best For |
|---|---|---|
| BroadcastHashJoin | Small table < threshold | Dim lookups, small joins |
| ShuffleHashJoin | AQE detects small side at runtime | Medium-sized joins |
| SortMergeJoin | Both sides large | Large-to-large joins |
| BroadcastNestedLoopJoin | Non-equi joins + small table | Range joins, theta joins |

**Force specific join strategies with hints:**
```sql
SELECT /*+ BROADCAST(small_table) */ ...  -- Avoid shuffle entirely
SELECT /*+ MERGE(table1, table2) */ ...   -- Force sort-merge join
SELECT /*+ SHUFFLE_HASH(table1) */ ...    -- Shuffle hash join
```

**Common join anti-patterns:** Cartesian/Cross joins, joining on non-selective keys, missing join predicates, joining large tables without filtering first.

### 4G. Caching Strategies

| Cache Type | Scope | Use When |
|---|---|---|
| **Disk Cache** (Delta cache) | Automatic on SSD nodes | Always (auto-enabled) |
| **Spark Cache** (.cache()) | DataFrame/RDD | Multiple actions on same DataFrame |
| **DBSQL Result Cache** | SQL Warehouse | Repeated identical queries |
| **UI Cache** | Per-user in DBSQL UI | Dashboard display |

**Disable result cache for benchmarking:** `SET use_cached_result = false;`

### 4H. Adaptive Query Execution (AQE) Tuning

AQE is enabled by default. Only tune when defaults aren't working:

```sql
SET spark.sql.adaptive.enabled = true;
SET spark.sql.shuffle.partitions = auto;            -- Auto-tune partitions
SET spark.sql.adaptive.forceApply = true;           -- Force AQE even without exchanges
SET spark.sql.adaptive.autoBroadcastJoinThreshold = 104857600;  -- 100MB runtime conversion
```

### 4I. Photon Engine

Photon is a C++ vectorized engine. Automatically enabled on SQL Warehouses and Photon-enabled clusters.

**Operations NOT supported by Photon (fall back to Spark):** UDFs, some complex nested data types, some window functions with complex frames.

**Recommendation:** Replace UDFs with built-in Spark SQL functions. See `resources/code-examples.md` for examples.

### 4J. SQL Warehouse (DBSQL) Specific Tuning

| Warehouse Size | Cores | Memory | Best For |
|---|---|---|---|
| 2X-Small | 4 | 16 GB | Light testing |
| Small | 16 | 64 GB | Medium analytics |
| Medium | 32 | 128 GB | Heavy analytics |
| Large | 64 | 256 GB | Large concurrent workloads |

**Scaling:** Set min/max clusters for concurrency. Use Serverless for auto Photon + result caching. Set Auto Stop to 5-10 min for cost savings.

**Query optimization tips:** Use column pruning (avoid `SELECT *`), apply filters early in CTEs, use `TABLESAMPLE` for exploration.

### 4K. Structured Streaming Performance

Enable these for stateful streaming — see `resources/code-examples.md` for full configuration:
- **RocksDB State Store** — Recommended for all stateful queries
- **Changelog Checkpointing** — Recommended on DBR 13.3+
- **Asynchronous Checkpointing** — For checkpoint-bound workloads

| Issue | Symptom | Solution |
|---|---|---|
| High batch duration | Batch time >> trigger interval | Increase cluster size, optimize transformations |
| State store growth | State rows growing unboundedly | Add watermark, ensure state TTL |
| Checkpoint bottleneck | Processing done fast but batch duration high | Enable async + changelog checkpointing |
| Small file output | Many tiny files per batch | Enable trigger.availableNow or auto-compaction |
| Backpressure | Input rate >> processing rate | Scale up, optimize joins, reduce state |

---

## Step 5: Validate — Measure Improvement

After applying optimizations, compare against the baseline:

1. **Run the same query/job** with optimizations applied
2. **Compare metrics** side-by-side: total duration, stage durations, shuffle, spill, skew ratio
3. **Check for regressions** in other queries
4. **Run at production scale** — test with full data volume, not samples
5. **Disable caching for benchmarks:** `SET use_cached_result = false;`

---

## Step 6: Document — Record Findings

Create a summary using this structure (or use the `google-docs` skill for a formatted doc):

```markdown
## Performance Analysis Summary

### Workload Description
- [What the workload does]
- [Current performance] → [Target performance]

### Bottlenecks Identified
1. [Bottleneck]: [Evidence from Spark UI / Query Profile]

### Optimizations Applied
| Optimization | Setting/Change | Impact |
|---|---|---|
| [What was changed] | [Specific config/code] | [Before → After] |

### Results
- Duration: [Before] → [After] ([X]% improvement)

### Recommendations
1. [Ongoing recommendation]
```

---

## Quick Reference

For detailed reference material, see the resources folder:
- **Anti-patterns checklist:** `resources/anti-patterns.md`
- **Spark config reference:** `resources/spark-config-reference.md`
- **Decision trees:** `resources/decision-trees.md` (start here for quick triage)
- **Internal resources:** `resources/internal-resources.md` (go-links, Confluence, Slack channels)

## Example Invocations

```
User: My customer's ETL job went from 2 hours to 8 hours after a data volume increase
→ Gather Spark UI screenshots, check for skew/spill in longest stage,
  review shuffle partitions, check cluster sizing

User: This SQL query takes 45 seconds, should be under 5 seconds
→ Get EXPLAIN output, check Query Profile, look for missing predicates,
  check table clustering/partitioning, review join strategies

User: Help me optimize this PySpark notebook for a customer
→ Review code for anti-patterns (UDFs, collect, repartition),
  check Spark UI for bottlenecks, recommend configuration changes

User: Customer streaming job has increasing latency
→ Check batch duration trend, state store size, watermark config,
  enable RocksDB + changelog checkpointing, review cluster sizing
```
