# Architecture Diagram Generator

Generate professional architecture diagrams with vendor icons (Databricks, AWS, GCP, Azure, Snowflake, Kafka) using mingrammer/diagrams (Python) or Mermaid. Features visual feedback-driven refinement via Chrome DevTools.

## How to Invoke

### Slash Command

```
/architecture-diagram
```

### Example Prompts

```
"Create a data lakehouse architecture diagram with Kafka, Delta Lake, and Snowflake"
"Generate a current-state vs future-state architecture diagram for a Snowflake migration"
"Build an AWS architecture diagram showing our customer's streaming pipeline into Databricks"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Graphviz | `brew install graphviz` for the diagrams engine |
| Python diagrams | Auto-installed in `~/.vibe/diagrams/` via setup script |
| Mermaid CLI (optional) | `npm install -g @mermaid-js/mermaid-cli` for Mermaid engine |
| Chrome DevTools MCP | For visual feedback loop (viewing and critiquing rendered diagrams) |

## What This Skill Does

1. Determines the best engine (mingrammer/diagrams for icon-heavy diagrams, Mermaid for simple flowcharts)
2. Selects or customizes a template matching the requested architecture
3. Generates the diagram with proper spacing, grouping, and vendor icons
4. Uses Chrome DevTools to view the rendered output and critique layout
5. Iterates on spacing, arrows, and grouping until the diagram looks professional
6. Exports in multiple formats (PNG, SVG, source file) for documentation or draw.io editing

## Key Resources

| File | Description |
|------|-------------|
| `resources/icons/` | Bundled custom icons for Databricks, Snowflake, Kafka, Confluent |
| `resources/templates/` | Pre-built diagram templates (data pipeline, lakehouse, multi-cloud, GenAI, etc.) |
| `resources/scripts/setup_diagrams_env.sh` | Setup script to install Python diagrams environment |
| `resources/references/` | Customization guide and Databricks product references |

## Related Skills

- `/product-question-research` - For validating architecture recommendations against product capabilities
- `/google-docs` - For inserting diagrams into Google Docs
- `/google-slides` - For inserting diagrams into presentations
