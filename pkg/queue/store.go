package queue

import (
	"context"
	"time"
)

// Store is the persistence contract for the job queue. Implementations must be
// safe for concurrent use by multiple goroutines.
type Store interface {
	// Enqueue persists a new job. The store assigns the ID field.
	Enqueue(ctx context.Context, job *Job) error

	// Dequeue atomically claims the highest-priority pending job whose run_at
	// has passed from the given queues. Returns ErrJobNotFound when no work is
	// available. The returned job has status=running and attempts incremented.
	Dequeue(ctx context.Context, queues []string) (*Job, error)

	// Complete marks a running job as completed.
	Complete(ctx context.Context, id string) error

	// Fail records a job failure. If the job has remaining attempts the store
	// reschedules it to pending with run_at = now + retryAfter. Otherwise the
	// job is marked dead.
	Fail(ctx context.Context, id string, errMsg string, retryAfter time.Duration) error

	// Reap reclaims jobs stuck in running state beyond their expected timeout.
	// Returns the number of reclaimed jobs.
	Reap(ctx context.Context, queues []string) (int, error)

	// Ping verifies store connectivity.
	Ping(ctx context.Context) error

	// Close releases any resources held by the store.
	Close() error
}
