package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wmsimpson/claude-vibe/cli/internal/util"
	"gopkg.in/yaml.v3"
)

// Profile represents a named collection of plugins, MCP servers, and permissions
type Profile struct {
	Version            int               `yaml:"version"`
	Name               string            `yaml:"name"`
	Description        string            `yaml:"description,omitempty"`
	Email              string            `yaml:"email,omitempty"`
	GitEmail           string            `yaml:"git_email,omitempty"`
	EnvFile            string            `yaml:"env_file,omitempty"`
	Integrations       map[string]bool   `yaml:"integrations,omitempty"`
	Plugins            []string          `yaml:"plugins,omitempty"`
	MCPServers         map[string]bool   `yaml:"mcp_servers,omitempty"`
	Permissions        PermissionSet     `yaml:"permissions,omitempty"`
	DatabricksProfile  string            `yaml:"databricks_profile,omitempty"`
	EnvOverrides       map[string]string `yaml:"env_overrides,omitempty"`
}

// PermissionSet represents allow and deny permission lists
type PermissionSet struct {
	Allow []string `yaml:"allow,omitempty"`
	Deny  []string `yaml:"deny,omitempty"`
}

// ProfilesDir returns the path to ~/.vibe/profiles
func ProfilesDir() string {
	return filepath.Join(VibeDir(), "profiles")
}

// ListProfiles returns the names of all available profiles
func ListProfiles() []string {
	profilesDir := ProfilesDir()
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return []string{}
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			// Remove extension to get profile name
			profileName := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
			profiles = append(profiles, profileName)
		}
	}

	return profiles
}

// LoadProfile loads a profile by name from ~/.vibe/profiles/<name>.yaml
func LoadProfile(name string) (*Profile, error) {
	profilePath := filepath.Join(ProfilesDir(), name+".yaml")

	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, err
	}

	profile := &Profile{}
	if err := yaml.Unmarshal(data, profile); err != nil {
		return nil, err
	}

	// Ensure slices are never nil (YAML unmarshal leaves them nil if not present)
	if profile.Plugins == nil {
		profile.Plugins = []string{}
	}
	if profile.MCPServers == nil {
		profile.MCPServers = make(map[string]bool)
	}
	if profile.Permissions.Allow == nil {
		profile.Permissions.Allow = []string{}
	}
	if profile.Permissions.Deny == nil {
		profile.Permissions.Deny = []string{}
	}
	if profile.Integrations == nil {
		profile.Integrations = make(map[string]bool)
	}

	return profile, nil
}

// NewProfile creates a new empty profile with the given name
func NewProfile(name string) *Profile {
	return &Profile{
		Version:       1,
		Name:          name,
		Plugins:       []string{},
		MCPServers:    make(map[string]bool),
		EnvOverrides:  make(map[string]string),
		Permissions: PermissionSet{
			Allow: []string{},
			Deny:  []string{},
		},
	}
}

// Save writes the profile to ~/.vibe/profiles/<name>.yaml
func (p *Profile) Save() error {
	profilesDir := ProfilesDir()
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return err
	}

	profilePath := filepath.Join(profilesDir, p.Name+".yaml")

	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}

	return os.WriteFile(profilePath, data, 0644)
}

// DeleteProfile deletes a profile by name from ~/.vibe/profiles/<name>.yaml
func DeleteProfile(name string) error {
	profilePath := filepath.Join(ProfilesDir(), name+".yaml")
	return os.Remove(profilePath)
}

// RenameProfile renames a profile by changing both the file name and the internal name
func RenameProfile(oldName, newName string) error {
	// Load the profile
	profile, err := LoadProfile(oldName)
	if err != nil {
		return err
	}

	// Delete the old file
	if err := DeleteProfile(oldName); err != nil {
		return err
	}

	// Update the name and save
	profile.Name = newName
	return profile.Save()
}

// ProfileExists checks if a profile with the given name exists
func ProfileExists(name string) bool {
	profilePath := filepath.Join(ProfilesDir(), name+".yaml")
	_, err := os.Stat(profilePath)
	return err == nil
}

// ValidateProfileName checks if a profile name is valid (no empty, no special chars except dash/underscore)
func ValidateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Check for invalid characters
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("profile name can only contain letters, numbers, dashes, and underscores")
		}
	}

	return nil
}

// ApplyResult contains information about what was applied
type ApplyResult struct {
	PluginsInstalled   int
	MCPServersEnabled  int
	PermissionsApplied int
	PluginErrors       []string
	IsHomeDirectory    bool     // true if the profile was applied to the home directory
	Warnings           []string // warnings about the application
}

// ErrApplyToHomeDirectory is returned when attempting to apply a profile to the home directory
// without explicitly confirming the operation
var ErrApplyToHomeDirectory = fmt.Errorf("cannot apply profile to home directory: this would modify user-scoped settings")

// IsHomeDirectoryError checks if an error is due to home directory safety check
func IsHomeDirectoryError(err error) bool {
	return err == ErrApplyToHomeDirectory
}

// ApplyOptions contains options for applying a profile
type ApplyOptions struct {
	// Force allows applying to home directory (dangerous - use with caution)
	Force bool
}

// Apply applies the profile settings to a project directory
// It creates/updates:
// - <dir>/.claude/settings.json (permissions)
// - <dir>/.claude.json (MCP servers) - MERGES with existing content
// - Installs plugins at project scope using `claude plugin install`
//
// If dir is the home directory, returns ErrApplyToHomeDirectory unless Force is set.
// Even with Force, the home directory application will merge settings carefully
// to avoid destroying critical application state.
func (p *Profile) Apply(dir string) (*ApplyResult, error) {
	return p.ApplyWithOptions(dir, ApplyOptions{})
}

// ApplyWithOptions applies the profile with the given options
func (p *Profile) ApplyWithOptions(dir string, opts ApplyOptions) (*ApplyResult, error) {
	result := &ApplyResult{}

	// Check if applying to home directory
	isHome := util.IsHomeDirectory(dir)
	result.IsHomeDirectory = isHome

	if isHome {
		if !opts.Force {
			return nil, ErrApplyToHomeDirectory
		}
		// Add warnings about what we're doing
		result.Warnings = append(result.Warnings,
			"Applying profile to home directory (~) - this modifies user-scoped settings",
			"MCP servers will be merged into ~/.claude.json (preserving other settings)",
			"Plugins will be installed at user scope, not project scope",
		)
	}

	// Create .claude directory
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return nil, err
	}

	// Write settings.json with permissions and plugins
	if err := p.writeProjectSettings(claudeDir); err != nil {
		return nil, err
	}
	result.PermissionsApplied = len(p.Permissions.Allow) + len(p.Permissions.Deny)

	// Write .claude.json with MCP servers (using merge to preserve existing data)
	if err := p.writeProjectClaudeJSON(dir, isHome); err != nil {
		return nil, err
	}
	// Count enabled MCP servers
	for _, enabled := range p.MCPServers {
		if enabled {
			result.MCPServersEnabled++
		}
	}

	// Install plugins - use user scope if in home directory
	var pluginErrors []string
	if isHome {
		pluginErrors = p.installPluginsUserScope()
	} else {
		pluginErrors = p.installPlugins(dir)
	}
	result.PluginsInstalled = len(p.Plugins) - len(pluginErrors)
	result.PluginErrors = pluginErrors

	// Sync Databricks profile to ~/.mcp.json and ~/.vibe/env
	if p.DatabricksProfile != "" {
		if err := p.syncDatabricksProfile(); err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Failed to sync Databricks profile: %v", err))
		}
	}

	// Sync env overrides to ~/.vibe/env
	if len(p.EnvOverrides) > 0 {
		if err := p.syncEnvOverrides(); err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Failed to sync env overrides: %v", err))
		}
	}

	return result, nil
}

// installPlugins installs the profile's plugins at project scope
// It returns a list of error messages for any plugins that failed to install
func (p *Profile) installPlugins(dir string) []string {
	var errors []string

	for _, plugin := range p.Plugins {
		if err := installPluginAtProjectScope(plugin, dir); err != nil {
			errors = append(errors, plugin+": "+err.Error())
		}
	}

	return errors
}

// installPluginsUserScope installs the profile's plugins at user scope
// This is used when applying a profile to the home directory
func (p *Profile) installPluginsUserScope() []string {
	var errors []string

	for _, plugin := range p.Plugins {
		if err := installPluginAtUserScope(plugin); err != nil {
			errors = append(errors, plugin+": "+err.Error())
		}
	}

	return errors
}

// installPluginAtProjectScope runs `claude plugin install <name>@claude-vibe --scope project`
// This is a variable so it can be mocked in tests
var installPluginAtProjectScope = func(pluginName, dir string) error {
	return defaultInstallPlugin(pluginName, dir, "project")
}

// installPluginAtUserScope runs `claude plugin install <name>@claude-vibe --scope user`
// This is a variable so it can be mocked in tests
var installPluginAtUserScope = func(pluginName string) error {
	return defaultInstallPlugin(pluginName, "", "user")
}

// defaultInstallPlugin is the real implementation of plugin installation
func defaultInstallPlugin(pluginName, dir, scope string) error {
	cmd := exec.Command("claude", "plugin", "install", pluginName+"@claude-vibe", "--scope", scope)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(output))
	}
	return nil
}

// syncDatabricksProfile updates ~/.mcp.json and ~/.vibe/env with the Databricks profile
func (p *Profile) syncDatabricksProfile() error {
	if p.DatabricksProfile == "" {
		return nil
	}

	// Update ~/.mcp.json
	home, _ := os.UserHomeDir()
	mcpPath := filepath.Join(home, ".mcp.json")

	var mcpData map[string]interface{}
	if data, err := os.ReadFile(mcpPath); err == nil {
		json.Unmarshal(data, &mcpData)
	}
	if mcpData == nil {
		mcpData = make(map[string]interface{})
	}

	servers, ok := mcpData["mcpServers"].(map[string]interface{})
	if !ok {
		servers = make(map[string]interface{})
		mcpData["mcpServers"] = servers
	}

	dbServer, ok := servers["databricks"].(map[string]interface{})
	if ok {
		env, ok := dbServer["env"].(map[string]interface{})
		if !ok {
			env = make(map[string]interface{})
			dbServer["env"] = env
		}
		env["DATABRICKS_CONFIG_PROFILE"] = p.DatabricksProfile

		data, err := json.MarshalIndent(mcpData, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(mcpPath, data, 0644); err != nil {
			return err
		}
	}

	// Update DATABRICKS_CONFIG_PROFILE in ~/.vibe/env
	return updateEnvVar(filepath.Join(home, ".vibe", "env"),
		"DATABRICKS_CONFIG_PROFILE", p.DatabricksProfile)
}

// syncEnvOverrides writes env override values to ~/.vibe/env
func (p *Profile) syncEnvOverrides() error {
	home, _ := os.UserHomeDir()
	envPath := filepath.Join(home, ".vibe", "env")

	for key, value := range p.EnvOverrides {
		if err := updateEnvVar(envPath, key, value); err != nil {
			return err
		}
	}
	return nil
}

// updateEnvVar updates or adds an environment variable in a shell env file
func updateEnvVar(envPath, key, value string) error {
	data, err := os.ReadFile(envPath)
	if err != nil {
		// File doesn't exist, create it
		content := fmt.Sprintf("export %s=%q\n", key, value)
		return os.WriteFile(envPath, []byte(content), 0644)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	found := false
	prefix := "export " + key + "="

	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = fmt.Sprintf("export %s=%q", key, value)
			found = true
			break
		}
	}

	if !found {
		// Append before the last line if it's empty, otherwise add new line
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = append(lines[:len(lines)-1], fmt.Sprintf("export %s=%q", key, value), "")
		} else {
			lines = append(lines, fmt.Sprintf("export %s=%q", key, value))
		}
	}

	return os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644)
}

func (p *Profile) writeProjectSettings(claudeDir string) error {
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Load existing settings or create new
	var settings map[string]interface{}
	existingData, err := os.ReadFile(settingsPath)
	if err == nil {
		json.Unmarshal(existingData, &settings)
	}
	if settings == nil {
		settings = make(map[string]interface{})
	}

	// Ensure allow and deny are never nil (JSON marshals nil slices as null, but we want [])
	allow := p.Permissions.Allow
	if allow == nil {
		allow = []string{}
	}
	deny := p.Permissions.Deny
	if deny == nil {
		deny = []string{}
	}

	settings["permissions"] = map[string]interface{}{
		"allow": allow,
		"deny":  deny,
	}

	// Add enabled plugins - merge with existing entries
	if len(p.Plugins) > 0 {
		// Get existing enabledPlugins or create new
		existingPlugins := make(map[string]interface{})
		if ep, ok := settings["enabledPlugins"].(map[string]interface{}); ok {
			existingPlugins = ep
		}

		// Update entries - find matching keys (with or without @marketplace suffix)
		for _, plugin := range p.Plugins {
			fullName := plugin + "@claude-vibe"
			if _, exists := existingPlugins[fullName]; exists {
				existingPlugins[fullName] = true
			} else if _, exists := existingPlugins[plugin]; exists {
				existingPlugins[plugin] = true
			} else {
				// No existing entry - use full name format for consistency
				existingPlugins[fullName] = true
			}
		}

		settings["enabledPlugins"] = existingPlugins
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0644)
}

// SyncEnvFile copies the profile's env file to ~/.vibe/env.
// Uses EnvFile field if set, otherwise falls back to <name>.env.
// No-op if neither exists.
func (p *Profile) SyncEnvFile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	profilesDir := filepath.Join(home, ".vibe", "profiles")
	targetPath := filepath.Join(home, ".vibe", "env")

	var sourcePath string
	if p.EnvFile != "" {
		sourcePath = filepath.Join(profilesDir, p.EnvFile)
	} else {
		candidate := filepath.Join(profilesDir, p.Name+".env")
		if _, err := os.Stat(candidate); err == nil {
			sourcePath = candidate
		}
	}

	if sourcePath == "" {
		return nil
	}

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading profile env file %s: %w", sourcePath, err)
	}

	return os.WriteFile(targetPath, data, 0644)
}

// writeProjectClaudeJSON writes MCP servers to .claude.json
// If isHomeDir is true, it carefully MERGES the mcpServers key while preserving
// all other keys in the file (like OAuth, userID, feature flags, etc.)
// SyncDatabricksProfile exports the Databricks profile sync for use by CLI commands.
func (p *Profile) SyncDatabricksProfile() error {
	return p.syncDatabricksProfile()
}

// SyncEnvOverrides exports the env override sync for use by CLI commands.
func (p *Profile) SyncEnvOverrides() error {
	return p.syncEnvOverrides()
}

func (p *Profile) writeProjectClaudeJSON(dir string, isHomeDir bool) error {
	claudeJSONPath := filepath.Join(dir, ".claude.json")

	// Build MCP servers config from profile
	mcpServers := make(map[string]interface{})

	// Load global MCP config to get server details
	mc := NewMCPConfig()
	globalServers := mc.ListServers()

	for serverName, enabled := range p.MCPServers {
		if !enabled {
			continue
		}

		// Find the server config in global config
		for _, gs := range globalServers {
			if gs.Name == serverName {
				serverCfg := map[string]interface{}{
					"command": gs.Command,
					"args":    gs.Args,
				}
				if len(gs.Env) > 0 {
					serverCfg["env"] = gs.Env
				}
				mcpServers[serverName] = serverCfg
				break
			}
		}
	}

	// Load existing file content - CRITICAL for home directory to preserve state
	var claudeJSON map[string]interface{}
	existingData, err := os.ReadFile(claudeJSONPath)
	if err == nil {
		if jsonErr := json.Unmarshal(existingData, &claudeJSON); jsonErr != nil {
			// File exists but is invalid JSON - only proceed if not home directory
			if isHomeDir {
				return fmt.Errorf("cannot modify ~/.claude.json: file contains invalid JSON")
			}
			claudeJSON = make(map[string]interface{})
		}
	} else {
		claudeJSON = make(map[string]interface{})
	}

	// For home directory, we MERGE the mcpServers rather than replace
	if isHomeDir && len(mcpServers) > 0 {
		// Get existing mcpServers
		existingMCPServers := make(map[string]interface{})
		if existing, ok := claudeJSON["mcpServers"].(map[string]interface{}); ok {
			existingMCPServers = existing
		}

		// Merge new servers into existing (new servers override existing with same name)
		for name, cfg := range mcpServers {
			existingMCPServers[name] = cfg
		}

		claudeJSON["mcpServers"] = existingMCPServers
	} else if len(mcpServers) > 0 {
		// For non-home directories, we can replace mcpServers entirely
		// but still preserve other keys if they exist
		claudeJSON["mcpServers"] = mcpServers
	}

	// For non-home directories, always create the file for consistency
	// (even if empty - this ensures .claude.json exists for project scope)
	// For home directories, only write if we actually have MCP servers to add
	if isHomeDir && len(mcpServers) == 0 {
		// Don't modify ~/.claude.json if we have no MCP servers to add
		return nil
	}

	data, err := json.MarshalIndent(claudeJSON, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(claudeJSONPath, data, 0644)
}
