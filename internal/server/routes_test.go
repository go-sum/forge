package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"starter/internal/model"
	"starter/internal/routes"
	"starter/internal/service"
	"starter/pkg/auth"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type routeHandlers struct{}

func (routeHandlers) HealthCheck(c *echo.Context) error { return c.String(http.StatusOK, "health") }
func (routeHandlers) Home(c *echo.Context) error        { return c.String(http.StatusOK, "home") }
func (routeHandlers) LoginPage(c *echo.Context) error   { return c.String(http.StatusOK, "login-page") }
func (routeHandlers) Login(c *echo.Context) error       { return c.String(http.StatusOK, "login") }
func (routeHandlers) RegisterPage(c *echo.Context) error {
	return c.String(http.StatusOK, "register-page")
}
func (routeHandlers) Register(c *echo.Context) error { return c.String(http.StatusOK, "register") }
func (routeHandlers) Logout(c *echo.Context) error   { return c.String(http.StatusOK, "logout") }
func (routeHandlers) ComponentExamples(c *echo.Context) error {
	return c.String(http.StatusOK, "components")
}
func (routeHandlers) UserList(c *echo.Context) error     { return c.String(http.StatusOK, "users") }
func (routeHandlers) UserEditForm(c *echo.Context) error { return c.String(http.StatusOK, "user-edit") }
func (routeHandlers) UserRow(c *echo.Context) error      { return c.String(http.StatusOK, "user-row") }
func (routeHandlers) UserUpdate(c *echo.Context) error   { return c.String(http.StatusOK, "user-update") }
func (routeHandlers) UserDelete(c *echo.Context) error   { return c.String(http.StatusOK, "user-delete") }

type routeUserRepo struct {
	user model.User
	err  error
}

func (r routeUserRepo) Create(context.Context, string, string, string) (model.User, error) {
	return model.User{}, errors.New("unexpected Create call")
}

func (r routeUserRepo) GetByID(context.Context, uuid.UUID) (model.User, error) {
	if r.err != nil {
		return model.User{}, r.err
	}
	return r.user, nil
}

func (r routeUserRepo) GetByEmail(context.Context, string) (model.User, error) {
	return model.User{}, errors.New("unexpected GetByEmail call")
}

func (r routeUserRepo) List(context.Context, int32, int32) ([]model.User, error) {
	return nil, errors.New("unexpected List call")
}

func (r routeUserRepo) Update(context.Context, uuid.UUID, string, string, string) (model.User, error) {
	return model.User{}, errors.New("unexpected Update call")
}

func (r routeUserRepo) Delete(context.Context, uuid.UUID) error {
	return errors.New("unexpected Delete call")
}

func (r routeUserRepo) Count(context.Context) (int64, error) {
	return 0, errors.New("unexpected Count call")
}

func TestRegisterRoutesRedirectsUnauthenticatedProtectedRequests(t *testing.T) {
	e := echo.New()
	sessions := auth.NewSessionStore(testSessionConfig())
	RegisterRoutes(e, routeHandlers{}, sessions, service.NewUserService(routeUserRepo{}), "/public", "public")

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
	sessions := auth.NewSessionStore(testSessionConfig())
	RegisterRoutes(e, routeHandlers{}, sessions, service.NewUserService(routeUserRepo{}), "/public", "public")

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
	sessions := auth.NewSessionStore(testSessionConfig())
	RegisterRoutes(e, routeHandlers{}, sessions, service.NewUserService(routeUserRepo{}), "/public", "public")

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
	sessions := auth.NewSessionStore(testSessionConfig())
	users := service.NewUserService(routeUserRepo{user: model.User{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Role: "user"}})
	RegisterRoutes(e, routeHandlers{}, sessions, users, "/public", "public")

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

func testSessionConfig() auth.SessionConfig {
	return auth.SessionConfig{
		Name:       "test-session",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
		MaxAge:     3600,
	}
}

func addSessionCookie(t *testing.T, sessions *auth.SessionManager, req *http.Request, userID string) {
	t.Helper()
	rec := httptest.NewRecorder()
	if err := sessions.SetUserID(rec, req, userID); err != nil {
		t.Fatalf("SetUserID() error = %v", err)
	}
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}
}
