package auth

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
)

// fakeSessionState is an in-memory SessionState backed by a map.
// encoding/json is used so that typed Get roundtrips work identically to
// the real session implementation.
type fakeSessionState struct {
	data map[string]any
}

func newFakeSessionState() *fakeSessionState {
	return &fakeSessionState{data: make(map[string]any)}
}

func (s *fakeSessionState) ID() string { return "fake-session-id" }

func (s *fakeSessionState) Get(key string, dst any) (bool, error) {
	v, ok := s.data[key]
	if !ok {
		return false, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(b, dst)
}

func (s *fakeSessionState) Put(key string, v any) error {
	s.data[key] = v
	return nil
}

func (s *fakeSessionState) Delete(key string) {
	delete(s.data, key)
}

func TestSetAuthRoundtrip(t *testing.T) {
	state := newFakeSessionState()
	if err := setAuth(state, "11111111-1111-1111-1111-111111111111", "Ada Lovelace"); err != nil {
		t.Fatalf("setAuth() error = %v", err)
	}

	id, ok := getUserID(state)
	if !ok || id != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("getUserID() = %q, %v", id, ok)
	}

	name, ok := getDisplayName(state)
	if !ok || name != "Ada Lovelace" {
		t.Fatalf("getDisplayName() = %q, %v", name, ok)
	}
}

func TestSetAuthEmptyDisplayName(t *testing.T) {
	state := newFakeSessionState()
	if err := setAuth(state, "11111111-1111-1111-1111-111111111111", ""); err != nil {
		t.Fatalf("setAuth() error = %v", err)
	}

	_, ok := getDisplayName(state)
	if ok {
		t.Fatal("getDisplayName() should return false for empty display name")
	}
}

func TestGetUserIDMissingKey(t *testing.T) {
	state := newFakeSessionState()
	_, ok := getUserID(state)
	if ok {
		t.Fatal("getUserID() should return false when no user ID is stored")
	}
}

func TestSetPendingFlowRoundtrip(t *testing.T) {
	state := newFakeSessionState()
	flow := model.PendingFlow{
		Purpose:   model.FlowPurposeSignin,
		Email:     "ada@example.com",
		UserID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	}

	if err := setPendingFlow(state, flow); err != nil {
		t.Fatalf("setPendingFlow() error = %v", err)
	}

	got, ok := getPendingFlow(state)
	if !ok {
		t.Fatal("getPendingFlow() returned false")
	}
	if got.Email != flow.Email || got.Purpose != flow.Purpose || got.Secret != flow.Secret {
		t.Fatalf("getPendingFlow() = %#v, want %#v", got, flow)
	}
}

func TestGetPendingFlowMissingKey(t *testing.T) {
	state := newFakeSessionState()
	_, ok := getPendingFlow(state)
	if ok {
		t.Fatal("getPendingFlow() should return false when no flow is stored")
	}
}
