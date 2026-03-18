from dataclasses import dataclass, field
from typing import Optional


@dataclass
class TestCase:
    name: str
    prompt: str
    expected_skill: Optional[str] = None  # Single skill (legacy)
    expected_skills: Optional[list[str]] = None  # Multiple skills (AND logic)
    expected_skill_one_of: Optional[list[str]] = None  # Any of these (OR logic)
    max_turns: int = 5  # Number of turns to check for skill invocations
    model: Optional[str] = None  # Model to use: haiku, sonnet, opus
    platform: Optional[str] = None  # Only run on this platform (e.g., "darwin", "linux")
    judge_criteria: Optional[str] = None  # Natural language criteria for LLM judge
    judge_model: str = "sonnet"  # Model for judging


@dataclass
class TestResult:
    name: str
    passed: bool
    expected: str
    actual: Optional[str]
    error: Optional[str] = None
    verbose_output: Optional[str] = None
    timed_out: bool = False
    skipped: bool = False
    skip_reason: Optional[str] = None
    judge_passed: Optional[bool] = None
    judge_reasoning: Optional[str] = None
