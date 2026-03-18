package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
)

// Confirm is a Y/N confirmation dialog.
type Confirm struct {
	prompt       string
	defaultYes   bool
	confirmed    bool
	answered     bool
	theme        tui.Theme
	showBorder   bool
	width        int
	affirmative  string
	negative     string
	affirmKey    string
	negKey       string
}

// ConfirmOption configures a Confirm dialog.
type ConfirmOption func(*Confirm)

// WithConfirmTheme sets a custom theme.
func WithConfirmTheme(theme tui.Theme) ConfirmOption {
	return func(c *Confirm) {
		c.theme = theme
	}
}

// WithConfirmDefault sets the default response.
func WithConfirmDefault(yes bool) ConfirmOption {
	return func(c *Confirm) {
		c.defaultYes = yes
	}
}

// WithConfirmBorder shows/hides the border.
func WithConfirmBorder(show bool) ConfirmOption {
	return func(c *Confirm) {
		c.showBorder = show
	}
}

// WithConfirmWidth sets the width of the dialog.
func WithConfirmWidth(width int) ConfirmOption {
	return func(c *Confirm) {
		c.width = width
	}
}

// WithConfirmLabels sets custom labels for yes/no.
func WithConfirmLabels(affirmative, negative string) ConfirmOption {
	return func(c *Confirm) {
		c.affirmative = affirmative
		c.negative = negative
	}
}

// WithConfirmKeys sets custom keys for yes/no.
func WithConfirmKeys(affirm, neg string) ConfirmOption {
	return func(c *Confirm) {
		c.affirmKey = affirm
		c.negKey = neg
	}
}

// NewConfirm creates a new Confirm dialog.
func NewConfirm(prompt string, opts ...ConfirmOption) Confirm {
	c := Confirm{
		prompt:      prompt,
		defaultYes:  false,
		theme:       tui.DefaultTheme,
		showBorder:  false,
		affirmative: "Yes",
		negative:    "No",
		affirmKey:   "y",
		negKey:      "n",
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

// Init implements tea.Model.
func (c Confirm) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (c Confirm) Update(msg tea.Msg) (Confirm, tea.Cmd) {
	if c.answered {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case c.affirmKey, "Y":
			c.confirmed = true
			c.answered = true
			return c, func() tea.Msg {
				return ConfirmResultMsg{Confirmed: true}
			}

		case c.negKey, "N":
			c.confirmed = false
			c.answered = true
			return c, func() tea.Msg {
				return ConfirmResultMsg{Confirmed: false}
			}

		case "enter":
			c.confirmed = c.defaultYes
			c.answered = true
			return c, func() tea.Msg {
				return ConfirmResultMsg{Confirmed: c.confirmed}
			}

		case "esc":
			c.confirmed = false
			c.answered = true
			return c, func() tea.Msg {
				return ConfirmCancelledMsg{}
			}
		}
	}

	return c, nil
}

// View implements tea.Model.
func (c Confirm) View() string {
	promptStyle := lipgloss.NewStyle().Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(c.theme.Muted)
	activeStyle := lipgloss.NewStyle().Foreground(c.theme.Primary).Bold(true)

	var yesLabel, noLabel string

	if c.defaultYes {
		yesLabel = activeStyle.Render("[" + c.affirmative + "]")
		noLabel = mutedStyle.Render(c.negative)
	} else {
		yesLabel = mutedStyle.Render(c.affirmative)
		noLabel = activeStyle.Render("[" + c.negative + "]")
	}

	hint := mutedStyle.Render(" (" + c.affirmKey + "/" + c.negKey + ")")

	content := promptStyle.Render(c.prompt) + " " + yesLabel + "/" + noLabel + hint

	if c.showBorder {
		borderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.theme.Primary).
			Padding(1, 2)

		if c.width > 0 {
			borderStyle = borderStyle.Width(c.width)
		}

		return borderStyle.Render(content)
	}

	return content
}

// Confirmed returns true if the user confirmed.
func (c Confirm) Confirmed() bool {
	return c.confirmed
}

// Answered returns true if the user has answered.
func (c Confirm) Answered() bool {
	return c.answered
}

// Reset resets the dialog for reuse.
func (c *Confirm) Reset() {
	c.confirmed = false
	c.answered = false
}

// SetPrompt updates the prompt text.
func (c *Confirm) SetPrompt(prompt string) {
	c.prompt = prompt
}

// ConfirmResultMsg is sent when the user confirms or denies.
type ConfirmResultMsg struct {
	Confirmed bool
}

// ConfirmCancelledMsg is sent when the user cancels (Esc).
type ConfirmCancelledMsg struct{}
