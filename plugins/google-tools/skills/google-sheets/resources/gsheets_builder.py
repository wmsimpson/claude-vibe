#!/usr/bin/env python3
"""
Google Sheets Builder - Build spreadsheets with proper data and formatting

This script helps create well-formatted Google Sheets by:
1. Creating spreadsheets and managing sheets
2. Updating cell values and formulas
3. Applying formatting, colors, and borders
4. Creating charts and conditional formatting
5. Find and replace operations

Usage:
    # Create a new spreadsheet
    python3 gsheets_builder.py create --title "My Spreadsheet"

    # Update cells
    python3 gsheets_builder.py update-cells --sheet-id "SHEET_ID" --range "A1:B2" --values '[["A","B"],["C","D"]]'

    # Format table
    python3 gsheets_builder.py add-table --sheet-id "SHEET_ID" --data '[["Name","Score"],["Alice","95"]]'
"""

import argparse
import json
import sys
from typing import Dict, List, Optional, Any

from google_api_utils import api_call_with_retry


def api_call(method: str, url: str, data: Optional[Dict] = None) -> Dict:
    """Make an API call using curl with retry logic."""
    return api_call_with_retry(method, url, data=data)


def create_spreadsheet(title: str, sheets: Optional[List[str]] = None) -> str:
    """Create a new Google Sheet and return its ID."""
    data = {"properties": {"title": title}}

    if sheets:
        data["sheets"] = [{"properties": {"title": sheet_title}} for sheet_title in sheets]

    response = api_call("POST", "https://sheets.googleapis.com/v4/spreadsheets", data)

    if "error" in response:
        raise RuntimeError(f"Failed to create spreadsheet: {response['error']['message']}")

    return response["spreadsheetId"]


def get_spreadsheet(sheet_id: str) -> Dict:
    """Get full spreadsheet structure."""
    return api_call("GET", f"https://sheets.googleapis.com/v4/spreadsheets/{sheet_id}")


def get_sheet_info(sheet_id: str) -> List[Dict]:
    """Get list of sheets with their IDs and properties."""
    response = get_spreadsheet(sheet_id)
    return [
        {
            "sheetId": sheet["properties"]["sheetId"],
            "title": sheet["properties"]["title"],
            "index": sheet["properties"]["index"],
            "rowCount": sheet["properties"]["gridProperties"]["rowCount"],
            "columnCount": sheet["properties"]["gridProperties"]["columnCount"]
        }
        for sheet in response.get("sheets", [])
    ]


def batch_update(sheet_id: str, requests: List[Dict]) -> Dict:
    """Execute a batchUpdate on the spreadsheet."""
    return api_call(
        "POST",
        f"https://sheets.googleapis.com/v4/spreadsheets/{sheet_id}:batchUpdate",
        {"requests": requests}
    )


def get_values(sheet_id: str, range_name: str) -> List[List[Any]]:
    """Get values from a range."""
    response = api_call("GET", f"https://sheets.googleapis.com/v4/spreadsheets/{sheet_id}/values/{range_name}")
    return response.get("values", [])


def update_values(sheet_id: str, range_name: str, values: List[List[Any]], value_input_option: str = "USER_ENTERED") -> Dict:
    """Update values in a range."""
    return api_call(
        "PUT",
        f"https://sheets.googleapis.com/v4/spreadsheets/{sheet_id}/values/{range_name}?valueInputOption={value_input_option}",
        {"range": range_name, "values": values}
    )


def append_values(sheet_id: str, range_name: str, values: List[List[Any]], value_input_option: str = "USER_ENTERED") -> Dict:
    """Append values to a range."""
    return api_call(
        "POST",
        f"https://sheets.googleapis.com/v4/spreadsheets/{sheet_id}/values/{range_name}:append?valueInputOption={value_input_option}",
        {"values": values}
    )


def clear_values(sheet_id: str, range_name: str) -> Dict:
    """Clear values in a range."""
    return api_call(
        "POST",
        f"https://sheets.googleapis.com/v4/spreadsheets/{sheet_id}/values/{range_name}:clear",
        {}
    )


def format_header_row(
    sheet_id: str,
    sheet_index: int,
    columns: int,
    bg_color: Optional[Dict] = None,
    text_color: Optional[Dict] = None
) -> Dict:
    """Format the first row as a header."""
    if bg_color is None:
        bg_color = {"red": 0.9, "green": 0.9, "blue": 0.9}
    if text_color is None:
        text_color = {"red": 0, "green": 0, "blue": 0}

    request = {
        "repeatCell": {
            "range": {
                "sheetId": sheet_index,
                "startRowIndex": 0,
                "endRowIndex": 1,
                "startColumnIndex": 0,
                "endColumnIndex": columns
            },
            "cell": {
                "userEnteredFormat": {
                    "backgroundColor": bg_color,
                    "textFormat": {
                        "foregroundColor": text_color,
                        "bold": True,
                        "fontSize": 11
                    },
                    "horizontalAlignment": "CENTER",
                    "verticalAlignment": "MIDDLE"
                }
            },
            "fields": "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)"
        }
    }

    return batch_update(sheet_id, [request])


def add_borders(
    sheet_id: str,
    sheet_index: int,
    start_row: int,
    end_row: int,
    start_col: int,
    end_col: int,
    style: str = "SOLID"
) -> Dict:
    """Add borders to a range."""
    border_style = {
        "style": style,
        "width": 1,
        "color": {"red": 0, "green": 0, "blue": 0}
    }

    request = {
        "updateBorders": {
            "range": {
                "sheetId": sheet_index,
                "startRowIndex": start_row,
                "endRowIndex": end_row,
                "startColumnIndex": start_col,
                "endColumnIndex": end_col
            },
            "top": border_style,
            "bottom": border_style,
            "left": border_style,
            "right": border_style,
            "innerHorizontal": border_style,
            "innerVertical": border_style
        }
    }

    return batch_update(sheet_id, [request])


def format_range(
    sheet_id: str,
    sheet_index: int,
    start_row: int,
    end_row: int,
    start_col: int,
    end_col: int,
    bg_color: Optional[Dict] = None,
    text_color: Optional[Dict] = None,
    bold: bool = False,
    italic: bool = False,
    font_size: Optional[int] = None,
    h_align: Optional[str] = None,
    v_align: Optional[str] = None,
    number_format: Optional[Dict] = None
) -> Dict:
    """Format a range of cells."""
    cell_format = {}
    fields = []

    if bg_color:
        cell_format["backgroundColor"] = bg_color
        fields.append("backgroundColor")

    text_format = {}
    if text_color:
        text_format["foregroundColor"] = text_color
    if bold:
        text_format["bold"] = True
    if italic:
        text_format["italic"] = True
    if font_size:
        text_format["fontSize"] = font_size

    if text_format:
        cell_format["textFormat"] = text_format
        fields.append("textFormat")

    if h_align:
        cell_format["horizontalAlignment"] = h_align
        fields.append("horizontalAlignment")

    if v_align:
        cell_format["verticalAlignment"] = v_align
        fields.append("verticalAlignment")

    if number_format:
        cell_format["numberFormat"] = number_format
        fields.append("numberFormat")

    request = {
        "repeatCell": {
            "range": {
                "sheetId": sheet_index,
                "startRowIndex": start_row,
                "endRowIndex": end_row,
                "startColumnIndex": start_col,
                "endColumnIndex": end_col
            },
            "cell": {
                "userEnteredFormat": cell_format
            },
            "fields": "userEnteredFormat(" + ",".join(fields) + ")"
        }
    }

    return batch_update(sheet_id, [request])


def add_table(
    sheet_id: str,
    sheet_index: int,
    start_row: int,
    start_col: int,
    data: List[List[Any]],
    header_color: Optional[Dict] = None,
    with_borders: bool = True
) -> Dict:
    """Add a formatted table with data."""
    if not data:
        raise ValueError("Data cannot be empty")

    rows = len(data)
    cols = len(data[0])

    # Convert to range notation
    start_cell = f"{chr(65 + start_col)}{start_row + 1}"
    end_cell = f"{chr(65 + start_col + cols - 1)}{start_row + rows}"
    range_name = f"Sheet1!{start_cell}:{end_cell}"

    # Update values
    update_values(sheet_id, range_name, data)

    # Format header
    if header_color:
        format_header_row(sheet_id, sheet_index, cols, bg_color=header_color)

    # Add borders
    if with_borders:
        add_borders(
            sheet_id, sheet_index,
            start_row, start_row + rows,
            start_col, start_col + cols
        )

    return {"status": "success"}


def find_replace(
    sheet_id: str,
    find: str,
    replace: str,
    sheet_index: Optional[int] = None,
    match_case: bool = False,
    match_entire_cell: bool = False,
    use_regex: bool = False
) -> Dict:
    """Find and replace text in spreadsheet."""
    request = {
        "findReplace": {
            "find": find,
            "replacement": replace,
            "matchCase": match_case,
            "matchEntireCell": match_entire_cell,
            "searchByRegex": use_regex
        }
    }

    if sheet_index is not None:
        request["findReplace"]["sheetId"] = sheet_index
    else:
        request["findReplace"]["allSheets"] = True

    return batch_update(sheet_id, [request])


def add_sheet(
    sheet_id: str,
    title: str,
    row_count: int = 1000,
    col_count: int = 26,
    tab_color: Optional[Dict] = None
) -> Dict:
    """Add a new sheet to the spreadsheet."""
    properties = {
        "title": title,
        "gridProperties": {
            "rowCount": row_count,
            "columnCount": col_count
        }
    }

    if tab_color:
        properties["tabColor"] = tab_color

    request = {
        "addSheet": {
            "properties": properties
        }
    }

    return batch_update(sheet_id, [request])


def delete_sheet(sheet_id: str, sheet_index: int) -> Dict:
    """Delete a sheet from the spreadsheet."""
    request = {
        "deleteSheet": {
            "sheetId": sheet_index
        }
    }

    return batch_update(sheet_id, [request])


def rename_sheet(sheet_id: str, sheet_index: int, new_title: str) -> Dict:
    """Rename a sheet."""
    request = {
        "updateSheetProperties": {
            "properties": {
                "sheetId": sheet_index,
                "title": new_title
            },
            "fields": "title"
        }
    }

    return batch_update(sheet_id, [request])


def freeze_rows(sheet_id: str, sheet_index: int, row_count: int) -> Dict:
    """Freeze top rows."""
    request = {
        "updateSheetProperties": {
            "properties": {
                "sheetId": sheet_index,
                "gridProperties": {
                    "frozenRowCount": row_count
                }
            },
            "fields": "gridProperties.frozenRowCount"
        }
    }

    return batch_update(sheet_id, [request])


def freeze_columns(sheet_id: str, sheet_index: int, col_count: int) -> Dict:
    """Freeze left columns."""
    request = {
        "updateSheetProperties": {
            "properties": {
                "sheetId": sheet_index,
                "gridProperties": {
                    "frozenColumnCount": col_count
                }
            },
            "fields": "gridProperties.frozenColumnCount"
        }
    }

    return batch_update(sheet_id, [request])


def auto_resize_columns(
    sheet_id: str,
    sheet_index: int,
    start_col: int = 0,
    end_col: Optional[int] = None
) -> Dict:
    """Auto-resize columns to fit content."""
    request = {
        "autoResizeDimensions": {
            "dimensions": {
                "sheetId": sheet_index,
                "dimension": "COLUMNS",
                "startIndex": start_col
            }
        }
    }

    if end_col is not None:
        request["autoResizeDimensions"]["dimensions"]["endIndex"] = end_col

    return batch_update(sheet_id, [request])


def add_formula(sheet_id: str, range_name: str, formula: str) -> Dict:
    """Add a formula to a cell or range."""
    return update_values(sheet_id, range_name, [[formula]], "USER_ENTERED")


def add_conditional_format(
    sheet_id: str,
    sheet_index: int,
    start_row: int,
    end_row: int,
    start_col: int,
    end_col: int,
    condition_type: str,
    value: str,
    bg_color: Optional[Dict] = None,
    text_color: Optional[Dict] = None
) -> Dict:
    """Add conditional formatting rule."""
    format_spec = {}
    if bg_color:
        format_spec["backgroundColor"] = bg_color
    if text_color:
        format_spec["textFormat"] = {"foregroundColor": text_color}

    request = {
        "addConditionalFormatRule": {
            "rule": {
                "ranges": [
                    {
                        "sheetId": sheet_index,
                        "startRowIndex": start_row,
                        "endRowIndex": end_row,
                        "startColumnIndex": start_col,
                        "endColumnIndex": end_col
                    }
                ],
                "booleanRule": {
                    "condition": {
                        "type": condition_type,
                        "values": [{"userEnteredValue": value}]
                    },
                    "format": format_spec
                }
            },
            "index": 0
        }
    }

    return batch_update(sheet_id, [request])


def add_chart(
    sheet_id: str,
    sheet_index: int,
    chart_type: str,
    title: str,
    domain_range: str,
    series_ranges: List[str],
    position_row: int = 0,
    position_col: int = 4,
    legend_position: str = "BOTTOM_LEGEND"
) -> Dict:
    """Add a chart to the spreadsheet."""

    # Parse range notation (e.g., "A1:A13")
    def parse_range(range_str: str) -> Dict:
        """Parse A1 notation to grid range."""
        parts = range_str.split(":")
        start_cell = parts[0]
        end_cell = parts[1] if len(parts) > 1 else start_cell

        def cell_to_indices(cell: str):
            col = ord(cell[0]) - ord('A')
            row = int(cell[1:]) - 1
            return row, col

        start_row, start_col = cell_to_indices(start_cell)
        end_row, end_col = cell_to_indices(end_cell)

        return {
            "sheetId": sheet_index,
            "startRowIndex": start_row,
            "endRowIndex": end_row + 1,
            "startColumnIndex": start_col,
            "endColumnIndex": end_col + 1
        }

    # Build chart spec
    if chart_type.upper() == "PIE":
        chart_spec = {
            "pieChart": {
                "legendPosition": legend_position,
                "domain": {
                    "sourceRange": {
                        "sources": [parse_range(domain_range)]
                    }
                },
                "series": {
                    "sourceRange": {
                        "sources": [parse_range(series_ranges[0])]
                    }
                },
                "threeDimensional": False
            }
        }
    else:
        # Basic chart (COLUMN, BAR, LINE, etc.)
        series_list = []
        for series_range in series_ranges:
            series_list.append({
                "series": {
                    "sourceRange": {
                        "sources": [parse_range(series_range)]
                    }
                },
                "targetAxis": "LEFT_AXIS"
            })

        chart_spec = {
            "basicChart": {
                "chartType": chart_type.upper(),
                "legendPosition": legend_position,
                "domains": [
                    {
                        "domain": {
                            "sourceRange": {
                                "sources": [parse_range(domain_range)]
                            }
                        }
                    }
                ],
                "series": series_list,
                "headerCount": 1
            }
        }

    request = {
        "addChart": {
            "chart": {
                "spec": {
                    "title": title,
                    **chart_spec
                },
                "position": {
                    "overlayPosition": {
                        "anchorCell": {
                            "sheetId": sheet_index,
                            "rowIndex": position_row,
                            "columnIndex": position_col
                        },
                        "offsetXPixels": 10,
                        "offsetYPixels": 10
                    }
                }
            }
        }
    }

    return batch_update(sheet_id, [request])


def main():
    parser = argparse.ArgumentParser(
        description="Build and manage Google Sheets",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    subparsers = parser.add_subparsers(dest="command", required=True)

    # Create command
    create_parser = subparsers.add_parser("create", help="Create a new spreadsheet")
    create_parser.add_argument("--title", required=True, help="Spreadsheet title")
    create_parser.add_argument("--sheets", help="Comma-separated list of sheet titles")

    # Info command
    info_parser = subparsers.add_parser("info", help="Get spreadsheet info")
    info_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")

    # Update cells command
    update_parser = subparsers.add_parser("update-cells", help="Update cell values")
    update_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")
    update_parser.add_argument("--range", required=True, help="Range in A1 notation (e.g., Sheet1!A1:B2)")
    update_parser.add_argument("--values", required=True, help="JSON 2D array of values")

    # Append rows command
    append_parser = subparsers.add_parser("append-rows", help="Append rows to sheet")
    append_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")
    append_parser.add_argument("--range", required=True, help="Range in A1 notation (e.g., Sheet1!A:B)")
    append_parser.add_argument("--values", required=True, help="JSON 2D array of values")

    # Format header command
    header_parser = subparsers.add_parser("format-header", help="Format header row")
    header_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")
    header_parser.add_argument("--sheet-index", type=int, default=0, help="Sheet index")
    header_parser.add_argument("--columns", type=int, required=True, help="Number of columns")
    header_parser.add_argument("--bg-color", help="Background color as JSON")

    # Add table command
    table_parser = subparsers.add_parser("add-table", help="Add formatted table")
    table_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")
    table_parser.add_argument("--sheet-index", type=int, default=0, help="Sheet index")
    table_parser.add_argument("--start-row", type=int, default=0, help="Start row")
    table_parser.add_argument("--start-col", type=int, default=0, help="Start column")
    table_parser.add_argument("--data", required=True, help="JSON 2D array of data")
    table_parser.add_argument("--header-color", help="Header background color as JSON")

    # Find replace command
    find_parser = subparsers.add_parser("find-replace", help="Find and replace text")
    find_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")
    find_parser.add_argument("--find", required=True, help="Text to find")
    find_parser.add_argument("--replace", required=True, help="Replacement text")
    find_parser.add_argument("--sheet-index", type=int, help="Sheet index (all sheets if not specified)")
    find_parser.add_argument("--regex", action="store_true", help="Use regex")

    # Add sheet command
    add_sheet_parser = subparsers.add_parser("add-sheet", help="Add a new sheet")
    add_sheet_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")
    add_sheet_parser.add_argument("--title", required=True, help="Sheet title")
    add_sheet_parser.add_argument("--tab-color", help="Tab color as JSON")

    # Add formula command
    formula_parser = subparsers.add_parser("add-formula", help="Add formula to cell")
    formula_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")
    formula_parser.add_argument("--range", required=True, help="Range in A1 notation")
    formula_parser.add_argument("--formula", required=True, help="Formula (e.g., =SUM(A1:A10))")

    # Add chart command
    chart_parser = subparsers.add_parser("add-chart", help="Add a chart")
    chart_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")
    chart_parser.add_argument("--sheet-index", type=int, default=0, help="Sheet index")
    chart_parser.add_argument("--chart-type", required=True, help="Chart type (COLUMN, BAR, LINE, PIE)")
    chart_parser.add_argument("--title", required=True, help="Chart title")
    chart_parser.add_argument("--domain-range", required=True, help="Domain range (e.g., A1:A13)")
    chart_parser.add_argument("--series-range", required=True, help="Series range (e.g., B1:B13)")
    chart_parser.add_argument("--position-col", type=int, default=4, help="Position column")

    # Add conditional format command
    cond_parser = subparsers.add_parser("add-conditional-format", help="Add conditional formatting")
    cond_parser.add_argument("--sheet-id", required=True, help="Spreadsheet ID")
    cond_parser.add_argument("--sheet-index", type=int, default=0, help="Sheet index")
    cond_parser.add_argument("--start-row", type=int, required=True, help="Start row")
    cond_parser.add_argument("--end-row", type=int, required=True, help="End row")
    cond_parser.add_argument("--start-col", type=int, required=True, help="Start column")
    cond_parser.add_argument("--end-col", type=int, required=True, help="End column")
    cond_parser.add_argument("--condition", required=True, help="Condition type (e.g., NUMBER_GREATER)")
    cond_parser.add_argument("--value", required=True, help="Condition value")
    cond_parser.add_argument("--bg-color", help="Background color as JSON")

    args = parser.parse_args()

    try:
        if args.command == "create":
            sheets = args.sheets.split(",") if hasattr(args, 'sheets') and args.sheets else None
            sheet_id = create_spreadsheet(args.title, sheets)
            print(json.dumps({
                "spreadsheetId": sheet_id,
                "url": f"https://docs.google.com/spreadsheets/d/{sheet_id}/edit"
            }, indent=2))

        elif args.command == "info":
            info = get_sheet_info(args.sheet_id)
            print(json.dumps(info, indent=2))

        elif args.command == "update-cells":
            values = json.loads(args.values)
            result = update_values(args.sheet_id, args.range, values)
            print(json.dumps(result, indent=2))

        elif args.command == "append-rows":
            values = json.loads(args.values)
            result = append_values(args.sheet_id, args.range, values)
            print(json.dumps(result, indent=2))

        elif args.command == "format-header":
            bg_color = json.loads(args.bg_color) if hasattr(args, 'bg_color') and args.bg_color else None
            result = format_header_row(args.sheet_id, args.sheet_index, args.columns, bg_color=bg_color)
            print(json.dumps(result, indent=2))

        elif args.command == "add-table":
            data = json.loads(args.data)
            header_color = json.loads(args.header_color) if hasattr(args, 'header_color') and args.header_color else None
            result = add_table(args.sheet_id, args.sheet_index, args.start_row, args.start_col, data, header_color=header_color)
            print(json.dumps(result, indent=2))

        elif args.command == "find-replace":
            sheet_index = args.sheet_index if hasattr(args, 'sheet_index') and args.sheet_index is not None else None
            use_regex = hasattr(args, 'regex') and args.regex
            result = find_replace(args.sheet_id, args.find, args.replace, sheet_index=sheet_index, use_regex=use_regex)
            print(json.dumps(result, indent=2))

        elif args.command == "add-sheet":
            tab_color = json.loads(args.tab_color) if hasattr(args, 'tab_color') and args.tab_color else None
            result = add_sheet(args.sheet_id, args.title, tab_color=tab_color)
            print(json.dumps(result, indent=2))

        elif args.command == "add-formula":
            result = add_formula(args.sheet_id, args.range, args.formula)
            print(json.dumps(result, indent=2))

        elif args.command == "add-chart":
            result = add_chart(
                args.sheet_id, args.sheet_index, args.chart_type, args.title,
                args.domain_range, [args.series_range], position_col=args.position_col
            )
            print(json.dumps(result, indent=2))

        elif args.command == "add-conditional-format":
            bg_color = json.loads(args.bg_color) if hasattr(args, 'bg_color') and args.bg_color else None
            result = add_conditional_format(
                args.sheet_id, args.sheet_index,
                args.start_row, args.end_row, args.start_col, args.end_col,
                args.condition, args.value, bg_color=bg_color
            )
            print(json.dumps(result, indent=2))

    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
