# Adding Evals

Every new skill **MUST** have at least one eval. Evals verify that Claude Code invokes the correct skill for relevant prompts.

- **Core plugins** (databricks-tools, google-tools, fe-salesforce-tools, fe-internal-tools, vibe-setup, workflows): add evals to `evals/test-cases/skill-routing.yaml`
- **Non-core plugins** (all others): create a per-plugin file at `evals/test-cases/<plugin-name>.yaml`

## Test Case Format

### Single Expected Skill

```yaml
tests:
  - name: "descriptive-test-name"
    prompt: "Natural language prompt that should trigger the skill"
    expected_skill: "plugin-name:skill-name"
    max_turns: 5
    model: sonnet
```

### Any of Multiple Skills (OR Logic)

Use when multiple skills could validly handle a prompt:

```yaml
tests:
  - name: "auth-uses-some-auth-skill"
    prompt: "authenticate with databricks"
    expected_skill_one_of:
      - "databricks-tools:databricks-authentication"
      - "fe-internal-tools:aws-authentication"
    max_turns: 3
    model: haiku
```

### Multiple Skills Required (AND Logic)

Use when a prompt should trigger multiple skills:

```yaml
tests:
  - name: "deploy-app-uses-multiple-skills"
    prompt: "deploy my app to the workspace I created"
    expected_skills:
      - "databricks-tools:databricks-fe-vm-workspace-deployment"
      - "databricks-tools:databricks-apps"
    max_turns: 10
    model: opus
```

## Test Case Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Descriptive test name (kebab-case) |
| `prompt` | Yes | Natural language prompt to test |
| `expected_skill` | One of | Single skill that should be invoked |
| `expected_skill_one_of` | One of | List of valid skills (any one is acceptable) |
| `expected_skills` | One of | List of skills that must ALL be invoked |
| `max_turns` | Yes | Maximum API turns before timeout |
| `model` | Yes | Model to use: `haiku`, `sonnet`, or `opus` |

## Core vs. Non-Core Plugin Evals

**Core plugins** are the 6 plugins installed in CI and tested on every PR: databricks-tools, google-tools, fe-salesforce-tools, fe-internal-tools, vibe-setup, and workflows. Their evals live in `evals/test-cases/skill-routing.yaml` and failures block PRs.

**Non-core plugins** (jira-tools, fe-file-expenses, fe-dnb-hunting, etc.) each get their own eval file at `evals/test-cases/<plugin-name>.yaml`. CI auto-detects changes to non-core plugins and runs their evals, but failures are informational only and never block merges.

Both use the exact same YAML test format -- the only difference is which file you add tests to.

## Running Evals

```bash
cd evals

# Run default core tests
uv run skill-evals

# Run specific test file (core or non-core)
uv run skill-evals test-cases/my-tests.yaml

# Run all YAML files in a directory
uv run skill-evals --dir test-cases/

# Filter tests by name substring
uv run skill-evals -k jira

# Verbose output for debugging
uv run skill-evals --verbose

# Combine flags
uv run skill-evals --dir test-cases/ -k logfood --verbose
```

## Writing Good Test Prompts

**Do:**
- Use natural language that users would actually type
- Test edge cases and variations
- Include context that disambiguates the skill

**Don't:**
- Use the skill name directly in the prompt (unless testing that)
- Write overly specific prompts that only match one phrasing
- Assume context from previous turns

## Examples

```yaml
tests:
  # Good: Natural user request
  - name: "file-expenses-from-receipts"
    prompt: "I need to file my expense report for last week's travel"
    expected_skill: "fe-file-expenses:file-expenses"
    max_turns: 5
    model: sonnet

  # Good: Ambiguous prompt with multiple valid handlers
  - name: "query-data-uses-sql-skill"
    prompt: "run a query to get all users"
    expected_skill_one_of:
      - "databricks-tools:databricks-query"
      - "fe-internal-tools:logfood-querier"
    max_turns: 3
    model: haiku

  # Bad: Too specific, uses skill name
  - name: "bad-example"
    prompt: "use the databricks-authentication skill"  # Don't do this
    expected_skill: "databricks-tools:databricks-authentication"
    max_turns: 3
    model: haiku
```

## Troubleshooting

**Eval fails but skill works manually:**
- Check if prompt is ambiguous
- Increase `max_turns`
- Try a more capable model

**Flaky tests:**
- Prompts may be too ambiguous
- Consider using `expected_skill_one_of`
- Remove or mark as known-flaky
