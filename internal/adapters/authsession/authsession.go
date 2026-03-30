// Package authsession provides session helpers for the auth domain and
// the KV store adapter used to back server-side session storage.
package authsession

import (
	"github.com/go-sum/auth/model"
	"github.com/go-sum/session"
)

const (
	keyUserID      = "auth.user_id"
	keyDisplayName = "auth.display_name"
	keyPendingFlow = "auth.pending_flow"
)

// SetAuth writes the authenticated user's ID and display name into the session
// and clears any pending verification flow.
func SetAuth(s *session.State, userID, displayName string) error {
	if err := s.Put(keyUserID, userID); err != nil {
		return err
	}
	if err := s.Put(keyDisplayName, displayName); err != nil {
		return err
	}
	s.Delete(keyPendingFlow)
	return nil
}

// GetUserID returns the authenticated user ID from the session.
func GetUserID(s *session.State) (string, bool) {
	var id string
	ok, _ := s.Get(keyUserID, &id)
	return id, ok && id != ""
}

// GetDisplayName returns the authenticated user's display name from the session.
func GetDisplayName(s *session.State) (string, bool) {
	var name string
	ok, _ := s.Get(keyDisplayName, &name)
	return name, ok && name != ""
}

// SetPendingFlow stores a pending verification flow in the session.
func SetPendingFlow(s *session.State, flow model.PendingFlow) error {
	return s.Put(keyPendingFlow, flow)
}

// GetPendingFlow retrieves the pending verification flow from the session.
func GetPendingFlow(s *session.State) (model.PendingFlow, bool) {
	var flow model.PendingFlow
	ok, _ := s.Get(keyPendingFlow, &flow)
	return flow, ok
}
