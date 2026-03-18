// Package components provides reusable Bubble Tea UI components.
package components

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
)

// Table wraps bubbles/table with custom styling and selection handling.
type Table struct {
	table  table.Model
	theme  tui.Theme
	width  int
	height int
}

// TableColumn defines a column in the table.
type TableColumn struct {
	Title string
	Width int
}

// TableRow represents a row of data.
type TableRow []string

// TableOption configures a Table.
type TableOption func(*Table)

// WithTableTheme sets a custom theme for the table.
func WithTableTheme(theme tui.Theme) TableOption {
	return func(t *Table) {
		t.theme = theme
		t.applyStyles()
	}
}

// WithTableHeight sets the visible height of the table.
func WithTableHeight(height int) TableOption {
	return func(t *Table) {
		t.height = height
		t.table.SetHeight(height)
	}
}

// WithTableWidth sets the width of the table.
func WithTableWidth(width int) TableOption {
	return func(t *Table) {
		t.width = width
		t.table.SetWidth(width)
	}
}

// NewTable creates a new Table with the given columns and rows.
func NewTable(columns []TableColumn, rows []TableRow, opts ...TableOption) Table {
	// Convert to bubbles table columns
	cols := make([]table.Column, len(columns))
	for i, c := range columns {
		cols[i] = table.Column{
			Title: c.Title,
			Width: c.Width,
		}
	}

	// Convert to bubbles table rows
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row(r)
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(tableRows),
		table.WithFocused(true),
	)

	tbl := Table{
		table:  t,
		theme:  tui.DefaultTheme,
		height: 10,
	}

	// Apply options
	for _, opt := range opts {
		opt(&tbl)
	}

	tbl.applyStyles()

	return tbl
}

// applyStyles applies the theme to the table.
func (t *Table) applyStyles() {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.theme.Muted).
		BorderBottom(true).
		Bold(true).
		Foreground(t.theme.Primary)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(t.theme.Primary).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("#D1D5DB"))

	t.table.SetStyles(s)
}

// Init implements tea.Model.
func (t Table) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (t Table) Update(msg tea.Msg) (Table, tea.Cmd) {
	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	return t, cmd
}

// View implements tea.Model.
func (t Table) View() string {
	return t.table.View()
}

// SelectedRow returns the currently selected row, or nil if no selection.
func (t Table) SelectedRow() TableRow {
	row := t.table.SelectedRow()
	if row == nil {
		return nil
	}
	return TableRow(row)
}

// Cursor returns the current cursor position.
func (t Table) Cursor() int {
	return t.table.Cursor()
}

// SetRows updates the table rows.
func (t *Table) SetRows(rows []TableRow) {
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row(r)
	}
	t.table.SetRows(tableRows)
}

// SetColumns updates the table columns.
func (t *Table) SetColumns(columns []TableColumn) {
	cols := make([]table.Column, len(columns))
	for i, c := range columns {
		cols[i] = table.Column{
			Title: c.Title,
			Width: c.Width,
		}
	}
	t.table.SetColumns(cols)
}

// Focus focuses the table for keyboard input.
func (t *Table) Focus() {
	t.table.Focus()
}

// Blur removes focus from the table.
func (t *Table) Blur() {
	t.table.Blur()
}

// Focused returns true if the table is focused.
func (t Table) Focused() bool {
	return t.table.Focused()
}

// TableSelectMsg is sent when a row is selected (Enter pressed).
type TableSelectMsg struct {
	Row   TableRow
	Index int
}
