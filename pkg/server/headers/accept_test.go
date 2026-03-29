package headers_test

import (
	"testing"

	"github.com/go-sum/server/headers"
)

// --- ParseAcceptLanguage ---

func TestParseAcceptLanguageEmpty(t *testing.T) {
	al := headers.ParseAcceptLanguage("")
	if len(al) != 0 {
		t.Errorf("len = %d, want 0", len(al))
	}
}

func TestParseAcceptLanguageBlank(t *testing.T) {
	al := headers.ParseAcceptLanguage("   ")
	if len(al) != 0 {
		t.Errorf("len = %d, want 0", len(al))
	}
}

func TestParseAcceptLanguageSingleNoQ(t *testing.T) {
	al := headers.ParseAcceptLanguage("en")
	if len(al) != 1 {
		t.Fatalf("len = %d, want 1", len(al))
	}
	if al[0].Tag != "en" {
		t.Errorf("Tag = %q, want %q", al[0].Tag, "en")
	}
	if al[0].Quality != 1.0 {
		t.Errorf("Quality = %v, want 1.0", al[0].Quality)
	}
}

func TestParseAcceptLanguageSingleWithQ(t *testing.T) {
	al := headers.ParseAcceptLanguage("fr;q=0.8")
	if len(al) != 1 {
		t.Fatalf("len = %d, want 1", len(al))
	}
	if al[0].Quality != 0.8 {
		t.Errorf("Quality = %v, want 0.8", al[0].Quality)
	}
}

func TestParseAcceptLanguageMultiSortedByQ(t *testing.T) {
	al := headers.ParseAcceptLanguage("fr;q=0.5, en-US, de;q=0.9")
	if len(al) != 3 {
		t.Fatalf("len = %d, want 3", len(al))
	}
	if al[0].Tag != "en-US" {
		t.Errorf("first Tag = %q, want %q", al[0].Tag, "en-US")
	}
	if al[1].Tag != "de" {
		t.Errorf("second Tag = %q, want %q", al[1].Tag, "de")
	}
	if al[2].Tag != "fr" {
		t.Errorf("third Tag = %q, want %q", al[2].Tag, "fr")
	}
}

func TestParseAcceptLanguageQZeroExcluded(t *testing.T) {
	al := headers.ParseAcceptLanguage("en, fr;q=0")
	if len(al) != 1 {
		t.Fatalf("len = %d, want 1 (q=0 excluded)", len(al))
	}
	if al[0].Tag != "en" {
		t.Errorf("Tag = %q, want %q", al[0].Tag, "en")
	}
}

func TestParseAcceptLanguageMalformedQDropped(t *testing.T) {
	al := headers.ParseAcceptLanguage("en, fr;q=bad, de")
	if len(al) != 2 {
		t.Fatalf("len = %d, want 2 (malformed q item dropped)", len(al))
	}
	tags := []string{al[0].Tag, al[1].Tag}
	for _, want := range []string{"en", "de"} {
		found := false
		for _, got := range tags {
			if got == want {
				found = true
			}
		}
		if !found {
			t.Errorf("tag %q not found in result %v", want, tags)
		}
	}
}

// --- AcceptLanguage.Preferred ---

func TestAcceptLanguagePreferredExactMatch(t *testing.T) {
	al := headers.ParseAcceptLanguage("fr")
	got := al.Preferred([]string{"en", "fr", "de"})
	if got != "fr" {
		t.Errorf("Preferred = %q, want %q", got, "fr")
	}
}

func TestAcceptLanguagePreferredCaseInsensitive(t *testing.T) {
	al := headers.ParseAcceptLanguage("FR")
	got := al.Preferred([]string{"fr"})
	if got != "fr" {
		t.Errorf("Preferred = %q, want %q", got, "fr")
	}
}

func TestAcceptLanguagePreferredSubtagPrefix(t *testing.T) {
	// "en-US" in header should match candidate "en"
	al := headers.ParseAcceptLanguage("en-US")
	got := al.Preferred([]string{"de", "en", "fr"})
	if got != "en" {
		t.Errorf("Preferred = %q, want %q (subtag prefix match)", got, "en")
	}
}

func TestAcceptLanguagePreferredNoReverseSubtag(t *testing.T) {
	// "en" in header must NOT match candidate "en-US"
	al := headers.ParseAcceptLanguage("en")
	got := al.Preferred([]string{"en-US"})
	if got != "" {
		t.Errorf("Preferred = %q, want %q (no reverse subtag match)", got, "")
	}
}

func TestAcceptLanguagePreferredWildcard(t *testing.T) {
	al := headers.ParseAcceptLanguage("*")
	got := al.Preferred([]string{"de", "fr"})
	if got != "de" {
		t.Errorf("Preferred = %q, want %q (wildcard returns first candidate)", got, "de")
	}
}

func TestAcceptLanguagePreferredNoMatch(t *testing.T) {
	al := headers.ParseAcceptLanguage("ja")
	got := al.Preferred([]string{"en", "fr"})
	if got != "" {
		t.Errorf("Preferred = %q, want %q (no match)", got, "")
	}
}

func TestAcceptLanguagePreferredEmptySlice(t *testing.T) {
	al := headers.ParseAcceptLanguage("")
	got := al.Preferred([]string{"en"})
	if got != "" {
		t.Errorf("Preferred = %q, want %q (empty AcceptLanguage)", got, "")
	}
}

// --- AcceptLanguage.String ---

func TestAcceptLanguageStringEmpty(t *testing.T) {
	al := headers.ParseAcceptLanguage("")
	if al.String() != "" {
		t.Errorf("String() = %q, want %q", al.String(), "")
	}
}

func TestAcceptLanguageStringRoundTrip(t *testing.T) {
	// q=1.0 items omit q; q<1.0 items include it.
	al := headers.ParseAcceptLanguage("en-US, fr;q=0.8, de;q=0.5")
	got := al.String()
	want := "en-US, fr;q=0.8, de;q=0.5"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// --- ParseAccept ---

func TestParseAcceptEmpty(t *testing.T) {
	a := headers.ParseAccept("")
	if len(a) != 0 {
		t.Errorf("len = %d, want 0", len(a))
	}
}

func TestParseAcceptSingleNoQ(t *testing.T) {
	a := headers.ParseAccept("text/html")
	if len(a) != 1 {
		t.Fatalf("len = %d, want 1", len(a))
	}
	if a[0].Type != "text/html" {
		t.Errorf("Type = %q, want %q", a[0].Type, "text/html")
	}
	if a[0].Quality != 1.0 {
		t.Errorf("Quality = %v, want 1.0", a[0].Quality)
	}
}

func TestParseAcceptQZeroExcluded(t *testing.T) {
	a := headers.ParseAccept("text/html, application/json;q=0")
	if len(a) != 1 {
		t.Fatalf("len = %d, want 1 (q=0 excluded)", len(a))
	}
	if a[0].Type != "text/html" {
		t.Errorf("Type = %q, want %q", a[0].Type, "text/html")
	}
}

func TestParseAcceptSortedByQThenSpecificity(t *testing.T) {
	// At equal q, exact type > type/* > */*
	a := headers.ParseAccept("*/*;q=0.8, text/*;q=0.8, text/html;q=0.8")
	if len(a) != 3 {
		t.Fatalf("len = %d, want 3", len(a))
	}
	if a[0].Type != "text/html" {
		t.Errorf("first = %q, want %q (most specific)", a[0].Type, "text/html")
	}
	if a[1].Type != "text/*" {
		t.Errorf("second = %q, want %q", a[1].Type, "text/*")
	}
	if a[2].Type != "*/*" {
		t.Errorf("third = %q, want %q", a[2].Type, "*/*")
	}
}

func TestParseAcceptMalformedQDropped(t *testing.T) {
	a := headers.ParseAccept("text/html, application/json;q=xyz")
	if len(a) != 1 {
		t.Fatalf("len = %d, want 1 (malformed q dropped)", len(a))
	}
	if a[0].Type != "text/html" {
		t.Errorf("Type = %q, want %q", a[0].Type, "text/html")
	}
}

// --- Accept.Preferred ---

func TestAcceptPreferredExact(t *testing.T) {
	a := headers.ParseAccept("application/json, text/html")
	got := a.Preferred([]string{"text/html", "application/json"})
	if got != "application/json" {
		t.Errorf("Preferred = %q, want %q", got, "application/json")
	}
}

func TestAcceptPreferredWildcardSubtype(t *testing.T) {
	a := headers.ParseAccept("text/*")
	got := a.Preferred([]string{"application/json", "text/html"})
	if got != "text/html" {
		t.Errorf("Preferred = %q, want %q (text/* matches text/html)", got, "text/html")
	}
}

func TestAcceptPreferredGlobalWildcard(t *testing.T) {
	a := headers.ParseAccept("*/*")
	got := a.Preferred([]string{"application/json"})
	if got != "application/json" {
		t.Errorf("Preferred = %q, want %q (*/* returns first candidate)", got, "application/json")
	}
}

func TestAcceptPreferredNoMatch(t *testing.T) {
	a := headers.ParseAccept("application/json")
	got := a.Preferred([]string{"text/html"})
	if got != "" {
		t.Errorf("Preferred = %q, want %q (no match)", got, "")
	}
}

func TestAcceptPreferredEmptySlice(t *testing.T) {
	a := headers.ParseAccept("")
	got := a.Preferred([]string{"text/html"})
	if got != "" {
		t.Errorf("Preferred = %q, want %q (empty Accept)", got, "")
	}
}

// --- Accept.String ---

func TestAcceptStringEmpty(t *testing.T) {
	a := headers.ParseAccept("")
	if a.String() != "" {
		t.Errorf("String() = %q, want %q", a.String(), "")
	}
}

func TestAcceptStringRoundTrip(t *testing.T) {
	a := headers.ParseAccept("text/html, application/json;q=0.9")
	got := a.String()
	want := "text/html, application/json;q=0.9"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
