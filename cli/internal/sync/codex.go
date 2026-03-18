package sync

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// CodexTarget syncs MCP servers and skills to OpenAI Codex CLI.
type CodexTarget struct {
	homeDir     string
	manifestDir string // override for tests; empty = use default ~/.vibe/sync/
}

// NewCodexTarget creates a new CodexTarget using the current user's home directory.
func NewCodexTarget() *CodexTarget {
	home, _ := os.UserHomeDir()
	return &CodexTarget{homeDir: home}
}

// NewCodexTargetWithHome creates a CodexTarget with a custom home directory (for testing).
func NewCodexTargetWithHome(home string) *CodexTarget {
	mDir := filepath.Join(home, ".vibe", "sync")
	return &CodexTarget{homeDir: home, manifestDir: mDir}
}

func (c *CodexTarget) configDir() string {
	return filepath.Join(c.homeDir, ".codex")
}

func (c *CodexTarget) configPath() string {
	return filepath.Join(c.configDir(), "config.toml")
}

func (c *CodexTarget) skillsDir() string {
	return filepath.Join(c.configDir(), "skills")
}

// Name returns the target identifier.
func (c *CodexTarget) Name() string { return "codex" }

// IsInstalled returns true if ~/.codex/ exists.
func (c *CodexTarget) IsInstalled() bool {
	info, err := os.Stat(c.configDir())
	return err == nil && info.IsDir()
}

// codexMCPServer represents an MCP server entry in Codex's config.toml.
type codexMCPServer struct {
	Command string            `toml:"command"`
	Args    []string          `toml:"args,omitempty"`
	Env     map[string]string `toml:"env,omitempty"`
}

// SyncMCP writes MCP servers to ~/.codex/config.toml under [mcp_servers.*] tables.
func (c *CodexTarget) SyncMCP(servers []MCPServerConfig, opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{Target: c.Name()}

	if len(servers) == 0 {
		result.Details = append(result.Details, "No MCP servers to sync")
		return result, nil
	}

	// Load existing config preserving non-MCP settings
	existing := make(map[string]interface{})
	data, err := os.ReadFile(c.configPath())
	if err == nil {
		if _, err := toml.Decode(string(data), &existing); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", c.configPath(), err)
		}
	}

	// Load manifest to know which servers we manage
	manifest, _ := LoadManifest(c.manifestDir, c.Name())

	// Build new mcp_servers map
	mcpServers := make(map[string]interface{})

	// Preserve user-added servers from existing config
	if existingMCP, ok := existing["mcp_servers"].(map[string]interface{}); ok {
		for name, cfg := range existingMCP {
			if !manifest.IsManagedServer(name) {
				mcpServers[name] = cfg
			}
		}
	}

	// Add vibe-managed servers
	var managedNames []string
	for _, s := range servers {
		entry := codexMCPServer{
			Command: s.Command,
			Args:    s.Args,
		}
		if len(s.Env) > 0 {
			entry.Env = s.Env
		}
		mcpServers[s.Name] = entry
		managedNames = append(managedNames, s.Name)
		result.ItemsSynced++
		result.Details = append(result.Details, fmt.Sprintf("  MCP server: %s", s.Name))
	}

	existing["mcp_servers"] = mcpServers

	if opts.DryRun {
		result.Details = append(result.Details, fmt.Sprintf("Would write %s", c.configPath()))
		return result, nil
	}

	// Write config
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(existing); err != nil {
		return nil, fmt.Errorf("failed to encode config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(c.configPath()), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	if err := os.WriteFile(c.configPath(), buf.Bytes(), 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", c.configPath(), err)
	}

	// Update manifest
	manifest.UpdateMCPServers(managedNames)
	manifest.Save(c.manifestDir, c.Name())

	return result, nil
}

// SyncSkills copies skill directories to ~/.codex/skills/<name>/.
func (c *CodexTarget) SyncSkills(skills []SkillSource, opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{Target: c.Name()}

	if len(skills) == 0 {
		result.Details = append(result.Details, "No skills to sync")
		return result, nil
	}

	manifest, _ := LoadManifest(c.manifestDir, c.Name())

	var managedNames []string
	checksums := make(map[string]string)

	for _, skill := range skills {
		targetName := skill.Name
		targetPath := filepath.Join(c.skillsDir(), targetName)

		if opts.DryRun {
			result.Details = append(result.Details, fmt.Sprintf("Would copy skill: %s -> %s", skill.Name, targetPath))
			result.ItemsSynced++
			managedNames = append(managedNames, targetName)
			continue
		}

		// Remove existing to handle renames/deletions
		os.RemoveAll(targetPath)

		if err := CopyDir(skill.SourcePath, targetPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to copy skill %s: %v", skill.Name, err))
			continue
		}

		checksum, _ := DirChecksum(targetPath)
		checksums[targetName] = checksum
		managedNames = append(managedNames, targetName)
		result.ItemsSynced++
		result.Details = append(result.Details, fmt.Sprintf("  Skill: %s", targetName))
	}

	if !opts.DryRun {
		manifest.UpdateSkills(managedNames, checksums)
		manifest.Save(c.manifestDir, c.Name())
	}

	return result, nil
}

// Status returns the sync status for Codex.
func (c *CodexTarget) Status(servers []MCPServerConfig, skills []SkillSource) (*SyncStatus, error) {
	status := &SyncStatus{Target: c.Name()}

	manifest, err := LoadManifest(c.manifestDir, c.Name())
	if err != nil {
		return status, nil
	}
	status.LastSynced = manifest.LastSynced

	// Check MCP status by reading config.toml
	existing := make(map[string]interface{})
	data, err := os.ReadFile(c.configPath())
	if err == nil {
		toml.Decode(string(data), &existing)
	}

	existingMCP := make(map[string]interface{})
	if mcp, ok := existing["mcp_servers"].(map[string]interface{}); ok {
		existingMCP = mcp
	}

	sourceServerNames := make(map[string]bool)
	for _, s := range servers {
		sourceServerNames[s.Name] = true
		if _, exists := existingMCP[s.Name]; exists {
			status.MCPInSync = append(status.MCPInSync, s.Name)
		} else {
			status.MCPMissing = append(status.MCPMissing, s.Name)
		}
	}

	for name := range existingMCP {
		if manifest.IsManagedServer(name) && !sourceServerNames[name] {
			status.MCPExtra = append(status.MCPExtra, name)
		}
	}

	// Check skill status
	for _, skill := range skills {
		targetName := skill.Name
		targetPath := filepath.Join(c.skillsDir(), targetName)

		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			status.SkillsMissing = append(status.SkillsMissing, targetName)
			continue
		}

		currentChecksum, _ := DirChecksum(targetPath)
		if savedChecksum, ok := manifest.Checksums[targetName]; ok && savedChecksum == currentChecksum {
			sourceChecksum, _ := DirChecksum(skill.SourcePath)
			if sourceChecksum == currentChecksum {
				status.SkillsInSync = append(status.SkillsInSync, targetName)
			} else {
				status.SkillsStale = append(status.SkillsStale, targetName)
			}
		} else {
			status.SkillsStale = append(status.SkillsStale, targetName)
		}
	}

	return status, nil
}
