package interactive

import (
	"strings"
	"testing"

	testutil "github.com/go-sum/componentry/testutil"
)

func TestThemeScriptRendersExpectedBootstrapCode(t *testing.T) {
	got := testutil.RenderNode(t, ThemeScript())
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
	got := testutil.RenderNode(t, ThemeSelector())
	checks := []string{
		` data-theme-toggle=""`,
		` aria-label="Toggle theme"`,
		`theme-icon-light`,
		`theme-icon-dark`,
		`theme-icon-system`,
		`focus-visible:ring-[3px]`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("ThemeSelector() output missing %q in %s", check, got)
		}
	}
}
