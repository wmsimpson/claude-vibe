package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// mockTarget implements SyncTarget for testing the orchestrator.
type mockTarget struct {
	name      string
	installed bool

	mcpResult   *SyncResult
	mcpErr      error
	skillResult *SyncResult
	skillErr    error
	statusResult *SyncStatus
	statusErr   error

	mcpCalled   bool
	skillCalled bool
	statusCalled bool
}

func (m *mockTarget) Name() string       { return m.name }
func (m *mockTarget) IsInstalled() bool   { return m.installed }

func (m *mockTarget) SyncMCP(servers []MCPServerConfig, opts SyncOptions) (*SyncResult, error) {
	m.mcpCalled = true
	return m.mcpResult, m.mcpErr
}

func (m *mockTarget) SyncSkills(skills []SkillSource, opts SyncOptions) (*SyncResult, error) {
	m.skillCalled = true
	return m.skillResult, m.skillErr
}

func (m *mockTarget) Status(servers []MCPServerConfig, skills []SkillSource) (*SyncStatus, error) {
	m.statusCalled = true
	return m.statusResult, m.statusErr
}

func TestFilterTargets_Empty(t *testing.T) {
	targets := []SyncTarget{
		&mockTarget{name: "codex", installed: true},
		&mockTarget{name: "cursor", installed: true},
	}

	// Empty names returns all targets
	filtered := FilterTargets(targets, nil)
	if len(filtered) != 2 {
		t.Errorf("expected 2 targets, got %d", len(filtered))
	}
}

func TestFilterTargets_Specific(t *testing.T) {
	targets := []SyncTarget{
		&mockTarget{name: "codex", installed: true},
		&mockTarget{name: "cursor", installed: true},
	}

	filtered := FilterTargets(targets, []string{"codex"})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 target, got %d", len(filtered))
	}
	if filtered[0].Name() != "codex" {
		t.Errorf("expected codex, got %s", filtered[0].Name())
	}
}

func TestFilterTargets_CaseInsensitive(t *testing.T) {
	targets := []SyncTarget{
		&mockTarget{name: "codex", installed: true},
		&mockTarget{name: "cursor", installed: true},
	}

	filtered := FilterTargets(targets, []string{"CODEX"})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 target, got %d", len(filtered))
	}
	if filtered[0].Name() != "codex" {
		t.Errorf("expected codex, got %s", filtered[0].Name())
	}
}

func TestFilterTargets_NoMatch(t *testing.T) {
	targets := []SyncTarget{
		&mockTarget{name: "codex", installed: true},
	}

	filtered := FilterTargets(targets, []string{"windsurf"})
	if len(filtered) != 0 {
		t.Errorf("expected 0 targets, got %d", len(filtered))
	}
}

func TestRunSyncWithTargets_InstalledTarget(t *testing.T) {
	target := &mockTarget{
		name:      "test-agent",
		installed: true,
		mcpResult: &SyncResult{Target: "test-agent", ItemsSynced: 3},
		skillResult: &SyncResult{Target: "test-agent", ItemsSynced: 2},
	}

	servers := []MCPServerConfig{
		{Name: "s1", Command: "cmd1"},
	}
	skills := []SkillSource{
		{Name: "sk1", PluginName: "p1", SourcePath: "/tmp/sk1"},
	}

	results := RunSyncWithTargets([]SyncTarget{target}, servers, skills, SyncOptions{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Target != "test-agent" {
		t.Errorf("expected target 'test-agent', got %q", r.Target)
	}
	if r.ItemsSynced != 5 {
		t.Errorf("expected 5 items synced, got %d", r.ItemsSynced)
	}
	if !target.mcpCalled {
		t.Error("expected SyncMCP to be called")
	}
	if !target.skillCalled {
		t.Error("expected SyncSkills to be called")
	}
}

func TestRunSyncWithTargets_NotInstalled(t *testing.T) {
	target := &mockTarget{
		name:      "missing-agent",
		installed: false,
	}

	results := RunSyncWithTargets([]SyncTarget{target}, nil, nil, SyncOptions{})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.ItemsSkipped != 1 {
		t.Errorf("expected 1 item skipped, got %d", r.ItemsSkipped)
	}
	if len(r.Errors) != 1 {
		t.Errorf("expected 1 error message, got %d", len(r.Errors))
	}
	if target.mcpCalled || target.skillCalled {
		t.Error("sync methods should not be called on uninstalled target")
	}
}

func TestRunSyncWithTargets_MCPError(t *testing.T) {
	target := &mockTarget{
		name:        "error-agent",
		installed:   true,
		mcpErr:      fmt.Errorf("toml write failed"),
		skillResult: &SyncResult{Target: "error-agent", ItemsSynced: 1},
	}

	results := RunSyncWithTargets([]SyncTarget{target}, nil, nil, SyncOptions{})
	r := results[0]

	if len(r.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(r.Errors))
	}
	// Skills should still be synced even if MCP fails
	if r.ItemsSynced != 1 {
		t.Errorf("expected 1 item synced from skills, got %d", r.ItemsSynced)
	}
}

func TestRunSyncWithTargets_SkillError(t *testing.T) {
	target := &mockTarget{
		name:      "error-agent",
		installed: true,
		mcpResult: &SyncResult{Target: "error-agent", ItemsSynced: 2},
		skillErr:  fmt.Errorf("copy failed"),
	}

	results := RunSyncWithTargets([]SyncTarget{target}, nil, nil, SyncOptions{})
	r := results[0]

	if len(r.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(r.Errors))
	}
	// MCP should still count even if skills fail
	if r.ItemsSynced != 2 {
		t.Errorf("expected 2 items synced from MCP, got %d", r.ItemsSynced)
	}
}

func TestRunSyncWithTargets_MultipleTargets(t *testing.T) {
	t1 := &mockTarget{
		name:        "codex",
		installed:   true,
		mcpResult:   &SyncResult{Target: "codex", ItemsSynced: 3},
		skillResult: &SyncResult{Target: "codex", ItemsSynced: 2},
	}
	t2 := &mockTarget{
		name:      "cursor",
		installed: false,
	}
	t3 := &mockTarget{
		name:        "other",
		installed:   true,
		mcpResult:   &SyncResult{Target: "other", ItemsSynced: 1},
		skillResult: &SyncResult{Target: "other", ItemsSynced: 0},
	}

	results := RunSyncWithTargets([]SyncTarget{t1, t2, t3}, nil, nil, SyncOptions{})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// codex: synced
	if results[0].ItemsSynced != 5 {
		t.Errorf("codex: expected 5 synced, got %d", results[0].ItemsSynced)
	}
	// cursor: skipped
	if results[1].ItemsSkipped != 1 {
		t.Errorf("cursor: expected 1 skipped, got %d", results[1].ItemsSkipped)
	}
	// other: synced
	if results[2].ItemsSynced != 1 {
		t.Errorf("other: expected 1 synced, got %d", results[2].ItemsSynced)
	}
}

func TestRunStatusWithTargets_Installed(t *testing.T) {
	target := &mockTarget{
		name:      "test-agent",
		installed: true,
		statusResult: &SyncStatus{
			Target:     "test-agent",
			MCPInSync:  []string{"s1", "s2"},
			MCPMissing: []string{"s3"},
		},
	}

	statuses := RunStatusWithTargets([]SyncTarget{target}, nil, nil)
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}

	s := statuses[0]
	if len(s.MCPInSync) != 2 {
		t.Errorf("expected 2 in-sync, got %d", len(s.MCPInSync))
	}
	if len(s.MCPMissing) != 1 {
		t.Errorf("expected 1 missing, got %d", len(s.MCPMissing))
	}
}

func TestRunStatusWithTargets_NotInstalled(t *testing.T) {
	target := &mockTarget{
		name:      "missing",
		installed: false,
	}

	statuses := RunStatusWithTargets([]SyncTarget{target}, nil, nil)
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Target != "missing" {
		t.Errorf("expected target 'missing', got %q", statuses[0].Target)
	}
	if target.statusCalled {
		t.Error("Status should not be called on uninstalled target")
	}
}

func TestCopyDir(t *testing.T) {
	// Create source directory with files
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dest")

	writeFile(t, filepath.Join(srcDir, "file1.txt"), "hello")
	mkdirAll(t, filepath.Join(srcDir, "subdir"))
	writeFile(t, filepath.Join(srcDir, "subdir", "file2.txt"), "world")

	if err := CopyDir(srcDir, dstDir); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	// Verify files were copied
	assertFileContains(t, filepath.Join(dstDir, "file1.txt"), "hello")
	assertFileContains(t, filepath.Join(dstDir, "subdir", "file2.txt"), "world")
}

// helpers

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to mkdir %s: %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, expected string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if string(data) != expected {
		t.Errorf("file %s: expected %q, got %q", path, expected, string(data))
	}
}
