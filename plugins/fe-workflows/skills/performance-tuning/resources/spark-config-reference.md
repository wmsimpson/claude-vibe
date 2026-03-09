# Spark Configuration Quick Reference

## Memory and Execution

| Config | Default | Description | When to Tune |
|---|---|---|---|
| `spark.executor.memory` | Varies by instance | Executor heap memory | Spill problems |
| `spark.executor.memoryOverhead` | 10% or 384MB | Off-heap memory | Native memory issues |
| `spark.driver.memory` | Varies | Driver heap memory | Driver OOM on collect/broadcast |
| `spark.sql.shuffle.partitions` | 200 (or auto) | Shuffle output partitions | **Set to `auto`** for AQE |
| `spark.sql.adaptive.enabled` | true | Enable AQE | Leave enabled |

## Join and Broadcast

| Config | Default | Description | When to Tune |
|---|---|---|---|
| `spark.sql.autoBroadcastJoinThreshold` | 10MB | Max table size for broadcast | Increase for larger dim tables |
| `spark.sql.adaptive.autoBroadcastJoinThreshold` | Same | AQE runtime broadcast threshold | Increase for runtime conversion |
| `spark.sql.adaptive.skewJoin.skewedPartitionFactor` | 5 | Skew detection multiplier | Lower for more aggressive detection |
| `spark.sql.adaptive.skewJoin.skewedPartitionThresholdInBytes` | 256MB | Min size to consider skewed | Lower for smaller datasets |

## Delta / Storage

| Config | Default | Description | When to Tune |
|---|---|---|---|
| `spark.databricks.delta.optimizeWrite.enabled` | false | Coalesce partitions on write | Enable for better file sizes |
| `spark.databricks.delta.autoCompact.enabled` | false | Auto-compact after writes | Enable for streaming / frequent writes |
| `spark.databricks.delta.merge.enableLowShuffle` | false | Reduce shuffle in MERGE | Enable for MERGE-heavy workloads |
| `delta.tuneFileSizesForRewrites` | false | Smaller files for DML tables | Enable for MERGE/UPDATE-heavy tables |

## Streaming

| Config | Default | Description | When to Tune |
|---|---|---|---|
| `spark.sql.streaming.stateStore.providerClass` | default | State store implementation | Set to RocksDB for stateful queries |
| `spark.sql.streaming.stateStore.rocksdb.changelogCheckpointing.enabled` | false | Changelog-based checkpointing | Enable on DBR 13.3+ |
| `spark.databricks.streaming.statefulOperator.asyncCheckpoint.enabled` | false | Async checkpointing | Enable for checkpoint-bound queries |
