package install

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// VerificationStep verifies the installation
type VerificationStep struct{}

func (s *VerificationStep) ID() string          { return "verification" }
func (s *VerificationStep) Name() string        { return "Verify Install" }
func (s *VerificationStep) Description() string { return "Verify installation was successful" }
func (s *VerificationStep) ActiveForm() string  { return "Verifying installation" }
func (s *VerificationStep) CanSkip(opts *Options) bool { return false }
func (s *VerificationStep) NeedsSudo() bool     { return false }

func (s *VerificationStep) Check(ctx *Context) (bool, error) {
	// Always run verification
	return false, nil
}

func (s *VerificationStep) Run(ctx *Context) StepResult {
	var checks []string
	var warnings []string
	var failures []string

	// Check Claude is installed
	if ctx.IsCommandInstalled("claude") {
		ver, _ := ctx.RunCommand("claude", "--version")
		checks = append(checks, fmt.Sprintf("Claude Code: %s", strings.TrimSpace(ver)))
	} else {
		failures = append(failures, "Claude Code not installed")
	}

	// Check marketplace is registered
	output, err := exec.Command("claude", "plugin", "marketplace", "list").Output()
	if err == nil && strings.Contains(string(output), "claude-vibe") {
		checks = append(checks, "Marketplace: claude-vibe registered")
	} else {
		warnings = append(warnings, "Marketplace may not be registered")
	}

	// Check settings.json exists
	settingsFile := filepath.Join(ctx.ClaudeDir, "settings.json")
	if ctx.FileExists(settingsFile) {
		checks = append(checks, "Settings: ~/.claude/settings.json exists")
	} else {
		warnings = append(warnings, "Settings file not found")
	}

	// Check MCP config exists
	mcpConfig := filepath.Join(ctx.ConfigDir, "mcp", "config.json")
	if ctx.FileExists(mcpConfig) {
		checks = append(checks, "MCP: ~/.config/mcp/config.json exists")
	} else {
		warnings = append(warnings, "MCP config not found")
	}

	// Check required tools
	tools := []struct {
		cmd  string
		name string
	}{
		{"gh", "GitHub CLI"},
		{"jq", "jq"},
		{"yq", "yq"},
		{"python3.10", "Python 3.10"},
		{"aws", "AWS CLI"},
	}

	for _, tool := range tools {
		if ctx.IsCommandInstalled(tool.cmd) {
			checks = append(checks, fmt.Sprintf("%s: installed", tool.name))
		} else {
			warnings = append(warnings, fmt.Sprintf("%s not installed", tool.name))
		}
	}

	// Build details string
	var details strings.Builder
	details.WriteString("Checks:\n")
	for _, c := range checks {
		details.WriteString("  [OK] " + c + "\n")
	}
	if len(warnings) > 0 {
		details.WriteString("\nWarnings:\n")
		for _, w := range warnings {
			details.WriteString("  [!] " + w + "\n")
		}
	}
	if len(failures) > 0 {
		details.WriteString("\nFailures:\n")
		for _, f := range failures {
			details.WriteString("  [X] " + f + "\n")
		}
	}

	if len(failures) > 0 {
		return FailureWithHint(
			"Installation verification failed",
			nil,
			"Run 'vibe doctor' for detailed diagnostics",
		)
	}

	if len(warnings) > 0 {
		return SuccessWithDetails(
			"Installation verified with warnings",
			details.String(),
		)
	}

	return SuccessWithDetails(
		"Installation verified successfully",
		details.String(),
	)
}
