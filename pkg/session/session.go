// Package session provides a generic session state engine.
//
// It manages the lifecycle of browser-bound state: Load, mutate, Commit, Destroy.
// The package knows nothing about auth, CSRF, theme, or any specific consumer —
// consumers operate on *State using namespaced keys.
//
// Two stores are supported:
//   - "cookie": entire session stored in a signed+encrypted cookie (gorilla/securecookie)
//   - "server": cookie carries only an opaque session ID; data persisted via BlobStore
//
// The package does not create any external connections. Server-side persistence is
// injected via the BlobStore interface, which callers satisfy with an adapter.
package session

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Manager defines the session lifecycle operations.
type Manager interface {
	// Load reads session state from the request. Returns a new empty State if
	// no valid session exists.
	Load(r *http.Request) (*State, error)

	// Commit persists the session state and sets the response cookie.
	Commit(w http.ResponseWriter, r *http.Request, s *State) error

	// Destroy deletes the session from the store and expires the cookie.
	Destroy(w http.ResponseWriter, r *http.Request) error

	// RotateID generates a new session ID, migrates state, and persists.
	// Use after authentication to prevent session fixation.
	RotateID(w http.ResponseWriter, r *http.Request, s *State) error
}

// Config holds all configuration needed to construct a Manager.
type Config struct {
	Store      string    // "cookie" (default) or "server"
	CookieName string    // default "_session"
	AuthKey    string    // HMAC signing key (32 or 64 bytes)
	EncryptKey string    // AES encryption key (16, 24, or 32 bytes)
	MaxAge     int       // session TTL in seconds
	Secure     bool      // cookie Secure flag
	SameSite   string    // "strict" (default), "lax", or "none"
	BlobStore  BlobStore // required when Store="server"; nil otherwise
	KeyPrefix  string    // prefix for blob keys (default "session:")
}

// NewManager constructs a Manager with the configured backend.
func NewManager(cfg Config) (Manager, error) {
	cookieName := cfg.CookieName
	if cookieName == "" {
		cookieName = "_session"
	}
	maxAge := time.Duration(cfg.MaxAge) * time.Second
	if maxAge == 0 {
		maxAge = 24 * time.Hour
	}

	if err := validateKeys(cfg.AuthKey, cfg.EncryptKey); err != nil {
		return nil, err
	}

	backend := cfg.Store
	if backend == "" {
		backend = "cookie"
	}

	m := &manager{
		cookieName: cookieName,
		maxAge:     maxAge,
		secure:     cfg.Secure,
		sameSite:   parseSameSite(cfg.SameSite),
	}

	switch backend {
	case "cookie":
		cb, err := newCookieStore([]byte(cfg.AuthKey), []byte(cfg.EncryptKey))
		if err != nil {
			return nil, err
		}
		m.store = cb
		m.cookie = cb
		m.isServer = false

	case "server":
		if cfg.BlobStore == nil {
			return nil, errors.New("session: backend=server requires a BlobStore")
		}
		m.store = newServerStore(cfg.BlobStore, cfg.KeyPrefix)
		// Server store still needs cookie codec for signing the session ID cookie
		cb, err := newCookieStore([]byte(cfg.AuthKey), []byte(cfg.EncryptKey))
		if err != nil {
			return nil, err
		}
		m.cookie = cb
		m.isServer = true

	default:
		return nil, fmt.Errorf("session: unknown backend %q", backend)
	}

	return m, nil
}

// Converts a config string to http.SameSite.
func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}

func validateKeys(authKey, encryptKey string) error {
	switch len(authKey) {
	case 32, 64:
	default:
		return fmt.Errorf("session: auth_key must be 32 or 64 bytes, got %d", len(authKey))
	}
	switch len(encryptKey) {
	case 16, 24, 32:
	default:
		return fmt.Errorf("session: encrypt_key must be 16, 24, or 32 bytes, got %d", len(encryptKey))
	}
	return nil
}
