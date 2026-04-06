package auth

import "github.com/go-sum/auth/model"

const (
	sessionKeyUserID      = "auth.user_id"
	sessionKeyDisplayName = "auth.display_name"
	sessionKeyPendingFlow = "auth.pending_flow"
)

func setAuth(s SessionState, userID, displayName string) error {
	if err := s.Put(sessionKeyUserID, userID); err != nil {
		return err
	}
	if err := s.Put(sessionKeyDisplayName, displayName); err != nil {
		return err
	}
	s.Delete(sessionKeyPendingFlow)
	return nil
}

func getUserID(s SessionState) (string, bool) {
	var id string
	ok, _ := s.Get(sessionKeyUserID, &id)
	return id, ok && id != ""
}

func getDisplayName(s SessionState) (string, bool) {
	var name string
	ok, _ := s.Get(sessionKeyDisplayName, &name)
	return name, ok && name != ""
}

func setPendingFlow(s SessionState, flow model.PendingFlow) error {
	return s.Put(sessionKeyPendingFlow, flow)
}

func getPendingFlow(s SessionState) (model.PendingFlow, bool) {
	var flow model.PendingFlow
	ok, _ := s.Get(sessionKeyPendingFlow, &flow)
	return flow, ok
}
