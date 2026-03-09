"""Cross-industry CDC helpers for Polars-based synthetic data generation.

REFERENCE IMPLEMENTATION — This file is not an importable module. Claude reads it
for patterns and adapts the code inline for user scripts. Uses pure Polars + NumPy
for Tier 1 local CDC generation (zero JVM overhead).

NOTE: Seed reproducibility differs from the previous random-module version.
All randomness now flows through np.random.default_rng(seed).
"""

import numpy as np
from datetime import datetime

import polars as pl


def add_cdc_operations(df: pl.DataFrame, weights: dict[str, int] | None = None,
                       seed: int = 42) -> pl.DataFrame:
    """Add CDC operation and operation_date columns to a Polars DataFrame.

    Args:
        df: Source DataFrame
        weights: Dict of operation weights, e.g. {"APPEND": 50, "UPDATE": 30, "DELETE": 10}
        seed: Random seed for reproducibility
    """
    if weights is None:
        weights = {"APPEND": 50, "UPDATE": 30, "DELETE": 10}

    rng = np.random.default_rng(seed)
    ops = np.array(list(weights.keys()))
    wts = np.array(list(weights.values()), dtype=np.float64)
    operations = rng.choice(ops, size=len(df), p=wts / wts.sum())

    return df.with_columns(
        pl.Series("operation", operations),
        pl.lit(datetime.now()).alias("operation_date"),
    )
