package install

import (
	"testing"
)

func TestAllSteps(t *testing.T) {
	steps := AllSteps()

	// Verify we have the expected number of steps
	expectedCount := 14
	if len(steps) != expectedCount {
		t.Errorf("Expected %d steps, got %d", expectedCount, len(steps))
	}

	// Verify each step has required fields
	seenIDs := make(map[string]bool)
	for i, step := range steps {
		// Check ID is unique
		if seenIDs[step.ID()] {
			t.Errorf("Step %d has duplicate ID: %s", i, step.ID())
		}
		seenIDs[step.ID()] = true

		// Check required fields are not empty
		if step.ID() == "" {
			t.Errorf("Step %d has empty ID", i)
		}
		if step.Name() == "" {
			t.Errorf("Step %d (%s) has empty Name", i, step.ID())
		}
		if step.Description() == "" {
			t.Errorf("Step %d (%s) has empty Description", i, step.ID())
		}
		if step.ActiveForm() == "" {
			t.Errorf("Step %d (%s) has empty ActiveForm", i, step.ID())
		}
	}
}

func TestStepOrder(t *testing.T) {
	steps := AllSteps()

	// Verify critical steps come in the right order.
	// Only local steps are included — no network-dependent tool installs.
	expectedOrder := []string{
		"preflight",         // Local checks only — macOS, disk space
		"shell_detection",   // Local — detect shell and RC file
		"path_config",       // Local — ensure ~/.local/bin is in PATH
		"claude_setup",      // Local — verify Claude Code is installed
		"download_vibe",     // Local — locate repo via VIBE_REPO_PATH
		"marketplace_setup", // Local — write known_marketplaces.json directly
		"permissions_sync",  // Local — merge permissions into settings.json
		"plugins_install",   // Local — copy plugin files + write JSON directly
		"hooks_sync",        // Local — write hooks to settings.json
		"model_config",      // Local — set default model in settings.json
		"mcp_sync",          // Local — sync MCP server config
		"mcp_configure",     // Local — ensure MCP entries have name fields
		"verification",      // Local — verify key files exist
		"agent_sync",        // Local — sync to Codex/Cursor if configured
	}

	if len(steps) != len(expectedOrder) {
		t.Errorf("Step count mismatch: expected %d, got %d", len(expectedOrder), len(steps))
		return
	}

	for i, step := range steps {
		if step.ID() != expectedOrder[i] {
			t.Errorf("Step %d: expected ID %q, got %q", i, expectedOrder[i], step.ID())
		}
	}
}

func TestOptionsDefaults(t *testing.T) {
	opts := DefaultOptions()

	if opts.ForceReinstall {
		t.Error("ForceReinstall should default to false")
	}
	if opts.CleanOnly {
		t.Error("CleanOnly should default to false")
	}
	if opts.SkipJAMF {
		t.Error("SkipJAMF should default to false")
	}
	if opts.SkipPlugins {
		t.Error("SkipPlugins should default to false")
	}
	if opts.NoInteractive {
		t.Error("NoInteractive should default to false")
	}
	if opts.Resume {
		t.Error("Resume should default to false")
	}
	if opts.Verbose {
		t.Error("Verbose should default to false")
	}
	if opts.NoBrew {
		t.Error("NoBrew should default to false")
	}
}

func TestStepResults(t *testing.T) {
	// Test Success result
	result := Success("test message")
	if result.Status != StepComplete {
		t.Error("Success should have StepComplete status")
	}
	if result.Message != "test message" {
		t.Error("Success should preserve message")
	}

	// Test Skip result
	result = Skip("skipped")
	if result.Status != StepSkipped {
		t.Error("Skip should have StepSkipped status")
	}

	// Test Failure result
	result = Failure("failed", nil)
	if result.Status != StepFailed {
		t.Error("Failure should have StepFailed status")
	}

	// Test FailureWithHint result
	result = FailureWithHint("failed", nil, "try this")
	if result.Status != StepFailed {
		t.Error("FailureWithHint should have StepFailed status")
	}
	if result.RepairHint != "try this" {
		t.Error("FailureWithHint should preserve hint")
	}

	// Test NeedsSudo result
	result = NeedsSudo("needs sudo", "bash", []string{"-c", "echo hello"})
	if result.Status != StepNeedsSudo {
		t.Error("NeedsSudo should have StepNeedsSudo status")
	}
	if result.SudoCmd != "bash" {
		t.Error("NeedsSudo should preserve SudoCmd")
	}
	if len(result.SudoArgs) != 2 || result.SudoArgs[0] != "-c" {
		t.Error("NeedsSudo should preserve SudoArgs")
	}
}

func TestStateMarking(t *testing.T) {
	state := NewState()

	// Test MarkComplete
	state.MarkComplete("step1")
	if !state.IsStepComplete("step1") {
		t.Error("Step should be marked complete")
	}
	if state.IsStepComplete("step2") {
		t.Error("Unmarked step should not be complete")
	}

	// Test MarkSkipped
	state.MarkSkipped("step2")
	if !state.IsStepSkipped("step2") {
		t.Error("Step should be marked skipped")
	}

	// Test no duplicates
	state.MarkComplete("step1")
	count := 0
	for _, id := range state.CompletedSteps {
		if id == "step1" {
			count++
		}
	}
	if count != 1 {
		t.Error("MarkComplete should not create duplicates")
	}
}

func TestCanSkipOptions(t *testing.T) {
	opts := DefaultOptions()

	// JAMF step should be skippable when SkipJAMF is set
	opts.SkipJAMF = true
	jamfStep := &JAMFCertsStep{}
	if !jamfStep.CanSkip(opts) {
		t.Error("JAMF step should be skippable with SkipJAMF option")
	}

	// Plugins step should be skippable when SkipPlugins is set
	opts.SkipPlugins = true
	pluginsStep := &PluginsInstallStep{}
	if !pluginsStep.CanSkip(opts) {
		t.Error("Plugins step should be skippable with SkipPlugins option")
	}

	// Preflight should never be skippable
	preflightStep := &PreflightStep{}
	if preflightStep.CanSkip(opts) {
		t.Error("Preflight step should not be skippable")
	}

	// Homebrew step should be skippable when NoBrew is set
	opts.NoBrew = true
	homebrewStep := &HomebrewStep{}
	if !homebrewStep.CanSkip(opts) {
		t.Error("Homebrew step should be skippable with NoBrew option")
	}

	// BrewTools step should be skippable when NoBrew is set
	brewToolsStep := &BrewToolsStep{}
	if !brewToolsStep.CanSkip(opts) {
		t.Error("BrewTools step should be skippable with NoBrew option")
	}

	// Node step should be skippable when NoBrew is set
	nodeStep := &NodeNpmStep{}
	if !nodeStep.CanSkip(opts) {
		t.Error("Node step should be skippable with NoBrew option")
	}

	// Gcloud step should be skippable when NoBrew is set
	gcloudStep := &GcloudCliStep{}
	if !gcloudStep.CanSkip(opts) {
		t.Error("Gcloud step should be skippable with NoBrew option")
	}

	// Databricks step should be skippable when NoBrew is set
	databricksStep := &DatabricksCliStep{}
	if !databricksStep.CanSkip(opts) {
		t.Error("Databricks step should be skippable with NoBrew option")
	}
}
