// Package util provides common utility functions for the vibe CLI.
package util

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ to the user's home directory and environment variables.
// It handles paths like "~/.vibe" and "$HOME/.vibe".
func ExpandPath(path string) string {
	if path == "" {
		return ""
	}

	// Expand environment variables first
	path = os.ExpandEnv(path)

	// Expand tilde
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}

	return path
}

// VibeDir returns the path to the vibe configuration directory (~/.vibe).
func VibeDir() string {
	return ExpandPath("~/.vibe")
}

// MarketplaceDir returns the path to the marketplace directory (~/.vibe/marketplace).
func MarketplaceDir() string {
	return filepath.Join(VibeDir(), "marketplace")
}

// ProfilesDir returns the path to the profiles directory (~/.vibe/profiles).
func ProfilesDir() string {
	return filepath.Join(VibeDir(), "profiles")
}

// ConfigFile returns the path to the vibe config file (~/.vibe/config.yaml).
func ConfigFile() string {
	return filepath.Join(VibeDir(), "config.yaml")
}

// ClaudeDir returns the path to the Claude configuration directory (~/.claude).
func ClaudeDir() string {
	return ExpandPath("~/.claude")
}

// ClaudeSettings returns the path to Claude's settings file (~/.claude/settings.json).
func ClaudeSettings() string {
	return filepath.Join(ClaudeDir(), "settings.json")
}

// ClaudePlugins returns the path to Claude's installed plugins file (~/.claude/plugins/installed_plugins.json).
func ClaudePlugins() string {
	return filepath.Join(ClaudeDir(), "plugins", "installed_plugins.json")
}

// MCPConfig returns the path to the MCP configuration file (~/.config/mcp/config.json).
func MCPConfig() string {
	return ExpandPath("~/.config/mcp/config.json")
}

// ClaudeJSON returns the path to the Claude JSON file (~/.claude.json).
func ClaudeJSON() string {
	return ExpandPath("~/.claude.json")
}

// IsHomeDirectory checks if the given path is the user's home directory.
// It handles exact matches, trailing slashes, symlinks, and relative paths.
// An empty path or "." is resolved to CWD first.
func IsHomeDirectory(path string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Handle empty path or "." by resolving to absolute path
	if path == "" || path == "." {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return false
		}
	}

	// Expand ~ in path
	path = ExpandPath(path)

	// Clean both paths (removes trailing slashes, resolves . and ..)
	path = filepath.Clean(path)
	home = filepath.Clean(home)

	// Direct comparison first
	if path == home {
		return true
	}

	// Try resolving symlinks for both
	resolvedPath, err1 := filepath.EvalSymlinks(path)
	resolvedHome, err2 := filepath.EvalSymlinks(home)

	if err1 == nil && err2 == nil {
		return filepath.Clean(resolvedPath) == filepath.Clean(resolvedHome)
	}

	return false
}

// IsInHomeDirectory checks if the given path is inside (or is) the home directory.
// This is useful for detecting when operations might affect user-scoped settings.
func IsInHomeDirectory(path string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Handle empty path or "." by resolving to absolute path
	if path == "" || path == "." {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return false
		}
	}

	// Expand ~ in path
	path = ExpandPath(path)

	// Clean both paths
	path = filepath.Clean(path)
	home = filepath.Clean(home)

	// Check if path starts with home
	return strings.HasPrefix(path, home)
}
