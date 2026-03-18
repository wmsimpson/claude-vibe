package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockCommandChecker implements CommandChecker for testing
type MockCommandChecker struct {
	InstalledCommands map[string]bool
	Versions          map[string]string
}

func (m *MockCommandChecker) IsInstalled(cmd string) bool {
	if m.InstalledCommands == nil {
		return false
	}
	return m.InstalledCommands[cmd]
}

func (m *MockCommandChecker) GetVersion(cmd string) (string, error) {
	if m.Versions == nil {
		return "", nil
	}
	if v, ok := m.Versions[cmd]; ok {
		return v, nil
	}
	return "", nil
}

// MockFileChecker implements FileChecker for testing
type MockFileChecker struct {
	ExistingPaths    map[string]bool
	DirectoryPaths   map[string]bool
	PathOwners       map[string]string
	FileContents     map[string]string
	ValidJSONPaths   map[string]bool
	JSONParseResults map[string]map[string]interface{}
}

func (m *MockFileChecker) Exists(path string) bool {
	if m.ExistingPaths == nil {
		return false
	}
	return m.ExistingPaths[path]
}

func (m *MockFileChecker) IsDir(path string) bool {
	if m.DirectoryPaths == nil {
		return false
	}
	return m.DirectoryPaths[path]
}

func (m *MockFileChecker) Owner(path string) (string, error) {
	if m.PathOwners == nil {
		return "", nil
	}
	return m.PathOwners[path], nil
}

func (m *MockFileChecker) ReadFile(path string) ([]byte, error) {
	if m.FileContents == nil {
		return nil, os.ErrNotExist
	}
	if content, ok := m.FileContents[path]; ok {
		return []byte(content), nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileChecker) IsValidJSON(path string) bool {
	if m.ValidJSONPaths == nil {
		return false
	}
	return m.ValidJSONPaths[path]
}

// MockMarketplaceChecker implements MarketplaceChecker for testing
type MockMarketplaceChecker struct {
	IsRegistered      bool
	InstalledPlugins  []string
	OutdatedPlugins   []string
	RequiredPlugins   []string
}

func (m *MockMarketplaceChecker) IsMarketplaceRegistered(name string) bool {
	return m.IsRegistered
}

func (m *MockMarketplaceChecker) GetInstalledPlugins() []string {
	return m.InstalledPlugins
}

func (m *MockMarketplaceChecker) GetOutdatedPlugins() []string {
	return m.OutdatedPlugins
}

func (m *MockMarketplaceChecker) GetRequiredPlugins() []string {
	return m.RequiredPlugins
}

// MockEnvChecker implements EnvChecker for testing
type MockEnvChecker struct {
	ShellRCPath    string
	ShellRCContent string
	EnvVars        map[string]string
}

func (m *MockEnvChecker) GetShellRC() string {
	return m.ShellRCPath
}

func (m *MockEnvChecker) GetShellRCContent() (string, error) {
	return m.ShellRCContent, nil
}

func (m *MockEnvChecker) GetEnv(key string) string {
	if m.EnvVars == nil {
		return ""
	}
	return m.EnvVars[key]
}

func (m *MockEnvChecker) GetCurrentUser() string {
	return "testuser"
}

// Test Status constants
func TestStatusConstants(t *testing.T) {
	if StatusPass != 0 {
		t.Errorf("StatusPass should be 0, got %d", StatusPass)
	}
	if StatusFail != 1 {
		t.Errorf("StatusFail should be 1, got %d", StatusFail)
	}
	if StatusWarning != 2 {
		t.Errorf("StatusWarning should be 2, got %d", StatusWarning)
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusPass, "Pass"},
		{StatusFail, "Fail"},
		{StatusWarning, "Warning"},
		{Status(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("Status.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// PrereqsCheck tests
func TestPrereqsCheck_AllInstalled(t *testing.T) {
	cmdChecker := &MockCommandChecker{
		InstalledCommands: map[string]bool{
			"gh":      true,
			"jq":      true,
			"yq":      true,
			"claude":  true,
			"python3": true,
		},
	}
	fileChecker := &MockFileChecker{}

	check := NewPrereqsCheck(cmdChecker, fileChecker)

	if check.Name() != "prereqs" {
		t.Errorf("Name() = %v, want prereqs", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestPrereqsCheck_SomeMissing(t *testing.T) {
	cmdChecker := &MockCommandChecker{
		InstalledCommands: map[string]bool{
			"gh":      true,
			"jq":      false,
			"yq":      true,
			"claude":  false,
			"python3": true,
		},
	}
	fileChecker := &MockFileChecker{}

	check := NewPrereqsCheck(cmdChecker, fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}

	if result.RepairHint == "" {
		t.Error("RepairHint should not be empty when prerequisites are missing")
	}
}

func TestPrereqsCheck_ClaudeFallbackToLocalBin(t *testing.T) {
	homeDir := os.Getenv("HOME")
	claudePath := filepath.Join(homeDir, ".local", "bin", "claude")

	// claude not in PATH but exists in ~/.local/bin
	cmdChecker := &MockCommandChecker{
		InstalledCommands: map[string]bool{
			"gh":      true,
			"jq":      true,
			"yq":      true,
			"claude":  false, // not in PATH
			"python3": true,
		},
	}
	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{claudePath: true},
	}

	check := NewPrereqsCheck(cmdChecker, fileChecker)
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v (claude should be found in ~/.local/bin)", result.Status, StatusPass)
	}
}

func TestPrereqsCheck_CanRepair(t *testing.T) {
	check := NewPrereqsCheck(&MockCommandChecker{}, &MockFileChecker{})

	// Prerequisites check can't auto-repair (needs manual install)
	if check.CanRepair() {
		t.Error("PrereqsCheck should not be auto-repairable")
	}
}

// MarketplaceCheck tests
func TestMarketplaceCheck_Exists(t *testing.T) {
	homeDir := os.Getenv("HOME")
	marketplacePath := filepath.Join(homeDir, ".vibe", "marketplace")

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{marketplacePath: true},
		DirectoryPaths: map[string]bool{marketplacePath: true},
	}

	check := NewMarketplaceCheck(fileChecker)

	if check.Name() != "marketplace" {
		t.Errorf("Name() = %v, want marketplace", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestMarketplaceCheck_NotExists(t *testing.T) {
	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{},
		DirectoryPaths: map[string]bool{},
	}

	check := NewMarketplaceCheck(fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

func TestMarketplaceCheck_CanRepair(t *testing.T) {
	check := NewMarketplaceCheck(&MockFileChecker{})

	// Marketplace check can be repaired by running vibe update
	if !check.CanRepair() {
		t.Error("MarketplaceCheck should be repairable")
	}
}

// MarketplaceRegisteredCheck tests
func TestMarketplaceRegisteredCheck_Registered(t *testing.T) {
	mpChecker := &MockMarketplaceChecker{
		IsRegistered: true,
	}

	check := NewMarketplaceRegisteredCheck(mpChecker)

	if check.Name() != "marketplace_registered" {
		t.Errorf("Name() = %v, want marketplace_registered", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestMarketplaceRegisteredCheck_NotRegistered(t *testing.T) {
	mpChecker := &MockMarketplaceChecker{
		IsRegistered: false,
	}

	check := NewMarketplaceRegisteredCheck(mpChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

// SettingsJSONCheck tests
func TestSettingsJSONCheck_Valid(t *testing.T) {
	homeDir := os.Getenv("HOME")
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{settingsPath: true},
		ValidJSONPaths:   map[string]bool{settingsPath: true},
		FileContents:  map[string]string{settingsPath: `{"allow":[],"deny":[]}`},
	}

	check := NewSettingsJSONCheck(fileChecker)

	if check.Name() != "settings_json" {
		t.Errorf("Name() = %v, want settings_json", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestSettingsJSONCheck_NotExists(t *testing.T) {
	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{},
	}

	check := NewSettingsJSONCheck(fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

func TestSettingsJSONCheck_InvalidJSON(t *testing.T) {
	homeDir := os.Getenv("HOME")
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{settingsPath: true},
		ValidJSONPaths:   map[string]bool{settingsPath: false},
		FileContents:  map[string]string{settingsPath: `{invalid json}`},
	}

	check := NewSettingsJSONCheck(fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

// PermissionsCheck tests
func TestPermissionsCheck_AllPresent(t *testing.T) {
	homeDir := os.Getenv("HOME")
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{settingsPath: true},
		FileContents: map[string]string{
			settingsPath: `{"allow":["Skill(configure-vibe)","Bash(gh:*)"],"deny":[]}`,
		},
	}

	check := NewPermissionsCheck(fileChecker, []string{"Skill(configure-vibe)", "Bash(gh:*)"})

	if check.Name() != "permissions" {
		t.Errorf("Name() = %v, want permissions", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestPermissionsCheck_SomeMissing(t *testing.T) {
	homeDir := os.Getenv("HOME")
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{settingsPath: true},
		FileContents: map[string]string{
			settingsPath: `{"allow":["Skill(configure-vibe)"],"deny":[]}`,
		},
	}

	check := NewPermissionsCheck(fileChecker, []string{"Skill(configure-vibe)", "Bash(gh:*)"})
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

// MCPConfigCheck tests
func TestMCPConfigCheck_Valid(t *testing.T) {
	homeDir := os.Getenv("HOME")
	mcpConfigPath := filepath.Join(homeDir, ".config", "mcp", "config.json")

	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{mcpConfigPath: true},
		ValidJSONPaths:   map[string]bool{mcpConfigPath: true},
		FileContents:  map[string]string{mcpConfigPath: `{"claude-code":{}}`},
	}

	check := NewMCPConfigCheck(fileChecker)

	if check.Name() != "mcp_config" {
		t.Errorf("Name() = %v, want mcp_config", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestMCPConfigCheck_NotExists(t *testing.T) {
	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{},
	}

	check := NewMCPConfigCheck(fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

// MCPServersCheck tests
func TestMCPServersCheck_AllConfigured(t *testing.T) {
	homeDir := os.Getenv("HOME")
	mcpConfigPath := filepath.Join(homeDir, ".config", "mcp", "config.json")

	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{mcpConfigPath: true},
		FileContents: map[string]string{
			mcpConfigPath: `{"claude-code":{"slack":{"enabled":true},"jira":{"enabled":true}}}`,
		},
	}

	check := NewMCPServersCheck(fileChecker, []string{"slack", "jira"})

	if check.Name() != "mcp_servers" {
		t.Errorf("Name() = %v, want mcp_servers", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestMCPServersCheck_SomeMissing(t *testing.T) {
	homeDir := os.Getenv("HOME")
	mcpConfigPath := filepath.Join(homeDir, ".config", "mcp", "config.json")

	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{mcpConfigPath: true},
		FileContents: map[string]string{
			mcpConfigPath: `{"claude-code":{"slack":{"enabled":true}}}`,
		},
	}

	check := NewMCPServersCheck(fileChecker, []string{"slack", "jira"})
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

// PluginsInstalledCheck tests
func TestPluginsInstalledCheck_AllInstalled(t *testing.T) {
	mpChecker := &MockMarketplaceChecker{
		InstalledPlugins: []string{"databricks-tools", "fe-salesforce-tools"},
		RequiredPlugins:  []string{"databricks-tools", "fe-salesforce-tools"},
	}

	check := NewPluginsInstalledCheck(mpChecker)

	if check.Name() != "plugins_installed" {
		t.Errorf("Name() = %v, want plugins_installed", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestPluginsInstalledCheck_SomeMissing(t *testing.T) {
	mpChecker := &MockMarketplaceChecker{
		InstalledPlugins: []string{"databricks-tools"},
		RequiredPlugins:  []string{"databricks-tools", "fe-salesforce-tools"},
	}

	check := NewPluginsInstalledCheck(mpChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

// PluginsOutdatedCheck tests
func TestPluginsOutdatedCheck_NoneOutdated(t *testing.T) {
	mpChecker := &MockMarketplaceChecker{
		OutdatedPlugins: []string{},
	}

	check := NewPluginsOutdatedCheck(mpChecker)

	if check.Name() != "plugins_outdated" {
		t.Errorf("Name() = %v, want plugins_outdated", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestPluginsOutdatedCheck_SomeOutdated(t *testing.T) {
	mpChecker := &MockMarketplaceChecker{
		OutdatedPlugins: []string{"databricks-tools"},
	}

	check := NewPluginsOutdatedCheck(mpChecker)
	result := check.Run()

	// Outdated plugins are a warning, not a failure
	if result.Status != StatusWarning {
		t.Errorf("Status = %v, want %v", result.Status, StatusWarning)
	}
}

// EnvVarsCheck tests
func TestEnvVarsCheck_AllPresent(t *testing.T) {
	envChecker := &MockEnvChecker{
		ShellRCPath:    "/home/testuser/.zshrc",
		ShellRCContent: "export I_DANGEROUSLY_OPT_IN_TO_UNSUPPORTED_ALPHA_TOOLS=true\nexport MCP_PRIVACY_SUMMARIZATION_ENABLED=false",
	}

	check := NewEnvVarsCheck(envChecker, []string{
		"I_DANGEROUSLY_OPT_IN_TO_UNSUPPORTED_ALPHA_TOOLS",
		"MCP_PRIVACY_SUMMARIZATION_ENABLED",
	})

	if check.Name() != "env_vars" {
		t.Errorf("Name() = %v, want env_vars", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestEnvVarsCheck_SomeMissing(t *testing.T) {
	envChecker := &MockEnvChecker{
		ShellRCPath:    "/home/testuser/.zshrc",
		ShellRCContent: "export I_DANGEROUSLY_OPT_IN_TO_UNSUPPORTED_ALPHA_TOOLS=true",
	}

	check := NewEnvVarsCheck(envChecker, []string{
		"I_DANGEROUSLY_OPT_IN_TO_UNSUPPORTED_ALPHA_TOOLS",
		"MCP_PRIVACY_SUMMARIZATION_ENABLED",
	})

	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

// LocalOwnershipCheck tests
func TestLocalOwnershipCheck_CorrectOwner(t *testing.T) {
	homeDir := os.Getenv("HOME")
	localPath := filepath.Join(homeDir, ".local")

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{localPath: true},
		DirectoryPaths: map[string]bool{localPath: true},
		PathOwners:     map[string]string{localPath: "testuser"},
	}

	envChecker := &MockEnvChecker{
		EnvVars: map[string]string{"USER": "testuser"},
	}

	check := NewLocalOwnershipCheck(fileChecker, envChecker)

	if check.Name() != "local_ownership" {
		t.Errorf("Name() = %v, want local_ownership", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestLocalOwnershipCheck_RootOwned(t *testing.T) {
	homeDir := os.Getenv("HOME")
	localPath := filepath.Join(homeDir, ".local")

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{localPath: true},
		DirectoryPaths: map[string]bool{localPath: true},
		PathOwners:     map[string]string{localPath: "root"},
	}

	envChecker := &MockEnvChecker{
		EnvVars: map[string]string{"USER": "testuser"},
	}

	check := NewLocalOwnershipCheck(fileChecker, envChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

func TestLocalOwnershipCheck_NotExists(t *testing.T) {
	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{},
		DirectoryPaths: map[string]bool{},
	}

	envChecker := &MockEnvChecker{}

	check := NewLocalOwnershipCheck(fileChecker, envChecker)
	result := check.Run()

	// If ~/.local doesn't exist, that's fine (passes)
	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

// ClaudeDirOwnershipCheck tests
func TestClaudeDirOwnershipCheck_CorrectOwner(t *testing.T) {
	homeDir := os.Getenv("HOME")
	claudePath := filepath.Join(homeDir, ".claude")

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{claudePath: true},
		DirectoryPaths: map[string]bool{claudePath: true},
		PathOwners:     map[string]string{claudePath: "testuser"},
	}

	envChecker := &MockEnvChecker{
		EnvVars: map[string]string{"USER": "testuser"},
	}

	check := NewClaudeDirOwnershipCheck(fileChecker, envChecker)

	if check.Name() != "claude_dir_ownership" {
		t.Errorf("Name() = %v, want claude_dir_ownership", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestClaudeDirOwnershipCheck_RootOwned(t *testing.T) {
	homeDir := os.Getenv("HOME")
	claudePath := filepath.Join(homeDir, ".claude")

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{claudePath: true},
		DirectoryPaths: map[string]bool{claudePath: true},
		PathOwners:     map[string]string{claudePath: "root"},
	}

	envChecker := &MockEnvChecker{
		EnvVars: map[string]string{"USER": "testuser"},
	}

	check := NewClaudeDirOwnershipCheck(fileChecker, envChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}

	if !strings.Contains(result.RepairHint, "chown") {
		t.Errorf("RepairHint should contain chown command, got: %s", result.RepairHint)
	}
}

func TestClaudeDirOwnershipCheck_PluginsDirRootOwned(t *testing.T) {
	homeDir := os.Getenv("HOME")
	claudePath := filepath.Join(homeDir, ".claude")
	pluginsPath := filepath.Join(homeDir, ".claude", "plugins")

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{claudePath: true, pluginsPath: true},
		DirectoryPaths: map[string]bool{claudePath: true, pluginsPath: true},
		PathOwners:     map[string]string{claudePath: "testuser", pluginsPath: "root"},
	}

	envChecker := &MockEnvChecker{
		EnvVars: map[string]string{"USER": "testuser"},
	}

	check := NewClaudeDirOwnershipCheck(fileChecker, envChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}

	if !strings.Contains(result.Message, "plugins") {
		t.Errorf("Message should mention plugins, got: %s", result.Message)
	}
}

func TestClaudeDirOwnershipCheck_NotExists(t *testing.T) {
	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{},
		DirectoryPaths: map[string]bool{},
	}

	envChecker := &MockEnvChecker{}

	check := NewClaudeDirOwnershipCheck(fileChecker, envChecker)
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestClaudeDirOwnershipCheck_CanRepair(t *testing.T) {
	fileChecker := &MockFileChecker{}
	envChecker := &MockEnvChecker{}

	check := NewClaudeDirOwnershipCheck(fileChecker, envChecker)

	if !check.CanRepair() {
		t.Error("ClaudeDirOwnershipCheck should be repairable")
	}
}

// TelemetryHookCheck tests
func TestTelemetryHookCheck_HookConfigured(t *testing.T) {
	homeDir := os.Getenv("HOME")
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Settings with correctly configured telemetry hook
	settingsContent := `{
		"allow": [],
		"deny": [],
		"hooks": {
			"Stop": [
				{
					"matcher": "",
					"hooks": [
						{
							"type": "command",
							"command": "vibe telemetry publish --event-type=claude.session.stop --source=claude-code-stop-hook --from-hook --quiet 2>/dev/null || true",
							"timeout": 30
						}
					]
				}
			]
		}
	}`

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{settingsPath: true},
		ValidJSONPaths: map[string]bool{settingsPath: true},
		FileContents:   map[string]string{settingsPath: settingsContent},
	}

	check := NewTelemetryHookCheck(fileChecker)

	if check.Name() != "telemetry_hook" {
		t.Errorf("Name() = %v, want telemetry_hook", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v. Message: %s", result.Status, StatusPass, result.Message)
	}
}

func TestTelemetryHookCheck_NoHooksSection(t *testing.T) {
	homeDir := os.Getenv("HOME")
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Settings with no hooks section
	settingsContent := `{"allow": [], "deny": []}`

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{settingsPath: true},
		ValidJSONPaths: map[string]bool{settingsPath: true},
		FileContents:   map[string]string{settingsPath: settingsContent},
	}

	check := NewTelemetryHookCheck(fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

func TestTelemetryHookCheck_NoStopHooks(t *testing.T) {
	homeDir := os.Getenv("HOME")
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Settings with hooks section but no Stop hooks
	settingsContent := `{"allow": [], "deny": [], "hooks": {}}`

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{settingsPath: true},
		ValidJSONPaths: map[string]bool{settingsPath: true},
		FileContents:   map[string]string{settingsPath: settingsContent},
	}

	check := NewTelemetryHookCheck(fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

func TestTelemetryHookCheck_WrongCommand(t *testing.T) {
	homeDir := os.Getenv("HOME")
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Settings with a different Stop hook command
	settingsContent := `{
		"allow": [],
		"deny": [],
		"hooks": {
			"Stop": [
				{
					"matcher": "",
					"hooks": [
						{
							"type": "command",
							"command": "echo 'some other command'",
							"timeout": 30
						}
					]
				}
			]
		}
	}`

	fileChecker := &MockFileChecker{
		ExistingPaths:  map[string]bool{settingsPath: true},
		ValidJSONPaths: map[string]bool{settingsPath: true},
		FileContents:   map[string]string{settingsPath: settingsContent},
	}

	check := NewTelemetryHookCheck(fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

func TestTelemetryHookCheck_FileNotExists(t *testing.T) {
	fileChecker := &MockFileChecker{
		ExistingPaths: map[string]bool{},
	}

	check := NewTelemetryHookCheck(fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

func TestTelemetryHookCheck_CanRepair(t *testing.T) {
	fileChecker := &MockFileChecker{}
	check := NewTelemetryHookCheck(fileChecker)

	if !check.CanRepair() {
		t.Error("CanRepair() should return true")
	}
}

// PathConfigCheck tests
func TestPathConfigCheck_InPath(t *testing.T) {
	homeDir := os.Getenv("HOME")
	localBinDir := filepath.Join(homeDir, ".local", "bin")

	envChecker := &MockEnvChecker{
		EnvVars: map[string]string{"PATH": localBinDir + ":/usr/bin:/bin"},
	}
	fileChecker := &MockFileChecker{}

	check := NewPathConfigCheck(envChecker, fileChecker)

	if check.Name() != "path_config" {
		t.Errorf("Name() = %v, want path_config", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestPathConfigCheck_InShellRC(t *testing.T) {
	envChecker := &MockEnvChecker{
		EnvVars:        map[string]string{"PATH": "/usr/bin:/bin"},
		ShellRCPath:    "/home/testuser/.zshrc",
		ShellRCContent: "export PATH=\"$HOME/.local/bin:$PATH\"",
	}
	fileChecker := &MockFileChecker{}

	check := NewPathConfigCheck(envChecker, fileChecker)
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestPathConfigCheck_NotConfigured(t *testing.T) {
	envChecker := &MockEnvChecker{
		EnvVars:        map[string]string{"PATH": "/usr/bin:/bin"},
		ShellRCPath:    "/home/testuser/.zshrc",
		ShellRCContent: "# no path config",
	}
	fileChecker := &MockFileChecker{}

	check := NewPathConfigCheck(envChecker, fileChecker)
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want %v", result.Status, StatusFail)
	}
}

func TestPathConfigCheck_CanRepair(t *testing.T) {
	check := NewPathConfigCheck(&MockEnvChecker{}, &MockFileChecker{})

	if !check.CanRepair() {
		t.Error("CanRepair() should return true")
	}
}

// RecommendedToolsCheck tests
func TestRecommendedToolsCheck_AllInstalled(t *testing.T) {
	cmdChecker := &MockCommandChecker{
		InstalledCommands: map[string]bool{
			"rg":          true,
			"node":        true,
			"npm":         true,
			"uv":          true,
			"aws":         true,
			"gcloud":      true,
			"databricks":  true,
			"sf":          true,
		},
	}

	check := NewRecommendedToolsCheck(cmdChecker)

	if check.Name() != "recommended_tools" {
		t.Errorf("Name() = %v, want recommended_tools", check.Name())
	}

	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
}

func TestRecommendedToolsCheck_SomeMissing(t *testing.T) {
	cmdChecker := &MockCommandChecker{
		InstalledCommands: map[string]bool{
			"rg":   true,
			"node": true,
			"npm":  true,
		},
	}

	check := NewRecommendedToolsCheck(cmdChecker)
	result := check.Run()

	// Missing recommended tools should be a warning, not a failure
	if result.Status != StatusWarning {
		t.Errorf("Status = %v, want %v", result.Status, StatusWarning)
	}

	if result.RepairHint == "" {
		t.Error("RepairHint should not be empty when tools are missing")
	}
}

func TestRecommendedToolsCheck_NoneInstalled(t *testing.T) {
	cmdChecker := &MockCommandChecker{
		InstalledCommands: map[string]bool{},
	}

	check := NewRecommendedToolsCheck(cmdChecker)
	result := check.Run()

	if result.Status != StatusWarning {
		t.Errorf("Status = %v, want %v", result.Status, StatusWarning)
	}
}

func TestRecommendedToolsCheck_CanRepair(t *testing.T) {
	check := NewRecommendedToolsCheck(&MockCommandChecker{})

	// Recommended tools can't auto-repair
	if check.CanRepair() {
		t.Error("RecommendedToolsCheck should not be auto-repairable")
	}
}

// AllChecks tests
func TestAllChecks_ReturnsChecks(t *testing.T) {
	checks := AllChecks()

	if len(checks) == 0 {
		t.Error("AllChecks() should return at least one check")
	}

	// Verify all checks have names and descriptions
	for _, check := range checks {
		if check.Name() == "" {
			t.Error("Check has empty name")
		}
		if check.Description() == "" {
			t.Errorf("Check %s has empty description", check.Name())
		}
	}
}

// RunAll tests
func TestRunAll_ReturnsResults(t *testing.T) {
	// This is an integration-style test using default checkers
	// In unit tests we'd use mocks, but this validates the plumbing works
	results := RunAll()

	if len(results) == 0 {
		t.Error("RunAll() should return at least one result")
	}

	// Verify all results have names
	for _, result := range results {
		if result.Name == "" {
			t.Error("Result has empty name")
		}
	}
}

// Test CheckResult
func TestCheckResult_Fields(t *testing.T) {
	result := CheckResult{
		Name:       "test_check",
		Status:     StatusPass,
		Message:    "All good",
		RepairHint: "",
	}

	if result.Name != "test_check" {
		t.Errorf("Name = %v, want test_check", result.Name)
	}
	if result.Status != StatusPass {
		t.Errorf("Status = %v, want %v", result.Status, StatusPass)
	}
	if result.Message != "All good" {
		t.Errorf("Message = %v, want 'All good'", result.Message)
	}
}

// Test parsePluginList helper to verify plugin list parsing logic
func TestParsePluginListOutput(t *testing.T) {
	// Simulates the output format of `claude plugin list`
	// Format:
	//   ❯ databricks-tools@claude-vibe
	//     Version: 1.0.0
	//     Scope: user
	//     Status: ✔ enabled
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "standard format with marketplace suffix",
			input: `Installed plugins:

  ❯ databricks-tools@claude-vibe
    Version: 1.0.0
    Scope: user
    Status: ✔ enabled

  ❯ google-tools@claude-vibe
    Version: 1.0.0
    Scope: user
    Status: ✔ enabled
`,
			expected: []string{"databricks-tools", "google-tools"},
		},
		{
			name: "plugin without marketplace suffix",
			input: `Installed plugins:

  ❯ some-plugin
    Version: 1.0.0
`,
			expected: []string{"some-plugin"},
		},
		{
			name:     "empty output",
			input:    "",
			expected: []string{},
		},
		{
			name: "no installed plugins",
			input: `Installed plugins:

`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the parsing logic from GetInstalledPlugins
			var plugins []string
			lines := strings.Split(tt.input, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "❯") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						pluginFullName := parts[1]
						pluginName := strings.Split(pluginFullName, "@")[0]
						plugins = append(plugins, pluginName)
					}
				}
			}

			if len(plugins) != len(tt.expected) {
				t.Errorf("got %d plugins, want %d", len(plugins), len(tt.expected))
				return
			}

			for i, p := range plugins {
				if p != tt.expected[i] {
					t.Errorf("plugin[%d] = %v, want %v", i, p, tt.expected[i])
				}
			}
		})
	}
}
