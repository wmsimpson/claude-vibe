# Vibe Profile

Instructions for configuring and customizing your vibe profile.

## Overview

The vibe profile (`~/.vibe/profile`) stores your personal info, accounts, projects, channels, and preferences. It can be:
- **Auto-generated** from available data sources (Slack MCP, GitHub, etc.)
- **Manually configured** by editing the YAML file directly
- **Customized** by asking Claude to add/remove/update entries

## Building a Profile

Generate a profile file in YAML format. Try to infer inputs from available tools/MCPs if configured — like Slack. Otherwise build from user-provided information.

## Customizing a Profile

Ask Claude to make specific changes:

### Example Customization Requests

- "Add a new project called my-app to my profile"
- "Update my GitHub username"
- "Add the Slack channel #dev-team to my profile"
- "Set my preferred timezone to America/New_York"
- "Add Jane Doe to my recent contacts"

### How Customizations Work

1. The agent reads your existing profile
2. Makes the requested changes
3. Records the change in the `customizations` section
4. Writes the updated profile

## Profile Structure

```yaml
# Vibe User Profile
# Generated: <timestamp>
# Regenerate by asking Claude to rebuild your vibe profile

user:
  name: Your Name
  email: you@example.com
  username: your-github-username
  role: developer  # developer, designer, manager, etc.
  title: Software Engineer
  location: San Francisco, CA
  timezone: America/Los_Angeles
  github_username: your-github-username
  running_todo_doc_url: https://docs.google.com/YOUR_DOC_URL  # or null

# Active projects (personal or professional)
projects:
  - name: my-mobile-app
    type: mobile  # mobile, web, api, library
    repo: https://github.com/your-username/my-mobile-app
    tech_stack:
      - React Native
      - Expo
      - TypeScript
    status: active
    notes: "iOS/Android app for personal finance tracking"

  - name: my-web-app
    type: web
    repo: https://github.com/your-username/my-web-app
    tech_stack:
      - React
      - Node.js
      - Tailwind CSS
    status: active
    deployed_url: https://my-web-app.vercel.app

# Team members / contacts
contacts:
  - name: Jane Doe
    email: jane@example.com
    role: Designer
    github: janedoe
    last_interaction: 2026-01-15

  - name: Bob Smith
    email: bob@example.com
    role: Backend Engineer
    github: bobsmith
    last_interaction: 2026-01-10

# Slack configuration (if Slack MCP is configured)
slack:
  workspace: your-workspace.slack.com
  channels:
    - name: dev-team
      id: C0XXXXXXX  # IMPORTANT: Always include channel ID
      description: Main dev team channel
    - name: project-updates
      id: C0YYYYYYY
      description: Project status updates

# Integration credentials (do NOT store secrets here — use env vars)
integrations:
  github:
    username: your-github-username
    # token: set as GITHUB_TOKEN env var
  gcp:
    # quota_project: set as GCP_QUOTA_PROJECT env var
  databricks:
    # profile: set in ~/.databrickscfg
  vercel:
    # token: set as VERCEL_TOKEN env var

# User preferences for plugin behavior
preferences:
  default_project: my-mobile-app  # default project context for skills
  code_style: typescript          # preferred language/style
  test_framework: jest
  deployment_target: vercel       # vercel, netlify, firebase, etc.
  mobile_platform: react-native   # react-native, expo, swift, flutter

# Metadata
generated_at: 2026-01-01T00:00:00Z
data_sources:
  - manual
  - github
  - slack

notes:
  - Auto-generated notes about missing data
  - Information that couldn't be discovered

# Manual customizations (audit trail)
customizations:
  - date: 2026-01-01
    action: added
    type: project
    details: "Added my-mobile-app project"
  - date: 2026-01-02
    action: updated
    type: field
    details: "Updated deployment_target to vercel"
```

## Customizations Section

The `customizations` section provides an audit trail of manual changes. Each entry includes:

| Field | Description |
|-------|-------------|
| `date` | When the change was made (YYYY-MM-DD) |
| `action` | Type of action: `added`, `removed`, or `updated` |
| `type` | What was changed: `project`, `contact`, `channel`, or `field` |
| `details` | Human-readable description of the change |

## Channel ID Discovery

When adding Slack channels, the agent will automatically discover the channel ID using the Slack MCP. Channel IDs are required for all channel entries (format: `C0XXXXXXX`).

If a channel cannot be found, the agent will ask you to provide the ID manually.

## Preferences Section

The `preferences` section stores user preferences for plugin behavior. Skills check this section to customize their output. See individual skill documentation for available preferences.

## Security Notes

- **Never store API tokens or passwords** in the profile file
- Use environment variables for secrets (GITHUB_TOKEN, VERCEL_TOKEN, etc.)
- The profile is stored at `~/.vibe/profile` — not committed to git
