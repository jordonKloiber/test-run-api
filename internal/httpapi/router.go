package httpapi

import "net/http"

func NewRouter(h *RunsHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /runs", h.CreateRun)
	mux.HandleFunc("GET /runs/{id}", h.GetRun)
	return mux
}
