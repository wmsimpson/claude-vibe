"""Oil & gas industry synthetic data generators.

REFERENCE IMPLEMENTATION — This file is not an importable module. Claude reads it
for patterns and adapts the code inline for user notebooks and scripts.

Generates upstream oil & gas production data for the Texas Permian Basin using
ARPS decline curve methodology. Uses a per-formation generation + union pattern:
well headers and type curves are generated per formation then unioned, and daily
production is generated per well using the well's ARPS parameters.
"""

# CONNECT COMPATIBILITY NOTES:
#   Catalyst-safe (works over Connect):
#     - values=/weights=, minValue/maxValue, begin/end dates, expr=, percentNulls=, omit=
#   UDF-dependent (notebook only — apply workarounds for Connect):
#     - text=mimesisText() → values=["James","Mary",...], random=True
#     - template=r"ddddd" → expr="lpad(cast(floor(rand()*100000) as string), 5, '0')"
#     - distribution=Gamma/Beta → random=True or expr= math
#     - .withConstraint() → .build().filter("condition")

import random
from datetime import date, timedelta

import dbldatagen as dg
from dbldatagen.config import OutputDataset
from pyspark.sql import DataFrame

# ARPS decline curve parameters per formation
TYPE_CURVE_PARAMS = {
    "FORMATION_A": {"q_i": 6000, "d": 0.01, "b": 0.8},
    "FORMATION_B": {"q_i": 7000, "d": 0.011, "b": 0.7},
    "FORMATION_C": {"q_i": 5500, "d": 0.009, "b": 0.8},
    "FORMATION_D": {"q_i": 5750, "d": 0.011, "b": 0.7},
}


def generate_well_headers(
    spark,
    rows_per_formation=50,
    seed=42,
    output: OutputDataset | None = None,
) -> DataFrame | None:
    """Generate well header data for all formations.

    Creates one DataGenerator per formation with formation-specific ARPS
    parameters, then unions all resulting DataFrames.

    Args:
        spark: Active SparkSession.
        rows_per_formation: Number of wells per formation (default 50).
        seed: Base random seed for reproducibility.
        output: Optional OutputDataset for saveAsDataset mode.

    Returns:
        Unioned DataFrame of all well headers, or None if output is provided.
    """
    partitions = max(2, rows_per_formation // 50)
    dfs = []

    for i, (formation, params) in enumerate(TYPE_CURVE_PARAMS.items()):
        spec = (
            dg.DataGenerator(
                spark,
                rows=rows_per_formation,
                partitions=partitions,
                randomSeed=seed + i,
                name=formation,
            )
            .withColumn(
                "API_NUMBER", "bigint",
                minValue=42000000000000, maxValue=42999999999999, random=True,
            )
            .withColumn(
                "FIELD_NAME", "string",
                values=["Field_1", "Field_2", "Field_3", "Field_4", "Field_5"],
                random=True,
            )
            .withColumn(
                "LATITUDE", "float",
                minValue=31.00, maxValue=32.50, step=1e-6, random=True,
            )
            .withColumn(
                "LONGITUDE", "float",
                minValue=-104.00, maxValue=-101.00, step=1e-6, random=True,
            )
            .withColumn(
                "COUNTY", "string",
                values=["Reeves", "Midland", "Ector", "Loving", "Ward"],
                random=True,
            )
            .withColumn("STATE", "string", values=["Texas"])
            .withColumn("COUNTRY", "string", values=["USA"])
            .withColumn("WELL_TYPE", "string", values=["Oil"])
            .withColumn("WELL_ORIENTATION", "string", values=["Horizontal"])
            .withColumn("PRODUCING_FORMATION", "string", values=[formation])
            .withColumn(
                "CURRENT_STATUS", "string",
                values=["Producing", "Shut-in", "Plugged and Abandoned", "Planned"],
                weights=[80, 10, 5, 5],
                random=True,
            )
            .withColumn(
                "TOTAL_DEPTH", "integer",
                minValue=12000, maxValue=20000, random=True,
            )
            .withColumn(
                "SPUD_DATE", "date",
                begin="2020-01-01", end="2025-02-14", random=True,
            )
            .withColumn(
                "COMPLETION_DATE", "date",
                begin="2020-01-01", end="2025-02-14", random=True,
            )
            .withColumn(
                "SURFACE_CASING_DEPTH", "integer",
                minValue=500, maxValue=800, random=True,
            )
            .withColumn("OPERATOR_NAME", "string", values=["OPERATOR_XYZ"])
            .withColumn(
                "PERMIT_DATE", "date",
                begin="2019-01-01", end="2025-02-14", random=True,
            )
            .withColumn("q_i", "double", values=[params["q_i"]])
            .withColumn("d", "double", values=[params["d"]])
            .withColumn("b", "double", values=[params["b"]])
        )

        if output:
            spec.saveAsDataset(dataset=output)
        else:
            dfs.append(spec.build())

    if output:
        return None

    result = dfs[0]
    for df in dfs[1:]:
        result = result.unionByName(df)
    return result


def generate_daily_production(
    spark,
    wells_df: DataFrame,
    seed=42,
    output: OutputDataset | None = None,
) -> DataFrame | None:
    """Generate daily production data for all wells.

    Iterates over wells from wells_df, generating 100-700 days of ARPS-based
    production data per well with shut-in simulation and production noise.

    Args:
        spark: Active SparkSession.
        wells_df: Well headers DataFrame (output of generate_well_headers).
        seed: Base random seed for reproducibility.
        output: Optional OutputDataset for saveAsDataset mode.

    Returns:
        Unioned DataFrame of all daily production, or None if output is provided.
    """
    wells_dict = wells_df.toPandas().to_dict()
    n_wells = len(next(iter(wells_dict.values())))
    dfs = []

    for i in range(n_wells):
        well_num = wells_dict["API_NUMBER"][i]
        q_i = wells_dict["q_i"][i]
        d_val = wells_dict["d"][i]
        b_val = wells_dict["b"][i]

        days_to_generate = int(round(random.uniform(100, 700)))
        well_seed = seed + i

        spec = (
            dg.DataGenerator(
                spark,
                rows=days_to_generate,
                partitions=max(2, days_to_generate // 200),
                randomSeed=well_seed,
                name="daily_production",
            )
            .withColumn("well_num", "bigint", values=[well_num])
            .withColumn(
                "day_from_first_production", "integer",
                minValue=1, maxValue=1000,
            )
            .withColumn(
                "first_production_date", "date",
                values=[date.today() - timedelta(days=days_to_generate)],
            )
            .withColumn(
                "date", "date",
                expr="date_add(first_production_date, day_from_first_production)",
            )
            .withColumn("q_i", "double", values=[q_i])
            .withColumn("d", "double", values=[d_val])
            .withColumn("b", "double", values=[b_val])
            .withColumn(
                "q_i_multiplier", "double",
                values=[1.0, 0],
                weights=[97, 3],
                random=True,
            )
            .withColumn("variation", "double", expr="rand() * 0.1 + 0.95")
            .withColumn(
                "actuals_bopd", "double",
                baseColumn=["q_i", "d", "b", "q_i_multiplier", "variation"],
                expr="(q_i * q_i_multiplier) / power(1 + b * d * variation * day_from_first_production, 1/b)",
            )
        )

        if output:
            spec.saveAsDataset(dataset=output)
        else:
            dfs.append(spec.build())

    if output:
        return None

    result = dfs[0]
    for df in dfs[1:]:
        result = result.unionByName(df)
    return result


def generate_type_curves(
    spark,
    rows_per_formation=2000,
    seed=42,
    output: OutputDataset | None = None,
) -> DataFrame | None:
    """Generate type curve forecast data for all formations.

    Creates one DataGenerator per formation with formation-specific ARPS
    parameters for production forecasting, then unions all DataFrames.

    Args:
        spark: Active SparkSession.
        rows_per_formation: Number of forecast days per formation (default 2000).
        seed: Base random seed for reproducibility.
        output: Optional OutputDataset for saveAsDataset mode.

    Returns:
        Unioned DataFrame of all type curves, or None if output is provided.
    """
    partitions = max(2, rows_per_formation // 500)
    dfs = []

    for i, (formation, params) in enumerate(TYPE_CURVE_PARAMS.items()):
        spec = (
            dg.DataGenerator(
                spark,
                rows=rows_per_formation,
                partitions=partitions,
                randomSeed=seed + i,
                name="type_curve",
            )
            .withColumn("formation", "string", values=[formation])
            .withColumn(
                "day_from_first_production", "integer",
                minValue=1, maxValue=1000,
            )
            .withColumn("q_i", "double", values=[params["q_i"]])
            .withColumn("d", "double", values=[params["d"]])
            .withColumn("b", "double", values=[params["b"]])
            .withColumn("variation", "double", expr="rand() * 0.1 + 0.95")
            .withColumn(
                "forecast_bopd", "double",
                baseColumn=["q_i", "d", "b", "day_from_first_production"],
                expr="q_i / power(1 + b * d * day_from_first_production, 1/b)",
            )
        )

        if output:
            spec.saveAsDataset(dataset=output)
        else:
            dfs.append(spec.build())

    if output:
        return None

    result = dfs[0]
    for df in dfs[1:]:
        result = result.unionByName(df)
    return result


def generate_oil_gas_demo(
    spark,
    catalog,
    schema="oil_gas",
    rows_per_formation=50,
    rows_per_type_curve=2000,
    seed=42,
):
    """Generate complete oil & gas demo dataset with all tables.

    Orchestrates the full generation pipeline:
    1. Generate well headers for all formations (per-formation + union)
    2. Generate daily production per well using ARPS formula
    3. Generate type curve forecasts per formation
    4. Write all tables to Unity Catalog as Delta

    Args:
        spark: Active SparkSession.
        catalog: Unity Catalog catalog name.
        schema: Schema name (default "oil_gas").
        rows_per_formation: Wells per formation (default 50).
        rows_per_type_curve: Forecast days per formation (default 2000).
        seed: Base random seed for reproducibility.
    """
    # 1. Generate well headers for all formations
    well_headers = generate_well_headers(
        spark, rows_per_formation=rows_per_formation, seed=seed,
    )

    # 2. Generate daily production per well (uses wells_df for iteration)
    daily_production = generate_daily_production(
        spark, wells_df=well_headers, seed=seed,
    )

    # 3. Generate type curves per formation
    type_curves = generate_type_curves(
        spark, rows_per_formation=rows_per_type_curve, seed=seed,
    )

    # 4. Write all tables to Unity Catalog
    tables = {
        "well_headers": well_headers,
        "daily_production": daily_production,
        "type_curves": type_curves,
    }

    for name, df in tables.items():
        df.write.format("delta").mode("overwrite").saveAsTable(
            f"{catalog}.{schema}.{name}"
        )

    return tables
