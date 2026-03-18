package install

// ShellDetectionStep detects the user's shell and sets the RC file path
type ShellDetectionStep struct{}

func (s *ShellDetectionStep) ID() string          { return "shell_detection" }
func (s *ShellDetectionStep) Name() string        { return "Detect Shell" }
func (s *ShellDetectionStep) Description() string { return "Detect shell and configuration file" }
func (s *ShellDetectionStep) ActiveForm() string  { return "Detecting shell" }
func (s *ShellDetectionStep) CanSkip(opts *Options) bool { return false }
func (s *ShellDetectionStep) NeedsSudo() bool     { return false }

func (s *ShellDetectionStep) Check(ctx *Context) (bool, error) {
	// Shell is already detected during context creation
	return ctx.ShellType != "" && ctx.ShellRC != "", nil
}

func (s *ShellDetectionStep) Run(ctx *Context) StepResult {
	// Shell is detected in context creation, just report it
	if ctx.ShellType == "" || ctx.ShellRC == "" {
		return Failure("Could not detect shell", nil)
	}

	details := "  - Shell: " + ctx.ShellType + "\n"
	details += "  - RC file: " + ctx.ShellRC + "\n"
	details += "  - Architecture: " + ctx.Arch + "\n"
	details += "  - Homebrew prefix: " + ctx.HomebrewPrefix

	return SuccessWithDetails(
		"Detected "+ctx.ShellType+" shell",
		details,
	)
}
