// Package views provides TUI view implementations for the vibe CLI.
package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/config"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui/components"
)

// ProfilesView states
type profileViewState int

const (
	stateList profileViewState = iota
	stateDetail           // Legacy - will redirect to stateDetailMenu
	stateDetailMenu       // Shows submenu with Plugins/MCP/Permissions
	stateDetailPlugins    // Paginated plugins list
	stateDetailMCP        // Paginated MCP servers list
	stateDetailPermissions // Paginated permissions list
	stateNaming           // Creating a new profile - entering name
	stateRenaming         // Renaming an existing profile
	stateConfirmDelete    // Confirming profile deletion
)

// detailCategory represents the selected category in the detail menu
type detailCategory int

const (
	categoryPlugins detailCategory = iota
	categoryMCP
	categoryPermissions
)

// Items per page for pagination
const itemsPerPage = 12

// ProfilesView displays and manages vibe profiles.
type ProfilesView struct {
	list            components.List
	profiles        []string
	selectedProfile *config.Profile
	defaultProfile  *config.Profile
	theme           tui.Theme
	width           int
	height          int
	loading         bool
	state           profileViewState

	// Text input for naming/renaming
	textInput textinput.Model
	inputErr  string

	// Confirm dialog for deletion
	confirmDialog components.Confirm
	deleteTarget  string

	// Global settings for display
	globalPlugins     []config.Plugin
	globalMCPServers  []config.MCPServer
	globalPermissions *config.PermissionsConfig

	// Available plugins from marketplace
	availablePlugins []config.AvailablePlugin

	// Detail menu navigation
	detailMenuCursor int

	// Category list navigation (cursor within current page)
	categoryCursor int

	// Paginator for detail views
	paginator paginator.Model

	// Cached items for pagination
	pluginItems     []detailItem
	mcpItems        []detailItem
	permissionItems []detailItem
}

// detailItem represents an item in a detail list with global indicator
type detailItem struct {
	name     string
	isGlobal bool
	subType  string // For permissions: "allow" or "deny"
	enabled  bool   // For MCP servers
}

// ProfilesOption configures a ProfilesView.
type ProfilesOption func(*ProfilesView)

// WithProfilesTheme sets a custom theme.
func WithProfilesTheme(theme tui.Theme) ProfilesOption {
	return func(p *ProfilesView) {
		p.theme = theme
	}
}

// NewProfilesView creates a new ProfilesView.
func NewProfilesView(opts ...ProfilesOption) *ProfilesView {
	p := &ProfilesView{
		theme:   tui.DefaultTheme,
		loading: true,
		state:   stateList,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.list = components.NewList(nil,
		components.WithListTheme(p.theme),
		components.WithListHeight(10),
		components.WithListHelp(false),
	)

	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "profile-name"
	ti.CharLimit = 64
	ti.Width = 40
	p.textInput = ti

	// Initialize paginator
	pg := paginator.New()
	pg.Type = paginator.Dots
	pg.PerPage = itemsPerPage
	pg.ActiveDot = lipgloss.NewStyle().Foreground(p.theme.Primary).Render("●")
	pg.InactiveDot = lipgloss.NewStyle().Foreground(p.theme.Muted).Render("○")
	p.paginator = pg

	return p
}

// Init implements tea.Model.
func (p *ProfilesView) Init() tea.Cmd {
	return func() tea.Msg {
		profiles := config.ListProfiles()

		// Try to load the default profile
		var defaultProfile *config.Profile
		for _, name := range profiles {
			if name == "default" {
				if prof, err := config.LoadProfile("default"); err == nil {
					defaultProfile = prof
				}
				break
			}
		}

		// Load global settings
		pluginConfig := config.NewPluginConfig()
		globalPlugins := pluginConfig.ListInstalledUserScope()
		availablePlugins := pluginConfig.ListAvailable()

		mcpConfig := config.NewMCPConfig()
		globalMCPServers := mcpConfig.ListServers()

		globalPermissions := config.NewPermissionsConfig()

		return profilesLoadedMsg{
			profiles:          profiles,
			defaultProfile:    defaultProfile,
			globalPlugins:     globalPlugins,
			globalMCPServers:  globalMCPServers,
			globalPermissions: globalPermissions,
			availablePlugins:  availablePlugins,
		}
	}
}

// Update handles messages and returns the updated view.
func (p *ProfilesView) Update(msg tea.Msg) (*ProfilesView, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case profilesLoadedMsg:
		p.profiles = msg.profiles
		p.defaultProfile = msg.defaultProfile
		p.globalPlugins = msg.globalPlugins
		p.globalMCPServers = msg.globalMCPServers
		p.globalPermissions = msg.globalPermissions
		p.availablePlugins = msg.availablePlugins
		p.loading = false
		p.buildListItems()
		return p, nil

	case profileLoadedMsg:
		if msg.err != nil {
			return p, func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Error loading profile: %v", msg.err)}
			}
		}
		p.selectedProfile = msg.profile
		p.detailMenuCursor = 0
		p.buildDetailItems()
		p.state = stateDetailMenu
		return p, nil

	case profileCreatedMsg:
		if msg.err != nil {
			return p, func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Error creating profile: %v", msg.err)}
			}
		}
		// Reload profiles
		return p, tea.Batch(
			func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Created profile: %s", msg.name)}
			},
			p.Init(),
		)

	case profileDeletedMsg:
		if msg.err != nil {
			return p, func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Error deleting profile: %v", msg.err)}
			}
		}
		// Reload profiles
		return p, tea.Batch(
			func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Deleted profile: %s", msg.name)}
			},
			p.Init(),
		)

	case profileRenamedMsg:
		if msg.err != nil {
			return p, func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Error renaming profile: %v", msg.err)}
			}
		}
		// Reload profile and profiles list
		return p, tea.Batch(
			func() tea.Msg {
				return StatusMsg{Message: fmt.Sprintf("Renamed profile to: %s", msg.newName)}
			},
			p.loadProfile(msg.newName),
			p.Init(),
		)

	case components.ConfirmResultMsg:
		if p.state == stateConfirmDelete {
			if msg.Confirmed {
				// Delete the profile
				name := p.deleteTarget
				p.state = stateList
				p.deleteTarget = ""
				return p, p.deleteProfile(name)
			}
			p.state = stateList
			p.deleteTarget = ""
		}
		return p, nil

	case components.ConfirmCancelledMsg:
		if p.state == stateConfirmDelete {
			p.state = stateList
			p.deleteTarget = ""
		}
		return p, nil

	case components.ListItemSelectedMsg:
		// Profile selected - load its details
		if msg.Item.ID == "_new_profile" {
			// Show text input for new profile name
			p.state = stateNaming
			p.textInput.SetValue("")
			p.textInput.Focus()
			p.inputErr = ""
			return p, textinput.Blink
		}
		return p, p.loadProfile(msg.Item.ID)

	case tea.KeyMsg:
		switch p.state {
		case stateNaming:
			return p.handleNamingKeys(msg)
		case stateRenaming:
			return p.handleRenamingKeys(msg)
		case stateConfirmDelete:
			return p.handleConfirmDeleteKeys(msg)
		case stateDetail:
			// Redirect legacy state to new detail menu
			p.state = stateDetailMenu
			p.detailMenuCursor = 0
			return p, nil
		case stateDetailMenu:
			return p.handleDetailMenuKeys(msg)
		case stateDetailPlugins, stateDetailMCP, stateDetailPermissions:
			return p.handleDetailCategoryKeys(msg)
		default:
			return p.handleListKeys(msg)
		}
	}

	// Update text input if in naming/renaming state
	if p.state == stateNaming || p.state == stateRenaming {
		var cmd tea.Cmd
		p.textInput, cmd = p.textInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update confirm dialog if in delete confirmation state
	if p.state == stateConfirmDelete {
		var cmd tea.Cmd
		p.confirmDialog, cmd = p.confirmDialog.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update list (only if in list state)
	if p.state == stateList {
		newList, listCmd := p.list.Update(msg)
		p.list = newList
		if listCmd != nil {
			cmds = append(cmds, listCmd)
		}
	}

	return p, tea.Batch(cmds...)
}

func (p *ProfilesView) handleListKeys(msg tea.KeyMsg) (*ProfilesView, tea.Cmd) {
	switch msg.String() {
	case "n":
		// New profile - show text input
		p.state = stateNaming
		p.textInput.SetValue("")
		p.textInput.Focus()
		p.inputErr = ""
		return p, textinput.Blink
	case "d":
		// Delete selected profile
		if item := p.list.SelectedItem(); item != nil && item.ID != "_new_profile" {
			p.deleteTarget = item.ID
			p.confirmDialog = components.NewConfirm(
				fmt.Sprintf("Delete profile '%s'?", item.ID),
				components.WithConfirmTheme(p.theme),
			)
			p.state = stateConfirmDelete
		}
		return p, nil
	case "r":
		// Reload
		p.loading = true
		return p, p.Init()
	}

	// Pass key events to list component for navigation (up/down/j/k/enter/etc.)
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p *ProfilesView) handleDetailKeys(msg tea.KeyMsg) (*ProfilesView, tea.Cmd) {
	// Legacy function - redirect to detail menu
	p.state = stateDetailMenu
	p.detailMenuCursor = 0
	return p, nil
}

func (p *ProfilesView) handleDetailMenuKeys(msg tea.KeyMsg) (*ProfilesView, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace", "q":
		p.state = stateList
		p.selectedProfile = nil
		return p, nil
	case "up", "k":
		if p.detailMenuCursor > 0 {
			p.detailMenuCursor--
		}
		return p, nil
	case "down", "j":
		if p.detailMenuCursor < 2 {
			p.detailMenuCursor++
		}
		return p, nil
	case "enter":
		// Enter the selected category
		switch detailCategory(p.detailMenuCursor) {
		case categoryPlugins:
			p.paginator.SetTotalPages(p.calculatePages(len(p.pluginItems)))
			p.paginator.Page = 0
			p.categoryCursor = 0
			p.state = stateDetailPlugins
		case categoryMCP:
			p.paginator.SetTotalPages(p.calculatePages(len(p.mcpItems)))
			p.paginator.Page = 0
			p.categoryCursor = 0
			p.state = stateDetailMCP
		case categoryPermissions:
			p.paginator.SetTotalPages(p.calculatePages(len(p.permissionItems)))
			p.paginator.Page = 0
			p.categoryCursor = 0
			p.state = stateDetailPermissions
		}
		return p, nil
	case "r":
		// Rename profile
		if p.selectedProfile != nil {
			p.state = stateRenaming
			p.textInput.SetValue(p.selectedProfile.Name)
			p.textInput.Focus()
			p.inputErr = ""
			return p, textinput.Blink
		}
	}
	return p, nil
}

func (p *ProfilesView) handleDetailCategoryKeys(msg tea.KeyMsg) (*ProfilesView, tea.Cmd) {
	// Get the current items based on state
	var items []detailItem
	switch p.state {
	case stateDetailPlugins:
		items = p.pluginItems
	case stateDetailMCP:
		items = p.mcpItems
	case stateDetailPermissions:
		items = p.permissionItems
	}

	// Calculate items on current page
	start, end := p.paginator.GetSliceBounds(len(items))
	itemsOnPage := end - start

	switch msg.String() {
	case "esc", "backspace", "q":
		p.state = stateDetailMenu
		return p, nil
	case "left", "h":
		if p.paginator.Page > 0 {
			p.paginator.PrevPage()
			p.categoryCursor = 0 // Reset cursor when changing pages
		}
		return p, nil
	case "right", "l":
		if p.paginator.Page < p.paginator.TotalPages-1 {
			p.paginator.NextPage()
			p.categoryCursor = 0 // Reset cursor when changing pages
		}
		return p, nil
	case "up", "k":
		if p.categoryCursor > 0 {
			p.categoryCursor--
		}
		return p, nil
	case "down", "j":
		if p.categoryCursor < itemsOnPage-1 {
			p.categoryCursor++
		}
		return p, nil
	case " ": // Space to toggle
		if itemsOnPage == 0 {
			return p, nil
		}
		// Get the actual item index in the full list
		itemIndex := start + p.categoryCursor
		if itemIndex >= len(items) {
			return p, nil
		}
		item := items[itemIndex]

		// Don't allow toggling global items
		if item.isGlobal {
			return p, func() tea.Msg {
				return StatusMsg{Message: "Cannot toggle global items - they are managed at the user level"}
			}
		}

		// Toggle the item based on category
		return p.toggleItem(itemIndex, item)
	}
	return p, nil
}

// toggleItem toggles an item in the profile and saves the changes
func (p *ProfilesView) toggleItem(index int, item detailItem) (*ProfilesView, tea.Cmd) {
	if p.selectedProfile == nil {
		return p, nil
	}

	switch p.state {
	case stateDetailPlugins:
		return p.togglePlugin(item.name)
	case stateDetailMCP:
		return p.toggleMCP(item.name)
	case stateDetailPermissions:
		return p.togglePermission(item.name, item.subType)
	}

	return p, nil
}

// togglePlugin adds or removes a plugin from the profile
func (p *ProfilesView) togglePlugin(name string) (*ProfilesView, tea.Cmd) {
	prof := p.selectedProfile

	// Check if plugin is in profile
	found := -1
	for i, plugin := range prof.Plugins {
		if plugin == name {
			found = i
			break
		}
	}

	var action string
	if found >= 0 {
		// Remove from profile
		prof.Plugins = append(prof.Plugins[:found], prof.Plugins[found+1:]...)
		action = "disabled"
	} else {
		// Add to profile
		prof.Plugins = append(prof.Plugins, name)
		action = "enabled"
	}

	// Save and rebuild items
	if err := prof.Save(); err != nil {
		return p, func() tea.Msg {
			return StatusMsg{Message: fmt.Sprintf("Error saving profile: %v", err)}
		}
	}

	p.buildDetailItems()
	return p, func() tea.Msg {
		return StatusMsg{Message: fmt.Sprintf("Plugin '%s' %s", name, action)}
	}
}

// toggleMCP toggles an MCP server in the profile
func (p *ProfilesView) toggleMCP(name string) (*ProfilesView, tea.Cmd) {
	prof := p.selectedProfile

	if prof.MCPServers == nil {
		prof.MCPServers = make(map[string]bool)
	}

	var action string
	if enabled, exists := prof.MCPServers[name]; exists && enabled {
		// Disable it
		prof.MCPServers[name] = false
		action = "disabled"
	} else {
		// Enable it
		prof.MCPServers[name] = true
		action = "enabled"
	}

	// Save and rebuild items
	if err := prof.Save(); err != nil {
		return p, func() tea.Msg {
			return StatusMsg{Message: fmt.Sprintf("Error saving profile: %v", err)}
		}
	}

	p.buildDetailItems()
	return p, func() tea.Msg {
		return StatusMsg{Message: fmt.Sprintf("MCP server '%s' %s", name, action)}
	}
}

// togglePermission adds or removes a permission from the profile
func (p *ProfilesView) togglePermission(name, subType string) (*ProfilesView, tea.Cmd) {
	prof := p.selectedProfile

	var list *[]string
	if subType == "allow" {
		list = &prof.Permissions.Allow
	} else {
		list = &prof.Permissions.Deny
	}

	// Check if permission is in list
	found := -1
	for i, perm := range *list {
		if perm == name {
			found = i
			break
		}
	}

	var action string
	if found >= 0 {
		// Remove from list
		*list = append((*list)[:found], (*list)[found+1:]...)
		action = "removed"
	} else {
		// Add to list
		*list = append(*list, name)
		action = "added"
	}

	// Save and rebuild items
	if err := prof.Save(); err != nil {
		return p, func() tea.Msg {
			return StatusMsg{Message: fmt.Sprintf("Error saving profile: %v", err)}
		}
	}

	p.buildDetailItems()
	return p, func() tea.Msg {
		return StatusMsg{Message: fmt.Sprintf("Permission '%s' %s", name, action)}
	}
}

func (p *ProfilesView) calculatePages(totalItems int) int {
	if totalItems == 0 {
		return 1
	}
	pages := totalItems / itemsPerPage
	if totalItems%itemsPerPage > 0 {
		pages++
	}
	return pages
}

func (p *ProfilesView) handleNamingKeys(msg tea.KeyMsg) (*ProfilesView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		p.state = stateList
		p.inputErr = ""
		return p, nil
	case "enter":
		name := strings.TrimSpace(p.textInput.Value())

		// Validate name
		if err := config.ValidateProfileName(name); err != nil {
			p.inputErr = err.Error()
			return p, nil
		}

		// Check for duplicates
		if config.ProfileExists(name) {
			p.inputErr = "A profile with this name already exists"
			return p, nil
		}

		p.state = stateList
		p.inputErr = ""
		return p, p.createNewProfile(name)
	}

	// Update text input
	var cmd tea.Cmd
	p.textInput, cmd = p.textInput.Update(msg)
	return p, cmd
}

func (p *ProfilesView) handleRenamingKeys(msg tea.KeyMsg) (*ProfilesView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		p.state = stateDetailMenu
		p.inputErr = ""
		return p, nil
	case "enter":
		newName := strings.TrimSpace(p.textInput.Value())
		oldName := p.selectedProfile.Name

		// If name unchanged, just go back
		if newName == oldName {
			p.state = stateDetailMenu
			p.inputErr = ""
			return p, nil
		}

		// Validate name
		if err := config.ValidateProfileName(newName); err != nil {
			p.inputErr = err.Error()
			return p, nil
		}

		// Check for duplicates
		if config.ProfileExists(newName) {
			p.inputErr = "A profile with this name already exists"
			return p, nil
		}

		p.state = stateDetailMenu
		p.inputErr = ""
		return p, p.renameProfile(oldName, newName)
	}

	// Update text input
	var cmd tea.Cmd
	p.textInput, cmd = p.textInput.Update(msg)
	return p, cmd
}

func (p *ProfilesView) handleConfirmDeleteKeys(msg tea.KeyMsg) (*ProfilesView, tea.Cmd) {
	// Let the confirm dialog handle the keys
	var cmd tea.Cmd
	p.confirmDialog, cmd = p.confirmDialog.Update(msg)
	return p, cmd
}

// View renders the profiles view.
func (p *ProfilesView) View() string {
	if p.loading {
		return "Loading profiles..."
	}

	switch p.state {
	case stateNaming:
		return p.renderNamingView()
	case stateRenaming:
		return p.renderRenamingView()
	case stateConfirmDelete:
		return p.renderConfirmDeleteView()
	case stateDetail:
		// Legacy - redirect to detail menu
		return p.renderDetailMenuView()
	case stateDetailMenu:
		return p.renderDetailMenuView()
	case stateDetailPlugins:
		return p.renderDetailCategoryView("Plugins", p.pluginItems)
	case stateDetailMCP:
		return p.renderDetailCategoryView("MCP Servers", p.mcpItems)
	case stateDetailPermissions:
		return p.renderDetailCategoryView("Permissions", p.permissionItems)
	default:
		return p.renderListView()
	}
}

func (p *ProfilesView) renderNamingView() string {
	var lines []string

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.theme.Primary).
		MarginBottom(1)
	lines = append(lines, headerStyle.Render("New Profile"))
	lines = append(lines, "")

	lines = append(lines, "Enter a name for the new profile:")
	lines = append(lines, "")
	lines = append(lines, p.textInput.View())

	if p.inputErr != "" {
		errorStyle := lipgloss.NewStyle().Foreground(p.theme.Error)
		lines = append(lines, "")
		lines = append(lines, errorStyle.Render(p.inputErr))
	}

	lines = append(lines, "")
	helpStyle := p.theme.HelpStyle()
	lines = append(lines, helpStyle.Render("enter: create | esc: cancel"))

	return strings.Join(lines, "\n")
}

func (p *ProfilesView) renderRenamingView() string {
	var lines []string

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.theme.Primary).
		MarginBottom(1)
	lines = append(lines, headerStyle.Render("Rename Profile"))
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("Enter a new name for '%s':", p.selectedProfile.Name))
	lines = append(lines, "")
	lines = append(lines, p.textInput.View())

	if p.inputErr != "" {
		errorStyle := lipgloss.NewStyle().Foreground(p.theme.Error)
		lines = append(lines, "")
		lines = append(lines, errorStyle.Render(p.inputErr))
	}

	lines = append(lines, "")
	helpStyle := p.theme.HelpStyle()
	lines = append(lines, helpStyle.Render("enter: rename | esc: cancel"))

	return strings.Join(lines, "\n")
}

func (p *ProfilesView) renderConfirmDeleteView() string {
	var lines []string

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.theme.Primary).
		MarginBottom(1)
	lines = append(lines, headerStyle.Render("Delete Profile"))
	lines = append(lines, "")
	lines = append(lines, p.confirmDialog.View())

	return strings.Join(lines, "\n")
}

func (p *ProfilesView) renderListView() string {
	var lines []string

	// Section header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.theme.Primary).
		MarginBottom(1)
	lines = append(lines, headerStyle.Render("Profiles"))

	// Subtitle
	subtitleStyle := p.theme.MutedStyle()
	lines = append(lines, subtitleStyle.Render("Select a profile to view details or create a new one"))
	lines = append(lines, "")

	// Profile list
	if len(p.list.Items()) == 0 {
		lines = append(lines, p.theme.MutedStyle().Render("No profiles configured"))
		lines = append(lines, "")
		lines = append(lines, p.theme.MutedStyle().Render("Press 'n' to create a new profile"))
	} else {
		lines = append(lines, p.list.View())
	}

	// Help text
	lines = append(lines, "")
	helpStyle := p.theme.HelpStyle()
	lines = append(lines, helpStyle.Render("↑/↓/j/k: navigate | enter: view | n: new | d: delete | r: reload"))

	return strings.Join(lines, "\n")
}

func (p *ProfilesView) renderDetailView() string {
	// Legacy function - redirect to detail menu
	return p.renderDetailMenuView()
}

func (p *ProfilesView) renderDetailMenuView() string {
	var lines []string
	prof := p.selectedProfile

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.theme.Primary)
	lines = append(lines, headerStyle.Render("Profile: "+prof.Name))

	if prof.Description != "" {
		descStyle := p.theme.MutedStyle().Italic(true)
		lines = append(lines, descStyle.Render(prof.Description))
	}
	lines = append(lines, "")

	// Count plugin items by category
	globalPluginCount := p.countGlobalItems(p.pluginItems)
	enabledPluginCount := p.countEnabledNonGlobalItems(p.pluginItems)
	totalPluginCount := len(p.pluginItems)

	globalMCPCount := p.countGlobalItems(p.mcpItems)
	profileMCPCount := len(p.mcpItems) - globalMCPCount

	globalPermCount := p.countGlobalItems(p.permissionItems)
	profilePermCount := len(p.permissionItems) - globalPermCount

	// Menu items
	menuItems := []struct {
		title string
		count string
	}{
		{"Plugins", fmt.Sprintf("%d global, %d profile, %d available", globalPluginCount, enabledPluginCount, totalPluginCount)},
		{"MCP Servers", fmt.Sprintf("%d global, %d profile-specific", globalMCPCount, profileMCPCount)},
		{"Permissions", fmt.Sprintf("%d global, %d profile-specific", globalPermCount, profilePermCount)},
	}

	lines = append(lines, "Select a category to view:")
	lines = append(lines, "")

	for i, item := range menuItems {
		cursor := "  "
		if i == p.detailMenuCursor {
			cursor = "▸ "
		}

		titleStyle := lipgloss.NewStyle()
		if i == p.detailMenuCursor {
			titleStyle = titleStyle.Foreground(p.theme.Primary).Bold(true)
		}

		countStyle := p.theme.MutedStyle()

		line := cursor + titleStyle.Render(item.title) + " " + countStyle.Render("("+item.count+")")
		lines = append(lines, line)
	}

	// Help text
	lines = append(lines, "")
	helpStyle := p.theme.HelpStyle()
	lines = append(lines, helpStyle.Render("↑/↓/j/k: navigate | enter: view | r: rename | esc: back"))

	return strings.Join(lines, "\n")
}

func (p *ProfilesView) renderDetailCategoryView(categoryName string, items []detailItem) string {
	var lines []string
	prof := p.selectedProfile

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.theme.Primary)
	lines = append(lines, headerStyle.Render(categoryName+" - "+prof.Name))
	lines = append(lines, "")

	if len(items) == 0 {
		lines = append(lines, p.theme.MutedStyle().Render("(none)"))
	} else {
		// Calculate pagination
		start, end := p.paginator.GetSliceBounds(len(items))

		// Render items for current page
		for i := start; i < end; i++ {
			item := items[i]
			cursorIndex := i - start // Index within current page
			line := p.renderDetailItemWithCursor(item, cursorIndex == p.categoryCursor)
			lines = append(lines, line)
		}

		// Pad to maintain consistent height
		itemsOnPage := end - start
		for i := itemsOnPage; i < itemsPerPage; i++ {
			lines = append(lines, "")
		}

		// Page indicator
		if p.paginator.TotalPages > 1 {
			lines = append(lines, "")
			pageInfo := fmt.Sprintf("Page %d/%d", p.paginator.Page+1, p.paginator.TotalPages)
			paginatorLine := p.paginator.View() + "  " + p.theme.MutedStyle().Render(pageInfo)
			lines = append(lines, paginatorLine)
		}
	}

	// Help text
	lines = append(lines, "")
	helpStyle := p.theme.HelpStyle()
	lines = append(lines, helpStyle.Render("↑/↓: navigate | ←/→: page | space: toggle | esc: back"))

	return strings.Join(lines, "\n")
}

func (p *ProfilesView) renderDetailItem(item detailItem) string {
	return p.renderDetailItemWithCursor(item, false)
}

func (p *ProfilesView) renderDetailItemWithCursor(item detailItem, isSelected bool) string {
	globalStyle := p.theme.MutedStyle()
	dimmedStyle := p.theme.MutedStyle()

	// Cursor indicator
	cursor := "  "
	if isSelected {
		cursor = "> "
	}

	// Determine if item is enabled (for plugins, use the stored enabled field)
	isEnabled := item.enabled

	// Build checkbox and suffix based on item type
	var checkbox string
	var suffix string

	if item.isGlobal {
		// Global items: show checkmark and (global) suffix
		checkbox = "[x] "
		suffix = " (global)"
	} else if isEnabled {
		// Enabled in profile
		checkbox = "[x] "
	} else {
		// Not enabled
		checkbox = "[ ] "
	}

	var content string
	switch {
	case item.subType == "allow" || item.subType == "deny":
		// Permission item
		prefix := ""
		if item.subType == "allow" {
			prefix = "[+] "
		} else {
			prefix = "[-] "
		}
		display := item.name
		if len(display) > 45 {
			display = display[:42] + "..."
		}
		content = prefix + display
	case item.subType == "mcp":
		// MCP server
		content = item.name
	default:
		// Plugin or other item
		content = item.name
	}

	// Build the line
	line := cursor + checkbox + content + suffix

	// Apply styling based on selection and state
	if isSelected {
		selectedStyle := lipgloss.NewStyle().
			Foreground(p.theme.Primary).
			Bold(true)
		return selectedStyle.Render(line)
	}

	// Global items get muted style
	if item.isGlobal {
		return globalStyle.Render(line)
	}

	// Not enabled items get dimmed style
	if !isEnabled {
		return dimmedStyle.Render(line)
	}

	return line
}

// isItemEnabled checks if an item is enabled in the current profile
func (p *ProfilesView) isItemEnabled(item detailItem) bool {
	if p.selectedProfile == nil {
		return false
	}

	switch {
	case item.subType == "mcp":
		if enabled, exists := p.selectedProfile.MCPServers[item.name]; exists {
			return enabled
		}
		return item.enabled // Fall back to the item's enabled state
	case item.subType == "allow":
		for _, perm := range p.selectedProfile.Permissions.Allow {
			if perm == item.name {
				return true
			}
		}
		return false
	case item.subType == "deny":
		for _, perm := range p.selectedProfile.Permissions.Deny {
			if perm == item.name {
				return true
			}
		}
		return false
	default:
		// Plugin - check if in profile plugins
		for _, plugin := range p.selectedProfile.Plugins {
			if plugin == item.name {
				return true
			}
		}
		return false
	}
}

func (p *ProfilesView) countGlobalItems(items []detailItem) int {
	count := 0
	for _, item := range items {
		if item.isGlobal {
			count++
		}
	}
	return count
}

func (p *ProfilesView) countEnabledNonGlobalItems(items []detailItem) int {
	count := 0
	for _, item := range items {
		if !item.isGlobal && item.enabled {
			count++
		}
	}
	return count
}

func (p *ProfilesView) buildDetailItems() {
	if p.selectedProfile == nil {
		return
	}

	prof := p.selectedProfile

	// Build plugin items - show ALL marketplace plugins
	p.pluginItems = nil

	// Create lookup maps for quick access
	globalPluginMap := make(map[string]bool)
	for _, gp := range p.globalPlugins {
		if gp.Enabled {
			globalPluginMap[gp.Name] = true
		}
	}

	profilePluginMap := make(map[string]bool)
	for _, plugin := range prof.Plugins {
		profilePluginMap[plugin] = true
	}

	// Add ALL marketplace plugins
	for _, avail := range p.availablePlugins {
		isGlobal := globalPluginMap[avail.Name]
		isInProfile := profilePluginMap[avail.Name]

		p.pluginItems = append(p.pluginItems, detailItem{
			name:     avail.Name,
			isGlobal: isGlobal,
			enabled:  isGlobal || isInProfile,
		})
	}

	// Build MCP server items
	p.mcpItems = nil

	// Add global MCP servers first
	for _, gs := range p.globalMCPServers {
		if gs.Enabled {
			p.mcpItems = append(p.mcpItems, detailItem{
				name:     gs.Name,
				isGlobal: true,
				subType:  "mcp",
				enabled:  true,
			})
		}
	}

	// Add profile-specific MCP servers
	for server, enabled := range prof.MCPServers {
		// Check if not already global
		isGlobal := false
		for _, gs := range p.globalMCPServers {
			if gs.Name == server && gs.Enabled {
				isGlobal = true
				break
			}
		}
		if !isGlobal {
			p.mcpItems = append(p.mcpItems, detailItem{
				name:     server,
				isGlobal: false,
				subType:  "mcp",
				enabled:  enabled,
			})
		}
	}

	// Build permission items
	p.permissionItems = nil

	// Get global permissions
	globalAllow := p.globalPermissions.AllowList()
	globalDeny := p.globalPermissions.DenyList()

	// Add global allow permissions
	for _, perm := range globalAllow {
		p.permissionItems = append(p.permissionItems, detailItem{
			name:     perm,
			isGlobal: true,
			subType:  "allow",
			enabled:  true, // Global permissions are always enabled
		})
	}

	// Add profile-specific allow permissions
	for _, perm := range prof.Permissions.Allow {
		isGlobal := false
		for _, gp := range globalAllow {
			if gp == perm {
				isGlobal = true
				break
			}
		}
		if !isGlobal {
			p.permissionItems = append(p.permissionItems, detailItem{
				name:     perm,
				isGlobal: false,
				subType:  "allow",
				enabled:  true, // Profile permissions are enabled by being in the list
			})
		}
	}

	// Add global deny permissions
	for _, perm := range globalDeny {
		p.permissionItems = append(p.permissionItems, detailItem{
			name:     perm,
			isGlobal: true,
			subType:  "deny",
			enabled:  true, // Global permissions are always enabled
		})
	}

	// Add profile-specific deny permissions
	for _, perm := range prof.Permissions.Deny {
		isGlobal := false
		for _, gp := range globalDeny {
			if gp == perm {
				isGlobal = true
				break
			}
		}
		if !isGlobal {
			p.permissionItems = append(p.permissionItems, detailItem{
				name:     perm,
				isGlobal: false,
				subType:  "deny",
				enabled:  true, // Profile permissions are enabled by being in the list
			})
		}
	}
}

func (p *ProfilesView) renderPluginsSection(prof *config.Profile, itemStyle, globalStyle lipgloss.Style) []string {
	var lines []string

	// Create a map of profile plugins for quick lookup
	profilePlugins := make(map[string]bool)
	for _, plugin := range prof.Plugins {
		profilePlugins[plugin] = true
	}

	// Show global plugins first
	hasContent := false
	for _, gp := range p.globalPlugins {
		if gp.Enabled {
			indicator := globalStyle.Render(" (global)")
			lines = append(lines, itemStyle.Render("- "+gp.Name+indicator))
			hasContent = true
		}
	}

	// Show profile-specific plugins (ones not in global)
	for _, plugin := range prof.Plugins {
		// Check if this is not already shown as global
		isGlobal := false
		for _, gp := range p.globalPlugins {
			if gp.Name == plugin && gp.Enabled {
				isGlobal = true
				break
			}
		}
		if !isGlobal {
			lines = append(lines, itemStyle.Render("- "+plugin))
			hasContent = true
		}
	}

	if !hasContent {
		lines = append(lines, itemStyle.Render(p.theme.MutedStyle().Render("(none)")))
	}

	return lines
}

func (p *ProfilesView) renderMCPSection(prof *config.Profile, itemStyle, globalStyle lipgloss.Style) []string {
	var lines []string

	// Create a map of profile MCP servers for quick lookup
	profileServers := prof.MCPServers

	// Show global MCP servers first
	hasContent := false
	for _, gs := range p.globalMCPServers {
		if gs.Enabled {
			status := "enabled"
			statusStyle := p.theme.SuccessStyle()
			indicator := globalStyle.Render(" (global)")
			lines = append(lines, itemStyle.Render(
				fmt.Sprintf("- %s: %s%s", gs.Name, statusStyle.Render(status), indicator),
			))
			hasContent = true
		}
	}

	// Show profile-specific MCP servers (ones not in global)
	for server, enabled := range profileServers {
		// Check if this is not already shown as global
		isGlobal := false
		for _, gs := range p.globalMCPServers {
			if gs.Name == server && gs.Enabled {
				isGlobal = true
				break
			}
		}
		if !isGlobal {
			status := "disabled"
			statusStyle := p.theme.MutedStyle()
			if enabled {
				status = "enabled"
				statusStyle = p.theme.SuccessStyle()
			}
			lines = append(lines, itemStyle.Render(
				fmt.Sprintf("- %s: %s", server, statusStyle.Render(status)),
			))
			hasContent = true
		}
	}

	if !hasContent {
		lines = append(lines, itemStyle.Render(p.theme.MutedStyle().Render("(none)")))
	}

	return lines
}

func (p *ProfilesView) renderPermissionsSection(prof *config.Profile, itemStyle, globalStyle lipgloss.Style) []string {
	var lines []string

	// Get global permissions
	globalAllow := p.globalPermissions.AllowList()
	globalDeny := p.globalPermissions.DenyList()

	// Create maps for quick lookup
	profileAllow := make(map[string]bool)
	for _, perm := range prof.Permissions.Allow {
		profileAllow[perm] = true
	}
	profileDeny := make(map[string]bool)
	for _, perm := range prof.Permissions.Deny {
		profileDeny[perm] = true
	}

	hasContent := false

	// Show Allow section
	hasAllowContent := len(globalAllow) > 0 || len(prof.Permissions.Allow) > 0
	if hasAllowContent {
		lines = append(lines, itemStyle.Render("Allow:"))

		// Show global allow permissions first
		for _, perm := range globalAllow {
			indicator := globalStyle.Render(" (global)")
			display := perm
			if len(display) > 55 {
				display = display[:52] + "..."
			}
			lines = append(lines, itemStyle.Render("  - "+display+indicator))
			hasContent = true
		}

		// Show profile-specific allow permissions
		for _, perm := range prof.Permissions.Allow {
			// Check if this is not already shown as global
			isGlobal := false
			for _, gp := range globalAllow {
				if gp == perm {
					isGlobal = true
					break
				}
			}
			if !isGlobal {
				display := perm
				if len(display) > 60 {
					display = display[:57] + "..."
				}
				lines = append(lines, itemStyle.Render("  - "+display))
				hasContent = true
			}
		}
	}

	// Show Deny section
	hasDenyContent := len(globalDeny) > 0 || len(prof.Permissions.Deny) > 0
	if hasDenyContent {
		lines = append(lines, itemStyle.Render("Deny:"))

		// Show global deny permissions first
		for _, perm := range globalDeny {
			indicator := globalStyle.Render(" (global)")
			display := perm
			if len(display) > 55 {
				display = display[:52] + "..."
			}
			lines = append(lines, itemStyle.Render("  - "+display+indicator))
			hasContent = true
		}

		// Show profile-specific deny permissions
		for _, perm := range prof.Permissions.Deny {
			// Check if this is not already shown as global
			isGlobal := false
			for _, gp := range globalDeny {
				if gp == perm {
					isGlobal = true
					break
				}
			}
			if !isGlobal {
				display := perm
				if len(display) > 60 {
					display = display[:57] + "..."
				}
				lines = append(lines, itemStyle.Render("  - "+display))
				hasContent = true
			}
		}
	}

	if !hasContent {
		lines = append(lines, itemStyle.Render(p.theme.MutedStyle().Render("(none)")))
	}

	return lines
}

// SetSize sets the view dimensions.
func (p *ProfilesView) SetSize(width, height int) {
	p.width = width
	p.height = height
	// Adjust list height
	listHeight := height - 8
	if listHeight < 5 {
		listHeight = 5
	}
	p.list = components.NewList(p.list.Items(),
		components.WithListTheme(p.theme),
		components.WithListHeight(listHeight),
		components.WithListHelp(false),
	)
}

func (p *ProfilesView) buildListItems() {
	var items []components.ListItem

	for _, name := range p.profiles {
		desc := ""
		if name == "default" {
			desc = "Global settings applied to all projects"
		}

		items = append(items, components.ListItem{
			ID:          name,
			Title:       name,
			Description: desc,
			Enabled:     false,
			Disabled:    true, // Don't show checkbox
			Dimmed:      false,
		})
	}

	// Add "New Profile" option at the end
	items = append(items, components.ListItem{
		ID:          "_new_profile",
		Title:       "+ New Profile",
		Description: "Create a new profile",
		Enabled:     false,
		Disabled:    true,
		Dimmed:      true,
	})

	p.list.SetItems(items)
}

func (p *ProfilesView) loadProfile(name string) tea.Cmd {
	return func() tea.Msg {
		profile, err := config.LoadProfile(name)
		return profileLoadedMsg{profile: profile, err: err}
	}
}

func (p *ProfilesView) createNewProfile(name string) tea.Cmd {
	return func() tea.Msg {
		profile := config.NewProfile(name)
		err := profile.Save()
		return profileCreatedMsg{name: name, err: err}
	}
}

func (p *ProfilesView) deleteProfile(name string) tea.Cmd {
	return func() tea.Msg {
		err := config.DeleteProfile(name)
		return profileDeletedMsg{name: name, err: err}
	}
}

func (p *ProfilesView) renameProfile(oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		err := config.RenameProfile(oldName, newName)
		return profileRenamedMsg{oldName: oldName, newName: newName, err: err}
	}
}

func (p *ProfilesView) isFromDefaultProfile(itemType, name string) bool {
	if p.defaultProfile == nil || p.selectedProfile == nil {
		return false
	}
	if p.selectedProfile.Name == "default" {
		return false
	}

	switch itemType {
	case "plugin":
		for _, plugin := range p.defaultProfile.Plugins {
			if plugin == name {
				return true
			}
		}
	case "mcp":
		if _, exists := p.defaultProfile.MCPServers[name]; exists {
			return true
		}
	case "allow":
		for _, perm := range p.defaultProfile.Permissions.Allow {
			if perm == name {
				return true
			}
		}
	case "deny":
		for _, perm := range p.defaultProfile.Permissions.Deny {
			if perm == name {
				return true
			}
		}
	}

	return false
}

// Message types for profiles
type profilesLoadedMsg struct {
	profiles          []string
	defaultProfile    *config.Profile
	globalPlugins     []config.Plugin
	globalMCPServers  []config.MCPServer
	globalPermissions *config.PermissionsConfig
	availablePlugins  []config.AvailablePlugin
}

type profileLoadedMsg struct {
	profile *config.Profile
	err     error
}

type profileCreatedMsg struct {
	name string
	err  error
}

type profileDeletedMsg struct {
	name string
	err  error
}

type profileRenamedMsg struct {
	oldName string
	newName string
	err     error
}
