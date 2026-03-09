# Data Extraction Guide

Techniques for extracting data from Databricks AI/BI dashboards.

## Reading Tables from Snapshots

Snapshots capture table content directly:

```bash
mcp__chrome-devtools__take_snapshot
```

**Example snapshot output:**
```
uid=5_10 table
  uid=5_11 row "Name | Value | Status"
  uid=5_12 row "Item A | 123 | Active"
  uid=5_13 row "Item B | 456 | Inactive"
```

The snapshot shows:
- Column headers
- Row content
- Cell text values (pipe-separated)

---

## Extracting Tables with JavaScript

For more structured extraction:

```bash
mcp__chrome-devtools__evaluate_script '{"function": "() => { const table = document.querySelector(\"table\"); if (!table) return null; const headers = Array.from(table.querySelectorAll(\"thead th\")).map(th => th.textContent.trim()); const rows = Array.from(table.querySelectorAll(\"tbody tr\")).map(tr => Array.from(tr.querySelectorAll(\"td\")).map(td => td.textContent.trim())); return { headers, rows, count: rows.length }; }"}'
```

**Returns:**
```json
{
  "headers": ["Name", "Value", "Status"],
  "rows": [
    ["Item A", "123", "Active"],
    ["Item B", "456", "Inactive"]
  ],
  "count": 2
}
```

**Note:** `evaluate_script` requires a **function declaration** as a string, not arbitrary JavaScript code.

---

## Downloading Data via Ellipsis Menus

**Key Strategy:** Every visualization in Databricks AI/BI dashboards has an ellipsis menu (⋮) for downloading data.

### Menu Structure

```
⋮ (ellipsis button)
├── View fullscreen
├── Copy link to widget
├── Download ▸
│   ├── Data ▸
│   │   ├── Download CSV
│   │   ├── Download TSV
│   │   └── Download Excel
│   └── Download PNG
└── View dataset: {dataset_name}
```

**IMPORTANT:** The ellipsis button only appears when you **hover over the chart**.

### Step-by-Step Download Workflow

```bash
# 1. Take snapshot to find chart heading
mcp__chrome-devtools__take_snapshot
# Find: uid=X_Y heading "$DBUs by Product Line" level="4"

# 2. Hover over the chart heading to make ellipsis appear
mcp__chrome-devtools__hover '{"uid": "X_Y"}'
# Returns new snapshot with ellipsis button visible:
# uid=X_Z button expandable haspopup="menu"

# 3. Click the ellipsis button to open menu
mcp__chrome-devtools__click '{"uid": "X_Z"}'

# 4. USE KEYBOARD NAVIGATION for nested menus (more reliable than clicking)
mcp__chrome-devtools__press_key '{"key": "ArrowDown"}'  # Navigate to "Download"
mcp__chrome-devtools__press_key '{"key": "ArrowDown"}'  # (repeat as needed)
mcp__chrome-devtools__press_key '{"key": "ArrowRight"}' # Expand "Download" submenu
mcp__chrome-devtools__press_key '{"key": "ArrowRight"}' # Expand "Data" submenu
mcp__chrome-devtools__press_key '{"key": "Enter"}'      # Select "Download CSV"

# 5. File downloads automatically to ~/Downloads
# Filename pattern: {ChartName}-{YYYY-MM-DD}.csv
# Example: "DBUs by Product Line-2026-01-21.csv"
```

### Why Keyboard Navigation?

- Hover/click on nested menu items can cause menus to collapse
- Keyboard navigation (ArrowDown, ArrowRight, Enter) is more reliable
- Works consistently across all AI/BI dashboards

---

## Analyzing Downloaded Files

### Small Files (< 256KB)

Use the Read tool directly:

```bash
Read ~/Downloads/DBUs by Product Line-2026-01-21.csv
```

### Larger Files

Use Python for analysis:

```python
import csv
from datetime import datetime, timedelta
from collections import defaultdict

# Read the CSV
with open('DBUs by Product Line-2026-01-21.csv', 'r') as f:
    reader = csv.DictReader(f)
    data = list(reader)

# Example: Calculate totals by product line
totals = defaultdict(float)
for row in data:
    totals[row['product_line']] += float(row['dbus'])

# Example: Compare time periods
def parse_date(s):
    return datetime.strptime(s, '%Y-%m-%d')

recent = [r for r in data if parse_date(r['date']) > datetime.now() - timedelta(days=14)]
previous = [r for r in data if parse_date(r['date']) <= datetime.now() - timedelta(days=14)]
```

---

## Workflow: Answer Questions via Data Download

1. User asks a question about customer data
2. Identify which visualization/table contains relevant data
3. Hover over the chart to reveal the ellipsis button
4. Click ellipsis, then keyboard navigate: Download → Data → Download CSV
5. Read the downloaded file and analyze to answer the question

**Example question:** "What is the growth trend for Product X over the last month?"

```bash
# 1. Find and download the relevant chart data
# (follow download workflow above)

# 2. Analyze with Python
```

```python
import csv
from datetime import datetime, timedelta

with open('Revenue by Product-2026-01-21.csv', 'r') as f:
    reader = csv.DictReader(f)
    product_x = [r for r in reader if r['product'] == 'Product X']

# Sort by date
product_x.sort(key=lambda r: r['date'])

# Calculate week-over-week growth
last_week = sum(float(r['revenue']) for r in product_x[-7:])
prev_week = sum(float(r['revenue']) for r in product_x[-14:-7])
growth = (last_week - prev_week) / prev_week * 100

print(f"Week-over-week growth: {growth:.1f}%")
```

---

## Extracting Specific Values with JavaScript

### Get all text from a container

```bash
mcp__chrome-devtools__evaluate_script '{"function": "() => { const container = document.querySelector(\".dashboard-container\"); return container ? container.innerText : null; }"}'
```

### Get KPI/counter values

```bash
mcp__chrome-devtools__evaluate_script '{"function": "() => { const counters = Array.from(document.querySelectorAll(\"[data-testid*=counter], .kpi-value, .metric-value\")); return counters.map(c => ({ text: c.innerText, label: c.closest(\".widget\")?.querySelector(\"h4\")?.innerText })); }"}'
```

### Get chart legend items

```bash
mcp__chrome-devtools__evaluate_script '{"function": "() => { const legends = Array.from(document.querySelectorAll(\".legend-item, [class*=legend]\")); return legends.map(l => l.innerText); }"}'
```

---

## Handling Paginated Tables

If tables are paginated:

```bash
# 1. Take snapshot to find pagination controls
mcp__chrome-devtools__take_snapshot
# Look for: "Next", "Page 1 of 5", "Show all", etc.

# 2. Click "Next page" or increase rows per page
mcp__chrome-devtools__click '{"uid": "NEXT_PAGE_UID"}'

# 3. Wait for new data to load
mcp__chrome-devtools__wait_for '{"text": "Page 2", "timeout": 10000}'

# 4. Extract data from new page
mcp__chrome-devtools__take_snapshot
```

**For "Show all" option:**
```bash
# Look for dropdown to increase rows per page
mcp__chrome-devtools__take_snapshot
# Find rows-per-page selector
mcp__chrome-devtools__click '{"uid": "ROWS_SELECTOR_UID"}'
mcp__chrome-devtools__take_snapshot
# Select maximum option
mcp__chrome-devtools__click '{"uid": "MAX_ROWS_OPTION_UID"}'
```

---

## Saving Extracted Data

After extracting data, save it for later use:

```python
import json

# Save as JSON
data = {"headers": [...], "rows": [...]}
with open('extracted_data.json', 'w') as f:
    json.dump(data, f, indent=2)

# Save as CSV
import csv
with open('extracted_data.csv', 'w', newline='') as f:
    writer = csv.writer(f)
    writer.writerow(headers)
    writer.writerows(rows)
```

Use the Write tool to create files from extracted data during analysis.
