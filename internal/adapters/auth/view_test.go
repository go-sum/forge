package authadapter

import (
	"net/http"
	"strings"
	"testing"

	auth "github.com/go-sum/auth"
	"github.com/go-sum/auth/model"
	"github.com/go-sum/componentry/testutil"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/server/route"
	g "maragu.dev/gomponents"
	"github.com/labstack/echo/v5"
)

// minimalRequest builds an auth.Request suitable for renderer tests.
// PageFn is a no-op wrapper so we get the card content without a full layout.
// When passkeyEnabled is true, routes are seeded so URL resolution works.
func minimalRequest(preferred auth.MethodName, passkeyEnabled bool) auth.Request {
	req := auth.Request{
		CSRFToken:      "test-csrf",
		CSRFFieldName:  "csrf",
		PasskeyEnabled: passkeyEnabled,
		Preferred:      preferred,
		PageFn: func(title string, children ...g.Node) g.Node {
			return g.Group(children)
		},
	}
	if passkeyEnabled {
		e := echo.New()
		noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }
		authGrp := e.Group("/auth/passkeys")
		route.Add(authGrp, echo.Route{Method: http.MethodPost, Path: "/authenticate/begin", Name: "passkey.authenticate.begin", Handler: noOp})
		route.Add(authGrp, echo.Route{Method: http.MethodPost, Path: "/authenticate/finish", Name: "passkey.authenticate.finish", Handler: noOp})
		route.Add(authGrp, echo.Route{Method: http.MethodPost, Path: "/register/begin", Name: "passkey.register.begin", Handler: noOp})
		route.Add(authGrp, echo.Route{Method: http.MethodPost, Path: "/register/finish", Name: "passkey.register.finish", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/account/passkeys", Name: "passkey.list", Handler: noOp})
		req.State = view.Request{Routes: e.Router().Routes()}
	}
	return req
}

// TestSigninPageEmailFirstNoOrderClasses checks that when TOTP is preferred the
// rendered HTML carries no flex order classes — DOM order alone determines layout.
func TestSigninPageEmailFirstNoOrderClasses(t *testing.T) {
	r := NewRenderer()
	req := minimalRequest(auth.MethodEmailTOTP, true)
	node := r.SigninPage(req, nil, model.BeginSigninInput{}, "/signin", "/signup", "csrf")
	html := testutil.RenderNode(t, node)

	if !strings.Contains(html, `id="email"`) {
		t.Fatal("expected id=email in rendered HTML")
	}
	if !strings.Contains(html, "data-passkey-visible") {
		t.Fatal("expected data-passkey-visible in rendered HTML")
	}
	// Email block should be order-1 (default); no order-2 on it.
	// Passkey block should be order-2 (secondary). No order-1 on any element.
	if strings.Contains(html, "order-1") {
		t.Error("email-first layout: expected no order-1 class (passkey block should be order-2)")
	}
}

// TestSigninPagePasskeyFirstOrderClasses checks that when passkey is preferred
// the rendered HTML carries order-1 on the passkey block and order-2 on the
// email block so CSS flex renders passkey on top.
func TestSigninPagePasskeyFirstOrderClasses(t *testing.T) {
	r := NewRenderer()
	req := minimalRequest(auth.MethodPasskey, true)
	node := r.SigninPage(req, nil, model.BeginSigninInput{}, "/signin", "/signup", "csrf")
	html := testutil.RenderNode(t, node)

	if !strings.Contains(html, "order-1") {
		t.Error("passkey-first layout: expected order-1 class on passkey block")
	}
	if !strings.Contains(html, "order-2") {
		t.Error("passkey-first layout: expected order-2 class on email block")
	}
}

// TestSigninPagePasskeyFirstDividerText checks that the "or" divider reads
// "or continue with email" when the passkey block is rendered first.
func TestSigninPagePasskeyFirstDividerText(t *testing.T) {
	r := NewRenderer()
	req := minimalRequest(auth.MethodPasskey, true)
	node := r.SigninPage(req, nil, model.BeginSigninInput{}, "/signin", "/signup", "csrf")
	html := testutil.RenderNode(t, node)

	if !strings.Contains(html, "or continue with email") {
		t.Errorf("passkey-first layout: expected divider text %q", "or continue with email")
	}
}

// TestSigninPageEmailFirstDividerText checks the divider reads plain "or"
// when TOTP is the preferred method.
func TestSigninPageEmailFirstDividerText(t *testing.T) {
	r := NewRenderer()
	req := minimalRequest(auth.MethodEmailTOTP, true)
	node := r.SigninPage(req, nil, model.BeginSigninInput{}, "/signin", "/signup", "csrf")
	html := testutil.RenderNode(t, node)

	if strings.Contains(html, "or continue with email") {
		t.Errorf("email-first layout: divider should be plain %q, not %q", "or", "or continue with email")
	}
}

// TestSigninPagePasskeyPreferredButDisabled checks that when the passkey
// backend is unavailable, the passkey block is absent entirely regardless of
// the Preferred value.
func TestSigninPagePasskeyPreferredButDisabled(t *testing.T) {
	r := NewRenderer()
	req := minimalRequest(auth.MethodPasskey, false) // PasskeyEnabled=false
	node := r.SigninPage(req, nil, model.BeginSigninInput{}, "/signin", "/signup", "csrf")
	html := testutil.RenderNode(t, node)

	if strings.Contains(html, "data-passkey-visible") {
		t.Error("passkey disabled: expected no data-passkey-visible in HTML")
	}
	if strings.Contains(html, "data-passkey-enabled") {
		t.Error("passkey disabled: expected no data-passkey-enabled attribute in HTML")
	}
	if !strings.Contains(html, `id="email"`) {
		t.Error("passkey disabled: expected email field to be present")
	}
}
