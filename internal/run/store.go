package run

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("run not found")

type Store interface {
	CreateRun(ctx context.Context, r Run) (Run, error)
	GetRun(ctx context.Context, id int64) (Run, error)
}

type pgStore struct {
	pool *pgxpool.Pool
}

func NewPgStore(pool *pgxpool.Pool) Store {
	return &pgStore{pool: pool}
}

func (s *pgStore) CreateRun(ctx context.Context, r Run) (Run, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Run{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO runs (suite_name, source) VALUES ($1, $2)
		 RETURNING id, submitted_at`,
		r.SuiteName, r.Source,
	).Scan(&r.ID, &r.SubmittedAt)
	if err != nil {
		return Run{}, err
	}

	for _, tr := range r.Results {
		_, err = tx.Exec(ctx,
			`INSERT INTO test_results (run_id, test_name, status, duration_ms, error_message)
			 VALUES ($1, $2, $3, $4, $5)`,
			r.ID, tr.TestName, string(tr.Status), tr.DurationMS, tr.ErrorMessage,
		)
		if err != nil {
			return Run{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Run{}, err
	}

	return r, nil
}

func (s *pgStore) GetRun(ctx context.Context, id int64) (Run, error) {
	var r Run
	err := s.pool.QueryRow(ctx,
		`SELECT id, suite_name, source, submitted_at FROM runs WHERE id = $1`,
		id,
	).Scan(&r.ID, &r.SuiteName, &r.Source, &r.SubmittedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Run{}, ErrNotFound
		}
		return Run{}, err
	}

	rows, err := s.pool.Query(ctx,
		`SELECT test_name, status, duration_ms, error_message
		 FROM test_results WHERE run_id = $1 ORDER BY id`,
		id,
	)
	if err != nil {
		return Run{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var tr TestResult
		var status string
		if err := rows.Scan(&tr.TestName, &status, &tr.DurationMS, &tr.ErrorMessage); err != nil {
			return Run{}, err
		}
		tr.Status = Status(status)
		r.Results = append(r.Results, tr)
	}
	if err := rows.Err(); err != nil {
		return Run{}, err
	}

	return r, nil
}
