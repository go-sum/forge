package authsession

import (
	"context"
	"errors"
	"time"

	"github.com/go-sum/kv"
	"github.com/go-sum/session"
)

// WrapKV adapts a kv.Store to the session.BlobStore interface for use as
// server-side session storage. Returns nil if store is nil, which allows
// cookie-mode sessions to operate without a KV store.
func WrapKV(store kv.Store) session.BlobStore {
	if store == nil {
		return nil
	}
	return &kvAdapter{store: store}
}

type kvAdapter struct{ store kv.Store }

func (a *kvAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := a.store.Get(ctx, key)
	if errors.Is(err, kv.ErrNotFound) {
		return nil, session.ErrBlobNotFound
	}
	return val, err
}

func (a *kvAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return a.store.Set(ctx, key, value, kv.SetOptions{TTL: ttl})
}

func (a *kvAdapter) Delete(ctx context.Context, key string) error {
	return a.store.Delete(ctx, key)
}
