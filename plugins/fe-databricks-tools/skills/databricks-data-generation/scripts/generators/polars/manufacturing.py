"""Manufacturing industry synthetic data generators (Polars + NumPy).

REFERENCE IMPLEMENTATION — This file is not an importable module. Claude reads it
for patterns and adapts the code inline for user scripts. Uses Polars + NumPy for
vectorized Tier 1 local generation (<500K rows, zero JVM overhead).

NOTE: Seed reproducibility differs from the previous random-module version.
All randomness now flows through np.random.default_rng(seed).
"""

import numpy as np
import polars as pl

EQUIPMENT_TYPES = [
    "CNC Machine", "Assembly Robot", "Conveyor Belt", "Press",
    "Welder", "Packaging Unit", "Quality Scanner", "HVAC System",
]
EQUIPMENT_MANUFACTURERS = ["Siemens", "ABB", "Fanuc", "Bosch", "Honeywell"]
ZONES = ["Zone A", "Zone B", "Zone C", "Zone D", "Zone E", "Zone F"]
MAINTENANCE_TYPES = ["Preventive", "Corrective", "Predictive", "Emergency"]
MAINTENANCE_TYPE_WEIGHTS = [50, 30, 15, 5]
PRIORITIES = ["Critical", "High", "Medium", "Low"]
PRIORITY_WEIGHTS = [5, 20, 50, 25]


def generate_equipment(rows: int = 500, seed: int = 42) -> pl.DataFrame:
    """Generate equipment/asset registry for manufacturing floor."""
    rng = np.random.default_rng(seed)

    install_start = np.datetime64("2015-01-01")
    install_span = (np.datetime64("2024-12-31") - install_start).astype(int)
    maint_start = np.datetime64("2024-01-01")
    maint_span = (np.datetime64("2024-12-31") - maint_start).astype(int)

    equipment_ids = np.arange(1000, 1000 + rows)
    serial_nums = rng.integers(10_000_000, 100_000_000, size=rows)
    equipment_types = rng.choice(EQUIPMENT_TYPES, size=rows)
    manufacturers = rng.choice(EQUIPMENT_MANUFACTURERS, size=rows)
    install_dates = install_start + rng.integers(0, install_span + 1, size=rows).astype("timedelta64[D]")
    location_zones = rng.choice(ZONES, size=rows)

    _stat_w = np.array([80, 10, 7, 3], dtype=np.float64)
    statuses = rng.choice(
        ["Operational", "Maintenance", "Idle", "Decommissioned"],
        size=rows, p=_stat_w / _stat_w.sum()
    )

    last_maint_dates = maint_start + rng.integers(0, maint_span + 1, size=rows).astype("timedelta64[D]")
    expected_lifespans = np.clip(rng.normal(15, 5, size=rows).astype(int), 5, 25)

    df = pl.DataFrame({
        "equipment_id": equipment_ids,
        "_serial_num": serial_nums,
        "equipment_type": equipment_types,
        "manufacturer": manufacturers,
        "install_date": install_dates,
        "location_zone": location_zones,
        "status": statuses,
        "last_maintenance_date": last_maint_dates,
        "expected_lifespan_years": expected_lifespans,
    })

    return df.with_columns(
        pl.format("EQ-{}", pl.col("_serial_num")).alias("serial_number"),
        pl.format("Model-{}", pl.col("equipment_id")).alias("model"),
    ).drop("_serial_num")


def generate_sensor_data(rows: int = 25_000, n_equipment: int = 500,
                         fault_rate: float = 0.10, seed: int = 42) -> pl.DataFrame:
    """Generate multi-sensor time-series with sine wave patterns and fault injection.

    Based on industrial IoT patterns (wind turbine vibration monitoring).
    Faulty equipment (~fault_rate of total) gets 15% outlier readings with
    sigma multiplied by 8-20x.
    """
    rng = np.random.default_rng(seed)

    start = np.datetime64("2024-01-01")
    span = int((np.datetime64("2024-12-31T23:59:59") - start) / np.timedelta64(1, "ms"))
    n_faulty = int(n_equipment * fault_rate)
    faulty_max_id = 1000 + n_faulty - 1

    reading_ids = np.arange(1, rows + 1)
    equipment_ids = rng.integers(1000, 1000 + n_equipment, size=rows)
    timestamps = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")
    idx = np.arange(rows, dtype=np.float64)

    # Vectorized fault/outlier logic
    is_faulty = equipment_ids <= faulty_max_id
    is_outlier = is_faulty & (rng.random(size=rows) < 0.15)
    mult = np.where(is_outlier, 8 + rng.random(size=rows) * 12, 1.0)

    # Vectorized sensor values with sinusoidal patterns
    sensor_a = np.round(rng.uniform(-1, 1, size=rows) * 1 * mult, 4)
    sensor_b = np.round(rng.uniform(-1, 1, size=rows) * 2 * mult, 4)
    sensor_c = np.round(rng.uniform(-1, 1, size=rows) * 3 * mult, 4)
    sensor_d = np.round((2 * np.exp(np.sin(0.1 * idx)) + rng.uniform(-1.5, 1.5, size=rows)) * mult, 4)
    sensor_e = np.round((2 * np.exp(np.sin(0.01 * idx)) + rng.uniform(-2, 2, size=rows)) * mult, 4)
    sensor_f = np.round((2 * np.exp(np.sin(0.2 * idx)) + rng.uniform(-1, 1, size=rows)) * mult, 4)

    # Vectorized energy with 0.5% null injection
    energy = np.round(rng.uniform(0, 100, size=rows), 2)
    energy_nulls = rng.random(rows) < 0.005

    # Vectorized anomaly detection via threshold checks
    is_anomaly = (
        (np.abs(sensor_a) > 9) | (np.abs(sensor_b) > 18) | (np.abs(sensor_c) > 27) |
        (np.abs(sensor_d) > 12) | (np.abs(sensor_e) > 15) | (np.abs(sensor_f) > 12)
    )

    df = pl.DataFrame({
        "reading_id": reading_ids,
        "equipment_id": equipment_ids,
        "timestamp": timestamps,
        "sensor_A": sensor_a,
        "sensor_B": sensor_b,
        "sensor_C": sensor_c,
        "sensor_D": sensor_d,
        "sensor_E": sensor_e,
        "sensor_F": sensor_f,
        "energy": energy,
        "is_anomaly": is_anomaly,
    })

    # Apply energy null mask
    return df.with_columns(
        pl.when(pl.Series(energy_nulls)).then(None)
        .otherwise(pl.col("energy"))
        .alias("energy"),
    )


def generate_maintenance_records(rows: int = 5_000, n_equipment: int = 500,
                                 seed: int = 42) -> pl.DataFrame:
    """Generate maintenance work order and event records."""
    rng = np.random.default_rng(seed)

    start = np.datetime64("2024-01-01")
    span = int((np.datetime64("2024-12-31T23:59:59") - start) / np.timedelta64(1, "ms"))

    maintenance_ids = np.arange(1, rows + 1)
    equipment_ids = rng.integers(1000, 1000 + n_equipment, size=rows)
    scheduled_dates = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")
    los_hours = np.clip(np.floor(rng.exponential(20.0, size=rows)).astype(int), 1, 72)

    _mt_w = np.array(MAINTENANCE_TYPE_WEIGHTS, dtype=np.float64)
    maintenance_types = rng.choice(MAINTENANCE_TYPES, size=rows, p=_mt_w / _mt_w.sum())

    _pr_w = np.array(PRIORITY_WEIGHTS, dtype=np.float64)
    priorities = rng.choice(PRIORITIES, size=rows, p=_pr_w / _pr_w.sum())

    duration_hours = np.clip(np.floor(rng.exponential(20.0, size=rows)).astype(int), 1, 72)
    costs = np.clip(np.round(np.maximum(100, rng.exponential(10_000.0, size=rows)), 2), 100, 50_000)
    technician_ids = rng.integers(100, 201, size=rows)

    _st_w = np.array([80, 10, 8, 2], dtype=np.float64)
    statuses = rng.choice(
        ["Completed", "In Progress", "Scheduled", "Cancelled"],
        size=rows, p=_st_w / _st_w.sum()
    )
    parts_replaced = np.clip(np.floor(rng.exponential(2.0, size=rows)).astype(int), 0, 10)

    df = pl.DataFrame({
        "maintenance_id": maintenance_ids,
        "equipment_id": equipment_ids,
        "scheduled_date": scheduled_dates,
        "_los_hours": los_hours,
        "maintenance_type": maintenance_types,
        "priority": priorities,
        "duration_hours": duration_hours,
        "cost": costs,
        "technician_id": technician_ids,
        "status": statuses,
        "parts_replaced": parts_replaced,
    })

    # Vectorized completion_date: scheduled + LOS, with 12% null mask
    comp_nulls = pl.Series(rng.random(rows) < 0.12)
    return df.with_columns(
        pl.format("WO-{}", pl.col("equipment_id")).alias("description"),
        (pl.col("scheduled_date") + pl.duration(hours=pl.col("_los_hours")))
        .alias("completion_date"),
    ).with_columns(
        pl.when(comp_nulls).then(None)
        .otherwise(pl.col("completion_date"))
        .alias("completion_date"),
    ).drop("_los_hours")
