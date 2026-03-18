// Package agent provides an abstraction layer for coding agents (Claude Code, Cursor, etc.)
// This allows vibe-cli to support multiple agents while sharing common functionality.
package agent

import (
	"sort"
	"sync"
)

// Agent represents a coding agent (Claude Code, Cursor, etc.)
type Agent interface {
	// Name returns the agent identifier (e.g., "claude", "cursor")
	Name() string

	// IsInstalled checks if the agent is available on the system
	IsInstalled() bool

	// Version returns the installed version of the agent
	Version() (string, error)

	// LaunchSession starts an interactive session with the agent
	LaunchSession(opts SessionOptions) error

	// ConfigPaths returns paths to agent configuration files
	ConfigPaths() AgentPaths
}

// SessionOptions configures an agent session
type SessionOptions struct {
	// WorkDir is the working directory for the session
	WorkDir string

	// Profile is the name of the profile to apply (empty for default)
	Profile string
}

// AgentPaths contains paths to agent configuration files
type AgentPaths struct {
	// Settings is the path to the settings file (e.g., ~/.claude/settings.json)
	Settings string

	// Plugins is the path to the installed plugins file (e.g., ~/.claude/plugins/installed_plugins.json)
	Plugins string

	// MCPConfig is the path to the MCP configuration file (e.g., ~/.config/mcp/config.json)
	MCPConfig string

	// GlobalConfig is the path to the global configuration file (e.g., ~/.claude.json)
	GlobalConfig string
}

// registry holds registered agent implementations
var (
	registryMu      sync.RWMutex
	registry        = make(map[string]Agent)
	registryOrder   []string // maintains insertion order
	defaultAgentKey string
)

// Register adds an agent implementation to the registry.
// The first registered agent becomes the default unless "claude" is registered.
func Register(name string, agent Agent) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[name]; !exists {
		registryOrder = append(registryOrder, name)
	}
	registry[name] = agent

	// If this is claude, make it the default
	if name == "claude" {
		defaultAgentKey = "claude"
	}
}

// Get retrieves an agent by name from the registry.
// Returns the agent and true if found, nil and false otherwise.
func Get(name string) (Agent, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	agent, ok := registry[name]
	return agent, ok
}

// Default returns the default agent.
// Returns the agent set via SetDefault, or "claude" if registered,
// or the first registered agent, or nil if the registry is empty.
func Default() Agent {
	registryMu.RLock()
	defer registryMu.RUnlock()

	// If a default was explicitly set and exists, return it
	if defaultAgentKey != "" {
		if agent, ok := registry[defaultAgentKey]; ok {
			return agent
		}
	}

	// Try to return claude
	if agent, ok := registry["claude"]; ok {
		return agent
	}

	// Return first registered agent
	if len(registryOrder) > 0 {
		return registry[registryOrder[0]]
	}

	return nil
}

// SetDefault sets the default agent by name.
// If the name is not in the registry, this is a no-op.
func SetDefault(name string) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, ok := registry[name]; ok {
		defaultAgentKey = name
	}
}

// List returns the names of all registered agents in sorted order.
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// clearRegistry removes all agents from the registry.
// This is intended for testing purposes only.
func clearRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()

	registry = make(map[string]Agent)
	registryOrder = nil
	defaultAgentKey = ""
}
