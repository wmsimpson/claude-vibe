"""Retail industry synthetic data generators (Polars + NumPy + Mimesis).

REFERENCE IMPLEMENTATION — This file is not an importable module. Claude reads it
for patterns and adapts the code inline for user scripts. Uses Polars + NumPy for
vectorized Tier 1 local generation (<500K rows, zero JVM overhead).

NOTE: Seed reproducibility differs from the previous random-module version.
All randomness now flows through np.random.default_rng(seed).
"""

import numpy as np
import polars as pl
from mimesis import Generic
from mimesis.locales import Locale

STATES = ["NY", "CA", "IL", "TX", "AZ", "PA", "FL", "OH", "NC", "GA"]
CATEGORIES = ["Electronics", "Clothing", "Home & Garden", "Sports",
              "Beauty", "Food & Grocery", "Toys", "Automotive"]
BRANDS = ["BrandA", "BrandB", "BrandC", "BrandD", "BrandE", "Generic"]
PAYMENT_METHODS = ["Credit Card", "Debit Card", "Cash", "Digital Wallet", "Gift Card"]
PAYMENT_WEIGHTS = [40, 25, 15, 15, 5]


def generate_customers(rows: int = 10_000, seed: int = 42) -> pl.DataFrame:
    """Generate retail customer master data with demographics and loyalty tiers."""
    rng = np.random.default_rng(seed)
    g = Generic(locale=Locale.EN, seed=seed)

    pool = min(1_000, rows)
    _first = np.array([g.person.first_name() for _ in range(pool)])
    _last = np.array([g.person.last_name() for _ in range(pool)])
    _phone = np.array([g.person.telephone() for _ in range(pool)])
    _city = np.array([g.address.city() for _ in range(pool)])

    first_names = _first[rng.integers(0, pool, size=rows)]
    last_names = _last[rng.integers(0, pool, size=rows)]
    phones = _phone[rng.integers(0, pool, size=rows)]
    cities = _city[rng.integers(0, pool, size=rows)]

    customer_ids = np.arange(1_000_000, 1_000_000 + rows)
    states = rng.choice(STATES, size=rows)
    zip_codes = rng.integers(10_000, 100_000, size=rows).astype(str)

    start = np.datetime64("2020-01-01")
    span = (np.datetime64("2024-12-31") - start).astype(int)
    signup_dates = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[D]")

    lifetime_values = np.round(rng.gamma(2.0, 2.0, size=rows) * 2500, 2)
    is_active = rng.random(size=rows) < 0.85

    df = pl.DataFrame({
        "customer_id": customer_ids,
        "first_name": first_names,
        "last_name": last_names,
        "phone": phones,
        "city": cities,
        "state": states,
        "zip_code": zip_codes,
        "signup_date": signup_dates,
        "lifetime_value": lifetime_values,
        "is_active": is_active,
    })

    # Derived: email from first + last name
    df = df.with_columns(
        pl.concat_str([
            pl.col("first_name").str.to_lowercase(),
            pl.lit("."),
            pl.col("last_name").str.to_lowercase(),
            (pl.arange(0, pl.count()) % 1000).cast(pl.Utf8),
            pl.lit("@example.com"),
        ]).alias("email")
    )

    # Null injection: email 2%, phone 1%
    email_nulls = pl.Series(rng.random(rows) < 0.02)
    phone_nulls = pl.Series(rng.random(rows) < 0.01)
    df = df.with_columns(
        pl.when(email_nulls).then(None).otherwise(pl.col("email")).alias("email"),
        pl.when(phone_nulls).then(None).otherwise(pl.col("phone")).alias("phone"),
    )

    # Derived: loyalty_tier based on tenure
    from datetime import date
    today = date.today()
    return df.with_columns(
        pl.when((pl.lit(today) - pl.col("signup_date").cast(pl.Date)).dt.total_days() > 48 * 30)
        .then(pl.lit("Platinum"))
        .when((pl.lit(today) - pl.col("signup_date").cast(pl.Date)).dt.total_days() > 24 * 30)
        .then(pl.lit("Gold"))
        .when((pl.lit(today) - pl.col("signup_date").cast(pl.Date)).dt.total_days() > 12 * 30)
        .then(pl.lit("Silver"))
        .otherwise(pl.lit("Bronze"))
        .alias("loyalty_tier")
    )


def generate_products(rows: int = 5_000, seed: int = 42) -> pl.DataFrame:
    """Generate product catalog with categories, pricing, and cost data."""
    rng = np.random.default_rng(seed)

    product_ids = np.arange(10_000, 10_000 + rows)
    sku_nums = rng.integers(10_000_000, 100_000_000, size=rows)
    categories = rng.choice(CATEGORIES, size=rows)
    brands = rng.choice(BRANDS, size=rows)
    unit_prices = np.round(rng.uniform(5, 500, size=rows), 2)
    weights_kg = np.round(rng.uniform(0.1, 50, size=rows), 2)
    is_active = rng.random(size=rows) < 0.95

    start = np.datetime64("2018-01-01")
    span = (np.datetime64("2024-12-31") - start).astype(int)
    created_dates = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[D]")

    df = pl.DataFrame({
        "product_id": product_ids,
        "_sku_num": sku_nums,
        "category": categories,
        "brand": brands,
        "unit_price": unit_prices,
        "weight_kg": weights_kg,
        "is_active": is_active,
        "created_date": created_dates,
    })

    # Derived: SKU string, product_name, cost
    return df.with_columns(
        pl.format("SKU-{}", pl.col("_sku_num")).alias("sku"),
        pl.format("Product-{}", pl.col("product_id")).alias("product_name"),
        (pl.col("unit_price") * (0.4 + pl.Series(rng.random(rows)) * 0.3)).round(2).alias("cost"),
    ).drop("_sku_num")


def generate_transactions(rows: int = 50_000, n_customers: int = 10_000,
                          seed: int = 42) -> pl.DataFrame:
    """Generate retail transactions with payment methods and computed totals."""
    rng = np.random.default_rng(seed)

    txn_ids = np.arange(1, rows + 1)
    customer_ids = rng.integers(1_000_000, 1_000_000 + n_customers, size=rows)
    store_ids = rng.integers(100, 151, size=rows)

    start = np.datetime64("2024-01-01")
    span = int((np.datetime64("2024-12-31T23:59:59") - start) / np.timedelta64(1, "ms"))
    txn_timestamps = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")

    _pay_w = np.array(PAYMENT_WEIGHTS, dtype=np.float64)
    payment_methods = rng.choice(PAYMENT_METHODS, size=rows, p=_pay_w / _pay_w.sum())

    subtotals = np.round(rng.gamma(2.0, 2.0, size=rows) * 50, 2)
    discount_pcts = rng.beta(2, 5, size=rows) * 0.3
    items_counts = np.clip(np.floor(rng.exponential(5.0, size=rows)).astype(int), 1, 20)

    df = pl.DataFrame({
        "txn_id": txn_ids,
        "customer_id": customer_ids,
        "store_id": store_ids,
        "txn_timestamp": txn_timestamps,
        "payment_method": payment_methods,
        "_subtotal": subtotals,
        "_discount_pct": discount_pcts,
        "items_count": items_counts,
    })

    # Vectorized derived columns — replaces the 14-line for-loop
    disc_nulls = pl.Series(rng.random(rows) < 0.02)
    df = df.with_columns(
        (pl.col("_subtotal") * pl.col("_discount_pct")).round(2).alias("discount_amount"),
    ).with_columns(
        pl.when(disc_nulls).then(None).otherwise(pl.col("discount_amount")).alias("discount_amount"),
    ).with_columns(
        ((pl.col("_subtotal") - pl.col("discount_amount").fill_null(0)) * 0.08).round(2).alias("tax_amount"),
    ).with_columns(
        (pl.col("_subtotal") - pl.col("discount_amount").fill_null(0) + pl.col("tax_amount")).round(2).alias("total_amount"),
    ).drop("_subtotal", "_discount_pct")

    return df


def generate_line_items(rows: int = 150_000, n_transactions: int = 50_000,
                        n_products: int = 5_000, seed: int = 42) -> pl.DataFrame:
    """Generate transaction line items linking transactions to products."""
    rng = np.random.default_rng(seed)

    line_item_ids = np.arange(1, rows + 1)
    txn_ids = rng.integers(1, n_transactions + 1, size=rows)
    product_ids = rng.integers(10_000, 10_000 + n_products, size=rows)
    quantities = np.clip(np.floor(rng.exponential(3.3, size=rows)).astype(int), 1, 10)
    unit_prices = np.round(rng.uniform(5, 500, size=rows), 2)

    df = pl.DataFrame({
        "line_item_id": line_item_ids,
        "txn_id": txn_ids,
        "product_id": product_ids,
        "quantity": quantities,
        "unit_price": unit_prices,
    })

    return df.with_columns(
        (pl.col("quantity") * pl.col("unit_price")).round(2).alias("line_total")
    )


def generate_inventory(rows: int = 50_000, n_products: int = 5_000,
                       seed: int = 42) -> pl.DataFrame:
    """Generate inventory levels by product and store location."""
    rng = np.random.default_rng(seed)

    start = np.datetime64("2024-01-01")
    span = int((np.datetime64("2024-12-31T23:59:59") - start) / np.timedelta64(1, "ms"))

    inventory_ids = np.arange(1, rows + 1)
    product_ids = rng.integers(10_000, 10_000 + n_products, size=rows)
    store_ids = rng.integers(100, 151, size=rows)
    qty_on_hand = np.clip(np.floor(rng.exponential(100.0, size=rows)).astype(int), 0, 500)
    reorder_points = rng.integers(10, 51, size=rows)
    last_updated = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")

    return pl.DataFrame({
        "inventory_id": inventory_ids,
        "product_id": product_ids,
        "store_id": store_ids,
        "quantity_on_hand": qty_on_hand,
        "reorder_point": reorder_points,
        "last_updated": last_updated,
    })
