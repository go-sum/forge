// Package pgstore implements queue.Store using PostgreSQL with pgx/v5.
// It uses FOR UPDATE SKIP LOCKED for concurrent-safe job claiming.
package pgstore

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/go-sum/queue"
	queuedb "github.com/go-sum/queue/pgstore/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compile-time interface check.
var _ queue.Store = (*PgStore)(nil)

//go:embed sql/schema.sql
var createTableSQL string

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
	queries       *queuedb.Queries
	reapThreshold time.Duration
}

// New creates a PgStore. The pool is externally managed and not closed by Close().
func New(cfg Config) *PgStore {
	reap := time.Duration(cfg.ReapThreshold) * time.Second
	if reap <= 0 {
		reap = 5 * time.Minute
	}
	return &PgStore{
		pool:          cfg.Pool,
		queries:       queuedb.New(cfg.Pool),
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
	row, err := s.queries.Enqueue(ctx, queuedb.EnqueueParams{
		Queue:       job.Queue,
		Priority:    int32(job.Priority),
		Payload:     job.Payload,
		Status:      string(job.Status),
		MaxAttempts: int32(job.MaxAttempts),
		RunAt:       job.RunAt,
	})
	if err != nil {
		return fmt.Errorf("pgstore: enqueue: %w", err)
	}
	job.ID = row.ID.String()
	job.CreatedAt = row.CreatedAt
	job.UpdatedAt = row.UpdatedAt
	return nil
}

// Dequeue atomically claims the highest-priority pending job from the given
// queues. Returns queue.ErrJobNotFound when no work is available.
func (s *PgStore) Dequeue(ctx context.Context, queues []string) (*queue.Job, error) {
	row, err := s.queries.Dequeue(ctx, queues)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, queue.ErrJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("pgstore: dequeue: %w", err)
	}
	return toJobModel(row), nil
}

// Complete marks a job as completed.
func (s *PgStore) Complete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("pgstore: complete: invalid id: %w", err)
	}
	if err := s.queries.Complete(ctx, uid); err != nil {
		return fmt.Errorf("pgstore: complete: %w", err)
	}
	return nil
}

// Fail records a failure. If the job has retries remaining it is rescheduled
// after retryAfter; otherwise it is marked dead.
func (s *PgStore) Fail(ctx context.Context, id string, errMsg string, retryAfter time.Duration) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("pgstore: fail: invalid id: %w", err)
	}
	if err := s.queries.Fail(ctx, queuedb.FailParams{
		ID:        uid,
		LastError: errMsg,
		Column3:   durationToInterval(retryAfter),
	}); err != nil {
		return fmt.Errorf("pgstore: fail: %w", err)
	}
	return nil
}

// Reap reclaims jobs stuck in running state beyond the reap threshold.
func (s *PgStore) Reap(ctx context.Context, queues []string) (int, error) {
	n, err := s.queries.Reap(ctx, queuedb.ReapParams{
		Column1: queues,
		Column2: durationToInterval(s.reapThreshold),
	})
	if err != nil {
		return 0, fmt.Errorf("pgstore: reap: %w", err)
	}
	return int(n), nil
}

// Ping verifies database connectivity.
func (s *PgStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Close is a no-op because the pool is externally managed.
func (s *PgStore) Close() error {
	return nil
}

// toJobModel converts a sqlc-generated db.QueueJob to the domain model.
func toJobModel(r queuedb.QueueJob) *queue.Job {
	return &queue.Job{
		ID:          r.ID.String(),
		Queue:       r.Queue,
		Priority:    queue.Priority(r.Priority),
		Payload:     r.Payload,
		Status:      queue.JobStatus(r.Status),
		Attempts:    int(r.Attempts),
		MaxAttempts: int(r.MaxAttempts),
		LastError:   r.LastError,
		RunAt:       r.RunAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// durationToInterval converts a time.Duration to a pgtype.Interval for use
// in parameterized queries that accept a PostgreSQL interval type.
func durationToInterval(d time.Duration) pgtype.Interval {
	return pgtype.Interval{
		Microseconds: d.Microseconds(),
		Valid:        true,
	}
}
