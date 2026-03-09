# Healthcare Industry Patterns

HIPAA-safe synthetic data models for healthcare demos.

**Important**: All data generated is 100% synthetic and contains no real patient information.

## Data Model Overview

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Patients   │────<│  Encounters  │>────│  Providers   │
└──────────────┘     └──────────────┘     └──────────────┘
       │                    │                    │
       │                    │                    │
       ▼                    ▼                    ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Insurance   │     │    Claims    │     │  Facilities  │
└──────────────┘     └──────────────┘     └──────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │  Diagnoses   │
                    └──────────────┘
```

## Table Schemas

### Patients

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `patient_id` | LONG | Primary key | Unique |
| `mrn` | STRING | Medical Record Number | Template: `MRNdddddddddd` |
| `first_name` | STRING | First name | Mimesis |
| `last_name` | STRING | Last name | Mimesis |
| `date_of_birth` | DATE | Birth date | Age 0-100 distribution |
| `gender` | STRING | M/F/Other | Weighted values |
| `ssn_last4` | STRING | Last 4 of SSN | Template: `dddd` |
| `address` | STRING | Street address | Mimesis |
| `city` | STRING | City | Mimesis |
| `state` | STRING | State | Values list |
| `zip_code` | STRING | ZIP | Template |
| `phone` | STRING | Phone | Template |
| `email` | STRING | Email | Template |
| `insurance_id` | STRING | Insurance member ID | Template |
| `pcp_id` | LONG | Primary care provider | FK to providers |

```python
import dbldatagen as dg
from utils.mimesis_text import mimesisText
from datetime import date, timedelta

patients = (
    dg.DataGenerator(spark, rows=100_000, partitions=10)
    .withColumn("patient_id", "long", minValue=1_000_000, uniqueValues=100_000)
    .withColumn("mrn", "string", template=r"MRNdddddddddd")
    .withColumn("first_name", "string", text=mimesisText("person.first_name"))
    .withColumn("last_name", "string", text=mimesisText("person.last_name"))
    .withColumn("date_of_birth", "date", begin="1924-01-01", end="2024-01-01", random=True)
    .withColumn("gender", "string", values=["M", "F", "Other"], weights=[49, 49, 2])
    .withColumn("ssn_last4", "string", template=r"dddd")
    .withColumn("address", "string", text=mimesisText("address.address"))
    .withColumn("city", "string", values=["Boston", "New York", "Chicago", "Houston", "Phoenix"])
    .withColumn("state", "string", values=["MA", "NY", "IL", "TX", "AZ"])
    .withColumn("zip_code", "string", template=r"ddddd")
    .withColumn("phone", "string", text=mimesisText("person.telephone"))
    .withColumn("email", "string", text=mimesisText("person.email"))
    .withColumn("insurance_id", "string", template=r"INS-dddddddd")
    .withColumn("pcp_id", "long", minValue=5000, maxValue=5999)
    .build()
)
```

### Providers

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `provider_id` | LONG | Primary key | Unique |
| `npi` | STRING | National Provider ID | Template: 10 digits |
| `first_name` | STRING | First name | Mimesis |
| `last_name` | STRING | Last name | Mimesis |
| `specialty` | STRING | Medical specialty | Values list |
| `facility_id` | LONG | Primary facility | FK |
| `license_state` | STRING | State | Values list |
| `active` | BOOLEAN | Currently practicing | 95% true |

```python
import dbldatagen as dg
from utils.mimesis_text import mimesisText

specialties = [
    "Family Medicine", "Internal Medicine", "Pediatrics", "Cardiology",
    "Orthopedics", "Dermatology", "Neurology", "Psychiatry",
    "Oncology", "Emergency Medicine", "Radiology", "Anesthesiology"
]

providers = (
    dg.DataGenerator(spark, rows=5_000, partitions=4)
    .withColumn("provider_id", "long", minValue=5_000, uniqueValues=5_000)
    .withColumn("npi", "string", template=r"dddddddddd")
    .withColumn("first_name", "string", text=mimesisText("person.first_name"))
    .withColumn("last_name", "string", text=mimesisText("person.last_name"))
    .withColumn("specialty", "string", values=specialties)
    .withColumn("facility_id", "long", minValue=100, maxValue=200)
    .withColumn("license_state", "string", values=["MA", "NY", "CA", "TX", "FL"])
    .withColumn("active", "boolean", expr="rand() < 0.95")
    .build()
)
```

### Encounters

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `encounter_id` | LONG | Primary key | Unique |
| `patient_id` | LONG | Foreign key | Match patient range |
| `provider_id` | LONG | Attending provider | FK |
| `facility_id` | LONG | Location | FK |
| `encounter_type` | STRING | Visit type | Values list |
| `admit_datetime` | TIMESTAMP | Admission time | Random |
| `discharge_datetime` | TIMESTAMP | Discharge time | After admit |
| `status` | STRING | Encounter status | Values list |
| `chief_complaint` | STRING | Reason for visit | Values list |

```python
encounter_types = ["Outpatient", "Inpatient", "Emergency", "Observation", "Telehealth"]
complaints = [
    "Chest pain", "Shortness of breath", "Abdominal pain", "Headache",
    "Back pain", "Fever", "Cough", "Fatigue", "Dizziness", "Follow-up"
]

encounters = (
    dg.DataGenerator(spark, rows=500_000, partitions=20)
    .withColumn("encounter_id", "long", minValue=1, uniqueValues=500_000)
    .withColumn("patient_id", "long", minValue=1_000_000, maxValue=1_099_999)
    .withColumn("provider_id", "long", minValue=5_000, maxValue=9_999)
    .withColumn("facility_id", "long", minValue=100, maxValue=200)
    .withColumn("encounter_type", "string", values=encounter_types, weights=[50, 20, 15, 10, 5])
    .withColumn("admit_datetime", "timestamp", begin="2024-01-01", end="2024-12-31", random=True)
    .withColumn("los_hours", "integer", minValue=1, maxValue=168, distribution="exponential", omit=True)
    .withColumn("discharge_datetime", "timestamp", expr="admit_datetime + interval los_hours hours")
    .withColumn("status", "string", values=["Completed", "Active", "Cancelled"], weights=[85, 10, 5])
    .withColumn("chief_complaint", "string", values=complaints)
    .build()
)
```

### Claims

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `claim_id` | LONG | Primary key | Unique |
| `encounter_id` | LONG | Foreign key | Match encounter range |
| `patient_id` | LONG | Foreign key | Match patient range |
| `payer_id` | STRING | Insurance company | Values list |
| `claim_type` | STRING | Professional/Institutional | Values |
| `service_date` | DATE | Date of service | From encounter |
| `submitted_date` | DATE | Claim submission | After service |
| `paid_date` | DATE | Payment date | After submitted |
| `billed_amount` | DECIMAL(12,2) | Amount billed | Range |
| `allowed_amount` | DECIMAL(12,2) | Contracted amount | 60-90% of billed |
| `paid_amount` | DECIMAL(12,2) | Paid by payer | 70-100% of allowed |
| `patient_responsibility` | DECIMAL(10,2) | Patient owes | Remainder |
| `status` | STRING | Claim status | Values list |

```python
payers = ["Medicare", "Medicaid", "Blue Cross", "Aetna", "United", "Cigna", "Self-Pay"]

claims = (
    dg.DataGenerator(spark, rows=1_000_000, partitions=20)
    .withColumn("claim_id", "long", minValue=1, uniqueValues=1_000_000)
    .withColumn("encounter_id", "long", minValue=1, maxValue=500_000)
    .withColumn("patient_id", "long", minValue=1_000_000, maxValue=1_099_999)
    .withColumn("payer_id", "string", values=payers, weights=[25, 15, 20, 15, 15, 8, 2])
    .withColumn("claim_type", "string", values=["Professional", "Institutional"], weights=[70, 30])
    .withColumn("service_date", "date", begin="2024-01-01", end="2024-12-31", random=True)
    .withColumn("submitted_date", "date", expr="date_add(service_date, cast(rand() * 30 as int))")
    .withColumn("paid_date", "date", expr="date_add(submitted_date, cast(rand() * 45 as int))")
    .withColumn("billed_amount", "decimal(12,2)", minValue=50, maxValue=50000, distribution="exponential")
    .withColumn("allowed_pct", "float", expr="0.6 + rand() * 0.3", omit=True)
    .withColumn("allowed_amount", "decimal(12,2)", expr="billed_amount * allowed_pct")
    .withColumn("paid_pct", "float", expr="0.7 + rand() * 0.3", omit=True)
    .withColumn("paid_amount", "decimal(12,2)", expr="allowed_amount * paid_pct")
    .withColumn("patient_responsibility", "decimal(10,2)", expr="allowed_amount - paid_amount")
    .withColumn("status", "string", values=["Paid", "Pending", "Denied", "Appealed"], weights=[70, 15, 10, 5])
    .build()
)
```

### Diagnoses

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `diagnosis_id` | LONG | Primary key | Unique |
| `encounter_id` | LONG | Foreign key | Match encounter range |
| `patient_id` | LONG | Foreign key | Match patient range |
| `icd10_code` | STRING | ICD-10 code | Pattern |
| `description` | STRING | Diagnosis text | Values list |
| `is_primary` | BOOLEAN | Primary diagnosis | First per encounter |
| `diagnosis_type` | STRING | Admitting/Final | Values |

```python
# Common ICD-10 patterns
icd10_codes = [
    ("I10", "Essential hypertension"),
    ("E11.9", "Type 2 diabetes mellitus"),
    ("J06.9", "Upper respiratory infection"),
    ("M54.5", "Low back pain"),
    ("R10.9", "Abdominal pain"),
    ("R05.9", "Cough"),
    ("J18.9", "Pneumonia"),
    ("N39.0", "Urinary tract infection"),
    ("K21.0", "GERD"),
    ("F41.1", "Generalized anxiety disorder"),
]

diagnoses = (
    dg.DataGenerator(spark, rows=1_500_000, partitions=20)  # ~3 per encounter
    .withColumn("diagnosis_id", "long", minValue=1, uniqueValues=1_500_000)
    .withColumn("encounter_id", "long", minValue=1, maxValue=500_000)
    .withColumn("patient_id", "long", minValue=1_000_000, maxValue=1_099_999)
    .withColumn("code_idx", "integer", minValue=0, maxValue=9, random=True, omit=True)
    .withColumn("icd10_code", "string", values=[c[0] for c in icd10_codes])
    .withColumn("description", "string", values=[c[1] for c in icd10_codes])
    .withColumn("is_primary", "boolean", expr="rand() < 0.33")
    .withColumn("diagnosis_type", "string", values=["Admitting", "Working", "Final"], weights=[20, 30, 50])
    .build()
)
```

## HIPAA Considerations

### What Makes Data HIPAA-Safe

This synthetic data is HIPAA-safe because:

1. **No Real Identifiers**: All names, SSNs, MRNs are randomly generated
2. **No Real Dates**: Birth dates and encounter dates are synthetic
3. **No Geographic Detail**: ZIP codes are random 5-digit strings
4. **No Real Insurance IDs**: Member IDs are randomly generated
5. **No Real NPIs**: Provider NPIs are random 10-digit strings

### Patterns to Avoid

```python
# BAD: Using real data as seed
real_patients = spark.table("production.patients")  # Never do this

# GOOD: Fully synthetic
patients = dg.DataGenerator(spark, rows=100000).build()
```

## CDC Generation

```python
def generate_healthcare_cdc(spark, volume_path, n_encounters=500_000, n_batches=5, seed=42):
    """Generate healthcare CDC data — encounter updates and claim status changes."""
    import dbldatagen as dg

    # Initial encounter load — all INSERTs
    initial = (
        dg.DataGenerator(spark, rows=n_encounters, partitions=20, randomSeed=seed)
        .withColumn("encounter_id", "long", minValue=1, uniqueValues=n_encounters)
        .withColumn("patient_id", "long", minValue=1_000_000, maxValue=1_099_999)
        .withColumn("provider_id", "long", minValue=5_000, maxValue=9_999)
        .withColumn("encounter_type", "string",
                    values=["Outpatient", "Inpatient", "Emergency", "Observation", "Telehealth"],
                    weights=[50, 20, 15, 10, 5])
        .withColumn("status", "string", values=["Active"])
        .withColumn("operation", "string", values=["INSERT"])
        .withColumn("operation_date", "timestamp", begin="2024-01-01", end="2024-01-01")
        .build()
    )
    initial.write.format("json").mode("overwrite").save(f"{volume_path}/encounters/batch_0")

    # Incremental batches — encounter updates (status changes) and new encounters
    for batch in range(1, n_batches + 1):
        batch_df = (
            dg.DataGenerator(spark, rows=n_encounters // 10, partitions=4, randomSeed=seed + batch)
            .withColumn("encounter_id", "long", minValue=1, maxValue=n_encounters)
            .withColumn("patient_id", "long", minValue=1_000_000, maxValue=1_099_999)
            .withColumn("provider_id", "long", minValue=5_000, maxValue=9_999)
            .withColumn("encounter_type", "string",
                        values=["Outpatient", "Inpatient", "Emergency", "Observation", "Telehealth"],
                        weights=[50, 20, 15, 10, 5])
            .withColumn("status", "string",
                        values=["Active", "Completed", "Cancelled"],
                        weights=[20, 70, 10])
            .withColumn("operation", "string", values=["INSERT", "UPDATE"], weights=[30, 70])
            .withColumn("operation_date", "timestamp",
                        expr=f"cast('2024-01-{batch + 1:02d}' as timestamp) + make_interval(0,0,0,0, cast(rand() * 23 as int), cast(rand() * 59 as int), 0)")
            .build()
        )
        batch_df.write.format("json").mode("overwrite").save(f"{volume_path}/encounters/batch_{batch}")

    # Claim status changes — Synthea-style encounter-to-claims temporal flow
    for batch in range(1, n_batches + 1):
        claims_cdc = (
            dg.DataGenerator(spark, rows=n_encounters // 5, partitions=4, randomSeed=seed + batch + 100)
            .withColumn("claim_id", "long", minValue=1, maxValue=n_encounters * 2)
            .withColumn("encounter_id", "long", minValue=1, maxValue=n_encounters)
            .withColumn("status", "string",
                        values=["Submitted", "In Review", "Paid", "Denied", "Appealed"],
                        weights=[20, 25, 35, 15, 5])
            .withColumn("paid_amount", "decimal(12,2)", minValue=0, maxValue=25000, distribution="exponential")
            .withColumn("operation", "string", values=["INSERT", "UPDATE"], weights=[40, 60])
            .withColumn("operation_date", "timestamp",
                        expr=f"cast('2024-01-{batch + 1:02d}' as timestamp) + make_interval(0,0,0,0, cast(rand() * 23 as int), cast(rand() * 59 as int), 0)")
            .build()
        )
        claims_cdc.write.format("json").mode("overwrite").save(f"{volume_path}/claims/batch_{batch}")
```

## Data Quality Injection

```python
from utils.mimesis_text import mimesisText

# Healthcare-appropriate quality issues
# 1% null DOB (incomplete patient intake), 3% null addresses
.withColumn("date_of_birth", "date", begin="1924-01-01", end="2024-01-01", random=True, percentNulls=0.01)
.withColumn("address", "string", text=mimesisText("address.address"), percentNulls=0.03)

# Duplicate encounter records (system integration issues — 2% rate)
clean_encounters = generate_encounters(spark, n_encounters)
encounters_with_dupes = clean_encounters.union(clean_encounters.sample(0.02))
```

## Medallion Output

```python
# Bronze: Raw to Volumes — Auto Loader picks up
patients_df.write.format("json").save(f"/Volumes/{catalog}/{schema}/{volume}/patients")
encounters_df.write.format("json").save(f"/Volumes/{catalog}/{schema}/{volume}/encounters")
claims_df.write.format("json").save(f"/Volumes/{catalog}/{schema}/{volume}/claims")

# Silver: Cleaned and deduplicated
# (handled by Spark Declarative Pipeline — APPLY CHANGES INTO for CDC)

# Gold: Patient summaries and aggregated metrics
# (materialized views in Spark Declarative Pipelines)
```

## Synthea-Style Patterns

```python
# Encounter-to-claims temporal flow:
# 1. Encounter created (admit_datetime)
# 2. Encounter completed (discharge_datetime = admit + LOS)
# 3. Claim submitted (2-30 days after discharge)
# 4. Claim adjudicated (15-45 days after submission)

.withColumn("admit_datetime", "timestamp", begin="2024-01-01", end="2024-12-31", random=True)
.withColumn("los_hours", "integer", minValue=1, maxValue=168, distribution="exponential", omit=True)
.withColumn("discharge_datetime", "timestamp", expr="admit_datetime + make_interval(0,0,0,0, los_hours, 0, 0)")
.withColumn("claim_lag_days", "integer", minValue=2, maxValue=30, random=True, omit=True)
.withColumn("claim_submitted_date", "date", expr="date_add(cast(discharge_datetime as date), claim_lag_days)")
.withColumn("adjudication_lag_days", "integer", minValue=15, maxValue=45, random=True, omit=True)
.withColumn("claim_paid_date", "date", expr="date_add(claim_submitted_date, adjudication_lag_days)")

# Age-weighted encounter frequency (elderly patients have more encounters)
.withColumn("patient_age", "integer",
            expr="floor(datediff(current_date(), date_of_birth) / 365)", omit=True)
.withColumn("encounter_weight", "float",
            expr="""case
                when patient_age >= 65 then 3.0
                when patient_age >= 45 then 1.5
                when patient_age >= 18 then 1.0
                else 1.2
            end""")
```

## Complete Healthcare Demo

```python
def generate_healthcare_demo(
    spark,
    n_patients: int = 100_000,
    n_providers: int = 5_000,
    n_encounters: int = 500_000,
    catalog: str = "demo",
    schema: str = "healthcare"
):
    """Generate HIPAA-safe healthcare demo dataset."""

    patients = generate_patients(spark, n_patients)
    providers = generate_providers(spark, n_providers)
    facilities = generate_facilities(spark, 100)
    encounters = generate_encounters(spark, n_encounters, n_patients, n_providers)
    claims = generate_claims(spark, n_encounters * 2, n_encounters, n_patients)
    diagnoses = generate_diagnoses(spark, n_encounters * 3, n_encounters, n_patients)

    tables = {
        "patients": patients,
        "providers": providers,
        "facilities": facilities,
        "encounters": encounters,
        "claims": claims,
        "diagnoses": diagnoses,
    }

    for name, df in tables.items():
        df.write.format("delta").mode("overwrite").saveAsTable(f"{catalog}.{schema}.{name}")

    return tables
```

## Common Demo Queries

### Patient Summary
```sql
SELECT
    p.patient_id,
    p.first_name,
    p.last_name,
    COUNT(DISTINCT e.encounter_id) as total_encounters,
    COUNT(DISTINCT d.icd10_code) as unique_diagnoses,
    SUM(c.paid_amount) as total_claims_paid
FROM patients p
LEFT JOIN encounters e ON p.patient_id = e.patient_id
LEFT JOIN diagnoses d ON e.encounter_id = d.encounter_id
LEFT JOIN claims c ON e.encounter_id = c.encounter_id
GROUP BY p.patient_id, p.first_name, p.last_name
```

### Top Diagnoses
```sql
SELECT
    d.icd10_code,
    d.description,
    COUNT(*) as occurrence_count,
    COUNT(DISTINCT d.patient_id) as unique_patients
FROM diagnoses d
GROUP BY d.icd10_code, d.description
ORDER BY occurrence_count DESC
LIMIT 20
```

### Claims by Payer
```sql
SELECT
    payer_id,
    COUNT(*) as claim_count,
    SUM(billed_amount) as total_billed,
    SUM(paid_amount) as total_paid,
    AVG(paid_amount / billed_amount) * 100 as avg_payment_rate
FROM claims
WHERE status = 'Paid'
GROUP BY payer_id
ORDER BY total_paid DESC
```

---

## Clinical Trials Variant

A specialized extension of the healthcare data model for clinical trial analytics.
Generates realistic synthetic data with correlated lab measurements, adverse events,
and participant outcomes across study visits.

**Important**: All data generated is 100% synthetic and contains no real patient information.

### Clinical Trials Data Model

```
┌──────────────────┐
│ Clinical Trials  │
└────────┬─────────┘
         │ trial_id
         ▼
┌──────────────────┐
│   Study Sites    │
└────────┬─────────┘
         │ site_id
         ▼
┌──────────────────────┐
│  Study Participants  │
└────┬────────────┬────┘
     │            │
     │ participant_id
     ▼            ▼
┌────────────┐  ┌──────────────────┐
│  Adverse   │  │ Lab Measurements │
│  Events    │  │                  │
└────────────┘  └──────────────────┘
```

### Clinical Trial Tables

#### Clinical Trials

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `trial_id` | INT | Primary key | Unique 10000-10099 |
| `nct_number` | STRING | ClinicalTrials.gov ID | Template: `NCT########` |
| `sponsor_company` | STRING | Pharma sponsor | Mimesis business.company |
| `phase` | STRING | Trial phase | Values: Phase I-IV |
| `status` | STRING | Trial status | Values: Active/Completed/Suspended/Terminated |
| `therapeutic_area` | STRING | Disease area | 6 areas (Oncology, Cardiology, etc.) |
| `study_drug` | STRING | Study drug | 20 real drug names |
| `trial_title` | STRING | Descriptive title | concat(study_drug, therapeutic_area) |
| `target_enrollment` | INT | Enrollment target | Phase-dependent CASE |

```python
import dbldatagen as dg
from utils.mimesis_text import mimesisText

trials = (
    dg.DataGenerator(spark, name="clinical_trials", rows=100, partitions=1, randomSeed=42)
    .withColumn("trial_id", "integer", minValue=10000, maxValue=10099, uniqueValues=100)
    .withColumn("nct_number", "string", template="NCT########")
    .withColumn("sponsor_company", "string", text=mimesisText("business.company"))
    .withColumn("phase", "string",
                values=["Phase I", "Phase II", "Phase III", "Phase IV"], random=True)
    .withColumn("status", "string",
                values=["Active", "Completed", "Suspended", "Terminated"], random=True)
    .withColumn("therapeutic_area", "string",
                values=["Oncology", "Cardiology", "Neurology",
                        "Immunology", "Endocrinology", "Rheumatology"], random=True)
    .withColumn("study_drug", "string",
                values=["Pembrolizumab", "Nivolumab", "Atezolizumab", "Durvalumab",
                        "Atorvastatin", "Rosuvastatin", "Evolocumab", "Alirocumab",
                        "Lecanemab", "Aducanumab", "Donepezil", "Memantine",
                        "Adalimumab", "Etanercept", "Infliximab", "Tocilizumab",
                        "Semaglutide", "Tirzepatide", "Empagliflozin", "Dapagliflozin"],
                random=True)
    .withColumn("trial_title", "string",
                baseColumn=["study_drug", "therapeutic_area"],
                expr="concat('Study of ', study_drug, ' in ', therapeutic_area)")
    .withColumn("target_enrollment", "integer",
                baseColumn="phase",
                expr="""
                CASE
                    WHEN phase = 'Phase I' THEN cast(20 + rand() * 60 as int)
                    WHEN phase = 'Phase II' THEN cast(100 + rand() * 200 as int)
                    WHEN phase = 'Phase III' THEN cast(500 + rand() * 1500 as int)
                    ELSE cast(200 + rand() * 800 as int)
                END
                """)
    .build()
)
```

#### Study Sites

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `site_id` | INT | Primary key | Unique 20000-20299 |
| `trial_id` | INT | FK to trials | 10000-10099 |
| `site_name` | STRING | Site name | Derived: mimesis city + " Medical Center" |
| `principal_investigator` | STRING | Lead PI | Derived: "Dr. " + mimesis name |
| `phone` | STRING | Contact phone | Template |
| `site_status` | STRING | Enrollment status | Weighted 5/3/2: Active/Enrolling/Closed |

```python
sites = (
    dg.DataGenerator(spark, name="study_sites", rows=300, partitions=4, randomSeed=42)
    .withColumn("site_id", "integer", minValue=20000, maxValue=20299, uniqueValues=300)
    .withColumn("trial_id", "integer", minValue=10000, maxValue=10099, random=True)
    .withColumn("city_base", "string", text=mimesisText("address.city"), omit=True)
    .withColumn("site_name", "string",
                baseColumn="city_base", expr="concat(city_base, ' Medical Center')")
    .withColumn("pi_name_base", "string", text=mimesisText("person.full_name"), omit=True)
    .withColumn("principal_investigator", "string",
                baseColumn="pi_name_base", expr="concat('Dr. ', pi_name_base)")
    .withColumn("phone", "string", template=r"(\ddd) \ddd-\dddd")
    .withColumn("site_status", "string",
                values=["Active", "Enrolling", "Closed"], weights=[5, 3, 2])
    .build()
)
```

#### Study Participants

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `participant_id` | INT | Primary key | Unique 30000-99999 |
| `site_id` | INT | FK to sites | 20000-20299 |
| `subject_id` | STRING | Blinded ID | Template: `SUBJ-#####` |
| `date_of_birth` | DATE | DOB | Age 25-65 distribution |
| `gender` | STRING | Gender | Values: Male/Female |
| `treatment_arm` | STRING | Randomisation | Weighted 6:4 Active Drug/Placebo |
| `baseline_weight_kg` | DOUBLE | Weight (kg) | 55-125 range |
| `baseline_bmi` | DOUBLE | BMI | 18.5-36.5 range |
| `baseline_disease_severity` | STRING | Severity | Weighted 3/5/2: Mild/Moderate/Severe |
| `prior_treatments` | INT | Prior lines | CASE on severity |
| `enrollment_date` | DATE | Enrollment | Random 2020-2024 |
| `completion_status` | STRING | Outcome | Weighted 6/2/1.5/0.5 |

```python
participants = (
    dg.DataGenerator(spark, name="study_participants", rows=3000, partitions=4, randomSeed=42)
    .withColumn("participant_id", "integer", minValue=30000, maxValue=99999, uniqueValues=3000)
    .withColumn("site_id", "integer", minValue=20000, maxValue=20299, random=True)
    .withColumn("subject_id", "string", template="SUBJ-#####")
    .withColumn("date_of_birth", "date",
                expr="date_add('2020-01-01', -cast(rand()*365*40 + 365*25 as int))")
    .withColumn("gender", "string", values=["Male", "Female"], random=True)
    .withColumn("treatment_arm", "string",
                values=["Active Drug", "Placebo"], weights=[6, 4])
    .withColumn("baseline_weight_kg", "double", expr="55 + rand() * 70")
    .withColumn("baseline_bmi", "double", expr="18.5 + rand() * 18")
    .withColumn("baseline_disease_severity", "string",
                values=["Mild", "Moderate", "Severe"], weights=[3, 5, 2])
    .withColumn("prior_treatments", "integer",
                baseColumn="baseline_disease_severity",
                expr="""
                CASE
                    WHEN baseline_disease_severity = 'Mild' THEN cast(rand() * 2 as int)
                    WHEN baseline_disease_severity = 'Moderate' THEN cast(1 + rand() * 3 as int)
                    ELSE cast(2 + rand() * 4 as int)
                END
                """)
    .withColumn("enrollment_date", "date",
                expr="date_add('2020-01-01', cast(rand()*datediff('2024-12-31', '2020-01-01') as int))")
    .withColumn("completion_status", "string",
                values=["Completed", "Ongoing", "Discontinued", "Lost to Follow-up"],
                weights=[6, 2, 1.5, 0.5])
    .build()
)
```

#### Adverse Events

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `ae_id` | INT | Primary key | Unique 40000-99999 |
| `participant_id` | INT | FK to participants | 30000-32999 |
| `ae_term` | STRING | MedDRA term | Weighted: Nausea/Headache/Fatigue/... |
| `severity` | STRING | Grade | Weighted 6/3/1: Mild/Moderate/Severe |
| `onset_day` | INT | Day of onset | CASE on severity (severe=earlier) |
| `resolution_days` | INT | Days to resolve | CASE on severity (severe=longer) |
| `related_to_study_drug` | BOOLEAN | Drug-related | CASE on severity (severe=70%) |
| `action_taken` | STRING | Clinical action | Nested CASE on severity |
| `ae_description` | STRING | Free text | concat(ae_term, severity, day) |

```python
adverse_events = (
    dg.DataGenerator(spark, name="adverse_events", rows=2000, partitions=4, randomSeed=42)
    .withColumn("ae_id", "integer", minValue=40000, maxValue=99999, uniqueValues=2000)
    .withColumn("participant_id", "integer", minValue=30000, maxValue=32999, random=True)
    .withColumn("ae_term", "string",
                values=["Nausea", "Headache", "Fatigue",
                        "Dizziness", "Injection Site Reaction", "Diarrhea"],
                weights=[3, 2.5, 3, 1.5, 2, 2])
    .withColumn("severity", "string",
                values=["Mild", "Moderate", "Severe"], weights=[6, 3, 1])
    .withColumn("onset_day", "integer",
                baseColumn="severity",
                expr="""
                CASE
                    WHEN severity = 'Severe' THEN cast(1 + rand() * 30 as int)
                    WHEN severity = 'Moderate' THEN cast(1 + rand() * 90 as int)
                    ELSE cast(1 + rand() * 180 as int)
                END
                """)
    .withColumn("resolution_days", "integer",
                baseColumn="severity",
                expr="""
                CASE
                    WHEN severity = 'Severe' THEN cast(7 + rand() * 21 as int)
                    WHEN severity = 'Moderate' THEN cast(3 + rand() * 10 as int)
                    ELSE cast(1 + rand() * 5 as int)
                END
                """)
    .withColumn("related_to_study_drug", "boolean",
                baseColumn="severity",
                expr="CASE WHEN severity = 'Severe' THEN rand() < 0.7 "
                     "WHEN severity = 'Moderate' THEN rand() < 0.5 ELSE rand() < 0.3 END")
    .withColumn("action_taken", "string",
                baseColumn="severity",
                expr="""
                CASE
                    WHEN severity = 'Severe' THEN
                        CASE WHEN rand() < 0.6 THEN 'Dose Reduced'
                             WHEN rand() < 0.8 THEN 'Treatment Interrupted'
                             ELSE 'Treatment Discontinued' END
                    WHEN severity = 'Moderate' THEN
                        CASE WHEN rand() < 0.5 THEN 'Dose Reduced'
                             WHEN rand() < 0.8 THEN 'No Action'
                             ELSE 'Treatment Interrupted' END
                    ELSE 'No Action'
                END
                """)
    .withColumn("report_day", "integer", minValue=1, maxValue=180, random=True, omit=True)
    .withColumn("ae_description", "string",
                baseColumn=["ae_term", "severity", "report_day"],
                expr="concat(ae_term, ': ', lower(severity), '. Reported on day ', "
                     "cast(report_day as string), ' of treatment.')")
    .build()
)
```

#### Lab Measurements

| Column | Type | Description | Generation Pattern |
|--------|------|-------------|-------------------|
| `measurement_id` | INT | Primary key | Unique 50000-99999 |
| `participant_id` | INT | FK to participants | 30000-32999 |
| `visit_number` | INT | Visit (0-6) | Random |
| `visit_name` | STRING | Visit label | CASE on visit_number |
| `visit_date` | DATE | Visit date | Random 2020-2024 |
| `lab_test` | STRING | Test name | 10 tests, weighted |
| `result_value` | DOUBLE | Result | Complex CASE with treatment effects |
| `result_units` | STRING | Unit | CASE on lab_test |
| `reference_min` | DOUBLE | Normal min | CASE on lab_test |
| `reference_max` | DOUBLE | Normal max | CASE on lab_test |
| `abnormal_flag` | BOOLEAN | Out-of-range | Computed from result vs reference |
| `change_from_baseline` | DOUBLE | Abs change | Null for visits 0-1 |
| `percent_change_from_baseline` | DOUBLE | Pct change | Null for visits 0-1 |
| `clinically_significant` | BOOLEAN | >30% out | Compound flag |
| `specimen_type` | STRING | Sample type | Weighted: Whole Blood/Serum/Plasma |
| `fasting_status` | STRING | Fasting | Weighted: Fasting/Non-Fasting/Unknown |
| `sample_quality` | STRING | Quality | Weighted: 90% Acceptable |
| `retest_flag` | BOOLEAN | Needs retest | True if quality != Acceptable |
| `lab_technician` | STRING | Technician | Mimesis person.full_name |
| `reviewed_by_physician` | STRING | Reviewer | Derived: "Dr. " + mimesis name |

```python
lab = (
    dg.DataGenerator(spark, name="lab_measurements", rows=6000, partitions=4, randomSeed=42)
    .withColumn("measurement_id", "integer", minValue=50000, maxValue=99999, uniqueValues=6000)
    .withColumn("participant_id", "integer", minValue=30000, maxValue=32999, random=True)
    .withColumn("visit_number", "integer", minValue=0, maxValue=6, random=True)
    .withColumn("visit_name", "string",
                baseColumn="visit_number",
                expr="""
                CASE
                    WHEN visit_number = 0 THEN 'Screening'
                    WHEN visit_number = 1 THEN 'Baseline'
                    WHEN visit_number = 2 THEN 'Week 4'
                    WHEN visit_number = 3 THEN 'Week 8'
                    WHEN visit_number = 4 THEN 'Week 12'
                    WHEN visit_number = 5 THEN 'Week 24'
                    ELSE 'End of Study'
                END
                """)
    .withColumn("visit_date", "date",
                expr="date_add('2020-01-01', cast(rand()*datediff('2024-12-31', '2020-01-01') as int))")
    .withColumn("lab_test", "string",
                values=["Hemoglobin", "WBC Count", "ALT", "AST", "Creatinine",
                        "BUN", "Glucose", "HbA1c", "LDL", "HDL"],
                weights=[2, 2, 1.5, 1.5, 1.5, 1, 1.5, 1.5, 1, 1])
    .withColumn("result_value", "double",
                baseColumn=["lab_test", "visit_number"],
                expr="""
                CASE
                    WHEN lab_test = 'Hemoglobin' THEN
                        10 + rand() * 8 + (visit_number * 0.2 * CASE WHEN rand() < 0.6 THEN 1 ELSE -0.5 END)
                    WHEN lab_test = 'WBC Count' THEN
                        3 + rand() * 9 + (visit_number * 0.15 * CASE WHEN rand() < 0.6 THEN -1 ELSE 0.5 END)
                    WHEN lab_test = 'ALT' THEN
                        10 + rand() * 80 + (visit_number * 2 * CASE WHEN rand() < 0.7 THEN -1 ELSE 1 END)
                    WHEN lab_test = 'AST' THEN
                        10 + rand() * 80 + (visit_number * 1.8 * CASE WHEN rand() < 0.7 THEN -1 ELSE 1 END)
                    WHEN lab_test = 'Creatinine' THEN
                        0.5 + rand() * 2 + (visit_number * 0.02 * CASE WHEN rand() < 0.5 THEN -1 ELSE 1 END)
                    WHEN lab_test = 'BUN' THEN
                        7 + rand() * 23 + (visit_number * 0.5 * CASE WHEN rand() < 0.5 THEN -1 ELSE 1 END)
                    WHEN lab_test = 'Glucose' THEN
                        70 + rand() * 100 + (visit_number * -2 * CASE WHEN rand() < 0.65 THEN 1 ELSE -0.5 END)
                    WHEN lab_test = 'HbA1c' THEN
                        5.0 + rand() * 5 + (visit_number * -0.15 * CASE WHEN rand() < 0.65 THEN 1 ELSE -0.5 END)
                    WHEN lab_test = 'LDL' THEN
                        80 + rand() * 120 + (visit_number * -3 * CASE WHEN rand() < 0.7 THEN 1 ELSE -0.3 END)
                    WHEN lab_test = 'HDL' THEN
                        35 + rand() * 45 + (visit_number * 1 * CASE WHEN rand() < 0.6 THEN 1 ELSE -0.5 END)
                    ELSE 50 + rand() * 100
                END
                """)
    .withColumn("result_units", "string",
                baseColumn="lab_test",
                expr="""
                CASE
                    WHEN lab_test = 'Hemoglobin' THEN 'g/dL'
                    WHEN lab_test = 'WBC Count' THEN '10E9/L'
                    WHEN lab_test IN ('ALT', 'AST') THEN 'U/L'
                    WHEN lab_test IN ('Creatinine', 'BUN', 'Glucose', 'LDL', 'HDL') THEN 'mg/dL'
                    WHEN lab_test = 'HbA1c' THEN '%'
                    ELSE 'units'
                END
                """)
    .withColumn("reference_min", "double",
                baseColumn="lab_test",
                expr="""
                CASE
                    WHEN lab_test = 'Hemoglobin' THEN 12.0
                    WHEN lab_test = 'WBC Count' THEN 4.0
                    WHEN lab_test = 'ALT' THEN 7.0
                    WHEN lab_test = 'AST' THEN 10.0
                    WHEN lab_test = 'Creatinine' THEN 0.6
                    WHEN lab_test = 'BUN' THEN 7.0
                    WHEN lab_test = 'Glucose' THEN 70.0
                    WHEN lab_test = 'HbA1c' THEN 4.0
                    WHEN lab_test = 'LDL' THEN 0.0
                    WHEN lab_test = 'HDL' THEN 40.0
                    ELSE 10.0
                END
                """)
    .withColumn("reference_max", "double",
                baseColumn="lab_test",
                expr="""
                CASE
                    WHEN lab_test = 'Hemoglobin' THEN 16.0
                    WHEN lab_test = 'WBC Count' THEN 11.0
                    WHEN lab_test = 'ALT' THEN 56.0
                    WHEN lab_test = 'AST' THEN 40.0
                    WHEN lab_test = 'Creatinine' THEN 1.2
                    WHEN lab_test = 'BUN' THEN 20.0
                    WHEN lab_test = 'Glucose' THEN 99.0
                    WHEN lab_test = 'HbA1c' THEN 5.6
                    WHEN lab_test = 'LDL' THEN 100.0
                    WHEN lab_test = 'HDL' THEN 999.0
                    ELSE 100.0
                END
                """)
    .withColumn("abnormal_flag", "boolean",
                baseColumn=["result_value", "reference_min", "reference_max"],
                expr="result_value < reference_min OR result_value > reference_max")
    .withColumn("change_from_baseline", "double",
                baseColumn=["result_value", "visit_number"],
                expr="CASE WHEN visit_number > 1 "
                     "THEN (result_value - result_value * (1 - visit_number * 0.02)) ELSE NULL END")
    .withColumn("percent_change_from_baseline", "double",
                baseColumn=["change_from_baseline", "result_value"],
                expr="CASE WHEN change_from_baseline IS NOT NULL "
                     "THEN (change_from_baseline / result_value) * 100 ELSE NULL END")
    .withColumn("clinically_significant", "boolean",
                baseColumn=["abnormal_flag", "result_value", "reference_min", "reference_max"],
                expr="abnormal_flag AND (result_value < reference_min * 0.7 "
                     "OR result_value > reference_max * 1.3)")
    .withColumn("specimen_type", "string",
                values=["Whole Blood", "Serum", "Plasma"], weights=[3, 5, 2])
    .withColumn("fasting_status", "string",
                values=["Fasting", "Non-Fasting", "Unknown"], weights=[4, 5, 1])
    .withColumn("sample_quality", "string",
                values=["Acceptable", "Hemolyzed", "Lipemic", "Icteric"],
                weights=[9, 0.5, 0.3, 0.2])
    .withColumn("retest_flag", "boolean",
                baseColumn="sample_quality",
                expr="CASE WHEN sample_quality != 'Acceptable' THEN true ELSE rand() < 0.03 END")
    .withColumn("lab_technician", "string", text=mimesisText("person.full_name"))
    .withColumn("physician_name_base", "string",
                text=mimesisText("person.full_name"), omit=True)
    .withColumn("reviewed_by_physician", "string",
                baseColumn="physician_name_base",
                expr="concat('Dr. ', physician_name_base)")
    .build()
)
```

### Clinical Trials Demo Queries

#### Treatment Effect Over Time
```sql
SELECT
    lab_test,
    visit_name,
    COUNT(*) as measurement_count,
    ROUND(AVG(result_value), 2) as avg_result,
    ROUND(AVG(change_from_baseline), 2) as avg_change,
    ROUND(AVG(percent_change_from_baseline), 2) as avg_pct_change
FROM lab_measurements
WHERE visit_number > 1
GROUP BY lab_test, visit_name, visit_number
ORDER BY lab_test, visit_number
```

#### Adverse Events by Severity
```sql
SELECT
    severity,
    action_taken,
    COUNT(*) as event_count,
    ROUND(AVG(onset_day), 1) as avg_onset_day,
    ROUND(AVG(resolution_days), 1) as avg_resolution_days,
    ROUND(AVG(CASE WHEN related_to_study_drug THEN 1.0 ELSE 0.0 END) * 100, 1) as pct_drug_related
FROM adverse_events
GROUP BY severity, action_taken
ORDER BY severity, event_count DESC
```

#### Participant Baseline Characteristics
```sql
SELECT
    treatment_arm,
    baseline_disease_severity,
    COUNT(*) as participant_count,
    ROUND(AVG(baseline_weight_kg), 1) as avg_weight,
    ROUND(AVG(baseline_bmi), 1) as avg_bmi,
    ROUND(AVG(prior_treatments), 1) as avg_prior_tx
FROM study_participants
GROUP BY treatment_arm, baseline_disease_severity
ORDER BY treatment_arm, baseline_disease_severity
```

### Complete Clinical Trials Demo

```python
def generate_clinical_trials_demo(
    spark,
    catalog: str,
    schema: str = "clinical_trials",
    volume: str = "raw_data",
    base_rows: int = 1_000,
    seed: int = 42,
):
    """Generate complete clinical-trials demo dataset.

    Table sizes scale from base_rows:
        - clinical_trials:     100          (fixed)
        - study_sites:         300          (fixed)
        - study_participants:  base_rows * 3
        - adverse_events:      base_rows * 2
        - lab_measurements:    base_rows * 6
    """
    n_participants = base_rows * 3

    trials = generate_clinical_trials(spark, rows=100, seed=seed)
    sites = generate_study_sites(spark, rows=300, seed=seed)
    participants = generate_study_participants(spark, rows=n_participants, seed=seed)
    adverse_events = generate_adverse_events(
        spark, rows=base_rows * 2, n_participants=n_participants, seed=seed)
    lab_measurements = generate_lab_measurements(
        spark, rows=base_rows * 6, n_participants=n_participants, seed=seed)

    tables = {
        "clinical_trials": trials,
        "study_sites": sites,
        "study_participants": participants,
        "adverse_events": adverse_events,
        "lab_measurements": lab_measurements,
    }

    for name, df in tables.items():
        df.write.format("delta").mode("overwrite").saveAsTable(f"{catalog}.{schema}.{name}")

    return tables
```
