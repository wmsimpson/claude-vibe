// Package config provides configuration management for the vibe CLI.
// It handles loading and saving configuration from various sources including
// vibe config (~/.vibe/config.yaml), MCP servers, plugins, permissions, and profiles.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main vibe configuration stored in ~/.vibe/config.yaml
type Config struct {
	Version  int      `yaml:"version"`
	Settings Settings `yaml:"settings"`
}

// Settings contains user preferences and global settings
type Settings struct {
	Theme           string   `yaml:"theme"`
	DefaultAgent    string   `yaml:"default_agent"`
	AutoUpdateCheck bool     `yaml:"auto_update_check"`
	ExtraPlugins    []string `yaml:"extra_plugins,omitempty"`
	SyncTargets     []string `yaml:"sync_targets,omitempty"`
	AutoSync        bool     `yaml:"auto_sync,omitempty"`
}

// VibeDir returns the path to the vibe configuration directory (~/.vibe)
func VibeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vibe")
}

// VibeConfigPath returns the path to the vibe config file (~/.vibe/config.yaml)
func VibeConfigPath() string {
	return filepath.Join(VibeDir(), "config.yaml")
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Settings: Settings{
			Theme:           "default",
			DefaultAgent:    "claude",
			AutoUpdateCheck: true,
		},
	}
}

// Load reads the vibe configuration from ~/.vibe/config.yaml
// If the file doesn't exist, it returns the default configuration
func Load() (*Config, error) {
	configPath := VibeConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the configuration to ~/.vibe/config.yaml
// It creates the ~/.vibe directory if it doesn't exist
func (c *Config) Save() error {
	vibeDir := VibeDir()
	if err := os.MkdirAll(vibeDir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(VibeConfigPath(), data, 0644)
}
