// Package interactive provides theme script and selector components.
package interactive

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	componenticons "starter/pkg/components/icons"
	iconrender "starter/pkg/components/icons/render"
	core "starter/pkg/components/ui/core"
)

// themeScriptContent is the exact JavaScript emitted by ThemeScript().
// Defined as a constant so its SHA-256 hash is computed once (ScriptCSPHash)
// and the rendered <script> always matches what the CSP authorises.
//
// Sets data-theme-preference on <html> so CSS icon-visibility rules fire on
// first paint (prevents FOUC for the theme toggle icons).
const themeScriptContent = `(function(){` +
	`var p=localStorage.getItem('themePreference')||'system';` +
	`document.documentElement.dataset.themePreference=p;` +
	`var dark=p==='dark'||(p==='system'&&window.matchMedia('(prefers-color-scheme: dark)').matches);` +
	`if(dark)document.documentElement.classList.add('dark');` +
	`})()`

// ScriptCSPHash is the ready-to-embed CSP token for the ThemeScript inline
// script.  Add it to script-src in your Content-Security-Policy header:
//
//	script-src 'self' <ScriptCSPHash>
var ScriptCSPHash string

func init() {
	// Hash the bytes the browser actually receives — the inner text of the
	// rendered <script> element — so any future change to ThemeScript() or its
	// gomponents wrapper automatically keeps the hash in sync.
	var buf strings.Builder
	ThemeScript().Render(&buf)
	rendered := buf.String()
	inner := strings.TrimPrefix(rendered, "<script>")
	inner = strings.TrimSuffix(inner, "</script>")
	ScriptCSPHash = cspHash(inner)
}

// cspHash returns the 'sha256-...' token for an inline script or style value.
func cspHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
}

// ThemeScript returns a synchronous inline <script> that must be placed
// inside <head> before any body content renders.  It reads the stored
// 'themePreference' key from localStorage ('light', 'dark', or 'system')
// and immediately adds the "dark" class to <html> when needed, preventing
// a flash of unstyled light-mode content on dark-preference page loads.
//
// The script is intentionally minified — it must not be deferred, and
// keeping it small reduces the blocking time to near zero.
func ThemeScript() g.Node {
	return h.Script(g.Raw(themeScriptContent))
}

// ThemeSelector renders a ghost icon button that cycles the user's preference
// through light → dark → system on each click.  The active state is persisted
// to localStorage and the .dark class on <html> is updated in place so the
// change is instant without a page reload.
//
// Icon visibility is controlled by CSS rules keyed on data-theme-preference on
// <html>; no inline JS is needed.  The delegated click handler in
// static/js/app.js responds to data-theme-toggle.
func ThemeSelector() g.Node {
	return h.Button(
		g.Attr("data-theme-toggle", ""),
		h.Type("button"),
		g.Attr("aria-label", "Toggle theme"),
		h.Class("inline-flex items-center justify-center size-9 rounded-md text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"),
		// Sun — visible when data-theme-preference="light" on <html>.
		h.Span(h.Class("contents theme-icon-light"), core.Icon(iconrender.PropsFor(componenticons.ThemeLight, core.IconProps{}))),
		// Moon — visible when data-theme-preference="dark" on <html>.
		h.Span(h.Class("contents theme-icon-dark"), core.Icon(iconrender.PropsFor(componenticons.ThemeDark, core.IconProps{}))),
		// Monitor — visible when data-theme-preference="system" (default).
		h.Span(h.Class("contents theme-icon-system"), core.Icon(iconrender.PropsFor(componenticons.ThemeSystem, core.IconProps{}))),
	)
}
