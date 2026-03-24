package token

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

var testKey = []byte("test-signing-key-32-bytes-padded!")

func TestIssueAndVerify(t *testing.T) {
	tok, err := Issue(testKey, "csrf", time.Hour)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if tok == "" {
		t.Fatal("Issue() returned empty token")
	}
	if err := Verify(testKey, "csrf", tok); err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
}

func TestIssueProducesUniqueTokens(t *testing.T) {
	a, _ := Issue(testKey, "csrf", time.Hour)
	b, _ := Issue(testKey, "csrf", time.Hour)
	if a == b {
		t.Fatal("Issue() returned identical tokens on successive calls")
	}
}

func TestVerifyWrongScope(t *testing.T) {
	tok, _ := Issue(testKey, "scope-a", time.Hour)
	err := Verify(testKey, "scope-b", tok)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Verify() wrong scope = %v, want ErrInvalid", err)
	}
}

func TestVerifyWrongKey(t *testing.T) {
	tok, _ := Issue([]byte("key-a"), "csrf", time.Hour)
	err := Verify([]byte("key-b"), "csrf", tok)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Verify() wrong key = %v, want ErrInvalid", err)
	}
}

func TestVerifyExpired(t *testing.T) {
	tok, _ := Issue(testKey, "csrf", -time.Second)
	err := Verify(testKey, "csrf", tok)
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("Verify() expired token = %v, want ErrExpired", err)
	}
}

func TestVerifyTamperedNonce(t *testing.T) {
	tok, _ := Issue(testKey, "csrf", time.Hour)
	b, _ := base64.RawURLEncoding.DecodeString(tok)
	b[0] ^= 0xFF // flip first nonce byte
	tampered := base64.RawURLEncoding.EncodeToString(b)
	if err := Verify(testKey, "csrf", tampered); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Verify() tampered nonce = %v, want ErrInvalid", err)
	}
}

func TestVerifyTamperedExpiry(t *testing.T) {
	tok, _ := Issue(testKey, "csrf", time.Hour)
	b, _ := base64.RawURLEncoding.DecodeString(tok)
	// Flip a bit in the exp field (bytes 24–31) to extend expiry.
	b[24] ^= 0x01
	tampered := base64.RawURLEncoding.EncodeToString(b)
	if err := Verify(testKey, "csrf", tampered); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Verify() tampered expiry = %v, want ErrInvalid", err)
	}
}

func TestVerifyTamperedMAC(t *testing.T) {
	tok, _ := Issue(testKey, "csrf", time.Hour)
	b, _ := base64.RawURLEncoding.DecodeString(tok)
	b[32] ^= 0xFF // flip first MAC byte
	tampered := base64.RawURLEncoding.EncodeToString(b)
	if err := Verify(testKey, "csrf", tampered); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Verify() tampered MAC = %v, want ErrInvalid", err)
	}
}

func TestVerifyShortToken(t *testing.T) {
	if err := Verify(testKey, "csrf", "abc"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Verify() short token = %v, want ErrInvalid", err)
	}
}

func TestVerifyEmpty(t *testing.T) {
	if err := Verify(testKey, "csrf", ""); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Verify() empty = %v, want ErrInvalid", err)
	}
}

func TestVerifyInvalidBase64(t *testing.T) {
	if err := Verify(testKey, "csrf", "not-valid-base64!!!"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Verify() bad base64 = %v, want ErrInvalid", err)
	}
}

// TestMACCheckedBeforeExpiry ensures that a tampered-but-expired token still
// returns ErrInvalid (not ErrExpired) — MAC is always verified first.
func TestMACCheckedBeforeExpiry(t *testing.T) {
	tok, _ := Issue(testKey, "csrf", -time.Second) // expired
	b, _ := base64.RawURLEncoding.DecodeString(tok)
	b[0] ^= 0xFF // also tampered
	tampered := base64.RawURLEncoding.EncodeToString(b)
	if err := Verify(testKey, "csrf", tampered); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Verify() tampered+expired = %v, want ErrInvalid (MAC checked first)", err)
	}
}
