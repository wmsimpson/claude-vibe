package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetActiveProfile_NoFile(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	name := GetActiveProfile()
	if name != "" {
		t.Errorf("expected empty string when no active profile, got %q", name)
	}
}

func TestSetActiveProfile(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	os.MkdirAll(filepath.Join(tempDir, ".vibe"), 0755)

	err := SetActiveProfile("personal")
	if err != nil {
		t.Fatalf("SetActiveProfile() error: %v", err)
	}

	name := GetActiveProfile()
	if name != "personal" {
		t.Errorf("expected 'personal', got %q", name)
	}
}

func TestSetActiveProfile_CreatesVibeDir(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	err := SetActiveProfile("test")
	if err != nil {
		t.Fatalf("SetActiveProfile() error: %v", err)
	}

	vibeDir := filepath.Join(tempDir, ".vibe")
	if _, err := os.Stat(vibeDir); os.IsNotExist(err) {
		t.Error(".vibe directory was not created")
	}
}

func TestSetActiveProfile_OverwritesPrevious(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	SetActiveProfile("first")
	SetActiveProfile("second")

	name := GetActiveProfile()
	if name != "second" {
		t.Errorf("expected 'second', got %q", name)
	}
}

func TestGetActiveProfile_TrimsWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	vibeDir := filepath.Join(tempDir, ".vibe")
	os.MkdirAll(vibeDir, 0755)
	os.WriteFile(filepath.Join(vibeDir, "active-profile"), []byte("personal\n"), 0644)

	name := GetActiveProfile()
	if name != "personal" {
		t.Errorf("expected 'personal', got %q", name)
	}
}
