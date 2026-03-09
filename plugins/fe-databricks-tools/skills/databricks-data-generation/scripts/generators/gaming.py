"""Gaming industry synthetic data generators.

REFERENCE IMPLEMENTATION — This file is not an importable module. Claude reads it
for patterns and adapts the code inline for user notebooks and scripts.
"""

# CONNECT COMPATIBILITY NOTES:
#   This generator uses baseColumnType="hash" which is NOTEBOOK-ONLY (Tier 3).
#   It does NOT work over Databricks Connect.

import dbldatagen as dg
from dbldatagen.config import OutputDataset
from pyspark.sql import DataFrame
from pyspark.sql.types import DoubleType, StringType, TimestampType, LongType
from pyspark.sql.functions import col, hour, to_date


def generate_login_events(
    spark,
    rows=4_500_000,
    n_users=200_000,
    n_devices=250_000,
    n_ips=40_000,
    start_timestamp="2025-03-01 00:00:00",
    end_timestamp="2025-03-30 00:00:00",
    seed=42,
    output: OutputDataset | None = None,
) -> DataFrame | None:
    """Generate gaming login event telemetry with device fingerprinting and GeoIP.

    Produces a single wide table of login events enriched with deterministic
    account/device IDs (hex-formatted), SHA-256 session/client IDs, country-weighted
    GeoIP fields (city, ISP, coordinates), and derived time columns.

    Uses ``baseColumnType="hash"`` for deterministic ID generation — this is
    **notebook-only** (Tier 3) and does NOT work over Databricks Connect.

    Args:
        spark: SparkSession.
        rows: Total login events to generate.
        n_users: Number of unique accounts (ACCOUNTID cardinality).
        n_devices: Number of unique devices (DEVICEID cardinality).
        n_ips: Number of unique IP addresses.
        start_timestamp: Start of the event timestamp range.
        end_timestamp: End of the event timestamp range.
        seed: Random seed for reproducibility.
        output: Optional OutputDataset for direct write (skips .build()).

    Returns:
        DataFrame with login events, or None if output is provided.
    """
    partitions = max(50, rows // 1_000_000)

    spec = (
        dg.DataGenerator(
            spark, name="login_events", rows=rows,
            partitions=partitions, randomSeed=seed,
        )
        .withColumn(
            "EVENT_TIMESTAMP",
            TimestampType(),
            data_range=dg.DateRange(start_timestamp, end_timestamp, "seconds=1"),
            random=True,
        )  # Random event timestamp within the specified range
        .withColumn(
            "internal_ACCOUNTID",
            LongType(),
            minValue=0x1000000000000,
            uniqueValues=n_users,
            omit=True,
            baseColumnType="hash",
        )  # Internal unique account id, omitted from output, used for deterministic hashing
        .withColumn(
            "ACCOUNTID", StringType(), format="0x%032x", baseColumn="internal_ACCOUNTID",
        )  # Public account id as hex string
        .withColumn(
            "internal_DEVICEID",
            LongType(),
            minValue=0x1000000000000,
            uniqueValues=n_devices,
            omit=True,
            baseColumnType="hash",
            baseColumn="internal_ACCOUNTID",
        )  # Internal device id, based on account, omitted from output
        .withColumn(
            "DEVICEID", StringType(), format="0x%032x", baseColumn="internal_DEVICEID",
        )  # Public device id as hex string
        .withColumn(
            "APP_VERSION", StringType(), values=["current"],
        )  # Static app version
        .withColumn(
            "AUTHMETHOD", StringType(), values=["OAuth", "password"],
        )  # Auth method, random selection
        # Assign clientName based on DEVICEID deterministically
        .withColumn(
            "CLIENTNAME",
            StringType(),
            expr="""
                element_at(
                    array('SwitchGameClient','XboxGameClient','PlaystationGameClient','PCGameClient'),
                    (pmod(abs(hash(DEVICEID)), 4) + 1)
                )
            """,
        )
        .withColumn(
            "CLIENTID",
            StringType(),
            expr="sha2(concat(ACCOUNTID, CLIENTNAME), 256)",
            baseColumn=["ACCOUNTID", "CLIENTNAME"],
        )  # Deterministic clientId based on ACCOUNTID and clientName
        .withColumn(
            "SESSION_ID",
            StringType(),
            expr="sha2(concat(ACCOUNTID, CLIENTID), 256)",
        )  # Session correlation id, deterministic hash
        .withColumn(
            "country",
            StringType(),
            values=["USA", "UK", "AUS"],
            weights=[0.6, 0.2, 0.2],
            baseColumn="ACCOUNTID",
            random=True,
        )  # Assign country with 60% USA, 20% UK, 20% AUS
        .withColumn(
            "APPENV", StringType(), values=["prod"],
        )  # Static environment value
        .withColumn(
            "EVENT_TYPE", StringType(), values=["account_login_success"],
        )  # Static event type
        # Assign geoip_city_name based on country and ACCOUNTID
        .withColumn(
            "CITY",
            StringType(),
            expr="""
                CASE
                    WHEN country = 'USA' THEN element_at(array('New York', 'San Francisco', 'Chicago'), pmod(abs(hash(ACCOUNTID)), 3) + 1)
                    WHEN country = 'UK' THEN 'London'
                    WHEN country = 'AUS' THEN 'Sydney'
                END
            """,
            baseColumn=["country", "ACCOUNTID"],
        )
        .withColumn(
            "COUNTRY_CODE2",
            StringType(),
            expr="CASE WHEN country = 'USA' THEN 'US' WHEN country = 'UK' THEN 'UK' WHEN country = 'AUS' THEN 'AU' END",
            baseColumn=["country"],
        )  # Country code
        # Assign ISP based on country and ACCOUNTID
        .withColumn(
            "ISP",
            StringType(),
            expr="""
                CASE
                    WHEN country = 'USA' THEN element_at(array('Comcast', 'AT&T', 'Verizon', 'Spectrum', 'Cox'), pmod(abs(hash(ACCOUNTID)), 5) + 1)
                    WHEN country = 'UK' THEN element_at(array('BT', 'Sky', 'Virgin Media', 'TalkTalk', 'EE'), pmod(abs(hash(ACCOUNTID)), 5) + 1)
                    WHEN country = 'AUS' THEN element_at(array('Telstra', 'Optus', 'TPG', 'Aussie Broadband', 'iiNet'), pmod(abs(hash(ACCOUNTID)), 5) + 1)
                    ELSE 'Unknown ISP'
                END
            """,
            baseColumn=["country", "ACCOUNTID"],
        )
        # Assign latitude based on city
        .withColumn(
            "LATITUDE",
            DoubleType(),
            expr="""
                CASE
                    WHEN CITY = 'New York' THEN 40.7128
                    WHEN CITY = 'San Francisco' THEN 37.7749
                    WHEN CITY = 'Chicago' THEN 41.8781
                    WHEN CITY = 'London' THEN 51.5074
                    WHEN CITY = 'Sydney' THEN -33.8688
                    ELSE 0.0
                END
            """,
            baseColumn="CITY",
        )
        # Assign longitude based on city
        .withColumn(
            "LONGITUDE",
            DoubleType(),
            expr="""
                CASE
                    WHEN CITY = 'New York' THEN -74.0060
                    WHEN CITY = 'San Francisco' THEN -122.4194
                    WHEN CITY = 'Chicago' THEN -87.6298
                    WHEN CITY = 'London' THEN -0.1278
                    WHEN CITY = 'Sydney' THEN 151.2093
                    ELSE 0.0
                END
            """,
            baseColumn="CITY",
        )
        # Assign region name based on country and city
        .withColumn(
            "REGION_NAME",
            StringType(),
            expr="""
                CASE
                    WHEN country = 'USA' THEN
                        CASE
                            WHEN CITY = 'New York' THEN 'New York'
                            WHEN CITY = 'San Francisco' THEN 'California'
                            WHEN CITY = 'Chicago' THEN 'Illinois'
                            ELSE 'Unknown'
                        END
                    WHEN country = 'UK' THEN 'England'
                    WHEN country = 'AUS' THEN 'New South Wales'
                    ELSE 'Unknown'
                END
            """,
            baseColumn=["country", "CITY"],
        )
        # Internal IP address as integer, unique per device, omitted from output
        .withColumn(
            "internal_REQUESTIPADDRESS",
            LongType(),
            minValue=0x1000000000000,
            uniqueValues=n_ips,
            omit=True,
            baseColumnType="hash",
            baseColumn="internal_DEVICEID",
        )
        # Convert internal IP integer to dotted quad string
        .withColumn(
            "REQUESTIPADDRESS",
            StringType(),
            expr="""
                concat(
                    cast((internal_REQUESTIPADDRESS >> 24) & 255 as string), '.',
                    cast((internal_REQUESTIPADDRESS >> 16) & 255 as string), '.',
                    cast((internal_REQUESTIPADDRESS >> 8) & 255 as string), '.',
                    cast(internal_REQUESTIPADDRESS & 255 as string)
                )
            """,
            baseColumn="internal_REQUESTIPADDRESS",
        )
        # Generate user agent string using clientName and SESSION_ID
        .withColumn(
            "USERAGENT",
            StringType(),
            expr="concat('Launch/1.0+', CLIENTNAME, '(', CLIENTNAME, '/)/', SESSION_ID)",
            baseColumn=["CLIENTNAME", "SESSION_ID"],
        )
    )

    if output:
        spec.saveAsDataset(dataset=output)
        return None

    df = spec.build()
    df = (
        df
        .withColumn("EVENT_HOUR", hour(col("EVENT_TIMESTAMP")))
        .withColumn("EVENT_DATE", to_date(col("EVENT_TIMESTAMP")))
    )
    return df


def generate_gaming_demo(
    spark,
    catalog,
    schema="gaming",
    volume="raw_data",
    rows=4_500_000,
    seed=42,
):
    """Generate complete gaming demo dataset and write to Unity Catalog.

    Creates the login_events table with post-processed time columns and writes
    raw JSON to a UC Volume for bronze ingestion, plus a clean Delta table.

    Args:
        spark: SparkSession.
        catalog: Unity Catalog catalog name.
        schema: Schema/database name (default: "gaming").
        volume: Volume name for raw data output (default: "raw_data").
        rows: Total login events to generate.
        seed: Random seed for reproducibility.
    """
    from ..utils.output import write_medallion

    login_events = generate_login_events(spark, rows=rows, seed=seed)

    write_medallion(
        tables={
            "login_events": login_events,
        },
        catalog=catalog,
        schema=schema,
        volume=volume,
    )
