package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CursorTarget syncs MCP servers and skills to Cursor.
type CursorTarget struct {
	homeDir     string
	manifestDir string // override for tests; empty = use default ~/.vibe/sync/
}

// NewCursorTarget creates a new CursorTarget using the current user's home directory.
func NewCursorTarget() *CursorTarget {
	home, _ := os.UserHomeDir()
	return &CursorTarget{homeDir: home}
}

// NewCursorTargetWithHome creates a CursorTarget with a custom home directory (for testing).
func NewCursorTargetWithHome(home string) *CursorTarget {
	mDir := filepath.Join(home, ".vibe", "sync")
	return &CursorTarget{homeDir: home, manifestDir: mDir}
}

func (c *CursorTarget) configDir() string {
	return filepath.Join(c.homeDir, ".cursor")
}

func (c *CursorTarget) mcpPath() string {
	return filepath.Join(c.configDir(), "mcp.json")
}

func (c *CursorTarget) skillsDir() string {
	return filepath.Join(c.configDir(), "skills")
}

// Name returns the target identifier.
func (c *CursorTarget) Name() string { return "cursor" }

// IsInstalled returns true if ~/.cursor/ exists.
func (c *CursorTarget) IsInstalled() bool {
	info, err := os.Stat(c.configDir())
	return err == nil && info.IsDir()
}

// SyncMCP writes MCP servers to ~/.cursor/mcp.json.
func (c *CursorTarget) SyncMCP(servers []MCPServerConfig, opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{Target: c.Name()}

	if len(servers) == 0 {
		result.Details = append(result.Details, "No MCP servers to sync")
		return result, nil
	}

	// Load existing config
	existing := make(map[string]interface{})
	data, err := os.ReadFile(c.mcpPath())
	if err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", c.mcpPath(), err)
		}
	}

	// Load manifest to know which servers we manage
	manifest, _ := LoadManifest(c.manifestDir, c.Name())

	// Get existing mcpServers or create new map
	mcpServers := make(map[string]interface{})
	if existingServers, ok := existing["mcpServers"].(map[string]interface{}); ok {
		// Preserve user-added servers (not in manifest)
		for name, cfg := range existingServers {
			if !manifest.IsManagedServer(name) {
				mcpServers[name] = cfg
			}
		}
	}

	// Add vibe-managed servers
	var managedNames []string
	for _, s := range servers {
		serverCfg := map[string]interface{}{
			"command": s.Command,
			"args":    s.Args,
		}
		if len(s.Env) > 0 {
			serverCfg["env"] = s.Env
		}
		mcpServers[s.Name] = serverCfg
		managedNames = append(managedNames, s.Name)
		result.ItemsSynced++
		result.Details = append(result.Details, fmt.Sprintf("  MCP server: %s", s.Name))
	}

	existing["mcpServers"] = mcpServers

	if opts.DryRun {
		result.Details = append(result.Details, fmt.Sprintf("Would write %s", c.mcpPath()))
		return result, nil
	}

	// Write config
	output, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(c.mcpPath()), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	if err := os.WriteFile(c.mcpPath(), output, 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", c.mcpPath(), err)
	}

	// Update manifest
	manifest.UpdateMCPServers(managedNames)
	manifest.Save(c.manifestDir, c.Name())

	return result, nil
}

// SyncSkills copies skill directories to ~/.cursor/skills/<name>/.
func (c *CursorTarget) SyncSkills(skills []SkillSource, opts SyncOptions) (*SyncResult, error) {
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

// Status returns the sync status for Cursor.
func (c *CursorTarget) Status(servers []MCPServerConfig, skills []SkillSource) (*SyncStatus, error) {
	status := &SyncStatus{Target: c.Name()}

	manifest, err := LoadManifest(c.manifestDir, c.Name())
	if err != nil {
		return status, nil
	}
	status.LastSynced = manifest.LastSynced

	// Check MCP status
	existing := make(map[string]interface{})
	data, err := os.ReadFile(c.mcpPath())
	if err == nil {
		json.Unmarshal(data, &existing)
	}

	existingServers := make(map[string]interface{})
	if servers, ok := existing["mcpServers"].(map[string]interface{}); ok {
		existingServers = servers
	}

	sourceServerNames := make(map[string]bool)
	for _, s := range servers {
		sourceServerNames[s.Name] = true
		if _, exists := existingServers[s.Name]; exists {
			status.MCPInSync = append(status.MCPInSync, s.Name)
		} else {
			status.MCPMissing = append(status.MCPMissing, s.Name)
		}
	}

	for name := range existingServers {
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

		// Compare checksums
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
