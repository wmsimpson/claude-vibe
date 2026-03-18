// Package views provides Bubble Tea views for the vibe CLI.
package views

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/install"
	"github.com/wmsimpson/claude-vibe/cli/internal/marketplace"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui/components"
	"github.com/wmsimpson/claude-vibe/cli/internal/util"
)

// UpdateState represents the current state of the update view.
type UpdateState int

const (
	// UpdateStateLoading is the initial state while fetching releases.
	UpdateStateLoading UpdateState = iota
	// UpdateStateSelectVersion shows the version selection table.
	UpdateStateSelectVersion
	// UpdateStateConfirm shows the confirmation dialog.
	UpdateStateConfirm
	// UpdateStateUpdating shows the progress of the update.
	UpdateStateUpdating
	// UpdateStateComplete shows the completion message.
	UpdateStateComplete
	// UpdateStateError shows an error message.
	UpdateStateError
)

// UpdateStep represents a step in the update process.
type UpdateStep int

const (
	StepDownload UpdateStep = iota
	StepExtract
	StepBinary
	StepMarketplace
	StepPermissions
	StepHooks
	StepModelConfig
	StepMCPServers
	StepPlugins
)

// UpdateView is the Bubble Tea model for the update command.
type UpdateView struct {
	state          UpdateState
	theme          tui.Theme
	releases       []marketplace.Release
	table          components.Table
	progress       components.Progress
	confirm        components.Confirm
	spinner        spinner.Model
	currentVersion string
	selectedTag    string
	errorMsg       string
	width          int
	height         int
	tempDir        string
	extractedDir   string
	currentStep    int
	totalSteps     int
	// Fields for sudo binary update
	pendingSudoBinaryPath string
	pendingSudoTargetPath string
	// Real-time log display (like install view)
	logLines    []string
	maxLogLines int
	logChan     chan string
}

// UpdateOption configures an UpdateView.
type UpdateOption func(*UpdateView)

// WithCurrentVersion sets the current installed version.
func WithCurrentVersion(version string) UpdateOption {
	return func(v *UpdateView) {
		v.currentVersion = version
	}
}

// WithSelectedVersion pre-selects a version (skips selection UI).
func WithSelectedVersion(tag string) UpdateOption {
	return func(v *UpdateView) {
		v.selectedTag = tag
	}
}

// NewUpdateView creates a new update view.
func NewUpdateView(opts ...UpdateOption) *UpdateView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(tui.DefaultTheme.Primary)

	steps := []string{
		"Downloading release",
		"Extracting files",
		"Updating vibe binary",
		"Updating marketplace",
		"Syncing permissions",
		"Syncing hooks",
		"Configuring default model",
		"Syncing MCP servers",
		"Configuring Isaac",
		"Reinstalling plugins",
	}

	v := &UpdateView{
		state:          UpdateStateLoading,
		theme:          tui.DefaultTheme,
		spinner:        s,
		currentVersion: "dev",
		totalSteps:     len(steps),
		logLines:       make([]string, 0),
		maxLogLines:    6,
		logChan:        make(chan string, 100),
		progress: components.NewProgress(
			components.WithSteps(steps),
			components.WithProgressBar(false),
		),
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Init implements tui.View.
func (v *UpdateView) Init() tea.Cmd {
	// If a version was pre-selected, skip to confirm/update
	if v.selectedTag != "" {
		v.state = UpdateStateUpdating
		return tea.Batch(
			v.spinner.Tick,
			v.startUpdate(),
		)
	}

	return tea.Batch(
		v.spinner.Tick,
		v.fetchReleases(),
	)
}

// Update implements tui.View.
func (v *UpdateView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Clean up temp dir if it exists
			if v.tempDir != "" {
				os.RemoveAll(v.tempDir)
			}
			return v, tea.Quit

		case "enter":
			if v.state == UpdateStateSelectVersion {
				row := v.table.SelectedRow()
				if row != nil {
					v.selectedTag = row[0]
					v.state = UpdateStateConfirm
					v.confirm = components.NewConfirm(
						fmt.Sprintf("Update vibe to %s?", v.selectedTag),
						components.WithConfirmDefault(true),
					)
					return v, nil
				}
			} else if v.state == UpdateStateComplete {
				return v, tea.Quit
			} else if v.state == UpdateStateError {
				return v, tea.Quit
			}

		case "esc":
			if v.state == UpdateStateConfirm {
				v.state = UpdateStateSelectVersion
				return v, nil
			}
		}

	case updateReleasesMsg:
		v.releases = msg.releases
		v.state = UpdateStateSelectVersion
		v.buildTable()
		return v, nil

	case updateErrMsg:
		v.state = UpdateStateError
		v.errorMsg = msg.err.Error()
		return v, nil

	case updateLogMsg:
		// Add log message to buffer
		v.logLines = append(v.logLines, msg.message)
		if len(v.logLines) > v.maxLogLines {
			v.logLines = v.logLines[1:]
		}
		// Keep listening for more logs
		if v.state == UpdateStateUpdating {
			return v, v.waitForLog()
		}
		return v, nil

	case updateLogTickMsg:
		// Continue listening for logs if still updating
		if v.state == UpdateStateUpdating {
			return v, v.waitForLog()
		}
		return v, nil

	case updateStepStartMsg:
		v.progress.StartStep(int(msg.step))
		// Clear log lines for new step
		v.logLines = make([]string, 0)
		// Start the actual work for this step and log listener
		return v, tea.Batch(v.spinner.Tick, v.runStep(msg.step), v.waitForLog())

	case updateStepDoneMsg:
		v.progress.CompleteStep(int(msg.step))
		v.currentStep = int(msg.step) + 1

		// Continue to next step or complete
		if v.currentStep > 9 {
			v.state = UpdateStateComplete
			// Clean up temp dir
			if v.tempDir != "" {
				os.RemoveAll(v.tempDir)
			}
			return v, nil
		}

		// Emit start message for next step (which triggers runStep via handler)
		return v, func() tea.Msg {
			return updateStepStartMsg{step: UpdateStep(v.currentStep)}
		}

	case updateStepFailMsg:
		v.progress.FailStep(int(msg.step))
		v.state = UpdateStateError
		v.errorMsg = msg.err.Error()
		// Clean up temp dir
		if v.tempDir != "" {
			os.RemoveAll(v.tempDir)
		}
		return v, nil

	case updateSudoRequiredMsg:
		// Store the paths for use after sudo completes
		v.pendingSudoBinaryPath = msg.binaryPath
		v.pendingSudoTargetPath = msg.targetPath

		// Use tea.ExecProcess to properly suspend the TUI while sudo runs
		// This gives sudo full terminal control for password input
		backupPath := msg.targetPath + ".old"
		cmd := exec.Command("sudo", "-p", "\nPassword required to update vibe: ", "mv", msg.targetPath, backupPath)
		return v, tea.ExecProcess(cmd, func(err error) tea.Msg {
			if err != nil {
				return updateSudoCompleteMsg{err: fmt.Errorf("failed to backup current binary (sudo): %w", err)}
			}
			// First sudo succeeded, now copy the new binary
			cpCmd := exec.Command("sudo", "cp", msg.binaryPath, msg.targetPath)
			if err := cpCmd.Run(); err != nil {
				// Try to restore backup
				exec.Command("sudo", "mv", backupPath, msg.targetPath).Run()
				return updateSudoCompleteMsg{err: fmt.Errorf("failed to install new binary (sudo): %w", err)}
			}
			// Set permissions
			exec.Command("sudo", "chmod", "755", msg.targetPath).Run()
			// Remove backup
			exec.Command("sudo", "rm", "-f", backupPath).Run()
			// Install dictation helper binary (macOS only, best-effort)
			binDir := filepath.Dir(msg.targetPath)
			installDictationBinary(v.selectedTag, v.tempDir, binDir, true)
			return updateSudoCompleteMsg{err: nil}
		})

	case updateSudoCompleteMsg:
		if msg.err != nil {
			v.progress.FailStep(int(StepBinary))
			v.state = UpdateStateError
			v.errorMsg = msg.err.Error()
			if v.tempDir != "" {
				os.RemoveAll(v.tempDir)
			}
			return v, nil
		}
		// Sudo succeeded, mark step complete and continue
		v.progress.CompleteStep(int(StepBinary))
		v.currentStep = int(StepBinary) + 1
		return v, func() tea.Msg {
			return updateStepStartMsg{step: UpdateStep(v.currentStep)}
		}

	case components.ConfirmResultMsg:
		if msg.Confirmed {
			v.state = UpdateStateUpdating
			v.currentStep = 0
			return v, tea.Batch(
				v.spinner.Tick,
				v.startUpdate(),
			)
		}
		v.state = UpdateStateSelectVersion
		return v, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		cmds = append(cmds, cmd)

		// Also update progress spinner
		var progCmd tea.Cmd
		v.progress, progCmd = v.progress.Update(msg)
		cmds = append(cmds, progCmd)
	}

	// Update sub-components based on state
	switch v.state {
	case UpdateStateSelectVersion:
		newTable, cmd := v.table.Update(msg)
		v.table = newTable
		cmds = append(cmds, cmd)

	case UpdateStateConfirm:
		newConfirm, cmd := v.confirm.Update(msg)
		v.confirm = newConfirm
		cmds = append(cmds, cmd)

	case UpdateStateUpdating:
		newProgress, cmd := v.progress.Update(msg)
		v.progress = newProgress
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

// View implements tui.View.
func (v *UpdateView) View() string {
	var content string

	titleStyle := v.theme.TitleStyle().
		Bold(true).
		Padding(0, 0, 1, 0)

	switch v.state {
	case UpdateStateLoading:
		content = titleStyle.Render("Vibe Update") + "\n\n"
		content += v.spinner.View() + " Fetching available releases..."

	case UpdateStateSelectVersion:
		content = titleStyle.Render("Vibe Update") + "\n\n"
		content += v.theme.SubtitleStyle().Render(
			fmt.Sprintf("Current version: %s", v.currentVersion),
		) + "\n\n"
		content += "Select a version to install:\n\n"
		content += v.table.View() + "\n\n"
		content += v.theme.HelpStyle().Render("Press Enter to select, q to quit")

	case UpdateStateConfirm:
		content = titleStyle.Render("Vibe Update") + "\n\n"
		content += v.confirm.View()

	case UpdateStateUpdating:
		content = titleStyle.Render("Updating Vibe") + "\n\n"
		content += v.theme.SubtitleStyle().Render(
			fmt.Sprintf("%s → %s", v.currentVersion, v.selectedTag),
		) + "\n\n"
		content += v.progress.View() + "\n\n"

		// Show progress percentage
		if v.totalSteps > 0 {
			progressPercent := float64(v.currentStep) / float64(v.totalSteps) * 100
			progressBar := v.renderProgressBar(progressPercent)
			content += "  " + progressBar + fmt.Sprintf(" %3.0f%%", progressPercent) + "\n"
		}

		// Show real-time log output
		if len(v.logLines) > 0 {
			content += "\n"
			mutedStyle := v.theme.MutedStyle()
			activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

			for i, line := range v.logLines {
				// Truncate long lines
				if len(line) > 70 {
					line = line[:67] + "..."
				}

				isActive := i == len(v.logLines)-1
				if isActive {
					content += "  " + v.spinner.View() + " " + activeStyle.Render(line) + "\n"
				} else {
					content += mutedStyle.Render("  ✓ "+line) + "\n"
				}
			}
		}

	case UpdateStateComplete:
		content = titleStyle.Render("Update Complete!") + "\n\n"
		content += v.theme.SuccessStyle().Render(
			fmt.Sprintf("%s Vibe has been updated to %s", v.theme.SuccessIcon(), v.selectedTag),
		) + "\n\n"
		content += v.theme.MutedStyle().Render("You may need to restart Claude Code for changes to take effect.") + "\n\n"
		content += v.theme.HelpStyle().Render("Press Enter or q to exit")

	case UpdateStateError:
		content = titleStyle.Render("Update Failed") + "\n\n"
		content += v.theme.ErrorStyle().Render(
			fmt.Sprintf("%s Error: %s", v.theme.ErrorIcon(), v.errorMsg),
		) + "\n\n"
		content += v.theme.HelpStyle().Render("Press Enter or q to exit")
	}

	// Add border
	borderStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Margin(1, 2)

	return borderStyle.Render(content)
}

// renderProgressBar renders a simple text-based progress bar.
func (v *UpdateView) renderProgressBar(percent float64) string {
	width := 40
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	// Style the filled portion with primary color
	filledStyle := lipgloss.NewStyle().Foreground(v.theme.Primary)
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	return filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", width-filled))
}

// buildTable creates the version selection table from releases.
func (v *UpdateView) buildTable() {
	columns := []components.TableColumn{
		{Title: "Version", Width: 15},
		{Title: "Status", Width: 12},
		{Title: "Published", Width: 20},
	}

	var rows []components.TableRow
	for _, r := range v.releases {
		// Skip drafts
		if r.IsDraft {
			continue
		}

		// Skip desktop releases (they are for the CustomerIQ desktop app, not vibe CLI)
		if strings.Contains(strings.ToLower(r.Tag), "desktop") {
			continue
		}

		status := ""
		if r.IsLatest {
			status = "Latest"
		} else if r.IsPrerelease {
			status = "Pre-release"
		}

		// Check if this is newer than current
		if marketplace.CompareVersions(r.Tag, v.currentVersion) > 0 {
			if status == "" {
				status = "New"
			} else {
				status += " (New)"
			}
		} else if r.Tag == v.currentVersion || strings.TrimPrefix(r.Tag, "v") == strings.TrimPrefix(v.currentVersion, "v") {
			status = "Installed"
		}

		dateStr := r.PublishedAt.Format("Jan 2, 2006")
		rows = append(rows, components.TableRow{r.Tag, status, dateStr})
	}

	v.table = components.NewTable(
		columns,
		rows,
		components.WithTableHeight(10),
	)
}

// Message types for async operations (prefixed to avoid conflicts with other views)

type updateReleasesMsg struct {
	releases []marketplace.Release
}

type updateErrMsg struct {
	err error
}

type updateStepStartMsg struct {
	step UpdateStep
}

type updateStepDoneMsg struct {
	step UpdateStep
}

type updateStepFailMsg struct {
	step UpdateStep
	err  error
}

// updateSudoRequiredMsg signals that sudo is needed for binary update
type updateSudoRequiredMsg struct {
	binaryPath string
	targetPath string
}

// updateSudoCompleteMsg signals that sudo command completed
type updateSudoCompleteMsg struct {
	err error
}

// updateLogMsg is sent when a log message is received
type updateLogMsg struct {
	message string
}

// updateLogTickMsg is sent periodically to check for logs
type updateLogTickMsg struct{}

// Command functions

// log sends a message to the log channel for display
func (v *UpdateView) log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	select {
	case v.logChan <- msg:
	default:
		// Channel full, skip message
	}
}

// waitForLog returns a command that waits for a log message with timeout
func (v *UpdateView) waitForLog() tea.Cmd {
	return func() tea.Msg {
		if v.logChan == nil {
			return updateLogTickMsg{}
		}
		select {
		case msg, ok := <-v.logChan:
			if !ok {
				return updateLogTickMsg{}
			}
			return updateLogMsg{message: msg}
		case <-time.After(100 * time.Millisecond):
			// Timeout - return tick to keep checking
			return updateLogTickMsg{}
		}
	}
}

func (v *UpdateView) fetchReleases() tea.Cmd {
	return func() tea.Msg {
		releases, err := marketplace.ListReleases("")
		if err != nil {
			return updateErrMsg{err: err}
		}
		return updateReleasesMsg{releases: releases}
	}
}

func (v *UpdateView) startUpdate() tea.Cmd {
	return func() tea.Msg {
		// Create temp directory
		tempDir, err := os.MkdirTemp("", "vibe-update-*")
		if err != nil {
			return updateStepFailMsg{step: StepDownload, err: err}
		}
		v.tempDir = tempDir

		return updateStepStartMsg{step: StepDownload}
	}
}

func (v *UpdateView) runStep(step UpdateStep) tea.Cmd {
	return func() tea.Msg {
		// Run the actual step
		var err error
		switch step {
		case StepDownload:
			err = v.doDownload()
		case StepExtract:
			err = v.doExtract()
		case StepBinary:
			// Binary step is special - it may need to return a sudo message
			return v.doBinaryStep()
		case StepMarketplace:
			err = v.doMarketplace()
		case StepPermissions:
			err = v.doPermissions()
		case StepHooks:
			err = v.doHooks()
		case StepModelConfig:
			err = v.doModelConfig()
		case StepMCPServers:
			err = v.doMCPServers()
		case StepPlugins:
			err = v.doPlugins()
		}

		if err != nil {
			return updateStepFailMsg{step: step, err: err}
		}

		// Signal step completion
		return updateStepDoneMsg{step: step}
	}
}

func (v *UpdateView) doDownload() error {
	v.log("Downloading %s from GitHub...", v.selectedTag)
	tarballPath, err := marketplace.DownloadRelease("", v.selectedTag, v.tempDir)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	v.log("Downloaded to %s", filepath.Base(tarballPath))
	// Store tarball path for extraction step
	v.extractedDir = tarballPath
	return nil
}

func (v *UpdateView) doExtract() error {
	tarballPath := v.extractedDir

	v.log("Extracting %s...", filepath.Base(tarballPath))
	extractedDir, err := marketplace.ExtractTarball(tarballPath, v.tempDir)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	v.log("Extracted to %s", filepath.Base(extractedDir))
	v.extractedDir = extractedDir
	return nil
}

// doBinaryStep handles the binary update step, returning a tea.Msg.
// This is separate from other steps because it may need to trigger sudo
// which requires suspending the TUI via tea.ExecProcess.
func (v *UpdateView) doBinaryStep() tea.Msg {
	// Determine the asset name based on OS and architecture
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map architecture names
	archName := goarch
	if goarch == "amd64" {
		archName = "amd64"
	} else if goarch == "arm64" {
		archName = "arm64"
	} else {
		return updateStepFailMsg{step: StepBinary, err: fmt.Errorf("unsupported architecture: %s", goarch)}
	}

	// Only support darwin and linux
	if goos != "darwin" && goos != "linux" {
		return updateStepFailMsg{step: StepBinary, err: fmt.Errorf("unsupported operating system: %s", goos)}
	}

	assetName := fmt.Sprintf("vibe-%s-%s", goos, archName)
	v.log("Downloading binary: %s", assetName)

	// Download the binary asset.
	// Always specify --hostname github.com to avoid Databricks GHE interception.
	binaryPath := filepath.Join(v.tempDir, assetName)
	_, stderr, err := util.RunCommand("gh", "release", "download", v.selectedTag,
		"--repo", marketplace.DefaultRepo,
		"--pattern", assetName,
		"--output", binaryPath,
		"--hostname", "github.com",
	)
	if err != nil {
		return updateStepFailMsg{step: StepBinary, err: fmt.Errorf("failed to download binary: %w\n%s", err, stderr)}
	}

	// Make the downloaded binary executable
	v.log("Setting executable permissions")
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return updateStepFailMsg{step: StepBinary, err: fmt.Errorf("failed to make binary executable: %w", err)}
	}

	// Get the path of the currently running binary
	currentBinary, err := os.Executable()
	if err != nil {
		return updateStepFailMsg{step: StepBinary, err: fmt.Errorf("failed to get current executable path: %w", err)}
	}

	// Resolve any symlinks to get the actual binary location
	currentBinary, err = filepath.EvalSymlinks(currentBinary)
	if err != nil {
		return updateStepFailMsg{step: StepBinary, err: fmt.Errorf("failed to resolve symlinks: %w", err)}
	}

	// Check if we need sudo by testing write permission
	needsSudo := !canWriteToDir(filepath.Dir(currentBinary))

	if needsSudo {
		v.log("Sudo required for %s", currentBinary)
		// Signal that sudo is required - the Update handler will use tea.ExecProcess
		// to properly suspend the TUI while sudo runs
		return updateSudoRequiredMsg{binaryPath: binaryPath, targetPath: currentBinary}
	}

	v.log("Installing to %s", currentBinary)
	// No sudo needed - proceed with normal update
	// Create backup of current binary
	backupPath := currentBinary + ".old"
	if err := os.Rename(currentBinary, backupPath); err != nil {
		return updateStepFailMsg{step: StepBinary, err: fmt.Errorf("failed to backup current binary: %w", err)}
	}

	// Move new binary to current location
	if err := os.Rename(binaryPath, currentBinary); err != nil {
		// Try to restore backup on failure
		os.Rename(backupPath, currentBinary)
		return updateStepFailMsg{step: StepBinary, err: fmt.Errorf("failed to install new binary: %w", err)}
	}

	// Remove backup
	os.Remove(backupPath)

	// Install dictation helper binary (macOS only, best-effort)
	binDir := filepath.Dir(currentBinary)
	if err := installDictationBinary(v.selectedTag, v.tempDir, binDir, false); err != nil {
		v.log("Warning: failed to install dictation binary: %v", err)
	} else if runtime.GOOS == "darwin" {
		v.log("Installed vibe-dictation to %s", binDir)
	}

	return updateStepDoneMsg{step: StepBinary}
}

// installDictationBinary downloads and installs the vibe-dictation helper binary
// next to the main vibe binary. This is macOS-only and best-effort — if it fails,
// the error is returned but callers may choose to treat it as non-fatal.
func installDictationBinary(version, tempDir, targetDir string, needsSudo bool) error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	archName := runtime.GOARCH
	assetName := fmt.Sprintf("vibe-dictation-darwin-%s", archName)
	dictationPath := filepath.Join(tempDir, assetName)

	_, stderr, err := util.RunCommand("gh", "release", "download", version,
		"--repo", marketplace.DefaultRepo,
		"--pattern", assetName,
		"--output", dictationPath,
		"--hostname", "github.com",
	)
	if err != nil {
		return fmt.Errorf("failed to download dictation binary: %w\n%s", err, stderr)
	}

	if err := os.Chmod(dictationPath, 0755); err != nil {
		return fmt.Errorf("failed to make dictation binary executable: %w", err)
	}

	targetPath := filepath.Join(targetDir, "vibe-dictation")

	if needsSudo {
		// Remove any existing dictation binary first
		exec.Command("sudo", "rm", "-f", targetPath).Run()
		cmd := exec.Command("sudo", "cp", dictationPath, targetPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dictation binary (sudo): %w", err)
		}
		exec.Command("sudo", "chmod", "755", targetPath).Run()
	} else {
		// Remove any existing dictation binary first
		os.Remove(targetPath)
		if err := os.Rename(dictationPath, targetPath); err != nil {
			// Fall back to copy if rename fails (cross-device)
			data, readErr := os.ReadFile(dictationPath)
			if readErr != nil {
				return fmt.Errorf("failed to install dictation binary: %w", err)
			}
			if writeErr := os.WriteFile(targetPath, data, 0755); writeErr != nil {
				return fmt.Errorf("failed to install dictation binary: %w", writeErr)
			}
		}
	}

	return nil
}

// canWriteToDir checks if we have write permission to a directory
func canWriteToDir(dir string) bool {
	// Try to create a temp file in the directory
	testFile := filepath.Join(dir, ".vibe-write-test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)
	return true
}

// replaceBinaryWithSudo uses sudo to replace the binary when we don't have write permissions.
// This is used by the non-interactive update path. For TUI updates, we use tea.ExecProcess
// in the Update handler to properly suspend the TUI while sudo runs.
func replaceBinaryWithSudo(newBinary, targetPath string) error {
	backupPath := targetPath + ".old"

	// Create backup using sudo mv
	cmd := exec.Command("sudo", "-p", "\nPassword required to update vibe: ", "mv", targetPath, backupPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to backup current binary (sudo): %w", err)
	}

	// Copy new binary to target using sudo cp (can't mv across filesystems)
	cmd = exec.Command("sudo", "cp", newBinary, targetPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Try to restore backup
		exec.Command("sudo", "mv", backupPath, targetPath).Run()
		return fmt.Errorf("failed to install new binary (sudo): %w", err)
	}

	// Make sure the new binary has correct permissions
	exec.Command("sudo", "chmod", "755", targetPath).Run()

	// Remove backup using sudo
	exec.Command("sudo", "rm", "-f", backupPath).Run()

	return nil
}

func (v *UpdateView) doMarketplace() error {
	// Copy to permanent marketplace location
	marketplacePath := marketplace.MarketplacePath()

	v.log("Removing old marketplace...")
	// Remove old marketplace
	if err := os.RemoveAll(marketplacePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old marketplace: %w", err)
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(marketplacePath), 0755); err != nil {
		return fmt.Errorf("failed to create marketplace directory: %w", err)
	}

	v.log("Copying new marketplace to %s", marketplacePath)
	// Copy new marketplace
	if err := marketplace.CopyDir(v.extractedDir, marketplacePath); err != nil {
		return fmt.Errorf("failed to copy marketplace: %w", err)
	}

	// Update Claude Code marketplace
	// First remove old marketplace if it exists
	v.log("Running: claude plugin marketplace remove claude-vibe")
	exec.Command("claude", "plugin", "marketplace", "remove", "claude-vibe").Run()

	// Add new marketplace
	v.log("Running: claude plugin marketplace add")
	cmd := exec.Command("claude", "plugin", "marketplace", "add", marketplacePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add marketplace: %w\n%s", err, string(output))
	}

	return nil
}

func (v *UpdateView) doPermissions() error {
	home, _ := os.UserHomeDir()
	permissionsFile := filepath.Join(v.extractedDir, "permissions.yaml")
	settingsFile := filepath.Join(home, ".claude", "settings.json")

	// Check if permissions file exists
	if _, err := os.Stat(permissionsFile); os.IsNotExist(err) {
		v.log("No permissions.yaml found, skipping")
		return nil // No permissions to sync
	}

	v.log("Reading permissions from permissions.yaml")
	// Ensure settings directory exists
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Ensure settings.json exists
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		if err := os.WriteFile(settingsFile, []byte(`{"allow":[],"deny":[]}`), 0644); err != nil {
			return fmt.Errorf("failed to create settings.json: %w", err)
		}
	}

	// Use yq and jq to merge permissions
	// Extract allow list from YAML
	v.log("Extracting permissions with yq")
	yqCmd := exec.Command("yq", "-r", ".allow | @json", permissionsFile)
	allowPerms, err := yqCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to extract permissions: %w", err)
	}

	// Create temp file for jq output
	tmpFile, err := os.CreateTemp("", "settings-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Merge using jq
	v.log("Merging into ~/.claude/settings.json")
	jqCmd := exec.Command("jq",
		"--argjson", "new_perms", strings.TrimSpace(string(allowPerms)),
		`.allow = (.allow // [] | . + $new_perms | unique)`,
		settingsFile,
	)
	output, err := jqCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to merge permissions: %w", err)
	}

	// Write merged settings
	if err := os.WriteFile(settingsFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

func (v *UpdateView) doHooks() error {
	home, _ := os.UserHomeDir()
	settingsFile := filepath.Join(home, ".claude", "settings.json")

	v.log("Reading ~/.claude/settings.json")
	// Ensure settings directory exists
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Read or create settings.json
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsFile); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse settings.json: %w", err)
		}
	} else if os.IsNotExist(err) {
		settings = map[string]interface{}{
			"allow": []interface{}{},
			"deny":  []interface{}{},
		}
	} else {
		return fmt.Errorf("failed to read settings.json: %w", err)
	}

	// Define the telemetry stop hook using the new matcher format
	telemetryCommand := "vibe telemetry publish --event-type=claude.session.stop --source=claude-code-stop-hook --from-hook --quiet 2>/dev/null || true"
	telemetryHookEntry := map[string]interface{}{
		"matcher": "", // Empty matcher matches all Stop events
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": telemetryCommand,
				"timeout": 30,
			},
		},
	}

	// Get or create hooks section
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
	}

	// Get or create Stop hooks array
	stopHooks, ok := hooks["Stop"].([]interface{})
	if !ok {
		stopHooks = []interface{}{}
	}

	// Check if our hook already exists (check nested hooks array)
	hookExists := false
	for _, h := range stopHooks {
		if hookMap, ok := h.(map[string]interface{}); ok {
			if nestedHooks, ok := hookMap["hooks"].([]interface{}); ok {
				for _, nh := range nestedHooks {
					if nhMap, ok := nh.(map[string]interface{}); ok {
						if cmd, ok := nhMap["command"].(string); ok {
							if cmd == telemetryCommand {
								hookExists = true
								break
							}
						}
					}
				}
			}
		}
		if hookExists {
			break
		}
	}

	// Add hook if it doesn't exist
	if !hookExists {
		v.log("Adding telemetry Stop hook")
		stopHooks = append(stopHooks, telemetryHookEntry)
		hooks["Stop"] = stopHooks
		settings["hooks"] = hooks

		// Write back settings
		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal settings: %w", err)
		}

		if err := os.WriteFile(settingsFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write settings.json: %w", err)
		}
	} else {
		v.log("Telemetry hook already exists")
	}

	return nil
}

func (v *UpdateView) doModelConfig() error {
	home, _ := os.UserHomeDir()
	settingsFile := filepath.Join(home, ".claude", "settings.json")

	v.log("Reading ~/.claude/settings.json")
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return fmt.Errorf("failed to read settings.json: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings.json: %w", err)
	}

	currentModel, _ := settings["model"].(string)
	if currentModel == "opus" {
		v.log("Default model already set to opus")
		return nil
	}

	v.log("Setting default model to opus")
	settings["model"] = "opus"

	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}

	return nil
}

func (v *UpdateView) doMCPServers() error {
	home, _ := os.UserHomeDir()
	mcpServersFile := filepath.Join(v.extractedDir, "mcp-servers.yaml")
	mcpConfig := filepath.Join(home, ".config", "mcp", "config.json")

	// Check if MCP servers file exists
	if _, err := os.Stat(mcpServersFile); os.IsNotExist(err) {
		v.log("No mcp-servers.yaml found, skipping")
		return nil // No MCP servers to sync
	}

	v.log("Reading MCP servers from mcp-servers.yaml")
	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Join(home, ".config", "mcp"), 0755); err != nil {
		return fmt.Errorf("failed to create MCP config directory: %w", err)
	}

	// Ensure config.json exists
	if _, err := os.Stat(mcpConfig); os.IsNotExist(err) {
		if err := os.WriteFile(mcpConfig, []byte(`{"claude-code":{}}`), 0644); err != nil {
			return fmt.Errorf("failed to create MCP config: %w", err)
		}
	}

	// Extract servers from YAML and expand ~ to $HOME
	v.log("Extracting server configs with yq")
	yqCmd := exec.Command("yq", "-r", ".servers | @json", mcpServersFile)
	yqCmd.Env = append(os.Environ(), fmt.Sprintf("HOME=%s", home))
	serversJSON, err := yqCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to extract MCP servers: %w", err)
	}

	// Expand ~ in paths using jq
	expandCmd := exec.Command("jq", `walk(
		if type == "string" then
			if startswith("~/") then
				sub("^~/"; env.HOME + "/")
			elif contains("=~/") then
				gsub("=~/"; "=" + env.HOME + "/")
			else
				.
			end
		else
			.
		end
	)`)
	expandCmd.Env = append(os.Environ(), fmt.Sprintf("HOME=%s", home))
	expandCmd.Stdin = strings.NewReader(string(serversJSON))
	expandedJSON, err := expandCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to expand paths: %w", err)
	}

	// Transform servers to add enabled and name fields
	transformCmd := exec.Command("jq", `
		to_entries | map({
			key: .key,
			value: ((.value + {enabled: true, name: .key}) | del(.type))
		}) | from_entries
	`)
	transformCmd.Stdin = strings.NewReader(string(expandedJSON))
	transformedJSON, err := transformCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to transform servers: %w", err)
	}

	// Merge into MCP config
	v.log("Merging into ~/.config/mcp/config.json")
	mergeCmd := exec.Command("jq",
		"--argjson", "new_servers", strings.TrimSpace(string(transformedJSON)),
		`.["claude-code"] = ((.["claude-code"] // {}) * $new_servers)`,
		mcpConfig,
	)
	output, err := mergeCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to merge MCP config: %w", err)
	}

	// Write merged config
	if err := os.WriteFile(mcpConfig, output, 0644); err != nil {
		return fmt.Errorf("failed to write MCP config: %w", err)
	}

	return nil
}

func (v *UpdateView) doPlugins() error {
	plugins := install.AllPluginsToInstall()

	for _, plugin := range plugins {
		v.log("Installing %s@claude-vibe", plugin)
		// Small delay to allow UI to render the log message
		time.Sleep(50 * time.Millisecond)
		cmd := exec.Command("claude", "plugin", "install", plugin+"@claude-vibe")
		// Ignore errors for individual plugins - some may not be available
		cmd.Run()
	}

	return nil
}

// RunUpdateTUI runs the update TUI with version selection.
func RunUpdateTUI(currentVersion string, targetVersion string) error {
	opts := []UpdateOption{
		WithCurrentVersion(currentVersion),
	}

	if targetVersion != "" {
		opts = append(opts, WithSelectedVersion(targetVersion))
	}

	view := NewUpdateView(opts...)
	app := tui.NewApp(view)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunUpdate runs the update to latest stable version with progress UI.
// This is the default behavior for `vibe update` without flags.
func RunUpdate(currentVersion string, targetVersion string) error {
	// If no target version specified, fetch the latest stable release
	if targetVersion == "" {
		fmt.Print("Checking for updates... ")
		release, err := marketplace.GetLatestRelease("")
		if err != nil {
			fmt.Println("failed")
			return fmt.Errorf("failed to get latest release: %w", err)
		}
		targetVersion = release.Tag
		fmt.Printf("found %s\n", targetVersion)

		// Check if already up to date
		if marketplace.CompareVersions(targetVersion, currentVersion) <= 0 {
			fmt.Printf("\nAlready up to date (current: %s)\n", currentVersion)
			return nil
		}
		fmt.Printf("Updating %s → %s\n\n", currentVersion, targetVersion)
	}

	// Run the TUI with the target version pre-selected (skips version selection)
	opts := []UpdateOption{
		WithCurrentVersion(currentVersion),
		WithSelectedVersion(targetVersion),
	}

	view := NewUpdateView(opts...)
	app := tui.NewApp(view)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunUpdateNonInteractive runs the update process without TUI.
func RunUpdateNonInteractive(currentVersion, targetVersion string) error {
	fmt.Println("Vibe Update")
	fmt.Println()

	// Determine target version
	if targetVersion == "" || targetVersion == "latest" {
		fmt.Print("Fetching latest release... ")
		release, err := marketplace.GetLatestRelease("")
		if err != nil {
			fmt.Println("FAILED")
			return fmt.Errorf("failed to get latest release: %w", err)
		}
		targetVersion = release.Tag
		fmt.Printf("found %s\n", targetVersion)
	}

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Target version: %s\n", targetVersion)
	fmt.Println()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "vibe-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download
	fmt.Print("Downloading release... ")
	tarballPath, err := marketplace.DownloadRelease("", targetVersion, tempDir)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("done")

	// Extract
	fmt.Print("Extracting files... ")
	extractedDir, err := marketplace.ExtractTarball(tarballPath, tempDir)
	if err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("done")

	// Update binary
	fmt.Print("Updating vibe binary... ")
	if err := updateBinaryNonInteractive(targetVersion, tempDir); err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("done")

	// Update marketplace
	fmt.Print("Updating marketplace... ")
	marketplacePath := marketplace.MarketplacePath()
	os.RemoveAll(marketplacePath)
	os.MkdirAll(filepath.Dir(marketplacePath), 0755)
	if err := marketplace.CopyDir(extractedDir, marketplacePath); err != nil {
		fmt.Println("FAILED")
		return err
	}
	exec.Command("claude", "plugin", "marketplace", "remove", "claude-vibe").Run()
	if err := exec.Command("claude", "plugin", "marketplace", "add", marketplacePath).Run(); err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("done")

	// Sync permissions
	fmt.Print("Syncing permissions... ")
	// Call the sync script if available
	syncScript := filepath.Join(extractedDir, "plugins", "vibe-setup", "skills", "configure-vibe", "resources", "vibe_sync.sh")
	if _, err := os.Stat(syncScript); err == nil {
		exec.Command("bash", syncScript).Run()
	}
	fmt.Println("done")

	// Configure default model
	fmt.Print("Configuring default model... ")
	if err := setDefaultModelNonInteractive(); err != nil {
		fmt.Println("FAILED")
		fmt.Printf("Warning: %v\n", err)
	} else {
		fmt.Println("done")
	}

	// Reinstall plugins
	fmt.Print("Reinstalling plugins... ")
	for _, plugin := range install.AllPluginsToInstall() {
		exec.Command("claude", "plugin", "install", plugin+"@claude-vibe").Run()
	}
	fmt.Println("done")

	fmt.Println()
	fmt.Printf("Vibe has been updated to %s\n", targetVersion)
	fmt.Println("You may need to restart Claude Code for changes to take effect.")

	return nil
}

// updateBinaryNonInteractive downloads and installs the new vibe binary.
func updateBinaryNonInteractive(targetVersion, tempDir string) error {
	// Determine the asset name based on OS and architecture
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map architecture names
	archName := goarch
	if goarch == "amd64" {
		archName = "amd64"
	} else if goarch == "arm64" {
		archName = "arm64"
	} else {
		return fmt.Errorf("unsupported architecture: %s", goarch)
	}

	// Only support darwin and linux
	if goos != "darwin" && goos != "linux" {
		return fmt.Errorf("unsupported operating system: %s", goos)
	}

	assetName := fmt.Sprintf("vibe-%s-%s", goos, archName)

	// Download the binary asset.
	// Always specify --hostname github.com to avoid Databricks GHE interception.
	binaryPath := filepath.Join(tempDir, assetName)
	_, stderr, err := util.RunCommand("gh", "release", "download", targetVersion,
		"--repo", marketplace.DefaultRepo,
		"--pattern", assetName,
		"--output", binaryPath,
		"--hostname", "github.com",
	)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w\n%s", err, stderr)
	}

	// Make the downloaded binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Get the path of the currently running binary
	currentBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Resolve any symlinks to get the actual binary location
	currentBinary, err = filepath.EvalSymlinks(currentBinary)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Check if we need sudo by testing write permission
	needsSudo := !canWriteToDir(filepath.Dir(currentBinary))

	binDir := filepath.Dir(currentBinary)

	if needsSudo {
		// Use sudo to replace the binary
		if err := replaceBinaryWithSudo(binaryPath, currentBinary); err != nil {
			return err
		}
		// Install dictation helper binary (macOS only, best-effort)
		installDictationBinary(targetVersion, tempDir, binDir, true)
		return nil
	}

	// Create backup of current binary
	backupPath := currentBinary + ".old"
	if err := os.Rename(currentBinary, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary to current location
	if err := os.Rename(binaryPath, currentBinary); err != nil {
		// Try to restore backup on failure
		os.Rename(backupPath, currentBinary)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Remove backup
	os.Remove(backupPath)

	// Install dictation helper binary (macOS only, best-effort)
	installDictationBinary(targetVersion, tempDir, binDir, false)

	return nil
}

// setDefaultModelNonInteractive sets the default model to opus in settings.json.
func setDefaultModelNonInteractive() error {
	home, _ := os.UserHomeDir()
	settingsFile := filepath.Join(home, ".claude", "settings.json")

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return fmt.Errorf("failed to read settings.json: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings.json: %w", err)
	}

	settings["model"] = "opus"

	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	return os.WriteFile(settingsFile, output, 0644)
}

