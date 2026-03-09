---
name: showcase
description: Document and share a Vibe session as a showcase writeup. Creates a Google Doc saved to a folder of your choice. Use this skill when the user says "showcase", "writeup", "document this session", or wants to share what they accomplished with Vibe.
user-invocable: true
---

# Vibe Showcase Skill

This skill guides you through documenting a Vibe session as a showcase writeup and saving it as a Google Doc. Showcases help you and others learn from real-world Vibe use cases.

## Quick Start

This skill is designed to be run **during or at the end of a Vibe session**, or by resuming a previous session. You have full access to the conversation history — use it. Do NOT ask the user what they accomplished or what tools they used. Extract everything directly from the conversation context:

1. **What was accomplished** - The task, deliverables, and time saved
2. **Prompting strategy** - The key prompts the user gave, in order, with "why this works" explanations
3. **Tools and skills invoked** - Table of tools/skills, their purpose, and how they were used
4. **Workflow architecture** - How data flowed from inputs through processing to outputs
5. **Key learnings and gotchas** - What worked, what didn't, tips for others

## Workflow

### Step 1: Analyze the Conversation

Review the full conversation history to extract:

- **User prompts** - The actual messages the user sent, especially the key ones that shaped the workflow. Quote these verbatim using `>` blockquotes.
- **Tools and skills used** - Every `/skill` invocation, MCP tool call, web search, parallel agent, etc.
- **Deliverables** - Google Docs, Sheets, dashboards, or other artifacts produced (extract URLs from tool results).
- **Workflow structure** - How the work was parallelized, what data sources were combined, what the pipeline looked like.
- **Gotchas encountered** - Any errors, retries, or workarounds that other SAs should know about.

Do not ask the user for this information — it is all in the conversation.

### Step 2: Draft the Showcase Markdown

Use the template in `resources/SHOWCASE_TEMPLATE.md` as the structural reference. Write the showcase markdown to `/tmp/showcase_<slug>.md` where `<slug>` is a short descriptive name (e.g., `showcase_ai_account_tiering.md`).

Key formatting requirements:
- Use `>` blockquotes for actual prompts the user gave to Vibe
- Include **"Why this works"** explanations after each prompt step
- Use tables for tools/skills and results summaries
- Use code blocks for workflow architecture diagrams
- Include links to output artifacts where available

### Step 3: Convert to Google Doc

Use the `markdown_to_gdocs.py` script from the google-docs skill to create the Google Doc:

```bash
python3 ~/.claude/plugins/cache/claude-vibe/fe-google-tools/*/skills/google-docs/resources/markdown_to_gdocs.py \
  --input /tmp/showcase_<slug>.md \
  --title "Vibe Showcase: <Title>"
```

### Step 4: Move to a Folder (Optional)

If the user wants the Google Doc saved to a specific Google Drive folder, ask them where to save it (or if they have a preferred folder ID). Then move it using the Drive API:

```bash
TOKEN=$(python3 ~/.claude/plugins/cache/claude-vibe/fe-google-tools/*/skills/google-auth/resources/google_auth.py token)

curl -s -X PATCH \
  "https://www.googleapis.com/drive/v3/files/<DOC_ID>?addParents=<FOLDER_ID>&fields=id,name,parents" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

If the user has not specified a folder, leave the doc in the default location (My Drive root) and share the link directly.

### Step 5: Return Result

Return:
1. The Google Doc URL
2. Confirmation of where it was saved (folder if specified, or My Drive)
3. A brief summary of the showcase

## Example Invocations

```
User: /showcase
-> Analyzes conversation history, drafts writeup from context, creates Google Doc

User: Write up what we just did as a showcase
-> Same - extracts everything from the session, creates formatted Google Doc

User: Document this Vibe session for the showcase library
-> Same - no questions asked, just analyzes and publishes
```

## Resources

- `resources/SHOWCASE_TEMPLATE.md` - Template structure for showcase writeups
- `fe-google-tools/skills/google-docs/resources/markdown_to_gdocs.py` - Markdown to Google Docs converter
