# Element Interaction Guide

Detailed patterns for interacting with dashboard UI elements.

## Finding Elements

Always start with a snapshot to get element UIDs:

```bash
mcp__chrome-devtools__take_snapshot
```

**Snapshot output format:**
```
uid=1_5 button "Apply Filters"
uid=2_7 textbox "Search..."
uid=3_5 combobox "Category Filter"
uid=4_2 table
  uid=4_3 row "Name | Value | Status"
  uid=4_4 row "Item A | 123 | Active"
```

Look for elements by:
- **Role**: button, link, textbox, combobox, table, heading, tab
- **Text content**: "Apply", "Filter", "Download"
- **UID**: `uid=1_5` for interaction

---

## Clicking Buttons and Links

```bash
# 1. Get UID from snapshot
mcp__chrome-devtools__take_snapshot
# Find: uid=1_12 button "Apply Filters"

# 2. Click using UID
mcp__chrome-devtools__click '{"uid": "1_12"}'

# For double-click
mcp__chrome-devtools__click '{"uid": "1_12", "dblClick": true}'
```

---

## Filling Text Fields

```bash
# 1. Get UID from snapshot
# Find: uid=2_7 textbox "Search..."

# 2. Fill using UID
mcp__chrome-devtools__fill '{"uid": "2_7", "value": "search term"}'
```

---

## Working with Dropdowns

Standard dropdown workflow:

```bash
# 1. Take snapshot to find dropdown
mcp__chrome-devtools__take_snapshot
# Find: uid=3_5 combobox "Category Filter"

# 2. Click to open dropdown
mcp__chrome-devtools__click '{"uid": "3_5"}'

# 3. Take new snapshot to see options
mcp__chrome-devtools__take_snapshot
# Find: uid=3_8 option "Option A"
#       uid=3_9 option "Option B"

# 4. Click desired option
mcp__chrome-devtools__click '{"uid": "3_8"}'
```

---

## React Controlled Inputs (Autocomplete Filters)

Databricks dashboards use React controlled inputs that don't respond to normal `fill()` or JavaScript `input.value =`. Use the native setter technique:

```bash
# 1. Click the combobox to focus it
mcp__chrome-devtools__click '{"uid": "COMBOBOX_UID"}'

# 2. Use evaluate_script with native setter
mcp__chrome-devtools__evaluate_script '{"function": "() => { const input = document.activeElement; const nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, \"value\").set; nativeInputValueSetter.call(input, \"Samsara, Inc\"); const event = new Event(\"input\", { bubbles: true }); input.dispatchEvent(event); return { newValue: input.value }; }"}'

# 3. Take snapshot to see filtered options
mcp__chrome-devtools__take_snapshot

# 4. Click on the desired option from the dropdown
mcp__chrome-devtools__click '{"uid": "OPTION_UID"}'
```

**Why this works:** React intercepts the native value setter. By using `Object.getOwnPropertyDescriptor` to get the original HTMLInputElement setter and calling it directly, we bypass React's interception.

---

## Navigating Dashboard Tabs

AI/BI dashboards often have multiple tabs (pages).

**Identifying tabs in snapshots:**
```
uid=X_1 tab "Overview" selectable
uid=X_2 tab "Customer View" selectable selected
uid=X_3 tab "Admin View" selectable
```

The `selected` attribute indicates the current tab.

**Switching tabs by clicking:**
```bash
mcp__chrome-devtools__click '{"uid": "X_3"}'
```

**Switching tabs via URL (faster):**

Tabs have their own page ID in the URL:
- Overview: `/pages/abc123`
- Customer View: `/pages/def456`
- Admin View: `/pages/ghi789`

```bash
mcp__chrome-devtools__navigate_page '{"type": "url", "url": "https://adb-XXX.azuredatabricks.net/dashboardsv3/UUID/published/pages/ghi789?o=ORGID"}'
```

---

## Hovering to Reveal Hidden Elements

Some elements (like ellipsis menus) only appear on hover:

```bash
# 1. Take snapshot to find the parent element (e.g., chart heading)
mcp__chrome-devtools__take_snapshot
# Find: uid=X_Y heading "$DBUs by Product Line" level="4"

# 2. Hover over the element
mcp__chrome-devtools__hover '{"uid": "X_Y"}'
# The hover action returns a new snapshot automatically

# 3. Look for newly revealed elements in the response
# Find: uid=X_Z button expandable haspopup="menu"

# 4. Click the revealed element
mcp__chrome-devtools__click '{"uid": "X_Z"}'
```

---

## Keyboard Navigation

For menus and complex UI interactions, keyboard navigation is often more reliable:

```bash
# Navigate menu items
mcp__chrome-devtools__press_key '{"key": "ArrowDown"}'
mcp__chrome-devtools__press_key '{"key": "ArrowUp"}'

# Expand/collapse submenus
mcp__chrome-devtools__press_key '{"key": "ArrowRight"}'
mcp__chrome-devtools__press_key '{"key": "ArrowLeft"}'

# Select current item
mcp__chrome-devtools__press_key '{"key": "Enter"}'

# Close menu/dialog
mcp__chrome-devtools__press_key '{"key": "Escape"}'

# Keyboard shortcuts
mcp__chrome-devtools__press_key '{"key": "Control+A"}'  # Select all
mcp__chrome-devtools__press_key '{"key": "Control+C"}'  # Copy
```

**When to use keyboard vs click:**
- **Keyboard**: Nested menus (they collapse easily on hover), form navigation, shortcuts
- **Click**: Buttons, links, simple dropdowns, tabs

---

## Scrolling

### Window Scrolling (rarely works for dashboards)

```bash
mcp__chrome-devtools__evaluate_script '{"function": "() => { window.scrollTo(0, 500); return { scrollY: window.scrollY }; }"}'
```

### Container Scrolling (use this for dashboards)

Dashboard content is typically in a scrollable container:

```bash
mcp__chrome-devtools__evaluate_script '{"function": "() => { const container = document.querySelector(\"[class*=dbsql-legacy-mfe-page]\"); if (container) { container.scrollBy(0, 600); return { scrolled: true, newScrollTop: container.scrollTop }; } return { scrolled: false }; }"}'
```

**Signs you need container scrolling:**
- `window.scrollY` stays at 0 after scroll attempts
- Charts/tables are cut off in screenshots
- Snapshot shows elements that aren't visible in screenshot

---

## Handling Dialogs

If a browser dialog appears (alert, confirm, prompt):

```bash
# Accept dialog
mcp__chrome-devtools__handle_dialog '{"action": "accept"}'

# Dismiss dialog
mcp__chrome-devtools__handle_dialog '{"action": "dismiss"}'

# Accept with text input (for prompts)
mcp__chrome-devtools__handle_dialog '{"action": "accept", "promptText": "user input"}'
```

---

## Waiting for Elements

After interactions, wait for the page to update:

```bash
# Wait for specific text to appear
mcp__chrome-devtools__wait_for '{"text": "Loading complete", "timeout": 15000}'

# Wait for content after filter change
mcp__chrome-devtools__wait_for '{"text": "Showing results", "timeout": 10000}'
```

If no specific text to wait for, use a brief delay via evaluate_script:
```bash
mcp__chrome-devtools__evaluate_script '{"function": "() => new Promise(r => setTimeout(r, 2000))"}'
```
