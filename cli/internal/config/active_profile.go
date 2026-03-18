package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ActiveProfilePath returns the path to ~/.vibe/active-profile
func ActiveProfilePath() string {
	return filepath.Join(VibeDir(), "active-profile")
}

// GetActiveProfile returns the name of the currently active profile.
// Returns empty string if no active profile is set.
func GetActiveProfile() string {
	data, err := os.ReadFile(ActiveProfilePath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// SetActiveProfile persists the given profile name as the active profile.
// Creates the ~/.vibe directory if it doesn't exist.
func SetActiveProfile(name string) error {
	vibeDir := VibeDir()
	if err := os.MkdirAll(vibeDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(ActiveProfilePath(), []byte(name+"\n"), 0644)
}
