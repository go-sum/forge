-- Queue job queries
-- Executed via pgx parameterized statements. The -- name: annotations follow
-- sqlc conventions for documentation, though these are not processed by sqlc.

-- name: Enqueue :one
INSERT INTO queue_jobs (queue, priority, payload, status, max_attempts, run_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at;

-- name: Dequeue :one
-- Atomically claims the highest-priority pending job from the given queues.
-- Uses FOR UPDATE SKIP LOCKED for concurrent-safe worker claiming.
WITH next AS (
    SELECT id FROM queue_jobs
    WHERE queue = ANY($1)
      AND status = 'pending'
      AND run_at <= NOW()
    ORDER BY priority ASC, run_at ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
UPDATE queue_jobs
SET status = 'running',
    attempts = attempts + 1,
    updated_at = NOW()
FROM next
WHERE queue_jobs.id = next.id
RETURNING queue_jobs.id, queue_jobs.queue, queue_jobs.priority,
          queue_jobs.payload, queue_jobs.status, queue_jobs.attempts,
          queue_jobs.max_attempts, queue_jobs.last_error, queue_jobs.run_at,
          queue_jobs.created_at, queue_jobs.updated_at;

-- name: Complete :exec
UPDATE queue_jobs
SET status = 'completed', updated_at = NOW()
WHERE id = $1;

-- name: Fail :exec
-- Reschedules a job for retry or marks it dead.
-- $1 = job ID, $2 = error message, $3 = retry_after interval (e.g. '10 seconds').
UPDATE queue_jobs
SET last_error  = $2,
    updated_at  = NOW(),
    status      = CASE WHEN attempts >= max_attempts THEN 'dead' ELSE 'pending' END,
    run_at      = CASE WHEN attempts >= max_attempts THEN run_at ELSE NOW() + $3::interval END
WHERE id = $1;

-- name: Reap :exec
-- Reclaims running jobs stuck beyond the stale threshold.
-- $1 = queue names, $2 = stale threshold interval (e.g. '300 seconds').
UPDATE queue_jobs
SET status = 'pending', updated_at = NOW()
WHERE id IN (
    SELECT id FROM queue_jobs
    WHERE queue = ANY($1)
      AND status = 'running'
      AND updated_at < NOW() - $2::interval
    FOR UPDATE SKIP LOCKED
);
