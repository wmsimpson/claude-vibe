package install

import (
	"os/exec"
)

// NodeNpmStep installs Node.js and npm via Homebrew
type NodeNpmStep struct{}

func (s *NodeNpmStep) ID() string          { return "node_npm" }
func (s *NodeNpmStep) Name() string        { return "Node.js/npm" }
func (s *NodeNpmStep) Description() string { return "Install Node.js for MCP servers" }
func (s *NodeNpmStep) ActiveForm() string  { return "Installing Node.js" }
func (s *NodeNpmStep) CanSkip(opts *Options) bool { return opts.NoBrew }
func (s *NodeNpmStep) NeedsSudo() bool     { return false }

func (s *NodeNpmStep) Check(ctx *Context) (bool, error) {
	return ctx.IsCommandInstalled("node") && ctx.IsCommandInstalled("npm"), nil
}

func (s *NodeNpmStep) Run(ctx *Context) StepResult {
	if ctx.IsCommandInstalled("node") && ctx.IsCommandInstalled("npm") {
		// Get version for details
		nodeVer, _ := ctx.RunCommand("node", "--version")
		npmVer, _ := ctx.RunCommand("npm", "--version")
		return SuccessWithDetails(
			"Node.js already installed",
			"  Node: "+nodeVer+"  npm: "+npmVer,
		)
	}

	// In --no-brew mode, report missing and continue
	if ctx.Options.NoBrew {
		return SuccessWithDetails(
			"Brew skipped: node/npm not found",
			"  Missing tools (install manually for best results):\n    - node (Node.js runtime)\n    - npm (Node.js package manager)\n",
		)
	}

	ctx.Log("Installing Node.js...")

	cmd := exec.Command("brew", "install", "node")
	cmd.Stdin = nil
	if output, err := cmd.CombinedOutput(); err != nil {
		ctx.Log("Failed to install node: %s", string(output))
		return FailureWithHint(
			"Failed to install Node.js",
			err,
			"Try running manually: brew install node",
		)
	}

	// Verify installation
	if !ctx.IsCommandInstalled("node") || !ctx.IsCommandInstalled("npm") {
		return Failure("Node.js installed but not found in PATH", nil)
	}

	nodeVer, _ := ctx.RunCommand("node", "--version")
	npmVer, _ := ctx.RunCommand("npm", "--version")

	return SuccessWithDetails(
		"Node.js installed",
		"  Node: "+nodeVer+"  npm: "+npmVer,
	)
}
