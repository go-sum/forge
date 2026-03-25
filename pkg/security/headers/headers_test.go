package headers

import "testing"

func TestInjectDirectiveSourcesStyleSrc(t *testing.T) {
	csp := "default-src 'self'; script-src 'self'; style-src 'self'; font-src 'self'"
	got := InjectDirectiveSources(csp, "style-src", []string{"https://fonts.googleapis.com", "'sha256-abc'"})
	want := "default-src 'self'; script-src 'self'; style-src https://fonts.googleapis.com 'sha256-abc' 'self'; font-src 'self'"
	if got != want {
		t.Fatalf("InjectDirectiveSources(style-src) = %q, want %q", got, want)
	}
}

func TestInjectDirectiveSourcesFontSrc(t *testing.T) {
	csp := "default-src 'self'; style-src 'self'; font-src 'self'"
	got := InjectDirectiveSources(csp, "font-src", []string{"https://fonts.gstatic.com"})
	want := "default-src 'self'; style-src 'self'; font-src https://fonts.gstatic.com 'self'"
	if got != want {
		t.Fatalf("InjectDirectiveSources(font-src) = %q, want %q", got, want)
	}
}

func TestInjectDirectiveSourcesStripsBlankSources(t *testing.T) {
	csp := "style-src 'self'"
	got := InjectDirectiveSources(csp, "style-src", []string{"", "  ", "https://example.com"})
	want := "style-src https://example.com 'self'"
	if got != want {
		t.Fatalf("InjectDirectiveSources() = %q, want %q", got, want)
	}
}

func TestInjectDirectiveSourcesLeavesUntouchedWhenNoSources(t *testing.T) {
	csp := "default-src 'self'; style-src 'self'"
	if got := InjectDirectiveSources(csp, "style-src", nil); got != csp {
		t.Fatalf("InjectDirectiveSources() = %q, want %q", got, csp)
	}
}

func TestInjectDirectiveSourcesLeavesUntouchedWhenDirectiveMissing(t *testing.T) {
	csp := "default-src 'self'; script-src 'self'"
	got := InjectDirectiveSources(csp, "style-src", []string{"https://example.com"})
	if got != csp {
		t.Fatalf("InjectDirectiveSources() modified csp when directive absent: %q", got)
	}
}
