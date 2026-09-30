package run

// PassRate returns the fraction of non-skipped results that passed, in [0,1].
// ok is false if there are no non-skipped results to compute from.
func PassRate(results []TestResult) (rate float64, ok bool) {
	var total, passed int
	for _, r := range results {
		if r.Status == StatusSkip {
			continue
		}
		total++
		if r.Status == StatusPass {
			passed++
		}
	}
	if total == 0 {
		return 0, false
	}
	return float64(passed) / float64(total), true
}

// IsFlaky reports whether results show both passes and failures.
func IsFlaky(results []TestResult) bool {
	var sawPass, sawFail bool
	for _, r := range results {
		switch r.Status {
		case StatusPass:
			sawPass = true
		case StatusFail:
			sawFail = true
		}
	}
	return sawPass && sawFail
}

// AverageDurationMS returns the mean duration across non-skipped results.
// ok is false if there are none.
func AverageDurationMS(results []TestResult) (avgMS float64, ok bool) {
	var total, sum int
	for _, r := range results {
		if r.Status == StatusSkip {
			continue
		}
		total++
		sum += r.DurationMS
	}
	if total == 0 {
		return 0, false
	}
	return float64(sum) / float64(total), true
}
