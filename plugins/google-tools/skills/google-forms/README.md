# Google Forms

Create, edit, and manage Google Forms surveys, quizzes, and feedback forms. Add questions of any type, configure form settings, and read/analyze responses.

## How to Invoke

### Slash Command

```
/google-forms
```

### Example Prompts

```
"Create a customer feedback survey with rating and comment questions"
"Add a multiple choice question to my form about preferred contact method"
"Show me the responses from my team satisfaction survey"
```

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Google Auth | Run `/google-auth` first to authenticate with Google Workspace |
| gcloud CLI | Must be installed (`brew install --cask google-cloud-sdk`) |
| Quota Project | Uses your GCP quota project (`$GCP_QUOTA_PROJECT`) for API billing |

## What This Skill Does

1. Creates Google Forms with titles and descriptions
2. Adds questions of any type (short answer, paragraph, multiple choice, checkbox, dropdown, scale, date, time)
3. Configures form settings (quiz mode, confirmation messages)
4. Reads and summarizes form responses
5. Updates and reorders existing form questions
6. Lists forms from Google Drive

## Key Resources

| File | Description |
|------|-------------|
| `resources/gforms_builder.py` | Complete form operations (create-form, add-question, list-responses, response-summary, update-settings) |

## Related Skills

- `/google-auth` - Required authentication before using Forms
- `/google-sheets` - Export form responses to Sheets for analysis
- `/google-docs` - Reference form results in documents
