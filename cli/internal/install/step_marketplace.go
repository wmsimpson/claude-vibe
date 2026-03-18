package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// MarketplaceSetupStep registers the vibe repo as a local marketplace by writing
// directly to ~/.claude/plugins/known_marketplaces.json. This avoids any network
// calls that `claude plugin marketplace add` might make.
type MarketplaceSetupStep struct{}

func (s *MarketplaceSetupStep) ID() string                      { return "marketplace_setup" }
func (s *MarketplaceSetupStep) Name() string                    { return "Setup Marketplace" }
func (s *MarketplaceSetupStep) Description() string             { return "Register vibe marketplace locally" }
func (s *MarketplaceSetupStep) ActiveForm() string              { return "Setting up marketplace" }
func (s *MarketplaceSetupStep) CanSkip(opts *Options) bool      { return false }
func (s *MarketplaceSetupStep) NeedsSudo() bool                 { return false }

func (s *MarketplaceSetupStep) Check(ctx *Context) (bool, error) {
	knownPath := filepath.Join(ctx.ClaudeDir, "plugins", "known_marketplaces.json")
	data, err := os.ReadFile(knownPath)
	if err != nil {
		return false, nil
	}

	var marketplaces map[string]interface{}
	if err := json.Unmarshal(data, &marketplaces); err != nil {
		return false, nil
	}

	entry, ok := marketplaces["claude-vibe"].(map[string]interface{})
	if !ok {
		return false, nil
	}

	src, ok := entry["source"].(map[string]interface{})
	if !ok {
		return false, nil
	}

	return src["path"] == ctx.VibeDir, nil
}

func (s *MarketplaceSetupStep) Run(ctx *Context) StepResult {
	if ctx.VibeDir == "" {
		return Failure("Vibe directory not set (download step failed?)", nil)
	}

	knownPath := filepath.Join(ctx.ClaudeDir, "plugins", "known_marketplaces.json")

	// Read existing entries to preserve other marketplaces (e.g. claude-plugins-official)
	var marketplaces map[string]interface{}
	if data, err := os.ReadFile(knownPath); err == nil {
		json.Unmarshal(data, &marketplaces)
	}
	if marketplaces == nil {
		marketplaces = make(map[string]interface{})
	}

	// Register our directory marketplace
	marketplaces["claude-vibe"] = map[string]interface{}{
		"source": map[string]interface{}{
			"source": "directory",
			"path":   ctx.VibeDir,
		},
		"installLocation": ctx.VibeDir,
		"lastUpdated":     time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(marketplaces, "", "  ")
	if err != nil {
		return Failure("Failed to marshal marketplace config", err)
	}

	if err := os.MkdirAll(filepath.Dir(knownPath), 0755); err != nil {
		return Failure("Failed to create plugins directory", err)
	}

	if err := os.WriteFile(knownPath, data, 0644); err != nil {
		return Failure("Failed to write marketplace config", err)
	}

	ctx.Log("Registered claude-vibe → %s", ctx.VibeDir)
	return SuccessWithDetails(
		"Marketplace registered",
		"  claude-vibe → "+ctx.VibeDir,
	)
}
