"""Healthcare industry synthetic data generators (Polars + NumPy + Mimesis, HIPAA-safe).

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

STATES = ["MA", "NY", "IL", "TX", "AZ", "CA", "FL", "PA"]
GENDERS = ["M", "F", "Other"]
GENDER_WEIGHTS = [49, 49, 2]
ENCOUNTER_TYPES = ["Outpatient", "Inpatient", "Emergency", "Observation", "Telehealth"]
ENCOUNTER_WEIGHTS = [50, 20, 15, 10, 5]
COMPLAINTS = [
    "Chest pain", "Shortness of breath", "Abdominal pain", "Headache",
    "Back pain", "Fever", "Cough", "Fatigue", "Dizziness", "Follow-up",
    "Routine checkup", "Medication refill", "Post-surgical", "Lab review",
]
PAYERS = ["Medicare", "Medicaid", "Blue Cross", "Aetna", "United", "Cigna", "Self-Pay"]
PAYER_WEIGHTS = [25, 15, 20, 15, 15, 8, 2]


def generate_patients(rows: int = 10_000, seed: int = 42) -> pl.DataFrame:
    """Generate synthetic patient demographics (100% HIPAA-safe)."""
    rng = np.random.default_rng(seed)
    g = Generic(locale=Locale.EN, seed=seed)

    pool = min(1_000, rows)
    _first = np.array([g.person.first_name() for _ in range(pool)])
    _last = np.array([g.person.last_name() for _ in range(pool)])
    _email = np.array([g.person.email() for _ in range(pool)])
    _phone = np.array([g.person.telephone() for _ in range(pool)])
    _address = np.array([g.address.address() for _ in range(pool)])
    _city = np.array([g.address.city() for _ in range(pool)])

    first_names = _first[rng.integers(0, pool, size=rows)]
    last_names = _last[rng.integers(0, pool, size=rows)]
    emails = _email[rng.integers(0, pool, size=rows)]
    phones = _phone[rng.integers(0, pool, size=rows)]
    addresses = _address[rng.integers(0, pool, size=rows)]
    cities = _city[rng.integers(0, pool, size=rows)]

    patient_ids = np.arange(1_000_000, 1_000_000 + rows)
    mrn_nums = rng.integers(1_000_000_000, 10_000_000_000, size=rows)
    ssn_last4 = rng.integers(0, 10_000, size=rows)

    dob_start = np.datetime64("1924-01-01")
    dob_span = (np.datetime64("2024-01-01") - dob_start).astype(int)
    dobs = dob_start + rng.integers(0, dob_span + 1, size=rows).astype("timedelta64[D]")

    _gen_w = np.array(GENDER_WEIGHTS, dtype=np.float64)
    genders = rng.choice(GENDERS, size=rows, p=_gen_w / _gen_w.sum())
    states = rng.choice(STATES, size=rows)
    zip_codes = rng.integers(10_000, 100_000, size=rows).astype(str)
    ins_nums = rng.integers(10_000_000, 100_000_000, size=rows)
    is_active = rng.random(size=rows) < 0.9

    df = pl.DataFrame({
        "patient_id": patient_ids,
        "_mrn_num": mrn_nums,
        "first_name": first_names,
        "last_name": last_names,
        "date_of_birth": dobs,
        "gender": genders,
        "_ssn_num": ssn_last4,
        "email": emails,
        "phone": phones,
        "address": addresses,
        "city": cities,
        "state": states,
        "zip_code": zip_codes,
        "_ins_num": ins_nums,
        "is_active": is_active,
    })

    # String formatting for MRN, SSN last4, insurance_id
    df = df.with_columns(
        pl.format("MRN{}", pl.col("_mrn_num")).alias("mrn"),
        pl.col("_ssn_num").cast(pl.Utf8).str.pad_start(4, "0").alias("ssn_last4"),
        pl.format("INS-{}", pl.col("_ins_num")).alias("insurance_id"),
    ).drop("_mrn_num", "_ssn_num", "_ins_num")

    # Null injection: email 5%, phone 2%
    email_nulls = pl.Series(rng.random(rows) < 0.05)
    phone_nulls = pl.Series(rng.random(rows) < 0.02)
    return df.with_columns(
        pl.when(email_nulls).then(None).otherwise(pl.col("email")).alias("email"),
        pl.when(phone_nulls).then(None).otherwise(pl.col("phone")).alias("phone"),
    )


def generate_encounters(rows: int = 30_000, n_patients: int = 10_000,
                        seed: int = 42) -> pl.DataFrame:
    """Generate clinical encounter data linked to patients."""
    rng = np.random.default_rng(seed)

    start = np.datetime64("2024-01-01")
    span = int((np.datetime64("2024-12-31T23:59:59") - start) / np.timedelta64(1, "ms"))

    encounter_ids = np.arange(1, rows + 1)
    patient_ids = rng.integers(1_000_000, 1_000_000 + n_patients, size=rows)
    provider_ids = rng.integers(5_000, 10_000, size=rows)
    facility_ids = rng.integers(100, 201, size=rows)

    _enc_w = np.array(ENCOUNTER_WEIGHTS, dtype=np.float64)
    encounter_types = rng.choice(ENCOUNTER_TYPES, size=rows, p=_enc_w / _enc_w.sum())

    admit_datetimes = start + rng.integers(0, span + 1, size=rows).astype("timedelta64[ms]")
    los_hours = np.clip(np.floor(rng.exponential(20.0, size=rows)).astype(int), 1, 168)

    _status_w = np.array([85, 10, 5], dtype=np.float64)
    statuses = rng.choice(["Completed", "Active", "Cancelled"], size=rows,
                          p=_status_w / _status_w.sum())
    chief_complaints = rng.choice(COMPLAINTS, size=rows)

    df = pl.DataFrame({
        "encounter_id": encounter_ids,
        "patient_id": patient_ids,
        "provider_id": provider_ids,
        "facility_id": facility_ids,
        "encounter_type": encounter_types,
        "admit_datetime": admit_datetimes,
        "_los_hours": los_hours,
        "status": statuses,
        "chief_complaint": chief_complaints,
    })

    # Vectorized discharge_datetime: admit + LOS, with 10% null mask
    discharge_nulls = pl.Series(rng.random(rows) < 0.1)
    df = df.with_columns(
        (pl.col("admit_datetime") + pl.duration(hours=pl.col("_los_hours")))
        .alias("discharge_datetime"),
    ).with_columns(
        pl.when(discharge_nulls).then(None)
        .otherwise(pl.col("discharge_datetime"))
        .alias("discharge_datetime"),
    ).drop("_los_hours")

    return df


def generate_claims(rows: int = 30_000, n_encounters: int = 30_000,
                    n_patients: int = 10_000, seed: int = 42) -> pl.DataFrame:
    """Generate insurance claims data linked to encounters and patients."""
    rng = np.random.default_rng(seed)

    svc_start = np.datetime64("2024-01-01")
    svc_span = (np.datetime64("2024-12-31") - svc_start).astype(int)

    claim_ids = np.arange(1, rows + 1)
    encounter_ids = rng.integers(1, n_encounters + 1, size=rows)
    patient_ids = rng.integers(1_000_000, 1_000_000 + n_patients, size=rows)

    _payer_w = np.array(PAYER_WEIGHTS, dtype=np.float64)
    payer_ids = rng.choice(PAYERS, size=rows, p=_payer_w / _payer_w.sum())

    _claim_w = np.array([70, 30], dtype=np.float64)
    claim_types = rng.choice(["Professional", "Institutional"], size=rows,
                             p=_claim_w / _claim_w.sum())

    service_dates = svc_start + rng.integers(0, svc_span + 1, size=rows).astype("timedelta64[D]")

    billed_amounts = np.clip(np.round(rng.exponential(1000.0, size=rows), 2), 0, 50_000)
    allowed_amounts = np.round(billed_amounts * (0.6 + rng.random(size=rows) * 0.3), 2)
    paid_amounts = np.round(allowed_amounts * (0.7 + rng.random(size=rows) * 0.3), 2)

    _stat_w = np.array([70, 15, 10, 5], dtype=np.float64)
    statuses = rng.choice(["Paid", "Pending", "Denied", "Appealed"], size=rows,
                          p=_stat_w / _stat_w.sum())

    df = pl.DataFrame({
        "claim_id": claim_ids,
        "encounter_id": encounter_ids,
        "patient_id": patient_ids,
        "payer_id": payer_ids,
        "claim_type": claim_types,
        "service_date": service_dates,
        "billed_amount": billed_amounts,
        "allowed_amount": allowed_amounts,
        "paid_amount": paid_amounts,
        "status": statuses,
    })

    # Null injection: paid_amount 5%
    paid_nulls = pl.Series(rng.random(rows) < 0.05)
    return df.with_columns(
        pl.when(paid_nulls).then(None).otherwise(pl.col("paid_amount")).alias("paid_amount"),
    )
