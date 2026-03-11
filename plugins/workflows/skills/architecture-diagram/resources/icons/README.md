# Custom Icons for Architecture Diagrams

This directory contains custom vendor icons **not available** in the mingrammer/diagrams package.

**Note**: The diagrams library includes **300+ built-in icons** for AWS, GCP, Azure, Kubernetes, and on-premises tools (Kafka, Spark, Airflow, PostgreSQL, etc.). These custom icons are only needed for:
- Databricks-specific products (not in the library)
- Snowflake (not in the library)
- Alternative icon styles

## Icon Specifications

- **Format**: PNG with transparent background
- **Size**: 256x256 pixels
- **Style**: Clean, professional, consistent with official branding

## Directory Structure

```
icons/
├── databricks/           # Databricks product icons
│   ├── workspace.png
│   ├── unity_catalog.png
│   ├── delta_lake.png
│   ├── lakehouse.png
│   ├── sql_warehouse.png
│   └── model_serving.png
├── cloud/                # Third-party cloud service icons
│   ├── snowflake.png
│   ├── kafka.png
│   ├── confluent.png
│   └── airflow.png
└── README.md
```

## Icon Sources & Licensing

### Databricks Icons
- **Source**: Official Databricks brand assets and documentation
- **Usage**: Internal Field Engineering use for customer architecture diagrams
- **Note**: These icons represent Databricks products and should be used in accordance with Databricks brand guidelines

### Cloud Service Icons
- **Snowflake**: Official Snowflake logo/icon
- **Kafka**: Apache Kafka official logo
- **Confluent**: Confluent official logo
- **Airflow**: Apache Airflow official logo

## Creating New Icons

1. Obtain the official logo/icon from the vendor's brand assets page
2. Convert to PNG if necessary (SVG → PNG at 256x256)
3. Ensure transparent background
4. Add to appropriate subdirectory
5. Update this README with source information

## Usage in Diagrams

```python
from diagrams.custom import Custom
from glob import glob
import os

def find_icons():
    patterns = [
        os.path.expanduser("~/.claude/plugins/cache/claude-vibe/workflows/*/skills/architecture-diagram/resources/icons"),
        os.path.expanduser("~/code/vibe/plugins/workflows/skills/architecture-diagram/resources/icons"),
    ]
    for pattern in patterns:
        matches = glob(pattern)
        if matches:
            return matches[0]
    raise FileNotFoundError("Icons directory not found")

ICONS = find_icons()

# Use custom icon
databricks_ws = Custom("Workspace", f"{ICONS}/databricks/workspace.png")
snowflake = Custom("Snowflake", f"{ICONS}/cloud/snowflake.png")
```

## Icon Placeholders

If icons are missing or you see placeholder images, you can:

1. Download official icons from vendor brand asset pages
2. Use the built-in icons from mingrammer/diagrams where available
3. Create simple text-based placeholders

### Vendor Brand Asset Pages

- **Databricks**: Internal brand resources
- **Snowflake**: https://www.snowflake.com/brand-guidelines/
- **Confluent**: https://www.confluent.io/brand/
- **Apache Projects**: https://apache.org/foundation/press/kit/
