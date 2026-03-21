package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"starter/internal/apperr"
	"starter/internal/model"
	"starter/internal/service"
	"starter/pkg/auth"
	"starter/pkg/ctxkeys"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const testLoginPath = "/login"

func TestRequireAuthRedirectsToLogin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// ctxkeys.UserID is not set — simulates LoadSession finding no session.

	called := false
	err := RequireAuth(testLoginPath)(func(c *echo.Context) error {
		called = true
		return nil
	})(c)

	if called {
		t.Fatal("next handler should not be called")
	}
	if err != nil {
		t.Fatalf("RequireAuth() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != testLoginPath {
		t.Fatalf("location = %q", got)
	}
}

func TestRequireAuthSetsHTMXRedirectHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	err := RequireAuth(testLoginPath)(func(c *echo.Context) error {
		called = true
		return nil
	})(c)

	if called {
		t.Fatal("next handler should not be called")
	}
	if err != nil {
		t.Fatalf("RequireAuth() error = %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(htmxRedirectHeader); got != testLoginPath {
		t.Fatalf("HX-Redirect = %q", got)
	}
}

func TestLoadSessionSetsUserIDWhenSessionExists(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	sessions := auth.NewSessionStore(testSessionConfig())
	if err := sessions.SetUserID(rec, req, "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Fatalf("SetUserID() error = %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/users", nil)
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	c := e.NewContext(req, httptest.NewRecorder())

	err := LoadSession(sessions)(func(c *echo.Context) error {
		if got, _ := c.Get(string(ctxkeys.UserID)).(string); got != "11111111-1111-1111-1111-111111111111" {
			t.Fatalf("user ID = %q", got)
		}
		return nil
	})(c)

	if err != nil {
		t.Fatalf("LoadSession() error = %v", err)
	}
}

func TestLoadUserRoleHandlesOutcomes(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		repo       middlewareUserRepo
		wantStatus int
		wantRole   string
		expectNext bool
	}{
		{
			name:       "missing session user ID",
			repo:       middlewareUserRepo{},
			expectNext: true,
		},
		{
			name:       "invalid UUID",
			userID:     "not-a-uuid",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "user not found",
			userID:     "11111111-1111-1111-1111-111111111111",
			repo:       middlewareUserRepo{getByIDErr: model.ErrUserNotFound},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "repository unavailable",
			userID:     "11111111-1111-1111-1111-111111111111",
			repo:       middlewareUserRepo{getByIDErr: errors.New("db down")},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "success",
			userID:     "11111111-1111-1111-1111-111111111111",
			repo:       middlewareUserRepo{user: model.User{Role: "admin"}},
			wantRole:   "admin",
			expectNext: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tc.userID != "" {
				c.Set(string(ctxkeys.UserID), tc.userID)
			}

			nextCalled := false
			err := LoadUserRole(service.NewUserService(tc.repo))(func(c *echo.Context) error {
				nextCalled = true
				if tc.wantRole != "" {
					if got, _ := c.Get(string(ctxkeys.UserRole)).(string); got != tc.wantRole {
						t.Fatalf("role = %q", got)
					}
				}
				return nil
			})(c)

			if tc.expectNext != nextCalled {
				t.Fatalf("next called = %v, want %v", nextCalled, tc.expectNext)
			}
			if tc.wantStatus == 0 {
				if err != nil {
					t.Fatalf("LoadUserRole() error = %v", err)
				}
				return
			}
			assertMiddlewareAppErrorStatus(t, err, tc.wantStatus)
		})
	}
}

func TestRequireAdminRejectsNonAdmin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(ctxkeys.UserRole), "user")

	called := false
	err := RequireAdmin()(func(c *echo.Context) error {
		called = true
		return nil
	})(c)

	if called {
		t.Fatal("next handler should not be called")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %T, want *apperr.Error", err)
	}
	if appErr.Status != http.StatusForbidden {
		t.Fatalf("status = %d", appErr.Status)
	}
}

func TestRequireAdminAllowsAdmin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(ctxkeys.UserRole), "admin")

	called := false
	err := RequireAdmin()(func(c *echo.Context) error {
		called = true
		return nil
	})(c)

	if err != nil {
		t.Fatalf("RequireAdmin() error = %v", err)
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}

type middlewareUserRepo struct {
	user       model.User
	getByIDErr error
}

func (r middlewareUserRepo) Create(context.Context, string, string, string) (model.User, error) {
	return model.User{}, errors.New("unexpected Create call")
}

func (r middlewareUserRepo) GetByID(context.Context, uuid.UUID) (model.User, error) {
	if r.getByIDErr != nil {
		return model.User{}, r.getByIDErr
	}
	return r.user, nil
}

func (r middlewareUserRepo) GetByEmail(context.Context, string) (model.User, error) {
	return model.User{}, errors.New("unexpected GetByEmail call")
}

func (r middlewareUserRepo) List(context.Context, int32, int32) ([]model.User, error) {
	return nil, errors.New("unexpected List call")
}

func (r middlewareUserRepo) Update(context.Context, uuid.UUID, string, string, string) (model.User, error) {
	return model.User{}, errors.New("unexpected Update call")
}

func (r middlewareUserRepo) Delete(context.Context, uuid.UUID) error {
	return errors.New("unexpected Delete call")
}

func (r middlewareUserRepo) Count(context.Context) (int64, error) {
	return 0, errors.New("unexpected Count call")
}

func testSessionConfig() auth.SessionConfig {
	return auth.SessionConfig{
		Name:       "test-session",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
		MaxAge:     3600,
	}
}

func assertMiddlewareAppErrorStatus(t *testing.T, err error, status int) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %T, want *apperr.Error", err)
	}
	if appErr.Status != status {
		t.Fatalf("status = %d, want %d", appErr.Status, status)
	}
}
