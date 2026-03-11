---
name: google-sheets
description: Create and manage well-formatted Google Sheets with cells, formulas, charts, colors, and multiple sheets using gcloud CLI + curl
---

# Google Sheets Skill

Create beautiful, well-formatted Google Sheets using gcloud CLI + curl. This skill provides patterns and utilities for working with spreadsheets including cell updates, find/replace, rich formatting with colors, multiple sheets with references, formulas, and charts.

## Authentication

**Run `/google-auth` first** to authenticate with Google Workspace, or use the shared auth module:

```bash
# Check authentication status
python3 ../google-auth/resources/google_auth.py status

# Login if needed (includes automatic retry if OAuth times out)
python3 ../google-auth/resources/google_auth.py login

# Get access token for API calls
TOKEN=$(python3 ../google-auth/resources/google_auth.py token)
```

All Google skills share the same authentication. See `/google-auth` for details on scopes and troubleshooting.

### CRITICAL: If Authentication Fails

**If the login command fails**, it means the user did NOT complete the OAuth flow in the browser.

**DO NOT:**
- Try alternative authentication methods
- Create OAuth credentials manually
- Attempt to set up service accounts

**ONLY solution:**
- Re-run `python3 ../google-auth/resources/google_auth.py login`
- The script includes automatic retry logic with clear instructions
- The user MUST click "Allow" in the browser window

### Quota Project

All API calls require a quota project header:

```bash
-H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

## Core Concepts

### Cell Addressing

Google Sheets uses A1 notation for cell references:
- Single cell: `A1`, `B5`, `Z100`
- Range: `A1:B10`, `C5:E20`
- Named sheet: `Sheet1!A1:B10`
- Entire column: `A:A`
- Entire row: `1:1`

### Grid Coordinates

API requests use zero-based grid coordinates:
- Row 1 = rowIndex 0
- Column A = columnIndex 0
- Column B = columnIndex 1, etc.

### Color Format

Colors use RGB values from 0.0 to 1.0:
```json
{
  "red": 1.0,
  "green": 0.0,
  "blue": 0.0
}
```

Common colors:
- Red: `{"red": 1, "green": 0, "blue": 0}`
- Green: `{"red": 0, "green": 1, "blue": 0}`
- Blue: `{"red": 0, "green": 0, "blue": 1}`
- Yellow: `{"red": 1, "green": 1, "blue": 0}`
- Orange: `{"red": 1, "green": 0.65, "blue": 0}`
- Purple: `{"red": 0.5, "green": 0, "blue": 0.5}`
- Light Gray: `{"red": 0.9, "green": 0.9, "blue": 0.9}`
- Dark Gray: `{"red": 0.4, "green": 0.4, "blue": 0.4}`

## API Reference

### Create a Spreadsheet

```bash
TOKEN=$(gcloud auth application-default print-access-token)
curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "properties": {
      "title": "My Spreadsheet"
    }
  }'
```

### Read a Spreadsheet

```bash
# Get spreadsheet metadata (sheets, properties)
curl -s "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Get cell values in a range
curl -s "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}/values/Sheet1!A1:D10" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"

# Get multiple ranges at once
curl -s "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}/values:batchGet?ranges=Sheet1!A1:B10&ranges=Sheet2!C5:D15" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Batch Update

The primary way to modify spreadsheets. Atomic - all requests succeed or all fail:

```bash
curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}:batchUpdate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{"requests": [...]}'
```

**Key benefit**: One batchUpdate counts as one request in quota, even with 1000+ operations inside.

## Cell Operations

### Update Individual Cells

```json
{
  "updateCells": {
    "range": {
      "sheetId": 0,
      "startRowIndex": 0,
      "endRowIndex": 1,
      "startColumnIndex": 0,
      "endColumnIndex": 1
    },
    "rows": [
      {
        "values": [
          {
            "userEnteredValue": {"stringValue": "Hello World"}
          }
        ]
      }
    ],
    "fields": "userEnteredValue"
  }
}
```

### Update Cell Range (A1 notation)

```bash
# Simple value update
curl -s -X PUT "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}/values/Sheet1!A1:B2?valueInputOption=USER_ENTERED" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "range": "Sheet1!A1:B2",
    "values": [
      ["Name", "Score"],
      ["Alice", "95"]
    ]
  }'
```

Value input options:
- `RAW` - Values stored as-is
- `USER_ENTERED` - Parse as if user typed it (converts "=SUM(A1:A10)" to formula, "3/1/2024" to date)

### Update Multiple Ranges

```bash
curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}/values:batchUpdate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "valueInputOption": "USER_ENTERED",
    "data": [
      {
        "range": "Sheet1!A1:B1",
        "values": [["Name", "Score"]]
      },
      {
        "range": "Sheet1!A2:B4",
        "values": [
          ["Alice", "95"],
          ["Bob", "87"],
          ["Charlie", "92"]
        ]
      }
    ]
  }'
```

### Append Rows

```bash
curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}/values/Sheet1!A:B:append?valueInputOption=USER_ENTERED" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "values": [
      ["David", "88"],
      ["Eve", "94"]
    ]
  }'
```

### Clear Range

```bash
curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}/values/Sheet1!A1:Z100:clear" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json"
```

## Find and Replace

### Find and Replace Text

```json
{
  "findReplace": {
    "find": "old text",
    "replacement": "new text",
    "matchCase": true,
    "matchEntireCell": false,
    "searchByRegex": false,
    "allSheets": true
  }
}
```

### Find and Replace in Specific Range

```json
{
  "findReplace": {
    "find": "TODO",
    "replacement": "DONE",
    "range": {
      "sheetId": 0,
      "startRowIndex": 0,
      "endRowIndex": 100,
      "startColumnIndex": 0,
      "endColumnIndex": 10
    }
  }
}
```

### Find and Replace with Regex

```json
{
  "findReplace": {
    "find": "[0-9]{3}-[0-9]{3}-[0-9]{4}",
    "replacement": "XXX-XXX-XXXX",
    "searchByRegex": true,
    "allSheets": false,
    "sheetId": 0
  }
}
```

## Formatting and Colors

### Format Cells (Bold, Colors, Fonts)

```json
{
  "repeatCell": {
    "range": {
      "sheetId": 0,
      "startRowIndex": 0,
      "endRowIndex": 1,
      "startColumnIndex": 0,
      "endColumnIndex": 5
    },
    "cell": {
      "userEnteredFormat": {
        "backgroundColor": {"red": 0.2, "green": 0.5, "blue": 0.8},
        "textFormat": {
          "foregroundColor": {"red": 1, "green": 1, "blue": 1},
          "fontSize": 12,
          "bold": true,
          "italic": false
        },
        "horizontalAlignment": "CENTER",
        "verticalAlignment": "MIDDLE"
      }
    },
    "fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)"
  }
}
```

### Format Header Row

```json
{
  "repeatCell": {
    "range": {
      "sheetId": 0,
      "startRowIndex": 0,
      "endRowIndex": 1
    },
    "cell": {
      "userEnteredFormat": {
        "backgroundColor": {"red": 0.9, "green": 0.9, "blue": 0.9},
        "textFormat": {
          "bold": true,
          "fontSize": 11
        },
        "horizontalAlignment": "CENTER"
      }
    },
    "fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)"
  }
}
```

### Number Formatting

```json
{
  "repeatCell": {
    "range": {
      "sheetId": 0,
      "startRowIndex": 1,
      "startColumnIndex": 2,
      "endColumnIndex": 3
    },
    "cell": {
      "userEnteredFormat": {
        "numberFormat": {
          "type": "CURRENCY",
          "pattern": "$#,##0.00"
        }
      }
    },
    "fields": "userEnteredFormat.numberFormat"
  }
}
```

Number format types:
- `NUMBER` - `#,##0.00`
- `CURRENCY` - `$#,##0.00`
- `PERCENT` - `0.00%`
- `DATE` - `M/d/yyyy`
- `TIME` - `h:mm:ss am/pm`
- `DATE_TIME` - `M/d/yyyy h:mm:ss`
- `SCIENTIFIC` - `0.00E+00`

### Cell Borders

```json
{
  "updateBorders": {
    "range": {
      "sheetId": 0,
      "startRowIndex": 0,
      "endRowIndex": 10,
      "startColumnIndex": 0,
      "endColumnIndex": 5
    },
    "top": {
      "style": "SOLID",
      "width": 1,
      "color": {"red": 0, "green": 0, "blue": 0}
    },
    "bottom": {
      "style": "SOLID",
      "width": 1,
      "color": {"red": 0, "green": 0, "blue": 0}
    },
    "left": {
      "style": "SOLID",
      "width": 1,
      "color": {"red": 0, "green": 0, "blue": 0}
    },
    "right": {
      "style": "SOLID",
      "width": 1,
      "color": {"red": 0, "green": 0, "blue": 0}
    }
  }
}
```

Border styles: `SOLID`, `DOTTED`, `DASHED`, `SOLID_MEDIUM`, `SOLID_THICK`, `DOUBLE`

### Conditional Formatting

```json
{
  "addConditionalFormatRule": {
    "rule": {
      "ranges": [
        {
          "sheetId": 0,
          "startRowIndex": 1,
          "startColumnIndex": 2,
          "endColumnIndex": 3
        }
      ],
      "booleanRule": {
        "condition": {
          "type": "NUMBER_GREATER",
          "values": [{"userEnteredValue": "90"}]
        },
        "format": {
          "backgroundColor": {"red": 0.6, "green": 0.9, "blue": 0.6}
        }
      }
    },
    "index": 0
  }
}
```

Condition types:
- `NUMBER_GREATER`, `NUMBER_GREATER_THAN_EQ`
- `NUMBER_LESS`, `NUMBER_LESS_THAN_EQ`
- `NUMBER_EQ`, `NUMBER_NOT_EQ`
- `NUMBER_BETWEEN`, `NUMBER_NOT_BETWEEN`
- `TEXT_CONTAINS`, `TEXT_NOT_CONTAINS`
- `TEXT_STARTS_WITH`, `TEXT_ENDS_WITH`
- `TEXT_EQ`, `TEXT_IS_EMAIL`, `TEXT_IS_URL`
- `DATE_EQ`, `DATE_BEFORE`, `DATE_AFTER`
- `BLANK`, `NOT_BLANK`
- `CUSTOM_FORMULA` - For complex conditions

### Gradient Conditional Formatting

```json
{
  "addConditionalFormatRule": {
    "rule": {
      "ranges": [
        {
          "sheetId": 0,
          "startRowIndex": 1,
          "endRowIndex": 100,
          "startColumnIndex": 2,
          "endColumnIndex": 3
        }
      ],
      "gradientRule": {
        "minpoint": {
          "color": {"red": 1, "green": 0, "blue": 0},
          "type": "MIN"
        },
        "midpoint": {
          "color": {"red": 1, "green": 1, "blue": 0},
          "type": "PERCENTILE",
          "value": "50"
        },
        "maxpoint": {
          "color": {"red": 0, "green": 1, "blue": 0},
          "type": "MAX"
        }
      }
    }
  }
}
```

## Rich Tables

### Create Formatted Table

```json
{
  "requests": [
    {
      "updateCells": {
        "range": {
          "sheetId": 0,
          "startRowIndex": 0,
          "endRowIndex": 4,
          "startColumnIndex": 0,
          "endColumnIndex": 3
        },
        "rows": [
          {"values": [
            {"userEnteredValue": {"stringValue": "Name"}},
            {"userEnteredValue": {"stringValue": "Department"}},
            {"userEnteredValue": {"stringValue": "Salary"}}
          ]},
          {"values": [
            {"userEnteredValue": {"stringValue": "Alice"}},
            {"userEnteredValue": {"stringValue": "Engineering"}},
            {"userEnteredValue": {"numberValue": 95000}}
          ]},
          {"values": [
            {"userEnteredValue": {"stringValue": "Bob"}},
            {"userEnteredValue": {"stringValue": "Sales"}},
            {"userEnteredValue": {"numberValue": 87000}}
          ]},
          {"values": [
            {"userEnteredValue": {"stringValue": "Charlie"}},
            {"userEnteredValue": {"stringValue": "Marketing"}},
            {"userEnteredValue": {"numberValue": 92000}}
          ]}
        ],
        "fields": "userEnteredValue"
      }
    },
    {
      "repeatCell": {
        "range": {
          "sheetId": 0,
          "startRowIndex": 0,
          "endRowIndex": 1,
          "startColumnIndex": 0,
          "endColumnIndex": 3
        },
        "cell": {
          "userEnteredFormat": {
            "backgroundColor": {"red": 0.2, "green": 0.4, "blue": 0.8},
            "textFormat": {
              "foregroundColor": {"red": 1, "green": 1, "blue": 1},
              "bold": true
            },
            "horizontalAlignment": "CENTER"
          }
        },
        "fields": "userEnteredFormat"
      }
    },
    {
      "repeatCell": {
        "range": {
          "sheetId": 0,
          "startRowIndex": 1,
          "startColumnIndex": 2,
          "endColumnIndex": 3
        },
        "cell": {
          "userEnteredFormat": {
            "numberFormat": {
              "type": "CURRENCY",
              "pattern": "$#,##0"
            }
          }
        },
        "fields": "userEnteredFormat.numberFormat"
      }
    },
    {
      "updateBorders": {
        "range": {
          "sheetId": 0,
          "startRowIndex": 0,
          "endRowIndex": 4,
          "startColumnIndex": 0,
          "endColumnIndex": 3
        },
        "innerHorizontal": {"style": "SOLID", "width": 1, "color": {"red": 0.8, "green": 0.8, "blue": 0.8}},
        "innerVertical": {"style": "SOLID", "width": 1, "color": {"red": 0.8, "green": 0.8, "blue": 0.8}},
        "top": {"style": "SOLID_MEDIUM", "width": 2, "color": {"red": 0, "green": 0, "blue": 0}},
        "bottom": {"style": "SOLID_MEDIUM", "width": 2, "color": {"red": 0, "green": 0, "blue": 0}},
        "left": {"style": "SOLID_MEDIUM", "width": 2, "color": {"red": 0, "green": 0, "blue": 0}},
        "right": {"style": "SOLID_MEDIUM", "width": 2, "color": {"red": 0, "green": 0, "blue": 0}}
      }
    }
  ]
}
```

### Freeze Rows/Columns

```json
{
  "updateSheetProperties": {
    "properties": {
      "sheetId": 0,
      "gridProperties": {
        "frozenRowCount": 1,
        "frozenColumnCount": 0
      }
    },
    "fields": "gridProperties.frozenRowCount,gridProperties.frozenColumnCount"
  }
}
```

### Auto-Resize Columns

```json
{
  "autoResizeDimensions": {
    "dimensions": {
      "sheetId": 0,
      "dimension": "COLUMNS",
      "startIndex": 0,
      "endIndex": 10
    }
  }
}
```

### Merge Cells

```json
{
  "mergeCells": {
    "range": {
      "sheetId": 0,
      "startRowIndex": 0,
      "endRowIndex": 1,
      "startColumnIndex": 0,
      "endColumnIndex": 3
    },
    "mergeType": "MERGE_ALL"
  }
}
```

Merge types: `MERGE_ALL`, `MERGE_COLUMNS`, `MERGE_ROWS`

## Multiple Sheets

### Add a New Sheet

```json
{
  "addSheet": {
    "properties": {
      "title": "Q2 Data",
      "gridProperties": {
        "rowCount": 1000,
        "columnCount": 26
      },
      "tabColor": {"red": 0, "green": 0.5, "blue": 1}
    }
  }
}
```

### Delete a Sheet

```json
{
  "deleteSheet": {
    "sheetId": 123456
  }
}
```

### Rename a Sheet

```json
{
  "updateSheetProperties": {
    "properties": {
      "sheetId": 0,
      "title": "Sales 2025"
    },
    "fields": "title"
  }
}
```

### Get Sheet IDs

```bash
curl -s "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}?fields=sheets.properties" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  | jq '.sheets[] | {sheetId: .properties.sheetId, title: .properties.title}'
```

### Copy Sheet to Another Spreadsheet

```bash
curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets/${SOURCE_SHEET_ID}/sheets/${SHEET_ID}:copyTo" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "destinationSpreadsheetId": "'${DEST_SHEET_ID}'"
  }'
```

### Duplicate Sheet Within Spreadsheet

```json
{
  "duplicateSheet": {
    "sourceSheetId": 0,
    "insertSheetIndex": 1,
    "newSheetName": "Copy of Sheet1"
  }
}
```

## Formulas

### Insert Formula

```bash
# Using A1 notation (simpler)
curl -s -X PUT "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}/values/Sheet1!D2?valueInputOption=USER_ENTERED" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "values": [["=SUM(A2:C2)"]]
  }'
```

### Common Formulas

```
SUM: =SUM(A1:A10)
AVERAGE: =AVERAGE(B1:B10)
COUNT: =COUNT(C1:C10)
IF: =IF(A1>100, "High", "Low")
VLOOKUP: =VLOOKUP(A1, Sheet2!A:B, 2, FALSE)
CONCATENATE: =CONCATENATE(A1, " ", B1)
TODAY: =TODAY()
NOW: =NOW()
```

### Cross-Sheet References

```
=Sheet2!A1
=SUM(Sheet2!A1:A10)
='Q1 Data'!B5
=VLOOKUP(A1, 'Reference Data'!A:B, 2, FALSE)
```

### Array Formulas

```bash
# Apply formula to entire column
curl -s -X PUT "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}/values/Sheet1!D2:D100?valueInputOption=USER_ENTERED" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "values": [["=ARRAYFORMULA(IF(ROW(A2:A100)=1,\"\",A2:A100*B2:B100))"]]
  }'
```

### Named Ranges

Create named range:
```json
{
  "addNamedRange": {
    "namedRange": {
      "name": "SalesData",
      "range": {
        "sheetId": 0,
        "startRowIndex": 0,
        "endRowIndex": 100,
        "startColumnIndex": 0,
        "endColumnIndex": 5
      }
    }
  }
}
```

Use in formula: `=SUM(SalesData)`

## Charts

### Create Column Chart

```json
{
  "addChart": {
    "chart": {
      "spec": {
        "title": "Sales by Month",
        "basicChart": {
          "chartType": "COLUMN",
          "legendPosition": "BOTTOM_LEGEND",
          "axis": [
            {
              "position": "BOTTOM_AXIS",
              "title": "Month"
            },
            {
              "position": "LEFT_AXIS",
              "title": "Sales ($)"
            }
          ],
          "domains": [
            {
              "domain": {
                "sourceRange": {
                  "sources": [
                    {
                      "sheetId": 0,
                      "startRowIndex": 0,
                      "endRowIndex": 13,
                      "startColumnIndex": 0,
                      "endColumnIndex": 1
                    }
                  ]
                }
              }
            }
          ],
          "series": [
            {
              "series": {
                "sourceRange": {
                  "sources": [
                    {
                      "sheetId": 0,
                      "startRowIndex": 0,
                      "endRowIndex": 13,
                      "startColumnIndex": 1,
                      "endColumnIndex": 2
                    }
                  ]
                }
              },
              "targetAxis": "LEFT_AXIS"
            }
          ],
          "headerCount": 1
        }
      },
      "position": {
        "overlayPosition": {
          "anchorCell": {
            "sheetId": 0,
            "rowIndex": 0,
            "columnIndex": 4
          },
          "offsetXPixels": 10,
          "offsetYPixels": 10
        }
      }
    }
  }
}
```

### Create Pie Chart

```json
{
  "addChart": {
    "chart": {
      "spec": {
        "title": "Market Share",
        "pieChart": {
          "legendPosition": "RIGHT_LEGEND",
          "domain": {
            "sourceRange": {
              "sources": [
                {
                  "sheetId": 0,
                  "startRowIndex": 1,
                  "endRowIndex": 6,
                  "startColumnIndex": 0,
                  "endColumnIndex": 1
                }
              ]
            }
          },
          "series": {
            "sourceRange": {
              "sources": [
                {
                  "sheetId": 0,
                  "startRowIndex": 1,
                  "endRowIndex": 6,
                  "startColumnIndex": 1,
                  "endColumnIndex": 2
                }
              ]
            }
          },
          "threeDimensional": false
        }
      },
      "position": {
        "overlayPosition": {
          "anchorCell": {
            "sheetId": 0,
            "rowIndex": 8,
            "columnIndex": 0
          }
        }
      }
    }
  }
}
```

### Create Line Chart

```json
{
  "addChart": {
    "chart": {
      "spec": {
        "title": "Trend Over Time",
        "basicChart": {
          "chartType": "LINE",
          "legendPosition": "BOTTOM_LEGEND",
          "axis": [
            {
              "position": "BOTTOM_AXIS",
              "title": "Date"
            },
            {
              "position": "LEFT_AXIS",
              "title": "Value"
            }
          ],
          "domains": [
            {
              "domain": {
                "sourceRange": {
                  "sources": [
                    {
                      "sheetId": 0,
                      "startRowIndex": 0,
                      "endRowIndex": 30,
                      "startColumnIndex": 0,
                      "endColumnIndex": 1
                    }
                  ]
                }
              }
            }
          ],
          "series": [
            {
              "series": {
                "sourceRange": {
                  "sources": [
                    {
                      "sheetId": 0,
                      "startRowIndex": 0,
                      "endRowIndex": 30,
                      "startColumnIndex": 1,
                      "endColumnIndex": 2
                    }
                  ]
                }
              },
              "targetAxis": "LEFT_AXIS"
            }
          ],
          "headerCount": 1
        }
      },
      "position": {
        "newSheet": true
      }
    }
  }
}
```

Chart position options:
- `overlayPosition` - Float over cells
- `newSheet` - Create chart in its own sheet

Chart types:
- `COLUMN`, `BAR`, `LINE`, `AREA`, `SCATTER`, `COMBO`
- `PIE` (use `pieChart` instead of `basicChart`)
- `HISTOGRAM`, `CANDLESTICK`, `WATERFALL`

### Update Chart

```json
{
  "updateChartSpec": {
    "chartId": 12345,
    "spec": {
      "title": "Updated Title"
    }
  }
}
```

### Delete Chart

```json
{
  "deleteEmbeddedObject": {
    "objectId": 12345
  }
}
```

## Comments

### Add a Comment

Comments use the Drive API and can mention users with @email syntax:

```bash
TOKEN=$(gcloud auth application-default print-access-token)

curl -s -X POST "https://www.googleapis.com/drive/v3/files/${SHEET_ID}/comments?fields=*" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "@user@example.com Please review this cell",
    "anchor": "{\"type\":\"cell\",\"r\":5,\"c\":2}"
  }'
```

The anchor format for Sheets:
- `r` - Row number (0-based, so row 1 = 0, row 2 = 1, etc.)
- `c` - Column number (0-based, so column A = 0, column B = 1, etc.)

For a range comment:
```json
{
  "anchor": "{\"type\":\"cell\",\"r\":5,\"c\":2,\"r2\":10,\"c2\":4}"
}
```

### List Comments

```bash
curl -s "https://www.googleapis.com/drive/v3/files/${SHEET_ID}/comments?fields=*" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

### Reply to Comment

```bash
curl -s -X POST "https://www.googleapis.com/drive/v3/files/${SHEET_ID}/comments/${COMMENT_ID}/replies?fields=*" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Thanks for the feedback!"
  }'
```

### Delete Comment

```bash
curl -s -X DELETE "https://www.googleapis.com/drive/v3/files/${SHEET_ID}/comments/${COMMENT_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

## Helper Scripts

See `/resources/` directory for Python helper scripts:

### gsheets_builder.py - Build spreadsheets with proper data and formatting

```bash
# Create a new spreadsheet
python3 gsheets_builder.py create --title "My Spreadsheet"

# Read spreadsheet info
python3 gsheets_builder.py info --sheet-id "SHEET_ID"

# Update cell values
python3 gsheets_builder.py update-cells --sheet-id "SHEET_ID" \
  --range "Sheet1!A1:B2" \
  --values '[["Name","Score"],["Alice","95"]]'

# Append rows
python3 gsheets_builder.py append-rows --sheet-id "SHEET_ID" \
  --range "Sheet1!A:B" \
  --values '[["Bob","87"],["Charlie","92"]]'

# Format header row
python3 gsheets_builder.py format-header --sheet-id "SHEET_ID" \
  --sheet-index 0 --columns 5

# Add colored table with borders
python3 gsheets_builder.py add-table --sheet-id "SHEET_ID" \
  --sheet-index 0 --start-row 0 --start-col 0 \
  --data '[["Name","Dept","Salary"],["Alice","Eng",95000]]' \
  --header-color '{"red":0.2,"green":0.4,"blue":0.8}'

# Find and replace
python3 gsheets_builder.py find-replace --sheet-id "SHEET_ID" \
  --find "TODO" --replace "DONE"

# Add a new sheet
python3 gsheets_builder.py add-sheet --sheet-id "SHEET_ID" \
  --title "Q2 Data" --tab-color '{"red":0,"green":0.5,"blue":1}'

# Add formula
python3 gsheets_builder.py add-formula --sheet-id "SHEET_ID" \
  --range "Sheet1!D2" --formula "=SUM(A2:C2)"

# Create column chart
python3 gsheets_builder.py add-chart --sheet-id "SHEET_ID" \
  --sheet-index 0 --chart-type "COLUMN" \
  --title "Sales by Month" \
  --domain-range "A1:A13" --series-range "B1:B13" \
  --position-col 4

# Add conditional formatting
python3 gsheets_builder.py add-conditional-format --sheet-id "SHEET_ID" \
  --sheet-index 0 --start-row 1 --start-col 2 --end-col 3 \
  --condition "NUMBER_GREATER" --value "90" \
  --bg-color '{"red":0.6,"green":0.9,"blue":0.6}'
```

### gsheets_auth.py - Authentication management

```bash
python3 gsheets_auth.py status    # Check auth status
python3 gsheets_auth.py login     # Login with required scopes
python3 gsheets_auth.py token     # Get access token
python3 gsheets_auth.py validate  # Validate current token
```

## Best Practices

1. **Use batchUpdate for multiple operations** - Atomic and counts as one API request
2. **Read before complex updates** - Get current state to avoid conflicts
3. **Use A1 notation for simple updates** - Easier to read and write
4. **Use grid coordinates for complex formatting** - More precise control
5. **Freeze header rows** - Better UX for large tables
6. **Auto-resize columns** - Ensures content is visible
7. **Use named ranges** - Makes formulas more readable
8. **Test formulas manually first** - Verify they work before automating
9. **Limit chart complexity** - Some chart types not fully supported by API
10. **Validate colors** - Use 0.0-1.0 range, not 0-255

## Example: Create Complete Spreadsheet with Table and Chart

```bash
#!/bin/bash
TOKEN=$(gcloud auth application-default print-access-token)
QUOTA_PROJECT="${GCP_QUOTA_PROJECT:-your-gcp-project-id}"

# 1. Create spreadsheet
SHEET_ID=$(curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: $QUOTA_PROJECT" \
  -H "Content-Type: application/json" \
  -d '{"properties": {"title": "Sales Report 2025"}}' | jq -r '.spreadsheetId')

echo "Created spreadsheet: $SHEET_ID"

# 2. Add data
curl -s -X PUT "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}/values/Sheet1!A1:C13?valueInputOption=USER_ENTERED" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: $QUOTA_PROJECT" \
  -H "Content-Type: application/json" \
  -d '{
    "values": [
      ["Month", "Sales", "Expenses"],
      ["Jan", "50000", "30000"],
      ["Feb", "55000", "32000"],
      ["Mar", "60000", "35000"],
      ["Apr", "58000", "33000"],
      ["May", "62000", "36000"],
      ["Jun", "65000", "38000"],
      ["Jul", "70000", "40000"],
      ["Aug", "68000", "39000"],
      ["Sep", "72000", "41000"],
      ["Oct", "75000", "43000"],
      ["Nov", "80000", "45000"],
      ["Dec", "85000", "47000"]
    ]
  }'

# 3. Format table with chart
curl -s -X POST "https://sheets.googleapis.com/v4/spreadsheets/${SHEET_ID}:batchUpdate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: $QUOTA_PROJECT" \
  -H "Content-Type: application/json" \
  -d '{
    "requests": [
      {
        "repeatCell": {
          "range": {"sheetId": 0, "startRowIndex": 0, "endRowIndex": 1},
          "cell": {
            "userEnteredFormat": {
              "backgroundColor": {"red": 0.2, "green": 0.4, "blue": 0.8},
              "textFormat": {"foregroundColor": {"red": 1, "green": 1, "blue": 1}, "bold": true},
              "horizontalAlignment": "CENTER"
            }
          },
          "fields": "userEnteredFormat"
        }
      },
      {
        "repeatCell": {
          "range": {"sheetId": 0, "startRowIndex": 1, "startColumnIndex": 1, "endColumnIndex": 3},
          "cell": {
            "userEnteredFormat": {
              "numberFormat": {"type": "CURRENCY", "pattern": "$#,##0"}
            }
          },
          "fields": "userEnteredFormat.numberFormat"
        }
      },
      {
        "updateBorders": {
          "range": {"sheetId": 0, "startRowIndex": 0, "endRowIndex": 13, "startColumnIndex": 0, "endColumnIndex": 3},
          "innerHorizontal": {"style": "SOLID", "width": 1, "color": {"red": 0.8, "green": 0.8, "blue": 0.8}},
          "innerVertical": {"style": "SOLID", "width": 1, "color": {"red": 0.8, "green": 0.8, "blue": 0.8}}
        }
      },
      {
        "updateSheetProperties": {
          "properties": {"sheetId": 0, "gridProperties": {"frozenRowCount": 1}},
          "fields": "gridProperties.frozenRowCount"
        }
      },
      {
        "addChart": {
          "chart": {
            "spec": {
              "title": "Monthly Sales vs Expenses",
              "basicChart": {
                "chartType": "COLUMN",
                "legendPosition": "BOTTOM_LEGEND",
                "domains": [{"domain": {"sourceRange": {"sources": [{"sheetId": 0, "startRowIndex": 0, "endRowIndex": 13, "startColumnIndex": 0, "endColumnIndex": 1}]}}}],
                "series": [
                  {"series": {"sourceRange": {"sources": [{"sheetId": 0, "startRowIndex": 0, "endRowIndex": 13, "startColumnIndex": 1, "endColumnIndex": 2}]}}, "targetAxis": "LEFT_AXIS"},
                  {"series": {"sourceRange": {"sources": [{"sheetId": 0, "startRowIndex": 0, "endRowIndex": 13, "startColumnIndex": 2, "endColumnIndex": 3}]}}, "targetAxis": "LEFT_AXIS"}
                ],
                "headerCount": 1
              }
            },
            "position": {"overlayPosition": {"anchorCell": {"sheetId": 0, "rowIndex": 0, "columnIndex": 4}}}
          }
        }
      }
    ]
  }'

echo "Spreadsheet URL: https://docs.google.com/spreadsheets/d/${SHEET_ID}/edit"
```

## Sources

- [Method: spreadsheets.batchUpdate | Google Sheets](https://developers.google.com/workspace/sheets/api/reference/rest/v4/spreadsheets/batchUpdate)
- [Update spreadsheets | Google Sheets](https://developers.google.com/workspace/sheets/api/guides/batchupdate)
- [Conditional formatting | Google Sheets](https://developers.google.com/workspace/sheets/api/samples/conditional-formatting)
- [Basic formatting | Google Sheets](https://developers.google.com/sheets/api/samples/formatting)
- [Charts | Google Sheets](https://developers.google.com/workspace/sheets/api/samples/charts)
- [Data operations | Google Sheets](https://developers.google.com/sheets/api/samples/data)
