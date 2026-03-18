package views

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/install"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui/components"
)

// InstallState represents the state of the install view
type InstallState int

const (
	InstallStateWelcome InstallState = iota
	InstallStateConfirm
	InstallStateRunning
	InstallStateSudoPrompt
	InstallStateError
	InstallStateComplete
)

// InstallView handles the installation TUI
type InstallView struct {
	theme        tui.Theme
	state        InstallState
	options      *install.Options
	ctx          *install.Context
	installState *install.State
	steps        []install.Step
	currentStep  int
	stepResults  []install.StepResult
	spinner      spinner.Model
	confirm      components.Confirm
	progress     components.Progress
	progressBar  progress.Model // Animated progress bar for current step
	errorMsg     string
	errorHint    string
	width        int
	height       int

	// Real-time log display
	logLines    []string // Recent log messages
	maxLogLines int      // Maximum number of log lines to keep

	// Sudo handling
	pendingSudoCmd  string   // Command pending sudo execution
	pendingSudoArgs []string // Args for pending sudo command
}

// InstallViewOption is a functional option for InstallView
type InstallViewOption func(*InstallView)

// WithInstallTheme sets the theme
func WithInstallTheme(theme tui.Theme) InstallViewOption {
	return func(v *InstallView) {
		v.theme = theme
	}
}

// WithInstallOptions sets the install options
func WithInstallOptions(opts *install.Options) InstallViewOption {
	return func(v *InstallView) {
		v.options = opts
	}
}

// NewInstallView creates a new install view
func NewInstallView(opts ...InstallViewOption) *InstallView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	// Create progress bar with custom styling (shows overall install progress)
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(50),
	)

	v := &InstallView{
		theme:       tui.DefaultTheme,
		state:       InstallStateWelcome,
		options:     install.DefaultOptions(),
		spinner:     s,
		progressBar: p,
		logLines:    make([]string, 0),
		maxLogLines: 8,
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// stepResultMsg is sent when a step completes
type stepResultMsg struct {
	stepIdx int
	result  install.StepResult
}

// logMsg is sent when a log message is received
type logMsg struct {
	message string
}

// installSudoCompleteMsg signals that sudo command completed
type installSudoCompleteMsg struct {
	err error
}

// installExecCompleteMsg signals that an exec command completed
type installExecCompleteMsg struct {
	err error
}

// Init initializes the view
func (v *InstallView) Init() tea.Cmd {
	// Create context
	ctx, err := install.NewContext(v.options)
	if err != nil {
		v.errorMsg = "Failed to initialize: " + err.Error()
		v.state = InstallStateError
		return nil
	}
	v.ctx = ctx

	// Set up log channel for real-time output
	v.ctx.LogChan = make(chan string, 100)

	// Set up logging function to send to channel
	v.ctx.Log = func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		select {
		case v.ctx.LogChan <- msg:
		default:
			// Channel full, skip message
		}
	}

	// Get all steps
	v.steps = install.AllSteps()
	v.stepResults = make([]install.StepResult, len(v.steps))

	// Build step names for progress
	stepNames := make([]string, len(v.steps))
	for i, step := range v.steps {
		stepNames[i] = step.Name()
	}

	v.progress = components.NewProgress(
		components.WithProgressTheme(v.theme),
		components.WithSteps(stepNames),
	)

	// Create confirm dialog
	v.confirm = components.NewConfirm(
		"Ready to install vibe?",
		components.WithConfirmTheme(v.theme),
		components.WithConfirmLabels("Install", "Cancel"),
	)

	// Load existing state if resuming
	if v.options.Resume {
		state, _ := install.LoadState()
		if state != nil {
			v.installState = state
			v.currentStep = state.GetResumePoint(v.steps)

			// Mark completed steps
			for i := 0; i < v.currentStep; i++ {
				if state.IsStepSkipped(v.steps[i].ID()) {
					v.progress.SkipStep(i)
				} else {
					v.progress.CompleteStep(i)
				}
			}
		}
	} else {
		v.installState = install.NewState()
	}

	return tea.Batch(v.spinner.Tick, v.progress.Init())
}

// Update handles messages
func (v *InstallView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return v, tea.Quit
		case "q":
			if v.state == InstallStateComplete || v.state == InstallStateError {
				return v, tea.Quit
			}
		case "r":
			if v.state == InstallStateError {
				// Retry from current step
				v.state = InstallStateRunning
				return v, v.runCurrentStep()
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case stepResultMsg:
		return v, v.handleStepResult(msg)

	case logMsg:
		// Add log message to buffer
		v.logLines = append(v.logLines, msg.message)
		if len(v.logLines) > v.maxLogLines {
			v.logLines = v.logLines[1:]
		}
		// Continue listening for more logs
		if v.state == InstallStateRunning {
			return v, v.waitForLog()
		}
		return v, nil

	case logTickMsg:
		// Continue listening for logs if still running
		if v.state == InstallStateRunning {
			return v, v.waitForLog()
		}
		return v, nil

	case progress.FrameMsg:
		// Handle progress bar animation frames
		var cmd tea.Cmd
		progressModel, cmd := v.progressBar.Update(msg)
		v.progressBar = progressModel.(progress.Model)
		return v, cmd

	case components.ConfirmResultMsg:
		if msg.Confirmed {
			v.state = InstallStateRunning
			return v, v.runCurrentStep()
		} else {
			return v, tea.Quit
		}

	case components.ConfirmCancelledMsg:
		return v, tea.Quit

	case installSudoCompleteMsg:
		if msg.err != nil {
			// Sudo failed
			v.progress.FailStep(v.currentStep)
			v.state = InstallStateError
			v.errorMsg = "Sudo command failed"
			v.errorHint = msg.err.Error()
			return v, nil
		}
		// Sudo succeeded, mark step complete and continue
		step := v.steps[v.currentStep]
		v.progress.CompleteStep(v.currentStep)
		v.installState.MarkComplete(step.ID())
		v.installState.Save()

		// Move to next step
		v.currentStep++
		progressPercent := float64(v.currentStep) / float64(len(v.steps))
		progressCmd := v.progressBar.SetPercent(progressPercent)

		if v.currentStep >= len(v.steps) {
			v.state = InstallStateComplete
			v.installState.Delete()
			return v, progressCmd
		}
		return v, tea.Batch(progressCmd, v.runCurrentStep())

	case installExecCompleteMsg:
		if msg.err != nil {
			v.progress.FailStep(v.currentStep)
			v.state = InstallStateError
			v.errorMsg = v.steps[v.currentStep].Name() + " failed"
			v.errorHint = "Try running manually: " + msg.err.Error()
			return v, nil
		}
		// Exec succeeded, mark step complete and continue
		step := v.steps[v.currentStep]
		v.progress.CompleteStep(v.currentStep)
		v.installState.MarkComplete(step.ID())
		v.installState.Save()

		// Move to next step
		v.currentStep++
		progressPercent := float64(v.currentStep) / float64(len(v.steps))
		progressCmd := v.progressBar.SetPercent(progressPercent)

		if v.currentStep >= len(v.steps) {
			v.state = InstallStateComplete
			v.installState.Delete()
			return v, progressCmd
		}
		return v, tea.Batch(progressCmd, v.runCurrentStep())
	}

	// Update components based on state
	switch v.state {
	case InstallStateWelcome:
		// Auto-advance after showing welcome
		v.state = InstallStateConfirm
		return v, nil

	case InstallStateConfirm:
		var cmd tea.Cmd
		v.confirm, cmd = v.confirm.Update(msg)
		cmds = append(cmds, cmd)

	case InstallStateRunning:
		var cmd tea.Cmd
		v.progress, cmd = v.progress.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

// logTickMsg is sent periodically to check for logs
type logTickMsg struct{}


// waitForLog returns a command that waits for a log message with timeout
func (v *InstallView) waitForLog() tea.Cmd {
	return func() tea.Msg {
		if v.ctx == nil || v.ctx.LogChan == nil {
			return logTickMsg{}
		}
		select {
		case msg, ok := <-v.ctx.LogChan:
			if !ok {
				return logTickMsg{}
			}
			return logMsg{message: msg}
		case <-time.After(100 * time.Millisecond):
			// Timeout - return tick to keep checking
			return logTickMsg{}
		}
	}
}

// runCurrentStep runs the current installation step
func (v *InstallView) runCurrentStep() tea.Cmd {
	if v.currentStep >= len(v.steps) {
		v.state = InstallStateComplete
		return nil
	}

	// Clear log lines for new step
	v.logLines = make([]string, 0)

	step := v.steps[v.currentStep]
	v.progress.StartStep(v.currentStep)
	v.progress.SetMessage(step.ActiveForm())

	runStep := func() tea.Msg {
		// Check if step should be skipped
		if step.CanSkip(v.options) {
			return stepResultMsg{
				stepIdx: v.currentStep,
				result:  install.Skip("skipped by options"),
			}
		}

		// Check if step is already complete
		complete, _ := step.Check(v.ctx)
		if complete {
			return stepResultMsg{
				stepIdx: v.currentStep,
				result:  install.Success("already complete"),
			}
		}

		// Run the step
		result := step.Run(v.ctx)
		return stepResultMsg{
			stepIdx: v.currentStep,
			result:  result,
		}
	}

	// Start the step and log listener
	return tea.Batch(runStep, v.waitForLog())
}

// handleStepResult processes the result of a step
func (v *InstallView) handleStepResult(msg stepResultMsg) tea.Cmd {
	v.stepResults[msg.stepIdx] = msg.result
	step := v.steps[msg.stepIdx]

	switch msg.result.Status {
	case install.StepComplete:
		v.progress.CompleteStep(msg.stepIdx)
		v.installState.MarkComplete(step.ID())
		v.installState.Save()

	case install.StepSkipped:
		v.progress.SkipStep(msg.stepIdx)
		v.installState.MarkSkipped(step.ID())
		v.installState.Save()

	case install.StepFailed:
		v.progress.FailStep(msg.stepIdx)
		v.installState.MarkFailed(step.ID(), msg.result.Error)
		v.installState.Save()

		v.errorMsg = msg.result.Message
		v.errorHint = msg.result.RepairHint
		v.state = InstallStateError
		return nil

	case install.StepNeedsSudo:
		// Step needs sudo - use tea.ExecProcess to properly suspend TUI
		// and give sudo full terminal control for password input
		v.pendingSudoCmd = msg.result.SudoCmd
		v.pendingSudoArgs = msg.result.SudoArgs

		// Build sudo command with custom prompt
		sudoArgs := append([]string{"-p", "\nPassword required for " + step.Name() + ": ", msg.result.SudoCmd}, msg.result.SudoArgs...)
		cmd := exec.Command("sudo", sudoArgs...)

		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			return installSudoCompleteMsg{err: err}
		})

	case install.StepNeedsExec:
		// Step needs full terminal access - suspend TUI so the user can
		// interact with any prompts (e.g., SSH host key verification).
		cmd := exec.Command(msg.result.ExecCmd, msg.result.ExecArgs...)
		cmd.Env = os.Environ()
		if msg.result.ExecStdin != "" {
			cmd.Stdin = strings.NewReader(msg.result.ExecStdin)
		}

		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			return installExecCompleteMsg{err: err}
		})
	}

	// Move to next step
	v.currentStep++

	// Update overall progress bar
	progressPercent := float64(v.currentStep) / float64(len(v.steps))
	progressCmd := v.progressBar.SetPercent(progressPercent)

	if v.currentStep >= len(v.steps) {
		v.state = InstallStateComplete
		v.installState.Delete()
		return progressCmd
	}

	return tea.Batch(progressCmd, v.runCurrentStep())
}

// View renders the install view
func (v *InstallView) View() string {
	var sb strings.Builder

	// Header
	header := v.theme.HeaderStyle().Render("Vibe Installer")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	switch v.state {
	case InstallStateWelcome:
		sb.WriteString(v.renderWelcome())

	case InstallStateConfirm:
		sb.WriteString(v.renderConfirm())

	case InstallStateRunning:
		sb.WriteString(v.renderRunning())

	case InstallStateError:
		sb.WriteString(v.renderError())

	case InstallStateComplete:
		sb.WriteString(v.renderComplete())
	}

	return sb.String()
}

func (v *InstallView) renderWelcome() string {
	if v.options.NoBrew {
		return `This will install vibe and all required dependencies.

What will be installed:
  - Claude Code CLI
  - Vibe marketplace and plugins
  - MCP server configurations

What will be SKIPPED (--no-brew):
  - Homebrew packages (jq, yq, gh, python, etc.)
  - Node.js/npm, gcloud CLI, Databricks CLI

Missing tools will be listed at the end for manual installation.

`
	}

	return `This will install vibe and all required dependencies.

What will be installed:
  - Homebrew packages (jq, yq, gh, python, etc.)
  - Claude Code CLI
  - Vibe marketplace and plugins
  - MCP server configurations

`
}

func (v *InstallView) renderConfirm() string {
	var sb strings.Builder

	sb.WriteString(v.renderWelcome())
	sb.WriteString("\n")

	// Show options
	sb.WriteString(v.theme.MutedStyle().Render("Options:"))
	sb.WriteString("\n")
	if v.options.SkipJAMF {
		sb.WriteString("  - Skip JAMF: Yes\n")
	}
	if v.options.SkipPlugins {
		sb.WriteString("  - Skip plugins: Yes\n")
	}
	if v.options.NoBrew {
		sb.WriteString("  - No Homebrew: Yes (missing tools will be listed)\n")
	}
	if v.options.ForceReinstall {
		sb.WriteString("  - Force reinstall: Yes\n")
	}
	sb.WriteString("\n")

	sb.WriteString(v.confirm.View())

	return sb.String()
}

func (v *InstallView) renderRunning() string {
	var sb strings.Builder

	sb.WriteString(v.progress.View())
	sb.WriteString("\n\n")

	if v.currentStep < len(v.steps) {
		step := v.steps[v.currentStep]

		// Show current step name with spinner
		sb.WriteString("  ")
		sb.WriteString(v.spinner.View())
		sb.WriteString(" ")
		sb.WriteString(step.ActiveForm())
		sb.WriteString("\n\n")

		// Show overall progress bar with percentage
		progressPercent := float64(v.currentStep) / float64(len(v.steps)) * 100
		sb.WriteString("  ")
		sb.WriteString(v.progressBar.View())
		sb.WriteString(fmt.Sprintf(" %3.0f%%", progressPercent))
		sb.WriteString("\n")

		// Show real-time log output
		if len(v.logLines) > 0 {
			sb.WriteString("\n")

			// Styles for log lines
			completedStyle := v.theme.MutedStyle()
			activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")) // White

			for i, line := range v.logLines {
				// Truncate long lines
				if len(line) > 70 {
					line = line[:67] + "..."
				}

				isActive := i == len(v.logLines)-1
				if isActive {
					// Active line: spinner + white text
					sb.WriteString("  ")
					sb.WriteString(v.spinner.View())
					sb.WriteString(" ")
					sb.WriteString(activeStyle.Render(line))
				} else {
					// Completed line: checkmark + gray text
					sb.WriteString(completedStyle.Render("  ✓ " + line))
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

func (v *InstallView) renderError() string {
	var sb strings.Builder

	sb.WriteString(v.progress.View())
	sb.WriteString("\n\n")

	sb.WriteString(v.theme.ErrorStyle().Render("Error: " + v.errorMsg))
	sb.WriteString("\n")

	if v.errorHint != "" {
		sb.WriteString(v.theme.MutedStyle().Render("Hint: " + v.errorHint))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(v.theme.MutedStyle().Render("Press 'r' to retry, 'q' to quit"))

	return sb.String()
}

func (v *InstallView) renderComplete() string {
	var sb strings.Builder

	sb.WriteString(v.progress.View())
	sb.WriteString("\n\n")

	sb.WriteString(v.theme.SuccessStyle().Render("Installation complete!"))
	sb.WriteString("\n\n")

	sb.WriteString("What's been installed:\n")
	sb.WriteString("  " + v.theme.SuccessIcon() + " Vibe marketplace added\n")
	sb.WriteString("  " + v.theme.SuccessIcon() + " Plugins installed\n")
	sb.WriteString("  " + v.theme.SuccessIcon() + " Permissions configured\n")
	sb.WriteString("  " + v.theme.SuccessIcon() + " MCP servers configured\n")
	sb.WriteString("\n")

	// Show missing tools warning when --no-brew was used
	if v.options.NoBrew {
		missing := install.MissingBrewManagedTools()
		if len(missing) > 0 {
			warnStyle := lipgloss.NewStyle().Foreground(v.theme.Warning)
			sb.WriteString(warnStyle.Render(v.theme.WarningIcon()+" Missing tools (install manually for best results):"))
			sb.WriteString("\n")
			for _, tool := range missing {
				line := fmt.Sprintf("    - %s (%s)", tool.Command, tool.Description)
				if tool.InstallHint != "" {
					line += "\n      " + tool.InstallHint
				}
				sb.WriteString(warnStyle.Render(line))
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	// Detect shell for source command
	shell := os.Getenv("SHELL")
	shellName := filepath.Base(shell)
	var rcFile string
	switch shellName {
	case "zsh":
		rcFile = "~/.zshrc"
	case "bash":
		rcFile = "~/.bashrc"
	default:
		rcFile = "~/.bashrc"
	}

	sb.WriteString("Next steps:\n")
	sb.WriteString(fmt.Sprintf("  1. Source your shell config: %s\n", v.theme.TitleStyle().Render(fmt.Sprintf("source %s", rcFile))))
	sb.WriteString(fmt.Sprintf("  2. Run '%s' and tell it to 'configure vibe'\n", v.theme.TitleStyle().Render("vibe agent")))
	sb.WriteString("  3. Restart Claude Code once that is completed\n")
	sb.WriteString("  4. Vibe\n")
	sb.WriteString("\n")

	sb.WriteString(v.theme.MutedStyle().Render("Press 'q' to exit"))

	return sb.String()
}

// RunInstallTUI runs the install TUI
func RunInstallTUI(opts *install.Options) error {
	view := NewInstallView(WithInstallOptions(opts))
	app := tui.NewApp(view)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunInstallNonInteractive runs install without TUI
func RunInstallNonInteractive(opts *install.Options) error {
	runner, err := install.NewRunner(opts)
	if err != nil {
		return err
	}
	return runner.Run()
}
