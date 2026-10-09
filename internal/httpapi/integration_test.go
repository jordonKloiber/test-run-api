//go:build integration

package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jordonKloiber/test-run-api/internal/run"
)

// TestIntegration exercises the real HTTP server, real router + middleware,
// and a real Postgres database — the dedicated integration-test database,
// never the app's real .env DATABASE_URL. Run via:
//
//	DATABASE_URL="<test-db-connection-string>" go test -tags=integration ./...
func TestIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Fatal("DATABASE_URL must be set to the dedicated integration-test database")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	store := run.NewPgStore(pool)
	handler := &RunsHandler{Store: store}
	router := NewRouter(handler, testAPIKey)

	server := httptest.NewServer(router)
	defer server.Close()
	client := server.Client()

	truncate := func(t *testing.T) {
		if _, err := pool.Exec(ctx, "TRUNCATE runs CASCADE"); err != nil {
			t.Fatalf("failed to truncate test database: %v", err)
		}
	}

	countRuns := func(t *testing.T) int {
		var n int
		if err := pool.QueryRow(ctx, "SELECT count(*) FROM runs").Scan(&n); err != nil {
			t.Fatalf("failed to count rows: %v", err)
		}
		return n
	}

	validRunBody := `{
		"suite_name": "integration-suite",
		"source": "integration-test",
		"results": [
			{"test_name": "test-a", "status": "pass", "duration_ms": 100}
		]
	}`

	t.Run("create and get a run round-trips correctly", func(t *testing.T) {
		truncate(t)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/runs", strings.NewReader(validRunBody))
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}
		req.Header.Set("X-API-Key", testAPIKey)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST /runs failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}

		var created createRunResponse
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		getResp, err := client.Get(server.URL + "/runs/" + strconv.FormatInt(created.ID, 10))
		if err != nil {
			t.Fatalf("GET /runs/{id} failed: %v", err)
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", getResp.StatusCode, http.StatusOK)
		}

		var fetched getRunResponse
		if err := json.NewDecoder(getResp.Body).Decode(&fetched); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if fetched.SuiteName != "integration-suite" || len(fetched.Results) != 1 {
			t.Errorf("fetched run = %+v, want suite_name=integration-suite with 1 result", fetched)
		}
	})

	t.Run("create fails without API Key", func(t *testing.T) {
		truncate(t)

		before := countRuns(t)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/runs", strings.NewReader(validRunBody))
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST /runs failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}

		after := countRuns(t)
		if after != before {
			t.Errorf("row count = %d, want %d (unauthorized request should not create anything)", after, before)
		}
	})

	t.Run("create fails with wrong API Key", func(t *testing.T) {
		truncate(t)

		before := countRuns(t)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/runs", strings.NewReader(validRunBody))
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}
		req.Header.Set("X-API-Key", "wrong-key")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST /runs failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}

		after := countRuns(t)
		if after != before {
			t.Errorf("row count = %d, want %d (unauthorized request should not create anything)", after, before)
		}
	})

	t.Run("create fails with malformed json", func(t *testing.T) {
		truncate(t)

		before := countRuns(t)

		invalidRunBodyMalformed := `{
		"suite_name": "integration-suite",
		"source": "integration-test",
		"results": [
			{"test_name": "test-a", "status": "pass", "duration_ms": 100}
		]`

		req, err := http.NewRequest(http.MethodPost, server.URL+"/runs", strings.NewReader(invalidRunBodyMalformed))
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}
		req.Header.Set("X-API-Key", testAPIKey)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST /runs failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}

		after := countRuns(t)
		if after != before {
			t.Errorf("row count = %d, want %d (malformed request should not create anything)", after, before)
		}
	})

	t.Run("create fails with empty json", func(t *testing.T) {
		truncate(t)

		before := countRuns(t)

		invalidRunBodyEmpty := `{}`

		req, err := http.NewRequest(http.MethodPost, server.URL+"/runs", strings.NewReader(invalidRunBodyEmpty))
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}
		req.Header.Set("X-API-Key", testAPIKey)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST /runs failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}

		after := countRuns(t)
		if after != before {
			t.Errorf("row count = %d, want %d (empty request should not create anything)", after, before)
		}
	})

	t.Run("create fails with json missing required field", func(t *testing.T) {
		truncate(t)

		before := countRuns(t)

		invalidRunBodyMissing := `{
			"source": "integration-test",
			"results": [
				{"test_name": "test-a", "status": "pass", "duration_ms": 100}
			]
		}`

		req, err := http.NewRequest(http.MethodPost, server.URL+"/runs", strings.NewReader(invalidRunBodyMissing))
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}
		req.Header.Set("X-API-Key", testAPIKey)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST /runs failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}

		after := countRuns(t)
		if after != before {
			t.Errorf("row count = %d, want %d (request with missing required field should not create anything)", after, before)
		}
	})

	t.Run("get runs fails with an integer id that was never created", func(t *testing.T) {
		truncate(t)

		getResp, err := client.Get(server.URL + "/runs/" + "-1")
		if err != nil {
			t.Fatalf("GET /runs/-1 failed: %v", err)
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", getResp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("get runs fails with a non-integer id", func(t *testing.T) {
		truncate(t)

		getResp, err := client.Get(server.URL + "/runs/" + "n")
		if err != nil {
			t.Fatalf("GET /runs/-1 failed: %v", err)
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", getResp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("get flaky succeeds without API Key", func(t *testing.T) {
		truncate(t)

		req, err := http.NewRequest(http.MethodGet, server.URL+"/flaky", nil)
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("GET /flaky failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("get flaky results are correct", func(t *testing.T) {
		truncate(t)

		seedRun1 := `{
			"suite_name": "flaky-suite",
			"source": "flaky-seed",
			"results": [
				{"test_name": "checkout flow", "status": "pass", "duration_ms": 120},
				{"test_name": "homepage loads", "status": "pass", "duration_ms": 80}
			]
		}`

		seedRun2 := `{
			"suite_name": "flaky-suite",
			"source": "flaky-seed",
			"results": [
				{"test_name": "checkout flow", "status": "fail", "duration_ms": 200, "error_message": "assertion failed: expected true"},
				{"test_name": "homepage loads", "status": "pass", "duration_ms": 85}
			]
		}`

		seedRun3 := `{
			"suite_name": "flaky-suite",
			"source": "flaky-seed",
			"results": [
				{"test_name": "checkout flow", "status": "pass", "duration_ms": 110},
				{"test_name": "homepage loads", "status": "pass", "duration_ms": 78}
			]
		}`

		seedRun4 := `{
			"suite_name": "flaky-suite",
			"source": "flaky-seed",
			"results": [
				{"test_name": "checkout flow", "status": "fail", "duration_ms": 210, "error_message": "timeout waiting for element"},
				{"test_name": "homepage loads", "status": "pass", "duration_ms": 82}
			]
		}`

		seedRuns := []string{seedRun1, seedRun2, seedRun3, seedRun4}

		for _, body := range seedRuns {
			req, err := http.NewRequest(http.MethodPost, server.URL+"/runs", strings.NewReader(body))
			if err != nil {
				t.Fatalf("failed to build request: %v", err)
			}
			req.Header.Set("X-API-Key", testAPIKey)

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("POST /runs failed: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
			}
		}
		getResp, err := client.Get(server.URL + "/flaky?source=flaky-seed&n=10")
		if err != nil {
			t.Fatalf("GET /flaky failed: %v", err)
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", getResp.StatusCode, http.StatusOK)
		}

		var resp []flakyResultDTO
		if err := json.NewDecoder(getResp.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("response length = %d, want 1", len(resp))
		}
		got := resp[0]
		if got.TestName != "checkout flow" || got.RunsConsidered != 4 || got.PassRate != 0.5 || got.IsFlaky != true || got.AvgDurationMS != 160 {
			t.Errorf("response[0] = %+v, want {TestName:checkout flow RunsConsidered:4 PassRate:0.5 IsFlaky:true AvgDurationMS:100}", got)
		}

		got = resp[1]
		if got.TestName != "homepage loads" || got.RunsConsidered != 4 || got.PassRate != 1 || got.IsFlaky != false || got.AvgDurationMS != 81.25 {
			t.Errorf("response[0] = %+v, want {TestName:homepage loads RunsConsidered:4 PassRate:0.5 IsFlaky:false AvgDurationMS:81.25}", got)
		}
	})

	t.Run("get flaky fails when integer param is 0", func(t *testing.T) {
		truncate(t)

		req, err := http.NewRequest(http.MethodGet, server.URL+"/flaky?n=0", nil)
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("GET /flaky failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})
}
