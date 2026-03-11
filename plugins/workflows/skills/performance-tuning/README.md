# Performance Tuning

Systematic Databricks workload performance optimization using the "4 S's" framework (Skew, Spill, Shuffle, Small Files). Covers Spark jobs, Spark SQL, DBSQL warehouse queries, and Structured Streaming workloads.

## How to Invoke

### Slash Command

```
/performance-tuning
```

### Example Prompts

```
"My customer's ETL job went from 2 hours to 8 hours after a data volume increase"
"This SQL query takes 45 seconds, should be under 5 seconds"
"Customer streaming job has increasing latency, help me diagnose it"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Workspace Access or Artifacts | Either direct workspace access, or Spark UI screenshots / Query Profile exports / EXPLAIN output from the customer |

## What This Skill Does

1. **Intake** - Identifies workload type (Spark Jobs, Spark SQL, DBSQL, Streaming) and gathers artifacts
2. **Baseline** - Establishes current performance metrics from Spark UI, Query Profile, or system.query.history
3. **Diagnose** - Identifies bottlenecks using the "4 S's" framework (Skew, Spill, Shuffle, Small Files) by reading Spark UI stages, DBSQL Query Profile operators, and EXPLAIN plans
4. **Optimize** - Applies targeted fixes: AQE tuning, salting, broadcast joins, OPTIMIZE, Liquid Clustering, Photon, RocksDB state store, and more
5. **Validate** - Measures improvement against baseline metrics
6. **Document** - Records findings and recommendations

## Key Resources

| File | Description |
|------|-------------|
| `resources/diagnostic-queries.sql` | SQL queries for system.query.history analysis |
| `resources/code-examples.md` | Salting, UDF replacement, shuffle, and streaming code examples |
| `resources/spark-config-reference.md` | Complete Spark configuration quick reference |
| `resources/decision-trees.md` | Diagnostic decision trees for slow queries and streaming |
| `resources/anti-patterns.md` | Common anti-patterns checklist |
| `resources/internal-resources.md` | Internal go-links, Confluence pages, tools, and Slack channels |

## Related Skills

- `/databricks-troubleshooting` - For non-performance issues (cluster failures, auth errors, networking)
