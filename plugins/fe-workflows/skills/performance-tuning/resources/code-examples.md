# Performance Tuning Code Examples

## Data Skew: Salting Technique

When AQE isn't sufficient for extreme skew on hot keys:

### PySpark Salting

```python
from pyspark.sql.functions import concat, lit, col, rand, floor, explode, sequence

num_salts = 10  # Spread skewed key across 10 partitions

# Salt the skewed (large) table
large_df = large_df.withColumn("salt", floor(rand() * num_salts).cast("int"))
large_df = large_df.withColumn("salted_key", concat(col("join_key"), lit("_"), col("salt")))

# Explode the small table to match all salts
small_df = small_df.withColumn("salt", explode(sequence(lit(0), lit(num_salts - 1))))
small_df = small_df.withColumn("salted_key", concat(col("join_key"), lit("_"), col("salt")))

# Join on the salted key
result = large_df.join(small_df, "salted_key")
```

### SQL Salting

```sql
WITH salted_large AS (
  SELECT *, FLOOR(RAND() * 10) AS salt,
         CONCAT(join_key, '_', FLOOR(RAND() * 10)) AS salted_key
  FROM large_table
),
salted_small AS (
  SELECT *, s.salt,
         CONCAT(join_key, '_', s.salt) AS salted_key
  FROM small_table
  CROSS JOIN (SELECT EXPLODE(SEQUENCE(0, 9)) AS salt) s
)
SELECT * FROM salted_large JOIN salted_small USING (salted_key);
```

## Photon: Replacing UDFs with Built-in Functions

```python
# BAD: Python UDF (no Photon, no codegen)
@udf(StringType())
def my_upper(s):
    return s.upper() if s else None
df = df.withColumn("name", my_upper(col("name")))

# GOOD: Built-in function (Photon-optimized)
from pyspark.sql.functions import upper
df = df.withColumn("name", upper(col("name")))
```

## Shuffle: Avoiding Unnecessary Shuffles

```python
# BAD: repartition before write (unnecessary shuffle)
df.repartition(100).write.format("delta").save(path)

# GOOD: use coalesce (no shuffle) or rely on optimized writes
df.coalesce(100).write.format("delta").save(path)

# BEST: let Databricks handle it with optimized writes
df.write.option("optimizeWrite", "true").format("delta").save(path)
```

## Structured Streaming Configuration

### RocksDB State Store

```python
spark.conf.set(
    "spark.sql.streaming.stateStore.providerClass",
    "com.databricks.sql.streaming.state.RocksDBStateStoreProvider"
)
```

### Changelog Checkpointing (DBR 13.3+)

```python
spark.conf.set(
    "spark.sql.streaming.stateStore.rocksdb.changelogCheckpointing.enabled",
    "true"
)
```

### Asynchronous Checkpointing

```python
spark.conf.set(
    "spark.databricks.streaming.statefulOperator.asyncCheckpoint.enabled",
    "true"
)
```

### Trigger Interval Options

```python
# Process as fast as possible (default)
query.trigger(processingTime="0 seconds")

# Fixed interval (good for batch-like streaming)
query.trigger(processingTime="30 seconds")

# Process all available data once (good for scheduled streaming)
query.trigger(availableNow=True)
```
