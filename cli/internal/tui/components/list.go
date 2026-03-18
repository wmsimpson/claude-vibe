package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
)

// List is a selectable list with toggleable items.
type List struct {
	items    []ListItem
	cursor   int
	theme    tui.Theme
	height   int
	offset   int
	focused  bool
	showHelp bool
}

// ListItem represents an item in the list.
type ListItem struct {
	ID          string
	Title       string
	Description string
	Enabled     bool
	Disabled    bool // Cannot be toggled
	Dimmed      bool // Visually de-emphasized
}

// ListOption configures a List component.
type ListOption func(*List)

// WithListTheme sets a custom theme for the list.
func WithListTheme(theme tui.Theme) ListOption {
	return func(l *List) {
		l.theme = theme
	}
}

// WithListHeight sets the visible height of the list.
func WithListHeight(height int) ListOption {
	return func(l *List) {
		l.height = height
	}
}

// WithListHelp shows/hides the help text.
func WithListHelp(show bool) ListOption {
	return func(l *List) {
		l.showHelp = show
	}
}

// NewList creates a new List component.
func NewList(items []ListItem, opts ...ListOption) List {
	l := List{
		items:    items,
		cursor:   0,
		theme:    tui.DefaultTheme,
		height:   10,
		focused:  true,
		showHelp: true,
	}

	for _, opt := range opts {
		opt(&l)
	}

	return l
}

// Init implements tea.Model.
func (l List) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (l List) Update(msg tea.Msg) (List, tea.Cmd) {
	if !l.focused {
		return l, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if l.cursor > 0 {
				l.cursor--
				// Scroll up if needed
				if l.cursor < l.offset {
					l.offset = l.cursor
				}
			}

		case "down", "j":
			if l.cursor < len(l.items)-1 {
				l.cursor++
				// Scroll down if needed
				if l.cursor >= l.offset+l.height {
					l.offset = l.cursor - l.height + 1
				}
			}

		case " ":
			// Toggle selected item
			if l.cursor >= 0 && l.cursor < len(l.items) {
				item := &l.items[l.cursor]
				if !item.Disabled {
					item.Enabled = !item.Enabled
					return l, func() tea.Msg {
						return ListItemToggledMsg{
							Index:   l.cursor,
							Item:    *item,
							Enabled: item.Enabled,
						}
					}
				}
			}

		case "enter":
			if l.cursor >= 0 && l.cursor < len(l.items) {
				return l, func() tea.Msg {
					return ListItemSelectedMsg{
						Index: l.cursor,
						Item:  l.items[l.cursor],
					}
				}
			}

		case "a":
			// Enable all
			l.enableAll()
			return l, func() tea.Msg {
				return ListAllToggledMsg{Enabled: true}
			}

		case "n":
			// Disable all
			l.disableAll()
			return l, func() tea.Msg {
				return ListAllToggledMsg{Enabled: false}
			}

		case "home", "g":
			l.cursor = 0
			l.offset = 0

		case "end", "G":
			l.cursor = len(l.items) - 1
			if l.cursor >= l.height {
				l.offset = l.cursor - l.height + 1
			}
		}
	}

	return l, nil
}

// View implements tea.Model.
func (l List) View() string {
	var lines []string

	// Calculate visible range
	start := l.offset
	end := start + l.height
	if end > len(l.items) {
		end = len(l.items)
	}

	for i := start; i < end; i++ {
		item := l.items[i]
		line := l.renderItem(i, item)
		lines = append(lines, line)
	}

	// Pad to height
	for len(lines) < l.height {
		lines = append(lines, "")
	}

	result := strings.Join(lines, "\n")

	// Add scrollbar indicator if needed
	if len(l.items) > l.height {
		scrollInfo := lipgloss.NewStyle().
			Foreground(l.theme.Muted).
			Render(fmt.Sprintf("  [%d-%d of %d]", start+1, end, len(l.items)))
		result += "\n" + scrollInfo
	}

	// Add help text
	if l.showHelp {
		helpStyle := l.theme.HelpStyle()
		help := helpStyle.Render("↑/↓: navigate • space: toggle • a: all • n: none • enter: select")
		result += "\n\n" + help
	}

	return result
}

// renderItem renders a single list item.
func (l List) renderItem(index int, item ListItem) string {
	cursor := "  "
	if index == l.cursor && l.focused {
		cursor = "▸ "
	}

	// Checkbox shows enabled state
	// Disabled items (can't toggle) show the same checkbox but are visually muted
	checkbox := "[ ]"
	if item.Enabled {
		checkbox = "[" + l.theme.SuccessIcon() + "]"
	}

	// Apply styling based on state
	titleStyle := lipgloss.NewStyle()
	descStyle := lipgloss.NewStyle().Foreground(l.theme.Muted)

	if item.Dimmed || item.Disabled {
		titleStyle = titleStyle.Foreground(l.theme.Muted)
		descStyle = descStyle.Foreground(l.theme.Muted).Faint(true)
	} else if index == l.cursor && l.focused {
		titleStyle = titleStyle.Foreground(l.theme.Primary).Bold(true)
	}

	title := titleStyle.Render(item.Title)

	line := cursor + checkbox + " " + title

	if item.Description != "" {
		desc := descStyle.Render(" - " + item.Description)
		line += desc
	}

	return line
}

// enableAll enables all non-disabled items.
func (l *List) enableAll() {
	for i := range l.items {
		if !l.items[i].Disabled {
			l.items[i].Enabled = true
		}
	}
}

// disableAll disables all non-disabled items.
func (l *List) disableAll() {
	for i := range l.items {
		if !l.items[i].Disabled {
			l.items[i].Enabled = false
		}
	}
}

// SelectedItem returns the currently highlighted item.
func (l List) SelectedItem() *ListItem {
	if l.cursor >= 0 && l.cursor < len(l.items) {
		return &l.items[l.cursor]
	}
	return nil
}

// EnabledItems returns all enabled items.
func (l List) EnabledItems() []ListItem {
	var enabled []ListItem
	for _, item := range l.items {
		if item.Enabled {
			enabled = append(enabled, item)
		}
	}
	return enabled
}

// Cursor returns the current cursor position.
func (l List) Cursor() int {
	return l.cursor
}

// SetCursor sets the cursor position.
func (l *List) SetCursor(index int) {
	if index >= 0 && index < len(l.items) {
		l.cursor = index
		// Adjust offset to keep cursor visible
		if l.cursor < l.offset {
			l.offset = l.cursor
		} else if l.cursor >= l.offset+l.height {
			l.offset = l.cursor - l.height + 1
		}
	}
}

// Items returns all items.
func (l List) Items() []ListItem {
	return l.items
}

// SetItems updates the list items.
func (l *List) SetItems(items []ListItem) {
	l.items = items
	if l.cursor >= len(items) {
		l.cursor = len(items) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
	l.offset = 0
}

// SetEnabled sets the enabled state of an item by index.
func (l *List) SetEnabled(index int, enabled bool) {
	if index >= 0 && index < len(l.items) && !l.items[index].Disabled {
		l.items[index].Enabled = enabled
	}
}

// Focus focuses the list for keyboard input.
func (l *List) Focus() {
	l.focused = true
}

// Blur removes focus from the list.
func (l *List) Blur() {
	l.focused = false
}

// Focused returns true if the list is focused.
func (l List) Focused() bool {
	return l.focused
}

// ListItemToggledMsg is sent when an item is toggled.
type ListItemToggledMsg struct {
	Index   int
	Item    ListItem
	Enabled bool
}

// ListItemSelectedMsg is sent when an item is selected (Enter pressed).
type ListItemSelectedMsg struct {
	Index int
	Item  ListItem
}

// ListAllToggledMsg is sent when all items are toggled.
type ListAllToggledMsg struct {
	Enabled bool
}
