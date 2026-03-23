package echocomponentry

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/auth/model"
	authrepo "github.com/go-sum/auth/repository"
	"github.com/go-sum/auth/session"
	"github.com/go-sum/server/apperr"
	cfgs "github.com/go-sum/server/config"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const testSigninPath = "/signin"

var testContextKeys = ContextKeys{
	UserID:      "user_id",
	UserRole:    "user_role",
	DisplayName: "user_display_name",
}

func TestRequireAuthRedirectsToSignin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	err := RequireAuth(testSigninPath, testContextKeys)(func(c *echo.Context) error {
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
	if got := rec.Header().Get(echo.HeaderLocation); got != testSigninPath {
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
	err := RequireAuth(testSigninPath, testContextKeys)(func(c *echo.Context) error {
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
	if got := rec.Header().Get("HX-Redirect"); got != testSigninPath {
		t.Fatalf("HX-Redirect = %q", got)
	}
}

func TestLoadSessionSetsUserIDWhenSessionExists(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	sessions, err := session.NewSessionStore(testSessionConfig())
	if err != nil {
		t.Fatalf("NewSessionStore() error = %v", err)
	}
	if err := sessions.SetUserID(rec, req, "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Fatalf("SetUserID() error = %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/users", nil)
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	c := e.NewContext(req, httptest.NewRecorder())

	err = LoadSession(sessions, testContextKeys)(func(c *echo.Context) error {
		if got, _ := cfgs.Get[string](c, testContextKeys.UserID); got != "11111111-1111-1111-1111-111111111111" {
			t.Fatalf("user ID = %q", got)
		}
		return nil
	})(c)

	if err != nil {
		t.Fatalf("LoadSession() error = %v", err)
	}
}

func TestLoadUserContextHandlesOutcomes(t *testing.T) {
	tests := []struct {
		name            string
		userID          string
		repo            middlewareUserRepo
		wantStatus      int
		wantRole        string
		wantDisplayName string
		expectNext      bool
	}{
		{name: "missing session user ID", repo: middlewareUserRepo{}, expectNext: true},
		{name: "invalid UUID", userID: "not-a-uuid", wantStatus: http.StatusUnauthorized},
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
			name:            "success",
			userID:          "11111111-1111-1111-1111-111111111111",
			repo:            middlewareUserRepo{user: model.User{Role: "admin", DisplayName: "Alice"}},
			wantRole:        "admin",
			wantDisplayName: "Alice",
			expectNext:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tc.userID != "" {
				cfgs.Set(c, testContextKeys.UserID, tc.userID)
			}

			nextCalled := false
			err := LoadUserContext(tc.repo, testContextKeys)(func(c *echo.Context) error {
				nextCalled = true
				if tc.wantRole != "" {
					if got, _ := cfgs.Get[string](c, testContextKeys.UserRole); got != tc.wantRole {
						t.Fatalf("role = %q", got)
					}
				}
				if tc.wantDisplayName != "" {
					if got, _ := cfgs.Get[string](c, testContextKeys.DisplayName); got != tc.wantDisplayName {
						t.Fatalf("display name = %q", got)
					}
				}
				return nil
			})(c)

			if tc.expectNext != nextCalled {
				t.Fatalf("next called = %v, want %v", nextCalled, tc.expectNext)
			}
			if tc.wantStatus == 0 {
				if err != nil {
					t.Fatalf("LoadUserContext() error = %v", err)
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
	cfgs.Set(c, testContextKeys.UserRole, "user")

	called := false
	err := RequireAdmin(testContextKeys)(func(c *echo.Context) error {
		called = true
		return nil
	})(c)

	if called {
		t.Fatal("next handler should not be called")
	}
	assertMiddlewareAppErrorStatus(t, err, http.StatusForbidden)
}

func TestRequireAdminAllowsAdmin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	cfgs.Set(c, testContextKeys.UserRole, "admin")

	called := false
	err := RequireAdmin(testContextKeys)(func(c *echo.Context) error {
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

func (r middlewareUserRepo) GetByID(context.Context, uuid.UUID) (model.User, error) {
	if r.getByIDErr != nil {
		return model.User{}, r.getByIDErr
	}
	return r.user, nil
}

func (r middlewareUserRepo) GetByEmail(context.Context, string) (model.User, error) {
	return model.User{}, errors.New("unexpected GetByEmail call")
}

var _ authrepo.UserReader = middlewareUserRepo{}

func testSessionConfig() session.SessionConfig {
	return session.SessionConfig{
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
