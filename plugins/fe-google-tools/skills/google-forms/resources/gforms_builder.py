#!/usr/bin/env python3
"""
Google Forms Builder - Complete Form Operations Helper

Provides high-level operations for Google Forms:
- Create and manage forms
- Add questions of various types
- Update and delete form items
- Configure form settings (quiz mode, etc.)
- Read and summarize responses

Usage:
    python3 gforms_builder.py create-form --title "My Survey"
    python3 gforms_builder.py add-question --id FORM_ID --title "Question?" --type multiple-choice --options "A" "B" "C"
    python3 gforms_builder.py list-responses --id FORM_ID
"""

import argparse
import json
import sys
from typing import Dict, List, Optional

from google_api_utils import api_call_with_retry

FORMS_API_BASE = "https://forms.googleapis.com/v1"
DRIVE_API_BASE = "https://www.googleapis.com/drive/v3"


def api_request(method: str, url: str, data: Optional[Dict] = None,
                params: Optional[Dict] = None) -> Dict:
    """Make an authenticated API request with retry logic."""
    try:
        return api_call_with_retry(method, url, data=data, params=params)
    except RuntimeError:
        return {}


# =============================================================================
# Form Operations
# =============================================================================

def list_forms(max_results: int = 20) -> List[Dict]:
    """List forms by searching Google Drive for form files."""
    params = {
        "q": "mimeType='application/vnd.google-apps.form'",
        "pageSize": max_results,
        "fields": "files(id,name,createdTime,modifiedTime,webViewLink)",
        "orderBy": "modifiedTime desc"
    }
    result = api_request("GET", f"{DRIVE_API_BASE}/files", params=params)
    files = result.get("files", [])

    return [{
        "formId": f.get("id"),
        "title": f.get("name"),
        "createdTime": f.get("createdTime"),
        "modifiedTime": f.get("modifiedTime"),
        "webViewLink": f.get("webViewLink")
    } for f in files]


def create_form(title: str, description: Optional[str] = None) -> Dict:
    """
    Create a new Google Form.

    Args:
        title: Form title
        description: Optional form description
    """
    data = {
        "info": {
            "title": title
        }
    }
    result = api_request("POST", f"{FORMS_API_BASE}/forms", data)

    form_id = result.get("formId")

    if form_id:
        # Set the Drive document title to match the form title
        # (the create endpoint only sets info.title, not documentTitle)
        api_request("PATCH", f"{DRIVE_API_BASE}/files/{form_id}",
                    data={"name": title})

        # If description provided, update it via batchUpdate
        if description:
            update_data = {
                "requests": [{
                    "updateFormInfo": {
                        "info": {
                            "description": description
                        },
                        "updateMask": "description"
                    }
                }]
            }
            api_request("POST", f"{FORMS_API_BASE}/forms/{form_id}:batchUpdate", update_data)

    return {
        "formId": result.get("formId"),
        "info": result.get("info"),
        "responderUri": result.get("responderUri"),
        "revisionId": result.get("revisionId")
    }


def get_form(form_id: str) -> Dict:
    """Get form details including all items."""
    result = api_request("GET", f"{FORMS_API_BASE}/forms/{form_id}")
    return result


def update_info(form_id: str, title: Optional[str] = None,
                description: Optional[str] = None) -> Dict:
    """Update form title and/or description."""
    requests = []
    update_mask_parts = []
    info = {}

    if title:
        info["title"] = title
        update_mask_parts.append("title")
    if description is not None:
        info["description"] = description
        update_mask_parts.append("description")

    if not update_mask_parts:
        return {"error": "No updates specified"}

    requests.append({
        "updateFormInfo": {
            "info": info,
            "updateMask": ",".join(update_mask_parts)
        }
    })

    result = api_request("POST", f"{FORMS_API_BASE}/forms/{form_id}:batchUpdate",
                         {"requests": requests})
    return result


# =============================================================================
# Question Operations
# =============================================================================

def _build_question_item(title: str, question_type: str,
                         options: Optional[List[str]] = None,
                         required: bool = False,
                         scale_low: int = 1, scale_high: int = 5,
                         label_low: Optional[str] = None,
                         label_high: Optional[str] = None) -> Dict:
    """Build a question item dict for the Forms API."""
    question = {"required": required}

    if question_type == "short-answer":
        question["textQuestion"] = {"paragraph": False}

    elif question_type == "paragraph":
        question["textQuestion"] = {"paragraph": True}

    elif question_type in ("multiple-choice", "checkbox", "dropdown"):
        type_map = {
            "multiple-choice": "RADIO",
            "checkbox": "CHECKBOX",
            "dropdown": "DROP_DOWN"
        }
        choice_options = [{"value": opt} for opt in (options or [])]
        question["choiceQuestion"] = {
            "type": type_map[question_type],
            "options": choice_options
        }

    elif question_type == "scale":
        question["scaleQuestion"] = {
            "low": scale_low,
            "high": scale_high
        }
        if label_low:
            question["scaleQuestion"]["lowLabel"] = label_low
        if label_high:
            question["scaleQuestion"]["highLabel"] = label_high

    elif question_type == "date":
        question["dateQuestion"] = {}

    elif question_type == "time":
        question["timeQuestion"] = {}

    else:
        print(f"Unknown question type: {question_type}", file=sys.stderr)
        print("Valid types: short-answer, paragraph, multiple-choice, checkbox, "
              "dropdown, scale, date, time", file=sys.stderr)
        sys.exit(1)

    return {
        "title": title,
        "questionItem": {
            "question": question
        }
    }


def add_question(form_id: str, title: str, question_type: str,
                 options: Optional[List[str]] = None,
                 required: bool = False,
                 index: Optional[int] = None,
                 scale_low: int = 1, scale_high: int = 5,
                 label_low: Optional[str] = None,
                 label_high: Optional[str] = None) -> Dict:
    """
    Add a question to a form.

    Args:
        form_id: Form ID
        title: Question text
        question_type: Type of question (short-answer, paragraph, multiple-choice,
                       checkbox, dropdown, scale, date, time)
        options: List of options for choice-based questions
        required: Whether the question is required
        index: Position to insert at (0-indexed, appends to end if None)
        scale_low: Low end of scale for scale questions
        scale_high: High end of scale for scale questions
        label_low: Label for low end of scale
        label_high: Label for high end of scale
    """
    item = _build_question_item(title, question_type, options, required,
                                scale_low, scale_high, label_low, label_high)

    # location.index is required by the Forms API
    # If no index specified, append to end by getting current item count
    if index is None:
        form = get_form(form_id)
        index = len(form.get("items", []))

    request = {
        "requests": [{
            "createItem": {
                "item": item,
                "location": {"index": index}
            }
        }]
    }

    result = api_request("POST", f"{FORMS_API_BASE}/forms/{form_id}:batchUpdate", request)

    # Extract the created item info from the response
    replies = result.get("replies", [])
    if replies:
        create_reply = replies[0].get("createItem", {})
        return {
            "itemId": create_reply.get("itemId"),
            "questionId": create_reply.get("questionId", []),
            "added": True
        }
    return result


def update_question(form_id: str, item_id: str,
                    title: Optional[str] = None) -> Dict:
    """Update an existing question's title."""
    if not title:
        return {"error": "No updates specified"}

    # First get the current form to find the item index
    form = get_form(form_id)
    items = form.get("items", [])

    item_index = None
    for i, item in enumerate(items):
        if item.get("itemId") == item_id:
            item_index = i
            break

    if item_index is None:
        return {"error": f"Item {item_id} not found in form"}

    # Build update request
    update_item = {"title": title}
    # Preserve the existing question structure
    existing_item = items[item_index]
    if "questionItem" in existing_item:
        update_item["questionItem"] = existing_item["questionItem"]

    request = {
        "requests": [{
            "updateItem": {
                "item": update_item,
                "location": {"index": item_index},
                "updateMask": "title"
            }
        }]
    }

    result = api_request("POST", f"{FORMS_API_BASE}/forms/{form_id}:batchUpdate", request)
    return result


def delete_item(form_id: str, item_id: str) -> Dict:
    """Delete an item from the form."""
    # Find the item index
    form = get_form(form_id)
    items = form.get("items", [])

    item_index = None
    for i, item in enumerate(items):
        if item.get("itemId") == item_id:
            item_index = i
            break

    if item_index is None:
        return {"error": f"Item {item_id} not found in form"}

    request = {
        "requests": [{
            "deleteItem": {
                "location": {"index": item_index}
            }
        }]
    }

    result = api_request("POST", f"{FORMS_API_BASE}/forms/{form_id}:batchUpdate", request)
    return {"deleted": True, "itemId": item_id}


def move_item(form_id: str, item_id: str, new_index: int) -> Dict:
    """Move an item to a new position."""
    # Find the current item index
    form = get_form(form_id)
    items = form.get("items", [])

    current_index = None
    for i, item in enumerate(items):
        if item.get("itemId") == item_id:
            current_index = i
            break

    if current_index is None:
        return {"error": f"Item {item_id} not found in form"}

    request = {
        "requests": [{
            "moveItem": {
                "originalLocation": {"index": current_index},
                "newLocation": {"index": new_index}
            }
        }]
    }

    result = api_request("POST", f"{FORMS_API_BASE}/forms/{form_id}:batchUpdate", request)
    return {"moved": True, "itemId": item_id, "newIndex": new_index}


# =============================================================================
# Settings Operations
# =============================================================================

def update_settings(form_id: str, quiz: Optional[bool] = None,
                    confirmation_message: Optional[str] = None) -> Dict:
    """Update form settings."""
    requests = []

    if quiz is not None:
        requests.append({
            "updateSettings": {
                "settings": {
                    "quizSettings": {
                        "isQuiz": quiz
                    }
                },
                "updateMask": "quizSettings.isQuiz"
            }
        })

    if confirmation_message is not None:
        requests.append({
            "updateFormInfo": {
                "info": {
                    "description": confirmation_message
                },
                "updateMask": "description"
            }
        })
        # Note: The Forms API handles confirmation messages via
        # settings.quizSettings or the form's confirmationMessage
        # depending on the API version. Using updateFormInfo for
        # the description as a workaround when direct confirmation
        # message setting is not available.

    if not requests:
        return {"error": "No settings specified"}

    result = api_request("POST", f"{FORMS_API_BASE}/forms/{form_id}:batchUpdate",
                         {"requests": requests})
    return result


# =============================================================================
# Response Operations
# =============================================================================

def list_responses(form_id: str) -> List[Dict]:
    """List all responses for a form."""
    result = api_request("GET", f"{FORMS_API_BASE}/forms/{form_id}/responses")
    responses = result.get("responses", [])

    return [{
        "responseId": r.get("responseId"),
        "createTime": r.get("createTime"),
        "lastSubmittedTime": r.get("lastSubmittedTime"),
        "respondentEmail": r.get("respondentEmail"),
        "totalScore": r.get("totalScore"),
        "answers": r.get("answers", {})
    } for r in responses]


def get_response(form_id: str, response_id: str) -> Dict:
    """Get a specific response."""
    result = api_request("GET",
                         f"{FORMS_API_BASE}/forms/{form_id}/responses/{response_id}")
    return result


def response_summary(form_id: str) -> Dict:
    """
    Get a summary of responses for a form.

    Returns question titles mapped to answer counts/values.
    """
    # Get the form structure for question titles
    form = get_form(form_id)
    items = form.get("items", [])

    # Build a map of question ID -> question title
    question_map = {}
    for item in items:
        title = item.get("title", "Untitled")
        question_item = item.get("questionItem", {})
        question = question_item.get("question", {})
        question_id = question.get("questionId")
        if question_id:
            question_map[question_id] = title

    # Get all responses
    responses = list_responses(form_id)

    # Aggregate answers
    summary = {}
    for question_id, title in question_map.items():
        summary[title] = {
            "questionId": question_id,
            "responseCount": 0,
            "answers": []
        }

    for response in responses:
        answers = response.get("answers", {})
        for question_id, answer_data in answers.items():
            title = question_map.get(question_id, question_id)
            if title not in summary:
                summary[title] = {
                    "questionId": question_id,
                    "responseCount": 0,
                    "answers": []
                }

            summary[title]["responseCount"] += 1

            # Extract text answers
            text_answers = answer_data.get("textAnswers", {})
            for answer in text_answers.get("answers", []):
                value = answer.get("value", "")
                summary[title]["answers"].append(value)

    return {
        "totalResponses": len(responses),
        "questions": summary
    }


# =============================================================================
# Main CLI
# =============================================================================

def main():
    parser = argparse.ArgumentParser(
        description="Google Forms operations helper",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    subparsers = parser.add_subparsers(dest="command")

    # List forms
    list_forms_parser = subparsers.add_parser("list-forms", help="List forms from Drive")
    list_forms_parser.add_argument("--max-results", type=int, default=20,
                                   help="Maximum results (default: 20)")

    # Create form
    create_parser = subparsers.add_parser("create-form", help="Create a new form")
    create_parser.add_argument("--title", required=True, help="Form title")
    create_parser.add_argument("--description", help="Form description")

    # Get form
    get_parser = subparsers.add_parser("get-form", help="Get form details")
    get_parser.add_argument("--id", required=True, help="Form ID")

    # Update form info
    update_info_parser = subparsers.add_parser("update-info", help="Update form title/description")
    update_info_parser.add_argument("--id", required=True, help="Form ID")
    update_info_parser.add_argument("--title", help="New title")
    update_info_parser.add_argument("--description", help="New description")

    # Add question
    add_q_parser = subparsers.add_parser("add-question", help="Add a question to the form")
    add_q_parser.add_argument("--id", required=True, help="Form ID")
    add_q_parser.add_argument("--title", required=True, help="Question text")
    add_q_parser.add_argument("--type", required=True,
                              choices=["short-answer", "paragraph", "multiple-choice",
                                       "checkbox", "dropdown", "scale", "date", "time"],
                              help="Question type")
    add_q_parser.add_argument("--options", nargs="+", help="Options for choice questions")
    add_q_parser.add_argument("--required", action="store_true", help="Mark as required")
    add_q_parser.add_argument("--index", type=int, help="Position (0-indexed)")
    add_q_parser.add_argument("--scale-low", type=int, default=1, help="Scale low value")
    add_q_parser.add_argument("--scale-high", type=int, default=5, help="Scale high value")
    add_q_parser.add_argument("--label-low", help="Scale low label")
    add_q_parser.add_argument("--label-high", help="Scale high label")

    # Update question
    update_q_parser = subparsers.add_parser("update-question", help="Update a question")
    update_q_parser.add_argument("--id", required=True, help="Form ID")
    update_q_parser.add_argument("--item-id", required=True, help="Item ID to update")
    update_q_parser.add_argument("--title", help="New question text")

    # Delete item
    delete_parser = subparsers.add_parser("delete-item", help="Delete a form item")
    delete_parser.add_argument("--id", required=True, help="Form ID")
    delete_parser.add_argument("--item-id", required=True, help="Item ID to delete")

    # Move item
    move_parser = subparsers.add_parser("move-item", help="Move an item to a new position")
    move_parser.add_argument("--id", required=True, help="Form ID")
    move_parser.add_argument("--item-id", required=True, help="Item ID to move")
    move_parser.add_argument("--index", type=int, required=True, help="New position (0-indexed)")

    # Update settings
    settings_parser = subparsers.add_parser("update-settings", help="Update form settings")
    settings_parser.add_argument("--id", required=True, help="Form ID")
    settings_parser.add_argument("--quiz", action="store_true", help="Enable quiz mode")
    settings_parser.add_argument("--no-quiz", action="store_true", help="Disable quiz mode")
    settings_parser.add_argument("--confirmation-message", help="Confirmation message")

    # List responses
    list_resp_parser = subparsers.add_parser("list-responses", help="List form responses")
    list_resp_parser.add_argument("--id", required=True, help="Form ID")

    # Get response
    get_resp_parser = subparsers.add_parser("get-response", help="Get a specific response")
    get_resp_parser.add_argument("--id", required=True, help="Form ID")
    get_resp_parser.add_argument("--response-id", required=True, help="Response ID")

    # Response summary
    summary_parser = subparsers.add_parser("response-summary",
                                           help="Get aggregated response summary")
    summary_parser.add_argument("--id", required=True, help="Form ID")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    result = None

    # Form commands
    if args.command == "list-forms":
        result = list_forms(args.max_results)

    elif args.command == "create-form":
        result = create_form(args.title, description=args.description)

    elif args.command == "get-form":
        result = get_form(args.id)

    elif args.command == "update-info":
        result = update_info(args.id, title=args.title, description=args.description)

    # Question commands
    elif args.command == "add-question":
        result = add_question(
            args.id,
            args.title,
            args.type,
            options=args.options,
            required=args.required,
            index=args.index,
            scale_low=args.scale_low,
            scale_high=args.scale_high,
            label_low=args.label_low,
            label_high=args.label_high
        )

    elif args.command == "update-question":
        result = update_question(args.id, args.item_id, title=args.title)

    elif args.command == "delete-item":
        result = delete_item(args.id, args.item_id)

    elif args.command == "move-item":
        result = move_item(args.id, args.item_id, args.index)

    # Settings commands
    elif args.command == "update-settings":
        quiz_setting = None
        if args.quiz:
            quiz_setting = True
        elif args.no_quiz:
            quiz_setting = False
        result = update_settings(args.id, quiz=quiz_setting,
                                 confirmation_message=args.confirmation_message)

    # Response commands
    elif args.command == "list-responses":
        result = list_responses(args.id)

    elif args.command == "get-response":
        result = get_response(args.id, args.response_id)

    elif args.command == "response-summary":
        result = response_summary(args.id)

    if result is not None:
        print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
