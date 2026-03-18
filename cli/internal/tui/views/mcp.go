package views

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/config"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui/components"
	"github.com/wmsimpson/claude-vibe/cli/internal/util"
)

// MCPView displays and manages MCP server configurations.
type MCPView struct {
	list             components.List
	mcpConfig        *config.MCPConfig
	projectMCPConfig *config.ProjectMCPConfig
	theme            tui.Theme
	width            int
	height           int
	dirty            bool
	loading          bool
	scope            string // "user" or "project"
	projectDir       string
	isHomeDir        bool // true if CWD is the home directory
}

// MCPOption configures an MCPView.
type MCPOption func(*MCPView)

// WithMCPTheme sets a custom theme.
func WithMCPTheme(theme tui.Theme) MCPOption {
	return func(m *MCPView) {
		m.theme = theme
	}
}

// NewMCPView creates a new MCPView.
func NewMCPView(opts ...MCPOption) *MCPView {
	// Get current working directory for project scope
	cwd, _ := os.Getwd()

	// Check if CWD is the home directory
	isHome := util.IsHomeDirectory(cwd)

	m := &MCPView{
		theme:      tui.DefaultTheme,
		loading:    true,
		scope:      ScopeUser, // Default to user scope
		projectDir: cwd,
		isHomeDir:  isHome,
	}

	for _, opt := range opts {
		opt(m)
	}

	m.list = components.NewList(nil,
		components.WithListTheme(m.theme),
		components.WithListHeight(10),
		components.WithListHelp(false),
	)

	return m
}

// Init implements tea.Model.
func (m *MCPView) Init() tea.Cmd {
	return func() tea.Msg {
		mc := config.NewMCPConfig()

		var pmc *config.ProjectMCPConfig
		if m.scope == ScopeProject {
			pmc = config.NewProjectMCPConfig(m.projectDir)
		}

		return mcpLoadedMsg{config: mc, projectConfig: pmc}
	}
}

// Update handles messages and returns the updated view.
func (m *MCPView) Update(msg tea.Msg) (*MCPView, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case mcpLoadedMsg:
		m.mcpConfig = msg.config
		m.projectMCPConfig = msg.projectConfig
		m.loading = false
		m.buildListItems()
		return m, nil

	case mcpSavedMsg:
		m.dirty = false
		if msg.err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Error saving MCP config: %v", msg.err)}
			}
		}
		return m, func() tea.Msg {
			return StatusMsg{Message: "MCP configuration saved successfully"}
		}

	case components.ListItemToggledMsg:
		// Handle toggle event
		item := msg.Item
		if item.Disabled {
			// Global server in project scope - can't toggle
			return m, func() tea.Msg {
				return StatusMsg{Message: "Cannot toggle global MCP servers in project scope"}
			}
		}

		if m.scope == ScopeUser {
			if m.mcpConfig != nil {
				servers := m.mcpConfig.ListServers()
				if msg.Index < len(servers) {
					serverName := servers[msg.Index].Name
					if err := m.mcpConfig.SetEnabled(serverName, msg.Enabled); err == nil {
						m.dirty = true
					}
				}
			}
		} else {
			if m.projectMCPConfig != nil {
				merged := m.projectMCPConfig.ListMerged()
				if msg.Index < len(merged) {
					serverName := merged[msg.Index].Server.Name
					if err := m.projectMCPConfig.SetEnabled(serverName, msg.Enabled); err == nil {
						m.dirty = true
						// Auto-save project scope changes
						if saveErr := m.projectMCPConfig.Save(); saveErr != nil {
							return m, func() tea.Msg {
								return StatusMsg{Message: fmt.Sprintf("Error saving: %v", saveErr)}
							}
						}
						m.dirty = false
						return m, func() tea.Msg {
							return StatusMsg{Message: "Changes saved"}
						}
					}
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			// In user scope, 's' is for save if dirty, otherwise switch scope
			if m.scope == ScopeUser && m.dirty {
				return m, m.save()
			}
			// Block switching to project scope if CWD is home directory
			if m.isHomeDir {
				return m, func() tea.Msg {
					return StatusMsg{Message: "Cannot switch to project scope: current directory is home (~). Project scope would be the same as user scope."}
				}
			}
			// Switch scope
			m.loading = true
			if m.scope == ScopeUser {
				m.scope = ScopeProject
			} else {
				m.scope = ScopeUser
			}
			m.dirty = false
			return m, m.Init()
		case "ctrl+s":
			if m.dirty {
				return m, m.save()
			}
		case "r":
			// Reload
			m.loading = true
			return m, m.Init()
		}
	}

	// Update list
	newList, listCmd := m.list.Update(msg)
	m.list = newList
	if listCmd != nil {
		cmds = append(cmds, listCmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the MCP view.
func (m *MCPView) View() string {
	if m.loading {
		return "Loading MCP servers..."
	}

	var lines []string

	// Section header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		MarginBottom(1)
	lines = append(lines, headerStyle.Render("MCP Servers"))

	// Scope indicator
	lines = append(lines, m.renderScopeIndicator())

	// Warning if CWD is home directory
	if m.isHomeDir {
		warningStyle := lipgloss.NewStyle().
			Foreground(m.theme.Warning).
			Italic(true)
		lines = append(lines, warningStyle.Render("⚠ Current directory is ~ (home). Project scope unavailable."))
	}
	lines = append(lines, "")

	// Subtitle
	subtitleStyle := m.theme.MutedStyle()
	if m.scope == ScopeUser {
		lines = append(lines, subtitleStyle.Render("Toggle servers on/off with Space | s: switch scope"))
	} else {
		lines = append(lines, subtitleStyle.Render("Toggle servers on/off with Space | s: switch scope | (global) = user-level"))
	}
	lines = append(lines, "")

	// Server list
	serverCount := 0
	if m.scope == ScopeUser && m.mcpConfig != nil {
		serverCount = len(m.mcpConfig.ListServers())
	} else if m.scope == ScopeProject && m.projectMCPConfig != nil {
		serverCount = len(m.projectMCPConfig.ListMerged())
	}

	if serverCount == 0 {
		lines = append(lines, m.theme.MutedStyle().Render("No MCP servers configured"))
	} else {
		lines = append(lines, m.list.View())
	}

	// Dirty indicator (only for user scope)
	if m.dirty && m.scope == ScopeUser {
		lines = append(lines, "")
		dirtyStyle := lipgloss.NewStyle().Foreground(m.theme.Warning)
		lines = append(lines, dirtyStyle.Render("* Unsaved changes - press 's' to save"))
	}

	// Legend
	lines = append(lines, "")
	legendStyle := m.theme.MutedStyle()
	if m.scope == ScopeProject {
		lines = append(lines, legendStyle.Render("Legend: [x] enabled | [ ] disabled | (global) = from user scope"))
	}

	// Help text
	lines = append(lines, "")
	helpStyle := m.theme.HelpStyle()
	if m.scope == ScopeUser {
		lines = append(lines, helpStyle.Render("j/k: navigate | space: toggle | s: save/switch scope | r: reload"))
	} else {
		lines = append(lines, helpStyle.Render("j/k: navigate | space: toggle | s: switch scope | r: reload"))
	}

	return strings.Join(lines, "\n")
}

// renderScopeIndicator renders the scope toggle UI
func (m *MCPView) renderScopeIndicator() string {
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Primary).
		Background(m.theme.Background).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(0, 1)

	disabledStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Strikethrough(true).
		Padding(0, 1)

	var userStyle, projectStyle lipgloss.Style
	if m.scope == ScopeUser {
		userStyle = activeStyle
		if m.isHomeDir {
			projectStyle = disabledStyle
		} else {
			projectStyle = inactiveStyle
		}
	} else {
		userStyle = inactiveStyle
		projectStyle = activeStyle
	}

	user := userStyle.Render("User (~/.claude)")
	projectLabel := "Project (./.claude.json)"
	if m.isHomeDir {
		projectLabel = "Project (N/A in ~)"
	}
	project := projectStyle.Render(projectLabel)

	return lipgloss.JoinHorizontal(lipgloss.Center, user, " | ", project)
}

// SetSize sets the view dimensions.
func (m *MCPView) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Adjust list height, accounting for headers and footers
	listHeight := height - 10 // Extra space for scope indicator
	if listHeight < 5 {
		listHeight = 5
	}
	m.list = components.NewList(m.list.Items(),
		components.WithListTheme(m.theme),
		components.WithListHeight(listHeight),
		components.WithListHelp(false),
	)
}

func (m *MCPView) buildListItems() {
	if m.scope == ScopeUser {
		m.buildUserScopeItems()
	} else {
		m.buildProjectScopeItems()
	}
}

func (m *MCPView) buildUserScopeItems() {
	if m.mcpConfig == nil {
		return
	}

	servers := m.mcpConfig.ListServers()
	items := make([]components.ListItem, len(servers))

	for i, server := range servers {
		// Build command description
		cmdParts := []string{server.Command}
		cmdParts = append(cmdParts, server.Args...)
		cmdDisplay := strings.Join(cmdParts, " ")

		// Truncate if too long
		maxCmdLen := 60
		if len(cmdDisplay) > maxCmdLen {
			cmdDisplay = cmdDisplay[:maxCmdLen-3] + "..."
		}

		items[i] = components.ListItem{
			ID:          server.Name,
			Title:       server.Name,
			Description: cmdDisplay,
			Enabled:     server.Enabled,
			Disabled:    false,
			Dimmed:      !server.Enabled,
		}
	}

	m.list.SetItems(items)
}

func (m *MCPView) buildProjectScopeItems() {
	if m.projectMCPConfig == nil {
		m.list.SetItems(nil)
		return
	}

	merged := m.projectMCPConfig.ListMerged()
	items := make([]components.ListItem, len(merged))

	for i, swc := range merged {
		// Build command description
		cmdParts := []string{swc.Server.Command}
		cmdParts = append(cmdParts, swc.Server.Args...)
		cmdDisplay := strings.Join(cmdParts, " ")

		// Truncate if too long
		maxCmdLen := 50
		if len(cmdDisplay) > maxCmdLen {
			cmdDisplay = cmdDisplay[:maxCmdLen-3] + "..."
		}

		if swc.IsGlobal {
			cmdDisplay += " (global)"
		}

		items[i] = components.ListItem{
			ID:          swc.Server.Name,
			Title:       swc.Server.Name,
			Description: cmdDisplay,
			Enabled:     swc.Server.Enabled,
			Disabled:    swc.IsGlobal, // Global servers can't be toggled
			Dimmed:      !swc.Server.Enabled,
		}
	}

	m.list.SetItems(items)
}

func (m *MCPView) save() tea.Cmd {
	return func() tea.Msg {
		if m.scope == ScopeUser {
			if m.mcpConfig == nil {
				return mcpSavedMsg{err: fmt.Errorf("no MCP configuration loaded")}
			}
			err := m.mcpConfig.Save()
			return mcpSavedMsg{err: err}
		} else {
			if m.projectMCPConfig == nil {
				return mcpSavedMsg{err: fmt.Errorf("no project MCP configuration loaded")}
			}
			err := m.projectMCPConfig.Save()
			return mcpSavedMsg{err: err}
		}
	}
}

// Message types for MCP
type mcpLoadedMsg struct {
	config        *config.MCPConfig
	projectConfig *config.ProjectMCPConfig
}

type mcpSavedMsg struct {
	err error
}
