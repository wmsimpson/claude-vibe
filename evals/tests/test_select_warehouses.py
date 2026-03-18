"""Tests for select_warehouses.py — calculate_score() and select_top_warehouses()."""

from select_warehouses import calculate_score, select_top_warehouses


# ---------------------------------------------------------------------------
# calculate_score tests
# ---------------------------------------------------------------------------

class TestCalculateScore:
    def test_all_zeros(self):
        """Empty / minimal warehouse scores 0."""
        assert calculate_score({}) == 0

    def test_serverless_adds_one(self):
        wh = {"enable_serverless_compute": True, "name": "B-wh", "cluster_size": "Small"}
        assert calculate_score(wh) == 2  # serverless(1) + name-not-A(1)

    def test_size_greater_than_small(self):
        wh = {"cluster_size": "Medium", "name": "B-wh"}
        assert calculate_score(wh) == 2  # size(1) + name(1)

    def test_small_size_no_bonus(self):
        wh = {"cluster_size": "Small", "name": "B-wh"}
        assert calculate_score(wh) == 1  # name only

    def test_name_starts_with_a(self):
        wh = {"name": "Analytics Warehouse", "cluster_size": "Large"}
        assert calculate_score(wh) == 1  # size only, no name bonus

    def test_min_num_clusters(self):
        wh = {"name": "B-wh", "min_num_clusters": 3}
        assert calculate_score(wh) == 4  # name(1) + min_clusters(3)

    def test_max_num_clusters(self):
        wh = {"name": "B-wh", "max_num_clusters": 5}
        # floor(0.5 * 5) = 2, + name(1) = 3
        assert calculate_score(wh) == 3

    def test_full_score(self):
        wh = {
            "enable_serverless_compute": True,
            "cluster_size": "Large",
            "name": "Big-wh",
            "min_num_clusters": 2,
            "max_num_clusters": 4,
        }
        # serverless(1) + size(1) + name(1) + min(2) + floor(0.5*4)=2 = 7
        assert calculate_score(wh) == 7


# ---------------------------------------------------------------------------
# select_top_warehouses tests
# ---------------------------------------------------------------------------

class TestSelectTopWarehouses:
    def test_empty_list(self):
        assert select_top_warehouses([]) == []

    def test_single_warehouse(self):
        warehouses = [{"name": "wh-1", "id": "id-1"}]
        result = select_top_warehouses(warehouses)
        assert len(result) == 1
        assert result[0]["warehouse_name"] == "wh-1"
        assert result[0]["warehouse_id"] == "id-1"

    def test_returns_max_three(self):
        warehouses = [
            {"name": f"wh-{i}", "id": f"id-{i}", "min_num_clusters": i}
            for i in range(5)
        ]
        result = select_top_warehouses(warehouses)
        assert len(result) == 3

    def test_sorted_by_score_descending(self):
        warehouses = [
            {"name": "Low", "id": "1"},
            {"name": "High", "id": "2", "enable_serverless_compute": True, "cluster_size": "Large", "min_num_clusters": 5},
            {"name": "Mid", "id": "3", "cluster_size": "Medium"},
        ]
        result = select_top_warehouses(warehouses)
        assert result[0]["warehouse_name"] == "High"

    def test_tie_breaking_by_name(self):
        """Warehouses with equal scores are sorted alphabetically by name."""
        warehouses = [
            {"name": "Zebra", "id": "z"},
            {"name": "Baker", "id": "b"},
        ]
        result = select_top_warehouses(warehouses)
        # Both have score 1 (name not starting with A), tie-break = alphabetical
        assert result[0]["warehouse_name"] == "Baker"
        assert result[1]["warehouse_name"] == "Zebra"

    def test_dict_format_input(self):
        data = {"warehouses": [{"name": "wh-1", "id": "id-1"}]}
        result = select_top_warehouses(data)
        assert len(result) == 1
        assert result[0]["warehouse_name"] == "wh-1"
