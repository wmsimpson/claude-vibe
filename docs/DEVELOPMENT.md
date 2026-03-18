# claude-vibe — Development Tracker

> Auto-updated after each work session. Tracks completed work, active bugs, and next steps.

---

## Current State (2026-02-26)

**Overall Status:** 🟢 Personal-ready — chrome-first setup, self-sufficient Google auth, full mobile/web dev support

| Layer | Status | Notes |
|-------|--------|-------|
| Plugin cleanup (disabled internal plugins) | ✅ Done | 10 plugins → `plugins/_disabled/` |
| Config files (mcp-servers, permissions, manifests) | ✅ Done | All cleaned and commented |
| configure-vibe skill | ✅ Done | Chrome/chrome-devtools as Step 1; personal Gmail sign-in; GCP walkthrough |
| validate-mcp-access skill | ✅ Done | 5-phase active self-diagnostic; uses chrome-devtools for visual debugging |
| google_auth.py quota project fix | ✅ Done | Was hardcoded gcp-sandbox-field-eng; now reads GCP_QUOTA_PROJECT env var |
| google_api_utils.py quota project fix | ✅ Done | Same fix; x-goog-user-project header conditional |
| jira-actions skill | ✅ Done | Generalized, ES-specific kept as reference |
| databricks-tools (internal deployment removed) | ✅ Done | vm/oneenv → `skills/_disabled/` |
| Go CLI module path updated | ✅ Done | `will-simpson/claude-vibe` |
| CLAUDE.md + README.md rewritten | ✅ Done | Personal/non-enterprise docs |
| workflows generalization | ✅ Done | showcase, support-escalation, draft-rca, security-questionnaire, fe-account-transition, slack-discovery-agent, rca-doc |
| GitHub repo created | ✅ Done | https://github.com/will-simpson_data/claude-vibe (private) |
| Go 1.26.0 installed | ✅ Done | `/opt/homebrew/bin/go` |
| CLI built from source | ✅ Done | `vibe dev` at `~/.local/bin/vibe` |
| vibe install (local) | ✅ Done | `vibe local` + manual permissions/MCP merge |
| All 9 plugins installed | ✅ Done | All enabled in Claude Code (8 original + app-dev) |
| Permissions cleaned | ✅ Done | 54 skill permissions; 4 new app-dev skills added |
| MCP servers cleaned | ✅ Done | Only chrome-devtools active; old Databricks servers removed |
| **macos-scheduler test** | ✅ Done | Lists tasks correctly ("No scheduled jobs found") |
| **Google tools test** | ✅ Done | Created Google Doc via markdown_to_gdocs.py |
| **Databricks tools test** | ✅ Done | SQL via REST API works; logfood profile valid |
| **Mermaid diagrams test** | ✅ Done | mmdc v11.12.0 generates PNG flowchart |
| **fe-architecture-diagram test** | ✅ Done | graphviz + Python diagrams package both working |
| **lean-sigma-tools test** | ⚠️ Partial | Sheets API blocked — needs personal GCP quota project (see B10) |
| **Internal references cleanup** | ✅ Done | gcp-sandbox-field-eng, @databricks.com emails, go/ shortlinks, FE_ACCESS_WORKFLOW all generalized |
| **app-dev plugin created** | ✅ Done | app-setup, github-workflow, web-deployment, app-debug skills |
| **configure-vibe rewritten (v2)** | ✅ Done | Chrome Step 1, personal Gmail, GCP project walkthrough, mobile/web tools |
| **validate-mcp-access rewritten** | ✅ Done | Active 5-phase diagnostic; chrome-devtools visual debug; per-API status |
| **README rewritten** | ✅ Done | Personal developer focus, correct install path (~/.local/bin), env vars table |

---

## Architecture Overview

```
claude-vibe/
├── cli/                         ← Go source — builds vibe binary
│   ├── cmd/vibe/main.go
│   └── internal/
│       ├── install/             ← vibe install logic
│       ├── marketplace/         ← plugin download/sync
│       ├── config/              ← settings, profiles, permissions
│       ├── tui/                 ← Bubble Tea terminal UI
│       └── sync/                ← MCP server sync
├── plugins/                     ← Active plugins
│   ├── databricks-tools/     ← Databricks workspace, apps, queries
│   ├── google-tools/         ← Gmail, Docs, Sheets, Slides, Drive
│   ├── jira-tools/           ← JIRA (any instance, MCP required)
│   ├── specialized-agents/   ← Mermaid diagrams, web devloop
│   ├── vibe-setup/           ← configure-vibe, validate-mcp, profile
│   ├── workflows/            ← Architecture, competitive, RCA, POC, etc.
│   ├── macos-scheduler/      ← macOS launchd scheduler
│   ├── lean-sigma-tools/        ← SIPOC, FMEA, process maps
│   ├── mcp-servers/          ← MCP server framework (future)
│   └── shared-resources/     ← Shared Python utilities
├── plugins/_disabled/           ← Archived (Databricks-internal, ref only)
│   ├── fe-salesforce-tools/
│   ├── fe-internal-tools/
│   ├── fe-file-expenses/
│   ├── fe-ssagent/
│   ├── fe-financialforce-tools/
│   ├── fe-dnb-hunting/
│   ├── fe-ee/
│   ├── fe-manager/
│   ├── fe-meeting-notes/
│   └── fe-social-media-tools/
├── .claude-plugin/
│   ├── plugin.json              ← Root manifest (active skills + agents)
│   └── marketplace.json         ← Plugin registry
├── permissions.yaml             ← Merged into ~/.claude/settings.json
├── mcp-servers.yaml             ← Merged into ~/.claude.json (chrome-devtools active)
├── build.sh                     ← Local build script
└── docs/
    ├── DEVELOPMENT.md           ← This file
    └── generate_mindmap.py      ← Visual mindmap generator
```

### How Skills Load

```
vibe install
  → reads permissions.yaml   → merges into ~/.claude/settings.json
  → reads mcp-servers.yaml   → merges into ~/.claude.json
  → reads plugin.json        → Claude Code loads skill .md files

Claude Code session
  → user prompt triggers skill routing
  → Skill tool invoked with skill name
  → skill SKILL.md content used as prompt context
```

---

## Completed Work Log

### 2026-02-26 — Initial generalization from databricks-field-eng/vibe

**Config layer:**
- `mcp-servers.yaml`: Commented out glean, slack, jira, confluence (Databricks internal pex files). Added public replacement references (mcp-atlassian, @modelcontextprotocol/server-slack).
- `permissions.yaml`: Removed 25+ internal skill permissions. Internal skills kept as commented reference.
- `.claude-plugin/plugin.json`: Reduced from 17 skills paths to 8 active.
- `.claude-plugin/marketplace.json`: Reduced from 20+ plugins to 10 active.

**Plugins disabled (→ plugins/_disabled/):**
- fe-salesforce-tools, fe-internal-tools, fe-file-expenses, fe-ssagent, fe-financialforce-tools, fe-dnb-hunting, fe-ee, fe-manager, fe-meeting-notes, fe-social-media-tools

**Plugins modified:**
- vibe-setup/configure-vibe: Removed steps 4 (Databricks CLI), 5 (Salesforce CLI), 8 (Okta/AWS config download)
- vibe-setup/validate-mcp-access: Fully rewritten — removed logfood workspace dependency
- jira-tools/jira-actions: Generalized from ES-ticket-only to any JIRA instance; ES fields in collapsible reference
- databricks-tools: Moved databricks-fe-vm-workspace-deployment and databricks-oneenv-workspace-deployment to skills/_disabled/

**CLI:**
- Go module path: `github.com/databricks-field-eng/vibe/cli` → `github.com/will-simpson/claude-vibe/cli`
- DefaultRepo: `databricks-field-eng/vibe` → `will-simpson/claude-vibe`
- Added `build.sh` local build script
- Removed `.github/workflows/` (Databricks-internal CI/CD)

**Documentation:**
- CLAUDE.md: Rewritten for personal use
- README.md: Rewritten with local build instructions
- Google Doc design doc created and updated: https://docs.google.com/document/d/1pWq-VSLG3wC8Zo94cIPCyE17eRfHuEp6s2d11u3M__o/edit

---

## Active Bugs / Issues

| # | Issue | Severity | Status |
|---|-------|----------|--------|
| B1 | ~~Go not installed~~ | High | ✅ Fixed — Go 1.26.0 installed |
| B2 | Jamf MDM blocks running any binary at `/usr/local/bin/vibe` — kills with SIGKILL | High | ✅ Worked Around — installed to `~/.local/bin/vibe` (takes PATH precedence) |
| B3 | `vibe local` fails with "FAILED" when run non-interactively | Medium | ✅ Worked Around — used `claude plugin marketplace add` + `claude plugin install` directly |
| B4 | `vibe sync` only syncs to Cursor/Codex, does NOT merge permissions.yaml into Claude settings | Medium | ✅ Fixed — manually merged using yq + jq (same logic as step_permissions.go) |
| B5 | Old Databricks skill permissions lingered in settings.json after `vibe sync` | Low | ✅ Fixed — clean replace instead of append |
| B6 | Old Databricks MCP servers (glean, jira, confluence, slack) persisted in `~/.claude.json` | Low | ✅ Fixed — removed, only chrome-devtools remains |
| B7 | `rg` (ripgrep) aliased through Claude Code binary — may not work in non-Claude shells | Low | Open |
| B8 | Go module path set to `will-simpson/claude-vibe` but personal non-EMU GitHub account TBD | Low | Open — will update when pushing to personal account |
| B9 | Stop hook in `~/.claude/settings.json` runs `vibe telemetry publish` — uses our new vibe (OK, command exists) | Low | Monitor — confirm on next session stop |
| B10 | Google Sheets API 403 on personal machine — `google_api_utils.py` defaults quota project to `gcp-sandbox-field-eng`; no personal GCP project configured | Medium | Fix: `gcloud auth application-default set-quota-project <your-project>` or set `GCP_QUOTA_PROJECT` env var. Google Docs API works fine (doesn't require quota project header). |
| B11 | lean-sigma-sipoc SKILL.md hardcodes cache path `fe-vibe/` in example code blocks (cosmetic only — skill dynamically locates scripts at runtime) | Low | Open — update example paths to use `claude-vibe/` in next skill edit |

---

## Next Steps (Ordered)

### Immediate (functional testing)

- [x] **N1** — Install Go: `brew install go` ✅
- [x] **N2** — Build vibe CLI: `~/.local/bin/vibe version` → `vibe dev` ✅
- [x] **N3** — All 8 plugins installed and enabled ✅
- [x] **N4** — 50 clean skill permissions in Claude settings ✅
- [x] **N5** — chrome-devtools MCP active, old Databricks MCPs removed ✅
- [ ] **N6** — **TEST: configure-vibe** — run in fresh terminal, verify no Databricks prompts
- [x] **N7** — **TEST: google-tools** — ✅ Created test doc via `markdown_to_gdocs.py`; URL: https://docs.google.com/document/d/1blQZM0_UX6n_7R9Eyd0Kapv-xfkxSapEart2mYiDAAQ/edit
- [x] **N8** — **TEST: databricks-tools** — ✅ SQL via `databricks api post /api/2.0/sql/statements/` against logfood; note: `databricks sql` subcommand does NOT exist in v0.253.0 (use REST API as skill docs correctly specify)
- [x] **N9** — **TEST: mermaid-diagrams** agent — ✅ Generated `/tmp/test_flowchart.png` using mmdc v11.12.0
- [x] **N10** — **TEST: workflows:fe-architecture-diagram** — ✅ graphviz 14.1.2 installed (`/opt/homebrew/bin/dot`); Python `diagrams` package in `~/.vibe/diagrams/.venv`; both working
- [x] **N11** — **TEST: lean-sigma-tools** — ⚠️ Partial pass: DOT graphviz rendering works; Google Sheets API blocked by missing quota project (see B10). Fix: configure personal GCP project.
- [x] **N12** — **TEST: macos-scheduler** — ✅ `scheduler_manager.py list` works; reports "No scheduled jobs found"

### Short-term (post-build validation)

- [x] **N13** — Generalize all internal references (gcp-sandbox, @databricks.com, go/ shortlinks) ✅
- [x] **N14** — Create app-dev plugin (app-setup, github-workflow, web-deployment, app-debug) ✅
- [x] **N15** — Rewrite configure-vibe for personal use (Chrome Step 1, personal Gmail, GCP walkthrough, mobile tools) ✅
- [x] **N16** — Rewrite README with personal developer focus + correct install instructions ✅
- [x] **N28** — Fix google_auth.py hardcoded quota project (second instance, missed in N13) ✅
- [x] **N29** — Rewrite validate-mcp-access as active 5-phase self-diagnostic with chrome-devtools ✅
- [ ] **N17** — Set up personal GCP quota project → fix lean-sigma-sipoc Google Sheets (B10)
- [ ] **N18** — Decide personal GitHub account for final push (non-EMU)
- [ ] **N19** — Update Go module path to match final GitHub username
- [ ] **N20** — Push to personal non-EMU GitHub

### Future (MCP server buildout + integrations)

- [ ] **N21** — Enable Slack MCP: `@modelcontextprotocol/server-slack`, configure bot token
- [ ] **N22** — Enable JIRA MCP: `mcp-atlassian`, configure Atlassian API token
- [ ] **N23** — Enable Confluence MCP: configure alongside JIRA
- [ ] **N24** — Add Firebase integration skill (app-setup → Firebase backend)
- [ ] **N25** — Add Supabase integration skill
- [ ] **N26** — Add App Store Connect workflow (automated TestFlight distribution)
- [ ] **N27** — Add Expo EAS workflow to github-workflow CI/CD templates

---

## Environment State

| Tool | Version | Status |
|------|---------|--------|
| macOS | Darwin 24.6.0 | ✅ |
| Homebrew | 5.0.14 | ✅ |
| Claude Code | 2.1.59 | ✅ |
| vibe CLI | dev (personal build, `~/.local/bin/vibe`) | ✅ |
| Go | 1.26.0 | ✅ |
| Node.js | v25.6.1 | ✅ |
| npm | present | ✅ |
| gcloud | 557.0.0 | ✅ |
| jq | 1.7.1 | ✅ |
| yq | v4.52.4 | ✅ |
| uv | 0.10.4 | ✅ |
| ripgrep | 14.1.1 | ✅ |
| Databricks CLI | v0.253.0 | ✅ |

---

*Last updated: 2026-02-26 (session 3) by Claude Sonnet 4.6*
