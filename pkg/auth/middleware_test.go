package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/auth/model"
	authrepo "github.com/go-sum/auth/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const testSigninPath = "/signin"

// fakeMiddlewareSessionManager always returns the same state so tests can
// pre-populate it without simulating a cookie roundtrip.
type fakeMiddlewareSessionManager struct {
	state *fakeSessionState
}

func newFakeMiddlewareSessionManager() *fakeMiddlewareSessionManager {
	return &fakeMiddlewareSessionManager{state: newFakeSessionState()}
}

func (m *fakeMiddlewareSessionManager) Load(r *http.Request) (SessionState, error) {
	return m.state, nil
}

func (m *fakeMiddlewareSessionManager) Commit(w http.ResponseWriter, r *http.Request, s SessionState) error {
	return nil
}

func (m *fakeMiddlewareSessionManager) Destroy(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (m *fakeMiddlewareSessionManager) RotateID(w http.ResponseWriter, r *http.Request, s SessionState) error {
	return nil
}

func TestRequireAuthRedirectsToSignin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	err := RequireAuth(testSigninPath)(func(c *echo.Context) error {
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
	err := RequireAuth(testSigninPath)(func(c *echo.Context) error {
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
	mgr := newFakeMiddlewareSessionManager()
	if err := setAuth(mgr.state, "11111111-1111-1111-1111-111111111111", ""); err != nil {
		t.Fatalf("setAuth() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	err := LoadSession(mgr)(func(c *echo.Context) error {
		if got, _ := c.Get(ContextKeyUserID).(string); got != "11111111-1111-1111-1111-111111111111" {
			t.Fatalf("user ID = %q", got)
		}
		return nil
	})(c)

	if err != nil {
		t.Fatalf("LoadSession() error = %v", err)
	}
}

func TestLoadSessionSetsDisplayNameWhenPresentInSession(t *testing.T) {
	e := echo.New()
	mgr := newFakeMiddlewareSessionManager()
	if err := setAuth(mgr.state, "11111111-1111-1111-1111-111111111111", "Ada Lovelace"); err != nil {
		t.Fatalf("setAuth() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	err := LoadSession(mgr)(func(c *echo.Context) error {
		if got, _ := c.Get(ContextKeyDisplayName).(string); got != "Ada Lovelace" {
			t.Fatalf("display name = %q, want %q", got, "Ada Lovelace")
		}
		return nil
	})(c)

	if err != nil {
		t.Fatalf("LoadSession() error = %v", err)
	}
}

func TestLoadSessionDoesNotSetDisplayNameWhenAbsentFromSession(t *testing.T) {
	e := echo.New()
	mgr := newFakeMiddlewareSessionManager()
	// Set only the user ID, not the display name.
	if err := mgr.state.Put(sessionKeyUserID, "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	err := LoadSession(mgr)(func(c *echo.Context) error {
		if got := c.Get(ContextKeyDisplayName); got != nil {
			t.Fatalf("display name = %v, want nil", got)
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
		repo       fakeMiddlewareUserRepo
		wantStatus int
		wantRole   string
		expectNext bool
	}{
		{name: "missing session user ID", repo: fakeMiddlewareUserRepo{}, expectNext: true},
		{name: "invalid UUID", userID: "not-a-uuid", wantStatus: http.StatusUnauthorized},
		{
			name:       "user not found",
			userID:     "11111111-1111-1111-1111-111111111111",
			repo:       fakeMiddlewareUserRepo{getByIDErr: model.ErrUserNotFound},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "repository unavailable",
			userID:     "11111111-1111-1111-1111-111111111111",
			repo:       fakeMiddlewareUserRepo{getByIDErr: errors.New("db down")},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "success",
			userID:     "11111111-1111-1111-1111-111111111111",
			repo:       fakeMiddlewareUserRepo{user: model.User{Role: "admin", DisplayName: "Alice"}},
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
				c.Set(ContextKeyUserID, tc.userID)
			}

			nextCalled := false
			err := LoadUserRole(tc.repo)(func(c *echo.Context) error {
				nextCalled = true
				if tc.wantRole != "" {
					if got, _ := c.Get(ContextKeyUserRole).(string); got != tc.wantRole {
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
			assertHTTPErrorStatus(t, err, tc.wantStatus)
		})
	}
}

func TestRequireAdminRejectsNonAdmin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(ContextKeyUserRole, "user")

	called := false
	err := RequireAdmin()(func(c *echo.Context) error {
		called = true
		return nil
	})(c)

	if called {
		t.Fatal("next handler should not be called")
	}
	assertHTTPErrorStatus(t, err, http.StatusForbidden)
}

func TestRequireAdminAllowsAdmin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(ContextKeyUserRole, "admin")

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

type fakeMiddlewareUserRepo struct {
	user       model.User
	getByIDErr error
}

func (r fakeMiddlewareUserRepo) GetByID(context.Context, uuid.UUID) (model.User, error) {
	if r.getByIDErr != nil {
		return model.User{}, r.getByIDErr
	}
	return r.user, nil
}

func (r fakeMiddlewareUserRepo) GetByEmail(context.Context, string) (model.User, error) {
	return model.User{}, errors.New("unexpected GetByEmail call")
}

var _ authrepo.UserReader = fakeMiddlewareUserRepo{}

func assertHTTPErrorStatus(t *testing.T, err error, status int) {
	t.Helper()
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("err = %T (%v), want *HTTPError", err, err)
	}
	if httpErr.StatusCode() != status {
		t.Fatalf("status = %d, want %d", httpErr.StatusCode(), status)
	}
}
