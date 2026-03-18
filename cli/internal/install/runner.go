package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Runner handles non-interactive installation
type Runner struct {
	ctx     *Context
	state   *State
	steps   []Step
	verbose bool
}

// NewRunner creates a new installation runner
func NewRunner(opts *Options) (*Runner, error) {
	ctx, err := NewContext(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	// Set up logging
	if opts.Verbose {
		ctx.Log = func(format string, args ...interface{}) {
			fmt.Printf("  [debug] "+format+"\n", args...)
		}
	}

	runner := &Runner{
		ctx:     ctx,
		steps:   AllSteps(),
		verbose: opts.Verbose,
	}

	// Load or create state
	if opts.Resume {
		state, err := LoadState()
		if err != nil {
			return nil, fmt.Errorf("failed to load state: %w", err)
		}
		if state == nil {
			return nil, fmt.Errorf("no installation to resume")
		}
		runner.state = state
	} else {
		runner.state = NewState()
	}

	return runner, nil
}

// Run executes all installation steps
func (r *Runner) Run() error {
	// Save state periodically
	defer r.state.Save()

	// Determine starting point
	startIdx := 0
	if r.ctx.Options.Resume {
		startIdx = r.state.GetResumePoint(r.steps)
		if startIdx > 0 {
			fmt.Printf("Resuming from step %d/%d\n\n", startIdx+1, len(r.steps))
		}
	}

	// Run each step
	for i := startIdx; i < len(r.steps); i++ {
		step := r.steps[i]

		// Check if step should be skipped
		if step.CanSkip(r.ctx.Options) {
			r.printStatus(i, step.Name(), "SKIP", "skipped by options")
			r.state.MarkSkipped(step.ID())
			r.state.Save()
			continue
		}

		// Check if step is already complete
		complete, err := step.Check(r.ctx)
		if err != nil {
			r.printStatus(i, step.Name(), "WARN", fmt.Sprintf("check failed: %v", err))
		}
		if complete {
			r.printStatus(i, step.Name(), "OK", "already complete")
			r.state.MarkComplete(step.ID())
			r.state.Save()
			continue
		}

		// Run the step
		r.printRunning(i, step.ActiveForm())
		r.state.CurrentStep = step.ID()
		r.state.Save()

		result := step.Run(r.ctx)

		switch result.Status {
		case StepComplete:
			r.printStatus(i, step.Name(), "OK", result.Message)
			if r.verbose && result.Details != "" {
				fmt.Println(result.Details)
			}
			r.state.MarkComplete(step.ID())

		case StepSkipped:
			r.printStatus(i, step.Name(), "SKIP", result.Message)
			r.state.MarkSkipped(step.ID())

		case StepFailed:
			r.printStatus(i, step.Name(), "FAIL", result.Message)
			if result.Error != nil {
				fmt.Printf("  Error: %v\n", result.Error)
			}
			if result.RepairHint != "" {
				fmt.Printf("  Hint: %s\n", result.RepairHint)
			}
			r.state.MarkFailed(step.ID(), result.Error)
			r.state.Save()
			return fmt.Errorf("installation failed at step %s: %s", step.ID(), result.Message)

		case StepNeedsSudo:
			// Run sudo command directly (interactive prompt in terminal)
			fmt.Printf("  Sudo required: %s\n", result.Message)
			sudoArgs := append([]string{"-p", "Password: ", result.SudoCmd}, result.SudoArgs...)
			cmd := exec.Command("sudo", sudoArgs...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				r.printStatus(i, step.Name(), "FAIL", "sudo command failed")
				fmt.Printf("  Error: %v\n", err)
				r.state.MarkFailed(step.ID(), err)
				r.state.Save()
				return fmt.Errorf("sudo command failed at step %s: %w", step.ID(), err)
			}

			r.printStatus(i, step.Name(), "OK", "completed with sudo")
			r.state.MarkComplete(step.ID())
		}

		r.state.Save()
	}

	// Clean up lock file
	lockFile := filepath.Join(r.ctx.HomeDir, ".vibe", "install.lock")
	os.Remove(lockFile)

	// Delete state file on success
	r.state.Delete()

	fmt.Println()
	fmt.Println("Installation complete!")
	fmt.Println()
	r.printSummary()

	return nil
}

func (r *Runner) printRunning(idx int, msg string) {
	fmt.Printf("[%2d/%d] %s...\n", idx+1, len(r.steps), msg)
}

func (r *Runner) printStatus(idx int, name, status, msg string) {
	icon := "  "
	switch status {
	case "OK":
		icon = "OK"
	case "SKIP":
		icon = "--"
	case "FAIL":
		icon = "!!"
	case "WARN":
		icon = "??"
	}
	fmt.Printf("[%2d/%d] [%s] %s: %s\n", idx+1, len(r.steps), icon, name, msg)
}

func (r *Runner) printSummary() {
	fmt.Println("What's been installed:")
	fmt.Println("  - Vibe marketplace added")
	fmt.Println("  - Plugins installed")
	fmt.Println("  - Permissions configured")
	fmt.Println("  - MCP servers configured")
	fmt.Println()

	// Show missing tools warning when --no-brew was used
	if r.ctx.Options.NoBrew {
		missing := MissingBrewManagedTools()
		if len(missing) > 0 {
			fmt.Println("⚠ Missing tools (install manually for best results):")
			for _, tool := range missing {
				fmt.Printf("    - %s (%s)\n", tool.Command, tool.Description)
				if tool.InstallHint != "" {
					fmt.Printf("      %s\n", tool.InstallHint)
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("Vibe CLI commands:")
	fmt.Println("  vibe update     - Download latest vibe and refresh everything")
	fmt.Println("  vibe status     - Show installation status")
	fmt.Println("  vibe plugins    - List available plugins")
	fmt.Println("  vibe doctor     - Diagnose and fix issues")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'vibe agent' and tell it to 'configure vibe'")
	fmt.Println("  2. Restart Claude Code once that is completed")
	fmt.Println("  3. Vibe")
}

// GetContext returns the installation context
func (r *Runner) GetContext() *Context {
	return r.ctx
}

// GetState returns the installation state
func (r *Runner) GetState() *State {
	return r.state
}

// GetSteps returns all steps
func (r *Runner) GetSteps() []Step {
	return r.steps
}
