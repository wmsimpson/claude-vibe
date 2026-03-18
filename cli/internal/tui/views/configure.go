// Package views provides TUI views for the vibe CLI.
package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui/components"
)

// TabName constants for the configure tabs.
const (
	TabSettings = "Settings"
	TabMCP      = "MCP Servers"
	TabPlugins  = "Plugins"
	TabProfiles = "Profiles"
)

// ConfigureView is the main configure view with tabbed interface.
type ConfigureView struct {
	tabs         components.Tabs
	settingsView *SettingsView
	mcpView      *MCPView
	pluginsView  *PluginsView
	profilesView *ProfilesView
	theme        tui.Theme
	width        int
	height       int
	statusMsg    string
	ready        bool
	initialScope string
}

// ConfigureOption configures a ConfigureView.
type ConfigureOption func(*ConfigureView)

// WithConfigureTheme sets a custom theme for the configure view.
func WithConfigureTheme(theme tui.Theme) ConfigureOption {
	return func(c *ConfigureView) {
		c.theme = theme
	}
}

// WithInitialTab sets the initial tab by name.
func WithInitialTab(tabName string) ConfigureOption {
	return func(c *ConfigureView) {
		tabs := []string{TabSettings, TabMCP, TabPlugins, TabProfiles}
		for i, name := range tabs {
			if strings.EqualFold(name, tabName) {
				c.tabs.SetActiveTab(i)
				break
			}
		}
	}
}

// WithInitialScope sets the initial scope for plugins/mcp views.
func WithInitialScope(scope string) ConfigureOption {
	return func(c *ConfigureView) {
		c.initialScope = scope
	}
}

// NewConfigureView creates a new ConfigureView.
func NewConfigureView(opts ...ConfigureOption) *ConfigureView {
	theme := tui.DefaultTheme

	c := &ConfigureView{
		tabs:         components.NewTabs([]string{TabSettings, TabMCP, TabPlugins, TabProfiles}, components.WithTabsTheme(theme)),
		settingsView: NewSettingsView(WithSettingsTheme(theme)),
		mcpView:      NewMCPView(WithMCPTheme(theme)),
		profilesView: NewProfilesView(WithProfilesTheme(theme)),
		theme:        theme,
	}

	// Apply options first to capture initialScope
	for _, opt := range opts {
		opt(c)
	}

	// Create plugins view with scope option if specified
	pluginOpts := []PluginsOption{WithPluginsTheme(theme)}
	if c.initialScope != "" {
		pluginOpts = append(pluginOpts, WithPluginsScope(c.initialScope))
	}
	c.pluginsView = NewPluginsView(pluginOpts...)

	return c
}

// Init implements tui.View.
func (c *ConfigureView) Init() tea.Cmd {
	return tea.Batch(
		c.settingsView.Init(),
		c.mcpView.Init(),
		c.pluginsView.Init(),
		c.profilesView.Init(),
	)
}

// Update implements tui.View.
func (c *ConfigureView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	var cmds []tea.Cmd
	var viewCmd tea.Cmd

	// Route view-specific messages to their views regardless of active tab.
	// This is important because Init() batches commands for ALL views, but
	// messages would otherwise only go to the active tab's view.
	switch msg.(type) {
	case mcpLoadedMsg, mcpSavedMsg:
		c.mcpView, viewCmd = c.mcpView.Update(msg)
		if viewCmd != nil {
			cmds = append(cmds, viewCmd)
		}
		return c, tea.Batch(cmds...)

	case pluginsLoadedMsg, pluginToggledMsg, pluginInstalledMsg, pluginUninstalledMsg:
		c.pluginsView, viewCmd = c.pluginsView.Update(msg)
		if viewCmd != nil {
			cmds = append(cmds, viewCmd)
		}
		return c, tea.Batch(cmds...)

	case profilesLoadedMsg, profileLoadedMsg, profileCreatedMsg:
		c.profilesView, viewCmd = c.profilesView.Update(msg)
		if viewCmd != nil {
			cmds = append(cmds, viewCmd)
		}
		return c, tea.Batch(cmds...)

	case settingsLoadedMsg, settingsSavedMsg:
		c.settingsView, viewCmd = c.settingsView.Update(msg)
		if viewCmd != nil {
			cmds = append(cmds, viewCmd)
		}
		return c, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.ready = true
		// Propagate to child views
		c.settingsView.SetSize(msg.Width, msg.Height-6)
		c.mcpView.SetSize(msg.Width, msg.Height-6)
		c.pluginsView.SetSize(msg.Width, msg.Height-6)
		c.profilesView.SetSize(msg.Width, msg.Height-6)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return c, tea.Quit
		}

	case components.TabChangedMsg:
		// Tab changed, reset status message
		c.statusMsg = ""

	case StatusMsg:
		c.statusMsg = msg.Message
	}

	// Update tabs
	newTabs, tabCmd := c.tabs.Update(msg)
	c.tabs = newTabs
	if tabCmd != nil {
		cmds = append(cmds, tabCmd)
	}

	// Update active tab's view
	switch c.tabs.ActiveTabName() {
	case TabSettings:
		c.settingsView, viewCmd = c.settingsView.Update(msg)
	case TabMCP:
		c.mcpView, viewCmd = c.mcpView.Update(msg)
	case TabPlugins:
		c.pluginsView, viewCmd = c.pluginsView.Update(msg)
	case TabProfiles:
		c.profilesView, viewCmd = c.profilesView.Update(msg)
	}
	if viewCmd != nil {
		cmds = append(cmds, viewCmd)
	}

	return c, tea.Batch(cmds...)
}

// View implements tui.View.
func (c *ConfigureView) View() string {
	if !c.ready {
		return "Loading..."
	}

	var content strings.Builder

	// Header
	headerStyle := c.theme.HeaderStyle().MarginBottom(0)
	content.WriteString(headerStyle.Render("Vibe Configuration"))
	content.WriteString("\n")

	// Tabs
	content.WriteString(c.tabs.View())
	content.WriteString("\n\n")

	// Active tab content
	switch c.tabs.ActiveTabName() {
	case TabSettings:
		content.WriteString(c.settingsView.View())
	case TabMCP:
		content.WriteString(c.mcpView.View())
	case TabPlugins:
		content.WriteString(c.pluginsView.View())
	case TabProfiles:
		content.WriteString(c.profilesView.View())
	}

	// Status bar at bottom
	if c.statusMsg != "" {
		content.WriteString("\n\n")
		statusStyle := c.theme.StatusBarStyle().Width(c.width)
		content.WriteString(statusStyle.Render(c.statusMsg))
	}

	// Global help
	content.WriteString("\n\n")
	helpStyle := c.theme.HelpStyle()
	content.WriteString(helpStyle.Render("Tab/Shift+Tab: switch tabs | Esc: quit"))

	// Add border around the whole view
	// MaxHeight ensures the header stays visible when content exceeds terminal height
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		MaxWidth(c.width).
		MaxHeight(c.height)

	return containerStyle.Render(content.String())
}

// StatusMsg is a message to display in the status bar.
type StatusMsg struct {
	Message string
}
