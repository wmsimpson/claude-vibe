package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMCPConfig_ListServers_Empty(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	mc := NewMCPConfig()
	servers := mc.ListServers()

	if len(servers) != 0 {
		t.Errorf("expected empty list, got %d servers", len(servers))
	}
}

func TestMCPConfig_ListServers_FromClaudeJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Create ~/.claude.json with mcpServers (legacy location)
	claudeJSON := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"chrome-devtools": map[string]interface{}{
				"command": "npx",
				"args":    []string{"chrome-devtools-mcp@latest"},
			},
			"slack": map[string]interface{}{
				"command": "python3.10",
				"args":    []string{"~/mcp/servers/slack_mcp.pex"},
				"env": map[string]string{
					"SOME_VAR": "value",
				},
			},
		},
	}

	data, _ := json.MarshalIndent(claudeJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(tempDir, ".claude.json"), data, 0644); err != nil {
		t.Fatalf("failed to write .claude.json: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	mc := NewMCPConfig()
	servers := mc.ListServers()

	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	// Find chrome-devtools server
	var chrome *MCPServer
	for i := range servers {
		if servers[i].Name == "chrome-devtools" {
			chrome = &servers[i]
			break
		}
	}

	if chrome == nil {
		t.Fatal("chrome-devtools server not found")
	}
	if chrome.Command != "npx" {
		t.Errorf("expected command 'npx', got %s", chrome.Command)
	}
	if !chrome.Enabled {
		t.Error("expected server to be enabled by default")
	}
}

func TestMCPConfig_ListServers_FromMCPConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create ~/.config/mcp/config.json
	mcpDir := filepath.Join(tempDir, ".config", "mcp")
	os.MkdirAll(mcpDir, 0755)

	mcpConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"glean": map[string]interface{}{
				"command": "python3.10",
				"args":    []string{"glean_mcp.pex"},
			},
		},
	}

	data, _ := json.MarshalIndent(mcpConfig, "", "  ")
	if err := os.WriteFile(filepath.Join(mcpDir, "config.json"), data, 0644); err != nil {
		t.Fatalf("failed to write mcp config.json: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	mc := NewMCPConfig()
	servers := mc.ListServers()

	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}

	if servers[0].Name != "glean" {
		t.Errorf("expected server name 'glean', got %s", servers[0].Name)
	}
}

func TestMCPConfig_ListServers_MergesBothSources(t *testing.T) {
	tempDir := t.TempDir()

	// Create ~/.claude.json
	claudeJSON := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"server1": map[string]interface{}{
				"command": "cmd1",
				"args":    []string{},
			},
		},
	}
	data, _ := json.MarshalIndent(claudeJSON, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".claude.json"), data, 0644)

	// Create ~/.config/mcp/config.json
	mcpDir := filepath.Join(tempDir, ".config", "mcp")
	os.MkdirAll(mcpDir, 0755)

	mcpConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"server2": map[string]interface{}{
				"command": "cmd2",
				"args":    []string{},
			},
		},
	}
	data, _ = json.MarshalIndent(mcpConfig, "", "  ")
	os.WriteFile(filepath.Join(mcpDir, "config.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	mc := NewMCPConfig()
	servers := mc.ListServers()

	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	names := make(map[string]bool)
	for _, s := range servers {
		names[s.Name] = true
	}

	if !names["server1"] || !names["server2"] {
		t.Error("expected both server1 and server2 to be present")
	}
}

func TestMCPConfig_SetEnabled(t *testing.T) {
	tempDir := t.TempDir()

	// Create ~/.claude.json with settings and mcpServers
	claudeJSON := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server": map[string]interface{}{
				"command": "test",
				"args":    []string{},
			},
		},
	}
	data, _ := json.MarshalIndent(claudeJSON, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".claude.json"), data, 0644)

	// Also create settings.json which is where enabledMcpServers would be stored
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	settingsJSON := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server": map[string]interface{}{
				"command": "test",
				"args":    []string{},
			},
		},
	}
	data, _ = json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	mc := NewMCPConfig()

	// Initially enabled
	servers := mc.ListServers()
	if len(servers) == 0 {
		t.Fatal("expected at least one server")
	}

	// Disable the server
	err := mc.SetEnabled("test-server", false)
	if err != nil {
		t.Fatalf("SetEnabled() error: %v", err)
	}

	// Verify it's disabled
	servers = mc.ListServers()
	for _, s := range servers {
		if s.Name == "test-server" && s.Enabled {
			t.Error("expected server to be disabled")
		}
	}

	// Re-enable
	err = mc.SetEnabled("test-server", true)
	if err != nil {
		t.Fatalf("SetEnabled() error: %v", err)
	}

	// Verify it's enabled
	servers = mc.ListServers()
	for _, s := range servers {
		if s.Name == "test-server" && !s.Enabled {
			t.Error("expected server to be enabled")
		}
	}
}

func TestMCPConfig_SetEnabled_NonExistent(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	mc := NewMCPConfig()

	err := mc.SetEnabled("nonexistent-server", false)
	if err == nil {
		t.Error("expected error for non-existent server")
	}
}

func TestMCPConfig_Save(t *testing.T) {
	tempDir := t.TempDir()

	// Create initial config files
	claudeDir := filepath.Join(tempDir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settingsJSON := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server": map[string]interface{}{
				"command": "original",
				"args":    []string{},
			},
		},
	}
	data, _ := json.MarshalIndent(settingsJSON, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	mc := NewMCPConfig()
	mc.SetEnabled("test-server", false)

	err := mc.Save()
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Reload and verify
	mc2 := NewMCPConfig()
	servers := mc2.ListServers()

	for _, s := range servers {
		if s.Name == "test-server" && s.Enabled {
			t.Error("expected server to remain disabled after reload")
		}
	}
}

func TestMCPServer_Fields(t *testing.T) {
	server := MCPServer{
		Name:    "test",
		Command: "python3",
		Args:    []string{"-m", "mcp_server"},
		Env:     map[string]string{"KEY": "value"},
		Enabled: true,
	}

	if server.Name != "test" {
		t.Errorf("expected name 'test', got %s", server.Name)
	}
	if server.Command != "python3" {
		t.Errorf("expected command 'python3', got %s", server.Command)
	}
	if len(server.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(server.Args))
	}
	if server.Env["KEY"] != "value" {
		t.Errorf("expected env KEY='value', got %s", server.Env["KEY"])
	}
	if !server.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestMCPConfigPath(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expected := filepath.Join(tempDir, ".config", "mcp", "config.json")
	actual := MCPConfigPath()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestClaudeJSONPath(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expected := filepath.Join(tempDir, ".claude.json")
	actual := ClaudeJSONPath()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}
