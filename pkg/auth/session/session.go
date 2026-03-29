package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-sum/auth/model"
	"github.com/gorilla/sessions"
)

// ErrNotAuthenticated is returned when the session contains no user ID.
var ErrNotAuthenticated = errors.New("not authenticated")

// ErrPendingFlowNotFound is returned when no pending verification flow is stored.
var ErrPendingFlowNotFound = errors.New("pending auth flow not found")

// sessionKeyUserID is the key used to store the authenticated user ID in the session.
const sessionKeyUserID = "user_id"

const sessionKeyPendingFlow = "pending_flow"

// SessionConfig holds cookie store configuration.
type SessionConfig struct {
	Name       string
	AuthKey    string // HMAC key, 32 or 64 bytes
	EncryptKey string // AES key, 16, 24, or 32 bytes
	MaxAge     int
	Secure     bool
}

// SessionManager wraps a gorilla CookieStore and provides typed session operations.
type SessionManager struct {
	name  string
	store sessions.Store
}

// NewSessionStore creates a SessionManager backed by a signed+encrypted cookie store.
// Returns an error if the key lengths are invalid:
//   - AuthKey must be 32 or 64 bytes (HMAC-SHA256 or HMAC-SHA512)
//   - EncryptKey must be 16, 24, or 32 bytes (AES-128, AES-192, or AES-256)
func NewSessionStore(cfg SessionConfig) (*SessionManager, error) {
	switch len(cfg.AuthKey) {
	case 32, 64:
	default:
		return nil, fmt.Errorf("session: auth_key must be 32 or 64 bytes, got %d", len(cfg.AuthKey))
	}
	switch len(cfg.EncryptKey) {
	case 16, 24, 32:
	default:
		return nil, fmt.Errorf("session: encrypt_key must be 16, 24, or 32 bytes, got %d", len(cfg.EncryptKey))
	}

	store := sessions.NewCookieStore([]byte(cfg.AuthKey), []byte(cfg.EncryptKey))
	store.Options = &sessions.Options{
		MaxAge:   cfg.MaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   cfg.Secure,
		Path:     "/",
	}
	name := cfg.Name
	if name == "" {
		name = "session"
	}
	return &SessionManager{name: name, store: store}, nil
}

// SetUserID stores the user ID in the session cookie.
func (m *SessionManager) SetUserID(w http.ResponseWriter, r *http.Request, userID string) error {
	session, err := m.store.Get(r, m.name)
	if err != nil {
		return err
	}
	session.Values[sessionKeyUserID] = userID
	return session.Save(r, w)
}

// GetUserID reads the user ID from the session cookie.
// Returns ("", ErrNotAuthenticated) when no user ID is present.
func (m *SessionManager) GetUserID(r *http.Request) (string, error) {
	session, err := m.store.Get(r, m.name)
	if err != nil {
		return "", ErrNotAuthenticated
	}
	v, ok := session.Values[sessionKeyUserID]
	if !ok {
		return "", ErrNotAuthenticated
	}
	userID, ok := v.(string)
	if !ok || userID == "" {
		return "", ErrNotAuthenticated
	}
	return userID, nil
}

// SetPendingFlow stores the browser-bound verification flow in the session cookie.
func (m *SessionManager) SetPendingFlow(w http.ResponseWriter, r *http.Request, flow model.PendingFlow) error {
	session, err := m.store.Get(r, m.name)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(flow)
	if err != nil {
		return err
	}
	session.Values[sessionKeyPendingFlow] = string(encoded)
	return session.Save(r, w)
}

// GetPendingFlow reads the pending verification flow from the session cookie.
func (m *SessionManager) GetPendingFlow(r *http.Request) (model.PendingFlow, error) {
	session, err := m.store.Get(r, m.name)
	if err != nil {
		return model.PendingFlow{}, ErrPendingFlowNotFound
	}
	v, ok := session.Values[sessionKeyPendingFlow]
	if !ok {
		return model.PendingFlow{}, ErrPendingFlowNotFound
	}
	raw, ok := v.(string)
	if !ok || raw == "" {
		return model.PendingFlow{}, ErrPendingFlowNotFound
	}
	var flow model.PendingFlow
	if err := json.Unmarshal([]byte(raw), &flow); err != nil {
		return model.PendingFlow{}, ErrPendingFlowNotFound
	}
	return flow, nil
}

// ClearPendingFlow removes the pending verification flow from the session.
func (m *SessionManager) ClearPendingFlow(w http.ResponseWriter, r *http.Request) error {
	session, err := m.store.Get(r, m.name)
	if err != nil {
		return err
	}
	delete(session.Values, sessionKeyPendingFlow)
	return session.Save(r, w)
}

// Clear invalidates the session by setting MaxAge to -1.
func (m *SessionManager) Clear(w http.ResponseWriter, r *http.Request) error {
	session, err := m.store.Get(r, m.name)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
