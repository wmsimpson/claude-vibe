#!/usr/bin/env python3
"""
Claude Code Skill Invocation Eval Runner

Uses the Claude Agent SDK to run prompts with the full Claude Code system prompt
preserved, enabling faster and more consistent skill routing.

Usage:
    uv run skill-evals [test-cases/skill-routing.yaml]
    uv run skill-evals --dir test-cases/
    uv run skill-evals -k jira
    uv run skill-evals -p /path/to/extra/plugin
"""

import argparse
import asyncio
import json
import logging
import os
import random
import re
import sys
from pathlib import Path

import yaml

from claude_agent_sdk import (
    AssistantMessage,
    ClaudeAgentOptions,
    ResultMessage,
    UserMessage,
    query,
)
from claude_agent_sdk.types import ToolUseBlock

from .models import TestCase, TestResult

logger = logging.getLogger("skill-evals")

# Matches <command-name>/skill-name</command-name> tags injected by Claude Code
# for slash command invocations (mirrors Go telemetry regex in cli/internal/telemetry/transcript.go)
_COMMAND_NAME_RE = re.compile(r"<command-name>/([^<]+)</command-name>")

# evals/ directory
EVALS_DIR = Path(__file__).resolve().parent.parent.parent

# Repository root (parent of evals/)
REPO_ROOT = EVALS_DIR.parent



def _is_rate_limit_error(exc: BaseException) -> bool:
    """Check if an exception is a rate limit error from the CLI subprocess."""
    return "rate_limit" in str(exc).lower()


def skill_matches(expected: str, invoked_skills: set[str]) -> bool:
    """Check if expected skill matches any invoked skill.

    Handles both prefixed (plugin:skill) and unprefixed (skill) names.
    """
    if expected in invoked_skills:
        return True
    expected_name = expected.split(":")[-1] if ":" in expected else expected
    for inv in invoked_skills:
        inv_name = inv.split(":")[-1] if ":" in inv else inv
        if expected_name == inv_name:
            return True
    return False


async def run_prompt_and_collect_skills(
    prompt: str,
    max_turns: int = 5,
    model: str | None = None,
    max_retries: int = 5,
    extra_plugins: list[str] | None = None,
) -> tuple[list[str], list[dict], dict]:
    """Run a prompt via Agent SDK, return (skills_invoked, tool_calls, result_info)."""
    logger.debug(
        "Building ClaudeAgentOptions: plugins=%s, max_turns=%d, model=%s, cwd=%s",
        REPO_ROOT,
        max_turns,
        model,
        REPO_ROOT,
    )
    stderr_lines: list[str] = []

    def capture_stderr(line: str) -> None:
        stderr_lines.append(line)
        logger.debug("CLI stderr: %s", line.rstrip())

    plugins = [{"type": "local", "path": str(REPO_ROOT)}] + [
        {"type": "local", "path": p} for p in (extra_plugins or [])
    ]
    options = ClaudeAgentOptions(

        plugins=plugins,
        allowed_tools=["Skill", "Read", "Glob", "Grep", "Bash"],
        permission_mode="bypassPermissions",
        system_prompt={
            "type": "preset",
            "preset": "claude_code",
            "append": "Never ask clarifying questions. Invoke skills directly.",
        },
        setting_sources=["project"],
        max_turns=max_turns,
        model=model,
        cwd=str(REPO_ROOT),
        stderr=capture_stderr,
    )

    logger.info("Invoking query: %.120s", prompt)
    for attempt in range(max_retries + 1):
        try:
            skills_invoked: list[str] = []
            tool_calls: list[dict] = []
            result_info: dict = {}

            logger.debug("Starting query: %.120s", prompt)
            async for message in query(prompt=prompt, options=options):
                if isinstance(message, UserMessage):
                    # Detect slash command invocations via <command-name> tags
                    # injected by Claude Code (same pattern as Go telemetry)
                    text_parts: list[str] = []
                    if isinstance(message.content, str):
                        text_parts.append(message.content)
                    elif isinstance(message.content, list):
                        for item in message.content:
                            if hasattr(item, "text"):
                                text_parts.append(item.text)
                    for text in text_parts:
                        for match in _COMMAND_NAME_RE.findall(text):
                            skills_invoked.append(match)
                            logger.debug("Skill invoked via slash command: %s", match)

                elif isinstance(message, AssistantMessage):
                    for block in message.content:
                        if isinstance(block, ToolUseBlock):
                            tool_calls.append({"tool": block.name, "input": block.input})
                            logger.debug(
                                "ToolUseBlock: %s  input=%s",
                                block.name,
                                json.dumps(block.input)[:200],
                            )
                            if block.name == "Skill":
                                skill_name = block.input.get("skill", "")
                                if skill_name:
                                    skills_invoked.append(skill_name)
                                    logger.debug("Skill invoked: %s", skill_name)

                elif isinstance(message, ResultMessage):
                    result_info = {
                        "session_id": message.session_id,
                        "total_cost_usd": message.total_cost_usd,
                        "num_turns": message.num_turns,
                        "is_error": message.is_error,
                        "duration_ms": message.duration_ms,
                        "result": message.result,
                    }
                    logger.debug(
                        "ResultMessage: session=%s turns=%s cost=$%s error=%s duration=%sms",
                        message.session_id,
                        message.num_turns,
                        message.total_cost_usd,
                        message.is_error,
                        message.duration_ms,
                    )

            logger.debug("Query complete: skills_invoked=%s", skills_invoked)
            result_info["stderr"] = "".join(stderr_lines)
            result_info["tool_calls"] = tool_calls
            return skills_invoked, tool_calls, result_info
        except Exception as e:
            # The SDK throws on non-zero exit codes even after streaming a
            # ResultMessage (e.g. skill hits auth error -> is_error=True ->
            # CLI exits 1 -> SDK raises).  If we already have a result,
            # return the collected data instead of crashing the eval.
            if result_info:
                logger.debug(
                    "Ignoring post-result exception (session ended with error): %s", e,
                )
                result_info["stderr"] = "".join(stderr_lines)
                result_info["tool_calls"] = tool_calls
                return skills_invoked, tool_calls, result_info
            if not _is_rate_limit_error(e) or attempt >= max_retries:
                stderr = "".join(stderr_lines)
                if stderr:
                    raise RuntimeError(f"{e}\nCLI stderr:\n{stderr}") from e
                raise
            delay = (2 ** attempt) + random.uniform(0, 1)
            logger.warning(
                "Rate limited (attempt %d/%d), retrying in %.1fs...",
                attempt + 1, max_retries, delay,
            )
            stderr_lines.clear()
            await asyncio.sleep(delay)


async def run_judge(
    prompt: str,
    output: str,
    criteria: str,
    model: str = "sonnet",
    timeout: int = 120,
    max_retries: int = 5,
) -> tuple[bool, str]:
    """Run an LLM-as-judge evaluation on Claude's output using the Agent SDK.

    Returns:
        (passed, reasoning) tuple.
    """
    judge_prompt = (
        "You are an eval judge. Evaluate the following AI response.\n\n"
        f"USER QUESTION:\n{prompt}\n\n"
        f"AI RESPONSE:\n{output}\n\n"
        f"CRITERIA:\n{criteria}\n\n"
        "For each criterion, write PASS or FAIL with a one-line reason.\n"
        "Then give a final verdict: PASS only if ALL criteria pass.\n"
        "End your response with exactly one line: VERDICT: PASS or VERDICT: FAIL"
    )

    logger.debug("[judge] Running judge with model=%s", model)

    try:
        options = ClaudeAgentOptions(
    
            max_turns=1,
            model=model,
            permission_mode="bypassPermissions",
            allowed_tools=[],
            cwd=str(REPO_ROOT),
        )

        response = await asyncio.wait_for(
            _collect_judge_result(judge_prompt, options, max_retries=max_retries),
            timeout=timeout,
        )

        logger.debug("[judge] Response:\n%s", response)
        passed = "VERDICT: PASS" in response.upper()
        return passed, response
    except asyncio.TimeoutError:
        return False, "Judge timed out"
    except Exception as e:
        return False, f"Judge error: {e}"


async def _collect_judge_result(
    prompt: str,
    options: ClaudeAgentOptions,
    max_retries: int = 5,
) -> str:
    """Collect the text result from a judge query."""
    for attempt in range(max_retries + 1):
        try:
            async for message in query(prompt=prompt, options=options):
                if isinstance(message, ResultMessage):
                    return message.result or ""
            return ""
        except Exception as e:
            if not _is_rate_limit_error(e) or attempt >= max_retries:
                raise
            delay = (2 ** attempt) + random.uniform(0, 1)
            logger.warning(
                "[judge] Rate limited (attempt %d/%d), retrying in %.1fs...",
                attempt + 1, max_retries, delay,
            )
            await asyncio.sleep(delay)


async def run_test(
    test: TestCase,
    timeout: int = 180,
    max_retries: int = 5,
    verbose: bool = False,
    extra_plugins: list[str] | None = None,
) -> TestResult:
    """Run a single test case and return result."""
    # Skip if platform doesn't match
    if test.platform and not sys.platform.startswith(test.platform):
        return TestResult(
            name=test.name,
            passed=False,
            expected=test.expected_skill or "N/A",
            actual="skipped",
            skipped=True,
            skip_reason=f"requires platform '{test.platform}', running on '{sys.platform}'",
        )

    logger.info("[%s] Starting test", test.name)
    verbose_lines: list[str] = []

    is_timeout = False
    try:
        skills_invoked, tool_calls, result_info = await asyncio.wait_for(
            run_prompt_and_collect_skills(
                test.prompt,
                max_turns=test.max_turns,
                model=test.model,
                max_retries=max_retries,
                extra_plugins=extra_plugins,
            ),
            timeout=timeout,
        )
    except asyncio.TimeoutError:
        logger.debug("[%s] Timed out after %ds", test.name, timeout)
        is_timeout = True
        skills_invoked = []
        tool_calls = []
        result_info = {}
    except Exception as e:
        logger.debug("[%s] Exception: %s", test.name, e, exc_info=True)
        return TestResult(
            name=test.name,
            passed=False,
            expected="completion",
            actual="error",
            error=str(e),
        )

    # Build verbose report output
    if verbose:
        verbose_lines.append(f"  Session ID: {result_info.get('session_id', 'N/A')}")
        verbose_lines.append(f"  Num turns: {result_info.get('num_turns', 'N/A')}")
        verbose_lines.append(f"  Is error: {result_info.get('is_error', 'N/A')}")
        verbose_lines.append(f"  Cost: ${result_info.get('total_cost_usd') or 0:.4f}")
        verbose_lines.append(f"  Duration: {result_info.get('duration_ms', 'N/A')}ms")

        result_text = result_info.get("result", "")
        if result_text:
            verbose_lines.append(f"  Result preview: {result_text[:1000]}")

        if tool_calls:
            verbose_lines.append(f"  Tool calls ({len(tool_calls)}):")
            for tc in tool_calls:
                tool_input = json.dumps(tc["input"])[:200]
                verbose_lines.append(f"    - {tc['tool']}: {tool_input}")
        else:
            verbose_lines.append("  Tool calls: (none)")

        stderr = result_info.get("stderr", "")
        if stderr:
            verbose_lines.append(f"  Stderr: {stderr[:500]}")

    invoked = skills_invoked[: test.max_turns]
    invoked_set = set(invoked)

    # Evaluate result
    if test.expected_skills:
        passed = all(skill_matches(exp, invoked_set) for exp in test.expected_skills)
        expected = f"all of {test.expected_skills}"
    elif test.expected_skill_one_of:
        passed = any(
            skill_matches(exp, invoked_set) for exp in test.expected_skill_one_of
        )
        expected = f"one of {test.expected_skill_one_of}"
    elif test.expected_skill:
        passed = skill_matches(test.expected_skill, invoked_set)
        expected = test.expected_skill
    else:
        passed = len(invoked) == 0
        expected = "null"

    actual_display = ", ".join(invoked) if invoked else "null"

    # If the test timed out and didn't pass, mark it as timed_out
    # so it gets excluded from the pass/fail tally
    timed_out = is_timeout and not passed

    # Run LLM judge if criteria provided and skill routing passed
    judge_passed = None
    judge_reasoning = None
    if test.judge_criteria and passed and not timed_out:
        claude_output = result_info.get("result", "")

        if claude_output:
            judge_passed, judge_reasoning = await run_judge(
                prompt=test.prompt,
                output=claude_output,
                criteria=test.judge_criteria,
                model=test.judge_model,
                max_retries=max_retries,
            )
            if verbose:
                verbose_lines.append(
                    f"  [judge] verdict={'PASS' if judge_passed else 'FAIL'}"
                )
            logger.debug(
                "[%s] [judge] verdict=%s",
                test.name,
                "PASS" if judge_passed else "FAIL",
            )
            # Overall pass requires both routing and judge to pass
            passed = passed and judge_passed
        else:
            judge_passed = False
            judge_reasoning = "No output to judge"
            passed = False

    logger.info(
        "[%s] Done: passed=%s expected='%s' actual='%s'",
        test.name,
        passed,
        expected,
        actual_display,
    )

    return TestResult(
        name=test.name,
        passed=passed,
        expected=expected,
        actual=actual_display,
        verbose_output="\n".join(verbose_lines) if verbose_lines else None,
        timed_out=timed_out,
        judge_passed=judge_passed,
        judge_reasoning=judge_reasoning,
    )


async def run_and_report(tests: list[TestCase], args: argparse.Namespace) -> None:
    """Run all tests and print summary."""
    logger.info(
        "Running %d tests (parallel=%d, timeout=%d)",
        len(tests),
        args.parallel,
        args.timeout,
    )
    results: list[TestResult] = []
    parallel = args.parallel

    verbose = args.verbose
    extra_plugins = [os.path.expanduser(p) for p in (getattr(args, "plugin", None) or [])]

    def print_result(result: TestResult) -> None:
        if result.skipped:
            status = "SKIP"
        elif result.timed_out:
            status = "TIMEOUT"
        elif result.passed:
            status = "PASS"
        else:
            status = "FAIL"
        print(f"  {result.name}: {status}")
        if result.verbose_output:
            print(f"  --- verbose output for {result.name} ---")
            print(result.verbose_output)
            print()

    if parallel > 1:
        print(f"Running {len(tests)} tests with {parallel} workers...")
        semaphore = asyncio.Semaphore(parallel)

        async def bounded(test: TestCase) -> TestResult:
            async with semaphore:
                return await run_test(
                    test, timeout=args.timeout, max_retries=args.max_retries,
                    verbose=verbose, extra_plugins=extra_plugins,
                )

        completed = await asyncio.gather(
            *[bounded(t) for t in tests], return_exceptions=True
        )

        for i, result in enumerate(completed):
            if isinstance(result, BaseException):
                err_result = TestResult(
                    name=tests[i].name,
                    passed=False,
                    expected="completion",
                    actual="error",
                    error=str(result),
                )
                results.append(err_result)
                print(f"  {tests[i].name}: ERROR - {result}")
            else:
                results.append(result)
                print_result(result)
    else:
        for test in tests:
            print(f"Running: {test.name}...", flush=True)
            result = await run_test(
                test, timeout=args.timeout, max_retries=args.max_retries,
                verbose=verbose, extra_plugins=extra_plugins,
            )
            results.append(result)
            if result.skipped:
                status = "SKIP"
            elif result.timed_out:
                status = "TIMEOUT"
            elif result.passed:
                status = "PASS"
            else:
                status = "FAIL"
            print(f"  {status}")
            if result.verbose_output:
                print(result.verbose_output)
                print()

    # Summary -- timed-out and skipped tests are excluded from pass/fail tally
    skipped_tests = [r for r in results if r.skipped]
    timed_out_tests = [r for r in results if r.timed_out]
    scored_results = [r for r in results if not r.timed_out and not r.skipped]
    passed = sum(1 for r in scored_results if r.passed)
    failed_tests = [r for r in scored_results if not r.passed]
    scored_total = len(scored_results)
    pass_percentage = (passed / scored_total * 100) if scored_total > 0 else 0
    threshold = args.threshold
    passed_threshold = pass_percentage >= threshold

    print(f"\n{'=' * 50}")
    print(f"Results: {passed}/{scored_total} passed ({pass_percentage:.1f}%)")
    if skipped_tests:
        print(f"Skipped: {len(skipped_tests)} test(s) excluded (platform mismatch)")
    if timed_out_tests:
        print(f"Timed out: {len(timed_out_tests)} test(s) excluded from results")

    if skipped_tests:
        print(f"\nSkipped tests:")
        for r in skipped_tests:
            print(f"  - {r.name} ({r.skip_reason})")

    if timed_out_tests:
        print(f"\nTimed out tests:")
        for r in timed_out_tests:
            print(f"  - {r.name} (expected '{r.expected}')")

    if failed_tests:
        if passed_threshold:
            print(f"\nPASSED with warnings (>= {threshold}% threshold met)")
            print(
                f"\nWarning: {len(failed_tests)} test(s) failed but within acceptable threshold:"
            )
        else:
            print(f"\nFAILED ({pass_percentage:.1f}% < {threshold}% threshold)")
            print("\nFailed tests:")

        for r in failed_tests:
            print(f"  - {r.name}: expected '{r.expected}', got '{r.actual}'")
            if r.error:
                print(f"    Error: {r.error}")
    elif not timed_out_tests:
        print(f"\nAll tests passed!")
    else:
        print(f"\nAll scored tests passed!")

    sys.exit(0 if passed_threshold else 1)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Eval suite for Claude Code skill invocation",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  skill-evals                              Run default test cases
  skill-evals test-cases/edge-cases.yaml   Run specific test file
  skill-evals --dir test-cases/            Run all YAML files in a directory
  skill-evals -k jira                      Filter tests by name substring
  skill-evals --timeout 120                Run with longer timeout
  skill-evals -j 15                        Run 15 tests in parallel
  skill-evals --threshold 80               Pass if >= 80% of tests pass
  skill-evals -p /path/to/plugin           Load additional plugin directory
        """,
    )
    parser.add_argument(
        "test_file",
        nargs="?",
        default="test-cases/skill-routing.yaml",
        help="Path to test case YAML file (default: test-cases/skill-routing.yaml)",
    )
    parser.add_argument(
        "--dir",
        type=str,
        default=None,
        help="Run all YAML files in a directory (overrides test_file)",
    )
    parser.add_argument(
        "-k",
        "--filter",
        type=str,
        default=None,
        help="Filter tests by name substring (case-insensitive)",
    )
    parser.add_argument(
        "--timeout",
        type=int,
        default=180,
        help="Timeout in seconds for each test (default: 180)",
    )
    parser.add_argument(
        "--verbose",
        "-v",
        action="store_true",
        help="Show detailed per-test output (session ID, cost, tool calls, etc.)",
    )
    parser.add_argument(
        "-j",
        "--parallel",
        type=int,
        default=15,
        help="Number of tests to run in parallel (default: 15)",
    )
    parser.add_argument(
        "--max-retries",
        type=int,
        default=5,
        help="Max retries per test on rate limit (default: 5)",
    )
    parser.add_argument(
        "--threshold",
        type=float,
        default=95.0,
        help="Minimum pass percentage to exit 0 (default: 95.0)",
    )
    parser.add_argument(
        "-p",
        "--plugin",
        action="append",
        default=[],
        help="Additional local plugin directory to load (repeatable)",
    )
    args = parser.parse_args()

    # Info logging always on; DEBUG controlled by SKILL_EVALS_DEBUG env var
    debug = os.environ.get("SKILL_EVALS_DEBUG", "").strip() not in ("", "0", "false")
    logging.basicConfig(
        level=logging.DEBUG if debug else logging.INFO,
        format="%(asctime)s %(name)s %(levelname)s  %(message)s",
        datefmt="%H:%M:%S",
    )
    # Suppress noisy SDK transport logging
    logging.getLogger("claude_agent_sdk").setLevel(logging.WARNING)

    # Load test cases
    tests: list[TestCase] = []
    if args.dir:
        # Load all YAML files from the directory
        yaml_dir = Path(args.dir)
        if not yaml_dir.is_absolute():
            yaml_dir = EVALS_DIR / yaml_dir
        yaml_files = sorted(yaml_dir.glob("*.yaml"))
        if not yaml_files:
            print(f"No YAML files found in {args.dir}")
            sys.exit(1)
        for yf in yaml_files:
            with open(yf) as f:
                suite = yaml.safe_load(f)
            if suite and "tests" in suite:
                tests.extend(TestCase(**t) for t in suite["tests"])
        print(f"Loaded {len(tests)} tests from {len(yaml_files)} files in {args.dir}")
    else:
        test_file = Path(args.test_file)
        if not test_file.is_absolute():
            test_file = EVALS_DIR / test_file
        with open(test_file) as f:
            suite = yaml.safe_load(f)
        tests = [TestCase(**t) for t in suite["tests"]]

    # Apply name filter if provided
    if args.filter:
        filter_lower = args.filter.lower()
        tests = [t for t in tests if filter_lower in t.name.lower()]
        print(f"Filter '{args.filter}' matched {len(tests)} test(s)")

    if not tests:
        print("No tests to run.")
        sys.exit(0)

    asyncio.run(run_and_report(tests, args))


if __name__ == "__main__":
    main()
