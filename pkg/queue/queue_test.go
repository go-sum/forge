package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Fake Store
// ---------------------------------------------------------------------------

type failRecord struct {
	ID         string
	ErrMsg     string
	RetryAfter time.Duration
}

type fakeStore struct {
	mu sync.Mutex

	// Enqueue tracking
	jobs       []*Job
	enqueueErr error

	// Dequeue control: returns items from dequeued in order, then ErrJobNotFound.
	dequeued   []*Job
	dequeueIdx int
	dequeueErr error

	// Complete / Fail tracking
	completed []string
	failed    []failRecord

	// Reap tracking
	reapCount int
	reapErr   error

	pingErr  error
	closeErr error
}

func (f *fakeStore) Enqueue(_ context.Context, job *Job) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.enqueueErr != nil {
		return f.enqueueErr
	}
	job.ID = fmt.Sprintf("job-%d", len(f.jobs)+1)
	f.jobs = append(f.jobs, job)
	return nil
}

func (f *fakeStore) Dequeue(_ context.Context, _ []string) (*Job, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.dequeueErr != nil {
		return nil, f.dequeueErr
	}
	if f.dequeueIdx < len(f.dequeued) {
		j := f.dequeued[f.dequeueIdx]
		f.dequeueIdx++
		return j, nil
	}
	return nil, ErrJobNotFound
}

func (f *fakeStore) Complete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.completed = append(f.completed, id)
	return nil
}

func (f *fakeStore) Fail(_ context.Context, id string, errMsg string, retryAfter time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failed = append(f.failed, failRecord{ID: id, ErrMsg: errMsg, RetryAfter: retryAfter})
	return nil
}

func (f *fakeStore) Reap(_ context.Context, _ []string) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reapCount++
	return 0, f.reapErr
}

func (f *fakeStore) Ping(_ context.Context) error { return f.pingErr }
func (f *fakeStore) Close() error                 { return f.closeErr }

// snapshot helpers (call under test, after Stop)
func (f *fakeStore) getJobs() []*Job {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*Job, len(f.jobs))
	copy(out, f.jobs)
	return out
}

func (f *fakeStore) getCompleted() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.completed))
	copy(out, f.completed)
	return out
}

func (f *fakeStore) getFailed() []failRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]failRecord, len(f.failed))
	copy(out, f.failed)
	return out
}

// ---------------------------------------------------------------------------
// Helper: create a Client with a single "email" queue
// ---------------------------------------------------------------------------

func newTestClient(store *fakeStore) *Client {
	return New(store, Config{
		Queues: []QueueConfig{
			{Name: "email", Priority: PriorityDefault, Workers: 1, MaxAttempts: 3, Timeout: 5, Backoff: 5},
		},
		PollInterval: 1,
		ShutdownWait: 5,
	})
}

// ---------------------------------------------------------------------------
// Config Defaults
// ---------------------------------------------------------------------------

func TestConfigDefaults(t *testing.T) {
	store := &fakeStore{}
	c := New(store, Config{
		Queues: []QueueConfig{
			{Name: "test"},
		},
	})

	if c.cfg.PollInterval != 1 {
		t.Errorf("PollInterval: got %d, want 1", c.cfg.PollInterval)
	}
	if c.cfg.ShutdownWait != 30 {
		t.Errorf("ShutdownWait: got %d, want 30", c.cfg.ShutdownWait)
	}
	qcfg := c.queues["test"]
	if qcfg.Workers != 1 {
		t.Errorf("Workers: got %d, want 1", qcfg.Workers)
	}
	if qcfg.MaxAttempts != 3 {
		t.Errorf("MaxAttempts: got %d, want 3", qcfg.MaxAttempts)
	}
	if qcfg.Timeout != 30 {
		t.Errorf("Timeout: got %d, want 30", qcfg.Timeout)
	}
	if qcfg.Backoff != 5 {
		t.Errorf("Backoff: got %d, want 5", qcfg.Backoff)
	}
}

// ---------------------------------------------------------------------------
// Register Tests
// ---------------------------------------------------------------------------

func TestRegister_UnknownQueue(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unknown queue, got none")
		}
		want := `queue: Register called with unknown queue "nope"`
		if got := fmt.Sprint(r); got != want {
			t.Errorf("panic message: got %q, want %q", got, want)
		}
	}()

	c.Register("nope", func(ctx context.Context, job Job) error { return nil })
}

func TestRegister_KnownQueue(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)

	// Should not panic.
	c.Register("email", func(ctx context.Context, job Job) error { return nil })

	if _, ok := c.handlers["email"]; !ok {
		t.Fatal("expected handler to be registered for 'email'")
	}
}

// ---------------------------------------------------------------------------
// Dispatch Tests
// ---------------------------------------------------------------------------

func TestDispatch_Success(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)

	payload := json.RawMessage(`{"to":"user@example.com"}`)
	runAt := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	err := c.Dispatch(context.Background(), DispatchOptions{
		Queue:    "email",
		Payload:  payload,
		Priority: PriorityHigh,
		RunAt:    runAt,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jobs := store.getJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job enqueued, got %d", len(jobs))
	}

	job := jobs[0]
	if job.Queue != "email" {
		t.Errorf("Queue: got %q, want %q", job.Queue, "email")
	}
	if job.Priority != PriorityHigh {
		t.Errorf("Priority: got %d, want %d", job.Priority, PriorityHigh)
	}
	if string(job.Payload) != `{"to":"user@example.com"}` {
		t.Errorf("Payload: got %s, want %s", job.Payload, payload)
	}
	if job.Status != StatusPending {
		t.Errorf("Status: got %q, want %q", job.Status, StatusPending)
	}
	if job.MaxAttempts != 3 {
		t.Errorf("MaxAttempts: got %d, want 3", job.MaxAttempts)
	}
	if !job.RunAt.Equal(runAt) {
		t.Errorf("RunAt: got %v, want %v", job.RunAt, runAt)
	}
	if job.ID != "job-1" {
		t.Errorf("ID: got %q, want %q", job.ID, "job-1")
	}
}

func TestDispatch_UnknownQueue(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)

	err := c.Dispatch(context.Background(), DispatchOptions{Queue: "nonexistent"})
	if !errors.Is(err, ErrQueueNotFound) {
		t.Errorf("expected ErrQueueNotFound, got %v", err)
	}
}

func TestDispatch_ClosedClient(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)
	c.Register("email", func(ctx context.Context, job Job) error { return nil })

	// Start then Stop to set closed=true
	c.Start(context.Background())
	if err := c.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	err := c.Dispatch(context.Background(), DispatchOptions{Queue: "email"})
	if !errors.Is(err, ErrAlreadyClosed) {
		t.Errorf("expected ErrAlreadyClosed, got %v", err)
	}
}

func TestDispatch_DefaultPriority(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)

	err := c.Dispatch(context.Background(), DispatchOptions{
		Queue:    "email",
		Payload:  json.RawMessage(`{}`),
		Priority: -1, // use queue default
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jobs := store.getJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Priority != PriorityDefault {
		t.Errorf("Priority: got %d, want %d (queue default)", jobs[0].Priority, PriorityDefault)
	}
}

func TestDispatch_ZeroRunAt(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)

	before := time.Now()
	err := c.Dispatch(context.Background(), DispatchOptions{
		Queue:    "email",
		Payload:  json.RawMessage(`{}`),
		Priority: PriorityLow,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := time.Now()

	jobs := store.getJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].RunAt.Before(before) || jobs[0].RunAt.After(after) {
		t.Errorf("RunAt %v not between %v and %v", jobs[0].RunAt, before, after)
	}
}

func TestDispatch_EnqueueError(t *testing.T) {
	store := &fakeStore{enqueueErr: errors.New("db down")}
	c := newTestClient(store)

	err := c.Dispatch(context.Background(), DispatchOptions{
		Queue:   "email",
		Payload: json.RawMessage(`{}`),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "db down" {
		t.Errorf("error: got %q, want %q", err.Error(), "db down")
	}
}

// ---------------------------------------------------------------------------
// DispatchPayload Tests
// ---------------------------------------------------------------------------

func TestDispatchPayload_Success(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)

	type emailPayload struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
	}
	err := c.DispatchPayload(context.Background(), "email", emailPayload{
		To:      "test@example.com",
		Subject: "Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jobs := store.getJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	want := `{"to":"test@example.com","subject":"Hello"}`
	if string(jobs[0].Payload) != want {
		t.Errorf("Payload: got %s, want %s", jobs[0].Payload, want)
	}
	// DispatchPayload uses Priority -1, so queue default should apply.
	if jobs[0].Priority != PriorityDefault {
		t.Errorf("Priority: got %d, want %d", jobs[0].Priority, PriorityDefault)
	}
}

func TestDispatchPayload_MarshalError(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)

	// Channels cannot be marshaled to JSON.
	err := c.DispatchPayload(context.Background(), "email", make(chan int))
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
	if len(store.getJobs()) != 0 {
		t.Error("expected no jobs enqueued on marshal failure")
	}
}

func TestDispatchPayload_UnknownQueue(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)

	err := c.DispatchPayload(context.Background(), "unknown", "data")
	if !errors.Is(err, ErrQueueNotFound) {
		t.Errorf("expected ErrQueueNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Worker Tests
// ---------------------------------------------------------------------------

func TestWorker_ProcessesJob(t *testing.T) {
	var handlerCalled bool
	var receivedJob Job

	store := &fakeStore{
		dequeued: []*Job{
			{ID: "j1", Queue: "email", Payload: json.RawMessage(`{"ok":true}`), Attempts: 1, MaxAttempts: 3},
		},
	}

	c := New(store, Config{
		Queues:       []QueueConfig{{Name: "email", Priority: PriorityDefault, Workers: 1, MaxAttempts: 3, Timeout: 5, Backoff: 5}},
		PollInterval: 1,
		ShutdownWait: 5,
	})
	c.Register("email", func(ctx context.Context, job Job) error {
		handlerCalled = true
		receivedJob = job
		return nil
	})

	c.Start(context.Background())
	// Give the worker time to pick up the job.
	time.Sleep(200 * time.Millisecond)
	if err := c.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if !handlerCalled {
		t.Fatal("handler was not called")
	}
	if receivedJob.ID != "j1" {
		t.Errorf("job ID: got %q, want %q", receivedJob.ID, "j1")
	}
	if string(receivedJob.Payload) != `{"ok":true}` {
		t.Errorf("job Payload: got %s, want %s", receivedJob.Payload, `{"ok":true}`)
	}

	completed := store.getCompleted()
	if len(completed) != 1 {
		t.Fatalf("expected 1 completed job, got %d", len(completed))
	}
	if completed[0] != "j1" {
		t.Errorf("completed job ID: got %q, want %q", completed[0], "j1")
	}
}

func TestWorker_HandlerError(t *testing.T) {
	store := &fakeStore{
		dequeued: []*Job{
			{ID: "j2", Queue: "email", Attempts: 2, MaxAttempts: 3},
		},
	}

	c := New(store, Config{
		Queues:       []QueueConfig{{Name: "email", Priority: PriorityDefault, Workers: 1, MaxAttempts: 3, Timeout: 5, Backoff: 5}},
		PollInterval: 1,
		ShutdownWait: 5,
	})
	c.Register("email", func(ctx context.Context, job Job) error {
		return errors.New("send failed")
	})

	c.Start(context.Background())
	time.Sleep(200 * time.Millisecond)
	if err := c.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	failed := store.getFailed()
	if len(failed) != 1 {
		t.Fatalf("expected 1 failed record, got %d", len(failed))
	}
	if failed[0].ID != "j2" {
		t.Errorf("failed job ID: got %q, want %q", failed[0].ID, "j2")
	}
	if failed[0].ErrMsg != "send failed" {
		t.Errorf("failed error message: got %q, want %q", failed[0].ErrMsg, "send failed")
	}
	// backoff = 5 * 2^(2-1) = 10s
	wantBackoff := 10 * time.Second
	if failed[0].RetryAfter != wantBackoff {
		t.Errorf("RetryAfter: got %v, want %v", failed[0].RetryAfter, wantBackoff)
	}

	if len(store.getCompleted()) != 0 {
		t.Error("expected no completed jobs on handler error")
	}
}

func TestWorker_HandlerPanic(t *testing.T) {
	store := &fakeStore{
		dequeued: []*Job{
			{ID: "j3", Queue: "email", Attempts: 1, MaxAttempts: 3},
		},
	}

	c := New(store, Config{
		Queues:       []QueueConfig{{Name: "email", Priority: PriorityDefault, Workers: 1, MaxAttempts: 3, Timeout: 5, Backoff: 5}},
		PollInterval: 1,
		ShutdownWait: 5,
	})
	c.Register("email", func(ctx context.Context, job Job) error {
		panic("unexpected nil pointer")
	})

	c.Start(context.Background())
	time.Sleep(200 * time.Millisecond)
	if err := c.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	failed := store.getFailed()
	if len(failed) != 1 {
		t.Fatalf("expected 1 failed record, got %d", len(failed))
	}
	if failed[0].ID != "j3" {
		t.Errorf("failed job ID: got %q, want %q", failed[0].ID, "j3")
	}
	wantMsg := "queue: handler panic: unexpected nil pointer"
	if failed[0].ErrMsg != wantMsg {
		t.Errorf("failed error message: got %q, want %q", failed[0].ErrMsg, wantMsg)
	}

	if len(store.getCompleted()) != 0 {
		t.Error("expected no completed jobs on handler panic")
	}
}

// ---------------------------------------------------------------------------
// Stop Tests
// ---------------------------------------------------------------------------

func TestStop_DoubleStop(t *testing.T) {
	store := &fakeStore{}
	c := newTestClient(store)
	c.Register("email", func(ctx context.Context, job Job) error { return nil })

	c.Start(context.Background())
	if err := c.Stop(); err != nil {
		t.Fatalf("first Stop failed: %v", err)
	}
	err := c.Stop()
	if !errors.Is(err, ErrAlreadyClosed) {
		t.Errorf("second Stop: expected ErrAlreadyClosed, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// computeBackoff Tests
// ---------------------------------------------------------------------------

func TestComputeBackoff(t *testing.T) {
	c := &Client{
		queues: map[string]QueueConfig{},
	}

	qcfg := QueueConfig{Backoff: 5}

	tests := []struct {
		name     string
		attempts int
		want     time.Duration
	}{
		{name: "attempt_0", attempts: 0, want: 5 * time.Second},  // shift=0 -> 5*1
		{name: "attempt_1", attempts: 1, want: 5 * time.Second},  // shift=0 -> 5*1
		{name: "attempt_2", attempts: 2, want: 10 * time.Second}, // shift=1 -> 5*2
		{name: "attempt_3", attempts: 3, want: 20 * time.Second}, // shift=2 -> 5*4
		{name: "attempt_4", attempts: 4, want: 40 * time.Second}, // shift=3 -> 5*8
		{name: "attempt_5", attempts: 5, want: 80 * time.Second}, // shift=4 -> 5*16
		{name: "attempt_10", attempts: 10, want: 5 * (1 << 9) * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.computeBackoff(qcfg, tt.attempts)
			if got != tt.want {
				t.Errorf("computeBackoff(%d): got %v, want %v", tt.attempts, got, tt.want)
			}
		})
	}
}

func TestComputeBackoff_CapsAt20(t *testing.T) {
	c := &Client{queues: map[string]QueueConfig{}}
	qcfg := QueueConfig{Backoff: 1}

	// attempts=22 -> shift capped at 20 -> 1*2^20 = 1048576s
	got := c.computeBackoff(qcfg, 22)
	want := time.Duration(1<<20) * time.Second
	if got != want {
		t.Errorf("computeBackoff(22): got %v, want %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// Priority and JobStatus Constants
// ---------------------------------------------------------------------------

func TestPriorityValues(t *testing.T) {
	tests := []struct {
		name string
		p    Priority
		want int
	}{
		{"Critical", PriorityCritical, 0},
		{"High", PriorityHigh, 10},
		{"Default", PriorityDefault, 20},
		{"Low", PriorityLow, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.p) != tt.want {
				t.Errorf("%s: got %d, want %d", tt.name, tt.p, tt.want)
			}
		})
	}
}

func TestJobStatusValues(t *testing.T) {
	tests := []struct {
		name string
		s    JobStatus
		want string
	}{
		{"Pending", StatusPending, "pending"},
		{"Running", StatusRunning, "running"},
		{"Completed", StatusCompleted, "completed"},
		{"Failed", StatusFailed, "failed"},
		{"Dead", StatusDead, "dead"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.s) != tt.want {
				t.Errorf("%s: got %q, want %q", tt.name, tt.s, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error Sentinels
// ---------------------------------------------------------------------------

func TestErrorSentinels(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrQueueNotFound", ErrQueueNotFound, "queue: queue not found"},
		{"ErrJobNotFound", ErrJobNotFound, "queue: job not found"},
		{"ErrAlreadyClosed", ErrAlreadyClosed, "queue: client already closed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// New Client with multiple queues
// ---------------------------------------------------------------------------

func TestNew_MultipleQueues(t *testing.T) {
	store := &fakeStore{}
	c := New(store, Config{
		Queues: []QueueConfig{
			{Name: "email", Priority: PriorityHigh, Workers: 2},
			{Name: "reports", Priority: PriorityLow, MaxAttempts: 5},
		},
	})

	if len(c.queues) != 2 {
		t.Fatalf("expected 2 queues, got %d", len(c.queues))
	}

	email := c.queues["email"]
	if email.Workers != 2 {
		t.Errorf("email Workers: got %d, want 2", email.Workers)
	}
	if email.MaxAttempts != 3 {
		t.Errorf("email MaxAttempts: got %d, want 3 (default)", email.MaxAttempts)
	}

	reports := c.queues["reports"]
	if reports.MaxAttempts != 5 {
		t.Errorf("reports MaxAttempts: got %d, want 5", reports.MaxAttempts)
	}
	if reports.Workers != 1 {
		t.Errorf("reports Workers: got %d, want 1 (default)", reports.Workers)
	}
}

// ---------------------------------------------------------------------------
// Worker: no handler registered => queue skipped
// ---------------------------------------------------------------------------

func TestStart_NoHandlerSkipsQueue(t *testing.T) {
	store := &fakeStore{
		dequeued: []*Job{
			{ID: "j-skip", Queue: "email", Attempts: 1, MaxAttempts: 3},
		},
	}

	c := newTestClient(store)
	// Do NOT register a handler.

	c.Start(context.Background())
	time.Sleep(200 * time.Millisecond)
	if err := c.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// The job should remain in the dequeued list — no worker should have picked it up.
	if len(store.getCompleted()) != 0 {
		t.Error("expected no completed jobs when no handler is registered")
	}
	if len(store.getFailed()) != 0 {
		t.Error("expected no failed jobs when no handler is registered")
	}
}

// ---------------------------------------------------------------------------
// Sync mode (nil store)
// ---------------------------------------------------------------------------

func newSyncClient() *Client {
	return New(nil, Config{
		Queues: []QueueConfig{
			{Name: "email", MaxAttempts: 1, Timeout: 5, Backoff: 1},
		},
	})
}

func TestSyncDispatch_ExecutesHandlerInline(t *testing.T) {
	c := newSyncClient()

	var received Job
	c.Register("email", func(ctx context.Context, job Job) error {
		received = job
		return nil
	})

	payload := json.RawMessage(`{"to":"alice@example.com"}`)
	err := c.Dispatch(context.Background(), DispatchOptions{
		Queue:    "email",
		Payload:  payload,
		Priority: -1,
	})
	if err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if received.Queue != "email" {
		t.Errorf("expected queue %q, got %q", "email", received.Queue)
	}
	if string(received.Payload) != string(payload) {
		t.Errorf("expected payload %s, got %s", payload, received.Payload)
	}
}

func TestSyncDispatch_PropagatesHandlerError(t *testing.T) {
	c := newSyncClient()
	wantErr := errors.New("handler failed")

	c.Register("email", func(ctx context.Context, job Job) error {
		return wantErr
	})

	err := c.DispatchPayload(context.Background(), "email", map[string]string{"k": "v"})
	if !errors.Is(err, wantErr) {
		t.Errorf("expected error %v, got %v", wantErr, err)
	}
}

func TestSyncDispatch_RecoversPanic(t *testing.T) {
	c := newSyncClient()

	c.Register("email", func(ctx context.Context, job Job) error {
		panic("boom")
	})

	err := c.DispatchPayload(context.Background(), "email", "test")
	if err == nil {
		t.Fatal("expected error from panicking handler")
	}
	if got := err.Error(); got != "queue: handler panic: boom" {
		t.Errorf("unexpected error message: %q", got)
	}
}

func TestSyncDispatch_NoHandler(t *testing.T) {
	c := newSyncClient()
	// Do NOT register a handler.

	err := c.DispatchPayload(context.Background(), "email", "test")
	if err == nil {
		t.Fatal("expected error for missing handler")
	}
}

func TestSyncClient_Async(t *testing.T) {
	c := newSyncClient()
	if c.Async() {
		t.Error("sync client should report Async() == false")
	}

	storeClient := newTestClient(&fakeStore{})
	if !storeClient.Async() {
		t.Error("store client should report Async() == true")
	}
}

func TestSyncClient_StartStop(t *testing.T) {
	c := newSyncClient()
	c.Register("email", func(ctx context.Context, job Job) error { return nil })

	// Start and Stop should be safe no-ops in sync mode.
	c.Start(context.Background())
	if err := c.Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}

	// After stop, dispatch should fail.
	err := c.DispatchPayload(context.Background(), "email", "test")
	if !errors.Is(err, ErrAlreadyClosed) {
		t.Errorf("expected ErrAlreadyClosed, got %v", err)
	}
}
