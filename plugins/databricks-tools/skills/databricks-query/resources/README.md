# Databricks Query Pretty Printer

A self-contained Python script that formats Databricks SQL query results as beautiful ASCII tables.

## Features

- ✨ Beautiful Unicode box-drawing characters
- 📊 Intelligent column width calculation
- 🖥️  Automatic terminal width detection
- 📏 Handles tables with many columns
- 🗑️  Automatic file cleanup
- 📦 Uses only Python standard library (no dependencies!)

## Usage

### Pipe from databricks CLI (stdin):

```bash
databricks api post /api/2.0/sql/statements/ \
  --json='{"statement": "SELECT * FROM my_table LIMIT 10",
           "warehouse_id": "b21e2c56a857fe88",
           "format":"JSON_ARRAY",
           "wait_timeout":"50s"}' \
  --profile=<your-profile> | python3 databricks_query_pretty.py
```

### From a file:

```bash
# Save output to file
databricks api post /api/2.0/sql/statements/ \
  --json='{"statement": "SELECT 1", "warehouse_id": "b21e2c56a857fe88", "format":"JSON_ARRAY"}' \
  --profile=<your-profile> > output.json

# Format it (file will be automatically deleted after parsing)
python3 databricks_query_pretty.py output.json
```

### Keep the file after parsing:

```bash
python3 databricks_query_pretty.py output.json --no-delete
```

### Custom table width:

```bash
# Set maximum table width to 200 characters
cat output.json | python3 databricks_query_pretty.py --max-width 200

# Set maximum column width to 30 characters
cat output.json | python3 databricks_query_pretty.py --max-col-width 30
```

## Example Output

```
┌────┬──────────────┬───────────────────┬─────────────┬────────┐
│ id │     name     │       email       │  department │ salary │
├────┼──────────────┼───────────────────┼─────────────┼────────┤
│ 1  │ Alice Johnson│ alice@example.com │ Engineering │ 125000 │
│ 2  │ Bob Smith    │ bob@example.com   │ Marketing   │ 95000  │
│ 3  │ Carol Willi..│ carol@example.com │ Sales       │ 110000 │
└────┴──────────────┴───────────────────┴─────────────┴────────┘

3 row(s) in set
```

## Options

```
positional arguments:
  file                  JSON file containing Databricks API response
                        (will be deleted after parsing). If omitted, reads from stdin.

optional arguments:
  --max-width WIDTH     Maximum table width in characters (default: terminal width)
  --max-col-width WIDTH Maximum width for any single column (default: 50)
  --no-delete           Do not delete the input file after parsing
  -h, --help           Show help message
```

## Requirements

- Python 3.6+
- No external dependencies (uses only standard library)
