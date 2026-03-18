package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
)

// Progress combines a spinner and progress bar for package-manager style output.
type Progress struct {
	spinner      spinner.Model
	progress     progress.Model
	theme        tui.Theme
	percent      float64
	message      string
	steps        []Step
	currentStep  int
	showSpinner  bool
	showProgress bool
	width        int
}

// Step represents a step in a multi-step process.
type Step struct {
	Name   string
	Status StepStatus
}

// StepStatus indicates the state of a step.
type StepStatus int

const (
	// StepPending means the step has not started.
	StepPending StepStatus = iota
	// StepRunning means the step is currently executing.
	StepRunning
	// StepComplete means the step finished successfully.
	StepComplete
	// StepFailed means the step encountered an error.
	StepFailed
	// StepSkipped means the step was skipped.
	StepSkipped
)

// ProgressOption configures a Progress component.
type ProgressOption func(*Progress)

// WithProgressTheme sets a custom theme.
func WithProgressTheme(theme tui.Theme) ProgressOption {
	return func(p *Progress) {
		p.theme = theme
		p.applyTheme()
	}
}

// WithProgressWidth sets the width of the progress bar.
func WithProgressWidth(width int) ProgressOption {
	return func(p *Progress) {
		p.width = width
		p.progress.Width = width
	}
}

// WithSpinner enables/disables the spinner.
func WithSpinner(show bool) ProgressOption {
	return func(p *Progress) {
		p.showSpinner = show
	}
}

// WithProgressBar enables/disables the progress bar.
func WithProgressBar(show bool) ProgressOption {
	return func(p *Progress) {
		p.showProgress = show
	}
}

// WithSteps initializes the progress with steps.
func WithSteps(steps []string) ProgressOption {
	return func(p *Progress) {
		p.steps = make([]Step, len(steps))
		for i, name := range steps {
			p.steps[i] = Step{Name: name, Status: StepPending}
		}
	}
}

// NewProgress creates a new Progress component.
func NewProgress(opts ...ProgressOption) Progress {
	s := spinner.New()
	s.Spinner = spinner.Dot

	p := progress.New(progress.WithDefaultGradient())

	prog := Progress{
		spinner:      s,
		progress:     p,
		theme:        tui.DefaultTheme,
		showSpinner:  true,
		showProgress: true,
		width:        40,
	}

	for _, opt := range opts {
		opt(&prog)
	}

	prog.applyTheme()

	return prog
}

// applyTheme applies the theme to spinner and progress bar.
func (p *Progress) applyTheme() {
	p.spinner.Style = lipgloss.NewStyle().Foreground(p.theme.Primary)
	p.progress = progress.New(
		progress.WithGradient(string(p.theme.Primary), string(p.theme.Secondary)),
	)
	p.progress.Width = p.width
}

// Init implements tea.Model.
func (p Progress) Init() tea.Cmd {
	return p.spinner.Tick
}

// Update implements tea.Model.
func (p Progress) Update(msg tea.Msg) (Progress, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		p.spinner, cmd = p.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := p.progress.Update(msg)
		p.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return p, tea.Batch(cmds...)
}

// View implements tea.Model.
func (p Progress) View() string {
	var parts []string

	// Render steps if present
	if len(p.steps) > 0 {
		parts = append(parts, p.renderSteps())
	}

	// Current status line with spinner
	if p.message != "" {
		statusLine := ""
		if p.showSpinner {
			statusLine = p.spinner.View() + " "
		}
		statusLine += p.message
		parts = append(parts, statusLine)
	}

	// Progress bar
	if p.showProgress && p.percent > 0 {
		parts = append(parts, p.progress.ViewAs(p.percent))
	}

	return strings.Join(parts, "\n")
}

// renderSteps renders the step list.
func (p Progress) renderSteps() string {
	var lines []string

	for i, step := range p.steps {
		var icon string
		var style lipgloss.Style

		switch step.Status {
		case StepPending:
			icon = "○"
			style = lipgloss.NewStyle().Foreground(p.theme.Muted)
		case StepRunning:
			icon = p.spinner.View()
			style = lipgloss.NewStyle().Foreground(p.theme.Primary)
		case StepComplete:
			icon = p.theme.SuccessIcon()
			style = lipgloss.NewStyle().Foreground(p.theme.Success)
		case StepFailed:
			icon = p.theme.ErrorIcon()
			style = lipgloss.NewStyle().Foreground(p.theme.Error)
		case StepSkipped:
			icon = "○"
			style = lipgloss.NewStyle().Foreground(p.theme.Muted).Strikethrough(true)
		}

		line := fmt.Sprintf("%s %s", icon, style.Render(step.Name))

		// Indent sub-steps if this is not the current step
		if i != p.currentStep && step.Status == StepPending {
			line = lipgloss.NewStyle().Foreground(p.theme.Muted).Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// SetPercent sets the progress bar percentage (0.0 to 1.0).
func (p *Progress) SetPercent(percent float64) tea.Cmd {
	p.percent = percent
	return p.progress.SetPercent(percent)
}

// SetMessage sets the current status message.
func (p *Progress) SetMessage(msg string) {
	p.message = msg
}

// StartStep marks a step as running.
func (p *Progress) StartStep(index int) {
	if index >= 0 && index < len(p.steps) {
		p.currentStep = index
		p.steps[index].Status = StepRunning
	}
}

// CompleteStep marks a step as complete.
func (p *Progress) CompleteStep(index int) {
	if index >= 0 && index < len(p.steps) {
		p.steps[index].Status = StepComplete
	}
}

// FailStep marks a step as failed.
func (p *Progress) FailStep(index int) {
	if index >= 0 && index < len(p.steps) {
		p.steps[index].Status = StepFailed
	}
}

// SkipStep marks a step as skipped.
func (p *Progress) SkipStep(index int) {
	if index >= 0 && index < len(p.steps) {
		p.steps[index].Status = StepSkipped
	}
}

// CurrentStep returns the index of the current step.
func (p Progress) CurrentStep() int {
	return p.currentStep
}

// StepCount returns the total number of steps.
func (p Progress) StepCount() int {
	return len(p.steps)
}

// TickCmd returns a command to tick the spinner.
func (p Progress) TickCmd() tea.Cmd {
	return p.spinner.Tick
}

// ProgressCompleteMsg is sent when progress reaches 100%.
type ProgressCompleteMsg struct{}

// StepCompleteMsg is sent when a step completes.
type StepCompleteMsg struct {
	Index int
	Name  string
}
