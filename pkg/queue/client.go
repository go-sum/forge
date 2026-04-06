package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Client manages queue registration, job dispatch, and worker lifecycle.
// When created with a nil store (sync mode), Dispatch executes registered
// handlers inline — no persistence, no workers, no polling. This makes the
// Client a universal dispatch point: consumers always call DispatchPayload
// regardless of whether async processing is enabled.
type Client struct {
	store    Store // nil = sync mode
	cfg      Config
	queues   map[string]QueueConfig // keyed by queue name
	handlers map[string]HandlerFunc // keyed by queue name

	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.RWMutex
	closed bool
}

// New creates a Client with the given store and configuration. Pass a nil
// store to create a synchronous client that executes handlers inline during
// Dispatch. Zero-value config fields receive sensible defaults.
func New(store Store, cfg Config) *Client {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 1
	}
	if cfg.ShutdownWait <= 0 {
		cfg.ShutdownWait = 30
	}

	queues := make(map[string]QueueConfig, len(cfg.Queues))
	for i := range cfg.Queues {
		q := &cfg.Queues[i]
		if q.Workers <= 0 {
			q.Workers = 1
		}
		if q.MaxAttempts <= 0 {
			q.MaxAttempts = 3
		}
		if q.Timeout <= 0 {
			q.Timeout = 30
		}
		if q.Backoff <= 0 {
			q.Backoff = 5
		}
		queues[q.Name] = *q
	}

	return &Client{
		store:    store,
		cfg:      cfg,
		queues:   queues,
		handlers: make(map[string]HandlerFunc, len(cfg.Queues)),
	}
}

// Async reports whether the client is in async mode (has a persistent store).
func (c *Client) Async() bool {
	return c.store != nil
}

// Register associates a handler with a named queue. The queue must exist in the
// Config passed to New. Panics if the queue name is unknown (programming error).
// Must be called before Start.
func (c *Client) Register(queue string, handler HandlerFunc) {
	if _, ok := c.queues[queue]; !ok {
		panic(fmt.Sprintf("queue: Register called with unknown queue %q", queue))
	}
	c.handlers[queue] = handler
}

// Dispatch enqueues a job for async processing, or executes the handler
// inline when the client is in sync mode (nil store).
func (c *Client) Dispatch(ctx context.Context, opts DispatchOptions) error {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()
	if closed {
		return ErrAlreadyClosed
	}

	qcfg, ok := c.queues[opts.Queue]
	if !ok {
		return ErrQueueNotFound
	}

	priority := opts.Priority
	if priority < 0 {
		priority = qcfg.Priority
	}

	runAt := opts.RunAt
	if runAt.IsZero() {
		runAt = time.Now()
	}

	job := &Job{
		Queue:       opts.Queue,
		Priority:    priority,
		Payload:     opts.Payload,
		Status:      StatusPending,
		MaxAttempts: qcfg.MaxAttempts,
		RunAt:       runAt,
	}

	// Sync mode: execute handler inline.
	if c.store == nil {
		return c.dispatchSync(ctx, qcfg, job)
	}

	return c.store.Enqueue(ctx, job)
}

// dispatchSync executes the registered handler inline for sync mode.
func (c *Client) dispatchSync(ctx context.Context, qcfg QueueConfig, job *Job) error {
	handler, ok := c.handlers[job.Queue]
	if !ok {
		return fmt.Errorf("queue: no handler registered for %q", job.Queue)
	}

	timeout := time.Duration(qcfg.Timeout) * time.Second
	jobCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.safeExecute(jobCtx, handler, *job)
}

// DispatchPayload is a convenience method that JSON-marshals payload and
// dispatches to the named queue using its default priority.
func (c *Client) DispatchPayload(ctx context.Context, queue string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("queue: marshal payload: %w", err)
	}
	return c.Dispatch(ctx, DispatchOptions{
		Queue:    queue,
		Payload:  data,
		Priority: -1,
	})
}

// Start launches worker goroutines for each registered queue and a reaper
// goroutine. It is non-blocking. In sync mode (nil store) this is a no-op.
// Call Stop to shut down workers gracefully.
func (c *Client) Start(ctx context.Context) {
	if c.store == nil {
		return // sync mode — no workers needed
	}

	ctx, c.cancel = context.WithCancel(ctx)

	for name, qcfg := range c.queues {
		if _, ok := c.handlers[name]; !ok {
			continue // queue configured but no handler registered
		}
		queues := []string{name}
		for i := range qcfg.Workers {
			c.wg.Add(1)
			go c.runWorker(ctx, queues, i)
		}
	}

	c.wg.Add(1)
	go c.runReaper(ctx)

	slog.Info("queue workers started", "queues", len(c.handlers))
}

// Stop signals all workers to cease polling and waits up to ShutdownWait
// seconds for in-flight jobs to complete. In sync mode this marks the
// client as closed (preventing further dispatches).
func (c *Client) Stop() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrAlreadyClosed
	}
	c.closed = true
	c.mu.Unlock()

	if c.store == nil {
		return nil // sync mode — nothing to shut down
	}

	if c.cancel != nil {
		c.cancel()
	}

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(time.Duration(c.cfg.ShutdownWait) * time.Second):
		return fmt.Errorf("queue: shutdown timed out after %ds", c.cfg.ShutdownWait)
	}
}
