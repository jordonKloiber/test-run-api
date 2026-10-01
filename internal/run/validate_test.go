package run

import (
	"strings"
	"testing"
)

// ValidateTestResult — Hand-written tests
func TestValidateTestResult(t *testing.T) {
	warningMsg := "test passed but with a warning"

	tests := []struct {
		name    string // short description of the case
		input   TestResult
		wantErr bool
	}{
		{
			name: "valid pass result",
			input: TestResult{
				TestName:   "passing-test",
				Status:     StatusPass,
				DurationMS: 120,
			},
			wantErr: false,
		},
		{
			name: "valid fail result",
			input: TestResult{
				TestName:   "failing-test",
				Status:     StatusFail,
				DurationMS: 120,
			},
			wantErr: false,
		},
		{
			name: "valid skip result",
			input: TestResult{
				TestName:   "skip-test",
				Status:     StatusSkip,
				DurationMS: 120,
			},
			wantErr: false,
		},
		{
			name: "empty test name is invalid",
			input: TestResult{
				TestName:   "",
				Status:     StatusPass,
				DurationMS: 120,
			},
			wantErr: true,
		},
		{
			name: "invalid status value, pass",
			input: TestResult{
				TestName:   "placeholder-test",
				Status:     "passed",
				DurationMS: 120,
			},
			wantErr: true,
		},
		{
			name: "invalid status value, pass - uppercase",
			input: TestResult{
				TestName:   "placeholder-test",
				Status:     "PASS",
				DurationMS: 120,
			},
			wantErr: true,
		},
		{
			name: "invalid status value, fail",
			input: TestResult{
				TestName:   "placeholder-test",
				Status:     "failed",
				DurationMS: 120,
			},
			wantErr: true,
		},
		{
			name: "invalid status value, fail - uppercase",
			input: TestResult{
				TestName:   "placeholder-test",
				Status:     "FAIL",
				DurationMS: 120,
			},
			wantErr: true,
		},
		{
			name: "invalid status value, skip",
			input: TestResult{
				TestName:   "placeholder-test",
				Status:     "skipped",
				DurationMS: 120,
			},
			wantErr: true,
		},
		{
			name: "invalid status value, skip - uppercase",
			input: TestResult{
				TestName:   "placeholder-test",
				Status:     "SKIP",
				DurationMS: 120,
			},
			wantErr: true,
		},
		{
			name: "invalid duration value",
			input: TestResult{
				TestName:   "placeholder-test",
				Status:     StatusPass,
				DurationMS: -120,
			},
			wantErr: true,
		},
		{
			name: "valid zero duration",
			input: TestResult{
				TestName:   "placeholder-test",
				Status:     StatusPass,
				DurationMS: 0,
			},
			wantErr: false,
		},
		{
			name: "fail status with nil error message",
			input: TestResult{
				TestName:     "placeholder-test",
				Status:       StatusFail,
				DurationMS:   120,
				ErrorMessage: nil,
			},
			wantErr: false,
		},
		{
			name: "pass status with non-nil error message",
			input: TestResult{
				TestName:     "placeholder-test",
				Status:       StatusPass,
				DurationMS:   120,
				ErrorMessage: &warningMsg,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTestResult(tt.input)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("ValidateTestResult(%+v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// ValidateRun — skeleton only. Fill in the table below.
// Remember: per the plan, ValidateRun must COLLECT ALL violations, not stop
// at the first one — so at least one case here should assert that (see the
// commented-out example at the bottom for how to check error content, not
// just error/no-error).
func TestValidateRun(t *testing.T) {
	tests := []struct {
		name    string
		input   Run
		wantErr bool
	}{
		{
			name: "valid run with mixed results",
			input: Run{
				SuiteName: "checkout-e2e",
				Source:    "github:example/repo",
				Results: []TestResult{
					{TestName: "passing-test", Status: StatusPass, DurationMS: 100},
					{TestName: "failing-test", Status: StatusFail, DurationMS: 50},
				},
			},
			wantErr: false,
		},
		{
			name: "valid run with empty suite_name error",
			input: Run{
				SuiteName: "",
				Source:    "github:example/repo",
				Results: []TestResult{
					{TestName: "", Status: StatusPass, DurationMS: 100},
					{TestName: "failing-test", Status: StatusFail, DurationMS: 50},
				},
			},
			wantErr: true,
		},
		{
			name: "valid run with empty source error",
			input: Run{
				SuiteName: "checkout-e2e",
				Source:    "",
				Results: []TestResult{
					{TestName: "passing-test", Status: StatusPass, DurationMS: 100},
					{TestName: "failing-test", Status: StatusFail, DurationMS: 50},
				},
			},
			wantErr: true,
		},
		{
			name: "valid run with empty results",
			input: Run{
				SuiteName: "checkout-e2e",
				Source:    "github:example/repo",
				Results:   []TestResult{},
			},
			wantErr: true,
		},
		{
			name: "valid run with nil results",
			input: Run{
				SuiteName: "checkout-e2e",
				Source:    "github:example/repo",
				Results:   nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRun(tt.input)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("ValidateRun(%+v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// TestValidateRun_CollectsAllErrors proves ValidateRun collects every
// violation instead of stopping at the first — errors.Join's Error() joins
// all wrapped messages with newlines, so both substrings should be present.
func TestValidateRun_CollectsAllErrors(t *testing.T) {
	r := Run{
		SuiteName: "",
		Source:    "",
		Results:   []TestResult{{TestName: "t", Status: StatusPass, DurationMS: 1}},
	}
	err := ValidateRun(r)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "suite_name") {
		t.Errorf("expected error to mention suite_name, got: %s", msg)
	}
	if !strings.Contains(msg, "source") {
		t.Errorf("expected error to mention source, got: %s", msg)
	}
}

func TestValidateRun_CollectsAllResultErrors(t *testing.T) {
	r := Run{
		SuiteName: "checkout-e2e",
		Source:    "github:example/repo",
		Results: []TestResult{
			{TestName: "", Status: StatusPass, DurationMS: 100},            // bad: empty TestName
			{TestName: "sample-test", Status: "passed", DurationMS: 50},    // bad: invalid Status
			{TestName: "sample-test", Status: StatusPass, DurationMS: -10}, // bad: negative DurationMS
		},
	}
	err := ValidateRun(r)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "results[0]: test_name") {
		t.Errorf("expected error to mention result[0]'s test_name violation, got: %s", msg)
	}
	if !strings.Contains(msg, "results[1]: status") {
		t.Errorf("expected error to mention result[1]'s status violation, got: %s", msg)
	}
	if !strings.Contains(msg, "results[2]: duration_ms") {
		t.Errorf("expected error to mention result[2]'s duration_ms violation, got: %s", msg)
	}
}
