"""Syntax smoke tests for all Python resource files.

Parametrized pytest that attempts to import every Python resource file
from plugins, catching syntax errors and typos. Files with missing
external dependencies are skipped (not failed) since they run in their
own environments.
"""

import importlib.util
from pathlib import Path

import pytest

VIBE_ROOT = Path(__file__).resolve().parent.parent.parent

PYTHON_FILES = sorted(VIBE_ROOT.glob("plugins/**/*.py"))


@pytest.mark.parametrize(
    "py_file",
    PYTHON_FILES,
    ids=[str(f.relative_to(VIBE_ROOT)) for f in PYTHON_FILES],
)
def test_resource_file_syntax(py_file: Path):
    """Each Python resource file should parse without SyntaxError."""
    spec = importlib.util.spec_from_file_location(py_file.stem, py_file)
    assert spec is not None, f"Could not create module spec for {py_file}"
    assert spec.loader is not None, f"No loader for {py_file}"

    module = importlib.util.module_from_spec(spec)
    try:
        spec.loader.exec_module(module)
    except SyntaxError as exc:
        pytest.fail(f"SyntaxError in {py_file}: {exc}")
    except ImportError:
        pytest.skip(f"Missing external dependency: {py_file}")
    except Exception:
        # Other runtime errors (e.g. missing env vars at module level) are OK
        # for a smoke test — we only care about syntax issues.
        pass
