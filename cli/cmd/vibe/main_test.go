package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFindVibeRoot tests the findVibeRoot directory-walking function.
func TestFindVibeRoot(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(root string) string // returns the startDir to pass to findVibeRoot
		expectErr bool
	}{
		{
			name: "found via marketplace.json at start dir",
			setup: func(root string) string {
				pluginDir := filepath.Join(root, ".claude-plugin")
				if err := os.MkdirAll(pluginDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(pluginDir, "marketplace.json"), []byte("{}"), 0644); err != nil {
					t.Fatal(err)
				}
				return root
			},
			expectErr: false,
		},
		{
			name: "found by walking up from a subdirectory",
			setup: func(root string) string {
				pluginDir := filepath.Join(root, ".claude-plugin")
				if err := os.MkdirAll(pluginDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(pluginDir, "marketplace.json"), []byte("{}"), 0644); err != nil {
					t.Fatal(err)
				}
				subDir := filepath.Join(root, "plugins", "databricks-tools", "skills")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatal(err)
				}
				return subDir
			},
			expectErr: false,
		},
		{
			name: "found via plugin.json fallback when marketplace.json absent",
			setup: func(root string) string {
				pluginDir := filepath.Join(root, ".claude-plugin")
				if err := os.MkdirAll(pluginDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte("{}"), 0644); err != nil {
					t.Fatal(err)
				}
				return root
			},
			expectErr: false,
		},
		{
			name: "error when not in a vibe repo",
			setup: func(root string) string {
				// No .claude-plugin directory — plain empty dir
				return root
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			startDir := tc.setup(root)

			got, err := findVibeRoot(startDir)

			if tc.expectErr {
				if err == nil {
					t.Errorf("findVibeRoot(%q) expected error, got nil (returned %q)", startDir, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("findVibeRoot(%q) unexpected error: %v", startDir, err)
			}

			// The returned path should be the repo root, not the subdirectory we started from.
			// Evaluate symlinks so macOS /private/var/... paths compare correctly.
			wantRoot, _ := filepath.EvalSymlinks(root)
			gotResolved, _ := filepath.EvalSymlinks(got)
			if gotResolved != wantRoot {
				t.Errorf("findVibeRoot(%q) = %q, want %q", startDir, got, wantRoot)
			}
		})
	}
}
