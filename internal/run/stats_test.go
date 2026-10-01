package run

import "testing"

// TestPassRate covers the fraction-of-passes calculation: all-pass, all-fail,
// and mixed results compute the right rate; all-skip and empty slices report
// ok=false (nothing to compute from); skips are excluded from the denominator
// even when mixed in with pass/fail results that do count.
func TestPassRate(t *testing.T) {
	tests := []struct {
		name     string
		input    []TestResult
		wantRate float64
		wantOk   bool
	}{
		{
			name: "all pass",
			input: []TestResult{
				{Status: StatusPass},
				{Status: StatusPass},
			},
			wantRate: 1.0,
			wantOk:   true,
		},
		{
			name: "all fail",
			input: []TestResult{
				{Status: StatusFail},
				{Status: StatusFail},
			},
			wantRate: 0.0,
			wantOk:   true,
		},
		{
			name: "mixed results",
			input: []TestResult{
				{Status: StatusPass},
				{Status: StatusFail},
			},
			wantRate: 0.5,
			wantOk:   true,
		},
		{
			name: "all skip",
			input: []TestResult{
				{Status: StatusSkip},
				{Status: StatusSkip},
			},
			wantRate: 0.0,
			wantOk:   false,
		},
		{
			name:     "empty slice",
			input:    []TestResult{},
			wantRate: 0.0,
			wantOk:   false,
		},
		{
			name: "mixed with skips",
			input: []TestResult{
				{Status: StatusFail},
				{Status: StatusSkip},
				{Status: StatusSkip},
				{Status: StatusPass},
			},
			wantRate: 0.5,
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRate, gotOk := PassRate(tt.input)
			if gotOk != tt.wantOk {
				t.Errorf("PassRate(%+v) ok = %v, want %v", tt.input, gotOk, tt.wantOk)
			}
			if gotOk && gotRate != tt.wantRate {
				t.Errorf("PassRate(%+v) rate = %v, want %v", tt.input, gotRate, tt.wantRate)
			}
		})
	}
}

// TestIsFlaky covers the flakiness signal: only a genuine pass/fail
// disagreement counts as flaky. All-pass, all-fail, a single result, an
// empty slice, and skip paired with either pass or fail (or skip alone)
// should all report false — skip never contributes to the flaky signal.
func TestIsFlaky(t *testing.T) {
	tests := []struct {
		name  string
		input []TestResult
		want  bool
	}{
		{
			name: "pass and fail mixed is flaky",
			input: []TestResult{
				{Status: StatusPass},
				{Status: StatusFail},
			},
			want: true,
		},
		{
			name: "all pass is not flaky",
			input: []TestResult{
				{Status: StatusPass},
				{Status: StatusPass},
			},
			want: false,
		},
		{
			name: "all fail is not flaky",
			input: []TestResult{
				{Status: StatusFail},
				{Status: StatusFail},
			},
			want: false,
		},
		{
			name: "pass and skip mixed is not flaky",
			input: []TestResult{
				{Status: StatusPass},
				{Status: StatusSkip},
			},
			want: false,
		},
		{
			name: "fail and skip mixed is not flaky",
			input: []TestResult{
				{Status: StatusFail},
				{Status: StatusSkip},
			},
			want: false,
		},
		{
			name: "single result is not flaky",
			input: []TestResult{
				{Status: StatusPass},
			},
			want: false,
		},
		{
			name:  "empty slice is not flaky",
			input: []TestResult{},
			want:  false,
		},
		{
			name:  "all skip is not flaky",
			input: []TestResult{{Status: StatusSkip}, {Status: StatusSkip}},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFlaky(tt.input)
			if got != tt.want {
				t.Errorf("IsFlaky(%+v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestAverageDurationMS covers the mean-duration calculation: known
// durations average correctly, skips are excluded even when mixed with
// counted results, a single result returns its own duration, and both an
// empty slice and an all-skip (but non-empty) slice report ok=false.
func TestAverageDurationMS(t *testing.T) {
	tests := []struct {
		name    string
		input   []TestResult
		wantAvg float64
		wantOk  bool
	}{
		{
			name: "two known durations",
			input: []TestResult{
				{Status: StatusPass, DurationMS: 100},
				{Status: StatusPass, DurationMS: 200},
			},
			wantAvg: 150,
			wantOk:  true,
		},
		{
			name: "three known durations, one skip",
			input: []TestResult{
				{Status: StatusPass, DurationMS: 100},
				{Status: StatusFail, DurationMS: 200},
				{Status: StatusSkip, DurationMS: 1000},
			},
			wantAvg: 150,
			wantOk:  true,
		},
		{
			name:    "empty slice",
			input:   []TestResult{},
			wantAvg: 0,
			wantOk:  false,
		},
		{
			name: "single result",
			input: []TestResult{
				{Status: StatusPass, DurationMS: 100},
			},
			wantAvg: 100,
			wantOk:  true,
		},
		{
			name: "all skip, non-empty slice",
			input: []TestResult{
				{Status: StatusSkip, DurationMS: 100},
				{Status: StatusSkip, DurationMS: 200}},
			wantAvg: 0,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAvg, gotOk := AverageDurationMS(tt.input)
			if gotOk != tt.wantOk {
				t.Errorf("AverageDurationMS(%+v) ok = %v, want %v", tt.input, gotOk, tt.wantOk)
			}
			if gotOk && gotAvg != tt.wantAvg {
				t.Errorf("AverageDurationMS(%+v) avg = %v, want %v", tt.input, gotAvg, tt.wantAvg)
			}
		})
	}
}
