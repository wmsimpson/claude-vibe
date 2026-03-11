#!/usr/bin/env python3
"""
Databricks Warehouse Selector

Reads JSON output from `databricks warehouses list --output=json` and selects
the top 3 warehouses based on a scoring system.

Scoring criteria:
- 1 point for serverless
- 1 point if size > Small
- 1 point if name doesn't start with 'A'
- min_num_clusters points
- floor(0.5 * max_num_clusters) points
"""

import json
import sys
from typing import List, Dict


def calculate_score(warehouse: Dict) -> int:
    """Calculate the score for a warehouse based on the criteria."""
    score = 0

    # 1 point for serverless
    if warehouse.get('enable_serverless_compute', False):
        score += 1

    # 1 point if size > Small
    # Size hierarchy: 2X-Small < X-Small < Small < Medium < Large < X-Large < 2X-Large < 3X-Large < 4X-Large
    cluster_size = warehouse.get('cluster_size', '').upper()
    if cluster_size and cluster_size not in ['2X-SMALL', 'X-SMALL', 'SMALL']:
        score += 1

    # 1 point if name doesn't start with 'A'
    warehouse_name = warehouse.get('name', '')
    if warehouse_name and not warehouse_name.upper().startswith('A'):
        score += 1

    # Add min_num_clusters points
    min_num_clusters = warehouse.get('min_num_clusters', 0)
    score += min_num_clusters

    # Add floor(0.5 * max_num_clusters) points
    max_num_clusters = warehouse.get('max_num_clusters', 0)
    score += int(0.5 * max_num_clusters)

    return score


def select_top_warehouses(warehouses_data) -> List[Dict[str, str]]:
    """Select the top 3 warehouses based on scoring criteria. Returns empty list if no warehouses."""
    # Handle both list and dict formats
    if isinstance(warehouses_data, list):
        warehouses = warehouses_data
    elif isinstance(warehouses_data, dict):
        warehouses = warehouses_data.get('warehouses', [])
    else:
        warehouses = []

    # Return empty list if no warehouses
    if not warehouses:
        return []

    # Calculate scores for each warehouse
    scored_warehouses = []
    for warehouse in warehouses:
        score = calculate_score(warehouse)
        scored_warehouses.append({
            'warehouse': warehouse,
            'score': score
        })

    # Sort by score (descending), then by name (ascending) for ties
    scored_warehouses.sort(
        key=lambda x: (-x['score'], x['warehouse'].get('name', ''))
    )

    # Get top 3 and return in the requested format
    top_3 = []
    for item in scored_warehouses[:3]:
        warehouse = item['warehouse']
        top_3.append({
            'warehouse_name': warehouse.get('name', ''),
            'warehouse_id': warehouse.get('id', '')
        })

    return top_3


def main():
    """Main entry point."""
    # Read JSON from stdin
    try:
        input_data = sys.stdin.read()
        warehouses_data = json.loads(input_data)
    except json.JSONDecodeError as e:
        print(f"Error parsing JSON: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error reading input: {e}", file=sys.stderr)
        sys.exit(1)

    # Select top 3 warehouses
    top_warehouses = select_top_warehouses(warehouses_data)

    # Output as JSON
    print(json.dumps(top_warehouses, indent=2))


if __name__ == '__main__':
    main()
