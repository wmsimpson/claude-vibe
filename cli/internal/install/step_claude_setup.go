package install

// ClaudeSetupStep verifies Claude Code is installed and accessible.
type ClaudeSetupStep struct{}

func (s *ClaudeSetupStep) ID() string                 { return "claude_setup" }
func (s *ClaudeSetupStep) Name() string               { return "Claude Code" }
func (s *ClaudeSetupStep) Description() string        { return "Verify Claude Code CLI is installed" }
func (s *ClaudeSetupStep) ActiveForm() string         { return "Checking Claude Code" }
func (s *ClaudeSetupStep) CanSkip(opts *Options) bool { return false }
func (s *ClaudeSetupStep) NeedsSudo() bool            { return false }

func (s *ClaudeSetupStep) Check(ctx *Context) (bool, error) {
	return ctx.IsCommandInstalled("claude"), nil
}

func (s *ClaudeSetupStep) Run(ctx *Context) StepResult {
	if !ctx.IsCommandInstalled("claude") {
		return FailureWithHint(
			"Claude Code not installed",
			nil,
			"Install Claude Code: npm install -g @anthropic-ai/claude-code",
		)
	}
	ver, _ := ctx.RunCommand("claude", "--version")
	return SuccessWithDetails("Claude Code is installed", "  Version: "+ver)
}
