package queue

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// runWorker polls the store for pending jobs and executes them. It runs until
// ctx is cancelled, then returns after any in-flight job completes.
func (c *Client) runWorker(ctx context.Context, queues []string, workerID int) {
	defer c.wg.Done()

	poll := time.Duration(c.cfg.PollInterval) * time.Second
	timer := time.NewTimer(0) // fire immediately on first iteration
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		job, err := c.store.Dequeue(ctx, queues)
		if err != nil {
			if errors.Is(err, ErrJobNotFound) {
				timer.Reset(poll)
				continue
			}
			if ctx.Err() != nil {
				return
			}
			slog.ErrorContext(ctx, "queue: dequeue error",
				"queues", queues, "worker", workerID, "error", err)
			timer.Reset(poll)
			continue
		}

		c.executeJob(ctx, job)
		timer.Reset(0) // immediately check for more work
	}
}

// executeJob runs the registered handler for a job with a timeout-bounded
// context. On success it marks the job complete; on failure it records the
// error and schedules a retry (or marks dead).
func (c *Client) executeJob(ctx context.Context, job *Job) {
	qcfg, ok := c.queues[job.Queue]
	if !ok {
		slog.ErrorContext(ctx, "queue: no config for queue", "queue", job.Queue, "job_id", job.ID)
		return
	}

	handler, ok := c.handlers[job.Queue]
	if !ok {
		slog.ErrorContext(ctx, "queue: no handler for queue", "queue", job.Queue, "job_id", job.ID)
		return
	}

	timeout := time.Duration(qcfg.Timeout) * time.Second
	jobCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := c.safeExecute(jobCtx, handler, *job)

	// Use a background context for store operations so they complete even
	// if the parent context was cancelled during shutdown.
	storeCtx, storeCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer storeCancel()

	if err == nil {
		if completeErr := c.store.Complete(storeCtx, job.ID); completeErr != nil {
			slog.ErrorContext(ctx, "queue: complete failed",
				"job_id", job.ID, "error", completeErr)
		}
		return
	}

	retryAfter := c.computeBackoff(qcfg, job.Attempts)
	if failErr := c.store.Fail(storeCtx, job.ID, err.Error(), retryAfter); failErr != nil {
		slog.ErrorContext(ctx, "queue: fail record failed",
			"job_id", job.ID, "error", failErr)
	}
}

// safeExecute invokes the handler with panic recovery.
func (c *Client) safeExecute(ctx context.Context, handler HandlerFunc, job Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("queue: handler panic: %v", r)
		}
	}()
	return handler(ctx, job)
}

// computeBackoff returns the retry delay using exponential backoff:
// base * 2^(attempt-1).
func (c *Client) computeBackoff(qcfg QueueConfig, attempts int) time.Duration {
	base := time.Duration(qcfg.Backoff) * time.Second
	shift := attempts - 1
	if shift < 0 {
		shift = 0
	}
	if shift > 20 {
		shift = 20 // cap to avoid overflow
	}
	return base * (1 << shift)
}

// runReaper periodically reclaims jobs stuck in running state.
func (c *Client) runReaper(ctx context.Context) {
	defer c.wg.Done()

	interval := time.Duration(c.cfg.PollInterval*10) * time.Second
	if interval < 10*time.Second {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	allQueues := make([]string, 0, len(c.queues))
	for name := range c.queues {
		allQueues = append(allQueues, name)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reaped, err := c.store.Reap(ctx, allQueues)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				slog.ErrorContext(ctx, "queue: reap error", "error", err)
				continue
			}
			if reaped > 0 {
				slog.InfoContext(ctx, "queue: reaped stale jobs", "count", reaped)
			}
		}
	}
}
