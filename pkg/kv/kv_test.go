package kv_test

import (
	"errors"
	"testing"

	"github.com/go-sum/kv"
)

func TestErrNotFoundIdentity(t *testing.T) {
	if !errors.Is(kv.ErrNotFound, kv.ErrNotFound) {
		t.Fatal("ErrNotFound should be identifiable with errors.Is")
	}
}

func TestErrClosedIdentity(t *testing.T) {
	if !errors.Is(kv.ErrClosed, kv.ErrClosed) {
		t.Fatal("ErrClosed should be identifiable with errors.Is")
	}
}

func TestErrNotFoundAndErrClosedAreDistinct(t *testing.T) {
	if errors.Is(kv.ErrNotFound, kv.ErrClosed) {
		t.Fatal("ErrNotFound and ErrClosed must be distinct errors")
	}
	if errors.Is(kv.ErrClosed, kv.ErrNotFound) {
		t.Fatal("ErrClosed and ErrNotFound must be distinct errors")
	}
}
