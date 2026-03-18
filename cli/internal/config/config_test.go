package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if cfg.Settings.Theme != "default" {
		t.Errorf("expected theme 'default', got %s", cfg.Settings.Theme)
	}
	if cfg.Settings.DefaultAgent != "claude" {
		t.Errorf("expected default_agent 'claude', got %s", cfg.Settings.DefaultAgent)
	}
	if !cfg.Settings.AutoUpdateCheck {
		t.Error("expected auto_update_check to be true")
	}
}

func TestLoad_NonExistent(t *testing.T) {
	// Use a temp directory that doesn't exist
	tempDir := filepath.Join(os.TempDir(), "vibe-test-nonexistent")
	os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should return default config for non-existent file, got error: %v", err)
	}

	// Should return default config
	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	vibeDir := filepath.Join(tempDir, ".vibe")
	os.MkdirAll(vibeDir, 0755)

	configContent := `version: 1
settings:
  theme: dark
  default_agent: cursor
  auto_update_check: false
`
	configPath := filepath.Join(vibeDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if cfg.Settings.Theme != "dark" {
		t.Errorf("expected theme 'dark', got %s", cfg.Settings.Theme)
	}
	if cfg.Settings.DefaultAgent != "cursor" {
		t.Errorf("expected default_agent 'cursor', got %s", cfg.Settings.DefaultAgent)
	}
	if cfg.Settings.AutoUpdateCheck {
		t.Error("expected auto_update_check to be false")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	vibeDir := filepath.Join(tempDir, ".vibe")
	os.MkdirAll(vibeDir, 0755)

	// Write invalid YAML
	configPath := filepath.Join(vibeDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestConfig_Save(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{
		Version: 1,
		Settings: Settings{
			Theme:           "high-contrast",
			DefaultAgent:    "claude",
			AutoUpdateCheck: true,
		},
	}

	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify the file was created
	configPath := filepath.Join(tempDir, ".vibe", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Reload and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error after Save(): %v", err)
	}

	if loaded.Settings.Theme != "high-contrast" {
		t.Errorf("expected theme 'high-contrast', got %s", loaded.Settings.Theme)
	}
}

func TestConfig_SaveCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cfg := DefaultConfig()
	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify the .vibe directory was created
	vibeDir := filepath.Join(tempDir, ".vibe")
	info, err := os.Stat(vibeDir)
	if os.IsNotExist(err) {
		t.Error(".vibe directory was not created")
	}
	if !info.IsDir() {
		t.Error(".vibe should be a directory")
	}
}

func TestVibeConfigPath(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expected := filepath.Join(tempDir, ".vibe", "config.yaml")
	actual := VibeConfigPath()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestVibeDir(t *testing.T) {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	expected := filepath.Join(tempDir, ".vibe")
	actual := VibeDir()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}
