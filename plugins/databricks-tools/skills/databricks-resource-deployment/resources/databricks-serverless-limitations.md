# Serverless Compute Limitations

## Language & API Restrictions
- **R not supported**
- **Scala not supported** in notebooks
- **Only Spark Connect APIs** (RDD APIs unsupported)
- External UDFs: Internet access restrictions
- **Maven coordinates not supported**

## Storage & Access
- **DBFS root**: Available
- **AWS instance profile mounts**: Unavailable
- **Cross-workspace**: Requires same region, no IP ACLs/PrivateLink

## Query Limitations
- **Timeout**: 9000-second default (configurable for notebooks, none for jobs)
- **Spark UI**: Not available (use query profiling)
- **Spark logs**: Not available (client-side logs only)

## Feature Restrictions

### Compute Features (Unsupported)
- Policies
- Init scripts
- Instance pools
- Compute event logs
- **Environment variables** (use widgets for job parameters)

### Data Handling
- **DataFrame caching APIs** (cache, persist, unpersist): Throw exceptions
- **Global temporary views**: Not supported (use session views)
- **Hive SerDe tables**: Incompatible
- **Hive variable syntax**: Incompatible

### Streaming
- **Only Trigger.AvailableNow supported**
- Default and time-based triggers: Cannot be used

## Supported Data Sources
- **DML operations**: 13 sources (Delta, Kafka, Iceberg, etc.)
- **Read operations**: 27 sources (PostgreSQL, Snowflake, BigQuery, MongoDB, etc.)
