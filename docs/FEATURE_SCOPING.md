# Feature Scoping

**CRITICAL:** Define feature scope before starting work. Scope drift (making unrelated changes) creates confusing PRs and breaks worktree isolation.

## Before Making Changes

Explicitly define:

| Question | Example Answer |
|----------|----------------|
| **Primary goal** | Add a new skill for querying Lakebase |
| **In-scope files** | `plugins/databricks-tools/skills/lakebase-query/`, `permissions.yaml`, `evals/test-cases/skill-routing.yaml` |
| **Out-of-scope** | Other skills, agents, unrelated plugins, documentation |

## Key Rules

**NEVER:**
- Make "helpful" improvements to unrelated code
- Refactor outside scope
- Fix linter errors in unrelated files
- Add features not directly requested

**ALWAYS:**
- Ask before editing out-of-scope files
- Document scope at start
- Verify with `git status` regularly
- Keep PRs focused on one change

## Scope Boundaries by Feature Type

### New Skill

| In Scope | Out of Scope |
|----------|--------------|
| `plugins/<plugin>/skills/<skill-name>/` | Other skills in same plugin |
| `permissions.yaml` (add skill permission) | Other plugins |
| `evals/test-cases/skill-routing.yaml` | Agents |
| | Documentation files (unless skill-specific) |

### New Agent

| In Scope | Out of Scope |
|----------|--------------|
| `plugins/<plugin>/agents/<agent-name>.md` | Other agents |
| `permissions.yaml` (if needed) | Skills |
| | Other plugins |

### Bug Fix

| In Scope | Out of Scope |
|----------|--------------|
| Files directly causing the bug | Refactoring unrelated code |
| Related test files | "While I'm here" improvements |
| | Formatting changes to other files |

### Documentation Update

| In Scope | Out of Scope |
|----------|--------------|
| Specific doc files being updated | Code changes |
| Cross-references if links change | Unrelated documentation |

## Validation Checklist

Before creating a PR:

- [ ] Run `git status` - are all modified files in scope?
- [ ] Review `git diff` - any unintended changes?
- [ ] Check commit message - does it match the scope?
- [ ] If out-of-scope changes exist - create separate PR or ask user

## Examples

### Good: Focused Scope

```
Primary goal: Add databricks-lakebase skill
In scope:
  - plugins/databricks-tools/skills/databricks-lakebase/SKILL.md
  - plugins/databricks-tools/skills/databricks-lakebase/resources/
  - permissions.yaml
  - evals/test-cases/skill-routing.yaml
Out of scope:
  - Other databricks skills
  - fe-salesforce-tools plugin
  - CLAUDE.md
```

### Bad: Scope Drift

```
Primary goal: Add databricks-lakebase skill
Actual changes:
  - plugins/databricks-tools/skills/databricks-lakebase/SKILL.md ✓
  - permissions.yaml ✓
  - evals/test-cases/skill-routing.yaml ✓
  - plugins/databricks-tools/skills/databricks-query/SKILL.md ✗ (noticed typo)
  - CLAUDE.md ✗ (added lakebase to project structure)
  - plugins/google-tools/skills/google-docs/SKILL.md ✗ (fixed formatting)
```

## Handling Out-of-Scope Issues

If you notice issues outside your scope:

1. **Don't fix them in this PR**
2. **Document them** - create an issue or note
3. **Ask the user** if they want a separate PR
4. **Stay focused** on the original goal

## Quick Scope Check Commands

```bash
# See what's changed
git status

# See detailed changes
git diff

# See changes vs main
git diff main

# Reset unintended changes
git checkout -- <file>
```
