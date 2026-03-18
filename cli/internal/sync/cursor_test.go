package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCursorTarget_Name(t *testing.T) {
	target := NewCursorTargetWithHome(t.TempDir())
	if target.Name() != "cursor" {
		t.Errorf("expected name 'cursor', got %q", target.Name())
	}
}

func TestCursorTarget_IsInstalled(t *testing.T) {
	home := t.TempDir()

	// Not installed
	target := NewCursorTargetWithHome(home)
	if target.IsInstalled() {
		t.Error("expected not installed when .cursor dir missing")
	}

	// Installed
	os.MkdirAll(filepath.Join(home, ".cursor"), 0755)
	if !target.IsInstalled() {
		t.Error("expected installed when .cursor dir exists")
	}
}

func TestCursorTarget_SyncMCP(t *testing.T) {
	home := t.TempDir()
	cursorDir := filepath.Join(home, ".cursor")
	os.MkdirAll(cursorDir, 0755)

	target := NewCursorTargetWithHome(home)

	servers := []MCPServerConfig{
		{
			Name:    "test-server",
			Command: "npx",
			Args:    []string{"-y", "test-mcp@latest"},
		},
		{
			Name:    "env-server",
			Command: "python3",
			Args:    []string{"server.py"},
			Env:     map[string]string{"API_KEY": "test123"},
		},
	}

	result, err := target.SyncMCP(servers, SyncOptions{})
	if err != nil {
		t.Fatalf("SyncMCP failed: %v", err)
	}
	if result.ItemsSynced != 2 {
		t.Errorf("expected 2 items synced, got %d", result.ItemsSynced)
	}

	// Verify written file
	data, err := os.ReadFile(filepath.Join(cursorDir, "mcp.json"))
	if err != nil {
		t.Fatalf("failed to read mcp.json: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse mcp.json: %v", err)
	}

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("mcpServers not found in config")
	}

	if len(mcpServers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(mcpServers))
	}

	// Verify env-server has env field
	envServer, ok := mcpServers["env-server"].(map[string]interface{})
	if !ok {
		t.Fatal("env-server not found")
	}
	env, ok := envServer["env"].(map[string]interface{})
	if !ok {
		t.Fatal("env field not found on env-server")
	}
	if env["API_KEY"] != "test123" {
		t.Errorf("expected API_KEY=test123, got %v", env["API_KEY"])
	}
}

func TestCursorTarget_SyncMCP_PreservesUserServers(t *testing.T) {
	home := t.TempDir()
	cursorDir := filepath.Join(home, ".cursor")
	os.MkdirAll(cursorDir, 0755)

	// Write existing config with a user-added server
	existingConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"user-custom-server": map[string]interface{}{
				"command": "my-server",
				"args":    []interface{}{"--port", "8080"},
			},
		},
	}
	data, _ := json.MarshalIndent(existingConfig, "", "  ")
	os.WriteFile(filepath.Join(cursorDir, "mcp.json"), data, 0644)

	target := NewCursorTargetWithHome(home)

	// Sync vibe servers
	servers := []MCPServerConfig{
		{Name: "vibe-server", Command: "npx", Args: []string{"vibe-mcp"}},
	}

	_, err := target.SyncMCP(servers, SyncOptions{})
	if err != nil {
		t.Fatalf("SyncMCP failed: %v", err)
	}

	// Verify both servers exist
	data, _ = os.ReadFile(filepath.Join(cursorDir, "mcp.json"))
	var config map[string]interface{}
	json.Unmarshal(data, &config)

	mcpServers := config["mcpServers"].(map[string]interface{})
	if len(mcpServers) != 2 {
		t.Errorf("expected 2 servers (user + vibe), got %d", len(mcpServers))
	}
	if _, ok := mcpServers["user-custom-server"]; !ok {
		t.Error("user-custom-server was not preserved")
	}
	if _, ok := mcpServers["vibe-server"]; !ok {
		t.Error("vibe-server was not added")
	}
}

func TestCursorTarget_SyncMCP_DryRun(t *testing.T) {
	home := t.TempDir()
	cursorDir := filepath.Join(home, ".cursor")
	os.MkdirAll(cursorDir, 0755)

	target := NewCursorTargetWithHome(home)

	servers := []MCPServerConfig{
		{Name: "test-server", Command: "npx", Args: []string{"test"}},
	}

	result, err := target.SyncMCP(servers, SyncOptions{DryRun: true})
	if err != nil {
		t.Fatalf("SyncMCP dry-run failed: %v", err)
	}
	if result.ItemsSynced != 1 {
		t.Errorf("expected 1 item synced in dry-run, got %d", result.ItemsSynced)
	}

	// Verify no file was written
	if _, err := os.Stat(filepath.Join(cursorDir, "mcp.json")); err == nil {
		t.Error("mcp.json should not exist after dry-run")
	}
}

func TestCursorTarget_SyncSkills(t *testing.T) {
	home := t.TempDir()
	cursorDir := filepath.Join(home, ".cursor")
	os.MkdirAll(cursorDir, 0755)

	// Create a fake source skill
	sourceSkillDir := filepath.Join(t.TempDir(), "my-skill")
	os.MkdirAll(sourceSkillDir, 0755)
	os.WriteFile(filepath.Join(sourceSkillDir, "SKILL.md"), []byte("---\nname: my-skill\ndescription: A test skill\n---\n\nInstructions here."), 0644)
	os.MkdirAll(filepath.Join(sourceSkillDir, "resources"), 0755)
	os.WriteFile(filepath.Join(sourceSkillDir, "resources", "helper.py"), []byte("print('hello')"), 0644)

	target := NewCursorTargetWithHome(home)

	skills := []SkillSource{
		{Name: "my-skill", PluginName: "test-plugin", SourcePath: sourceSkillDir},
	}

	result, err := target.SyncSkills(skills, SyncOptions{})
	if err != nil {
		t.Fatalf("SyncSkills failed: %v", err)
	}
	if result.ItemsSynced != 1 {
		t.Errorf("expected 1 skill synced, got %d", result.ItemsSynced)
	}

	// Verify skill was copied without prefix
	targetSkillDir := filepath.Join(cursorDir, "skills", "my-skill")
	if _, err := os.Stat(filepath.Join(targetSkillDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found in target")
	}
	if _, err := os.Stat(filepath.Join(targetSkillDir, "resources", "helper.py")); err != nil {
		t.Error("resources/helper.py not found in target")
	}

	// Verify manifest was updated in ~/.vibe/sync/
	manifestDir := filepath.Join(home, ".vibe", "sync")
	manifest, err := LoadManifest(manifestDir, "cursor")
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	if !manifest.IsManagedSkill("my-skill") {
		t.Error("skill not tracked in manifest")
	}
}

func TestCursorTarget_SyncMCP_Empty(t *testing.T) {
	home := t.TempDir()
	cursorDir := filepath.Join(home, ".cursor")
	os.MkdirAll(cursorDir, 0755)

	target := NewCursorTargetWithHome(home)

	result, err := target.SyncMCP(nil, SyncOptions{})
	if err != nil {
		t.Fatalf("SyncMCP failed: %v", err)
	}
	if result.ItemsSynced != 0 {
		t.Errorf("expected 0 items synced, got %d", result.ItemsSynced)
	}
}
