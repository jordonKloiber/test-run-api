package httpapi

import "net/http"

func NewRouter(h *RunsHandler, apiKey string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /runs", requireAPIKey(apiKey, h.CreateRun))
	mux.HandleFunc("GET /runs/{id}", h.GetRun)
	mux.HandleFunc("GET /flaky", h.ListFlaky)
	return mux
}
