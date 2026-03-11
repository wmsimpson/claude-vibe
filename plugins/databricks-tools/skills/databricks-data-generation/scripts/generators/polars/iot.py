"""IoT and telematics synthetic data generators (Polars + NumPy).

REFERENCE IMPLEMENTATION — This file is not an importable module. Claude reads it
for patterns and adapts the code inline for user scripts. Uses Polars + NumPy for
vectorized Tier 1 local generation (<500K rows, zero JVM overhead).

NOTE: Seed reproducibility differs from the previous random-module version.
All randomness now flows through np.random.default_rng(seed).
"""

import numpy as np
import polars as pl

DEVICE_TYPES = [
    "Temperature Sensor", "Humidity Sensor", "Pressure Sensor",
    "Motion Detector", "Smart Meter", "GPS Tracker", "Camera", "HVAC Controller",
]
MANUFACTURERS = ["SensorCorp", "IoTech", "SmartDevices", "TelcoSystems", "DataSense"]
METRICS = ["temperature", "humidity", "pressure", "power", "vibration"]
UNITS = {"temperature": "celsius", "humidity": "percent", "pressure": "hPa",
         "power": "kWh", "vibration": "mm/s"}
EVENT_TYPES = [
    "Threshold Exceeded", "Device Offline", "Firmware Update",
    "Calibration Required", "Battery Low", "Connection Lost", "Maintenance Due",
]


def generate_devices(rows: int = 1_000, seed: int = 42) -> pl.DataFrame:
    """Generate IoT device registry with metadata and geolocation."""
    rng = np.random.default_rng(seed)

    install_start = np.datetime64("2020-01-01")
    install_span = (np.datetime64("2024-12-31") - install_start).astype(int)

    device_ids = np.arange(1000, 1000 + rows)
    serial_nums = rng.integers(10_000_000, 100_000_000, size=rows)
    device_types = rng.choice(DEVICE_TYPES, size=rows)
    manufacturers = rng.choice(MANUFACTURERS, size=rows)
    install_dates = install_start + rng.integers(0, install_span + 1, size=rows).astype("timedelta64[D]")
    latitudes = rng.uniform(25.0, 48.0, size=rows)
    longitudes = rng.uniform(-125.0, -70.0, size=rows)

    _stat_w = np.array([85, 8, 5, 2], dtype=np.float64)
    statuses = rng.choice(
        ["Online", "Offline", "Maintenance", "Decommissioned"],
        size=rows, p=_stat_w / _stat_w.sum()
    )

    df = pl.DataFrame({
        "device_id": device_ids,
        "_serial_num": serial_nums,
        "device_type": device_types,
        "manufacturer": manufacturers,
        "install_date": install_dates,
        "latitude": latitudes,
        "longitude": longitudes,
        "status": statuses,
    })

    return df.with_columns(
        pl.format("DEV-{}", pl.col("_serial_num")).alias("device_serial"),
    ).drop("_serial_num")


def generate_sensor_readings(rows: int = 50_000, n_devices: int = 1_000,
                             anomaly_rate: float = 0.02,
                             seed: int = 42) -> pl.DataFrame:
    """Generate sensor reading time-series with sine wave patterns and anomaly injection.

    Temperature and power readings follow daily/annual sinusoidal cycles.
    Anomalies are injected at the specified rate with 1.5-2.5x amplification.
    """
    rng = np.random.default_rng(seed)

    start = np.datetime64("2024-01-01")
    span = int((np.datetime64("2024-12-31T23:59:59") - start) / np.timedelta64(1, "ms"))

    reading_ids = np.arange(1, rows + 1)
    device_ids = rng.integers(1000, 1000 + n_devices, size=rows)
    timestamps = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")
    metric_names = rng.choice(METRICS, size=rows)
    is_anomalies = rng.random(size=rows) < anomaly_rate

    # Build the DataFrame first so we can use Polars datetime extraction
    df = pl.DataFrame({
        "reading_id": reading_ids,
        "device_id": device_ids,
        "timestamp": timestamps,
        "metric_name": metric_names,
        "is_anomaly": is_anomalies,
    })

    # Extract hour and day_of_year via Polars, convert to numpy for vectorized math
    hours = df["timestamp"].dt.hour().to_numpy().astype(np.float64)
    day_of_year = df["timestamp"].dt.ordinal_day().to_numpy().astype(np.float64)
    idx = np.arange(rows, dtype=np.float64)

    # Vectorized sinusoidal cycles
    daily_cycle = 5 * np.sin(2 * np.pi * (hours - 6) / 24)
    annual_cycle = 15 * np.sin(2 * np.pi * (day_of_year - 80) / 365)
    noise = rng.uniform(-2.5, 2.5, size=rows)

    # Metric-specific base values (vectorized)
    metrics_np = df["metric_name"].to_numpy()
    base = np.zeros(rows)
    m_temp = metrics_np == "temperature"
    m_hum = metrics_np == "humidity"
    m_pres = metrics_np == "pressure"
    m_pow = metrics_np == "power"
    m_vib = metrics_np == "vibration"

    base[m_temp] = 20 + annual_cycle[m_temp] + daily_cycle[m_temp]
    base[m_hum] = 60 + 15 * np.sin(2 * np.pi * day_of_year[m_hum] / 365)
    base[m_pres] = 1013 + 10 * np.sin(2 * np.pi * day_of_year[m_pres] / 365)
    base[m_pow] = 50 + 20 * np.sin(2 * np.pi * (hours[m_pow] - 8) / 24)
    base[m_vib] = 5 + 2 * np.sin(0.1 * idx[m_vib])

    # Anomaly amplification factor
    factor = np.where(is_anomalies, 1.5 + rng.random(size=rows), 1.0)
    metric_values = np.round((base + noise) * factor, 4)

    # Null injection on metric_value: 0.5%
    val_nulls = rng.random(rows) < 0.005

    # Quality scores via beta distribution
    quality_scores = np.clip((rng.beta(5, 2, size=rows) * 20 + 80).astype(int), 80, 100)

    # Map metric names to units
    unit_map = np.array([UNITS[m] for m in METRICS])
    metric_indices = np.zeros(rows, dtype=int)
    for i, m in enumerate(METRICS):
        metric_indices[metrics_np == m] = i
    units = unit_map[metric_indices]

    df = df.with_columns(
        pl.Series("metric_value", metric_values),
        pl.Series("unit", units),
        pl.Series("quality_score", quality_scores),
    )

    # Apply null mask on metric_value
    return df.with_columns(
        pl.when(pl.Series(val_nulls)).then(None)
        .otherwise(pl.col("metric_value"))
        .alias("metric_value"),
    )


def generate_events(rows: int = 10_000, n_devices: int = 1_000,
                    seed: int = 42) -> pl.DataFrame:
    """Generate device events (alerts, maintenance, status changes)."""
    rng = np.random.default_rng(seed)

    start = np.datetime64("2024-01-01")
    span = int((np.datetime64("2024-12-31T23:59:59") - start) / np.timedelta64(1, "ms"))

    event_ids = np.arange(1, rows + 1)
    device_ids = rng.integers(1000, 1000 + n_devices, size=rows)
    timestamps = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")
    event_types = rng.choice(EVENT_TYPES, size=rows)

    _sev_w = np.array([10, 30, 60], dtype=np.float64)
    severities = rng.choice(["Critical", "Warning", "Info"], size=rows,
                            p=_sev_w / _sev_w.sum())

    acknowledged = rng.random(size=rows) < 0.7
    resolved = acknowledged & (rng.random(size=rows) < 0.8)

    df = pl.DataFrame({
        "event_id": event_ids,
        "device_id": device_ids,
        "timestamp": timestamps,
        "event_type": event_types,
        "severity": severities,
        "acknowledged": acknowledged,
        "resolved": resolved,
    })

    # Description with 5% null injection
    desc_nulls = pl.Series(rng.random(rows) < 0.05)
    return df.with_columns(
        pl.when(desc_nulls).then(None)
        .otherwise(pl.format("Event on device {}", pl.col("device_id")))
        .alias("description"),
    )


def generate_telemetry(rows: int = 50_000, n_devices: int = 1_000,
                       seed: int = 42) -> pl.DataFrame:
    """Generate GPS/vehicle telemetry data."""
    rng = np.random.default_rng(seed)

    start = np.datetime64("2024-01-01")
    span = int((np.datetime64("2024-12-31T23:59:59") - start) / np.timedelta64(1, "ms"))

    telemetry_ids = np.arange(1, rows + 1)
    device_ids = rng.integers(1000, 1000 + n_devices, size=rows)
    timestamps = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")

    base_lats = rng.uniform(37.0, 38.0, size=rows)
    base_lons = rng.uniform(-122.5, -121.5, size=rows)
    latitudes = base_lats + rng.uniform(-0.005, 0.005, size=rows)
    longitudes = base_lons + rng.uniform(-0.005, 0.005, size=rows)

    speeds = rng.uniform(0, 80, size=rows)
    headings = rng.integers(0, 360, size=rows)

    fuel_levels = np.round(rng.uniform(10, 100, size=rows), 1)
    engine_temps = np.round(rng.normal(200, 10, size=rows), 1)

    df = pl.DataFrame({
        "telemetry_id": telemetry_ids,
        "device_id": device_ids,
        "timestamp": timestamps,
        "latitude": latitudes,
        "longitude": longitudes,
        "speed": speeds,
        "heading": headings,
        "fuel_level": fuel_levels,
        "engine_temp": engine_temps,
    })

    # Null injection: fuel_level 2%, engine_temp 1%
    fuel_nulls = pl.Series(rng.random(rows) < 0.02)
    temp_nulls = pl.Series(rng.random(rows) < 0.01)
    return df.with_columns(
        pl.when(fuel_nulls).then(None).otherwise(pl.col("fuel_level")).alias("fuel_level"),
        pl.when(temp_nulls).then(None).otherwise(pl.col("engine_temp")).alias("engine_temp"),
    )
