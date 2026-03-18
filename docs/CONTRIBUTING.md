# Contributing to Vibe

This guide covers the contribution workflow for the Vibe plugin marketplace.

## Project and Contributor Philosophy

1. **AI-assisted contributions are encouraged, but stay engaged.** Using AI to build contributions is great, but submitting pure, untested, unvalidated slop makes reviews hard and the interplay of skills/plugins more brittle. Please be actively engaged in the process of building things you'd like to contribute.

2. **Build for the 95%.** The project goal is to automate doing the right thing and make best practices the default behavior for most users. Build with a generic best practice or example in mind.

3. **Simplicity > complexity.** Even if simplicity is a bit more verbose, prefer it over clever or complex solutions.

4. **Test on real workflows.** Please test what you build on the actual workflow it's designed for. Ideally, run the behavior by others before submission.

5. **Show your work.** PRs with screenshots, videos, outputs (e.g., doc links), logs, and chat sessions are highly encouraged.

## Getting Write Access

To contribute to this repository you need write access. By default, FEs only have read access.

Request the `role.github-emu.field-eng.vibe-write` role via [Opal](https://app.opal.dev/groups/081fdf07-9f0f-4046-8f18-08eeb47e1116). Once approved, your EMU GitHub account will have push access to the repo and you can open PRs directly.

## Development Workflow

**ALWAYS use git worktrees before creating new files or making any changes to this repository.**

1. **Use the `/using-git-worktrees` skill FIRST** - Creates an isolated worktree
2. **Define feature scope** - See [Feature Scoping](FEATURE_SCOPING.md)
3. Make all changes in the worktree directory
4. Commit changes in the worktree
5. **Rebase on main before creating a PR** (required)

**Why worktrees:** Multiple Claude instances can work on different features simultaneously without conflicts.

## Creating Pull Requests

When creating a PR, you **MUST** complete all sections of the PR template (`.github/PULL_REQUEST_TEMPLATE.md`):

| Section | Required Content |
|---------|------------------|
| **Summary** | Describe what the PR does |
| **Type of Change** | Check applicable boxes AND describe the issue/feature |
| **AI Assistance** | Document prompts used to create/modify code |
| **Testing** | Describe local testing and prompts used |
| **Related Issues** | Link to relevant issues or discussions |
| **Screenshots/Demos** | Before/after for UI changes |

**Incomplete sections will block merge.**

## Rebasing Requirement

All PRs **MUST** be rebased on main before merging:

```bash
git fetch origin && git rebase origin/main
```

This keeps commit history clean and avoids merge commits.

## Plugin Versioning

**Every change to a plugin requires a version bump.** We use a rolling versioning scheme where each component rolls over at 10:

- **Patch** increments by 1 for each change (e.g., `1.0.0` → `1.0.1` → ... → `1.0.9`)
- When patch reaches 10, **minor** increments and patch resets (e.g., `1.0.9` → `1.1.0`)
- When minor reaches 10, **major** increments and minor/patch reset (e.g., `1.9.9` → `2.0.0`)

**Example progression:**

```
1.0.0 → 1.0.1 → 1.0.2 → ... → 1.0.9 → 1.1.0 → 1.1.1 → ... → 1.9.9 → 2.0.0
```

**Where to bump the version:**

You must update the version in **both** places:

1. `plugins/<plugin>/.claude-plugin/plugin.json` — the `"version"` field
2. `.claude-plugin/marketplace.json` — the `"version"` field for that plugin's entry

**When to bump:** Any change to a plugin's skills, agents, commands, resources, or configuration files counts as a change requiring a version bump.

## Check for Duplicate Functionality - REQUIRED for New Additions

**Before creating any new skill, agent, or plugin, check whether similar functionality already exists.** This does NOT apply to fixes or improvements to existing skills/agents/plugins.

This marketplace is shared by hundreds of Field Engineers. Duplicate functionality confuses users, fragments maintenance, and creates routing conflicts.

### What to check

1. **Existing skills and agents across ALL plugins** — including non-default plugins you may not have installed. Search skill descriptions, agent descriptions, and the marketplace catalog.
2. **Git history and past PRs** — someone may have already built and submitted something similar. Search commit messages and PR titles with `git log --grep` and `gh pr list --search`.
3. **Who built the similar functionality** — use `git log --format="%an (%ae)"` on the relevant paths and `gh pr list` to identify authors so you can coordinate with them.

### If overlap is found

- **Flag it before proceeding.** Note the existing skill/plugin name, what it does, and who submitted it.
- **Coordinate with the original author** if possible.
- **If you still want to proceed** (regional differences, different approach, personal preference), that's fine — document the rationale in your PR description.

## Code Review Checklist

Before submitting a PR, verify:

- [ ] **Checked for duplicate/overlapping functionality** across all plugins (see above)
- [ ] All changes are within defined scope
- [ ] New skills have 1-2 evals, max 2 per skill (core plugins: `evals/test-cases/skill-routing.yaml`, non-core plugins: `evals/test-cases/<plugin-name>.yaml`)
- [ ] New skills are added to `permissions.yaml`
- [ ] **New plugins registered in root `.claude-plugin/plugin.json`** — skills path in `skills` array, agent paths in `agents` array (entries not listed here are silently ignored)
- [ ] **New agents added to root `.claude-plugin/plugin.json`** `agents` array
- [ ] **Plugin version bumped** in both `plugin.json` and `marketplace.json`
- [ ] Unit tests pass locally (`cd evals && uv run pytest -v`)
- [ ] No unrelated changes included
- [ ] **Google Docs output uses markdown converter** (see below)

## Creating Google Docs Output

**If your skill or agent creates Google Docs, you MUST use the markdown converter pattern.**

❌ **Don't** manually construct Google Docs API batch updates (`insertText`, `updateTextStyle`, `createParagraphBullets`, etc.)

✅ **Do** use the `markdown_to_gdocs.py` script from the `google-docs` skill:

1. Write output as markdown to a temp file
2. Convert using `google-tools/skills/google-docs/resources/markdown_to_gdocs.py`

```bash
python3 ~/.claude/plugins/cache/fe-vibe/google-tools/*/skills/google-docs/resources/markdown_to_gdocs.py \
  --input /tmp/your_output.md \
  --title "Document Title"
```

**Why:** Manual API calls are error-prone (index calculations), slow (multiple API calls), and hard to maintain. The markdown converter handles headings, bold, embedded hyperlinks, tables, and bullets automatically.

See `CLAUDE.md` for full details and examples.

## Testing Changes

### Unit tests (required before every PR)

Unit tests validate Python resource files — syntax, logic in the lakeview builder, query pretty-printer, warehouse selector, and markdown parser. They run locally with no auth required.

```bash
cd evals && uv run pytest -v
```

All tests must pass. If you modified a Python resource file under `plugins/`, the import smoke tests will automatically pick it up.

### Skill routing evals (run in CI, optional locally)

These test that Claude invokes the correct skill for a given prompt. They require an LLM backend, so they run automatically in CI on PRs. To run locally:

```bash
cd evals && uv run skill-evals --verbose
```

### Plugin installation testing

Use `vibe local` to streamline local testing. Run it from anywhere inside your vibe repo checkout:

```bash
cd ~/code/vibe

# Register local marketplace + install a plugin + launch agent
vibe local databricks-tools

# Install multiple plugins
vibe local databricks-tools google-tools

# Register local marketplace only (no install or agent)
vibe local --no-agent

# Manual equivalent (what vibe local automates)
claude plugin marketplace add /path/to/vibe
claude plugin install <plugin-name>@fe-vibe
vibe agent
```

## Contributing Non-Core Plugins

Non-core plugins are any plugins not in the 6 core set (databricks-tools, google-tools, fe-salesforce-tools, fe-internal-tools, vibe-setup, workflows). Examples include jira-tools, fe-file-expenses, fe-dnb-hunting, and fe-financialforce-tools.

### Eval Files for Non-Core Plugins

Each non-core plugin gets its own eval file at `evals/test-cases/<plugin-name>.yaml` (e.g., `evals/test-cases/jira-tools.yaml`). The YAML format is identical to `skill-routing.yaml`:

```yaml
tests:
  - name: "jira-create-ticket"
    prompt: "Create a JIRA ticket for the authentication bug"
    expected_skill: "jira-tools:jira-actions"
    max_turns: 5
    model: sonnet
```

### Testing Locally

```bash
cd evals
uv run skill-evals test-cases/<plugin-name>.yaml --verbose
```

### CI Behavior

When a PR touches files under `plugins/<non-core-plugin>/` or `evals/test-cases/<non-core-plugin>.yaml`, CI automatically:

1. Detects the affected non-core plugins
2. Installs all 6 core plugins (non-core skills may depend on them)
3. Installs each detected non-core plugin
4. Runs evals from the matching `evals/test-cases/<plugin>.yaml` file

Non-core eval failures are **informational only** -- they never block PR merges.

### Checklist for Non-Core Plugin PRs

- [ ] Created `evals/test-cases/<plugin-name>.yaml` with at least one test case
- [ ] Eval file uses the same YAML format as `skill-routing.yaml`
- [ ] Tested locally with `uv run skill-evals test-cases/<plugin-name>.yaml --verbose`
- [ ] Plugin registered in root `.claude-plugin/plugin.json` (`skills` and `agents` arrays)
- [ ] Plugin version bumped in both `plugin.json` and `marketplace.json`
- [ ] New skills added to `permissions.yaml`

## Related Documentation

- [Adding Skills](ADDING_SKILLS.md)
- [Adding Evals](ADDING_EVALS.md)
- [Feature Scoping](FEATURE_SCOPING.md)
