# databricks-data-generation

A **Claude Code skill** that generates realistic synthetic data for Databricks demos, prototyping, and testing — from 500-row quick prototypes to billion-row production datasets.

## Why This Exists

Building Databricks demos and proofs-of-concept requires realistic data, but getting it is painful:

- **Production data is off-limits** — PII, compliance, and access restrictions make real data unusable for demos
- **Hand-crafted CSVs don't scale** — manually building test data is tedious and never looks realistic
- **Generic random data fails to convince** — uniform distributions, placeholder names like "John Doe", and missing referential integrity break the illusion instantly
- **Setting up data generation tooling is a project in itself** — learning dbldatagen, configuring Spark, wiring up Unity Catalog, handling distributions and constraints

This skill solves all of that. You describe the data you need in plain English, and Claude generates a complete, self-contained script with realistic PII (via [Mimesis](https://mimesis.name/)), proper statistical distributions (Gamma, Beta, Normal), referential integrity across tables, and correct Databricks output targeting — whether that's local parquet files or Unity Catalog Delta tables.

## What This Is (and Isn't)

**This is a [Claude Code skill](https://docs.anthropic.com/en/docs/claude-code)** — not a Python package, CLI tool, or library you install. There is no `pyproject.toml`, no `pip install`, and no importable modules.

You use it by asking Claude to generate synthetic data during a Claude Code session. Claude reads `SKILL.md` and the supporting reference files, then generates **inline, self-contained Python code** tailored to your exact request. The generated code runs in Databricks notebooks or locally via Databricks Connect.

## How to Use

Open a Claude Code session and describe the data you need:

```
> Generate 50K rows of retail customer data with realistic names and loyalty tiers

> Create a healthcare dataset with patients, encounters, and claims for a demo

> Generate streaming IoT sensor data for a Spark Declarative Pipeline

> Introspect prod.sales.customers and generate matching synthetic data

> Generate manufacturing sensor data with 5% anomaly rate and save to Unity Catalog
```

You can also invoke it directly as a skill:

```
> /databricks-data-generation retail --rows 50000
> /databricks-data-generation healthcare --rows 10000 --catalog demo
> /databricks-data-generation iot --tier 3
```

### What You Get

Claude generates a complete Python script that:

1. Creates all related tables for your industry (e.g., retail = customers + products + transactions + line_items + inventory)
2. Maintains foreign key relationships across tables (every `customer_id` in transactions exists in customers)
3. Uses realistic PII — actual-looking names, emails, phone numbers, addresses via Mimesis
4. Applies proper statistical distributions — Gamma-distributed account balances, Beta-distributed discount rates, weighted categorical tiers
5. Writes output to your chosen destination — local parquet, Unity Catalog Delta tables, or streaming sinks
6. Is self-contained and runnable — just copy, paste, and execute

## Three-Tier Architecture

The skill automatically selects the right generation engine based on your data volume and output requirements:

```
 Volume            Engine                          Output
 ──────            ──────                          ──────
 <500K rows   ──→  Polars + NumPy + Mimesis    ──→  Local parquet files
                                                    OR Unity Catalog (via Connect bridge)

 500K-5M rows ──→  dbldatagen + Connect        ──→  Unity Catalog Delta tables
                   (or Polars + NumPy + Mimesis
                    alternative when Mimesis PII needed)

 >5M rows     ──→  dbldatagen in notebooks     ──→  Unity Catalog Delta tables
                   (full feature set)
```

### Tier 1: Polars + NumPy + Mimesis (Local or UC)

**Best for:** Quick prototyping, offline work, small datasets, unit tests

- Zero JVM overhead — no Spark session, no cluster, runs in seconds
- **NumPy-vectorized randomness** — `np.random.default_rng(seed)` for all distributions and sampling
- **Pool-and-sample PII** — Mimesis generates ~1K unique values, then NumPy samples at array speed (100x faster than per-row Mimesis calls)
- NumPy statistical distributions — `rng.gamma()`, `rng.beta()`, `rng.normal()`
- Output: local parquet files (`./output/{industry}/`) **or** Unity Catalog via Connect bridge
- Practical limit: ~500K rows (vectorized NumPy, not single-threaded Python loops)

**When does Tier 1 write to UC?** When you say "save to Unity Catalog", "write to UC", provide `--catalog`, or otherwise indicate you want the data in Databricks. The skill generates with Polars locally, then bridges to UC via `spark.createDataFrame(df.to_pandas()).write.saveAsTable()`.

### Tier 2: dbldatagen + Databricks Connect (Primary)

**Best for:** Medium-scale datasets that need to land in Unity Catalog

- Runs from your local IDE via Databricks Connect + serverless compute
- dbldatagen's declarative API — `values=`, `weights=`, `minValue`/`maxValue`, `expr=`, `begin`/`end`
- All Catalyst-safe features work over Connect (no UDFs shipped to workers)
- Constraint workaround: `.build().filter("condition")` instead of `.withConstraint()`

**Alternative Tier 2:** When you need rich Mimesis PII or NumPy statistical distributions that dbldatagen can't run over Connect, the skill falls back to Polars + NumPy + Mimesis generation with a Connect write bridge.

### Tier 3: dbldatagen in Notebooks (Full Feature Set)

**Best for:** Large-scale generation, streaming, or when you need UDF-dependent features

- Full dbldatagen feature set — distributions (`Gamma`, `Beta`, `Normal`), `mimesisText()`, `.withConstraint()`, `template=`
- Streaming support via `withStreaming=True`
- Scale: millions to billions of rows with distributed Spark execution
- Libraries are pre-installed on Databricks Runtime — no dependency management

### Decision Tree

| Scenario                                  | Tier       | Engine                            |
| ----------------------------------------- | ---------- | --------------------------------- |
| Quick local prototype, no UC needed       | **1**      | Polars + NumPy + Mimesis → parquet        |
| Small dataset, save to Unity Catalog      | **1 + UC** | Polars + NumPy + Mimesis → Connect bridge |
| Medium dataset (500K-5M), UC target       | **2**      | dbldatagen + Connect (primary)            |
| Need Mimesis PII at scale for UC          | **2 alt**  | Polars + NumPy + Mimesis → Connect write  |
| Large dataset (>5M rows)                  | **3**      | dbldatagen in notebook            |
| Streaming or CDC with Volume landing      | **3**      | dbldatagen in notebook            |
| UDF features (distributions, mimesisText) | **3**      | dbldatagen in notebook            |
| User explicitly asks for Polars/local     | **1**      | Polars + NumPy + Mimesis          |
| User explicitly asks for notebook         | **3**      | dbldatagen in notebook            |

## Supported Industries

Each industry comes with a full set of interrelated tables, pre-defined schemas, and realistic data patterns:

### Retail

| Table          | Key Columns                                                    | Typical Size |
| -------------- | -------------------------------------------------------------- | ------------ |
| `customers`    | customer_id, first_name, last_name, email, loyalty_tier, state | 100K-1M      |
| `products`     | product_id, product_name, category, unit_price, sku            | 10K-100K     |
| `transactions` | txn_id, customer_id, store_id, total_amount, txn_timestamp     | 1M-100M      |
| `line_items`   | line_item_id, txn_id, product_id, quantity, line_total         | 3M-300M      |
| `inventory`    | inventory_id, product_id, store_id, quantity_on_hand           | 100K-1M      |

### Healthcare

| Table        | Key Columns                                           | Notes                    |
| ------------ | ----------------------------------------------------- | ------------------------ |
| `patients`   | patient_id, name, dob, mrn, insurance_id              | HIPAA-safe synthetic PII |
| `encounters` | encounter_id, patient_id, provider_id, admit_datetime | Clinical visits          |
| `claims`     | claim_id, encounter_id, icd10_code, billed_amount     | Insurance claims         |

### Financial

| Table          | Key Columns                                         | Use Case                         |
| -------------- | --------------------------------------------------- | -------------------------------- |
| `accounts`     | account_id, customer_id, type, balance, risk_rating | Gamma-distributed balances       |
| `trades`       | trade_id, account_id, symbol, quantity, price       | Trading activity                 |
| `transactions` | txn_id, account_id, type, amount, timestamp         | Deposits, withdrawals, transfers |

### IoT

| Table             | Key Columns                                     | Pattern                        |
| ----------------- | ----------------------------------------------- | ------------------------------ |
| `devices`         | device_id, device_type, location, install_date  | Device registry                |
| `sensor_readings` | device_id, timestamp, metric, value, is_anomaly | Sinusoidal + anomaly injection |
| `events`          | event_id, device_id, event_type, severity       | Alerts and notifications       |
| `telemetry`       | device_id, timestamp, lat, lon, speed           | GPS/telematics                 |

### Manufacturing

| Table                 | Key Columns                                          | Pattern                           |
| --------------------- | ---------------------------------------------------- | --------------------------------- |
| `equipment`           | equipment_id, type, manufacturer, zone, install_date | Asset registry                    |
| `sensor_data`         | equipment_id, timestamp, sensors A-F, is_anomaly     | Multi-sensor with fault injection |
| `maintenance_records` | work_order_id, equipment_id, type, priority, cost    | Predictive maintenance            |

### Supply Chain (CPG)

| Table                  | Key Columns                                                 | Typical Size | Pattern                     |
| ---------------------- | ----------------------------------------------------------- | ------------ | --------------------------- |
| `products`             | id, sku, category, brand, unit_cost, unit_price             | 500          | CPG product master          |
| `distribution_centers` | id, dc_code, region, capacity_pallets, utilization_pct      | 25           | DC network                  |
| `stores`               | id, store_code, format, distribution_center_id              | 1K           | Retail locations            |
| `orders`               | id, dc_id, product_id, scheduled_date, actual_date, status  | 10K          | Manufacturing/purchase orders |
| `inventory_snapshots`  | id, product_id, dc_id, store_id, quantity, stockout_risk    | 50K          | Point-in-time inventory     |
| `shipments`            | id, dc_id, store_id, product_id, transport_mode, carrier    | 30K          | Logistics & delivery tracking |

### Oil & Gas

| Table              | Key Columns                                             | Pattern                          |
| ------------------ | ------------------------------------------------------- | -------------------------------- |
| `well_headers`     | well_id, api_number, operator, formation, spud_date     | Well metadata (per-formation)    |
| `daily_production` | well_id, production_date, oil_bbl, gas_mcf, water_bbl  | ARPS decline curve production    |
| `type_curves`      | formation, month_on_production, oil_rate, gas_rate      | Forecast by formation            |

### Gaming

| Table          | Key Columns                                                     | Pattern                                          |
| -------------- | --------------------------------------------------------------- | ------------------------------------------------ |
| `login_events` | account_id, device_id, ip_address, country, login_ts, platform  | Hash-based IDs, country-weighted (Tier 3 only)   |

### Clinical Trials (Healthcare Variant)

| Table                | Key Columns                                                 | Pattern                      |
| -------------------- | ----------------------------------------------------------- | ---------------------------- |
| `clinical_trials`    | trial_id, phase, therapeutic_area, sponsor, status          | Trial protocol metadata      |
| `study_sites`        | site_id, trial_id, pi_name, institution                     | Investigator site network    |
| `study_participants` | participant_id, site_id, arm, consent_date, severity        | Randomized participants      |
| `adverse_events`     | ae_id, participant_id, term, severity, outcome              | Correlated AE patterns       |
| `lab_measurements`   | lab_id, participant_id, test_name, value, unit              | Treatment-effect lab results |

## Data Realism Features

What makes generated data look and behave like production data:

### Statistical Distributions

- **Gamma** — account balances, lifetime values (right-skewed, realistic wealth distribution)
- **Beta** — discount percentages, quality scores (bounded 0-1, asymmetric)
- **Normal/Gaussian** — sensor readings, temperatures (centered with natural variation)
- **Weighted categorical** — loyalty tiers (50% Bronze, 30% Silver, 15% Gold, 5% Platinum)

### Referential Integrity

- Foreign keys are always valid — `customer_id` in transactions always references an existing customer
- Cardinality is configurable — 1:N, N:M relationships between parent and child tables
- ID ranges match across tables automatically

### Realistic PII via Mimesis (Pool-and-Sample)

- Names, emails, phone numbers, addresses from locale-aware providers
- **Pool-and-sample pattern** — Mimesis generates ~1K unique values, then NumPy samples from the pool at array speed (avoids slow per-row Mimesis calls)
- Consistent formatting (not "test1@test.com" patterns)
- Seeded for reproducibility — same seed = same data every time (`np.random.default_rng(seed)` + `Generic(seed=seed)`)

### Data Quality Patterns

- **Null injection** — configurable `percentNulls` for any column
- **Duplicate generation** — for deduplication pipeline demos
- **Late-arriving data** — timestamps outside expected windows
- **Anomaly injection** — configurable anomaly rates for sensor/IoT data

### Time-Based Patterns

- **Seasonality** — monthly, weekly, and campaign-based traffic spikes
- **Time series** — sinusoidal sensor data with trend and noise components
- **CDC batches** — APPEND/UPDATE/DELETE operations across time windows

## Key Capabilities

Beyond basic data generation, the skill supports advanced Databricks patterns:

| Capability                  | Description                                                  | Reference                                                                       |
| --------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------- |
| **CDC Generation**          | APPEND/UPDATE/DELETE batch operations for pipeline demos     | [cdc-patterns.md](references/cdc-patterns.md)                                   |
| **Streaming**               | Real-time data via `withStreaming=True` and rate sources     | [streaming-patterns.md](references/streaming-patterns.md)                       |
| **Medallion Architecture**  | Bronze/Silver/Gold with UC Volumes                           | [medallion-patterns.md](references/medallion-patterns.md)                       |
| **Multi-Table Consistency** | FK relationships, cardinality control                        | [multi-table-patterns.md](references/multi-table-patterns.md)                   |
| **Schema Introspection**    | Generate synthetic data matching an existing UC table schema | [schema-introspection-patterns.md](references/schema-introspection-patterns.md) |
| **Data Quality Injection**  | Nulls, duplicates, late-arriving data for DQ demos           | [data-quality-patterns.md](references/data-quality-patterns.md)                 |
| **Seasonality**             | Monthly, weekly, and campaign time-based patterns            | [seasonality-patterns.md](references/seasonality-patterns.md)                   |
| **ML Features**             | Drift simulation, label imbalance, feature arrays            | [ml-feature-patterns.md](references/ml-feature-patterns.md)                     |
| **Time Series**             | Temporal data, sine waves, fault injection                   | [time-series-patterns.md](references/time-series-patterns.md)                   |

## Prerequisites

### For Tier 1 (local generation)

No setup required beyond having `uv` installed. Dependencies are supplied at runtime:

```bash
uv run --with polars --with numpy --with mimesis script.py
```

### For Tier 2/3 (Databricks Connect or notebooks)

```bash
# Configure Databricks CLI with the DEFAULT profile
databricks configure

# Tier 2 (dbldatagen + Connect)
uv run --with "databricks-connect>=16.4,<17.0" --with dbldatagen --with jmespath --with pyparsing script.py

# Tier 1 + UC write (Polars + NumPy + Mimesis → Connect bridge)
uv run --with polars --with numpy --with mimesis --with "databricks-connect>=16.4,<17.0" script.py
```

**Important version constraints:**

- Pin `databricks-connect` to **16.x** — newer versions have serverless compatibility issues
- Use **Python 3.12** — serverless compute runs 3.12.3; mismatched versions break UDF serialization
- Never install standalone `pyspark` — always use `databricks-connect` with `DatabricksSession`

## Project Structure

```
databricks-data-generation/
├── SKILL.md                           # Main skill file — Claude reads this first
├── README.md                          # This file
├── references/                        # 17 pattern guides + 8 industry patterns (loaded on demand)
│   ├── generator-api.md               # Function signatures and default row counts
│   ├── examples.md                    # End-to-end demo examples
│   ├── polars-generation-guide.md     # Tier 1: Polars + NumPy + Mimesis patterns
│   ├── dbldatagen-guide.md            # Core dbldatagen API reference
│   ├── dbldatagen-connect-patterns.md # Validated dbldatagen + Connect patterns
│   ├── mimesis-guide.md               # Mimesis provider reference
│   ├── databricks-connect-guide.md    # Local dev + pure PySpark patterns
│   ├── cdc-patterns.md               # Change data capture batches
│   ├── streaming-patterns.md          # Real-time generation
│   ├── medallion-patterns.md          # Bronze/Silver/Gold architecture
│   ├── multi-table-patterns.md        # FK consistency, cardinality control
│   ├── data-quality-patterns.md       # Null injection, duplicates
│   ├── seasonality-patterns.md        # Time-based patterns
│   ├── ml-feature-patterns.md         # Drift, label imbalance, feature arrays
│   ├── time-series-patterns.md        # Temporal data patterns
│   ├── schema-introspection-patterns.md # DataAnalyzer workflows
│   ├── troubleshooting.md             # Common issues and fixes
│   └── industry-patterns/             # Per-industry schema references
│       ├── retail.md
│       ├── healthcare.md              # Includes clinical trials variant
│       ├── financial.md
│       ├── iot.md
│       ├── manufacturing.md
│       ├── supply_chain.md
│       ├── oil_gas.md
│       └── gaming.md
├── scripts/                           # Reference implementations (not importable)
│   ├── generators/                    # Tier 2/3 Spark + dbldatagen generators
│   │   ├── retail.py
│   │   ├── healthcare.py
│   │   ├── financial.py
│   │   ├── iot.py
│   │   ├── manufacturing.py
│   │   ├── supply_chain.py
│   │   ├── oil_gas.py
│   │   ├── gaming.py
│   │   ├── clinical_trials.py
│   │   ├── cdc.py
│   │   └── polars/                    # Tier 1 Polars + NumPy + Mimesis generators
│   │       ├── retail.py
│   │       ├── healthcare.py
│   │       ├── financial.py
│   │       ├── iot.py
│   │       ├── manufacturing.py
│   │       └── cdc.py
│   └── utils/
│       ├── mimesis_text.py            # MimesisText PyfuncTextFactory (notebooks)
│       ├── introspect.py              # DataAnalyzer wrappers
│       ├── output.py                  # write_delta, write_medallion helpers
│       ├── local_output.py            # Tier 1 local file output
│       └── uc_setup.py                # SDK-based UC provisioning
└── assets/
    └── schemas/                       # 33 JSON schemas (one per table)
        ├── retail_customer.json
        ├── retail_product.json
        ├── retail_transaction.json
        ├── retail_line_item.json
        ├── retail_inventory.json
        ├── healthcare_patient.json
        ├── healthcare_encounter.json
        ├── healthcare_claim.json
        ├── financial_account.json
        ├── financial_trade.json
        ├── financial_transaction.json
        ├── iot_device.json
        ├── iot_sensor.json
        ├── iot_event.json
        ├── iot_telemetry.json
        ├── manufacturing_equipment.json
        ├── manufacturing_sensor_data.json
        ├── manufacturing_maintenance.json
        ├── supply_chain_product.json
        ├── supply_chain_distribution_center.json
        ├── supply_chain_store.json
        ├── supply_chain_order.json
        ├── supply_chain_inventory_snapshot.json
        ├── supply_chain_shipment.json
        ├── oil_gas_well_header.json
        ├── oil_gas_daily_production.json
        ├── oil_gas_type_curve.json
        ├── gaming_login_event.json
        ├── healthcare_clinical_trial.json
        ├── healthcare_study_site.json
        ├── healthcare_study_participant.json
        ├── healthcare_adverse_event.json
        └── healthcare_lab_measurement.json
```

### How Claude Uses This Structure

1. **`SKILL.md`** is read first — contains the three-tier decision tree, quick starts, code generation guidelines, and industry schemas
2. **`references/*.md`** are loaded on demand — when a user requests CDC data, Claude reads `cdc-patterns.md`; for streaming, it reads `streaming-patterns.md`
3. **`scripts/generators/`** are reference implementations — Claude reads them for patterns and adapts inline for the user's specific request
4. **`assets/schemas/`** define column names, types, and generation hints — Claude uses these to ensure consistent table structures

The scripts are **not importable modules** (no `__init__.py`). They exist solely as patterns for Claude to reference when generating code.

## Quick Examples

### Tier 1: Local Retail Data (sub-second, zero setup)

```python
import numpy as np
import polars as pl
from mimesis import Generic
from mimesis.locales import Locale

rng = np.random.default_rng(42)
g = Generic(locale=Locale.EN, seed=42)
rows = 10_000

# Pool-and-sample: generate ~1K unique Mimesis values, then NumPy picks from pool
pool = min(1_000, rows)
_first = np.array([g.person.first_name() for _ in range(pool)])
_last = np.array([g.person.last_name() for _ in range(pool)])

start = np.datetime64("2020-01-01")
span = (np.datetime64("2024-12-31") - start).astype(int)

_tier_w = np.array([50, 30, 15, 5], dtype=np.float64)

customers = pl.DataFrame({
    "customer_id": np.arange(1_000_000, 1_000_000 + rows),
    "first_name": _first[rng.integers(0, pool, size=rows)],
    "last_name": _last[rng.integers(0, pool, size=rows)],
    "loyalty_tier": rng.choice(
        ["Bronze", "Silver", "Gold", "Platinum"],
        size=rows, p=_tier_w / _tier_w.sum()),
    "signup_date": start + rng.integers(0, span + 1, size=rows).astype("timedelta64[D]"),
})
customers.write_parquet("output/retail/customers.parquet")
```

Run: `uv run --with polars --with numpy --with mimesis script.py`

### Tier 2: dbldatagen + Connect to Unity Catalog

```python
from databricks.connect import DatabricksSession
import dbldatagen as dg

spark = DatabricksSession.builder.serverless().getOrCreate()

customers = (
    dg.DataGenerator(spark, rows=500_000, partitions=8, randomSeed=42)
    .withColumn("customer_id", "long", minValue=1_000_000, uniqueValues=500_000)
    .withColumn("first_name", "string",
                values=["James","Mary","Robert","Patricia","John","Jennifer"],
                random=True)
    .withColumn("loyalty_tier", "string",
                values=["Bronze", "Silver", "Gold", "Platinum"],
                weights=[50, 30, 15, 5])
    .withColumn("signup_date", "date",
                begin="2020-01-01", end="2024-12-31", random=True)
    .build()
)
customers.write.format("delta").mode("overwrite").saveAsTable("demo.retail.customers")
```

Run: `uv run --with "databricks-connect>=16.4,<17.0" --with dbldatagen --with jmespath --with pyparsing script.py`

## Underlying Technologies

| Technology                                                                          | Role                                                    | Used In                    |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------- | -------------------------- |
| [Polars](https://pola.rs/)                                                          | Fast DataFrame library for local generation             | Tier 1, Tier 2 alt         |
| [NumPy](https://numpy.org/)                                                         | Vectorized randomness, distributions, pool sampling     | Tier 1, Tier 2 alt         |
| [Mimesis](https://mimesis.name/)                                                    | Locale-aware fake data (PII, addresses, etc.)           | Tier 1, Tier 2 alt, Tier 3 |
| [dbldatagen](https://github.com/databrickslabs/dbldatagen)                          | Databricks Labs synthetic data generator for Spark      | Tier 2, Tier 3             |
| [Databricks Connect](https://docs.databricks.com/dev-tools/databricks-connect.html) | Local-to-serverless Spark bridge                        | Tier 1+UC, Tier 2          |
| [Unity Catalog](https://docs.databricks.com/data-governance/unity-catalog/)         | Governed data catalog for Delta tables                  | Tier 1+UC, Tier 2, Tier 3  |

## Troubleshooting

Common issues and their fixes are documented in [troubleshooting.md](references/troubleshooting.md). The most frequent ones:

| Issue                                          | Fix                                                           |
| ---------------------------------------------- | ------------------------------------------------------------- |
| `ModuleNotFoundError` for polars/numpy/mimesis | Use `uv run --with polars --with numpy --with mimesis`        |
| `Serverless mode is not yet supported`         | Pin `databricks-connect` to 16.x                              |
| dbldatagen UDF features fail over Connect      | Use Polars alternative or Tier 3 notebook                     |
| FK mismatch between tables                     | Ensure child `maxValue` matches parent row count              |
| `jmespath`/`pyparsing` missing with dbldatagen | Add `--with jmespath --with pyparsing`                        |
