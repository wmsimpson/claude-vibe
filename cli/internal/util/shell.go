package util

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DetectShellRC returns the path to the user's shell RC file.
// It checks for zsh first (checking $SHELL and file existence),
// then falls back to bash.
func DetectShellRC() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(home, ".bashrc")
	}

	// Check if user is using zsh
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return filepath.Join(home, ".zshrc")
	}

	// Check if .zshrc exists (might be using zsh without SHELL set properly)
	zshrc := filepath.Join(home, ".zshrc")
	if _, err := os.Stat(zshrc); err == nil {
		return zshrc
	}

	// Default to bash
	return filepath.Join(home, ".bashrc")
}

// HasEnvVar checks if an environment variable export exists in the given RC file.
// It looks for lines like: export VAR_NAME="value" or export VAR_NAME=value
// It ignores commented lines.
func HasEnvVar(rcFile, varName string) bool {
	file, err := os.Open(rcFile)
	if err != nil {
		return false
	}
	defer file.Close()

	// Pattern to match: export VAR_NAME= (not commented)
	pattern := regexp.MustCompile(`^\s*export\s+` + regexp.QuoteMeta(varName) + `\s*=`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip commented lines
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if pattern.MatchString(line) {
			return true
		}
	}

	return false
}

// EnsureEnvVar adds an environment variable export to the RC file if it doesn't exist.
// If the variable is already present, it does nothing.
// If the file doesn't exist, it creates it.
func EnsureEnvVar(rcFile, varName, value string) error {
	// Check if variable already exists
	if HasEnvVar(rcFile, varName) {
		return nil
	}

	// Create directory if needed
	dir := filepath.Dir(rcFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open file for appending (create if doesn't exist)
	file, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", rcFile, err)
	}
	defer file.Close()

	// Add a newline before if file is not empty
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", rcFile, err)
	}

	var prefix string
	if info.Size() > 0 {
		prefix = "\n"
	}

	// Write the export line
	exportLine := fmt.Sprintf("%sexport %s=%q\n", prefix, varName, value)
	if _, err := file.WriteString(exportLine); err != nil {
		return fmt.Errorf("failed to write to %s: %w", rcFile, err)
	}

	return nil
}
