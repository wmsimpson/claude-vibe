#!/usr/bin/env python3
"""
Google Slides Builder - Build presentations with proper element management

This script helps create well-formatted Google Slides presentations by:
1. Creating presentations and slides with Databricks templates
2. Adding shapes, text, images, tables, and charts
3. Duplicating slides and managing layouts
4. Replacing text in placeholders and shapes
5. Copying slides between presentations

Usage:
    # Create from Databricks template
    python3 gslides_builder.py create-from-template --title "My Presentation"

    # Add a slide with template layout
    python3 gslides_builder.py add-template-slide --pres-id "PRES_ID" --layout "content_basic"

    # Replace placeholder text
    python3 gslides_builder.py replace-text --pres-id "PRES_ID" --find "{{TITLE}}" --replace "My Title"
"""

import argparse
import json
import sys
import uuid
from typing import Dict, List, Optional, Tuple

from google_api_utils import api_call_with_retry

# EMU (English Metric Units) conversion
# 1 inch = 914400 EMU, 1 pt = 12700 EMU
EMU_PER_INCH = 914400
EMU_PER_PT = 12700

# Standard slide dimensions (10" x 5.625" for 16:9)
SLIDE_WIDTH_EMU = 9144000   # 10 inches
SLIDE_HEIGHT_EMU = 5143500  # 5.625 inches

# =============================================================================
# SPATIAL AWARENESS - Slide Layout Constants (in inches)
# =============================================================================
# Standard slide: 10" x 5.625" (16:9 aspect ratio)
SLIDE_WIDTH = 10.0
SLIDE_HEIGHT = 5.625

# Margins and safe areas
MARGIN_LEFT = 0.5
MARGIN_RIGHT = 0.5
MARGIN_TOP = 0.5
MARGIN_BOTTOM = 0.5

# Content area (accounting for margins)
CONTENT_LEFT = MARGIN_LEFT
CONTENT_TOP = MARGIN_TOP
CONTENT_WIDTH = SLIDE_WIDTH - MARGIN_LEFT - MARGIN_RIGHT  # 9.0 inches
CONTENT_HEIGHT = SLIDE_HEIGHT - MARGIN_TOP - MARGIN_BOTTOM  # 4.625 inches

# Title area (when title placeholder is present)
TITLE_HEIGHT = 0.75  # Typical title placeholder height
BODY_TOP = MARGIN_TOP + TITLE_HEIGHT + 0.25  # Body starts below title
BODY_HEIGHT = CONTENT_HEIGHT - TITLE_HEIGHT - 0.25  # Remaining height for body

# Dark slide body area (Databricks dark templates have different content area)
# Content starts lower on dark slides due to graphical header elements
DARK_BODY_TOP = 2.25  # Dark slides content starts lower
DARK_BODY_HEIGHT = SLIDE_HEIGHT - DARK_BODY_TOP - MARGIN_BOTTOM  # ~2.875 inches

# Predefined positions (x, y, width, height) for common placements
POSITIONS = {
    # Full content area
    "full": (CONTENT_LEFT, BODY_TOP, CONTENT_WIDTH, BODY_HEIGHT),
    "full_no_title": (CONTENT_LEFT, CONTENT_TOP, CONTENT_WIDTH, CONTENT_HEIGHT),

    # Horizontal thirds
    "left_third": (CONTENT_LEFT, BODY_TOP, CONTENT_WIDTH / 3 - 0.1, BODY_HEIGHT),
    "center_third": (CONTENT_LEFT + CONTENT_WIDTH / 3, BODY_TOP, CONTENT_WIDTH / 3 - 0.1, BODY_HEIGHT),
    "right_third": (CONTENT_LEFT + 2 * CONTENT_WIDTH / 3, BODY_TOP, CONTENT_WIDTH / 3 - 0.1, BODY_HEIGHT),

    # Horizontal halves
    "left_half": (CONTENT_LEFT, BODY_TOP, CONTENT_WIDTH / 2 - 0.1, BODY_HEIGHT),
    "right_half": (CONTENT_LEFT + CONTENT_WIDTH / 2, BODY_TOP, CONTENT_WIDTH / 2 - 0.1, BODY_HEIGHT),

    # Vertical halves
    "top_half": (CONTENT_LEFT, BODY_TOP, CONTENT_WIDTH, BODY_HEIGHT / 2 - 0.1),
    "bottom_half": (CONTENT_LEFT, BODY_TOP + BODY_HEIGHT / 2, CONTENT_WIDTH, BODY_HEIGHT / 2 - 0.1),

    # Quadrants
    "top_left": (CONTENT_LEFT, BODY_TOP, CONTENT_WIDTH / 2 - 0.1, BODY_HEIGHT / 2 - 0.1),
    "top_right": (CONTENT_LEFT + CONTENT_WIDTH / 2, BODY_TOP, CONTENT_WIDTH / 2 - 0.1, BODY_HEIGHT / 2 - 0.1),
    "bottom_left": (CONTENT_LEFT, BODY_TOP + BODY_HEIGHT / 2, CONTENT_WIDTH / 2 - 0.1, BODY_HEIGHT / 2 - 0.1),
    "bottom_right": (CONTENT_LEFT + CONTENT_WIDTH / 2, BODY_TOP + BODY_HEIGHT / 2, CONTENT_WIDTH / 2 - 0.1, BODY_HEIGHT / 2 - 0.1),

    # Centered elements (various sizes)
    "center_large": (1.0, BODY_TOP, 8.0, BODY_HEIGHT),
    "center_medium": (2.0, BODY_TOP + 0.5, 6.0, BODY_HEIGHT - 1.0),
    "center_small": (3.0, BODY_TOP + 1.0, 4.0, BODY_HEIGHT - 2.0),

    # Table positions (optimized for readability) - LIGHT slides
    "table_full": (0.5, BODY_TOP, 9.0, 3.0),
    "table_left": (0.5, BODY_TOP, 4.25, 3.0),
    "table_right": (5.25, BODY_TOP, 4.25, 3.0),

    # Table positions for DARK slides (content starts lower)
    "table_full_dark": (0.5, DARK_BODY_TOP, 9.0, 2.5),
    "table_left_dark": (0.5, DARK_BODY_TOP, 4.25, 2.5),
    "table_right_dark": (5.25, DARK_BODY_TOP, 4.25, 2.5),

    # Chart positions
    "chart_full": (0.75, BODY_TOP + 0.25, 8.5, 3.25),
    "chart_left": (0.5, BODY_TOP, 4.5, 3.5),
    "chart_right": (5.0, BODY_TOP, 4.5, 3.5),

    # Chart positions for DARK slides
    "chart_full_dark": (0.75, DARK_BODY_TOP, 8.5, 2.75),
    "chart_left_dark": (0.5, DARK_BODY_TOP, 4.5, 2.75),
    "chart_right_dark": (5.0, DARK_BODY_TOP, 4.5, 2.75),

    # Image positions
    "image_left": (0.5, BODY_TOP, 4.0, 3.5),
    "image_right": (5.5, BODY_TOP, 4.0, 3.5),
    "image_center": (2.5, BODY_TOP, 5.0, 3.5),
    "image_background": (0, 0, SLIDE_WIDTH, SLIDE_HEIGHT),

    # Text box positions
    "text_title_area": (CONTENT_LEFT, MARGIN_TOP, CONTENT_WIDTH, TITLE_HEIGHT),
    "text_subtitle": (CONTENT_LEFT, MARGIN_TOP + TITLE_HEIGHT, CONTENT_WIDTH, 0.5),
    "text_footer": (CONTENT_LEFT, SLIDE_HEIGHT - 0.75, CONTENT_WIDTH, 0.5),
    "text_caption": (CONTENT_LEFT, SLIDE_HEIGHT - 1.0, CONTENT_WIDTH, 0.75),
}

# Common element sizes
SIZES = {
    "icon_small": (0.5, 0.5),
    "icon_medium": (1.0, 1.0),
    "icon_large": (1.5, 1.5),
    "logo_small": (1.5, 0.5),
    "logo_medium": (2.5, 0.8),
    "logo_large": (3.5, 1.2),
}

# Predefined layouts (generic Google Slides)
LAYOUTS = {
    "BLANK": "BLANK",
    "TITLE": "TITLE",
    "TITLE_AND_BODY": "TITLE_AND_BODY",
    "TITLE_AND_TWO_COLUMNS": "TITLE_AND_TWO_COLUMNS",
    "TITLE_ONLY": "TITLE_ONLY",
    "SECTION_HEADER": "SECTION_HEADER",
    "ONE_COLUMN_TEXT": "ONE_COLUMN_TEXT",
    "MAIN_POINT": "MAIN_POINT",
    "BIG_NUMBER": "BIG_NUMBER",
    "CAPTION_ONLY": "CAPTION_ONLY",
}

# Databricks Corporate Template
DATABRICKS_TEMPLATE_ID = "1p6-qcJw8sEcfVlsbLRKVDZAonCaYvcsBNhFxFU80Whk"

# Databricks brand colors (RGB 0-1 scale)
DATABRICKS_COLORS = {
    "red": {"red": 1.0, "green": 0.224, "blue": 0.161},        # #FF3621
    "orange": {"red": 1.0, "green": 0.439, "blue": 0.204},     # #FF7033
    "yellow": {"red": 0.984, "green": 0.702, "blue": 0.0},     # #FBB300
    "navy": {"red": 0.0, "green": 0.192, "blue": 0.349},       # #003159
    "dark_navy": {"red": 0.071, "green": 0.165, "blue": 0.271}, # #122A45
    "white": {"red": 1.0, "green": 1.0, "blue": 1.0},
    "light_gray": {"red": 0.969, "green": 0.969, "blue": 0.969}, # #F7F7F7
    "gray": {"red": 0.6, "green": 0.6, "blue": 0.6},
}

# Databricks Template Layouts - Light Theme
DATABRICKS_LAYOUTS_LIGHT = {
    # Standard layouts
    "title": "g324ba092b07_3_198",           # 3 Title Slide B - Light 1
    "title_alt": "g324ba092b07_3_125",       # 3 Title Slide B - Light 2
    "title_gradient": "g32c3cd6d0e3_1_88",   # 3 Title Slide B - Light 3
    "title_orange": "g32c3cd6d0e3_1_108",    # 3 Title Slide B - Light 4

    # Content layouts
    "content_basic": "g324ba092b07_3_45",    # 7 Content A - Basic
    "content_basic_white": "g324ba092b07_3_215",  # 8 Content A - Basic White 1
    "content_2col": "g324ba092b07_3_50",     # 9 Content B - 2 Column
    "content_2col_icon": "g324ba092b07_3_58", # 10 Content B - 2 Column w/ Icon Spot
    "content_3col": "g324ba092b07_3_66",     # 11 Content C - 3 Column
    "content_3col_icon": "g324ba092b07_3_76", # 12 Content C - 3 Column w/ Icon Spot
    "content_3col_cards": "g324ba092b07_3_92", # 13 Content C - 3 Column Cards
    "content_card_right": "g324ba092b07_3_105", # 14 Content D - Card Right
    "content_card_left": "g324ba092b07_3_112",  # 15 Content D - Card Left
    "content_card_large": "g324ba092b07_3_119", # 16 Content D - Card Large

    # Section breaks
    "section_break_1": "g324ba092b07_3_223", # Content E - Section Break 1
    "section_break_2": "g32fee89b7b9_0_39",  # Content E - Section Break 2
    "section_break_3": "g32fee89b7b9_0_49",  # Content E - Section Break 3
    "section_break_4": "g32fee89b7b9_0_69",  # Content E - Section Break 4
    "section_break_5": "g32fee89b7b9_0_74",  # Content E - Section Break 5
    "section_break_6": "g3344513dabd_0_5",   # Content E - Section Break 6
    "section_break_7": "g3344513dabd_0_15",  # Content E - Section Break 7
    "section_break_8": "g32fee89b7b9_0_79",  # Content E - Section Break 8

    # Special layouts
    "blank": "g3413a4b56ae_5_32",            # Content E - Blank
    "power_statement": "g324ba092b07_3_220", # Content E - Power Statement 1
    "power_statement_2": "g324ba092b07_3_229", # Content E - Power Statement 2
    "closing": "g324ba092b07_3_233",         # Z - Closing Light

    # Industry layouts
    "industry_media": "g324ba092b07_3_134",  # Light Industry: M&E
    "industry_retail": "g324ba092b07_3_142", # Light Industry: Retail
    "industry_healthcare": "g324ba092b07_3_150", # Light Industry: Healthcare
    "industry_manufacturing": "g324ba092b07_3_158", # Light Industry: Manufacturing
    "industry_financial": "g324ba092b07_3_174", # Light Industry: Financial Services
    "industry_public": "g32c3cd6d0e3_1_153", # Light Industry: Public
    "industry_consumer": "g324ba092b07_3_190", # Light Industry: Consumer Goods
}

# Databricks Template Layouts - Dark Theme
DATABRICKS_LAYOUTS_DARK = {
    "title": "g324ba092b07_3_482",           # 3 Title Slide B - Dark
    "title_alt": "g324ba092b07_3_494",       # 3 Title Slide B - Dark
    "content_basic": "g324ba092b07_3_515",   # 7 Content A - Basic
    "content_2col": "g324ba092b07_3_520",    # 9 Content B - 2 Column
    "content_2col_icon": "g324ba092b07_3_528", # 10 Content B - 2 Column w/ Icon Spot
    "content_3col": "g324ba092b07_3_536",    # 11 Content C - 3 Column
    "content_3col_icon": "g324ba092b07_3_546", # 12 Content C - 3 Column w/ Icon Spot
    "content_3col_cards": "g324ba092b07_3_556", # 13 Content C - 3 Column Cards
    "content_card_right": "g324ba092b07_3_589", # 14 Content D - Card Right Dark 1
    "content_card_left": "g324ba092b07_3_596",  # 15 Content D - Card Left Dark
    "content_card_large": "g324ba092b07_3_603", # 16 Content D - Card Large Dark
    "blank": "g3413a4b56ae_5_84",            # Content E - Blank
    "section_break_1": "g330123ec4c9_0_66",  # Content E - Section Break 1
    "section_break_2": "g330123ec4c9_0_71",  # Content E - Section Break 2
    "power_statement": "g324ba092b07_3_575", # Content E - Power Statement 1
    "closing": "g324ba092b07_3_379",         # Z - Closing Dark
}


def api_call(method: str, url: str, data: Optional[Dict] = None) -> Dict:
    """Make an API call using curl with retry logic."""
    return api_call_with_retry(method, url, data=data)


def generate_id() -> str:
    """Generate a unique object ID."""
    return f"obj_{uuid.uuid4().hex[:12]}"


def create_presentation(title: str) -> str:
    """Create a new Google Slides presentation and return its ID."""
    response = api_call(
        "POST",
        "https://slides.googleapis.com/v1/presentations",
        {"title": title}
    )

    if "error" in response:
        raise RuntimeError(f"Failed to create presentation: {response['error']['message']}")

    return response["presentationId"]


def get_presentation(pres_id: str) -> Dict:
    """Get full presentation structure."""
    return api_call("GET", f"https://slides.googleapis.com/v1/presentations/{pres_id}")


def batch_update(pres_id: str, requests: List[Dict]) -> Dict:
    """Execute a batchUpdate on the presentation."""
    return api_call(
        "POST",
        f"https://slides.googleapis.com/v1/presentations/{pres_id}:batchUpdate",
        {"requests": requests}
    )


def get_slide_ids(pres_id: str) -> List[str]:
    """Get list of slide IDs in presentation order."""
    pres = get_presentation(pres_id)
    return [slide["objectId"] for slide in pres.get("slides", [])]


def get_slide_elements(pres_id: str, page_id: str) -> List[Dict]:
    """Get all elements on a specific slide."""
    pres = get_presentation(pres_id)
    for slide in pres.get("slides", []):
        if slide["objectId"] == page_id:
            return slide.get("pageElements", [])
    return []


def find_placeholder(pres_id: str, page_id: str, placeholder_type: str) -> Optional[str]:
    """Find a placeholder element by type (TITLE, SUBTITLE, BODY, etc.)."""
    elements = get_slide_elements(pres_id, page_id)
    for elem in elements:
        shape = elem.get("shape", {})
        placeholder = shape.get("placeholder", {})
        if placeholder.get("type") == placeholder_type:
            return elem["objectId"]
    return None


def get_all_placeholders(pres_id: str, page_id: str) -> List[Dict]:
    """Get all placeholder elements on a slide with their info."""
    elements = get_slide_elements(pres_id, page_id)
    placeholders = []
    for elem in elements:
        shape = elem.get("shape", {})
        placeholder = shape.get("placeholder", {})
        if placeholder:
            placeholders.append({
                "objectId": elem["objectId"],
                "type": placeholder.get("type"),
                "index": placeholder.get("index"),
                "parentObjectId": placeholder.get("parentObjectId")
            })
    return placeholders


def get_text_content(pres_id: str, shape_id: str) -> str:
    """Get the text content of a shape."""
    pres = get_presentation(pres_id)
    for slide in pres.get("slides", []):
        for elem in slide.get("pageElements", []):
            if elem.get("objectId") == shape_id:
                shape = elem.get("shape", {})
                text = shape.get("text", {})
                content = ""
                for text_elem in text.get("textElements", []):
                    text_run = text_elem.get("textRun", {})
                    content += text_run.get("content", "")
                return content.strip()
    return ""


def inches_to_emu(inches: float) -> int:
    """Convert inches to EMU."""
    return int(inches * EMU_PER_INCH)


def pt_to_emu(pt: float) -> int:
    """Convert points to EMU."""
    return int(pt * EMU_PER_PT)


def get_position(position_name: str) -> Tuple[float, float, float, float]:
    """
    Get predefined position coordinates by name.

    Args:
        position_name: Name from POSITIONS dict (e.g., 'left_half', 'table_full')

    Returns:
        Tuple of (x, y, width, height) in inches

    Available positions:
        Full area: full, full_no_title
        Thirds: left_third, center_third, right_third
        Halves: left_half, right_half, top_half, bottom_half
        Quadrants: top_left, top_right, bottom_left, bottom_right
        Centered: center_large, center_medium, center_small
        Tables: table_full, table_left, table_right
        Charts: chart_full, chart_left, chart_right
        Images: image_left, image_right, image_center, image_background
        Text: text_title_area, text_subtitle, text_footer, text_caption
    """
    if position_name not in POSITIONS:
        available = ", ".join(sorted(POSITIONS.keys()))
        raise ValueError(f"Unknown position '{position_name}'. Available: {available}")
    return POSITIONS[position_name]


def get_size(size_name: str) -> Tuple[float, float]:
    """
    Get predefined size by name.

    Args:
        size_name: Name from SIZES dict (e.g., 'icon_small', 'logo_medium')

    Returns:
        Tuple of (width, height) in inches
    """
    if size_name not in SIZES:
        available = ", ".join(sorted(SIZES.keys()))
        raise ValueError(f"Unknown size '{size_name}'. Available: {available}")
    return SIZES[size_name]


def calculate_grid_position(
    row: int,
    col: int,
    rows: int,
    cols: int,
    padding: float = 0.1
) -> Tuple[float, float, float, float]:
    """
    Calculate position for an element in a grid layout.

    Args:
        row: Row index (0-based)
        col: Column index (0-based)
        rows: Total number of rows
        cols: Total number of columns
        padding: Padding between cells in inches

    Returns:
        Tuple of (x, y, width, height) in inches
    """
    cell_width = (CONTENT_WIDTH - (cols - 1) * padding) / cols
    cell_height = (BODY_HEIGHT - (rows - 1) * padding) / rows

    x = CONTENT_LEFT + col * (cell_width + padding)
    y = BODY_TOP + row * (cell_height + padding)

    return (x, y, cell_width, cell_height)


def list_positions() -> List[str]:
    """Return list of all available position names."""
    return sorted(POSITIONS.keys())


def list_sizes() -> List[str]:
    """Return list of all available size names."""
    return sorted(SIZES.keys())


# =============================================================================
# SLIDE OPERATIONS
# =============================================================================

def create_from_template(
    title: str,
    template_id: str = DATABRICKS_TEMPLATE_ID,
    delete_sample_slides: bool = True
) -> str:
    """
    Create a new presentation by copying a template.

    Args:
        title: Title for the new presentation
        template_id: Source template presentation ID
        delete_sample_slides: Whether to delete sample slides from template

    Returns:
        New presentation ID
    """
    # Copy the template using Drive API
    # Explicitly set parents to ["root"] so the copy lands in the user's
    # My Drive root instead of inheriting the template's parent folder.
    response = api_call(
        "POST",
        f"https://www.googleapis.com/drive/v3/files/{template_id}/copy",
        {"name": title, "parents": ["root"]}
    )

    if "error" in response:
        raise RuntimeError(f"Failed to copy template: {response['error']['message']}")

    new_pres_id = response["id"]

    # Optionally delete sample slides (keep only masters/layouts)
    if delete_sample_slides:
        slides = get_slide_ids(new_pres_id)
        # Delete all slides - we can't delete the last one, so leave one
        # We'll delete it when adding the first real slide
        if len(slides) > 1:
            for slide_id in slides[:-1]:
                batch_update(new_pres_id, [{"deleteObject": {"objectId": slide_id}}])

    return new_pres_id


def add_slide(
    pres_id: str,
    layout: str = "BLANK",
    insertion_index: Optional[int] = None,
    page_id: Optional[str] = None,
    layout_id: Optional[str] = None
) -> Dict:
    """
    Add a new slide to the presentation.

    Args:
        pres_id: Presentation ID
        layout: Predefined layout type (BLANK, TITLE, TITLE_AND_BODY, etc.)
                Used only if layout_id is not provided.
        insertion_index: Where to insert (None = end)
        page_id: Optional custom page ID
        layout_id: Custom layout object ID (for template layouts).
                   If provided, overrides the 'layout' parameter.

    Returns:
        API response with created slide info
    """
    page_id = page_id or generate_id()

    # Build the slide layout reference
    if layout_id:
        # Use custom layout ID (e.g., from Databricks template)
        slide_layout_ref = {"layoutId": layout_id}
    else:
        # Use predefined layout
        slide_layout_ref = {"predefinedLayout": layout}

    request = {
        "createSlide": {
            "objectId": page_id,
            "slideLayoutReference": slide_layout_ref
        }
    }

    if insertion_index is not None:
        request["createSlide"]["insertionIndex"] = insertion_index

    result = batch_update(pres_id, [request])

    if "error" not in result:
        result["pageId"] = page_id

    return result


def add_slide_from_template(
    pres_id: str,
    layout_name: str,
    theme: str = "light",
    insertion_index: Optional[int] = None,
    page_id: Optional[str] = None
) -> Dict:
    """
    Add a slide using a Databricks template layout by name.

    Args:
        pres_id: Presentation ID (must be created from Databricks template)
        layout_name: Layout name (e.g., 'title', 'content_basic', 'content_2col')
        theme: 'light' or 'dark'
        insertion_index: Where to insert (None = end)
        page_id: Optional custom page ID

    Available layout names:
        Title slides: title, title_alt, title_gradient, title_orange
        Content: content_basic, content_basic_white, content_2col, content_2col_icon,
                 content_3col, content_3col_icon, content_3col_cards,
                 content_card_right, content_card_left, content_card_large
        Section breaks: section_break_1 through section_break_8
        Special: blank, power_statement, power_statement_2, closing
        Industry: industry_media, industry_retail, industry_healthcare,
                  industry_manufacturing, industry_financial, industry_public,
                  industry_consumer

    Returns:
        API response with created slide info
    """
    layouts = DATABRICKS_LAYOUTS_DARK if theme == "dark" else DATABRICKS_LAYOUTS_LIGHT

    if layout_name not in layouts:
        available = ", ".join(sorted(layouts.keys()))
        raise ValueError(f"Unknown layout '{layout_name}'. Available: {available}")

    layout_id = layouts[layout_name]
    return add_slide(pres_id, layout_id=layout_id, insertion_index=insertion_index, page_id=page_id)


def list_layouts(pres_id: str) -> List[Dict]:
    """Get all available layouts in a presentation."""
    pres = get_presentation(pres_id)
    layouts = []
    for layout in pres.get("layouts", []):
        props = layout.get("layoutProperties", {})
        layouts.append({
            "objectId": layout["objectId"],
            "displayName": props.get("displayName", "unnamed"),
            "masterObjectId": props.get("masterObjectId")
        })
    return layouts


def find_layout_by_name(
    pres_id: str,
    name: str,
    fuzzy: bool = True,
    preferred_master: Optional[str] = None
) -> Optional[str]:
    """
    Find a layout ID by its display name.

    Args:
        pres_id: Presentation ID
        name: Layout display name to search for
        fuzzy: If True, matches if name is contained in display name (case-insensitive)
               If False, requires exact match
        preferred_master: If provided, prefer layouts from this master.
                          Use "light" for Databricks light theme (g324ba092b07_3_0)
                          Use "dark" for Databricks dark theme (g324ba092b07_3_358)

    Returns:
        Layout object ID if found, None otherwise
    """
    layouts = list_layouts(pres_id)
    name_lower = name.lower()

    # Resolve preferred master shorthand
    master_ids = {
        "light": "g324ba092b07_3_0",
        "dark": "g324ba092b07_3_358"
    }
    if preferred_master in master_ids:
        preferred_master = master_ids[preferred_master]

    # First try exact match (preferring layouts from specified master)
    exact_matches = []
    for layout in layouts:
        if layout["displayName"].lower() == name_lower:
            exact_matches.append(layout)

    if exact_matches:
        if preferred_master:
            for m in exact_matches:
                if m["masterObjectId"] == preferred_master:
                    return m["objectId"]
        # If preferred_master was specified but not found in exact matches,
        # don't return a layout from the wrong master - fall through to fuzzy
        if not preferred_master:
            return exact_matches[0]["objectId"]

    # If fuzzy matching, try partial match
    if fuzzy:
        fuzzy_matches = []
        for layout in layouts:
            if name_lower in layout["displayName"].lower():
                fuzzy_matches.append(layout)

        if fuzzy_matches:
            if preferred_master:
                for m in fuzzy_matches:
                    if m["masterObjectId"] == preferred_master:
                        return m["objectId"]
            return fuzzy_matches[0]["objectId"]

    return None


# Databricks layout display name mappings (for use with find_layout_by_name)
DATABRICKS_LAYOUT_NAMES = {
    # Light theme titles
    "title": "3 Title Slide B - Light",
    "title_light_2": "3 Title Slide B - Light 2",
    "title_light_3": "3 Title Slide B - Light 3",

    # Dark theme titles
    "title_dark": "3 Title Slide B - Dark",

    # Content layouts
    "content_basic": "7 Content A - Basic",
    "content_2col": "9 Content B - 2 Column",
    "content_2col_icon": "10 Content B - 2 Column w/ Icon Spot",
    "content_3col": "11 Content C - 3 Column",
    "content_3col_icon": "12 Content C - 3 Column w/ Icon Spot",
    "content_3col_cards": "13 Content C - 3 Column Cards",
    "content_card_right": "14 Content D - Card Right",
    "content_card_left": "15 Content D - Card Left",
    "content_card_large": "16 Content D - Card Large",

    # Section breaks
    "section_break_1": "Content E - Section Break 1",
    "section_break_2": "Content E - Section Break 2",
    "section_break_3": "Content E - Section Break 3",
    "section_break_4": "Content E - Section Break 4",
    "section_break_5": "Content E - Section Break 5",
    "section_break_6": "Content E - Section Break 6",
    "section_break_7": "Content E - Section Break 7",
    "section_break_8": "Content E - Section Break 8",

    # Special layouts
    "blank": "Content E - Blank",
    "power_statement": "Content E - Power Statement",
    "closing_light": "Z - Closing Light",
    "closing_dark": "Z - Closing Dark",
}


def add_template_slide_by_name(
    pres_id: str,
    layout_name: str,
    insertion_index: Optional[int] = None,
    page_id: Optional[str] = None,
    theme: str = "light"
) -> Dict:
    """
    Add a slide using a Databricks template layout by name.

    This function looks up the layout ID dynamically from the presentation,
    which works correctly even when the presentation was copied from a template
    (where layout IDs change but display names stay the same).

    Args:
        pres_id: Presentation ID
        layout_name: Either a key from DATABRICKS_LAYOUT_NAMES or a display name
        insertion_index: Where to insert (None = end)
        page_id: Optional custom page ID
        theme: "light" or "dark" - prefer layouts from this theme's master

    Returns:
        API response with created slide info
    """
    # First check if it's a shorthand name
    search_name = DATABRICKS_LAYOUT_NAMES.get(layout_name, layout_name)

    # Find the layout ID in this presentation, preferring the specified theme
    layout_id = find_layout_by_name(pres_id, search_name, preferred_master=theme)

    if not layout_id:
        # List available layouts in error message
        layouts = list_layouts(pres_id)
        available = [l["displayName"] for l in layouts if not l["displayName"].startswith("Title slide")]
        return {
            "error": {
                "message": f"Layout '{layout_name}' not found. Available layouts: {', '.join(available[:10])}..."
            }
        }

    return add_slide(pres_id, layout_id=layout_id, insertion_index=insertion_index, page_id=page_id)


def duplicate_slide(pres_id: str, page_id: str, new_page_id: Optional[str] = None) -> Dict:
    """
    Duplicate a slide within the same presentation.

    Args:
        pres_id: Presentation ID
        page_id: ID of slide to duplicate
        new_page_id: Optional ID for the new slide

    Returns:
        API response
    """
    new_page_id = new_page_id or generate_id()

    request = {
        "duplicateObject": {
            "objectId": page_id,
            "objectIds": {
                page_id: new_page_id
            }
        }
    }

    result = batch_update(pres_id, [request])
    if "error" not in result:
        result["newPageId"] = new_page_id

    return result


def delete_slide(pres_id: str, page_id: str) -> Dict:
    """Delete a slide from the presentation."""
    request = {
        "deleteObject": {
            "objectId": page_id
        }
    }
    return batch_update(pres_id, [request])


def move_slides(pres_id: str, slide_ids: List[str], insertion_index: int) -> Dict:
    """Move slides to a new position."""
    request = {
        "updateSlidesPosition": {
            "slideObjectIds": slide_ids,
            "insertionIndex": insertion_index
        }
    }
    return batch_update(pres_id, [request])


def set_slide_background(
    pres_id: str,
    page_id: str,
    color: Optional[Dict] = None,
    image_url: Optional[str] = None
) -> Dict:
    """
    Set slide background color or image.

    Args:
        pres_id: Presentation ID
        page_id: Slide ID
        color: RGB color dict {"red": 0-1, "green": 0-1, "blue": 0-1}
        image_url: URL of background image

    Returns:
        API response
    """
    page_properties = {}

    if color:
        page_properties["pageBackgroundFill"] = {
            "solidFill": {
                "color": {"rgbColor": color}
            }
        }
    elif image_url:
        page_properties["pageBackgroundFill"] = {
            "stretchedPictureFill": {
                "contentUrl": image_url
            }
        }

    request = {
        "updatePageProperties": {
            "objectId": page_id,
            "pageProperties": page_properties,
            "fields": "pageBackgroundFill"
        }
    }

    return batch_update(pres_id, [request])


# =============================================================================
# TEXT OPERATIONS
# =============================================================================

def insert_text(pres_id: str, shape_id: str, text: str, index: int = 0) -> Dict:
    """Insert text into a shape or placeholder."""
    request = {
        "insertText": {
            "objectId": shape_id,
            "text": text,
            "insertionIndex": index
        }
    }
    return batch_update(pres_id, [request])


def delete_text(pres_id: str, shape_id: str, start_index: int = 0, end_index: Optional[int] = None) -> Dict:
    """Delete text from a shape."""
    text_range = {"type": "FROM_START_INDEX", "startIndex": start_index}
    if end_index is not None:
        text_range = {
            "type": "FIXED_RANGE",
            "startIndex": start_index,
            "endIndex": end_index
        }

    request = {
        "deleteText": {
            "objectId": shape_id,
            "textRange": text_range
        }
    }
    return batch_update(pres_id, [request])


def replace_all_text(
    pres_id: str,
    find: str,
    replace: str,
    match_case: bool = False,
    page_ids: Optional[List[str]] = None
) -> Dict:
    """
    Replace all occurrences of text across the presentation.

    Args:
        pres_id: Presentation ID
        find: Text to find
        replace: Replacement text
        match_case: Whether to match case
        page_ids: Optional list of page IDs to limit replacement to

    Returns:
        API response with number of occurrences replaced
    """
    request = {
        "replaceAllText": {
            "containsText": {
                "text": find,
                "matchCase": match_case
            },
            "replaceText": replace
        }
    }

    if page_ids:
        request["replaceAllText"]["pageObjectIds"] = page_ids

    return batch_update(pres_id, [request])


def replace_shape_text(
    pres_id: str,
    shape_id: str,
    new_text: str,
    preserve_style: bool = True
) -> Dict:
    """
    Replace all text in a shape with new text.

    Args:
        pres_id: Presentation ID
        shape_id: Shape object ID
        new_text: New text to set
        preserve_style: If True, preserves existing text style (font, size, color)

    Returns:
        API response
    """
    # Check if shape has existing text
    existing_text = get_text_content(pres_id, shape_id)

    requests = []

    # Only delete if there's existing text (avoid error on empty shapes)
    if existing_text:
        requests.append({
            "deleteText": {
                "objectId": shape_id,
                "textRange": {"type": "ALL"}
            }
        })

    # Insert new text
    requests.append({
        "insertText": {
            "objectId": shape_id,
            "text": new_text,
            "insertionIndex": 0
        }
    })

    return batch_update(pres_id, requests)


def set_placeholder_text(pres_id: str, page_id: str, placeholder_type: str, text: str) -> Dict:
    """
    Set text in a placeholder (TITLE, SUBTITLE, BODY).

    Args:
        pres_id: Presentation ID
        page_id: Slide ID
        placeholder_type: TITLE, SUBTITLE, BODY, CENTERED_TITLE
        text: Text to insert

    Returns:
        API response
    """
    shape_id = find_placeholder(pres_id, page_id, placeholder_type)
    if not shape_id:
        raise RuntimeError(f"Placeholder {placeholder_type} not found on slide {page_id}")

    return replace_shape_text(pres_id, shape_id, text)


def update_text_style(
    pres_id: str,
    shape_id: str,
    start_index: int,
    end_index: int,
    bold: Optional[bool] = None,
    italic: Optional[bool] = None,
    underline: Optional[bool] = None,
    strikethrough: Optional[bool] = None,
    font_size: Optional[float] = None,
    font_family: Optional[str] = None,
    foreground_color: Optional[Dict] = None,
    link_url: Optional[str] = None
) -> Dict:
    """Update text style for a range of text."""
    style = {}
    fields = []

    if bold is not None:
        style["bold"] = bold
        fields.append("bold")

    if italic is not None:
        style["italic"] = italic
        fields.append("italic")

    if underline is not None:
        style["underline"] = underline
        fields.append("underline")

    if strikethrough is not None:
        style["strikethrough"] = strikethrough
        fields.append("strikethrough")

    if font_size is not None:
        style["fontSize"] = {"magnitude": font_size, "unit": "PT"}
        fields.append("fontSize")

    if font_family is not None:
        style["fontFamily"] = font_family
        fields.append("fontFamily")

    if foreground_color is not None:
        style["foregroundColor"] = {"opaqueColor": {"rgbColor": foreground_color}}
        fields.append("foregroundColor")

    if link_url is not None:
        style["link"] = {"url": link_url}
        fields.append("link")

    request = {
        "updateTextStyle": {
            "objectId": shape_id,
            "textRange": {
                "type": "FIXED_RANGE",
                "startIndex": start_index,
                "endIndex": end_index
            },
            "style": style,
            "fields": ",".join(fields)
        }
    }

    return batch_update(pres_id, [request])


def create_bullets(
    pres_id: str,
    shape_id: str,
    start_index: int,
    end_index: int,
    preset: str = "BULLET_DISC_CIRCLE_SQUARE"
) -> Dict:
    """Create bullet points in a text range."""
    request = {
        "createParagraphBullets": {
            "objectId": shape_id,
            "textRange": {
                "type": "FIXED_RANGE",
                "startIndex": start_index,
                "endIndex": end_index
            },
            "bulletPreset": preset
        }
    }
    return batch_update(pres_id, [request])


# =============================================================================
# SHAPE OPERATIONS
# =============================================================================

def create_shape(
    pres_id: str,
    page_id: str,
    shape_type: str,
    x: float,
    y: float,
    width: float,
    height: float,
    shape_id: Optional[str] = None
) -> Dict:
    """
    Create a shape on a slide.

    Args:
        pres_id: Presentation ID
        page_id: Slide ID
        shape_type: RECTANGLE, ELLIPSE, TEXT_BOX, etc.
        x, y: Position in inches from top-left
        width, height: Size in inches
        shape_id: Optional custom shape ID

    Returns:
        API response
    """
    shape_id = shape_id or generate_id()

    request = {
        "createShape": {
            "objectId": shape_id,
            "shapeType": shape_type,
            "elementProperties": {
                "pageObjectId": page_id,
                "size": {
                    "width": {"magnitude": inches_to_emu(width), "unit": "EMU"},
                    "height": {"magnitude": inches_to_emu(height), "unit": "EMU"}
                },
                "transform": {
                    "scaleX": 1,
                    "scaleY": 1,
                    "translateX": inches_to_emu(x),
                    "translateY": inches_to_emu(y),
                    "unit": "EMU"
                }
            }
        }
    }

    result = batch_update(pres_id, [request])
    if "error" not in result:
        result["shapeId"] = shape_id

    return result


def create_text_box(
    pres_id: str,
    page_id: str,
    text: str,
    x: float,
    y: float,
    width: float,
    height: float,
    font_size: float = 18,
    bold: bool = False,
    font_color: Optional[Dict] = None,
    text_box_id: Optional[str] = None
) -> Dict:
    """
    Create a text box with text on a slide.

    Args:
        pres_id: Presentation ID
        page_id: Slide ID
        text: Text content
        x, y: Position in inches
        width, height: Size in inches
        font_size: Font size in points
        bold: Whether text is bold
        font_color: Optional RGB color dict
        text_box_id: Optional custom ID

    Returns:
        API response with textBoxId
    """
    text_box_id = text_box_id or generate_id()

    style = {
        "fontSize": {"magnitude": font_size, "unit": "PT"},
        "bold": bold
    }
    fields = ["fontSize", "bold"]

    if font_color:
        style["foregroundColor"] = {"opaqueColor": {"rgbColor": font_color}}
        fields.append("foregroundColor")

    # Create shape and insert text in one batch
    requests = [
        {
            "createShape": {
                "objectId": text_box_id,
                "shapeType": "TEXT_BOX",
                "elementProperties": {
                    "pageObjectId": page_id,
                    "size": {
                        "width": {"magnitude": inches_to_emu(width), "unit": "EMU"},
                        "height": {"magnitude": inches_to_emu(height), "unit": "EMU"}
                    },
                    "transform": {
                        "scaleX": 1,
                        "scaleY": 1,
                        "translateX": inches_to_emu(x),
                        "translateY": inches_to_emu(y),
                        "unit": "EMU"
                    }
                }
            }
        },
        {
            "insertText": {
                "objectId": text_box_id,
                "text": text,
                "insertionIndex": 0
            }
        },
        {
            "updateTextStyle": {
                "objectId": text_box_id,
                "textRange": {"type": "ALL"},
                "style": style,
                "fields": ",".join(fields)
            }
        }
    ]

    result = batch_update(pres_id, requests)
    if "error" not in result:
        result["textBoxId"] = text_box_id

    return result


def update_shape_properties(
    pres_id: str,
    shape_id: str,
    fill_color: Optional[Dict] = None,
    outline_color: Optional[Dict] = None,
    outline_weight: Optional[float] = None
) -> Dict:
    """Update shape fill and outline properties."""
    properties = {}
    fields = []

    if fill_color is not None:
        properties["shapeBackgroundFill"] = {
            "solidFill": {"color": {"rgbColor": fill_color}}
        }
        fields.append("shapeBackgroundFill")

    if outline_color is not None or outline_weight is not None:
        outline = {}
        if outline_color is not None:
            outline["outlineFill"] = {
                "solidFill": {"color": {"rgbColor": outline_color}}
            }
            fields.append("outline.outlineFill")
        if outline_weight is not None:
            outline["weight"] = {"magnitude": outline_weight, "unit": "PT"}
            fields.append("outline.weight")
        properties["outline"] = outline

    request = {
        "updateShapeProperties": {
            "objectId": shape_id,
            "shapeProperties": properties,
            "fields": ",".join(fields)
        }
    }

    return batch_update(pres_id, [request])


# =============================================================================
# IMAGE OPERATIONS
# =============================================================================

def create_image(
    pres_id: str,
    page_id: str,
    image_url: str,
    x: float,
    y: float,
    width: float,
    height: float,
    image_id: Optional[str] = None
) -> Dict:
    """
    Insert an image on a slide.

    Args:
        pres_id: Presentation ID
        page_id: Slide ID
        image_url: URL of the image
        x, y: Position in inches
        width, height: Size in inches
        image_id: Optional custom ID

    Returns:
        API response with imageId
    """
    image_id = image_id or generate_id()

    request = {
        "createImage": {
            "objectId": image_id,
            "url": image_url,
            "elementProperties": {
                "pageObjectId": page_id,
                "size": {
                    "width": {"magnitude": inches_to_emu(width), "unit": "EMU"},
                    "height": {"magnitude": inches_to_emu(height), "unit": "EMU"}
                },
                "transform": {
                    "scaleX": 1,
                    "scaleY": 1,
                    "translateX": inches_to_emu(x),
                    "translateY": inches_to_emu(y),
                    "unit": "EMU"
                }
            }
        }
    }

    result = batch_update(pres_id, [request])
    if "error" not in result:
        result["imageId"] = image_id

    return result


def replace_image(
    pres_id: str,
    image_id: str,
    new_image_url: str
) -> Dict:
    """
    Replace an existing image with a new one.

    Args:
        pres_id: Presentation ID
        image_id: Existing image object ID
        new_image_url: URL of the new image

    Returns:
        API response
    """
    request = {
        "replaceImage": {
            "imageObjectId": image_id,
            "url": new_image_url,
            "imageReplaceMethod": "CENTER_INSIDE"
        }
    }

    return batch_update(pres_id, [request])


# =============================================================================
# TABLE OPERATIONS
# =============================================================================

def create_table(
    pres_id: str,
    page_id: str,
    rows: int,
    cols: int,
    x: float,
    y: float,
    width: float,
    height: float,
    table_id: Optional[str] = None
) -> Dict:
    """
    Create a table on a slide.

    Args:
        pres_id: Presentation ID
        page_id: Slide ID
        rows: Number of rows
        cols: Number of columns
        x, y: Position in inches
        width, height: Size in inches
        table_id: Optional custom ID

    Returns:
        API response with tableId
    """
    table_id = table_id or generate_id()

    request = {
        "createTable": {
            "objectId": table_id,
            "rows": rows,
            "columns": cols,
            "elementProperties": {
                "pageObjectId": page_id,
                "size": {
                    "width": {"magnitude": inches_to_emu(width), "unit": "EMU"},
                    "height": {"magnitude": inches_to_emu(height), "unit": "EMU"}
                },
                "transform": {
                    "scaleX": 1,
                    "scaleY": 1,
                    "translateX": inches_to_emu(x),
                    "translateY": inches_to_emu(y),
                    "unit": "EMU"
                }
            }
        }
    }

    result = batch_update(pres_id, [request])
    if "error" not in result:
        result["tableId"] = table_id

    return result


def fill_table(
    pres_id: str,
    table_id: str,
    data: List[List[str]],
    header_bold: bool = True
) -> Dict:
    """
    Fill a table with data.

    Args:
        pres_id: Presentation ID
        table_id: Table object ID
        data: 2D array of cell values
        header_bold: Bold the first row

    Returns:
        API response
    """
    requests = []

    for row_idx, row in enumerate(data):
        for col_idx, cell_text in enumerate(row):
            if cell_text:
                # Insert text into cell
                requests.append({
                    "insertText": {
                        "objectId": table_id,
                        "cellLocation": {
                            "rowIndex": row_idx,
                            "columnIndex": col_idx
                        },
                        "text": str(cell_text),
                        "insertionIndex": 0
                    }
                })

                # Bold header row
                if header_bold and row_idx == 0:
                    requests.append({
                        "updateTextStyle": {
                            "objectId": table_id,
                            "cellLocation": {
                                "rowIndex": row_idx,
                                "columnIndex": col_idx
                            },
                            "textRange": {"type": "ALL"},
                            "style": {"bold": True},
                            "fields": "bold"
                        }
                    })

    return batch_update(pres_id, requests)


def style_table_header(
    pres_id: str,
    table_id: str,
    cols: int,
    bg_color: Dict = None,
    text_color: Dict = None
) -> Dict:
    """
    Style the header row of a table with background and text color.

    Args:
        pres_id: Presentation ID
        table_id: Table object ID
        cols: Number of columns
        bg_color: Background color (default: Databricks navy)
        text_color: Text color (default: white)

    Returns:
        API response
    """
    if bg_color is None:
        bg_color = DATABRICKS_COLORS["navy"]
    if text_color is None:
        text_color = DATABRICKS_COLORS["white"]

    requests = []

    for col_idx in range(cols):
        # Set background color
        requests.append({
            "updateTableCellProperties": {
                "objectId": table_id,
                "tableRange": {
                    "location": {"rowIndex": 0, "columnIndex": col_idx},
                    "rowSpan": 1,
                    "columnSpan": 1
                },
                "tableCellProperties": {
                    "tableCellBackgroundFill": {
                        "solidFill": {"color": {"rgbColor": bg_color}}
                    }
                },
                "fields": "tableCellBackgroundFill"
            }
        })

        # Set text color to white (or specified color)
        requests.append({
            "updateTextStyle": {
                "objectId": table_id,
                "cellLocation": {"rowIndex": 0, "columnIndex": col_idx},
                "textRange": {"type": "ALL"},
                "style": {
                    "foregroundColor": {"opaqueColor": {"rgbColor": text_color}},
                    "bold": True
                },
                "fields": "foregroundColor,bold"
            }
        })

    return batch_update(pres_id, requests)


def style_table_cell(
    pres_id: str,
    table_id: str,
    row: int,
    col: int,
    bg_color: Optional[Dict] = None,
    bold: Optional[bool] = None,
    font_color: Optional[Dict] = None
) -> Dict:
    """Style a specific table cell."""
    requests = []

    if bg_color:
        requests.append({
            "updateTableCellProperties": {
                "objectId": table_id,
                "tableRange": {
                    "location": {"rowIndex": row, "columnIndex": col},
                    "rowSpan": 1,
                    "columnSpan": 1
                },
                "tableCellProperties": {
                    "tableCellBackgroundFill": {
                        "solidFill": {"color": {"rgbColor": bg_color}}
                    }
                },
                "fields": "tableCellBackgroundFill"
            }
        })

    style = {}
    fields = []
    if bold is not None:
        style["bold"] = bold
        fields.append("bold")
    if font_color:
        style["foregroundColor"] = {"opaqueColor": {"rgbColor": font_color}}
        fields.append("foregroundColor")

    if style:
        requests.append({
            "updateTextStyle": {
                "objectId": table_id,
                "cellLocation": {"rowIndex": row, "columnIndex": col},
                "textRange": {"type": "ALL"},
                "style": style,
                "fields": ",".join(fields)
            }
        })

    return batch_update(pres_id, requests) if requests else {"status": "no_changes"}


def style_table_body_text(
    pres_id: str,
    table_id: str,
    rows: int,
    cols: int,
    text_color: Dict = None,
    start_row: int = 1
) -> Dict:
    """
    Style text color for table body cells (non-header rows).

    This is useful for dark backgrounds where body text needs to be white/light.

    Args:
        pres_id: Presentation ID
        table_id: Table object ID
        rows: Total number of rows in table
        cols: Number of columns
        text_color: Text color (default: white)
        start_row: First row to style (default: 1, skips header)

    Returns:
        API response
    """
    if text_color is None:
        text_color = DATABRICKS_COLORS["white"]

    requests = []

    for row_idx in range(start_row, rows):
        for col_idx in range(cols):
            requests.append({
                "updateTextStyle": {
                    "objectId": table_id,
                    "cellLocation": {"rowIndex": row_idx, "columnIndex": col_idx},
                    "textRange": {"type": "ALL"},
                    "style": {
                        "foregroundColor": {"opaqueColor": {"rgbColor": text_color}}
                    },
                    "fields": "foregroundColor"
                }
            })

    return batch_update(pres_id, requests) if requests else {"status": "no_changes"}


def style_table_for_dark_background(
    pres_id: str,
    table_id: str,
    rows: int,
    cols: int,
    header_bg_color: Dict = None,
    text_color: Dict = None
) -> Dict:
    """
    Style a table for dark slide backgrounds.

    Sets header with Databricks red background and white text,
    and body cells with white text.

    Args:
        pres_id: Presentation ID
        table_id: Table object ID
        rows: Total number of rows
        cols: Number of columns
        header_bg_color: Header background color (default: Databricks red)
        text_color: Text color for all cells (default: white)

    Returns:
        API response
    """
    if header_bg_color is None:
        header_bg_color = DATABRICKS_COLORS["red"]
    if text_color is None:
        text_color = DATABRICKS_COLORS["white"]

    requests = []

    # Style header row with background and white text
    for col_idx in range(cols):
        # Header background
        requests.append({
            "updateTableCellProperties": {
                "objectId": table_id,
                "tableRange": {
                    "location": {"rowIndex": 0, "columnIndex": col_idx},
                    "rowSpan": 1,
                    "columnSpan": 1
                },
                "tableCellProperties": {
                    "tableCellBackgroundFill": {
                        "solidFill": {"color": {"rgbColor": header_bg_color}}
                    }
                },
                "fields": "tableCellBackgroundFill"
            }
        })

        # Header text style (bold + white)
        requests.append({
            "updateTextStyle": {
                "objectId": table_id,
                "cellLocation": {"rowIndex": 0, "columnIndex": col_idx},
                "textRange": {"type": "ALL"},
                "style": {
                    "foregroundColor": {"opaqueColor": {"rgbColor": text_color}},
                    "bold": True
                },
                "fields": "foregroundColor,bold"
            }
        })

    # Style body rows with white text
    for row_idx in range(1, rows):
        for col_idx in range(cols):
            requests.append({
                "updateTextStyle": {
                    "objectId": table_id,
                    "cellLocation": {"rowIndex": row_idx, "columnIndex": col_idx},
                    "textRange": {"type": "ALL"},
                    "style": {
                        "foregroundColor": {"opaqueColor": {"rgbColor": text_color}}
                    },
                    "fields": "foregroundColor"
                }
            })

    return batch_update(pres_id, requests)


# =============================================================================
# CHART OPERATIONS (requires Google Sheets)
# =============================================================================

def create_sheets_chart(
    pres_id: str,
    page_id: str,
    spreadsheet_id: str,
    chart_id: int,
    x: float,
    y: float,
    width: float,
    height: float,
    linked: bool = True,
    obj_id: Optional[str] = None
) -> Dict:
    """
    Embed a chart from Google Sheets.

    Args:
        pres_id: Presentation ID
        page_id: Slide ID
        spreadsheet_id: Google Sheets spreadsheet ID
        chart_id: Chart ID within the spreadsheet
        x, y: Position in inches
        width, height: Size in inches
        linked: If True, chart updates when sheet changes
        obj_id: Optional custom object ID

    Returns:
        API response with chartId
    """
    obj_id = obj_id or generate_id()

    request = {
        "createSheetsChart": {
            "objectId": obj_id,
            "spreadsheetId": spreadsheet_id,
            "chartId": chart_id,
            "linkingMode": "LINKED" if linked else "NOT_LINKED_IMAGE",
            "elementProperties": {
                "pageObjectId": page_id,
                "size": {
                    "width": {"magnitude": inches_to_emu(width), "unit": "EMU"},
                    "height": {"magnitude": inches_to_emu(height), "unit": "EMU"}
                },
                "transform": {
                    "scaleX": 1,
                    "scaleY": 1,
                    "translateX": inches_to_emu(x),
                    "translateY": inches_to_emu(y),
                    "unit": "EMU"
                }
            }
        }
    }

    result = batch_update(pres_id, [request])
    if "error" not in result:
        result["chartId"] = obj_id

    return result


def refresh_chart(pres_id: str, chart_id: str) -> Dict:
    """Refresh a linked Sheets chart."""
    request = {
        "refreshSheetsChart": {
            "objectId": chart_id
        }
    }
    return batch_update(pres_id, [request])


# =============================================================================
# COPY OPERATIONS
# =============================================================================

def copy_presentation(pres_id: str, new_title: str) -> str:
    """
    Copy an entire presentation using Drive API.

    Args:
        pres_id: Source presentation ID
        new_title: Title for the new presentation

    Returns:
        New presentation ID
    """
    # Explicitly set parents to ["root"] so the copy lands in the user's
    # My Drive root instead of inheriting the source file's parent folder.
    response = api_call(
        "POST",
        f"https://www.googleapis.com/drive/v3/files/{pres_id}/copy",
        {"name": new_title, "parents": ["root"]}
    )

    if "error" in response:
        raise RuntimeError(f"Failed to copy presentation: {response['error']['message']}")

    return response["id"]


def import_slides_from_presentation(
    target_pres_id: str,
    source_pres_id: str,
    slide_ids: Optional[List[str]] = None,
    insertion_index: Optional[int] = None
) -> Dict:
    """
    Import slides from another presentation.

    Note: This creates a copy of the source presentation, extracts slides,
    and then deletes unwanted slides. Theme is NOT preserved.

    Args:
        target_pres_id: Destination presentation ID
        source_pres_id: Source presentation ID
        slide_ids: List of slide IDs to import (None = all)
        insertion_index: Where to insert slides (None = end)

    Returns:
        Dict with imported slide IDs
    """
    # Get source presentation structure
    source_pres = get_presentation(source_pres_id)
    source_slides = source_pres.get("slides", [])

    if not source_slides:
        raise RuntimeError("Source presentation has no slides")

    # Filter slides if specific IDs requested
    if slide_ids:
        source_slides = [s for s in source_slides if s["objectId"] in slide_ids]

    # Get target presentation to find insertion point
    target_slides = get_slide_ids(target_pres_id)
    insert_at = insertion_index if insertion_index is not None else len(target_slides)

    # We need to recreate each slide manually since there's no direct import API
    # This is a simplified version - for complex slides, consider Apps Script
    imported_ids = []

    for slide_data in source_slides:
        # Create a blank slide
        new_id = generate_id()
        result = add_slide(target_pres_id, "BLANK", insert_at, new_id)

        if "error" in result:
            continue

        imported_ids.append(new_id)
        insert_at += 1

        # Copy elements (simplified - full copy would need all element types)
        # For complete slide copying, use Apps Script or copy entire presentation

    return {"importedSlideIds": imported_ids, "count": len(imported_ids)}


# =============================================================================
# TEMPLATE UTILITIES
# =============================================================================

def create_presentation_from_spec(
    title: str,
    slides: List[Dict],
    template_id: str = DATABRICKS_TEMPLATE_ID,
    theme: str = "light"
) -> Dict:
    """
    Create a complete presentation from a specification.

    Args:
        title: Presentation title
        slides: List of slide specs, each containing:
            - layout: Layout name (e.g., "title", "content_basic")
            - title: Optional slide title text
            - body: Optional body text
            - replacements: Optional dict of text replacements {find: replace}
        template_id: Template to use
        theme: "light" or "dark"

    Returns:
        Dict with presentationId and slideIds

    Example:
        slides = [
            {"layout": "title", "title": "My Presentation", "body": "Subtitle"},
            {"layout": "content_basic", "title": "Overview", "body": "Key points..."},
            {"layout": "closing"}
        ]
        result = create_presentation_from_spec("Demo Deck", slides)
    """
    # Create from template
    pres_id = create_from_template(title, template_id)

    # Delete the initial sample slide
    initial_slides = get_slide_ids(pres_id)
    if initial_slides:
        delete_slide(pres_id, initial_slides[0])

    slide_ids = []

    for spec in slides:
        layout_name = spec.get("layout", "content_basic")

        # Add the slide
        result = add_slide_from_template(pres_id, layout_name, theme=theme)

        if "error" in result:
            print(f"Warning: Failed to add slide with layout '{layout_name}': {result['error']}")
            continue

        page_id = result.get("pageId")
        if not page_id:
            continue

        slide_ids.append(page_id)

        # Set title if provided
        if "title" in spec:
            try:
                set_placeholder_text(pres_id, page_id, "TITLE", spec["title"])
            except RuntimeError:
                # Some layouts may not have a title placeholder
                pass

        # Set body if provided
        if "body" in spec:
            try:
                set_placeholder_text(pres_id, page_id, "BODY", spec["body"])
            except RuntimeError:
                # Some layouts may not have a body placeholder
                pass

        # Apply text replacements
        if "replacements" in spec:
            for find, replace in spec["replacements"].items():
                replace_all_text(pres_id, find, replace, page_ids=[page_id])

    return {
        "presentationId": pres_id,
        "url": f"https://docs.google.com/presentation/d/{pres_id}/edit",
        "slideIds": slide_ids
    }


# =============================================================================
# MAIN CLI
# =============================================================================

def main():
    parser = argparse.ArgumentParser(
        description="Build and manage Google Slides presentations",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    subparsers = parser.add_subparsers(dest="command", required=True)

    # Create presentation
    create_parser = subparsers.add_parser("create", help="Create a new presentation")
    create_parser.add_argument("--title", required=True, help="Presentation title")

    # Get presentation info
    info_parser = subparsers.add_parser("info", help="Get presentation info")
    info_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    info_parser.add_argument("--full", action="store_true", help="Show full JSON")

    # List slides
    list_parser = subparsers.add_parser("list-slides", help="List all slides")
    list_parser.add_argument("--pres-id", required=True, help="Presentation ID")

    # Add slide
    slide_parser = subparsers.add_parser("add-slide", help="Add a new slide")
    slide_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    slide_parser.add_argument("--layout", default="BLANK", help="Layout type")
    slide_parser.add_argument("--index", type=int, help="Insertion index")

    # Duplicate slide
    dup_parser = subparsers.add_parser("duplicate-slide", help="Duplicate a slide")
    dup_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    dup_parser.add_argument("--page-id", required=True, help="Slide ID to duplicate")

    # Delete slide
    del_parser = subparsers.add_parser("delete-slide", help="Delete a slide")
    del_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    del_parser.add_argument("--page-id", required=True, help="Slide ID to delete")

    # Set background
    bg_parser = subparsers.add_parser("set-background", help="Set slide background")
    bg_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    bg_parser.add_argument("--page-id", required=True, help="Slide ID")
    bg_parser.add_argument("--color", help="RGB color as JSON")
    bg_parser.add_argument("--image-url", help="Background image URL")

    # Add text box
    text_parser = subparsers.add_parser("add-text-box", help="Add a text box")
    text_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    text_parser.add_argument("--page-id", required=True, help="Slide ID")
    text_parser.add_argument("--text", required=True, help="Text content")
    text_parser.add_argument("--x", type=float, default=1, help="X position (inches)")
    text_parser.add_argument("--y", type=float, default=1, help="Y position (inches)")
    text_parser.add_argument("--width", type=float, default=3, help="Width (inches)")
    text_parser.add_argument("--height", type=float, default=1, help="Height (inches)")
    text_parser.add_argument("--font-size", type=float, default=18, help="Font size (pt)")
    text_parser.add_argument("--bold", action="store_true", help="Bold text")

    # Add image
    img_parser = subparsers.add_parser("add-image", help="Add an image")
    img_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    img_parser.add_argument("--page-id", required=True, help="Slide ID")
    img_parser.add_argument("--url", required=True, help="Image URL")
    img_parser.add_argument("--x", type=float, default=1, help="X position (inches)")
    img_parser.add_argument("--y", type=float, default=1, help="Y position (inches)")
    img_parser.add_argument("--width", type=float, default=3, help="Width (inches)")
    img_parser.add_argument("--height", type=float, default=2, help="Height (inches)")

    # Add table
    table_parser = subparsers.add_parser("add-table", help="Add a table")
    table_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    table_parser.add_argument("--page-id", required=True, help="Slide ID")
    table_parser.add_argument("--rows", type=int, required=True, help="Number of rows")
    table_parser.add_argument("--cols", type=int, required=True, help="Number of columns")
    table_parser.add_argument("--data", required=True, help="JSON 2D array of cell values")
    table_parser.add_argument("--position", help="Predefined position name (e.g., 'table_full', 'table_full_dark')")
    table_parser.add_argument("--x", type=float, help="X position (inches) - overrides position")
    table_parser.add_argument("--y", type=float, help="Y position (inches) - overrides position")
    table_parser.add_argument("--width", type=float, help="Width (inches) - overrides position")
    table_parser.add_argument("--height", type=float, help="Height (inches) - overrides position")
    table_parser.add_argument("--dark", action="store_true", help="Use dark styling (white text, orange header, dark positions)")

    # Add chart from Sheets
    chart_parser = subparsers.add_parser("add-chart", help="Add a chart from Sheets")
    chart_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    chart_parser.add_argument("--page-id", required=True, help="Slide ID")
    chart_parser.add_argument("--spreadsheet-id", required=True, help="Google Sheets ID")
    chart_parser.add_argument("--chart-id", type=int, required=True, help="Chart ID in sheet")
    chart_parser.add_argument("--x", type=float, default=0.5, help="X position (inches)")
    chart_parser.add_argument("--y", type=float, default=1.5, help="Y position (inches)")
    chart_parser.add_argument("--width", type=float, default=5, help="Width (inches)")
    chart_parser.add_argument("--height", type=float, default=3, help="Height (inches)")
    chart_parser.add_argument("--not-linked", action="store_true", help="Don't link to sheet")

    # Copy presentation
    copy_parser = subparsers.add_parser("copy", help="Copy entire presentation")
    copy_parser.add_argument("--pres-id", required=True, help="Source presentation ID")
    copy_parser.add_argument("--title", required=True, help="New presentation title")

    # Set placeholder text
    placeholder_parser = subparsers.add_parser("set-placeholder", help="Set placeholder text")
    placeholder_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    placeholder_parser.add_argument("--page-id", required=True, help="Slide ID")
    placeholder_parser.add_argument("--type", required=True, help="TITLE, SUBTITLE, BODY")
    placeholder_parser.add_argument("--text", required=True, help="Text content")

    # Create from Databricks template
    template_parser = subparsers.add_parser("create-from-template", help="Create presentation from Databricks template")
    template_parser.add_argument("--title", required=True, help="Presentation title")
    template_parser.add_argument("--template-id", default=DATABRICKS_TEMPLATE_ID, help="Template ID (default: Databricks Corporate)")
    template_parser.add_argument("--keep-samples", action="store_true", help="Keep sample slides from template")

    # Add slide from template layout
    tslide_parser = subparsers.add_parser("add-template-slide", help="Add slide using Databricks template layout")
    tslide_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    tslide_parser.add_argument("--layout", required=True, help="Layout name: shorthand (title, content_basic, closing_dark) or full display name (7 Content A - Basic)")
    tslide_parser.add_argument("--theme", default="light", choices=["light", "dark"], help="Theme for layout selection")
    tslide_parser.add_argument("--index", type=int, help="Insertion index")

    # List available layouts
    layouts_parser = subparsers.add_parser("list-layouts", help="List available layouts in presentation")
    layouts_parser.add_argument("--pres-id", required=True, help="Presentation ID")

    # List template layouts (without needing a presentation)
    tlayouts_parser = subparsers.add_parser("list-template-layouts", help="List available Databricks template layouts")
    tlayouts_parser.add_argument("--theme", default="light", choices=["light", "dark"], help="Theme (light or dark)")

    # Replace text across presentation
    replace_parser = subparsers.add_parser("replace-text", help="Replace text across presentation")
    replace_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    replace_parser.add_argument("--find", required=True, help="Text to find")
    replace_parser.add_argument("--replace", required=True, help="Replacement text")
    replace_parser.add_argument("--match-case", action="store_true", help="Match case")

    # List placeholders on a slide
    placeholders_parser = subparsers.add_parser("list-placeholders", help="List placeholders on a slide")
    placeholders_parser.add_argument("--pres-id", required=True, help="Presentation ID")
    placeholders_parser.add_argument("--page-id", required=True, help="Slide ID")

    # List predefined positions
    subparsers.add_parser("list-positions", help="List predefined positions for spatial layout")

    # Create from spec (JSON)
    spec_parser = subparsers.add_parser("create-from-spec", help="Create presentation from JSON spec")
    spec_parser.add_argument("--title", required=True, help="Presentation title")
    spec_parser.add_argument("--spec", required=True, help="JSON array of slide specs")
    spec_parser.add_argument("--theme", default="light", choices=["light", "dark"], help="Theme")

    args = parser.parse_args()

    try:
        if args.command == "create":
            pres_id = create_presentation(args.title)
            print(json.dumps({
                "presentationId": pres_id,
                "url": f"https://docs.google.com/presentation/d/{pres_id}/edit"
            }, indent=2))

        elif args.command == "info":
            pres = get_presentation(args.pres_id)
            if args.full:
                print(json.dumps(pres, indent=2))
            else:
                print(json.dumps({
                    "presentationId": pres["presentationId"],
                    "title": pres.get("title", ""),
                    "slideCount": len(pres.get("slides", [])),
                    "slideIds": [s["objectId"] for s in pres.get("slides", [])]
                }, indent=2))

        elif args.command == "list-slides":
            pres = get_presentation(args.pres_id)
            slides = []
            for i, slide in enumerate(pres.get("slides", [])):
                slides.append({
                    "index": i,
                    "objectId": slide["objectId"],
                    "elementCount": len(slide.get("pageElements", []))
                })
            print(json.dumps({"slides": slides}, indent=2))

        elif args.command == "add-slide":
            result = add_slide(args.pres_id, args.layout, args.index)
            print(json.dumps(result, indent=2))

        elif args.command == "duplicate-slide":
            result = duplicate_slide(args.pres_id, args.page_id)
            print(json.dumps(result, indent=2))

        elif args.command == "delete-slide":
            result = delete_slide(args.pres_id, args.page_id)
            print(json.dumps(result, indent=2))

        elif args.command == "set-background":
            color = json.loads(args.color) if args.color else None
            result = set_slide_background(args.pres_id, args.page_id, color, args.image_url)
            print(json.dumps(result, indent=2))

        elif args.command == "add-text-box":
            result = create_text_box(
                args.pres_id, args.page_id, args.text,
                args.x, args.y, args.width, args.height,
                args.font_size, args.bold
            )
            print(json.dumps(result, indent=2))

        elif args.command == "add-image":
            result = create_image(
                args.pres_id, args.page_id, args.url,
                args.x, args.y, args.width, args.height
            )
            print(json.dumps(result, indent=2))

        elif args.command == "add-table":
            data = json.loads(args.data)

            # Determine if using dark mode positioning
            use_dark = args.dark

            # Get position from name or use explicit coordinates
            if args.position:
                x, y, width, height = get_position(args.position)
            elif use_dark:
                # Default to dark table position
                x, y, width, height = get_position("table_full_dark")
            else:
                # Use explicit values or defaults
                x = args.x if args.x is not None else 0.5
                y = args.y if args.y is not None else BODY_TOP
                width = args.width if args.width is not None else 9.0
                height = args.height if args.height is not None else 3.0

            # Allow explicit coords to override position
            if args.x is not None:
                x = args.x
            if args.y is not None:
                y = args.y
            if args.width is not None:
                width = args.width
            if args.height is not None:
                height = args.height

            # Create table
            result = create_table(
                args.pres_id, args.page_id, args.rows, args.cols,
                x, y, width, height
            )
            if "error" not in result and "tableId" in result:
                # Fill table
                fill_result = fill_table(args.pres_id, result["tableId"], data)

                # Style based on dark mode
                if use_dark:
                    # Dark mode: orange header with white text, white body text
                    style_result = style_table_for_dark_background(
                        args.pres_id, result["tableId"], args.rows, args.cols
                    )
                else:
                    # Light mode: navy header with white text (default)
                    style_result = style_table_header(args.pres_id, result["tableId"], args.cols)

                result["filled"] = "error" not in fill_result
                result["styled"] = "error" not in style_result
                result["dark_mode"] = use_dark
                result["position"] = {"x": x, "y": y, "width": width, "height": height}
            print(json.dumps(result, indent=2))

        elif args.command == "add-chart":
            result = create_sheets_chart(
                args.pres_id, args.page_id, args.spreadsheet_id, args.chart_id,
                args.x, args.y, args.width, args.height,
                linked=not args.not_linked
            )
            print(json.dumps(result, indent=2))

        elif args.command == "copy":
            new_id = copy_presentation(args.pres_id, args.title)
            print(json.dumps({
                "presentationId": new_id,
                "url": f"https://docs.google.com/presentation/d/{new_id}/edit"
            }, indent=2))

        elif args.command == "set-placeholder":
            result = set_placeholder_text(args.pres_id, args.page_id, args.type, args.text)
            print(json.dumps(result, indent=2))

        elif args.command == "create-from-template":
            pres_id = create_from_template(
                args.title,
                args.template_id,
                delete_sample_slides=not args.keep_samples
            )
            print(json.dumps({
                "presentationId": pres_id,
                "url": f"https://docs.google.com/presentation/d/{pres_id}/edit",
                "template": "Databricks Corporate 2025"
            }, indent=2))

        elif args.command == "add-template-slide":
            # Use name-based lookup for dynamic layout resolution
            # This works correctly after copying a template (when IDs change)
            result = add_template_slide_by_name(
                args.pres_id,
                args.layout,
                theme=args.theme,
                insertion_index=args.index
            )
            print(json.dumps(result, indent=2))

        elif args.command == "list-layouts":
            layouts = list_layouts(args.pres_id)
            print(json.dumps({"layouts": layouts}, indent=2))

        elif args.command == "list-template-layouts":
            layouts = DATABRICKS_LAYOUTS_DARK if args.theme == "dark" else DATABRICKS_LAYOUTS_LIGHT
            print("Available Databricks template layouts ({} theme):".format(args.theme))
            print()
            categories = {
                "Title slides": ["title", "title_alt", "title_gradient", "title_orange"],
                "Content": ["content_basic", "content_basic_white", "content_2col", "content_2col_icon",
                           "content_3col", "content_3col_icon", "content_3col_cards",
                           "content_card_right", "content_card_left", "content_card_large"],
                "Section breaks": [f"section_break_{i}" for i in range(1, 9)],
                "Special": ["blank", "power_statement", "power_statement_2", "closing"],
                "Industry": ["industry_media", "industry_retail", "industry_healthcare",
                            "industry_manufacturing", "industry_financial", "industry_public",
                            "industry_consumer"]
            }
            for category, names in categories.items():
                available = [n for n in names if n in layouts]
                if available:
                    print(f"  {category}:")
                    for name in available:
                        print(f"    - {name}")
            print()

        elif args.command == "replace-text":
            result = replace_all_text(args.pres_id, args.find, args.replace, args.match_case)
            print(json.dumps(result, indent=2))

        elif args.command == "list-placeholders":
            placeholders = get_all_placeholders(args.pres_id, args.page_id)
            print(json.dumps({"placeholders": placeholders}, indent=2))

        elif args.command == "list-positions":
            print("Predefined positions for spatial layout:")
            print(f"\nSlide dimensions: {SLIDE_WIDTH}\" x {SLIDE_HEIGHT}\" (16:9)")
            print(f"Content area: {CONTENT_WIDTH}\" x {CONTENT_HEIGHT}\" (with {MARGIN_LEFT}\" margins)")
            print(f"Body area (light slides): starts at y={BODY_TOP}\", height={BODY_HEIGHT}\"")
            print(f"Body area (dark slides):  starts at y={DARK_BODY_TOP}\", height={DARK_BODY_HEIGHT:.2f}\"")
            print("\nAvailable positions (x, y, width, height in inches):")
            categories = {
                "Full area": ["full", "full_no_title"],
                "Horizontal thirds": ["left_third", "center_third", "right_third"],
                "Horizontal halves": ["left_half", "right_half"],
                "Vertical halves": ["top_half", "bottom_half"],
                "Quadrants": ["top_left", "top_right", "bottom_left", "bottom_right"],
                "Centered": ["center_large", "center_medium", "center_small"],
                "Tables (light)": ["table_full", "table_left", "table_right"],
                "Tables (dark)": ["table_full_dark", "table_left_dark", "table_right_dark"],
                "Charts (light)": ["chart_full", "chart_left", "chart_right"],
                "Charts (dark)": ["chart_full_dark", "chart_left_dark", "chart_right_dark"],
                "Images": ["image_left", "image_right", "image_center", "image_background"],
                "Text boxes": ["text_title_area", "text_subtitle", "text_footer", "text_caption"],
            }
            for category, names in categories.items():
                print(f"\n  {category}:")
                for name in names:
                    if name in POSITIONS:
                        x, y, w, h = POSITIONS[name]
                        print(f"    {name}: ({x:.2f}, {y:.2f}, {w:.2f}, {h:.2f})")

        elif args.command == "create-from-spec":
            slides = json.loads(args.spec)
            result = create_presentation_from_spec(args.title, slides, theme=args.theme)
            print(json.dumps(result, indent=2))

    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
