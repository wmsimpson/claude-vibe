package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestListProfiles_Empty(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profiles := ListProfiles()

	if len(profiles) != 0 {
		t.Errorf("expected empty list, got %d profiles", len(profiles))
	}
}

func TestListProfiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create profiles directory with some profiles
	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)

	// Create profile files
	profile1 := `version: 1
name: ml-project
description: "ML development profile"
plugins:
  - databricks-tools
mcp_servers:
  chrome-devtools: true
permissions:
  allow:
    - "Bash(python:*)"
`
	profile2 := `version: 1
name: web-dev
description: "Web development profile"
plugins:
  - google-tools
`

	os.WriteFile(filepath.Join(profilesDir, "ml-project.yaml"), []byte(profile1), 0644)
	os.WriteFile(filepath.Join(profilesDir, "web-dev.yaml"), []byte(profile2), 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profiles := ListProfiles()

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	// Check that both profiles are found (order may vary)
	names := make(map[string]bool)
	for _, name := range profiles {
		names[name] = true
	}

	if !names["ml-project"] || !names["web-dev"] {
		t.Errorf("expected 'ml-project' and 'web-dev' profiles, got %v", profiles)
	}
}

func TestLoadProfile(t *testing.T) {
	tempDir := t.TempDir()

	// Create profile
	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)

	profileContent := `version: 1
name: test-profile
description: "A test profile"
plugins:
  - databricks-tools
  - google-tools
mcp_servers:
  chrome-devtools: true
  slack: false
permissions:
  allow:
    - "Bash(python:*)"
    - "Read(~/projects/**)"
  deny:
    - "Write(/etc/**)"
`
	os.WriteFile(filepath.Join(profilesDir, "test-profile.yaml"), []byte(profileContent), 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profile, err := LoadProfile("test-profile")
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}

	if profile.Version != 1 {
		t.Errorf("expected version 1, got %d", profile.Version)
	}
	if profile.Name != "test-profile" {
		t.Errorf("expected name 'test-profile', got %s", profile.Name)
	}
	if profile.Description != "A test profile" {
		t.Errorf("expected description 'A test profile', got %s", profile.Description)
	}
	if len(profile.Plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(profile.Plugins))
	}
	if profile.MCPServers["chrome-devtools"] != true {
		t.Error("expected chrome-devtools to be enabled")
	}
	if profile.MCPServers["slack"] != false {
		t.Error("expected slack to be disabled")
	}
	if len(profile.Permissions.Allow) != 2 {
		t.Errorf("expected 2 allow permissions, got %d", len(profile.Permissions.Allow))
	}
	if len(profile.Permissions.Deny) != 1 {
		t.Errorf("expected 1 deny permission, got %d", len(profile.Permissions.Deny))
	}
}

func TestLoadProfile_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	_, err := LoadProfile("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
}

func TestLoadProfile_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()

	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)

	os.WriteFile(filepath.Join(profilesDir, "invalid.yaml"), []byte("invalid: yaml: content:"), 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	_, err := LoadProfile("invalid")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestProfile_Save(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profile := &Profile{
		Version:     1,
		Name:        "new-profile",
		Description: "A new profile",
		Plugins:     []string{"databricks-tools"},
		MCPServers: map[string]bool{
			"chrome-devtools": true,
		},
		Permissions: PermissionSet{
			Allow: []string{"Bash"},
			Deny:  []string{},
		},
	}

	err := profile.Save()
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify the file was created
	profilePath := filepath.Join(tempDir, ".vibe", "profiles", "new-profile.yaml")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		t.Error("profile file was not created")
	}

	// Reload and verify
	loaded, err := LoadProfile("new-profile")
	if err != nil {
		t.Fatalf("LoadProfile() after Save() error: %v", err)
	}

	if loaded.Name != "new-profile" {
		t.Errorf("expected name 'new-profile', got %s", loaded.Name)
	}
	if loaded.Description != "A new profile" {
		t.Errorf("expected description 'A new profile', got %s", loaded.Description)
	}
}

func TestProfile_SaveCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profile := &Profile{
		Version: 1,
		Name:    "test",
	}

	err := profile.Save()
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify the profiles directory was created
	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	info, err := os.Stat(profilesDir)
	if os.IsNotExist(err) {
		t.Error("profiles directory was not created")
	}
	if !info.IsDir() {
		t.Error("profiles should be a directory")
	}
}

func TestProfile_Apply(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "my-project")
	os.MkdirAll(projectDir, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock the plugin installer to avoid needing claude CLI
	originalInstaller := installPluginAtProjectScope
	var installedPlugins []string
	installPluginAtProjectScope = func(pluginName, dir string) error {
		installedPlugins = append(installedPlugins, pluginName)
		return nil
	}
	defer func() { installPluginAtProjectScope = originalInstaller }()

	profile := &Profile{
		Version:     1,
		Name:        "test-profile",
		Description: "Test profile",
		Plugins:     []string{"databricks-tools", "google-tools"},
		MCPServers: map[string]bool{
			"chrome-devtools": true,
			"slack":           false,
		},
		Permissions: PermissionSet{
			Allow: []string{"Bash", "Skill(test)"},
			Deny:  []string{},
		},
	}

	result, err := profile.Apply(projectDir)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	// Verify .claude directory was created
	claudeDir := filepath.Join(projectDir, ".claude")
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		t.Error(".claude directory was not created")
	}

	// Verify settings.json was created
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("settings.json was not created")
	}

	// Verify .claude.json was created in project root
	claudeJSONPath := filepath.Join(projectDir, ".claude.json")
	if _, err := os.Stat(claudeJSONPath); os.IsNotExist(err) {
		t.Error(".claude.json was not created")
	}

	// Verify result
	if result.PluginsInstalled != 2 {
		t.Errorf("expected 2 plugins installed, got %d", result.PluginsInstalled)
	}
	if result.MCPServersEnabled != 1 {
		t.Errorf("expected 1 MCP server enabled, got %d", result.MCPServersEnabled)
	}
	if result.PermissionsApplied != 2 {
		t.Errorf("expected 2 permissions, got %d", result.PermissionsApplied)
	}

	// Verify plugins were installed
	if len(installedPlugins) != 2 {
		t.Errorf("expected 2 plugins to be installed, got %d", len(installedPlugins))
	}
}

func TestProfile_ApplyCreatesClaudeDirectory(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "new-project")
	os.MkdirAll(projectDir, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock the plugin installer
	originalInstaller := installPluginAtProjectScope
	installPluginAtProjectScope = func(pluginName, dir string) error { return nil }
	defer func() { installPluginAtProjectScope = originalInstaller }()

	profile := &Profile{
		Version: 1,
		Name:    "minimal",
		Plugins: []string{},
	}

	_, err := profile.Apply(projectDir)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	claudeDir := filepath.Join(projectDir, ".claude")
	info, err := os.Stat(claudeDir)
	if os.IsNotExist(err) {
		t.Error(".claude directory was not created")
	}
	if !info.IsDir() {
		t.Error(".claude should be a directory")
	}
}

func TestProfile_Fields(t *testing.T) {
	profile := Profile{
		Version:     1,
		Name:        "test",
		Description: "Test description",
		Plugins:     []string{"plugin1", "plugin2"},
		MCPServers:  map[string]bool{"server1": true},
		Permissions: PermissionSet{
			Allow: []string{"perm1"},
			Deny:  []string{"perm2"},
		},
	}

	if profile.Version != 1 {
		t.Errorf("expected version 1, got %d", profile.Version)
	}
	if profile.Name != "test" {
		t.Errorf("expected name 'test', got %s", profile.Name)
	}
	if len(profile.Plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(profile.Plugins))
	}
	if len(profile.MCPServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(profile.MCPServers))
	}
	if len(profile.Permissions.Allow) != 1 {
		t.Errorf("expected 1 allow permission, got %d", len(profile.Permissions.Allow))
	}
	if len(profile.Permissions.Deny) != 1 {
		t.Errorf("expected 1 deny permission, got %d", len(profile.Permissions.Deny))
	}
}

func TestPermissionSet_Fields(t *testing.T) {
	ps := PermissionSet{
		Allow: []string{"Bash", "Read"},
		Deny:  []string{"Write"},
	}

	if len(ps.Allow) != 2 {
		t.Errorf("expected 2 allow permissions, got %d", len(ps.Allow))
	}
	if len(ps.Deny) != 1 {
		t.Errorf("expected 1 deny permission, got %d", len(ps.Deny))
	}
}

func TestProfilesDir(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expected := filepath.Join(tempDir, ".vibe", "profiles")
	actual := ProfilesDir()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestNewProfile(t *testing.T) {
	profile := NewProfile("my-profile")

	if profile.Version != 1 {
		t.Errorf("expected version 1, got %d", profile.Version)
	}
	if profile.Name != "my-profile" {
		t.Errorf("expected name 'my-profile', got %s", profile.Name)
	}
	if profile.Plugins == nil {
		t.Error("expected plugins to be initialized")
	}
	if profile.MCPServers == nil {
		t.Error("expected mcp_servers to be initialized")
	}
	if profile.Permissions.Allow == nil {
		t.Error("expected permissions.allow to be initialized")
	}
	if profile.Permissions.Deny == nil {
		t.Error("expected permissions.deny to be initialized")
	}
}

func TestProfile_ApplyResult(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "result-test")
	os.MkdirAll(projectDir, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock the plugin installer
	originalInstaller := installPluginAtProjectScope
	installPluginAtProjectScope = func(pluginName, dir string) error { return nil }
	defer func() { installPluginAtProjectScope = originalInstaller }()

	profile := &Profile{
		Version: 1,
		Name:    "result-test",
		Plugins: []string{"plugin1", "plugin2", "plugin3"},
		MCPServers: map[string]bool{
			"server1": true,
			"server2": true,
			"server3": false,
		},
		Permissions: PermissionSet{
			Allow: []string{"Bash", "Read", "Write"},
			Deny:  []string{"Delete"},
		},
	}

	result, err := profile.Apply(projectDir)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	if result.PluginsInstalled != 3 {
		t.Errorf("expected 3 plugins installed, got %d", result.PluginsInstalled)
	}
	if result.MCPServersEnabled != 2 {
		t.Errorf("expected 2 MCP servers enabled, got %d", result.MCPServersEnabled)
	}
	if result.PermissionsApplied != 4 {
		t.Errorf("expected 4 permissions, got %d", result.PermissionsApplied)
	}
	if len(result.PluginErrors) != 0 {
		t.Errorf("expected no plugin errors, got %d", len(result.PluginErrors))
	}
}

func TestProfile_ApplyWithPluginErrors(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "error-test")
	os.MkdirAll(projectDir, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock the plugin installer to fail for one plugin
	originalInstaller := installPluginAtProjectScope
	installPluginAtProjectScope = func(pluginName, dir string) error {
		if pluginName == "bad-plugin" {
			return &mockError{"installation failed"}
		}
		return nil
	}
	defer func() { installPluginAtProjectScope = originalInstaller }()

	profile := &Profile{
		Version: 1,
		Name:    "error-test",
		Plugins: []string{"good-plugin", "bad-plugin", "another-good-plugin"},
	}

	result, err := profile.Apply(projectDir)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	if result.PluginsInstalled != 2 {
		t.Errorf("expected 2 plugins installed, got %d", result.PluginsInstalled)
	}
	if len(result.PluginErrors) != 1 {
		t.Errorf("expected 1 plugin error, got %d", len(result.PluginErrors))
	}
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestProfile_ApplySettingsContent(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "settings-content")
	os.MkdirAll(projectDir, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock the plugin installer
	originalInstaller := installPluginAtProjectScope
	installPluginAtProjectScope = func(pluginName, dir string) error { return nil }
	defer func() { installPluginAtProjectScope = originalInstaller }()

	profile := &Profile{
		Version: 1,
		Name:    "settings-test",
		Plugins: []string{"databricks-tools"},
		Permissions: PermissionSet{
			Allow: []string{"Bash(python:*)", "Read(~/projects/**)"},
			Deny:  []string{"Write(/etc/**)"},
		},
	}

	_, err := profile.Apply(projectDir)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	// Read and verify settings.json content
	settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings.json: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to parse settings.json: %v", err)
	}

	// Verify permissions
	perms, ok := settings["permissions"].(map[string]interface{})
	if !ok {
		t.Fatal("settings.json missing permissions")
	}

	allow, ok := perms["allow"].([]interface{})
	if !ok {
		t.Fatal("settings.json missing permissions.allow")
	}
	if len(allow) != 2 {
		t.Errorf("expected 2 allow permissions, got %d", len(allow))
	}

	deny, ok := perms["deny"].([]interface{})
	if !ok {
		t.Fatal("settings.json missing permissions.deny")
	}
	if len(deny) != 1 {
		t.Errorf("expected 1 deny permission, got %d", len(deny))
	}

	// Verify enabledPlugins
	enabledPlugins, ok := settings["enabledPlugins"].(map[string]interface{})
	if !ok {
		t.Fatal("settings.json missing enabledPlugins")
	}
	// Plugins are added with @claude-vibe suffix by default
	if enabledPlugins["databricks-tools@claude-vibe"] != true {
		t.Errorf("expected databricks-tools@claude-vibe to be enabled, got: %v", enabledPlugins)
	}
}

func TestProfile_ApplyClaudeJSONContent(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "claude-json-content")
	os.MkdirAll(projectDir, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock the plugin installer
	originalInstaller := installPluginAtProjectScope
	installPluginAtProjectScope = func(pluginName, dir string) error { return nil }
	defer func() { installPluginAtProjectScope = originalInstaller }()

	// Create a mock MCP server config in settings.json
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	settingsContent := `{
		"mcpServers": {
			"test-server": {
				"command": "test-cmd",
				"args": ["--arg1", "--arg2"]
			}
		}
	}`
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(settingsContent), 0644)

	profile := &Profile{
		Version: 1,
		Name:    "claude-json-test",
		MCPServers: map[string]bool{
			"test-server": true,
		},
	}

	_, err := profile.Apply(projectDir)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	// Read and verify .claude.json content
	claudeJSONPath := filepath.Join(projectDir, ".claude.json")
	data, err := os.ReadFile(claudeJSONPath)
	if err != nil {
		t.Fatalf("Failed to read .claude.json: %v", err)
	}

	var claudeJSON map[string]interface{}
	if err := json.Unmarshal(data, &claudeJSON); err != nil {
		t.Fatalf("Failed to parse .claude.json: %v", err)
	}

	// Verify mcpServers
	mcpServers, ok := claudeJSON["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal(".claude.json missing mcpServers")
	}

	testServer, ok := mcpServers["test-server"].(map[string]interface{})
	if !ok {
		t.Fatal(".claude.json missing test-server")
	}
	if testServer["command"] != "test-cmd" {
		t.Errorf("expected command 'test-cmd', got %v", testServer["command"])
	}
}

func TestProfile_ApplyNoPlugins(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "no-plugins")
	os.MkdirAll(projectDir, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock the plugin installer - should never be called
	originalInstaller := installPluginAtProjectScope
	installCalled := false
	installPluginAtProjectScope = func(pluginName, dir string) error {
		installCalled = true
		return nil
	}
	defer func() { installPluginAtProjectScope = originalInstaller }()

	profile := &Profile{
		Version: 1,
		Name:    "no-plugins",
		Plugins: []string{},
		Permissions: PermissionSet{
			Allow: []string{"Bash"},
		},
	}

	result, err := profile.Apply(projectDir)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	if installCalled {
		t.Error("plugin installer should not be called when no plugins")
	}
	if result.PluginsInstalled != 0 {
		t.Errorf("expected 0 plugins installed, got %d", result.PluginsInstalled)
	}
}

func TestApplyResult_Fields(t *testing.T) {
	result := ApplyResult{
		PluginsInstalled:   5,
		MCPServersEnabled:  3,
		PermissionsApplied: 10,
		PluginErrors:       []string{"error1", "error2"},
	}

	if result.PluginsInstalled != 5 {
		t.Errorf("expected 5, got %d", result.PluginsInstalled)
	}
	if result.MCPServersEnabled != 3 {
		t.Errorf("expected 3, got %d", result.MCPServersEnabled)
	}
	if result.PermissionsApplied != 10 {
		t.Errorf("expected 10, got %d", result.PermissionsApplied)
	}
	if len(result.PluginErrors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.PluginErrors))
	}
}

func TestDeleteProfile(t *testing.T) {
	tempDir := t.TempDir()

	// Create profiles directory with a profile
	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)

	profileContent := `version: 1
name: to-delete
description: "Profile to be deleted"
`
	os.WriteFile(filepath.Join(profilesDir, "to-delete.yaml"), []byte(profileContent), 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Verify profile exists first
	if !ProfileExists("to-delete") {
		t.Fatal("Profile should exist before deletion")
	}

	// Delete the profile
	err := DeleteProfile("to-delete")
	if err != nil {
		t.Fatalf("DeleteProfile() error: %v", err)
	}

	// Verify profile no longer exists
	if ProfileExists("to-delete") {
		t.Error("Profile should not exist after deletion")
	}
}

func TestDeleteProfile_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	err := DeleteProfile("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
}

func TestRenameProfile(t *testing.T) {
	tempDir := t.TempDir()

	// Create profiles directory with a profile
	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)

	profileContent := `version: 1
name: old-name
description: "Profile to be renamed"
plugins:
  - test-plugin
`
	os.WriteFile(filepath.Join(profilesDir, "old-name.yaml"), []byte(profileContent), 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Rename the profile
	err := RenameProfile("old-name", "new-name")
	if err != nil {
		t.Fatalf("RenameProfile() error: %v", err)
	}

	// Verify old name no longer exists
	if ProfileExists("old-name") {
		t.Error("Old profile name should not exist after rename")
	}

	// Verify new name exists
	if !ProfileExists("new-name") {
		t.Error("New profile name should exist after rename")
	}

	// Verify content was preserved
	profile, err := LoadProfile("new-name")
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}
	if profile.Name != "new-name" {
		t.Errorf("expected name 'new-name', got %s", profile.Name)
	}
	if profile.Description != "Profile to be renamed" {
		t.Errorf("expected description preserved, got %s", profile.Description)
	}
	if len(profile.Plugins) != 1 || profile.Plugins[0] != "test-plugin" {
		t.Error("expected plugins to be preserved")
	}
}

func TestRenameProfile_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	err := RenameProfile("nonexistent", "new-name")
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
}

func TestProfileExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create profiles directory with a profile
	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)

	profileContent := `version: 1
name: existing-profile
`
	os.WriteFile(filepath.Join(profilesDir, "existing-profile.yaml"), []byte(profileContent), 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	if !ProfileExists("existing-profile") {
		t.Error("expected profile to exist")
	}

	if ProfileExists("nonexistent-profile") {
		t.Error("expected profile to not exist")
	}
}

func TestLoadProfile_MinimalYAML_InitializesSlices(t *testing.T) {
	tempDir := t.TempDir()

	// Create a minimal profile without plugins, mcp_servers, or permissions
	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)

	// This minimal profile should NOT cause nil slices
	minimalProfile := `version: 1
name: minimal-profile
`
	os.WriteFile(filepath.Join(profilesDir, "minimal-profile.yaml"), []byte(minimalProfile), 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profile, err := LoadProfile("minimal-profile")
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}

	// Verify slices are initialized (not nil)
	if profile.Plugins == nil {
		t.Error("expected Plugins to be initialized, got nil")
	}
	if profile.MCPServers == nil {
		t.Error("expected MCPServers to be initialized, got nil")
	}
	if profile.Permissions.Allow == nil {
		t.Error("expected Permissions.Allow to be initialized, got nil")
	}
	if profile.Permissions.Deny == nil {
		t.Error("expected Permissions.Deny to be initialized, got nil")
	}
}

func TestProfile_ApplyWithNilSlices_OutputsEmptyArrays(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "nil-slices-test")
	os.MkdirAll(projectDir, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Mock the plugin installer
	originalInstaller := installPluginAtProjectScope
	installPluginAtProjectScope = func(pluginName, dir string) error { return nil }
	defer func() { installPluginAtProjectScope = originalInstaller }()

	// Simulate a profile loaded from YAML that has nil slices
	// (this could happen if the YAML file doesn't have these fields)
	profile := &Profile{
		Version: 1,
		Name:    "nil-test",
		// Deliberately leave Plugins, MCPServers, Permissions as zero values (nil)
	}

	_, err := profile.Apply(projectDir)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	// Read the generated settings.json
	settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings.json: %v", err)
	}

	// The JSON should NOT contain "null" for allow or deny
	content := string(data)
	if content == "" {
		t.Fatal("settings.json is empty")
	}

	// Parse and verify the structure
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to parse settings.json: %v", err)
	}

	perms, ok := settings["permissions"].(map[string]interface{})
	if !ok {
		t.Fatal("settings.json missing permissions")
	}

	// Verify allow and deny are arrays, not null
	allow, ok := perms["allow"].([]interface{})
	if !ok {
		t.Fatalf("permissions.allow should be an array, got %T", perms["allow"])
	}
	if allow == nil {
		t.Error("permissions.allow should not be nil")
	}
	if len(allow) != 0 {
		t.Errorf("permissions.allow should be empty, got %d elements", len(allow))
	}

	deny, ok := perms["deny"].([]interface{})
	if !ok {
		t.Fatalf("permissions.deny should be an array, got %T", perms["deny"])
	}
	if deny == nil {
		t.Error("permissions.deny should not be nil")
	}
	if len(deny) != 0 {
		t.Errorf("permissions.deny should be empty, got %d elements", len(deny))
	}
}

func TestValidateProfileName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"valid lowercase", "my-profile", false},
		{"valid uppercase", "My-Profile", false},
		{"valid with numbers", "profile-123", false},
		{"valid with underscore", "my_profile", false},
		{"valid mixed", "My_Profile-123", false},
		{"empty", "", true},
		{"with spaces", "my profile", true},
		{"with special chars", "my@profile", true},
		{"with slash", "my/profile", true},
		{"with dot", "my.profile", true},
		{"with colon", "my:profile", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfileName(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
			}
		})
	}
}

func TestLoadProfile_WithIdentityFields(t *testing.T) {
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)

	profileContent := `version: 1
name: identity-test
description: "Profile with identity"
email: user@example.com
git_email: user@example.com
env_file: identity-test.env
integrations:
  github: true
  slack: false
databricks_profile: DEFAULT
`
	os.WriteFile(filepath.Join(profilesDir, "identity-test.yaml"), []byte(profileContent), 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profile, err := LoadProfile("identity-test")
	if err != nil {
		t.Fatalf("LoadProfile() error: %v", err)
	}

	if profile.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got %s", profile.Email)
	}
	if profile.GitEmail != "user@example.com" {
		t.Errorf("expected git_email 'user@example.com', got %s", profile.GitEmail)
	}
	if profile.EnvFile != "identity-test.env" {
		t.Errorf("expected env_file 'identity-test.env', got %s", profile.EnvFile)
	}
	if !profile.Integrations["github"] {
		t.Error("expected github integration to be true")
	}
	if profile.Integrations["slack"] {
		t.Error("expected slack integration to be false")
	}
}

func TestProfile_SyncEnvFile(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)
	os.MkdirAll(filepath.Join(tempDir, ".vibe"), 0755)

	envContent := "# env for test\nexport FOO=bar\n"
	os.WriteFile(filepath.Join(profilesDir, "test-profile.env"), []byte(envContent), 0644)

	profile := &Profile{Name: "test-profile", EnvFile: "test-profile.env"}
	err := profile.SyncEnvFile()
	if err != nil {
		t.Fatalf("SyncEnvFile() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, ".vibe", "env"))
	if err != nil {
		t.Fatalf("Failed to read ~/.vibe/env: %v", err)
	}
	if string(data) != envContent {
		t.Errorf("expected %q, got %q", envContent, string(data))
	}
}

func TestProfile_SyncEnvFile_NoEnvFile(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profile := &Profile{Name: "no-env"}
	err := profile.SyncEnvFile()
	if err != nil {
		t.Fatalf("SyncEnvFile() should be no-op, got: %v", err)
	}
}

func TestProfile_SyncEnvFile_FallbackToNamedEnv(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	profilesDir := filepath.Join(tempDir, ".vibe", "profiles")
	os.MkdirAll(profilesDir, 0755)
	os.MkdirAll(filepath.Join(tempDir, ".vibe"), 0755)

	envContent := "# fallback\nexport BAZ=qux\n"
	os.WriteFile(filepath.Join(profilesDir, "fb-profile.env"), []byte(envContent), 0644)

	profile := &Profile{Name: "fb-profile"}
	err := profile.SyncEnvFile()
	if err != nil {
		t.Fatalf("SyncEnvFile() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(tempDir, ".vibe", "env"))
	if string(data) != envContent {
		t.Errorf("expected fallback content, got %q", string(data))
	}
}
