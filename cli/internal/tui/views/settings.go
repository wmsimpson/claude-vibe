package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/config"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
)

// SettingsView displays and edits vibe settings.
type SettingsView struct {
	config      *config.Config
	isaacConfig *config.IsaacConfig
	theme       tui.Theme
	cursor      int
	width       int
	height      int
	options     []settingOption
	dirty       bool
	saving      bool
	saveError   error
}

type settingOption struct {
	label       string
	description string
	valueType   string // "select", "toggle"
	choices     []string
	value       int // Index for select, 0/1 for toggle
}

// SettingsOption configures a SettingsView.
type SettingsOption func(*SettingsView)

// WithSettingsTheme sets a custom theme.
func WithSettingsTheme(theme tui.Theme) SettingsOption {
	return func(s *SettingsView) {
		s.theme = theme
	}
}

// NewSettingsView creates a new SettingsView.
func NewSettingsView(opts ...SettingsOption) *SettingsView {
	s := &SettingsView{
		theme:  tui.DefaultTheme,
		cursor: 0,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Init implements tea.Model.
func (s *SettingsView) Init() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			return settingsLoadedMsg{err: err}
		}
		// Also load isaac config (optional - may not exist yet)
		isaacCfg, _ := config.LoadIsaacConfig()
		return settingsLoadedMsg{config: cfg, isaacConfig: isaacCfg}
	}
}

// Update handles messages and returns the updated view.
func (s *SettingsView) Update(msg tea.Msg) (*SettingsView, tea.Cmd) {
	switch msg := msg.(type) {
	case settingsLoadedMsg:
		if msg.err != nil {
			s.saveError = msg.err
			return s, nil
		}
		s.config = msg.config
		s.isaacConfig = msg.isaacConfig
		s.buildOptions()
		return s, nil

	case settingsSavedMsg:
		s.saving = false
		s.dirty = false
		if msg.err != nil {
			s.saveError = msg.err
			return s, func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Error saving: %v", msg.err)}
			}
		}
		return s, func() tea.Msg {
			return StatusMsg{Message: "Settings saved successfully"}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}

		case "down", "j":
			if s.cursor < len(s.options)-1 {
				s.cursor++
			}

		case "left", "h":
			s.decrementOption()

		case "right", "l":
			s.incrementOption()

		case " ", "enter":
			s.toggleOption()

		case "s", "ctrl+s":
			if !s.saving {
				s.saving = true
				return s, s.save()
			}
		}
	}

	return s, nil
}

// View renders the settings view.
func (s *SettingsView) View() string {
	if s.config == nil {
		return "Loading settings..."
	}

	var lines []string

	// Section header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(s.theme.Primary).
		MarginBottom(1)
	lines = append(lines, headerStyle.Render("General Settings"))

	// Settings list
	for i, opt := range s.options {
		line := s.renderOption(i, opt)
		lines = append(lines, line)
	}

	// Status indicator
	if s.saving {
		lines = append(lines, "")
		savingStyle := lipgloss.NewStyle().Foreground(s.theme.Primary)
		lines = append(lines, savingStyle.Render("Saving..."))
	} else if s.dirty {
		lines = append(lines, "")
		dirtyStyle := lipgloss.NewStyle().Foreground(s.theme.Warning)
		lines = append(lines, dirtyStyle.Render("* Unsaved changes - press 's' to save"))
	}

	// Help text
	lines = append(lines, "")
	helpStyle := s.theme.HelpStyle()
	lines = append(lines, helpStyle.Render("j/k: navigate | left/right: change value | space: toggle | s: save"))

	return strings.Join(lines, "\n")
}

// SetSize sets the view dimensions.
func (s *SettingsView) SetSize(width, height int) {
	s.width = width
	s.height = height
}

func (s *SettingsView) buildOptions() {
	if s.config == nil {
		return
	}

	// Theme selector
	themeIndex := 0
	themes := []string{"default", "dark", "light"}
	for i, t := range themes {
		if t == s.config.Settings.Theme {
			themeIndex = i
			break
		}
	}

	// Default agent selector
	agentIndex := 0
	agents := []string{"claude"}
	for i, a := range agents {
		if a == s.config.Settings.DefaultAgent {
			agentIndex = i
			break
		}
	}

	// Auto-update toggle
	autoUpdateValue := 0
	if s.config.Settings.AutoUpdateCheck {
		autoUpdateValue = 1
	}

	s.options = []settingOption{
		{
			label:       "Theme",
			description: "Color theme for the TUI",
			valueType:   "select",
			choices:     themes,
			value:       themeIndex,
		},
		{
			label:       "Default Agent",
			description: "Agent to use when running 'vibe agent'",
			valueType:   "select",
			choices:     agents,
			value:       agentIndex,
		},
		{
			label:       "Auto-Update Check",
			description: "Check for updates on startup",
			valueType:   "toggle",
			choices:     []string{"Off", "On"},
			value:       autoUpdateValue,
		},
	}
}

func (s *SettingsView) renderOption(index int, opt settingOption) string {
	cursor := "  "
	if index == s.cursor {
		cursor = "> "
	}

	labelStyle := lipgloss.NewStyle()
	valueStyle := lipgloss.NewStyle().Foreground(s.theme.Secondary)
	descStyle := lipgloss.NewStyle().Foreground(s.theme.Muted)

	if index == s.cursor {
		labelStyle = labelStyle.Bold(true).Foreground(s.theme.Primary)
	}

	var valueDisplay string
	switch opt.valueType {
	case "select":
		// Show with arrows for navigation hint
		valueDisplay = fmt.Sprintf("< %s >", opt.choices[opt.value])
	case "toggle":
		if opt.value == 1 {
			valueDisplay = "[x] Enabled"
		} else {
			valueDisplay = "[ ] Disabled"
		}
	}

	line := cursor + labelStyle.Render(opt.label) + ": " + valueStyle.Render(valueDisplay)
	if opt.description != "" && index == s.cursor {
		line += "\n    " + descStyle.Render(opt.description)
	}

	return line
}

func (s *SettingsView) incrementOption() {
	if s.cursor >= len(s.options) {
		return
	}

	opt := &s.options[s.cursor]
	if opt.valueType == "select" {
		if opt.value < len(opt.choices)-1 {
			opt.value++
			s.applyOption(s.cursor)
			s.dirty = true
		}
	} else if opt.valueType == "toggle" {
		if opt.value == 0 {
			opt.value = 1
			s.applyOption(s.cursor)
			s.dirty = true
		}
	}
}

func (s *SettingsView) decrementOption() {
	if s.cursor >= len(s.options) {
		return
	}

	opt := &s.options[s.cursor]
	if opt.valueType == "select" {
		if opt.value > 0 {
			opt.value--
			s.applyOption(s.cursor)
			s.dirty = true
		}
	} else if opt.valueType == "toggle" {
		if opt.value == 1 {
			opt.value = 0
			s.applyOption(s.cursor)
			s.dirty = true
		}
	}
}

func (s *SettingsView) toggleOption() {
	if s.cursor >= len(s.options) {
		return
	}

	opt := &s.options[s.cursor]
	if opt.valueType == "toggle" {
		if opt.value == 0 {
			opt.value = 1
		} else {
			opt.value = 0
		}
		s.applyOption(s.cursor)
		s.dirty = true
	} else if opt.valueType == "select" {
		// Cycle through options
		opt.value = (opt.value + 1) % len(opt.choices)
		s.applyOption(s.cursor)
		s.dirty = true
	}
}

func (s *SettingsView) applyOption(index int) {
	if s.config == nil || index >= len(s.options) {
		return
	}

	opt := s.options[index]
	switch opt.label {
	case "Theme":
		s.config.Settings.Theme = opt.choices[opt.value]
	case "Default Agent":
		s.config.Settings.DefaultAgent = opt.choices[opt.value]
	case "Auto-Update Check":
		s.config.Settings.AutoUpdateCheck = opt.value == 1
	}
}

func (s *SettingsView) save() tea.Cmd {
	return func() tea.Msg {
		if s.config == nil {
			return settingsSavedMsg{err: fmt.Errorf("no configuration loaded")}
		}
		if err := s.config.Save(); err != nil {
			return settingsSavedMsg{err: err}
		}
		return settingsSavedMsg{err: nil}
	}
}

// Message types for settings
type settingsLoadedMsg struct {
	config      *config.Config
	isaacConfig *config.IsaacConfig
	err         error
}

type settingsSavedMsg struct {
	err error
}
