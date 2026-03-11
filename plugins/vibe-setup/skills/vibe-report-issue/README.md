# Vibe Report Issue

Reports issues with vibe skills, agents, or behavior by collecting context from the current conversation and filing a GitHub issue.

## How to Invoke

### Slash Command

```
/vibe-report-issue
```

### Example Prompts

```
"Report a bug with the skill I just used"
"File an issue -- vibe invoked the wrong skill"
"Something went wrong, I want to report this"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| GitHub CLI | `gh` must be installed and authenticated (`gh auth login`) |

## What This Skill Does

1. Summarizes the current conversation (skills invoked, tool calls made, outcome) while sanitizing sensitive data
2. Asks the user to confirm the summary and provide issue details (type, expected vs actual behavior)
3. Creates a GitHub issue in the vibe repository with the sanitized summary, issue type, and user description
4. Returns the issue URL to the user

## Related Skills

- `/configure-vibe` - Set up `gh` CLI if not installed
