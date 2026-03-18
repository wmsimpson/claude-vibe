package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CleanupStep removes root-owned caches from prior sudo installs.
// During normal install: detects and fixes root-owned cache directories automatically.
// During --force-reinstall: also removes config files (.claude, .claude.json, etc.) for a clean slate.
type CleanupStep struct{}

func (s *CleanupStep) ID() string          { return "cleanup" }
func (s *CleanupStep) Name() string        { return "Clean Caches" }
func (s *CleanupStep) Description() string { return "Fix root-owned caches from prior sudo installs" }
func (s *CleanupStep) ActiveForm() string  { return "Checking for root-owned caches" }
func (s *CleanupStep) CanSkip(opts *Options) bool { return false }
func (s *CleanupStep) NeedsSudo() bool     { return true }

// Paths that are safe to remove (caches that get recreated) - checked during ALL installs
var cleanupCachePaths = []string{
	".pex",
	".local/state/mcp-servers",
}

// Paths only removed during --force-reinstall (config/state that shouldn't be removed normally)
var cleanupForceReinstallPaths = []string{
	".claude.json",
	".claude",
	"mcp",
}

// Paths that need ownership fixed (not removed) - checked during ALL installs
var cleanupOwnershipPaths = []string{
	".kube",
	".npm",
	".local",
}

// hasRootOwnedContents checks if a directory contains any root-owned files.
func hasRootOwnedContents(dir string) bool {
	cmd := exec.Command("find", dir, "-user", "root", "-maxdepth", "3", "-print", "-quit")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

func (s *CleanupStep) Check(ctx *Context) (bool, error) {
	// During force reinstall, always run
	if ctx.Options.ForceReinstall {
		return false, nil
	}

	// During normal install, check if any root-owned caches exist
	for _, p := range cleanupCachePaths {
		fullPath := filepath.Join(ctx.HomeDir, p)
		if !ctx.FileExists(fullPath) && !ctx.DirExists(fullPath) {
			continue
		}
		owner, _ := ctx.GetFileOwner(fullPath)
		if owner == "root" {
			return false, nil // needs to run
		}
		if ctx.DirExists(fullPath) && hasRootOwnedContents(fullPath) {
			return false, nil // needs to run
		}
	}

	for _, p := range cleanupOwnershipPaths {
		fullPath := filepath.Join(ctx.HomeDir, p)
		if !ctx.DirExists(fullPath) {
			continue
		}
		owner, _ := ctx.GetFileOwner(fullPath)
		if owner == "root" {
			return false, nil // needs to run
		}
		if hasRootOwnedContents(fullPath) {
			return false, nil // needs to run
		}
	}

	// No root-owned caches found
	return true, nil
}

func (s *CleanupStep) Run(ctx *Context) StepResult {
	// Determine which remove paths to check based on install mode
	removePaths := cleanupCachePaths
	if ctx.Options.ForceReinstall {
		// Force reinstall: also remove config/state paths
		removePaths = append(removePaths, cleanupForceReinstallPaths...)
	}

	// Collect paths that need sudo
	var pathsToRemove []string
	var pathsToChown []string

	// Check which paths exist and need cleanup
	for _, p := range removePaths {
		fullPath := filepath.Join(ctx.HomeDir, p)
		if !ctx.FileExists(fullPath) && !ctx.DirExists(fullPath) {
			continue
		}

		owner, _ := ctx.GetFileOwner(fullPath)
		if owner == "root" {
			pathsToRemove = append(pathsToRemove, fullPath)
		} else if ctx.DirExists(fullPath) && hasRootOwnedContents(fullPath) {
			// Directory is user-owned but contains root-owned files inside
			pathsToRemove = append(pathsToRemove, fullPath)
		} else if ctx.Options.ForceReinstall {
			// Force reinstall: remove even non-root-owned paths
			ctx.Log("Removing %s", fullPath)
			if err := os.RemoveAll(fullPath); err != nil {
				ctx.Log("Failed to remove %s without sudo: %v", fullPath, err)
				pathsToRemove = append(pathsToRemove, fullPath)
			}
		}
	}

	// Check ownership paths
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = os.Getenv("LOGNAME")
	}

	for _, p := range cleanupOwnershipPaths {
		fullPath := filepath.Join(ctx.HomeDir, p)
		if !ctx.DirExists(fullPath) {
			continue
		}
		owner, _ := ctx.GetFileOwner(fullPath)
		if owner == "root" {
			pathsToChown = append(pathsToChown, fullPath)
		} else if hasRootOwnedContents(fullPath) {
			pathsToChown = append(pathsToChown, fullPath)
		}
	}

	// If we need sudo for anything, return a sudo request
	if len(pathsToRemove) > 0 || len(pathsToChown) > 0 {
		// Build the sudo command
		var cmdParts []string

		if len(pathsToRemove) > 0 {
			cmdParts = append(cmdParts, fmt.Sprintf("rm -rf %s", strings.Join(pathsToRemove, " ")))
		}

		for _, p := range pathsToChown {
			cmdParts = append(cmdParts, fmt.Sprintf("chown -R %s:staff %s", currentUser, p))
		}

		sudoCmd := strings.Join(cmdParts, " && ")

		return NeedsSudo(
			"Root-owned files detected, sudo required to clean",
			"bash",
			[]string{"-c", sudoCmd},
		)
	}

	return Success("No root-owned cache issues found")
}
