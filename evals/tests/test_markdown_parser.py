"""Tests for markdown_to_gdocs.py — parse_inline_formatting() and parse_markdown()."""

from markdown_to_gdocs import (
    Paragraph,
    Table,
    TableCell,
    TextSpan,
    parse_inline_formatting,
    parse_markdown,
)


# ---------------------------------------------------------------------------
# parse_inline_formatting tests
# ---------------------------------------------------------------------------

class TestParseInlineFormatting:
    def test_plain_text(self):
        spans = parse_inline_formatting("hello world")
        assert len(spans) == 1
        assert spans[0].text == "hello world"
        assert not spans[0].bold
        assert not spans[0].italic

    def test_bold(self):
        spans = parse_inline_formatting("**bold text**")
        assert any(s.bold and s.text == "bold text" for s in spans)

    def test_italic(self):
        spans = parse_inline_formatting("*italic text*")
        assert any(s.italic and s.text == "italic text" for s in spans)

    def test_bold_italic(self):
        spans = parse_inline_formatting("***both***")
        assert any(s.bold and s.italic and s.text == "both" for s in spans)

    def test_link(self):
        spans = parse_inline_formatting("[click here](https://example.com)")
        assert any(s.link_url == "https://example.com" and s.text == "click here" for s in spans)

    def test_inline_code(self):
        spans = parse_inline_formatting("`code`")
        assert any(s.code and s.text == "code" for s in spans)

    def test_strikethrough(self):
        spans = parse_inline_formatting("~~struck~~")
        assert any(s.strikethrough and s.text == "struck" for s in spans)

    def test_mixed_formatting(self):
        spans = parse_inline_formatting("plain **bold** and *italic*")
        texts = [s.text for s in spans]
        assert "bold" in texts
        assert "italic" in texts



# ---------------------------------------------------------------------------
# parse_markdown tests
# ---------------------------------------------------------------------------

class TestParseMarkdown:
    def test_heading_levels(self):
        for level in range(1, 7):
            md = f"{'#' * level} Heading {level}"
            elements = parse_markdown(md)
            assert len(elements) == 1
            assert isinstance(elements[0], Paragraph)
            assert elements[0].style == f"HEADING_{level}"

    def test_bullet_list(self):
        md = "- item one\n- item two"
        elements = parse_markdown(md)
        assert len(elements) == 2
        assert all(isinstance(e, Paragraph) and e.bullet for e in elements)

    def test_numbered_list(self):
        md = "1. first\n2. second"
        elements = parse_markdown(md)
        assert len(elements) == 2
        assert all(isinstance(e, Paragraph) and e.numbered for e in elements)

    def test_blockquote(self):
        md = "> quote text"
        elements = parse_markdown(md)
        assert len(elements) == 1
        assert isinstance(elements[0], Paragraph)
        assert elements[0].blockquote is True

    def test_table(self):
        md = "| A | B |\n|---|---|\n| 1 | 2 |"
        elements = parse_markdown(md)
        assert len(elements) == 1
        assert isinstance(elements[0], Table)
        assert len(elements[0].rows) == 2  # header + 1 data row
        # Header cells are bold
        assert elements[0].rows[0][0].bold is True

    def test_table_with_link(self):
        md = "| Name | Link |\n|------|------|\n| foo | [bar](https://bar.com) |"
        elements = parse_markdown(md)
        table = elements[0]
        link_cell = table.rows[1][1]
        assert link_cell.link_url == "https://bar.com"
        assert link_cell.text == "bar"

    def test_code_block(self):
        md = "```python\nprint('hello')\n```"
        elements = parse_markdown(md)
        assert len(elements) == 1
        assert isinstance(elements[0], dict)
        assert elements[0]["type"] == "code_block"
        assert elements[0]["language"] == "python"
        assert "print('hello')" in elements[0]["code"]

    def test_horizontal_rule(self):
        md = "---"
        elements = parse_markdown(md)
        assert len(elements) == 1
        assert isinstance(elements[0], dict)
        assert elements[0]["type"] == "hr"

    def test_nested_bullet_list(self):
        md = "- top\n  - nested"
        elements = parse_markdown(md)
        assert elements[0].list_level == 0
        assert elements[1].list_level == 1

    def test_mixed_content(self):
        md = "# Title\n\nSome text.\n\n- bullet\n\n> quote"
        elements = parse_markdown(md)
        types = [(type(e).__name__, getattr(e, "style", None)) for e in elements]
        assert ("Paragraph", "HEADING_1") in types
