package sessions

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/app/testutil"
	"github.com/go-sum/session"
)

// ---------------------------------------------------------------------------
// in-memory BlobStore for server-mode session manager
// ---------------------------------------------------------------------------

type inMemBlobStore struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newInMemBlobStore() *inMemBlobStore {
	return &inMemBlobStore{data: make(map[string][]byte)}
}

func (s *inMemBlobStore) Get(_ context.Context, key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	if !ok {
		return nil, session.ErrBlobNotFound
	}
	return v, nil
}

func (s *inMemBlobStore) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *inMemBlobStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// ---------------------------------------------------------------------------
// fakeMultiManager
// ---------------------------------------------------------------------------

type fakeMultiManager struct {
	// Underlying manager handles Load/Commit/Destroy/RotateID.
	mgr session.Manager

	// Configurable list/error returns.
	listSessions []session.SessionMeta
	listErr      error

	// Capture args for DestroySession.
	destroyedIDs []string
	destroyErr   error
}

// newFakeMultiManager creates a fake backed by a cookie-mode manager.
// The session ID returned by Load will be a random value each call — suitable
// for tests that do NOT need to match the current session ID.
func newFakeMultiManager(t *testing.T) *fakeMultiManager {
	t.Helper()
	mgr, err := session.NewManager(session.Config{
		Store:      "cookie",
		CookieName: "_test",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 16),
	})
	if err != nil {
		t.Fatalf("newFakeMultiManager: %v", err)
	}
	return &fakeMultiManager{mgr: mgr}
}

// newFakeMultiManagerServer creates a fake backed by a server-mode manager
// with an in-memory blob store. Use this when you need a stable, reproducible
// session ID (by committing the session and replaying the cookie).
func newFakeMultiManagerServer(t *testing.T) (*fakeMultiManager, *inMemBlobStore) {
	t.Helper()
	store := newInMemBlobStore()
	mgr, err := session.NewManager(session.Config{
		Store:      "server",
		CookieName: "_test",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 16),
		MaxAge:     3600,
		BlobStore:  store,
	})
	if err != nil {
		t.Fatalf("newFakeMultiManagerServer: %v", err)
	}
	// mgr from server mode is itself a MultiManager; wrap in our fake so we
	// can inject listSessions / destroyedIDs etc.
	return &fakeMultiManager{mgr: mgr}, store
}

// Manager interface methods — delegate to real manager.

func (f *fakeMultiManager) Load(r *http.Request) (*session.State, error) {
	return f.mgr.Load(r)
}

func (f *fakeMultiManager) Commit(w http.ResponseWriter, r *http.Request, s *session.State) error {
	return f.mgr.Commit(w, r, s)
}

func (f *fakeMultiManager) Destroy(w http.ResponseWriter, r *http.Request) error {
	return f.mgr.Destroy(w, r)
}

func (f *fakeMultiManager) RotateID(w http.ResponseWriter, r *http.Request, s *session.State) error {
	return f.mgr.RotateID(w, r, s)
}

// MultiManager-specific methods.

func (f *fakeMultiManager) BindUser(_ context.Context, _, _ string, _ session.SessionMeta) error {
	return nil
}

func (f *fakeMultiManager) UnbindUser(_ context.Context, _, _ string) error {
	return nil
}

func (f *fakeMultiManager) ListUserSessions(_ context.Context, _ string) ([]session.SessionMeta, error) {
	return f.listSessions, f.listErr
}

func (f *fakeMultiManager) DestroySession(_ context.Context, sessionID, _ string) error {
	f.destroyedIDs = append(f.destroyedIDs, sessionID)
	return f.destroyErr
}

func (f *fakeMultiManager) DestroyUserSessions(_ context.Context, _ string) error {
	return nil
}

func (f *fakeMultiManager) TouchSession(_ context.Context, _, _ string) error {
	return nil
}

// ---------------------------------------------------------------------------
// test helper
// ---------------------------------------------------------------------------

func newTestSessionHandler(mgr session.MultiManager) *Handler {
	return NewHandler(
		&config.Config{
			Security: config.SecurityConfig{
				CSRF: config.CSRFConfig{ContextKey: "csrf"},
			},
		},
		mgr,
	)
}

// seedSession commits a session via mgr and returns its ID and the cookie to
// replay it on subsequent requests. Both the seed request/recorder and the
// returned cookie are valid for a server-mode manager.
func seedSession(t *testing.T, mgr session.Manager) (id string, cookie *http.Cookie) {
	t.Helper()
	seedReq := httptest.NewRequest(http.MethodGet, "/", nil)
	seedRec := httptest.NewRecorder()
	state, err := mgr.Load(seedReq)
	if err != nil {
		t.Fatalf("seedSession Load: %v", err)
	}
	if err := mgr.Commit(seedRec, seedReq, state); err != nil {
		t.Fatalf("seedSession Commit: %v", err)
	}
	cookies := seedRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("seedSession: no cookie set after Commit")
	}
	return state.ID(), cookies[0]
}

// ---------------------------------------------------------------------------
// List tests
// ---------------------------------------------------------------------------

func TestListRendersSessionRegion(t *testing.T) {
	mgr := newFakeMultiManager(t)
	now := time.Now()
	mgr.listSessions = []session.SessionMeta{
		{SessionID: "s1", AuthMethod: "passkey", CreatedAt: now, LastActiveAt: now},
		{SessionID: "s2", AuthMethod: "passkey", CreatedAt: now.Add(time.Second), LastActiveAt: now},
	}

	h := newTestSessionHandler(mgr)
	c, rec := testutil.NewRequestContext(http.MethodGet, "/account/sessions", nil)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	if err := h.List(c); err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "sessions-list-region") {
		t.Error("expected sessions-list-region in response body")
	}
}

func TestListNilManagerRendersEmptyState(t *testing.T) {
	h := newTestSessionHandler(nil)
	c, rec := testutil.NewRequestContext(http.MethodGet, "/account/sessions", nil)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	if err := h.List(c); err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "sessions-list-region") {
		t.Error("expected sessions-list-region in response body")
	}
	// With no manager there are no sessions, so no Sign out buttons.
	if strings.Contains(body, "Sign out") {
		t.Error("expected no Sign out buttons when manager is nil")
	}
}

func TestListServiceError(t *testing.T) {
	mgr := newFakeMultiManager(t)
	mgr.listErr = errors.New("kv down")

	h := newTestSessionHandler(mgr)
	c, _ := testutil.NewRequestContext(http.MethodGet, "/account/sessions", nil)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	err := h.List(c)
	testutil.AssertAppErrorStatus(t, err, http.StatusServiceUnavailable)
}

func TestListMarksCurrentSession(t *testing.T) {
	// Use a server-mode manager so we can replay the session cookie and get a
	// stable, matching ID inside the handler.
	mgr, _ := newFakeMultiManagerServer(t)
	currentID, cookie := seedSession(t, mgr)

	now := time.Now()
	mgr.listSessions = []session.SessionMeta{
		{SessionID: currentID, AuthMethod: "passkey", CreatedAt: now, LastActiveAt: now},
		{SessionID: "other-session", AuthMethod: "passkey", CreatedAt: now.Add(time.Second), LastActiveAt: now},
	}

	h := newTestSessionHandler(mgr)
	c, rec := testutil.NewRequestContext(http.MethodGet, "/account/sessions", nil)
	c.Request().AddCookie(cookie)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	if err := h.List(c); err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	// The current session row should include the "Current" badge text.
	body := rec.Body.String()
	if !strings.Contains(body, "Current") {
		t.Error("expected 'Current' badge in response body for current session")
	}
	// Other sessions should have a "Sign out" button.
	if !strings.Contains(body, "Sign out") {
		t.Error("expected Sign out button for non-current session")
	}
}

func TestListHTMXPartialReturnsFragment(t *testing.T) {
	mgr := newFakeMultiManager(t)
	mgr.listSessions = []session.SessionMeta{}

	h := newTestSessionHandler(mgr)
	c, rec := testutil.NewRequestContext(http.MethodGet, "/account/sessions", nil)
	c.Request().Header.Set("HX-Request", "true")
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	if err := h.List(c); err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	// Fragment-only: should contain the region div but not the full HTML document.
	if !strings.Contains(body, "sessions-list-region") {
		t.Error("expected sessions-list-region in fragment body")
	}
	if strings.Contains(body, "<html") {
		t.Error("HTMX partial must not return a full HTML document")
	}
}

// ---------------------------------------------------------------------------
// Revoke tests
// ---------------------------------------------------------------------------

func TestRevokeSuccess(t *testing.T) {
	mgr := newFakeMultiManager(t)
	h := newTestSessionHandler(mgr)

	c, rec := testutil.NewRequestContext(http.MethodDelete, "/account/sessions/abc123", nil)
	testutil.SetPathParam(c, "/account/sessions/:id", "id", "abc123")
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	if err := h.Revoke(c); err != nil {
		t.Fatalf("Revoke returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if len(mgr.destroyedIDs) != 1 || mgr.destroyedIDs[0] != "abc123" {
		t.Errorf("expected DestroySession(abc123), got %v", mgr.destroyedIDs)
	}
}

func TestRevokeMissingIDReturnsBadRequest(t *testing.T) {
	mgr := newFakeMultiManager(t)
	h := newTestSessionHandler(mgr)

	// No path param set — Param("id") returns "".
	c, _ := testutil.NewRequestContext(http.MethodDelete, "/account/sessions/", nil)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	err := h.Revoke(c)
	testutil.AssertAppErrorStatus(t, err, http.StatusBadRequest)
}

func TestRevokeNilManagerReturnsUnavailable(t *testing.T) {
	h := newTestSessionHandler(nil)

	c, _ := testutil.NewRequestContext(http.MethodDelete, "/account/sessions/some-id", nil)
	testutil.SetPathParam(c, "/account/sessions/:id", "id", "some-id")
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	err := h.Revoke(c)
	testutil.AssertAppErrorStatus(t, err, http.StatusServiceUnavailable)
}

func TestRevokeServiceErrorReturnsUnavailable(t *testing.T) {
	mgr := newFakeMultiManager(t)
	mgr.destroyErr = errors.New("kv down")

	h := newTestSessionHandler(mgr)

	// Use a cookie-mode manager whose Load always produces a random ID, so
	// "other-session-id" will never match the current session.
	c, _ := testutil.NewRequestContext(http.MethodDelete, "/account/sessions/other-session-id", nil)
	testutil.SetPathParam(c, "/account/sessions/:id", "id", "other-session-id")
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	err := h.Revoke(c)
	testutil.AssertAppErrorStatus(t, err, http.StatusServiceUnavailable)
}

func TestRevokeCurrentSessionRedirects(t *testing.T) {
	// Use server-mode manager so the session ID is stable across two Load calls
	// on requests that share the same cookie.
	mgr, _ := newFakeMultiManagerServer(t)
	currentID, cookie := seedSession(t, mgr)

	h := newTestSessionHandler(mgr)

	// The handler request carries the same cookie, so Load returns the same ID.
	c, rec := testutil.NewRequestContext(http.MethodDelete, "/account/sessions/"+currentID, nil)
	c.Request().AddCookie(cookie)
	testutil.SetPathParam(c, "/account/sessions/:id", "id", currentID)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	if err := h.Revoke(c); err != nil {
		t.Fatalf("Revoke returned unexpected error: %v", err)
	}
	// The handler must NOT call DestroySession for the current session.
	if len(mgr.destroyedIDs) > 0 {
		t.Errorf("expected no DestroySession call for current session, got %v", mgr.destroyedIDs)
	}
	// The handler should redirect (303 for non-HTMX) or set HX-Redirect.
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect for current session protection, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// RevokeAll tests
// ---------------------------------------------------------------------------

func TestRevokeAllDestroysOtherSessions(t *testing.T) {
	// Use server-mode manager so the session cookie gives a stable current ID.
	mgr, _ := newFakeMultiManagerServer(t)
	currentID, cookie := seedSession(t, mgr)

	now := time.Now()
	mgr.listSessions = []session.SessionMeta{
		{SessionID: "s1", CreatedAt: now, LastActiveAt: now},
		{SessionID: currentID, CreatedAt: now.Add(time.Second), LastActiveAt: now},
		{SessionID: "s3", CreatedAt: now.Add(2 * time.Second), LastActiveAt: now},
	}

	h := newTestSessionHandler(mgr)

	c, rec := testutil.NewRequestContext(http.MethodDelete, "/account/sessions", nil)
	c.Request().AddCookie(cookie)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	if err := h.RevokeAll(c); err != nil {
		t.Fatalf("RevokeAll returned unexpected error: %v", err)
	}
	// Must redirect.
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d", rec.Code)
	}
	// Must have destroyed s1 and s3 but NOT the current session.
	if len(mgr.destroyedIDs) != 2 {
		t.Errorf("expected 2 DestroySession calls, got %d: %v", len(mgr.destroyedIDs), mgr.destroyedIDs)
	}
	for _, id := range mgr.destroyedIDs {
		if id == currentID {
			t.Errorf("current session %q must not be destroyed by RevokeAll", currentID)
		}
	}
}

func TestRevokeAllNilManagerReturnsUnavailable(t *testing.T) {
	h := newTestSessionHandler(nil)

	c, _ := testutil.NewRequestContext(http.MethodDelete, "/account/sessions", nil)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	err := h.RevokeAll(c)
	testutil.AssertAppErrorStatus(t, err, http.StatusServiceUnavailable)
}

func TestRevokeAllListErrorReturnsUnavailable(t *testing.T) {
	mgr := newFakeMultiManager(t)
	mgr.listErr = errors.New("kv down")

	h := newTestSessionHandler(mgr)

	c, _ := testutil.NewRequestContext(http.MethodDelete, "/account/sessions", nil)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	err := h.RevokeAll(c)
	testutil.AssertAppErrorStatus(t, err, http.StatusServiceUnavailable)
}

func TestRevokeAllSkipsDestroyOnEmptySessionList(t *testing.T) {
	mgr := newFakeMultiManager(t)
	mgr.listSessions = []session.SessionMeta{}

	h := newTestSessionHandler(mgr)

	c, rec := testutil.NewRequestContext(http.MethodDelete, "/account/sessions", nil)
	testutil.SetUserID(c, "user-1")
	testutil.SetCSRFToken(c)

	if err := h.RevokeAll(c); err != nil {
		t.Fatalf("RevokeAll returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d", rec.Code)
	}
	if len(mgr.destroyedIDs) != 0 {
		t.Errorf("expected no DestroySession calls on empty list, got %v", mgr.destroyedIDs)
	}
}

// ---------------------------------------------------------------------------
// NewModule smoke test
// ---------------------------------------------------------------------------

func TestNewModuleWithServerModeManager(t *testing.T) {
	store := newInMemBlobStore()
	mgr, err := session.NewManager(session.Config{
		Store:      "server",
		CookieName: "_test",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 16),
		MaxAge:     3600,
		BlobStore:  store,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	cfg := &config.Config{
		Security: config.SecurityConfig{
			CSRF: config.CSRFConfig{ContextKey: "csrf"},
		},
	}
	mod := NewModule(cfg, mgr, nil)
	if mod == nil {
		t.Fatal("expected non-nil Module")
	}
	if mod.Handler() == nil {
		t.Fatal("expected non-nil Handler")
	}
	// Server-mode manager satisfies MultiManager; handler.mgr must not be nil.
	if mod.Handler().mgr == nil {
		t.Error("expected handler.mgr to be set for server-mode manager")
	}
}

func TestNewModuleWithCookieModeManagerHasNilMultiManager(t *testing.T) {
	mgr, err := session.NewManager(session.Config{
		Store:      "cookie",
		CookieName: "_test",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 16),
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	cfg := &config.Config{
		Security: config.SecurityConfig{
			CSRF: config.CSRFConfig{ContextKey: "csrf"},
		},
	}
	mod := NewModule(cfg, mgr, nil)
	if mod == nil {
		t.Fatal("expected non-nil Module")
	}
	// Cookie-mode manager does NOT satisfy MultiManager; handler.mgr must be nil.
	if mod.Handler().mgr != nil {
		t.Error("expected handler.mgr to be nil for cookie-mode manager")
	}
}
