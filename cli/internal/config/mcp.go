package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

// MCPServer represents an MCP server configuration
type MCPServer struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
	Enabled bool              `json:"enabled"`
}

// MCPConfig manages MCP server configurations from multiple sources
type MCPConfig struct {
	servers         map[string]*mcpServerEntry
	disabledServers map[string]bool
	settingsData    map[string]interface{}
	claudeJSONData  map[string]interface{}
}

// mcpServerEntry tracks a server and its source
type mcpServerEntry struct {
	server MCPServer
	source string // "settings" or "claudeJSON" or "mcpConfig"
}

// MCPConfigPath returns the path to ~/.config/mcp/config.json
func MCPConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mcp", "config.json")
}

// ClaudeJSONPath returns the path to ~/.claude.json
func ClaudeJSONPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude.json")
}

// ProjectClaudeJSONPath returns the path to ./.claude.json (project scope)
func ProjectClaudeJSONPath(projectDir string) string {
	return filepath.Join(projectDir, ".claude.json")
}

// NewMCPConfig creates a new MCPConfig and loads servers from config files
func NewMCPConfig() *MCPConfig {
	mc := &MCPConfig{
		servers:         make(map[string]*mcpServerEntry),
		disabledServers: make(map[string]bool),
		settingsData:    make(map[string]interface{}),
		claudeJSONData:  make(map[string]interface{}),
	}
	mc.load()
	return mc
}

// load reads MCP server configurations from all sources
func (mc *MCPConfig) load() {
	// Load from ~/.claude/settings.json (primary source)
	mc.loadFromSettings()

	// Load from ~/.claude.json (legacy/alternative location)
	mc.loadFromClaudeJSON()

	// Load from ~/.config/mcp/config.json
	mc.loadFromMCPConfig()
}

func (mc *MCPConfig) loadFromSettings() {
	settingsPath := ClaudeSettingsPath()
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, &mc.settingsData); err != nil {
		return
	}

	// Parse mcpServers
	if servers, ok := mc.settingsData["mcpServers"].(map[string]interface{}); ok {
		for name, cfg := range servers {
			server := mc.parseServerConfig(name, cfg)
			mc.servers[name] = &mcpServerEntry{server: server, source: "settings"}
		}
	}
}

func (mc *MCPConfig) loadFromClaudeJSON() {
	claudeJSONPath := ClaudeJSONPath()
	data, err := os.ReadFile(claudeJSONPath)
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, &mc.claudeJSONData); err != nil {
		return
	}

	// Parse mcpServers
	if servers, ok := mc.claudeJSONData["mcpServers"].(map[string]interface{}); ok {
		for name, cfg := range servers {
			// Only add if not already present from settings
			if _, exists := mc.servers[name]; !exists {
				server := mc.parseServerConfig(name, cfg)
				mc.servers[name] = &mcpServerEntry{server: server, source: "claudeJSON"}
			}
		}
	}
}

func (mc *MCPConfig) loadFromMCPConfig() {
	mcpConfigPath := MCPConfigPath()
	data, err := os.ReadFile(mcpConfigPath)
	if err != nil {
		return
	}

	var mcpData map[string]interface{}
	if err := json.Unmarshal(data, &mcpData); err != nil {
		return
	}

	// Parse mcpServers (may be nested under "mcpServers" or at root level)
	var servers map[string]interface{}
	if s, ok := mcpData["mcpServers"].(map[string]interface{}); ok {
		servers = s
	} else {
		servers = mcpData
	}

	for name, cfg := range servers {
		// Skip non-server entries
		if _, ok := cfg.(map[string]interface{}); !ok {
			continue
		}
		// Only add if not already present
		if _, exists := mc.servers[name]; !exists {
			server := mc.parseServerConfig(name, cfg)
			mc.servers[name] = &mcpServerEntry{server: server, source: "mcpConfig"}
		}
	}
}

func (mc *MCPConfig) parseServerConfig(name string, cfg interface{}) MCPServer {
	server := MCPServer{
		Name:    name,
		Enabled: true,
	}

	cfgMap, ok := cfg.(map[string]interface{})
	if !ok {
		return server
	}

	if cmd, ok := cfgMap["command"].(string); ok {
		server.Command = cmd
	}

	if args, ok := cfgMap["args"].([]interface{}); ok {
		for _, arg := range args {
			if s, ok := arg.(string); ok {
				server.Args = append(server.Args, s)
			}
		}
	}

	if env, ok := cfgMap["env"].(map[string]interface{}); ok {
		server.Env = make(map[string]string)
		for k, v := range env {
			if s, ok := v.(string); ok {
				server.Env[k] = s
			}
		}
	}

	// Check if disabled
	if mc.disabledServers[name] {
		server.Enabled = false
	}

	return server
}

// ListServers returns all configured MCP servers sorted by name
func (mc *MCPConfig) ListServers() []MCPServer {
	var servers []MCPServer
	for _, entry := range mc.servers {
		servers = append(servers, entry.server)
	}

	// Sort by name for consistent ordering
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Name < servers[j].Name
	})

	return servers
}

// SetEnabled enables or disables an MCP server
func (mc *MCPConfig) SetEnabled(name string, enabled bool) error {
	entry, exists := mc.servers[name]
	if !exists {
		return errors.New("MCP server not found: " + name)
	}

	entry.server.Enabled = enabled
	if enabled {
		delete(mc.disabledServers, name)
	} else {
		mc.disabledServers[name] = true
	}

	return nil
}

// Save writes the MCP configuration to the appropriate config files
func (mc *MCPConfig) Save() error {
	// Update settings.json
	if err := mc.saveToSettings(); err != nil {
		return err
	}

	return nil
}

func (mc *MCPConfig) saveToSettings() error {
	settingsPath := ClaudeSettingsPath()

	// Load existing settings or create new
	var settingsData map[string]interface{}
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		json.Unmarshal(data, &settingsData)
	}
	if settingsData == nil {
		settingsData = make(map[string]interface{})
	}

	// Build mcpServers map
	mcpServers := make(map[string]interface{})
	for name, entry := range mc.servers {
		if entry.server.Enabled {
			serverCfg := map[string]interface{}{
				"command": entry.server.Command,
				"args":    entry.server.Args,
			}
			if len(entry.server.Env) > 0 {
				serverCfg["env"] = entry.server.Env
			}
			mcpServers[name] = serverCfg
		}
	}

	settingsData["mcpServers"] = mcpServers

	// Write back
	output, err := json.MarshalIndent(settingsData, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(settingsPath, output, 0644)
}

// MCPServerWithScope represents an MCP server with scope information
type MCPServerWithScope struct {
	Server   MCPServer
	IsGlobal bool // true if from user scope (global)
}

// ProjectMCPConfig manages MCP server configurations for project scope
type ProjectMCPConfig struct {
	projectDir     string
	userServers    map[string]*MCPServer // servers from user scope
	projectServers map[string]*MCPServer // servers added at project scope
	claudeJSONData map[string]interface{}
	userMCPConfig  *MCPConfig // reference to user config
	dirty          bool
}

// NewProjectMCPConfig creates a new ProjectMCPConfig for the given project directory
func NewProjectMCPConfig(projectDir string) *ProjectMCPConfig {
	pmc := &ProjectMCPConfig{
		projectDir:     projectDir,
		userServers:    make(map[string]*MCPServer),
		projectServers: make(map[string]*MCPServer),
		claudeJSONData: make(map[string]interface{}),
		userMCPConfig:  NewMCPConfig(),
	}
	pmc.load()
	return pmc
}

// load reads MCP server configurations from user and project scopes
func (pmc *ProjectMCPConfig) load() {
	// Load user-scope servers
	for _, s := range pmc.userMCPConfig.ListServers() {
		server := s // copy
		pmc.userServers[server.Name] = &server
	}

	// Load project-scope settings from .claude.json
	pmc.loadProjectSettings()
}

func (pmc *ProjectMCPConfig) loadProjectSettings() {
	claudeJSONPath := ProjectClaudeJSONPath(pmc.projectDir)
	data, err := os.ReadFile(claudeJSONPath)
	if err != nil {
		return
	}

	if err := json.Unmarshal(data, &pmc.claudeJSONData); err != nil {
		return
	}

	// Parse mcpServers at project level
	if servers, ok := pmc.claudeJSONData["mcpServers"].(map[string]interface{}); ok {
		for name, cfg := range servers {
			server := pmc.parseServerConfig(name, cfg)
			// If not in user scope, add to project servers
			if _, exists := pmc.userServers[name]; !exists {
				pmc.projectServers[name] = &server
			}
		}
	}
}

func (pmc *ProjectMCPConfig) parseServerConfig(name string, cfg interface{}) MCPServer {
	server := MCPServer{
		Name:    name,
		Enabled: true,
	}

	cfgMap, ok := cfg.(map[string]interface{})
	if !ok {
		return server
	}

	if cmd, ok := cfgMap["command"].(string); ok {
		server.Command = cmd
	}

	if args, ok := cfgMap["args"].([]interface{}); ok {
		for _, arg := range args {
			if s, ok := arg.(string); ok {
				server.Args = append(server.Args, s)
			}
		}
	}

	if env, ok := cfgMap["env"].(map[string]interface{}); ok {
		server.Env = make(map[string]string)
		for k, v := range env {
			if s, ok := v.(string); ok {
				server.Env[k] = s
			}
		}
	}

	return server
}

// ListMerged returns all servers (user + project) with merged state
// User servers are marked as global and cannot be modified
func (pmc *ProjectMCPConfig) ListMerged() []MCPServerWithScope {
	var servers []MCPServerWithScope

	// Add user-scope servers (marked as global)
	for _, s := range pmc.userServers {
		servers = append(servers, MCPServerWithScope{
			Server:   *s,
			IsGlobal: true,
		})
	}

	// Add project-scope only servers
	for name, s := range pmc.projectServers {
		if _, exists := pmc.userServers[name]; !exists {
			servers = append(servers, MCPServerWithScope{
				Server:   *s,
				IsGlobal: false,
			})
		}
	}

	// Sort by name
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Server.Name < servers[j].Server.Name
	})

	return servers
}

// SetEnabled enables or disables an MCP server at project scope
func (pmc *ProjectMCPConfig) SetEnabled(name string, enabled bool) error {
	// Can only modify project-scope servers
	if _, exists := pmc.userServers[name]; exists {
		return errors.New("cannot modify global MCP server: " + name)
	}

	server, exists := pmc.projectServers[name]
	if !exists {
		return errors.New("MCP server not found: " + name)
	}

	server.Enabled = enabled
	pmc.dirty = true
	return nil
}

// IsDirty returns true if there are unsaved changes
func (pmc *ProjectMCPConfig) IsDirty() bool {
	return pmc.dirty
}

// Save writes the project-scope MCP configuration to .claude.json
func (pmc *ProjectMCPConfig) Save() error {
	claudeJSONPath := ProjectClaudeJSONPath(pmc.projectDir)

	// Load existing settings or create new
	var claudeJSONData map[string]interface{}
	data, err := os.ReadFile(claudeJSONPath)
	if err == nil {
		json.Unmarshal(data, &claudeJSONData)
	}
	if claudeJSONData == nil {
		claudeJSONData = make(map[string]interface{})
	}

	// Build mcpServers map from project servers only
	mcpServers := make(map[string]interface{})
	for name, server := range pmc.projectServers {
		if server.Enabled {
			serverCfg := map[string]interface{}{
				"command": server.Command,
				"args":    server.Args,
			}
			if len(server.Env) > 0 {
				serverCfg["env"] = server.Env
			}
			mcpServers[name] = serverCfg
		}
	}

	claudeJSONData["mcpServers"] = mcpServers

	// Write back
	output, err := json.MarshalIndent(claudeJSONData, "", "  ")
	if err != nil {
		return err
	}

	pmc.dirty = false
	return os.WriteFile(claudeJSONPath, output, 0644)
}
