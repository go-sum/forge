package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/session"
)

const (
	testAuthKey    = "12345678901234567890123456789012" // 32 bytes
	testEncryptKey = "1234567890123456"                 // 16 bytes
)

func testCookieManager(t *testing.T) session.Manager {
	t.Helper()
	mgr, err := session.NewManager(session.Config{
		Store:    "cookie",
		CookieName: "_test",
		AuthKey:    testAuthKey,
		EncryptKey: testEncryptKey,
		MaxAge:     3600,
		Secure:     false,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return mgr
}

// commitAndLoad performs a Commit, extracts cookies from the response, and
// Loads a fresh State from a new request with those cookies.
func commitAndLoad(t *testing.T, mgr session.Manager, state *session.State) *session.State {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := mgr.Commit(rec, req, state); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	readReq := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		readReq.AddCookie(c)
	}
	loaded, err := mgr.Load(readReq)
	if err != nil {
		t.Fatalf("Load after Commit: %v", err)
	}
	return loaded
}

func TestNewManagerRejectsInvalidAuthKey(t *testing.T) {
	_, err := session.NewManager(session.Config{
		AuthKey:    "tooshort",
		EncryptKey: testEncryptKey,
	})
	if err == nil {
		t.Fatal("expected error for short auth key")
	}
}

func TestNewManagerRejectsInvalidEncryptKey(t *testing.T) {
	_, err := session.NewManager(session.Config{
		AuthKey:    testAuthKey,
		EncryptKey: "short",
	})
	if err == nil {
		t.Fatal("expected error for short encrypt key")
	}
}

func TestNewManagerRequiresBlobStoreForServerBackend(t *testing.T) {
	_, err := session.NewManager(session.Config{
		Store:    "server",
		AuthKey:    testAuthKey,
		EncryptKey: testEncryptKey,
		BlobStore:  nil,
	})
	if err == nil {
		t.Fatal("expected error for server backend without BlobStore")
	}
}

func TestNewManagerRejectsUnknownBackend(t *testing.T) {
	_, err := session.NewManager(session.Config{
		Store:    "unknown",
		AuthKey:    testAuthKey,
		EncryptKey: testEncryptKey,
	})
	if err == nil {
		t.Fatal("expected error for unknown backend")
	}
}

func TestLoadReturnsNewSessionWithoutCookie(t *testing.T) {
	mgr := testCookieManager(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, err := mgr.Load(req)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !state.IsNew() {
		t.Fatal("expected new session without cookie")
	}
}

func TestStatePutGetRoundtrip(t *testing.T) {
	mgr := testCookieManager(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)

	if err := state.Put("auth.user_id", "user-123"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	loaded := commitAndLoad(t, mgr, state)

	var got string
	ok, err := loaded.Get("auth.user_id", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("key not found after roundtrip")
	}
	if got != "user-123" {
		t.Fatalf("Get = %q, want %q", got, "user-123")
	}
}

func TestStateMultipleKeysPersist(t *testing.T) {
	mgr := testCookieManager(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)

	_ = state.Put("auth.user_id", "u1")
	_ = state.Put("auth.display_name", "Alice")
	_ = state.Put("ui.theme", "dark")

	loaded := commitAndLoad(t, mgr, state)

	var uid, name, theme string
	ok1, _ := loaded.Get("auth.user_id", &uid)
	ok2, _ := loaded.Get("auth.display_name", &name)
	ok3, _ := loaded.Get("ui.theme", &theme)

	if !ok1 || uid != "u1" {
		t.Fatalf("user_id: ok=%v, val=%q", ok1, uid)
	}
	if !ok2 || name != "Alice" {
		t.Fatalf("display_name: ok=%v, val=%q", ok2, name)
	}
	if !ok3 || theme != "dark" {
		t.Fatalf("theme: ok=%v, val=%q", ok3, theme)
	}
}

func TestStateDeleteRemovesKey(t *testing.T) {
	mgr := testCookieManager(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)

	_ = state.Put("auth.user_id", "u1")
	_ = state.Put("auth.pending_flow", "flow-data")
	state.Delete("auth.pending_flow")

	loaded := commitAndLoad(t, mgr, state)

	if loaded.Has("auth.pending_flow") {
		t.Fatal("deleted key should not exist after roundtrip")
	}
	var uid string
	ok, _ := loaded.Get("auth.user_id", &uid)
	if !ok || uid != "u1" {
		t.Fatal("non-deleted key should survive")
	}
}

func TestStateHas(t *testing.T) {
	mgr := testCookieManager(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)

	if state.Has("missing") {
		t.Fatal("Has should be false for missing key")
	}
	_ = state.Put("present", "yes")
	if !state.Has("present") {
		t.Fatal("Has should be true after Put")
	}
}

func TestStateGetMissingKeyReturnsFalse(t *testing.T) {
	mgr := testCookieManager(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)

	var v string
	ok, err := state.Get("missing", &v)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ok {
		t.Fatal("Get should return false for missing key")
	}
}

func TestDestroyExpiresCookie(t *testing.T) {
	mgr := testCookieManager(t)

	// Create a session.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)
	_ = state.Put("auth.user_id", "u1")
	rec := httptest.NewRecorder()
	_ = mgr.Commit(rec, req, state)

	// Destroy it.
	destroyReq := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		destroyReq.AddCookie(c)
	}
	destroyRec := httptest.NewRecorder()
	if err := mgr.Destroy(destroyRec, destroyReq); err != nil {
		t.Fatalf("Destroy: %v", err)
	}

	// Cookie should have MaxAge=-1.
	found := false
	for _, c := range destroyRec.Result().Cookies() {
		if c.Name == "_test" {
			found = true
			if c.MaxAge != -1 {
				t.Fatalf("Destroy cookie MaxAge = %d, want -1", c.MaxAge)
			}
		}
	}
	if !found {
		t.Fatal("Destroy did not set cookie")
	}

	// Load from destroyed cookie should give new session.
	newReq := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range destroyRec.Result().Cookies() {
		newReq.AddCookie(c)
	}
	newState, _ := mgr.Load(newReq)
	if !newState.IsNew() {
		t.Fatal("Load after Destroy should return new session")
	}
}

func TestStatePutComplexType(t *testing.T) {
	mgr := testCookieManager(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)

	type flow struct {
		Purpose string `json:"purpose"`
		Email   string `json:"email"`
	}
	_ = state.Put("auth.pending_flow", flow{Purpose: "signin", Email: "a@b.com"})

	loaded := commitAndLoad(t, mgr, state)

	var got flow
	ok, err := loaded.Get("auth.pending_flow", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("key not found")
	}
	if got.Purpose != "signin" || got.Email != "a@b.com" {
		t.Fatalf("Get = %+v, want signin/a@b.com", got)
	}
}

func TestCookiePolicyAttributes(t *testing.T) {
	mgr := testCookieManager(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, _ := mgr.Load(req)
	_ = state.Put("k", "v")

	rec := httptest.NewRecorder()
	_ = mgr.Commit(rec, req, state)

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookie set")
	}
	c := cookies[0]
	if c.Name != "_test" {
		t.Fatalf("cookie name = %q, want _test", c.Name)
	}
	if !c.HttpOnly {
		t.Fatal("cookie should be HttpOnly")
	}
	if c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("SameSite = %v, want Strict", c.SameSite)
	}
	if c.Path != "/" {
		t.Fatalf("Path = %q, want /", c.Path)
	}
	if c.MaxAge != 3600 {
		t.Fatalf("MaxAge = %d, want 3600", c.MaxAge)
	}
}
