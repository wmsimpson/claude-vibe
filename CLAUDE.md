# Claude Vibe

Personal Claude Code plugin toolkit providing skills and agents for Databricks, Google Workspace, JIRA, diagrams, and workflow automation.

## Project Structure

```
claude-vibe/
├── bin/                          Shell CLI entry point
├── lib/                          Shell helpers (tty, profiles)
├── steps/                        8-step installer scripts
├── cli/                          Go CLI (TUI, doctor, configure, profiles)
├── evals/                        Python eval/test framework
├── plugins/                      Plugin collections
│   ├── app-dev/                  Mobile and web app development
│   ├── databricks-tools/         Databricks integration
│   ├── google-tools/             Google Workspace (Docs, Sheets, Gmail, etc.)
│   ├── jira-tools/               JIRA ticket management
│   ├── lean-sigma-tools/         Lean Six Sigma process tools
│   ├── macos-scheduler/          macOS launchd task scheduler
│   ├── specialized-agents/       Mermaid diagrams, web dev testing
│   ├── vibe-setup/               Environment setup and validation
│   └── workflows/                Workflow automation
├── scripts/                      Linting and utility scripts
├── docs/                         Developer documentation
├── .claude-plugin/               Plugin manifest and marketplace
├── permissions.yaml              Master permissions config
└── mcp-servers.yaml              MCP server configs
```

## Development Guidelines

### Adding Skills

1. Skills are markdown files in `plugins/<plugin-name>/skills/<skill-name>/SKILL.md`
2. Use YAML frontmatter for metadata
3. Register the plugin's skills path in `.claude-plugin/plugin.json` if new plugin

### Adding Agents

1. Agents are markdown files in `plugins/<plugin-name>/agents/<agent-name>.md`
2. Add the agent path to `.claude-plugin/plugin.json` agents array

### Testing

- Go CLI: `cd cli && go test ./... -v`
- Evals: `cd evals && uv run pytest -v`
- Skill size check: `bash scripts/check-skill-sizes.sh`

### Building the CLI

```bash
cd cli
go build -o ~/.local/bin/vibe ./cmd/vibe/
```

### Google Docs Output Pattern

Use the markdown converter — do NOT manually construct batch updates:

```bash
python3 plugins/google-tools/skills/google-docs/resources/markdown_to_gdocs.py \
  --input /tmp/output.md --title "Document Title"
```
