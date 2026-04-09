package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"sync"
	"time"
)

// SessionMeta holds metadata about a single session for a user.
type SessionMeta struct {
	SessionID    string    `json:"session_id"`
	AuthMethod   string    `json:"auth_method,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
}

// MultiManager is a superset of Manager that maintains a per-user index of
// session IDs and per-session metadata blobs.
type MultiManager interface {
	Manager
	BindUser(ctx context.Context, sessionID, userID string, meta SessionMeta) error
	UnbindUser(ctx context.Context, sessionID, userID string) error
	ListUserSessions(ctx context.Context, userID string) ([]SessionMeta, error)
	DestroySession(ctx context.Context, sessionID, userID string) error
	DestroyUserSessions(ctx context.Context, userID string) error
	TouchSession(ctx context.Context, sessionID, userID string) error
}

type multiManager struct {
	*manager
	blobStore   BlobStore
	keyPrefix   string        // for session data blobs, e.g. "session:"
	userPrefix  string        // for user index blobs, e.g. "user_sessions:"
	metaPrefix  string        // for per-session metadata blobs, e.g. "session_meta:"
	touchFence  sync.Map      // sessionID → time.Time of last committed touch
	touchWindow time.Duration // minimum interval between KV writes in TouchSession
}

// userIndex is the per-user set of active session IDs.
// Metadata for each session lives at its own key (metaPrefix+sessionID).
type userIndex struct {
	Sessions []string `json:"sessions"`
}

func (idx *userIndex) add(sessionID string) {
	if slices.Contains(idx.Sessions, sessionID) {
		return
	}
	idx.Sessions = append(idx.Sessions, sessionID)
}

func (idx *userIndex) remove(sessionID string) {
	for i, id := range idx.Sessions {
		if id == sessionID {
			idx.Sessions = append(idx.Sessions[:i], idx.Sessions[i+1:]...)
			return
		}
	}
}

func (mm *multiManager) loadUserIndex(ctx context.Context, userID string) (userIndex, error) {
	data, err := mm.blobStore.Get(ctx, mm.userPrefix+userID)
	if err != nil {
		if errors.Is(err, ErrBlobNotFound) {
			return userIndex{Sessions: []string{}}, nil
		}
		return userIndex{}, fmt.Errorf("multiManager.loadUserIndex: %w", err)
	}
	var idx userIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return userIndex{}, fmt.Errorf("multiManager.loadUserIndex: %w", err)
	}
	if idx.Sessions == nil {
		idx.Sessions = []string{}
	}
	return idx, nil
}

func (mm *multiManager) saveUserIndex(ctx context.Context, userID string, idx userIndex) error {
	if len(idx.Sessions) == 0 {
		return mm.blobStore.Delete(ctx, mm.userPrefix+userID)
	}
	data, err := json.Marshal(idx)
	if err != nil {
		return fmt.Errorf("multiManager.saveUserIndex: %w", err)
	}
	return mm.blobStore.Set(ctx, mm.userPrefix+userID, data, mm.manager.maxAge)
}

func (mm *multiManager) loadSessionMeta(ctx context.Context, sessionID string) (SessionMeta, error) {
	data, err := mm.blobStore.Get(ctx, mm.metaPrefix+sessionID)
	if err != nil {
		return SessionMeta{}, err
	}
	var meta SessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return SessionMeta{}, fmt.Errorf("multiManager.loadSessionMeta: %w", err)
	}
	return meta, nil
}

func (mm *multiManager) saveSessionMeta(ctx context.Context, sessionID string, meta SessionMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("multiManager.saveSessionMeta: %w", err)
	}
	return mm.blobStore.Set(ctx, mm.metaPrefix+sessionID, data, mm.manager.maxAge)
}

// BindUser upserts a session entry into the user's index and writes the session
// metadata blob. If metadata already exists for the session, CreatedAt is preserved.
func (mm *multiManager) BindUser(ctx context.Context, sessionID, userID string, meta SessionMeta) error {
	if existing, err := mm.loadSessionMeta(ctx, sessionID); err == nil {
		meta.CreatedAt = existing.CreatedAt
	}
	meta.LastActiveAt = time.Now()
	if err := mm.saveSessionMeta(ctx, sessionID, meta); err != nil {
		return err
	}
	idx, err := mm.loadUserIndex(ctx, userID)
	if err != nil {
		return err
	}
	idx.add(sessionID)
	return mm.saveUserIndex(ctx, userID, idx)
}

// UnbindUser removes a session from the user's index and deletes its metadata blob.
// Returns nil even if the session or user is not found (idempotent).
func (mm *multiManager) UnbindUser(ctx context.Context, sessionID, userID string) error {
	idx, err := mm.loadUserIndex(ctx, userID)
	if err != nil {
		return err
	}
	idx.remove(sessionID)
	if err := mm.saveUserIndex(ctx, userID, idx); err != nil {
		return err
	}
	mm.touchFence.Delete(sessionID)
	_ = mm.blobStore.Delete(ctx, mm.metaPrefix+sessionID)
	return nil
}

// ListUserSessions returns all active sessions for a user, sorted by CreatedAt ascending.
// Stale entries (where the session blob no longer exists) are pruned before returning.
func (mm *multiManager) ListUserSessions(ctx context.Context, userID string) ([]SessionMeta, error) {
	idx, err := mm.loadUserIndex(ctx, userID)
	if err != nil {
		return nil, err
	}

	var keepIDs []string
	var result []SessionMeta
	stale := false

	for _, sessionID := range idx.Sessions {
		_, getErr := mm.blobStore.Get(ctx, mm.keyPrefix+sessionID)
		if getErr != nil {
			if !errors.Is(getErr, ErrBlobNotFound) {
				return nil, fmt.Errorf("multiManager.ListUserSessions: %w", getErr)
			}
			// Session blob is gone (expired or deleted) — truly stale.
			_ = mm.blobStore.Delete(ctx, mm.metaPrefix+sessionID)
			stale = true
			continue
		}
		meta, metaErr := mm.loadSessionMeta(ctx, sessionID)
		if metaErr != nil {
			if !errors.Is(metaErr, ErrBlobNotFound) {
				return nil, fmt.Errorf("multiManager.ListUserSessions: %w", metaErr)
			}
			// Meta blob missing but session blob exists — drop from index.
			stale = true
			continue
		}
		keepIDs = append(keepIDs, sessionID)
		result = append(result, meta)
	}

	if stale {
		idx.Sessions = keepIDs
		if saveErr := mm.saveUserIndex(ctx, userID, idx); saveErr != nil {
			return nil, saveErr
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

// DestroySession deletes the session blob and removes it from the user's index.
// Returns ErrSessionNotOwned if sessionID is not present in userID's index.
func (mm *multiManager) DestroySession(ctx context.Context, sessionID, userID string) error {
	idx, err := mm.loadUserIndex(ctx, userID)
	if err != nil {
		return fmt.Errorf("multiManager.DestroySession: %w", err)
	}
	if !slices.Contains(idx.Sessions, sessionID) {
		return fmt.Errorf("multiManager.DestroySession: %w", ErrSessionNotOwned)
	}
	if err := mm.blobStore.Delete(ctx, mm.keyPrefix+sessionID); err != nil {
		return fmt.Errorf("multiManager.DestroySession: %w", err)
	}
	return mm.UnbindUser(ctx, sessionID, userID) // clears fence + deletes meta blob
}

// DestroyUserSessions deletes all session and metadata blobs for a user and
// removes the index.
func (mm *multiManager) DestroyUserSessions(ctx context.Context, userID string) error {
	idx, err := mm.loadUserIndex(ctx, userID)
	if err != nil {
		return err
	}
	for _, sessionID := range idx.Sessions {
		_ = mm.blobStore.Delete(ctx, mm.keyPrefix+sessionID)
		_ = mm.blobStore.Delete(ctx, mm.metaPrefix+sessionID)
		mm.touchFence.Delete(sessionID)
	}
	return mm.blobStore.Delete(ctx, mm.userPrefix+userID)
}

// TouchSession updates the LastActiveAt timestamp in the session's metadata blob.
// If the session was touched within the configured TouchWindow, the KV write is
// skipped and nil is returned immediately. If the session metadata is not found,
// it returns nil (noop).
func (mm *multiManager) TouchSession(ctx context.Context, sessionID, userID string) error {
	now := time.Now()
	if last, ok := mm.touchFence.Load(sessionID); ok {
		if now.Sub(last.(time.Time)) < mm.touchWindow {
			return nil
		}
	}
	meta, err := mm.loadSessionMeta(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrBlobNotFound) {
			return nil
		}
		return err
	}
	meta.LastActiveAt = now
	if err := mm.saveSessionMeta(ctx, sessionID, meta); err != nil {
		return err
	}
	mm.touchFence.Store(sessionID, now)
	return nil
}
