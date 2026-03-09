---
name: google-forms
description: Create, edit, and manage Google Forms surveys and quizzes. Collect and analyze form responses. Use this skill for ANY Google Forms operation - creating surveys, adding questions, configuring quiz settings, reading responses, or managing form settings.
---

# Google Forms Skill

Create and manage Google Forms using gcloud CLI + curl. This skill provides patterns and utilities for creating forms, adding questions, configuring settings, and reading responses.

## Authentication

**Run `/google-auth` first** to authenticate with Google Workspace, or use the shared auth module:

```bash
# Check authentication status
python3 ../google-auth/resources/google_auth.py status

# Login if needed (includes automatic retry if OAuth times out)
python3 ../google-auth/resources/google_auth.py login

# Get access token for API calls
TOKEN=$(python3 ../google-auth/resources/google_auth.py token)
```

All Google skills share the same authentication. See `/google-auth` for details on scopes and troubleshooting.

### CRITICAL: If Authentication Fails

**If the login command fails**, it means the user did NOT complete the OAuth flow in the browser.

**DO NOT:**
- Try alternative authentication methods
- Create OAuth credentials manually
- Attempt to set up service accounts

**ONLY solution:**
- Re-run `python3 ../google-auth/resources/google_auth.py login`
- The script includes automatic retry logic with clear instructions
- The user MUST click "Allow" in the browser window

### Quota Project

All API calls require a quota project header:

```bash
-H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

## Core Concepts

### Form Structure

A Google Form consists of:
- **Info**: Title and description displayed at the top
- **Items**: The questions, sections, page breaks, images, and videos in the form
- **Settings**: Quiz mode, response collection, confirmation message

### Question Types

| Type | Description | API `kind` |
|------|-------------|------------|
| Short answer | Single-line text input | `textQuestion` (paragraph: false) |
| Paragraph | Multi-line text input | `textQuestion` (paragraph: true) |
| Multiple choice | Select one option | `choiceQuestion` (type: RADIO) |
| Checkboxes | Select multiple options | `choiceQuestion` (type: CHECKBOX) |
| Dropdown | Select one from dropdown | `choiceQuestion` (type: DROP_DOWN) |
| Linear scale | Numeric scale (e.g., 1-5) | `scaleQuestion` |
| Multiple choice grid | Grid of radio buttons | `rowQuestion` with `questionGroupItem` |
| Checkbox grid | Grid of checkboxes | `rowQuestion` with `questionGroupItem` |
| Date | Date picker | `dateQuestion` |
| Time | Time picker | `timeQuestion` |
| File upload | File attachment | `fileUploadQuestion` |

### Form Settings

- **Quiz mode**: Enables point values, correct answers, and auto-grading
- **Collect email**: Requires respondent email addresses
- **Limit to 1 response**: Prevents multiple submissions per user
- **Confirmation message**: Custom message shown after submission

### Responses

- Each form submission creates a response
- Responses contain answers keyed by question ID
- Responses can be read individually or in bulk
- Response timestamps are in RFC 3339 format

## API Reference

### Base URL

```
https://forms.googleapis.com/v1
```

### Form Operations

#### Create a Form

```bash
TOKEN=$(python3 ../google-auth/resources/google_auth.py token)

curl -s -X POST "https://forms.googleapis.com/v1/forms" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "info": {
      "title": "Customer Feedback Survey"
    }
  }'
```

**Note:** Creating a form only sets the title. Use `batchUpdate` to add questions, description, and settings.

#### Get a Form

```bash
curl -s "https://forms.googleapis.com/v1/forms/${FORM_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

#### Update a Form (Batch Update)

All modifications to a form (adding questions, changing settings, updating items) use the `batchUpdate` endpoint:

```bash
curl -s -X POST "https://forms.googleapis.com/v1/forms/${FORM_ID}:batchUpdate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "requests": [
      {
        "updateFormInfo": {
          "info": {
            "description": "Please share your feedback about our product."
          },
          "updateMask": "description"
        }
      }
    ]
  }'
```

#### Add a Question

```bash
curl -s -X POST "https://forms.googleapis.com/v1/forms/${FORM_ID}:batchUpdate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}" \
  -H "Content-Type: application/json" \
  -d '{
    "requests": [
      {
        "createItem": {
          "item": {
            "title": "How satisfied are you with our product?",
            "questionItem": {
              "question": {
                "required": true,
                "choiceQuestion": {
                  "type": "RADIO",
                  "options": [
                    {"value": "Very satisfied"},
                    {"value": "Satisfied"},
                    {"value": "Neutral"},
                    {"value": "Dissatisfied"},
                    {"value": "Very dissatisfied"}
                  ]
                }
              }
            }
          },
          "location": {
            "index": 0
          }
        }
      }
    ]
  }'
```

### Response Operations

#### List Responses

```bash
curl -s "https://forms.googleapis.com/v1/forms/${FORM_ID}/responses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

#### Get a Single Response

```bash
curl -s "https://forms.googleapis.com/v1/forms/${FORM_ID}/responses/${RESPONSE_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "x-goog-user-project: ${GCP_QUOTA_PROJECT}"
```

## Helper Script

The `gforms_builder.py` script provides a convenient CLI for all Forms operations.

### Form Management

```bash
# Create a new form
python3 resources/gforms_builder.py create-form --title "Customer Feedback Survey"

# Create a form with description
python3 resources/gforms_builder.py create-form \
  --title "Customer Feedback Survey" \
  --description "Please share your thoughts about our product."

# Get form details
python3 resources/gforms_builder.py get-form --id FORM_ID

# Update form info (title/description)
python3 resources/gforms_builder.py update-info --id FORM_ID \
  --title "Updated Title" \
  --description "Updated description"

# Get the form's responder URL (the link to share with respondents)
python3 resources/gforms_builder.py get-form --id FORM_ID | jq -r '.responderUri'

# List forms (searches Drive for form files)
python3 resources/gforms_builder.py list-forms
python3 resources/gforms_builder.py list-forms --max-results 5
```

### Adding Questions

```bash
# Add a short answer question
python3 resources/gforms_builder.py add-question --id FORM_ID \
  --title "What is your name?" \
  --type short-answer \
  --required

# Add a paragraph question
python3 resources/gforms_builder.py add-question --id FORM_ID \
  --title "Describe your experience" \
  --type paragraph

# Add a multiple choice question
python3 resources/gforms_builder.py add-question --id FORM_ID \
  --title "How did you hear about us?" \
  --type multiple-choice \
  --options "Search engine" "Social media" "Friend referral" "Advertisement" "Other"

# Add a checkbox question
python3 resources/gforms_builder.py add-question --id FORM_ID \
  --title "Which features do you use?" \
  --type checkbox \
  --options "Dashboard" "Reports" "API" "Mobile app" \
  --required

# Add a dropdown question
python3 resources/gforms_builder.py add-question --id FORM_ID \
  --title "Select your department" \
  --type dropdown \
  --options "Engineering" "Sales" "Marketing" "Support" "Other"

# Add a linear scale question
python3 resources/gforms_builder.py add-question --id FORM_ID \
  --title "How likely are you to recommend us?" \
  --type scale \
  --scale-low 1 --scale-high 10 \
  --label-low "Not likely" --label-high "Very likely"

# Add a date question
python3 resources/gforms_builder.py add-question --id FORM_ID \
  --title "Preferred meeting date" \
  --type date

# Add a time question
python3 resources/gforms_builder.py add-question --id FORM_ID \
  --title "Preferred meeting time" \
  --type time

# Add a question at a specific position (0-indexed)
python3 resources/gforms_builder.py add-question --id FORM_ID \
  --title "First question" \
  --type short-answer \
  --index 0
```

### Modifying Items

```bash
# Update a question title
python3 resources/gforms_builder.py update-question --id FORM_ID \
  --item-id ITEM_ID \
  --title "Updated question text"

# Delete a question
python3 resources/gforms_builder.py delete-item --id FORM_ID \
  --item-id ITEM_ID

# Move a question to a different position
python3 resources/gforms_builder.py move-item --id FORM_ID \
  --item-id ITEM_ID \
  --index 2
```

### Form Settings

```bash
# Enable quiz mode
python3 resources/gforms_builder.py update-settings --id FORM_ID --quiz

# Set confirmation message
python3 resources/gforms_builder.py update-settings --id FORM_ID \
  --confirmation-message "Thank you for your feedback!"
```

### Reading Responses

```bash
# List all responses
python3 resources/gforms_builder.py list-responses --id FORM_ID

# Get a specific response
python3 resources/gforms_builder.py get-response --id FORM_ID \
  --response-id RESPONSE_ID

# Get response summary (count of responses per question)
python3 resources/gforms_builder.py response-summary --id FORM_ID
```

## Helper Script Reference

### gforms_builder.py Commands

| Command | Description |
|---------|-------------|
| `list-forms` | List forms from Google Drive |
| `create-form --title TITLE` | Create a new form |
| `get-form --id ID` | Get form details and structure |
| `update-info --id ID` | Update form title/description |
| `add-question --id ID --title TITLE --type TYPE` | Add a question to the form |
| `update-question --id ID --item-id ITEM_ID` | Update an existing question |
| `delete-item --id ID --item-id ITEM_ID` | Delete a form item |
| `move-item --id ID --item-id ITEM_ID --index N` | Move an item to a new position |
| `update-settings --id ID` | Update form settings (quiz mode, etc.) |
| `list-responses --id ID` | List all form responses |
| `get-response --id ID --response-id RID` | Get a specific response |
| `response-summary --id ID` | Get aggregated response summary |

### Common Options

- `--id FORM_ID` - Form ID (from the form URL or create output)
- `--title TEXT` - Question or form title
- `--type TYPE` - Question type: `short-answer`, `paragraph`, `multiple-choice`, `checkbox`, `dropdown`, `scale`, `date`, `time`
- `--options OPT1 OPT2 ...` - Choice options for multiple-choice, checkbox, or dropdown
- `--required` - Mark question as required
- `--index N` - Position index for inserting/moving items (0-indexed)
- `--scale-low N` / `--scale-high N` - Scale bounds for linear scale questions
- `--label-low TEXT` / `--label-high TEXT` - Scale endpoint labels

## Troubleshooting

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `401 Unauthorized` | Token expired or invalid | Run `/google-auth` to re-authenticate |
| `403 Forbidden` | Missing Forms scope | Re-run `/google-auth` with `--force` to get all scopes |
| `404 Not Found` | Invalid form ID | Verify the form ID from the URL or `list-forms` |
| `400 Bad Request` | Invalid request structure | Check question type and options format |

### Checking Authentication

```bash
# Verify you have Forms scopes
python3 ../google-auth/resources/google_auth.py status

# Look for:
#   https://www.googleapis.com/auth/forms.body
#   https://www.googleapis.com/auth/forms.responses.readonly
```

### Form ID from URL

Google Forms URLs look like:
```
https://docs.google.com/forms/d/FORM_ID/edit
```

Extract the `FORM_ID` portion between `/d/` and `/edit`.

## Best Practices

1. **Create form first, then add questions** - The Forms API requires creating a blank form first, then using batchUpdate to add content
2. **Use the helper script** - It handles the two-step create + update pattern automatically
3. **Set required questions** - Mark critical questions as required to ensure complete responses
4. **Include a description** - Help respondents understand the purpose of the form
5. **Use appropriate question types** - Multiple choice for single select, checkboxes for multi-select
6. **Check authentication first** - Always verify `/google-auth` before operations
7. **Share the responderUri** - Use `get-form` to retrieve the public URL for respondents

## Complete Example: Customer Feedback Survey

```bash
#!/bin/bash

# 1. Create the form
FORM=$(python3 resources/gforms_builder.py create-form \
  --title "Customer Feedback Survey" \
  --description "Help us improve our product by sharing your experience.")
FORM_ID=$(echo $FORM | jq -r '.formId')
echo "Created form: $FORM_ID"

# 2. Add questions
python3 resources/gforms_builder.py add-question --id $FORM_ID \
  --title "What is your name?" \
  --type short-answer \
  --required

python3 resources/gforms_builder.py add-question --id $FORM_ID \
  --title "How satisfied are you with our product?" \
  --type multiple-choice \
  --options "Very satisfied" "Satisfied" "Neutral" "Dissatisfied" "Very dissatisfied" \
  --required

python3 resources/gforms_builder.py add-question --id $FORM_ID \
  --title "Which features do you use most?" \
  --type checkbox \
  --options "Dashboard" "Reports" "API" "Mobile app" "Integrations"

python3 resources/gforms_builder.py add-question --id $FORM_ID \
  --title "How likely are you to recommend us?" \
  --type scale \
  --scale-low 1 --scale-high 10 \
  --label-low "Not likely" --label-high "Very likely" \
  --required

python3 resources/gforms_builder.py add-question --id $FORM_ID \
  --title "Any additional feedback?" \
  --type paragraph

# 3. Set confirmation message
python3 resources/gforms_builder.py update-settings --id $FORM_ID \
  --confirmation-message "Thank you for your feedback! We appreciate your time."

# 4. Get the shareable link
FORM_DATA=$(python3 resources/gforms_builder.py get-form --id $FORM_ID)
RESPONDER_URL=$(echo $FORM_DATA | jq -r '.responderUri')
echo "Share this link: $RESPONDER_URL"

# 5. Later - check responses
python3 resources/gforms_builder.py list-responses --id $FORM_ID
python3 resources/gforms_builder.py response-summary --id $FORM_ID
```

## Sources

- [Google Forms API Reference](https://developers.google.com/forms/api/reference/rest)
- [Forms Resource](https://developers.google.com/forms/api/reference/rest/v1/forms)
- [Responses Resource](https://developers.google.com/forms/api/reference/rest/v1/forms.responses)
- [BatchUpdate Requests](https://developers.google.com/forms/api/reference/rest/v1/forms/batchUpdate)
