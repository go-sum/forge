package session_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-sum/session"
)

// testMultiManager constructs a server-mode manager with custom prefixes and
// returns it as session.MultiManager. The test fails if the type assertion fails.
func testMultiManager(t *testing.T, store session.BlobStore) session.MultiManager {
	t.Helper()
	mgr, err := session.NewManager(session.Config{
		Store:       "server",
		CookieName:  "_test",
		AuthKey:     testAuthKey,
		EncryptKey:  testEncryptKey,
		MaxAge:      3600,
		Secure:      false,
		BlobStore:   store,
		KeyPrefix:   "sess:",
		UserPrefix:  "usr:",
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	mm, ok := mgr.(session.MultiManager)
	if !ok {
		t.Fatalf("server-mode manager does not implement MultiManager")
	}
	return mm
}

// testMultiManagerWithWindow constructs a manager with an explicit TouchWindow.
func testMultiManagerWithWindow(t *testing.T, store session.BlobStore, window time.Duration) session.MultiManager {
	t.Helper()
	mgr, err := session.NewManager(session.Config{
		Store:       "server",
		CookieName:  "_test",
		AuthKey:     testAuthKey,
		EncryptKey:  testEncryptKey,
		MaxAge:      3600,
		Secure:      false,
		BlobStore:   store,
		KeyPrefix:   "sess:",
		UserPrefix:  "usr:",
		TouchWindow: window,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	mm, ok := mgr.(session.MultiManager)
	if !ok {
		t.Fatalf("server-mode manager does not implement MultiManager")
	}
	return mm
}

// countingBlobStore wraps a BlobStore and counts Set calls.
type countingBlobStore struct {
	session.BlobStore
	mu   sync.Mutex
	sets int
}

func (c *countingBlobStore) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	c.mu.Lock()
	c.sets++
	c.mu.Unlock()
	return c.BlobStore.Set(ctx, key, val, ttl)
}

func (c *countingBlobStore) Sets() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sets
}

// commitSession commits a new session through the manager and returns the session ID
// and the set-cookie response. The session blob will exist in the fake store.
func commitSession(t *testing.T, mgr session.Manager) (sessionID string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, err := mgr.Load(req)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	_ = state.Put("k", "v")
	rec := httptest.NewRecorder()
	if err := mgr.Commit(rec, req, state); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	return state.ID()
}

// makeMeta constructs a SessionMeta with the given session ID and controlled timestamps.
func makeMeta(sessionID string, createdAt, lastActiveAt time.Time) session.SessionMeta {
	return session.SessionMeta{
		SessionID:    sessionID,
		AuthMethod:   "email_totp",
		IPAddress:    "1.2.3.4",
		UserAgent:    "TestAgent/1.0",
		CreatedAt:    createdAt,
		LastActiveAt: lastActiveAt,
	}
}

// TestBindUserCreatesIndex verifies that BindUser on a fresh user creates an
// index entry and that ListUserSessions returns the expected metadata fields.
func TestBindUserCreatesIndex(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sid := commitSession(t, mm)
	now := time.Now().Truncate(time.Millisecond)
	meta := makeMeta(sid, now, now)

	if err := mm.BindUser(ctx, sid, "user-a", meta); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	sessions, err := mm.ListUserSessions(ctx, "user-a")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("ListUserSessions returned %d entries, want 1", len(sessions))
	}

	got := sessions[0]
	if got.SessionID != sid {
		t.Errorf("SessionID = %q, want %q", got.SessionID, sid)
	}
	if got.AuthMethod != "email_totp" {
		t.Errorf("AuthMethod = %q, want %q", got.AuthMethod, "email_totp")
	}
	if got.IPAddress != "1.2.3.4" {
		t.Errorf("IPAddress = %q, want %q", got.IPAddress, "1.2.3.4")
	}
	if got.UserAgent != "TestAgent/1.0" {
		t.Errorf("UserAgent = %q, want %q", got.UserAgent, "TestAgent/1.0")
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if got.LastActiveAt.IsZero() {
		t.Error("LastActiveAt is zero")
	}
}

// TestBindUserMultipleSessions verifies that two BindUser calls with different
// session IDs for the same user result in two index entries.
func TestBindUserMultipleSessions(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sid1 := commitSession(t, mm)
	sid2 := commitSession(t, mm)

	now := time.Now().Truncate(time.Millisecond)
	if err := mm.BindUser(ctx, sid1, "user-b", makeMeta(sid1, now, now)); err != nil {
		t.Fatalf("BindUser sid1: %v", err)
	}
	if err := mm.BindUser(ctx, sid2, "user-b", makeMeta(sid2, now.Add(time.Millisecond), now.Add(time.Millisecond))); err != nil {
		t.Fatalf("BindUser sid2: %v", err)
	}

	sessions, err := mm.ListUserSessions(ctx, "user-b")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("ListUserSessions returned %d entries, want 2", len(sessions))
	}
}

// TestBindUserPreservesCreatedAt verifies that a second BindUser for the same
// session preserves the original CreatedAt and updates LastActiveAt.
func TestBindUserPreservesCreatedAt(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sid := commitSession(t, mm)

	t1 := time.Now().Add(-1 * time.Hour).Truncate(time.Millisecond)
	meta1 := makeMeta(sid, t1, t1)
	if err := mm.BindUser(ctx, sid, "user-c", meta1); err != nil {
		t.Fatalf("first BindUser: %v", err)
	}

	// Second BindUser with a fresh time — CreatedAt should be preserved from first call.
	t2 := time.Now().Truncate(time.Millisecond)
	meta2 := makeMeta(sid, t2, t2)
	if err := mm.BindUser(ctx, sid, "user-c", meta2); err != nil {
		t.Fatalf("second BindUser: %v", err)
	}

	sessions, err := mm.ListUserSessions(ctx, "user-c")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("ListUserSessions returned %d entries, want 1", len(sessions))
	}

	got := sessions[0]
	// CreatedAt should match t1, not t2.
	if !got.CreatedAt.Equal(t1) {
		t.Errorf("CreatedAt = %v, want %v (original)", got.CreatedAt, t1)
	}
	// LastActiveAt should be after t1.
	if !got.LastActiveAt.After(t1) {
		t.Errorf("LastActiveAt = %v should be after t1 = %v", got.LastActiveAt, t1)
	}
}

// TestUnbindUserRemovesSession verifies that UnbindUser removes the session from
// the index and ListUserSessions returns an empty slice.
func TestUnbindUserRemovesSession(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sid := commitSession(t, mm)
	now := time.Now()
	if err := mm.BindUser(ctx, sid, "user-d", makeMeta(sid, now, now)); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	if err := mm.UnbindUser(ctx, sid, "user-d"); err != nil {
		t.Fatalf("UnbindUser: %v", err)
	}

	sessions, err := mm.ListUserSessions(ctx, "user-d")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("ListUserSessions returned %d entries after UnbindUser, want 0", len(sessions))
	}
}

// TestUnbindUserIdempotent verifies that UnbindUser on a non-existent
// session or user returns nil and does not panic.
func TestUnbindUserIdempotent(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	// First call — user and session do not exist.
	if err := mm.UnbindUser(ctx, "no-such-session", "no-such-user"); err != nil {
		t.Fatalf("first UnbindUser: %v", err)
	}
	// Second call — still does not exist.
	if err := mm.UnbindUser(ctx, "no-such-session", "no-such-user"); err != nil {
		t.Fatalf("second UnbindUser: %v", err)
	}
}

// TestUnbindUserDeletesIndexWhenEmpty verifies that once the last session is
// unbound, the user index blob key is removed from the store.
func TestUnbindUserDeletesIndexWhenEmpty(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sid := commitSession(t, mm)
	now := time.Now()
	if err := mm.BindUser(ctx, sid, "user-e", makeMeta(sid, now, now)); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	// After bind: session blob + user index blob = at least 2 entries.
	if store.Len() < 2 {
		t.Fatalf("expected at least 2 blobs after BindUser, got %d", store.Len())
	}

	if err := mm.UnbindUser(ctx, sid, "user-e"); err != nil {
		t.Fatalf("UnbindUser: %v", err)
	}

	// The user index key "usr:user-e" should be gone.
	_, err := store.Get(ctx, "usr:user-e")
	if err == nil {
		t.Error("user index blob should have been deleted after last session unbound")
	}
}

// TestListUserSessionsPrunesStale verifies that ListUserSessions removes entries
// whose session blobs have been deleted, and writes the pruned index back to the store.
func TestListUserSessionsPrunesStale(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sid1 := commitSession(t, mm)
	sid2 := commitSession(t, mm)

	now := time.Now().Truncate(time.Millisecond)
	if err := mm.BindUser(ctx, sid1, "user-f", makeMeta(sid1, now, now)); err != nil {
		t.Fatalf("BindUser sid1: %v", err)
	}
	if err := mm.BindUser(ctx, sid2, "user-f", makeMeta(sid2, now.Add(time.Millisecond), now.Add(time.Millisecond))); err != nil {
		t.Fatalf("BindUser sid2: %v", err)
	}

	// Delete the blob for sid1 directly, simulating a stale/expired session.
	if err := store.Delete(ctx, "sess:"+sid1); err != nil {
		t.Fatalf("Delete sid1 blob: %v", err)
	}

	sessions, err := mm.ListUserSessions(ctx, "user-f")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("ListUserSessions returned %d entries, want 1 (stale pruned)", len(sessions))
	}
	if sessions[0].SessionID != sid2 {
		t.Errorf("SessionID = %q, want %q", sessions[0].SessionID, sid2)
	}

	// Verify that sid1 is no longer referenced in the stored index by calling
	// ListUserSessions again — it should still return only sid2.
	sessions2, err := mm.ListUserSessions(ctx, "user-f")
	if err != nil {
		t.Fatalf("second ListUserSessions: %v", err)
	}
	if len(sessions2) != 1 {
		t.Fatalf("second ListUserSessions returned %d entries, want 1", len(sessions2))
	}
}

// TestListUserSessionsEmptyUser verifies that ListUserSessions for an unknown
// user returns an empty (or nil) slice and no error, without panicking.
func TestListUserSessionsEmptyUser(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sessions, err := mm.ListUserSessions(ctx, "unknown-user")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("ListUserSessions returned %d entries, want 0", len(sessions))
	}
}

// TestDestroySessionDeletesBlobAndIndex verifies that DestroySession removes
// the session blob and removes the session from the user index.
func TestDestroySessionDeletesBlobAndIndex(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sid := commitSession(t, mm)
	now := time.Now()
	if err := mm.BindUser(ctx, sid, "user-g", makeMeta(sid, now, now)); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	if err := mm.DestroySession(ctx, sid, "user-g"); err != nil {
		t.Fatalf("DestroySession: %v", err)
	}

	// Session blob should be gone.
	_, err := store.Get(ctx, "sess:"+sid)
	if err == nil {
		t.Error("session blob should have been deleted by DestroySession")
	}

	// User index should be empty.
	sessions, err := mm.ListUserSessions(ctx, "user-g")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("ListUserSessions returned %d entries after DestroySession, want 0", len(sessions))
	}
}

// TestDestroyUserSessionsClearsAll verifies that DestroyUserSessions removes all
// session blobs for a user and clears the index.
func TestDestroyUserSessionsClearsAll(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sid1 := commitSession(t, mm)
	sid2 := commitSession(t, mm)
	sid3 := commitSession(t, mm)

	now := time.Now().Truncate(time.Millisecond)
	if err := mm.BindUser(ctx, sid1, "user-h", makeMeta(sid1, now, now)); err != nil {
		t.Fatalf("BindUser sid1: %v", err)
	}
	if err := mm.BindUser(ctx, sid2, "user-h", makeMeta(sid2, now.Add(time.Millisecond), now.Add(time.Millisecond))); err != nil {
		t.Fatalf("BindUser sid2: %v", err)
	}
	if err := mm.BindUser(ctx, sid3, "user-h", makeMeta(sid3, now.Add(2*time.Millisecond), now.Add(2*time.Millisecond))); err != nil {
		t.Fatalf("BindUser sid3: %v", err)
	}

	// Store should have 3 session blobs + 3 meta blobs + 1 index blob = 7 entries.
	if store.Len() != 7 {
		t.Fatalf("expected 7 blobs before DestroyUserSessions, got %d", store.Len())
	}

	if err := mm.DestroyUserSessions(ctx, "user-h"); err != nil {
		t.Fatalf("DestroyUserSessions: %v", err)
	}

	// All blobs should be gone.
	if store.Len() != 0 {
		t.Fatalf("expected 0 blobs after DestroyUserSessions, got %d", store.Len())
	}

	sessions, err := mm.ListUserSessions(ctx, "user-h")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("ListUserSessions returned %d entries after DestroyUserSessions, want 0", len(sessions))
	}
}

// TestTouchSessionUpdatesLastActive verifies that TouchSession updates
// LastActiveAt for the session in the user index.
func TestTouchSessionUpdatesLastActive(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	sid := commitSession(t, mm)

	// Set a clearly past LastActiveAt.
	t1 := time.Now().Add(-1 * time.Hour).Truncate(time.Millisecond)
	meta := makeMeta(sid, t1, t1)
	if err := mm.BindUser(ctx, sid, "user-i", meta); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	if err := mm.TouchSession(ctx, sid, "user-i"); err != nil {
		t.Fatalf("TouchSession: %v", err)
	}

	sessions, err := mm.ListUserSessions(ctx, "user-i")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("ListUserSessions returned %d entries, want 1", len(sessions))
	}

	if !sessions[0].LastActiveAt.After(t1) {
		t.Errorf("LastActiveAt = %v, want after %v", sessions[0].LastActiveAt, t1)
	}
}

// TestTouchSessionNonExistentIsNoop verifies that TouchSession for an unknown
// session/user returns nil without panicking.
func TestTouchSessionNonExistentIsNoop(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	if err := mm.TouchSession(ctx, "no-such-session", "no-such-user"); err != nil {
		t.Fatalf("TouchSession on non-existent session: %v", err)
	}
}

// TestCookieModeManagerIsNotMultiManager verifies that a cookie-mode manager
// does not implement the MultiManager interface.
func TestCookieModeManagerIsNotMultiManager(t *testing.T) {
	mgr, err := session.NewManager(session.Config{
		Store:      "cookie",
		CookieName: "_test",
		AuthKey:    testAuthKey,
		EncryptKey: testEncryptKey,
		MaxAge:     3600,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	_, ok := mgr.(session.MultiManager)
	if ok {
		t.Fatal("cookie-mode manager should not implement MultiManager")
	}
}

// TestServerModeManagerIsMultiManager verifies that a server-mode manager
// implements the MultiManager interface.
func TestServerModeManagerIsMultiManager(t *testing.T) {
	store := newFakeBlobStore()
	mgr, err := session.NewManager(session.Config{
		Store:      "server",
		CookieName: "_test",
		AuthKey:    testAuthKey,
		EncryptKey: testEncryptKey,
		MaxAge:     3600,
		BlobStore:  store,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	_, ok := mgr.(session.MultiManager)
	if !ok {
		t.Fatal("server-mode manager should implement MultiManager")
	}
}

// TestTouchSessionWithinWindowSkipsKVWrite verifies that a second TouchSession
// call within the touch window does not write to the blob store.
func TestTouchSessionWithinWindowSkipsKVWrite(t *testing.T) {
	inner := newFakeBlobStore()
	store := &countingBlobStore{BlobStore: inner}
	mm := testMultiManagerWithWindow(t, store, time.Hour)
	ctx := context.Background()

	sid := commitSession(t, mm)
	now := time.Now()
	if err := mm.BindUser(ctx, sid, "user-j", makeMeta(sid, now, now)); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	setsBefore := store.Sets()
	// First touch — outside window (no prior fence entry); must write.
	if err := mm.TouchSession(ctx, sid, "user-j"); err != nil {
		t.Fatalf("first TouchSession: %v", err)
	}
	if store.Sets() == setsBefore {
		t.Error("first TouchSession should have written to the blob store")
	}

	setsAfterFirst := store.Sets()
	// Second touch — within the 1h window; must skip the write.
	if err := mm.TouchSession(ctx, sid, "user-j"); err != nil {
		t.Fatalf("second TouchSession: %v", err)
	}
	if store.Sets() != setsAfterFirst {
		t.Errorf("second TouchSession within window wrote %d extra time(s), want 0", store.Sets()-setsAfterFirst)
	}
}

// TestTouchSessionOutsideWindowWritesKV verifies that a TouchSession call after
// the touch window has elapsed does write to the blob store.
func TestTouchSessionOutsideWindowWritesKV(t *testing.T) {
	inner := newFakeBlobStore()
	store := &countingBlobStore{BlobStore: inner}
	mm := testMultiManagerWithWindow(t, store, time.Millisecond)
	ctx := context.Background()

	sid := commitSession(t, mm)
	now := time.Now()
	if err := mm.BindUser(ctx, sid, "user-k", makeMeta(sid, now, now)); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	if err := mm.TouchSession(ctx, sid, "user-k"); err != nil {
		t.Fatalf("first TouchSession: %v", err)
	}

	time.Sleep(2 * time.Millisecond) // exceed the 1ms window

	setsBefore := store.Sets()
	if err := mm.TouchSession(ctx, sid, "user-k"); err != nil {
		t.Fatalf("second TouchSession: %v", err)
	}
	if store.Sets() == setsBefore {
		t.Error("TouchSession after window expiry should have written to the blob store")
	}
}

// TestDestroySessionRejectsUnownedSession verifies that DestroySession returns
// ErrSessionNotOwned when the session ID is not in the user's index.
func TestDestroySessionRejectsUnownedSession(t *testing.T) {
	store := newFakeBlobStore()
	mm := testMultiManager(t, store)
	ctx := context.Background()

	// Create and bind a session for user-A.
	sidA := commitSession(t, mm)
	now := time.Now()
	if err := mm.BindUser(ctx, sidA, "user-a", makeMeta(sidA, now, now)); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	// user-B attempts to destroy user-A's session — must be rejected.
	err := mm.DestroySession(ctx, sidA, "user-b")
	if !errors.Is(err, session.ErrSessionNotOwned) {
		t.Errorf("DestroySession with foreign session returned %v, want ErrSessionNotOwned", err)
	}

	// user-A's session blob must still be intact.
	sessions, err := mm.ListUserSessions(ctx, "user-a")
	if err != nil {
		t.Fatalf("ListUserSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("ListUserSessions returned %d entries after rejected revoke, want 1", len(sessions))
	}
}

// errBlobStore wraps a BlobStore and returns a fixed error for Get on a specific key.
type errBlobStore struct {
	session.BlobStore
	failKey string
	failErr error
}

func (e *errBlobStore) Get(ctx context.Context, key string) ([]byte, error) {
	if key == e.failKey {
		return nil, e.failErr
	}
	return e.BlobStore.Get(ctx, key)
}

// TestListUserSessionsTransientBlobError verifies that a non-ErrBlobNotFound
// error from the session blob read is propagated and does NOT prune the index.
func TestListUserSessionsTransientBlobError(t *testing.T) {
	inner := newFakeBlobStore()
	ctx := context.Background()

	sid := commitSession(t, testMultiManager(t, inner))
	now := time.Now()
	// Bind using the raw inner store so we can later inject errors at the wrapping layer.
	mmInner := testMultiManager(t, inner)
	if err := mmInner.BindUser(ctx, sid, "user-x", makeMeta(sid, now, now)); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	transientErr := errors.New("kv: connection reset")
	wrapped := &errBlobStore{BlobStore: inner, failKey: "sess:" + sid, failErr: transientErr}
	mmWrapped := testMultiManager(t, wrapped)

	_, err := mmWrapped.ListUserSessions(ctx, "user-x")
	if err == nil {
		t.Fatal("ListUserSessions should have returned an error on transient blob failure")
	}
	if !errors.Is(err, transientErr) {
		t.Errorf("ListUserSessions returned %v, want wrapping transientErr", err)
	}

	// Index must not have been modified — user-x should still have 1 session.
	sessions, err := mmInner.ListUserSessions(ctx, "user-x")
	if err != nil {
		t.Fatalf("ListUserSessions on inner: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("index was modified after transient error: got %d sessions, want 1", len(sessions))
	}
}

// TestUnbindUserClearsFence verifies that UnbindUser (the signout path) evicts
// the touchFence entry so that a subsequent touch on a re-bound session always writes.
func TestUnbindUserClearsFence(t *testing.T) {
	inner := newFakeBlobStore()
	store := &countingBlobStore{BlobStore: inner}
	mm := testMultiManagerWithWindow(t, store, time.Hour)
	ctx := context.Background()

	sid := commitSession(t, mm)
	now := time.Now()
	if err := mm.BindUser(ctx, sid, "user-y", makeMeta(sid, now, now)); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	// Populate the fence via a touch.
	if err := mm.TouchSession(ctx, sid, "user-y"); err != nil {
		t.Fatalf("TouchSession: %v", err)
	}

	// Signout path: UnbindUser should clear the fence.
	if err := mm.UnbindUser(ctx, sid, "user-y"); err != nil {
		t.Fatalf("UnbindUser: %v", err)
	}

	// Re-bind the same session ID under the same user and touch — fence is gone,
	// so a new KV write must happen regardless of the 1h window.
	if err := mm.BindUser(ctx, sid, "user-y", makeMeta(sid, now, now)); err != nil {
		t.Fatalf("second BindUser: %v", err)
	}
	setsBefore := store.Sets()
	if err := mm.TouchSession(ctx, sid, "user-y"); err != nil {
		t.Fatalf("TouchSession after UnbindUser: %v", err)
	}
	if store.Sets() == setsBefore {
		t.Error("TouchSession after UnbindUser should write (fence was cleared)")
	}
}

// TestDestroySessionClearsFence verifies that DestroySession evicts the fence
// entry so that a re-bound session always writes on the next touch.
func TestDestroySessionClearsFence(t *testing.T) {
	inner := newFakeBlobStore()
	store := &countingBlobStore{BlobStore: inner}
	mm := testMultiManagerWithWindow(t, store, time.Hour)
	ctx := context.Background()

	sid := commitSession(t, mm)
	now := time.Now()
	if err := mm.BindUser(ctx, sid, "user-l", makeMeta(sid, now, now)); err != nil {
		t.Fatalf("BindUser: %v", err)
	}

	// Touch once to populate the fence.
	if err := mm.TouchSession(ctx, sid, "user-l"); err != nil {
		t.Fatalf("TouchSession: %v", err)
	}

	// Destroy clears the fence.
	if err := mm.DestroySession(ctx, sid, "user-l"); err != nil {
		t.Fatalf("DestroySession: %v", err)
	}

	// Recreate and re-bind the session under the same ID simulation:
	// bind a new session and touch — it must write (fence was cleared).
	sid2 := commitSession(t, mm)
	if err := mm.BindUser(ctx, sid2, "user-l", makeMeta(sid2, now, now)); err != nil {
		t.Fatalf("second BindUser: %v", err)
	}
	setsBefore := store.Sets()
	if err := mm.TouchSession(ctx, sid2, "user-l"); err != nil {
		t.Fatalf("TouchSession after destroy: %v", err)
	}
	if store.Sets() == setsBefore {
		t.Error("TouchSession on fresh session after DestroySession should write")
	}
}
