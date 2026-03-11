# Lucid Diagram

Generate architecture, data flow, and sequence diagrams as Graphviz DOT files and convert them to PNG images and Lucid Chart compatible XML for import.

## How to Invoke

### Slash Command

```
/lucid-diagram
```

### Example Prompts

```
"Create an architecture diagram of our microservices"
"Generate a data flow diagram for the ETL pipeline in this repo"
"Make a sequence diagram showing the authentication flow and export it for Lucid Chart"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Graphviz | Must be installed (`brew install graphviz` on macOS) |
| graphviz2drawio | Auto-installed by the conversion script for XML generation |

## What This Skill Does

1. Analyzes the codebase or user description to identify components and relationships
2. Confirms the output directory with the user (defaults to `docs/diagrams/`)
3. Generates a Graphviz `.dot` file with styled nodes, edges, and subgraphs
4. Runs `convert_to_lucid.py` to produce a PNG image and Lucid Chart XML
5. Provides instructions for embedding the PNG in docs or importing the XML into Lucid Chart

## Key Resources

| File | Description |
|------|-------------|
| `scripts/convert_to_lucid.py` | Converts DOT files to PNG and Lucid Chart XML |
| `references/graphviz_syntax.md` | Complete Graphviz DOT language reference |
