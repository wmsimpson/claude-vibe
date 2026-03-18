package install

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// MCPConfigureStep ensures MCP server config entries have name fields.
type MCPConfigureStep struct{}

func (s *MCPConfigureStep) ID() string                 { return "mcp_configure" }
func (s *MCPConfigureStep) Name() string               { return "Configure MCP" }
func (s *MCPConfigureStep) Description() string        { return "Finalise MCP server configuration" }
func (s *MCPConfigureStep) ActiveForm() string         { return "Configuring MCP" }
func (s *MCPConfigureStep) CanSkip(opts *Options) bool { return false }
func (s *MCPConfigureStep) NeedsSudo() bool            { return false }

func (s *MCPConfigureStep) Check(ctx *Context) (bool, error) {
	return false, nil // always run to keep config up to date
}

func (s *MCPConfigureStep) Run(ctx *Context) StepResult {
	mcpConfig := filepath.Join(ctx.ConfigDir, "mcp", "config.json")
	if !ctx.FileExists(mcpConfig) {
		return Skip("No MCP config found (will be created when MCPs are enabled)")
	}

	data, err := os.ReadFile(mcpConfig)
	if err != nil {
		return Failure("Failed to read MCP config", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return Failure("Failed to parse MCP config", err)
	}

	// Ensure every server entry has a name field
	modified := false
	if claudeCode, ok := config["claude-code"].(map[string]interface{}); ok {
		for name, server := range claudeCode {
			if serverMap, ok := server.(map[string]interface{}); ok {
				if _, hasName := serverMap["name"]; !hasName {
					serverMap["name"] = name
					modified = true
				}
			}
		}
	}

	if modified {
		data, _ = json.MarshalIndent(config, "", "  ")
		if err := os.WriteFile(mcpConfig, data, 0644); err != nil {
			return Failure("Failed to write MCP config", err)
		}
	}

	return Success("MCP configuration verified")
}
