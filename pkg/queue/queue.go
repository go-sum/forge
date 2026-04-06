// Package queue provides a background job queue with multiple named queues,
// configurable priority levels, and pluggable persistence via the Store interface.
package queue

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// Error sentinels.
var (
	ErrQueueNotFound = errors.New("queue: queue not found")
	ErrJobNotFound   = errors.New("queue: job not found")
	ErrAlreadyClosed = errors.New("queue: client already closed")
)

// Priority determines the processing order of jobs. Lower values are processed first.
type Priority int

const (
	PriorityCritical Priority = 0
	PriorityHigh     Priority = 10
	PriorityDefault  Priority = 20
	PriorityLow      Priority = 30
)

// JobStatus represents the lifecycle state of a job.
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusDead      JobStatus = "dead" // exhausted all retry attempts
)

// Job represents a unit of work persisted in the store.
type Job struct {
	ID          string
	Queue       string
	Priority    Priority
	Payload     json.RawMessage
	Status      JobStatus
	Attempts    int
	MaxAttempts int
	LastError   string
	RunAt       time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// QueueConfig defines the behavior of a single named queue.
type QueueConfig struct {
	Name        string
	Priority    Priority
	Workers     int // goroutines polling this queue; default 1
	MaxAttempts int // retry limit; default 3
	Timeout     int // seconds per attempt; default 30
	Backoff     int // seconds base for exponential backoff; default 5
}

// Config configures the queue client.
type Config struct {
	Queues       []QueueConfig
	PollInterval int // seconds between poll attempts; default 1
	ShutdownWait int // seconds to wait for in-flight jobs on stop; default 30
}

// HandlerFunc processes a single job. Return a non-nil error to trigger a retry.
type HandlerFunc func(ctx context.Context, job Job) error

// DispatchOptions controls how a job is enqueued.
type DispatchOptions struct {
	Queue    string
	Payload  json.RawMessage
	Priority Priority  // -1 uses the queue's default priority
	RunAt    time.Time // zero value means execute immediately
}
