package session

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

// ErrNotAuthenticated is returned when the session contains no user ID.
var ErrNotAuthenticated = errors.New("not authenticated")

// sessionKeyUserID is the key used to store the authenticated user ID in the session.
const sessionKeyUserID = "user_id"

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
func NewSessionStore(cfg SessionConfig) *SessionManager {
	store := sessions.NewCookieStore([]byte(cfg.AuthKey), []byte(cfg.EncryptKey))
	store.Options = &sessions.Options{
		MaxAge:   cfg.MaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   cfg.Secure,
		Path:     "/",
	}
	name := cfg.Name
	if name == "" {
		name = "session"
	}
	return &SessionManager{name: name, store: store}
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



// Clear invalidates the session by setting MaxAge to -1.
func (m *SessionManager) Clear(w http.ResponseWriter, r *http.Request) error {
	session, err := m.store.Get(r, m.name)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
