package install

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MCPSyncStep merges mcp-servers.yaml into MCP config
type MCPSyncStep struct{}

func (s *MCPSyncStep) ID() string          { return "mcp_sync" }
func (s *MCPSyncStep) Name() string        { return "Sync MCP" }
func (s *MCPSyncStep) Description() string { return "Merge MCP server configurations" }
func (s *MCPSyncStep) ActiveForm() string  { return "Syncing MCP servers" }
func (s *MCPSyncStep) CanSkip(opts *Options) bool { return false }
func (s *MCPSyncStep) NeedsSudo() bool     { return false }

func (s *MCPSyncStep) Check(ctx *Context) (bool, error) {
	// Always sync to ensure latest config
	return false, nil
}

func (s *MCPSyncStep) Run(ctx *Context) StepResult {
	ctx.Log("Finding MCP servers configuration...")

	// Find MCP servers file
	mcpServersFile := ""
	if ctx.VibeDir != "" {
		mcpServersFile = filepath.Join(ctx.VibeDir, "mcp-servers.yaml")
	}
	if mcpServersFile == "" || !ctx.FileExists(mcpServersFile) {
		mcpServersFile = filepath.Join(ctx.MarketplaceDir, "mcp-servers.yaml")
	}

	if !ctx.FileExists(mcpServersFile) {
		return Skip("mcp-servers.yaml not found")
	}

	ctx.Log("Reading %s", filepath.Base(mcpServersFile))

	mcpConfigDir := filepath.Join(ctx.ConfigDir, "mcp")
	mcpConfig := filepath.Join(mcpConfigDir, "config.json")

	// Ensure MCP config directory exists
	if err := os.MkdirAll(mcpConfigDir, 0755); err != nil {
		return Failure("Failed to create .config/mcp directory", err)
	}

	// Save existing enabled states before modification
	var savedStates map[string]bool
	if ctx.FileExists(mcpConfig) {
		data, _ := os.ReadFile(mcpConfig)
		var config map[string]interface{}
		if json.Unmarshal(data, &config) == nil {
			if claudeCode, ok := config["claude-code"].(map[string]interface{}); ok {
				savedStates = make(map[string]bool)
				for name, server := range claudeCode {
					if serverMap, ok := server.(map[string]interface{}); ok {
						if enabled, ok := serverMap["enabled"].(bool); ok {
							savedStates[name] = enabled
						}
					}
				}
			}
		}
	}

	// Create MCP config if it doesn't exist
	if !ctx.FileExists(mcpConfig) {
		initial := map[string]interface{}{
			"claude-code": map[string]interface{}{},
		}
		data, _ := json.MarshalIndent(initial, "", "  ")
		os.WriteFile(mcpConfig, data, 0644)
	}

	// Check for yq
	if !ctx.IsCommandInstalled("yq") {
		return FailureWithHint(
			"yq not installed",
			nil,
			"Install yq: brew install yq",
		)
	}

	// Extract servers from YAML and expand paths
	ctx.Log("Extracting MCP servers from %s", mcpServersFile)

	// Build jq filter to expand ~ to home directory
	expandFilter := `walk(
		if type == "string" then
			if startswith("~/") then
				sub("^~/"; "` + ctx.HomeDir + `/")
			elif contains("=~/") then
				gsub("=~/"; "=` + ctx.HomeDir + `/")
			else
				.
			end
		else
			.
		end
	)`

	// Get servers JSON from YAML
	serversOutput, err := exec.Command("yq", "-r", ".servers | @json", mcpServersFile).Output()
	if err != nil {
		return Failure("Failed to parse mcp-servers.yaml", err)
	}

	// Expand paths using jq
	cmd := exec.Command("jq", expandFilter)
	cmd.Stdin = strings.NewReader(string(serversOutput))
	expandedOutput, err := cmd.Output()
	if err != nil {
		return Failure("Failed to expand paths", err)
	}

	// Add enabled and name fields, remove type
	transformFilter := `to_entries | map({
		key: .key,
		value: ((.value + {enabled: true, name: .key}) | del(.type))
	}) | from_entries`

	cmd = exec.Command("jq", transformFilter)
	cmd.Stdin = strings.NewReader(string(expandedOutput))
	transformedOutput, err := cmd.Output()
	if err != nil {
		return Failure("Failed to transform servers", err)
	}

	// Read current config
	configData, err := os.ReadFile(mcpConfig)
	if err != nil {
		return Failure("Failed to read MCP config", err)
	}

	// Merge new servers into config
	mergeFilter := `."claude-code" = (
		(."claude-code" // {}) as $existing |
		$existing |
		. * $new_servers
	)`

	cmd = exec.Command("jq", "--argjson", "new_servers", string(transformedOutput), mergeFilter)
	cmd.Stdin = strings.NewReader(string(configData))
	merged, err := cmd.Output()
	if err != nil {
		return Failure("Failed to merge MCP config", err)
	}

	// Write merged config
	if err := os.WriteFile(mcpConfig, merged, 0644); err != nil {
		return Failure("Failed to write MCP config", err)
	}

	ctx.Log("Merged MCP servers into config")

	// Restore saved enabled states
	if len(savedStates) > 0 {
		ctx.Log("Restoring %d saved server states", len(savedStates))
		// Re-read the config
		data, _ := os.ReadFile(mcpConfig)
		var config map[string]interface{}
		if json.Unmarshal(data, &config) == nil {
			if claudeCode, ok := config["claude-code"].(map[string]interface{}); ok {
				for name, enabled := range savedStates {
					if serverMap, ok := claudeCode[name].(map[string]interface{}); ok {
						serverMap["enabled"] = enabled
					}
				}
				// Write back with restored states
				data, _ = json.MarshalIndent(config, "", "  ")
				os.WriteFile(mcpConfig, data, 0644)
			}
		}
	}

	return Success("MCP servers synced")
}
