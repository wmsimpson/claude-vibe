// Package sync provides functionality for syncing vibe's MCP server configurations
// and skills from Claude Code (source of truth) to other coding agents like
// OpenAI Codex CLI and Cursor.
package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wmsimpson/claude-vibe/cli/internal/config"
)

// SyncTarget represents a destination agent that can receive synced configs.
type SyncTarget interface {
	// Name returns the target identifier (e.g., "codex", "cursor")
	Name() string

	// IsInstalled returns true if the target's config directory exists
	IsInstalled() bool

	// SyncMCP writes MCP server configurations to the target's format
	SyncMCP(servers []MCPServerConfig, opts SyncOptions) (*SyncResult, error)

	// SyncSkills copies skill files to the target
	SyncSkills(skills []SkillSource, opts SyncOptions) (*SyncResult, error)

	// Status returns the current sync state compared to source
	Status(servers []MCPServerConfig, skills []SkillSource) (*SyncStatus, error)
}

// MCPServerConfig is a normalized MCP server config for syncing.
type MCPServerConfig struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
}

// SkillSource represents a skill directory to sync.
type SkillSource struct {
	Name       string // skill name (e.g., "databricks-query")
	PluginName string // parent plugin (e.g., "databricks-tools")
	SourcePath string // absolute path to skill directory in plugin cache
}

// SyncOptions controls sync behavior.
type SyncOptions struct {
	DryRun  bool
	Verbose bool
}

// SyncResult contains the outcome of a sync operation.
type SyncResult struct {
	Target       string
	ItemsSynced  int
	ItemsSkipped int
	Errors       []string
	Details      []string // verbose info about what was written
}

// SyncStatus tracks what is in sync vs stale for a target.
type SyncStatus struct {
	Target        string
	MCPInSync     []string
	MCPStale      []string
	MCPMissing    []string // in source but not in target
	MCPExtra      []string // in target (vibe-managed) but not in source
	SkillsInSync  []string
	SkillsStale   []string
	SkillsMissing []string
	LastSynced    time.Time
}

// AllTargets returns all available sync targets.
func AllTargets() []SyncTarget {
	return []SyncTarget{
		NewCodexTarget(),
		NewCursorTarget(),
	}
}

// FilterTargets returns only the targets matching the given names.
// If names is empty, returns all targets.
func FilterTargets(targets []SyncTarget, names []string) []SyncTarget {
	if len(names) == 0 {
		return targets
	}

	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[strings.ToLower(n)] = true
	}

	var filtered []SyncTarget
	for _, t := range targets {
		if nameSet[t.Name()] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// RunSync executes sync against the given targets.
// It loads MCP servers and skills from Claude Code config automatically.
func RunSync(targets []SyncTarget, opts SyncOptions) []*SyncResult {
	return RunSyncWithTargets(targets, LoadMCPServers(), LoadInstalledSkills(), opts)
}

// RunSyncWithTargets executes sync with explicitly provided servers and skills.
// This is useful for testing and for callers that already have the data loaded.
func RunSyncWithTargets(targets []SyncTarget, servers []MCPServerConfig, skills []SkillSource, opts SyncOptions) []*SyncResult {
	var results []*SyncResult
	for _, target := range targets {
		if !target.IsInstalled() {
			results = append(results, &SyncResult{
				Target:       target.Name(),
				ItemsSkipped: 1,
				Errors:       []string{fmt.Sprintf("%s config directory not found, skipping", target.Name())},
			})
			continue
		}

		result := &SyncResult{Target: target.Name()}

		// Sync MCP servers
		mcpResult, err := target.SyncMCP(servers, opts)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("MCP sync failed: %v", err))
		} else if mcpResult != nil {
			result.ItemsSynced += mcpResult.ItemsSynced
			result.ItemsSkipped += mcpResult.ItemsSkipped
			result.Errors = append(result.Errors, mcpResult.Errors...)
			result.Details = append(result.Details, mcpResult.Details...)
		}

		// Sync skills
		skillResult, err := target.SyncSkills(skills, opts)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Skill sync failed: %v", err))
		} else if skillResult != nil {
			result.ItemsSynced += skillResult.ItemsSynced
			result.ItemsSkipped += skillResult.ItemsSkipped
			result.Errors = append(result.Errors, skillResult.Errors...)
			result.Details = append(result.Details, skillResult.Details...)
		}

		results = append(results, result)
	}
	return results
}

// RunStatus checks sync status for the given targets.
// It loads MCP servers and skills from Claude Code config automatically.
func RunStatus(targets []SyncTarget) []*SyncStatus {
	return RunStatusWithTargets(targets, LoadMCPServers(), LoadInstalledSkills())
}

// RunStatusWithTargets checks sync status with explicitly provided servers and skills.
func RunStatusWithTargets(targets []SyncTarget, servers []MCPServerConfig, skills []SkillSource) []*SyncStatus {
	var statuses []*SyncStatus
	for _, target := range targets {
		if !target.IsInstalled() {
			statuses = append(statuses, &SyncStatus{
				Target: target.Name(),
			})
			continue
		}

		status, err := target.Status(servers, skills)
		if err != nil {
			statuses = append(statuses, &SyncStatus{
				Target: target.Name(),
			})
			continue
		}
		statuses = append(statuses, status)
	}
	return statuses
}

// LoadMCPServers loads MCP server configurations from Claude Code config.
func LoadMCPServers() []MCPServerConfig {
	mc := config.NewMCPConfig()
	servers := mc.ListServers()

	var result []MCPServerConfig
	for _, s := range servers {
		if !s.Enabled {
			continue
		}
		result = append(result, MCPServerConfig{
			Name:    s.Name,
			Command: s.Command,
			Args:    s.Args,
			Env:     s.Env,
		})
	}
	return result
}

// LoadInstalledSkills discovers skill directories from installed plugins.
func LoadInstalledSkills() []SkillSource {
	pc := config.NewPluginConfig()
	installed := pc.ListInstalled()

	var skills []SkillSource
	for _, plugin := range installed {
		if !plugin.Enabled || plugin.InstallPath == "" {
			continue
		}

		skillsDir := filepath.Join(plugin.InstallPath, "skills")
		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			continue // no skills directory
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillPath := filepath.Join(skillsDir, entry.Name())
			skillMD := filepath.Join(skillPath, "SKILL.md")
			if _, err := os.Stat(skillMD); err != nil {
				continue // no SKILL.md
			}

			skills = append(skills, SkillSource{
				Name:       entry.Name(),
				PluginName: plugin.Name,
				SourcePath: skillPath,
			})
		}
	}
	return skills
}

// CopyDir recursively copies a directory from src to dst.
func CopyDir(src, dst string) error {
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
			if err := CopyDir(srcPath, dstPath); err != nil {
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

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, srcInfo.Mode())
}
