# Adding Skills

Skills are markdown files that provide Claude Code with specialized instructions for specific tasks.

## Quick Start

1. Create `plugins/<plugin>/skills/<skill-name>/SKILL.md`
2. Add YAML frontmatter with name and description
3. Add `Skill(skill-name)` to `permissions.yaml`
4. Add at least one eval to `evals/test-cases/skill-routing.yaml`
5. Bump the plugin version (see [Contributing - Plugin Versioning](CONTRIBUTING.md#plugin-versioning))
6. Test with `/skill-name` in Claude Code

## Skill File Structure

```
plugins/<plugin>/skills/<skill-name>/
├── SKILL.md              # Main skill file (required)
└── resources/            # Optional supporting files
    ├── schema.yaml
    ├── helper.py
    └── examples/
```

## SKILL.md Format

```yaml
---
name: skill-name
description: Brief description shown in skill list
user-invocable: true  # Optional: allows /skill-name invocation
---

# Skill Title

Brief overview of what this skill does.

## Prerequisites

List any required setup, authentication, or dependencies.

## Workflow

Step-by-step instructions for Claude to follow.

## Examples

Show example usage and expected outputs.

## Troubleshooting

Common issues and solutions.
```

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Skill identifier (kebab-case) |
| `description` | Yes | Brief description for skill list |
| `user-invocable` | No | If true, users can invoke with `/skill-name` |

## Writing Effective Skills

### Be Specific
- Include exact commands, not just descriptions
- Specify file paths, API endpoints, parameter names
- Show expected output formats

### Be Defensive
- Include error handling instructions
- Document common failure modes
- Provide fallback approaches

### Use Resources
- Put schemas, configs, and helpers in `resources/`
- Reference them with relative paths
- Keep SKILL.md focused on workflow

## Permissions

Every skill MUST be added to `permissions.yaml`:

```yaml
permissions:
  allow:
    - "Skill(skill-name)"
```

Without this, users will get permission prompts when invoking the skill.

## Evals

Every skill MUST have at least one eval. See [Adding Evals](ADDING_EVALS.md).

```yaml
tests:
  - name: "skill-name-basic-usage"
    prompt: "Natural prompt that should trigger this skill"
    expected_skill: "plugin-name:skill-name"
    max_turns: 5
    model: sonnet
```

## Example: Complete Skill

```markdown
---
name: databricks-query
description: Execute SQL queries on Databricks
user-invocable: true
---

# Databricks Query

Execute SQL queries against Databricks warehouses.

## Prerequisites

- Authenticated with Databricks (use `/databricks-authentication` if needed)
- Active SQL warehouse

## Workflow

1. Check authentication status
2. Select appropriate warehouse
3. Execute query using Databricks MCP
4. Format and display results

## Example

User: "Query the trips table for rides over $50"

Response: Execute via databricks MCP:
\`\`\`sql
SELECT * FROM samples.nyctaxi.trips WHERE total_amount > 50 LIMIT 100
\`\`\`
```

## Testing Skills Locally

```bash
# Add local marketplace
claude plugin marketplace add /path/to/vibe

# Install plugin containing the skill
claude plugin install <plugin-name>@fe-vibe

# Test the skill
# In Claude Code: /skill-name
```

## Related Documentation

- [Adding Evals](ADDING_EVALS.md)
- [Feature Scoping](FEATURE_SCOPING.md)
- [Contributing](CONTRIBUTING.md)
