package install

import (
	"fmt"
	"strings"

	"github.com/wmsimpson/claude-vibe/cli/internal/config"
	vibesync "github.com/wmsimpson/claude-vibe/cli/internal/sync"
)

// AgentSyncStep syncs MCP servers and skills to other coding agents (Codex, Cursor).
type AgentSyncStep struct{}

func (s *AgentSyncStep) ID() string          { return "agent_sync" }
func (s *AgentSyncStep) Name() string        { return "Agent Sync" }
func (s *AgentSyncStep) Description() string { return "Sync MCP and skills to Codex/Cursor" }
func (s *AgentSyncStep) ActiveForm() string  { return "Syncing to other agents" }
func (s *AgentSyncStep) NeedsSudo() bool     { return false }

func (s *AgentSyncStep) CanSkip(opts *Options) bool {
	// Skip if auto_sync is not enabled in config
	cfg, err := config.Load()
	if err != nil {
		return true
	}
	return !cfg.Settings.AutoSync
}

func (s *AgentSyncStep) Check(ctx *Context) (bool, error) {
	// Always run during install to ensure everything is in sync
	return false, nil
}

func (s *AgentSyncStep) Run(ctx *Context) StepResult {
	targets := vibesync.AllTargets()

	// Filter by configured targets if set
	cfg, err := config.Load()
	if err == nil && len(cfg.Settings.SyncTargets) > 0 {
		targets = vibesync.FilterTargets(targets, cfg.Settings.SyncTargets)
	}

	// Only sync to installed targets
	var installed []vibesync.SyncTarget
	for _, t := range targets {
		if t.IsInstalled() {
			installed = append(installed, t)
		}
	}

	if len(installed) == 0 {
		return Skip("No sync targets installed (Codex/Cursor)")
	}

	ctx.Log("Syncing to %d target(s)...", len(installed))

	results := vibesync.RunSync(installed, vibesync.SyncOptions{})

	totalSynced := 0
	var errors []string
	var targetNames []string

	for _, r := range results {
		totalSynced += r.ItemsSynced
		targetNames = append(targetNames, r.Target)
		for _, e := range r.Errors {
			errors = append(errors, fmt.Sprintf("%s: %s", r.Target, e))
		}
		for _, d := range r.Details {
			ctx.Log("  %s: %s", r.Target, d)
		}
	}

	if len(errors) > 0 {
		return FailureWithHint(
			fmt.Sprintf("Sync completed with errors: %s", strings.Join(errors, "; ")),
			nil,
			"Run 'vibe sync --status' to check what's out of sync",
		)
	}

	return Success(fmt.Sprintf("Synced %d items to %s", totalSynced, strings.Join(targetNames, ", ")))
}
