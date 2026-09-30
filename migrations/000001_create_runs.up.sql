CREATE TABLE runs (
    id           BIGSERIAL PRIMARY KEY,
    suite_name   TEXT NOT NULL,
    source       TEXT NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_runs_suite_name ON runs (suite_name);
CREATE INDEX idx_runs_source ON runs (source);
