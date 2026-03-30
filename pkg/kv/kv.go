// Package kv defines a backend-agnostic key-value store contract.
//
// Implementations live in sub-packages (e.g. redisstore). The interface
// is intentionally minimal: basic CRUD, TTL, existence checks, and
// prefix scanning. Transactions, watches, and pub/sub are out of scope.
package kv

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a key does not exist.
var ErrNotFound = errors.New("kv: key not found")

// ErrClosed is returned when the store has been closed.
var ErrClosed = errors.New("kv: store closed")

// SetOptions configures a Set operation.
type SetOptions struct {
	TTL time.Duration // 0 means no expiry
}

// Store is the minimal key-value store contract.
type Store interface {
	// Ping verifies the store is reachable.
	Ping(ctx context.Context) error

	// Get retrieves the value for key. Returns ErrNotFound if the key does not exist.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores value under key with optional TTL.
	Set(ctx context.Context, key string, value []byte, opts SetOptions) error

	// Delete removes one or more keys. Non-existent keys are ignored.
	Delete(ctx context.Context, keys ...string) error

	// Exists returns the count of keys that exist.
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Close releases resources held by the store.
	Close() error
}

// Scanner extends Store with pattern-based iteration.
type Scanner interface {
	// Scan iterates over keys matching the given pattern, calling fn for each key.
	// Returning a non-nil error from fn stops iteration.
	Scan(ctx context.Context, pattern string, fn func(key string) error) error
}
