package session_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-sum/session"
)

// fakeBlobStore is an in-memory BlobStore for testing the server backend.
type fakeBlobStore struct {
	mu   sync.Mutex
	data map[string]blobEntry
}

type blobEntry struct {
	value     []byte
	expiresAt time.Time
}

func newFakeBlobStore() *fakeBlobStore {
	return &fakeBlobStore{data: make(map[string]blobEntry)}
}

func (s *fakeBlobStore) Get(_ context.Context, key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.data[key]
	if !ok || (!e.expiresAt.IsZero() && time.Now().After(e.expiresAt)) {
		return nil, session.ErrBlobNotFound
	}
	cp := make([]byte, len(e.value))
	copy(cp, e.value)
	return cp, nil
}

func (s *fakeBlobStore) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]byte, len(value))
	copy(cp, value)
	e := blobEntry{value: cp}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}
	s.data[key] = e
	return nil
}

func (s *fakeBlobStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *fakeBlobStore) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}

func testServerManager(t *testing.T, store session.BlobStore) session.Manager {
	t.Helper()
	mgr, err := session.NewManager(session.Config{
		Store:    "server",
		CookieName: "_test",
		AuthKey:    testAuthKey,
		EncryptKey: testEncryptKey,
		MaxAge:     3600,
		Secure:     false,
		BlobStore:  store,
		KeyPrefix:  "sess:",
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return mgr
}

func TestServerBackendPutGetRoundtrip(t *testing.T) {
	store := newFakeBlobStore()
	mgr := testServerManager(t, store)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)
	_ = state.Put("auth.user_id", "u1")

	loaded := commitAndLoad(t, mgr, state)

	var got string
	ok, _ := loaded.Get("auth.user_id", &got)
	if !ok || got != "u1" {
		t.Fatalf("Get = (%v, %q), want (true, u1)", ok, got)
	}
}

func TestServerBackendStoresDataInBlob(t *testing.T) {
	store := newFakeBlobStore()
	mgr := testServerManager(t, store)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)
	_ = state.Put("k", "v")

	rec := httptest.NewRecorder()
	_ = mgr.Commit(rec, req, state)

	if store.Len() != 1 {
		t.Fatalf("blob store should have 1 entry, got %d", store.Len())
	}
}

func TestServerBackendDestroyDeletesBlob(t *testing.T) {
	store := newFakeBlobStore()
	mgr := testServerManager(t, store)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)
	_ = state.Put("k", "v")
	rec := httptest.NewRecorder()
	_ = mgr.Commit(rec, req, state)

	if store.Len() != 1 {
		t.Fatal("expected 1 blob before destroy")
	}

	destroyReq := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		destroyReq.AddCookie(c)
	}
	destroyRec := httptest.NewRecorder()
	_ = mgr.Destroy(destroyRec, destroyReq)

	if store.Len() != 0 {
		t.Fatalf("blob store should be empty after destroy, got %d", store.Len())
	}
}

func TestServerBackendMultipleKeys(t *testing.T) {
	store := newFakeBlobStore()
	mgr := testServerManager(t, store)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)
	_ = state.Put("auth.user_id", "u1")
	_ = state.Put("ui.theme", "dark")

	loaded := commitAndLoad(t, mgr, state)

	var uid, theme string
	ok1, _ := loaded.Get("auth.user_id", &uid)
	ok2, _ := loaded.Get("ui.theme", &theme)

	if !ok1 || uid != "u1" {
		t.Fatalf("user_id: (%v, %q)", ok1, uid)
	}
	if !ok2 || theme != "dark" {
		t.Fatalf("theme: (%v, %q)", ok2, theme)
	}
}

func TestServerBackendCookieContainsOnlySessionID(t *testing.T) {
	store := newFakeBlobStore()
	mgr := testServerManager(t, store)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)
	_ = state.Put("auth.user_id", "u1")

	rec := httptest.NewRecorder()
	_ = mgr.Commit(rec, req, state)

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookie set")
	}
	// Server backend cookie should contain only the session ID (64 hex chars).
	// It should NOT contain the actual session data.
	cookieVal := cookies[0].Value
	if len(cookieVal) == 0 {
		t.Fatal("cookie value is empty")
	}
	// The cookie value should be the session ID, not the full data blob.
	// Session IDs are 64 hex chars. Cookie-encoded data would be much longer.
	// We can't check exact length since the server backend stores the raw ID in cookie,
	// but it should be reasonably short (under 200 chars for a 64-char hex ID).
	if len(cookieVal) > 200 {
		t.Fatalf("cookie value too long for session ID (%d chars) — data may be leaking into cookie", len(cookieVal))
	}
}

func TestServerBackendMissingBlobReturnsNewSession(t *testing.T) {
	store := newFakeBlobStore()
	mgr := testServerManager(t, store)

	// Create a session.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)
	_ = state.Put("k", "v")
	rec := httptest.NewRecorder()
	_ = mgr.Commit(rec, req, state)

	// Delete the blob directly.
	store.mu.Lock()
	for k := range store.data {
		delete(store.data, k)
	}
	store.mu.Unlock()

	// Load with the old cookie should give a new session.
	readReq := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		readReq.AddCookie(c)
	}
	newState, err := mgr.Load(readReq)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !newState.IsNew() {
		t.Fatal("expected new session when blob is missing")
	}
}
