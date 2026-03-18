package install

// StepStatus represents the current status of an installation step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepComplete
	StepFailed
	StepSkipped
	StepNeedsSudo // Step requires sudo to complete
	StepNeedsExec // Step needs to run a command with full terminal access (suspends TUI)
)

func (s StepStatus) String() string {
	switch s {
	case StepPending:
		return "pending"
	case StepRunning:
		return "running"
	case StepComplete:
		return "complete"
	case StepFailed:
		return "failed"
	case StepSkipped:
		return "skipped"
	case StepNeedsSudo:
		return "needs_sudo"
	case StepNeedsExec:
		return "needs_exec"
	default:
		return "unknown"
	}
}

// StepResult contains the outcome of running a step
type StepResult struct {
	Status     StepStatus
	Message    string
	Error      error
	RepairHint string   // Instructions if manual intervention needed
	Details    string   // Additional details (for verbose mode)
	SudoCmd    string   // Command to run with sudo (when Status is StepNeedsSudo)
	SudoArgs   []string // Arguments for sudo command
	ExecCmd    string   // Command to run with full terminal access (when Status is StepNeedsExec)
	ExecArgs   []string // Arguments for exec command
	ExecStdin  string   // Optional stdin to feed to the command
}

// Success creates a successful step result
func Success(message string) StepResult {
	return StepResult{
		Status:  StepComplete,
		Message: message,
	}
}

// SuccessWithDetails creates a successful step result with details
func SuccessWithDetails(message, details string) StepResult {
	return StepResult{
		Status:  StepComplete,
		Message: message,
		Details: details,
	}
}

// Skip creates a skipped step result
func Skip(message string) StepResult {
	return StepResult{
		Status:  StepSkipped,
		Message: message,
	}
}

// Failure creates a failed step result
func Failure(message string, err error) StepResult {
	return StepResult{
		Status:  StepFailed,
		Message: message,
		Error:   err,
	}
}

// FailureWithHint creates a failed step result with recovery hint
func FailureWithHint(message string, err error, hint string) StepResult {
	return StepResult{
		Status:     StepFailed,
		Message:    message,
		Error:      err,
		RepairHint: hint,
	}
}

// NeedsSudo creates a result indicating sudo is required to complete the step
func NeedsSudo(message string, cmd string, args []string) StepResult {
	return StepResult{
		Status:   StepNeedsSudo,
		Message:  message,
		SudoCmd:  cmd,
		SudoArgs: args,
	}
}

// NeedsExec creates a result indicating the step needs to run a command with
// full terminal access. The TUI will suspend, giving the command direct access
// to stdin/stdout/stderr, then resume when the command completes. Use this for
// commands that may prompt for user input (e.g., SSH host key verification).
func NeedsExec(message string, cmd string, args []string, stdin string) StepResult {
	return StepResult{
		Status:    StepNeedsExec,
		Message:   message,
		ExecCmd:   cmd,
		ExecArgs:  args,
		ExecStdin: stdin,
	}
}

// Step defines the interface for all installation steps
type Step interface {
	// ID returns a unique identifier for state persistence
	ID() string

	// Name returns a short display name
	Name() string

	// Description returns a human-readable description
	Description() string

	// ActiveForm returns the present continuous form (e.g., "Installing...")
	ActiveForm() string

	// CanSkip returns true if this step can be skipped based on options
	CanSkip(opts *Options) bool

	// NeedsSudo returns true if this step requires elevated privileges
	NeedsSudo() bool

	// Check returns true if this step is already complete
	Check(ctx *Context) (bool, error)

	// Run executes the step and returns the result
	Run(ctx *Context) StepResult
}
