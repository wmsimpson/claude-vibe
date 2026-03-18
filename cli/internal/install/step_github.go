package install

import (
	"os/exec"
	"strings"
)

// GitHubAuthStep ensures GitHub CLI is authenticated
type GitHubAuthStep struct{}

func (s *GitHubAuthStep) ID() string          { return "github_auth" }
func (s *GitHubAuthStep) Name() string        { return "GitHub Auth" }
func (s *GitHubAuthStep) Description() string { return "Authenticate GitHub CLI" }
func (s *GitHubAuthStep) ActiveForm() string  { return "Authenticating GitHub" }
func (s *GitHubAuthStep) CanSkip(opts *Options) bool { return false }
func (s *GitHubAuthStep) NeedsSudo() bool     { return false }

func (s *GitHubAuthStep) Check(ctx *Context) (bool, error) {
	if !ctx.IsCommandInstalled("gh") {
		return false, nil
	}

	// Always specify --hostname github.com — machines with Databricks GHE
	// configured will otherwise default to the internal host and hang on auth.
	cmd := exec.Command("gh", "auth", "status", "--hostname", "github.com")
	err := cmd.Run()
	return err == nil, nil
}

func (s *GitHubAuthStep) Run(ctx *Context) StepResult {
	if !ctx.IsCommandInstalled("gh") {
		return FailureWithHint(
			"GitHub CLI not installed",
			nil,
			"Install gh first: brew install gh",
		)
	}

	// Always specify --hostname github.com to avoid hitting internal GHE hosts.
	cmd := exec.Command("gh", "auth", "status", "--hostname", "github.com")
	if err := cmd.Run(); err == nil {
		// Get user info — specify hostname to avoid GHE GraphQL hang
		user, _ := ctx.RunCommand("gh", "api", "user", "-q", ".login", "--hostname", "github.com")
		return SuccessWithDetails(
			"GitHub CLI authenticated",
			"  Logged in as: "+strings.TrimSpace(user),
		)
	}

	// Interactive auth (gh auth login --web) cannot run inside the TUI because
	// bubbletea captures the terminal — the device code would be hidden and the
	// command would hang indefinitely. Instruct the user to authenticate first.
	return FailureWithHint(
		"GitHub CLI not authenticated",
		nil,
		"Run this BEFORE vibe install:\n  gh auth login --web --hostname github.com --git-protocol ssh",
	)
}
