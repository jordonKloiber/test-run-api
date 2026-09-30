CREATE TABLE test_results (
    id            BIGSERIAL PRIMARY KEY,
    run_id        BIGINT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    test_name     TEXT NOT NULL,
    status        TEXT NOT NULL CHECK (status IN ('pass', 'fail', 'skip')),
    duration_ms   INTEGER NOT NULL CHECK (duration_ms >= 0),
    error_message TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_test_results_run_id ON test_results (run_id);
CREATE INDEX idx_test_results_test_name ON test_results (test_name);
