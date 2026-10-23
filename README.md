# test-run-api

A small Go REST API that ingests structured test-run results from CI pipelines (pass/fail/duration per test) and tracks flakiness over time. It doesn't run any tests itself; it's the service other pipelines report their results *into*.

Live at `https://test-run-api-796881713635.us-central1.run.app`, currently receiving real results from [jordonkloiber.com](https://github.com/jordonKloiber/jordonkloiber.com)'s Playwright suite on every CI run.

## Stack

Go 1.27 (stdlib `net/http`, no framework) · PostgreSQL via [pgx](https://github.com/jackc/pgx), hosted on [Neon](https://neon.tech) · [golang-migrate](https://github.com/golang-migrate/migrate) for schema migrations · Docker + Google Cloud Run for deployment · GitHub Actions for CI and auto-deploy (via Workload Identity Federation, no stored key)

## Why this exists

Built to get hands-on with Go and Postgres on a real (but small) service, with schema design, validation, and test coverage you can  read.

Partly by hand with assistance from Claude and all tests written by hand in order to practice my Go. This is meant to be a portfolio project to showcase full-stack skills with REST APIs and Go. Ultimately all major architectural decisions made by me, so feel free to ask any questions you might have. I'd be happy to explain how it works and decision rationale in an interview!

## Quickstart

```bash
git clone https://github.com/jordonKloiber/test-run-api.git
cd test-run-api
cp .env.example .env   # fill in DATABASE_URL (a Postgres instance) and API_KEY (any random string)

make migrate-up        # applies the schema
make run                # starts the server on :8080 (override with PORT)
```

```bash
make test               # unit tests, no database needed
make test-integration   # real HTTP + real Postgres, point DATABASE_URL at a disposable test database first
```

## Endpoints

| Route | Auth | Does |
|---|---|---|
| `POST /runs` | `X-API-Key` required | Ingests one suite run and all of its test results in a single transaction |
| `GET /runs/{id}` | public | Returns one stored run with its results |
| `GET /flaky` | public | Pass rate, flakiness, and mean duration per test over the last `n` runs, optionally filtered by `source` |

Only the write endpoint is gated: the data is test history, not secrets. Writes are protected so nobody can poison the flakiness record; reads stay open so dashboards and a quick `curl` need no credentials.

### Example: reporting a run

```bash
curl -X POST https://test-run-api-796881713635.us-central1.run.app/runs \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <your key>" \
  -d '{
    "suite_name": "my-app-e2e",
    "source": "my-app-ci",
    "results": [
      { "test_name": "checkout flow completes", "status": "pass", "duration_ms": 842 },
      { "test_name": "search returns results",   "status": "fail", "duration_ms": 1203, "error_message": "timed out waiting for selector" }
    ]
  }'
```

`status` is one of `pass`, `fail`, `skip`. `error_message` is optional.

### Example: checking for flaky tests

```bash
curl "https://test-run-api-796881713635.us-central1.run.app/flaky?source=my-app-ci&n=10"
```

Returns, per test, how many of its last `n` runs were considered, the pass rate, whether it's flagged flaky, and the average duration.

## CI and deployment

GitHub Actions gates every PR and push to `main`: build, `go vet`, `gofmt` check, and the unit test suite. A separate workflow deploys automatically on every merge to `main`: it builds the Docker image, pushes it to Artifact Registry tagged with the commit SHA, and redeploys Cloud Run. Authentication to GCP uses Workload Identity Federation rather than a stored service-account key.

## Status

Core API, auth, flaky-detection, CI, and auto-deploy are done and live. One real external consumer (jordonkloiber.com) reporting production data. Known gaps: no pagination on `/flaky`, a single static API key with no rotation, no rate limiting.
