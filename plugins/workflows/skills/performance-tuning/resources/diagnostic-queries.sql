-- Databricks Performance Tuning: Diagnostic SQL Queries
-- Run these against system.query.history to establish baselines and identify bottlenecks

-- ============================================================================
-- 1. Find the top 20 slowest queries in the last 7 days
-- ============================================================================
SELECT
  statement_id,
  SUBSTRING(statement_text, 1, 200) AS query_preview,
  executed_by,
  execution_status,
  start_time,
  total_duration_ms,
  execution_duration_ms,
  compilation_duration_ms,
  result_fetch_duration_ms,
  total_duration_ms - execution_duration_ms - compilation_duration_ms AS queue_and_overhead_ms,
  read_rows,
  produced_rows,
  read_bytes,
  spilled_local_bytes,
  compute.warehouse_id
FROM system.query.history
WHERE start_time >= CURRENT_DATE - INTERVAL 7 DAYS
  AND execution_status = 'FINISHED'
ORDER BY total_duration_ms DESC
LIMIT 20;

-- ============================================================================
-- 2. Analyze time breakdown for a specific query
-- ============================================================================
SELECT
  statement_id,
  total_duration_ms,
  compilation_duration_ms,
  execution_duration_ms,
  result_fetch_duration_ms,
  ROUND(compilation_duration_ms * 100.0 / total_duration_ms, 1) AS pct_compilation,
  ROUND(execution_duration_ms * 100.0 / total_duration_ms, 1) AS pct_execution,
  ROUND(result_fetch_duration_ms * 100.0 / total_duration_ms, 1) AS pct_fetch
FROM system.query.history
WHERE statement_id = '<STATEMENT_ID>';

-- ============================================================================
-- 3. Find queries with high spill (indicates memory pressure)
-- ============================================================================
SELECT
  statement_id,
  SUBSTRING(statement_text, 1, 200) AS query_preview,
  total_duration_ms,
  read_bytes,
  spilled_local_bytes,
  ROUND(spilled_local_bytes * 100.0 / NULLIF(read_bytes, 0), 1) AS spill_pct
FROM system.query.history
WHERE start_time >= CURRENT_DATE - INTERVAL 7 DAYS
  AND spilled_local_bytes > 0
ORDER BY spilled_local_bytes DESC
LIMIT 20;

-- ============================================================================
-- 4. Detect performance regression: compare avg duration over two periods
-- ============================================================================
SELECT
  SUBSTRING(statement_text, 1, 100) AS query_signature,
  COUNT(*) AS executions,
  AVG(CASE WHEN start_time >= CURRENT_DATE - INTERVAL 7 DAYS THEN total_duration_ms END) AS avg_ms_last_7d,
  AVG(CASE WHEN start_time BETWEEN CURRENT_DATE - INTERVAL 14 DAYS AND CURRENT_DATE - INTERVAL 7 DAYS THEN total_duration_ms END) AS avg_ms_prior_7d,
  ROUND(
    (AVG(CASE WHEN start_time >= CURRENT_DATE - INTERVAL 7 DAYS THEN total_duration_ms END) -
     AVG(CASE WHEN start_time BETWEEN CURRENT_DATE - INTERVAL 14 DAYS AND CURRENT_DATE - INTERVAL 7 DAYS THEN total_duration_ms END)) * 100.0 /
    NULLIF(AVG(CASE WHEN start_time BETWEEN CURRENT_DATE - INTERVAL 14 DAYS AND CURRENT_DATE - INTERVAL 7 DAYS THEN total_duration_ms END), 0),
    1
  ) AS pct_change
FROM system.query.history
WHERE start_time >= CURRENT_DATE - INTERVAL 14 DAYS
  AND execution_status = 'FINISHED'
GROUP BY SUBSTRING(statement_text, 1, 100)
HAVING COUNT(*) >= 5
ORDER BY pct_change DESC
LIMIT 20;
