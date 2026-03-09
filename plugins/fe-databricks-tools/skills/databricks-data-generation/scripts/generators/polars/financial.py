"""Financial services synthetic data generators (Polars + NumPy).

REFERENCE IMPLEMENTATION — This file is not an importable module. Claude reads it
for patterns and adapts the code inline for user scripts. Uses Polars + NumPy for
vectorized Tier 1 local generation (<500K rows, zero JVM overhead).

NOTE: Seed reproducibility differs from the previous random-module version.
All randomness now flows through np.random.default_rng(seed).
"""

import numpy as np
import polars as pl

ACCOUNT_TYPES = ["Checking", "Savings", "Investment", "Retirement", "Credit Card", "Loan"]
ACCOUNT_TYPE_WEIGHTS = [30, 25, 20, 10, 10, 5]
CURRENCIES = ["USD", "EUR", "GBP", "JPY"]
CURRENCY_WEIGHTS = [70, 15, 10, 5]
SYMBOLS = ["AAPL", "GOOGL", "MSFT", "AMZN", "META", "NVDA", "TSLA", "JPM", "BAC", "WMT"]
ORDER_TYPES = ["Market", "Limit", "Stop", "Stop-Limit"]
ORDER_TYPE_WEIGHTS = [60, 25, 10, 5]
VENUES = ["NYSE", "NASDAQ", "ARCA", "BATS"]
VENUE_WEIGHTS = [35, 35, 15, 15]
TXN_TYPES = ["Deposit", "Withdrawal", "Transfer In", "Transfer Out",
             "Payment", "Fee", "Interest"]
TXN_TYPE_WEIGHTS = [20, 15, 15, 15, 25, 5, 5]
CHANNELS = ["Online", "Mobile", "Branch", "ATM", "Wire"]
CHANNEL_WEIGHTS = [35, 30, 15, 15, 5]


def generate_accounts(rows: int = 10_000, seed: int = 42) -> pl.DataFrame:
    """Generate financial account data with types, balances, and risk ratings."""
    rng = np.random.default_rng(seed)

    open_start = np.datetime64("2015-01-01")
    open_span = (np.datetime64("2024-12-31") - open_start).astype(int)

    account_ids = np.arange(1_000_000, 1_000_000 + rows)
    n_unique_customers = int(rows * 0.8)
    customer_ids = rng.integers(500_000, 500_000 + n_unique_customers, size=rows)
    account_numbers = rng.integers(100_000_000_000, 1_000_000_000_000, size=rows).astype(str)

    _acct_w = np.array(ACCOUNT_TYPE_WEIGHTS, dtype=np.float64)
    account_types = rng.choice(ACCOUNT_TYPES, size=rows, p=_acct_w / _acct_w.sum())

    _stat_w = np.array([85, 10, 5], dtype=np.float64)
    statuses = rng.choice(["Active", "Dormant", "Closed"], size=rows,
                          p=_stat_w / _stat_w.sum())

    open_dates = open_start + rng.integers(0, open_span + 1, size=rows).astype("timedelta64[D]")
    balances = np.round(np.maximum(0, rng.gamma(2.0, 2.0, size=rows) * 500_000), 2)

    _cur_w = np.array(CURRENCY_WEIGHTS, dtype=np.float64)
    currencies = rng.choice(CURRENCIES, size=rows, p=_cur_w / _cur_w.sum())

    df = pl.DataFrame({
        "account_id": account_ids,
        "customer_id": customer_ids,
        "account_number": account_numbers,
        "account_type": account_types,
        "status": statuses,
        "open_date": open_dates,
        "balance": balances,
        "currency": currencies,
    })

    return df.with_columns(
        pl.when(pl.col("balance") > 5_000_000).then(pl.lit("Low"))
        .when(pl.col("balance") > 500_000).then(pl.lit("Medium"))
        .otherwise(pl.lit("High"))
        .alias("risk_rating")
    )


def generate_trades(rows: int = 50_000, n_accounts: int = 10_000,
                    seed: int = 42) -> pl.DataFrame:
    """Generate trading activity data with execution details."""
    rng = np.random.default_rng(seed)

    start = np.datetime64("2024-01-02T09:30:00")
    end = np.datetime64("2024-12-31T16:00:00")
    span = int((end - start) / np.timedelta64(1, "ms"))

    trade_ids = np.arange(1, rows + 1)
    account_ids = rng.integers(1_000_000, 1_000_000 + n_accounts, size=rows)
    symbols = rng.choice(SYMBOLS, size=rows)

    _tt_w = np.array([55, 45], dtype=np.float64)
    trade_types = rng.choice(["BUY", "SELL"], size=rows, p=_tt_w / _tt_w.sum())

    _ot_w = np.array(ORDER_TYPE_WEIGHTS, dtype=np.float64)
    order_types = rng.choice(ORDER_TYPES, size=rows, p=_ot_w / _ot_w.sum())

    quantities = np.clip(np.floor(rng.exponential(500.0, size=rows)).astype(int), 1, 10_000)
    prices = np.round(rng.uniform(10, 1000, size=rows), 4)

    trade_timestamps = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")

    _ts_w = np.array([90, 5, 3, 2], dtype=np.float64)
    statuses = rng.choice(["Filled", "Partial", "Cancelled", "Rejected"], size=rows,
                          p=_ts_w / _ts_w.sum())

    _v_w = np.array(VENUE_WEIGHTS, dtype=np.float64)
    execution_venues = rng.choice(VENUES, size=rows, p=_v_w / _v_w.sum())

    df = pl.DataFrame({
        "trade_id": trade_ids,
        "account_id": account_ids,
        "symbol": symbols,
        "trade_type": trade_types,
        "order_type": order_types,
        "quantity": quantities,
        "price": prices,
        "trade_timestamp": trade_timestamps,
        "status": statuses,
        "execution_venue": execution_venues,
    })

    # Vectorized derived columns: trade_value, commission, volume_factor
    return df.with_columns(
        (pl.col("quantity").cast(pl.Float64) * pl.col("price")).round(2).alias("trade_value"),
    ).with_columns(
        pl.col("trade_value").clip(lower_bound=0.99).mul(0.001).round(2).alias("commission"),
        # Vectorized volume_factor — replaces the per-row for-loop
        pl.when(pl.col("trade_timestamp").dt.hour().is_between(9, 10))
        .then(3.0)
        .when(pl.col("trade_timestamp").dt.hour().is_between(15, 16))
        .then(2.5)
        .when(pl.col("trade_timestamp").dt.hour().is_between(11, 14))
        .then(1.0)
        .otherwise(0.3)
        .alias("volume_factor"),
    )


def generate_transactions(rows: int = 50_000, n_accounts: int = 10_000,
                          seed: int = 42) -> pl.DataFrame:
    """Generate financial transactions (deposits, withdrawals, transfers)."""
    rng = np.random.default_rng(seed)

    start = np.datetime64("2024-01-01")
    span = int((np.datetime64("2024-12-31T23:59:59") - start) / np.timedelta64(1, "ms"))

    txn_ids = np.arange(1, rows + 1)
    account_ids = rng.integers(1_000_000, 1_000_000 + n_accounts, size=rows)

    _txn_w = np.array(TXN_TYPE_WEIGHTS, dtype=np.float64)
    txn_types = rng.choice(TXN_TYPES, size=rows, p=_txn_w / _txn_w.sum())

    amounts = np.round(rng.gamma(2.0, 2.0, size=rows) * 5_000, 2)
    txn_timestamps = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")

    _ch_w = np.array(CHANNEL_WEIGHTS, dtype=np.float64)
    channels = rng.choice(CHANNELS, size=rows, p=_ch_w / _ch_w.sum())

    is_international = rng.random(size=rows) < 0.05

    _st_w = np.array([90, 5, 3, 2], dtype=np.float64)
    statuses = rng.choice(["Completed", "Pending", "Failed", "Reversed"], size=rows,
                          p=_st_w / _st_w.sum())

    return pl.DataFrame({
        "txn_id": txn_ids,
        "account_id": account_ids,
        "txn_type": txn_types,
        "amount": amounts,
        "txn_timestamp": txn_timestamps,
        "channel": channels,
        "is_international": is_international,
        "status": statuses,
    })
