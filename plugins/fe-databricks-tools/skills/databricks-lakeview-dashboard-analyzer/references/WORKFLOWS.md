# Dashboard Analysis Workflows

Complete workflow patterns for common dashboard analysis tasks.

## Workflow 1: Dashboard Summary

Generate a comprehensive overview of a dashboard's structure and key insights.

```bash
# 1. Navigate and wait for load
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL"}'
mcp__chrome-devtools__wait_for '{"text": "Dashboard", "timeout": 15000}'

# 2. Take snapshot for structure
mcp__chrome-devtools__take_snapshot
# Analyze: titles, filters, widget labels, table headers

# 3. Take screenshot for visuals
mcp__chrome-devtools__take_screenshot
# Analyze: chart types, trends, colors, layouts
```

**Generate summary including:**
- Dashboard purpose
- Key metrics (KPIs, counters)
- Available filters
- Main insights from charts
- Data quality notes

---

## Workflow 2: Filtered Data Extraction

Apply filters and extract specific data subsets.

```bash
# 1. Navigate to dashboard
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL"}'

# 2. Take snapshot to find filters
mcp__chrome-devtools__take_snapshot
# Look for combobox, textbox, or dropdown elements

# 3. Apply filters (click combobox, select options)
mcp__chrome-devtools__click '{"uid": "FILTER_UID"}'
mcp__chrome-devtools__take_snapshot  # Get option UIDs
mcp__chrome-devtools__click '{"uid": "OPTION_UID"}'

# 4. Wait for refresh
mcp__chrome-devtools__wait_for '{"text": "Loaded", "timeout": 10000}'

# 5. Extract table data from snapshot
mcp__chrome-devtools__take_snapshot
# Parse table rows from output
```

**Alternative: URL-based filtering (faster)**

```bash
# Modify URL parameters directly instead of clicking
# Pattern: &f_PAGEID%7EFILTERID=VALUE
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL&f_abc123%7Edef456=FilterValue"}'
```

---

## Workflow 3: Time Series Analysis

Analyze trends and patterns in time-based visualizations.

```bash
# 1. Navigate to dashboard with time series charts
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL"}'
mcp__chrome-devtools__wait_for '{"text": "Dashboard", "timeout": 15000}'

# 2. Take screenshot of line/area charts
mcp__chrome-devtools__take_screenshot
```

**Visual analysis checklist:**
- Identify trends (upward, downward, flat)
- Note seasonality patterns
- Spot anomalies or outliers
- Calculate approximate growth rates from visual

```bash
# 3. Extract exact values if available in table
mcp__chrome-devtools__take_snapshot
# Look for table elements with time series data
```

---

## Workflow 4: Interactive Q&A

Answer specific questions by navigating to relevant data.

**Example: "What is the total sales for Product A?"**

```bash
# 1. Navigate to dashboard
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL"}'

# 2. Take snapshot to find product filter
mcp__chrome-devtools__take_snapshot
# Look for filter labeled "Product" or similar

# 3. Apply filter for "Product A"
mcp__chrome-devtools__click '{"uid": "PRODUCT_FILTER_UID"}'
mcp__chrome-devtools__take_snapshot
mcp__chrome-devtools__click '{"uid": "PRODUCT_A_OPTION_UID"}'

# 4. Take snapshot to read KPI counter or table sum
mcp__chrome-devtools__wait_for '{"text": "Product A", "timeout": 10000}'
mcp__chrome-devtools__take_snapshot
# Extract numeric value from counter or table

# 5. Answer with context from visualizations
```

---

## Workflow 5: Multi-Tab Dashboard with Filters

Navigate dashboards with multiple tabs and interdependent filters.

```bash
# 1. Navigate and discover structure
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL"}'
mcp__chrome-devtools__take_snapshot
```

**Identify tabs in snapshot output:**
```
uid=X_1 tab "Overview" selectable selected
uid=X_2 tab "Details" selectable
uid=X_3 tab "Admin View" selectable
```

```bash
# 2. Navigate to appropriate tab
# Option A: Click the tab
mcp__chrome-devtools__click '{"uid": "X_2"}'

# Option B: Navigate via URL (tabs have page IDs)
# Current: .../pages/abc123  → Details tab: .../pages/def456
mcp__chrome-devtools__navigate_page '{"type": "url", "url": ".../pages/def456?o=ORGID"}'

# 3. Apply filters via URL manipulation (faster)
# Pattern: &f_PAGEID%7EFILTER1=Value1&f_PAGEID%7EFILTER2=Value2
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "...&f_PAGEID%7EFILTER1=NewValue"}'

# 4. Scroll to visualizations (container scrolling required)
mcp__chrome-devtools__evaluate_script '{"function": "() => { const c = document.querySelector(\"[class*=dbsql-legacy-mfe-page]\"); if(c){c.scrollBy(0,600);return{scrolled:true};} return{scrolled:false}; }"}'

# 5. Take screenshot to analyze charts
mcp__chrome-devtools__take_screenshot

# 6. Compare across filter values by changing URL and re-capturing
# IMPORTANT: Check Y-axis scales - they may differ between views!
```

**Key Patterns:**
- Tabs appear as `tab` role elements in snapshots
- Filter parameters follow pattern `f_PAGEID%7EFILTERID=VALUE` (`%7E` = `~`)
- Container scrolling required - `window.scrollTo()` won't work
- Y-axis scales vary between filtered views - always check when comparing

---

## Workflow 6: Answering Comparative Questions

Compare values across categories (e.g., "Which product has highest X?", "Compare A vs B").

```bash
# 1. Navigate and identify relevant visualization
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL"}'
mcp__chrome-devtools__take_snapshot
# Look for relevant chart titles, table headers, KPI labels

# 2. For each category to compare:
```

**Loop for each category:**
```bash
# a. Apply filter (URL manipulation is fastest)
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL&f_...=CategoryA"}'

# b. Scroll to the relevant chart
mcp__chrome-devtools__evaluate_script '{"function": "() => { const c = document.querySelector(\"[class*=dbsql-legacy-mfe-page]\"); if(c){c.scrollBy(0,600);return{scrolled:true};} return{scrolled:false}; }"}'

# c. Take screenshot
mcp__chrome-devtools__take_screenshot

# d. Record the value (note the Y-axis scale!)
```

```bash
# 3. Compile comparison
# Example: "Peak for Category A: 30M, Peak for Category B: 7M"
# Note: If scales differ, mention it - don't just compare raw visual heights
```

**Pitfalls to Avoid:**
- Don't assume same Y-axis scale across filter changes
- Check which tab has the granularity you need (Overview vs Detail views)
- Session may expire during long comparisons - save progress

---

## Workflow 7: Data Export for Analysis

Download underlying data for detailed analysis outside the dashboard.

```bash
# 1. Navigate to dashboard and find target visualization
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "DASHBOARD_URL"}'
mcp__chrome-devtools__take_snapshot
# Find: uid=X_Y heading "Chart Title" level="4"

# 2. Hover over chart heading to reveal ellipsis button
mcp__chrome-devtools__hover '{"uid": "X_Y"}'
# Returns new snapshot with ellipsis button visible:
# uid=X_Z button expandable haspopup="menu"

# 3. Click ellipsis button
mcp__chrome-devtools__click '{"uid": "X_Z"}'

# 4. Navigate menu with keyboard (more reliable than clicking)
mcp__chrome-devtools__press_key '{"key": "ArrowDown"}'  # Navigate to "Download"
mcp__chrome-devtools__press_key '{"key": "ArrowDown"}'  # (repeat as needed)
mcp__chrome-devtools__press_key '{"key": "ArrowRight"}' # Expand "Download" submenu
mcp__chrome-devtools__press_key '{"key": "ArrowRight"}' # Expand "Data" submenu
mcp__chrome-devtools__press_key '{"key": "Enter"}'      # Select "Download CSV"

# 5. File downloads to ~/Downloads/{ChartName}-{YYYY-MM-DD}.csv
# Wait a moment for download to complete

# 6. Read and analyze the downloaded CSV
```

**For smaller files (< 256KB):**
```bash
Read ~/Downloads/ChartName-2026-01-21.csv
```

**For larger files, use Python:**
```python
import csv
from collections import defaultdict

with open('ChartName-2026-01-21.csv', 'r') as f:
    reader = csv.DictReader(f)
    # Aggregate, filter, analyze as needed
```
