# Common Anti-Patterns Checklist

Run through this checklist when reviewing customer code:

| Anti-Pattern | Why It's Bad | Fix |
|---|---|---|
| `SELECT *` | Reads all columns, defeats column pruning | Select only needed columns |
| Python/Scala UDFs | Defeats Photon, requires serialization | Use built-in SQL functions |
| `.collect()` on large data | Brings all data to driver, OOM risk | Use aggregations or `.take(n)` |
| `.count()` for existence check | Scans entire dataset | Use `.limit(1).count()` |
| Cartesian joins | O(n*m) explosion | Add join condition or use broadcast |
| Union of many small queries | Creates many stages, lots of overhead | Combine into single scan |
| `.repartition()` before write | Unnecessary full shuffle | Use `.coalesce()` or optimized writes |
| Caching everything | Wastes memory, evicts useful data | Only cache DataFrames used by multiple actions |
| `ORDER BY` without `LIMIT` | Sorts entire dataset | Add `LIMIT` or remove if not needed |
| Repeated subqueries | Same data scanned multiple times | Use CTEs or temp views |
| Not using Delta | Miss out on pruning, caching, ACID | Convert to Delta format |
| Manual partitioning on new tables | Fragile, hard to change | Use Liquid Clustering |
| Skipping ANALYZE TABLE | Optimizer makes bad decisions | Run ANALYZE after major data changes |
