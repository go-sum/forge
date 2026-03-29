package headers_test

import (
	"testing"

	"github.com/go-sum/server/headers"
)

// --- ParseCacheControl ---

func TestParseCacheControlEmpty(t *testing.T) {
	cc := headers.ParseCacheControl("")
	if cc.NoStore() {
		t.Error("NoStore() = true, want false for empty header")
	}
	if cc.NoCache() {
		t.Error("NoCache() = true, want false for empty header")
	}
	if cc.Public() {
		t.Error("Public() = true, want false for empty header")
	}
	if cc.Private() {
		t.Error("Private() = true, want false for empty header")
	}
	if cc.MustRevalidate() {
		t.Error("MustRevalidate() = true, want false for empty header")
	}
	if cc.Immutable() {
		t.Error("Immutable() = true, want false for empty header")
	}
	if _, ok := cc.MaxAge(); ok {
		t.Error("MaxAge() ok = true, want false for empty header")
	}
	if _, ok := cc.SMaxAge(); ok {
		t.Error("SMaxAge() ok = true, want false for empty header")
	}
}

func TestParseCacheControlNoStore(t *testing.T) {
	cc := headers.ParseCacheControl("no-store")
	if !cc.NoStore() {
		t.Error("NoStore() = false, want true")
	}
}

func TestParseCacheControlNoCache(t *testing.T) {
	cc := headers.ParseCacheControl("no-cache")
	if !cc.NoCache() {
		t.Error("NoCache() = false, want true")
	}
}

func TestParseCacheControlPublic(t *testing.T) {
	cc := headers.ParseCacheControl("public")
	if !cc.Public() {
		t.Error("Public() = false, want true")
	}
}

func TestParseCacheControlPrivate(t *testing.T) {
	cc := headers.ParseCacheControl("private")
	if !cc.Private() {
		t.Error("Private() = false, want true")
	}
}

func TestParseCacheControlMustRevalidate(t *testing.T) {
	cc := headers.ParseCacheControl("must-revalidate")
	if !cc.MustRevalidate() {
		t.Error("MustRevalidate() = false, want true")
	}
}

func TestParseCacheControlImmutable(t *testing.T) {
	cc := headers.ParseCacheControl("immutable")
	if !cc.Immutable() {
		t.Error("Immutable() = false, want true")
	}
}

func TestParseCacheControlMaxAge(t *testing.T) {
	cc := headers.ParseCacheControl("max-age=3600")
	secs, ok := cc.MaxAge()
	if !ok {
		t.Fatal("MaxAge() ok = false, want true")
	}
	if secs != 3600 {
		t.Errorf("MaxAge() = %d, want 3600", secs)
	}
}

func TestParseCacheControlSMaxAge(t *testing.T) {
	cc := headers.ParseCacheControl("s-maxage=300")
	secs, ok := cc.SMaxAge()
	if !ok {
		t.Fatal("SMaxAge() ok = false, want true")
	}
	if secs != 300 {
		t.Errorf("SMaxAge() = %d, want 300", secs)
	}
}

func TestParseCacheControlAbsentMaxAge(t *testing.T) {
	cc := headers.ParseCacheControl("no-store")
	secs, ok := cc.MaxAge()
	if ok {
		t.Errorf("MaxAge() ok = true, want false (absent); secs = %d", secs)
	}
}

func TestParseCacheControlUnknownDirective(t *testing.T) {
	cc := headers.ParseCacheControl("stale-while-revalidate=60")
	if !cc.Has("stale-while-revalidate") {
		t.Error("Has(stale-while-revalidate) = false, want true")
	}
}

func TestParseCacheControlHasCaseInsensitive(t *testing.T) {
	cc := headers.ParseCacheControl("No-Store")
	if !cc.NoStore() {
		t.Error("NoStore() = false, want true (case-insensitive)")
	}
}

func TestParseCacheControlFirstOccurrenceWins(t *testing.T) {
	cc := headers.ParseCacheControl("max-age=60, max-age=120")
	secs, ok := cc.MaxAge()
	if !ok {
		t.Fatal("MaxAge() ok = false, want true")
	}
	if secs != 60 {
		t.Errorf("MaxAge() = %d, want 60 (first occurrence wins)", secs)
	}
}

func TestParseCacheControlMultiDirective(t *testing.T) {
	cc := headers.ParseCacheControl("public, max-age=86400, immutable")
	if !cc.Public() {
		t.Error("Public() = false, want true")
	}
	if !cc.Immutable() {
		t.Error("Immutable() = false, want true")
	}
	secs, ok := cc.MaxAge()
	if !ok {
		t.Fatal("MaxAge() ok = false, want true")
	}
	if secs != 86400 {
		t.Errorf("MaxAge() = %d, want 86400", secs)
	}
}

func TestParseCacheControlStringSorted(t *testing.T) {
	// String output is sorted alphabetically.
	cc := headers.ParseCacheControl("public, max-age=3600, immutable")
	got := cc.String()
	want := "immutable, max-age=3600, public"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestParseCacheControlStringEmpty(t *testing.T) {
	cc := headers.ParseCacheControl("")
	if cc.String() != "" {
		t.Errorf("String() = %q, want %q (empty)", cc.String(), "")
	}
}

// --- Builder ---

func TestBuilderSingleDirective(t *testing.T) {
	got := headers.NewCacheControl().NoStore().String()
	want := "no-store"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBuilderMultiDirectiveInsertionOrder(t *testing.T) {
	got := headers.NewCacheControl().Public().MaxAge(86400).Immutable().String()
	want := "public, max-age=86400, immutable"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBuilderDuplicateIsNoOp(t *testing.T) {
	// Calling MaxAge twice; first value (60) must win.
	got := headers.NewCacheControl().MaxAge(60).MaxAge(120).String()
	want := "max-age=60"
	if got != want {
		t.Errorf("String() = %q, want %q (first-call-wins)", got, want)
	}
}

func TestBuilderNoCache(t *testing.T) {
	got := headers.NewCacheControl().NoCache().String()
	want := "no-cache"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBuilderPrivateMustRevalidate(t *testing.T) {
	got := headers.NewCacheControl().Private().MustRevalidate().String()
	want := "private, must-revalidate"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBuilderSMaxAge(t *testing.T) {
	got := headers.NewCacheControl().Public().SMaxAge(300).String()
	want := "public, s-maxage=300"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
