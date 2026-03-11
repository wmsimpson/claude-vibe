---
name: competitive-analysis
description: Competitive analysis comparing Databricks to competitors like Snowflake, Redshift, Fabric, Polaris, Purview. Use for "compare", "vs", "why Databricks over X", "competitive", "battlecard" questions. Fast mode (default) for quick answers using registry differentiators. Comprehensive mode when user requests "create", "generate", "comprehensive", "detailed", "battlecard", "document", or "research" - includes Glean searches and battlecard documents.
user-invocable: true
---

# Competitive Analysis Skill

This skill provides competitive analysis comparing Databricks to competitors with two modes: fast (default) for quick chat answers, and comprehensive for detailed research with document generation.

## Quick Start

**Fast Mode (Default):**
1. Read `resources/competitive/battlecard-registry.yaml` for pre-defined differentiators
2. Optionally read one battlecard if available
3. Return concise chat response in < 30 seconds

**Comprehensive Mode (When Requested):**
1. Read battlecard registry and discover relevant battlecards
2. Search Glean for internal competitive intelligence
3. Research Databricks and competitor features
4. Generate Google Doc following `resources/competitive/compete-framework.md`

**Mode selection:** Fast mode unless prompt contains "create", "generate", "comprehensive", "detailed", "battlecard", "document", or "research".

## Supported Competitors

| Compete Type | Competitors | Analysis Scope |
|--------------|-------------|----------------|
| Platform | Snowflake, Amazon Redshift, Microsoft Fabric | Architecture, compute, storage, ML, streaming, governance |
| Governance | Polaris, Microsoft Purview, AWS Glue Catalog | Catalog, lineage, access control, data quality |
| OLTP | Lakebase, Supabase | Use case differentiation (analytics/ML vs transactions) |

## Example Invocations

```
User: Why Databricks over Snowflake for streaming?
→ Fast mode: Registry differentiators, quick chat answer

User: Compare Unity Catalog to Polaris
→ Fast mode: Governance-focused, registry-based answer

User: Create a comprehensive battlecard comparing Databricks to Redshift
→ Comprehensive mode: Full research, Glean searches, Google Doc

User: Generate a detailed competitive analysis document for Snowflake
→ Comprehensive mode: Multiple battlecards, competitor research, formatted doc
```

## Workflow

### Fast Mode (Default)

1. **Parse question** - Extract competitor name, product area, use case
2. **Read registry** - `resources/competitive/battlecard-registry.yaml` for differentiators
3. **Optional battlecard** - Read one if available (often skipped for speed)
4. **Generate answer** - Concise chat response using registry differentiators

**Skip in fast mode:** Glean searches, competitor research, Google Doc generation

### Comprehensive Mode (When Requested)

1. **Parse question** - Extract competitor, compete type, product area, use case, customer context
2. **Discover battlecards** - Find relevant battlecards from registry
3. **Read battlecards** - Extract differentiators, gaps, proof points
4. **Search Glean** - Internal competitive intelligence, win/loss stories, compete docs
5. **Research Databricks** - Use `/product-question-research` for feature capabilities
6. **Research competitor** - Use `WebSearch` and `WebFetch` for competitor docs
7. **Generate output** - Chat response or Google Doc (if requested)

## Document Formatting Rules

| Element | Format |
|---------|--------|
| Differentiators | Bold key terms, specific metrics (not vague claims) |
| Competitor assessment | Fair and accurate (don't understate strengths) |
| Sources | All claims cited with links (battlecards, Glean docs, public docs) |
| Tables | Feature comparison with proper alignment |
| Links | Clickable hyperlinks (not bare URLs) |

## Document Structure

For comprehensive mode Google Docs, follow `resources/competitive/compete-framework.md`:

1. Quick Answer (30-second elevator pitch)
2. Feature Comparison (scaled to compete type)
3. Why Databricks (3-5 key differentiators)
4. Competitor Limitations (gaps, fairly stated)
5. Use Case Fit (when each solution wins)
6. Objection Handling (common questions with responses)
7. Proof Points (customer examples, benchmarks)
8. Next Steps (demo suggestions, POC guidance)
9. Sources (all citations)

## Output Format

**Fast Mode:**
- Chat response only
- 2-3 differentiators from registry
- Concise format (see `resources/competitive/examples.md`)

**Comprehensive Mode:**
- Chat response for simple questions
- Google Doc when user requests "create", "generate", "document", "battlecard"
- Use `markdown_to_gdocs.py` script for formatting

## Resources

- `resources/competitive/battlecard-registry.yaml` - Competitor differentiators and battlecard URLs
- `resources/competitive/compete-framework.md` - Standard template for compete analysis documents
- `resources/competitive/examples.md` - Example outputs and interaction patterns

## Integration with Other Skills

| Skill/Tool | Used For | Mode |
|------------|----------|------|
| `/google-docs` | Reading battlecards from Google Drive | Comprehensive only |
| `mcp__glean__glean_read_api_call` | Internal competitive intelligence, win/loss stories | Comprehensive only |
| `/product-question-research` | Researching Databricks features | Both (minimal in fast mode) |
| `WebSearch` / `WebFetch` | Competitor information and documentation | Comprehensive only |
| `battlecard-registry.yaml` | Pre-defined differentiators and focus areas | Both (primary in fast mode) |

**Markdown converter:** Uses `google-tools` markdown_to_gdocs.py script for formatted output (comprehensive mode only).
