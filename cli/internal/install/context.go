package install

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Context holds shared state across installation steps
type Context struct {
	// Options for this installation
	Options *Options

	// Home directory
	HomeDir string

	// Shell configuration
	ShellType string // "zsh" or "bash"
	ShellRC   string // Path to shell rc file (e.g., ~/.zshrc)

	// Architecture
	Arch       string // "arm64" or "amd64"
	IsAppleSi  bool   // True if Apple Silicon Mac
	HomebrewPrefix string // /opt/homebrew or /usr/local

	// Downloaded vibe directory (set by download step)
	VibeDir string

	// Marketplace directory (permanent location)
	MarketplaceDir string

	// Paths
	LocalBinDir string // ~/.local/bin
	ConfigDir   string // ~/.config
	VibeDataDir string // ~/.vibe
	ClaudeDir   string // ~/.claude

	// Sudo password cached (if collected)
	SudoPassword string

	// Verbose logging function
	Log func(format string, args ...interface{})

	// LogChan is used to send log messages to the TUI for real-time display
	LogChan chan string
}

// NewContext creates a new installation context
func NewContext(opts *Options) (*Context, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Force gh CLI to use github.com as the default host for the duration of
	// the install process. On machines where Databricks GHE was previously
	// configured as the default gh host, any gh invocation without an explicit
	// --hostname flag (including those made by subprocesses like
	// `claude plugin install`) will otherwise hit the internal GHE host,
	// triggering a reauthentication prompt that hangs the TUI indefinitely.
	os.Setenv("GH_HOST", "github.com")

	ctx := &Context{
		Options:        opts,
		HomeDir:        home,
		LocalBinDir:    filepath.Join(home, ".local", "bin"),
		ConfigDir:      filepath.Join(home, ".config"),
		VibeDataDir:    filepath.Join(home, ".vibe"),
		ClaudeDir:      filepath.Join(home, ".claude"),
		MarketplaceDir: filepath.Join(home, ".vibe", "marketplace"),
		Log:            func(format string, args ...interface{}) {}, // no-op by default
	}

	// Detect architecture
	ctx.Arch = runtime.GOARCH
	ctx.IsAppleSi = runtime.GOARCH == "arm64"
	if ctx.IsAppleSi {
		ctx.HomebrewPrefix = "/opt/homebrew"
	} else {
		ctx.HomebrewPrefix = "/usr/local"
	}

	// Detect shell
	ctx.detectShell()

	return ctx, nil
}

// detectShell determines the user's shell and rc file
func (ctx *Context) detectShell() {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		ctx.ShellType = "zsh"
		ctx.ShellRC = filepath.Join(ctx.HomeDir, ".zshrc")
	} else if strings.Contains(shell, "bash") {
		ctx.ShellType = "bash"
		ctx.ShellRC = filepath.Join(ctx.HomeDir, ".bashrc")
	} else {
		// Default to bash
		ctx.ShellType = "bash"
		ctx.ShellRC = filepath.Join(ctx.HomeDir, ".bashrc")
	}
}

// IsCommandInstalled checks if a command is available in PATH
func (ctx *Context) IsCommandInstalled(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetCommandPath returns the full path to a command
func (ctx *Context) GetCommandPath(cmd string) (string, error) {
	return exec.LookPath(cmd)
}

// RunCommand executes a command and returns the output
func (ctx *Context) RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCommandSilent executes a command without capturing output
func (ctx *Context) RunCommandSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// RunSudoCommand executes a command with sudo
func (ctx *Context) RunSudoCommand(args ...string) (string, error) {
	sudoArgs := append([]string{"-S"}, args...)
	cmd := exec.Command("sudo", sudoArgs...)

	// If we have a cached password, use it
	if ctx.SudoPassword != "" {
		cmd.Stdin = strings.NewReader(ctx.SudoPassword + "\n")
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// FileExists checks if a file exists
func (ctx *Context) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func (ctx *Context) DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// EnsureDir creates a directory if it doesn't exist
func (ctx *Context) EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// GetFileOwner returns the owner of a file (macOS specific)
func (ctx *Context) GetFileOwner(path string) (string, error) {
	output, err := ctx.RunCommand("stat", "-f", "%Su", path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// ExpandPath expands ~ to home directory
func (ctx *Context) ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(ctx.HomeDir, path[2:])
	}
	return path
}
