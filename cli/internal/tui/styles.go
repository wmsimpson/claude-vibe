// Package tui provides Bubble Tea components and styles for the vibe CLI.
package tui

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for the TUI.
type Theme struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Muted      lipgloss.Color
	Background lipgloss.Color
}

// DefaultTheme provides the standard vibe color scheme.
var DefaultTheme = Theme{
	Primary:    lipgloss.Color("#7C3AED"),
	Secondary:  lipgloss.Color("#10B981"),
	Success:    lipgloss.Color("#10B981"),
	Warning:    lipgloss.Color("#F59E0B"),
	Error:      lipgloss.Color("#EF4444"),
	Muted:      lipgloss.Color("#6B7280"),
	Background: lipgloss.Color("#1F2937"),
}

// HeaderStyle returns a styled header with the primary color.
func (t Theme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		MarginBottom(1)
}

// SelectedStyle returns styling for selected/highlighted items.
func (t Theme) SelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		Background(lipgloss.Color("#374151"))
}

// MutedStyle returns styling for de-emphasized text.
func (t Theme) MutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Muted)
}

// SuccessStyle returns styling for success messages.
func (t Theme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Success)
}

// WarningStyle returns styling for warning messages.
func (t Theme) WarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Warning)
}

// ErrorStyle returns styling for error messages.
func (t Theme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Error)
}

// SuccessIcon returns a styled checkmark icon.
func (t Theme) SuccessIcon() string {
	return lipgloss.NewStyle().
		Foreground(t.Success).
		Render("✓")
}

// WarningIcon returns a styled warning icon.
func (t Theme) WarningIcon() string {
	return lipgloss.NewStyle().
		Foreground(t.Warning).
		Render("⚠")
}

// ErrorIcon returns a styled error icon.
func (t Theme) ErrorIcon() string {
	return lipgloss.NewStyle().
		Foreground(t.Error).
		Render("✗")
}

// InfoIcon returns a styled info icon.
func (t Theme) InfoIcon() string {
	return lipgloss.NewStyle().
		Foreground(t.Primary).
		Render("●")
}

// BorderStyle returns a style for bordered containers.
func (t Theme) BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2)
}

// TitleStyle returns a style for titles.
func (t Theme) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		Padding(0, 1)
}

// SubtitleStyle returns a style for subtitles.
func (t Theme) SubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Muted).
		Italic(true)
}

// ActiveTabStyle returns styling for the active tab.
func (t Theme) ActiveTabStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(t.Primary).
		Padding(0, 2)
}

// InactiveTabStyle returns styling for inactive tabs.
func (t Theme) InactiveTabStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Muted).
		Padding(0, 2)
}

// StatusBarStyle returns styling for status bars.
func (t Theme) StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#374151")).
		Foreground(lipgloss.Color("#D1D5DB")).
		Padding(0, 1)
}

// HelpStyle returns styling for help text.
func (t Theme) HelpStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Muted).
		Italic(true)
}
