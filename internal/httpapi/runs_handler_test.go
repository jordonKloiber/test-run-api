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

// fakeStore is a test double for run.Store — each test case plugs in exactly
// the behavior it needs via these function fields, without a real database.
type fakeStore struct {
	createRunFunc   func(ctx context.Context, r run.Run) (run.Run, error)
	getRunFunc      func(ctx context.Context, id int64) (run.Run, error)
	createRunCalled bool
}

func (f *fakeStore) CreateRun(ctx context.Context, r run.Run) (run.Run, error) {
	f.createRunCalled = true
	return f.createRunFunc(ctx, r)
}

func (f *fakeStore) GetRun(ctx context.Context, id int64) (run.Run, error) {
	return f.getRunFunc(ctx, id)
}

// TestCreateRun covers POST /runs: a valid payload persists (via the fake)
// and returns 201 with the expected response shape; an invalid payload and a
// malformed JSON body both return 400 without ever touching the store; and a
// store failure (independent of validation) returns 500.
func TestCreateRun(t *testing.T) {
	t.Run("valid payload returns 201", func(t *testing.T) {
		store := &fakeStore{
			createRunFunc: func(ctx context.Context, r run.Run) (run.Run, error) {
				r.ID = 42
				r.SubmittedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				return r, nil
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler)

		body := `{
			"suite_name": "checkout-e2e",
			"source": "github:example/repo",
			"results": [
				{"test_name": "test-a", "status": "pass", "duration_ms": 100}
			]
		}`
		req := httptest.NewRequest(http.MethodPost, "/runs", strings.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}
		if !store.createRunCalled {
			t.Error("expected Store.CreateRun to be called")
		}

		var resp createRunResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if resp.ID != 42 {
			t.Errorf("response ID = %d, want 42", resp.ID)
		}
		if resp.ResultCount != 1 {
			t.Errorf("response ResultCount = %d, want 1", resp.ResultCount)
		}
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		// createRunFunc is deliberately nil — if ValidateRun ever fails to
		// short-circuit before the store is touched, calling it would panic
		// loudly instead of silently returning a fake success.
		store := &fakeStore{}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler)

		body := `{
			"suite_name": "",
			"source": "github:example/repo",
			"results": [
				{"test_name": "test-a", "status": "pass", "duration_ms": 100}
			]
		}`
		req := httptest.NewRequest(http.MethodPost, "/runs", strings.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		if store.createRunCalled {
			t.Error("expected Store.CreateRun NOT to be called")
		}
	})

	t.Run("malformed JSON body returns 400", func(t *testing.T) {
		// createRunFunc is deliberately nil, same reasoning as above — a
		// malformed body should never reach the store at all.
		store := &fakeStore{}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler)

		body := `{not json`
		req := httptest.NewRequest(http.MethodPost, "/runs", strings.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		if store.createRunCalled {
			t.Error("expected Store.CreateRun NOT to be called")
		}
	})

	t.Run("store error returns 500", func(t *testing.T) {
		store := &fakeStore{
			createRunFunc: func(ctx context.Context, r run.Run) (run.Run, error) {
				return run.Run{}, errors.New("db connection lost")
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler)

		body := `{
			"suite_name": "checkout-e2e",
			"source": "github:example/repo",
			"results": [
				{"test_name": "test-a", "status": "pass", "duration_ms": 100}
			]
		}`
		req := httptest.NewRequest(http.MethodPost, "/runs", strings.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusInternalServerError, rec.Body.String())
		}
	})
}

// TestGetRun covers GET /runs/{id}: an existing id returns 200 with the
// run's data from the fake, including its nested results' actual field
// values (not just a count); a missing id (run.ErrNotFound from the fake)
// returns 404; any other store error returns 500, confirming that branch is
// distinct from the 404 path rather than everything falling through to one;
// and a non-integer id in the path returns 400 without touching the store.
func TestGetRun(t *testing.T) {
	t.Run("existing id returns 200", func(t *testing.T) {
		store := &fakeStore{
			getRunFunc: func(ctx context.Context, id int64) (run.Run, error) {
				return run.Run{
					ID:          id,
					SuiteName:   "checkout-e2e",
					Source:      "github:example/repo",
					SubmittedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Results: []run.TestResult{
						{TestName: "test-a", Status: run.StatusPass, DurationMS: 100},
					},
				}, nil
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/runs/42", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp getRunResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if resp.ID != 42 {
			t.Errorf("response ID = %d, want 42", resp.ID)
		}
		if len(resp.Results) != 1 {
			t.Fatalf("response Results length = %d, want 1", len(resp.Results))
		}
		got := resp.Results[0]
		if got.TestName != "test-a" || got.Status != "pass" || got.DurationMS != 100 {
			t.Errorf("response Results[0] = %+v, want {TestName:test-a Status:pass DurationMS:100}", got)
		}
	})

	t.Run("missing id returns 404", func(t *testing.T) {
		store := &fakeStore{
			getRunFunc: func(ctx context.Context, id int64) (run.Run, error) {
				return run.Run{}, run.ErrNotFound
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/runs/999", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})

	t.Run("store error returns 500", func(t *testing.T) {
		store := &fakeStore{
			getRunFunc: func(ctx context.Context, id int64) (run.Run, error) {
				return run.Run{}, errors.New("db connection lost")
			},
		}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/runs/42", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusInternalServerError, rec.Body.String())
		}
	})

	t.Run("non-integer id returns 400", func(t *testing.T) {
		// getRunFunc is deliberately nil — an unparseable id should be
		// rejected before the store is ever touched.
		store := &fakeStore{}
		handler := &RunsHandler{Store: store}
		router := NewRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/runs/abc", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d. body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}
