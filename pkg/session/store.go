package session

import (
	"context"
	"errors"
	"time"
)

// Store persists and retrieves serialized session state.
type Store interface {
	Load(id string) (data []byte, exists bool, err error)
	Save(id string, data []byte, maxAge time.Duration) error
	Delete(id string) error
}

// BlobStore is the minimal persistence contract for server-side sessions.
// Satisfied by an adapter over any key-value store (e.g. kv.Store).
type BlobStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// ErrBlobNotFound must be returned by BlobStore.Get when a key does not exist.
var ErrBlobNotFound = errors.New("session: blob not found")

// Returned by DestroySession when the session ID is not present in the given user's index.
var ErrSessionNotOwned = errors.New("session: session does not belong to user")
