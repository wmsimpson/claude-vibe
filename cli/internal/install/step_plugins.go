package install

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wmsimpson/claude-vibe/cli/internal/config"
)

// PluginsInstallStep installs all vibe plugins by copying files directly to the
// Claude plugin cache and updating installed_plugins.json. This bypasses
// `claude plugin install` entirely — no CLI invocations, no network calls.
type PluginsInstallStep struct{}

func (s *PluginsInstallStep) ID() string                      { return "plugins_install" }
func (s *PluginsInstallStep) Name() string                    { return "Install Plugins" }
func (s *PluginsInstallStep) Description() string             { return "Install all vibe plugins" }
func (s *PluginsInstallStep) ActiveForm() string              { return "Installing plugins" }
func (s *PluginsInstallStep) CanSkip(opts *Options) bool      { return opts.SkipPlugins }
func (s *PluginsInstallStep) NeedsSudo() bool                 { return false }

// DefaultPlugins is the list of plugins installed by default.
var DefaultPlugins = []string{
	"databricks-tools",
	"google-tools",
	"specialized-agents",
	"vibe-setup",
	"mcp-servers",
	"jira-tools",
	"workflows",
}

// AllPluginsToInstall returns the full list of plugins to install:
// default plugins merged with any extra plugins saved in ~/.vibe/config.yaml.
func AllPluginsToInstall() []string {
	cfg, err := config.Load()
	if err != nil {
		return DefaultPlugins
	}
	return mergePlugins(DefaultPlugins, cfg.Settings.ExtraPlugins)
}

// mergePlugins combines defaults and extras, removing duplicates.
func mergePlugins(defaults, extras []string) []string {
	if len(extras) == 0 {
		return defaults
	}

	seen := make(map[string]bool, len(defaults))
	for _, p := range defaults {
		seen[p] = true
	}

	merged := make([]string, len(defaults))
	copy(merged, defaults)

	for _, p := range extras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !seen[p] {
			seen[p] = true
			merged = append(merged, p)
		}
	}

	return merged
}

func (s *PluginsInstallStep) Check(ctx *Context) (bool, error) {
	// Always reinstall to get latest versions
	return false, nil
}

func (s *PluginsInstallStep) Run(ctx *Context) StepResult {
	if ctx.VibeDir == "" {
		return Failure("Vibe directory not set (download step failed?)", nil)
	}

	cfg, _ := config.Load()
	savedExtras := cfg.Settings.ExtraPlugins
	cliExtras := ctx.Options.ExtraPlugins
	allExtras := mergePlugins(savedExtras, cliExtras)
	plugins := mergePlugins(DefaultPlugins, allExtras)

	// Persist any new CLI extras
	if len(cliExtras) > 0 {
		defaultSet := make(map[string]bool, len(DefaultPlugins))
		for _, p := range DefaultPlugins {
			defaultSet[p] = true
		}
		seenExtra := make(map[string]bool)
		var newExtras []string
		for _, p := range append(savedExtras, cliExtras...) {
			p = strings.TrimSpace(p)
			if p == "" || defaultSet[p] || seenExtra[p] {
				continue
			}
			seenExtra[p] = true
			newExtras = append(newExtras, p)
		}
		cfg.Settings.ExtraPlugins = newExtras
		cfg.Save()
	}

	// Paths
	cacheRoot := filepath.Join(ctx.ClaudeDir, "plugins", "cache", "claude-vibe")
	installedPluginsPath := filepath.Join(ctx.ClaudeDir, "plugins", "installed_plugins.json")
	settingsFile := filepath.Join(ctx.ClaudeDir, "settings.json")

	// Read plugin versions from marketplace.json
	versions := readPluginVersions(ctx.VibeDir)

	// Load or create installed_plugins.json
	installedData := map[string]interface{}{
		"version": 2,
		"plugins": map[string]interface{}{},
	}
	if data, err := os.ReadFile(installedPluginsPath); err == nil {
		json.Unmarshal(data, &installedData)
	}
	pluginsMap, ok := installedData["plugins"].(map[string]interface{})
	if !ok {
		pluginsMap = make(map[string]interface{})
		installedData["plugins"] = pluginsMap
	}

	// Load or create settings.json
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsFile); err == nil {
		json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]interface{})
	}
	enabledPlugins, ok := settings["enabledPlugins"].(map[string]interface{})
	if !ok {
		enabledPlugins = make(map[string]interface{})
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var installed []string
	var skipped []string
	var failed []string
	var details strings.Builder

	for _, plugin := range plugins {
		srcDir := filepath.Join(ctx.VibeDir, "plugins", plugin)
		if !ctx.DirExists(srcDir) {
			skipped = append(skipped, plugin)
			ctx.Log("Skipping %s: not found at %s", plugin, srcDir)
			continue
		}

		version := versions[plugin]
		if version == "" {
			version = "1.0.0"
		}

		// Copy plugin directory to cache
		destDir := filepath.Join(cacheRoot, plugin, version)
		os.RemoveAll(destDir) // remove stale version
		if err := copyDir(srcDir, destDir); err != nil {
			ctx.Log("Failed to copy %s: %v", plugin, err)
			failed = append(failed, plugin)
			details.WriteString(fmt.Sprintf("  %s: FAILED (%v)\n", plugin, err))
			continue
		}

		// Register in installed_plugins.json
		fullName := plugin + "@claude-vibe"
		pluginsMap[fullName] = []interface{}{
			map[string]interface{}{
				"scope":       "user",
				"installPath": destDir,
				"version":     version,
				"installedAt": now,
				"lastUpdated": now,
			},
		}

		// Enable in settings.json
		enabledPlugins[fullName] = true

		installed = append(installed, plugin)
		details.WriteString(fmt.Sprintf("  %s@%s: OK\n", plugin, version))
	}

	// Write installed_plugins.json
	if err := os.MkdirAll(filepath.Dir(installedPluginsPath), 0755); err == nil {
		if data, err := json.MarshalIndent(installedData, "", "  "); err == nil {
			os.WriteFile(installedPluginsPath, data, 0644)
		}
	}

	// Write settings.json
	settings["enabledPlugins"] = enabledPlugins
	if data, err := json.MarshalIndent(settings, "", "  "); err == nil {
		os.WriteFile(settingsFile, data, 0644)
	}

	if len(skipped) > 0 {
		details.WriteString(fmt.Sprintf("  Skipped (not found): %s\n", strings.Join(skipped, ", ")))
	}

	if len(failed) > 0 {
		return FailureWithHint(
			fmt.Sprintf("Some plugins failed to install: %s", strings.Join(failed, ", ")),
			nil,
			"Check that plugins/ directories exist in VIBE_REPO_PATH",
		)
	}

	msg := "All plugins installed"
	if len(installed) == 0 {
		msg = "No plugins installed"
	}
	return SuccessWithDetails(msg, details.String())
}

// readPluginVersions reads plugin versions from the marketplace.json in the repo root.
func readPluginVersions(vibeDir string) map[string]string {
	versions := make(map[string]string)
	data, err := os.ReadFile(filepath.Join(vibeDir, ".claude-plugin", "marketplace.json"))
	if err != nil {
		return versions
	}
	var marketplace struct {
		Plugins []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(data, &marketplace); err != nil {
		return versions
	}
	for _, p := range marketplace.Plugins {
		if p.Version != "" {
			versions[p.Name] = p.Version
		}
	}
	return versions
}

// copyDir recursively copies a directory tree from src to dst.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
