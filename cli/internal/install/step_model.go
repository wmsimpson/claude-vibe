package install

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ModelConfigStep sets the default Claude model to opus
type ModelConfigStep struct{}

func (s *ModelConfigStep) ID() string          { return "model_config" }
func (s *ModelConfigStep) Name() string        { return "Configure Default Model" }
func (s *ModelConfigStep) Description() string { return "Set default Claude model to opus" }
func (s *ModelConfigStep) ActiveForm() string  { return "Configuring default model" }
func (s *ModelConfigStep) CanSkip(opts *Options) bool { return false }
func (s *ModelConfigStep) NeedsSudo() bool     { return false }

func (s *ModelConfigStep) Check(ctx *Context) (bool, error) {
	// Always run to ensure model is set correctly
	return false, nil
}

func (s *ModelConfigStep) Run(ctx *Context) StepResult {
	settingsFile := filepath.Join(ctx.ClaudeDir, "settings.json")

	// Read current settings
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return Failure("Failed to read settings.json", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return Failure("Failed to parse settings.json", err)
	}

	// Set default model to opus
	settings["model"] = "opus"

	// Write back
	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return Failure("Failed to marshal settings", err)
	}

	if err := os.WriteFile(settingsFile, output, 0644); err != nil {
		return Failure("Failed to write settings.json", err)
	}

	ctx.Log("Default model set to opus")
	return Success("Default model set to opus")
}
