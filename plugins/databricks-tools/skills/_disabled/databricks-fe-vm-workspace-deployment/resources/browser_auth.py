#!/usr/bin/env python3
"""
FE Vending Machine Browser Authentication

Uses Chrome DevTools MCP to navigate to FEVM, wait for SSO login,
and extract the session cookie automatically.

This script is designed to be called by Claude Code with MCP access.
It outputs instructions and commands for Claude to execute.

Usage:
    python3 browser_auth.py authenticate
    python3 browser_auth.py extract-cookie
    python3 browser_auth.py check-login-status
"""

import sys
import json
import argparse
from pathlib import Path

# Import the environment manager for saving session
SCRIPT_DIR = Path(__file__).parent
sys.path.insert(0, str(SCRIPT_DIR))
from environment_manager import save_session, get_session, FEVM_DIR


FEVM_URL = "https://vending-machine-main-2481552415672103.aws.databricksapps.com/"
FEVM_COOKIE_NAME = "__Host-databricksapps"


def print_auth_instructions():
    """Print instructions for Claude to authenticate via Chrome DevTools MCP."""
    instructions = """
## FE Vending Machine Authentication Required

To authenticate with FEVM, Claude needs to use the Chrome DevTools MCP. Follow these steps:

### Step 1: Navigate to FEVM

```bash
mcp-cli call chrome-devtools/navigate_page '{"type": "url", "url": "https://vending-machine-main-2481552415672103.aws.databricksapps.com/"}'
```

### Step 2: Check if Already Logged In

Take a snapshot to see the current page state:

```bash
mcp-cli call chrome-devtools/take_snapshot '{}'
```

**If you see "Logged as <email>"**: The user is already authenticated. Skip to Step 4.

**If you see "Login" or SSO prompt**: The user needs to complete SSO login. Continue to Step 3.

### Step 3: Complete SSO Login (If Required)

If on a login page:

1. Click "Continue with SSO" if visible
2. The user will need to complete Okta authentication (may require Touch ID)
3. Wait for redirect back to FEVM
4. Take another snapshot to verify login succeeded

```bash
# Click SSO button (adjust uid based on snapshot)
mcp-cli call chrome-devtools/click '{"uid": "<sso_button_uid>"}'

# Wait and check status
sleep 5
mcp-cli call chrome-devtools/take_snapshot '{}'
```

### Step 4: Extract the Session Cookie

Once logged in (you see "Logged as <email>" in the page):

```bash
mcp-cli call chrome-devtools/evaluate_script '{"function": "() => { return document.cookie; }"}'
```

Then parse the cookie to find `__Host-databricksapps`:

```python
# Parse cookies and find the FEVM session cookie
cookies = "<cookie_string_from_above>"
for cookie in cookies.split(";"):
    if "__Host-databricksapps" in cookie:
        session_cookie = cookie.split("=", 1)[1].strip()
        break
```

### Step 5: Save the Session

Save the extracted cookie using this script:

```bash
python3 browser_auth.py save-cookie "<session_cookie_value>"
```

This saves the cookie to ~/.vibe/fe-vm/session.json with a 24-hour expiry.

---

**IMPORTANT**: The session cookie expires after approximately 24-48 hours.
If API calls start failing with 401/403 errors, re-run authentication.
"""
    print(instructions)


def print_mcp_commands():
    """Print the MCP commands for Claude to execute."""
    commands = {
        "navigate": f'mcp-cli call chrome-devtools/navigate_page \'{{"type": "url", "url": "{FEVM_URL}"}}\'',
        "snapshot": "mcp-cli call chrome-devtools/take_snapshot '{}'",
        "get_cookies": 'mcp-cli call chrome-devtools/evaluate_script \'{"function": "() => { return document.cookie; }"}\'',
        "get_url": 'mcp-cli call chrome-devtools/evaluate_script \'{"function": "() => { return window.location.href; }"}\'',
    }
    print(json.dumps(commands, indent=2))


def check_session_status():
    """Check if we have a valid session."""
    session = get_session()

    if session:
        print(json.dumps({
            "status": "valid",
            "created_at": session.get("created_at"),
            "expires_at": session.get("expires_at"),
            "has_cookie": bool(session.get("cookie"))
        }, indent=2))
        return True
    else:
        print(json.dumps({
            "status": "invalid",
            "message": "No valid session found. Authentication required."
        }, indent=2))
        return False


def save_cookie_from_arg(cookie_value: str, expires_hours: int = 24):
    """Save a cookie value provided as argument."""
    if not cookie_value:
        print(json.dumps({"error": "Cookie value is required"}))
        sys.exit(1)

    # Clean up the cookie value if needed
    cookie_value = cookie_value.strip()
    if cookie_value.startswith(f"{FEVM_COOKIE_NAME}="):
        cookie_value = cookie_value.split("=", 1)[1]

    save_session(cookie_value, expires_hours=expires_hours)
    print(json.dumps({
        "status": "saved",
        "message": f"Session cookie saved to {FEVM_DIR / 'session.json'}",
        "expires_hours": expires_hours
    }, indent=2))


def parse_cookies_for_fevm(cookie_string: str) -> str:
    """
    Parse a cookie string and extract the FEVM session cookie.

    Args:
        cookie_string: Full cookie string from browser

    Returns:
        The session cookie value or empty string
    """
    for cookie in cookie_string.split(";"):
        cookie = cookie.strip()
        if cookie.startswith(f"{FEVM_COOKIE_NAME}="):
            return cookie.split("=", 1)[1]
    return ""


def generate_auth_script():
    """
    Generate a script that Claude can use to authenticate.

    This outputs a series of MCP commands and checks that Claude should execute.
    """
    script = """
# FE Vending Machine Authentication Script
# Claude should execute these commands in order

# 1. Navigate to FEVM
echo "Navigating to FEVM..."
mcp-cli call chrome-devtools/navigate_page '{"type": "url", "url": "https://vending-machine-main-2481552415672103.aws.databricksapps.com/"}'

# 2. Wait a moment for page load
sleep 2

# 3. Check page state
echo "Checking page state..."
mcp-cli call chrome-devtools/take_snapshot '{}'

# If the snapshot shows "Logged as <email>", user is authenticated.
# If it shows login page, user needs to complete SSO.

# 4. After user is logged in, extract cookie
echo "Extracting session cookie..."
COOKIES=$(mcp-cli call chrome-devtools/evaluate_script '{"function": "() => { return document.cookie; }"}' | jq -r '.result // .')

# 5. Parse and save the cookie
echo "Cookie string: $COOKIES"
# Use browser_auth.py to save:
# python3 browser_auth.py save-cookie "$COOKIE_VALUE"
"""
    print(script)


def main():
    parser = argparse.ArgumentParser(
        description="FE Vending Machine Browser Authentication Helper"
    )

    subparsers = parser.add_subparsers(dest="command", help="Commands")

    # authenticate - show full instructions
    subparsers.add_parser("authenticate", help="Show authentication instructions")

    # mcp-commands - show MCP commands for Claude
    subparsers.add_parser("mcp-commands", help="Show MCP commands in JSON format")

    # check-session - check if session is valid
    subparsers.add_parser("check-session", help="Check if session is valid")

    # save-cookie - save a cookie value
    save_parser = subparsers.add_parser("save-cookie", help="Save session cookie")
    save_parser.add_argument("cookie", help="Cookie value to save")
    save_parser.add_argument("--expires-hours", type=int, default=24,
                           help="Hours until expiry (default: 24)")

    # parse-cookies - parse a cookie string
    parse_parser = subparsers.add_parser("parse-cookies", help="Parse cookie string for FEVM cookie")
    parse_parser.add_argument("cookie_string", help="Full cookie string")

    # script - generate auth script
    subparsers.add_parser("script", help="Generate authentication script")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    if args.command == "authenticate":
        print_auth_instructions()

    elif args.command == "mcp-commands":
        print_mcp_commands()

    elif args.command == "check-session":
        if not check_session_status():
            sys.exit(1)

    elif args.command == "save-cookie":
        save_cookie_from_arg(args.cookie, args.expires_hours)

    elif args.command == "parse-cookies":
        cookie = parse_cookies_for_fevm(args.cookie_string)
        if cookie:
            print(json.dumps({"cookie": cookie}))
        else:
            print(json.dumps({"error": f"Cookie '{FEVM_COOKIE_NAME}' not found in string"}))
            sys.exit(1)

    elif args.command == "script":
        generate_auth_script()


if __name__ == "__main__":
    main()
