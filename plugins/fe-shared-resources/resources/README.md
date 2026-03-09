# Shared Resources

This plugin hosts Python utilities and configuration files that are used by **two or more** vibe plugins. It has no skills or agents — just resources that get deployed to the plugin cache so other plugins can reference them.

## How it works

When installed, files land in:

```
~/.claude/plugins/cache/fe-vibe/fe-shared-resources/<version>/resources/
```

Other skills reference them with a glob pattern (same pattern used for `markdown_to_gdocs.py`, `google_auth.py`, etc.):

```bash
python3 ~/.claude/plugins/cache/fe-vibe/fe-shared-resources/*/resources/fiscal_calendar.py
```

## Current resources

| File | Purpose | Used by |
|------|---------|---------|
| `fiscal_calendar.py` | Databricks fiscal year/quarter context (FY starts Feb 1) | fe-manager, fe-internal-tools |

## When to add a resource here

A resource belongs in `fe-shared-resources` when:

1. **It is used by 2+ plugins** — don't pre-share something only one plugin needs
2. **It is a utility, not a skill** — skills belong in their own plugin
3. **It is stable** — shared resources should change infrequently since updates affect all consumers

## How to add a new resource

1. Add the file to `resources/`
2. Update the table above
3. Bump the version in `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json`
4. In consumer skills, reference via: `~/.claude/plugins/cache/fe-vibe/fe-shared-resources/*/resources/<file>`
