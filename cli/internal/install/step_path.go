package install

import (
	"os"
	"strings"
)

// PathConfigStep ensures ~/.local/bin is in PATH
type PathConfigStep struct{}

func (s *PathConfigStep) ID() string          { return "path_config" }
func (s *PathConfigStep) Name() string        { return "Configure PATH" }
func (s *PathConfigStep) Description() string { return "Add ~/.local/bin to PATH" }
func (s *PathConfigStep) ActiveForm() string  { return "Configuring PATH" }
func (s *PathConfigStep) CanSkip(opts *Options) bool { return false }
func (s *PathConfigStep) NeedsSudo() bool     { return false }

func (s *PathConfigStep) Check(ctx *Context) (bool, error) {
	// Check if ~/.local/bin is in current PATH
	path := os.Getenv("PATH")
	if strings.Contains(path, ctx.LocalBinDir) {
		return true, nil
	}

	// Check if it's configured in shell rc
	if ctx.FileExists(ctx.ShellRC) {
		content, err := os.ReadFile(ctx.ShellRC)
		if err == nil {
			if strings.Contains(string(content), ".local/bin") {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *PathConfigStep) Run(ctx *Context) StepResult {
	pathExport := `export PATH="$HOME/.local/bin:$PATH"`

	// Check if already configured
	if ctx.FileExists(ctx.ShellRC) {
		content, err := os.ReadFile(ctx.ShellRC)
		if err == nil && strings.Contains(string(content), ".local/bin") {
			return Success("PATH already configured")
		}
	}

	// Append to shell rc
	f, err := os.OpenFile(ctx.ShellRC, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return Failure("Failed to open shell rc file", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n# Added by vibe installer\n" + pathExport + "\n"); err != nil {
		return Failure("Failed to write to shell rc file", err)
	}

	// Also set for current process
	os.Setenv("PATH", ctx.LocalBinDir+":"+os.Getenv("PATH"))

	return SuccessWithDetails(
		"Added ~/.local/bin to PATH",
		"Added to "+ctx.ShellRC,
	)
}
