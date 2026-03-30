package authui

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
	"github.com/go-sum/forge/internal/adapters/authsession"
	"github.com/go-sum/server/apperr"
	"github.com/go-sum/session"
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
	Verified:    true,
	CreatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
	UpdatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
}

type fakeAuthService struct {
	beginSigninFn      func(context.Context, model.BeginSigninInput, string) (model.PendingFlow, error)
	beginSignupFn      func(context.Context, model.BeginSignupInput, string) (model.PendingFlow, error)
	beginEmailChangeFn func(context.Context, uuid.UUID, model.BeginEmailChangeInput, string) (model.PendingFlow, error)
	resendPendingFn    func(context.Context, model.PendingFlow, string) (model.PendingFlow, error)
	verifyPendingFn    func(context.Context, model.PendingFlow, model.VerifyInput) (model.VerifyResult, error)
	verifyTokenFn      func(context.Context, string, model.VerifyInput) (model.VerifyResult, error)
	verifyPageStateFn  func(string) (model.VerifyPageState, error)
}

func (f fakeAuthService) BeginSignin(ctx context.Context, input model.BeginSigninInput, verifyURL string) (model.PendingFlow, error) {
	if f.beginSigninFn != nil {
		return f.beginSigninFn(ctx, input, verifyURL)
	}
	return model.PendingFlow{}, errors.New("unexpected BeginSignin call")
}

func (f fakeAuthService) BeginSignup(ctx context.Context, input model.BeginSignupInput, verifyURL string) (model.PendingFlow, error) {
	if f.beginSignupFn != nil {
		return f.beginSignupFn(ctx, input, verifyURL)
	}
	return model.PendingFlow{}, errors.New("unexpected BeginSignup call")
}

func (f fakeAuthService) BeginEmailChange(ctx context.Context, userID uuid.UUID, input model.BeginEmailChangeInput, verifyURL string) (model.PendingFlow, error) {
	if f.beginEmailChangeFn != nil {
		return f.beginEmailChangeFn(ctx, userID, input, verifyURL)
	}
	return model.PendingFlow{}, errors.New("unexpected BeginEmailChange call")
}

func (f fakeAuthService) ResendPendingFlow(ctx context.Context, flow model.PendingFlow, verifyURL string) (model.PendingFlow, error) {
	if f.resendPendingFn != nil {
		return f.resendPendingFn(ctx, flow, verifyURL)
	}
	return model.PendingFlow{}, errors.New("unexpected ResendPendingFlow call")
}

func (f fakeAuthService) VerifyPendingFlow(ctx context.Context, flow model.PendingFlow, input model.VerifyInput) (model.VerifyResult, error) {
	if f.verifyPendingFn != nil {
		return f.verifyPendingFn(ctx, flow, input)
	}
	return model.VerifyResult{}, errors.New("unexpected VerifyPendingFlow call")
}

func (f fakeAuthService) VerifyToken(ctx context.Context, token string, input model.VerifyInput) (model.VerifyResult, error) {
	if f.verifyTokenFn != nil {
		return f.verifyTokenFn(ctx, token, input)
	}
	return model.VerifyResult{}, errors.New("unexpected VerifyToken call")
}

func (f fakeAuthService) VerifyPageState(token string) (model.VerifyPageState, error) {
	if f.verifyPageStateFn != nil {
		return f.verifyPageStateFn(token)
	}
	return model.VerifyPageState{}, errors.New("unexpected VerifyPageState call")
}

func newTestHandler(svc authService) *Handler {
	mgr, err := session.NewManager(session.Config{
		CookieName: "test-session",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
		MaxAge:     3600,
	})
	if err != nil {
		panic(err)
	}
	return New(
		svc,
		mgr,
		servervalidate.New(),
		Config{
			SigninPath:       "/signin",
			SignupPath:       "/signup",
			VerifyPath:       "/verify",
			VerifyResendPath: "/verify/resend",
			VerifyURL:        "https://example.com/verify",
			EmailChangePath:  "/account/email",
			HomePath:         "/",
			CSRFField:        "_csrf",
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
	if !strings.Contains(body, "Send Code") || !strings.Contains(body, `value="`+testCSRFToken+`"`) {
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
	if !strings.Contains(rec.Body.String(), `value="not-an-email"`) {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestSigninRedirectsToVerifyAndStoresPendingFlow(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		beginSigninFn: func(_ context.Context, input model.BeginSigninInput, verifyURL string) (model.PendingFlow, error) {
			if input.Email != "ada@example.com" || verifyURL != "https://example.com/verify" {
				t.Fatalf("input=%#v verifyURL=%q", input, verifyURL)
			}
			return model.PendingFlow{
				Purpose:   model.FlowPurposeSignin,
				Email:     input.Email,
				Secret:    "SECRET",
				IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
				ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
			}, nil
		},
	})
	c, rec := newFormContext(http.MethodPost, "/signin", url.Values{
		"email": {"ada@example.com"},
	})
	setCSRFToken(c)

	if err := h.Signin(c); err != nil {
		t.Fatalf("Signin() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != "/verify" {
		t.Fatalf("location = %q", got)
	}

	req := httptest.NewRequest(http.MethodGet, "/verify", nil)
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	state, err := h.sessions.Load(req)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	flow, ok := authsession.GetPendingFlow(state)
	if !ok || flow.Email != "ada@example.com" || flow.Purpose != model.FlowPurposeSignin {
		t.Fatalf("flow=%#v ok=%v", flow, ok)
	}
}

func TestSignupConflictRenders409(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		beginSignupFn: func(context.Context, model.BeginSignupInput, string) (model.PendingFlow, error) {
			return model.PendingFlow{}, model.ErrEmailTaken
		},
	})
	c, rec := newFormContext(http.MethodPost, "/signup", url.Values{
		"email":        {"ada@example.com"},
		"display_name": {"Ada"},
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

func TestVerifyPagePrefillsTokenCode(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		verifyPageStateFn: func(token string) (model.VerifyPageState, error) {
			if token != "signed-token" {
				t.Fatalf("token = %q", token)
			}
			return model.VerifyPageState{
				Purpose: model.FlowPurposeSignup,
				Code:    "123456",
				Token:   token,
				Email:   "ada@example.com",
			}, nil
		},
	})
	c, rec := newRequestContext(http.MethodGet, "/verify?token=signed-token", nil)
	setCSRFToken(c)

	if err := h.VerifyPage(c); err != nil {
		t.Fatalf("VerifyPage() error = %v", err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `value="123456"`) || !strings.Contains(body, "ada@example.com") {
		t.Fatalf("body = %q", body)
	}
}

func TestResendVerifyRedirectsWithFreshPendingFlow(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		resendPendingFn: func(_ context.Context, flow model.PendingFlow, verifyURL string) (model.PendingFlow, error) {
			if flow.Purpose != model.FlowPurposeSignup || verifyURL != "https://example.com/verify" {
				t.Fatalf("flow=%#v verifyURL=%q", flow, verifyURL)
			}
			next := flow
			next.Secret = "NEWSECRET"
			return next, nil
		},
	})

	cookies := withPendingFlowSession(t, h, model.PendingFlow{
		Purpose:     model.FlowPurposeSignup,
		Email:       "ada@example.com",
		DisplayName: "Ada",
		Role:        model.RoleUser,
		Secret:      "OLDSECRET",
		IssuedAt:    time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt:   time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	req := httptest.NewRequest(http.MethodPost, "/verify/resend", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resendRec := httptest.NewRecorder()
	c := echo.New().NewContext(req, resendRec)
	setCSRFToken(c)

	if err := h.ResendVerify(c); err != nil {
		t.Fatalf("ResendVerify() error = %v", err)
	}
	if resendRec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", resendRec.Code)
	}
	if got := resendRec.Header().Get(echo.HeaderLocation); got != "/verify" {
		t.Fatalf("location = %q", got)
	}

	checkReq := httptest.NewRequest(http.MethodGet, "/verify", nil)
	for _, cookie := range resendRec.Result().Cookies() {
		checkReq.AddCookie(cookie)
	}
	state, err := h.sessions.Load(checkReq)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	flow, ok := authsession.GetPendingFlow(state)
	if !ok || flow.Secret != "NEWSECRET" {
		t.Fatalf("flow=%#v ok=%v", flow, ok)
	}
}

func TestVerifyRedirectsOnSuccess(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		verifyPendingFn: func(_ context.Context, flow model.PendingFlow, input model.VerifyInput) (model.VerifyResult, error) {
			if flow.Purpose != model.FlowPurposeSignin || input.Code != "123456" {
				t.Fatalf("flow=%#v input=%#v", flow, input)
			}
			return model.VerifyResult{
				Purpose: model.FlowPurposeSignin,
				User:    handlerTestUser,
			}, nil
		},
	})

	cookies := withPendingFlowSession(t, h, model.PendingFlow{
		Purpose:   model.FlowPurposeSignin,
		Email:     handlerTestUser.Email,
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(url.Values{"code": {"123456"}}.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	verifyRec := httptest.NewRecorder()
	c := echo.New().NewContext(req, verifyRec)
	setCSRFToken(c)

	if err := h.Verify(c); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if verifyRec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", verifyRec.Code)
	}
	if got := verifyRec.Header().Get(echo.HeaderLocation); got != "/" {
		t.Fatalf("location = %q", got)
	}
	if !strings.Contains(verifyRec.Header().Get(echo.HeaderSetCookie), "test-session=") {
		t.Fatalf("set-cookie = %q", verifyRec.Header().Get(echo.HeaderSetCookie))
	}
}

func TestEmailChangeRedirectsToVerify(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		beginEmailChangeFn: func(_ context.Context, userID uuid.UUID, input model.BeginEmailChangeInput, verifyURL string) (model.PendingFlow, error) {
			if userID != handlerTestUser.ID || input.Email != "new@example.com" || verifyURL != "https://example.com/verify" {
				t.Fatalf("userID=%s input=%#v verifyURL=%q", userID, input, verifyURL)
			}
			return model.PendingFlow{
				Purpose:   model.FlowPurposeEmailChange,
				Email:     input.Email,
				UserID:    userID,
				Secret:    "SECRET",
				IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
				ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
			}, nil
		},
	})

	cookies := withAuthSession(t, h, handlerTestUser.ID.String(), handlerTestUser.DisplayName)

	req := httptest.NewRequest(http.MethodPost, "/account/email", strings.NewReader(url.Values{"email": {"new@example.com"}}.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	changeRec := httptest.NewRecorder()
	c := echo.New().NewContext(req, changeRec)
	setCSRFToken(c)

	if err := h.BeginEmailChange(c); err != nil {
		t.Fatalf("BeginEmailChange() error = %v", err)
	}
	if changeRec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", changeRec.Code)
	}
	if got := changeRec.Header().Get(echo.HeaderLocation); got != "/verify" {
		t.Fatalf("location = %q", got)
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

func withPendingFlowSession(t *testing.T, h *Handler, flow model.PendingFlow) []*http.Cookie {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/verify", nil)
	rec := httptest.NewRecorder()
	state, err := h.sessions.Load(req)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := authsession.SetPendingFlow(state, flow); err != nil {
		t.Fatalf("authsession.SetPendingFlow() error = %v", err)
	}
	if err := h.sessions.Commit(rec, req, state); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	return rec.Result().Cookies()
}

func withAuthSession(t *testing.T, h *Handler, userID, displayName string) []*http.Cookie {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	state, err := h.sessions.Load(req)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := authsession.SetAuth(state, userID, displayName); err != nil {
		t.Fatalf("authsession.SetAuth() error = %v", err)
	}
	if err := h.sessions.Commit(rec, req, state); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	return rec.Result().Cookies()
}

func TestVerifyRejectsExpiredCode(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		verifyPendingFn: func(context.Context, model.PendingFlow, model.VerifyInput) (model.VerifyResult, error) {
			return model.VerifyResult{}, model.ErrVerificationExpired
		},
	})

	cookies := withPendingFlowSession(t, h, model.PendingFlow{
		Purpose:   model.FlowPurposeSignin,
		Email:     "ada@example.com",
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(url.Values{"code": {"123456"}}.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	setCSRFToken(c)

	if err := h.Verify(c); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d, want 410", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "expired") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestVerifyRejectsInvalidCode(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		verifyPendingFn: func(context.Context, model.PendingFlow, model.VerifyInput) (model.VerifyResult, error) {
			return model.VerifyResult{}, model.ErrInvalidVerificationCode
		},
	})

	cookies := withPendingFlowSession(t, h, model.PendingFlow{
		Purpose:   model.FlowPurposeSignin,
		Email:     "ada@example.com",
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(url.Values{"code": {"000000"}}.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	setCSRFToken(c)

	if err := h.Verify(c); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestVerifyWithMissingSessionReturnsBadRequest(t *testing.T) {
	h := newTestHandler(fakeAuthService{})
	c, _ := newFormContext(http.MethodPost, "/verify", url.Values{"code": {"123456"}})
	setCSRFToken(c)

	err := h.Verify(c)
	assertAppErrorStatus(t, err, http.StatusBadRequest)
}

func TestEmailChangeRejectsConflictingEmail(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		beginEmailChangeFn: func(context.Context, uuid.UUID, model.BeginEmailChangeInput, string) (model.PendingFlow, error) {
			return model.PendingFlow{}, model.ErrEmailTaken
		},
	})

	cookies := withAuthSession(t, h, handlerTestUser.ID.String(), handlerTestUser.DisplayName)

	req := httptest.NewRequest(http.MethodPost, "/account/email", strings.NewReader(url.Values{"email": {"taken@example.com"}}.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	changeRec := httptest.NewRecorder()
	c := echo.New().NewContext(req, changeRec)
	setCSRFToken(c)

	if err := h.BeginEmailChange(c); err != nil {
		t.Fatalf("BeginEmailChange() error = %v", err)
	}
	if changeRec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", changeRec.Code)
	}
	if !strings.Contains(changeRec.Body.String(), "Email already in use.") {
		t.Fatalf("body = %q", changeRec.Body.String())
	}
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
