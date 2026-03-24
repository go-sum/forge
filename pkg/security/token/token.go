// Package token issues and verifies HMAC-SHA256 signed, time-limited tokens.
// Tokens are self-contained (no server-side storage) and safe to embed in HTML
// forms, cookies, or URLs.
//
// Wire format (64 bytes, base64url-encoded without padding):
//
//	bytes  0–15  random nonce          (prevents precomputation)
//	bytes 16–23  iat as big-endian int64 unix seconds
//	bytes 24–31  exp as big-endian int64 unix seconds
//	bytes 32–63  HMAC-SHA256(key, scope \x00 nonce iat exp)
//
// The scope string is mixed into the MAC but not stored in the token wire
// format. Both Issue and Verify must be called with the same scope, so a token
// issued for one purpose (e.g. "csrf") cannot be replayed for another (e.g.
// "password-reset:uuid").
package token

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"time"
)

// ErrInvalid is returned when a token is malformed or its HMAC does not match.
var ErrInvalid = errors.New("token: invalid")

// ErrExpired is returned when a well-formed, authentic token has passed its
// expiry time. Callers may surface a "please refresh and try again" message.
var ErrExpired = errors.New("token: expired")

const (
	nonceLen = 16
	timeLen  = 8 // int64 unix seconds, big-endian
	macLen   = 32 // HMAC-SHA256
	tokenLen = nonceLen + timeLen + timeLen + macLen // 64 bytes → 86 base64url chars
)

// Issue generates a cryptographically random token scoped to scope that expires
// after ttl. It returns the base64url-encoded token string or an error if the
// system random source fails.
func Issue(key []byte, scope string, ttl time.Duration) (string, error) {
	var raw [tokenLen]byte

	if _, err := rand.Read(raw[:nonceLen]); err != nil {
		return "", err
	}

	now := time.Now().Unix()
	binary.BigEndian.PutUint64(raw[nonceLen:], uint64(now))
	binary.BigEndian.PutUint64(raw[nonceLen+timeLen:], uint64(now+int64(ttl.Seconds())))

	mac := computeMAC(key, scope, raw[:nonceLen+timeLen+timeLen])
	copy(raw[nonceLen+timeLen+timeLen:], mac)

	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

// Verify checks that raw is a valid, unexpired token for the given scope and
// signing key. It returns ErrInvalid for malformed or tampered tokens and
// ErrExpired for tokens that have passed their expiry time.
//
// MAC verification always runs before expiry is checked to avoid leaking
// timing information about partially valid tokens.
func Verify(key []byte, scope string, raw string) error {
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(b) != tokenLen {
		return ErrInvalid
	}

	expected := computeMAC(key, scope, b[:nonceLen+timeLen+timeLen])
	if !hmac.Equal(expected, b[nonceLen+timeLen+timeLen:]) {
		return ErrInvalid
	}

	exp := int64(binary.BigEndian.Uint64(b[nonceLen+timeLen:]))
	if time.Now().Unix() > exp {
		return ErrExpired
	}

	return nil
}

func computeMAC(key []byte, scope string, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(scope))
	h.Write([]byte{0}) // null separator prevents scope from bleeding into data
	h.Write(data)
	return h.Sum(nil)
}
