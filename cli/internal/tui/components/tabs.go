package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
)

// Tabs provides tab navigation with Tab/Shift+Tab keys.
type Tabs struct {
	tabs       []string
	activeTab  int
	theme      tui.Theme
	width      int
	showBorder bool
}

// TabsOption configures a Tabs component.
type TabsOption func(*Tabs)

// WithTabsTheme sets a custom theme for tabs.
func WithTabsTheme(theme tui.Theme) TabsOption {
	return func(t *Tabs) {
		t.theme = theme
	}
}

// WithTabsWidth sets the width for the tabs bar.
func WithTabsWidth(width int) TabsOption {
	return func(t *Tabs) {
		t.width = width
	}
}

// WithTabsBorder enables a border below the tabs.
func WithTabsBorder(show bool) TabsOption {
	return func(t *Tabs) {
		t.showBorder = show
	}
}

// NewTabs creates a new Tabs component.
func NewTabs(tabs []string, opts ...TabsOption) Tabs {
	t := Tabs{
		tabs:       tabs,
		activeTab:  0,
		theme:      tui.DefaultTheme,
		showBorder: true,
	}

	for _, opt := range opts {
		opt(&t)
	}

	return t
}

// Init implements tea.Model.
func (t Tabs) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (t Tabs) Update(msg tea.Msg) (Tabs, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			t.activeTab = (t.activeTab + 1) % len(t.tabs)
			return t, func() tea.Msg {
				return TabChangedMsg{Index: t.activeTab, Tab: t.tabs[t.activeTab]}
			}
		case "shift+tab":
			t.activeTab = (t.activeTab - 1 + len(t.tabs)) % len(t.tabs)
			return t, func() tea.Msg {
				return TabChangedMsg{Index: t.activeTab, Tab: t.tabs[t.activeTab]}
			}
		}
	}
	return t, nil
}

// View implements tea.Model.
func (t Tabs) View() string {
	var tabs []string

	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(t.theme.Primary).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(t.theme.Primary).
		Padding(0, 2)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(t.theme.Muted).
		Padding(0, 2)

	for i, tab := range t.tabs {
		if i == t.activeTab {
			tabs = append(tabs, activeStyle.Render(tab))
		} else {
			tabs = append(tabs, inactiveStyle.Render(tab))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	if t.showBorder {
		borderStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(t.theme.Muted)

		if t.width > 0 {
			borderStyle = borderStyle.Width(t.width)
		}

		// Add a separator line below tabs
		separator := strings.Repeat("─", max(lipgloss.Width(row), t.width))
		separatorStyle := lipgloss.NewStyle().Foreground(t.theme.Muted)

		return lipgloss.JoinVertical(lipgloss.Left, row, separatorStyle.Render(separator))
	}

	return row
}

// ActiveTab returns the index of the active tab.
func (t Tabs) ActiveTab() int {
	return t.activeTab
}

// ActiveTabName returns the name of the active tab.
func (t Tabs) ActiveTabName() string {
	if t.activeTab >= 0 && t.activeTab < len(t.tabs) {
		return t.tabs[t.activeTab]
	}
	return ""
}

// SetActiveTab sets the active tab by index.
func (t *Tabs) SetActiveTab(index int) {
	if index >= 0 && index < len(t.tabs) {
		t.activeTab = index
	}
}

// SetTabs updates the tab list.
func (t *Tabs) SetTabs(tabs []string) {
	t.tabs = tabs
	if t.activeTab >= len(tabs) {
		t.activeTab = len(tabs) - 1
	}
	if t.activeTab < 0 {
		t.activeTab = 0
	}
}

// TabCount returns the number of tabs.
func (t Tabs) TabCount() int {
	return len(t.tabs)
}

// TabChangedMsg is sent when the active tab changes.
type TabChangedMsg struct {
	Index int
	Tab   string
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
