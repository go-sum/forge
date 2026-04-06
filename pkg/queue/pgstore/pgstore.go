// Package pgstore implements queue.Store using PostgreSQL with pgx/v5.
// It uses FOR UPDATE SKIP LOCKED for concurrent-safe job claiming.
package pgstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-sum/queue"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds the PostgreSQL store configuration.
type Config struct {
	Pool *pgxpool.Pool

	// ReapThreshold is the duration after which a running job is considered
	// stuck and eligible for reaping. Defaults to 5 minutes.
	ReapThreshold int // seconds; 0 = 300 (5 minutes)
}

// PgStore implements queue.Store backed by PostgreSQL.
type PgStore struct {
	pool          *pgxpool.Pool
	reapThreshold time.Duration
}

// Compile-time interface check.
var _ queue.Store = (*PgStore)(nil)

// New creates a PgStore. The pool is externally managed and not closed by Close().
func New(cfg Config) *PgStore {
	reap := time.Duration(cfg.ReapThreshold) * time.Second
	if reap <= 0 {
		reap = 5 * time.Minute
	}
	return &PgStore{
		pool:          cfg.Pool,
		reapThreshold: reap,
	}
}

// Install creates the queue_jobs table and indexes idempotently.
func (s *PgStore) Install(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("pgstore: install schema: %w", err)
	}
	return nil
}

// Enqueue inserts a new job and sets its ID from the RETURNING clause.
func (s *PgStore) Enqueue(ctx context.Context, job *queue.Job) error {
	row := s.pool.QueryRow(ctx, enqueueSQL,
		job.Queue,
		int(job.Priority),
		job.Payload,
		string(job.Status),
		job.MaxAttempts,
		job.RunAt,
	)
	return row.Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)
}

// Dequeue atomically claims the highest-priority pending job from the given
// queues. Returns queue.ErrJobNotFound when no work is available.
func (s *PgStore) Dequeue(ctx context.Context, queues []string) (*queue.Job, error) {
	row := s.pool.QueryRow(ctx, dequeueSQL, queues)

	var job queue.Job
	var priority int
	var status string

	err := row.Scan(
		&job.ID, &job.Queue, &priority,
		&job.Payload, &status, &job.Attempts,
		&job.MaxAttempts, &job.LastError, &job.RunAt,
		&job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, queue.ErrJobNotFound
		}
		return nil, fmt.Errorf("pgstore: dequeue: %w", err)
	}

	job.Priority = queue.Priority(priority)
	job.Status = queue.JobStatus(status)
	return &job, nil
}

// Complete marks a job as completed.
func (s *PgStore) Complete(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, completeSQL, id)
	if err != nil {
		return fmt.Errorf("pgstore: complete: %w", err)
	}
	return nil
}

// Fail records a failure. If the job has retries remaining it is rescheduled
// after retryAfter; otherwise it is marked dead.
func (s *PgStore) Fail(ctx context.Context, id string, errMsg string, retryAfter time.Duration) error {
	interval := fmt.Sprintf("%d seconds", int(retryAfter.Seconds()))
	_, err := s.pool.Exec(ctx, failSQL, id, errMsg, interval)
	if err != nil {
		return fmt.Errorf("pgstore: fail: %w", err)
	}
	return nil
}

// Reap reclaims jobs stuck in running state beyond the reap threshold.
func (s *PgStore) Reap(ctx context.Context, queues []string) (int, error) {
	interval := fmt.Sprintf("%d seconds", int(s.reapThreshold.Seconds()))
	tag, err := s.pool.Exec(ctx, reapSQL, queues, interval)
	if err != nil {
		return 0, fmt.Errorf("pgstore: reap: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// Ping verifies database connectivity.
func (s *PgStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Close is a no-op because the pool is externally managed.
func (s *PgStore) Close() error {
	return nil
}
