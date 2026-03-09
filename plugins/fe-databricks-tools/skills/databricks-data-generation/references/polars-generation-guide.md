# Polars Generation Guide (Tier 1 Primary, Tier 2 Alternative)

Generate synthetic data locally with **Polars + NumPy + Mimesis** — zero JVM overhead, no Spark session required. NumPy vectorizes all randomness for ~5-10x speedup over list comprehensions. Ideal for quick prototyping, unit tests, and datasets under ~500K rows.

> Polars + NumPy + Mimesis is the **primary** engine for **Tier 1** (local parquet). For **Tier 2** (UC Delta tables via Connect), dbldatagen + Connect is primary — see [dbldatagen-connect-patterns.md](dbldatagen-connect-patterns.md). Polars is a valid **Tier 2 alternative** when you need Mimesis PII or NumPy statistical distributions that dbldatagen can't run over Connect.

## When to Use

- User wants **quick local data** without spinning up a Spark session
- Dataset is **<500K rows** (practical limit with NumPy vectorization)
- User is **offline** or doesn't have Databricks Connect configured
- User explicitly asks for **Polars**, **local**, or **offline** generation
- User needs data for **Polars-based** downstream analysis

**When NOT to use:** >500K rows — use **dbldatagen + Connect** (primary Tier 2, see [dbldatagen-connect-patterns.md](dbldatagen-connect-patterns.md)) or Polars + Connect as an alternative when Mimesis PII or NumPy distributions are needed. Need streaming or CDC with Volume landing zones → Tier 3 notebooks.

## Setup

```bash
# Minimal — no Spark, no JVM, no Databricks Connect
uv pip install polars numpy mimesis
```

## Core Patterns

### NumPy RNG Initialization

```python
import numpy as np
import polars as pl

rng = np.random.default_rng(42)  # Reproducible, vectorized RNG
```

All randomness flows through the `rng` object. Never use `random.seed()` or `random.choices()` — NumPy's Generator API provides vectorized equivalents for every distribution.

### Mimesis Pool-and-Sample for PII

```python
from mimesis import Generic
from mimesis.locales import Locale

g = Generic(locale=Locale.EN, seed=42)

# Generate ~1K unique values, then NumPy samples from the pool
pool = min(1_000, rows)
_first = np.array([g.person.first_name() for _ in range(pool)])
_last = np.array([g.person.last_name() for _ in range(pool)])

first_names = _first[rng.integers(0, pool, size=rows)]  # Instant for any row count
last_names = _last[rng.integers(0, pool, size=rows)]
```

The `seed=` parameter on `Generic` ensures reproducible Mimesis output. Pool size of 1,000 balances realism (1K unique names is plenty for demos) with speed (~10x fewer Mimesis calls).

### Weighted Random Choices

Replaces dbldatagen's `values=/weights=`:

```python
# Equivalent to: values=["Bronze","Silver","Gold","Platinum"], weights=[50,30,15,5]
_w = np.array([50, 30, 15, 5], dtype=np.float64)
tiers = rng.choice(
    ["Bronze", "Silver", "Gold", "Platinum"],
    size=rows, p=_w / _w.sum()
)
```

### Statistical Distributions

Replaces dbldatagen's `distribution=Gamma(...)`, `Beta(...)`, etc.:

```python
# Gamma — right-skewed positives (transaction amounts, account balances)
amounts = np.round(rng.gamma(2.0, 2.0, size=rows) * scale_factor, 2)

# Beta — bounded 0-1 values (discount rates, quality scores)
discount_pcts = rng.beta(2, 5, size=rows)

# Normal — symmetric data (sensor readings, lifespan estimates)
values = rng.normal(loc=15, scale=5, size=rows)

# Exponential — inter-arrival times, counts (clip to enforce min/max)
counts = np.clip(np.floor(rng.exponential(5.0, size=rows)).astype(int), 1, 20)
```

### Date and Timestamp Ranges

```python
# Random dates in range (day precision)
start = np.datetime64("2020-01-01")
span = (np.datetime64("2024-12-31") - start).astype(int)
dates = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[D]")

# Random timestamps in range (millisecond precision — Polars requires ms/us/ns)
ts_start = np.datetime64("2024-01-01")
ts_span = int((np.datetime64("2024-12-31T23:59:59") - ts_start) / np.timedelta64(1, "ms"))
timestamps = ts_start + rng.integers(0, ts_span + 1, size=rows).astype("timedelta64[ms]")
```

### Null Injection (Post-DataFrame)

Replaces dbldatagen's `percentNulls=`:

```python
# 5% nulls on email column — applied AFTER DataFrame construction
email_nulls = pl.Series(rng.random(rows) < 0.05)
df = df.with_columns(
    pl.when(email_nulls).then(None).otherwise(pl.col("email")).alias("email"),
)
```

### Foreign Key Consistency

Replaces dbldatagen's `minValue/maxValue` on FK columns:

```python
# Parent: customer_ids = np.arange(1_000_000, 1_000_000 + n_customers)
# Child: FK references into parent range
customer_fks = rng.integers(1_000_000, 1_000_000 + n_customers, size=rows)
```

### Derived Columns via Polars Expressions

Replaces dbldatagen's `expr=` for computed columns:

```python
df = pl.DataFrame({...})

# Computed column from other columns
df = df.with_columns(
    (pl.col("quantity") * pl.col("unit_price")).round(2).alias("line_total")
)

# Conditional/CASE expressions
df = df.with_columns(
    pl.when(pl.col("balance") > 5_000_000).then(pl.lit("Low"))
    .when(pl.col("balance") > 500_000).then(pl.lit("Medium"))
    .otherwise(pl.lit("High"))
    .alias("risk_rating")
)
```

### String Formatting via Polars

```python
# Generate int with NumPy, format with Polars, then drop temp column
df = pl.DataFrame({"_sku_num": rng.integers(10_000_000, 100_000_000, size=rows)})
df = df.with_columns(
    pl.format("SKU-{}", pl.col("_sku_num")).alias("sku")
).drop("_sku_num")
```

### Unique Sequential IDs

```python
# Equivalent to: minValue=1_000_000, uniqueValues=rows
customer_ids = np.arange(1_000_000, 1_000_000 + rows)
```

## Mimesis Provider Quick Reference

The `mimesisText("provider.method")` pattern used with dbldatagen maps directly to `g.provider.method()` calls in standalone Mimesis.

| dbldatagen (Spark) | Polars (standalone Mimesis) | Example Output |
|---|---|---|
| `text=mimesisText("person.first_name")` | `g.person.first_name()` | Jennifer |
| `text=mimesisText("person.last_name")` | `g.person.last_name()` | Williams |
| `text=mimesisText("person.email")` | `g.person.email()` | user@example.com |
| `text=mimesisText("person.telephone")` | `g.person.telephone()` | +1-555-123-4567 |
| `text=mimesisText("address.address")` | `g.address.address()` | 123 Main Street |
| `text=mimesisText("address.city")` | `g.address.city()` | Los Angeles |
| `text=mimesisText("address.state")` | `g.address.state()` | California |
| `text=mimesisText("finance.company")` | `g.finance.company()` | Acme Corp |

## Output Patterns

### Write Parquet

```python
import polars as pl

# Single file
df.write_parquet("output/retail/customers.parquet")

# Using the local_output utility
from utils.local_output import write_local_parquet
path = write_local_parquet(df, "customers", industry="retail")
```

### Write CSV

```python
df.write_csv("output/retail/customers.csv")
```

### Write to Unity Catalog (Connect Bridge)

When the user asks to **save to Unity Catalog** (even for small datasets), use the Connect bridge pattern. The generation code stays the same — only the output sink changes.

```python
from databricks.connect import DatabricksSession

spark = DatabricksSession.builder.serverless().getOrCreate()

CATALOG = "my_catalog"
SCHEMA = "retail"

# Ensure schema exists
existing = [row.databaseName for row in spark.sql(f"SHOW SCHEMAS IN {CATALOG}").collect()]
if SCHEMA not in existing:
    spark.sql(f"CREATE SCHEMA {CATALOG}.{SCHEMA}")

# Bridge: Polars → pandas → Spark DataFrame → UC Delta table
spark_df = spark.createDataFrame(df.to_pandas())
(spark_df.write.format("delta")
 .mode("overwrite")
 .option("overwriteSchema", "true")
 .saveAsTable(f"{CATALOG}.{SCHEMA}.customers"))

# Validate
print(f"Wrote {spark.table(f'{CATALOG}.{SCHEMA}.customers').count():,} rows")
```

Run: `uv run --with polars --with numpy --with mimesis --with "databricks-connect>=16.4,<17.0" script.py`

### Directory Convention (Local Output)

```
output/
  retail/
    customers.parquet
    products.parquet
    transactions.parquet
  healthcare/
    patients.parquet
    encounters.parquet
```

## Reading Data

```python
import polars as pl

# Single file
customers = pl.read_parquet("output/retail/customers.parquet")

# Multiple files (glob)
all_data = pl.read_parquet("output/retail/*.parquet")

# Quick inspection
print(customers.shape)
print(customers.schema)
customers.head()
customers.describe()
```

## Performance Notes

- NumPy vectorization makes Tier 1 practical up to **~500K rows per table**
- Mimesis pool-and-sample generates ~1K unique values, then NumPy samples at array speed (~10x faster than per-row calls)
- For 100K rows: ~0.1s for NumPy numeric columns, ~1s for Mimesis pool creation
- For 500K rows: ~0.5s for NumPy numerics, ~1s for Mimesis pool (same pool, more sampling)
- Parquet write is near-instant for datasets under 500K rows
- If you need >500K rows, switch to **Tier 2** (**dbldatagen + Connect** as primary, or Polars + Connect as alternative) or **Tier 3** (dbldatagen in notebooks)

## dbldatagen Equivalence Table

| dbldatagen Feature | Polars + NumPy + Mimesis Equivalent |
|---|---|
| `values=["A","B"], weights=[70,30]` | `rng.choice(["A","B"], size=rows, p=np.array([70,30])/100)` |
| `minValue=1, maxValue=1000, random=True` | `rng.integers(1, 1001, size=rows)` |
| `distribution=Gamma(shape, scale)` | `rng.gamma(shape, scale, size=rows) * factor` |
| `distribution=Beta(alpha, beta)` | `rng.beta(alpha, beta, size=rows)` |
| `distribution=Normal(mean, std)` | `rng.normal(mean, std, size=rows)` |
| `distribution=Exponential()` | `rng.exponential(scale, size=rows)` |
| `text=mimesisText("person.first_name")` | `_pool[rng.integers(0, pool, size=rows)]` (pool-and-sample) |
| `template=r"ddddd"` | `rng.integers(10000, 100000, size=rows).astype(str)` |
| `begin="2020-01-01", end="2024-12-31"` | `start + rng.integers(0, span+1, size=rows).astype("timedelta64[D]")` |
| `percentNulls=0.05` | `pl.when(pl.Series(rng.random(rows)<0.05)).then(None).otherwise(col)` |
| `omit=True` | Compute intermediate as `_temp` column, then `.drop("_temp")` |
| `expr="col_a * col_b"` | `df.with_columns((pl.col("a") * pl.col("b")).alias("c"))` |
| `.withConstraint(PositiveValues("x"))` | `df.filter(pl.col("x") > 0)` |
| `.withConstraint(SqlExpr("a <= b"))` | `df.filter(pl.col("a") <= pl.col("b"))` |
| `uniqueValues=rows` | `np.arange(start, start + rows)` |
