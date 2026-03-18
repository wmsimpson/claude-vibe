package views

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/config"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui/components"
	"github.com/wmsimpson/claude-vibe/cli/internal/util"
)

// Scope constants
const (
	ScopeUser    = "user"
	ScopeProject = "project"
)

// PluginsView displays and manages Claude Code plugins.
type PluginsView struct {
	list                components.List
	pluginConfig        *config.PluginConfig
	projectPluginConfig *config.ProjectPluginConfig
	installed           []config.Plugin
	available           []config.AvailablePlugin
	theme               tui.Theme
	width               int
	height              int
	loading             bool
	installing          bool
	confirmDelete       bool
	deleteIndex         int
	scope               string // "user" or "project"
	projectDir          string
	isHomeDir           bool // true if CWD is the home directory
}

// PluginsOption configures a PluginsView.
type PluginsOption func(*PluginsView)

// WithPluginsTheme sets a custom theme.
func WithPluginsTheme(theme tui.Theme) PluginsOption {
	return func(p *PluginsView) {
		p.theme = theme
	}
}

// WithPluginsScope sets the initial scope.
func WithPluginsScope(scope string) PluginsOption {
	return func(p *PluginsView) {
		if scope == ScopeProject || scope == ScopeUser {
			p.scope = scope
		}
	}
}

// NewPluginsView creates a new PluginsView.
func NewPluginsView(opts ...PluginsOption) *PluginsView {
	// Get current working directory for project scope
	cwd, _ := os.Getwd()

	// Check if CWD is the home directory
	isHome := util.IsHomeDirectory(cwd)

	p := &PluginsView{
		theme:      tui.DefaultTheme,
		loading:    true,
		scope:      ScopeUser, // Default to user scope
		projectDir: cwd,
		isHomeDir:  isHome,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.list = components.NewList(nil,
		components.WithListTheme(p.theme),
		components.WithListHeight(10),
		components.WithListHelp(false),
	)

	return p
}

// Init implements tea.Model.
func (p *PluginsView) Init() tea.Cmd {
	return func() tea.Msg {
		pc := config.NewPluginConfig()
		installed := pc.ListInstalled()
		available := pc.ListAvailable()

		var ppc *config.ProjectPluginConfig
		if p.scope == ScopeProject {
			ppc = config.NewProjectPluginConfig(p.projectDir)
		}

		return pluginsLoadedMsg{
			config:        pc,
			projectConfig: ppc,
			installed:     installed,
			available:     available,
		}
	}
}

// Update handles messages and returns the updated view.
func (p *PluginsView) Update(msg tea.Msg) (*PluginsView, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case pluginsLoadedMsg:
		p.pluginConfig = msg.config
		p.projectPluginConfig = msg.projectConfig
		p.installed = msg.installed
		p.available = msg.available
		p.loading = false
		p.buildListItems()
		return p, nil

	case pluginToggledMsg:
		if msg.err != nil {
			return p, func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Error: %v", msg.err)}
			}
		}
		// Refresh the list - set loading to true so UI shows loading state
		p.loading = true
		return p, p.Init()

	case pluginInstalledMsg:
		p.installing = false
		if msg.err != nil {
			return p, func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Install failed: %v", msg.err)}
			}
		}
		// Reload plugins - set loading to true so UI shows loading state
		// while we wait for the fresh data
		p.loading = true
		return p, tea.Batch(
			func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Installed %s successfully", msg.name)}
			},
			p.Init(),
		)

	case pluginUninstalledMsg:
		p.confirmDelete = false
		if msg.err != nil {
			return p, func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Uninstall failed: %v", msg.err)}
			}
		}
		// Reload plugins - set loading to true so UI shows loading state
		// while we wait for the fresh data
		p.loading = true
		return p, tea.Batch(
			func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Uninstalled %s successfully", msg.name)}
			},
			p.Init(),
		)

	case components.ListItemToggledMsg:
		// Handle toggle - for installed plugins, enable/disable
		// For not installed plugins, trigger install
		item := msg.Item
		if item.Dimmed {
			// Not installed - install it at current scope
			p.installing = true
			return p, p.installPlugin(item.ID, p.scope)
		} else if item.Disabled {
			// Global plugin in project scope - can't toggle
			return p, func() tea.Msg {
				return StatusMsg{Message: "Cannot toggle global plugins in project scope"}
			}
		} else {
			// Toggle enable/disable
			return p, p.togglePlugin(item.ID, msg.Enabled)
		}

	case components.ConfirmResultMsg:
		if p.confirmDelete && msg.Confirmed {
			// User confirmed deletion
			items := p.list.Items()
			if p.deleteIndex >= 0 && p.deleteIndex < len(items) {
				return p, p.uninstallPlugin(items[p.deleteIndex].ID, p.scope)
			}
		}
		p.confirmDelete = false
		return p, nil

	case tea.KeyMsg:
		if p.confirmDelete {
			// Handle confirm dialog keys
			switch msg.String() {
			case "y", "Y":
				items := p.list.Items()
				if p.deleteIndex >= 0 && p.deleteIndex < len(items) {
					p.confirmDelete = false
					return p, p.uninstallPlugin(items[p.deleteIndex].ID, p.scope)
				}
			case "n", "N", "esc":
				p.confirmDelete = false
				return p, nil
			}
			return p, nil
		}

		switch msg.String() {
		case "s", "tab":
			// Switch scope - but block if CWD is home directory
			if p.isHomeDir {
				return p, func() tea.Msg {
					return StatusMsg{Message: "Cannot switch to project scope: current directory is home (~). Project scope would be the same as user scope."}
				}
			}
			p.loading = true
			if p.scope == ScopeUser {
				p.scope = ScopeProject
			} else {
				p.scope = ScopeUser
			}
			return p, p.Init()
		case "d", "delete", "backspace":
			// Delete/uninstall selected plugin
			items := p.list.Items()
			cursor := p.list.Cursor()
			if cursor >= 0 && cursor < len(items) {
				item := items[cursor]
				if item.Dimmed {
					// Not installed - nothing to uninstall
					return p, func() tea.Msg {
						return StatusMsg{Message: "Plugin is not installed"}
					}
				}
				if p.scope == ScopeProject && item.Disabled {
					// Global plugin in project scope - can't uninstall here
					return p, func() tea.Msg {
						return StatusMsg{Message: "Global plugins can only be uninstalled from user scope"}
					}
				}
				// Can uninstall - show confirmation
				p.confirmDelete = true
				p.deleteIndex = cursor
				return p, nil
			}
		case "r":
			// Reload
			p.loading = true
			return p, p.Init()
		}
	}

	// Update list (only if not in confirm mode)
	if !p.confirmDelete {
		newList, listCmd := p.list.Update(msg)
		p.list = newList
		if listCmd != nil {
			cmds = append(cmds, listCmd)
		}
	}

	return p, tea.Batch(cmds...)
}

// View renders the plugins view.
func (p *PluginsView) View() string {
	if p.loading {
		return "Loading plugins..."
	}

	if p.installing {
		return "Installing plugin..."
	}

	var lines []string

	// Section header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.theme.Primary).
		MarginBottom(1)
	lines = append(lines, headerStyle.Render("Plugins"))

	// Scope indicator
	lines = append(lines, p.renderScopeIndicator())

	// Warning if CWD is home directory
	if p.isHomeDir {
		warningStyle := lipgloss.NewStyle().
			Foreground(p.theme.Warning).
			Italic(true)
		lines = append(lines, warningStyle.Render("⚠ Current directory is ~ (home). Project scope unavailable."))
	}
	lines = append(lines, "")

	// Subtitle
	subtitleStyle := p.theme.MutedStyle()
	if p.scope == ScopeUser {
		lines = append(lines, subtitleStyle.Render("Space: toggle/install | d: uninstall | s/Tab: switch scope"))
	} else {
		lines = append(lines, subtitleStyle.Render("Space: toggle | d: uninstall | s/Tab: switch scope | (global) = user-level"))
	}
	lines = append(lines, "")

	// Confirm delete dialog
	if p.confirmDelete {
		items := p.list.Items()
		if p.deleteIndex >= 0 && p.deleteIndex < len(items) {
			confirmStyle := lipgloss.NewStyle().
				Foreground(p.theme.Warning).
				Bold(true)
			lines = append(lines, confirmStyle.Render(
				fmt.Sprintf("Uninstall %s? (y/n)", items[p.deleteIndex].Title),
			))
			lines = append(lines, "")
		}
	}

	// Plugin list
	if len(p.list.Items()) == 0 {
		lines = append(lines, p.theme.MutedStyle().Render("No plugins available"))
	} else {
		lines = append(lines, p.list.View())
	}

	// Legend
	lines = append(lines, "")
	legendStyle := p.theme.MutedStyle()
	if p.scope == ScopeUser {
		lines = append(lines, legendStyle.Render("Legend: [x] enabled | [ ] disabled | dim = not installed"))
	} else {
		lines = append(lines, legendStyle.Render("Legend: [x] enabled | [ ] disabled | (global) = from user scope"))
	}

	// Help text
	lines = append(lines, "")
	helpStyle := p.theme.HelpStyle()
	lines = append(lines, helpStyle.Render("j/k: navigate | space: toggle/install | s: switch scope | r: reload"))

	return strings.Join(lines, "\n")
}

// renderScopeIndicator renders the scope toggle UI
func (p *PluginsView) renderScopeIndicator() string {
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.theme.Primary).
		Background(p.theme.Background).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(p.theme.Muted).
		Padding(0, 1)

	disabledStyle := lipgloss.NewStyle().
		Foreground(p.theme.Muted).
		Strikethrough(true).
		Padding(0, 1)

	var userStyle, projectStyle lipgloss.Style
	if p.scope == ScopeUser {
		userStyle = activeStyle
		if p.isHomeDir {
			projectStyle = disabledStyle
		} else {
			projectStyle = inactiveStyle
		}
	} else {
		userStyle = inactiveStyle
		projectStyle = activeStyle
	}

	user := userStyle.Render("User (~/.claude)")
	projectLabel := "Project (./.claude)"
	if p.isHomeDir {
		projectLabel = "Project (N/A in ~)"
	}
	project := projectStyle.Render(projectLabel)

	return lipgloss.JoinHorizontal(lipgloss.Center, user, " | ", project)
}

// SetSize sets the view dimensions.
func (p *PluginsView) SetSize(width, height int) {
	p.width = width
	p.height = height
	// Adjust list height
	listHeight := height - 12 // Extra space for scope indicator
	if listHeight < 5 {
		listHeight = 5
	}
	p.list = components.NewList(p.list.Items(),
		components.WithListTheme(p.theme),
		components.WithListHeight(listHeight),
		components.WithListHelp(false),
	)
}

func (p *PluginsView) buildListItems() {
	if p.scope == ScopeUser {
		p.buildUserScopeItems()
	} else {
		p.buildProjectScopeItems()
	}
}

func (p *PluginsView) buildUserScopeItems() {
	// Build a map of user-scoped installed plugins for quick lookup
	// Only include plugins with scope "user" - project-scoped plugins should
	// appear as "not installed" in the user scope view
	installedMap := make(map[string]config.Plugin)
	for _, plugin := range p.installed {
		if plugin.Scope == "user" {
			installedMap[plugin.Name] = plugin
		}
	}

	// Build unified list showing ALL marketplace plugins in order
	// User-scoped installed plugins show version, not installed show description and are dimmed
	var items []components.ListItem

	// Show all available plugins from marketplace in marketplace order
	for _, avail := range p.available {
		if installed, ok := installedMap[avail.Name]; ok {
			// Plugin is installed at user scope - show with version
			items = append(items, components.ListItem{
				ID:          installed.Name,
				Title:       installed.Name,
				Description: fmt.Sprintf("v%s (user)", installed.Version),
				Enabled:     installed.Enabled,
				Disabled:    false,
				Dimmed:      false,
			})
		} else {
			// Plugin is not installed at user scope - show description and dimmed
			desc := avail.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			items = append(items, components.ListItem{
				ID:          avail.Name,
				Title:       avail.Name,
				Description: desc,
				Enabled:     false,
				Disabled:    false,
				Dimmed:      true, // Not installed at user scope = dimmed
			})
		}
	}

	p.list.SetItems(items)
}

func (p *PluginsView) buildProjectScopeItems() {
	if p.projectPluginConfig == nil {
		p.list.SetItems(nil)
		return
	}

	// Build a map of installed plugins (user + project scope) for quick lookup
	// Exclude disabled global plugins - they should appear as "available to install"
	merged := p.projectPluginConfig.ListMerged()
	installedMap := make(map[string]config.PluginWithScope)
	for _, pwc := range merged {
		// Skip disabled global plugins - treat them as not installed in project scope
		if pwc.IsGlobal && !pwc.Plugin.Enabled {
			continue
		}
		installedMap[pwc.Plugin.Name] = pwc
	}

	// Build unified list showing ALL marketplace plugins
	// Installed plugins show version/scope, not installed show description and are dimmed
	var items []components.ListItem

	// Show all available plugins from marketplace in marketplace order
	for _, avail := range p.available {
		if pwc, ok := installedMap[avail.Name]; ok {
			// Plugin is installed - show with version and scope indicator
			desc := fmt.Sprintf("v%s", pwc.Plugin.Version)
			if pwc.IsGlobal {
				desc += " (global)"
			}
			items = append(items, components.ListItem{
				ID:          pwc.Plugin.Name,
				Title:       pwc.Plugin.Name,
				Description: desc,
				Enabled:     pwc.Plugin.Enabled,
				Disabled:    pwc.IsGlobal, // Global plugins can't be toggled
				Dimmed:      false,
			})
		} else {
			// Plugin is not installed - show description and dimmed
			desc := avail.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			items = append(items, components.ListItem{
				ID:          avail.Name,
				Title:       avail.Name,
				Description: desc,
				Enabled:     false,
				Disabled:    false,
				Dimmed:      true, // Not installed = dimmed
			})
		}
	}

	p.list.SetItems(items)
}

func (p *PluginsView) togglePlugin(name string, enabled bool) tea.Cmd {
	return func() tea.Msg {
		if p.scope == ScopeUser {
			if p.pluginConfig == nil {
				return pluginToggledMsg{name: name, err: fmt.Errorf("no plugin config")}
			}
			err := p.pluginConfig.SetEnabled(name, enabled)
			return pluginToggledMsg{name: name, err: err}
		} else {
			if p.projectPluginConfig == nil {
				return pluginToggledMsg{name: name, err: fmt.Errorf("no project plugin config")}
			}
			err := p.projectPluginConfig.SetEnabled(name, enabled)
			if err == nil {
				// Auto-save project scope changes
				err = p.projectPluginConfig.Save()
			}
			return pluginToggledMsg{name: name, err: err}
		}
	}
}

func (p *PluginsView) installPlugin(name string, scope string) tea.Cmd {
	return func() tea.Msg {
		// Use claude CLI to install the plugin with the appropriate scope
		cmd := exec.Command("claude", "plugin", "install", "--scope", scope, name+"@claude-vibe")
		err := cmd.Run()
		return pluginInstalledMsg{name: name, err: err}
	}
}

func (p *PluginsView) uninstallPlugin(name string, scope string) tea.Cmd {
	return func() tea.Msg {
		// Use claude CLI to uninstall the plugin with the appropriate scope
		cmd := exec.Command("claude", "plugin", "uninstall", "--scope", scope, name)
		err := cmd.Run()
		return pluginUninstalledMsg{name: name, err: err}
	}
}

// Message types for plugins
type pluginsLoadedMsg struct {
	config        *config.PluginConfig
	projectConfig *config.ProjectPluginConfig
	installed     []config.Plugin
	available     []config.AvailablePlugin
}

type pluginToggledMsg struct {
	name string
	err  error
}

type pluginInstalledMsg struct {
	name string
	err  error
}

type pluginUninstalledMsg struct {
	name string
	err  error
}
