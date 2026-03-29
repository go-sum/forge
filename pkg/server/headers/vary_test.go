package headers_test

import (
	"net/http"
	"testing"

	"github.com/go-sum/server/headers"
)

func TestAppendVaryToEmpty(t *testing.T) {
	h := http.Header{}
	headers.AppendVary(h, "Accept-Language")
	got := h.Get("Vary")
	want := "Accept-Language"
	if got != want {
		t.Errorf("Vary = %q, want %q", got, want)
	}
}

func TestAppendVaryTwoValues(t *testing.T) {
	h := http.Header{}
	headers.AppendVary(h, "Accept-Language", "Accept-Encoding")
	got := h.Get("Vary")
	want := "Accept-Language, Accept-Encoding"
	if got != want {
		t.Errorf("Vary = %q, want %q", got, want)
	}
}

func TestAppendVaryNoDuplicateExact(t *testing.T) {
	h := http.Header{}
	h.Set("Vary", "Accept-Language")
	headers.AppendVary(h, "Accept-Language")
	got := h.Get("Vary")
	want := "Accept-Language"
	if got != want {
		t.Errorf("Vary = %q, want %q (no duplicate)", got, want)
	}
}

func TestAppendVaryCaseInsensitiveDedup(t *testing.T) {
	// existing header uses lowercase; new value uses title case.
	h := http.Header{}
	h.Set("Vary", "accept-language")
	headers.AppendVary(h, "Accept-Language")
	got := h.Get("Vary")
	want := "accept-language"
	if got != want {
		t.Errorf("Vary = %q, want %q (case-insensitive dedup preserves original casing)", got, want)
	}
}

func TestAppendVaryNewValueAddedToExisting(t *testing.T) {
	h := http.Header{}
	h.Set("Vary", "Origin")
	headers.AppendVary(h, "Accept-Language")
	got := h.Get("Vary")
	want := "Origin, Accept-Language"
	if got != want {
		t.Errorf("Vary = %q, want %q", got, want)
	}
}

func TestAppendVaryNoOpEmptyVariadic(t *testing.T) {
	h := http.Header{}
	h.Set("Vary", "Origin")
	headers.AppendVary(h) // no values
	got := h.Get("Vary")
	want := "Origin"
	if got != want {
		t.Errorf("Vary = %q, want %q (unchanged on empty variadic)", got, want)
	}
}

func TestAppendVaryMultipleNewValues(t *testing.T) {
	h := http.Header{}
	h.Set("Vary", "Origin")
	headers.AppendVary(h, "Accept-Language", "Accept-Encoding")
	got := h.Get("Vary")
	want := "Origin, Accept-Language, Accept-Encoding"
	if got != want {
		t.Errorf("Vary = %q, want %q", got, want)
	}
}

func TestAppendVaryPartialDedup(t *testing.T) {
	// "Origin" is already present; only "Accept-Language" should be added.
	h := http.Header{}
	h.Set("Vary", "Origin")
	headers.AppendVary(h, "Origin", "Accept-Language")
	got := h.Get("Vary")
	want := "Origin, Accept-Language"
	if got != want {
		t.Errorf("Vary = %q, want %q (partial dedup)", got, want)
	}
}
