// Package doctor provides health checks and diagnostics for the vibe CLI.
package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// Status represents the result status of a health check.
type Status int

const (
	// StatusPass indicates the check passed.
	StatusPass Status = iota
	// StatusFail indicates the check failed.
	StatusFail
	// StatusWarning indicates the check passed with warnings.
	StatusWarning
)

// String returns a human-readable string for the status.
func (s Status) String() string {
	switch s {
	case StatusPass:
		return "Pass"
	case StatusFail:
		return "Fail"
	case StatusWarning:
		return "Warning"
	default:
		return "Unknown"
	}
}

// CheckResult contains the outcome of running a health check.
type CheckResult struct {
	Name       string
	Status     Status
	Message    string
	RepairHint string
}

// Check defines the interface for health checks.
type Check interface {
	// Name returns the identifier for this check.
	Name() string
	// Description returns a human-readable description.
	Description() string
	// Run executes the check and returns the result.
	Run() CheckResult
	// CanRepair indicates if this check supports auto-repair.
	CanRepair() bool
	// Repair attempts to fix the issue. Returns error if repair fails.
	Repair() error
}

// CommandChecker abstracts command existence and version checking.
type CommandChecker interface {
	IsInstalled(cmd string) bool
	GetVersion(cmd string) (string, error)
}

// FileChecker abstracts file system operations.
type FileChecker interface {
	Exists(path string) bool
	IsDir(path string) bool
	Owner(path string) (string, error)
	ReadFile(path string) ([]byte, error)
	IsValidJSON(path string) bool
}

// MarketplaceChecker abstracts marketplace operations.
type MarketplaceChecker interface {
	IsMarketplaceRegistered(name string) bool
	GetInstalledPlugins() []string
	GetOutdatedPlugins() []string
	GetRequiredPlugins() []string
}

// EnvChecker abstracts environment operations.
type EnvChecker interface {
	GetShellRC() string
	GetShellRCContent() (string, error)
	GetEnv(key string) string
	GetCurrentUser() string
}

// DefaultCommandChecker implements CommandChecker using real system commands.
type DefaultCommandChecker struct{}

// IsInstalled checks if a command is available in PATH.
func (d *DefaultCommandChecker) IsInstalled(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetVersion attempts to get the version of a command.
func (d *DefaultCommandChecker) GetVersion(cmd string) (string, error) {
	// Common version flags to try
	versionFlags := []string{"--version", "-version", "version"}

	for _, flag := range versionFlags {
		out, err := exec.Command(cmd, flag).Output()
		if err == nil {
			// Return first line of output
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0]), nil
			}
		}
	}

	return "", fmt.Errorf("could not determine version for %s", cmd)
}

// DefaultFileChecker implements FileChecker using real file system.
type DefaultFileChecker struct{}

// Exists checks if a path exists.
func (d *DefaultFileChecker) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory.
func (d *DefaultFileChecker) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Owner returns the username of the file owner.
func (d *DefaultFileChecker) Owner(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("could not get file owner")
	}

	u, err := user.LookupId(fmt.Sprintf("%d", stat.Uid))
	if err != nil {
		return "", err
	}

	return u.Username, nil
}

// ReadFile reads the contents of a file.
func (d *DefaultFileChecker) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// IsValidJSON checks if a file contains valid JSON.
func (d *DefaultFileChecker) IsValidJSON(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var js interface{}
	return json.Unmarshal(data, &js) == nil
}

// DefaultMarketplaceChecker implements MarketplaceChecker using claude CLI.
type DefaultMarketplaceChecker struct{}

// IsMarketplaceRegistered checks if a marketplace is registered.
func (d *DefaultMarketplaceChecker) IsMarketplaceRegistered(name string) bool {
	out, err := exec.Command("claude", "plugin", "marketplace", "list").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), name)
}

// GetInstalledPlugins returns a list of installed plugins.
func (d *DefaultMarketplaceChecker) GetInstalledPlugins() []string {
	out, err := exec.Command("claude", "plugin", "list").Output()
	if err != nil {
		return nil
	}

	var plugins []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines starting with ❯ which indicate plugin entries
		// Format: "❯ plugin-name@marketplace" or "❯ plugin-name"
		if strings.HasPrefix(line, "❯") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pluginFullName := parts[1] // e.g., "databricks-tools@claude-vibe"
				// Extract just the plugin name (before @)
				pluginName := strings.Split(pluginFullName, "@")[0]
				plugins = append(plugins, pluginName)
			}
		}
	}
	return plugins
}

// GetOutdatedPlugins returns a list of outdated plugins.
func (d *DefaultMarketplaceChecker) GetOutdatedPlugins() []string {
	// This would need to compare installed vs marketplace versions
	// For now, return empty (implementation depends on marketplace format)
	return []string{}
}

// GetRequiredPlugins returns the list of plugins that should be installed.
func (d *DefaultMarketplaceChecker) GetRequiredPlugins() []string {
	return []string{
		"databricks-tools",
		"google-tools",
		"specialized-agents",
		"vibe-setup",
		"mcp-servers",
		"jira-tools",
		"workflows",
	}
}

// DefaultEnvChecker implements EnvChecker using real environment.
type DefaultEnvChecker struct{}

// GetShellRC returns the path to the user's shell rc file.
func (d *DefaultEnvChecker) GetShellRC() string {
	shell := os.Getenv("SHELL")
	homeDir, _ := os.UserHomeDir()

	switch filepath.Base(shell) {
	case "zsh":
		return filepath.Join(homeDir, ".zshrc")
	case "bash":
		return filepath.Join(homeDir, ".bashrc")
	default:
		return filepath.Join(homeDir, ".bashrc")
	}
}

// GetShellRCContent returns the contents of the shell rc file.
func (d *DefaultEnvChecker) GetShellRCContent() (string, error) {
	content, err := os.ReadFile(d.GetShellRC())
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetEnv returns the value of an environment variable.
func (d *DefaultEnvChecker) GetEnv(key string) string {
	return os.Getenv(key)
}

// GetCurrentUser returns the current username.
func (d *DefaultEnvChecker) GetCurrentUser() string {
	u, err := user.Current()
	if err != nil {
		return os.Getenv("USER")
	}
	return u.Username
}

// PrereqsCheck verifies required tools are installed.
type PrereqsCheck struct {
	cmdChecker  CommandChecker
	fileChecker FileChecker
	required    []string
	localBinDir string
}

// NewPrereqsCheck creates a new prerequisites check.
func NewPrereqsCheck(cmdChecker CommandChecker, fileChecker FileChecker) *PrereqsCheck {
	homeDir, _ := os.UserHomeDir()
	return &PrereqsCheck{
		cmdChecker:  cmdChecker,
		fileChecker: fileChecker,
		required:    []string{"gh", "jq", "yq", "claude", "python3"},
		localBinDir: filepath.Join(homeDir, ".local", "bin"),
	}
}

// Name returns the check identifier.
func (c *PrereqsCheck) Name() string { return "prereqs" }

// Description returns a human-readable description.
func (c *PrereqsCheck) Description() string {
	return "Verify required tools (gh, jq, yq, claude, python3) are installed"
}

// Run executes the check.
func (c *PrereqsCheck) Run() CheckResult {
	var missing []string

	for _, cmd := range c.required {
		if !c.cmdChecker.IsInstalled(cmd) {
			// Special case for claude: also check ~/.local/bin/claude
			if cmd == "claude" {
				claudePath := filepath.Join(c.localBinDir, "claude")
				if c.fileChecker.Exists(claudePath) {
					continue // claude found in ~/.local/bin
				}
			}
			missing = append(missing, cmd)
		}
	}

	if len(missing) == 0 {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "All required tools are installed",
		}
	}

	// Build hint based on what's missing
	hint := fmt.Sprintf("Install missing tools: %s (via your package manager or 'vibe install')", strings.Join(missing, ", "))
	for _, m := range missing {
		if m == "claude" {
			hint = "Install claude with: curl -fsSL https://claude.ai/install.sh | bash"
			break
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    fmt.Sprintf("Missing tools: %s", strings.Join(missing, ", ")),
		RepairHint: hint,
	}
}

// CanRepair returns false as tools need manual installation.
func (c *PrereqsCheck) CanRepair() bool { return false }

// Repair is not supported for prerequisites.
func (c *PrereqsCheck) Repair() error {
	return fmt.Errorf("cannot auto-repair: please install missing tools manually")
}

// RecommendedToolsCheck verifies recommended tools are installed (any install method).
type RecommendedToolsCheck struct {
	cmdChecker CommandChecker
	tools      []recommendedTool
}

type recommendedTool struct {
	Command     string // command name to check in PATH
	Description string // human-readable description
}

// NewRecommendedToolsCheck creates a new recommended tools check.
func NewRecommendedToolsCheck(cmdChecker CommandChecker) *RecommendedToolsCheck {
	return &RecommendedToolsCheck{
		cmdChecker: cmdChecker,
		tools: []recommendedTool{
			{"rg", "ripgrep (fast file search)"},
			{"node", "Node.js runtime"},
			{"npm", "Node.js package manager"},
			{"uv", "Python package manager"},
			{"aws", "AWS CLI"},
			{"gcloud", "Google Cloud SDK"},
			{"databricks", "Databricks CLI"},
			{"sf", "Salesforce CLI"},
		},
	}
}

// Name returns the check identifier.
func (c *RecommendedToolsCheck) Name() string { return "recommended_tools" }

// Description returns a human-readable description.
func (c *RecommendedToolsCheck) Description() string {
	return "Check for recommended tools (rg, node, uv, aws, gcloud, databricks, sf)"
}

// Run executes the check.
func (c *RecommendedToolsCheck) Run() CheckResult {
	var missing []string

	for _, tool := range c.tools {
		if !c.cmdChecker.IsInstalled(tool.Command) {
			missing = append(missing, tool.Command)
		}
	}

	if len(missing) == 0 {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "All recommended tools are installed",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusWarning,
		Message:    fmt.Sprintf("Missing recommended tools: %s", strings.Join(missing, ", ")),
		RepairHint: "Install missing tools using your preferred package manager, or run 'vibe install' to install via Homebrew",
	}
}

// CanRepair returns false as tools need manual installation.
func (c *RecommendedToolsCheck) CanRepair() bool { return false }

// Repair is not supported for recommended tools.
func (c *RecommendedToolsCheck) Repair() error {
	return fmt.Errorf("cannot auto-repair: install missing tools using your preferred package manager")
}

// MarketplaceCheck verifies the marketplace directory exists.
type MarketplaceCheck struct {
	fileChecker FileChecker
	path        string
}

// NewMarketplaceCheck creates a new marketplace check.
func NewMarketplaceCheck(fileChecker FileChecker) *MarketplaceCheck {
	homeDir, _ := os.UserHomeDir()
	return &MarketplaceCheck{
		fileChecker: fileChecker,
		path:        filepath.Join(homeDir, ".vibe", "marketplace"),
	}
}

// Name returns the check identifier.
func (c *MarketplaceCheck) Name() string { return "marketplace" }

// Description returns a human-readable description.
func (c *MarketplaceCheck) Description() string {
	return "Verify ~/.vibe/marketplace directory exists"
}

// Run executes the check.
func (c *MarketplaceCheck) Run() CheckResult {
	if c.fileChecker.Exists(c.path) && c.fileChecker.IsDir(c.path) {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "Marketplace directory exists",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    "Marketplace directory not found",
		RepairHint: "Run 'vibe update' to download the marketplace",
	}
}

// CanRepair returns true as we can download the marketplace.
func (c *MarketplaceCheck) CanRepair() bool { return true }

// Repair downloads the marketplace.
func (c *MarketplaceCheck) Repair() error {
	// This would trigger a marketplace download
	// Implementation depends on marketplace package
	return fmt.Errorf("repair not implemented: run 'vibe update' manually")
}

// MarketplaceRegisteredCheck verifies the marketplace is registered with Claude.
type MarketplaceRegisteredCheck struct {
	mpChecker MarketplaceChecker
	name      string
}

// NewMarketplaceRegisteredCheck creates a new marketplace registration check.
func NewMarketplaceRegisteredCheck(mpChecker MarketplaceChecker) *MarketplaceRegisteredCheck {
	return &MarketplaceRegisteredCheck{
		mpChecker: mpChecker,
		name:      "claude-vibe",
	}
}

// Name returns the check identifier.
func (c *MarketplaceRegisteredCheck) Name() string { return "marketplace_registered" }

// Description returns a human-readable description.
func (c *MarketplaceRegisteredCheck) Description() string {
	return "Verify claude-vibe marketplace is registered with Claude Code"
}

// Run executes the check.
func (c *MarketplaceRegisteredCheck) Run() CheckResult {
	if c.mpChecker.IsMarketplaceRegistered(c.name) {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "Marketplace is registered",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    "Marketplace not registered with Claude Code",
		RepairHint: "Run 'claude plugin marketplace add ~/.vibe/marketplace'",
	}
}

// CanRepair returns true as we can register the marketplace.
func (c *MarketplaceRegisteredCheck) CanRepair() bool { return true }

// Repair registers the marketplace.
func (c *MarketplaceRegisteredCheck) Repair() error {
	homeDir, _ := os.UserHomeDir()
	marketplacePath := filepath.Join(homeDir, ".vibe", "marketplace")

	cmd := exec.Command("claude", "plugin", "marketplace", "add", marketplacePath)
	return cmd.Run()
}

// SettingsJSONCheck verifies settings.json exists and is valid.
type SettingsJSONCheck struct {
	fileChecker FileChecker
	path        string
}

// NewSettingsJSONCheck creates a new settings.json check.
func NewSettingsJSONCheck(fileChecker FileChecker) *SettingsJSONCheck {
	homeDir, _ := os.UserHomeDir()
	return &SettingsJSONCheck{
		fileChecker: fileChecker,
		path:        filepath.Join(homeDir, ".claude", "settings.json"),
	}
}

// Name returns the check identifier.
func (c *SettingsJSONCheck) Name() string { return "settings_json" }

// Description returns a human-readable description.
func (c *SettingsJSONCheck) Description() string {
	return "Verify ~/.claude/settings.json exists and is valid JSON"
}

// Run executes the check.
func (c *SettingsJSONCheck) Run() CheckResult {
	if !c.fileChecker.Exists(c.path) {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "settings.json does not exist",
			RepairHint: "Create default settings.json",
		}
	}

	if !c.fileChecker.IsValidJSON(c.path) {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "settings.json contains invalid JSON",
			RepairHint: "Fix or recreate settings.json",
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "settings.json is valid",
	}
}

// CanRepair returns true as we can create default settings.
func (c *SettingsJSONCheck) CanRepair() bool { return true }

// Repair creates or fixes settings.json.
func (c *SettingsJSONCheck) Repair() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	defaultSettings := `{"allow":[],"deny":[]}`
	return os.WriteFile(c.path, []byte(defaultSettings), 0644)
}

// PermissionsCheck verifies required permissions are present.
type PermissionsCheck struct {
	fileChecker FileChecker
	required    []string
	path        string
}

// NewPermissionsCheck creates a new permissions check.
func NewPermissionsCheck(fileChecker FileChecker, required []string) *PermissionsCheck {
	homeDir, _ := os.UserHomeDir()
	return &PermissionsCheck{
		fileChecker: fileChecker,
		required:    required,
		path:        filepath.Join(homeDir, ".claude", "settings.json"),
	}
}

// Name returns the check identifier.
func (c *PermissionsCheck) Name() string { return "permissions" }

// Description returns a human-readable description.
func (c *PermissionsCheck) Description() string {
	return "Verify required permissions are present in settings.json"
}

// Run executes the check.
func (c *PermissionsCheck) Run() CheckResult {
	data, err := c.fileChecker.ReadFile(c.path)
	if err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "Could not read settings.json",
			RepairHint: "Run 'vibe update' to sync permissions",
		}
	}

	var settings struct {
		Allow []string `json:"allow"`
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "Could not parse settings.json",
			RepairHint: "Run 'vibe update' to sync permissions",
		}
	}

	allowSet := make(map[string]bool)
	for _, perm := range settings.Allow {
		allowSet[perm] = true
	}

	var missing []string
	for _, req := range c.required {
		if !allowSet[req] {
			missing = append(missing, req)
		}
	}

	if len(missing) == 0 {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "All required permissions present",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    fmt.Sprintf("Missing permissions: %d", len(missing)),
		RepairHint: "Run 'vibe update' to sync permissions",
	}
}

// CanRepair returns true as we can sync permissions.
func (c *PermissionsCheck) CanRepair() bool { return true }

// Repair syncs permissions from marketplace.
func (c *PermissionsCheck) Repair() error {
	// This would trigger a permission sync
	// Implementation depends on marketplace package
	return fmt.Errorf("repair not implemented: run 'vibe update' manually")
}

// MCPConfigCheck verifies MCP config exists and is valid.
type MCPConfigCheck struct {
	fileChecker FileChecker
	path        string
}

// NewMCPConfigCheck creates a new MCP config check.
func NewMCPConfigCheck(fileChecker FileChecker) *MCPConfigCheck {
	homeDir, _ := os.UserHomeDir()
	return &MCPConfigCheck{
		fileChecker: fileChecker,
		path:        filepath.Join(homeDir, ".config", "mcp", "config.json"),
	}
}

// Name returns the check identifier.
func (c *MCPConfigCheck) Name() string { return "mcp_config" }

// Description returns a human-readable description.
func (c *MCPConfigCheck) Description() string {
	return "Verify ~/.config/mcp/config.json exists and is valid"
}

// Run executes the check.
func (c *MCPConfigCheck) Run() CheckResult {
	if !c.fileChecker.Exists(c.path) {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "MCP config does not exist",
			RepairHint: "Create default MCP config",
		}
	}

	if !c.fileChecker.IsValidJSON(c.path) {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "MCP config contains invalid JSON",
			RepairHint: "Fix or recreate MCP config",
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "MCP config is valid",
	}
}

// CanRepair returns true as we can create default config.
func (c *MCPConfigCheck) CanRepair() bool { return true }

// Repair creates or fixes MCP config.
func (c *MCPConfigCheck) Repair() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	defaultConfig := `{"claude-code":{}}`
	return os.WriteFile(c.path, []byte(defaultConfig), 0644)
}

// MCPServersCheck verifies required MCP servers are configured.
type MCPServersCheck struct {
	fileChecker FileChecker
	required    []string
	path        string
}

// NewMCPServersCheck creates a new MCP servers check.
func NewMCPServersCheck(fileChecker FileChecker, required []string) *MCPServersCheck {
	homeDir, _ := os.UserHomeDir()
	return &MCPServersCheck{
		fileChecker: fileChecker,
		required:    required,
		path:        filepath.Join(homeDir, ".config", "mcp", "config.json"),
	}
}

// Name returns the check identifier.
func (c *MCPServersCheck) Name() string { return "mcp_servers" }

// Description returns a human-readable description.
func (c *MCPServersCheck) Description() string {
	return "Verify required MCP servers are configured"
}

// Run executes the check.
func (c *MCPServersCheck) Run() CheckResult {
	data, err := c.fileChecker.ReadFile(c.path)
	if err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "Could not read MCP config",
			RepairHint: "Run 'vibe update' to sync MCP servers",
		}
	}

	var config struct {
		ClaudeCode map[string]interface{} `json:"claude-code"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "Could not parse MCP config",
			RepairHint: "Run 'vibe update' to sync MCP servers",
		}
	}

	var missing []string
	for _, server := range c.required {
		if _, exists := config.ClaudeCode[server]; !exists {
			missing = append(missing, server)
		}
	}

	if len(missing) == 0 {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "All required MCP servers configured",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    fmt.Sprintf("Missing MCP servers: %s", strings.Join(missing, ", ")),
		RepairHint: "Run 'vibe update' to sync MCP servers",
	}
}

// CanRepair returns true as we can sync MCP servers.
func (c *MCPServersCheck) CanRepair() bool { return true }

// Repair syncs MCP servers from marketplace.
func (c *MCPServersCheck) Repair() error {
	return fmt.Errorf("repair not implemented: run 'vibe update' manually")
}

// PluginsInstalledCheck verifies required plugins are installed.
type PluginsInstalledCheck struct {
	mpChecker MarketplaceChecker
}

// NewPluginsInstalledCheck creates a new plugins installed check.
func NewPluginsInstalledCheck(mpChecker MarketplaceChecker) *PluginsInstalledCheck {
	return &PluginsInstalledCheck{
		mpChecker: mpChecker,
	}
}

// Name returns the check identifier.
func (c *PluginsInstalledCheck) Name() string { return "plugins_installed" }

// Description returns a human-readable description.
func (c *PluginsInstalledCheck) Description() string {
	return "Verify required plugins are installed"
}

// Run executes the check.
func (c *PluginsInstalledCheck) Run() CheckResult {
	installed := make(map[string]bool)
	for _, p := range c.mpChecker.GetInstalledPlugins() {
		installed[p] = true
	}

	var missing []string
	for _, req := range c.mpChecker.GetRequiredPlugins() {
		if !installed[req] {
			missing = append(missing, req)
		}
	}

	if len(missing) == 0 {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "All required plugins installed",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    fmt.Sprintf("Missing plugins: %s", strings.Join(missing, ", ")),
		RepairHint: "Run 'vibe update' to install missing plugins",
	}
}

// CanRepair returns true as we can install plugins.
func (c *PluginsInstalledCheck) CanRepair() bool { return true }

// Repair installs missing plugins.
func (c *PluginsInstalledCheck) Repair() error {
	installed := make(map[string]bool)
	for _, p := range c.mpChecker.GetInstalledPlugins() {
		installed[p] = true
	}

	for _, req := range c.mpChecker.GetRequiredPlugins() {
		if !installed[req] {
			cmd := exec.Command("claude", "plugin", "install", req+"@claude-vibe")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install %s: %w", req, err)
			}
		}
	}

	return nil
}

// PluginsOutdatedCheck verifies no plugins are outdated.
type PluginsOutdatedCheck struct {
	mpChecker MarketplaceChecker
}

// NewPluginsOutdatedCheck creates a new plugins outdated check.
func NewPluginsOutdatedCheck(mpChecker MarketplaceChecker) *PluginsOutdatedCheck {
	return &PluginsOutdatedCheck{
		mpChecker: mpChecker,
	}
}

// Name returns the check identifier.
func (c *PluginsOutdatedCheck) Name() string { return "plugins_outdated" }

// Description returns a human-readable description.
func (c *PluginsOutdatedCheck) Description() string {
	return "Check for outdated plugins"
}

// Run executes the check.
func (c *PluginsOutdatedCheck) Run() CheckResult {
	outdated := c.mpChecker.GetOutdatedPlugins()

	if len(outdated) == 0 {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "All plugins up to date",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusWarning,
		Message:    fmt.Sprintf("Outdated plugins: %s", strings.Join(outdated, ", ")),
		RepairHint: "Run 'vibe update' to update plugins",
	}
}

// CanRepair returns true as we can update plugins.
func (c *PluginsOutdatedCheck) CanRepair() bool { return true }

// Repair updates outdated plugins.
func (c *PluginsOutdatedCheck) Repair() error {
	for _, plugin := range c.mpChecker.GetOutdatedPlugins() {
		cmd := exec.Command("claude", "plugin", "install", plugin+"@claude-vibe")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update %s: %w", plugin, err)
		}
	}
	return nil
}

// EnvVarsCheck verifies required environment variables in shell rc.
type EnvVarsCheck struct {
	envChecker EnvChecker
	required   []string
}

// NewEnvVarsCheck creates a new environment variables check.
func NewEnvVarsCheck(envChecker EnvChecker, required []string) *EnvVarsCheck {
	return &EnvVarsCheck{
		envChecker: envChecker,
		required:   required,
	}
}

// Name returns the check identifier.
func (c *EnvVarsCheck) Name() string { return "env_vars" }

// Description returns a human-readable description.
func (c *EnvVarsCheck) Description() string {
	return "Verify required environment variables are configured"
}

// Run executes the check.
func (c *EnvVarsCheck) Run() CheckResult {
	content, err := c.envChecker.GetShellRCContent()
	if err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "Could not read shell rc file",
			RepairHint: "Check shell configuration",
		}
	}

	var missing []string
	for _, envVar := range c.required {
		if !strings.Contains(content, envVar) {
			missing = append(missing, envVar)
		}
	}

	if len(missing) == 0 {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "All required environment variables configured",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    fmt.Sprintf("Missing env vars: %s", strings.Join(missing, ", ")),
		RepairHint: "Run 'vibe update' to add missing environment variables",
	}
}

// CanRepair returns true as we can add env vars.
func (c *EnvVarsCheck) CanRepair() bool { return true }

// Repair adds missing environment variables to shell rc.
func (c *EnvVarsCheck) Repair() error {
	rcPath := c.envChecker.GetShellRC()
	content, err := c.envChecker.GetShellRCContent()
	if err != nil {
		content = ""
	}

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	envDefaults := map[string]string{
		// Add default values for required env vars here
	}

	for _, envVar := range c.required {
		if !strings.Contains(content, envVar) {
			value := envDefaults[envVar]
			if value == "" {
				value = "true"
			}
			line := fmt.Sprintf("\nexport %s=%s\n", envVar, value)
			if _, err := f.WriteString(line); err != nil {
				return err
			}
		}
	}

	return nil
}

// LocalOwnershipCheck verifies ~/.local is not owned by root.
type LocalOwnershipCheck struct {
	fileChecker FileChecker
	envChecker  EnvChecker
	path        string
}

// NewLocalOwnershipCheck creates a new local ownership check.
func NewLocalOwnershipCheck(fileChecker FileChecker, envChecker EnvChecker) *LocalOwnershipCheck {
	homeDir, _ := os.UserHomeDir()
	return &LocalOwnershipCheck{
		fileChecker: fileChecker,
		envChecker:  envChecker,
		path:        filepath.Join(homeDir, ".local"),
	}
}

// Name returns the check identifier.
func (c *LocalOwnershipCheck) Name() string { return "local_ownership" }

// Description returns a human-readable description.
func (c *LocalOwnershipCheck) Description() string {
	return "Verify ~/.local is not owned by root"
}

// Run executes the check.
func (c *LocalOwnershipCheck) Run() CheckResult {
	if !c.fileChecker.Exists(c.path) {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "~/.local does not exist (OK)",
		}
	}

	owner, err := c.fileChecker.Owner(c.path)
	if err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusWarning,
			Message:    "Could not determine owner of ~/.local",
			RepairHint: "Check ~/.local ownership manually",
		}
	}

	if owner == "root" {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "~/.local is owned by root",
			RepairHint: fmt.Sprintf("Run: sudo chown -R %s:staff ~/.local", c.envChecker.GetCurrentUser()),
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("~/.local is owned by %s", owner),
	}
}

// CanRepair returns true as we can fix ownership.
func (c *LocalOwnershipCheck) CanRepair() bool { return true }

// Repair fixes the ownership of ~/.local.
func (c *LocalOwnershipCheck) Repair() error {
	user := c.envChecker.GetCurrentUser()
	group := "staff"
	if runtime.GOOS == "linux" {
		group = user
	}

	cmd := exec.Command("sudo", "-p", "Password required to fix ~/.local ownership: ", "chown", "-R", fmt.Sprintf("%s:%s", user, group), c.path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RootOwnedCachesCheck verifies no critical directories are owned by root.
// This catches common issues from running vibe with sudo.
type RootOwnedCachesCheck struct {
	fileChecker FileChecker
	envChecker  EnvChecker
	paths       []string
	homeDir     string
}

// NewRootOwnedCachesCheck creates a new root-owned caches check.
func NewRootOwnedCachesCheck(fileChecker FileChecker, envChecker EnvChecker) *RootOwnedCachesCheck {
	homeDir, _ := os.UserHomeDir()
	return &RootOwnedCachesCheck{
		fileChecker: fileChecker,
		envChecker:  envChecker,
		homeDir:     homeDir,
		paths: []string{
			".pex",
			".npm",
			".local/state/mcp-servers",
			".kube",
		},
	}
}

// Name returns the check identifier.
func (c *RootOwnedCachesCheck) Name() string { return "root_owned_caches" }

// Description returns a human-readable description.
func (c *RootOwnedCachesCheck) Description() string {
	return "Check for root-owned cache directories (common issue from prior sudo installs)"
}

// hasRootOwnedContents checks if a directory contains any root-owned files.
func (c *RootOwnedCachesCheck) hasRootOwnedContents(dir string) bool {
	// Use find to check for root-owned files inside the directory
	cmd := exec.Command("find", dir, "-user", "root", "-maxdepth", "3", "-print", "-quit")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// Run executes the check.
func (c *RootOwnedCachesCheck) Run() CheckResult {
	var rootOwned []string

	for _, p := range c.paths {
		fullPath := filepath.Join(c.homeDir, p)
		if !c.fileChecker.Exists(fullPath) {
			continue
		}

		owner, err := c.fileChecker.Owner(fullPath)
		if err != nil {
			continue
		}

		if owner == "root" {
			rootOwned = append(rootOwned, p)
		} else if c.fileChecker.IsDir(fullPath) && c.hasRootOwnedContents(fullPath) {
			// Directory is user-owned but contains root-owned files inside
			rootOwned = append(rootOwned, p+" (contains root-owned files)")
		}
	}

	// Also check /tmp/cached_hcvault_token (absolute path, not relative to home)
	vaultTokenPath := "/tmp/cached_hcvault_token"
	if c.fileChecker.Exists(vaultTokenPath) {
		tokenOwner, err := c.fileChecker.Owner(vaultTokenPath)
		if err == nil && tokenOwner == "root" {
			rootOwned = append(rootOwned, vaultTokenPath)
		}
	}

	if len(rootOwned) == 0 {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "No root-owned cache directories found",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    fmt.Sprintf("Root-owned paths found: %s", strings.Join(rootOwned, ", ")),
		RepairHint: "Run 'vibe doctor --repair' to fix, or manually: sudo rm -f " + rootOwned[0],
	}
}

// CanRepair returns true as we can fix ownership or remove caches.
func (c *RootOwnedCachesCheck) CanRepair() bool { return true }

// Repair fixes root-owned caches by removing them (they will be recreated).
func (c *RootOwnedCachesCheck) Repair() error {
	var pathsToRemove []string
	var pathsToChown []string

	// Paths that should be removed entirely (caches)
	removablePaths := map[string]bool{
		".pex":                     true,
		".local/state/mcp-servers": true,
	}

	// Paths that should have ownership fixed instead of removed
	chownPaths := map[string]bool{
		".npm":  true,
		".kube": true,
	}

	for _, p := range c.paths {
		fullPath := filepath.Join(c.homeDir, p)
		if !c.fileChecker.Exists(fullPath) {
			continue
		}

		owner, err := c.fileChecker.Owner(fullPath)
		if err != nil {
			continue
		}

		needsSudo := owner == "root"
		// Also check for root-owned files inside user-owned directories
		if !needsSudo && c.fileChecker.IsDir(fullPath) {
			needsSudo = c.hasRootOwnedContents(fullPath)
		}

		if !needsSudo {
			continue
		}

		if removablePaths[p] {
			pathsToRemove = append(pathsToRemove, fullPath)
		} else if chownPaths[p] {
			pathsToChown = append(pathsToChown, fullPath)
		}
	}

	// Also check /tmp/cached_hcvault_token (absolute path, not relative to home)
	vaultTokenPath := "/tmp/cached_hcvault_token"
	if c.fileChecker.Exists(vaultTokenPath) {
		tokenOwner, err := c.fileChecker.Owner(vaultTokenPath)
		if err == nil && tokenOwner == "root" {
			pathsToRemove = append(pathsToRemove, vaultTokenPath)
		} else {
			// Not root-owned, safe to remove without sudo
			os.Remove(vaultTokenPath)
		}
	}

	if len(pathsToRemove) == 0 && len(pathsToChown) == 0 {
		return nil
	}

	// Build the sudo command
	user := c.envChecker.GetCurrentUser()
	var cmdParts []string

	if len(pathsToRemove) > 0 {
		cmdParts = append(cmdParts, fmt.Sprintf("rm -rf %s", strings.Join(pathsToRemove, " ")))
	}

	for _, p := range pathsToChown {
		cmdParts = append(cmdParts, fmt.Sprintf("chown -R %s:staff %s", user, p))
	}

	sudoCmd := strings.Join(cmdParts, " && ")

	cmd := exec.Command("sudo", "-p", "Password required to clean root-owned caches: ", "bash", "-c", sudoCmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ClaudeDirOwnershipCheck verifies ~/.claude is not owned by root.
// Root ownership of ~/.claude or ~/.claude/plugins causes EACCES errors
// when Claude Code tries to write to known_marketplaces.json.
type ClaudeDirOwnershipCheck struct {
	fileChecker FileChecker
	envChecker  EnvChecker
	claudeDir   string
}

// NewClaudeDirOwnershipCheck creates a new Claude directory ownership check.
func NewClaudeDirOwnershipCheck(fileChecker FileChecker, envChecker EnvChecker) *ClaudeDirOwnershipCheck {
	homeDir, _ := os.UserHomeDir()
	return &ClaudeDirOwnershipCheck{
		fileChecker: fileChecker,
		envChecker:  envChecker,
		claudeDir:   filepath.Join(homeDir, ".claude"),
	}
}

// Name returns the check identifier.
func (c *ClaudeDirOwnershipCheck) Name() string { return "claude_dir_ownership" }

// Description returns a human-readable description.
func (c *ClaudeDirOwnershipCheck) Description() string {
	return "Verify ~/.claude directory is not owned by root"
}

// Run executes the check.
func (c *ClaudeDirOwnershipCheck) Run() CheckResult {
	if !c.fileChecker.Exists(c.claudeDir) {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "~/.claude does not exist (OK)",
		}
	}

	// Check if the directory itself is owned by root
	owner, err := c.fileChecker.Owner(c.claudeDir)
	if err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusWarning,
			Message:    "Could not determine owner of ~/.claude",
			RepairHint: "Check ~/.claude ownership manually",
		}
	}

	if owner == "root" {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "~/.claude is owned by root (causes EACCES errors for marketplace/plugins)",
			RepairHint: fmt.Sprintf("Run: sudo chown -R %s:staff ~/.claude", c.envChecker.GetCurrentUser()),
		}
	}

	// Also check ~/.claude/plugins specifically
	pluginsDir := filepath.Join(c.claudeDir, "plugins")
	if c.fileChecker.Exists(pluginsDir) {
		pluginsOwner, err := c.fileChecker.Owner(pluginsDir)
		if err == nil && pluginsOwner == "root" {
			return CheckResult{
				Name:       c.Name(),
				Status:     StatusFail,
				Message:    "~/.claude/plugins is owned by root (causes EACCES errors for marketplace)",
				RepairHint: fmt.Sprintf("Run: sudo chown -R %s:staff ~/.claude", c.envChecker.GetCurrentUser()),
			}
		}
	}

	// Check for root-owned files inside (e.g., known_marketplaces.json)
	cmd := exec.Command("find", c.claudeDir, "-user", "root", "-maxdepth", "3", "-print", "-quit")
	output, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		rootFile := strings.TrimSpace(string(output))
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    fmt.Sprintf("~/.claude contains root-owned files (e.g., %s)", filepath.Base(rootFile)),
			RepairHint: fmt.Sprintf("Run: sudo chown -R %s:staff ~/.claude", c.envChecker.GetCurrentUser()),
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("~/.claude is owned by %s", owner),
	}
}

// CanRepair returns true as we can fix ownership.
func (c *ClaudeDirOwnershipCheck) CanRepair() bool { return true }

// Repair fixes the ownership of ~/.claude.
func (c *ClaudeDirOwnershipCheck) Repair() error {
	user := c.envChecker.GetCurrentUser()
	group := "staff"
	if runtime.GOOS == "linux" {
		group = user
	}

	cmd := exec.Command("sudo", "-p", "Password required to fix ~/.claude ownership: ", "chown", "-R", fmt.Sprintf("%s:%s", user, group), c.claudeDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// TelemetryHookCheck verifies the telemetry Stop hook is configured.
type TelemetryHookCheck struct {
	fileChecker FileChecker
	path        string
}

// TelemetryHookCommand is the expected telemetry hook command.
const TelemetryHookCommand = "vibe telemetry publish --event-type=claude.session.stop --source=claude-code-stop-hook --from-hook --quiet 2>/dev/null || true"

// NewTelemetryHookCheck creates a new telemetry hook check.
func NewTelemetryHookCheck(fileChecker FileChecker) *TelemetryHookCheck {
	homeDir, _ := os.UserHomeDir()
	return &TelemetryHookCheck{
		fileChecker: fileChecker,
		path:        filepath.Join(homeDir, ".claude", "settings.json"),
	}
}

// Name returns the check identifier.
func (c *TelemetryHookCheck) Name() string { return "telemetry_hook" }

// Description returns a human-readable description.
func (c *TelemetryHookCheck) Description() string {
	return "Verify telemetry Stop hook is configured"
}

// Run executes the check.
func (c *TelemetryHookCheck) Run() CheckResult {
	data, err := c.fileChecker.ReadFile(c.path)
	if err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "Could not read settings.json",
			RepairHint: "Run 'vibe update' to configure hooks",
		}
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "Could not parse settings.json",
			RepairHint: "Run 'vibe update' to configure hooks",
		}
	}

	// Check for hooks.Stop array
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "No hooks configured in settings.json",
			RepairHint: "Run 'vibe update' to configure hooks",
		}
	}

	stopHooks, ok := hooks["Stop"].([]interface{})
	if !ok {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "No Stop hooks configured",
			RepairHint: "Run 'vibe update' to configure hooks",
		}
	}

	// Look for our telemetry hook command in nested hooks array
	for _, h := range stopHooks {
		hookMap, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		nestedHooks, ok := hookMap["hooks"].([]interface{})
		if !ok {
			continue
		}
		for _, nh := range nestedHooks {
			nhMap, ok := nh.(map[string]interface{})
			if !ok {
				continue
			}
			if cmd, ok := nhMap["command"].(string); ok {
				if cmd == TelemetryHookCommand {
					return CheckResult{
						Name:    c.Name(),
						Status:  StatusPass,
						Message: "Telemetry hook is configured",
					}
				}
			}
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    "Telemetry Stop hook not found",
		RepairHint: "Run 'vibe update' to configure hooks",
	}
}

// CanRepair returns true as we can add the hook.
func (c *TelemetryHookCheck) CanRepair() bool { return true }

// Repair adds the telemetry hook to settings.json.
func (c *TelemetryHookCheck) Repair() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return fmt.Errorf("could not read settings.json: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("could not parse settings.json: %w", err)
	}

	// Define the telemetry stop hook
	telemetryHookEntry := map[string]interface{}{
		"matcher": "",
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": TelemetryHookCommand,
				"timeout": 30,
			},
		},
	}

	// Get or create hooks section
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
	}

	// Get or create Stop hooks array
	stopHooks, ok := hooks["Stop"].([]interface{})
	if !ok {
		stopHooks = []interface{}{}
	}

	// Add the hook
	stopHooks = append(stopHooks, telemetryHookEntry)
	hooks["Stop"] = stopHooks
	settings["hooks"] = hooks

	// Write back
	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal settings: %w", err)
	}

	if err := os.WriteFile(c.path, output, 0644); err != nil {
		return fmt.Errorf("could not write settings.json: %w", err)
	}

	return nil
}

// PathConfigCheck verifies ~/.local/bin is in PATH or shell rc.
type PathConfigCheck struct {
	envChecker  EnvChecker
	fileChecker FileChecker
	localBinDir string
}

// NewPathConfigCheck creates a new PATH configuration check.
func NewPathConfigCheck(envChecker EnvChecker, fileChecker FileChecker) *PathConfigCheck {
	homeDir, _ := os.UserHomeDir()
	return &PathConfigCheck{
		envChecker:  envChecker,
		fileChecker: fileChecker,
		localBinDir: filepath.Join(homeDir, ".local", "bin"),
	}
}

// Name returns the check identifier.
func (c *PathConfigCheck) Name() string { return "path_config" }

// Description returns a human-readable description.
func (c *PathConfigCheck) Description() string {
	return "Verify ~/.local/bin is in PATH"
}

// Run executes the check.
func (c *PathConfigCheck) Run() CheckResult {
	// Check if ~/.local/bin is in current PATH
	path := c.envChecker.GetEnv("PATH")
	if strings.Contains(path, c.localBinDir) || strings.Contains(path, ".local/bin") {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "~/.local/bin is in PATH",
		}
	}

	// Check if it's configured in shell rc
	rcContent, err := c.envChecker.GetShellRCContent()
	if err == nil && strings.Contains(rcContent, ".local/bin") {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "~/.local/bin configured in shell rc (restart shell to activate)",
		}
	}

	return CheckResult{
		Name:       c.Name(),
		Status:     StatusFail,
		Message:    "~/.local/bin is not in PATH",
		RepairHint: "Run 'vibe install' to configure PATH",
	}
}

// CanRepair returns true as we can add PATH to shell rc.
func (c *PathConfigCheck) CanRepair() bool { return true }

// Repair adds ~/.local/bin to PATH in shell rc.
func (c *PathConfigCheck) Repair() error {
	rcPath := c.envChecker.GetShellRC()
	rcContent, _ := c.envChecker.GetShellRCContent()

	// Check if already configured
	if strings.Contains(rcContent, ".local/bin") {
		return nil
	}

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open shell rc: %w", err)
	}
	defer f.Close()

	pathExport := `export PATH="$HOME/.local/bin:$PATH"`
	if _, err := f.WriteString("\n# Added by vibe doctor\n" + pathExport + "\n"); err != nil {
		return fmt.Errorf("failed to write to shell rc: %w", err)
	}

	return nil
}

// DefaultModelCheck verifies the default model is set to opus in settings.json.
type DefaultModelCheck struct {
	fileChecker FileChecker
	path        string
}

// NewDefaultModelCheck creates a new default model check.
func NewDefaultModelCheck(fileChecker FileChecker) *DefaultModelCheck {
	homeDir, _ := os.UserHomeDir()
	return &DefaultModelCheck{
		fileChecker: fileChecker,
		path:        filepath.Join(homeDir, ".claude", "settings.json"),
	}
}

// Name returns the check identifier.
func (c *DefaultModelCheck) Name() string { return "default_model" }

// Description returns a human-readable description.
func (c *DefaultModelCheck) Description() string {
	return "Verify default model is set to opus"
}

// Run executes the check.
func (c *DefaultModelCheck) Run() CheckResult {
	data, err := c.fileChecker.ReadFile(c.path)
	if err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "Could not read settings.json",
			RepairHint: "Run 'vibe update' to configure default model",
		}
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    "Could not parse settings.json",
			RepairHint: "Run 'vibe update' to configure default model",
		}
	}

	model, ok := settings["model"].(string)
	if !ok || model != "opus" {
		currentModel := "not set"
		if ok {
			currentModel = model
		}
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    fmt.Sprintf("Default model is %q, expected \"opus\"", currentModel),
			RepairHint: "Run 'vibe update' to set default model to opus",
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Default model is set to opus",
	}
}

// CanRepair returns true as we can set the model.
func (c *DefaultModelCheck) CanRepair() bool { return true }

// Repair sets the default model to opus in settings.json.
func (c *DefaultModelCheck) Repair() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return fmt.Errorf("could not read settings.json: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("could not parse settings.json: %w", err)
	}

	settings["model"] = "opus"

	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal settings: %w", err)
	}

	if err := os.WriteFile(c.path, output, 0644); err != nil {
		return fmt.Errorf("could not write settings.json: %w", err)
	}

	return nil
}

// NpmPrefixPermissionsCheck verifies the npm global prefix directory is writable.
// This catches EACCES errors when running npm install -g (e.g., for Salesforce CLI).
type NpmPrefixPermissionsCheck struct {
	cmdChecker CommandChecker
	envChecker EnvChecker
	prefix     string // resolved lazily
}

// NewNpmPrefixPermissionsCheck creates a new npm prefix permissions check.
func NewNpmPrefixPermissionsCheck(cmdChecker CommandChecker, envChecker EnvChecker) *NpmPrefixPermissionsCheck {
	return &NpmPrefixPermissionsCheck{
		cmdChecker: cmdChecker,
		envChecker: envChecker,
	}
}

// Name returns the check identifier.
func (c *NpmPrefixPermissionsCheck) Name() string { return "npm_prefix_permissions" }

// Description returns a human-readable description.
func (c *NpmPrefixPermissionsCheck) Description() string {
	return "Verify npm global directory is writable"
}

// getNpmPrefix resolves and caches the npm global prefix directory.
func (c *NpmPrefixPermissionsCheck) getNpmPrefix() (string, error) {
	if c.prefix != "" {
		return c.prefix, nil
	}
	if !c.cmdChecker.IsInstalled("npm") {
		return "", fmt.Errorf("npm not installed")
	}
	out, err := exec.Command("npm", "config", "get", "prefix").Output()
	if err != nil {
		return "", fmt.Errorf("could not get npm prefix: %w", err)
	}
	c.prefix = strings.TrimSpace(string(out))
	return c.prefix, nil
}

// Run executes the check.
func (c *NpmPrefixPermissionsCheck) Run() CheckResult {
	if !c.cmdChecker.IsInstalled("npm") {
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "npm not installed (skipped)",
		}
	}

	prefix, err := c.getNpmPrefix()
	if err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusWarning,
			Message:    "Could not determine npm prefix",
			RepairHint: "Run: npm config get prefix",
		}
	}

	npmModulesDir := filepath.Join(prefix, "lib", "node_modules")

	// If node_modules doesn't exist, check the lib dir
	checkDir := npmModulesDir
	if _, err := os.Stat(checkDir); err != nil {
		checkDir = filepath.Join(prefix, "lib")
		if _, err := os.Stat(checkDir); err != nil {
			return CheckResult{
				Name:    c.Name(),
				Status:  StatusPass,
				Message: "npm prefix directories do not exist yet (OK)",
			}
		}
	}

	owner, err := (&DefaultFileChecker{}).Owner(checkDir)
	if err != nil {
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusWarning,
			Message:    fmt.Sprintf("Could not determine owner of %s", checkDir),
			RepairHint: fmt.Sprintf("Check permissions manually: ls -la %s", checkDir),
		}
	}

	if owner == "root" {
		currentUser := c.envChecker.GetCurrentUser()
		return CheckResult{
			Name:       c.Name(),
			Status:     StatusFail,
			Message:    fmt.Sprintf("%s is owned by root", checkDir),
			RepairHint: fmt.Sprintf("Run: sudo chown -R %s:staff %s", currentUser, checkDir),
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("npm global directory is writable (%s)", prefix),
	}
}

// CanRepair returns true as we can fix ownership.
func (c *NpmPrefixPermissionsCheck) CanRepair() bool { return true }

// Repair fixes the ownership of the npm global prefix directories.
func (c *NpmPrefixPermissionsCheck) Repair() error {
	prefix, err := c.getNpmPrefix()
	if err != nil {
		return err
	}

	currentUser := c.envChecker.GetCurrentUser()
	group := "staff"
	if runtime.GOOS == "linux" {
		group = currentUser
	}

	ownership := fmt.Sprintf("%s:%s", currentUser, group)

	// Fix lib/node_modules
	npmModulesDir := filepath.Join(prefix, "lib", "node_modules")
	if _, err := os.Stat(npmModulesDir); err == nil {
		cmd := exec.Command("sudo", "-p", "Password required to fix npm directory ownership: ", "chown", "-R", ownership, npmModulesDir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to fix %s: %w", npmModulesDir, err)
		}
	}

	// Fix bin directory too (where npm links global binaries)
	binDir := filepath.Join(prefix, "bin")
	if _, err := os.Stat(binDir); err == nil {
		owner, _ := (&DefaultFileChecker{}).Owner(binDir)
		if owner == "root" {
			cmd := exec.Command("sudo", "-p", "Password required to fix npm bin directory ownership: ", "chown", "-R", ownership, binDir)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to fix %s: %w", binDir, err)
			}
		}
	}

	return nil
}

// AllChecks returns all available health checks with default dependencies.
func AllChecks() []Check {
	cmdChecker := &DefaultCommandChecker{}
	fileChecker := &DefaultFileChecker{}
	mpChecker := &DefaultMarketplaceChecker{}
	envChecker := &DefaultEnvChecker{}

	// No required MCP servers — user configures their own in mcp-servers.yaml
	requiredMCPServers := []string{}

	// No required environment variables by default
	requiredEnvVars := []string{}

	return []Check{
		NewPrereqsCheck(cmdChecker, fileChecker),
		NewRecommendedToolsCheck(cmdChecker),
		NewPathConfigCheck(envChecker, fileChecker),
		NewMarketplaceCheck(fileChecker),
		NewMarketplaceRegisteredCheck(mpChecker),
		NewSettingsJSONCheck(fileChecker),
		NewDefaultModelCheck(fileChecker),
		NewTelemetryHookCheck(fileChecker),
		NewMCPConfigCheck(fileChecker),
		NewMCPServersCheck(fileChecker, requiredMCPServers),
		NewPluginsInstalledCheck(mpChecker),
		NewPluginsOutdatedCheck(mpChecker),
		NewEnvVarsCheck(envChecker, requiredEnvVars),
		NewClaudeDirOwnershipCheck(fileChecker, envChecker),
		NewLocalOwnershipCheck(fileChecker, envChecker),
		NewRootOwnedCachesCheck(fileChecker, envChecker),
		NewNpmPrefixPermissionsCheck(cmdChecker, envChecker),
	}
}

// RunAll executes all checks and returns results.
func RunAll() []CheckResult {
	checks := AllChecks()
	results := make([]CheckResult, len(checks))

	for i, check := range checks {
		results[i] = check.Run()
	}

	return results
}
