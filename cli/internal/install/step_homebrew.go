package install

import (
	"os"
	"os/exec"
)

// HomebrewStep ensures Homebrew is installed
type HomebrewStep struct{}

func (s *HomebrewStep) ID() string          { return "homebrew" }
func (s *HomebrewStep) Name() string        { return "Install Homebrew" }
func (s *HomebrewStep) Description() string { return "Install Homebrew package manager" }
func (s *HomebrewStep) ActiveForm() string  { return "Installing Homebrew" }
func (s *HomebrewStep) CanSkip(opts *Options) bool { return opts.NoBrew }
func (s *HomebrewStep) NeedsSudo() bool     { return false }

func (s *HomebrewStep) Check(ctx *Context) (bool, error) {
	return ctx.IsCommandInstalled("brew"), nil
}

func (s *HomebrewStep) Run(ctx *Context) StepResult {
	if ctx.IsCommandInstalled("brew") {
		return Success("Homebrew already installed")
	}

	// Install Homebrew using the official script
	ctx.Log("Installing Homebrew...")

	cmd := exec.Command("/bin/bash", "-c", "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "NONINTERACTIVE=1")

	if err := cmd.Run(); err != nil {
		return FailureWithHint(
			"Failed to install Homebrew",
			err,
			"Install manually: /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"",
		)
	}

	// Add Homebrew to PATH for current session
	brewPath := ctx.HomebrewPrefix + "/bin/brew"
	if ctx.FileExists(brewPath) {
		// Run shellenv to get the proper PATH exports
		output, err := exec.Command(brewPath, "shellenv").Output()
		if err == nil {
			// Execute the shellenv output to set up PATH
			exec.Command("/bin/bash", "-c", string(output)).Run()
		}
	}

	return Success("Homebrew installed")
}
