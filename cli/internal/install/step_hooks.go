package install

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// HooksSyncStep adds the telemetry stop hook to settings.json
type HooksSyncStep struct{}

func (s *HooksSyncStep) ID() string          { return "hooks_sync" }
func (s *HooksSyncStep) Name() string        { return "Sync Hooks" }
func (s *HooksSyncStep) Description() string { return "Configure Claude Code hooks for telemetry" }
func (s *HooksSyncStep) ActiveForm() string  { return "Syncing hooks" }
func (s *HooksSyncStep) CanSkip(opts *Options) bool { return false }
func (s *HooksSyncStep) NeedsSudo() bool     { return false }

func (s *HooksSyncStep) Check(ctx *Context) (bool, error) {
	// Always sync to ensure latest hooks
	return false, nil
}

func (s *HooksSyncStep) Run(ctx *Context) StepResult {
	settingsFile := filepath.Join(ctx.ClaudeDir, "settings.json")

	// Ensure claude directory exists
	if err := os.MkdirAll(ctx.ClaudeDir, 0755); err != nil {
		return Failure("Failed to create .claude directory", err)
	}

	// Read or create settings.json
	var settings map[string]interface{}
	if ctx.FileExists(settingsFile) {
		data, err := os.ReadFile(settingsFile)
		if err != nil {
			return Failure("Failed to read settings.json", err)
		}
		if err := json.Unmarshal(data, &settings); err != nil {
			return Failure("Failed to parse settings.json", err)
		}
	} else {
		settings = map[string]interface{}{
			"allow": []interface{}{},
			"deny":  []interface{}{},
		}
	}

	// Define the telemetry stop hook using the new matcher format
	// Uses --quiet to suppress all output (no stdout/stderr on error)
	telemetryCommand := "vibe telemetry publish --event-type=claude.session.stop --source=claude-code-stop-hook --from-hook --quiet 2>/dev/null || true"
	telemetryHookEntry := map[string]interface{}{
		"matcher": "", // Empty matcher matches all Stop events
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": telemetryCommand,
				"timeout": 30,
			},
		},
	}

	// Get or create hooks section
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
	}

	// Get or create Stop hooks array
	stopHooks, ok := hooks["Stop"].([]interface{})
	if !ok {
		stopHooks = []interface{}{}
	}

	// Check if our hook already exists (check nested hooks array)
	hookExists := false
	for _, h := range stopHooks {
		if hookMap, ok := h.(map[string]interface{}); ok {
			if nestedHooks, ok := hookMap["hooks"].([]interface{}); ok {
				for _, nh := range nestedHooks {
					if nhMap, ok := nh.(map[string]interface{}); ok {
						if cmd, ok := nhMap["command"].(string); ok {
							if cmd == telemetryCommand {
								hookExists = true
								break
							}
						}
					}
				}
			}
		}
		if hookExists {
			break
		}
	}

	// Add hook if it doesn't exist
	if !hookExists {
		stopHooks = append(stopHooks, telemetryHookEntry)
		hooks["Stop"] = stopHooks
		settings["hooks"] = hooks

		// Write back settings
		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return Failure("Failed to marshal settings", err)
		}

		if err := os.WriteFile(settingsFile, data, 0644); err != nil {
			return Failure("Failed to write settings.json", err)
		}

		ctx.Log("Added telemetry stop hook to settings.json")
		return Success("Hooks configured")
	}

	return Success("Hooks already configured")
}
