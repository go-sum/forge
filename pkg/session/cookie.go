package session

import (
	"fmt"
	"time"

	"github.com/gorilla/securecookie"
)

// cookieStore stores the entire session envelope in a signed+encrypted cookie.
type cookieStore struct {
	codec securecookie.Codec
}

func newCookieStore(authKey, encryptKey []byte) (*cookieStore, error) {
	sc := securecookie.New(authKey, encryptKey)
	sc.MaxAge(0) // no codec-level expiry; handled by cookie MaxAge
	return &cookieStore{codec: sc}, nil
}

// Encode serializes and signs+encrypts session data for cookie storage.
func (b *cookieStore) Encode(name string, data []byte) (string, error) {
	return securecookie.EncodeMulti(name, data, b.codec)
}

// Decode verifies and decrypts a cookie value back to session data.
func (b *cookieStore) Decode(name, value string) ([]byte, error) {
	var data []byte
	if err := securecookie.DecodeMulti(name, value, &data, b.codec); err != nil {
		return nil, fmt.Errorf("session: decode cookie: %w", err)
	}
	return data, nil
}

// Load is not used for cookie store — data comes from the cookie value directly.
func (b *cookieStore) Load(id string) ([]byte, bool, error) {
	return nil, false, nil
}

// Save is not used for cookie store — data goes into the cookie value directly.
func (b *cookieStore) Save(id string, data []byte, maxAge time.Duration) error {
	return nil
}

// Delete is not used for cookie store — handled by expiring the cookie.
func (b *cookieStore) Delete(id string) error {
	return nil
}
