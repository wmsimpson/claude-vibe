#!/usr/bin/env python3
"""
Shared Google API utilities for all Google Workspace resource scripts.

Provides:
- QUOTA_PROJECT constant
- find_gcloud() — gcloud path discovery with caching
- get_access_token() — ADC token retrieval
- api_call_with_retry() — curl wrapper with retry logic for rate limits and transient errors
"""

import json
import os
import random
import shutil
import subprocess
import time
from typing import Dict, Optional, Tuple
import urllib.parse


# =============================================================================
# Constants
# =============================================================================

# Set GCP_QUOTA_PROJECT env var to your GCP project ID if Sheets/Slides/Forms APIs
# return 403 "quota project required". Run:
#   gcloud auth application-default set-quota-project YOUR_PROJECT_ID
# or add to your shell profile:
#   export GCP_QUOTA_PROJECT=your-project-id
QUOTA_PROJECT = os.environ.get("GCP_QUOTA_PROJECT", "")

# HTTP status codes that are safe to retry
RETRYABLE_STATUS_CODES = {429, 500, 502, 503, 504}

# Cache access tokens for 50 minutes (tokens last 60 min, 10 min safety margin)
_TOKEN_CACHE_TTL = 50 * 60


# =============================================================================
# Token cache
# =============================================================================

_cached_token: Optional[str] = None
_cached_token_time: float = 0.0


# =============================================================================
# gcloud path discovery
# =============================================================================

# Module-level cache to avoid repeated lookups
_gcloud_path_cache: Optional[str] = None


def find_gcloud() -> Optional[str]:
    """
    Find gcloud CLI path dynamically.

    Searches in order:
    1. System PATH via shutil.which()
    2. Common installation locations

    Returns:
        Path to gcloud executable, or None if not found
    """
    global _gcloud_path_cache

    if _gcloud_path_cache:
        return _gcloud_path_cache

    # Try PATH first (works if gcloud is properly installed)
    gcloud_path = shutil.which("gcloud")
    if gcloud_path:
        _gcloud_path_cache = gcloud_path
        return gcloud_path

    # Check common installation locations
    common_paths = [
        os.path.expanduser("~/google-cloud-sdk/bin/gcloud"),
        os.path.expanduser("~/Downloads/google-cloud-sdk/bin/gcloud"),
        "/usr/local/bin/gcloud",
        "/opt/homebrew/bin/gcloud",
        "/opt/homebrew/share/google-cloud-sdk/bin/gcloud",
        "/usr/bin/gcloud",
        "/opt/google-cloud-sdk/bin/gcloud",
    ]

    for path in common_paths:
        if os.path.exists(path):
            _gcloud_path_cache = path
            return path

    return None


# =============================================================================
# Authentication
# =============================================================================

def get_access_token(force_refresh: bool = False) -> str:
    """
    Get an access token from gcloud ADC, with module-level caching.

    Tokens are cached for 50 minutes (they expire after 60). This avoids
    hitting Google's OAuth token grant rate limit when making many API calls.

    Args:
        force_refresh: If True, bypass the cache and fetch a fresh token

    Returns:
        Access token string

    Raises:
        RuntimeError: If gcloud is not found or token retrieval fails
    """
    global _cached_token, _cached_token_time

    # Return cached token if still valid
    if (
        not force_refresh
        and _cached_token
        and (time.time() - _cached_token_time) < _TOKEN_CACHE_TTL
    ):
        return _cached_token

    gcloud_path = find_gcloud()
    if not gcloud_path:
        raise RuntimeError(
            "gcloud CLI not found. Install Google Cloud SDK or run /google-auth."
        )

    result = subprocess.run(
        [gcloud_path, "auth", "application-default", "print-access-token"],
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(
            f"Failed to get access token: {result.stderr.strip()}\n"
            "Re-run /google-auth to refresh credentials."
        )

    _cached_token = result.stdout.strip()
    _cached_token_time = time.time()
    return _cached_token


# =============================================================================
# API call with retry
# =============================================================================

def api_call_with_retry(
    method: str,
    url: str,
    data: Optional[Dict] = None,
    params: Optional[Dict] = None,
    max_retries: int = 3,
    timeout: int = 30,
) -> Dict:
    """
    Make a Google API call via curl with automatic retry for transient errors.

    Retry conditions:
    - curl non-zero exit code (network/timeout error)
    - HTTP response with error code in RETRYABLE_STATUS_CODES (429, 500, 502, 503, 504)

    Non-retryable errors (400, 401, 403, 404, etc.) raise immediately.

    Args:
        method: HTTP method (GET, POST, PUT, PATCH, DELETE)
        url: Full URL to call (query params may be passed via `params`)
        data: JSON body payload (optional)
        params: Query string parameters dict (optional)
        max_retries: Maximum number of attempts (default 3)
        timeout: Per-attempt curl timeout in seconds (default 30)

    Returns:
        Parsed JSON response dict

    Raises:
        RuntimeError: On non-retryable errors or after all retries exhausted
    """
    token = get_access_token()

    if params:
        query_string = urllib.parse.urlencode(params)
        url = f"{url}?{query_string}"

    data_json = json.dumps(data) if data else None
    last_error: Optional[str] = None

    for attempt in range(max_retries):
        cmd = [
            "curl", "-s",
            "--max-time", str(timeout),
            "-X", method,
            url,
            "-H", f"Authorization: Bearer {token}",
            "-H", "Content-Type: application/json",
        ]

        if QUOTA_PROJECT:
            cmd.extend(["-H", f"x-goog-user-project: {QUOTA_PROJECT}"])

        if data_json:
            cmd.extend(["-d", data_json])

        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode != 0:
            # Network or timeout error — transient, retry with backoff
            last_error = f"curl failed (exit {result.returncode}): {result.stderr.strip()}"
        else:
            # Parse the response
            try:
                response = json.loads(result.stdout) if result.stdout else {}
            except json.JSONDecodeError:
                # Non-JSON response — treat as success (some APIs return empty body on DELETE)
                return {"raw": result.stdout}

            error_obj = response.get("error", {})
            error_code = error_obj.get("code") if isinstance(error_obj, dict) else None

            if error_code is None:
                # No error — success
                return response

            if error_code == 401:
                # Token expired/invalid — refresh once and retry
                token = get_access_token(force_refresh=True)
                continue

            if error_code in RETRYABLE_STATUS_CODES:
                # Rate limit or transient server error — retry with same token
                last_error = (
                    f"HTTP {error_code}: {error_obj.get('message', 'unknown error')}"
                )
            else:
                # Non-retryable error (400, 403, 404, etc.)
                raise RuntimeError(
                    f"API error {error_code}: {error_obj.get('message', 'unknown error')}"
                )

        # Backoff before next attempt (not after the last one)
        if attempt < max_retries - 1:
            backoff = 2 ** attempt + random.uniform(0, 1)
            time.sleep(backoff)

    raise RuntimeError(
        f"API call failed after {max_retries} attempts. Last error: {last_error}"
    )
