"""
Unit tests for google_api_utils.py — shared Google API retry/auth utilities.
"""

import json
import sys
from unittest.mock import MagicMock, patch
from pathlib import Path

import pytest

# Make the module importable without installing the plugin
sys.path.insert(
    0,
    str(Path(__file__).parent.parent.parent
        / "plugins/google-tools/skills/google-auth/resources"),
)

import google_api_utils
from google_api_utils import api_call_with_retry, RETRYABLE_STATUS_CODES


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _make_run_result(stdout: str, returncode: int = 0) -> MagicMock:
    """Return a mock subprocess.CompletedProcess-like object."""
    result = MagicMock()
    result.returncode = returncode
    result.stdout = stdout
    result.stderr = ""
    return result


def _patch_token(token: str = "test-token"):
    """Patch get_access_token to return a fixed token."""
    return patch.object(google_api_utils, "get_access_token", return_value=token)


# ---------------------------------------------------------------------------
# Test: success on first attempt
# ---------------------------------------------------------------------------

def test_success_on_first_attempt():
    payload = {"documentId": "abc123"}

    with _patch_token(), \
         patch("subprocess.run", return_value=_make_run_result(json.dumps(payload))):
        result = api_call_with_retry("GET", "https://docs.googleapis.com/v1/documents/abc123")

    assert result == payload


# ---------------------------------------------------------------------------
# Test: retry on 429 then succeed
# ---------------------------------------------------------------------------

def test_retry_on_429_then_success():
    rate_limit_resp = json.dumps({"error": {"code": 429, "message": "Rate limit exceeded"}})
    success_resp = json.dumps({"documentId": "abc123"})

    side_effects = [
        _make_run_result(rate_limit_resp),
        _make_run_result(success_resp),
    ]

    with _patch_token(), \
         patch("subprocess.run", side_effect=side_effects), \
         patch("time.sleep") as mock_sleep:
        result = api_call_with_retry("GET", "https://docs.googleapis.com/v1/documents/abc123")

    assert result == {"documentId": "abc123"}
    assert mock_sleep.call_count == 1  # one sleep between attempts


# ---------------------------------------------------------------------------
# Test: no retry on 404
# ---------------------------------------------------------------------------

def test_no_retry_on_404():
    not_found_resp = json.dumps({"error": {"code": 404, "message": "Not Found"}})

    with _patch_token(), \
         patch("subprocess.run", return_value=_make_run_result(not_found_resp)) as mock_run, \
         patch("time.sleep"):
        with pytest.raises(RuntimeError, match="404"):
            api_call_with_retry("GET", "https://docs.googleapis.com/v1/documents/missing")

    # Should have called subprocess only once — no retry
    assert mock_run.call_count == 1


# ---------------------------------------------------------------------------
# Test: raises after max retries exhausted
# ---------------------------------------------------------------------------

def test_raises_after_max_retries():
    error_resp = json.dumps({"error": {"code": 503, "message": "Service Unavailable"}})

    with _patch_token(), \
         patch("subprocess.run", return_value=_make_run_result(error_resp)), \
         patch("time.sleep"):
        with pytest.raises(RuntimeError, match="3 attempts"):
            api_call_with_retry(
                "GET",
                "https://docs.googleapis.com/v1/documents/abc",
                max_retries=3,
            )


# ---------------------------------------------------------------------------
# Test: backoff timing uses exponential formula
# ---------------------------------------------------------------------------

def test_backoff_timing():
    error_resp = json.dumps({"error": {"code": 500, "message": "Internal Server Error"}})

    sleep_calls = []

    def capture_sleep(seconds):
        sleep_calls.append(seconds)

    with _patch_token(), \
         patch("subprocess.run", return_value=_make_run_result(error_resp)), \
         patch("time.sleep", side_effect=capture_sleep), \
         patch("random.uniform", return_value=0.0):
        with pytest.raises(RuntimeError):
            api_call_with_retry("GET", "https://example.com", max_retries=3)

    # Expect sleeps of 2^0=1 and 2^1=2 (with uniform=0)
    assert len(sleep_calls) == 2
    assert sleep_calls[0] == pytest.approx(1.0, abs=0.01)
    assert sleep_calls[1] == pytest.approx(2.0, abs=0.01)


# ---------------------------------------------------------------------------
# Test: curl non-zero exit code triggers retry
# ---------------------------------------------------------------------------

def test_curl_failure_triggers_retry():
    failed = _make_run_result("", returncode=28)  # curl timeout
    success = _make_run_result(json.dumps({"ok": True}))

    with _patch_token(), \
         patch("subprocess.run", side_effect=[failed, success]), \
         patch("time.sleep"):
        result = api_call_with_retry("GET", "https://example.com")

    assert result == {"ok": True}


# ---------------------------------------------------------------------------
# Test: params are appended as query string
# ---------------------------------------------------------------------------

def test_params_appended_to_url():
    success_resp = json.dumps({"items": []})

    captured_cmd = []

    def fake_run(cmd, **kwargs):
        captured_cmd.extend(cmd)
        return _make_run_result(success_resp)

    with _patch_token(), patch("subprocess.run", side_effect=fake_run):
        api_call_with_retry(
            "GET",
            "https://www.googleapis.com/calendar/v3/calendars/primary/events",
            params={"maxResults": "10", "orderBy": "startTime"},
        )

    # The URL in the command should contain the query params
    url_in_cmd = [arg for arg in captured_cmd if arg.startswith("https://")]
    assert len(url_in_cmd) == 1
    assert "maxResults=10" in url_in_cmd[0]
    assert "orderBy=startTime" in url_in_cmd[0]


# ---------------------------------------------------------------------------
# Test: RETRYABLE_STATUS_CODES contains expected codes
# ---------------------------------------------------------------------------

def test_retryable_status_codes():
    assert 429 in RETRYABLE_STATUS_CODES
    assert 500 in RETRYABLE_STATUS_CODES
    assert 502 in RETRYABLE_STATUS_CODES
    assert 503 in RETRYABLE_STATUS_CODES
    assert 504 in RETRYABLE_STATUS_CODES
    # Non-retryable codes should NOT be in the set
    assert 400 not in RETRYABLE_STATUS_CODES
    assert 401 not in RETRYABLE_STATUS_CODES
    assert 403 not in RETRYABLE_STATUS_CODES
    assert 404 not in RETRYABLE_STATUS_CODES
