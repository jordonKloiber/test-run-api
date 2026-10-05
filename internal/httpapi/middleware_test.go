package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRequireAPIKey covers the auth middleware in isolation — no real
// server needed, requireAPIKey's returned handler is just called directly.
func TestRequireAPIKey(t *testing.T) {
	const configuredKey = "test-secret-key"

	// newNextSpy builds a stand-in "next" handler that records whether it
	// was called, so tests can assert the real handler was (or wasn't)
	// reached — same idea as fakeStore's createRunCalled bool.
	newNextSpy := func() (next http.HandlerFunc, wasCalled *bool) {
		called := false
		next = func(w http.ResponseWriter, req *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}
		return next, &called
	}

	t.Run("valid key reaches next", func(t *testing.T) {
		next, called := newNextSpy()
		wrapped := requireAPIKey(configuredKey, next)

		req := httptest.NewRequest(http.MethodPost, "/runs", nil)
		req.Header.Set("X-API-Key", configuredKey)
		rec := httptest.NewRecorder()

		wrapped(rec, req)

		if !*called {
			t.Error("expected next to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
	t.Run("missing X-API-Key results in 401", func(t *testing.T) {
		next, called := newNextSpy()
		wrapped := requireAPIKey(configuredKey, next)

		req := httptest.NewRequest(http.MethodPost, "/runs", nil)
		//req.Header.Set("X-API-Key", configuredKey)
		rec := httptest.NewRecorder()

		wrapped(rec, req)

		if *called {
			t.Error("expected next NOT to be called")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}

		var body map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if body["error"] == "" {
			t.Error("expected a non-empty error message in the response body")
		}
	})
	t.Run("incorrect X-API-Key results in 401", func(t *testing.T) {
		next, called := newNextSpy()
		wrapped := requireAPIKey(configuredKey, next)

		req := httptest.NewRequest(http.MethodPost, "/runs", nil)
		req.Header.Set("X-API-Key", "wrong key")
		rec := httptest.NewRecorder()

		wrapped(rec, req)

		if *called {
			t.Error("expected next NOT to be called")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
	t.Run("empty X-API-Key results in 401", func(t *testing.T) {
		next, called := newNextSpy()
		wrapped := requireAPIKey(configuredKey, next)

		req := httptest.NewRequest(http.MethodPost, "/runs", nil)
		req.Header.Set("X-API-Key", "")
		rec := httptest.NewRecorder()

		wrapped(rec, req)

		if *called {
			t.Error("expected next NOT to be called")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
	t.Run("empty configured key never authenticates", func(t *testing.T) {
		next, called := newNextSpy()
		wrapped := requireAPIKey("", next) // bypasses config.Load's validation on purpose

		req := httptest.NewRequest(http.MethodPost, "/runs", nil)
		rec := httptest.NewRecorder()

		wrapped(rec, req)

		if *called {
			t.Error("expected next NOT to be called, even with an empty configured key")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}
