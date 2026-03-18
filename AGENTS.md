# Vibe - Claude Code Marketplace

This is the Field Engineering Claude Code plugin marketplace. It provides specialized skills and agents for Databricks, Salesforce, Google Workspace, AWS, and internal analytics tools.

## Development Workflow - CRITICAL

**ALWAYS use git worktrees before creating new files or making any changes to this repository.**

This allows parallel work on the same repo without conflicts. Before making any edits:

1. **Define the feature scope** - Explicitly list which files/paths are in scope and out of scope (see [Feature Scoping](#feature-scoping---preventing-scope-drift) below)
2. **Use the `/using-git-worktrees` skill** - Create an isolated worktree for your changes
3. **Make all changes in the worktree directory** - Stay within your defined scope
4. **Commit changes in the worktree** - Verify only in-scope files are modified
5. **ALWAYS rebase on main before creating a PR** - This is required, not optional

**Why this matters:** Multiple Claude instances or parallel tasks can work on different features simultaneously without interfering with each other. This is especially important for vibe since it's a shared resource used by the Field Engineering team.

### NEVER Push Directly to Main

**All changes MUST go through a Pull Request.** This repository is used by many people on the Field Engineering team.

- **DO NOT** use `git push origin main` or merge directly to main
- **ALWAYS** create a feature branch and open a PR for review
- **ALWAYS** wait for PR approval before merging
- **NEVER create a PR until explicitly told** - Prepare the branch and commits, but wait for user approval before running `gh pr create`

### Creating Pull Requests

When creating a PR, you **MUST** complete all sections of the PR template (`.github/PULL_REQUEST_TEMPLATE.md`):

- **Summary**: Describe what the PR does
- **Type of Change**: Check applicable boxes AND describe what issue this fixes or what feature/improvement is added
- **AI Assistance**: If AI was used, document which prompts were used to create/modify the code
- **Testing**: Describe how this was tested locally and which prompts were used to test it
- **Related Issues**: Link to relevant issues or discussions
- **Screenshots/Demos**: Include before/after for UI changes

If any sections are incomplete, provide the missing information before the PR can be merged.

**Rebasing Requirement:** All PRs **MUST** be rebased on main before merging. Run `git fetch origin && git rebase origin/main` before pushing your branch. This keeps the commit history clean and avoids merge commits.

### Feature Scoping - Preventing Scope Drift

**CRITICAL:** Define feature scope before starting work. See `docs/FEATURE_SCOPING.md` for detailed guidance.

## Project Structure

```
vibe/
├── .claude-plugin/plugin.json        # Root manifest (registers all plugin skills/agents)
├── .claude-plugin/marketplace.json   # Marketplace catalog (defines all plugins)
├── plugins/                          # Plugin directories
│   ├── databricks-tools/          # Databricks integration skills + agents
│   ├── fe-salesforce-tools/          # Salesforce CRM operations
│   ├── google-tools/              # Google Docs/Slides/Drive integration
│   ├── specialized-agents/        # CLI executor, diagrams, web testing
│   ├── fe-internal-tools/            # Logfood analytics, AWS auth
│   ├── vibe-setup/                # Environment setup and validation
│   └── mcp-servers/               # Custom MCP server framework
├── permissions.yaml                  # Master permissions config
└── mcp-servers.yaml                  # External MCP server configs
```

## Development Guidelines

### Adding Skills

1. Skills are markdown files in `plugins/<plugin-name>/skills/`
2. Use YAML frontmatter for metadata:
   ```yaml
   ---
   name: skill-name
   description: Brief description
   ---
   ```
3. Include clear step-by-step instructions in the body
4. Reference resource files for schemas, configs, or helper scripts
5. **REQUIRED: Add 1-2 evals (max 2 per skill)** - Core plugins: `evals/test-cases/skill-routing.yaml`. Non-core plugins: `evals/test-cases/<plugin-name>.yaml`. See [Adding Evals](#adding-evals) below

### Adding Agents

1. Agents are markdown files in `plugins/<plugin-name>/agents/`
2. Document the agent's purpose, capabilities, and when to use it
3. Specify which tools the agent has access to
4. Include usage examples

### Creating Google Docs Output

**If your skill or agent creates Google Docs, you MUST use the markdown converter pattern.**

Do NOT manually construct Google Docs API batch updates with `insertText`, `updateTextStyle`, `createParagraphBullets`, etc. This approach is:
- Error-prone (requires calculating character indices)
- Slow (multiple API calls)
- Hard to maintain

**Instead, use the `markdown_to_gdocs.py` script from the `google-docs` skill:**

1. Write your output as a markdown file (to `/tmp/` or similar)
2. Convert using the script from `google-tools/skills/google-docs/resources/markdown_to_gdocs.py`

```bash
python3 ~/.claude/plugins/cache/fe-vibe/google-tools/*/skills/google-docs/resources/markdown_to_gdocs.py \
  --input /tmp/your_output.md \
  --title "Document Title"
```

**The script handles:**
- Headings (`#` → HEADING_1, `##` → HEADING_2, etc.)
- Bold (`**text**`) and italic (`*text*`)
- Embedded hyperlinks (`[text](url)` → clickable blue links, no visible URLs)
- Tables with links in cells (`| [Link](url) |`)
- Bullet and numbered lists
- Code blocks

**Example markdown structure:**
```markdown
# Document Title

## Section

Content with **bold text** and [embedded links](https://example.com).

| Column 1 | Column 2 |
|----------|----------|
| [Link](url) | Description |

- Bullet item 1
- Bullet item 2
```

See existing agents for examples: `product-question-researcher.md`, `rca-doc.md`, `customer-question-answerer.md`.

### Plugin Versioning

**Every change to a plugin requires a version bump.** We use a rolling versioning scheme where each component rolls over at 10:

- **Patch** increments by 1 for each change (e.g., `1.0.0` → `1.0.1` → ... → `1.0.9`)
- When patch reaches 10, **minor** increments and patch resets (e.g., `1.0.9` → `1.1.0`)
- When minor reaches 10, **major** increments and minor/patch reset (e.g., `1.9.9` → `2.0.0`)

**You must update the version in both places:**
1. `plugins/<plugin>/.claude-plugin/plugin.json` — the `"version"` field
2. `.claude-plugin/marketplace.json` — the `"version"` field for that plugin's entry

### Modifying Permissions

1. Edit `permissions.yaml` for new permissions
2. Permissions are merged into `~/.claude/settings.json` during installation
3. Common permission patterns:
   - `Skill(skill-name)` - Allow invoking a skill
   - `Bash(command:*)` - Allow specific bash commands
   - `Read(path/**)` - Allow reading files in a path
   - `Edit(path/**)` - Allow editing files in a path

### Adding Evals

Every new skill **MUST** have at least one eval and **no more than 2 evals per skill**. Evals verify that Claude Code invokes the correct skill for relevant prompts.

- **Core plugins** (databricks-tools, google-tools, fe-salesforce-tools, fe-internal-tools, vibe-setup, workflows): add evals to `evals/test-cases/skill-routing.yaml`
- **Non-core plugins** (all others): create `evals/test-cases/<plugin-name>.yaml`

**Test case format:**

```yaml
tests:
  # Single expected skill
  - name: "descriptive-test-name"
    prompt: "Natural language prompt that should trigger the skill"
    expected_skill: "plugin-name:skill-name"
    max_turns: 5
    model: sonnet

  # Any of multiple skills (OR logic)
  - name: "auth-uses-some-auth-skill"
    prompt: "authenticate with databricks"
    expected_skill_one_of:
      - "databricks-tools:databricks-authentication"
      - "fe-internal-tools:aws-authentication"
    max_turns: 3
    model: haiku

  # Multiple skills must ALL be called (AND logic)
  - name: "deploy-app-uses-multiple-skills"
    prompt: "deploy my app to the workspace I created"
    expected_skills:
      - "databricks-tools:databricks-fe-vm-workspace-deployment"
      - "databricks-tools:databricks-apps"
    max_turns: 10
    model: opus
```

**Running evals:**

```bash
cd evals
uv run skill-evals                           # Run core tests
uv run skill-evals test-cases/my-tests.yaml  # Run specific file
uv run skill-evals --dir test-cases/         # Run all YAML files in a directory
uv run skill-evals -k jira                   # Filter tests by name substring
uv run skill-evals --verbose                 # Debug output
```

### Updating the Marketplace

1. Edit `.claude-plugin/marketplace.json` to add/modify plugins
2. Each plugin needs an entry with: name, source, description, version
3. Test locally with `claude plugin marketplace add .`

## Testing Changes

### Unit tests (required before every PR)

```bash
cd evals && uv run pytest -v
```

These validate Python resource files — syntax, logic in the lakeview builder, query pretty-printer, warehouse selector, and markdown parser. No auth required. All tests must pass before submitting a PR.

### Plugin installation testing

Use `vibe local` to streamline local testing. Run it from anywhere inside your vibe repo checkout:

```bash
cd ~/code/vibe

# Register local marketplace + install a plugin + launch agent
vibe local databricks-tools

# Install multiple plugins
vibe local databricks-tools google-tools

# Register local marketplace only, no agent launch
vibe local --no-agent

# List installed plugins
claude plugin list

# Check marketplace
claude plugin marketplace list
```

## Key Files to Know

| File | Purpose |
|------|---------|
| `.claude-plugin/plugin.json` | Root manifest - registers all plugin skill/agent paths (skills silently ignored if missing) |
| `.claude-plugin/marketplace.json` | Plugin registry - defines all available plugins |
| `permissions.yaml` | Master permissions merged during install |
| `mcp-servers.yaml` | External MCP servers (Chrome DevTools, Slack, etc.) |
| `plugins/vibe-setup/skills/configure-vibe/SKILL.md` | Environment setup/validation |

## Duplicate Functionality Check - REQUIRED for New Skills, Agents, and Plugins

**Before creating any new skill, agent, or plugin, you MUST check whether similar functionality already exists in the marketplace.** This applies to all new additions — it does NOT apply to fixes or improvements to existing skills/agents/plugins.

### Why this matters

This is a shared marketplace used by hundreds of Field Engineers. Duplicate or overlapping functionality:
- Confuses users about which skill to use
- Fragments maintenance effort across redundant implementations
- Creates routing conflicts in skill invocation

### How to check

1. **Search existing skills and agents across ALL plugins** (including non-default/non-core plugins that the contributor may not have installed):
   ```bash
   # Search skill names and descriptions in all plugin manifests
   grep -r "description" plugins/*/skills/*/SKILL.md plugins/*/.claude-plugin/plugin.json
   # Search the marketplace catalog
   cat .claude-plugin/marketplace.json
   # Search agent descriptions
   grep -r "description" plugins/*/agents/*.md
   ```

2. **Search git history for similar past work** — someone may have already built and submitted something similar:
   ```bash
   # Search commit messages
   git log --oneline --all --grep="<keyword>"
   # Search PR titles
   gh pr list --state all --search "<keyword>"
   ```

3. **Identify the authors** of any similar existing functionality or past PRs:
   ```bash
   # Find who contributed the similar plugin/skill
   git log --format="%an (%ae)" -- plugins/<similar-plugin>/
   # Find PR authors for related work
   gh pr list --state all --search "<keyword>" --json number,title,author --jq '.[] | "\(.number): \(.title) by \(.author.login)"'
   ```

### What to do if overlap is found

- **Flag it to the user before proceeding.** Present what you found: the existing skill/plugin name, what it does, and who submitted it.
- **Suggest the user coordinate with the original author** — provide their GitHub username so they can reach out.
- **If the user still wants to proceed** (e.g., regional differences, different approach, personal preference), that's fine — respect their decision and continue with the PR. Document the rationale for the separate implementation in the PR description.

### Example output when overlap is found

```
⚠️ Similar functionality already exists:

  - Skill: workflows:competitive-analysis (plugins/workflows/skills/competitive-analysis/)
    Does: Compares Databricks to competitors like Snowflake, Redshift, Fabric
    Submitted by: @jane-doe (PR #234)

  - PR #301 (closed): "Add competitor comparison tool" by @john-smith

Consider coordinating with @jane-doe or @john-smith before creating a new plugin.
Do you still want to proceed with a separate implementation?
```

## Common Tasks

### Add a new skill to existing plugin
1. Create `plugins/<plugin>/skills/<skill-name>/SKILL.md`
2. Add any resource files (Python helpers, YAML configs)
3. **REQUIRED: Add `Skill(skill-name)` to `permissions.yaml`** - Every new skill MUST be added to the permissions file or users won't be able to invoke it without approval prompts
4. **REQUIRED: Add 1-2 evals (max 2 per skill)** - Core plugins: `evals/test-cases/skill-routing.yaml`. Non-core plugins: `evals/test-cases/<plugin-name>.yaml`
5. **REQUIRED: Bump the plugin version** in both `plugins/<plugin>/.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` (see [Plugin Versioning](#plugin-versioning))
6. Test with `/skill-name` in Claude Code

### Add a new agent to existing plugin
1. Create `plugins/<plugin>/agents/<agent-name>.md`
2. Define the agent's tools and instructions
3. **REQUIRED: Add agent path to root `.claude-plugin/plugin.json`** - Add the full path (e.g., `"./plugins/<plugin>/agents/<agent-name>.md"`) to the `agents` array in alphabetical order
4. Update `permissions.yaml` if needed
5. **REQUIRED: Bump the plugin version** in both `plugins/<plugin>/.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` (see [Plugin Versioning](#plugin-versioning))

### Create a new plugin
1. Create directory structure under `plugins/`
2. Add `plugin.json` in `plugins/<plugin>/.claude-plugin/`
3. **REQUIRED: Register in root `.claude-plugin/plugin.json`** - Add the plugin's skills path to the `skills` array (e.g., `"./plugins/<plugin>/skills/"`) and any agent paths to the `agents` array, in alphabetical order. **Skills and agents not listed here are silently ignored at runtime.**
4. Register in `.claude-plugin/marketplace.json`
5. Add `Skill(skill-name)` entries to `permissions.yaml` for each skill

### Modify an existing skill, agent, or resource
1. Make your changes to the relevant files
2. **REQUIRED: Bump the plugin version** in both `plugins/<plugin>/.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` (see [Plugin Versioning](#plugin-versioning))
