#!/usr/bin/env python3
"""
Pretty ASCII Table Formatter for Databricks SQL Query Results

Parses JSON output from databricks CLI and formats as a beautiful ASCII table.
Uses only Python standard library.

Usage:
    # From stdin (pipe output):
    databricks api post /api/2.0/sql/statements/ --json='...' | python databricks_query_pretty.py

    # From file:
    databricks api post /api/2.0/sql/statements/ --json='...' > output.json
    python databricks_query_pretty.py output.json

    # File will be automatically deleted after successful parsing
"""

import argparse
import json
import sys
import os
import shutil
from typing import List, Any, Tuple


class PrettyTableFormatter:
    """Formats data as a beautiful ASCII table with intelligent column sizing."""

    # Box drawing characters for pretty tables
    HORIZONTAL = '─'
    VERTICAL = '│'
    TOP_LEFT = '┌'
    TOP_RIGHT = '┐'
    BOTTOM_LEFT = '└'
    BOTTOM_RIGHT = '┘'
    TOP_JUNCTION = '┬'
    BOTTOM_JUNCTION = '┴'
    LEFT_JUNCTION = '├'
    RIGHT_JUNCTION = '┤'
    CROSS = '┼'

    def __init__(self, max_width: int = None, max_col_width: int = 50):
        """
        Initialize the formatter.

        Args:
            max_width: Maximum table width (defaults to terminal width)
            max_col_width: Maximum width for any single column
        """
        self.max_width = max_width or self._get_terminal_width()
        self.max_col_width = max_col_width

    def _get_terminal_width(self) -> int:
        """Get the current terminal width."""
        try:
            return shutil.get_terminal_size().columns
        except:
            return 120  # Default fallback

    def _truncate_string(self, s: str, max_len: int) -> str:
        """Truncate string to max length with ellipsis if needed."""
        s = str(s)
        if len(s) <= max_len:
            return s
        return s[:max_len-3] + '...'

    def _calculate_column_widths(self, headers: List[str], rows: List[List[str]]) -> List[int]:
        """
        Calculate optimal column widths based on content and terminal size.

        Args:
            headers: List of column headers
            rows: List of data rows

        Returns:
            List of column widths
        """
        num_cols = len(headers)
        if num_cols == 0:
            return []

        # Calculate minimum required width for each column
        col_widths = []
        for i in range(num_cols):
            # Start with header width
            max_width = len(headers[i])

            # Check all rows for this column
            for row in rows:
                if i < len(row):
                    max_width = max(max_width, len(str(row[i])))

            # Apply maximum column width limit
            col_widths.append(min(max_width, self.max_col_width))

        # Calculate total width needed (including borders and padding)
        # Format: │ col1 │ col2 │ col3 │
        # That's: num_cols + 1 vertical bars + 2 * num_cols spaces for padding
        border_width = (num_cols + 1) + (2 * num_cols)
        total_width = sum(col_widths) + border_width

        # If table is too wide, proportionally reduce column widths
        if total_width > self.max_width:
            available_width = self.max_width - border_width
            if available_width > 0:
                # Reduce each column proportionally
                scale_factor = available_width / sum(col_widths)
                col_widths = [max(3, int(w * scale_factor)) for w in col_widths]

        return col_widths

    def _format_row(self, values: List[str], widths: List[int], center_header: bool = False) -> str:
        """
        Format a single row with proper padding and borders.

        Args:
            values: List of cell values
            widths: List of column widths
            center_header: If True, center text in cells (for headers)

        Returns:
            Formatted row string
        """
        cells = []
        for i, (value, width) in enumerate(zip(values, widths)):
            value_str = self._truncate_string(str(value), width)
            if center_header:
                cells.append(value_str.center(width))
            else:
                cells.append(value_str.ljust(width))

        return f"{self.VERTICAL} {f' {self.VERTICAL} '.join(cells)} {self.VERTICAL}"

    def _format_separator(self, widths: List[int], position: str = 'middle') -> str:
        """
        Format a separator line.

        Args:
            widths: List of column widths
            position: 'top', 'middle', or 'bottom'

        Returns:
            Formatted separator string
        """
        if position == 'top':
            left = self.TOP_LEFT
            junction = self.TOP_JUNCTION
            right = self.TOP_RIGHT
        elif position == 'bottom':
            left = self.BOTTOM_LEFT
            junction = self.BOTTOM_JUNCTION
            right = self.BOTTOM_RIGHT
        else:  # middle
            left = self.LEFT_JUNCTION
            junction = self.CROSS
            right = self.RIGHT_JUNCTION

        segments = [self.HORIZONTAL * (w + 2) for w in widths]
        return f"{left}{junction.join(segments)}{right}"

    def format_table(self, headers: List[str], rows: List[List[Any]]) -> str:
        """
        Format data as a pretty ASCII table.

        Args:
            headers: List of column headers
            rows: List of data rows

        Returns:
            Formatted table string
        """
        if not headers:
            return "No data to display"

        # Convert all row values to strings and handle None
        str_rows = [[str(cell) if cell is not None else '' for cell in row] for row in rows]

        # Calculate column widths
        widths = self._calculate_column_widths(headers, str_rows)

        # Build the table
        lines = []

        # Top border
        lines.append(self._format_separator(widths, 'top'))

        # Header row (centered)
        lines.append(self._format_row(headers, widths, center_header=True))

        # Header separator
        lines.append(self._format_separator(widths, 'middle'))

        # Data rows
        for row in str_rows:
            # Pad row if it has fewer columns than headers
            padded_row = row + [''] * (len(headers) - len(row))
            lines.append(self._format_row(padded_row, widths))

        # Bottom border
        lines.append(self._format_separator(widths, 'bottom'))

        return '\n'.join(lines)


def parse_databricks_response(data: dict) -> Tuple[List[str], List[List[Any]]]:
    """
    Parse Databricks API JSON response into headers and rows.

    Args:
        data: JSON response from Databricks API

    Returns:
        Tuple of (headers, rows)
    """
    # Check if query was successful
    status = data.get('status', {}).get('state')
    if status != 'SUCCEEDED':
        error_info = data.get('status', {}).get('error', {})
        error_msg = error_info.get('message', 'Unknown error')
        error_type = error_info.get('error_code', 'ERROR')
        raise ValueError(f"Query failed [{error_type}]: {error_msg}")

    # Extract manifest (contains column schema)
    manifest = data.get('manifest', {})
    schema = manifest.get('schema', {})
    columns = schema.get('columns', [])

    # Extract headers
    headers = [col.get('name', f'col_{i}') for i, col in enumerate(columns)]

    # Extract result data
    result = data.get('result', {})

    # Handle different result formats
    if 'data_array' in result:
        # JSON_ARRAY format: list of lists
        rows = result['data_array']
    elif 'data_typed_array' in result:
        # Typed array format
        rows = result['data_typed_array']
    elif 'external_links' in result:
        # Large result set stored externally
        raise ValueError("Results are stored externally. This script only handles inline results.")
    else:
        rows = []

    return headers, rows


def read_input(file_path: str = None) -> dict:
    """
    Read JSON input from stdin or file.

    Args:
        file_path: Path to JSON file (None for stdin)

    Returns:
        Parsed JSON data
    """
    try:
        if file_path:
            with open(file_path, 'r') as f:
                data = json.load(f)
        else:
            data = json.load(sys.stdin)
        return data
    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON input - {e}", file=sys.stderr)
        sys.exit(1)
    except FileNotFoundError:
        print(f"Error: File not found - {file_path}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error reading input: {e}", file=sys.stderr)
        sys.exit(1)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Format Databricks SQL query results as pretty ASCII tables',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # From stdin (pipe):
  databricks api post /api/2.0/sql/statements/ \\
    --json='{"statement": "SELECT 1", "warehouse_id": "b21e2c56a857fe88", \\
            "format":"JSON_ARRAY", "wait_timeout":"50s"}' \\
    --profile=logfood | python databricks_query_pretty.py

  # From file (file will be deleted after parsing):
  databricks api post /api/2.0/sql/statements/ --json='...' > result.json
  python databricks_query_pretty.py result.json

  # Custom table width:
  cat result.json | python databricks_query_pretty.py --max-width 200
        """
    )

    parser.add_argument(
        'file',
        nargs='?',
        help='JSON file containing Databricks API response (will be deleted after parsing). If omitted, reads from stdin.'
    )
    parser.add_argument(
        '--max-width',
        type=int,
        help='Maximum table width in characters (default: terminal width)'
    )
    parser.add_argument(
        '--max-col-width',
        type=int,
        default=50,
        help='Maximum width for any single column (default: 50)'
    )
    parser.add_argument(
        '--no-delete',
        action='store_true',
        help='Do not delete the input file after parsing'
    )

    args = parser.parse_args()

    # Read JSON input
    data = read_input(args.file)

    # Parse the response
    try:
        headers, rows = parse_databricks_response(data)
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

    # Format and print table
    formatter = PrettyTableFormatter(
        max_width=args.max_width,
        max_col_width=args.max_col_width
    )
    table = formatter.format_table(headers, rows)
    print(table)

    # Print summary
    print(f"\n{len(rows)} row(s) in set")

    # Delete input file if it was provided and deletion is enabled
    if args.file and not args.no_delete:
        try:
            os.remove(args.file)
        except Exception as e:
            print(f"Warning: Failed to delete file {args.file}: {e}", file=sys.stderr)


if __name__ == '__main__':
    main()
