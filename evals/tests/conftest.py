"""Pytest configuration for vibe eval tests.

Adds plugin resource directories to sys.path so tests can import
Python modules from plugin resources directly.
"""

import sys
from pathlib import Path

# Root of the vibe repository (evals/../)
VIBE_ROOT = Path(__file__).resolve().parent.parent.parent

# All plugin resource directories that contain Python files
RESOURCE_DIRS = [
    "plugins/databricks-tools/skills/databricks-warehouse-selector/resources",
    "plugins/databricks-tools/skills/databricks-query/resources",
    "plugins/databricks-tools/skills/databricks-lakeview-dashboard/resources",
    "plugins/google-tools/skills/google-docs/resources",
]

for d in RESOURCE_DIRS:
    path = str(VIBE_ROOT / d)
    if path not in sys.path:
        sys.path.insert(0, path)
