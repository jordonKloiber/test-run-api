package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jordonKloiber/test-run-api/internal/config"
	"github.com/jordonKloiber/test-run-api/internal/httpapi"
	"github.com/jordonKloiber/test-run-api/internal/run"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	store := run.NewPgStore(pool)
	handler := &httpapi.RunsHandler{Store: store}
	mux := httpapi.NewRouter(handler)

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
