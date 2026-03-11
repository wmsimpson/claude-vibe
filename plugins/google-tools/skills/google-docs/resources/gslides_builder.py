#!/usr/bin/env python3
"""
Google Slides Builder - Build presentations with proper element management

This script helps create well-formatted Google Slides presentations by:
1. Creating presentations and slides
2. Adding shapes, text, images, tables, and charts
3. Duplicating slides and managing layouts
4. Copying slides between presentations

Usage:
    # Create a new presentation
    python3 gslides_builder.py create --title "My Presentation"

    # Add a slide with layout
    python3 gslides_builder.py add-slide --pres-id "PRES_ID" --layout "TITLE_AND_BODY"

    # Add text to a placeholder
    python3 gslides_builder.py add-text --pres-id "PRES_ID" --page-id "PAGE_ID" --text "Hello"
"""

import argparse
import json
import os
import sys
import uuid
from typing import Dict, List, Optional

# google_api_utils lives in the google-auth skill. Scripts outside that directory
# must resolve the path explicitly — this is the established pattern (see evals/tests/test_google_api_utils.py).
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "../../google-auth/resources"))
from google_api_utils import api_call_with_retry

# EMU (English Metric Units) conversion
# 1 inch = 914400 EMU, 1 pt = 12700 EMU
EMU_PER_INCH = 914400
EMU_PER_PT = 12700

# Standard slide dimensions (10" x 5.625" for 16:9)
SLIDE_WIDTH_EMU = 9144000   # 10 inches
SLIDE_HEIGHT_EMU = 5143500  # 5.625 inches

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


def inches_to_emu(inches: float) -> int:
    """Convert inches to EMU."""
    return int(inches * EMU_PER_INCH)


def pt_to_emu(pt: float) -> int:
    """Convert points to EMU."""
    return int(pt * EMU_PER_PT)


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
    response = api_call(
        "POST",
        f"https://www.googleapis.com/drive/v3/files/{template_id}/copy",
        {"name": title}
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

    return insert_text(pres_id, shape_id, text)


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
        text_box_id: Optional custom ID

    Returns:
        API response with textBoxId
    """
    text_box_id = text_box_id or generate_id()

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
                "style": {
                    "fontSize": {"magnitude": font_size, "unit": "PT"},
                    "bold": bold
                },
                "fields": "fontSize,bold"
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
    bg_color: Dict = {"red": 0.2, "green": 0.4, "blue": 0.7}
) -> Dict:
    """Style the header row of a table with background color."""
    requests = []

    for col_idx in range(cols):
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
    response = api_call(
        "POST",
        f"https://www.googleapis.com/drive/v3/files/{pres_id}/copy",
        {"name": new_title}
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
    table_parser.add_argument("--x", type=float, default=0.5, help="X position (inches)")
    table_parser.add_argument("--y", type=float, default=1.5, help="Y position (inches)")
    table_parser.add_argument("--width", type=float, default=9, help="Width (inches)")
    table_parser.add_argument("--height", type=float, default=3, help="Height (inches)")

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
    tslide_parser.add_argument("--index", type=int, help="Insertion index")

    # List available layouts
    layouts_parser = subparsers.add_parser("list-layouts", help="List available layouts in presentation")
    layouts_parser.add_argument("--pres-id", required=True, help="Presentation ID")

    # List template layouts (without needing a presentation)
    tlayouts_parser = subparsers.add_parser("list-template-layouts", help="List available Databricks template layouts")
    tlayouts_parser.add_argument("--theme", default="light", choices=["light", "dark"], help="Theme (light or dark)")

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
            # Create table
            result = create_table(
                args.pres_id, args.page_id, args.rows, args.cols,
                args.x, args.y, args.width, args.height
            )
            if "error" not in result and "tableId" in result:
                # Fill table
                fill_result = fill_table(args.pres_id, result["tableId"], data)
                # Style header
                style_result = style_table_header(args.pres_id, result["tableId"], args.cols)
                result["filled"] = "error" not in fill_result
                result["styled"] = "error" not in style_result
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
            result = add_template_slide_by_name(
                args.pres_id,
                args.layout,
                args.index
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

    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
