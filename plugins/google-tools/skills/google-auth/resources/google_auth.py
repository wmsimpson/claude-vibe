#!/usr/bin/env python3
"""
Google Workspace Authentication Manager

Unified authentication for all Google Workspace APIs:
- Google Docs
- Google Sheets
- Google Slides
- Google Drive
- Google Calendar
- Gmail
- Google Tasks
- Google Forms

This is the single source of truth for Google authentication across all
google-tools skills. Other skills should reference this module.

Usage:
    # Check auth status
    python3 google_auth.py status

    # Login with required scopes
    python3 google_auth.py login

    # Force re-authentication
    python3 google_auth.py login --force

    # Get access token (will prompt login if needed)
    python3 google_auth.py token

    # Validate token has required scopes
    python3 google_auth.py validate

    # Show the gcloud login command
    python3 google_auth.py show-login-command
"""

import argparse
import json
import os
import shutil
import subprocess
import sys
from typing import Dict, List, Optional, Tuple


QUOTA_PROJECT = os.environ.get("GCP_QUOTA_PROJECT", "")

# Required scopes for complete Google Workspace access
# All Google skills share these same scopes for unified auth
REQUIRED_SCOPES = [
    "https://www.googleapis.com/auth/drive",           # Drive file management
    "https://www.googleapis.com/auth/cloud-platform",  # GCP access
    "https://www.googleapis.com/auth/documents",       # Google Docs
    "https://www.googleapis.com/auth/presentations",   # Google Slides
    "https://www.googleapis.com/auth/spreadsheets",    # Google Sheets
    "https://www.googleapis.com/auth/calendar",        # Google Calendar
    "https://www.googleapis.com/auth/gmail.modify",    # Gmail (read/write)
    "https://www.googleapis.com/auth/tasks",           # Google Tasks
    "https://www.googleapis.com/auth/forms.body",      # Google Forms (read/write)
    "https://www.googleapis.com/auth/forms.responses.readonly"  # Google Forms responses
]

ADC_PATH = os.path.expanduser("~/.config/gcloud/application_default_credentials.json")


def find_gcloud() -> Optional[str]:
    """
    Find gcloud CLI path using 'which gcloud'.

    Returns:
        Path to gcloud binary, or None if not found
    """
    result = subprocess.run(
        ["which", "gcloud"],
        capture_output=True,
        text=True
    )

    if result.returncode == 0 and result.stdout.strip():
        return result.stdout.strip()

    return None


def install_gcloud_with_brew() -> bool:
    """
    Attempt to install gcloud using Homebrew.

    Returns:
        True if installation succeeded, False otherwise
    """
    print("gcloud CLI not found. Attempting to install via Homebrew...")

    # Check if brew is available
    brew_path = shutil.which("brew")
    if not brew_path:
        print("ERROR: Homebrew not found. Please install gcloud manually:", file=sys.stderr)
        print("  https://cloud.google.com/sdk/docs/install", file=sys.stderr)
        return False

    print("Installing google-cloud-sdk via Homebrew (this may take a few minutes)...")
    result = subprocess.run(
        [brew_path, "install", "--cask", "google-cloud-sdk"],
        text=True
    )

    if result.returncode != 0:
        print("ERROR: Failed to install gcloud via Homebrew.", file=sys.stderr)
        print("Please install manually: https://cloud.google.com/sdk/docs/install", file=sys.stderr)
        return False

    print("gcloud installed successfully!")
    print("\nIMPORTANT: You may need to restart your terminal or run:")
    print('  source "$(brew --prefix)/share/google-cloud-sdk/path.bash.inc"')
    return True


def get_gcloud_path() -> str:
    """
    Get gcloud CLI path, installing if necessary.

    Returns:
        Path to gcloud binary

    Raises:
        SystemExit if gcloud cannot be found or installed
    """
    # First try to find gcloud
    gcloud_path = find_gcloud()

    if gcloud_path:
        return gcloud_path

    # Try to install with brew
    if install_gcloud_with_brew():
        # Try to find it again after installation
        gcloud_path = find_gcloud()
        if gcloud_path:
            return gcloud_path

        # Check common Homebrew install locations
        homebrew_paths = [
            "/opt/homebrew/share/google-cloud-sdk/bin/gcloud",  # Apple Silicon
            "/usr/local/share/google-cloud-sdk/bin/gcloud",     # Intel Mac
            os.path.expanduser("~/google-cloud-sdk/bin/gcloud"),
        ]

        for path in homebrew_paths:
            if os.path.exists(path):
                return path

    print("ERROR: Could not find gcloud after installation.", file=sys.stderr)
    print("Please restart your terminal and try again, or install manually.", file=sys.stderr)
    sys.exit(1)


# Initialize gcloud path (will be set on first use)
_gcloud_path: Optional[str] = None


def gcloud_path() -> str:
    """Get cached gcloud path."""
    global _gcloud_path
    if _gcloud_path is None:
        _gcloud_path = get_gcloud_path()
    return _gcloud_path


def get_login_command() -> str:
    """Get the gcloud login command with required scopes."""
    scopes = ",".join(REQUIRED_SCOPES)
    return f'{gcloud_path()} auth application-default login --scopes="{scopes}"'


def read_adc_file() -> Optional[Dict]:
    """Read the ADC credentials file."""
    if not os.path.exists(ADC_PATH):
        return None

    try:
        with open(ADC_PATH, 'r') as f:
            return json.load(f)
    except (json.JSONDecodeError, IOError):
        return None


def get_adc_scopes() -> List[str]:
    """Get scopes from the ADC file or token info."""
    creds = read_adc_file()
    if not creds:
        return []

    # Scopes may be stored in the credentials file
    scopes = creds.get("scopes", [])

    # If no scopes in file, check the token info
    if not scopes:
        token = get_access_token_quiet()
        if token:
            info = get_token_info(token)
            if info:
                scopes = info.get("scope", "").split()

    return scopes


def check_required_scopes(current_scopes: List[str]) -> Tuple[bool, List[str]]:
    """Check if current scopes include all required scopes."""
    missing = [scope for scope in REQUIRED_SCOPES if scope not in current_scopes]
    return len(missing) == 0, missing


def get_access_token_quiet() -> Optional[str]:
    """Get access token without error output."""
    try:
        result = subprocess.run(
            [gcloud_path(), "auth", "application-default", "print-access-token"],
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            return None

        return result.stdout.strip()
    except Exception:
        return None


def get_access_token() -> str:
    """Get access token, prompting for login if needed."""
    token = get_access_token_quiet()

    if not token:
        print("No valid credentials found. Please authenticate.", file=sys.stderr)
        print(f"\nRun: {get_login_command()}", file=sys.stderr)
        sys.exit(1)

    return token


def get_token_info(token: str) -> Optional[Dict]:
    """Get token information from Google's tokeninfo endpoint."""
    try:
        result = subprocess.run(
            [
                "curl", "-s",
                f"https://oauth2.googleapis.com/tokeninfo?access_token={token}"
            ],
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            return None

        info = json.loads(result.stdout)
        if "error" in info:
            return None
        return info
    except (json.JSONDecodeError, Exception):
        return None


def test_api_access(token: str) -> Dict[str, bool]:
    """Test access to each Google API."""
    results = {}

    api_tests = [
        ("docs", "https://docs.googleapis.com/v1/documents/invalid", ["404", "200"]),
        ("drive", "https://www.googleapis.com/drive/v3/files?pageSize=1", ["200"]),
        ("slides", "https://slides.googleapis.com/v1/presentations/invalid", ["404", "200"]),
        ("sheets", "https://sheets.googleapis.com/v4/spreadsheets/invalid", ["404", "200"]),
        ("calendar", "https://www.googleapis.com/calendar/v3/calendars/primary", ["200"]),
        ("gmail", "https://gmail.googleapis.com/gmail/v1/users/me/profile", ["200"]),
        ("tasks", "https://tasks.googleapis.com/tasks/v1/users/@me/lists?maxResults=1", ["200"]),
        ("forms", "https://forms.googleapis.com/v1/forms/invalid", ["404", "200", "400"]),
    ]

    for api_name, url, valid_codes in api_tests:
        curl_cmd = [
            "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
            url,
            "-H", f"Authorization: Bearer {token}",
        ]
        if QUOTA_PROJECT:
            curl_cmd.extend(["-H", f"x-goog-user-project: {QUOTA_PROJECT}"])
        result = subprocess.run(curl_cmd, capture_output=True, text=True)
        results[api_name] = result.stdout in valid_codes

    return results


def do_login(force: bool = False, retry_on_timeout: bool = True) -> bool:
    """
    Perform authentication.

    Args:
        force: If True, always re-authenticate even if credentials exist
        retry_on_timeout: If True, offer to retry if OAuth flow times out

    Returns:
        True if login successful
    """
    if not force:
        # Check if we already have valid credentials with correct scopes
        token = get_access_token_quiet()
        if token:
            info = get_token_info(token)
            if info:
                current_scopes = info.get("scope", "").split()
                has_scopes, missing = check_required_scopes(current_scopes)
                if has_scopes:
                    print("Already authenticated with required scopes.")
                    return True
                else:
                    print(f"Current token missing scopes: {', '.join(missing)}")
                    print("Re-authenticating...")

    # Run the login command with retry logic
    scopes = ",".join(REQUIRED_SCOPES)

    while True:
        print("\n" + "=" * 70)
        print("Google Workspace Authentication")
        print("=" * 70)
        print("\nIMPORTANT: A browser window will open for OAuth authentication.")
        print("You MUST complete the OAuth flow in the browser to continue.")
        print("\nWhat to expect:")
        print("  1. A browser window/tab will open automatically")
        print("  2. You may need to select your Google account")
        print("  3. You'll be asked to grant permissions")
        print("  4. Click 'Allow' to complete authentication")
        print("\nIf you don't see a browser window:")
        print("  - Check if it opened in the background")
        print("  - Look for a URL in the terminal you can copy/paste")
        print("=" * 70)

        # Don't wait for user input in non-interactive mode
        if sys.stdin.isatty():
            try:
                input("\nPress ENTER to start authentication (or Ctrl+C to cancel)...")
            except KeyboardInterrupt:
                print("\n\nAuthentication cancelled by user.")
                return False
        else:
            print("\nStarting authentication...")

        print("\nOpening browser for authentication...")
        result = subprocess.run(
            [
                gcloud_path(), "auth", "application-default", "login",
                f"--scopes={scopes}"
            ],
            text=True
        )

        if result.returncode == 0:
            # Verify we actually got the credentials
            token = get_access_token_quiet()
            if token:
                print("\n" + "=" * 70)
                print("SUCCESS: Authentication completed!")
                print("=" * 70)
                print("You now have access to: Docs, Sheets, Slides, Drive, Calendar, Gmail, Tasks, Forms")
                print("=" * 70 + "\n")
                return True

        # Authentication failed
        print("\n" + "=" * 70)
        print("ERROR: Authentication Failed")
        print("=" * 70)

        # Check if we have any credentials now
        token = get_access_token_quiet()
        if token:
            # We have a token, but something went wrong
            info = get_token_info(token)
            if info:
                current_scopes = info.get("scope", "").split()
                has_scopes, missing = check_required_scopes(current_scopes)
                if not has_scopes:
                    print("\nYou have credentials, but they're missing required scopes:")
                    for scope in missing:
                        print(f"  - {scope}")
            else:
                print("\nCredentials exist but token validation failed.")
        else:
            print("\nAuthentication failed. This usually happens when:")
            print("  1. The OAuth flow wasn't completed in the browser")
            print("  2. You didn't click 'Allow' in the browser window")
            print("  3. The authentication was cancelled or timed out")
            print("  4. The browser window didn't open")

        print("\nDO NOT try alternative authentication methods.")
        print("DO NOT try to create OAuth client IDs or credentials manually.")
        print("The ONLY way to authenticate is to complete the browser OAuth flow.")
        print("=" * 70)

        if not retry_on_timeout:
            return False

        # Offer to retry in interactive mode
        if sys.stdin.isatty():
            print("\nWould you like to try again?")
            print("Make sure to complete the OAuth flow in the browser this time.")
            try:
                response = input("\nType 'yes' to retry, or anything else to exit: ").strip().lower()
            except (KeyboardInterrupt, EOFError):
                print("\n\nAuthentication cancelled.")
                return False

            if response not in ['yes', 'y']:
                print("\nAuthentication cancelled.")
                print("\nTo authenticate later, run: python3 google_auth.py login")
                return False

            print("\nRetrying authentication...\n")
        else:
            # In non-interactive mode, don't retry
            print("\nAuthentication failed. Run this command interactively to retry.")
            return False


def print_status():
    """Print current authentication status."""
    print("=" * 60)
    print("Google Workspace Authentication Status")
    print("=" * 60)

    # Check gcloud
    print(f"\ngcloud CLI:")
    try:
        gcloud = gcloud_path()
        print(f"  Path: {gcloud}")
        print("  Status: FOUND")
    except SystemExit:
        print("  Status: NOT FOUND")
        print("  Install with: brew install --cask google-cloud-sdk")
        return

    # Check ADC file
    print(f"\nCredentials File: {ADC_PATH}")
    if os.path.exists(ADC_PATH):
        print("  Status: EXISTS")
        creds = read_adc_file()
        if creds:
            print(f"  Type: {creds.get('type', 'unknown')}")
            if 'client_id' in creds:
                print(f"  Client ID: {creds['client_id'][:20]}...")
    else:
        print("  Status: NOT FOUND")
        print("\n  Run: " + get_login_command())
        return

    # Check token
    print("\nAccess Token:")
    token = get_access_token_quiet()
    if not token:
        print("  Status: INVALID/EXPIRED")
        print("\n  Run: " + get_login_command())
        return

    print("  Status: VALID")

    # Get token info
    info = get_token_info(token)
    if info:
        print(f"  Email: {info.get('email', 'unknown')}")

        # Check expiry
        expires_in = info.get('expires_in')
        if expires_in:
            print(f"  Expires in: {int(expires_in) // 60} minutes")

        # Check scopes
        current_scopes = info.get("scope", "").split()
        print(f"\nScopes ({len(current_scopes)}):")
        for scope in current_scopes:
            status = "OK" if scope in REQUIRED_SCOPES else "extra"
            print(f"  - {scope} [{status}]")

        has_scopes, missing = check_required_scopes(current_scopes)
        if missing:
            print(f"\nMISSING REQUIRED SCOPES:")
            for scope in missing:
                print(f"  - {scope}")
            print("\nRe-authenticate with: " + get_login_command())

    # Test API access
    print("\nAPI Access Test:")
    api_results = test_api_access(token)
    all_ok = True
    for api, success in api_results.items():
        status = "OK" if success else "FAILED"
        if not success:
            all_ok = False
        print(f"  {api.capitalize():10} API: {status}")

    if not all_ok:
        print("\nSome APIs failed. This may be due to:")
        print("  1. Missing scopes - run login command above")
        print(f"  2. Quota project permissions - check access to {QUOTA_PROJECT}")
    else:
        print("\nAll APIs accessible. Ready to use Google Workspace skills.")

    print("\n" + "=" * 60)


def main():
    parser = argparse.ArgumentParser(
        description="Unified Google Workspace authentication for all Google skills",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Supported Google APIs:
  - Google Docs     (documents)
  - Google Sheets   (spreadsheets)
  - Google Slides   (presentations)
  - Google Drive    (files)
  - Google Calendar (events)
  - Gmail           (messages)
  - Google Tasks    (tasks)
  - Google Forms    (forms)

Examples:
  google_auth.py status           # Check authentication status
  google_auth.py login            # Authenticate with Google
  google_auth.py login --force    # Force re-authentication
  google_auth.py token            # Get access token for API calls
"""
    )

    subparsers = parser.add_subparsers(dest="command")

    # Status command
    subparsers.add_parser("status", help="Show authentication status for all Google APIs")

    # Login command
    login_parser = subparsers.add_parser("login", help="Authenticate with Google Workspace")
    login_parser.add_argument("--force", "-f", action="store_true",
                              help="Force re-authentication even if credentials exist")
    login_parser.add_argument("--no-retry", action="store_true",
                              help="Don't offer to retry if authentication fails")

    # Token command
    subparsers.add_parser("token", help="Print access token (for use in scripts)")

    # Validate command
    subparsers.add_parser("validate", help="Validate current credentials have required scopes")

    # Show login command
    subparsers.add_parser("show-login-command", help="Print the gcloud login command")

    args = parser.parse_args()

    if not args.command or args.command == "status":
        print_status()

    elif args.command == "login":
        success = do_login(force=args.force, retry_on_timeout=not args.no_retry)
        sys.exit(0 if success else 1)

    elif args.command == "token":
        token = get_access_token()
        print(token)

    elif args.command == "validate":
        token = get_access_token_quiet()
        if not token:
            print("No valid credentials. Run: python3 google_auth.py login", file=sys.stderr)
            sys.exit(1)

        info = get_token_info(token)
        if not info:
            print("Token validation failed. Run: python3 google_auth.py login", file=sys.stderr)
            sys.exit(1)

        current_scopes = info.get("scope", "").split()
        has_scopes, missing = check_required_scopes(current_scopes)

        if has_scopes:
            print("OK: All required scopes present")
            sys.exit(0)
        else:
            print(f"MISSING SCOPES: {', '.join(missing)}", file=sys.stderr)
            print(f"\nRun: {get_login_command()}", file=sys.stderr)
            sys.exit(1)

    elif args.command == "show-login-command":
        print(get_login_command())


if __name__ == "__main__":
    main()
