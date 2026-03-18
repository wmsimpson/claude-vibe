package install

// AllSteps returns all installation steps in order.
// Only steps that can complete without any network access are included.
// Prerequisites (Claude Code, Homebrew, jq, yq, gh, etc.) are assumed
// to already be installed on the target machine. Tool installation steps
// have been removed to eliminate GitHub API calls that cause hangs.
// Sudo-requiring steps (cleanup, local_ownership) are also excluded —
// on a fresh personal Mac they are not needed and would block non-interactive runs.
func AllSteps() []Step {
	return []Step{
		&PreflightStep{},          // Local checks only — macOS, disk space
		&ShellDetectionStep{},     // Local — detect shell and RC file
		&PathConfigStep{},         // Local — ensure ~/.local/bin is in PATH
		&ClaudeSetupStep{},        // Local — verify Claude Code is installed
		&DownloadVibeStep{},       // Local — locate repo via VIBE_REPO_PATH
		&MarketplaceSetupStep{},   // Local — write known_marketplaces.json directly
		&PermissionsSyncStep{},    // Local — merge permissions into settings.json
		&PluginsInstallStep{},     // Local — copy plugin files + write JSON directly
		&HooksSyncStep{},          // Local — write hooks to settings.json
		&ModelConfigStep{},        // Local — set default model in settings.json
		&MCPSyncStep{},            // Local — sync MCP server config
		&MCPConfigureStep{},       // Local — ensure MCP entries have name fields
		&VerificationStep{},       // Local — verify key files exist
		&AgentSyncStep{},          // Local — sync to Codex/Cursor if configured
	}
}

// StepCount returns the total number of steps
func StepCount() int {
	return len(AllSteps())
}
