package install

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

// PreflightStep performs pre-installation checks (local only — no network calls)
type PreflightStep struct{}

func (s *PreflightStep) ID() string                      { return "preflight" }
func (s *PreflightStep) Name() string                    { return "Pre-flight Check" }
func (s *PreflightStep) Description() string             { return "Verify system requirements" }
func (s *PreflightStep) ActiveForm() string              { return "Verifying system requirements" }
func (s *PreflightStep) CanSkip(opts *Options) bool      { return false }
func (s *PreflightStep) NeedsSudo() bool                 { return false }

func (s *PreflightStep) Check(ctx *Context) (bool, error) {
	// Pre-flight always runs
	return false, nil
}

func (s *PreflightStep) Run(ctx *Context) StepResult {
	var details []string

	// Check platform
	switch runtime.GOOS {
	case "darwin":
		details = append(details, "macOS detected")
		if runtime.GOARCH == "arm64" {
			details = append(details, "Apple Silicon (arm64)")
		} else if runtime.GOARCH == "amd64" {
			details = append(details, "Intel (amd64)")
		}
	case "linux":
		details = append(details, "Linux detected ("+runtime.GOARCH+")")
	default:
		return FailureWithHint(
			fmt.Sprintf("Unsupported platform: %s", runtime.GOOS),
			errors.New("vibe supports macOS and Linux (including WSL)"),
			"On Windows, install WSL first: wsl --install",
		)
	}

	// Check disk space (need at least 500MB free)
	var stat syscall.Statfs_t
	home, _ := os.UserHomeDir()
	if err := syscall.Statfs(home, &stat); err == nil {
		freeBytes := stat.Bavail * uint64(stat.Bsize)
		freeMB := freeBytes / (1024 * 1024)
		if freeMB < 500 {
			return FailureWithHint(
				fmt.Sprintf("Low disk space: %dMB free", freeMB),
				errors.New("at least 500MB required"),
				"Free up disk space and try again",
			)
		}
		details = append(details, fmt.Sprintf("Disk space: %dMB free", freeMB))
	}

	// Check for concurrent installation (lock file)
	lockFile := filepath.Join(home, ".vibe", "install.lock")
	if _, err := os.Stat(lockFile); err == nil {
		ctx.Log("Lock file exists at %s", lockFile)
	}

	// Create lock file
	if err := os.MkdirAll(filepath.Dir(lockFile), 0755); err == nil {
		if f, err := os.Create(lockFile); err == nil {
			f.WriteString(fmt.Sprintf("%d", os.Getpid()))
			f.Close()
		}
	}

	// Check Claude Code is installed (warn only — install separately)
	if ctx.IsCommandInstalled("claude") {
		details = append(details, "Claude Code: installed")
	} else {
		details = append(details, "Claude Code: NOT FOUND — install before proceeding:")
		details = append(details, "  curl -fsSL https://claude.ai/install.sh | bash")
	}

	// Check jq/yq (needed for MCP and permissions sync)
	for _, tool := range []string{"jq", "yq"} {
		if ctx.IsCommandInstalled(tool) {
			details = append(details, tool+": installed")
		} else {
			details = append(details, tool+": NOT FOUND — install with: brew install "+tool)
		}
	}

	return SuccessWithDetails(
		"System requirements verified",
		joinDetails(details),
	)
}

func joinDetails(details []string) string {
	result := ""
	for _, d := range details {
		result += "  - " + d + "\n"
	}
	return result
}
