package install

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocalOwnershipStep fixes ownership of ~/.local if owned by root
type LocalOwnershipStep struct{}

func (s *LocalOwnershipStep) ID() string          { return "local_ownership" }
func (s *LocalOwnershipStep) Name() string        { return "Fix Ownership" }
func (s *LocalOwnershipStep) Description() string { return "Fix ~/.local directory ownership" }
func (s *LocalOwnershipStep) ActiveForm() string  { return "Fixing directory ownership" }
func (s *LocalOwnershipStep) CanSkip(opts *Options) bool { return false }
func (s *LocalOwnershipStep) NeedsSudo() bool     { return true }

func (s *LocalOwnershipStep) Check(ctx *Context) (bool, error) {
	localDir := filepath.Join(ctx.HomeDir, ".local")

	// If directory doesn't exist, nothing to fix
	if !ctx.DirExists(localDir) {
		return true, nil
	}

	// Check ownership
	owner, err := ctx.GetFileOwner(localDir)
	if err != nil {
		return false, err
	}

	// If not owned by root, nothing to fix
	return owner != "root", nil
}

func (s *LocalOwnershipStep) Run(ctx *Context) StepResult {
	localDir := filepath.Join(ctx.HomeDir, ".local")

	// Create directory if it doesn't exist
	if !ctx.DirExists(localDir) {
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return Failure("Failed to create ~/.local", err)
		}
		return Success("Created ~/.local directory")
	}

	// Check ownership
	owner, err := ctx.GetFileOwner(localDir)
	if err != nil {
		return Failure("Failed to check ~/.local ownership", err)
	}

	if owner == "root" {
		// Fix ownership
		currentUser := os.Getenv("USER")
		if currentUser == "" {
			currentUser = os.Getenv("LOGNAME")
		}

		_, err := ctx.RunSudoCommand("chown", "-R", fmt.Sprintf("%s:staff", currentUser), localDir)
		if err != nil {
			return FailureWithHint(
				"Failed to fix ~/.local ownership",
				err,
				fmt.Sprintf("Run manually: sudo chown -R %s:staff %s", currentUser, localDir),
			)
		}

		return Success("Fixed ~/.local ownership")
	}

	return Success("~/.local ownership is correct")
}
