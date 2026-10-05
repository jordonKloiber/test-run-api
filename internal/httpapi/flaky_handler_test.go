package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jordonKloiber/test-run-api/internal/run"
)

// TestListFlaky covers GET /flaky: defaults, source/n overrides, n
// validation, store errors, and empty results.
func TestListFlaky(t *testing.T) {
	t.Run("no query params uses defaults", func(t *testing.T) {
		var gotSource string
		store := &fakeStore{
			listRecentResultsFunc: func(ctx context.Context, source string) ([]run.RunResult, error) {
				gotSource = source
				return []run.RunResult{
					{TestName: "checkout works", Status: run.StatusPass, DurationMS: 100, SubmittedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
					{TestName: "checkout works", Status: run.StatusFail, DurationMS: 50, SubmittedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
				}, nil
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler, testAPIKey)

		req := httptest.NewRequest(http.MethodGet, "/flaky", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if gotSource != "" {
			t.Errorf("ListRecentResults called with source = %q, want empty", gotSource)
		}

		var resp []flakyResultDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if len(resp) != 1 {
			t.Fatalf("response length = %d, want 1", len(resp))
		}
		got := resp[0]
		if got.TestName != "checkout works" || got.RunsConsidered != 2 || got.PassRate != 0.5 || !got.IsFlaky || got.AvgDurationMS != 75 {
			t.Errorf("response[0] = %+v, want {TestName:checkout works RunsConsidered:2 PassRate:0.5 IsFlaky:true AvgDurationMS:75}", got)
		}
	})

	t.Run("source query param is passed through", func(t *testing.T) {
		var gotSource string
		store := &fakeStore{
			listRecentResultsFunc: func(ctx context.Context, source string) ([]run.RunResult, error) {
				gotSource = source
				return []run.RunResult{}, nil
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler, testAPIKey)

		req := httptest.NewRequest(http.MethodGet, "/flaky?source=github:example/repo", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if gotSource != "github:example/repo" {
			t.Errorf("ListRecentResults called with source = %q, want %q", gotSource, "github:example/repo")
		}
	})

	t.Run("n query param overrides the default", func(t *testing.T) {
		// n=5 should drop the 2 oldest (failing) rows below. if truncation
		// didn't work, the stats would reflect all 7 instead.
		store := &fakeStore{
			listRecentResultsFunc: func(ctx context.Context, source string) ([]run.RunResult, error) {
				return []run.RunResult{
					{TestName: "ran many times", Status: run.StatusPass, DurationMS: 100, SubmittedAt: time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)},
					{TestName: "ran many times", Status: run.StatusPass, DurationMS: 100, SubmittedAt: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)},
					{TestName: "ran many times", Status: run.StatusPass, DurationMS: 100, SubmittedAt: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)},
					{TestName: "ran many times", Status: run.StatusPass, DurationMS: 100, SubmittedAt: time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)},
					{TestName: "ran many times", Status: run.StatusPass, DurationMS: 100, SubmittedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
					{TestName: "ran many times", Status: run.StatusFail, DurationMS: 1000, SubmittedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
					{TestName: "ran many times", Status: run.StatusFail, DurationMS: 1000, SubmittedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
				}, nil
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler, testAPIKey)

		req := httptest.NewRequest(http.MethodGet, "/flaky?n=5", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp []flakyResultDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if len(resp) != 1 {
			t.Fatalf("response length = %d, want 1", len(resp))
		}
		got := resp[0]
		if got.RunsConsidered != 5 || got.PassRate != 1 || got.IsFlaky || got.AvgDurationMS != 100 {
			t.Errorf("response[0] = %+v, want {RunsConsidered:5 PassRate:1 IsFlaky:false AvgDurationMS:100}", got)
		}
	})

	invalidNCases := []string{"0", "-3", "abc"}
	for _, n := range invalidNCases {
		t.Run("n="+n+" returns 400", func(t *testing.T) {
			storeCalled := false
			store := &fakeStore{
				listRecentResultsFunc: func(ctx context.Context, source string) ([]run.RunResult, error) {
					storeCalled = true
					return []run.RunResult{}, nil
				},
			}
			handler := &RunsHandler{Store: store}
			router := NewRouter(handler, testAPIKey)

			req := httptest.NewRequest(http.MethodGet, "/flaky?n="+n, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
			if storeCalled {
				t.Error("expected Store.ListRecentResults NOT to be called")
			}
		})
	}

	t.Run("store error returns 500", func(t *testing.T) {
		store := &fakeStore{
			listRecentResultsFunc: func(ctx context.Context, source string) ([]run.RunResult, error) {
				return nil, errors.New("db connection lost")
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler, testAPIKey)

		req := httptest.NewRequest(http.MethodGet, "/flaky", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusInternalServerError, rec.Body.String())
		}
	})

	t.Run("empty result set returns 200 with a JSON array, not null", func(t *testing.T) {
		store := &fakeStore{
			listRecentResultsFunc: func(ctx context.Context, source string) ([]run.RunResult, error) {
				return []run.RunResult{}, nil
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler, testAPIKey)

		req := httptest.NewRequest(http.MethodGet, "/flaky", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		// Raw body check, not decode-then-len: "null" also decodes to len 0.
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Errorf("response body = %q, want %q", got, "[]")
		}
	})
}
