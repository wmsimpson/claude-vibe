"""Clinical trials synthetic data generators.

REFERENCE IMPLEMENTATION — This file is not an importable module. Claude reads it
for patterns and adapts the code inline for user notebooks and scripts.
"""

# CONNECT COMPATIBILITY NOTES:
#   Catalyst-safe (works over Connect):
#     - values=/weights=, minValue/maxValue, begin/end dates, expr=, percentNulls=, omit=
#   UDF-dependent (notebook only — apply workarounds for Connect):
#     - text=mimesisText() → values=["James","Mary",...], random=True

import dbldatagen as dg
from dbldatagen.config import OutputDataset
from pyspark.sql import DataFrame
from utils.mimesis_text import mimesisText


def generate_clinical_trials(spark, rows=100, seed=42, output: OutputDataset | None = None) -> DataFrame | None:
    """Generate synthetic clinical trial protocol data.

    Each row represents a clinical trial with phase, therapeutic area,
    study drug, and phase-dependent target enrollment.
    """
    partitions = max(1, rows // 100)
    spec = (
        dg.DataGenerator(spark, name="clinical_trials", rows=rows,
                         partitions=partitions, randomSeed=seed)
        .withColumn("trial_id", "integer",
                    minValue=10000, maxValue=10000 + rows - 1, uniqueValues=rows)
        .withColumn("nct_number", "string", template="NCT########")
        .withColumn("sponsor_company", "string",
                    text=mimesisText("business.company"))
        .withColumn("phase", "string",
                    values=["Phase I", "Phase II", "Phase III", "Phase IV"],
                    random=True)
        .withColumn("status", "string",
                    values=["Active", "Completed", "Suspended", "Terminated"],
                    random=True)
        .withColumn("therapeutic_area", "string",
                    values=["Oncology", "Cardiology", "Neurology",
                            "Immunology", "Endocrinology", "Rheumatology"],
                    random=True)
        .withColumn("study_drug", "string",
                    values=[
                        "Pembrolizumab", "Nivolumab", "Atezolizumab", "Durvalumab",
                        "Atorvastatin", "Rosuvastatin", "Evolocumab", "Alirocumab",
                        "Lecanemab", "Aducanumab", "Donepezil", "Memantine",
                        "Adalimumab", "Etanercept", "Infliximab", "Tocilizumab",
                        "Semaglutide", "Tirzepatide", "Empagliflozin", "Dapagliflozin",
                    ],
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
    )
    if output:
        spec.saveAsDataset(dataset=output)
        return None
    return spec.build()


def generate_study_sites(spark, rows=300, seed=42, output: OutputDataset | None = None) -> DataFrame | None:
    """Generate synthetic study-site data linked to clinical trials.

    Site names are derived from mimesis city names + ' Medical Center'.
    Principal investigators use 'Dr. ' prefix with mimesis names.
    """
    partitions = max(1, rows // 100)
    spec = (
        dg.DataGenerator(spark, name="study_sites", rows=rows,
                         partitions=partitions, randomSeed=seed)
        .withColumn("site_id", "integer",
                    minValue=20000, maxValue=20000 + rows - 1, uniqueValues=rows)
        .withColumn("trial_id", "integer",
                    minValue=10000, maxValue=10099, random=True)
        .withColumn("city_base", "string",
                    text=mimesisText("address.city"), omit=True)
        .withColumn("site_name", "string",
                    baseColumn="city_base",
                    expr="concat(city_base, ' Medical Center')")
        .withColumn("pi_name_base", "string",
                    text=mimesisText("person.full_name"), omit=True)
        .withColumn("principal_investigator", "string",
                    baseColumn="pi_name_base",
                    expr="concat('Dr. ', pi_name_base)")
        .withColumn("phone", "string",
                    template=r"(\ddd) \ddd-\dddd")
        .withColumn("site_status", "string",
                    values=["Active", "Enrolling", "Closed"],
                    weights=[5, 3, 2])
    )
    if output:
        spec.saveAsDataset(dataset=output)
        return None
    return spec.build()


def generate_study_participants(spark, rows=3_000, seed=42, output: OutputDataset | None = None) -> DataFrame | None:
    """Generate synthetic study-participant data with baseline characteristics.

    Includes treatment arm randomisation, disease severity, and
    severity-dependent prior-treatment counts.
    """
    partitions = max(4, rows // 1_000)
    start_date = "2020-01-01"
    end_date = "2024-12-31"
    spec = (
        dg.DataGenerator(spark, name="study_participants", rows=rows,
                         partitions=partitions, randomSeed=seed)
        .withColumn("participant_id", "integer",
                    minValue=30000, maxValue=99999, uniqueValues=rows)
        .withColumn("site_id", "integer",
                    minValue=20000, maxValue=20299, random=True)
        .withColumn("subject_id", "string", template="SUBJ-#####")
        .withColumn("date_of_birth", "date",
                    expr=f"date_add('{start_date}', -cast(rand()*365*40 + 365*25 as int))")
        .withColumn("gender", "string",
                    values=["Male", "Female"], random=True)
        .withColumn("treatment_arm", "string",
                    values=["Active Drug", "Placebo"],
                    weights=[6, 4])
        .withColumn("baseline_weight_kg", "double",
                    expr="55 + rand() * 70")
        .withColumn("baseline_bmi", "double",
                    expr="18.5 + rand() * 18")
        .withColumn("baseline_disease_severity", "string",
                    values=["Mild", "Moderate", "Severe"],
                    weights=[3, 5, 2])
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
                    expr=f"date_add('{start_date}', cast(rand()*datediff('{end_date}', '{start_date}') as int))")
        .withColumn("completion_status", "string",
                    values=["Completed", "Ongoing", "Discontinued", "Lost to Follow-up"],
                    weights=[6, 2, 1.5, 0.5])
    )
    if output:
        spec.saveAsDataset(dataset=output)
        return None
    return spec.build()


def generate_adverse_events(spark, rows=2_000, n_participants=3_000, seed=42,
                            output: OutputDataset | None = None) -> DataFrame | None:
    """Generate synthetic adverse-event data correlated with severity.

    Severe events have earlier onset, longer resolution, higher drug-relatedness,
    and escalated clinical actions.
    """
    partitions = max(4, rows // 1_000)
    spec = (
        dg.DataGenerator(spark, name="adverse_events", rows=rows,
                         partitions=partitions, randomSeed=seed)
        .withColumn("ae_id", "integer",
                    minValue=40000, maxValue=99999, uniqueValues=rows)
        .withColumn("participant_id", "integer",
                    minValue=30000, maxValue=30000 + n_participants - 1,
                    random=True)
        .withColumn("ae_term", "string",
                    values=["Nausea", "Headache", "Fatigue",
                            "Dizziness", "Injection Site Reaction", "Diarrhea"],
                    weights=[3, 2.5, 3, 1.5, 2, 2])
        .withColumn("severity", "string",
                    values=["Mild", "Moderate", "Severe"],
                    weights=[6, 3, 1])
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
                         "WHEN severity = 'Moderate' THEN rand() < 0.5 "
                         "ELSE rand() < 0.3 END")
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
        .withColumn("report_day", "integer",
                    minValue=1, maxValue=180, random=True, omit=True)
        .withColumn("ae_description", "string",
                    baseColumn=["ae_term", "severity", "report_day"],
                    expr="concat(ae_term, ': ', lower(severity), "
                         "'. Reported on day ', cast(report_day as string), ' of treatment.')")
    )
    if output:
        spec.saveAsDataset(dataset=output)
        return None
    return spec.build()


def generate_lab_measurements(spark, rows=6_000, n_participants=3_000, seed=42,
                              output: OutputDataset | None = None) -> DataFrame | None:
    """Generate synthetic lab-measurement data with treatment effects.

    Result values vary by lab test and visit number, simulating the
    progressive treatment effect seen in real clinical trials. Includes
    reference ranges, abnormality flags, and sample-quality metadata.
    """
    partitions = max(4, rows // 1_000)
    start_date = "2020-01-01"
    end_date = "2024-12-31"
    spec = (
        dg.DataGenerator(spark, name="lab_measurements", rows=rows,
                         partitions=partitions, randomSeed=seed)
        .withColumn("measurement_id", "integer",
                    minValue=50000, maxValue=99999, uniqueValues=rows)
        .withColumn("participant_id", "integer",
                    minValue=30000, maxValue=30000 + n_participants - 1,
                    random=True)
        .withColumn("visit_number", "integer",
                    minValue=0, maxValue=6, random=True)
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
                    expr=f"date_add('{start_date}', "
                         f"cast(rand()*datediff('{end_date}', '{start_date}') as int))")
        .withColumn("lab_test", "string",
                    values=["Hemoglobin", "WBC Count", "ALT", "AST", "Creatinine",
                            "BUN", "Glucose", "HbA1c", "LDL", "HDL"],
                    weights=[2, 2, 1.5, 1.5, 1.5, 1, 1.5, 1.5, 1, 1])
        # Result values with treatment effect — active drug shows improvement over visits
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
                         "THEN (result_value - result_value * (1 - visit_number * 0.02)) "
                         "ELSE NULL END")
        .withColumn("percent_change_from_baseline", "double",
                    baseColumn=["change_from_baseline", "result_value"],
                    expr="CASE WHEN change_from_baseline IS NOT NULL "
                         "THEN (change_from_baseline / result_value) * 100 "
                         "ELSE NULL END")
        .withColumn("clinically_significant", "boolean",
                    baseColumn=["abnormal_flag", "result_value",
                                "reference_min", "reference_max"],
                    expr="abnormal_flag AND "
                         "(result_value < reference_min * 0.7 OR result_value > reference_max * 1.3)")
        .withColumn("specimen_type", "string",
                    values=["Whole Blood", "Serum", "Plasma"],
                    weights=[3, 5, 2])
        .withColumn("fasting_status", "string",
                    values=["Fasting", "Non-Fasting", "Unknown"],
                    weights=[4, 5, 1])
        .withColumn("sample_quality", "string",
                    values=["Acceptable", "Hemolyzed", "Lipemic", "Icteric"],
                    weights=[9, 0.5, 0.3, 0.2])
        .withColumn("retest_flag", "boolean",
                    baseColumn="sample_quality",
                    expr="CASE WHEN sample_quality != 'Acceptable' THEN true "
                         "ELSE rand() < 0.03 END")
        .withColumn("lab_technician", "string",
                    text=mimesisText("person.full_name"))
        .withColumn("physician_name_base", "string",
                    text=mimesisText("person.full_name"), omit=True)
        .withColumn("reviewed_by_physician", "string",
                    baseColumn="physician_name_base",
                    expr="concat('Dr. ', physician_name_base)")
    )
    if output:
        spec.saveAsDataset(dataset=output)
        return None
    return spec.build()


def generate_clinical_trials_demo(spark, catalog, schema="clinical_trials",
                                  volume="raw_data", base_rows=1_000, seed=42):
    """Generate complete clinical-trials demo dataset with all five tables.

    Table sizes scale from *base_rows*:
        - clinical_trials:     100          (fixed — trial protocols)
        - study_sites:         300          (fixed — participating sites)
        - study_participants:  base_rows * 3
        - adverse_events:      base_rows * 2
        - lab_measurements:    base_rows * 6
    """
    from ..utils.output import write_medallion

    n_participants = base_rows * 3

    trials = generate_clinical_trials(spark, rows=100, seed=seed)
    sites = generate_study_sites(spark, rows=300, seed=seed)
    participants = generate_study_participants(spark, rows=n_participants, seed=seed)
    adverse_events = generate_adverse_events(
        spark, rows=base_rows * 2, n_participants=n_participants, seed=seed)
    lab_measurements = generate_lab_measurements(
        spark, rows=base_rows * 6, n_participants=n_participants, seed=seed)

    write_medallion(
        tables={
            "clinical_trials": trials,
            "study_sites": sites,
            "study_participants": participants,
            "adverse_events": adverse_events,
            "lab_measurements": lab_measurements,
        },
        catalog=catalog,
        schema=schema,
        volume=volume,
    )
