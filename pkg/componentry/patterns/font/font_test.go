package font

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/go-sum/componentry/testutil"

	g "maragu.dev/gomponents"
)

func TestGoogleEmitsPreconnectsAndStylesheet(t *testing.T) {
	got := renderNodes(t, Google("Inter:wght@400;700").Nodes())

	want := []string{
		`rel="preconnect" href="https://fonts.googleapis.com"`,
		`rel="preconnect" href="https://fonts.gstatic.com" crossorigin`,
		`rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;700&amp;display=swap"`,
	}
	assertContains(t, got, want)
}

func TestGoogleCombinesMultipleFamiliesIntoSingleRequest(t *testing.T) {
	got := renderNodes(t, Google("Inter:wght@400;700", "Roboto:wght@400").Nodes())

	if strings.Count(got, `rel="preconnect"`) != 2 {
		t.Fatalf("expected exactly 2 preconnect links, got:\n%s", got)
	}
	if !strings.Contains(got, "family=Inter:wght@400;700&amp;family=Roboto:wght@400") {
		t.Fatalf("missing combined family params in:\n%s", got)
	}
}

func TestGoogleEmitsNothingWhenNoFamilies(t *testing.T) {
	nodes := Google().Nodes()
	if len(nodes) != 0 {
		t.Fatalf("Google() with no families: want 0 nodes, got %d", len(nodes))
	}
}

func TestBunnyEmitsPreconnectAndStylesheet(t *testing.T) {
	got := renderNodes(t, Bunny("inter:400,700").Nodes())

	want := []string{
		`rel="preconnect" href="https://fonts.bunny.net"`,
		`rel="stylesheet" href="https://fonts.bunny.net/css?family=inter:400,700&amp;display=swap"`,
	}
	assertContains(t, got, want)
}

func TestBunnyEmitsNothingWhenNoFamilies(t *testing.T) {
	nodes := Bunny().Nodes()
	if len(nodes) != 0 {
		t.Fatalf("Bunny() with no families: want 0 nodes, got %d", len(nodes))
	}
}

func TestAdobeEmitsStylesheetLink(t *testing.T) {
	got := renderNodes(t, Adobe("abc1234").Nodes())

	if !strings.Contains(got, `href="https://use.typekit.net/abc1234.css"`) {
		t.Fatalf("Adobe() missing expected stylesheet link in:\n%s", got)
	}
}

func TestAdobeEmitsNothingWhenNoProjectID(t *testing.T) {
	nodes := Adobe("").Nodes()
	if len(nodes) != 0 {
		t.Fatalf("Adobe() with empty projectID: want 0 nodes, got %d", len(nodes))
	}
}

func TestSelfEmitsPreloadAndFontFaceStyle(t *testing.T) {
	got := renderNodes(t, Self(Face{
		Family: "Inter",
		URL:    "/public/fonts/inter-regular.woff2",
		Format: "woff2",
		Weight: "400",
		Style:  "normal",
	}).Nodes())

	want := []string{
		`rel="preload"`,
		`href="/public/fonts/inter-regular.woff2"`,
		`as="font"`,
		`type="font/woff2"`,
		`crossorigin`,
		`@font-face`,
		`font-family:'Inter'`,
		`src:url('/public/fonts/inter-regular.woff2') format('woff2')`,
		`font-weight:400`,
		`font-style:normal`,
		`font-display:swap`,
	}
	assertContains(t, got, want)
}

func TestSelfAppliesDefaultsForMissingFields(t *testing.T) {
	got := renderNodes(t, Self(Face{
		Family: "MyFont",
		URL:    "/public/fonts/myfont.woff2",
	}).Nodes())

	// Should default Format=woff2, Weight=400, Style=normal.
	want := []string{
		`type="font/woff2"`,
		`font-weight:400`,
		`font-style:normal`,
	}
	assertContains(t, got, want)
}

func TestSelfMultipleFacesEmitMultiplePreloads(t *testing.T) {
	got := renderNodes(t, Self(
		Face{Family: "Inter", URL: "/public/fonts/inter-400.woff2", Weight: "400"},
		Face{Family: "Inter", URL: "/public/fonts/inter-700.woff2", Weight: "700"},
	).Nodes())

	preloadCount := strings.Count(got, `rel="preload"`)
	if preloadCount != 2 {
		t.Fatalf("expected 2 preload links, got %d in:\n%s", preloadCount, got)
	}
	faceCount := strings.Count(got, "@font-face")
	if faceCount != 2 {
		t.Fatalf("expected 2 @font-face rules, got %d in:\n%s", faceCount, got)
	}
}

func TestSelfEmitsNothingWhenNoFaces(t *testing.T) {
	nodes := Self().Nodes()
	if len(nodes) != 0 {
		t.Fatalf("Self() with no faces: want 0 nodes, got %d", len(nodes))
	}
}

func TestSelfSkipsFacesWithEmptyURL(t *testing.T) {
	got := renderNodes(t, Self(
		Face{Family: "ValidFont", URL: "/public/fonts/valid.woff2"},
		Face{Family: "BadFont", URL: ""},
	).Nodes())

	if strings.Contains(got, "BadFont") {
		t.Fatalf("Self() should skip faces with empty URL, but rendered:\n%s", got)
	}
}

func TestNodesCollectsAllProviders(t *testing.T) {
	got := renderNodes(t, Nodes(
		Google("Inter:wght@400"),
		Bunny("inter:400"),
	))

	if !strings.Contains(got, "fonts.googleapis.com") {
		t.Fatalf("Nodes() missing Google preconnect in:\n%s", got)
	}
	if !strings.Contains(got, "fonts.bunny.net") {
		t.Fatalf("Nodes() missing Bunny preconnect in:\n%s", got)
	}
}

func TestNodesReturnsNilWhenNoProviders(t *testing.T) {
	nodes := Nodes()
	if nodes != nil {
		t.Fatalf("Nodes() with no providers: want nil, got %v", nodes)
	}
}

// --- CSPSources tests ---

func TestGoogleCSPSources(t *testing.T) {
	srcs := Google("Inter:wght@400").CSPSources()
	assertStringSliceContains(t, srcs.StyleSources, "https://fonts.googleapis.com")
	assertStringSliceContains(t, srcs.FontSources, "https://fonts.gstatic.com")
	if len(srcs.StyleInlineHashes) != 0 {
		t.Fatalf("Google CSPSources: want no StyleInlineHashes, got %v", srcs.StyleInlineHashes)
	}
}

func TestGoogleCSPSourcesEmptyWhenNoFamilies(t *testing.T) {
	srcs := Google().CSPSources()
	if len(srcs.StyleSources) != 0 || len(srcs.FontSources) != 0 {
		t.Fatalf("Google().CSPSources(): want zero-value, got %+v", srcs)
	}
}

func TestBunnyCSPSources(t *testing.T) {
	srcs := Bunny("inter:400").CSPSources()
	assertStringSliceContains(t, srcs.StyleSources, "https://fonts.bunny.net")
	assertStringSliceContains(t, srcs.FontSources, "https://fonts.bunny.net")
}

func TestBunnyCSPSourcesEmptyWhenNoFamilies(t *testing.T) {
	srcs := Bunny().CSPSources()
	if len(srcs.StyleSources) != 0 || len(srcs.FontSources) != 0 {
		t.Fatalf("Bunny().CSPSources(): want zero-value, got %+v", srcs)
	}
}

func TestAdobeCSPSources(t *testing.T) {
	srcs := Adobe("abc1234").CSPSources()
	assertStringSliceContains(t, srcs.StyleSources, "https://use.typekit.net")
	assertStringSliceContains(t, srcs.FontSources, "https://use.typekit.net")
}

func TestAdobeCSPSourcesEmptyWhenNoProjectID(t *testing.T) {
	srcs := Adobe("").CSPSources()
	if len(srcs.StyleSources) != 0 || len(srcs.FontSources) != 0 {
		t.Fatalf("Adobe(\"\").CSPSources(): want zero-value, got %+v", srcs)
	}
}

func TestSelfCSPSourcesInlineHash(t *testing.T) {
	face := Face{Family: "Inter", URL: "/public/fonts/inter-regular.woff2", Format: "woff2", Weight: "400", Style: "normal"}
	srcs := Self(face).CSPSources()

	if len(srcs.StyleInlineHashes) != 1 {
		t.Fatalf("Self CSPSources: want 1 hash, got %d: %v", len(srcs.StyleInlineHashes), srcs.StyleInlineHashes)
	}

	// The hash must match sha256 of the generated CSS.
	css := buildFontFaceCSS([]Face{face})
	sum := sha256.Sum256([]byte(css))
	wantHash := fmt.Sprintf("'sha256-%s'", base64.StdEncoding.EncodeToString(sum[:]))
	if srcs.StyleInlineHashes[0] != wantHash {
		t.Fatalf("Self CSPSources hash = %q, want %q", srcs.StyleInlineHashes[0], wantHash)
	}
}

func TestSelfCSPSourcesEmptyWhenNoFaces(t *testing.T) {
	srcs := Self().CSPSources()
	if len(srcs.StyleInlineHashes) != 0 {
		t.Fatalf("Self().CSPSources(): want zero-value, got %+v", srcs)
	}
}

func TestCollectCSPSourcesMergesProviders(t *testing.T) {
	providers := []Provider{
		Google("Inter:wght@400"),
		Bunny("inter:400"),
		Self(Face{Family: "MyFont", URL: "/public/fonts/myfont.woff2"}),
	}
	srcs := CollectCSPSources(providers)

	assertStringSliceContains(t, srcs.StyleSources, "https://fonts.googleapis.com")
	assertStringSliceContains(t, srcs.StyleSources, "https://fonts.bunny.net")
	assertStringSliceContains(t, srcs.FontSources, "https://fonts.gstatic.com")
	assertStringSliceContains(t, srcs.FontSources, "https://fonts.bunny.net")
	if len(srcs.StyleInlineHashes) != 1 {
		t.Fatalf("CollectCSPSources: want 1 inline hash, got %d", len(srcs.StyleInlineHashes))
	}
}

func TestCollectCSPSourcesEmptyForNoProviders(t *testing.T) {
	srcs := CollectCSPSources(nil)
	if len(srcs.StyleSources) != 0 || len(srcs.FontSources) != 0 || len(srcs.StyleInlineHashes) != 0 {
		t.Fatalf("CollectCSPSources(nil): want zero-value, got %+v", srcs)
	}
}

// assertStringSliceContains fails if want is not found in slice.
func assertStringSliceContains(t *testing.T, slice []string, want string) {
	t.Helper()
	for _, s := range slice {
		if s == want {
			return
		}
	}
	t.Fatalf("expected %q in %v", want, slice)
}

// renderNodes renders a []g.Node slice into a single HTML string.
func renderNodes(t *testing.T, nodes []g.Node) string {
	t.Helper()
	if len(nodes) == 0 {
		return ""
	}
	return testutil.RenderNode(t, g.Group(nodes))
}

// assertContains fails the test if any want string is absent from got.
func assertContains(t *testing.T, got string, want []string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Fatalf("missing %q in rendered output:\n%s", w, got)
		}
	}
}
