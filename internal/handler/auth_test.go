package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"starter/internal/model"
	"starter/internal/routes"

	"github.com/labstack/echo/v5"
)

func TestLoginPageRenders(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{}, nil)
	c, rec := newRequestContext(http.MethodGet, routes.Login, nil)
	setCSRFToken(c)
	setUserID(c, testUser.ID.String())

	if err := h.LoginPage(c); err != nil {
		t.Fatalf("LoginPage() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Sign In") || !strings.Contains(body, `value="`+testCSRFToken+`"`) {
		t.Fatalf("body = %q", body)
	}
}

func TestLoginValidationFailureRenders422(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{}, nil)
	c, rec := newFormContext(http.MethodPost, routes.Login, url.Values{
		"email": {"not-an-email"},
	})
	setCSRFToken(c)

	if err := h.Login(c); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Sign In") || !strings.Contains(body, `value="not-an-email"`) {
		t.Fatalf("body = %q", body)
	}
}

func TestLoginInvalidCredentialsRenders401(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		loginFn: func(context.Context, model.LoginInput) (model.User, error) {
			return model.User{}, model.ErrInvalidCredentials
		},
	}, fakeUserService{}, nil)
	c, rec := newFormContext(http.MethodPost, routes.Login, url.Values{
		"email":    {"ada@example.com"},
		"password": {"wrong-password"},
	})
	setCSRFToken(c)

	if err := h.Login(c); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Invalid email or password.") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestLoginRedirectsOnSuccess(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		loginFn: func(_ context.Context, input model.LoginInput) (model.User, error) {
			if input.Email != "ada@example.com" || input.Password != "correct-password" {
				t.Fatalf("input = %#v", input)
			}
			return testUser, nil
		},
	}, fakeUserService{}, nil)
	c, rec := newFormContext(http.MethodPost, routes.Login, url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct-password"},
	})

	if err := h.Login(c); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != routes.Home {
		t.Fatalf("location = %q", got)
	}
	if !strings.Contains(rec.Header().Get(echo.HeaderSetCookie), "test-session=") {
		t.Fatalf("set-cookie = %q", rec.Header().Get(echo.HeaderSetCookie))
	}
}

func TestLoginRedirectsHTMXOnSuccess(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		loginFn: func(context.Context, model.LoginInput) (model.User, error) {
			return testUser, nil
		},
	}, fakeUserService{}, nil)
	c, rec := newFormContext(http.MethodPost, routes.Login, url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct-password"},
	})
	c.Request().Header.Set("HX-Request", "true")

	if err := h.Login(c); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("HX-Redirect"); got != routes.Home {
		t.Fatalf("HX-Redirect = %q", got)
	}
}

func TestRegisterPageRenders(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{}, nil)
	c, rec := newRequestContext(http.MethodGet, routes.Register, nil)
	setCSRFToken(c)

	if err := h.RegisterPage(c); err != nil {
		t.Fatalf("RegisterPage() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Create Account") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRegisterValidationFailureRenders422(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{}, nil)
	c, rec := newFormContext(http.MethodPost, routes.Register, url.Values{
		"email":        {"ada@example.com"},
		"display_name": {"Ada"},
		"password":     {"short"},
	})
	setCSRFToken(c)

	if err := h.Register(c); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Create Account") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRegisterConflictRenders409(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		registerFn: func(context.Context, model.CreateUserInput) (model.User, error) {
			return model.User{}, model.ErrEmailTaken
		},
	}, fakeUserService{}, nil)
	c, rec := newFormContext(http.MethodPost, routes.Register, url.Values{
		"email":        {"ada@example.com"},
		"display_name": {"Ada"},
		"password":     {"correct-password"},
	})
	setCSRFToken(c)

	if err := h.Register(c); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Email already in use.") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRegisterRedirectsOnSuccess(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		registerFn: func(_ context.Context, input model.CreateUserInput) (model.User, error) {
			if input.Email != "ada@example.com" || input.DisplayName != "Ada" || input.Password != "correct-password" {
				t.Fatalf("input = %#v", input)
			}
			return testUser, nil
		},
	}, fakeUserService{}, nil)
	c, rec := newFormContext(http.MethodPost, routes.Register, url.Values{
		"email":        {"ada@example.com"},
		"display_name": {"Ada"},
		"password":     {"correct-password"},
	})

	if err := h.Register(c); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != routes.Login {
		t.Fatalf("location = %q", got)
	}
	if !strings.Contains(rec.Header().Get(echo.HeaderSetCookie), "flash=") {
		t.Fatalf("set-cookie = %q", rec.Header().Get(echo.HeaderSetCookie))
	}
}

func TestLogoutClearsSession(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{}, nil)
	c, rec := newRequestContext(http.MethodPost, routes.Logout, nil)

	if err := h.Logout(c); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != routes.Login {
		t.Fatalf("location = %q", got)
	}
	if !strings.Contains(rec.Header().Get(echo.HeaderSetCookie), "Max-Age=0") &&
		!strings.Contains(rec.Header().Get(echo.HeaderSetCookie), "Max-Age=-1") {
		t.Fatalf("set-cookie = %q", rec.Header().Get(echo.HeaderSetCookie))
	}
}

func TestLoginInternalErrorPropagates(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		loginFn: func(context.Context, model.LoginInput) (model.User, error) {
			return model.User{}, errors.New("database timeout")
		},
	}, fakeUserService{}, nil)
	c, _ := newFormContext(http.MethodPost, routes.Login, url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct-password"},
	})

	err := h.Login(c)
	assertAppErrorStatus(t, err, http.StatusInternalServerError)
}
