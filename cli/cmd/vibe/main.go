// Package main provides the entry point for the vibe CLI.
// Vibe is the Field Engineering Claude Code Plugin Manager.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	// earlyinit MUST be imported first to suppress spurious SDK init warnings
	"github.com/wmsimpson/claude-vibe/cli/internal/earlyinit"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wmsimpson/claude-vibe/cli/internal/config"
	"github.com/wmsimpson/claude-vibe/cli/internal/doctor"
	"github.com/wmsimpson/claude-vibe/cli/internal/install"
	"github.com/wmsimpson/claude-vibe/cli/internal/marketplace"
	vibesync "github.com/wmsimpson/claude-vibe/cli/internal/sync"
	"github.com/wmsimpson/claude-vibe/cli/internal/telemetry"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui"
	"github.com/wmsimpson/claude-vibe/cli/internal/tui/views"
	"github.com/wmsimpson/claude-vibe/cli/internal/util"
	"github.com/wmsimpson/claude-vibe/cli/internal/voice"
	"github.com/spf13/cobra"
)

// Build-time variables set via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Global flags
var (
	profile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vibe",
	Short: "Field Engineering Claude Code Plugin Manager",
	Long: `Vibe CLI manages plugins, MCP servers, and configurations for Claude Code
and other AI coding agents.

Use vibe to install, configure, and manage plugins from the Field Engineering
marketplace, run health checks on your environment, and launch agent sessions.`,
	// Run without subcommand shows help
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// versionCmd prints the version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the vibe version",
	Long:  `Print the version, commit hash, and build date of vibe.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("vibe %s\n", version)
		if verbose {
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		}
	},
}

// updateCmd handles self-update functionality
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update vibe from the local repository",
	Long: `Update vibe from a local git repository.

Vibe is distributed as a git bundle via email, not via GitHub releases.
To update, apply the new bundle and re-run install:

  cd $VIBE_REPO_PATH && git pull ~/Downloads/claude-vibe-latest.bundle main
  ./install_cli.sh
  vibe install`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := os.Getenv("VIBE_REPO_PATH")
		if repoPath == "" {
			home, _ := os.UserHomeDir()
			repoPath = filepath.Join(home, "claude-vibe")
		}

		fmt.Println("")
		fmt.Println("Vibe is distributed as a git bundle — there are no GitHub releases to download.")
		fmt.Println("")
		fmt.Println("To update:")
		fmt.Println("")
		fmt.Printf("  1. Save the new bundle to ~/Downloads/claude-vibe-latest.bundle\n")
		fmt.Printf("  2. cd %s\n", repoPath)
		fmt.Printf("     git pull ~/Downloads/claude-vibe-latest.bundle main\n")
		fmt.Printf("  3. ./install_cli.sh\n")
		fmt.Printf("  4. vibe install\n")
		fmt.Println("")
	},
}

// doctorCmd handles health checks
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run health checks on your environment",
	Long: `Run health checks to verify your vibe installation and environment.

By default, runs all checks and prompts to repair any issues found.
Use --repair to automatically fix issues without prompting.
Use --collect to generate a diagnostic tar.gz for sharing.`,
	Run: func(cmd *cobra.Command, args []string) {
		repair, _ := cmd.Flags().GetBool("repair")
		collect, _ := cmd.Flags().GetBool("collect")

		if collect {
			runDoctorCollect()
			return
		}

		runDoctorChecks(repair)
	},
}

// runDoctorChecks runs all health checks
func runDoctorChecks(autoRepair bool) {
	if autoRepair {
		// Non-interactive mode: run checks and auto-repair
		runDoctorNonInteractive()
		return
	}

	// Interactive TUI mode
	doctorView := views.NewDoctorView(
		views.WithDoctorTheme(tui.DefaultTheme),
	)

	app := tui.NewApp(doctorView)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running doctor: %v\n", err)
		os.Exit(1)
	}
}

// runDoctorNonInteractive runs doctor checks and auto-repairs without TUI
func runDoctorNonInteractive() {
	fmt.Println("Running vibe doctor...")
	fmt.Println()

	// Get all checks from the doctor package
	checks := doctor.AllChecks()
	results := make([]doctor.CheckResult, 0, len(checks))

	// Run all checks
	for _, check := range checks {
		result := check.Run()
		results = append(results, result)

		var icon string
		switch result.Status {
		case doctor.StatusPass:
			icon = "✓"
		case doctor.StatusFail:
			icon = "✗"
		case doctor.StatusWarning:
			icon = "⚠"
		}

		fmt.Printf("  %s %s: %s\n", icon, check.Description(), result.Message)
	}

	fmt.Println()

	// Check if there are any issues
	if !doctor.HasIssues(results) {
		fmt.Println("All checks passed! Vibe is healthy.")
		return
	}

	// Auto-repair mode
	fmt.Println("Repairing issues...")
	fmt.Println()

	repairResults := doctor.RepairAll(checks, results)

	for _, result := range repairResults {
		var icon string
		if result.Repaired {
			icon = "✓"
		} else if result.Skipped {
			icon = "○"
		} else {
			icon = "✗"
		}

		fmt.Printf("  %s %s: %s\n", icon, result.CheckName, result.Message)
	}

	fmt.Println()

	// Count repair outcomes
	repaired := 0
	repairFailed := 0
	for _, r := range repairResults {
		if r.Repaired {
			repaired++
		} else if !r.Skipped {
			repairFailed++
		}
	}

	if repairFailed > 0 {
		fmt.Printf("Repairs completed: %d succeeded, %d failed\n", repaired, repairFailed)
		fmt.Println("Some issues could not be auto-repaired. See hints above for manual fixes.")
	} else if repaired > 0 {
		fmt.Printf("All %d repairs completed successfully!\n", repaired)
	} else {
		fmt.Println("No auto-repairs were available for the issues found.")
		fmt.Println("See hints above for manual fixes.")
	}
}

// runDoctorCollect generates diagnostic tarball
func runDoctorCollect() {
	fmt.Println("Collecting diagnostic information...")

	// Use the collector from internal/doctor package
	info, err := doctor.Collect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting diagnostics: %v\n", err)
		os.Exit(1)
	}

	// Save to a tar.gz file
	path, err := info.SaveToDefaultLocation()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving diagnostics: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Diagnostics saved to: %s\n", path)
	fmt.Println()
	fmt.Println("This file can be shared with support for troubleshooting.")
	fmt.Println("Sensitive data (passwords, tokens, etc.) has been redacted.")
}

// configureCmd handles configuration UI
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure vibe settings",
	Long: `Open the configuration interface to manage vibe settings,
MCP servers, plugins, and profiles.

Use --tab to jump directly to a specific tab:
  - settings: General vibe settings
  - mcp: MCP server configuration
  - plugins: Plugin management
  - profiles: Profile management`,
	Run: func(cmd *cobra.Command, args []string) {
		tabName, _ := cmd.Flags().GetString("tab")

		// Create configure view with options
		opts := []views.ConfigureOption{
			views.WithConfigureTheme(tui.DefaultTheme),
		}
		if tabName != "" {
			opts = append(opts, views.WithInitialTab(tabName))
		}

		configureView := views.NewConfigureView(opts...)
		app := tui.NewApp(configureView)
		p := tea.NewProgram(app, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running configure: %v\n", err)
			os.Exit(1)
		}
	},
}

// pluginsCmd handles plugin management
var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Manage Claude Code plugins",
	Long: `List, install, update, and remove Claude Code plugins.

Without flags, lists all marketplace plugins with their status.
Press Enter on any plugin to open the configuration menu.
Use --installed to show only installed plugins.
Use --global or --user to open the interactive TUI for user-scope plugins.
Use --profile <name> to manage plugins for a specific profile.
Use --publish [path] to publish a plugin to the vibe marketplace.
Use --update to refresh the marketplace from the latest release and reinstall all plugins.
Use --update --latest to pull the marketplace from the main branch (includes unreleased plugins).`,
	Run: func(cmd *cobra.Command, args []string) {
		installedOnly, _ := cmd.Flags().GetBool("installed")
		globalScope, _ := cmd.Flags().GetBool("global")
		userScope, _ := cmd.Flags().GetBool("user")
		publish, _ := cmd.Flags().GetString("publish")
		updatePlugins, _ := cmd.Flags().GetBool("update")
		useLatest, _ := cmd.Flags().GetBool("latest")

		// --update flag: refresh marketplace and reinstall all plugins
		if updatePlugins {
			runPluginsUpdate(useLatest)
			return
		}

		// --publish flag: run the vibe-publish-plugin skill
		if cmd.Flags().Changed("publish") {
			runPublishPlugin(publish)
			return
		}

		// --profile flag (global): open TUI for profile's plugins
		if profile != "" {
			runProfilePluginsTUI(profile)
			return
		}

		// --global or --user flag: open TUI for user-scope plugins
		if globalScope || userScope {
			runUserPluginsTUI()
			return
		}

		// --installed flag: print only installed plugins
		if installedOnly {
			printInstalledPlugins()
			return
		}

		// Default: print all marketplace plugins with status
		printAllPlugins()
	},
}

// pluginTableModel is a simple Bubble Tea model for displaying plugins in a table
type pluginTableModel struct {
	table         table.Model
	title         string
	quitting      bool
	openConfigure bool
	isHomeDir     bool
}

func (m pluginTableModel) Init() tea.Cmd { return nil }

func (m pluginTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.openConfigure = true
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m pluginTableModel) View() string {
	if m.quitting || m.openConfigure {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1)

	tableStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	return titleStyle.Render(m.title) + "\n" +
		tableStyle.Render(m.table.View()) + "\n" +
		helpStyle.Render("↑/↓: navigate | enter: configure | q/esc: quit")
}

// printInstalledPlugins displays installed plugins in a TUI table
func printInstalledPlugins() {
	// Get plugin configuration
	pluginConfig := config.NewPluginConfig()
	installed := pluginConfig.ListInstalled()

	if len(installed) == 0 {
		fmt.Println("No plugins installed.")
		return
	}

	// Check if we're in the home directory
	cwd, _ := os.Getwd()
	isHome := util.IsHomeDirectory(cwd)

	// Define columns
	columns := []table.Column{
		{Title: "Status", Width: 8},
		{Title: "Plugin", Width: 28},
		{Title: "Version", Width: 10},
		{Title: "Scope", Width: 10},
	}

	// Build rows
	rows := make([]table.Row, 0, len(installed))
	for _, p := range installed {
		status := "✓"
		if !p.Enabled {
			status = "○"
		}
		rows = append(rows, table.Row{status, p.Name, p.Version, p.Scope})
	}

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows)+1, 15)),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("#7C3AED")).
		Bold(false)
	t.SetStyles(s)

	m := pluginTableModel{table: t, title: "Installed Plugins", isHomeDir: isHome}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check if user pressed Enter to open configure
	if fm, ok := finalModel.(pluginTableModel); ok && fm.openConfigure {
		if fm.isHomeDir {
			runUserPluginsTUI()
		} else {
			runProjectPluginsTUI()
		}
	}
}

// printAllPlugins displays all marketplace plugins in a TUI table
func printAllPlugins() {
	// Get plugin configuration
	pluginConfig := config.NewPluginConfig()
	installed := pluginConfig.ListInstalled()
	available := pluginConfig.ListAvailable()

	if len(available) == 0 {
		fmt.Println("No plugins found in marketplace.")
		fmt.Println("Make sure vibe is installed correctly.")
		return
	}

	// Check if we're in the home directory
	cwd, _ := os.Getwd()
	isHome := util.IsHomeDirectory(cwd)

	// Build map of installed plugins for quick lookup
	installedMap := make(map[string]config.Plugin)
	for _, p := range installed {
		installedMap[p.Name] = p
	}

	// Define columns
	columns := []table.Column{
		{Title: "Status", Width: 8},
		{Title: "Plugin", Width: 28},
		{Title: "Version", Width: 10},
		{Title: "Scope", Width: 10},
		{Title: "Description", Width: 40},
	}

	// Build rows
	rows := make([]table.Row, 0, len(available))
	for _, avail := range available {
		var status, version, scope, desc string

		if inst, ok := installedMap[avail.Name]; ok {
			// Installed plugin
			if inst.Enabled {
				status = "✓"
			} else {
				status = "○"
			}
			version = inst.Version
			scope = inst.Scope
			desc = ""
		} else {
			// Not installed
			status = "—"
			version = ""
			scope = ""
			desc = avail.Description
			if len(desc) > 38 {
				desc = desc[:35] + "..."
			}
		}

		rows = append(rows, table.Row{status, avail.Name, version, scope, desc})
	}

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows)+1, 20)),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("#7C3AED")).
		Bold(false)
	t.SetStyles(s)

	m := pluginTableModel{table: t, title: "All Marketplace Plugins (✓ enabled, ○ disabled, — not installed)", isHomeDir: isHome}
	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check if user pressed Enter to open configure
	if fm, ok := finalModel.(pluginTableModel); ok && fm.openConfigure {
		if fm.isHomeDir {
			runUserPluginsTUI()
		} else {
			runProjectPluginsTUI()
		}
	}
}

// runPublishPlugin launches the vibe-publish-plugin skill via claude
func runPublishPlugin(path string) {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// If a path was provided, resolve it relative to CWD
	if path != "" && path != "." {
		if !filepath.IsAbs(path) {
			path = filepath.Join(workDir, path)
		}
		workDir = path
	}

	// Verify the directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory does not exist: %s\n", workDir)
		os.Exit(1)
	}

	fmt.Printf("Publishing plugin from %s...\n", workDir)

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: 'claude' command not found. Please install Claude Code CLI.\n")
		os.Exit(1)
	}

	cmd := exec.Command(claudePath, "/vibe-publish-plugin")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running publish: %v\n", err)
		os.Exit(1)
	}
}

// runPluginsUpdate refreshes the marketplace from the local repo and reinstalls all plugins.
// fromMain is ignored — there are no GitHub releases; the local repo is always used.
func runPluginsUpdate(_ bool) {
	fmt.Println("Updating plugins from local repository...")
	fmt.Println()

	// Find the local repo
	repoPath := os.Getenv("VIBE_REPO_PATH")
	if repoPath == "" {
		home, _ := os.UserHomeDir()
		repoPath = filepath.Join(home, "claude-vibe")
	}
	if _, err := os.Stat(filepath.Join(repoPath, ".claude-plugin")); err != nil {
		fmt.Fprintf(os.Stderr, "Vibe repo not found at %s\n", repoPath)
		fmt.Fprintf(os.Stderr, "Set VIBE_REPO_PATH to the repo location and retry.\n")
		os.Exit(1)
	}
	fmt.Printf("Using repo: %s\n\n", repoPath)

	// Copy local repo to marketplace directory
	fmt.Print("Updating marketplace... ")
	marketplacePath := marketplace.MarketplacePath()
	os.RemoveAll(marketplacePath)
	os.MkdirAll(filepath.Dir(marketplacePath), 0755)
	if err := marketplace.CopyDir(repoPath, marketplacePath); err != nil {
		fmt.Println("FAILED")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	exec.Command("claude", "plugin", "marketplace", "remove", "claude-vibe").Run()
	if err := exec.Command("claude", "plugin", "marketplace", "add", marketplacePath).Run(); err != nil {
		fmt.Println("FAILED")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("done")

	// Reinstall all plugins (defaults + extras)
	plugins := install.AllPluginsToInstall()
	fmt.Printf("Reinstalling %d plugins...\n", len(plugins))

	var failed []string
	for _, plugin := range plugins {
		fmt.Printf("  Installing %s... ", plugin)
		cmd := exec.Command("claude", "plugin", "install", plugin+"@claude-vibe")
		if err := cmd.Run(); err != nil {
			fmt.Println("FAILED")
			failed = append(failed, plugin)
		} else {
			fmt.Println("ok")
		}
	}

	fmt.Println()
	if len(failed) > 0 {
		fmt.Printf("Warning: %d plugin(s) failed to install: %s\n", len(failed), strings.Join(failed, ", "))
		fmt.Println("Try manually: claude plugin install <plugin-name>@claude-vibe")
	} else {
		fmt.Printf("All %d plugins installed successfully.\n", len(plugins))
	}

	// Auto-sync to other agents if configured
	runAutoSync()
}

// runUserPluginsTUI launches the configure TUI at the Plugins tab for user scope
func runUserPluginsTUI() {
	opts := []views.ConfigureOption{
		views.WithConfigureTheme(tui.DefaultTheme),
		views.WithInitialTab("plugins"),
	}

	configureView := views.NewConfigureView(opts...)
	app := tui.NewApp(configureView)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running plugins TUI: %v\n", err)
		os.Exit(1)
	}
}

// runProjectPluginsTUI launches the configure TUI at the Plugins tab for project scope
func runProjectPluginsTUI() {
	opts := []views.ConfigureOption{
		views.WithConfigureTheme(tui.DefaultTheme),
		views.WithInitialTab("plugins"),
		views.WithInitialScope("project"),
	}

	configureView := views.NewConfigureView(opts...)
	app := tui.NewApp(configureView)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running plugins TUI: %v\n", err)
		os.Exit(1)
	}
}

// runProfilePluginsTUI launches the TUI for managing a profile's plugins
func runProfilePluginsTUI(profileName string) {
	// Check if profile exists
	if !config.ProfileExists(profileName) {
		// Profile doesn't exist - ask if they want to create it
		fmt.Printf("Profile '%s' does not exist.\n", profileName)
		fmt.Print("Would you like to create it? [y/N] ")

		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Aborted.")
			return
		}

		// Create the profile
		newProfile := config.NewProfile(profileName)
		if err := newProfile.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating profile: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created profile '%s'\n", profileName)
	}

	// Launch the configure TUI at the Profiles tab
	// The user will need to select the profile to edit its plugins
	opts := []views.ConfigureOption{
		views.WithConfigureTheme(tui.DefaultTheme),
		views.WithInitialTab("profiles"),
	}

	configureView := views.NewConfigureView(opts...)
	app := tui.NewApp(configureView)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running plugins TUI: %v\n", err)
		os.Exit(1)
	}
}

// localCmd installs local plugin changes and launches an agent session for testing
var localCmd = &cobra.Command{
	Use:   "local [plugin-name...]",
	Short: "Test local plugin changes",
	Long: `Test local plugin changes by installing from the local vibe repository.

Must be run from within a local vibe repo checkout (anywhere in the tree).
Automatically finds the repo root, registers it as the claude-vibe marketplace
source, reinstalls the specified plugin(s), and launches an agent session.

If no plugin names are provided, only the local marketplace is registered.
Use --no-agent to skip launching the agent after installation.

Examples:
  cd ~/code/vibe
  vibe local                                     # Register local marketplace only
  vibe local databricks-tools                 # Install one plugin and launch agent
  vibe local databricks-tools google-tools # Install multiple plugins
  vibe local databricks-tools --no-agent      # Install without launching agent`,
	Run: func(cmd *cobra.Command, args []string) {
		noAgent, _ := cmd.Flags().GetBool("no-agent")
		runLocalDev(args, noAgent)
	},
}

// findVibeRoot walks up from the current directory to find the vibe repository root
// by looking for .claude-plugin/marketplace.json or .claude-plugin/plugin.json
func findVibeRoot(startDir string) (string, error) {
	dir := startDir
	for {
		// Check for marketplace.json (preferred marker)
		marketplacePath := filepath.Join(dir, ".claude-plugin", "marketplace.json")
		if _, err := os.Stat(marketplacePath); err == nil {
			return dir, nil
		}

		// Check for plugin.json as fallback
		pluginPath := filepath.Join(dir, ".claude-plugin", "plugin.json")
		if _, err := os.Stat(pluginPath); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding vibe repo
			return "", fmt.Errorf("not in a vibe repository (looked for .claude-plugin/marketplace.json or .claude-plugin/plugin.json)")
		}
		dir = parent
	}
}

// runLocalDev registers the current directory as the claude-vibe marketplace,
// reinstalls the specified plugins, and optionally launches an agent session.
func runLocalDev(pluginNames []string, noAgent bool) {
	startDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Find the vibe repository root
	localPath, err := findVibeRoot(startDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please run 'vibe local' from within the vibe repository.\n")
		os.Exit(1)
	}

	fmt.Printf("Local marketplace: %s\n", localPath)
	fmt.Println()

	// Register local directory as the claude-vibe marketplace source.
	// Remove first to ensure the local version takes precedence over any
	// previously cached release. The error is intentionally ignored — if
	// claude-vibe isn't currently registered, the remove is a no-op and we
	// proceed straight to the add.
	fmt.Print("Registering local marketplace... ")
	exec.Command("claude", "plugin", "marketplace", "remove", "claude-vibe").Run()
	if err := exec.Command("claude", "plugin", "marketplace", "add", localPath).Run(); err != nil {
		fmt.Println("FAILED")
		fmt.Fprintf(os.Stderr, "Error adding marketplace: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("done")

	// Reinstall each specified plugin from the local marketplace.
	if len(pluginNames) > 0 {
		fmt.Printf("Installing %d plugin(s) from local source...\n", len(pluginNames))
		var failed []string
		for _, plugin := range pluginNames {
			fmt.Printf("  %s... ", plugin)
			installCmd := exec.Command("claude", "plugin", "install", plugin+"@claude-vibe")
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if err := installCmd.Run(); err != nil {
				fmt.Println("FAILED")
				failed = append(failed, plugin)
			} else {
				fmt.Println("ok")
			}
		}
		fmt.Println()
		if len(failed) > 0 {
			fmt.Printf("Warning: %d plugin(s) failed to install: %s\n", len(failed), strings.Join(failed, ", "))
			fmt.Println("Try manually: claude plugin install <plugin-name>@claude-vibe")
		}
	}

	if noAgent {
		fmt.Println("Local marketplace registered. Skipping agent launch (--no-agent).")
		return
	}

	fmt.Println("Launching agent session for local testing...")
	launchAgentSession("", profile, false, false, false, 3.0)
}

// agentCmd launches an agent session
var agentCmd = &cobra.Command{
	Use:   "agent [name]",
	Short: "Launch an agent session",
	Long: `Launch a Claude Code agent session.

If a name is provided, creates or enters that folder in the current directory
before launching the session.

If --profile is specified (and not "default"), applies the profile's plugins,
MCP servers, and permissions to the current directory before launching.

Use --voice to dictate your prompt using macOS native speech recognition.
The transcribed text is sent to Claude as the initial prompt.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		voiceMode, _ := cmd.Flags().GetBool("voice")
		voiceTimeout, _ := cmd.Flags().GetFloat64("voice-timeout")

		var projectName string
		if len(args) > 0 {
			projectName = args[0]
		}

		launchAgentSession(projectName, profile, force, dryRun, voiceMode, voiceTimeout)
	},
}

// installCmd handles vibe installation
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install vibe and all dependencies",
	Long: `Install vibe and all required dependencies.

This command installs:
  - Homebrew packages (jq, yq, gh, python, etc.)
  - Claude Code CLI
  - Vibe marketplace and plugins
  - MCP server configurations

The installer tracks progress and can resume from failures.
Use --resume to continue from a failed installation.
Use --no-brew to skip Homebrew and all brew-based installations.
  Missing tools will be listed so you can install them manually.
Use --extra-plugins to install additional plugins alongside defaults.
  Extra plugins are saved and automatically reinstalled on future updates.`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := install.DefaultOptions()

		// Parse flags
		opts.ForceReinstall, _ = cmd.Flags().GetBool("force-reinstall")
		opts.CleanOnly, _ = cmd.Flags().GetBool("clean-only")
		opts.SkipJAMF, _ = cmd.Flags().GetBool("skip-jamf")
		opts.SkipPlugins, _ = cmd.Flags().GetBool("skip-plugins")
		opts.NoBrew, _ = cmd.Flags().GetBool("no-brew")
		opts.NoInteractive, _ = cmd.Flags().GetBool("no-interactive")
		opts.Resume, _ = cmd.Flags().GetBool("resume")
		opts.Verbose = verbose

		// Parse --extra-plugins flag
		extraPluginsStr, _ := cmd.Flags().GetString("extra-plugins")
		if extraPluginsStr != "" {
			for _, p := range strings.Split(extraPluginsStr, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					opts.ExtraPlugins = append(opts.ExtraPlugins, p)
				}
			}
		}

		// Default to non-interactive. TUI must be explicitly requested
		// with --tui. The bubbletea TUI captures stdin and causes any
		// subprocess that prompts for input to hang indefinitely.
		tui, _ := cmd.Flags().GetBool("tui")
		if tui {
			if err := views.RunInstallTUI(opts); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := views.RunInstallNonInteractive(opts); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

// launchAgentSession starts an agent session
func launchAgentSession(name, profileName string, force, dryRun, voiceMode bool, voiceTimeout float64) {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// If name provided, create/enter that folder
	if name != "" {
		projectDir := workDir + "/" + name
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			if dryRun {
				fmt.Printf("Would create directory: %s\n", projectDir)
			} else {
				if err := os.MkdirAll(projectDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", projectDir, err)
					os.Exit(1)
				}
				fmt.Printf("Created directory: %s\n", projectDir)
			}
		}
		workDir = projectDir
		if !dryRun {
			if err := os.Chdir(workDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error changing to directory %s: %v\n", workDir, err)
				os.Exit(1)
			}
		}
	}

	// Resolve profile: explicit -p flag takes precedence, then active profile
	if profileName == "" {
		profileName = config.GetActiveProfile()
	}

	// If profile specified and not "default", apply profile settings
	if profileName != "" && profileName != "default" {
		claudeDir := workDir + "/.claude"
		settingsFile := claudeDir + "/settings.json"

		// Check if project-local settings exist
		if _, err := os.Stat(settingsFile); err == nil && !force {
			if dryRun {
				fmt.Printf("Would prompt: Project settings exist at %s. Overwrite? [y/N]\n", settingsFile)
			} else {
				fmt.Printf("Project settings already exist at %s\n", settingsFile)
				fmt.Print("Overwrite with profile settings? [y/N] ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Keeping existing settings.")
				} else {
					applyProfile(workDir, profileName, dryRun)
				}
			}
		} else {
			applyProfile(workDir, profileName, dryRun)
		}
	}

	if dryRun {
		_, dbexecErr := exec.LookPath("dbexec")
		_, isaacErr := exec.LookPath("isaac")
		launcher := "claude"
		if dbexecErr == nil {
			launcher = "dbexec repo run isaac"
		} else if isaacErr == nil {
			launcher = "isaac"
		}
		if voiceMode {
			fmt.Printf("Would run: voice capture then %s (in %s)\n", launcher, workDir)
		} else {
			fmt.Printf("Would run: %s (in %s)\n", launcher, workDir)
		}
		return
	}

	// Voice mode: capture speech before launching the agent
	var initialPrompt string
	if voiceMode {
		text, err := voice.Capture(voiceTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Voice capture error: %v\n", err)
			os.Exit(1)
		}
		if text == "" {
			fmt.Println("No speech captured.")
			return
		}
		initialPrompt = text
	}

	// Launch the agent session
	if initialPrompt != "" {
		fmt.Printf("\nSending: %s\n\n", initialPrompt)
	} else {
		fmt.Printf("Launching agent session in %s...\n", workDir)
	}

	// Prefer dbexec → isaac → claude (direct fallback for personal computers)
	dbexecPath, dbexecErr := exec.LookPath("dbexec")
	isaacPath, isaacErr := exec.LookPath("isaac")

	if dbexecErr == nil {
		// Databricks work machine: use dbexec repo run isaac
		args := []string{"repo", "run", "isaac"}
		if initialPrompt != "" {
			args = append(args, "--", initialPrompt)
		}
		cmd := exec.Command(dbexecPath, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error launching agent: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if isaacErr == nil {
		// Direct isaac (less common)
		var cmd *exec.Cmd
		if initialPrompt != "" {
			cmd = exec.Command(isaacPath, initialPrompt)
		} else {
			cmd = exec.Command(isaacPath)
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error launching agent: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Personal computer fallback: run claude directly
	claudePath, claudeErr := exec.LookPath("claude")
	if claudeErr != nil {
		fmt.Fprintf(os.Stderr, "Error: 'claude' command not found\n")
		fmt.Fprintf(os.Stderr, "Install Claude Code: npm install -g @anthropic-ai/claude-code\n")
		os.Exit(1)
	}

	if err := os.Chdir(workDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error changing to %s: %v\n", workDir, err)
		os.Exit(1)
	}

	claudeArgs := []string{"claude"}
	if initialPrompt != "" {
		claudeArgs = append(claudeArgs, initialPrompt)
	}

	// syscall.Exec replaces this process entirely — cleaner than cmd.Run()
	if err := syscall.Exec(claudePath, claudeArgs, os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "Error launching claude: %v\n", err)
		os.Exit(1)
	}
}

// applyProfile applies a profile's settings to a directory
func applyProfile(dir, profileName string, dryRun bool) {
	if dryRun {
		fmt.Printf("Would apply profile '%s' to %s\n", profileName, dir)
		return
	}

	fmt.Printf("Applying profile '%s' to %s...\n", profileName, dir)

	// Load the profile from ~/.vibe/profiles/<name>.yaml
	loadedProfile, err := config.LoadProfile(profileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading profile '%s': %v\n", profileName, err)
		os.Exit(1)
	}

	// Try to apply the profile - this may fail if dir is home directory
	result, err := loadedProfile.Apply(dir)
	if err != nil {
		// Check if this is a home directory safety error
		if config.IsHomeDirectoryError(err) {
			fmt.Println()
			fmt.Println("⚠️  WARNING: You are attempting to apply a profile to your home directory (~).")
			fmt.Println()
			fmt.Println("This will modify your USER-SCOPED settings, not project-scoped settings:")
			fmt.Println("  • ~/.claude/settings.json - permissions and plugins")
			fmt.Println("  • ~/.claude.json - MCP servers (will be MERGED, not replaced)")
			fmt.Println("  • Plugins will be installed at USER scope (affects all projects)")
			fmt.Println()
			fmt.Print("Are you sure you want to continue? [y/N] ")

			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted.")
				return
			}

			// Retry with Force option
			result, err = loadedProfile.ApplyWithOptions(dir, config.ApplyOptions{Force: true})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error applying profile: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error applying profile: %v\n", err)
			os.Exit(1)
		}
	}

	// Print warnings if any
	for _, warning := range result.Warnings {
		fmt.Printf("⚠️  %s\n", warning)
	}

	// Print what was applied
	fmt.Printf("Applied profile '%s':\n", profileName)
	if len(loadedProfile.Plugins) > 0 {
		scope := "project"
		if result.IsHomeDirectory {
			scope = "user"
		}
		fmt.Printf("  Plugins: %d installed (%s scope)", result.PluginsInstalled, scope)
		if len(result.PluginErrors) > 0 {
			fmt.Printf(" (%d failed)", len(result.PluginErrors))
		}
		fmt.Println()
		// Always show plugin errors so users can see what went wrong
		if len(result.PluginErrors) > 0 {
			for _, err := range result.PluginErrors {
				fmt.Printf("    Error: %s\n", err)
			}
		}
	}
	if result.MCPServersEnabled > 0 {
		action := "enabled"
		if result.IsHomeDirectory {
			action = "merged"
		}
		fmt.Printf("  MCP Servers: %d %s\n", result.MCPServersEnabled, action)
	}
	if result.PermissionsApplied > 0 {
		fmt.Printf("  Permissions: %d configured\n", result.PermissionsApplied)
	}
}

// syncCmd syncs MCP servers and skills to other coding agents
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync MCP servers and skills to Codex and Cursor",
	Long: `Sync vibe's MCP server configurations and skills from Claude Code
(source of truth) to other AI coding agents like OpenAI Codex CLI and Cursor.

By default, syncs to all installed targets. Use --target to sync to a specific one.
Use --status to check what is currently in sync without making changes.
Use --dry-run to preview what would be written.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetName, _ := cmd.Flags().GetString("target")
		showStatus, _ := cmd.Flags().GetBool("status")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if showStatus {
			runSyncStatus(targetName)
			return
		}

		runSync(targetName, dryRun)
	},
}

// runSync executes the sync operation
func runSync(targetName string, dryRun bool) {
	targets := vibesync.AllTargets()

	// Filter by target name if specified
	if targetName != "" {
		targets = vibesync.FilterTargets(targets, []string{targetName})
		if len(targets) == 0 {
			fmt.Fprintf(os.Stderr, "Error: unknown target %q (available: codex, cursor)\n", targetName)
			os.Exit(1)
		}
	} else {
		// If no target specified, check config for sync_targets
		cfg, err := config.Load()
		if err == nil && len(cfg.Settings.SyncTargets) > 0 {
			targets = vibesync.FilterTargets(targets, cfg.Settings.SyncTargets)
		}
	}

	opts := vibesync.SyncOptions{
		DryRun:  dryRun,
		Verbose: verbose,
	}

	if dryRun {
		fmt.Println("Dry run - no files will be written")
		fmt.Println()
	}

	results := vibesync.RunSync(targets, opts)

	// Print results
	anySuccess := false
	for _, r := range results {
		if r.ItemsSkipped > 0 && len(r.Errors) > 0 {
			fmt.Printf("%s: %s\n", r.Target, r.Errors[0])
			continue
		}

		fmt.Printf("%s: %d items synced", r.Target, r.ItemsSynced)
		if dryRun {
			fmt.Print(" (dry-run)")
		}
		fmt.Println()

		if verbose || dryRun {
			for _, detail := range r.Details {
				fmt.Printf("  %s\n", detail)
			}
		}

		if len(r.Errors) > 0 {
			for _, e := range r.Errors {
				fmt.Printf("  Warning: %s\n", e)
			}
		}

		if r.ItemsSynced > 0 || len(r.Errors) == 0 {
			anySuccess = true
		}
	}

	if !anySuccess && !dryRun {
		os.Exit(1)
	}
}

// runSyncStatus displays the current sync status
func runSyncStatus(targetName string) {
	targets := vibesync.AllTargets()

	if targetName != "" {
		targets = vibesync.FilterTargets(targets, []string{targetName})
		if len(targets) == 0 {
			fmt.Fprintf(os.Stderr, "Error: unknown target %q (available: codex, cursor)\n", targetName)
			os.Exit(1)
		}
	}

	statuses := vibesync.RunStatus(targets)

	for _, s := range statuses {
		fmt.Printf("=== %s ===\n", s.Target)

		if !s.LastSynced.IsZero() {
			fmt.Printf("  Last synced: %s\n", s.LastSynced.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Println("  Never synced")
		}

		// MCP status
		if len(s.MCPInSync) > 0 {
			fmt.Printf("  MCP in sync: %s\n", strings.Join(s.MCPInSync, ", "))
		}
		if len(s.MCPMissing) > 0 {
			fmt.Printf("  MCP missing: %s\n", strings.Join(s.MCPMissing, ", "))
		}
		if len(s.MCPExtra) > 0 {
			fmt.Printf("  MCP extra (will be removed on sync): %s\n", strings.Join(s.MCPExtra, ", "))
		}

		// Skills status
		if len(s.SkillsInSync) > 0 {
			fmt.Printf("  Skills in sync: %s\n", strings.Join(s.SkillsInSync, ", "))
		}
		if len(s.SkillsStale) > 0 {
			fmt.Printf("  Skills stale: %s\n", strings.Join(s.SkillsStale, ", "))
		}
		if len(s.SkillsMissing) > 0 {
			fmt.Printf("  Skills missing: %s\n", strings.Join(s.SkillsMissing, ", "))
		}

		fmt.Println()
	}
}

// runAutoSync is called after install/update to auto-sync if configured.
// It runs silently and never fails the parent operation.
func runAutoSync() {
	cfg, err := config.Load()
	if err != nil || !cfg.Settings.AutoSync {
		return
	}

	targets := vibesync.AllTargets()
	if len(cfg.Settings.SyncTargets) > 0 {
		targets = vibesync.FilterTargets(targets, cfg.Settings.SyncTargets)
	}

	// Only sync to installed targets
	var installed []vibesync.SyncTarget
	for _, t := range targets {
		if t.IsInstalled() {
			installed = append(installed, t)
		}
	}

	if len(installed) == 0 {
		return
	}

	fmt.Println()
	fmt.Print("Auto-syncing to other agents... ")

	results := vibesync.RunSync(installed, vibesync.SyncOptions{})

	totalSynced := 0
	for _, r := range results {
		totalSynced += r.ItemsSynced
	}

	if totalSynced > 0 {
		fmt.Printf("done (%d items synced)\n", totalSynced)
	} else {
		fmt.Println("done (nothing to sync)")
	}
}

// telemetryCmd is the parent command for telemetry operations
var telemetryCmd = &cobra.Command{
	Use:    "telemetry",
	Short:  "Manage telemetry and event publishing",
	Hidden: true, // Hide from help output - internal use only
	Long: `Manage telemetry operations for vibe.

Use 'vibe telemetry publish' to send usage events.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// telemetryPublishCmd publishes telemetry events
var telemetryPublishCmd = &cobra.Command{
	Use:   "publish [json-payload]",
	Short: "Publish a telemetry event",
	Long: `Publish a telemetry event.

The event type is specified with --event-type flag.
The JSON payload can be provided as an argument, via stdin, or parsed from a transcript.

Examples:
  # Publish with inline JSON
  vibe telemetry publish --event-type="claude.stop" '{"session_id": "abc123", "duration": 120}'

  # Publish from stdin
  echo '{"session_id": "abc123"}' | vibe telemetry publish --event-type="claude.stop"

  # Publish from a Claude Code transcript file (auto-extracts stats)
  vibe telemetry publish --event-type="claude.session.stop" --transcript=/path/to/transcript.jsonl

  # Use in a Claude Code stop hook (reads transcript_path from hook input on stdin)
  vibe telemetry publish --event-type="claude.session.stop" --from-hook

  # Use --quiet to suppress all output (recommended for hooks)
  vibe telemetry publish --event-type="claude.session.stop" --from-hook --quiet`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eventType, _ := cmd.Flags().GetString("event-type")
		source, _ := cmd.Flags().GetString("source")
		transcriptPath, _ := cmd.Flags().GetString("transcript")
		fromHook, _ := cmd.Flags().GetBool("from-hook")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		quiet, _ := cmd.Flags().GetBool("quiet")

		// Helper to exit silently in quiet mode
		exitWithError := func(format string, args ...interface{}) {
			if !quiet {
				fmt.Fprintf(os.Stderr, format+"\n", args...)
			}
			os.Exit(1)
		}

		// Check if telemetry is configured (unless dry-run)
		if !dryRun && !telemetry.IsConfigured() {
			if !quiet {
				fmt.Fprintln(os.Stderr, "Error: telemetry is not configured")
				fmt.Fprintln(os.Stderr, "This binary was not built with telemetry credentials.")
				if verbose {
					cfg := telemetry.GetConfiguration()
					fmt.Fprintf(os.Stderr, "\nConfiguration:\n")
					fmt.Fprintf(os.Stderr, "  client_id:         %s\n", cfg["client_id"])
					fmt.Fprintf(os.Stderr, "  secret:            %s\n", cfg["secret"])
					fmt.Fprintf(os.Stderr, "  endpoint:          %s\n", cfg["endpoint"])
					fmt.Fprintf(os.Stderr, "  unity_catalog_url: %s\n", cfg["unity_catalog_url"])
					fmt.Fprintf(os.Stderr, "  table:             %s\n", cfg["table"])
				}
			}
			os.Exit(1)
		}

		var event *telemetry.Event

		// Handle --from-hook: read hook input from stdin to get transcript_path
		if fromHook {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				hookInput, err := os.ReadFile("/dev/stdin")
				if err != nil {
					exitWithError("Error reading hook input from stdin: %v", err)
				}

				// Parse hook input to get transcript_path
				var hookData struct {
					TranscriptPath string `json:"transcript_path"`
					SessionID      string `json:"session_id"`
				}
				if err := json.Unmarshal(hookInput, &hookData); err != nil {
					exitWithError("Error parsing hook input: %v", err)
				}

				if hookData.TranscriptPath == "" {
					exitWithError("Error: no transcript_path in hook input")
				}
				transcriptPath = hookData.TranscriptPath
			} else {
				exitWithError("Error: --from-hook requires hook input on stdin")
			}
		}

		// Handle --transcript: parse transcript file
		if transcriptPath != "" {
			stats, err := telemetry.ParseTranscript(transcriptPath)
			if err != nil {
				exitWithError("Error parsing transcript: %v", err)
			}
			event = stats.ToEvent(eventType)

			if (verbose || dryRun) && !quiet {
				fmt.Printf("Parsed transcript: %s\n", transcriptPath)
				fmt.Printf("  Session ID:         %s\n", stats.SessionID)
				fmt.Printf("  Version:            %s\n", stats.Version)
				fmt.Printf("  Model:              %s\n", stats.Model)
				fmt.Printf("  Duration:           %dms\n", stats.DurationMs)
				fmt.Printf("  User messages:      %d\n", stats.UserMessages)
				fmt.Printf("  Assistant messages: %d\n", stats.AssistMessages)
				fmt.Printf("  Token usage:\n")
				fmt.Printf("    Input:            %d\n", stats.TokenUsage.Input)
				fmt.Printf("    Output:           %d\n", stats.TokenUsage.Output)
				fmt.Printf("    Cache creation:   %d\n", stats.TokenUsage.CacheCreation)
				fmt.Printf("    Cache read:       %d\n", stats.TokenUsage.CacheRead)
				fmt.Printf("  Tools invoked:      %v\n", stats.ToolsInvoked)
				fmt.Printf("  Skills invoked:     %v\n", stats.SkillsInvoked)
				fmt.Printf("  Agents spawned:     %v\n", stats.AgentsSpawned)
				fmt.Printf("  MCP tools invoked:  %v\n", stats.MCPToolsInvoked)
				if stats.Summary != "" {
					fmt.Printf("  Summary:            %s\n", stats.Summary)
				}
			}

			if dryRun {
				if !quiet {
					fmt.Println("\nDry run - not publishing")
				}
				return
			}
		} else if len(args) > 0 {
			// Raw JSON payload from args - validate it's valid JSON
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(args[0]), &payload); err != nil {
				exitWithError("Error parsing JSON payload: %v", err)
			}
			event = &telemetry.Event{
				EventType: eventType,
				Payload:   args[0], // Store as JSON string
			}
		} else {
			// Check if stdin has data
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				jsonPayload, err := os.ReadFile("/dev/stdin")
				if err != nil {
					exitWithError("Error reading from stdin: %v", err)
				}
				// Validate it's valid JSON
				var payload map[string]interface{}
				if err := json.Unmarshal(jsonPayload, &payload); err != nil {
					exitWithError("Error parsing JSON payload: %v", err)
				}
				event = &telemetry.Event{
					EventType: eventType,
					Payload:   string(jsonPayload), // Store as JSON string
				}
			} else {
				exitWithError("Error: no payload provided. Use JSON argument, stdin, or --transcript")
			}
		}

		// Set source if provided
		if source != "" {
			event.Source = source
		}

		// Create publisher and publish
		publisher, err := telemetry.NewPublisher()
		if err != nil {
			exitWithError("Error creating publisher: %v", err)
		}

		ctx := context.Background()
		if err := publisher.Publish(ctx, event); err != nil {
			exitWithError("Error publishing event: %v", err)
		}

		if verbose && !quiet {
			fmt.Printf("Published event type '%s' successfully\n", eventType)
		}
	},
}

// profileCmd is the parent command for profile management
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage vibe profiles",
	Long: `Manage vibe profiles for switching between configurations.

Without a subcommand, lists all available profiles.`,
	Run: func(cmd *cobra.Command, args []string) {
		printProfileList()
	},
}

// profileListCmd lists all available profiles
var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Run: func(cmd *cobra.Command, args []string) {
		printProfileList()
	},
}

// profileCurrentCmd shows the currently active profile
var profileCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the active profile",
	Run: func(cmd *cobra.Command, args []string) {
		active := config.GetActiveProfile()
		if active == "" {
			fmt.Println("No active profile set.")
			fmt.Println("Use 'vibe profile switch <name>' to set one.")
			return
		}

		fmt.Printf("Active profile: %s\n", active)

		p, err := config.LoadProfile(active)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load profile details: %v\n", err)
			return
		}

		if p.Description != "" {
			fmt.Printf("Description:    %s\n", p.Description)
		}
		if p.DatabricksProfile != "" {
			fmt.Printf("Databricks:     %s\n", p.DatabricksProfile)
		}
		fmt.Printf("Plugins:        %d\n", len(p.Plugins))
		enabledMCP := 0
		for _, v := range p.MCPServers {
			if v {
				enabledMCP++
			}
		}
		fmt.Printf("MCP servers:    %d\n", enabledMCP)
	},
}

// profileShowCmd displays full details of a profile
var profileShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show profile details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if !config.ProfileExists(name) {
			fmt.Fprintf(os.Stderr, "Error: profile '%s' not found\n", name)
			os.Exit(1)
		}

		profilePath := filepath.Join(config.ProfilesDir(), name+".yaml")
		data, err := os.ReadFile(profilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading profile: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Profile: %s\n---\n%s", name, string(data))
	},
}

// profileSwitchCmd switches the active profile
var profileSwitchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch the active profile",
	Long: `Switch the active profile. The new profile will be applied
the next time 'vibe agent' is launched (affects all instances).

If the profile has a Databricks profile configured, it will be
synced to ~/.mcp.json and ~/.vibe/env immediately.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if !config.ProfileExists(name) {
			fmt.Fprintf(os.Stderr, "Error: profile '%s' not found\n", name)
			fmt.Fprintf(os.Stderr, "Available profiles: %s\n", strings.Join(config.ListProfiles(), ", "))
			os.Exit(1)
		}

		p, err := config.LoadProfile(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profile '%s': %v\n", name, err)
			os.Exit(1)
		}

		if err := config.SetActiveProfile(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting active profile: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Switched to profile '%s'\n", name)

		if p.Description != "" {
			fmt.Printf("  Description: %s\n", p.Description)
		}

		if p.DatabricksProfile != "" {
			if err := p.SyncDatabricksProfile(); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: failed to sync Databricks profile: %v\n", err)
			} else {
				fmt.Printf("  Databricks profile: %s (synced)\n", p.DatabricksProfile)
			}
		}

		if len(p.EnvOverrides) > 0 {
			if err := p.SyncEnvOverrides(); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: failed to sync env overrides: %v\n", err)
			}
		}

		fmt.Println("\nProfile will be fully applied on next 'vibe agent' launch.")
	},
}

// profileCreateCmd creates a new empty profile
var profileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := config.ValidateProfileName(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if config.ProfileExists(name) {
			fmt.Fprintf(os.Stderr, "Error: profile '%s' already exists\n", name)
			os.Exit(1)
		}

		p := config.NewProfile(name)
		if err := p.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating profile: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created profile '%s'\n", name)
		fmt.Printf("Edit at: %s/%s.yaml\n", config.ProfilesDir(), name)
	},
}

// profileDeleteCmd deletes a profile
var profileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if !config.ProfileExists(name) {
			fmt.Fprintf(os.Stderr, "Error: profile '%s' not found\n", name)
			os.Exit(1)
		}

		active := config.GetActiveProfile()
		if name == active {
			fmt.Fprintf(os.Stderr, "Error: cannot delete the active profile '%s'\n", name)
			fmt.Fprintf(os.Stderr, "Switch to a different profile first: vibe profile switch <other>\n")
			os.Exit(1)
		}

		fmt.Printf("Delete profile '%s'? [y/N] ", name)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted.")
			return
		}

		if err := config.DeleteProfile(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting profile: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Deleted profile '%s'\n", name)
	},
}

func printProfileList() {
	profiles := config.ListProfiles()
	if len(profiles) == 0 {
		fmt.Println("No profiles configured.")
		fmt.Println("Use 'vibe profile create <name>' to create one.")
		return
	}

	active := config.GetActiveProfile()
	fmt.Println("Profiles:")
	for _, name := range profiles {
		marker := "  "
		if name == active {
			marker = "* "
		}

		p, err := config.LoadProfile(name)
		if err != nil {
			fmt.Printf("  %s%s (error loading)\n", marker, name)
			continue
		}

		desc := ""
		if p.Description != "" {
			desc = " - " + p.Description
		}
		dbInfo := ""
		if p.DatabricksProfile != "" {
			dbInfo = fmt.Sprintf(" [databricks: %s]", p.DatabricksProfile)
		}
		fmt.Printf("  %s%s%s%s\n", marker, name, desc, dbInfo)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "Specify profile name")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Update command flags
	updateCmd.Flags().String("version", "", "Install a specific version (e.g., v1.2.3)")
	updateCmd.Flags().Bool("versions", false, "List all available versions")

	// Doctor command flags
	doctorCmd.Flags().Bool("repair", false, "Automatically repair issues without prompting")
	doctorCmd.Flags().Bool("collect", false, "Generate diagnostic tar.gz for sharing")

	// Configure command flags
	configureCmd.Flags().StringP("tab", "t", "", "Jump to specific tab (settings, mcp, plugins, profiles)")

	// Plugins command flags
	pluginsCmd.Flags().Bool("installed", false, "Show only installed plugins")
	pluginsCmd.Flags().Bool("global", false, "Open interactive TUI for user-scope plugins")
	pluginsCmd.Flags().Bool("user", false, "Open interactive TUI for user-scope plugins (alias for --global)")
	pluginsCmd.Flags().String("publish", "", "Publish a plugin to vibe marketplace (optional: path to plugin directory)")
	pluginsCmd.Flags().Lookup("publish").NoOptDefVal = "."
	pluginsCmd.Flags().Bool("update", false, "Refresh marketplace from latest release and reinstall all plugins")
	pluginsCmd.Flags().Bool("latest", false, "With --update, pull marketplace from main branch instead of latest release")

	// Local command flags
	localCmd.Flags().Bool("no-agent", false, "Skip launching the agent after installing plugins")

	// Agent command flags
	agentCmd.Flags().Bool("force", false, "Don't prompt for overwrite, just do it")
	agentCmd.Flags().Bool("dry-run", false, "Show what would be done without doing it")
	agentCmd.Flags().Bool("voice", false, "Use macOS voice dictation for the prompt")
	agentCmd.Flags().Float64("voice-timeout", 3.0, "Seconds of silence before auto-sending voice input")

	// Install command flags
	installCmd.Flags().Bool("force-reinstall", false, "Clean-slate installation (removes caches/configs)")
	installCmd.Flags().Bool("clean-only", false, "Only run cleanup, don't reinstall (use with --force-reinstall)")
	installCmd.Flags().Bool("skip-jamf", false, "Skip Databricks CLI installation")
	installCmd.Flags().Bool("skip-plugins", false, "Skip plugin installation")
	installCmd.Flags().Bool("no-brew", false, "Skip Homebrew and brew-based installations (lists missing tools instead)")
	installCmd.Flags().Bool("no-interactive", false, "Run without TUI — deprecated, non-interactive is now the default")
	installCmd.Flags().Bool("tui", false, "Run with interactive TUI (opt-in)")
	installCmd.Flags().Bool("resume", false, "Resume from failed installation")
	installCmd.Flags().String("extra-plugins", "", "Comma-separated list of additional plugins to install alongside defaults")

	// Sync command flags
	syncCmd.Flags().String("target", "", "Sync to a specific target (codex, cursor)")
	syncCmd.Flags().Bool("status", false, "Show sync status without making changes")
	syncCmd.Flags().Bool("dry-run", false, "Preview what would be written without writing")

	// Telemetry publish command flags
	telemetryPublishCmd.Flags().String("event-type", "", "Event type identifier (required)")
	telemetryPublishCmd.Flags().String("source", "", "Source identifier for the event (e.g., 'claude-code-stop-hook')")
	telemetryPublishCmd.Flags().String("transcript", "", "Path to Claude Code transcript JSONL file")
	telemetryPublishCmd.Flags().Bool("from-hook", false, "Read transcript_path from Claude Code hook input on stdin")
	telemetryPublishCmd.Flags().Bool("dry-run", false, "Parse and display stats without publishing")
	telemetryPublishCmd.Flags().Bool("quiet", false, "Suppress all output (recommended for hooks)")
	telemetryPublishCmd.MarkFlagRequired("event-type")

	// Add telemetry subcommands
	telemetryCmd.AddCommand(telemetryPublishCmd)

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(pluginsCmd)
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(localCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(telemetryCmd)

	// Profile subcommands
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileCurrentCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileSwitchCmd)
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	rootCmd.AddCommand(profileCmd)
}

// ensureLocalBinInPath adds ~/.local/bin to PATH if not already present.
// This ensures commands like 'claude' can be found even if the user hasn't
// restarted their shell after installation.
func ensureLocalBinInPath() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	localBin := filepath.Join(homeDir, ".local", "bin")
	currentPath := os.Getenv("PATH")

	// Check if already in PATH
	for _, p := range filepath.SplitList(currentPath) {
		if p == localBin {
			return
		}
	}

	// Prepend to PATH
	os.Setenv("PATH", localBin+string(os.PathListSeparator)+currentPath)
}

func main() {
	// Ensure ~/.local/bin is in PATH so we can find claude
	ensureLocalBinInPath()

	// Flush stderr buffer
	earlyinit.FlushFiltered()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
