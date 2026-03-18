package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// CommandFinder is a function that finds a command by name.
// It returns the path to the command or an error if not found.
type CommandFinder func(name string) (string, error)

// CommandRunner is a function that runs a command and returns its output.
type CommandRunner func(name string, args ...string) (string, error)

// SessionLauncher is a function that launches an interactive session.
// It receives the working directory, command name, and arguments.
type SessionLauncher func(dir, name string, args ...string) error

// ClaudeAgentOptions configures a ClaudeAgent instance.
type ClaudeAgentOptions struct {
	// HomeDir overrides the home directory (default: os.UserHomeDir)
	HomeDir string

	// BinaryName overrides the binary name (default: "claude")
	BinaryName string

	// CommandFinder overrides how commands are found (default: exec.LookPath)
	CommandFinder CommandFinder

	// CommandRunner overrides how commands are run (default: exec.Command)
	CommandRunner CommandRunner

	// SessionLauncher overrides how sessions are launched (default: syscall.Exec)
	SessionLauncher SessionLauncher
}

// ClaudeAgent implements the Agent interface for Claude Code.
type ClaudeAgent struct {
	homeDir         string
	binaryName      string
	commandFinder   CommandFinder
	commandRunner   CommandRunner
	sessionLauncher SessionLauncher
}

// NewClaudeAgent creates a new ClaudeAgent with default configuration.
func NewClaudeAgent() *ClaudeAgent {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	return NewClaudeAgentWithHome(homeDir)
}

// NewClaudeAgentWithHome creates a new ClaudeAgent with a custom home directory.
func NewClaudeAgentWithHome(homeDir string) *ClaudeAgent {
	return NewClaudeAgentWithOptions(ClaudeAgentOptions{
		HomeDir: homeDir,
	})
}

// NewClaudeAgentWithOptions creates a new ClaudeAgent with custom options.
func NewClaudeAgentWithOptions(opts ClaudeAgentOptions) *ClaudeAgent {
	agent := &ClaudeAgent{
		homeDir:    opts.HomeDir,
		binaryName: opts.BinaryName,
	}

	// Set defaults
	if agent.homeDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			agent.homeDir = homeDir
		}
	}

	if agent.binaryName == "" {
		agent.binaryName = "claude"
	}

	if opts.CommandFinder != nil {
		agent.commandFinder = opts.CommandFinder
	} else {
		agent.commandFinder = exec.LookPath
	}

	if opts.CommandRunner != nil {
		agent.commandRunner = opts.CommandRunner
	} else {
		agent.commandRunner = defaultCommandRunner
	}

	if opts.SessionLauncher != nil {
		agent.sessionLauncher = opts.SessionLauncher
	} else {
		agent.sessionLauncher = defaultSessionLauncher
	}

	return agent
}

// Name returns the agent identifier.
func (c *ClaudeAgent) Name() string {
	return "claude"
}

// IsInstalled checks if Claude Code is available on the system.
func (c *ClaudeAgent) IsInstalled() bool {
	_, err := c.commandFinder(c.binaryName)
	return err == nil
}

// Version returns the installed version of Claude Code.
func (c *ClaudeAgent) Version() (string, error) {
	output, err := c.commandRunner(c.binaryName, "--version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// LaunchSession starts an interactive Claude Code session.
func (c *ClaudeAgent) LaunchSession(opts SessionOptions) error {
	workDir := opts.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// Profile application happens before this call in the vibe agent command,
	// so we don't pass profile flags to claude directly
	return c.sessionLauncher(workDir, c.binaryName)
}

// ConfigPaths returns paths to Claude Code configuration files.
func (c *ClaudeAgent) ConfigPaths() AgentPaths {
	return AgentPaths{
		Settings:     filepath.Join(c.homeDir, ".claude", "settings.json"),
		Plugins:      filepath.Join(c.homeDir, ".claude", "plugins", "installed_plugins.json"),
		MCPConfig:    filepath.Join(c.homeDir, ".config", "mcp", "config.json"),
		GlobalConfig: filepath.Join(c.homeDir, ".claude.json"),
	}
}

// defaultCommandRunner runs a command and returns its output as a string.
func defaultCommandRunner(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// defaultSessionLauncher launches an interactive session by replacing the current process.
func defaultSessionLauncher(dir, name string, args ...string) error {
	binary, err := exec.LookPath(name)
	if err != nil {
		return err
	}

	// Change to the working directory
	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			return err
		}
	}

	// Build the full argument list (command name must be first)
	fullArgs := append([]string{name}, args...)

	// Replace the current process with claude
	return syscall.Exec(binary, fullArgs, os.Environ())
}

// init registers the Claude agent when the package is imported.
// This is commented out to allow explicit registration control.
// Uncomment if auto-registration is desired.
// func init() {
//     Register("claude", NewClaudeAgent())
// }
