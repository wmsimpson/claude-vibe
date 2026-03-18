package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPermissionsConfig_Empty(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPermissionsConfig()

	allow := pc.AllowList()
	deny := pc.DenyList()

	if len(allow) != 0 {
		t.Errorf("expected empty allow list, got %d items", len(allow))
	}
	if len(deny) != 0 {
		t.Errorf("expected empty deny list, got %d items", len(deny))
	}
}

func TestPermissionsConfig_LoadFromSettings(t *testing.T) {
	tempDir := t.TempDir()

	// Create ~/.claude/settings.json
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settingsJSON := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []string{
				"Bash",
				"Read(~/**)",
				"Skill(databricks-query)",
			},
			"deny": []string{
				"Write(/etc/**)",
			},
		},
	}

	data, _ := json.MarshalIndent(settingsJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPermissionsConfig()

	allow := pc.AllowList()
	deny := pc.DenyList()

	if len(allow) != 3 {
		t.Fatalf("expected 3 allow items, got %d", len(allow))
	}
	if len(deny) != 1 {
		t.Fatalf("expected 1 deny item, got %d", len(deny))
	}

	// Check specific permissions
	if !pc.HasPermission("Bash") {
		t.Error("expected 'Bash' permission to be present")
	}
	if !pc.HasPermission("Read(~/**)") {
		t.Error("expected 'Read(~/**)' permission to be present")
	}
}

func TestPermissionsConfig_LoadFromTopLevelAllow(t *testing.T) {
	tempDir := t.TempDir()

	// Some Claude configs use top-level "allow" instead of nested permissions
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settingsJSON := map[string]interface{}{
		"allow": []string{
			"Bash",
			"Skill",
		},
	}

	data, _ := json.MarshalIndent(settingsJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPermissionsConfig()

	allow := pc.AllowList()

	if len(allow) != 2 {
		t.Fatalf("expected 2 allow items, got %d", len(allow))
	}
}

func TestPermissionsConfig_HasPermission(t *testing.T) {
	tempDir := t.TempDir()

	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settingsJSON := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []string{
				"Bash",
				"Skill(test-skill)",
			},
			"deny": []string{},
		},
	}

	data, _ := json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPermissionsConfig()

	if !pc.HasPermission("Bash") {
		t.Error("expected 'Bash' permission to exist")
	}
	if !pc.HasPermission("Skill(test-skill)") {
		t.Error("expected 'Skill(test-skill)' permission to exist")
	}
	if pc.HasPermission("NonExistent") {
		t.Error("expected 'NonExistent' permission to not exist")
	}
}

func TestPermissionsConfig_AddPermission(t *testing.T) {
	tempDir := t.TempDir()

	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settingsJSON := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []string{"Bash"},
			"deny":  []string{},
		},
	}

	data, _ := json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPermissionsConfig()

	// Add new permission
	err := pc.AddPermission("Skill(new-skill)")
	if err != nil {
		t.Fatalf("AddPermission() error: %v", err)
	}

	// Verify it's added
	if !pc.HasPermission("Skill(new-skill)") {
		t.Error("expected new permission to be added")
	}

	// Adding duplicate should not error
	err = pc.AddPermission("Bash")
	if err != nil {
		t.Fatalf("AddPermission() for duplicate error: %v", err)
	}
}

func TestPermissionsConfig_RemovePermission(t *testing.T) {
	tempDir := t.TempDir()

	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settingsJSON := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []string{"Bash", "Skill(remove-me)"},
			"deny":  []string{},
		},
	}

	data, _ := json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPermissionsConfig()

	// Remove permission
	err := pc.RemovePermission("Skill(remove-me)")
	if err != nil {
		t.Fatalf("RemovePermission() error: %v", err)
	}

	// Verify it's removed
	if pc.HasPermission("Skill(remove-me)") {
		t.Error("expected permission to be removed")
	}

	// Bash should still be there
	if !pc.HasPermission("Bash") {
		t.Error("expected 'Bash' to still be present")
	}
}

func TestPermissionsConfig_RemovePermission_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settingsJSON := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []string{"Bash"},
			"deny":  []string{},
		},
	}

	data, _ := json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPermissionsConfig()

	// Remove non-existent permission should not error
	err := pc.RemovePermission("NonExistent")
	if err != nil {
		t.Fatalf("RemovePermission() for non-existent should not error: %v", err)
	}
}

func TestPermissionsConfig_Save(t *testing.T) {
	tempDir := t.TempDir()

	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settingsJSON := map[string]interface{}{
		"permissions": map[string]interface{}{
			"allow": []string{"Bash"},
			"deny":  []string{},
		},
	}

	data, _ := json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPermissionsConfig()
	pc.AddPermission("Skill(saved-skill)")

	err := pc.Save()
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Reload and verify
	pc2 := NewPermissionsConfig()
	if !pc2.HasPermission("Skill(saved-skill)") {
		t.Error("expected saved permission to persist after reload")
	}
}

func TestPermissionsConfig_SavePreservesOtherFields(t *testing.T) {
	tempDir := t.TempDir()

	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Settings with multiple fields
	settingsJSON := map[string]interface{}{
		"model": "opus",
		"permissions": map[string]interface{}{
			"allow": []string{"Bash"},
			"deny":  []string{},
		},
		"mcpServers": map[string]interface{}{
			"test": map[string]interface{}{
				"command": "test",
			},
		},
	}

	data, _ := json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	pc := NewPermissionsConfig()
	pc.AddPermission("NewPerm")
	pc.Save()

	// Read the file and verify other fields are preserved
	content, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	var result map[string]interface{}
	json.Unmarshal(content, &result)

	if result["model"] != "opus" {
		t.Error("expected 'model' field to be preserved")
	}
	if result["mcpServers"] == nil {
		t.Error("expected 'mcpServers' field to be preserved")
	}
}

func TestClaudeSettingsPath(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expected := filepath.Join(tempDir, ".claude", "settings.json")
	actual := ClaudeSettingsPath()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}
