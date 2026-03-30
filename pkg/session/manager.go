package session

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// manager implements the Manager interface.
type manager struct {
	store      Store
	cookie     *cookieStore // non-nil only for cookie store (encode/decode)
	cookieName string
	maxAge     time.Duration
	secure     bool
	isServer   bool // true when using server store (BlobStore)
}

func (m *manager) Load(r *http.Request) (*State, error) {
	c, err := r.Cookie(m.cookieName)
	if err != nil {
		// No cookie — new session
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		return newState(id, true), nil
	}

	if m.isServer {
		return m.loadServer(c.Value)
	}
	return m.loadCookie(c.Value)
}

func (m *manager) loadServer(sessionID string) (*State, error) {
	data, exists, err := m.store.Load(sessionID)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Session expired or deleted — create new
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		return newState(id, true), nil
	}
	state := newState(sessionID, false)
	if err := state.decode(data); err != nil {
		// Corrupt data — create new session
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		return newState(id, true), nil
	}
	return state, nil
}

func (m *manager) loadCookie(cookieValue string) (*State, error) {
	data, err := m.cookie.Decode(m.cookieName, cookieValue)
	if err != nil {
		// Invalid/expired cookie — create new session
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		return newState(id, true), nil
	}
	state := newState("cookie", false)
	if err := state.decode(data); err != nil {
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		return newState(id, true), nil
	}
	return state, nil
}

func (m *manager) Commit(w http.ResponseWriter, r *http.Request, s *State) error {
	data, err := s.encode()
	if err != nil {
		return err
	}

	if m.isServer {
		if err := m.store.Save(s.id, data, m.maxAge); err != nil {
			return err
		}
		m.setCookie(w, s.id)
	} else {
		encoded, err := m.cookie.Encode(m.cookieName, data)
		if err != nil {
			return err
		}
		m.setCookie(w, encoded)
	}
	s.dirty = false
	s.isNew = false
	return nil
}

func (m *manager) Destroy(w http.ResponseWriter, r *http.Request) error {
	c, err := r.Cookie(m.cookieName)
	if err == nil && m.isServer && c.Value != "" {
		_ = m.store.Delete(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   m.secure,
	})
	return nil
}

func (m *manager) RotateID(w http.ResponseWriter, r *http.Request, s *State) error {
	if m.isServer {
		// Delete old session from backend
		_ = m.store.Delete(s.id)
	}
	newID, err := generateID()
	if err != nil {
		return err
	}
	s.id = newID
	s.dirty = true
	return m.Commit(w, r, s)
}

func (m *manager) setCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(m.maxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   m.secure,
	})
}

func generateID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
