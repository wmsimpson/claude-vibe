package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// PermissionsConfig manages permissions from ~/.claude/settings.json
type PermissionsConfig struct {
	allowList    []string
	denyList     []string
	settingsData map[string]interface{}
}

// ClaudeSettingsPath returns the path to ~/.claude/settings.json
func ClaudeSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

// ClaudeDir returns the path to ~/.claude
func ClaudeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

// NewPermissionsConfig creates a new PermissionsConfig and loads permissions
func NewPermissionsConfig() *PermissionsConfig {
	pc := &PermissionsConfig{
		allowList:    []string{},
		denyList:     []string{},
		settingsData: make(map[string]interface{}),
	}
	pc.load()
	return pc
}

// load reads permissions from ~/.claude/settings.json
func (pc *PermissionsConfig) load() {
	settingsPath := ClaudeSettingsPath()
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, &pc.settingsData); err != nil {
		return
	}

	// Try nested permissions object first
	if perms, ok := pc.settingsData["permissions"].(map[string]interface{}); ok {
		if allow, ok := perms["allow"].([]interface{}); ok {
			pc.allowList = interfaceSliceToStringSlice(allow)
		}
		if deny, ok := perms["deny"].([]interface{}); ok {
			pc.denyList = interfaceSliceToStringSlice(deny)
		}
	}

	// Also check top-level "allow" (some configs use this format)
	if allow, ok := pc.settingsData["allow"].([]interface{}); ok {
		// Merge with existing allow list, avoiding duplicates
		topLevelAllow := interfaceSliceToStringSlice(allow)
		for _, perm := range topLevelAllow {
			if !contains(pc.allowList, perm) {
				pc.allowList = append(pc.allowList, perm)
			}
		}
	}
}

// AllowList returns the list of allowed permissions
func (pc *PermissionsConfig) AllowList() []string {
	return pc.allowList
}

// DenyList returns the list of denied permissions
func (pc *PermissionsConfig) DenyList() []string {
	return pc.denyList
}

// HasPermission checks if a permission is in the allow list
func (pc *PermissionsConfig) HasPermission(perm string) bool {
	return contains(pc.allowList, perm)
}

// AddPermission adds a permission to the allow list
func (pc *PermissionsConfig) AddPermission(perm string) error {
	if !contains(pc.allowList, perm) {
		pc.allowList = append(pc.allowList, perm)
	}
	return nil
}

// RemovePermission removes a permission from the allow list
func (pc *PermissionsConfig) RemovePermission(perm string) error {
	pc.allowList = remove(pc.allowList, perm)
	return nil
}

// Save writes the permissions to ~/.claude/settings.json
func (pc *PermissionsConfig) Save() error {
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

	// Update permissions
	perms, ok := settingsData["permissions"].(map[string]interface{})
	if !ok {
		perms = make(map[string]interface{})
	}

	perms["allow"] = pc.allowList
	perms["deny"] = pc.denyList
	settingsData["permissions"] = perms

	// Also update top-level allow for compatibility
	settingsData["allow"] = pc.allowList

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

// Helper functions

func interfaceSliceToStringSlice(slice []interface{}) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func remove(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
