"""Supply chain (CPG) industry synthetic data generators.

REFERENCE IMPLEMENTATION — This file is not an importable module. Claude reads it
for patterns and adapts the code inline for user notebooks and scripts.
"""

# CONNECT COMPATIBILITY NOTES:
#   Catalyst-safe (works over Connect):
#     - values=/weights=, minValue/maxValue, begin/end dates, expr=, percentNulls=, omit=
#   UDF-dependent (notebook only — apply workarounds for Connect):
#     - text=mimesisText() → values=["James","Mary",...], random=True
#     - template=r"ddddd" → expr="lpad(cast(floor(rand()*100000) as string), 5, '0')"
#     - distribution=Gamma/Beta → random=True or expr= math
#     - .withConstraint() → .build().filter("condition")

import dbldatagen as dg
from dbldatagen.config import OutputDataset
from pyspark.sql import DataFrame
from pyspark.sql import functions as F


def generate_products(spark, rows=500, seed=42,
                      output: OutputDataset | None = None) -> DataFrame | None:
    """Generate CPG product master data with categories, pricing, and packaging."""
    partitions = max(2, rows // 500)

    product_categories = [
        "Beverages", "Snacks", "Dairy", "Bakery",
        "Frozen Foods", "Personal Care", "Household",
    ]
    brands = [
        "Premium Brand A", "Value Brand B", "Store Brand C",
        "Organic Brand D", "Brand E",
    ]

    spec = (
        dg.DataGenerator(spark, rows=rows, partitions=partitions, randomSeed=seed)
        .withColumn("id", "long", minValue=1, uniqueValues=rows)
        .withColumn("sku", "string",
                    expr="concat('SKU-', lpad(cast(id as string), 6, '0'))",
                    uniqueValues=rows)
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
    )
    if output:
        spec.saveAsDataset(dataset=output)
        return None
    return spec.build()


def generate_distribution_centers(spark, rows=25, seed=42,
                                  output: OutputDataset | None = None) -> DataFrame | None:
    """Generate distribution center network with capacity and geographic data."""
    partitions = max(2, rows // 25)

    spec = (
        dg.DataGenerator(spark, rows=rows, partitions=partitions, randomSeed=seed)
        .withColumn("id", "long", minValue=1, uniqueValues=rows)
        .withColumn("distribution_center_code", "string",
                    expr="concat('DC-', lpad(cast(id as string), 4, '0'))",
                    uniqueValues=rows)
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
    )
    if output:
        spec.saveAsDataset(dataset=output)
        return None
    return spec.build()


def generate_stores(spark, rows=1_000, n_distribution_centers=25, seed=42,
                    output: OutputDataset | None = None) -> DataFrame | None:
    """Generate retail store locations with format and distribution center assignment."""
    partitions = max(4, rows // 1_000)

    store_formats = ["Hypermarket", "Supermarket", "Convenience", "Online", "Club Store"]
    retailers = ["RetailCo", "MegaMart", "QuickStop", "FreshGrocer", "ValueMart"]

    spec = (
        dg.DataGenerator(spark, rows=rows, partitions=partitions, randomSeed=seed)
        .withColumn("id", "long", minValue=1, uniqueValues=rows)
        .withColumn("store_code", "string",
                    expr="concat('STORE-', lpad(cast(id as string), 6, '0'))",
                    uniqueValues=rows)
        .withColumn("retailer", "string", values=retailers, random=True)
        .withColumn("store_format", "string", values=store_formats, random=True)
        .withColumn("region", "string",
                    values=["Northeast", "Southeast", "Midwest", "Southwest", "West"],
                    random=True)
        .withColumn("square_footage", "integer", minValue=2000, maxValue=200000, random=True)
        .withColumn("distribution_center_id", "integer",
                    minValue=1, maxValue=n_distribution_centers, random=True)
        .withColumn("latitude", "decimal(9,6)", minValue=25.0, maxValue=49.0, random=True)
        .withColumn("longitude", "decimal(9,6)", minValue=-125.0, maxValue=-65.0, random=True)
        .withColumn("opened_date", "date", begin="2010-01-01", end="2024-01-01", random=True)
    )
    if output:
        spec.saveAsDataset(dataset=output)
        return None
    return spec.build()


def generate_orders(spark, rows=10_000, n_distribution_centers=25, n_products=500,
                    seed=42, output: OutputDataset | None = None) -> DataFrame | None:
    """Generate manufacturing/purchase orders with schedule tracking and production metrics.

    Uses post-processing to derive scheduled_start, scheduled_end, actual_start,
    actual_end, quantity_produced, and status from seed columns.

    Note: When using OutputDataset, the raw spec output (before post-processing)
    is written. Users can apply post-processing separately.
    """
    partitions = max(4, rows // 10_000)

    order_status = ["Scheduled", "In Progress", "Completed", "Delayed", "Quality Hold"]

    spec = (
        dg.DataGenerator(spark, rows=rows, partitions=partitions, randomSeed=seed)
        .withColumn("id", "long", minValue=1, uniqueValues=rows)
        .withColumn("order_number", "string",
                    expr="concat('PO-', lpad(cast(id as string), 8, '0'))",
                    uniqueValues=rows)
        # Foreign keys
        .withColumn("distribution_center_id", "integer",
                    minValue=1, maxValue=n_distribution_centers, random=True)
        .withColumn("product_id", "integer", minValue=1, maxValue=n_products, random=True)
        # Base timestamp
        .withColumn("order_date", "timestamp",
                    begin="2024-01-01 00:00:00", end="2025-09-29 23:59:59", random=True)
        # Seed columns for post-processing (will be dropped)
        .withColumn("scheduled_start_days", "integer", minValue=0, maxValue=10, random=True)
        .withColumn("scheduled_duration_days", "integer", minValue=1, maxValue=6, random=True)
        .withColumn("start_delay_hours", "integer", minValue=-12, maxValue=12, random=True)
        .withColumn("actual_duration_hours", "integer", minValue=24, maxValue=144, random=True)
        .withColumn("start_probability", "double", minValue=0, maxValue=1, random=True)
        .withColumn("completion_probability", "double", minValue=0, maxValue=1, random=True)
        .withColumn("quantity_ordered", "integer", minValue=500, maxValue=50000, random=True)
        .withColumn("order_variance", "double", minValue=0.85, maxValue=1.0, random=True)
        .withColumn("status_rand", "integer", minValue=1, maxValue=10000, random=True)
        .withColumn("line_efficiency_pct", "decimal(5,2)",
                    minValue=75.0, maxValue=98.0, random=True)
        .withColumn("production_cost", "decimal(12,2)",
                    minValue=5000, maxValue=500000, random=True)
    )

    if output:
        spec.saveAsDataset(dataset=output)
        return None

    df = spec.build()

    # Post-processing: derive schedule, actuals, status
    df = (
        df
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
    return df


def generate_inventory_snapshots(spark, rows=50_000, n_products=500,
                                 n_distribution_centers=25, n_stores=1_000,
                                 seed=42,
                                 output: OutputDataset | None = None) -> DataFrame | None:
    """Generate multi-echelon inventory snapshots with stock levels and risk metrics.

    Uses weighted location_type (30% DC / 70% Store), safe division for
    days_of_supply, and CASE-based stockout_risk classification.

    Note: When using OutputDataset, the raw spec output (before dropping
    intermediate columns) is written.
    """
    partitions = max(4, rows // 50_000)

    spec = (
        dg.DataGenerator(spark, rows=rows, partitions=partitions, randomSeed=seed)
        .withColumn("id", "long", minValue=1, uniqueValues=rows)
        .withColumn("snapshot_date", "date",
                    begin="2024-01-01", end="2025-09-29", random=True)
        # Weighted distribution: 30% distribution_center, 70% Store
        .withColumn("location_type_seed", "double", minValue=0, maxValue=1, random=True)
        .withColumn("location_type", "string", expr="""
            CASE
                WHEN location_type_seed < 0.3 THEN 'distribution_center'
                ELSE 'Store'
            END
        """)
        .withColumn("location_id", "integer", expr=f"""
            CASE
                WHEN location_type = 'distribution_center' THEN (id % {n_distribution_centers}) + 1
                ELSE (id % {n_stores}) + 1
            END
        """)
        # Foreign key
        .withColumn("product_id", "integer", minValue=1, maxValue=n_products, random=True)
        # Inventory quantities
        .withColumn("quantity_on_hand", "integer", minValue=0, maxValue=10000, random=True)
        .withColumn("reserve_factor", "double", minValue=0, maxValue=0.5, random=True)
        .withColumn("quantity_reserved", "integer",
                    expr="cast(quantity_on_hand * reserve_factor as int)")
        .withColumn("quantity_available", "integer",
                    expr="quantity_on_hand - quantity_reserved")
        .withColumn("reorder_point", "integer", minValue=100, maxValue=2000, random=True)
        # Demand and supply metrics
        .withColumn("daily_demand", "double", minValue=50.0, maxValue=150.0, random=True)
        .withColumn("days_of_supply", "decimal(8,2)", expr="""
            CASE
                WHEN daily_demand > 0 THEN cast(quantity_available / daily_demand as decimal(8,2))
                ELSE NULL
            END
        """)
        .withColumn("inventory_value", "decimal(12,2)",
                    minValue=1000, maxValue=500000, random=True)
        .withColumn("days_offset", "integer", minValue=0, maxValue=60, random=True)
        .withColumn("last_received_date", "date", expr="date_sub(snapshot_date, days_offset)")
        # Risk classification
        .withColumn("stockout_risk", "string", expr="""
            CASE
                WHEN days_of_supply IS NULL OR days_of_supply < 3 THEN 'High'
                WHEN days_of_supply < 7 THEN 'Medium'
                ELSE 'Low'
            END
        """)
    )

    if output:
        spec.saveAsDataset(dataset=output)
        return None

    df = spec.build()
    df = df.drop("reserve_factor", "days_offset", "location_type_seed")
    return df


def generate_shipments(spark, rows=30_000, n_distribution_centers=25,
                       n_products=500, n_stores=1_000, seed=42,
                       output: OutputDataset | None = None) -> DataFrame | None:
    """Generate transportation and logistics shipment data with delivery tracking.

    Uses weighted distributions for transport_mode (60% Truck / 15% Rail /
    20% Intermodal / 5% Air) and status (25% In Transit / 50% Delivered /
    5% Delayed / 10% At Hub / 10% Out for Delivery). Approximately 80% of
    shipments have an actual_delivery date.

    Note: When using OutputDataset, the raw spec output (before dropping
    intermediate columns) is written.
    """
    partitions = max(4, rows // 30_000)

    spec = (
        dg.DataGenerator(spark, rows=rows, partitions=partitions, randomSeed=seed)
        .withColumn("id", "long", minValue=1, uniqueValues=rows)
        .withColumn("shipment_id", "string",
                    expr="concat('SHP-', lpad(cast(id as string), 10, '0'))",
                    uniqueValues=rows)
        # Origin is always a distribution center
        .withColumn("origin_distribution_center_id", "integer",
                    minValue=1, maxValue=n_distribution_centers, random=True)
        # Destination: 30% DC, 70% Store
        .withColumn("destination_type_seed", "double", minValue=0, maxValue=1, random=True)
        .withColumn("destination_type", "string", expr="""
            CASE
                WHEN destination_type_seed < 0.3 THEN 'distribution_center'
                ELSE 'Store'
            END
        """)
        .withColumn("destination_id", "integer", expr=f"""
            CASE
                WHEN destination_type = 'distribution_center' THEN (id % {n_distribution_centers}) + 1
                ELSE (id % {n_stores}) + 1
            END
        """)
        .withColumn("product_id", "integer", minValue=1, maxValue=n_products, random=True)
        # Shipment dates
        .withColumn("ship_date", "timestamp",
                    begin="2024-01-01 00:00:00", end="2025-09-29 23:59:59", random=True)
        .withColumn("transit_days", "integer", minValue=1, maxValue=6, random=True)
        .withColumn("actual_transit_days", "integer", minValue=1, maxValue=8, random=True)
        .withColumn("delivery_probability", "double", minValue=0, maxValue=1, random=True)
        .withColumn("expected_delivery", "timestamp",
                    expr="date_add(ship_date, transit_days)")
        # Actual delivery: 80% delivered (probability > 0.2)
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
                    values=["FastFreight", "ReliableLogistics",
                            "ExpressTransport", "GlobalShippers"],
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

    if output:
        spec.saveAsDataset(dataset=output)
        return None

    df = spec.build()
    df = df.drop(
        "transit_days", "actual_transit_days", "delivery_probability",
        "destination_type_seed", "transport_mode_seed", "status_seed"
    )
    return df


def generate_supply_chain_demo(spark, catalog, schema="supply_chain",
                               volume="raw_data", seed=42):
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
