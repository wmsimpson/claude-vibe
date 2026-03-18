package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// State represents the persistent state of an installation
type State struct {
	// StartedAt is when the installation began
	StartedAt time.Time `json:"started_at"`

	// CurrentStep is the step currently being executed or last failed
	CurrentStep string `json:"current_step"`

	// CompletedSteps lists all successfully completed step IDs
	CompletedSteps []string `json:"completed_steps"`

	// SkippedSteps lists all skipped step IDs
	SkippedSteps []string `json:"skipped_steps"`

	// FailedStep contains information about a failed step
	FailedStep *FailedStepInfo `json:"failed_step,omitempty"`

	// Version of vibe being installed
	Version string `json:"version,omitempty"`
}

// FailedStepInfo contains details about a failed step
type FailedStepInfo struct {
	StepID   string    `json:"step_id"`
	Error    string    `json:"error"`
	FailedAt time.Time `json:"failed_at"`
	Attempts int       `json:"attempts"`
}

// StateFilePath returns the path to the state file
func StateFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vibe", "install-state.json")
}

// LoadState loads the installation state from disk
func LoadState() (*State, error) {
	path := StateFilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No state file
		}
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// Save persists the state to disk
func (s *State) Save() error {
	path := StateFilePath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Delete removes the state file
func (s *State) Delete() error {
	return os.Remove(StateFilePath())
}

// NewState creates a new installation state
func NewState() *State {
	return &State{
		StartedAt:      time.Now(),
		CompletedSteps: []string{},
		SkippedSteps:   []string{},
	}
}

// MarkComplete marks a step as completed
func (s *State) MarkComplete(stepID string) {
	// Avoid duplicates
	for _, id := range s.CompletedSteps {
		if id == stepID {
			return
		}
	}
	s.CompletedSteps = append(s.CompletedSteps, stepID)
	s.CurrentStep = stepID

	// Clear failed step if it was this one
	if s.FailedStep != nil && s.FailedStep.StepID == stepID {
		s.FailedStep = nil
	}
}

// MarkSkipped marks a step as skipped
func (s *State) MarkSkipped(stepID string) {
	// Avoid duplicates
	for _, id := range s.SkippedSteps {
		if id == stepID {
			return
		}
	}
	s.SkippedSteps = append(s.SkippedSteps, stepID)
	s.CurrentStep = stepID
}

// MarkFailed marks a step as failed
func (s *State) MarkFailed(stepID string, err error) {
	attempts := 1
	if s.FailedStep != nil && s.FailedStep.StepID == stepID {
		attempts = s.FailedStep.Attempts + 1
	}

	errMsg := "unknown error"
	if err != nil {
		errMsg = err.Error()
	}

	s.FailedStep = &FailedStepInfo{
		StepID:   stepID,
		Error:    errMsg,
		FailedAt: time.Now(),
		Attempts: attempts,
	}
	s.CurrentStep = stepID
}

// IsStepComplete returns true if a step has been completed
func (s *State) IsStepComplete(stepID string) bool {
	for _, id := range s.CompletedSteps {
		if id == stepID {
			return true
		}
	}
	return false
}

// IsStepSkipped returns true if a step was skipped
func (s *State) IsStepSkipped(stepID string) bool {
	for _, id := range s.SkippedSteps {
		if id == stepID {
			return true
		}
	}
	return false
}

// GetResumePoint returns the step to resume from (after last completed)
func (s *State) GetResumePoint(allSteps []Step) int {
	if len(s.CompletedSteps) == 0 {
		return 0
	}

	lastCompleted := s.CompletedSteps[len(s.CompletedSteps)-1]
	for i, step := range allSteps {
		if step.ID() == lastCompleted {
			return i + 1 // Resume from next step
		}
	}

	return 0
}
