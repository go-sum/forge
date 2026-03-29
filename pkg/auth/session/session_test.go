package session

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

func TestSessionManagerUserRoundTripAndClear(t *testing.T) {
	manager, err := NewSessionStore(SessionConfig{
		Name:       "test-session",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
		MaxAge:     3600,
	})
	if err != nil {
		t.Fatalf("NewSessionStore() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := manager.SetUserID(rec, req, "user-123"); err != nil {
		t.Fatalf("SetUserID() error = %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	userID, err := manager.GetUserID(req)
	if err != nil || userID != "user-123" {
		t.Fatalf("GetUserID() userID=%q err=%v", userID, err)
	}

	rec = httptest.NewRecorder()
	if err := manager.Clear(rec, req); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if !strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=0") &&
		!strings.Contains(rec.Header().Get("Set-Cookie"), "Max-Age=-1") {
		t.Fatalf("set-cookie = %q", rec.Header().Get("Set-Cookie"))
	}
}

func TestSessionManagerDefaultsAndMissingUser(t *testing.T) {
	manager, err := NewSessionStore(SessionConfig{
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
	})
	if err != nil {
		t.Fatalf("NewSessionStore() error = %v", err)
	}
	if manager.name != "session" {
		t.Fatalf("default name = %q", manager.name)
	}
	if _, err := manager.GetUserID(httptest.NewRequest(http.MethodGet, "/", nil)); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("GetUserID() err = %v", err)
	}
	if _, err := manager.GetPendingFlow(httptest.NewRequest(http.MethodGet, "/", nil)); !errors.Is(err, ErrPendingFlowNotFound) {
		t.Fatalf("GetPendingFlow() err = %v", err)
	}
}

func TestSessionManagerPendingFlowRoundTripAndClear(t *testing.T) {
	manager, err := NewSessionStore(SessionConfig{
		Name:       "test-session",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
		MaxAge:     3600,
	})
	if err != nil {
		t.Fatalf("NewSessionStore() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	flow := model.PendingFlow{
		Purpose:   model.FlowPurposeEmailChange,
		Email:     "new@example.com",
		UserID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	}
	if err := manager.SetPendingFlow(rec, req, flow); err != nil {
		t.Fatalf("SetPendingFlow() error = %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	stored, err := manager.GetPendingFlow(req)
	if err != nil || stored != flow {
		t.Fatalf("GetPendingFlow() flow=%#v err=%v", stored, err)
	}

	rec = httptest.NewRecorder()
	if err := manager.ClearPendingFlow(rec, req); err != nil {
		t.Fatalf("ClearPendingFlow() error = %v", err)
	}
}

func TestSessionManagerDisplayNameRoundTrip(t *testing.T) {
	manager, err := NewSessionStore(SessionConfig{
		Name:       "test-session",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
		MaxAge:     3600,
	})
	if err != nil {
		t.Fatalf("NewSessionStore() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := manager.SetUserID(rec, req, "user-123"); err != nil {
		t.Fatalf("SetUserID() error = %v", err)
	}
	if err := manager.SetDisplayName(rec, req, "Ada Lovelace"); err != nil {
		t.Fatalf("SetDisplayName() error = %v", err)
	}

	// Use only the last Set-Cookie header — simulates browser behaviour where the
	// most recent Set-Cookie for a given name overwrites the previous value.
	cookies := rec.Result().Cookies()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookies[len(cookies)-1])
	name, err := manager.GetDisplayName(req)
	if err != nil || name != "Ada Lovelace" {
		t.Fatalf("GetDisplayName() name=%q err=%v", name, err)
	}
}

func TestSessionManagerGetDisplayNameMissingReturnsError(t *testing.T) {
	manager, err := NewSessionStore(SessionConfig{
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
	})
	if err != nil {
		t.Fatalf("NewSessionStore() error = %v", err)
	}
	if _, err := manager.GetDisplayName(httptest.NewRequest(http.MethodGet, "/", nil)); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("GetDisplayName() err = %v, want ErrNotAuthenticated", err)
	}
}

func TestNewSessionStoreRejectsInvalidKeys(t *testing.T) {
	tests := []struct {
		name       string
		authKey    string
		encryptKey string
	}{
		{"auth key too short", strings.Repeat("a", 16), strings.Repeat("b", 32)},
		{"auth key wrong length", strings.Repeat("a", 48), strings.Repeat("b", 32)},
		{"encrypt key too short", strings.Repeat("a", 32), strings.Repeat("b", 8)},
		{"encrypt key wrong length", strings.Repeat("a", 32), strings.Repeat("b", 48)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewSessionStore(SessionConfig{AuthKey: tt.authKey, EncryptKey: tt.encryptKey}); err == nil {
				t.Fatal("NewSessionStore() error = nil, want error")
			}
		})
	}
}

func TestSessionManagerPropagatesStoreErrors(t *testing.T) {
	manager := &SessionManager{name: "test-session", store: failingStore{err: errors.New("store failure")}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	if err := manager.SetUserID(rec, req, "user-123"); err == nil {
		t.Fatal("SetUserID() unexpectedly succeeded")
	}
	if err := manager.Clear(rec, req); err == nil {
		t.Fatal("Clear() unexpectedly succeeded")
	}
}

type failingStore struct {
	err error
}

func (s failingStore) Get(*http.Request, string) (*sessions.Session, error) {
	return nil, s.err
}

func (s failingStore) New(*http.Request, string) (*sessions.Session, error) {
	return sessions.NewSession(s, "test-session"), nil
}

func (s failingStore) Save(*http.Request, http.ResponseWriter, *sessions.Session) error {
	return s.err
}
