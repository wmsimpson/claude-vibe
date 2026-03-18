package install

import (
	"os/exec"
)

// DatabricksCliStep installs the Databricks CLI via brew tap
type DatabricksCliStep struct{}

func (s *DatabricksCliStep) ID() string          { return "databricks_cli" }
func (s *DatabricksCliStep) Name() string        { return "Databricks CLI" }
func (s *DatabricksCliStep) Description() string { return "Install Databricks CLI" }
func (s *DatabricksCliStep) ActiveForm() string  { return "Installing Databricks CLI" }
func (s *DatabricksCliStep) CanSkip(opts *Options) bool { return opts.NoBrew }
func (s *DatabricksCliStep) NeedsSudo() bool     { return false }

func (s *DatabricksCliStep) Check(ctx *Context) (bool, error) {
	return ctx.IsCommandInstalled("databricks"), nil
}

func (s *DatabricksCliStep) Run(ctx *Context) StepResult {
	if ctx.IsCommandInstalled("databricks") {
		return Success("Databricks CLI already installed")
	}

	// In --no-brew mode, report missing and continue
	if ctx.Options.NoBrew {
		return SuccessWithDetails(
			"Brew skipped: databricks CLI not found",
			"  Missing tools (install manually for best results):\n    - databricks (Databricks CLI)\n      Install: https://docs.databricks.com/dev-tools/cli/install.html\n",
		)
	}

	ctx.Log("Adding Databricks tap...")

	// Add the Databricks tap
	cmd := exec.Command("brew", "tap", "databricks/tap")
	if output, err := cmd.CombinedOutput(); err != nil {
		ctx.Log("Failed to tap databricks/tap: %s", string(output))
		return FailureWithHint(
			"Failed to add Databricks tap",
			err,
			"Try running manually: brew tap databricks/tap && brew install databricks",
		)
	}

	ctx.Log("Installing Databricks CLI...")

	// Install databricks
	cmd = exec.Command("brew", "install", "databricks")
	cmd.Stdin = nil
	if output, err := cmd.CombinedOutput(); err != nil {
		ctx.Log("Failed to install databricks: %s", string(output))
		return FailureWithHint(
			"Failed to install Databricks CLI",
			err,
			"Try running manually: brew install databricks/tap/databricks",
		)
	}

	return Success("Databricks CLI installed")
}
