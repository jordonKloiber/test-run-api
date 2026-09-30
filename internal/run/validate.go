package run

import (
	"errors"
	"fmt"
)

// ValidateRun checks a Run for business-rule validity, collecting every
// violation (not just the first) across the run and all its results.
// Returns nil if valid.
func ValidateRun(r Run) error {
	var errs []error

	if r.SuiteName == "" {
		errs = append(errs, errors.New("suite_name is required"))
	}
	if r.Source == "" {
		errs = append(errs, errors.New("source is required"))
	}
	if len(r.Results) == 0 {
		errs = append(errs, errors.New("results must not be empty"))
	}

	for i, tr := range r.Results {
		if err := ValidateTestResult(tr); err != nil {
			errs = append(errs, fmt.Errorf("results[%d]: %w", i, err))
		}
	}

	return errors.Join(errs...)
}

// ValidateTestResult checks a single TestResult in isolation. Called by
// ValidateRun per-result, exported so it's independently testable.
func ValidateTestResult(tr TestResult) error {
	var errs []error

	if tr.TestName == "" {
		errs = append(errs, errors.New("test_name is required"))
	}

	switch tr.Status {
	case StatusPass, StatusFail, StatusSkip:
	default:
		errs = append(errs, fmt.Errorf("status %q is invalid", tr.Status))
	}

	if tr.DurationMS < 0 {
		errs = append(errs, errors.New("duration_ms must be >= 0"))
	}

	return errors.Join(errs...)
}
