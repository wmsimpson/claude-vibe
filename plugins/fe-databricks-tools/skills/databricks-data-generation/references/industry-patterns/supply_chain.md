# Supply Chain (CPG) Industry Patterns

Comprehensive data models and generation patterns for CPG supply chain demos.

## Data Model Overview

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Products   │────<│    Orders    │>────│ Distribution │
└──────────────┘     └──────────────┘     │   Centers    │
       │                                   └──────────────┘
       │                                          │
       ▼                                          │
┌──────────────┐     ┌──────────────┐            │
│  Inventory   │     │  Shipments   │>───────────┘
│  Snapshots   │     └──────────────┘
└──────────────┘            │
       ▲                    │
       │                    ▼
       │             ┌──────────────┐
       └─────────────│    Stores    │
                     └──────────────┘
```

## Table Schemas

### Products

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `id` | LONG | Primary key | Unique, starting at 1 |
| `sku` | STRING | Stock keeping unit | expr: `concat('SKU-', lpad(...))` |
| `product_name` | STRING | Product display name | Template: `\\w \\w Product` |
| `category` | STRING | CPG product category | 7 values, random |
| `brand` | STRING | Brand name | 5 values, random |
| `unit_cost` | DECIMAL(10,2) | Cost per unit | Range 0.50-50.00 |
| `unit_price` | DECIMAL(10,2) | Price per unit | Range 1.00-100.00 |
| `units_per_case` | INTEGER | Case pack size | Values: 6, 12, 24, 48 |
| `weight_kg` | DECIMAL(8,2) | Product weight | Range 0.1-25.0 |
| `shelf_life_days` | INTEGER | Days until expiry | Range 30-730 |
| `created_date` | DATE | Product creation date | Random 2020-2024 |

```python
import dbldatagen as dg

product_categories = ["Beverages", "Snacks", "Dairy", "Bakery",
                      "Frozen Foods", "Personal Care", "Household"]
brands = ["Premium Brand A", "Value Brand B", "Store Brand C",
          "Organic Brand D", "Brand E"]

products = (
    dg.DataGenerator(spark, rows=500, partitions=4, randomSeed=42)
    .withColumn("id", "long", minValue=1, uniqueValues=500)
    .withColumn("sku", "string",
                expr="concat('SKU-', lpad(cast(id as string), 6, '0'))",
                uniqueValues=500)
    .withColumn("product_name", "string", template=r"\\w \\w Product")
    .withColumn("category", "string", values=product_categories, random=True)
    .withColumn("brand", "string", values=brands, random=True)
    .withColumn("unit_cost", "decimal(10,2)", minValue=0.5, maxValue=50.0, random=True)
    .withColumn("unit_price", "decimal(10,2)", minValue=1.0, maxValue=100.0, random=True)
    .withColumn("units_per_case", "integer", values=[6, 12, 24, 48], random=True)
    .withColumn("weight_kg", "decimal(8,2)", minValue=0.1, maxValue=25.0, random=True)
    .withColumn("shelf_life_days", "integer", minValue=30, maxValue=730, random=True)
    .withColumn("created_date", "date", begin="2020-01-01", end="2024-01-01",
                interval="1 day", random=True)
    .build()
)
```

### Distribution Centers

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `id` | LONG | Primary key | Unique, starting at 1 |
| `distribution_center_code` | STRING | DC identifier code | expr: `concat('DC-', lpad(...))` |
| `distribution_center_name` | STRING | Facility name | Template: `\\w Distribution Center` |
| `region` | STRING | US geographic region | 5 values, random |
| `capacity_pallets` | INTEGER | Max pallet capacity | Range 5,000-50,000 |
| `current_utilization_pct` | DECIMAL(5,2) | Current usage | Range 45.0-95.0 |
| `latitude` | DECIMAL(9,6) | Geographic latitude | Range 25.0-49.0 |
| `longitude` | DECIMAL(9,6) | Geographic longitude | Range -125.0 to -65.0 |
| `operating_cost_daily` | DECIMAL(10,2) | Daily operating cost | Range 5,000-50,000 |
| `opened_date` | DATE | Facility open date | Random 2015-2023 |

```python
distribution_centers = (
    dg.DataGenerator(spark, rows=25, partitions=4, randomSeed=42)
    .withColumn("id", "long", minValue=1, uniqueValues=25)
    .withColumn("distribution_center_code", "string",
                expr="concat('DC-', lpad(cast(id as string), 4, '0'))",
                uniqueValues=25)
    .withColumn("distribution_center_name", "string",
                template=r"\\w Distribution Center")
    .withColumn("region", "string",
                values=["Northeast", "Southeast", "Midwest", "Southwest", "West"],
                random=True)
    .withColumn("capacity_pallets", "integer", minValue=5000, maxValue=50000, random=True)
    .withColumn("current_utilization_pct", "decimal(5,2)",
                minValue=45.0, maxValue=95.0, random=True)
    .withColumn("latitude", "decimal(9,6)", minValue=25.0, maxValue=49.0, random=True)
    .withColumn("longitude", "decimal(9,6)", minValue=-125.0, maxValue=-65.0, random=True)
    .withColumn("operating_cost_daily", "decimal(10,2)",
                minValue=5000, maxValue=50000, random=True)
    .withColumn("opened_date", "date", begin="2015-01-01", end="2023-01-01", random=True)
    .build()
)
```

### Stores

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `id` | LONG | Primary key | Unique, starting at 1 |
| `store_code` | STRING | Store identifier code | expr: `concat('STORE-', lpad(...))` |
| `retailer` | STRING | Retailer name | 5 values, random |
| `store_format` | STRING | Store type | 5 values, random |
| `region` | STRING | US geographic region | 5 values, random |
| `square_footage` | INTEGER | Store area | Range 2,000-200,000 |
| `distribution_center_id` | INTEGER | FK to distribution_centers | Range 1-25 |
| `latitude` | DECIMAL(9,6) | Geographic latitude | Range 25.0-49.0 |
| `longitude` | DECIMAL(9,6) | Geographic longitude | Range -125.0 to -65.0 |
| `opened_date` | DATE | Store open date | Random 2010-2024 |

```python
store_formats = ["Hypermarket", "Supermarket", "Convenience", "Online", "Club Store"]
retailers = ["RetailCo", "MegaMart", "QuickStop", "FreshGrocer", "ValueMart"]

stores = (
    dg.DataGenerator(spark, rows=1_000, partitions=8, randomSeed=42)
    .withColumn("id", "long", minValue=1, uniqueValues=1_000)
    .withColumn("store_code", "string",
                expr="concat('STORE-', lpad(cast(id as string), 6, '0'))",
                uniqueValues=1_000)
    .withColumn("retailer", "string", values=retailers, random=True)
    .withColumn("store_format", "string", values=store_formats, random=True)
    .withColumn("region", "string",
                values=["Northeast", "Southeast", "Midwest", "Southwest", "West"],
                random=True)
    .withColumn("square_footage", "integer", minValue=2000, maxValue=200000, random=True)
    .withColumn("distribution_center_id", "integer", minValue=1, maxValue=25, random=True)
    .withColumn("latitude", "decimal(9,6)", minValue=25.0, maxValue=49.0, random=True)
    .withColumn("longitude", "decimal(9,6)", minValue=-125.0, maxValue=-65.0, random=True)
    .withColumn("opened_date", "date", begin="2010-01-01", end="2024-01-01", random=True)
    .build()
)
```

### Orders

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `id` | LONG | Primary key | Unique, starting at 1 |
| `order_number` | STRING | Purchase order number | expr: `concat('PO-', lpad(...))` |
| `distribution_center_id` | INTEGER | FK to distribution_centers | Range 1-25 |
| `product_id` | INTEGER | FK to products | Range 1-500 |
| `order_date` | TIMESTAMP | Order creation time | Random 2024-2025 |
| `scheduled_start` | DATE | Planned start (post-proc) | `date_add(order_date, days)` |
| `scheduled_end` | DATE | Planned end (post-proc) | `date_add(scheduled_start, days)` |
| `actual_start` | TIMESTAMP | Actual start (post-proc) | Conditional on probability |
| `actual_end` | TIMESTAMP | Actual end (post-proc) | Conditional on start + probability |
| `quantity_ordered` | INTEGER | Units ordered | Range 500-50,000 |
| `quantity_produced` | INTEGER | Units produced (post-proc) | Variance of ordered if completed |
| `status` | STRING | Order status (post-proc) | 5 values via modulo |
| `line_efficiency_pct` | DECIMAL(5,2) | Production efficiency | Range 75.0-98.0 |
| `production_cost` | DECIMAL(12,2) | Total production cost | Range 5,000-500,000 |

```python
from pyspark.sql import functions as F

order_status = ["Scheduled", "In Progress", "Completed", "Delayed", "Quality Hold"]

order_spec = (
    dg.DataGenerator(spark, rows=10_000, partitions=8, randomSeed=42)
    .withColumn("id", "long", minValue=1, uniqueValues=10_000)
    .withColumn("order_number", "string",
                expr="concat('PO-', lpad(cast(id as string), 8, '0'))",
                uniqueValues=10_000)
    .withColumn("distribution_center_id", "integer", minValue=1, maxValue=25, random=True)
    .withColumn("product_id", "integer", minValue=1, maxValue=500, random=True)
    .withColumn("order_date", "timestamp",
                begin="2024-01-01 00:00:00", end="2025-09-29 23:59:59", random=True)
    # Seed columns for post-processing
    .withColumn("scheduled_start_days", "integer", minValue=0, maxValue=10, random=True)
    .withColumn("scheduled_duration_days", "integer", minValue=1, maxValue=6, random=True)
    .withColumn("start_delay_hours", "integer", minValue=-12, maxValue=12, random=True)
    .withColumn("actual_duration_hours", "integer", minValue=24, maxValue=144, random=True)
    .withColumn("start_probability", "double", minValue=0, maxValue=1, random=True)
    .withColumn("completion_probability", "double", minValue=0, maxValue=1, random=True)
    .withColumn("quantity_ordered", "integer", minValue=500, maxValue=50000, random=True)
    .withColumn("order_variance", "double", minValue=0.85, maxValue=1.0, random=True)
    .withColumn("status_rand", "integer", minValue=1, maxValue=10000, random=True)
    .withColumn("line_efficiency_pct", "decimal(5,2)", minValue=75.0, maxValue=98.0, random=True)
    .withColumn("production_cost", "decimal(12,2)", minValue=5000, maxValue=500000, random=True)
)

df_orders = order_spec.build()

# Post-processing: derive schedule, actuals, status
df_orders = (
    df_orders
    .withColumn("scheduled_start",
                F.expr("date_add(order_date, scheduled_start_days)"))
    .withColumn("scheduled_end",
                F.expr("date_add(scheduled_start, scheduled_duration_days)"))
    .withColumn("actual_start",
                F.when(F.col("start_probability") > 0.3,
                       F.expr("timestampadd(HOUR, start_delay_hours, scheduled_start)"))
                .otherwise(None))
    .withColumn("actual_end",
                F.when((F.col("actual_start").isNotNull()) &
                       (F.col("completion_probability") > 0.2),
                       F.expr("timestampadd(HOUR, actual_duration_hours, actual_start)"))
                .otherwise(None))
    .withColumn("quantity_produced",
                F.when(F.col("actual_end").isNotNull(),
                       (F.col("quantity_ordered") * F.col("order_variance")).cast("integer"))
                .otherwise(0))
    .withColumn("status_index", F.col("status_rand") % 5)
    .withColumn("status",
                F.array([F.lit(s) for s in order_status]).getItem(F.col("status_index")))
    .drop("scheduled_start_days", "scheduled_duration_days", "start_delay_hours",
          "actual_duration_hours", "start_probability", "completion_probability",
          "order_variance", "status_rand", "status_index")
)
```

### Inventory Snapshots

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `id` | LONG | Primary key | Unique, starting at 1 |
| `snapshot_date` | DATE | Snapshot date | Random 2024-2025 |
| `location_type` | STRING | DC or Store | Weighted: 30% DC / 70% Store |
| `location_id` | INTEGER | FK to DC or Store | CASE on location_type |
| `product_id` | INTEGER | FK to products | Range 1-500 |
| `quantity_on_hand` | INTEGER | Current stock | Range 0-10,000 |
| `quantity_reserved` | INTEGER | Reserved stock | Derived from on_hand |
| `quantity_available` | INTEGER | Available stock | on_hand - reserved |
| `reorder_point` | INTEGER | Reorder trigger | Range 100-2,000 |
| `daily_demand` | DOUBLE | Daily demand rate | Range 50.0-150.0 |
| `days_of_supply` | DECIMAL(8,2) | Supply coverage days | Safe division |
| `inventory_value` | DECIMAL(12,2) | Dollar value of stock | Range 1,000-500,000 |
| `last_received_date` | DATE | Last receipt date | Derived from snapshot_date |
| `stockout_risk` | STRING | Risk level | CASE: High/Medium/Low |

```python
inventory_spec = (
    dg.DataGenerator(spark, rows=50_000, partitions=8, randomSeed=42)
    .withColumn("id", "long", minValue=1, uniqueValues=50_000)
    .withColumn("snapshot_date", "date",
                begin="2024-01-01", end="2025-09-29", random=True)
    # Weighted: 30% distribution_center, 70% Store
    .withColumn("location_type_seed", "double", minValue=0, maxValue=1, random=True)
    .withColumn("location_type", "string", expr="""
        CASE
            WHEN location_type_seed < 0.3 THEN 'distribution_center'
            ELSE 'Store'
        END
    """)
    .withColumn("location_id", "integer", expr="""
        CASE
            WHEN location_type = 'distribution_center' THEN (id % 25) + 1
            ELSE (id % 1000) + 1
        END
    """)
    .withColumn("product_id", "integer", minValue=1, maxValue=500, random=True)
    .withColumn("quantity_on_hand", "integer", minValue=0, maxValue=10000, random=True)
    .withColumn("reserve_factor", "double", minValue=0, maxValue=0.5, random=True)
    .withColumn("quantity_reserved", "integer",
                expr="cast(quantity_on_hand * reserve_factor as int)")
    .withColumn("quantity_available", "integer",
                expr="quantity_on_hand - quantity_reserved")
    .withColumn("reorder_point", "integer", minValue=100, maxValue=2000, random=True)
    .withColumn("daily_demand", "double", minValue=50.0, maxValue=150.0, random=True)
    .withColumn("days_of_supply", "decimal(8,2)", expr="""
        CASE
            WHEN daily_demand > 0 THEN cast(quantity_available / daily_demand as decimal(8,2))
            ELSE NULL
        END
    """)
    .withColumn("inventory_value", "decimal(12,2)", minValue=1000, maxValue=500000, random=True)
    .withColumn("days_offset", "integer", minValue=0, maxValue=60, random=True)
    .withColumn("last_received_date", "date", expr="date_sub(snapshot_date, days_offset)")
    .withColumn("stockout_risk", "string", expr="""
        CASE
            WHEN days_of_supply IS NULL OR days_of_supply < 3 THEN 'High'
            WHEN days_of_supply < 7 THEN 'Medium'
            ELSE 'Low'
        END
    """)
)

df_inventory = inventory_spec.build().drop("reserve_factor", "days_offset", "location_type_seed")
```

### Shipments

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `id` | LONG | Primary key | Unique, starting at 1 |
| `shipment_id` | STRING | Shipment tracking ID | expr: `concat('SHP-', lpad(...))` |
| `origin_distribution_center_id` | INTEGER | FK to distribution_centers | Range 1-25 |
| `destination_type` | STRING | DC or Store | Weighted: 30% DC / 70% Store |
| `destination_id` | INTEGER | FK to DC or Store | CASE on destination_type |
| `product_id` | INTEGER | FK to products | Range 1-500 |
| `ship_date` | TIMESTAMP | Shipment departure | Random 2024-2025 |
| `expected_delivery` | TIMESTAMP | Expected arrival | `date_add(ship_date, transit_days)` |
| `actual_delivery` | TIMESTAMP | Actual arrival | 80% delivered |
| `on_time` | BOOLEAN | Delivered on/before expected | Computed from dates |
| `delay_hours` | INTEGER | Hours of delay | Computed from dates |
| `quantity` | INTEGER | Units shipped | Range 100-5,000 |
| `transport_mode` | STRING | Transport method | Weighted: 60/15/20/5 |
| `carrier` | STRING | Carrier name | 4 values, random |
| `status` | STRING | Shipment status | Weighted: 25/50/5/10/10 |
| `shipping_cost` | DECIMAL(10,2) | Cost of shipment | Range 50-5,000 |
| `distance_miles` | INTEGER | Shipping distance | Range 50-2,500 |

```python
shipments_spec = (
    dg.DataGenerator(spark, rows=30_000, partitions=8, randomSeed=42)
    .withColumn("id", "long", minValue=1, uniqueValues=30_000)
    .withColumn("shipment_id", "string",
                expr="concat('SHP-', lpad(cast(id as string), 10, '0'))",
                uniqueValues=30_000)
    .withColumn("origin_distribution_center_id", "integer",
                minValue=1, maxValue=25, random=True)
    # Weighted: 30% DC, 70% Store
    .withColumn("destination_type_seed", "double", minValue=0, maxValue=1, random=True)
    .withColumn("destination_type", "string", expr="""
        CASE
            WHEN destination_type_seed < 0.3 THEN 'distribution_center'
            ELSE 'Store'
        END
    """)
    .withColumn("destination_id", "integer", expr="""
        CASE
            WHEN destination_type = 'distribution_center' THEN (id % 25) + 1
            ELSE (id % 1000) + 1
        END
    """)
    .withColumn("product_id", "integer", minValue=1, maxValue=500, random=True)
    .withColumn("ship_date", "timestamp",
                begin="2024-01-01 00:00:00", end="2025-09-29 23:59:59", random=True)
    .withColumn("transit_days", "integer", minValue=1, maxValue=6, random=True)
    .withColumn("actual_transit_days", "integer", minValue=1, maxValue=8, random=True)
    .withColumn("delivery_probability", "double", minValue=0, maxValue=1, random=True)
    .withColumn("expected_delivery", "timestamp",
                expr="date_add(ship_date, transit_days)")
    .withColumn("actual_delivery", "timestamp", expr="""
        CASE
            WHEN delivery_probability > 0.2 THEN date_add(ship_date, actual_transit_days)
            ELSE NULL
        END
    """)
    .withColumn("on_time", "boolean", expr="""
        actual_delivery IS NOT NULL AND actual_delivery <= expected_delivery
    """)
    .withColumn("delay_hours", "integer", expr="""
        CASE
            WHEN actual_delivery IS NOT NULL THEN
                cast((unix_timestamp(actual_delivery) - unix_timestamp(expected_delivery)) / 3600 as int)
            ELSE NULL
        END
    """)
    .withColumn("quantity", "integer", minValue=100, maxValue=5000, random=True)
    # Transport mode: 60% Truck, 15% Rail, 20% Intermodal, 5% Air
    .withColumn("transport_mode_seed", "double", minValue=0, maxValue=1, random=True)
    .withColumn("transport_mode", "string", expr="""
        CASE
            WHEN transport_mode_seed < 0.60 THEN 'Truck'
            WHEN transport_mode_seed < 0.75 THEN 'Rail'
            WHEN transport_mode_seed < 0.95 THEN 'Intermodal'
            ELSE 'Air'
        END
    """)
    .withColumn("carrier", "string",
                values=["FastFreight", "ReliableLogistics", "ExpressTransport", "GlobalShippers"],
                random=True)
    # Status: 25% In Transit, 50% Delivered, 5% Delayed, 10% At Hub, 10% Out for Delivery
    .withColumn("status_seed", "double", minValue=0, maxValue=1, random=True)
    .withColumn("status", "string", expr="""
        CASE
            WHEN status_seed < 0.25 THEN 'In Transit'
            WHEN status_seed < 0.75 THEN 'Delivered'
            WHEN status_seed < 0.80 THEN 'Delayed'
            WHEN status_seed < 0.90 THEN 'At Hub'
            ELSE 'Out for Delivery'
        END
    """)
    .withColumn("shipping_cost", "decimal(10,2)", minValue=50, maxValue=5000, random=True)
    .withColumn("distance_miles", "integer", minValue=50, maxValue=2500, random=True)
)

df_shipments = shipments_spec.build().drop(
    "transit_days", "actual_transit_days", "delivery_probability",
    "destination_type_seed", "transport_mode_seed", "status_seed"
)
```

## Common Demo Queries

### Inventory Health
```sql
SELECT
    location_type,
    stockout_risk,
    COUNT(*) as item_count,
    SUM(inventory_value) as total_value,
    ROUND(AVG(days_of_supply), 1) as avg_days_supply
FROM inventory
WHERE snapshot_date = (SELECT MAX(snapshot_date) FROM inventory)
GROUP BY location_type, stockout_risk
ORDER BY location_type,
    CASE stockout_risk
        WHEN 'High' THEN 1
        WHEN 'Medium' THEN 2
        WHEN 'Low' THEN 3
    END
```

### Carrier Performance
```sql
SELECT
    carrier,
    COUNT(*) as total_shipments,
    ROUND(AVG(CASE WHEN on_time = true THEN 100.0 ELSE 0.0 END), 1) as otd_pct,
    ROUND(AVG(shipping_cost), 2) as avg_cost,
    ROUND(AVG(distance_miles), 0) as avg_distance,
    ROUND(AVG(shipping_cost / distance_miles), 3) as cost_per_mile
FROM shipments
WHERE actual_delivery IS NOT NULL
GROUP BY carrier
ORDER BY total_shipments DESC
```

### Network Overview
```sql
SELECT
    dc.distribution_center_code,
    dc.region,
    dc.capacity_pallets,
    ROUND(dc.current_utilization_pct, 1) as utilization_pct,
    COUNT(DISTINCT i.product_id) as active_skus,
    SUM(i.inventory_value) as inventory_value,
    COUNT(DISTINCT s.id) as outbound_shipments_last_30d,
    ROUND(AVG(CASE WHEN s.on_time = true THEN 100.0 ELSE 0.0 END), 1) as otd_pct
FROM distribution_centers dc
LEFT JOIN inventory i ON dc.id = i.location_id
    AND i.location_type = 'distribution_center'
    AND i.snapshot_date = (SELECT MAX(snapshot_date) FROM inventory)
LEFT JOIN shipments s ON dc.id = s.origin_distribution_center_id
    AND s.ship_date >= CURRENT_DATE - INTERVAL 30 DAY
GROUP BY dc.distribution_center_code, dc.region,
         dc.capacity_pallets, dc.current_utilization_pct
ORDER BY inventory_value DESC
```

## Complete Demo

```python
def generate_supply_chain_demo(
    spark,
    catalog: str,
    schema: str = "supply_chain",
    volume: str = "raw_data",
    seed: int = 42,
):
    """Generate complete supply chain demo dataset with all tables.

    Writes raw JSON to UC Volume and clean Delta tables to the catalog.
    """
    from ..utils.output import write_medallion

    n_products = 500
    n_distribution_centers = 25
    n_stores = 1_000

    products = generate_products(spark, rows=n_products, seed=seed)
    distribution_centers = generate_distribution_centers(
        spark, rows=n_distribution_centers, seed=seed
    )
    stores = generate_stores(
        spark, rows=n_stores, n_distribution_centers=n_distribution_centers, seed=seed
    )
    orders = generate_orders(
        spark, rows=10_000, n_distribution_centers=n_distribution_centers,
        n_products=n_products, seed=seed
    )
    inventory_snapshots = generate_inventory_snapshots(
        spark, rows=50_000, n_products=n_products,
        n_distribution_centers=n_distribution_centers,
        n_stores=n_stores, seed=seed
    )
    shipments = generate_shipments(
        spark, rows=30_000, n_distribution_centers=n_distribution_centers,
        n_products=n_products, n_stores=n_stores, seed=seed
    )

    write_medallion(
        tables={
            "products": products,
            "distribution_centers": distribution_centers,
            "stores": stores,
            "orders": orders,
            "inventory_snapshots": inventory_snapshots,
            "shipments": shipments,
        },
        catalog=catalog,
        schema=schema,
        volume=volume,
    )
```
