package install

import (
	"fmt"
	"os/exec"
	"strings"
)

// BrewToolsStep installs required tools via Homebrew
type BrewToolsStep struct{}

func (s *BrewToolsStep) ID() string          { return "brew_tools" }
func (s *BrewToolsStep) Name() string        { return "Install Tools" }
func (s *BrewToolsStep) Description() string { return "Install required tools via Homebrew" }
func (s *BrewToolsStep) ActiveForm() string  { return "Installing tools" }
func (s *BrewToolsStep) CanSkip(opts *Options) bool { return opts.NoBrew }
func (s *BrewToolsStep) NeedsSudo() bool     { return false }

// brewToolEntry maps a brew formula to its command name and description
type brewToolEntry struct {
	Formula     string // brew formula name (e.g., "python@3.10")
	CommandName string // command to check in PATH (e.g., "python3.10")
	Description string // human-readable description for --no-brew output
}

var requiredBrewTools = []brewToolEntry{
	{"jq", "jq", "JSON processor"},
	{"yq", "yq", "YAML processor"},
	{"gh", "gh", "GitHub CLI"},
	{"python@3.10", "python3.10", "Python 3.10"},
	{"uv", "uv", "Python package manager"},
	{"pipx", "pipx", "Python app installer"},
	{"terminal-notifier", "terminal-notifier", "macOS notifications"},
	{"awscli", "aws", "AWS CLI"},
}

func (s *BrewToolsStep) Check(ctx *Context) (bool, error) {
	for _, tool := range requiredBrewTools {
		if !ctx.IsCommandInstalled(tool.CommandName) {
			return false, nil
		}
	}
	return true, nil
}

func (s *BrewToolsStep) Run(ctx *Context) StepResult {
	// In --no-brew mode, check what's present and report what's missing
	if ctx.Options.NoBrew {
		return s.runNoBrew(ctx)
	}

	var installed []string
	var alreadyPresent []string
	var failed []string

	for _, tool := range requiredBrewTools {
		if ctx.IsCommandInstalled(tool.CommandName) {
			alreadyPresent = append(alreadyPresent, tool.Formula)
			continue
		}

		ctx.Log("Installing %s...", tool.Formula)

		// Run brew install with stdin redirected from /dev/null to prevent prompts
		cmd := exec.Command("brew", "install", tool.Formula)
		cmd.Stdin = nil // Prevents interactive prompts
		output, err := cmd.CombinedOutput()

		if err != nil {
			ctx.Log("Failed to install %s: %s", tool.Formula, string(output))
			failed = append(failed, tool.Formula)
		} else {
			installed = append(installed, tool.Formula)
		}
	}

	// Build details
	var details string
	if len(installed) > 0 {
		details += "  Installed: " + strings.Join(installed, ", ") + "\n"
	}
	if len(alreadyPresent) > 0 {
		details += "  Already present: " + strings.Join(alreadyPresent, ", ") + "\n"
	}
	if len(failed) > 0 {
		details += "  Failed: " + strings.Join(failed, ", ") + "\n"
	}

	if len(failed) > 0 {
		return FailureWithHint(
			"Some tools failed to install",
			nil,
			"Try running manually: brew install "+strings.Join(failed, " "),
		)
	}

	msg := "All tools installed"
	if len(installed) == 0 {
		msg = "All tools already present"
	}

	return SuccessWithDetails(msg, details)
}

// BrewManagedTool represents a tool normally installed via Homebrew
type BrewManagedTool struct {
	Command     string // command name to check in PATH
	Description string // human-readable description
	InstallHint string // optional install URL/instructions
}

// MissingBrewManagedTools returns all brew-managed tools that are not found in PATH.
// This is used to show a warning when --no-brew mode was used.
func MissingBrewManagedTools() []BrewManagedTool {
	allTools := []BrewManagedTool{
		// From requiredBrewTools
		{"jq", "JSON processor", ""},
		{"yq", "YAML processor", ""},
		{"gh", "GitHub CLI", "https://github.com/cli/cli#installation"},
		{"python3.10", "Python 3.10", "https://www.python.org/downloads/"},
		{"uv", "Python package manager", "https://docs.astral.sh/uv/"},
		{"pipx", "Python app installer", "https://pipx.pypa.io/latest/installation/"},
		{"terminal-notifier", "macOS notifications", ""},
		{"aws", "AWS CLI", "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"},
		// Node/npm
		{"node", "Node.js runtime", "https://nodejs.org/"},
		{"npm", "Node.js package manager", ""},
		// gcloud
		{"gcloud", "Google Cloud SDK", "https://cloud.google.com/sdk/docs/install"},
		// databricks
		{"databricks", "Databricks CLI", "https://docs.databricks.com/dev-tools/cli/install.html"},
	}

	var missing []BrewManagedTool
	for _, tool := range allTools {
		if _, err := exec.LookPath(tool.Command); err != nil {
			missing = append(missing, tool)
		}
	}
	return missing
}

// runNoBrew checks which tools are present and reports missing ones
func (s *BrewToolsStep) runNoBrew(ctx *Context) StepResult {
	var present []string
	var missing []brewToolEntry

	for _, tool := range requiredBrewTools {
		if ctx.IsCommandInstalled(tool.CommandName) {
			present = append(present, tool.CommandName)
		} else {
			missing = append(missing, tool)
		}
	}

	var details string
	if len(present) > 0 {
		details += "  Found: " + strings.Join(present, ", ") + "\n"
	}

	if len(missing) == 0 {
		return SuccessWithDetails("All tools already present (brew skipped)", details)
	}

	// Build a helpful message listing what needs to be installed
	details += "\n  Missing tools (install manually for best results):\n"
	for _, tool := range missing {
		details += fmt.Sprintf("    - %s (%s)\n", tool.CommandName, tool.Description)
	}

	return SuccessWithDetails(
		fmt.Sprintf("Brew skipped: %d tools found, %d missing", len(present), len(missing)),
		details,
	)
}
