package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// SyncManifest tracks what vibe has synced to a target.
type SyncManifest struct {
	LastSynced time.Time         `json:"last_synced"`
	MCPServers []string          `json:"mcp_servers"`
	Skills     []string          `json:"skills,omitempty"`
	Checksums  map[string]string `json:"checksums,omitempty"`
}

// manifestDir returns the directory where sync manifests are stored (~/.vibe/sync/).
func manifestDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vibe", "sync")
}

// manifestPath returns the path for a target's manifest file.
func manifestPath(targetName string) string {
	return filepath.Join(manifestDir(), targetName+"-manifest.json")
}

// ManifestPathForTarget returns the manifest path for a given target name.
// Exported for use in tests that need a custom base directory.
func ManifestPathForTarget(baseDir, targetName string) string {
	return filepath.Join(baseDir, targetName+"-manifest.json")
}

// LoadManifest reads a sync manifest for the given target.
// If manifestDir is empty, uses the default ~/.vibe/sync/ location.
func LoadManifest(dir string, targetName string) (*SyncManifest, error) {
	var path string
	if dir != "" {
		path = ManifestPathForTarget(dir, targetName)
	} else {
		path = manifestPath(targetName)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &SyncManifest{
				Checksums: make(map[string]string),
			}, nil
		}
		return nil, err
	}

	var m SyncManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	if m.Checksums == nil {
		m.Checksums = make(map[string]string)
	}
	return &m, nil
}

// Save writes the manifest for the given target.
// If manifestDir is empty, uses the default ~/.vibe/sync/ location.
func (m *SyncManifest) Save(dir string, targetName string) error {
	var path string
	if dir != "" {
		path = ManifestPathForTarget(dir, targetName)
	} else {
		path = manifestPath(targetName)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// IsManagedServer returns true if the server was synced by vibe.
func (m *SyncManifest) IsManagedServer(name string) bool {
	for _, s := range m.MCPServers {
		if s == name {
			return true
		}
	}
	return false
}

// IsManagedSkill returns true if the skill was synced by vibe.
func (m *SyncManifest) IsManagedSkill(name string) bool {
	for _, s := range m.Skills {
		if s == name {
			return true
		}
	}
	return false
}

// UpdateMCPServers sets the list of managed server names and updates the timestamp.
func (m *SyncManifest) UpdateMCPServers(names []string) {
	sort.Strings(names)
	m.MCPServers = names
	m.LastSynced = time.Now()
}

// UpdateSkills sets the list of managed skill names and their checksums.
func (m *SyncManifest) UpdateSkills(names []string, checksums map[string]string) {
	sort.Strings(names)
	m.Skills = names
	m.Checksums = checksums
	m.LastSynced = time.Now()
}

// DirChecksum computes a checksum of a directory's contents by hashing
// all file paths and contents.
func DirChecksum(dir string) (string, error) {
	h := sha256.New()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Write relative path
		rel, _ := filepath.Rel(dir, path)
		io.WriteString(h, rel)

		if info.IsDir() {
			return nil
		}

		// Write file contents
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
