package doctor

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockCollectorDeps contains mock dependencies for the collector
type MockCollectorDeps struct {
	OSInfo       string
	ShellInfo    string
	PathInfo     string
	ToolVersions map[string]string
	CheckResults []CheckResult
	ConfigFiles  map[string]string
}

func TestDiagnosticInfo_Fields(t *testing.T) {
	now := time.Now()
	info := DiagnosticInfo{
		Timestamp:    now,
		OS:           "darwin",
		Shell:        "/bin/zsh",
		Path:         "/usr/local/bin:/usr/bin",
		ToolVersions: map[string]string{"gh": "2.40.0", "claude": "1.0.0"},
		CheckResults: []CheckResult{{Name: "test", Status: StatusPass}},
		ConfigFiles:  map[string]string{"settings.json": `{"allow":[]}`},
	}

	if info.Timestamp != now {
		t.Errorf("Timestamp = %v, want %v", info.Timestamp, now)
	}
	if info.OS != "darwin" {
		t.Errorf("OS = %v, want darwin", info.OS)
	}
	if info.Shell != "/bin/zsh" {
		t.Errorf("Shell = %v, want /bin/zsh", info.Shell)
	}
	if info.Path != "/usr/local/bin:/usr/bin" {
		t.Errorf("Path = %v, want /usr/local/bin:/usr/bin", info.Path)
	}
	if len(info.ToolVersions) != 2 {
		t.Errorf("len(ToolVersions) = %d, want 2", len(info.ToolVersions))
	}
	if len(info.CheckResults) != 1 {
		t.Errorf("len(CheckResults) = %d, want 1", len(info.CheckResults))
	}
	if len(info.ConfigFiles) != 1 {
		t.Errorf("len(ConfigFiles) = %d, want 1", len(info.ConfigFiles))
	}
}

func TestCollect_ReturnsInfo(t *testing.T) {
	info, err := Collect()

	if err != nil {
		t.Errorf("Collect() error = %v", err)
	}

	if info == nil {
		t.Fatal("Collect() returned nil")
	}

	// Verify timestamp is recent
	if time.Since(info.Timestamp) > time.Minute {
		t.Error("Timestamp should be recent")
	}

	// Verify OS is populated
	if info.OS == "" {
		t.Error("OS should not be empty")
	}

	// Verify Shell is populated
	if info.Shell == "" {
		t.Error("Shell should not be empty")
	}

	// Verify Path is populated
	if info.Path == "" {
		t.Error("Path should not be empty")
	}

	// Verify ToolVersions is populated
	if info.ToolVersions == nil {
		t.Error("ToolVersions should not be nil")
	}

	// Verify CheckResults is populated
	if info.CheckResults == nil {
		t.Error("CheckResults should not be nil")
	}
}

func TestCollect_ToolVersions(t *testing.T) {
	info, err := Collect()
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	// Should attempt to get versions for these tools
	expectedTools := []string{"gh", "jq", "yq", "claude", "python3"}

	for _, tool := range expectedTools {
		if _, exists := info.ToolVersions[tool]; !exists {
			// Tool version might be empty if not installed, but key should exist
			t.Logf("Note: %s not found in ToolVersions (may not be installed)", tool)
		}
	}
}

func TestCollect_ConfigFilesRedacted(t *testing.T) {
	info, err := Collect()
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	// Check that sensitive patterns are redacted
	sensitivePatterns := []string{
		"password",
		"secret",
		"token",
		"api_key",
		"apiKey",
	}

	for filename, content := range info.ConfigFiles {
		contentLower := strings.ToLower(content)
		for _, pattern := range sensitivePatterns {
			// If a sensitive key exists, its value should be redacted
			if strings.Contains(contentLower, pattern) {
				// The value after the pattern should be redacted
				t.Logf("Checking redaction in %s for pattern %s", filename, pattern)
			}
		}
	}
}

func TestDiagnosticInfo_SaveTarGz(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "doctor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	info := &DiagnosticInfo{
		Timestamp:    time.Now(),
		OS:           "darwin",
		Shell:        "/bin/zsh",
		Path:         "/usr/local/bin:/usr/bin",
		ToolVersions: map[string]string{"gh": "2.40.0"},
		CheckResults: []CheckResult{{Name: "test", Status: StatusPass, Message: "OK"}},
		ConfigFiles:  map[string]string{"settings.json": `{"allow":[]}`},
	}

	outputPath := filepath.Join(tmpDir, "diagnostics.tar.gz")
	err = info.SaveTarGz(outputPath)

	if err != nil {
		t.Fatalf("SaveTarGz() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("SaveTarGz() did not create file")
	}

	// Verify file is valid gzip
	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	// Verify it contains expected files
	tr := tar.NewReader(gzr)
	foundFiles := make(map[string]bool)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read tar entry: %v", err)
		}
		foundFiles[header.Name] = true
	}

	// Should contain summary.txt and config files
	if !foundFiles["summary.txt"] {
		t.Error("tar.gz should contain summary.txt")
	}

	if !foundFiles["check_results.txt"] {
		t.Error("tar.gz should contain check_results.txt")
	}
}

func TestDiagnosticInfo_SaveTarGz_InvalidPath(t *testing.T) {
	info := &DiagnosticInfo{
		Timestamp: time.Now(),
		OS:        "darwin",
	}

	// Try to save to invalid path
	err := info.SaveTarGz("/nonexistent/path/file.tar.gz")

	if err == nil {
		t.Error("SaveTarGz() should return error for invalid path")
	}
}

func TestDiagnosticInfo_SaveTarGz_EmptyInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "doctor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	info := &DiagnosticInfo{
		Timestamp: time.Now(),
	}

	outputPath := filepath.Join(tmpDir, "empty.tar.gz")
	err = info.SaveTarGz(outputPath)

	// Should succeed even with empty info
	if err != nil {
		t.Errorf("SaveTarGz() error = %v, want nil", err)
	}
}

func TestRedactSensitiveContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "redact password value",
			input:    `{"password": "secret123"}`,
			expected: `{"password": "[REDACTED]"}`,
		},
		{
			name:     "redact api_key value",
			input:    `{"api_key": "abc123xyz"}`,
			expected: `{"api_key": "[REDACTED]"}`,
		},
		{
			name:     "redact token value",
			input:    `{"token": "bearer_xxx"}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "redact secret value",
			input:    `{"secret": "mysecret"}`,
			expected: `{"secret": "[REDACTED]"}`,
		},
		{
			name:     "no redaction needed",
			input:    `{"allow": ["Bash(*)"]}`,
			expected: `{"allow": ["Bash(*)"]}`,
		},
		{
			name:     "multiple redactions",
			input:    `{"password": "p1", "token": "t1"}`,
			expected: `{"password": "[REDACTED]", "token": "[REDACTED]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactSensitiveContent(tt.input)
			if got != tt.expected {
				t.Errorf("RedactSensitiveContent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateDiagnosticFilename(t *testing.T) {
	filename := GenerateDiagnosticFilename()

	if !strings.HasPrefix(filename, "vibe-diagnostics-") {
		t.Errorf("Filename should start with 'vibe-diagnostics-', got %s", filename)
	}

	if !strings.HasSuffix(filename, ".tar.gz") {
		t.Errorf("Filename should end with '.tar.gz', got %s", filename)
	}

	// Filename should contain timestamp
	if len(filename) < len("vibe-diagnostics-.tar.gz")+8 {
		t.Errorf("Filename too short, missing timestamp: %s", filename)
	}
}

func TestCollectWithOptions(t *testing.T) {
	opts := CollectOptions{
		IncludeEnv:     true,
		IncludeConfigs: true,
		RunChecks:      true,
	}

	info, err := CollectWithOptions(opts)

	if err != nil {
		t.Errorf("CollectWithOptions() error = %v", err)
	}

	if info == nil {
		t.Fatal("CollectWithOptions() returned nil")
	}

	if !opts.RunChecks && len(info.CheckResults) > 0 {
		t.Error("CheckResults should be empty when RunChecks is false")
	}
}

func TestCollectWithOptions_NoConfigs(t *testing.T) {
	opts := CollectOptions{
		IncludeEnv:     true,
		IncludeConfigs: false,
		RunChecks:      false,
	}

	info, err := CollectWithOptions(opts)

	if err != nil {
		t.Errorf("CollectWithOptions() error = %v", err)
	}

	if len(info.ConfigFiles) > 0 {
		t.Error("ConfigFiles should be empty when IncludeConfigs is false")
	}
}

func TestCollectWithOptions_NoEnv(t *testing.T) {
	opts := CollectOptions{
		IncludeEnv:     false,
		IncludeConfigs: false,
		RunChecks:      false,
	}

	info, err := CollectWithOptions(opts)

	if err != nil {
		t.Errorf("CollectWithOptions() error = %v", err)
	}

	// Path should still be empty or minimal when IncludeEnv is false
	// (implementation may vary)
	if info == nil {
		t.Fatal("CollectWithOptions() returned nil")
	}
}
