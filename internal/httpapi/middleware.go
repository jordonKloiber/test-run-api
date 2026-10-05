package httpapi

import (
	"crypto/subtle"
	"net/http"
)

// requireAPIKey wraps a handler so requests missing or presenting a wrong
// X-API-Key header get a 401 with the standard {"error": "..."} body,
// instead of reaching next. Comparison is constant-time to avoid leaking
// the key via response-time side channels.
//
// An empty apiKey always rejects, regardless of what's sent — config.Load
// already refuses to start with an empty API_KEY, but this guard means
// requireAPIKey doesn't rely solely on that upstream guarantee: without it,
// an empty configured key would match an absent header (both compare as
// empty), silently disabling auth entirely.
func requireAPIKey(apiKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		got := req.Header.Get("X-API-Key")
		if apiKey == "" || subtle.ConstantTimeCompare([]byte(got), []byte(apiKey)) != 1 {
			writeError(w, http.StatusUnauthorized, "invalid or missing API key")
			return
		}
		next(w, req)
	}
}
