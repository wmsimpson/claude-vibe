package install

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PermissionsSyncStep merges permissions.yaml into settings.json
type PermissionsSyncStep struct{}

func (s *PermissionsSyncStep) ID() string          { return "permissions_sync" }
func (s *PermissionsSyncStep) Name() string        { return "Sync Permissions" }
func (s *PermissionsSyncStep) Description() string { return "Merge permissions into Claude settings" }
func (s *PermissionsSyncStep) ActiveForm() string  { return "Syncing permissions" }
func (s *PermissionsSyncStep) CanSkip(opts *Options) bool { return false }
func (s *PermissionsSyncStep) NeedsSudo() bool     { return false }

func (s *PermissionsSyncStep) Check(ctx *Context) (bool, error) {
	// Always sync to ensure latest permissions
	return false, nil
}

func (s *PermissionsSyncStep) Run(ctx *Context) StepResult {
	// Find permissions file
	permissionsFile := ""
	if ctx.VibeDir != "" {
		permissionsFile = filepath.Join(ctx.VibeDir, "permissions.yaml")
	}
	if permissionsFile == "" || !ctx.FileExists(permissionsFile) {
		permissionsFile = filepath.Join(ctx.MarketplaceDir, "permissions.yaml")
	}

	if !ctx.FileExists(permissionsFile) {
		return Skip("permissions.yaml not found")
	}

	settingsFile := filepath.Join(ctx.ClaudeDir, "settings.json")

	// Ensure claude directory exists
	if err := os.MkdirAll(ctx.ClaudeDir, 0755); err != nil {
		return Failure("Failed to create .claude directory", err)
	}

	// Create settings.json if it doesn't exist
	if !ctx.FileExists(settingsFile) {
		initial := map[string]interface{}{
			"allow": []string{},
			"deny":  []string{},
		}
		data, _ := json.MarshalIndent(initial, "", "  ")
		os.WriteFile(settingsFile, data, 0644)
	}

	// Check for yq
	if !ctx.IsCommandInstalled("yq") {
		return FailureWithHint(
			"yq not installed",
			nil,
			"Install yq: brew install yq",
		)
	}

	// Extract allow list from YAML
	ctx.Log("Extracting permissions from %s", permissionsFile)
	output, err := exec.Command("yq", "-r", ".allow | @json", permissionsFile).Output()
	if err != nil {
		return Failure("Failed to parse permissions.yaml", err)
	}

	allowPerms := string(output)

	// Check for jq
	if !ctx.IsCommandInstalled("jq") {
		return FailureWithHint(
			"jq not installed",
			nil,
			"Install jq: brew install jq",
		)
	}

	// Read current settings
	settingsData, err := os.ReadFile(settingsFile)
	if err != nil {
		return Failure("Failed to read settings.json", err)
	}

	// Merge permissions using jq
	cmd := exec.Command("jq", "--argjson", "new_perms", allowPerms,
		`.allow = (.allow // [] | . + $new_perms | unique)`)
	cmd.Stdin = strings.NewReader(string(settingsData))
	merged, err := cmd.Output()
	if err != nil {
		return Failure("Failed to merge permissions", err)
	}

	// Write back
	if err := os.WriteFile(settingsFile, merged, 0644); err != nil {
		return Failure("Failed to write settings.json", err)
	}

	// Remove invalid bash find permissions (known issue)
	exec.Command("sed", "-i", "", `/"Bash(find:/d`, settingsFile).Run()

	return Success("Permissions synced")
}
