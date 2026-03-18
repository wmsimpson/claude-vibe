package install

import (
	"os"
	"path/filepath"
)

// DownloadVibeStep locates the vibe repository on the local machine.
// Since the CLI is built from source, the repo is always present locally.
// We look for it via VIBE_REPO_PATH env var or common default locations.
type DownloadVibeStep struct{}

func (s *DownloadVibeStep) ID() string          { return "download_vibe" }
func (s *DownloadVibeStep) Name() string        { return "Locate Vibe Repo" }
func (s *DownloadVibeStep) Description() string { return "Locate the vibe repository on this machine" }
func (s *DownloadVibeStep) ActiveForm() string  { return "Locating vibe repository" }
func (s *DownloadVibeStep) CanSkip(opts *Options) bool { return false }
func (s *DownloadVibeStep) NeedsSudo() bool     { return false }

func (s *DownloadVibeStep) Check(ctx *Context) (bool, error) {
	return false, nil // always run to set ctx.VibeDir
}

func (s *DownloadVibeStep) Run(ctx *Context) StepResult {
	// Check VIBE_REPO_PATH env var first (user override)
	repoPath := os.Getenv("VIBE_REPO_PATH")

	if repoPath == "" {
		// Check common local locations — the CLI is built from source, so the
		// repo must already be on this machine somewhere.
		candidates := []string{
			filepath.Join(ctx.HomeDir, "claude-vibe"),
			filepath.Join(ctx.HomeDir, "vibe"),
			filepath.Join(ctx.HomeDir, ".vibe", "src", "claude-vibe"),
		}
		for _, candidate := range candidates {
			if ctx.DirExists(filepath.Join(candidate, ".claude-plugin")) {
				repoPath = candidate
				break
			}
		}
	}

	if repoPath == "" {
		return FailureWithHint(
			"Vibe repository not found",
			nil,
			"Set VIBE_REPO_PATH to your repo location before running vibe install:\n"+
				"  export VIBE_REPO_PATH=~/claude-vibe\n"+
				"Or clone it first, then re-run vibe install.",
		)
	}

	// Validate the repo looks correct
	if !ctx.DirExists(filepath.Join(repoPath, ".claude-plugin")) {
		return FailureWithHint(
			"Directory found but missing .claude-plugin — is this the right repo?",
			nil,
			"Set VIBE_REPO_PATH to the correct repo root (should contain .claude-plugin/).",
		)
	}

	ctx.VibeDir = repoPath
	ctx.Log("Using local vibe repo at %s", repoPath)

	return SuccessWithDetails(
		"Vibe repository located",
		"  Location: "+repoPath,
	)
}
