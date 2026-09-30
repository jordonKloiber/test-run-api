package run

import "time"

type Status string

const (
	StatusPass Status = "pass"
	StatusFail Status = "fail"
	StatusSkip Status = "skip"
)

type TestResult struct {
	TestName     string
	Status       Status
	DurationMS   int
	ErrorMessage *string
}

type Run struct {
	ID          int64
	SuiteName   string
	Source      string
	SubmittedAt time.Time
	Results     []TestResult
}
