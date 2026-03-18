package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ClaudeInstallStep installs Claude Code CLI
type ClaudeInstallStep struct{}

func (s *ClaudeInstallStep) ID() string          { return "claude_install" }
func (s *ClaudeInstallStep) Name() string        { return "Install Claude" }
func (s *ClaudeInstallStep) Description() string { return "Install Claude Code CLI" }
func (s *ClaudeInstallStep) ActiveForm() string  { return "Installing Claude Code" }
func (s *ClaudeInstallStep) CanSkip(opts *Options) bool { return false }
func (s *ClaudeInstallStep) NeedsSudo() bool     { return false } // May need sudo only to fix root-owned binary

func (s *ClaudeInstallStep) Check(ctx *Context) (bool, error) {
	if !ctx.IsCommandInstalled("claude") {
		// Also check ~/.local/bin/claude directly
		claudePath := filepath.Join(ctx.LocalBinDir, "claude")
		return ctx.FileExists(claudePath), nil
	}
	return true, nil
}

func (s *ClaudeInstallStep) Run(ctx *Context) StepResult {
	// Check for existing claude binary
	claudePath, _ := ctx.GetCommandPath("claude")
	if claudePath == "" {
		claudePath = filepath.Join(ctx.LocalBinDir, "claude")
	}

	// Check if claude exists and is owned by root
	if ctx.FileExists(claudePath) {
		owner, _ := ctx.GetFileOwner(claudePath)
		if owner == "root" {
			if ctx.Options.ForceReinstall {
				ctx.Log("Removing root-owned claude binary at %s", claudePath)
				// Try without sudo first
				if err := os.Remove(claudePath); err != nil {
					// Need sudo
					ctx.RunSudoCommand("rm", "-f", claudePath)
				}
			} else {
				return FailureWithHint(
					fmt.Sprintf("Claude binary at %s is owned by root", claudePath),
					errors.New("previous installation was run with sudo"),
					"Re-run with --force-reinstall to fix this",
				)
			}
		} else if !ctx.Options.ForceReinstall {
			ver, _ := ctx.RunCommand("claude", "--version")
			return SuccessWithDetails(
				"Claude Code already installed",
				"  Version: "+ver,
			)
		}
	}

	// Clean up locks before installation
	locksDir := filepath.Join(ctx.HomeDir, ".local", "state", "claude", "locks")
	os.RemoveAll(locksDir)

	ctx.Log("Installing Claude Code...")

	// Use the official installer
	cmd := exec.Command("/bin/bash", "-c", "curl -fsSL https://claude.ai/install.sh | bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return FailureWithHint(
			"Failed to install Claude Code",
			err,
			"Try running manually: curl -fsSL https://claude.ai/install.sh | bash",
		)
	}

	// Verify installation
	claudePath = filepath.Join(ctx.LocalBinDir, "claude")
	if !ctx.FileExists(claudePath) && !ctx.IsCommandInstalled("claude") {
		return Failure("Claude Code installed but not found", nil)
	}

	// Get version
	ver, _ := ctx.RunCommand("claude", "--version")

	return SuccessWithDetails("Claude Code installed", "  Version: "+ver)
}
