package echocomponentry

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/go-sum/auth/model"
	"github.com/go-sum/auth/session"
	"github.com/go-sum/server/apperr"
	servervalidate "github.com/go-sum/server/validate"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const testCSRFToken = "csrf-token"

var handlerTestUser = model.User{
	ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	Email:       "ada@example.com",
	DisplayName: "Ada Lovelace",
	Role:        "admin",
	CreatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
	UpdatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
}

type fakeAuthService struct {
	loginFn    func(context.Context, model.LoginInput) (model.User, error)
	registerFn func(context.Context, model.CreateUserInput) (model.User, error)
}

func (f fakeAuthService) Login(ctx context.Context, input model.LoginInput) (model.User, error) {
	if f.loginFn != nil {
		return f.loginFn(ctx, input)
	}
	return model.User{}, errors.New("unexpected Login call")
}

func (f fakeAuthService) Register(ctx context.Context, input model.CreateUserInput) (model.User, error) {
	if f.registerFn != nil {
		return f.registerFn(ctx, input)
	}
	return model.User{}, errors.New("unexpected Register call")
}

func newTestHandler(svc authService) *Handler {
	sessions, err := session.NewSessionStore(session.SessionConfig{
		Name:       "test-session",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
		MaxAge:     3600,
	})
	if err != nil {
		panic(err)
	}
	return New(
		svc,
		sessions,
		servervalidate.New(),
		Config{
			LoginPath:    "/login",
			RegisterPath: "/register",
			HomePath:     "/",
			CSRFField:    "_csrf",
			RequestFn: func(*echo.Context) Request {
				return Request{
					CSRFToken: testCSRFToken,
					PageFn: func(_ string, children ...g.Node) g.Node {
						return h.Div(children...)
					},
				}
			},
		},
	)
}

func newRequestContext(method, target string, body io.Reader) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, target, body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func newFormContext(method, target string, values url.Values) (*echo.Context, *httptest.ResponseRecorder) {
	c, rec := newRequestContext(method, target, strings.NewReader(values.Encode()))
	c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	return c, rec
}

func setCSRFToken(c *echo.Context) {
	c.Set(echomw.DefaultCSRFConfig.ContextKey, testCSRFToken)
}

func TestLoginPageRenders(t *testing.T) {
	h := newTestHandler(fakeAuthService{})
	c, rec := newRequestContext(http.MethodGet, "/login", nil)
	setCSRFToken(c)

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
	h := newTestHandler(fakeAuthService{})
	c, rec := newFormContext(http.MethodPost, "/login", url.Values{
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
	})
	c, rec := newFormContext(http.MethodPost, "/login", url.Values{
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
			return handlerTestUser, nil
		},
	})
	c, rec := newFormContext(http.MethodPost, "/login", url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct-password"},
	})

	if err := h.Login(c); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != "/" {
		t.Fatalf("location = %q", got)
	}
	if !strings.Contains(rec.Header().Get(echo.HeaderSetCookie), "test-session=") {
		t.Fatalf("set-cookie = %q", rec.Header().Get(echo.HeaderSetCookie))
	}
}

func TestLoginRedirectsHTMXOnSuccess(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		loginFn: func(context.Context, model.LoginInput) (model.User, error) {
			return handlerTestUser, nil
		},
	})
	c, rec := newFormContext(http.MethodPost, "/login", url.Values{
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
	if got := rec.Header().Get("HX-Redirect"); got != "/" {
		t.Fatalf("HX-Redirect = %q", got)
	}
}

func TestRegisterPageRenders(t *testing.T) {
	h := newTestHandler(fakeAuthService{})
	c, rec := newRequestContext(http.MethodGet, "/register", nil)
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
	h := newTestHandler(fakeAuthService{})
	c, rec := newFormContext(http.MethodPost, "/register", url.Values{
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
	})
	c, rec := newFormContext(http.MethodPost, "/register", url.Values{
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
			return handlerTestUser, nil
		},
	})
	c, rec := newFormContext(http.MethodPost, "/register", url.Values{
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
	if got := rec.Header().Get(echo.HeaderLocation); got != "/login" {
		t.Fatalf("location = %q", got)
	}
	if !strings.Contains(rec.Header().Get(echo.HeaderSetCookie), "flash=") {
		t.Fatalf("set-cookie = %q", rec.Header().Get(echo.HeaderSetCookie))
	}
}

func TestLogoutClearsSession(t *testing.T) {
	h := newTestHandler(fakeAuthService{})
	c, rec := newRequestContext(http.MethodPost, "/logout", nil)

	if err := h.Logout(c); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != "/login" {
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
	})
	c, _ := newFormContext(http.MethodPost, "/login", url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct-password"},
	})

	err := h.Login(c)
	assertAppErrorStatus(t, err, http.StatusInternalServerError)
}

func assertAppErrorStatus(t *testing.T, err error, status int) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %T, want *apperr.Error", err)
	}
	if appErr.Status != status {
		t.Fatalf("status = %d, want %d", appErr.Status, status)
	}
}
