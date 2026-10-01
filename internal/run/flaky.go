package run

// FlakyResult is one test name's flaky-analysis summary over the most
// recent n runs considered.
type FlakyResult struct {
	TestName       string
	RunsConsidered int
	PassRate       float64
	IsFlaky        bool
	AvgDurationMS  float64
}

// AnalyzeFlaky groups rows by TestName and keeps only the n most recent per
// group, then computes PassRate/IsFlaky/AverageDurationMS per group.
//
// Precondition: rows must already be sorted by TestName, then SubmittedAt
// descending (exactly what Store.ListRecentResults returns) — this does a
// single linear pass, no re-sorting.
func AnalyzeFlaky(rows []RunResult, n int) []FlakyResult {
	result := []FlakyResult{}
	i := 0
	for i < len(rows) {
		currentName := rows[i].TestName
		// j scans forward to find where this test name's group ends
		j := i
		for j < len(rows) && rows[j].TestName == currentName {

			j++
		}
		group := rows[i:j] // all rows for this one test name
		i = j

		limit := n
		if len(group) < limit {
			limit = len(group)
		}
		recent := group[:limit]

		testResults := make([]TestResult, 0, len(recent))
		for _, row := range recent {
			testResults = append(testResults, TestResult{
				Status:     row.Status,
				DurationMS: row.DurationMS,
			})
		}

		passRate, ok := PassRate(testResults)
		if !ok {
			passRate = 0
		}
		avgDuration, ok := AverageDurationMS(testResults)
		if !ok {
			avgDuration = 0
		}
		isFlaky := IsFlaky(testResults)

		result = append(result, FlakyResult{
			TestName:       currentName,
			RunsConsidered: len(recent),
			PassRate:       passRate,
			IsFlaky:        isFlaky,
			AvgDurationMS:  avgDuration,
		})

	}

	return (result)
}
