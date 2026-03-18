// Package doctor provides health checks and diagnostics for the vibe CLI.
//
// The doctor package implements a comprehensive system for:
//   - Running health checks to verify the vibe installation is correct
//   - Auto-repairing common issues when possible
//   - Collecting diagnostic information for troubleshooting
//
// # Health Checks
//
// Health checks are implemented as types that satisfy the Check interface.
// Each check has a name, description, can be run to produce a CheckResult,
// and may support auto-repair.
//
// Available checks include:
//   - PrereqsCheck: Verifies required tools (gh, jq, yq, claude, python3) are installed
//   - MarketplaceCheck: Verifies ~/.vibe/marketplace directory exists
//   - MarketplaceRegisteredCheck: Verifies claude-vibe marketplace is registered with Claude
//   - SettingsJSONCheck: Verifies ~/.claude/settings.json exists and is valid JSON
//   - PermissionsCheck: Verifies required permissions are present
//   - MCPConfigCheck: Verifies ~/.config/mcp/config.json exists and is valid
//   - MCPServersCheck: Verifies required MCP servers are configured
//   - PluginsInstalledCheck: Verifies required plugins are installed
//   - PluginsOutdatedCheck: Checks for outdated plugins
//   - EnvVarsCheck: Verifies required environment variables are configured
//   - LocalOwnershipCheck: Verifies ~/.local is not owned by root
//
// # Usage
//
// To run all checks:
//
//	results := doctor.RunAll()
//	for _, r := range results {
//	    fmt.Printf("%s: %s - %s\n", r.Status, r.Name, r.Message)
//	}
//
// To repair failed checks:
//
//	checks := doctor.AllChecks()
//	results := doctor.RunAll()
//	repairResults := doctor.RepairAll(checks, results)
//
// To collect diagnostics:
//
//	info, err := doctor.Collect()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	path, err := info.SaveToDefaultLocation()
//
// # Dependency Injection
//
// All checks support dependency injection for testing. The package defines
// interfaces (CommandChecker, FileChecker, MarketplaceChecker, EnvChecker)
// that can be mocked in tests. Default implementations are provided for
// production use.
package doctor
