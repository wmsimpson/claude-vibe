package install

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GcloudCliStep installs the Google Cloud SDK
type GcloudCliStep struct{}

func (s *GcloudCliStep) ID() string          { return "gcloud_cli" }
func (s *GcloudCliStep) Name() string        { return "gcloud CLI" }
func (s *GcloudCliStep) Description() string { return "Install Google Cloud SDK" }
func (s *GcloudCliStep) ActiveForm() string  { return "Installing gcloud CLI" }
func (s *GcloudCliStep) CanSkip(opts *Options) bool { return opts.NoBrew }
func (s *GcloudCliStep) NeedsSudo() bool     { return false }

// gcloudKnownPaths returns paths where gcloud may be installed
func gcloudKnownPaths(ctx *Context) []string {
	return []string{
		// Homebrew (current cask name: gcloud-cli, alias: google-cloud-sdk)
		filepath.Join(ctx.HomebrewPrefix, "bin", "gcloud"),
		filepath.Join(ctx.HomebrewPrefix, "share", "google-cloud-sdk", "bin", "gcloud"),
		// Google's official standalone installer (curl https://sdk.cloud.google.com | bash)
		filepath.Join(ctx.HomeDir, "google-cloud-sdk", "bin", "gcloud"),
		// System-wide installs
		"/usr/local/google-cloud-sdk/bin/gcloud",
		"/usr/local/bin/gcloud",
		"/usr/bin/gcloud",
		// Snap (Linux)
		"/snap/bin/gcloud",
		"/snap/google-cloud-sdk/current/bin/gcloud",
	}
}

// findGcloud returns the path to gcloud if found, or empty string
func findGcloud(ctx *Context) string {
	// Check PATH first (covers any install method that updated PATH)
	if path, err := exec.LookPath("gcloud"); err == nil {
		return path
	}

	// Check known installation locations
	for _, p := range gcloudKnownPaths(ctx) {
		if ctx.FileExists(p) {
			return p
		}
	}

	// Last resort: ask the shell (catches installs that set PATH in shell RC
	// but aren't in the current process PATH)
	if output, err := exec.Command("bash", "-l", "-c", "which gcloud").Output(); err == nil {
		if p := strings.TrimSpace(string(output)); p != "" && ctx.FileExists(p) {
			return p
		}
	}

	return ""
}

func (s *GcloudCliStep) Check(ctx *Context) (bool, error) {
	return findGcloud(ctx) != "", nil
}

func (s *GcloudCliStep) Run(ctx *Context) StepResult {
	// Check if gcloud is already installed (PATH + known paths)
	if gcloudPath := findGcloud(ctx); gcloudPath != "" {
		ver, _ := ctx.RunCommand(gcloudPath, "--version")
		firstLine := strings.Split(ver, "\n")[0]
		return SuccessWithDetails("gcloud CLI already installed", "  "+firstLine)
	}

	// In --no-brew mode, report missing and continue
	if ctx.Options.NoBrew {
		return SuccessWithDetails(
			"Brew skipped: gcloud not found",
			"  Missing tools (install manually for best results):\n    - gcloud (Google Cloud SDK)\n      Install: https://cloud.google.com/sdk/docs/install\n",
		)
	}

	ctx.Log("Installing Google Cloud SDK via Homebrew...")

	// Install via Homebrew cask (google-cloud-sdk is an alias for gcloud-cli)
	cmd := exec.Command("brew", "install", "--cask", "google-cloud-sdk")
	cmd.Stdin = nil
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		ctx.Log("Homebrew cask install failed: %s", outputStr)

		// Check if brew says it's already installed
		if !strings.Contains(outputStr, "already installed") {
			// Try the formula instead
			cmd = exec.Command("brew", "install", "google-cloud-sdk")
			cmd.Stdin = nil
			if output, err := cmd.CombinedOutput(); err != nil {
				ctx.Log("Homebrew formula install also failed: %s", string(output))
			}
		}
	}

	// After install attempts, check if gcloud is actually available
	// (brew may report errors but gcloud could still be functional)
	if gcloudPath := findGcloud(ctx); gcloudPath != "" {
		s.configureShellPath(ctx)
		ver, _ := ctx.RunCommand(gcloudPath, "--version")
		firstLine := strings.Split(ver, "\n")[0]
		return SuccessWithDetails("gcloud CLI installed", "  "+firstLine)
	}

	// Also try adding homebrew bin to PATH for this process and re-check
	os.Setenv("PATH", os.Getenv("PATH")+":"+filepath.Join(ctx.HomebrewPrefix, "bin"))
	if gcloudPath := findGcloud(ctx); gcloudPath != "" {
		s.configureShellPath(ctx)
		ver, _ := ctx.RunCommand(gcloudPath, "--version")
		firstLine := strings.Split(ver, "\n")[0]
		return SuccessWithDetails("gcloud CLI installed", "  "+firstLine)
	}

	return FailureWithHint(
		"Failed to install Google Cloud SDK",
		err,
		"Try running manually: brew install --cask google-cloud-sdk",
	)
}

// configureShellPath adds gcloud path sourcing to the user's shell RC file
func (s *GcloudCliStep) configureShellPath(ctx *Context) {
	// Check both current and legacy path.zsh.inc locations
	pathFiles := []string{
		filepath.Join(ctx.HomebrewPrefix, "share", "google-cloud-sdk", "path.zsh.inc"),
		filepath.Join(ctx.HomebrewPrefix, "Caskroom", "gcloud-cli", "latest", "google-cloud-sdk", "path.zsh.inc"),
		filepath.Join(ctx.HomebrewPrefix, "Caskroom", "google-cloud-sdk", "latest", "google-cloud-sdk", "path.zsh.inc"),
	}

	for _, gcloudPathFile := range pathFiles {
		if ctx.FileExists(gcloudPathFile) {
			content, err := os.ReadFile(ctx.ShellRC)
			if err == nil && !strings.Contains(string(content), "google-cloud-sdk") {
				f, err := os.OpenFile(ctx.ShellRC, os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					f.WriteString("\n# Google Cloud SDK\nsource \"" + gcloudPathFile + "\"\n")
					f.Close()
				}
			}
			return
		}
	}
}
