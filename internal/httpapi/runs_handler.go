package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jordonKloiber/test-run-api/internal/run"
)

type RunsHandler struct {
	Store run.Store
}

type testResultDTO struct {
	TestName     string  `json:"test_name"`
	Status       string  `json:"status"`
	DurationMS   int     `json:"duration_ms"`
	ErrorMessage *string `json:"error_message"`
}

type createRunRequest struct {
	SuiteName string          `json:"suite_name"`
	Source    string          `json:"source"`
	Results   []testResultDTO `json:"results"`
}

type createRunResponse struct {
	ID          int64  `json:"id"`
	SuiteName   string `json:"suite_name"`
	Source      string `json:"source"`
	SubmittedAt string `json:"submitted_at"`
	ResultCount int    `json:"result_count"`
}

type getRunResponse struct {
	ID          int64           `json:"id"`
	SuiteName   string          `json:"suite_name"`
	Source      string          `json:"source"`
	SubmittedAt string          `json:"submitted_at"`
	Results     []testResultDTO `json:"results"`
}

// POST /runs
func (h *RunsHandler) CreateRun(w http.ResponseWriter, req *http.Request) {
	var body createRunRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	r := run.Run{
		SuiteName: body.SuiteName,
		Source:    body.Source,
	}
	for _, tr := range body.Results {
		r.Results = append(r.Results, run.TestResult{
			TestName:     tr.TestName,
			Status:       run.Status(tr.Status),
			DurationMS:   tr.DurationMS,
			ErrorMessage: tr.ErrorMessage,
		})
	}

	if err := run.ValidateRun(r); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.Store.CreateRun(req.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create run")
		return
	}

	writeJSON(w, http.StatusCreated, createRunResponse{
		ID:          created.ID,
		SuiteName:   created.SuiteName,
		Source:      created.Source,
		SubmittedAt: created.SubmittedAt.UTC().Format(time.RFC3339),
		ResultCount: len(created.Results),
	})
}

// GET /runs/{id}
func (h *RunsHandler) GetRun(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be an integer")
		return
	}

	r, err := h.Store.GetRun(req.Context(), id)
	if err != nil {
		if errors.Is(err, run.ErrNotFound) {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get run")
		return
	}

	results := make([]testResultDTO, 0, len(r.Results))
	for _, tr := range r.Results {
		results = append(results, testResultDTO{
			TestName:     tr.TestName,
			Status:       string(tr.Status),
			DurationMS:   tr.DurationMS,
			ErrorMessage: tr.ErrorMessage,
		})
	}

	writeJSON(w, http.StatusOK, getRunResponse{
		ID:          r.ID,
		SuiteName:   r.SuiteName,
		Source:      r.Source,
		SubmittedAt: r.SubmittedAt.UTC().Format(time.RFC3339),
		Results:     results,
	})
}
