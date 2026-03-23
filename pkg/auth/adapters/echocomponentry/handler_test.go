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

	"github.com/go-sum/auth/model"
	"github.com/go-sum/auth/session"
	"github.com/go-sum/server/apperr"
	servervalidate "github.com/go-sum/server/validate"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
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
	signinFn func(context.Context, model.SigninInput) (model.User, error)
	signupFn func(context.Context, model.SignupInput) (model.User, error)
}

func (f fakeAuthService) Signin(ctx context.Context, input model.SigninInput) (model.User, error) {
	if f.signinFn != nil {
		return f.signinFn(ctx, input)
	}
	return model.User{}, errors.New("unexpected Signin call")
}

func (f fakeAuthService) Signup(ctx context.Context, input model.SignupInput) (model.User, error) {
	if f.signupFn != nil {
		return f.signupFn(ctx, input)
	}
	return model.User{}, errors.New("unexpected Signup call")
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
			SigninPath: "/signin",
			SignupPath: "/signup",
			HomePath:   "/",
			CSRFField:  "_csrf",
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

func TestSigninPageRenders(t *testing.T) {
	h := newTestHandler(fakeAuthService{})
	c, rec := newRequestContext(http.MethodGet, "/signin", nil)
	setCSRFToken(c)

	if err := h.SigninPage(c); err != nil {
		t.Fatalf("SigninPage() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Sign In") || !strings.Contains(body, `value="`+testCSRFToken+`"`) {
		t.Fatalf("body = %q", body)
	}
}

func TestSigninValidationFailureRenders422(t *testing.T) {
	h := newTestHandler(fakeAuthService{})
	c, rec := newFormContext(http.MethodPost, "/signin", url.Values{
		"email": {"not-an-email"},
	})
	setCSRFToken(c)

	if err := h.Signin(c); err != nil {
		t.Fatalf("Signin() error = %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Sign In") || !strings.Contains(body, `value="not-an-email"`) {
		t.Fatalf("body = %q", body)
	}
}

func TestSigninInvalidCredentialsRenders401(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		signinFn: func(context.Context, model.SigninInput) (model.User, error) {
			return model.User{}, model.ErrInvalidCredentials
		},
	})
	c, rec := newFormContext(http.MethodPost, "/signin", url.Values{
		"email":    {"ada@example.com"},
		"password": {"wrong-password"},
	})
	setCSRFToken(c)

	if err := h.Signin(c); err != nil {
		t.Fatalf("Signin() error = %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Invalid email or password.") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestSigninRedirectsOnSuccess(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		signinFn: func(_ context.Context, input model.SigninInput) (model.User, error) {
			if input.Email != "ada@example.com" || input.Password != "correct-password" {
				t.Fatalf("input = %#v", input)
			}
			return handlerTestUser, nil
		},
	})
	c, rec := newFormContext(http.MethodPost, "/signin", url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct-password"},
	})

	if err := h.Signin(c); err != nil {
		t.Fatalf("Signin() error = %v", err)
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

func TestSigninRedirectsHTMXOnSuccess(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		signinFn: func(context.Context, model.SigninInput) (model.User, error) {
			return handlerTestUser, nil
		},
	})
	c, rec := newFormContext(http.MethodPost, "/signin", url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct-password"},
	})
	c.Request().Header.Set("HX-Request", "true")

	if err := h.Signin(c); err != nil {
		t.Fatalf("Signin() error = %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("HX-Redirect"); got != "/" {
		t.Fatalf("HX-Redirect = %q", got)
	}
}

func TestSignupPageRenders(t *testing.T) {
	h := newTestHandler(fakeAuthService{})
	c, rec := newRequestContext(http.MethodGet, "/signup", nil)
	setCSRFToken(c)

	if err := h.SignupPage(c); err != nil {
		t.Fatalf("SignupPage() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Create Account") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestSignupValidationFailureRenders422(t *testing.T) {
	h := newTestHandler(fakeAuthService{})
	c, rec := newFormContext(http.MethodPost, "/signup", url.Values{
		"email":        {"ada@example.com"},
		"display_name": {"Ada"},
		"password":     {"short"},
	})
	setCSRFToken(c)

	if err := h.Signup(c); err != nil {
		t.Fatalf("Signup() error = %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Create Account") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestSignupConflictRenders409(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		signupFn: func(context.Context, model.SignupInput) (model.User, error) {
			return model.User{}, model.ErrEmailTaken
		},
	})
	c, rec := newFormContext(http.MethodPost, "/signup", url.Values{
		"email":        {"ada@example.com"},
		"display_name": {"Ada"},
		"password":     {"correct-password"},
	})
	setCSRFToken(c)

	if err := h.Signup(c); err != nil {
		t.Fatalf("Signup() error = %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Email already in use.") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestSignupRedirectsOnSuccess(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		signupFn: func(_ context.Context, input model.SignupInput) (model.User, error) {
			if input.Email != "ada@example.com" || input.DisplayName != "Ada" || input.Password != "correct-password" {
				t.Fatalf("input = %#v", input)
			}
			return handlerTestUser, nil
		},
	})
	c, rec := newFormContext(http.MethodPost, "/signup", url.Values{
		"email":        {"ada@example.com"},
		"display_name": {"Ada"},
		"password":     {"correct-password"},
	})

	if err := h.Signup(c); err != nil {
		t.Fatalf("Signup() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != "/signin" {
		t.Fatalf("location = %q", got)
	}
	if !strings.Contains(rec.Header().Get(echo.HeaderSetCookie), "flash=") {
		t.Fatalf("set-cookie = %q", rec.Header().Get(echo.HeaderSetCookie))
	}
}

func TestSignoutClearsSession(t *testing.T) {
	h := newTestHandler(fakeAuthService{})
	c, rec := newRequestContext(http.MethodPost, "/signout", nil)

	if err := h.Signout(c); err != nil {
		t.Fatalf("Signout() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != "/signin" {
		t.Fatalf("location = %q", got)
	}
	if !strings.Contains(rec.Header().Get(echo.HeaderSetCookie), "Max-Age=0") &&
		!strings.Contains(rec.Header().Get(echo.HeaderSetCookie), "Max-Age=-1") {
		t.Fatalf("set-cookie = %q", rec.Header().Get(echo.HeaderSetCookie))
	}
}

func TestSigninInternalErrorPropagates(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		signinFn: func(context.Context, model.SigninInput) (model.User, error) {
			return model.User{}, errors.New("database timeout")
		},
	})
	c, _ := newFormContext(http.MethodPost, "/signin", url.Values{
		"email":    {"ada@example.com"},
		"password": {"correct-password"},
	})

	err := h.Signin(c)
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
