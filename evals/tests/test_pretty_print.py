"""Tests for databricks_query_pretty_print.py — PrettyTableFormatter and parse_databricks_response."""

import pytest

from databricks_query_pretty_print import PrettyTableFormatter, parse_databricks_response


# ---------------------------------------------------------------------------
# PrettyTableFormatter tests
# ---------------------------------------------------------------------------

class TestPrettyTableFormatter:
    def setup_method(self):
        self.fmt = PrettyTableFormatter(max_width=120, max_col_width=50)

    def test_empty_headers(self):
        assert self.fmt.format_table([], []) == "No data to display"

    def test_multiple_rows(self):
        headers = ["ID", "Value"]
        rows = [["1", "a"], ["2", "b"], ["3", "c"]]
        table = self.fmt.format_table(headers, rows)
        for val in ["1", "2", "3", "a", "b", "c"]:
            assert val in table

    def test_none_values_rendered_as_empty(self):
        table = self.fmt.format_table(["Col"], [[None]])
        # None becomes empty string; no "None" text
        lines = table.split("\n")
        data_line = lines[3]  # header, separator, data
        assert "None" not in data_line

    def test_truncation(self):
        fmt = PrettyTableFormatter(max_width=120, max_col_width=10)
        long_val = "a" * 50
        table = fmt.format_table(["Col"], [[long_val]])
        assert "..." in table

    def test_row_shorter_than_headers_padded(self):
        """Rows with fewer columns than headers should be padded."""
        table = self.fmt.format_table(["A", "B", "C"], [["1"]])
        # Should not error and should produce a table
        assert "A" in table



# ---------------------------------------------------------------------------
# parse_databricks_response tests
# ---------------------------------------------------------------------------

class TestParseDatabricksResponse:
    def test_success_response(self):
        data = {
            "status": {"state": "SUCCEEDED"},
            "manifest": {
                "schema": {
                    "columns": [
                        {"name": "id"},
                        {"name": "value"},
                    ]
                }
            },
            "result": {
                "data_array": [["1", "hello"], ["2", "world"]]
            },
        }
        headers, rows = parse_databricks_response(data)
        assert headers == ["id", "value"]
        assert len(rows) == 2
        assert rows[0] == ["1", "hello"]

    def test_failed_query_raises(self):
        data = {
            "status": {
                "state": "FAILED",
                "error": {
                    "error_code": "SYNTAX_ERROR",
                    "message": "bad sql",
                },
            }
        }
        with pytest.raises(ValueError, match="bad sql"):
            parse_databricks_response(data)

    def test_empty_result(self):
        data = {
            "status": {"state": "SUCCEEDED"},
            "manifest": {"schema": {"columns": [{"name": "x"}]}},
            "result": {},
        }
        headers, rows = parse_databricks_response(data)
        assert headers == ["x"]
        assert rows == []

    def test_external_links_raises(self):
        data = {
            "status": {"state": "SUCCEEDED"},
            "manifest": {"schema": {"columns": [{"name": "x"}]}},
            "result": {"external_links": [{"url": "https://example.com"}]},
        }
        with pytest.raises(ValueError, match="externally"):
            parse_databricks_response(data)

    def test_missing_column_name_fallback(self):
        """Columns without a name key get a col_N fallback."""
        data = {
            "status": {"state": "SUCCEEDED"},
            "manifest": {"schema": {"columns": [{}]}},
            "result": {"data_array": [["val"]]},
        }
        headers, _ = parse_databricks_response(data)
        assert headers == ["col_0"]
