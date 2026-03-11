# Performance Tuning Decision Trees

## "My query is slow" — Where to start?

```
Is it a DBSQL query or a Spark job/notebook?
├── DBSQL Query
│   ├── Check Query Profile for the slow operator
│   ├── Is compilation time high (>30% of total)?
│   │   └── Yes → Check for complex views, deeply nested CTEs, or stats issues
│   ├── Is execution time high?
│   │   ├── Check for spill → Increase warehouse size
│   │   ├── Check for skew → AQE should handle; if not, salt or restructure
│   │   ├── Check scan size → Add filters, use partition pruning, check clustering
│   │   └── Check join strategy → Broadcast small tables
│   └── Is queue/wait time high?
│       └── Increase warehouse cluster count for concurrency
│
└── Spark Job/Notebook
    ├── Open Spark UI → Jobs → Sort by duration
    ├── Click longest job → find longest stage
    ├── Check stage for the 4 S's:
    │   ├── Skew? → AQE / salting / broadcast join
    │   ├── Spill? → More memory / more partitions / broadcast
    │   ├── Shuffle? → Broadcast joins / reduce shuffles / better join order
    │   └── Small files? → OPTIMIZE / auto-compact / optimized writes
    ├── Check for UDFs (defeats Photon) → Replace with built-in functions
    └── Check cluster sizing → Right-size instances and node count
```

## "My streaming job is falling behind"

```
Is batch duration > trigger interval?
├── Yes → Processing is too slow
│   ├── Check Spark UI for the 4 S's in each batch
│   ├── Is state store large?
│   │   ├── Yes → Enable RocksDB state store + changelog checkpointing
│   │   └── Check watermark — is it properly configured?
│   ├── Is the cluster fully utilized?
│   │   ├── No → Reduce trigger interval or increase input rate
│   │   └── Yes → Scale up cluster (more/bigger nodes)
│   └── Are you using UDFs? → Replace with built-in functions
│
└── No → Something else is slow
    ├── Is checkpoint time high?
    │   └── Yes → Enable async + changelog checkpointing
    ├── Is scheduling delay high?
    │   └── Yes → Check cluster health, consider dedicated cluster
    └── Is input deserialization slow?
        └── Yes → Check source format, use Delta or Kafka with binary
```
