#!/usr/bin/env python3
"""
Quicksizer API client for Databricks cost estimation.

Supports single-shot and multi-turn conversations via a messages file.

Prerequisites:
- Databricks CLI installed and configured with a profile that has access to the Quicksizer endpoint

Usage:
    # Single request (uses DATABRICKS_PROFILE env var or 'DEFAULT' profile)
    python3 quicksizer_api.py "Size an ETL workload: 500GB daily on AWS Premium"

    # Specify a Databricks CLI profile
    python3 quicksizer_api.py --profile myprofile "Size an ETL workload: 500GB daily on AWS Premium"

    # Get discovery questions
    python3 quicksizer_api.py --discovery "data warehousing"

    # Multi-turn: first call (creates messages file)
    python3 quicksizer_api.py --messages-file /tmp/qs_chat.json "Size ETL: 500GB daily, AWS"

    # Multi-turn: follow-up (reads history, appends, calls API)
    python3 quicksizer_api.py --messages-file /tmp/qs_chat.json "Yes, 50GB per job run"
"""

import argparse
import json
import subprocess
import sys


ENDPOINT_URL = "https://adb-2548836972759138.18.azuredatabricks.net/serving-endpoints/agents_main-team_fieldeng_sizingagent-quicksizer_agent/invocations"


def get_token(profile: str = "") -> str:
    """Get Databricks token from the configured profile."""
    import os
    profile = profile or os.environ.get("DATABRICKS_PROFILE", "DEFAULT")
    cmd = ["databricks", "auth", "token"]
    if profile and profile != "DEFAULT":
        cmd += ["-p", profile]
    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            check=True,
        )
        token_data = json.loads(result.stdout)
        return token_data.get("access_token", "")
    except subprocess.CalledProcessError as e:
        print(f"Error getting token: {e.stderr}", file=sys.stderr)
        print(
            f"Please run 'databricks auth login' to authenticate (profile: {profile})",
            file=sys.stderr,
        )
        sys.exit(1)
    except json.JSONDecodeError:
        print("Error parsing token response", file=sys.stderr)
        sys.exit(1)


def get_user_email() -> str:
    """Get user email from git config or environment."""
    try:
        result = subprocess.run(
            ["git", "config", "user.email"],
            capture_output=True,
            text=True,
            check=True,
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError:
        import os

        return os.environ.get("USER_EMAIL", "unknown@example.com")


def call_quicksizer(messages: list[dict], user_email: str, token: str) -> dict:
    """Call the Quicksizer agent endpoint."""
    payload = {
        "input": messages,
        "custom_inputs": {"user_email": user_email},
    }

    curl_cmd = [
        "curl",
        "-s",
        "-X",
        "POST",
        ENDPOINT_URL,
        "-H",
        f"Authorization: Bearer {token}",
        "-H",
        "Content-Type: application/json",
        "-d",
        json.dumps(payload),
        "--max-time",
        "300",
    ]

    try:
        result = subprocess.run(
            curl_cmd, capture_output=True, text=True, timeout=320
        )
        if result.returncode != 0:
            print(f"curl error: {result.stderr}", file=sys.stderr)
            sys.exit(1)
        return json.loads(result.stdout)
    except subprocess.TimeoutExpired:
        print("Request timed out after 300 seconds", file=sys.stderr)
        sys.exit(1)
    except json.JSONDecodeError as e:
        print(f"Error parsing response: {e}", file=sys.stderr)
        print(f"Raw response: {result.stdout[:500]}", file=sys.stderr)
        sys.exit(1)


def extract_response_text(response: dict) -> str:
    """Extract the text response from the API response."""
    try:
        output = response.get("output", [])
        for item in reversed(output):
            if item.get("type") == "message" and item.get("content"):
                for content in item["content"]:
                    if content.get("type") in ("text", "output_text"):
                        return content.get("text", "")
        return "No response text found"
    except (KeyError, IndexError, TypeError):
        return json.dumps(response, indent=2)


def load_messages(path: str) -> list[dict]:
    """Load conversation history from a JSON file."""
    try:
        with open(path) as f:
            messages = json.load(f)
        if not isinstance(messages, list):
            print(f"Messages file must contain a JSON array, got {type(messages).__name__}", file=sys.stderr)
            sys.exit(1)
        return messages
    except FileNotFoundError:
        return []
    except json.JSONDecodeError as e:
        print(f"Error parsing messages file: {e}", file=sys.stderr)
        sys.exit(1)


def save_messages(path: str, messages: list[dict]):
    """Save conversation history to a JSON file."""
    with open(path, "w") as f:
        json.dump(messages, f, indent=2)


def main():
    parser = argparse.ArgumentParser(
        description="Get Databricks cost estimates using Quicksizer"
    )
    parser.add_argument("prompt", nargs="?", help="The sizing request or question")
    parser.add_argument(
        "--discovery",
        metavar="USE_CASE",
        help="Get discovery questions for a use case (etl, data_warehousing, ml, interactive, lakebase, lakeflow_connect, apps)",
    )
    parser.add_argument(
        "--messages-file",
        metavar="PATH",
        help="JSON file for multi-turn conversation history. Created if it doesn't exist. "
        "Each call appends the new user message and assistant response.",
    )
    parser.add_argument("--profile", help="Databricks CLI profile to use for authentication (default: DATABRICKS_PROFILE env var or 'DEFAULT')")
    parser.add_argument("--email", help="Override user email for tracking")
    parser.add_argument(
        "--raw", action="store_true", help="Output raw JSON response"
    )

    args = parser.parse_args()

    # Build the prompt
    if args.discovery:
        prompt = f"What discovery questions should I ask for a {args.discovery} use case?"
    elif args.prompt:
        prompt = args.prompt
    else:
        parser.print_help()
        sys.exit(1)

    # Get authentication
    token = get_token(args.profile or "")
    user_email = args.email or get_user_email()

    # Load existing conversation or start fresh
    if args.messages_file:
        messages = load_messages(args.messages_file)
    else:
        messages = []

    # Append the new user message
    messages.append({"role": "user", "content": prompt})

    # Call the API
    response = call_quicksizer(messages, user_email, token)

    if args.raw:
        print(json.dumps(response, indent=2))
    else:
        response_text = extract_response_text(response)
        print(response_text)

        # Append assistant response and save if using messages file
        if args.messages_file:
            messages.append({"role": "assistant", "content": response_text})
            save_messages(args.messages_file, messages)


if __name__ == "__main__":
    main()
