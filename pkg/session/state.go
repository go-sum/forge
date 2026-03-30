package session

import (
	"encoding/json"
)

// State holds session data as namespaced key-value pairs.
// Consumers use Get/Put/Delete with convention-namespaced keys
// (e.g. "auth.user_id", "security.csrf", "ui.theme").
type State struct {
	id     string
	values map[string]json.RawMessage
	dirty  bool
	isNew  bool
}

func newState(id string, isNew bool) *State {
	return &State{
		id:     id,
		values: make(map[string]json.RawMessage),
		isNew:  isNew,
	}
}

// ID returns the opaque session identifier.
func (s *State) ID() string { return s.id }

// IsNew returns true if the session was freshly created (not loaded from storage).
func (s *State) IsNew() bool { return s.isNew }

// Get deserializes the value for key into dst. Returns false if key is absent.
func (s *State) Get(key string, dst any) (bool, error) {
	raw, ok := s.values[key]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return false, err
	}
	return true, nil
}

// Put serializes v and stores it under key. Marks the state as dirty.
func (s *State) Put(key string, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	s.values[key] = raw
	s.dirty = true
	return nil
}

// Delete removes the key. Marks the state as dirty.
func (s *State) Delete(key string) {
	if _, ok := s.values[key]; ok {
		delete(s.values, key)
		s.dirty = true
	}
}

// Has returns true if key exists.
func (s *State) Has(key string) bool {
	_, ok := s.values[key]
	return ok
}

// encode serializes the full state to JSON for persistence.
func (s *State) encode() ([]byte, error) {
	return json.Marshal(s.values)
}

// decode populates the state from persisted JSON.
func (s *State) decode(data []byte) error {
	return json.Unmarshal(data, &s.values)
}
