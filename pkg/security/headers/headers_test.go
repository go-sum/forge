package headers

import "testing"

func TestInjectScriptHashes(t *testing.T) {
	got := InjectScriptHashes("default-src 'self'; script-src 'self'; style-src 'self'", []string{"'sha256-a'", "", " 'sha256-b' "})
	want := "default-src 'self'; script-src 'sha256-a' 'sha256-b' 'self'; style-src 'self'"
	if got != want {
		t.Fatalf("InjectScriptHashes() = %q, want %q", got, want)
	}
}

func TestInjectScriptHashesLeavesUntouchedWhenNoHashes(t *testing.T) {
	csp := "default-src 'self'; script-src 'self'"
	if got := InjectScriptHashes(csp, nil); got != csp {
		t.Fatalf("InjectScriptHashes() = %q, want %q", got, csp)
	}
}
