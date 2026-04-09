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
	"cmp"
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
	Store       string        // "cookie" (default) or "server"
	CookieName  string        // default "_session"
	AuthKey     string        // HMAC signing key (32 or 64 bytes)
	EncryptKey  string        // AES encryption key (16, 24, or 32 bytes)
	MaxAge      int           // session TTL in seconds
	Secure      bool          // cookie Secure flag
	SameSite    string        // "strict" (default), "lax", or "none"
	BlobStore   BlobStore     // required when Store="server"; nil otherwise
	KeyPrefix   string        // prefix for session data blobs (default "session:")
	UserPrefix  string        // prefix for user index blobs (default "user_sessions:")
	MetaPrefix  string        // prefix for per-session metadata blobs (default "session_meta:")
	TouchWindow time.Duration // minimum interval between TouchSession KV writes (default 60s)
}

// defaultConfig holds the zero-omitted defaults applied by NewManager.
// Edit here to change package-wide defaults.
var defaultConfig = Config{
	Store:       "cookie",
	CookieName:  "_session",
	MaxAge:      86400, // 24h in seconds
	KeyPrefix:   "session:",
	UserPrefix:  "user_sessions:",
	MetaPrefix:  "session_meta:",
	TouchWindow: 60 * time.Second,
}

// NewManager constructs a Manager with the configured backend.
func NewManager(cfg Config) (Manager, error) {
	cookieName := cmp.Or(cfg.CookieName, defaultConfig.CookieName)
	maxAge := time.Duration(cmp.Or(cfg.MaxAge, defaultConfig.MaxAge)) * time.Second

	if err := validateKeys(cfg.AuthKey, cfg.EncryptKey); err != nil {
		return nil, err
	}

	backend := cmp.Or(cfg.Store, defaultConfig.Store)

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
		keyPrefix := cmp.Or(cfg.KeyPrefix, defaultConfig.KeyPrefix)
		m.store = newServerStore(cfg.BlobStore, keyPrefix)
		// Server store still needs cookie codec for signing the session ID cookie
		cb, err := newCookieStore([]byte(cfg.AuthKey), []byte(cfg.EncryptKey))
		if err != nil {
			return nil, err
		}
		m.cookie = cb
		m.isServer = true

		return &multiManager{
			manager:     m,
			blobStore:   cfg.BlobStore,
			keyPrefix:   keyPrefix,
			userPrefix:  cmp.Or(cfg.UserPrefix, defaultConfig.UserPrefix),
			metaPrefix:  cmp.Or(cfg.MetaPrefix, defaultConfig.MetaPrefix),
			touchWindow: cmp.Or(cfg.TouchWindow, defaultConfig.TouchWindow),
		}, nil

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
