# test-run-api

A small Go REST API that ingests structured test-run results from CI pipelines (pass/fail/duration per test) and tracks flakiness over time. It doesn't test anything itself — it's the service other pipelines report their results *into*.

**Status: actively in development.** This is a personal portfolio project, not a finished product.

## Stack

Go 1.27 (stdlib `net/http`, no framework) · PostgreSQL via [pgx](https://github.com/jackc/pgx) · [golang-migrate](https://github.com/golang-migrate/migrate) for schema migrations

## Why this exists

Built to get hands-on with Go and Postgres on a real (but small) service. With schema design, validation, and test coverage you can actually read. Writing muost of it by hand with assistance from AI, rather than written by AI, in order to practice my Go.
