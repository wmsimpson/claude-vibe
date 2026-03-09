# Oil & Gas Industry Patterns

Comprehensive data models and generation patterns for upstream oil & gas production analytics in the Texas Permian Basin.

## Data Model Overview

```
                        ┌──────────────────────┐
                        │  TYPE_CURVE_PARAMS    │
                        │  (per formation)      │
                        │  q_i, d, b            │
                        └──────────┬───────────┘
                                   │
                    ┌──────────────┼──────────────┐
                    ▼                              ▼
         ┌──────────────────┐           ┌──────────────────┐
         │   Well Headers   │           │   Type Curves    │
         │  (50 per form.)  │           │ (2000 per form.) │
         │   ~200 total     │           │  ~8000 total     │
         └────────┬─────────┘           └──────────────────┘
                  │
                  ▼
         ┌──────────────────┐
         │ Daily Production │
         │ (100-700 days    │
         │  per well)       │
         └──────────────────┘
```

**Generation Strategy:** Per-formation generation with union. Each formation has its own
ARPS decline curve parameters. Well headers and type curves are generated per formation
then unioned. Daily production is generated per well using the well's ARPS parameters.

## ARPS Decline Curve Formula

The [ARPS hyperbolic decline curve](https://petrowiki.spe.org/Production_decline_curve_analysis) is the
industry-standard method for forecasting oil well production:

```
q(t) = q_i / (1 + b * d * t)^(1/b)
```

Where:
- **q(t)** — production rate at time t (BOPD: barrels of oil per day)
- **q_i** — initial production rate (BOPD)
- **d** — initial decline rate (fraction/day)
- **b** — hyperbolic exponent (0 < b < 1 for hyperbolic decline)
- **t** — time from first production (days)

## Formation Parameters

| Formation | q_i (BOPD) | d (decline) | b (exponent) | Character |
|-----------|-----------|-------------|--------------|-----------|
| FORMATION_A | 6,000 | 0.010 | 0.8 | Moderate IP, slow decline |
| FORMATION_B | 7,000 | 0.011 | 0.7 | High IP, steeper decline |
| FORMATION_C | 5,500 | 0.009 | 0.8 | Lower IP, slowest decline |
| FORMATION_D | 5,750 | 0.011 | 0.7 | Mid IP, steeper decline |

## Table Schemas

### Well Headers

50 wells per formation, ~200 total. Each well belongs to a single producing formation
and carries its own ARPS parameters for downstream production generation.

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `API_NUMBER` | BIGINT | API well number | Random in 42T range |
| `FIELD_NAME` | STRING | Producing field | 5 values, random |
| `LATITUDE` | FLOAT | Surface latitude | 31.00 - 32.50 |
| `LONGITUDE` | FLOAT | Surface longitude | -104.00 to -101.00 |
| `COUNTY` | STRING | Texas county | Reeves/Midland/Ector/Loving/Ward |
| `STATE` | STRING | State | Fixed: Texas |
| `COUNTRY` | STRING | Country | Fixed: USA |
| `WELL_TYPE` | STRING | Well type | Fixed: Oil |
| `WELL_ORIENTATION` | STRING | Drilling orientation | Fixed: Horizontal |
| `PRODUCING_FORMATION` | STRING | Target formation | Fixed per formation loop |
| `CURRENT_STATUS` | STRING | Well status | Weighted 80/10/5/5 |
| `TOTAL_DEPTH` | INTEGER | Total measured depth (ft) | 12,000 - 20,000 |
| `SPUD_DATE` | DATE | Spud date | 2020-01-01 to 2025-02-14 |
| `COMPLETION_DATE` | DATE | Completion date | 2020-01-01 to 2025-02-14 |
| `SURFACE_CASING_DEPTH` | INTEGER | Surface casing depth (ft) | 500 - 800 |
| `OPERATOR_NAME` | STRING | Operator | Fixed: OPERATOR_XYZ |
| `PERMIT_DATE` | DATE | Permit date | 2019-01-01 to 2025-02-14 |
| `q_i` | DOUBLE | ARPS initial rate | Fixed per formation |
| `d` | DOUBLE | ARPS decline rate | Fixed per formation |
| `b` | DOUBLE | ARPS b-factor | Fixed per formation |

```python
import dbldatagen as dg

TYPE_CURVE_PARAMS = {
    "FORMATION_A": {"q_i": 6000, "d": 0.01, "b": 0.8},
    "FORMATION_B": {"q_i": 7000, "d": 0.011, "b": 0.7},
    "FORMATION_C": {"q_i": 5500, "d": 0.009, "b": 0.8},
    "FORMATION_D": {"q_i": 5750, "d": 0.011, "b": 0.7},
}

# Generate per formation, then union
dfs = []
for i, (formation, params) in enumerate(TYPE_CURVE_PARAMS.items()):
    spec = (
        dg.DataGenerator(spark, rows=50, partitions=2, randomSeed=42 + i)
        .withColumn("API_NUMBER", "bigint",
                    minValue=42000000000000, maxValue=42999999999999, random=True)
        .withColumn("FIELD_NAME", "string",
                    values=["Field_1", "Field_2", "Field_3", "Field_4", "Field_5"],
                    random=True)
        .withColumn("LATITUDE", "float",
                    minValue=31.00, maxValue=32.50, step=1e-6, random=True)
        .withColumn("LONGITUDE", "float",
                    minValue=-104.00, maxValue=-101.00, step=1e-6, random=True)
        .withColumn("COUNTY", "string",
                    values=["Reeves", "Midland", "Ector", "Loving", "Ward"],
                    random=True)
        .withColumn("STATE", "string", values=["Texas"])
        .withColumn("COUNTRY", "string", values=["USA"])
        .withColumn("WELL_TYPE", "string", values=["Oil"])
        .withColumn("WELL_ORIENTATION", "string", values=["Horizontal"])
        .withColumn("PRODUCING_FORMATION", "string", values=[formation])
        .withColumn("CURRENT_STATUS", "string",
                    values=["Producing", "Shut-in", "Plugged and Abandoned", "Planned"],
                    weights=[80, 10, 5, 5], random=True)
        .withColumn("TOTAL_DEPTH", "integer",
                    minValue=12000, maxValue=20000, random=True)
        .withColumn("SPUD_DATE", "date",
                    begin="2020-01-01", end="2025-02-14", random=True)
        .withColumn("COMPLETION_DATE", "date",
                    begin="2020-01-01", end="2025-02-14", random=True)
        .withColumn("SURFACE_CASING_DEPTH", "integer",
                    minValue=500, maxValue=800, random=True)
        .withColumn("OPERATOR_NAME", "string", values=["OPERATOR_XYZ"])
        .withColumn("PERMIT_DATE", "date",
                    begin="2019-01-01", end="2025-02-14", random=True)
        .withColumn("q_i", "double", values=[params["q_i"]])
        .withColumn("d", "double", values=[params["d"]])
        .withColumn("b", "double", values=[params["b"]])
        .build()
    )
    dfs.append(spec)

wells_df = dfs[0]
for df in dfs[1:]:
    wells_df = wells_df.unionByName(df)
```

### Daily Production

100-700 days per well. Each row represents a single day of production for a single well.
The ARPS formula calculates actual production with random variation and occasional shut-in events.

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `well_num` | BIGINT | Well API number | Fixed per well loop |
| `day_from_first_production` | INTEGER | Days since first production | Sequential 1-N |
| `first_production_date` | DATE | First production date | Computed from days |
| `date` | DATE | Calendar date | `date_add(first_production_date, day)` |
| `q_i` | DOUBLE | ARPS initial rate | From well header |
| `d` | DOUBLE | ARPS decline rate | From well header |
| `b` | DOUBLE | ARPS b-factor | From well header |
| `q_i_multiplier` | DOUBLE | Shut-in multiplier | Weighted 97/3 (1.0 vs 0) |
| `variation` | DOUBLE | Production noise | `rand() * 0.1 + 0.95` |
| `actuals_bopd` | DOUBLE | Actual oil rate (BOPD) | ARPS formula |

**ARPS Production Formula:**
```
actuals_bopd = (q_i * q_i_multiplier) / power(1 + b * d * variation * day_from_first_production, 1/b)
```

```python
import dbldatagen as dg
from datetime import date, timedelta
import random

def generate_daily_production(spark, well_num, q_i, d, b, q_i_multiplier=1.0,
                              partitions=4, seed=None):
    days_to_generate = int(round(random.uniform(100, 700)))
    if seed is None:
        seed = int(round(random.uniform(20, 1000)))

    return (
        dg.DataGenerator(spark, rows=days_to_generate, partitions=partitions,
                        randomSeed=seed, name="type_curve")
        .withColumn("well_num", "bigint", values=[well_num])
        .withColumn("day_from_first_production", "integer",
                    minValue=1, maxValue=1000)
        .withColumn("first_production_date", "date",
                    values=[date.today() - timedelta(days=days_to_generate)])
        .withColumn("date", "date",
                    expr="date_add(first_production_date, day_from_first_production)")
        .withColumn("q_i", "double", values=[q_i])
        .withColumn("d", "double", values=[d])
        .withColumn("b", "double", values=[b])
        .withColumn("q_i_multiplier", "double",
                    values=[q_i_multiplier, 0], weights=[97, 3], random=True)
        .withColumn("variation", "double", expr="rand() * 0.1 + 0.95")
        .withColumn("actuals_bopd", "double",
                    baseColumn=["q_i", "d", "b", "q_i_multiplier", "variation"],
                    expr="(q_i * q_i_multiplier) / power(1 + b * d * variation * day_from_first_production, 1/b)")
        .build()
    )

# Iterate over wells from wells_df
wells_dict = wells_df.toPandas().to_dict()
prod_dfs = [
    generate_daily_production(
        spark,
        wells_dict["API_NUMBER"][i],
        wells_dict["q_i"][i],
        wells_dict["d"][i],
        wells_dict["b"][i],
        1.0,
    )
    for i in range(len(next(iter(wells_dict.values()))))
]
daily_production_df = prod_dfs[0]
for df in prod_dfs[1:]:
    daily_production_df = daily_production_df.unionByName(df)
```

### Type Curves

2,000 forecast days per formation. Type curves represent the expected production profile
for a formation, used as benchmarks for comparing actual well performance.

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `formation` | STRING | Formation name | Fixed per formation loop |
| `day_from_first_production` | INTEGER | Forecast day | Sequential 1-1000 |
| `q_i` | DOUBLE | ARPS initial rate | From formation params |
| `d` | DOUBLE | ARPS decline rate | From formation params |
| `b` | DOUBLE | ARPS b-factor | From formation params |
| `variation` | DOUBLE | Forecast noise | `rand() * 0.1 + 0.95` |
| `forecast_bopd` | DOUBLE | Forecasted rate (BOPD) | ARPS formula |

**ARPS Forecast Formula:**
```
forecast_bopd = q_i / power(1 + b * d * day_from_first_production, 1/b)
```

```python
import dbldatagen as dg

def generate_type_curve(spark, formation, q_i, d, b, partitions=4, seed=None):
    if seed is None:
        seed = int(round(random.uniform(20, 1000)))

    return (
        dg.DataGenerator(spark, rows=2000, partitions=partitions,
                        randomSeed=seed, name="type_curve")
        .withColumn("formation", "string", values=[formation])
        .withColumn("day_from_first_production", "integer",
                    minValue=1, maxValue=1000)
        .withColumn("q_i", "double", values=[q_i])
        .withColumn("d", "double", values=[d])
        .withColumn("b", "double", values=[b])
        .withColumn("variation", "double", expr="rand() * 0.1 + 0.95")
        .withColumn("forecast_bopd", "double",
                    baseColumn=["q_i", "d", "b", "day_from_first_production"],
                    expr="q_i / power(1 + b * d * day_from_first_production, 1/b)")
        .build()
    )

# Generate per formation, then union
tc_dfs = [
    generate_type_curve(spark, formation, params["q_i"], params["d"], params["b"])
    for formation, params in TYPE_CURVE_PARAMS.items()
]
type_curve_df = tc_dfs[0]
for df in tc_dfs[1:]:
    type_curve_df = type_curve_df.unionByName(df)
```

## Unique Generation Pattern

This industry differs from others because it uses **per-formation generation + union**
rather than a single `DataGenerator` call for all rows:

1. **Well Headers** — Loop over `TYPE_CURVE_PARAMS`, create one `DataGenerator` per formation
   with formation-specific ARPS parameters baked in, then `unionByName` all DataFrames.
2. **Daily Production** — Convert well headers to a dict, loop over each well, generate
   100-700 days of production per well using that well's ARPS parameters, then union.
3. **Type Curves** — Loop over formations, generate 2,000 forecast days per formation,
   then union.

This pattern ensures each formation/well gets its own deterministic ARPS parameters
rather than using random distributions across a single generator.

## Geographic Context: Texas Permian Basin

All generated wells are located in the **Permian Basin** of West Texas:
- **Latitude:** 31.00 to 32.50 (southern Permian Basin)
- **Longitude:** -104.00 to -101.00
- **Counties:** Reeves, Midland, Ector, Loving, Ward
- **State:** Texas
- **Well Type:** Horizontal oil wells (typical of Permian unconventional plays)

## Common Demo Queries

### Decline Curve Analysis
```sql
SELECT
    well_num,
    day_from_first_production,
    actuals_bopd,
    q_i / power(1 + b * d * day_from_first_production, 1/b) AS forecast_bopd,
    actuals_bopd - (q_i / power(1 + b * d * day_from_first_production, 1/b)) AS variance
FROM daily_production
WHERE q_i_multiplier > 0
ORDER BY well_num, day_from_first_production
```

### Formation Comparison
```sql
SELECT
    wh.PRODUCING_FORMATION,
    COUNT(DISTINCT dp.well_num) AS well_count,
    AVG(dp.actuals_bopd) AS avg_bopd,
    MAX(dp.actuals_bopd) AS peak_bopd,
    MIN(dp.actuals_bopd) AS min_bopd
FROM daily_production dp
JOIN well_headers wh ON dp.well_num = wh.API_NUMBER
WHERE dp.q_i_multiplier > 0
GROUP BY wh.PRODUCING_FORMATION
ORDER BY avg_bopd DESC
```

### Well Ranking by Cumulative Production
```sql
SELECT
    dp.well_num,
    wh.PRODUCING_FORMATION,
    wh.COUNTY,
    SUM(dp.actuals_bopd) AS cumulative_oil_bbl,
    COUNT(*) AS producing_days,
    MAX(dp.actuals_bopd) AS peak_rate_bopd
FROM daily_production dp
JOIN well_headers wh ON dp.well_num = wh.API_NUMBER
WHERE dp.q_i_multiplier > 0
GROUP BY dp.well_num, wh.PRODUCING_FORMATION, wh.COUNTY
ORDER BY cumulative_oil_bbl DESC
LIMIT 20
```

### Actual vs Type Curve Performance
```sql
SELECT
    wh.PRODUCING_FORMATION AS formation,
    dp.day_from_first_production,
    AVG(dp.actuals_bopd) AS avg_actual_bopd,
    tc.forecast_bopd,
    AVG(dp.actuals_bopd) - tc.forecast_bopd AS performance_delta
FROM daily_production dp
JOIN well_headers wh ON dp.well_num = wh.API_NUMBER
JOIN type_curves tc
    ON wh.PRODUCING_FORMATION = tc.formation
    AND dp.day_from_first_production = tc.day_from_first_production
WHERE dp.q_i_multiplier > 0
GROUP BY wh.PRODUCING_FORMATION, dp.day_from_first_production, tc.forecast_bopd
ORDER BY formation, dp.day_from_first_production
```

### Shut-in Analysis
```sql
SELECT
    wh.PRODUCING_FORMATION,
    COUNT(CASE WHEN dp.q_i_multiplier = 0 THEN 1 END) AS shutin_days,
    COUNT(*) AS total_days,
    ROUND(100.0 * COUNT(CASE WHEN dp.q_i_multiplier = 0 THEN 1 END) / COUNT(*), 2) AS shutin_pct
FROM daily_production dp
JOIN well_headers wh ON dp.well_num = wh.API_NUMBER
GROUP BY wh.PRODUCING_FORMATION
ORDER BY shutin_pct DESC
```
