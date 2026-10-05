package run

import (
	"reflect"
	"testing"
	"time"
)

// TestAnalyzeFlaky covers the grouping + truncation + stats-wiring logic:
// single/multiple test names, n smaller and larger than a group's row
// count, all-skip groups, and an empty input slice.
func TestAnalyzeFlaky(t *testing.T) {
	day := func(d int) time.Time {
		return time.Date(2026, 1, d, 0, 0, 0, 0, time.UTC)
	}

	tests := []struct {
		name  string
		input []RunResult
		n     int
		want  []FlakyResult
	}{
		{
			name: "single test name, mixed results within n",
			input: []RunResult{
				{TestName: "checkout works", Status: StatusFail, DurationMS: 50, SubmittedAt: day(3)},
				{TestName: "checkout works", Status: StatusPass, DurationMS: 100, SubmittedAt: day(2)},
			},
			n: 10,
			want: []FlakyResult{
				{TestName: "checkout works", RunsConsidered: 2, PassRate: 0.5, IsFlaky: true, AvgDurationMS: 75},
			},
		},
		{
			name: "n larger than available rows uses all of them",
			input: []RunResult{
				{TestName: "only ran twice", Status: StatusPass, DurationMS: 100, SubmittedAt: day(2)},
				{TestName: "only ran twice", Status: StatusFail, DurationMS: 50, SubmittedAt: day(1)},
			},
			n: 10,
			want: []FlakyResult{
				{TestName: "only ran twice", RunsConsidered: 2, PassRate: 0.5, IsFlaky: true, AvgDurationMS: 75},
			},
		},
		{
			name: "more rows than n for one test name",
			input: []RunResult{
				{TestName: "ran three times", Status: StatusPass, DurationMS: 100, SubmittedAt: day(3)},
				{TestName: "ran three times", Status: StatusFail, DurationMS: 50, SubmittedAt: day(2)},
				{TestName: "ran three times", Status: StatusPass, DurationMS: 150, SubmittedAt: day(1)},
			},
			n: 2,
			want: []FlakyResult{
				{TestName: "ran three times", RunsConsidered: 2, PassRate: 0.5, IsFlaky: true, AvgDurationMS: 75},
			},
		},
		{
			name: "multiple distinct test names",
			input: []RunResult{
				{TestName: "ran passing once", Status: StatusPass, DurationMS: 100, SubmittedAt: day(3)},
				{TestName: "ran passing once", Status: StatusFail, DurationMS: 100, SubmittedAt: day(3)},
				{TestName: "ran failing once", Status: StatusFail, DurationMS: 50, SubmittedAt: day(2)},
				{TestName: "ran failing once", Status: StatusPass, DurationMS: 50, SubmittedAt: day(2)},
				{TestName: "ran passing second time", Status: StatusPass, DurationMS: 150, SubmittedAt: day(1)},
				{TestName: "ran passing second time", Status: StatusFail, DurationMS: 150, SubmittedAt: day(1)},
			},
			n: 10,
			want: []FlakyResult{
				{TestName: "ran passing once", RunsConsidered: 2, PassRate: 0.5, IsFlaky: true, AvgDurationMS: 100},
				{TestName: "ran failing once", RunsConsidered: 2, PassRate: 0.5, IsFlaky: true, AvgDurationMS: 50},
				{TestName: "ran passing second time", RunsConsidered: 2, PassRate: 0.5, IsFlaky: true, AvgDurationMS: 150},
			},
		},
		{
			name: "all-skip is still reported",
			input: []RunResult{
				{TestName: "ran skipped test", Status: StatusSkip, DurationMS: 100, SubmittedAt: day(3)},
				{TestName: "ran skipped test", Status: StatusSkip, DurationMS: 50, SubmittedAt: day(2)},
				{TestName: "ran skipped test", Status: StatusSkip, DurationMS: 150, SubmittedAt: day(1)},
			},
			n: 3,
			want: []FlakyResult{
				{TestName: "ran skipped test", RunsConsidered: 3, PassRate: 0, IsFlaky: false, AvgDurationMS: 0},
			},
		},
		{
			name:  "empty slice results in empty output",
			input: []RunResult{},
			n:     10,
			want:  []FlakyResult{},
		},
		{
			name: "n == 1 results in one single result",
			input: []RunResult{
				{TestName: "ran mixed test", Status: StatusPass, DurationMS: 100, SubmittedAt: day(3)},
				{TestName: "ran mixed test", Status: StatusFail, DurationMS: 50, SubmittedAt: day(2)},
				{TestName: "ran mixed test", Status: StatusSkip, DurationMS: 150, SubmittedAt: day(1)},
			},
			n: 1,
			want: []FlakyResult{
				{TestName: "ran mixed test", RunsConsidered: 1, PassRate: 1, IsFlaky: false, AvgDurationMS: 100},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnalyzeFlaky(tt.input, tt.n)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AnalyzeFlaky(%+v, %d) = %+v, want %+v", tt.input, tt.n, got, tt.want)
			}
		})
	}
}
