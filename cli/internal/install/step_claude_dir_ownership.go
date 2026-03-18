package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ClaudeDirOwnershipStep fixes ownership of ~/.claude if owned by root.
// This prevents EACCES errors when Claude Code tries to write to
// ~/.claude/plugins/known_marketplaces.json during marketplace setup.
type ClaudeDirOwnershipStep struct{}

func (s *ClaudeDirOwnershipStep) ID() string          { return "claude_dir_ownership" }
func (s *ClaudeDirOwnershipStep) Name() string        { return "Fix Claude Dir Permissions" }
func (s *ClaudeDirOwnershipStep) Description() string { return "Fix ~/.claude directory ownership" }
func (s *ClaudeDirOwnershipStep) ActiveForm() string  { return "Fixing ~/.claude permissions" }
func (s *ClaudeDirOwnershipStep) CanSkip(opts *Options) bool { return false }
func (s *ClaudeDirOwnershipStep) NeedsSudo() bool     { return true }

func (s *ClaudeDirOwnershipStep) Check(ctx *Context) (bool, error) {
	claudeDir := ctx.ClaudeDir

	// If directory doesn't exist, nothing to fix
	if !ctx.DirExists(claudeDir) {
		return true, nil
	}

	// Check if the directory itself is owned by root
	owner, err := ctx.GetFileOwner(claudeDir)
	if err != nil {
		return false, err
	}
	if owner == "root" {
		return false, nil
	}

	// Check if ~/.claude/plugins/ exists and is owned by root
	pluginsDir := filepath.Join(claudeDir, "plugins")
	if ctx.DirExists(pluginsDir) {
		pluginsOwner, err := ctx.GetFileOwner(pluginsDir)
		if err != nil {
			return false, err
		}
		if pluginsOwner == "root" {
			return false, nil
		}
	}

	// Check for any root-owned files inside ~/.claude (max depth 3 for speed)
	cmd := exec.Command("find", claudeDir, "-user", "root", "-maxdepth", "3", "-print", "-quit")
	output, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return false, nil
	}

	return true, nil
}

func (s *ClaudeDirOwnershipStep) Run(ctx *Context) StepResult {
	claudeDir := ctx.ClaudeDir

	// Create directory if it doesn't exist
	if !ctx.DirExists(claudeDir) {
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			return Failure("Failed to create ~/.claude", err)
		}
		return Success("Created ~/.claude directory")
	}

	// Check if anything is owned by root
	needsFix := false

	// Check directory itself
	owner, err := ctx.GetFileOwner(claudeDir)
	if err != nil {
		return Failure("Failed to check ~/.claude ownership", err)
	}
	if owner == "root" {
		needsFix = true
	}

	// Check plugins dir
	if !needsFix {
		pluginsDir := filepath.Join(claudeDir, "plugins")
		if ctx.DirExists(pluginsDir) {
			pluginsOwner, err := ctx.GetFileOwner(pluginsDir)
			if err == nil && pluginsOwner == "root" {
				needsFix = true
			}
		}
	}

	// Check for any root-owned files inside
	if !needsFix {
		cmd := exec.Command("find", claudeDir, "-user", "root", "-maxdepth", "3", "-print", "-quit")
		output, err := cmd.Output()
		if err == nil && len(strings.TrimSpace(string(output))) > 0 {
			needsFix = true
		}
	}

	if !needsFix {
		return Success("~/.claude ownership is correct")
	}

	// Fix ownership
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = os.Getenv("LOGNAME")
	}

	_, err = ctx.RunSudoCommand("chown", "-R", fmt.Sprintf("%s:staff", currentUser), claudeDir)
	if err != nil {
		return FailureWithHint(
			"Failed to fix ~/.claude ownership",
			err,
			fmt.Sprintf("Run manually: sudo chown -R %s:staff %s", currentUser, claudeDir),
		)
	}

	return Success("Fixed ~/.claude ownership")
}
