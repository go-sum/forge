package interactive

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
)

func renderNode(t *testing.T, node g.Node) string {
	t.Helper()

	var buf strings.Builder
	if err := node.Render(&buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.String()
}

func TestThemeScriptRendersExpectedBootstrapCode(t *testing.T) {
	got := renderNode(t, ThemeScript())
	checks := []string{"themePreference", "matchMedia('(prefers-color-scheme: dark)')", "<script>"}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("ThemeScript() output missing %q in %s", check, got)
		}
	}
	if ScriptCSPHash == "" {
		t.Fatal("ScriptCSPHash should be populated")
	}
}

func TestThemeSelectorRendersToggleButton(t *testing.T) {
	got := renderNode(t, ThemeSelector())
	checks := []string{` data-theme-toggle=""`, ` aria-label="Toggle theme"`, `theme-icon-light`, `theme-icon-dark`, `theme-icon-system`}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("ThemeSelector() output missing %q in %s", check, got)
		}
	}
}
