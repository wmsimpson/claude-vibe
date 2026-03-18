package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ============================================================================
// Path Tests
// ============================================================================

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde at start",
			input:    "~/.vibe",
			expected: filepath.Join(home, ".vibe"),
		},
		{
			name:     "tilde alone",
			input:    "~",
			expected: home,
		},
		{
			name:     "no tilde",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "tilde in middle (not expanded)",
			input:    "/path/~/file",
			expected: "/path/~/file",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "relative path",
			input:    "relative/path",
			expected: "relative/path",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ExpandPath(tc.input)
			if result != tc.expected {
				t.Errorf("ExpandPath(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestExpandPathWithEnvVar(t *testing.T) {
	// Set a test environment variable
	testValue := "/test/value"
	os.Setenv("VIBE_TEST_VAR", testValue)
	defer os.Unsetenv("VIBE_TEST_VAR")

	result := ExpandPath("$VIBE_TEST_VAR/subdir")
	expected := filepath.Join(testValue, "subdir")
	if result != expected {
		t.Errorf("ExpandPath with env var = %q, expected %q", result, expected)
	}
}

func TestPathConstants(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		got      string
		contains string
	}{
		{"VibeDir", VibeDir(), ".vibe"},
		{"MarketplaceDir", MarketplaceDir(), "marketplace"},
		{"ProfilesDir", ProfilesDir(), "profiles"},
		{"ConfigFile", ConfigFile(), "config.yaml"},
		{"ClaudeDir", ClaudeDir(), ".claude"},
		{"ClaudeSettings", ClaudeSettings(), "settings.json"},
		{"ClaudePlugins", ClaudePlugins(), "installed_plugins.json"},
		{"ClaudeJSON", ClaudeJSON(), ".claude.json"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.HasPrefix(tc.got, home) {
				t.Errorf("%s = %q, expected to start with home dir %q", tc.name, tc.got, home)
			}
			if !strings.Contains(tc.got, tc.contains) {
				t.Errorf("%s = %q, expected to contain %q", tc.name, tc.got, tc.contains)
			}
		})
	}
}

func TestMCPConfig(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	result := MCPConfig()
	if !strings.HasPrefix(result, home) {
		t.Errorf("MCPConfig() = %q, expected to start with home dir", result)
	}
	if !strings.Contains(result, "config.json") {
		t.Errorf("MCPConfig() = %q, expected to contain 'config.json'", result)
	}
}

// ============================================================================
// Shell Tests
// ============================================================================

func TestDetectShellRC(t *testing.T) {
	result := DetectShellRC()

	// Should return a path ending in .zshrc or .bashrc
	if !strings.HasSuffix(result, ".zshrc") && !strings.HasSuffix(result, ".bashrc") {
		t.Errorf("DetectShellRC() = %q, expected to end with .zshrc or .bashrc", result)
	}

	// Should be in home directory
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(result, home) {
		t.Errorf("DetectShellRC() = %q, expected to start with home dir %q", result, home)
	}
}

func TestHasEnvVar(t *testing.T) {
	// Create a temporary file with some content
	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".testrc")

	content := `# Some shell config
export PATH="/usr/local/bin:$PATH"
export VIBE_EXISTING_VAR="value1"
# export COMMENTED_VAR="value2"
export ANOTHER_VAR="value3"
`
	if err := os.WriteFile(rcFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"existing var", "VIBE_EXISTING_VAR", true},
		{"another var", "ANOTHER_VAR", true},
		{"commented var", "COMMENTED_VAR", false},
		{"nonexistent var", "NONEXISTENT_VAR", false},
		{"PATH var", "PATH", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := HasEnvVar(rcFile, tc.varName)
			if result != tc.expected {
				t.Errorf("HasEnvVar(%q, %q) = %v, expected %v", rcFile, tc.varName, result, tc.expected)
			}
		})
	}
}

func TestHasEnvVarNonexistentFile(t *testing.T) {
	result := HasEnvVar("/nonexistent/path/.rc", "SOME_VAR")
	if result {
		t.Error("HasEnvVar should return false for nonexistent file")
	}
}

func TestEnsureEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".testrc")

	// Create initial file
	initialContent := "# Shell config\nexport EXISTING_VAR=\"value1\"\n"
	if err := os.WriteFile(rcFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test adding a new variable
	t.Run("add new variable", func(t *testing.T) {
		err := EnsureEnvVar(rcFile, "NEW_VAR", "new_value")
		if err != nil {
			t.Fatalf("EnsureEnvVar failed: %v", err)
		}

		content, _ := os.ReadFile(rcFile)
		if !strings.Contains(string(content), `export NEW_VAR="new_value"`) {
			t.Errorf("File should contain new export, got: %s", content)
		}
	})

	// Test that existing variable is not duplicated
	t.Run("skip existing variable", func(t *testing.T) {
		err := EnsureEnvVar(rcFile, "EXISTING_VAR", "different_value")
		if err != nil {
			t.Fatalf("EnsureEnvVar failed: %v", err)
		}

		content, _ := os.ReadFile(rcFile)
		// Should still have original value, not be duplicated
		count := strings.Count(string(content), "EXISTING_VAR")
		if count != 1 {
			t.Errorf("EXISTING_VAR should appear exactly once, appeared %d times", count)
		}
	})
}

func TestEnsureEnvVarCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".newrc")

	err := EnsureEnvVar(rcFile, "NEW_VAR", "value")
	if err != nil {
		t.Fatalf("EnsureEnvVar failed to create file: %v", err)
	}

	if _, err := os.Stat(rcFile); os.IsNotExist(err) {
		t.Error("EnsureEnvVar should create the file if it doesn't exist")
	}
}

// ============================================================================
// Version Tests
// ============================================================================

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		major         int
		minor         int
		patch         int
		expectErr     bool
	}{
		{"standard version", "1.2.3", 1, 2, 3, false},
		{"with v prefix", "v1.2.3", 1, 2, 3, false},
		{"zeros", "0.0.0", 0, 0, 0, false},
		{"large numbers", "10.20.30", 10, 20, 30, false},
		{"only major.minor", "1.2", 1, 2, 0, false},
		{"only major", "1", 1, 0, 0, false},
		{"with prerelease", "1.2.3-beta", 1, 2, 3, false},
		{"with build", "1.2.3+build", 1, 2, 3, false},
		{"empty string", "", 0, 0, 0, true},
		{"invalid", "not.a.version", 0, 0, 0, true},
		{"negative", "-1.2.3", 0, 0, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			major, minor, patch, err := ParseVersion(tc.version)
			if tc.expectErr {
				if err == nil {
					t.Errorf("ParseVersion(%q) expected error, got nil", tc.version)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseVersion(%q) unexpected error: %v", tc.version, err)
				return
			}
			if major != tc.major || minor != tc.minor || patch != tc.patch {
				t.Errorf("ParseVersion(%q) = (%d, %d, %d), expected (%d, %d, %d)",
					tc.version, major, minor, patch, tc.major, tc.minor, tc.patch)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"equal", "1.2.3", "1.2.3", 0},
		{"v1 newer major", "2.0.0", "1.9.9", 1},
		{"v1 older major", "1.9.9", "2.0.0", -1},
		{"v1 newer minor", "1.3.0", "1.2.9", 1},
		{"v1 older minor", "1.2.9", "1.3.0", -1},
		{"v1 newer patch", "1.2.4", "1.2.3", 1},
		{"v1 older patch", "1.2.3", "1.2.4", -1},
		{"with v prefix", "v1.2.3", "1.2.3", 0},
		{"both with prefix", "v1.2.4", "v1.2.3", 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CompareVersions(tc.v1, tc.v2)
			if result != tc.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, expected %d", tc.v1, tc.v2, result, tc.expected)
			}
		})
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{"newer major", "2.0.0", "1.0.0", true},
		{"newer minor", "1.1.0", "1.0.0", true},
		{"newer patch", "1.0.1", "1.0.0", true},
		{"equal", "1.0.0", "1.0.0", false},
		{"older", "1.0.0", "2.0.0", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNewer(tc.v1, tc.v2)
			if result != tc.expected {
				t.Errorf("IsNewer(%q, %q) = %v, expected %v", tc.v1, tc.v2, result, tc.expected)
			}
		})
	}
}

// ============================================================================
// Exec Tests
// ============================================================================

func TestRunCommand(t *testing.T) {
	t.Run("successful command", func(t *testing.T) {
		stdout, stderr, err := RunCommand("echo", "hello")
		if err != nil {
			t.Fatalf("RunCommand failed: %v", err)
		}
		if !strings.Contains(stdout, "hello") {
			t.Errorf("stdout = %q, expected to contain 'hello'", stdout)
		}
		if stderr != "" {
			t.Errorf("stderr = %q, expected empty", stderr)
		}
	})

	t.Run("command with error", func(t *testing.T) {
		_, _, err := RunCommand("false")
		if err == nil {
			t.Error("RunCommand('false') should return error")
		}
	})

	t.Run("nonexistent command", func(t *testing.T) {
		_, _, err := RunCommand("nonexistent_command_12345")
		if err == nil {
			t.Error("RunCommand with nonexistent command should return error")
		}
	})

	t.Run("command with stderr", func(t *testing.T) {
		// Use shell to redirect to stderr
		if runtime.GOOS != "windows" {
			stdout, stderr, err := RunCommand("sh", "-c", "echo error >&2")
			if err != nil {
				t.Fatalf("RunCommand failed: %v", err)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, expected empty", stdout)
			}
			if !strings.Contains(stderr, "error") {
				t.Errorf("stderr = %q, expected to contain 'error'", stderr)
			}
		}
	})
}

func TestRunCommandInDir(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("command in directory", func(t *testing.T) {
		stdout, _, err := RunCommandInDir(tmpDir, "pwd")
		if err != nil {
			t.Fatalf("RunCommandInDir failed: %v", err)
		}
		// Normalize paths for comparison (handle symlinks like /private on macOS)
		stdoutClean := strings.TrimSpace(stdout)
		if !strings.HasSuffix(stdoutClean, filepath.Base(tmpDir)) {
			t.Errorf("pwd in %q returned %q", tmpDir, stdout)
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		_, _, err := RunCommandInDir("/nonexistent/path/12345", "pwd")
		if err == nil {
			t.Error("RunCommandInDir in nonexistent directory should return error")
		}
	})
}

func TestCommandExists(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"echo exists", "echo", true},
		{"ls exists", "ls", true},
		{"nonexistent", "nonexistent_command_xyz_12345", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CommandExists(tc.command)
			if result != tc.expected {
				t.Errorf("CommandExists(%q) = %v, expected %v", tc.command, result, tc.expected)
			}
		})
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestPathsAreAbsolute(t *testing.T) {
	paths := []struct {
		name string
		path string
	}{
		{"VibeDir", VibeDir()},
		{"MarketplaceDir", MarketplaceDir()},
		{"ProfilesDir", ProfilesDir()},
		{"ConfigFile", ConfigFile()},
		{"ClaudeDir", ClaudeDir()},
		{"ClaudeSettings", ClaudeSettings()},
		{"ClaudePlugins", ClaudePlugins()},
		{"MCPConfig", MCPConfig()},
		{"ClaudeJSON", ClaudeJSON()},
	}

	for _, p := range paths {
		t.Run(p.name, func(t *testing.T) {
			if !filepath.IsAbs(p.path) {
				t.Errorf("%s = %q is not absolute", p.name, p.path)
			}
		})
	}
}

// ============================================================================
// Home Directory Tests
// ============================================================================

func TestIsHomeDirectory(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Create a temp directory for non-home tests
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"home directory exact", home, true},
		{"home with tilde", "~", true},
		{"home with trailing slash", home + "/", true},
		{"temp directory", tmpDir, false},
		{"subdirectory of home", filepath.Join(home, "Documents"), false},
		{"root directory", "/", false},
		{"relative subdirectory notation", home + "/./", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsHomeDirectory(tc.path)
			if result != tc.expected {
				t.Errorf("IsHomeDirectory(%q) = %v, expected %v", tc.path, result, tc.expected)
			}
		})
	}
}

func TestIsHomeDirectoryWithCWD(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer os.Chdir(originalDir)

	// Test with CWD as home
	if err := os.Chdir(home); err != nil {
		t.Skipf("cannot chdir to home: %v", err)
	}

	t.Run("empty string when CWD is home", func(t *testing.T) {
		if !IsHomeDirectory("") {
			t.Error("IsHomeDirectory('') should return true when CWD is home")
		}
	})

	t.Run("dot when CWD is home", func(t *testing.T) {
		if !IsHomeDirectory(".") {
			t.Error("IsHomeDirectory('.') should return true when CWD is home")
		}
	})

	// Test with CWD as non-home
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("cannot chdir to temp: %v", err)
	}

	t.Run("empty string when CWD is not home", func(t *testing.T) {
		if IsHomeDirectory("") {
			t.Error("IsHomeDirectory('') should return false when CWD is not home")
		}
	})

	t.Run("dot when CWD is not home", func(t *testing.T) {
		if IsHomeDirectory(".") {
			t.Error("IsHomeDirectory('.') should return false when CWD is not home")
		}
	})
}

func TestIsInHomeDirectory(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"home directory itself", home, true},
		{"subdirectory of home", filepath.Join(home, "Documents"), true},
		{"deep subdirectory", filepath.Join(home, "a", "b", "c"), true},
		{"home with tilde", "~", true},
		{"tilde subdirectory", "~/.vibe", true},
		{"root directory", "/", false},
		{"tmp directory", "/tmp", false},
		{"usr directory", "/usr/local", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsInHomeDirectory(tc.path)
			if result != tc.expected {
				t.Errorf("IsInHomeDirectory(%q) = %v, expected %v", tc.path, result, tc.expected)
			}
		})
	}
}
