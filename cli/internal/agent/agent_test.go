package agent

import (
	"testing"
)

// MockAgent is a test double for the Agent interface
type MockAgent struct {
	name        string
	installed   bool
	version     string
	versionErr  error
	launchErr   error
	paths       AgentPaths
	launchCalls []SessionOptions
}

func NewMockAgent(name string) *MockAgent {
	return &MockAgent{
		name:        name,
		installed:   true,
		version:     "1.0.0",
		launchCalls: make([]SessionOptions, 0),
		paths: AgentPaths{
			Settings:     "/mock/.agent/settings.json",
			Plugins:      "/mock/.agent/plugins/installed_plugins.json",
			MCPConfig:    "/mock/.config/mcp/config.json",
			GlobalConfig: "/mock/.agent.json",
		},
	}
}

func (m *MockAgent) Name() string {
	return m.name
}

func (m *MockAgent) IsInstalled() bool {
	return m.installed
}

func (m *MockAgent) Version() (string, error) {
	if m.versionErr != nil {
		return "", m.versionErr
	}
	return m.version, nil
}

func (m *MockAgent) LaunchSession(opts SessionOptions) error {
	m.launchCalls = append(m.launchCalls, opts)
	return m.launchErr
}

func (m *MockAgent) ConfigPaths() AgentPaths {
	return m.paths
}

// SetInstalled sets whether the mock agent is installed
func (m *MockAgent) SetInstalled(installed bool) *MockAgent {
	m.installed = installed
	return m
}

// SetVersion sets the version the mock agent returns
func (m *MockAgent) SetVersion(version string) *MockAgent {
	m.version = version
	return m
}

// SetVersionError sets an error to return from Version()
func (m *MockAgent) SetVersionError(err error) *MockAgent {
	m.versionErr = err
	return m
}

// SetLaunchError sets an error to return from LaunchSession()
func (m *MockAgent) SetLaunchError(err error) *MockAgent {
	m.launchErr = err
	return m
}

// SetPaths sets the config paths the mock agent returns
func (m *MockAgent) SetPaths(paths AgentPaths) *MockAgent {
	m.paths = paths
	return m
}

// LaunchCallCount returns the number of times LaunchSession was called
func (m *MockAgent) LaunchCallCount() int {
	return len(m.launchCalls)
}

// LastLaunchOptions returns the options from the last LaunchSession call
func (m *MockAgent) LastLaunchOptions() (SessionOptions, bool) {
	if len(m.launchCalls) == 0 {
		return SessionOptions{}, false
	}
	return m.launchCalls[len(m.launchCalls)-1], true
}

// TestAgentInterface verifies that MockAgent implements the Agent interface
func TestAgentInterface(t *testing.T) {
	var _ Agent = (*MockAgent)(nil)
}

// TestMockAgentDefaults verifies the default state of a new MockAgent
func TestMockAgentDefaults(t *testing.T) {
	mock := NewMockAgent("test-agent")

	if mock.Name() != "test-agent" {
		t.Errorf("expected name 'test-agent', got '%s'", mock.Name())
	}

	if !mock.IsInstalled() {
		t.Error("expected IsInstalled to be true by default")
	}

	version, err := mock.Version()
	if err != nil {
		t.Errorf("unexpected error from Version: %v", err)
	}
	if version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", version)
	}

	paths := mock.ConfigPaths()
	if paths.Settings == "" {
		t.Error("expected Settings path to be non-empty")
	}
}

// TestMockAgentSetters verifies the setter methods work correctly
func TestMockAgentSetters(t *testing.T) {
	mock := NewMockAgent("test-agent")

	// Test SetInstalled
	mock.SetInstalled(false)
	if mock.IsInstalled() {
		t.Error("expected IsInstalled to be false after SetInstalled(false)")
	}

	// Test SetVersion
	mock.SetVersion("2.0.0")
	version, _ := mock.Version()
	if version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got '%s'", version)
	}

	// Test SetVersionError
	testErr := &testError{msg: "version error"}
	mock.SetVersionError(testErr)
	_, err := mock.Version()
	if err != testErr {
		t.Errorf("expected error '%v', got '%v'", testErr, err)
	}

	// Test SetPaths
	customPaths := AgentPaths{
		Settings:     "/custom/settings.json",
		Plugins:      "/custom/plugins.json",
		MCPConfig:    "/custom/mcp.json",
		GlobalConfig: "/custom/global.json",
	}
	mock.SetPaths(customPaths)
	paths := mock.ConfigPaths()
	if paths.Settings != "/custom/settings.json" {
		t.Errorf("expected Settings '/custom/settings.json', got '%s'", paths.Settings)
	}
}

// TestMockAgentLaunchSession verifies LaunchSession tracking
func TestMockAgentLaunchSession(t *testing.T) {
	mock := NewMockAgent("test-agent")

	// Initially no calls
	if mock.LaunchCallCount() != 0 {
		t.Errorf("expected 0 launch calls, got %d", mock.LaunchCallCount())
	}

	_, ok := mock.LastLaunchOptions()
	if ok {
		t.Error("expected LastLaunchOptions to return false when no calls made")
	}

	// Make a call
	opts := SessionOptions{
		WorkDir: "/test/dir",
		Profile: "test-profile",
	}
	err := mock.LaunchSession(opts)
	if err != nil {
		t.Errorf("unexpected error from LaunchSession: %v", err)
	}

	if mock.LaunchCallCount() != 1 {
		t.Errorf("expected 1 launch call, got %d", mock.LaunchCallCount())
	}

	lastOpts, ok := mock.LastLaunchOptions()
	if !ok {
		t.Error("expected LastLaunchOptions to return true")
	}
	if lastOpts.WorkDir != "/test/dir" {
		t.Errorf("expected WorkDir '/test/dir', got '%s'", lastOpts.WorkDir)
	}
	if lastOpts.Profile != "test-profile" {
		t.Errorf("expected Profile 'test-profile', got '%s'", lastOpts.Profile)
	}

	// Test launch error
	launchErr := &testError{msg: "launch error"}
	mock.SetLaunchError(launchErr)
	err = mock.LaunchSession(opts)
	if err != launchErr {
		t.Errorf("expected error '%v', got '%v'", launchErr, err)
	}
}

// TestRegistry verifies the agent registry functions
func TestRegistry(t *testing.T) {
	// Clear registry before test
	clearRegistry()

	// Test Get on empty registry
	_, ok := Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for nonexistent agent")
	}

	// Test List on empty registry
	agents := List()
	if len(agents) != 0 {
		t.Errorf("expected empty list, got %d agents", len(agents))
	}

	// Test Default on empty registry
	defaultAgent := Default()
	if defaultAgent != nil {
		t.Error("expected Default to return nil on empty registry")
	}

	// Register an agent
	mockAgent := NewMockAgent("mock-agent")
	Register("mock-agent", mockAgent)

	// Test Get
	agent, ok := Get("mock-agent")
	if !ok {
		t.Error("expected Get to return true for registered agent")
	}
	if agent != mockAgent {
		t.Error("expected Get to return the registered agent")
	}

	// Test List
	agents = List()
	if len(agents) != 1 {
		t.Errorf("expected 1 agent in list, got %d", len(agents))
	}
	if agents[0] != "mock-agent" {
		t.Errorf("expected 'mock-agent' in list, got '%s'", agents[0])
	}

	// Register another agent
	mockAgent2 := NewMockAgent("mock-agent-2")
	Register("mock-agent-2", mockAgent2)

	agents = List()
	if len(agents) != 2 {
		t.Errorf("expected 2 agents in list, got %d", len(agents))
	}
}

// TestRegistryDefault verifies the default agent behavior
func TestRegistryDefault(t *testing.T) {
	// Clear registry before test
	clearRegistry()

	// Register claude agent - should become default
	claudeAgent := NewMockAgent("claude")
	Register("claude", claudeAgent)

	defaultAgent := Default()
	if defaultAgent != claudeAgent {
		t.Error("expected claude to be the default agent")
	}

	// Register another agent - default should still be claude
	otherAgent := NewMockAgent("other")
	Register("other", otherAgent)

	defaultAgent = Default()
	if defaultAgent != claudeAgent {
		t.Error("expected claude to remain the default agent")
	}
}

// TestRegistryDefaultFallback verifies fallback behavior when claude is not registered
func TestRegistryDefaultFallback(t *testing.T) {
	// Clear registry before test
	clearRegistry()

	// Register non-claude agent first
	mockAgent := NewMockAgent("not-claude")
	Register("not-claude", mockAgent)

	// Default should return the first registered when claude is not present
	defaultAgent := Default()
	if defaultAgent != mockAgent {
		t.Error("expected first registered agent to be default when claude not present")
	}
}

// TestSetDefault verifies setting a custom default agent
func TestSetDefault(t *testing.T) {
	// Clear registry before test
	clearRegistry()

	// Register multiple agents
	agent1 := NewMockAgent("agent-1")
	agent2 := NewMockAgent("agent-2")
	Register("agent-1", agent1)
	Register("agent-2", agent2)

	// Set custom default
	SetDefault("agent-2")

	defaultAgent := Default()
	if defaultAgent != agent2 {
		t.Error("expected agent-2 to be the default after SetDefault")
	}

	// Setting invalid default should be a no-op
	SetDefault("nonexistent")
	defaultAgent = Default()
	if defaultAgent != agent2 {
		t.Error("expected agent-2 to remain default after invalid SetDefault")
	}
}

// TestSessionOptions verifies SessionOptions struct
func TestSessionOptions(t *testing.T) {
	opts := SessionOptions{
		WorkDir: "/test/workdir",
		Profile: "test-profile",
	}

	if opts.WorkDir != "/test/workdir" {
		t.Errorf("expected WorkDir '/test/workdir', got '%s'", opts.WorkDir)
	}
	if opts.Profile != "test-profile" {
		t.Errorf("expected Profile 'test-profile', got '%s'", opts.Profile)
	}
}

// TestAgentPaths verifies AgentPaths struct
func TestAgentPaths(t *testing.T) {
	paths := AgentPaths{
		Settings:     "/home/user/.claude/settings.json",
		Plugins:      "/home/user/.claude/plugins/installed_plugins.json",
		MCPConfig:    "/home/user/.config/mcp/config.json",
		GlobalConfig: "/home/user/.claude.json",
	}

	if paths.Settings != "/home/user/.claude/settings.json" {
		t.Errorf("unexpected Settings path: %s", paths.Settings)
	}
	if paths.Plugins != "/home/user/.claude/plugins/installed_plugins.json" {
		t.Errorf("unexpected Plugins path: %s", paths.Plugins)
	}
	if paths.MCPConfig != "/home/user/.config/mcp/config.json" {
		t.Errorf("unexpected MCPConfig path: %s", paths.MCPConfig)
	}
	if paths.GlobalConfig != "/home/user/.claude.json" {
		t.Errorf("unexpected GlobalConfig path: %s", paths.GlobalConfig)
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
