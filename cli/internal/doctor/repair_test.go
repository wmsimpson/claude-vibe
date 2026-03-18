package doctor

import (
	"errors"
	"testing"
)

// MockCheck implements Check interface for testing
type MockCheck struct {
	name        string
	description string
	result      CheckResult
	canRepair   bool
	repairError error
	repairCalls int
}

func (m *MockCheck) Name() string {
	return m.name
}

func (m *MockCheck) Description() string {
	return m.description
}

func (m *MockCheck) Run() CheckResult {
	return m.result
}

func (m *MockCheck) CanRepair() bool {
	return m.canRepair
}

func (m *MockCheck) Repair() error {
	m.repairCalls++
	return m.repairError
}

func TestRepair_SuccessfulRepair(t *testing.T) {
	check := &MockCheck{
		name:        "test_check",
		description: "Test check",
		canRepair:   true,
		repairError: nil,
	}

	err := Repair(check)

	if err != nil {
		t.Errorf("Repair() error = %v, want nil", err)
	}

	if check.repairCalls != 1 {
		t.Errorf("repairCalls = %d, want 1", check.repairCalls)
	}
}

func TestRepair_FailedRepair(t *testing.T) {
	expectedError := errors.New("repair failed")
	check := &MockCheck{
		name:        "test_check",
		description: "Test check",
		canRepair:   true,
		repairError: expectedError,
	}

	err := Repair(check)

	if err == nil {
		t.Error("Repair() should return error")
	}

	if !errors.Is(err, expectedError) {
		t.Errorf("Repair() error = %v, want %v", err, expectedError)
	}
}

func TestRepair_CannotRepair(t *testing.T) {
	check := &MockCheck{
		name:        "test_check",
		description: "Test check",
		canRepair:   false,
	}

	err := Repair(check)

	if err == nil {
		t.Error("Repair() should return error when CanRepair() is false")
	}

	if check.repairCalls != 0 {
		t.Errorf("repairCalls = %d, want 0 (should not call Repair)", check.repairCalls)
	}
}

func TestRepairAll_AllSuccessful(t *testing.T) {
	checks := []Check{
		&MockCheck{
			name:      "check1",
			canRepair: true,
			result:    CheckResult{Name: "check1", Status: StatusFail},
		},
		&MockCheck{
			name:      "check2",
			canRepair: true,
			result:    CheckResult{Name: "check2", Status: StatusFail},
		},
	}

	results := []CheckResult{
		{Name: "check1", Status: StatusFail},
		{Name: "check2", Status: StatusFail},
	}

	repairResults := RepairAll(checks, results)

	if len(repairResults) != 2 {
		t.Errorf("len(repairResults) = %d, want 2", len(repairResults))
	}

	for _, result := range repairResults {
		if result.Error != nil {
			t.Errorf("RepairResult for %s has error: %v", result.CheckName, result.Error)
		}
		if !result.Repaired {
			t.Errorf("RepairResult for %s should be repaired", result.CheckName)
		}
	}
}

func TestRepairAll_SomeSkipped(t *testing.T) {
	checks := []Check{
		&MockCheck{
			name:      "passing_check",
			canRepair: true,
			result:    CheckResult{Name: "passing_check", Status: StatusPass},
		},
		&MockCheck{
			name:      "failing_check",
			canRepair: true,
			result:    CheckResult{Name: "failing_check", Status: StatusFail},
		},
	}

	results := []CheckResult{
		{Name: "passing_check", Status: StatusPass},
		{Name: "failing_check", Status: StatusFail},
	}

	repairResults := RepairAll(checks, results)

	// Only failing checks should be repaired
	if len(repairResults) != 1 {
		t.Errorf("len(repairResults) = %d, want 1", len(repairResults))
	}

	if repairResults[0].CheckName != "failing_check" {
		t.Errorf("Repaired check = %s, want failing_check", repairResults[0].CheckName)
	}
}

func TestRepairAll_CannotRepair(t *testing.T) {
	checks := []Check{
		&MockCheck{
			name:      "unrepairable",
			canRepair: false,
			result:    CheckResult{Name: "unrepairable", Status: StatusFail},
		},
	}

	results := []CheckResult{
		{Name: "unrepairable", Status: StatusFail},
	}

	repairResults := RepairAll(checks, results)

	if len(repairResults) != 1 {
		t.Errorf("len(repairResults) = %d, want 1", len(repairResults))
	}

	if repairResults[0].Repaired {
		t.Error("unrepairable check should not be marked as repaired")
	}

	if repairResults[0].Skipped != true {
		t.Error("unrepairable check should be marked as skipped")
	}
}

func TestRepairAll_RepairFailed(t *testing.T) {
	expectedError := errors.New("repair failed")
	checks := []Check{
		&MockCheck{
			name:        "failing_repair",
			canRepair:   true,
			repairError: expectedError,
			result:      CheckResult{Name: "failing_repair", Status: StatusFail},
		},
	}

	results := []CheckResult{
		{Name: "failing_repair", Status: StatusFail},
	}

	repairResults := RepairAll(checks, results)

	if len(repairResults) != 1 {
		t.Errorf("len(repairResults) = %d, want 1", len(repairResults))
	}

	if repairResults[0].Repaired {
		t.Error("failed repair should not be marked as repaired")
	}

	if repairResults[0].Error == nil {
		t.Error("failed repair should have error")
	}
}

func TestRepairAll_EmptyResults(t *testing.T) {
	repairResults := RepairAll(nil, nil)

	if len(repairResults) != 0 {
		t.Errorf("len(repairResults) = %d, want 0", len(repairResults))
	}
}

func TestRepairAll_WarningsIncluded(t *testing.T) {
	checks := []Check{
		&MockCheck{
			name:      "warning_check",
			canRepair: true,
			result:    CheckResult{Name: "warning_check", Status: StatusWarning},
		},
	}

	results := []CheckResult{
		{Name: "warning_check", Status: StatusWarning},
	}

	repairResults := RepairAll(checks, results)

	// Warnings should also be repaired if they can be
	if len(repairResults) != 1 {
		t.Errorf("len(repairResults) = %d, want 1", len(repairResults))
	}

	if repairResults[0].CheckName != "warning_check" {
		t.Errorf("Repaired check = %s, want warning_check", repairResults[0].CheckName)
	}
}

// Test RepairResult struct
func TestRepairResult_Fields(t *testing.T) {
	result := RepairResult{
		CheckName: "test_check",
		Repaired:  true,
		Skipped:   false,
		Error:     nil,
		Message:   "Repair successful",
	}

	if result.CheckName != "test_check" {
		t.Errorf("CheckName = %v, want test_check", result.CheckName)
	}
	if !result.Repaired {
		t.Error("Repaired should be true")
	}
	if result.Skipped {
		t.Error("Skipped should be false")
	}
}
