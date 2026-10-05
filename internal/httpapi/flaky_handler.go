package httpapi

import (
	"net/http"
	"strconv"

	"github.com/jordonKloiber/test-run-api/internal/run"
)

type flakyResultDTO struct {
	TestName       string  `json:"test_name"`
	RunsConsidered int     `json:"runs_considered"`
	PassRate       float64 `json:"pass_rate"`
	IsFlaky        bool    `json:"is_flaky"`
	AvgDurationMS  float64 `json:"avg_duration_ms"`
}

// GET /flaky?source=<optional>&n=<optional, default 10>
func (h *RunsHandler) ListFlaky(w http.ResponseWriter, req *http.Request) {
	source := req.URL.Query().Get("source")

	nStr := req.URL.Query().Get("n")
	n := 10 // default if the caller does not specify run number
	if nStr != "" {
		parsed, err := strconv.Atoi(nStr)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "n must be a positive integer")
			return
		}
		n = parsed
	}

	rows, err := h.Store.ListRecentResults(req.Context(), source)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list results")
		return
	}

	flakyResults := run.AnalyzeFlaky(rows, n)

	dtos := make([]flakyResultDTO, 0, len(flakyResults))
	for _, fr := range flakyResults {
		dtos = append(dtos, flakyResultDTO{
			TestName:       fr.TestName,
			RunsConsidered: fr.RunsConsidered,
			PassRate:       fr.PassRate,
			IsFlaky:        fr.IsFlaky,
			AvgDurationMS:  fr.AvgDurationMS,
		})
	}

	writeJSON(w, http.StatusOK, dtos)
}
