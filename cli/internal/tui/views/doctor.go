// Package views provides TUI view implementations for the vibe CLI.
package views

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/doctor"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui/components"
)

// DoctorState represents the current state of the doctor view.
type DoctorState int

const (
	// DoctorStateRunningChecks indicates checks are being run.
	DoctorStateRunningChecks DoctorState = iota
	// DoctorStateShowingResults indicates checks are complete and results are shown.
	DoctorStateShowingResults
	// DoctorStatePromptRepair indicates prompting user for repair.
	DoctorStatePromptRepair
	// DoctorStateRunningRepairs indicates repairs are being run.
	DoctorStateRunningRepairs
	// DoctorStateComplete indicates everything is done.
	DoctorStateComplete
)

// DoctorView is the TUI view for the vibe doctor command.
type DoctorView struct {
	theme        tui.Theme
	state        DoctorState
	checks       []doctor.Check
	results      []doctor.CheckResult
	repairResults []doctor.RepairResult
	currentCheck int
	spinner      spinner.Model
	confirm      components.Confirm
	width        int
	height       int
	err          error
	quitting     bool
}

// DoctorOption configures the DoctorView.
type DoctorOption func(*DoctorView)

// WithDoctorTheme sets a custom theme.
func WithDoctorTheme(theme tui.Theme) DoctorOption {
	return func(v *DoctorView) {
		v.theme = theme
	}
}

// NewDoctorView creates a new DoctorView.
func NewDoctorView(opts ...DoctorOption) *DoctorView {
	s := spinner.New()
	s.Spinner = spinner.Dot

	v := &DoctorView{
		theme:        tui.DefaultTheme,
		state:        DoctorStateRunningChecks,
		checks:       doctor.AllChecks(),
		results:      make([]doctor.CheckResult, 0),
		currentCheck: 0,
		spinner:      s,
	}

	for _, opt := range opts {
		opt(v)
	}

	v.spinner.Style = lipgloss.NewStyle().Foreground(v.theme.Primary)

	return v
}

// Init implements tui.View.
func (v *DoctorView) Init() tea.Cmd {
	return tea.Batch(
		v.spinner.Tick,
		v.runNextCheck(),
	)
}

// checkResultMsg is sent when a check completes.
type checkResultMsg struct {
	result doctor.CheckResult
	err    error
}

// repairResultMsg is sent when a repair completes.
type repairResultMsg struct {
	result doctor.RepairResult
	err    error
}

// allRepairsCompleteMsg is sent when all repairs are done.
type allRepairsCompleteMsg struct {
	results []doctor.RepairResult
}

// repairRunner implements tea.ExecCommand to run repairs with proper stdin access.
// This allows sudo and other interactive commands to work correctly.
type repairRunner struct {
	checks  []doctor.Check
	results []doctor.CheckResult
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	// repairResults stores the results after Run() completes
	repairResults []doctor.RepairResult
}

func (r *repairRunner) SetStdin(stdin io.Reader)   { r.stdin = stdin }
func (r *repairRunner) SetStdout(stdout io.Writer) { r.stdout = stdout }
func (r *repairRunner) SetStderr(stderr io.Writer) { r.stderr = stderr }

func (r *repairRunner) Run() error {
	checkMap := make(map[string]doctor.Check)
	for _, check := range r.checks {
		checkMap[check.Name()] = check
	}

	for _, result := range r.results {
		if result.Status == doctor.StatusPass {
			continue
		}

		check, exists := checkMap[result.Name]
		if !exists {
			r.repairResults = append(r.repairResults, doctor.RepairResult{
				CheckName: result.Name,
				Repaired:  false,
				Skipped:   true,
				Message:   "Check not found",
			})
			continue
		}

		if !check.CanRepair() {
			r.repairResults = append(r.repairResults, doctor.RepairResult{
				CheckName: result.Name,
				Repaired:  false,
				Skipped:   true,
				Message:   fmt.Sprintf("Cannot auto-repair: %s", result.RepairHint),
			})
			continue
		}

		// Print what we're repairing so user knows what's happening
		fmt.Fprintf(r.stdout, "Repairing: %s\n", check.Description())

		err := check.Repair()
		if err != nil {
			r.repairResults = append(r.repairResults, doctor.RepairResult{
				CheckName: result.Name,
				Repaired:  false,
				Skipped:   false,
				Error:     err,
				Message:   fmt.Sprintf("Repair failed: %v", err),
			})
			fmt.Fprintf(r.stderr, "  Failed: %v\n", err)
		} else {
			r.repairResults = append(r.repairResults, doctor.RepairResult{
				CheckName: result.Name,
				Repaired:  true,
				Skipped:   false,
				Message:   "Repair successful",
			})
			fmt.Fprintf(r.stdout, "  Success\n")
		}
	}

	fmt.Fprintln(r.stdout, "\nPress Enter to continue...")
	// Wait for user to acknowledge results before returning to TUI
	buf := make([]byte, 1)
	r.stdin.Read(buf)

	return nil
}

// runNextCheck runs the next check in the list.
func (v *DoctorView) runNextCheck() tea.Cmd {
	if v.currentCheck >= len(v.checks) {
		return nil
	}

	check := v.checks[v.currentCheck]
	return func() tea.Msg {
		// Add a small delay to show the spinner
		time.Sleep(100 * time.Millisecond)
		result := check.Run()
		return checkResultMsg{result: result}
	}
}

// runRepairs runs all repairs for failed checks.
// Uses tea.Exec to temporarily suspend the TUI, allowing interactive commands
// like sudo to properly read from stdin.
func (v *DoctorView) runRepairs() tea.Cmd {
	runner := &repairRunner{
		checks:  v.checks,
		results: v.results,
	}

	return tea.Exec(runner, func(err error) tea.Msg {
		return allRepairsCompleteMsg{results: runner.repairResults}
	})
}

// Update implements tui.View.
func (v *DoctorView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			v.quitting = true
			return v, tea.Quit
		case "esc":
			if v.state == DoctorStatePromptRepair {
				v.state = DoctorStateComplete
				return v, nil
			}
		}

	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case checkResultMsg:
		if msg.err != nil {
			v.err = msg.err
			v.state = DoctorStateComplete
			return v, nil
		}

		v.results = append(v.results, msg.result)
		v.currentCheck++

		if v.currentCheck >= len(v.checks) {
			// All checks complete
			if doctor.HasIssues(v.results) {
				v.state = DoctorStatePromptRepair
				v.confirm = components.NewConfirm(
					"Would you like to repair?",
					components.WithConfirmDefault(false),
				)
			} else {
				v.state = DoctorStateComplete
			}
		} else {
			cmds = append(cmds, v.runNextCheck())
		}

	case components.ConfirmResultMsg:
		if msg.Confirmed {
			v.state = DoctorStateRunningRepairs
			v.repairResults = nil
			cmds = append(cmds, v.runRepairs())
		} else {
			v.state = DoctorStateComplete
		}

	case components.ConfirmCancelledMsg:
		v.state = DoctorStateComplete

	case allRepairsCompleteMsg:
		// Run repairs completed, use the results from the runner
		v.repairResults = msg.results
		v.state = DoctorStateComplete
	}

	// Update confirm dialog if in prompt state
	if v.state == DoctorStatePromptRepair {
		var cmd tea.Cmd
		v.confirm, cmd = v.confirm.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

// View implements tui.View.
func (v *DoctorView) View() string {
	if v.quitting {
		return ""
	}

	var b strings.Builder

	// Header
	headerStyle := v.theme.HeaderStyle()
	b.WriteString(headerStyle.Render("Vibe Doctor"))
	b.WriteString("\n\n")

	switch v.state {
	case DoctorStateRunningChecks:
		b.WriteString(v.renderChecksInProgress())

	case DoctorStateShowingResults, DoctorStatePromptRepair:
		b.WriteString(v.renderResults())
		if v.state == DoctorStatePromptRepair {
			b.WriteString("\n\n")
			b.WriteString(v.confirm.View())
		}

	case DoctorStateRunningRepairs:
		b.WriteString(v.renderResults())
		b.WriteString("\n\n")
		b.WriteString(v.spinner.View() + " Running repairs...")

	case DoctorStateComplete:
		b.WriteString(v.renderResults())
		if len(v.repairResults) > 0 {
			b.WriteString("\n\n")
			b.WriteString(v.renderRepairResults())
		}
		b.WriteString("\n\n")
		b.WriteString(v.renderSummary())
	}

	// Help text
	b.WriteString("\n\n")
	helpStyle := v.theme.HelpStyle()
	if v.state == DoctorStateComplete {
		b.WriteString(helpStyle.Render("Press q to quit"))
	} else if v.state == DoctorStatePromptRepair {
		b.WriteString(helpStyle.Render("Press y to repair, n to skip, q to quit"))
	} else {
		b.WriteString(helpStyle.Render("Press q to quit"))
	}

	return b.String()
}

// renderChecksInProgress renders the checks currently running.
func (v *DoctorView) renderChecksInProgress() string {
	var b strings.Builder

	for i, check := range v.checks {
		var icon string
		var style lipgloss.Style

		if i < v.currentCheck {
			// Completed check
			result := v.results[i]
			switch result.Status {
			case doctor.StatusPass:
				icon = v.theme.SuccessIcon()
				style = lipgloss.NewStyle().Foreground(v.theme.Success)
			case doctor.StatusFail:
				icon = v.theme.ErrorIcon()
				style = lipgloss.NewStyle().Foreground(v.theme.Error)
			case doctor.StatusWarning:
				icon = v.theme.WarningIcon()
				style = lipgloss.NewStyle().Foreground(v.theme.Warning)
			}
		} else if i == v.currentCheck {
			// Currently running
			icon = v.spinner.View()
			style = lipgloss.NewStyle().Foreground(v.theme.Primary)
		} else {
			// Pending
			icon = "○"
			style = lipgloss.NewStyle().Foreground(v.theme.Muted)
		}

		line := fmt.Sprintf("%s %s", icon, style.Render(check.Description()))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// renderResults renders all check results.
func (v *DoctorView) renderResults() string {
	var b strings.Builder

	for i, result := range v.results {
		var icon string
		var style lipgloss.Style

		switch result.Status {
		case doctor.StatusPass:
			icon = v.theme.SuccessIcon()
			style = lipgloss.NewStyle().Foreground(v.theme.Success)
		case doctor.StatusFail:
			icon = v.theme.ErrorIcon()
			style = lipgloss.NewStyle().Foreground(v.theme.Error)
		case doctor.StatusWarning:
			icon = v.theme.WarningIcon()
			style = lipgloss.NewStyle().Foreground(v.theme.Warning)
		}

		checkDesc := v.checks[i].Description()
		line := fmt.Sprintf("%s %s", icon, style.Render(checkDesc))
		b.WriteString(line)
		b.WriteString("\n")

		// Show message for failures/warnings
		if result.Status != doctor.StatusPass {
			mutedStyle := lipgloss.NewStyle().Foreground(v.theme.Muted)
			b.WriteString(fmt.Sprintf("    %s\n", mutedStyle.Render(result.Message)))
			if result.RepairHint != "" {
				b.WriteString(fmt.Sprintf("    %s\n", mutedStyle.Render("Hint: "+result.RepairHint)))
			}
		}
	}

	return b.String()
}

// renderRepairResults renders the repair results.
func (v *DoctorView) renderRepairResults() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Repair Results:"))
	b.WriteString("\n")

	for _, result := range v.repairResults {
		var icon string
		var style lipgloss.Style

		if result.Repaired {
			icon = v.theme.SuccessIcon()
			style = lipgloss.NewStyle().Foreground(v.theme.Success)
		} else if result.Skipped {
			icon = "○"
			style = lipgloss.NewStyle().Foreground(v.theme.Muted)
		} else {
			icon = v.theme.ErrorIcon()
			style = lipgloss.NewStyle().Foreground(v.theme.Error)
		}

		line := fmt.Sprintf("%s %s: %s", icon, result.CheckName, style.Render(result.Message))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// renderSummary renders the final summary.
func (v *DoctorView) renderSummary() string {
	pass, fail, warning := doctor.CountByStatus(v.results)

	var b strings.Builder

	if fail == 0 && warning == 0 {
		successStyle := lipgloss.NewStyle().Foreground(v.theme.Success).Bold(true)
		b.WriteString(successStyle.Render("All checks passed! Vibe is healthy."))
	} else {
		summaryStyle := lipgloss.NewStyle().Bold(true)
		b.WriteString(summaryStyle.Render(fmt.Sprintf(
			"Summary: %d passed, %d failed, %d warnings",
			pass, fail, warning,
		)))

		if len(v.repairResults) > 0 {
			repaired := 0
			repairFailed := 0
			for _, r := range v.repairResults {
				if r.Repaired {
					repaired++
				} else if !r.Skipped {
					repairFailed++
				}
			}
			if repaired > 0 || repairFailed > 0 {
				b.WriteString("\n")
				b.WriteString(fmt.Sprintf("Repairs: %d succeeded, %d failed", repaired, repairFailed))
			}
		}
	}

	return b.String()
}

// HasIssues returns whether any issues were found.
func (v *DoctorView) HasIssues() bool {
	return doctor.HasIssues(v.results)
}

// Results returns the check results.
func (v *DoctorView) Results() []doctor.CheckResult {
	return v.results
}

// RepairResults returns the repair results.
func (v *DoctorView) RepairResults() []doctor.RepairResult {
	return v.repairResults
}
