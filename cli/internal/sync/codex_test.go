package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexTarget_Name(t *testing.T) {
	target := NewCodexTargetWithHome(t.TempDir())
	if target.Name() != "codex" {
		t.Errorf("expected name 'codex', got %q", target.Name())
	}
}

func TestCodexTarget_IsInstalled(t *testing.T) {
	home := t.TempDir()

	target := NewCodexTargetWithHome(home)
	if target.IsInstalled() {
		t.Error("expected not installed when .codex dir missing")
	}

	os.MkdirAll(filepath.Join(home, ".codex"), 0755)
	if !target.IsInstalled() {
		t.Error("expected installed when .codex dir exists")
	}
}

func TestCodexTarget_SyncMCP(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	os.MkdirAll(codexDir, 0755)

	target := NewCodexTargetWithHome(home)

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

	// Verify config.toml was written
	data, err := os.ReadFile(filepath.Join(codexDir, "config.toml"))
	if err != nil {
		t.Fatalf("failed to read config.toml: %v", err)
	}

	content := string(data)

	// Verify server entries exist
	if !strings.Contains(content, "[mcp_servers.test-server]") &&
		!strings.Contains(content, `[mcp_servers."test-server"]`) {
		t.Error("test-server not found in config.toml")
	}
	if !strings.Contains(content, `command = "npx"`) {
		t.Error("command field not found")
	}
	if !strings.Contains(content, "API_KEY") {
		t.Error("env field not found for env-server")
	}
}

func TestCodexTarget_SyncMCP_PreservesExistingConfig(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	os.MkdirAll(codexDir, 0755)

	// Write existing config with non-MCP settings and a user server
	existingConfig := `model = "gpt-5-codex"
approval_policy = "on-request"

[mcp_servers.user-server]
command = "my-server"
args = ["--port", "8080"]
`
	os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(existingConfig), 0644)

	target := NewCodexTargetWithHome(home)

	servers := []MCPServerConfig{
		{Name: "vibe-server", Command: "npx", Args: []string{"vibe-mcp"}},
	}

	_, err := target.SyncMCP(servers, SyncOptions{})
	if err != nil {
		t.Fatalf("SyncMCP failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(codexDir, "config.toml"))
	if err != nil {
		t.Fatalf("failed to read config.toml: %v", err)
	}

	content := string(data)

	// Non-MCP settings should be preserved
	if !strings.Contains(content, "gpt-5-codex") {
		t.Error("existing model setting was not preserved")
	}
	if !strings.Contains(content, "on-request") {
		t.Error("existing approval_policy was not preserved")
	}

	// User server should be preserved (not in manifest)
	if !strings.Contains(content, "user-server") {
		t.Error("user-server was not preserved")
	}

	// Vibe server should be added
	if !strings.Contains(content, "vibe-server") {
		t.Error("vibe-server was not added")
	}
}

func TestCodexTarget_SyncMCP_DryRun(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	os.MkdirAll(codexDir, 0755)

	target := NewCodexTargetWithHome(home)

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
	if _, err := os.Stat(filepath.Join(codexDir, "config.toml")); err == nil {
		t.Error("config.toml should not exist after dry-run")
	}
}

func TestCodexTarget_SyncSkills(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	os.MkdirAll(codexDir, 0755)

	// Create a fake source skill
	sourceSkillDir := filepath.Join(t.TempDir(), "my-skill")
	os.MkdirAll(sourceSkillDir, 0755)
	os.WriteFile(filepath.Join(sourceSkillDir, "SKILL.md"), []byte("---\nname: my-skill\ndescription: A test skill\n---\n\nInstructions here."), 0644)
	os.MkdirAll(filepath.Join(sourceSkillDir, "resources"), 0755)
	os.WriteFile(filepath.Join(sourceSkillDir, "resources", "helper.py"), []byte("print('hello')"), 0644)

	target := NewCodexTargetWithHome(home)

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
	targetSkillDir := filepath.Join(codexDir, "skills", "my-skill")
	if _, err := os.Stat(filepath.Join(targetSkillDir, "SKILL.md")); err != nil {
		t.Error("SKILL.md not found in target")
	}
	if _, err := os.Stat(filepath.Join(targetSkillDir, "resources", "helper.py")); err != nil {
		t.Error("resources/helper.py not found in target")
	}

	// Verify manifest was updated in ~/.vibe/sync/
	manifestDir := filepath.Join(home, ".vibe", "sync")
	manifest, err := LoadManifest(manifestDir, "codex")
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	if !manifest.IsManagedSkill("my-skill") {
		t.Error("skill not tracked in manifest")
	}
	if _, ok := manifest.Checksums["my-skill"]; !ok {
		t.Error("checksum not stored in manifest")
	}
}

func TestCodexTarget_SyncSkills_ReplacesExisting(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	os.MkdirAll(codexDir, 0755)

	// Create source skill
	sourceSkillDir := filepath.Join(t.TempDir(), "my-skill")
	os.MkdirAll(sourceSkillDir, 0755)
	os.WriteFile(filepath.Join(sourceSkillDir, "SKILL.md"), []byte("version 2"), 0644)

	// Create existing target skill with different content
	targetSkillDir := filepath.Join(codexDir, "skills", "my-skill")
	os.MkdirAll(targetSkillDir, 0755)
	os.WriteFile(filepath.Join(targetSkillDir, "SKILL.md"), []byte("version 1"), 0644)
	os.WriteFile(filepath.Join(targetSkillDir, "old-file.txt"), []byte("should be removed"), 0644)

	target := NewCodexTargetWithHome(home)

	skills := []SkillSource{
		{Name: "my-skill", PluginName: "test-plugin", SourcePath: sourceSkillDir},
	}

	_, err := target.SyncSkills(skills, SyncOptions{})
	if err != nil {
		t.Fatalf("SyncSkills failed: %v", err)
	}

	// Verify new content
	data, _ := os.ReadFile(filepath.Join(targetSkillDir, "SKILL.md"))
	if string(data) != "version 2" {
		t.Errorf("expected 'version 2', got %q", string(data))
	}

	// Verify old file was removed
	if _, err := os.Stat(filepath.Join(targetSkillDir, "old-file.txt")); err == nil {
		t.Error("old-file.txt should have been removed")
	}
}

func TestCodexTarget_SyncMCP_Empty(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	os.MkdirAll(codexDir, 0755)

	target := NewCodexTargetWithHome(home)

	result, err := target.SyncMCP(nil, SyncOptions{})
	if err != nil {
		t.Fatalf("SyncMCP failed: %v", err)
	}
	if result.ItemsSynced != 0 {
		t.Errorf("expected 0 items synced, got %d", result.ItemsSynced)
	}
}
