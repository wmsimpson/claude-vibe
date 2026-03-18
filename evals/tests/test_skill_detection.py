"""Tests for skill invocation detection in the eval runner."""

from skill_evals.runner import _COMMAND_NAME_RE, skill_matches


class TestCommandNameRegex:
    """Test <command-name> tag parsing for slash command detection."""

    def test_simple_slash_command(self):
        text = "<command-name>/linkedin-post-generator</command-name>"
        matches = _COMMAND_NAME_RE.findall(text)
        assert matches == ["linkedin-post-generator"]

    def test_prefixed_slash_command(self):
        text = "<command-name>/fe-social-media-tools:linkedin-post-generator</command-name>"
        matches = _COMMAND_NAME_RE.findall(text)
        assert matches == ["fe-social-media-tools:linkedin-post-generator"]

    def test_slash_command_with_surrounding_text(self):
        text = (
            "<command-message>fe-databricks-tools:databricks-lakeview-dashboard</command-message>\n"
            "<command-name>/fe-databricks-tools:databricks-lakeview-dashboard</command-name>"
        )
        matches = _COMMAND_NAME_RE.findall(text)
        assert matches == ["fe-databricks-tools:databricks-lakeview-dashboard"]

    def test_multiple_slash_commands(self):
        text = (
            "<command-name>/fe-google-tools:google-docs</command-name> "
            "<command-name>/fe-google-tools:google-slides</command-name>"
        )
        matches = _COMMAND_NAME_RE.findall(text)
        assert matches == ["fe-google-tools:google-docs", "fe-google-tools:google-slides"]

    def test_no_match_without_tags(self):
        text = "Generate a LinkedIn post from my promotional emails"
        matches = _COMMAND_NAME_RE.findall(text)
        assert matches == []

    def test_no_match_without_slash(self):
        text = "<command-name>linkedin-post-generator</command-name>"
        matches = _COMMAND_NAME_RE.findall(text)
        assert matches == []


class TestSkillMatches:
    """Test skill matching logic."""

    def test_exact_match(self):
        assert skill_matches("fe-social-media-tools:linkedin-post-generator",
                             {"fe-social-media-tools:linkedin-post-generator"})

    def test_suffix_match(self):
        assert skill_matches("fe-social-media-tools:linkedin-post-generator",
                             {"linkedin-post-generator"})

    def test_unprefixed_expected_matches_prefixed_invoked(self):
        assert skill_matches("linkedin-post-generator",
                             {"fe-social-media-tools:linkedin-post-generator"})

    def test_no_match(self):
        assert not skill_matches("fe-social-media-tools:linkedin-post-generator",
                                 {"fe-google-tools:google-docs"})

    def test_empty_invoked_set(self):
        assert not skill_matches("fe-social-media-tools:linkedin-post-generator", set())
