package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest_NotExist(t *testing.T) {
	dir := t.TempDir()

	m, err := LoadManifest(dir, "codex")
	if err != nil {
		t.Fatalf("expected no error for missing manifest, got %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil manifest")
	}
	if len(m.MCPServers) != 0 {
		t.Errorf("expected empty MCPServers, got %v", m.MCPServers)
	}
	if len(m.Skills) != 0 {
		t.Errorf("expected empty Skills, got %v", m.Skills)
	}
	if m.Checksums == nil {
		t.Error("expected non-nil Checksums map")
	}
}

func TestManifest_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	m := &SyncManifest{
		MCPServers: []string{"server-a", "server-b"},
		Skills:     []string{"skill-1"},
		Checksums:  map[string]string{"skill-1": "abc123"},
	}

	if err := m.Save(dir, "codex"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists at <dir>/codex-manifest.json
	manifestPath := ManifestPathForTarget(dir, "codex")
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("manifest file not found: %v", err)
	}

	// Load and verify
	loaded, err := LoadManifest(dir, "codex")
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if len(loaded.MCPServers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(loaded.MCPServers))
	}
	if len(loaded.Skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(loaded.Skills))
	}
	if loaded.Checksums["skill-1"] != "abc123" {
		t.Errorf("expected checksum 'abc123', got %q", loaded.Checksums["skill-1"])
	}
}

func TestManifest_SeparatePerTarget(t *testing.T) {
	dir := t.TempDir()

	codexManifest := &SyncManifest{
		MCPServers: []string{"server-a"},
		Checksums:  make(map[string]string),
	}
	cursorManifest := &SyncManifest{
		MCPServers: []string{"server-b"},
		Checksums:  make(map[string]string),
	}

	codexManifest.Save(dir, "codex")
	cursorManifest.Save(dir, "cursor")

	// Load each and verify they're independent
	loaded1, _ := LoadManifest(dir, "codex")
	loaded2, _ := LoadManifest(dir, "cursor")

	if len(loaded1.MCPServers) != 1 || loaded1.MCPServers[0] != "server-a" {
		t.Errorf("codex manifest wrong: %v", loaded1.MCPServers)
	}
	if len(loaded2.MCPServers) != 1 || loaded2.MCPServers[0] != "server-b" {
		t.Errorf("cursor manifest wrong: %v", loaded2.MCPServers)
	}
}

func TestManifest_IsManagedServer(t *testing.T) {
	m := &SyncManifest{
		MCPServers: []string{"chrome-devtools", "slack"},
	}

	if !m.IsManagedServer("chrome-devtools") {
		t.Error("expected chrome-devtools to be managed")
	}
	if !m.IsManagedServer("slack") {
		t.Error("expected slack to be managed")
	}
	if m.IsManagedServer("user-custom") {
		t.Error("expected user-custom to NOT be managed")
	}
}

func TestManifest_IsManagedSkill(t *testing.T) {
	m := &SyncManifest{
		Skills: []string{"databricks-query", "google-docs"},
	}

	if !m.IsManagedSkill("databricks-query") {
		t.Error("expected databricks-query to be managed")
	}
	if m.IsManagedSkill("my-custom-skill") {
		t.Error("expected my-custom-skill to NOT be managed")
	}
}

func TestManifest_UpdateMCPServers(t *testing.T) {
	m := &SyncManifest{
		MCPServers: []string{"old-server"},
		Checksums:  make(map[string]string),
	}

	m.UpdateMCPServers([]string{"b-server", "a-server"})

	if len(m.MCPServers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(m.MCPServers))
	}
	// Should be sorted
	if m.MCPServers[0] != "a-server" {
		t.Errorf("expected sorted: first should be 'a-server', got %q", m.MCPServers[0])
	}
	if m.LastSynced.IsZero() {
		t.Error("expected LastSynced to be set")
	}
}

func TestManifest_UpdateSkills(t *testing.T) {
	m := &SyncManifest{
		Skills:    []string{"old-skill"},
		Checksums: map[string]string{"old-skill": "old-hash"},
	}

	m.UpdateSkills(
		[]string{"skill-b", "skill-a"},
		map[string]string{"skill-a": "hash-a", "skill-b": "hash-b"},
	)

	if len(m.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(m.Skills))
	}
	// Should be sorted
	if m.Skills[0] != "skill-a" {
		t.Errorf("expected sorted: first should be 'skill-a', got %q", m.Skills[0])
	}
	if m.Checksums["skill-a"] != "hash-a" {
		t.Errorf("expected checksum 'hash-a', got %q", m.Checksums["skill-a"])
	}
	// Old checksums should be replaced
	if _, ok := m.Checksums["old-skill"]; ok {
		t.Error("old-skill checksum should have been replaced")
	}
}

func TestDirChecksum_Deterministic(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(dir, "sub", "file2.txt"), []byte("world"), 0644)

	hash1, err := DirChecksum(dir)
	if err != nil {
		t.Fatalf("DirChecksum failed: %v", err)
	}
	if hash1 == "" {
		t.Fatal("expected non-empty checksum")
	}

	// Same content should produce same checksum
	hash2, err := DirChecksum(dir)
	if err != nil {
		t.Fatalf("DirChecksum failed: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("expected deterministic checksum: %q != %q", hash1, hash2)
	}
}

func TestDirChecksum_DiffersOnChange(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("original"), 0644)

	hash1, _ := DirChecksum(dir)

	// Modify file
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("modified"), 0644)

	hash2, _ := DirChecksum(dir)

	if hash1 == hash2 {
		t.Error("expected different checksums after file modification")
	}
}

func TestLoadManifest_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(ManifestPathForTarget(dir, "codex"), []byte("not json{"), 0644)

	_, err := LoadManifest(dir, "codex")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
