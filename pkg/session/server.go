package session

import (
	"cmp"
	"context"
	"errors"
	"time"
)

// serverStore stores session data in a BlobStore, with only the session ID in the cookie.
type serverStore struct {
	store     BlobStore
	keyPrefix string
}

func newServerStore(store BlobStore, keyPrefix string) *serverStore {
	return &serverStore{store: store, keyPrefix: cmp.Or(keyPrefix, "session:")}
}

func (b *serverStore) Load(id string) ([]byte, bool, error) {
	data, err := b.store.Get(context.Background(), b.keyPrefix+id)
	if err != nil {
		if errors.Is(err, ErrBlobNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return data, true, nil
}

func (b *serverStore) Save(id string, data []byte, maxAge time.Duration) error {
	return b.store.Set(context.Background(), b.keyPrefix+id, data, maxAge)
}

func (b *serverStore) Delete(id string) error {
	return b.store.Delete(context.Background(), b.keyPrefix+id)
}
