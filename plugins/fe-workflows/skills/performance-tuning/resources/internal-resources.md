# Internal Resources (Databricks Employees Only)

## Key Go-Links

| Go-Link | Description |
|---|---|
| `go/slowquery` | Systematic SQL query performance tuning approach |
| `go/performance-guide` | Field Guide to DBSQL Performance / IWM |
| `go/aqe-configuration` | Complete AQE configuration reference |
| `go/perfsme` | Performance and Delta SME knowledge base |
| `go/queryprofile` | Query Profile project documentation |
| `go/db-assistant/optimize` | AI-based query optimization dashboard |

## Internal Confluence Pages

| Page | URL |
|---|---|
| Overall Performance Tuning and Cost Optimization | https://databricks.atlassian.net/wiki/spaces/FE/pages/3351478545 |
| Compute Level - Spark Optimizations | https://databricks.atlassian.net/wiki/spaces/FE/pages/3351478572 |
| Storage Level - Cost Optimization | https://databricks.atlassian.net/wiki/spaces/FE/pages/3351249159 |
| AQE Configuration Reference | https://databricks.atlassian.net/wiki/spaces/UN/pages/2511634440 |

## Internal Tools

| Tool | Description |
|---|---|
| **DBR Doctor** | Unified debug/performance tool — single pane of glass for DBR analysis. See https://databricks.atlassian.net/wiki/spaces/UN/pages/5544247341 |
| **Logfood** | Query `prod_ds.spark_logs` for event log analysis. Replay Spark event logs to recreate Spark UI for terminated workloads. |
| **system.query.history** | Query execution records (365-day retention) for SQL warehouses and serverless compute |

## Slack Channels for Escalation

| Channel | Topic |
|---|---|
| #dbsql-perf-help | DBSQL performance questions |
| #query-history-system-table | system.query.history questions |
| #streaming-help | Structured Streaming performance |
| #delta-users | Delta Lake optimization |
| #photon | Photon engine questions |
| #aqe | Adaptive Query Execution |
