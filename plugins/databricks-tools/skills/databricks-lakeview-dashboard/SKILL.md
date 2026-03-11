---
name: databricks-lakeview-dashboard
description: Programmatically create and manage Lakeview dashboards in Databricks
---

# Databricks Lakeview Dashboard Skill

This skill enables programmatic creation of beautiful Lakeview dashboards in Databricks using the Dashboard API.

## Overview

Lakeview dashboards are stored as JSON documents with a `serialized_dashboard` payload. This skill provides the schema and helper utilities to create dashboards programmatically.

## Prerequisites

1. Databricks CLI authenticated: `databricks auth login --profile <profile>`
2. SQL warehouse available for dashboard queries
3. Unity Catalog tables/views for data sources

## Choosing the Right Visualization Type

> ⚠️ **Important:** Before creating any visualization, think carefully about what chart type best represents your data. Not everything should be a bar chart! Consider the data pattern, the story you want to tell, and the audience.

### Visualization Selection Guide

| Data Pattern | Best Visualization | When to Use |
|-------------|-------------------|-------------|
| **Trend over time** | Line, Area | Showing how metrics change over days/weeks/months |
| **Comparing categories** | Bar (horizontal for many categories) | Comparing values across distinct groups |
| **Part of a whole** | Pie, Funnel | Showing proportions that sum to 100% |
| **Distribution** | Histogram, Box | Understanding data spread, quartiles, outliers |
| **Relationship between variables** | Scatter, Bubble | Correlation analysis, finding patterns |
| **Geographic data** | Choropleth Map, Point Map | Regional comparisons, location-based insights |
| **Flow/Process stages** | Sankey, Funnel | Conversion tracking, user journeys, data flow |
| **Single KPI** | Counter | Highlighting one important metric |
| **Cumulative changes** | Waterfall | Financial analysis, showing positive/negative contributions |
| **Two metrics on different scales** | Combo (dual-axis) | Comparing revenue vs. count, etc. |
| **Retention/Cohort analysis** | Cohort | User retention over time by signup cohort |
| **Pattern in 2D categories** | Heatmap | Finding hotspots in category intersections |
| **Detailed data exploration** | Table, Pivot | When users need to see raw values |

### Available Visualization Types in Lakeview

1. **Area** - Combines line and bar to show cumulative changes over time across groups
2. **Bar** - Compare values across categories; supports stacked, 100% stacked, or grouped
3. **Box** - Show distribution summaries with quartiles; identify outliers
4. **Bubble** - Scatter with size dimension; visualize 3 variables at once
5. **Choropleth Map** - Color regions (countries, states) by aggregate values
6. **Cohort** - Track user retention and behavior patterns over time
7. **Combo** - Mix line and bar on dual axes for different scales
8. **Counter** - Display a single prominent value (KPI)
9. **Funnel** - Analyze metric changes through sequential stages
10. **Heatmap** - Use color intensity to show values in a grid
11. **Histogram** - Plot frequency distribution of values
12. **Line** - Standard time series and trend analysis
13. **Pie** - Show proportions (avoid for time series or many categories)
14. **Pivot** - Tabular summaries with conditional formatting
15. **Point Map** - Plot data at lat/long coordinates
16. **Sankey** - Visualize flow between stages
17. **Scatter** - Show relationship between two numeric variables
18. **Table** - Raw data display with formatting
19. **Waterfall** - Show cumulative effect of sequential values

### Anti-Patterns to Avoid

- ❌ **Pie charts with many slices** - Use bar chart if more than 5-6 categories
- ❌ **Bar charts for time series** - Use line chart to show trends over time
- ❌ **Line charts for categorical data** - Use bar chart for discrete categories
- ❌ **Tables when a visual would be clearer** - Reserve tables for detail views
- ❌ **Multiple counters when comparison matters** - Use bar chart to enable comparison

Reference: [Databricks Lakeview Visualization Types Documentation](https://docs.databricks.com/en/dashboards/lakeview-visualization-types.html)

## API Endpoints

### Create Dashboard
```bash
databricks api post /api/2.0/lakeview/dashboards --profile <profile> --json '{
  "display_name": "My Dashboard",
  "warehouse_id": "<warehouse_id>",
  "parent_path": "/Users/<email>",
  "serialized_dashboard": "<json_string>"
}'
```

### Update Dashboard
```bash
databricks api patch /api/2.0/lakeview/dashboards/<dashboard_id> --profile <profile> --json '{
  "display_name": "Updated Name",
  "serialized_dashboard": "<json_string>"
}'
```

### Get Dashboard
```bash
databricks api get /api/2.0/lakeview/dashboards/<dashboard_id> --profile <profile>
```

### List Dashboards
```bash
databricks api get /api/2.0/lakeview/dashboards --profile <profile>
```

### Publish Dashboard
```bash
databricks api post /api/2.0/lakeview/dashboards/<dashboard_id>/published --profile <profile>
```

## Serialized Dashboard Schema

The `serialized_dashboard` is a JSON string with this structure:

```json
{
  "datasets": [...],
  "pages": [...],
  "uiSettings": {...}
}
```

### Datasets

Datasets define the SQL queries that power visualizations:

```json
{
  "datasets": [
    {
      "name": "unique_id_123",
      "displayName": "Sales Data",
      "queryLines": [
        "SELECT * FROM catalog.schema.table"
      ]
    }
  ]
}
```

### Pages

Pages contain the layout of widgets:

```json
{
  "pages": [
    {
      "name": "page_id_123",
      "displayName": "Overview",
      "pageType": "PAGE_TYPE_CANVAS",
      "layout": [...]
    }
  ]
}
```

### Layout and Positioning

Widgets are positioned on a 6-column grid:

```json
{
  "layout": [
    {
      "widget": {...},
      "position": {
        "x": 0,       // Column (0-5)
        "y": 0,       // Row
        "width": 2,   // Columns to span (1-6)
        "height": 3   // Rows to span
      }
    }
  ]
}
```

## Widget Types

### Bar Chart
```json
{
  "widget": {
    "name": "widget_id",
    "queries": [{
      "name": "main_query",
      "query": {
        "datasetName": "dataset_id",
        "fields": [
          {"name": "category", "expression": "`category`"},
          {"name": "sum_amount", "expression": "SUM(`amount`)"}
        ],
        "disaggregated": false
      }
    }],
    "spec": {
      "version": 3,
      "widgetType": "bar",
      "encodings": {
        "x": {
          "fieldName": "category",
          "scale": {"type": "categorical"},
          "displayName": "Category"
        },
        "y": {
          "fieldName": "sum_amount",
          "scale": {"type": "quantitative"},
          "displayName": "Total Amount"
        },
        "label": {"show": true}
      },
      "frame": {
        "showTitle": true,
        "title": "Sales by Category"
      },
      "mark": {
        "colors": ["#FFAB00", "#00A972", "#FF3621"]
      }
    }
  }
}
```

### Line Chart
```json
{
  "widget": {
    "name": "widget_id",
    "queries": [{
      "name": "main_query",
      "query": {
        "datasetName": "dataset_id",
        "fields": [
          {"name": "date", "expression": "DATE_TRUNC(\"MONTH\", `sale_date`)"},
          {"name": "series", "expression": "`category`"},
          {"name": "value", "expression": "SUM(`amount`)"}
        ],
        "disaggregated": false
      }
    }],
    "spec": {
      "version": 3,
      "widgetType": "line",
      "encodings": {
        "x": {
          "fieldName": "date",
          "scale": {"type": "temporal"},
          "displayName": "Date"
        },
        "y": {
          "fieldName": "value",
          "scale": {"type": "quantitative"},
          "displayName": "Amount"
        },
        "color": {
          "fieldName": "series",
          "scale": {"type": "categorical"},
          "displayName": "Category"
        }
      }
    }
  }
}
```

### Pie Chart
```json
{
  "widget": {
    "name": "widget_id",
    "queries": [{
      "name": "main_query",
      "query": {
        "datasetName": "dataset_id",
        "fields": [
          {"name": "count", "expression": "COUNT(`*`)"},
          {"name": "category", "expression": "`category`"}
        ],
        "disaggregated": false
      }
    }],
    "spec": {
      "version": 3,
      "widgetType": "pie",
      "encodings": {
        "angle": {
          "fieldName": "count",
          "scale": {"type": "quantitative"},
          "displayName": "Count"
        },
        "color": {
          "fieldName": "category",
          "scale": {"type": "categorical"},
          "displayName": "Category"
        }
      },
      "frame": {
        "showTitle": true,
        "title": "Distribution by Category"
      }
    }
  }
}
```

### Counter (KPI)
```json
{
  "widget": {
    "name": "widget_id",
    "queries": [{
      "name": "main_query",
      "query": {
        "datasetName": "dataset_id",
        "fields": [
          {"name": "total", "expression": "SUM(`amount`)"}
        ],
        "disaggregated": true
      }
    }],
    "spec": {
      "version": 2,
      "widgetType": "counter",
      "encodings": {
        "value": {
          "fieldName": "total",
          "displayName": "Total Revenue"
        }
      },
      "frame": {
        "showTitle": true,
        "title": "Total Revenue"
      }
    }
  }
}
```

### Scatter Plot
```json
{
  "widget": {
    "name": "widget_id",
    "queries": [{
      "name": "main_query",
      "query": {
        "datasetName": "dataset_id",
        "fields": [
          {"name": "x_val", "expression": "`price`"},
          {"name": "y_val", "expression": "`quantity`"},
          {"name": "group", "expression": "`category`"}
        ],
        "disaggregated": true
      }
    }],
    "spec": {
      "version": 3,
      "widgetType": "scatter",
      "encodings": {
        "x": {
          "fieldName": "x_val",
          "scale": {"type": "quantitative"},
          "displayName": "Price"
        },
        "y": {
          "fieldName": "y_val",
          "scale": {"type": "quantitative"},
          "displayName": "Quantity"
        },
        "color": {
          "fieldName": "group",
          "scale": {"type": "categorical"},
          "displayName": "Category"
        }
      }
    }
  }
}
```

### Area Chart
```json
{
  "spec": {
    "version": 3,
    "widgetType": "area",
    "encodings": {
      "x": {"fieldName": "date", "scale": {"type": "temporal"}},
      "y": {"fieldName": "value", "scale": {"type": "quantitative"}},
      "color": {"fieldName": "series", "scale": {"type": "categorical"}}
    }
  }
}
```

### Histogram
```json
{
  "spec": {
    "version": 3,
    "widgetType": "histogram",
    "encodings": {
      "x": {
        "fieldName": "bin_field",
        "scale": {"type": "categorical", "sort": {"by": "natural-order"}}
      },
      "y": {
        "fieldName": "count",
        "scale": {"type": "quantitative"}
      },
      "color": {
        "fieldName": "category",
        "scale": {
          "type": "categorical",
          "mappings": [
            {"value": "good", "color": "#00A972"},
            {"value": "bad", "color": "#FF3621"}
          ]
        }
      }
    }
  }
}
```

### Table
```json
{
  "widget": {
    "name": "widget_id",
    "queries": [{
      "name": "main_query",
      "query": {
        "datasetName": "dataset_id",
        "fields": [
          {"name": "col1", "expression": "`column1`"},
          {"name": "col2", "expression": "`column2`"}
        ],
        "disaggregated": true
      }
    }],
    "spec": {
      "version": 1,
      "widgetType": "table",
      "encodings": {
        "columns": [
          {
            "fieldName": "col1",
            "type": "string",
            "displayAs": "string",
            "title": "Column 1",
            "displayName": "Column 1"
          },
          {
            "fieldName": "col2",
            "type": "float",
            "displayAs": "number",
            "numberFormat": "0.00",
            "title": "Column 2",
            "alignContent": "right"
          }
        ]
      }
    }
  }
}
```

## Filter Widgets

### Date Range Picker
```json
{
  "spec": {
    "version": 2,
    "widgetType": "filter-date-range-picker",
    "encodings": {
      "fields": [{
        "fieldName": "date_field",
        "displayName": "Date",
        "queryName": "filter_query_name"
      }]
    },
    "frame": {"showTitle": true, "title": "Select Date Range"}
  }
}
```

### Single Select Dropdown
```json
{
  "spec": {
    "version": 2,
    "widgetType": "filter-single-select",
    "encodings": {
      "fields": [{
        "fieldName": "category",
        "displayName": "Category",
        "queryName": "filter_query_name"
      }]
    },
    "frame": {"showTitle": true, "title": "Select Category"}
  }
}
```

### Multi-Select
```json
{
  "spec": {
    "version": 2,
    "widgetType": "filter-multi-select",
    "encodings": {
      "fields": [{
        "fieldName": "region",
        "displayName": "Region",
        "queryName": "filter_query_name"
      }]
    }
  }
}
```

### Text Entry Filter
```json
{
  "spec": {
    "version": 2,
    "widgetType": "filter-text-entry",
    "encodings": {
      "fields": [{
        "fieldName": "search_field",
        "displayName": "Search",
        "queryName": "filter_query_name"
      }]
    }
  }
}
```

## Scale Types

- `categorical` - For discrete categories
- `quantitative` - For numeric values
- `temporal` - For date/time values

## Sorting Charts (Top K / Most X Visualizations)

**Important:** When a chart is visualizing "top K" or "most X" data (e.g., "Top 10 Users", "Most Active Accounts", "Highest Revenue Categories"), the chart should automatically sort by the y-axis value in **descending order** so the largest values appear first.

To show largest values first (left-to-right), add a `sort` configuration to the x-axis scale with `"by": "y-reversed"`:

```json
{
  "x": {
    "fieldName": "user_name",
    "scale": {
      "type": "categorical",
      "sort": {
        "by": "y-reversed"
      }
    },
    "displayName": "User"
  }
}
```

### Example: Most Active Users by Tokens Consumed

```json
{
  "widget": {
    "name": "top_users_widget",
    "queries": [{
      "name": "main_query",
      "query": {
        "datasetName": "usage_data",
        "fields": [
          {"name": "user_name", "expression": "`user_name`"},
          {"name": "sum_tokens", "expression": "SUM(`tokens_consumed`)"}
        ],
        "disaggregated": false
      }
    }],
    "spec": {
      "version": 3,
      "widgetType": "bar",
      "encodings": {
        "x": {
          "fieldName": "user_name",
          "scale": {
            "type": "categorical",
            "sort": {
              "by": "y-reversed"
            }
          },
          "displayName": "User"
        },
        "y": {
          "fieldName": "sum_tokens",
          "scale": {"type": "quantitative"},
          "displayName": "Total Tokens Consumed"
        },
        "label": {"show": true}
      },
      "frame": {
        "showTitle": true,
        "title": "Most Active Users by Tokens Consumed"
      }
    }
  }
}
```

### Sort Options

- `"by": "y"` - Sort by y-axis value, smallest first (ascending)
- `"by": "y-reversed"` - Sort by y-axis value, largest first (descending)
- `"by": "x"` - Sort by x-axis value alphabetically (A-Z)
- `"by": "x-reversed"` - Sort by x-axis value reverse alphabetically (Z-A)
- `"by": "natural-order"` - Preserve the order from the query

**Best Practice:** Use `"by": "y-reversed"` when the chart title contains words like "Top", "Most", "Highest", "Largest", or similar superlatives indicating a ranking - this places the highest values first (left-to-right).

> ⚠️ **Common Mistake:** Do NOT use `"direction": "descending"` or `"direction": "ascending"` syntax - this does not work in Lakeview. The correct syntax uses the `-reversed` suffix on the sort key (e.g., `"by": "y-reversed"` not `"by": "y", "direction": "descending"`).

## Color Palettes

Default Databricks colors:
```json
["#FFAB00", "#00A972", "#FF3621", "#8BCAE7", "#AB4057", "#99DDB4", "#FCA4A1", "#919191", "#BF7080"]
```

Custom color mappings:
```json
{
  "scale": {
    "type": "categorical",
    "mappings": [
      {"value": "Success", "color": "#00A972"},
      {"value": "Warning", "color": "#FFAB00"},
      {"value": "Error", "color": "#FF3621"}
    ]
  }
}
```

## UI Settings

```json
{
  "uiSettings": {
    "theme": {
      "widgetHeaderAlignment": "ALIGNMENT_UNSPECIFIED"
    },
    "applyModeEnabled": false
  }
}
```

## Complete Example

See `resources/example_dashboard.json` for a complete working example.

## Python Helper

Use `resources/lakeview_builder.py` for a Python class that simplifies dashboard creation:

```python
from lakeview_builder import LakeviewDashboard

dashboard = LakeviewDashboard("My Sales Dashboard")

# Add dataset
dashboard.add_dataset(
    "sales",
    "Sales Data",
    "SELECT * FROM catalog.schema.sales"
)

# Add bar chart
dashboard.add_bar_chart(
    dataset_name="sales",
    x_field="category",
    y_field="amount",
    y_agg="SUM",
    title="Sales by Category",
    position={"x": 0, "y": 0, "width": 3, "height": 4}
)

# Add counter
dashboard.add_counter(
    dataset_name="sales",
    value_field="amount",
    value_agg="SUM",
    title="Total Sales",
    position={"x": 3, "y": 0, "width": 1, "height": 2}
)

# Get JSON for API
json_payload = dashboard.to_json()
```

## Tips

1. **Widget IDs**: Generate unique 8-character hex IDs for widget and dataset names
2. **Grid Layout**: Dashboard uses a 6-column grid. Plan layout before building.
3. **Dataset Reuse**: Multiple widgets can share the same dataset
4. **Filter Queries**: Filters need special query names for associativity
5. **Disaggregated**: Set to `true` for raw data (tables), `false` for aggregations

## Troubleshooting

- **Dashboard not rendering**: Check serialized_dashboard is valid JSON string (escaped)
- **Widget empty**: Verify dataset name matches exactly
- **Filters not working**: Ensure filter query names follow the pattern
