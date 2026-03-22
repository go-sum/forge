package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authmodel "github.com/y-goweb/auth/model"
	authrepo "github.com/y-goweb/auth/repository"
	"github.com/y-goweb/auth/session"
	"github.com/y-goweb/foundry/internal/routes"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type routeHandlers struct{}

func (routeHandlers) HealthCheck(c *echo.Context) error       { return c.String(http.StatusOK, "health") }
func (routeHandlers) Home(c *echo.Context) error              { return c.String(http.StatusOK, "home") }
func (routeHandlers) ComponentExamples(c *echo.Context) error { return c.String(http.StatusOK, "components") }
func (routeHandlers) UserList(c *echo.Context) error          { return c.String(http.StatusOK, "users") }
func (routeHandlers) UserEditForm(c *echo.Context) error      { return c.String(http.StatusOK, "user-edit") }
func (routeHandlers) UserRow(c *echo.Context) error           { return c.String(http.StatusOK, "user-row") }
func (routeHandlers) UserUpdate(c *echo.Context) error        { return c.String(http.StatusOK, "user-update") }
func (routeHandlers) UserDelete(c *echo.Context) error        { return c.String(http.StatusOK, "user-delete") }

type routeAuthHandlers struct{}

func (routeAuthHandlers) LoginPage(c *echo.Context) error    { return c.String(http.StatusOK, "login-page") }
func (routeAuthHandlers) Login(c *echo.Context) error        { return c.String(http.StatusOK, "login") }
func (routeAuthHandlers) RegisterPage(c *echo.Context) error { return c.String(http.StatusOK, "register-page") }
func (routeAuthHandlers) Register(c *echo.Context) error     { return c.String(http.StatusOK, "register") }
func (routeAuthHandlers) Logout(c *echo.Context) error       { return c.String(http.StatusOK, "logout") }

// routeAuthUserReader is a test double implementing authrepo.UserReader.
type routeAuthUserReader struct {
	user authmodel.User
	err  error
}

func (r routeAuthUserReader) GetByID(_ context.Context, _ uuid.UUID) (authmodel.User, error) {
	if r.err != nil {
		return authmodel.User{}, r.err
	}
	return r.user, nil
}

func (r routeAuthUserReader) GetByEmail(_ context.Context, _ string) (authmodel.User, error) {
	return authmodel.User{}, errors.New("unexpected GetByEmail call")
}

var _ authrepo.UserReader = routeAuthUserReader{}

func TestRegisterRoutesRedirectsUnauthenticatedProtectedRequests(t *testing.T) {
	e := echo.New()
	sessions := session.NewSessionStore(testSessionConfig())
	RegisterRoutes(e, routeHandlers{}, routeAuthHandlers{}, sessions, routeAuthUserReader{}, "/public", "public")

	req := httptest.NewRequest(http.MethodGet, routes.Users, nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != routes.Login {
		t.Fatalf("location = %q", got)
	}
}

func TestRegisterRoutesUsesHTMXRedirectForUnauthenticatedProtectedRequests(t *testing.T) {
	e := echo.New()
	sessions := session.NewSessionStore(testSessionConfig())
	RegisterRoutes(e, routeHandlers{}, routeAuthHandlers{}, sessions, routeAuthUserReader{}, "/public", "public")

	req := httptest.NewRequest(http.MethodGet, routes.Users, nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("HX-Redirect"); got != routes.Login {
		t.Fatalf("HX-Redirect = %q", got)
	}
}

func TestRegisterRoutesAllowsAuthenticatedComponentsRoute(t *testing.T) {
	e := echo.New()
	sessions := session.NewSessionStore(testSessionConfig())
	RegisterRoutes(e, routeHandlers{}, routeAuthHandlers{}, sessions, routeAuthUserReader{}, "/public", "public")

	req := httptest.NewRequest(http.MethodGet, routes.Components, nil)
	addSessionCookie(t, sessions, req, "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%q", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "components" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRegisterRoutesAllowsAuthenticatedUsersToReachUserManagement(t *testing.T) {
	e := echo.New()
	sessions := session.NewSessionStore(testSessionConfig())
	users := routeAuthUserReader{user: authmodel.User{
		ID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Role: "user",
	}}
	RegisterRoutes(e, routeHandlers{}, routeAuthHandlers{}, sessions, users, "/public", "public")

	req := httptest.NewRequest(http.MethodGet, routes.Users, nil)
	addSessionCookie(t, sessions, req, "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%q", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "users" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func testSessionConfig() session.SessionConfig {
	return session.SessionConfig{
		Name:       "test-session",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
		MaxAge:     3600,
	}
}

func addSessionCookie(t *testing.T, sessions *session.SessionManager, req *http.Request, userID string) {
	t.Helper()
	rec := httptest.NewRecorder()
	if err := sessions.SetUserID(rec, req, userID); err != nil {
		t.Fatalf("SetUserID() error = %v", err)
	}
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}
}
