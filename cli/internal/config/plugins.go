package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Plugin represents an installed plugin
type Plugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Scope       string `json:"scope"` // "user" or "project"
	InstallPath string `json:"installPath"`
	Enabled     bool   `json:"enabled"`
}

// AvailablePlugin represents a plugin available in the marketplace
type AvailablePlugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// PluginConfig manages plugin configurations
type PluginConfig struct {
	installed      map[string]*Plugin
	enabledPlugins map[string]bool
	settingsData   map[string]interface{}
}

// InstalledPluginsPath returns the path to ~/.claude/plugins/installed_plugins.json
func InstalledPluginsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "plugins", "installed_plugins.json")
}

// MarketplacePath returns the path to the marketplace.json file
func MarketplacePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vibe", "marketplace", ".claude-plugin", "marketplace.json")
}

// ProjectClaudeSettingsPath returns the path to ./.claude/settings.json (project scope)
func ProjectClaudeSettingsPath(projectDir string) string {
	return filepath.Join(projectDir, ".claude", "settings.json")
}

// ProjectPluginConfig manages plugin configurations for project scope
type ProjectPluginConfig struct {
	projectDir       string
	userPlugins      map[string]*Plugin // plugins from user scope
	projectPlugins   map[string]*Plugin // plugins added at project scope
	enabledPlugins   map[string]bool    // enabled state at project scope
	settingsData     map[string]interface{}
	userPluginConfig *PluginConfig // reference to user config
	dirty            bool
}

// NewProjectPluginConfig creates a new ProjectPluginConfig for the given project directory
func NewProjectPluginConfig(projectDir string) *ProjectPluginConfig {
	ppc := &ProjectPluginConfig{
		projectDir:       projectDir,
		userPlugins:      make(map[string]*Plugin),
		projectPlugins:   make(map[string]*Plugin),
		enabledPlugins:   make(map[string]bool),
		settingsData:     make(map[string]interface{}),
		userPluginConfig: NewPluginConfig(),
	}
	ppc.load()
	return ppc
}

// load reads plugin configurations from user and project scopes
func (ppc *ProjectPluginConfig) load() {
	// Load plugins and separate by scope
	for _, p := range ppc.userPluginConfig.ListInstalled() {
		plugin := p // copy
		if plugin.Scope == "user" {
			ppc.userPlugins[plugin.Name] = &plugin
		}
	}

	// Load project-scoped plugins for this specific project
	ppc.loadProjectPlugins()

	// Load project-scope settings (enabled/disabled overrides)
	ppc.loadProjectSettings()
}

// loadProjectPlugins loads plugins installed at project scope for this project
func (ppc *ProjectPluginConfig) loadProjectPlugins() {
	installedPath := InstalledPluginsPath()
	data, err := os.ReadFile(installedPath)
	if err != nil {
		return
	}

	var installedData struct {
		Version int                                  `json:"version"`
		Plugins map[string][]map[string]interface{} `json:"plugins"`
	}

	if err := json.Unmarshal(data, &installedData); err != nil {
		return
	}

	for fullName, installations := range installedData.Plugins {
		// Extract base name (before @)
		name := fullName
		if idx := strings.Index(fullName, "@"); idx > 0 {
			name = fullName[:idx]
		}

		// Look for project-scoped installation matching this project
		for _, install := range installations {
			scope, _ := install["scope"].(string)
			projectPath, _ := install["projectPath"].(string)

			if scope == "project" && projectPath == ppc.projectDir {
				plugin := &Plugin{
					Name:    name,
					Scope:   "project",
					Enabled: true,
				}
				if version, ok := install["version"].(string); ok {
					plugin.Version = version
				}
				if installPath, ok := install["installPath"].(string); ok {
					plugin.InstallPath = installPath
				}

				// Check enabled state
				if enabled, exists := ppc.enabledPlugins[name]; exists {
					plugin.Enabled = enabled
				}

				ppc.projectPlugins[name] = plugin
				break
			}
		}
	}
}

func (ppc *ProjectPluginConfig) loadProjectSettings() {
	settingsPath := ProjectClaudeSettingsPath(ppc.projectDir)
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, &ppc.settingsData); err != nil {
		return
	}

	// Parse enabledPlugins at project level
	// Normalize keys to base name (strip @marketplace suffix) for consistent lookup
	if enabled, ok := ppc.settingsData["enabledPlugins"].(map[string]interface{}); ok {
		for name, val := range enabled {
			if b, ok := val.(bool); ok {
				// Normalize to base name
				baseName := name
				if idx := strings.Index(name, "@"); idx > 0 {
					baseName = name[:idx]
				}
				ppc.enabledPlugins[baseName] = b
			}
		}
	}
}

// ListMerged returns all plugins (user + project) with merged state
// User plugins are marked as global and cannot be modified
func (ppc *ProjectPluginConfig) ListMerged() []PluginWithScope {
	var plugins []PluginWithScope

	// Track which plugins we've added (to avoid duplicates)
	added := make(map[string]bool)

	// Add user-scope plugins (marked as global)
	// But if user-scope plugin is disabled AND project-scope version exists, skip it
	// (user can install at project scope to override a globally disabled plugin)
	for _, p := range ppc.userPlugins {
		// If user plugin is disabled at user scope and project version exists,
		// skip the user version - show project version instead
		if !p.Enabled {
			if _, hasProject := ppc.projectPlugins[p.Name]; hasProject {
				continue // Will be added as project-scope below
			}
		}

		// Show user-scope plugin as global (read-only in project scope)
		// Use the user scope's enabled state - don't apply project overrides to global plugins
		pwc := PluginWithScope{
			Plugin:   *p,
			IsGlobal: true,
		}
		plugins = append(plugins, pwc)
		added[p.Name] = true
	}

	// Add project-scope plugins (if not already added from user scope)
	for name, p := range ppc.projectPlugins {
		if added[name] {
			continue
		}
		pwc := PluginWithScope{
			Plugin:   *p,
			IsGlobal: false,
		}
		// Check if there's an enabled override (keys are normalized to base name)
		if enabled, exists := ppc.enabledPlugins[name]; exists {
			pwc.Plugin.Enabled = enabled
		}
		plugins = append(plugins, pwc)
	}

	// Sort by name
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Plugin.Name < plugins[j].Plugin.Name
	})

	return plugins
}

// SetEnabled enables or disables a plugin at project scope
func (ppc *ProjectPluginConfig) SetEnabled(name string, enabled bool) error {
	// Can only modify plugins that exist
	if _, exists := ppc.userPlugins[name]; !exists {
		if _, exists := ppc.projectPlugins[name]; !exists {
			return errors.New("plugin not found: " + name)
		}
	}

	ppc.enabledPlugins[name] = enabled
	ppc.dirty = true
	return nil
}

// IsDirty returns true if there are unsaved changes
func (ppc *ProjectPluginConfig) IsDirty() bool {
	return ppc.dirty
}

// Save writes the project-scope plugin configuration
func (ppc *ProjectPluginConfig) Save() error {
	settingsPath := ProjectClaudeSettingsPath(ppc.projectDir)

	// Load existing settings or create new
	var settingsData map[string]interface{}
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		json.Unmarshal(data, &settingsData)
	}
	if settingsData == nil {
		settingsData = make(map[string]interface{})
	}

	// Get existing enabledPlugins or create new
	existingPlugins := make(map[string]interface{})
	if ep, ok := settingsData["enabledPlugins"].(map[string]interface{}); ok {
		existingPlugins = ep
	}

	// Update existing entries - find matching keys (with or without @marketplace suffix)
	for name, enabled := range ppc.enabledPlugins {
		// Look for existing entry with full name (name@claude-vibe)
		fullName := name + "@claude-vibe"
		if _, exists := existingPlugins[fullName]; exists {
			existingPlugins[fullName] = enabled
		} else if _, exists := existingPlugins[name]; exists {
			// Entry exists with base name
			existingPlugins[name] = enabled
		} else {
			// No existing entry - use full name format for consistency
			existingPlugins[fullName] = enabled
		}
	}

	settingsData["enabledPlugins"] = existingPlugins

	// Write back
	output, err := json.MarshalIndent(settingsData, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	ppc.dirty = false
	return os.WriteFile(settingsPath, output, 0644)
}

// ListAvailable returns all plugins available in the marketplace
func (ppc *ProjectPluginConfig) ListAvailable() []AvailablePlugin {
	return ppc.userPluginConfig.ListAvailable()
}

// PluginWithScope represents a plugin with scope information
type PluginWithScope struct {
	Plugin   Plugin
	IsGlobal bool // true if from user scope (global)
}

// NewPluginConfig creates a new PluginConfig and loads plugins from config files
func NewPluginConfig() *PluginConfig {
	pc := &PluginConfig{
		installed:      make(map[string]*Plugin),
		enabledPlugins: make(map[string]bool),
		settingsData:   make(map[string]interface{}),
	}
	pc.load()
	return pc
}

// load reads plugin configurations from all sources
func (pc *PluginConfig) load() {
	pc.loadEnabledPlugins()
	pc.loadInstalledPlugins()
}

func (pc *PluginConfig) loadEnabledPlugins() {
	settingsPath := ClaudeSettingsPath()
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, &pc.settingsData); err != nil {
		return
	}

	// Parse enabledPlugins
	if enabled, ok := pc.settingsData["enabledPlugins"].(map[string]interface{}); ok {
		for name, val := range enabled {
			if b, ok := val.(bool); ok {
				pc.enabledPlugins[name] = b
			}
		}
	}
}

func (pc *PluginConfig) loadInstalledPlugins() {
	installedPath := InstalledPluginsPath()
	data, err := os.ReadFile(installedPath)
	if err != nil {
		return
	}

	var installedData struct {
		Version int                                  `json:"version"`
		Plugins map[string][]map[string]interface{} `json:"plugins"`
	}

	if err := json.Unmarshal(data, &installedData); err != nil {
		return
	}

	for fullName, installations := range installedData.Plugins {
		if len(installations) == 0 {
			continue
		}

		// Prefer user-scoped installation over project-scoped
		// This ensures the user scope view shows user-installed plugins
		install := installations[0]
		for _, inst := range installations {
			if scope, ok := inst["scope"].(string); ok && scope == "user" {
				install = inst
				break
			}
		}

		// Extract base name (before @) for matching with marketplace
		name := fullName
		if idx := strings.Index(fullName, "@"); idx > 0 {
			name = fullName[:idx]
		}

		plugin := &Plugin{
			Name:    name,
			Enabled: true, // Default to enabled
		}

		if version, ok := install["version"].(string); ok {
			plugin.Version = version
		}
		if scope, ok := install["scope"].(string); ok {
			plugin.Scope = scope
		}
		if installPath, ok := install["installPath"].(string); ok {
			plugin.InstallPath = installPath
		}

		// Check if explicitly enabled/disabled in settings (check both full and base name)
		if enabled, exists := pc.enabledPlugins[fullName]; exists {
			plugin.Enabled = enabled
		} else if enabled, exists := pc.enabledPlugins[name]; exists {
			plugin.Enabled = enabled
		}

		pc.installed[name] = plugin
	}
}

// ListInstalled returns all installed plugins sorted by name
func (pc *PluginConfig) ListInstalled() []Plugin {
	var plugins []Plugin
	for _, p := range pc.installed {
		plugins = append(plugins, *p)
	}

	// Sort by name for consistent ordering
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	return plugins
}

// ListInstalledUserScope returns plugins installed at user scope (global)
func (pc *PluginConfig) ListInstalledUserScope() []Plugin {
	var plugins []Plugin
	for _, p := range pc.installed {
		if p.Scope == "user" {
			plugins = append(plugins, *p)
		}
	}

	// Sort by name for consistent ordering
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	return plugins
}

// ListAvailable returns all plugins available in the marketplace
func (pc *PluginConfig) ListAvailable() []AvailablePlugin {
	marketplacePath := MarketplacePath()
	data, err := os.ReadFile(marketplacePath)
	if err != nil {
		return nil
	}

	var marketplace struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name        string          `json:"name"`
			Version     string          `json:"version"`
			Description string          `json:"description"`
			Source      json.RawMessage `json:"source"`
		} `json:"plugins"`
	}

	if err := json.Unmarshal(data, &marketplace); err != nil {
		return nil
	}

	var plugins []AvailablePlugin
	for _, p := range marketplace.Plugins {
		// Handle source which can be string or object
		var source string
		if len(p.Source) > 0 {
			// Try to unmarshal as string first
			if err := json.Unmarshal(p.Source, &source); err != nil {
				// If that fails, it's an object - just use the plugin name
				source = p.Name
			}
		}
		plugins = append(plugins, AvailablePlugin{
			Name:        p.Name,
			Version:     p.Version,
			Description: p.Description,
			Source:      source,
		})
	}

	return plugins
}

// SetEnabled enables or disables a plugin
func (pc *PluginConfig) SetEnabled(name string, enabled bool) error {
	plugin, exists := pc.installed[name]
	if !exists {
		return errors.New("plugin not installed: " + name)
	}

	plugin.Enabled = enabled

	// Update both base name and any full name (with @marketplace suffix) entries
	// to ensure consistency since loadInstalledPlugins checks fullName first
	pc.enabledPlugins[name] = enabled

	// Also update any existing entries with marketplace suffixes
	for key := range pc.enabledPlugins {
		if strings.HasPrefix(key, name+"@") {
			pc.enabledPlugins[key] = enabled
		}
	}

	return pc.Save()
}

// Save writes the plugin configuration to settings.json
func (pc *PluginConfig) Save() error {
	settingsPath := ClaudeSettingsPath()

	// Load existing settings or create new
	var settingsData map[string]interface{}
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		json.Unmarshal(data, &settingsData)
	}
	if settingsData == nil {
		settingsData = make(map[string]interface{})
	}

	// Build enabledPlugins map
	enabledPlugins := make(map[string]bool)
	for name, enabled := range pc.enabledPlugins {
		enabledPlugins[name] = enabled
	}

	settingsData["enabledPlugins"] = enabledPlugins

	// Write back
	output, err := json.MarshalIndent(settingsData, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(settingsPath, output, 0644)
}
