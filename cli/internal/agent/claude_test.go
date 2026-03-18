package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestClaudeAgentImplementsInterface verifies ClaudeAgent implements Agent
func TestClaudeAgentImplementsInterface(t *testing.T) {
	var _ Agent = (*ClaudeAgent)(nil)
}

// TestClaudeAgentName verifies the agent name
func TestClaudeAgentName(t *testing.T) {
	agent := NewClaudeAgent()
	if agent.Name() != "claude" {
		t.Errorf("expected name 'claude', got '%s'", agent.Name())
	}
}

// TestClaudeAgentConfigPaths verifies config paths are correct
func TestClaudeAgentConfigPaths(t *testing.T) {
	agent := NewClaudeAgent()
	paths := agent.ConfigPaths()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	expectedSettings := filepath.Join(homeDir, ".claude", "settings.json")
	if paths.Settings != expectedSettings {
		t.Errorf("expected Settings '%s', got '%s'", expectedSettings, paths.Settings)
	}

	expectedPlugins := filepath.Join(homeDir, ".claude", "plugins", "installed_plugins.json")
	if paths.Plugins != expectedPlugins {
		t.Errorf("expected Plugins '%s', got '%s'", expectedPlugins, paths.Plugins)
	}

	expectedMCPConfig := filepath.Join(homeDir, ".config", "mcp", "config.json")
	if paths.MCPConfig != expectedMCPConfig {
		t.Errorf("expected MCPConfig '%s', got '%s'", expectedMCPConfig, paths.MCPConfig)
	}

	expectedGlobalConfig := filepath.Join(homeDir, ".claude.json")
	if paths.GlobalConfig != expectedGlobalConfig {
		t.Errorf("expected GlobalConfig '%s', got '%s'", expectedGlobalConfig, paths.GlobalConfig)
	}
}

// TestClaudeAgentWithCustomHomeDir verifies paths work with custom home directory
func TestClaudeAgentWithCustomHomeDir(t *testing.T) {
	customHome := "/custom/home"
	agent := NewClaudeAgentWithHome(customHome)
	paths := agent.ConfigPaths()

	expectedSettings := filepath.Join(customHome, ".claude", "settings.json")
	if paths.Settings != expectedSettings {
		t.Errorf("expected Settings '%s', got '%s'", expectedSettings, paths.Settings)
	}

	expectedPlugins := filepath.Join(customHome, ".claude", "plugins", "installed_plugins.json")
	if paths.Plugins != expectedPlugins {
		t.Errorf("expected Plugins '%s', got '%s'", expectedPlugins, paths.Plugins)
	}

	expectedMCPConfig := filepath.Join(customHome, ".config", "mcp", "config.json")
	if paths.MCPConfig != expectedMCPConfig {
		t.Errorf("expected MCPConfig '%s', got '%s'", expectedMCPConfig, paths.MCPConfig)
	}

	expectedGlobalConfig := filepath.Join(customHome, ".claude.json")
	if paths.GlobalConfig != expectedGlobalConfig {
		t.Errorf("expected GlobalConfig '%s', got '%s'", expectedGlobalConfig, paths.GlobalConfig)
	}
}

// TestClaudeAgentIsInstalled tests the IsInstalled method
func TestClaudeAgentIsInstalled(t *testing.T) {
	agent := NewClaudeAgent()

	// This test depends on whether claude is actually installed
	// We just verify it returns a boolean without panicking
	_ = agent.IsInstalled()
}

// TestClaudeAgentIsInstalledWithMockFinder tests IsInstalled with mock command finder
func TestClaudeAgentIsInstalledWithMockFinder(t *testing.T) {
	tests := []struct {
		name     string
		finder   CommandFinder
		expected bool
	}{
		{
			name: "claude found",
			finder: func(name string) (string, error) {
				if name == "claude" {
					return "/usr/local/bin/claude", nil
				}
				return "", os.ErrNotExist
			},
			expected: true,
		},
		{
			name: "claude not found",
			finder: func(name string) (string, error) {
				return "", os.ErrNotExist
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			agent := NewClaudeAgentWithOptions(ClaudeAgentOptions{
				CommandFinder: tc.finder,
			})
			if agent.IsInstalled() != tc.expected {
				t.Errorf("expected IsInstalled() to be %v", tc.expected)
			}
		})
	}
}

// TestClaudeAgentVersion tests the Version method
func TestClaudeAgentVersion(t *testing.T) {
	agent := NewClaudeAgent()

	// This test depends on whether claude is actually installed
	// If not installed, Version should return an error
	version, err := agent.Version()
	if err != nil {
		// Not installed or error - that's fine for this test
		return
	}

	// If we got a version, it should be non-empty
	if version == "" {
		t.Error("expected non-empty version string")
	}
}

// TestClaudeAgentVersionWithMockRunner tests Version with mock command runner
func TestClaudeAgentVersionWithMockRunner(t *testing.T) {
	tests := []struct {
		name        string
		runner      CommandRunner
		expected    string
		expectError bool
	}{
		{
			name: "version returned",
			runner: func(name string, args ...string) (string, error) {
				if name == "claude" && len(args) > 0 && args[0] == "--version" {
					return "claude-code 1.2.3\n", nil
				}
				return "", os.ErrNotExist
			},
			expected:    "claude-code 1.2.3",
			expectError: false,
		},
		{
			name: "command fails",
			runner: func(name string, args ...string) (string, error) {
				return "", os.ErrNotExist
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "version with extra whitespace",
			runner: func(name string, args ...string) (string, error) {
				return "  claude-code 2.0.0  \n\n", nil
			},
			expected:    "claude-code 2.0.0",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			agent := NewClaudeAgentWithOptions(ClaudeAgentOptions{
				CommandRunner: tc.runner,
			})
			version, err := agent.Version()
			if tc.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if version != tc.expected {
					t.Errorf("expected version '%s', got '%s'", tc.expected, version)
				}
			}
		})
	}
}

// TestClaudeAgentLaunchSession tests the LaunchSession method
func TestClaudeAgentLaunchSession(t *testing.T) {
	// Track what command was launched
	var launchedCommand string
	var launchedArgs []string
	var launchedDir string

	mockLauncher := func(dir, name string, args ...string) error {
		launchedDir = dir
		launchedCommand = name
		launchedArgs = args
		return nil
	}

	agent := NewClaudeAgentWithOptions(ClaudeAgentOptions{
		SessionLauncher: mockLauncher,
	})

	opts := SessionOptions{
		WorkDir: "/test/workdir",
		Profile: "",
	}

	err := agent.LaunchSession(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if launchedDir != "/test/workdir" {
		t.Errorf("expected workdir '/test/workdir', got '%s'", launchedDir)
	}

	if launchedCommand != "claude" {
		t.Errorf("expected command 'claude', got '%s'", launchedCommand)
	}

	// Should not have any extra args for empty profile
	if len(launchedArgs) != 0 {
		t.Errorf("expected no args, got %v", launchedArgs)
	}
}

// TestClaudeAgentLaunchSessionWithProfile tests LaunchSession with a profile
func TestClaudeAgentLaunchSessionWithProfile(t *testing.T) {
	var launchedArgs []string

	mockLauncher := func(dir, name string, args ...string) error {
		launchedArgs = args
		return nil
	}

	agent := NewClaudeAgentWithOptions(ClaudeAgentOptions{
		SessionLauncher: mockLauncher,
	})

	opts := SessionOptions{
		WorkDir: "/test/workdir",
		Profile: "my-profile",
	}

	err := agent.LaunchSession(opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Profile is stored in opts but not passed to claude directly
	// (profile application happens before launch in the vibe agent command)
	if len(launchedArgs) != 0 {
		t.Errorf("expected no args passed to claude, got %v", launchedArgs)
	}
}

// TestClaudeAgentLaunchSessionError tests LaunchSession error handling
func TestClaudeAgentLaunchSessionError(t *testing.T) {
	expectedErr := &testError{msg: "launch failed"}

	mockLauncher := func(dir, name string, args ...string) error {
		return expectedErr
	}

	agent := NewClaudeAgentWithOptions(ClaudeAgentOptions{
		SessionLauncher: mockLauncher,
	})

	err := agent.LaunchSession(SessionOptions{})
	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

// TestClaudeAgentRegistration tests that Claude agent auto-registers
func TestClaudeAgentRegistration(t *testing.T) {
	// Clear registry
	clearRegistry()

	// Create and register claude agent
	agent := NewClaudeAgent()
	Register("claude", agent)

	// Should be retrievable
	retrieved, ok := Get("claude")
	if !ok {
		t.Error("expected claude to be registered")
	}
	if retrieved != agent {
		t.Error("expected retrieved agent to match registered agent")
	}

	// Should be the default
	defaultAgent := Default()
	if defaultAgent != agent {
		t.Error("expected claude to be the default agent")
	}
}

// TestClaudeAgentBinaryName tests the binary name used
func TestClaudeAgentBinaryName(t *testing.T) {
	agent := NewClaudeAgent()

	// The binary name should be "claude"
	if agent.binaryName != "claude" {
		t.Errorf("expected binary name 'claude', got '%s'", agent.binaryName)
	}
}

// TestClaudeAgentWithCustomBinary tests using a custom binary name
func TestClaudeAgentWithCustomBinary(t *testing.T) {
	agent := NewClaudeAgentWithOptions(ClaudeAgentOptions{
		BinaryName: "claude-dev",
	})

	if agent.binaryName != "claude-dev" {
		t.Errorf("expected binary name 'claude-dev', got '%s'", agent.binaryName)
	}
}

// TestParseVersion tests version string parsing
func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"claude-code 1.2.3", "claude-code 1.2.3"},
		{"1.2.3", "1.2.3"},
		{"  claude-code 1.2.3  \n", "claude-code 1.2.3"},
		{"", ""},
	}

	for _, tc := range tests {
		result := strings.TrimSpace(tc.input)
		if result != tc.expected {
			t.Errorf("parsing '%s': expected '%s', got '%s'", tc.input, tc.expected, result)
		}
	}
}
