"""Tests for lakeview_builder.py — LakeviewDashboard builder class.

Schema validation is derived from 3 production Lakeview dashboards:
  - Amer Customer Insights (01ee3d32a76119d58aaa11f6ee39300d)
  - DBSQL AI Insights (01ef8c1903ff186ab84488c2378ebdad)
  - GTM Analytics Hub (01efb0ef62ff17dc883211afe1b41ced)
"""

import json

import pytest

from lakeview_builder import LakeviewDashboard

# ---------------------------------------------------------------------------
# Schema constants derived from production dashboards
# ---------------------------------------------------------------------------

REQUIRED_DASHBOARD_KEYS = {"datasets", "pages"}
REQUIRED_DATASET_KEYS = {"name", "displayName", "queryLines"}
REQUIRED_PAGE_KEYS = {"name", "displayName", "layout", "pageType"}
REQUIRED_POSITION_KEYS = {"x", "y", "width", "height"}
REQUIRED_WIDGET_KEYS = {"name", "queries", "spec"}
REQUIRED_QUERY_KEYS = {"name", "query"}
REQUIRED_QUERY_INNER_KEYS = {"datasetName", "fields", "disaggregated"}
REQUIRED_FIELD_KEYS = {"name", "expression"}
REQUIRED_SPEC_KEYS = {"version", "widgetType", "encodings", "frame"}
REQUIRED_FRAME_KEYS = {"showTitle", "title"}

VALID_WIDGET_TYPES = {
    "bar", "line", "pie", "counter", "scatter", "table", "combo",
    "filter-single-select", "filter-multi-select",
    "filter-date-picker", "filter-date-range-picker",
}

REQUIRED_ENCODINGS = {
    "bar": {"x", "y"},
    "line": {"x", "y"},
    "pie": {"angle", "color"},
    "counter": {"value"},
    "scatter": {"x", "y"},
    "table": {"columns"},
    "filter-single-select": {"fields"},
    "filter-multi-select": {"fields"},
    "filter-date-range-picker": {"fields"},
}

REQUIRED_AXIS_KEYS = {"fieldName", "scale", "displayName"}
VALID_SCALE_TYPES = {"quantitative", "temporal", "categorical"}


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _get_widget(dashboard, page_idx=0, widget_idx=0):
    return dashboard.pages[page_idx]["layout"][widget_idx]["widget"]


def _get_position(dashboard, page_idx=0, widget_idx=0):
    return dashboard.pages[page_idx]["layout"][widget_idx]["position"]


def _validate_dashboard_schema(dashboard):
    """Validate the full dashboard against the production schema."""
    data = dashboard.to_dict()

    assert REQUIRED_DASHBOARD_KEYS.issubset(data.keys()), \
        f"Missing top-level keys: {REQUIRED_DASHBOARD_KEYS - data.keys()}"

    for ds in data["datasets"]:
        assert REQUIRED_DATASET_KEYS.issubset(ds.keys()), \
            f"Dataset {ds.get('name')} missing keys: {REQUIRED_DATASET_KEYS - ds.keys()}"
        assert isinstance(ds["queryLines"], list), "queryLines must be a list"

    for page in data["pages"]:
        assert REQUIRED_PAGE_KEYS.issubset(page.keys()), \
            f"Page {page.get('name')} missing keys: {REQUIRED_PAGE_KEYS - page.keys()}"

        for layout_item in page["layout"]:
            pos = layout_item["position"]
            assert REQUIRED_POSITION_KEYS == set(pos.keys()), \
                f"Position keys mismatch: {set(pos.keys())}"
            for k in REQUIRED_POSITION_KEYS:
                assert isinstance(pos[k], int), f"position.{k} must be int"

            widget = layout_item["widget"]
            assert REQUIRED_WIDGET_KEYS.issubset(widget.keys()), \
                f"Widget missing keys: {REQUIRED_WIDGET_KEYS - widget.keys()}"

            for q in widget["queries"]:
                assert REQUIRED_QUERY_KEYS.issubset(q.keys())
                inner = q["query"]
                assert REQUIRED_QUERY_INNER_KEYS.issubset(inner.keys()), \
                    f"Query.query missing: {REQUIRED_QUERY_INNER_KEYS - inner.keys()}"
                for field in inner["fields"]:
                    assert REQUIRED_FIELD_KEYS.issubset(field.keys()), \
                        f"Field missing: {REQUIRED_FIELD_KEYS - field.keys()}"

            spec = widget["spec"]
            assert REQUIRED_SPEC_KEYS.issubset(spec.keys()), \
                f"Spec missing: {REQUIRED_SPEC_KEYS - spec.keys()}"
            wtype = spec["widgetType"]
            assert wtype in VALID_WIDGET_TYPES, f"Unknown widgetType: {wtype}"

            frame = spec["frame"]
            assert REQUIRED_FRAME_KEYS.issubset(frame.keys()), \
                f"Frame missing: {REQUIRED_FRAME_KEYS - frame.keys()}"

            encodings = spec["encodings"]
            if wtype in REQUIRED_ENCODINGS:
                required = REQUIRED_ENCODINGS[wtype]
                assert required.issubset(encodings.keys()), \
                    f"{wtype} encodings missing: {required - encodings.keys()}"

            for axis_key in ("x", "y"):
                if axis_key in encodings and isinstance(encodings[axis_key], dict):
                    ax = encodings[axis_key]
                    if "fieldName" in ax:
                        assert REQUIRED_AXIS_KEYS.issubset(ax.keys()), \
                            f"{wtype}.{axis_key} missing: {REQUIRED_AXIS_KEYS - ax.keys()}"
                        assert ax["scale"].get("type") in VALID_SCALE_TYPES, \
                            f"{wtype}.{axis_key} invalid scale type: {ax['scale'].get('type')}"

    assert json.loads(dashboard.to_json()) == data, "JSON round-trip mismatch"


# ---------------------------------------------------------------------------
# Init
# ---------------------------------------------------------------------------

class TestLakeviewDashboardInit:
    def test_default_name(self):
        assert LakeviewDashboard().name == "New Dashboard"

    def test_custom_name(self):
        assert LakeviewDashboard("My Dashboard").name == "My Dashboard"

    def test_creates_default_page(self):
        d = LakeviewDashboard()
        assert len(d.pages) == 1
        assert d.pages[0]["displayName"] == "Overview"


# ---------------------------------------------------------------------------
# Datasets
# ---------------------------------------------------------------------------

class TestAddDataset:
    def test_adds_dataset(self):
        d = LakeviewDashboard()
        name = d.add_dataset("sales", "Sales Data", "SELECT * FROM sales")
        assert name == "sales"
        assert d.datasets[0]["queryLines"] == ["SELECT * FROM sales"]



# ---------------------------------------------------------------------------
# Pages
# ---------------------------------------------------------------------------

class TestAddPage:
    def test_adds_second_page(self):
        d = LakeviewDashboard()
        d.add_page("Details")
        assert len(d.pages) == 2
        assert d.pages[1]["displayName"] == "Details"



# ---------------------------------------------------------------------------
# Bar chart
# ---------------------------------------------------------------------------

class TestAddBarChart:
    def test_schema_validation(self):
        d = LakeviewDashboard()
        d.add_dataset("ds", "DS", "SELECT 1")
        d.add_bar_chart("ds", "category", "amount", title="Revenue")
        _validate_dashboard_schema(d)

    def test_position(self):
        d = LakeviewDashboard()
        d.add_bar_chart("ds", "x", "y", position={"x": 1, "y": 2, "width": 4, "height": 5})
        assert _get_position(d) == {"x": 1, "y": 2, "width": 4, "height": 5}

    def test_sort_descending(self):
        d = LakeviewDashboard()
        d.add_bar_chart("ds", "x", "y", sort_descending=True)
        enc = _get_widget(d)["spec"]["encodings"]
        assert enc["x"]["scale"]["sort"] == {"by": "y-reversed"}

    def test_color_field(self):
        d = LakeviewDashboard()
        d.add_bar_chart("ds", "x", "y", color_field="region")
        enc = _get_widget(d)["spec"]["encodings"]
        assert enc["color"]["fieldName"] == "region"
        assert enc["color"]["scale"]["type"] in VALID_SCALE_TYPES

    def test_title_shown(self):
        d = LakeviewDashboard()
        d.add_bar_chart("ds", "x", "y", title="Revenue")
        frame = _get_widget(d)["spec"]["frame"]
        assert frame["showTitle"] is True
        assert frame["title"] == "Revenue"

    def test_no_title(self):
        d = LakeviewDashboard()
        d.add_bar_chart("ds", "x", "y")
        assert _get_widget(d)["spec"]["frame"]["showTitle"] is False

    def test_default_colors(self):
        d = LakeviewDashboard()
        d.add_bar_chart("ds", "x", "y")
        colors = _get_widget(d)["spec"]["mark"]["colors"]
        assert isinstance(colors, list) and len(colors) > 0
        assert all(c.startswith("#") for c in colors)


# ---------------------------------------------------------------------------
# Line chart
# ---------------------------------------------------------------------------

class TestAddLineChart:
    def test_schema_validation(self):
        d = LakeviewDashboard()
        d.add_dataset("ds", "DS", "SELECT 1")
        d.add_line_chart("ds", "date_col", "amount", time_grain="MONTH", title="Trend")
        _validate_dashboard_schema(d)

    def test_temporal_scale(self):
        d = LakeviewDashboard()
        d.add_line_chart("ds", "date_col", "amount", time_grain="MONTH")
        assert _get_widget(d)["spec"]["encodings"]["x"]["scale"]["type"] == "temporal"

    def test_categorical_without_time_grain(self):
        d = LakeviewDashboard()
        d.add_line_chart("ds", "category", "amount")
        assert _get_widget(d)["spec"]["encodings"]["x"]["scale"]["type"] == "categorical"

    def test_color_field(self):
        d = LakeviewDashboard()
        d.add_line_chart("ds", "date", "amount", color_field="region")
        assert "color" in _get_widget(d)["spec"]["encodings"]


# ---------------------------------------------------------------------------
# Pie chart
# ---------------------------------------------------------------------------

class TestAddPieChart:
    def test_schema_validation(self):
        d = LakeviewDashboard()
        d.add_dataset("ds", "DS", "SELECT 1")
        d.add_pie_chart("ds", "id", "region", "COUNT", title="Distribution")
        _validate_dashboard_schema(d)

    def test_encodings(self):
        d = LakeviewDashboard()
        d.add_pie_chart("ds", "id", "region", "COUNT")
        enc = _get_widget(d)["spec"]["encodings"]
        assert enc["angle"]["scale"]["type"] == "quantitative"
        assert enc["color"]["scale"]["type"] == "categorical"


# ---------------------------------------------------------------------------
# Counter
# ---------------------------------------------------------------------------

class TestAddCounter:
    def test_schema_validation(self):
        d = LakeviewDashboard()
        d.add_dataset("ds", "DS", "SELECT 1")
        d.add_counter("ds", "amount", "SUM", "Total Revenue")
        _validate_dashboard_schema(d)

    def test_counter_version(self):
        """Production counters use version 2, not 3."""
        d = LakeviewDashboard()
        d.add_counter("ds", "amount", "SUM")
        assert _get_widget(d)["spec"]["version"] == 2



# ---------------------------------------------------------------------------
# Scatter plot
# ---------------------------------------------------------------------------

class TestAddScatterPlot:
    def test_schema_validation(self):
        d = LakeviewDashboard()
        d.add_dataset("ds", "DS", "SELECT 1")
        d.add_scatter_plot("ds", "x_col", "y_col", title="Scatter")
        _validate_dashboard_schema(d)

    def test_quantitative_axes(self):
        d = LakeviewDashboard()
        d.add_scatter_plot("ds", "x_col", "y_col")
        enc = _get_widget(d)["spec"]["encodings"]
        assert enc["x"]["scale"]["type"] == "quantitative"
        assert enc["y"]["scale"]["type"] == "quantitative"


# ---------------------------------------------------------------------------
# Table
# ---------------------------------------------------------------------------

class TestAddTable:
    def test_schema_validation(self):
        d = LakeviewDashboard()
        d.add_dataset("ds", "DS", "SELECT 1")
        columns = [
            {"field": "name", "title": "Name", "type": "string"},
            {"field": "amount", "title": "Amount", "type": "float"},
        ]
        d.add_table("ds", columns, title="Data Table")
        _validate_dashboard_schema(d)

    def test_columns_encoding(self):
        d = LakeviewDashboard()
        d.add_table("ds", [{"field": "name", "type": "string"}])
        enc = _get_widget(d)["spec"]["encodings"]
        assert len(enc["columns"]) == 1
        assert enc["columns"][0]["fieldName"] == "name"


# ---------------------------------------------------------------------------
# Filters
# ---------------------------------------------------------------------------

class TestAddFilters:
    def test_dropdown_schema(self):
        d = LakeviewDashboard()
        d.add_dataset("ds", "DS", "SELECT 1")
        d.add_filter_dropdown("ds", "category", title="Category")
        _validate_dashboard_schema(d)

    def test_multi_select_type(self):
        d = LakeviewDashboard()
        d.add_filter_dropdown("ds", "category", multi_select=True)
        assert _get_widget(d)["spec"]["widgetType"] == "filter-multi-select"

    def test_single_select_type(self):
        d = LakeviewDashboard()
        d.add_filter_dropdown("ds", "category", multi_select=False)
        assert _get_widget(d)["spec"]["widgetType"] == "filter-single-select"

    def test_date_filter_schema(self):
        d = LakeviewDashboard()
        d.add_dataset("ds", "DS", "SELECT 1")
        d.add_date_filter("ds", "date_col", title="Date")
        _validate_dashboard_schema(d)


# ---------------------------------------------------------------------------
# Full dashboard / JSON output
# ---------------------------------------------------------------------------

class TestToJson:
    def test_full_dashboard_schema(self):
        """Build a realistic multi-widget dashboard and validate against production schema."""
        d = LakeviewDashboard("Sales Analytics")
        d.add_dataset("sales", "Sales Data", "SELECT * FROM main.default.sales")

        d.add_filter_dropdown("sales", "category", "Category",
                              position={"x": 0, "y": 0, "width": 1, "height": 2})
        d.add_counter("sales", "amount", "SUM", "Total Revenue",
                      position={"x": 0, "y": 2, "width": 1, "height": 2})
        d.add_bar_chart("sales", "category", "amount", "SUM", "Revenue by Category",
                        position={"x": 0, "y": 4, "width": 3, "height": 4})
        d.add_pie_chart("sales", "id", "region", "COUNT", "By Region",
                        position={"x": 3, "y": 4, "width": 3, "height": 4})
        d.add_line_chart("sales", "sale_date", "amount", "SUM",
                         time_grain="MONTH", title="Monthly Trend",
                         position={"x": 0, "y": 8, "width": 6, "height": 4})

        _validate_dashboard_schema(d)

    def test_to_dict_has_uisettings(self):
        assert "uiSettings" in LakeviewDashboard().to_dict()

    def test_get_api_payload(self):
        d = LakeviewDashboard("Payload Test")
        payload = d.get_api_payload("wh-123", "/Users/user@email.com")
        assert payload["display_name"] == "Payload Test"
        assert payload["warehouse_id"] == "wh-123"
        assert payload["parent_path"] == "/Users/user@email.com"
        assert REQUIRED_DASHBOARD_KEYS.issubset(json.loads(payload["serialized_dashboard"]).keys())

