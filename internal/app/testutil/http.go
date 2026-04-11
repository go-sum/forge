package testutil

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	auth "github.com/go-sum/auth"
	"github.com/go-sum/server/apperr"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const CSRFToken = "csrf-token"

func NewRequestContext(method, target string, body io.Reader) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	RegisterTestRoutes(e)
	req := httptest.NewRequest(method, target, body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func NewFormContext(method, target string, values url.Values) (*echo.Context, *httptest.ResponseRecorder) {
	c, rec := NewRequestContext(method, target, strings.NewReader(values.Encode()))
	c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	return c, rec
}

func SetCSRFToken(c *echo.Context) {
	c.Set(middleware.DefaultCSRFConfig.ContextKey, CSRFToken)
}

func SetUserID(c *echo.Context, userID string) {
	c.Set(auth.ContextKeyUserID, userID)
}

func SetPathParam(c *echo.Context, path, name, value string) {
	c.SetPath(path)
	c.SetPathValues(echo.PathValues{{
		Name:  name,
		Value: value,
	}})
}

func AssertAppErrorStatus(t *testing.T, err error, status int) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %T, want *apperr.Error", err)
	}
	if appErr.Status != status {
		t.Fatalf("status = %d, want %d", appErr.Status, status)
	}
}

// RegisterTestRoutes mirrors the app route names used by feature handler tests.
func RegisterTestRoutes(e *echo.Echo) {
	noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signup", Name: "signup.get", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/robots.txt", Name: "robots.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/sitemap.xml", Name: "sitemap.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/contact", Name: "contact.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodPost, Path: "/contact", Name: "contact.submit", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/admin/elevate", Name: "admin.elevate", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodPost, Path: "/admin/elevate", Name: "admin.elevate.post", Handler: noOp})

	adminUsers := e.Group("/admin/users")
	route.Add(adminUsers, echo.Route{Method: http.MethodGet, Path: "", Name: "admin.user.list", Handler: noOp})
	route.Add(adminUsers, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "admin.user.edit", Handler: noOp})
	route.Add(adminUsers, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "admin.user.row", Handler: noOp})
	route.Add(adminUsers, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "admin.user.update", Handler: noOp})
	route.Add(adminUsers, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "admin.user.delete", Handler: noOp})

	profile := e.Group("/profile")
	route.Add(profile, echo.Route{Method: http.MethodGet, Path: "/email", Name: "profile.email.get", Handler: noOp})
	route.Add(profile, echo.Route{Method: http.MethodPost, Path: "/signout", Name: "profile.signout.post", Handler: noOp})
	route.Add(profile, echo.Route{Method: http.MethodPost, Path: "/email", Name: "profile.email.post", Handler: noOp})
	route.Add(profile, echo.Route{Method: http.MethodGet, Path: "/sessions", Name: "profile.session.list", Handler: noOp})
	route.Add(profile, echo.Route{Method: http.MethodDelete, Path: "/sessions/:id", Name: "profile.session.revoke", Handler: noOp})
	route.Add(profile, echo.Route{Method: http.MethodDelete, Path: "/sessions", Name: "profile.session.revoke.all", Handler: noOp})

	passkeyAuthenticate := e.Group("/auth/passkeys/authenticate")
	route.Add(passkeyAuthenticate, echo.Route{Method: http.MethodPost, Path: "/begin", Name: "passkey.authenticate.begin", Handler: noOp})
	route.Add(passkeyAuthenticate, echo.Route{Method: http.MethodPost, Path: "/finish", Name: "passkey.authenticate.finish", Handler: noOp})

	passkeyRegister := e.Group("/auth/passkeys/register")
	route.Add(passkeyRegister, echo.Route{Method: http.MethodPost, Path: "/begin", Name: "passkey.register.begin", Handler: noOp})
	route.Add(passkeyRegister, echo.Route{Method: http.MethodPost, Path: "/finish", Name: "passkey.register.finish", Handler: noOp})

	passkeys := e.Group("/account/passkeys")
	route.Add(passkeys, echo.Route{Method: http.MethodGet, Path: "", Name: "passkey.list", Handler: noOp})
	route.Add(passkeys, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "passkey.row", Handler: noOp})
	route.Add(passkeys, echo.Route{Method: http.MethodGet, Path: "/:id/rename", Name: "passkey.rename.form", Handler: noOp})
	route.Add(passkeys, echo.Route{Method: http.MethodPost, Path: "/:id/rename", Name: "passkey.rename", Handler: noOp})
	route.Add(passkeys, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "passkey.delete", Handler: noOp})
}
