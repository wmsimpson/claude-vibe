package doctor

import (
	"fmt"
)

// RepairResult contains the outcome of a repair attempt.
type RepairResult struct {
	CheckName string
	Repaired  bool
	Skipped   bool
	Error     error
	Message   string
}

// Repair attempts to fix the issue identified by a check.
// Returns an error if the check cannot be repaired or if repair fails.
func Repair(check Check) error {
	if !check.CanRepair() {
		return fmt.Errorf("check %s cannot be auto-repaired", check.Name())
	}

	return check.Repair()
}

// RepairAll attempts to repair all failed or warning checks.
// It returns a slice of RepairResult for each check that was attempted.
func RepairAll(checks []Check, results []CheckResult) []RepairResult {
	if checks == nil || results == nil {
		return []RepairResult{}
	}

	// Build a map from check name to check for quick lookup
	checkMap := make(map[string]Check)
	for _, check := range checks {
		checkMap[check.Name()] = check
	}

	var repairResults []RepairResult

	for _, result := range results {
		// Skip passing checks
		if result.Status == StatusPass {
			continue
		}

		check, exists := checkMap[result.Name]
		if !exists {
			repairResults = append(repairResults, RepairResult{
				CheckName: result.Name,
				Repaired:  false,
				Skipped:   true,
				Error:     fmt.Errorf("check not found"),
				Message:   "Check not found in check list",
			})
			continue
		}

		if !check.CanRepair() {
			repairResults = append(repairResults, RepairResult{
				CheckName: result.Name,
				Repaired:  false,
				Skipped:   true,
				Message:   fmt.Sprintf("Cannot auto-repair: %s", result.RepairHint),
			})
			continue
		}

		err := check.Repair()
		if err != nil {
			repairResults = append(repairResults, RepairResult{
				CheckName: result.Name,
				Repaired:  false,
				Skipped:   false,
				Error:     err,
				Message:   fmt.Sprintf("Repair failed: %v", err),
			})
		} else {
			repairResults = append(repairResults, RepairResult{
				CheckName: result.Name,
				Repaired:  true,
				Skipped:   false,
				Message:   "Repair successful",
			})
		}
	}

	return repairResults
}

// RepairInteractive runs repairs with confirmation for each step.
// The confirm function is called before each repair and should return
// true to proceed or false to skip.
func RepairInteractive(checks []Check, results []CheckResult, confirm func(checkName, repairHint string) bool) []RepairResult {
	if checks == nil || results == nil {
		return []RepairResult{}
	}

	checkMap := make(map[string]Check)
	for _, check := range checks {
		checkMap[check.Name()] = check
	}

	var repairResults []RepairResult

	for _, result := range results {
		if result.Status == StatusPass {
			continue
		}

		check, exists := checkMap[result.Name]
		if !exists {
			continue
		}

		if !check.CanRepair() {
			repairResults = append(repairResults, RepairResult{
				CheckName: result.Name,
				Repaired:  false,
				Skipped:   true,
				Message:   fmt.Sprintf("Cannot auto-repair: %s", result.RepairHint),
			})
			continue
		}

		// Ask for confirmation
		if !confirm(result.Name, result.RepairHint) {
			repairResults = append(repairResults, RepairResult{
				CheckName: result.Name,
				Repaired:  false,
				Skipped:   true,
				Message:   "Skipped by user",
			})
			continue
		}

		err := check.Repair()
		if err != nil {
			repairResults = append(repairResults, RepairResult{
				CheckName: result.Name,
				Repaired:  false,
				Skipped:   false,
				Error:     err,
				Message:   fmt.Sprintf("Repair failed: %v", err),
			})
		} else {
			repairResults = append(repairResults, RepairResult{
				CheckName: result.Name,
				Repaired:  true,
				Skipped:   false,
				Message:   "Repair successful",
			})
		}
	}

	return repairResults
}

// RepairSingle repairs a single check by name.
// Returns an error if the check is not found, cannot be repaired, or repair fails.
func RepairSingle(checks []Check, checkName string) error {
	for _, check := range checks {
		if check.Name() == checkName {
			return Repair(check)
		}
	}

	return fmt.Errorf("check %s not found", checkName)
}

// GetRepairableChecks returns only the checks that can be auto-repaired.
func GetRepairableChecks(checks []Check) []Check {
	var repairable []Check
	for _, check := range checks {
		if check.CanRepair() {
			repairable = append(repairable, check)
		}
	}
	return repairable
}

// GetFailedResults returns only the failed or warning results.
func GetFailedResults(results []CheckResult) []CheckResult {
	var failed []CheckResult
	for _, result := range results {
		if result.Status != StatusPass {
			failed = append(failed, result)
		}
	}
	return failed
}

// CountByStatus counts results by their status.
func CountByStatus(results []CheckResult) (pass, fail, warning int) {
	for _, result := range results {
		switch result.Status {
		case StatusPass:
			pass++
		case StatusFail:
			fail++
		case StatusWarning:
			warning++
		}
	}
	return
}

// HasIssues returns true if any results are not passing.
func HasIssues(results []CheckResult) bool {
	_, fail, warning := CountByStatus(results)
	return fail > 0 || warning > 0
}

// HasCriticalIssues returns true if any results are failing (not just warnings).
func HasCriticalIssues(results []CheckResult) bool {
	_, fail, _ := CountByStatus(results)
	return fail > 0
}
