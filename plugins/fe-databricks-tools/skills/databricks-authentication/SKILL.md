---
name: databricks-authentication
description: This skill provides instructions for authenticating with databricks. Do this before any databricks operations.
---

# Databricks Authentication Skill

This skill ensures you are authenticated into the correct Databricks workspace using the Databricks CLI before running any commands.

---

## Step 1 — Identify Your Workspace

Ask the user which Databricks workspace to use if not already specified. The workspace URL looks like:
```
https://<workspace-id>.azuredatabricks.net    (Azure)
https://<workspace-id>.cloud.databricks.com   (AWS)
https://<workspace-id>.gcp.databricks.com     (GCP)
```

If you need a workspace for testing or demos and don't have one, you can sign up for a free trial at https://www.databricks.com/try-databricks

---

## Step 2 — Check for Existing Profile

```bash
databricks auth profiles
```

If a valid profile exists for the target workspace (shows `YES` under the valid column), use it with `--profile=<profile-name>` in subsequent commands and skip to Step 4.

---

## Step 3 — Authenticate

If no valid profile exists, authenticate interactively:

```bash
# Replace <workspace-url> and <profile-name> with your values
databricks auth login <workspace-url> --profile=<profile-name>
```

This opens a browser for OAuth login. No Okta or SSO required for personal Databricks accounts — log in with your account email and password.

For personal access token (PAT) auth instead:
```bash
databricks configure --profile=<profile-name>
# Prompts for: workspace URL + personal access token
# Generate a PAT at: Workspace → Settings → Developer → Access tokens
```

---

## Step 4 — Verify Authentication

```bash
databricks workspace list / --profile=<profile-name>
```

If this lists files, you're authenticated correctly.

---

## Step 5 — Proceed

After authenticating, use the appropriate skill for your task:
- **Query data**: use `databricks-query`
- **Deploy resources**: use `databricks-resource-deployment`
- **Build apps**: use `databricks-apps`
- **Create demos**: use `databricks-demo`

Always pass `--profile=<profile-name>` to every CLI command to target the correct workspace.

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `auth login` opens wrong browser | Run with `--no-browser` to get a URL to paste manually |
| PAT expired | Generate new token at workspace → Settings → Developer → Access tokens |
| Profile not found | Re-run `databricks auth login` with correct workspace URL |
| Workspace URL wrong format | Must be full URL with `https://`, no trailing slash |
