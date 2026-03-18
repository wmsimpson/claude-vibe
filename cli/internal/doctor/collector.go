package doctor

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// DiagnosticInfo contains collected diagnostic information.
type DiagnosticInfo struct {
	Timestamp    time.Time
	OS           string
	Shell        string
	Path         string
	ToolVersions map[string]string
	CheckResults []CheckResult
	ConfigFiles  map[string]string // contents are redacted
}

// CollectOptions configures what information to collect.
type CollectOptions struct {
	IncludeEnv     bool
	IncludeConfigs bool
	RunChecks      bool
}

// Collect gathers diagnostic information with default options.
func Collect() (*DiagnosticInfo, error) {
	return CollectWithOptions(CollectOptions{
		IncludeEnv:     true,
		IncludeConfigs: true,
		RunChecks:      true,
	})
}

// CollectWithOptions gathers diagnostic information with specified options.
func CollectWithOptions(opts CollectOptions) (*DiagnosticInfo, error) {
	info := &DiagnosticInfo{
		Timestamp:    time.Now(),
		OS:           fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		ToolVersions: make(map[string]string),
		ConfigFiles:  make(map[string]string),
	}

	// Collect environment info
	if opts.IncludeEnv {
		info.Shell = os.Getenv("SHELL")
		info.Path = os.Getenv("PATH")
	}

	// Collect tool versions
	tools := []string{"gh", "jq", "yq", "claude", "python3"}
	for _, tool := range tools {
		version, err := getToolVersion(tool)
		if err != nil {
			info.ToolVersions[tool] = "(not installed)"
		} else {
			info.ToolVersions[tool] = version
		}
	}

	// Run health checks
	if opts.RunChecks {
		info.CheckResults = RunAll()
	}

	// Collect config files
	if opts.IncludeConfigs {
		collectConfigFiles(info)
	}

	return info, nil
}

// getToolVersion attempts to get the version of a command.
func getToolVersion(cmd string) (string, error) {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return "", err
	}

	// Try common version flags
	versionFlags := []string{"--version", "-version", "version"}

	for _, flag := range versionFlags {
		out, err := exec.Command(path, flag).Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0]), nil
			}
		}
	}

	return "(version unknown)", nil
}

// collectConfigFiles collects and redacts configuration files.
func collectConfigFiles(info *DiagnosticInfo) {
	homeDir, _ := os.UserHomeDir()

	configPaths := map[string]string{
		"settings.json":   filepath.Join(homeDir, ".claude", "settings.json"),
		"mcp_config.json": filepath.Join(homeDir, ".config", "mcp", "config.json"),
		"vibe_config":     filepath.Join(homeDir, ".vibe", "config.yaml"),
	}

	for name, path := range configPaths {
		content, err := os.ReadFile(path)
		if err != nil {
			info.ConfigFiles[name] = fmt.Sprintf("(error reading: %v)", err)
			continue
		}

		// Redact sensitive content
		redacted := RedactSensitiveContent(string(content))
		info.ConfigFiles[name] = redacted
	}
}

// RedactSensitiveContent removes sensitive values from content.
func RedactSensitiveContent(content string) string {
	// Patterns to redact (JSON format)
	patterns := []string{
		`"password"\s*:\s*"[^"]*"`,
		`"api_key"\s*:\s*"[^"]*"`,
		`"apiKey"\s*:\s*"[^"]*"`,
		`"token"\s*:\s*"[^"]*"`,
		`"secret"\s*:\s*"[^"]*"`,
		`"credential"\s*:\s*"[^"]*"`,
		`"auth"\s*:\s*"[^"]*"`,
	}

	result := content

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			// Find the key name
			keyRe := regexp.MustCompile(`"([^"]+)"\s*:`)
			keyMatch := keyRe.FindStringSubmatch(match)
			if len(keyMatch) > 1 {
				return fmt.Sprintf(`"%s": "[REDACTED]"`, keyMatch[1])
			}
			return `"[REDACTED]": "[REDACTED]"`
		})
	}

	return result
}

// SaveTarGz saves the diagnostic info to a gzipped tar archive.
func (d *DiagnosticInfo) SaveTarGz(path string) error {
	// Create the output file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Create gzip writer
	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	// Create tar writer
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add summary.txt
	summary := d.generateSummary()
	if err := addFileToTar(tw, "summary.txt", []byte(summary)); err != nil {
		return err
	}

	// Add check_results.txt
	checkResults := d.generateCheckResults()
	if err := addFileToTar(tw, "check_results.txt", []byte(checkResults)); err != nil {
		return err
	}

	// Add config files
	for name, content := range d.ConfigFiles {
		filename := fmt.Sprintf("configs/%s", name)
		if err := addFileToTar(tw, filename, []byte(content)); err != nil {
			return err
		}
	}

	// Add tool versions as JSON
	versionsJSON, _ := json.MarshalIndent(d.ToolVersions, "", "  ")
	if err := addFileToTar(tw, "tool_versions.json", versionsJSON); err != nil {
		return err
	}

	return nil
}

// addFileToTar adds a file with the given content to the tar archive.
func addFileToTar(tw *tar.Writer, name string, content []byte) error {
	header := &tar.Header{
		Name:    name,
		Mode:    0644,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header for %s: %w", name, err)
	}

	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("failed to write content for %s: %w", name, err)
	}

	return nil
}

// generateSummary creates a human-readable summary of the diagnostic info.
func (d *DiagnosticInfo) generateSummary() string {
	var sb strings.Builder

	sb.WriteString("Vibe Diagnostics Report\n")
	sb.WriteString("=======================\n\n")

	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", d.Timestamp.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("OS: %s\n", d.OS))
	sb.WriteString(fmt.Sprintf("Shell: %s\n", d.Shell))
	sb.WriteString("\n")

	sb.WriteString("Tool Versions:\n")
	for tool, version := range d.ToolVersions {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", tool, version))
	}
	sb.WriteString("\n")

	sb.WriteString("PATH:\n")
	for _, p := range strings.Split(d.Path, ":") {
		sb.WriteString(fmt.Sprintf("  %s\n", p))
	}

	return sb.String()
}

// generateCheckResults creates a formatted list of check results.
func (d *DiagnosticInfo) generateCheckResults() string {
	var sb strings.Builder

	sb.WriteString("Health Check Results\n")
	sb.WriteString("====================\n\n")

	pass, fail, warning := CountByStatus(d.CheckResults)
	sb.WriteString(fmt.Sprintf("Summary: %d passed, %d failed, %d warnings\n\n", pass, fail, warning))

	for _, result := range d.CheckResults {
		var statusIcon string
		switch result.Status {
		case StatusPass:
			statusIcon = "[PASS]"
		case StatusFail:
			statusIcon = "[FAIL]"
		case StatusWarning:
			statusIcon = "[WARN]"
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", statusIcon, result.Name))
		sb.WriteString(fmt.Sprintf("  Message: %s\n", result.Message))
		if result.RepairHint != "" {
			sb.WriteString(fmt.Sprintf("  Hint: %s\n", result.RepairHint))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GenerateDiagnosticFilename creates a filename with timestamp.
func GenerateDiagnosticFilename() string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("vibe-diagnostics-%s.tar.gz", timestamp)
}

// SaveToDefaultLocation saves diagnostics to the current directory.
func (d *DiagnosticInfo) SaveToDefaultLocation() (string, error) {
	filename := GenerateDiagnosticFilename()
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	path := filepath.Join(cwd, filename)
	if err := d.SaveTarGz(path); err != nil {
		return "", err
	}

	return path, nil
}

// ToJSON returns the diagnostic info as JSON.
func (d *DiagnosticInfo) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

// PrintSummary prints a human-readable summary to stdout.
func (d *DiagnosticInfo) PrintSummary() {
	fmt.Println(d.generateSummary())
}

// PrintCheckResults prints check results to stdout.
func (d *DiagnosticInfo) PrintCheckResults() {
	fmt.Println(d.generateCheckResults())
}
