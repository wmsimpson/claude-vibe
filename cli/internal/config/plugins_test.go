package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPluginConfig_ListInstalled_Empty(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPluginConfig()
	plugins := pc.ListInstalled()

	if len(plugins) != 0 {
		t.Errorf("expected empty list, got %d plugins", len(plugins))
	}
}

func TestPluginConfig_ListInstalled(t *testing.T) {
	tempDir := t.TempDir()

	// Create ~/.claude/plugins/installed_plugins.json
	pluginsDir := filepath.Join(tempDir, ".claude", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	installedPlugins := map[string]interface{}{
		"version": 2,
		"plugins": map[string]interface{}{
			"databricks-tools@claude-vibe": []map[string]interface{}{
				{
					"scope":       "user",
					"installPath": "/path/to/plugin",
					"version":     "1.0.0",
					"installedAt": "2024-01-01T00:00:00.000Z",
				},
			},
			"google-tools@claude-vibe": []map[string]interface{}{
				{
					"scope":       "project",
					"projectPath": "/some/project",
					"installPath": "/path/to/plugin2",
					"version":     "1.1.0",
					"installedAt": "2024-01-02T00:00:00.000Z",
				},
			},
		},
	}

	data, _ := json.MarshalIndent(installedPlugins, "", "  ")
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), data, 0644); err != nil {
		t.Fatalf("failed to write installed_plugins.json: %v", err)
	}

	// Also create settings.json with enabledPlugins
	claudeDir := filepath.Join(tempDir, ".claude")
	settingsJSON := map[string]interface{}{
		"enabledPlugins": map[string]bool{
			"databricks-tools@claude-vibe": true,
			"google-tools@claude-vibe":     false,
		},
	}
	data, _ = json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPluginConfig()
	plugins := pc.ListInstalled()

	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	// Find databricks plugin (name is extracted from full name, without @marketplace)
	var databricks *Plugin
	for i := range plugins {
		if plugins[i].Name == "databricks-tools" {
			databricks = &plugins[i]
			break
		}
	}

	if databricks == nil {
		t.Fatal("databricks-tools plugin not found")
	}
	if databricks.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", databricks.Version)
	}
	if databricks.Scope != "user" {
		t.Errorf("expected scope 'user', got %s", databricks.Scope)
	}
	if !databricks.Enabled {
		t.Error("expected plugin to be enabled")
	}
}

func TestPluginConfig_ListAvailable(t *testing.T) {
	tempDir := t.TempDir()

	// Create marketplace.json at ~/.vibe/marketplace/.claude-plugin/marketplace.json
	marketplaceDir := filepath.Join(tempDir, ".vibe", "marketplace", ".claude-plugin")
	os.MkdirAll(marketplaceDir, 0755)

	marketplace := map[string]interface{}{
		"name": "claude-vibe",
		"plugins": []map[string]interface{}{
			{
				"name":        "databricks-tools",
				"version":     "1.0.0",
				"description": "Databricks tools",
			},
			{
				"name":        "google-tools",
				"version":     "1.1.0",
				"description": "Google tools",
			},
		},
	}

	data, _ := json.MarshalIndent(marketplace, "", "  ")
	if err := os.WriteFile(filepath.Join(marketplaceDir, "marketplace.json"), data, 0644); err != nil {
		t.Fatalf("failed to write marketplace.json: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPluginConfig()
	available := pc.ListAvailable()

	if len(available) != 2 {
		t.Fatalf("expected 2 available plugins, got %d", len(available))
	}

	// Check first plugin
	found := false
	for _, p := range available {
		if p.Name == "databricks-tools" {
			found = true
			if p.Version != "1.0.0" {
				t.Errorf("expected version '1.0.0', got %s", p.Version)
			}
		}
	}
	if !found {
		t.Error("databricks-tools not found in available plugins")
	}
}

func TestPluginConfig_ListAvailable_Empty(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPluginConfig()
	available := pc.ListAvailable()

	if len(available) != 0 {
		t.Errorf("expected empty list, got %d plugins", len(available))
	}
}

func TestPluginConfig_SetEnabled(t *testing.T) {
	tempDir := t.TempDir()

	// Create installed plugins
	pluginsDir := filepath.Join(tempDir, ".claude", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	installedPlugins := map[string]interface{}{
		"version": 2,
		"plugins": map[string]interface{}{
			"test-plugin@test": []map[string]interface{}{
				{
					"scope":       "user",
					"installPath": "/path/to/plugin",
					"version":     "1.0.0",
				},
			},
		},
	}
	data, _ := json.MarshalIndent(installedPlugins, "", "  ")
	os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), data, 0644)

	// Create settings.json
	claudeDir := filepath.Join(tempDir, ".claude")
	settingsJSON := map[string]interface{}{
		"enabledPlugins": map[string]bool{
			"test-plugin@test": true,
		},
	}
	data, _ = json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPluginConfig()

	// Disable the plugin (use base name since that's what's stored now)
	err := pc.SetEnabled("test-plugin", false)
	if err != nil {
		t.Fatalf("SetEnabled() error: %v", err)
	}

	// Verify it's disabled
	plugins := pc.ListInstalled()
	for _, p := range plugins {
		if p.Name == "test-plugin" && p.Enabled {
			t.Error("expected plugin to be disabled")
		}
	}

	// Re-enable
	err = pc.SetEnabled("test-plugin", true)
	if err != nil {
		t.Fatalf("SetEnabled() error: %v", err)
	}

	// Verify it's enabled
	plugins = pc.ListInstalled()
	for _, p := range plugins {
		if p.Name == "test-plugin" && !p.Enabled {
			t.Error("expected plugin to be enabled")
		}
	}
}

func TestPluginConfig_SetEnabled_NotInstalled(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPluginConfig()

	err := pc.SetEnabled("nonexistent-plugin@test", false)
	if err == nil {
		t.Error("expected error for non-installed plugin")
	}
}

func TestPlugin_Fields(t *testing.T) {
	plugin := Plugin{
		Name:        "test-plugin@marketplace",
		Version:     "1.2.3",
		Scope:       "user",
		InstallPath: "/path/to/install",
		Enabled:     true,
	}

	if plugin.Name != "test-plugin@marketplace" {
		t.Errorf("expected name 'test-plugin@marketplace', got %s", plugin.Name)
	}
	if plugin.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %s", plugin.Version)
	}
	if plugin.Scope != "user" {
		t.Errorf("expected scope 'user', got %s", plugin.Scope)
	}
	if plugin.InstallPath != "/path/to/install" {
		t.Errorf("expected installPath '/path/to/install', got %s", plugin.InstallPath)
	}
	if !plugin.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestInstalledPluginsPath(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expected := filepath.Join(tempDir, ".claude", "plugins", "installed_plugins.json")
	actual := InstalledPluginsPath()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestMarketplacePath(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expected := filepath.Join(tempDir, ".vibe", "marketplace", ".claude-plugin", "marketplace.json")
	actual := MarketplacePath()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestAvailablePlugin_Fields(t *testing.T) {
	plugin := AvailablePlugin{
		Name:        "test-plugin",
		Version:     "2.0.0",
		Description: "A test plugin",
		Source:      "./plugins/test-plugin",
	}

	if plugin.Name != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got %s", plugin.Name)
	}
	if plugin.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %s", plugin.Version)
	}
	if plugin.Description != "A test plugin" {
		t.Errorf("expected description 'A test plugin', got %s", plugin.Description)
	}
}

func TestPluginConfig_ListInstalledUserScope(t *testing.T) {
	tempDir := t.TempDir()

	// Create ~/.claude/plugins/installed_plugins.json
	pluginsDir := filepath.Join(tempDir, ".claude", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	installedPlugins := map[string]interface{}{
		"version": 2,
		"plugins": map[string]interface{}{
			"databricks-tools@claude-vibe": []map[string]interface{}{
				{
					"scope":       "user",
					"installPath": "/path/to/plugin1",
					"version":     "1.0.0",
				},
			},
			"google-tools@claude-vibe": []map[string]interface{}{
				{
					"scope":       "project",
					"projectPath": "/some/project",
					"installPath": "/path/to/plugin2",
					"version":     "1.1.0",
				},
			},
			"fe-salesforce-tools@claude-vibe": []map[string]interface{}{
				{
					"scope":       "user",
					"installPath": "/path/to/plugin3",
					"version":     "1.2.0",
				},
			},
		},
	}

	data, _ := json.MarshalIndent(installedPlugins, "", "  ")
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), data, 0644); err != nil {
		t.Fatalf("failed to write installed_plugins.json: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPluginConfig()
	userPlugins := pc.ListInstalledUserScope()

	// Should only return user-scoped plugins (2 out of 3)
	if len(userPlugins) != 2 {
		t.Fatalf("expected 2 user-scoped plugins, got %d", len(userPlugins))
	}

	// Verify all returned plugins have user scope
	for _, p := range userPlugins {
		if p.Scope != "user" {
			t.Errorf("expected scope 'user', got %s for plugin %s", p.Scope, p.Name)
		}
	}

	// Verify google-tools (project scope) is not included
	for _, p := range userPlugins {
		if p.Name == "google-tools" {
			t.Error("project-scoped plugin should not be included in user scope list")
		}
	}
}

// TestProjectPluginConfig_Save_MergesExistingEntries tests that Save() updates
// existing entries with full names instead of creating duplicates with base names.
// This was a bug where toggling a plugin would create "plugin-name": false
// alongside the existing "plugin-name@claude-vibe": true entry.
func TestProjectPluginConfig_Save_MergesExistingEntries(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	os.MkdirAll(projectDir, 0755)

	// Create project .claude/settings.json with full name entry
	claudeDir := filepath.Join(projectDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	existingSettings := map[string]interface{}{
		"enabledPlugins": map[string]interface{}{
			"test-plugin@claude-vibe": true,
		},
	}
	data, _ := json.MarshalIndent(existingSettings, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	// Create installed plugins in HOME
	homeDir := filepath.Join(tempDir, "home")
	pluginsDir := filepath.Join(homeDir, ".claude", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	installedPlugins := map[string]interface{}{
		"version": 2,
		"plugins": map[string]interface{}{
			"test-plugin@claude-vibe": []map[string]interface{}{
				{
					"scope":       "project",
					"projectPath": projectDir,
					"installPath": "/path/to/plugin",
					"version":     "1.0.0",
				},
			},
		},
	}
	data, _ = json.MarshalIndent(installedPlugins, "", "  ")
	os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", originalHome)

	// Create project plugin config and disable the plugin
	ppc := NewProjectPluginConfig(projectDir)
	err := ppc.SetEnabled("test-plugin", false)
	if err != nil {
		t.Fatalf("SetEnabled() error: %v", err)
	}

	err = ppc.Save()
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Read back the settings file and verify no duplicate entries
	savedData, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read saved settings: %v", err)
	}

	var savedSettings map[string]interface{}
	if err := json.Unmarshal(savedData, &savedSettings); err != nil {
		t.Fatalf("failed to parse saved settings: %v", err)
	}

	enabledPlugins, ok := savedSettings["enabledPlugins"].(map[string]interface{})
	if !ok {
		t.Fatal("enabledPlugins not found in saved settings")
	}

	// Should have exactly one entry (the full name), not two
	if len(enabledPlugins) != 1 {
		t.Errorf("expected 1 entry in enabledPlugins, got %d: %v", len(enabledPlugins), enabledPlugins)
	}

	// The full name entry should be updated to false
	if val, exists := enabledPlugins["test-plugin@claude-vibe"]; exists {
		if val != false {
			t.Errorf("expected test-plugin@claude-vibe to be false, got %v", val)
		}
	} else {
		t.Error("test-plugin@claude-vibe entry not found")
	}

	// Should NOT have a duplicate entry with just the base name
	if _, exists := enabledPlugins["test-plugin"]; exists {
		t.Error("should not have duplicate entry with base name 'test-plugin'")
	}
}

// TestProjectPluginConfig_ListMerged_EnabledLookupBothFormats tests that
// ListMerged() checks both base name and full name when looking up enabled state.
// This was a bug where the enabled state from settings.json (with full name)
// wasn't being applied because the lookup used only the base name.
func TestProjectPluginConfig_ListMerged_EnabledLookupBothFormats(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	os.MkdirAll(projectDir, 0755)

	// Create project .claude/settings.json with full name entry set to false
	claudeDir := filepath.Join(projectDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	existingSettings := map[string]interface{}{
		"enabledPlugins": map[string]interface{}{
			"test-plugin@claude-vibe": false, // Using full name format
		},
	}
	data, _ := json.MarshalIndent(existingSettings, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	// Create installed plugins in HOME
	homeDir := filepath.Join(tempDir, "home")
	pluginsDir := filepath.Join(homeDir, ".claude", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	installedPlugins := map[string]interface{}{
		"version": 2,
		"plugins": map[string]interface{}{
			"test-plugin@claude-vibe": []map[string]interface{}{
				{
					"scope":       "project",
					"projectPath": projectDir,
					"installPath": "/path/to/plugin",
					"version":     "1.0.0",
				},
			},
		},
	}
	data, _ = json.MarshalIndent(installedPlugins, "", "  ")
	os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", originalHome)

	// Create project plugin config
	ppc := NewProjectPluginConfig(projectDir)
	merged := ppc.ListMerged()

	if len(merged) != 1 {
		t.Fatalf("expected 1 merged plugin, got %d", len(merged))
	}

	// The plugin should be disabled (reading the full name entry from settings.json)
	if merged[0].Plugin.Enabled {
		t.Error("expected plugin to be disabled based on settings.json entry with full name")
	}
}

// TestProjectPluginConfig_ListMerged_ProjectScopePlugins tests that
// project-scoped plugins are correctly identified and not marked as global.
func TestProjectPluginConfig_ListMerged_ProjectScopePlugins(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	os.MkdirAll(projectDir, 0755)

	// Create HOME directory structure
	homeDir := filepath.Join(tempDir, "home")
	pluginsDir := filepath.Join(homeDir, ".claude", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	// Create installed plugins - one user scope, one project scope
	installedPlugins := map[string]interface{}{
		"version": 2,
		"plugins": map[string]interface{}{
			"user-plugin@claude-vibe": []map[string]interface{}{
				{
					"scope":       "user",
					"installPath": "/path/to/user-plugin",
					"version":     "1.0.0",
				},
			},
			"project-plugin@claude-vibe": []map[string]interface{}{
				{
					"scope":       "project",
					"projectPath": projectDir,
					"installPath": "/path/to/project-plugin",
					"version":     "2.0.0",
				},
			},
		},
	}
	data, _ := json.MarshalIndent(installedPlugins, "", "  ")
	os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", originalHome)

	ppc := NewProjectPluginConfig(projectDir)
	merged := ppc.ListMerged()

	if len(merged) != 2 {
		t.Fatalf("expected 2 merged plugins, got %d", len(merged))
	}

	// Find and verify each plugin
	var userPlugin, projectPlugin *PluginWithScope
	for i := range merged {
		if merged[i].Plugin.Name == "user-plugin" {
			userPlugin = &merged[i]
		} else if merged[i].Plugin.Name == "project-plugin" {
			projectPlugin = &merged[i]
		}
	}

	if userPlugin == nil {
		t.Fatal("user-plugin not found in merged list")
	}
	if !userPlugin.IsGlobal {
		t.Error("user-scoped plugin should be marked as global")
	}

	if projectPlugin == nil {
		t.Fatal("project-plugin not found in merged list")
	}
	if projectPlugin.IsGlobal {
		t.Error("project-scoped plugin should NOT be marked as global")
	}
}
