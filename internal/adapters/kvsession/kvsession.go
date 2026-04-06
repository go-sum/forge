package kvsession

import (
	"context"
	"errors"
	"time"

	"github.com/go-sum/kv"
	"github.com/go-sum/session"
)

// New wraps a kv.Store as a session.BlobStore.
// Returns nil when store is nil, allowing cookie-mode sessions to operate
// without a KV store.
func New(store kv.Store) session.BlobStore {
	if store == nil {
		return nil
	}
	return &kvSessionAdapter{store: store}
}

type kvSessionAdapter struct{ store kv.Store }

func (a *kvSessionAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := a.store.Get(ctx, key)
	if errors.Is(err, kv.ErrNotFound) {
		return nil, session.ErrBlobNotFound
	}
	return val, err
}

func (a *kvSessionAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return a.store.Set(ctx, key, value, kv.SetOptions{TTL: ttl})
}

func (a *kvSessionAdapter) Delete(ctx context.Context, key string) error {
	return a.store.Delete(ctx, key)
}
