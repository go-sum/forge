package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-sum/auth/model"
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

// ── Fakes ────────────────────────────────────────────────────────────────────

// fakeHandlerSessionManager is a stateful in-memory session manager.
// All Load calls return the same fakeSessionState regardless of the request,
// so tests can pre-populate state without simulating cookie roundtrips.
type fakeHandlerSessionManager struct {
	state      *fakeSessionState
	commitErr  error
	destroyErr error
	rotateErr  error
	loadErr    error

	bindCalled    bool
	bindSessionID string
	bindUserID    string
	bindMeta      SessionMeta
	bindErr       error

	unbindCalled    bool
	unbindSessionID string
	unbindUserID    string
	unbindErr       error
}

func newFakeHandlerSessionManager() *fakeHandlerSessionManager {
	return &fakeHandlerSessionManager{state: newFakeSessionState()}
}

func (m *fakeHandlerSessionManager) Load(r *http.Request) (SessionState, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.state, nil
}

func (m *fakeHandlerSessionManager) BindSession(_ context.Context, sessionID, userID string, meta SessionMeta) error {
	m.bindCalled = true
	m.bindSessionID = sessionID
	m.bindUserID = userID
	m.bindMeta = meta
	return m.bindErr
}

func (m *fakeHandlerSessionManager) UnbindSession(_ context.Context, sessionID, userID string) error {
	m.unbindCalled = true
	m.unbindSessionID = sessionID
	m.unbindUserID = userID
	return m.unbindErr
}

func (m *fakeHandlerSessionManager) Commit(w http.ResponseWriter, r *http.Request, s SessionState) error {
	if m.commitErr != nil {
		return m.commitErr
	}
	http.SetCookie(w, &http.Cookie{Name: "test-session", Value: "x", MaxAge: 3600})
	return nil
}

func (m *fakeHandlerSessionManager) Destroy(w http.ResponseWriter, r *http.Request) error {
	if m.destroyErr != nil {
		return m.destroyErr
	}
	m.state = newFakeSessionState()
	http.SetCookie(w, &http.Cookie{Name: "test-session", Value: "", MaxAge: -1})
	return nil
}

func (m *fakeHandlerSessionManager) RotateID(w http.ResponseWriter, r *http.Request, s SessionState) error {
	if m.rotateErr != nil {
		return m.rotateErr
	}
	http.SetCookie(w, &http.Cookie{Name: "test-session", Value: "rotated", MaxAge: 3600})
	return nil
}

func (m *fakeHandlerSessionManager) TouchSession(_ context.Context, _, _ string) error {
	return nil
}

// fakeFormParser reads `form:` struct tags, populates string fields from
// request form values, and validates using `validate:` tags.
type fakeFormParser struct{}

func (p *fakeFormParser) Parse(c *echo.Context, dest any) FormSubmission {
	_ = c.Request().ParseForm()
	sub := &fakeFormSubmission{valid: true}
	rv := reflect.ValueOf(dest).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		formTag := field.Tag.Get("form")
		if formTag == "" || formTag == "-" {
			continue
		}
		val := c.Request().FormValue(formTag)
		if rv.Field(i).Kind() == reflect.String {
			rv.Field(i).SetString(val)
		}
		validateFormField(field.Name, val, field.Tag.Get("validate"), sub)
	}
	return sub
}

func validateFormField(name, val, tag string, sub *fakeFormSubmission) {
	parts := strings.Split(tag, ",")
	for _, p := range parts {
		if strings.TrimSpace(p) == "omitempty" && val == "" {
			return
		}
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch {
		case part == "required":
			if val == "" {
				sub.SetFieldError(name, "This field is required.")
				return
			}
		case part == "email":
			if _, err := mail.ParseAddress(val); err != nil {
				sub.SetFieldError(name, "Enter a valid email address.")
				return
			}
		case strings.HasPrefix(part, "len="):
			n, _ := strconv.Atoi(strings.TrimPrefix(part, "len="))
			if len(val) != n {
				sub.SetFieldError(name, fmt.Sprintf("Must be exactly %d characters.", n))
				return
			}
		case part == "numeric":
			for _, ch := range val {
				if ch < '0' || ch > '9' {
					sub.SetFieldError(name, "Must contain only digits.")
					return
				}
			}
		}
	}
}

// fakeFormSubmission satisfies FormSubmission.
type fakeFormSubmission struct {
	valid       bool
	fieldErrors map[string][]string
	formErrors  []string
}

func (s *fakeFormSubmission) IsValid() bool { return s.valid && len(s.fieldErrors) == 0 && len(s.formErrors) == 0 }

func (s *fakeFormSubmission) GetFieldErrors(field string) []string {
	if s.fieldErrors == nil {
		return nil
	}
	return s.fieldErrors[field]
}

func (s *fakeFormSubmission) SetFieldError(field, msg string) {
	if s.fieldErrors == nil {
		s.fieldErrors = make(map[string][]string)
	}
	s.fieldErrors[field] = append(s.fieldErrors[field], msg)
}

func (s *fakeFormSubmission) GetFormErrors() []string { return s.formErrors }

func (s *fakeFormSubmission) SetFormError(msg string) {
	s.formErrors = append(s.formErrors, msg)
}

func (s *fakeFormSubmission) GetErrors() map[string][]string {
	errors := make(map[string][]string, len(s.fieldErrors)+1)
	for field, messages := range s.fieldErrors {
		errors[field] = append([]string(nil), messages...)
	}
	if len(s.formErrors) > 0 {
		errors["_"] = append([]string(nil), s.formErrors...)
	}
	return errors
}

// fakeFlasher discards all flash messages (flash is not under test here).
type fakeFlasher struct{}

func (f *fakeFlasher) Success(w http.ResponseWriter, text string) error { return nil }
func (f *fakeFlasher) Error(w http.ResponseWriter, text string) error   { return nil }

// fakeRedirector performs a real 303 redirect so Location/status checks pass.
type fakeRedirector struct{}

func (r *fakeRedirector) Redirect(w http.ResponseWriter, req *http.Request, url string) error {
	http.Redirect(w, req, url, http.StatusSeeOther)
	return nil
}

// fakePageRenderer renders enough HTML for test assertions without importing
// any UI package.
type fakePageRenderer struct{}

func (r *fakePageRenderer) SigninPage(req Request, sub FormSubmission, input model.BeginSigninInput, signinPath, signupPath, csrfField string) g.Node {
	nodes := []g.Node{
		g.Text("Send Code"),
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
	}
	if input.Email != "" {
		nodes = append(nodes, h.Input(h.Name("email"), h.Value(input.Email)))
	}
	if sub != nil {
		for _, e := range sub.GetFieldErrors("Email") {
			nodes = append(nodes, g.Text(e))
		}
		for _, e := range sub.GetFormErrors() {
			nodes = append(nodes, g.Text(e))
		}
	}
	return h.Div(g.Group(nodes))
}

func (r *fakePageRenderer) SignupPage(req Request, sub FormSubmission, input model.BeginSignupInput, signinPath, signupPath, csrfField string) g.Node {
	nodes := []g.Node{
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
	}
	if sub != nil {
		for _, e := range sub.GetFieldErrors("Email") {
			nodes = append(nodes, g.Text(e))
		}
		for _, e := range sub.GetFormErrors() {
			nodes = append(nodes, g.Text(e))
		}
	}
	return h.Div(g.Group(nodes))
}

func (r *fakePageRenderer) VerifyPage(req Request, sub FormSubmission, input model.VerifyInput, state model.VerifyPageState, stateErrors []string, verifyPath, resendPath, csrfField string) g.Node {
	code := input.Code
	if code == "" {
		code = state.Code
	}
	nodes := []g.Node{
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
	}
	if code != "" {
		nodes = append(nodes, h.Input(h.Value(code)))
	}
	if state.Email != "" {
		nodes = append(nodes, g.Text(state.Email))
	}
	for _, e := range stateErrors {
		nodes = append(nodes, g.Text(e))
	}
	if sub != nil {
		for _, e := range sub.GetFormErrors() {
			nodes = append(nodes, g.Text(e))
		}
		for _, e := range sub.GetFieldErrors("Code") {
			nodes = append(nodes, g.Text(e))
		}
	}
	return h.Div(g.Group(nodes))
}

func (r *fakePageRenderer) EmailChangePage(req Request, sub FormSubmission, input model.BeginEmailChangeInput, actionPath, csrfField string) g.Node {
	nodes := []g.Node{
		h.Input(h.Type("hidden"), h.Name(csrfField), h.Value(req.CSRFToken)),
	}
	if sub != nil {
		for _, e := range sub.GetFieldErrors("Email") {
			nodes = append(nodes, g.Text(e))
		}
		for _, e := range sub.GetFormErrors() {
			nodes = append(nodes, g.Text(e))
		}
	}
	return h.Div(g.Group(nodes))
}

// fakeAuthService satisfies authService.
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

// ── Helpers ───────────────────────────────────────────────────────────────────

func newTestHandler(svc Service, mgr *fakeHandlerSessionManager) *Handler {
	if mgr == nil {
		mgr = newFakeHandlerSessionManager()
	}
	return NewHandler(svc, HandlerConfig{
		Sessions:         mgr,
		Forms:            &fakeFormParser{},
		Flash:            &fakeFlasher{},
		Redirect:         &fakeRedirector{},
		Pages:            &fakePageRenderer{},
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
				PageFn:    func(_ string, children ...g.Node) g.Node { return h.Div(children...) },
			}
		},
	})
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

// withPendingFlowSession pre-populates the fake session with a pending flow.
func withPendingFlowSession(t *testing.T, mgr *fakeHandlerSessionManager, flow model.PendingFlow) {
	t.Helper()
	if err := setPendingFlow(mgr.state, flow); err != nil {
		t.Fatalf("setPendingFlow() error = %v", err)
	}
}

// withAuthSession pre-populates the fake session with auth state.
func withAuthSession(t *testing.T, mgr *fakeHandlerSessionManager, userID, displayName string) {
	t.Helper()
	if err := setAuth(mgr.state, userID, displayName); err != nil {
		t.Fatalf("setAuth() error = %v", err)
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestSigninPageRenders(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, nil)
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
	h := newTestHandler(fakeAuthService{}, nil)
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
	mgr := newFakeHandlerSessionManager()
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
	}, mgr)
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

	flow, ok := getPendingFlow(mgr.state)
	if !ok || flow.Email != "ada@example.com" || flow.Purpose != model.FlowPurposeSignin {
		t.Fatalf("flow=%#v ok=%v", flow, ok)
	}
}

func TestSignupConflictRenders409(t *testing.T) {
	h := newTestHandler(fakeAuthService{
		beginSignupFn: func(context.Context, model.BeginSignupInput, string) (model.PendingFlow, error) {
			return model.PendingFlow{}, model.ErrEmailTaken
		},
	}, nil)
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
	}, nil)
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
	mgr := newFakeHandlerSessionManager()
	h := newTestHandler(fakeAuthService{
		resendPendingFn: func(_ context.Context, flow model.PendingFlow, verifyURL string) (model.PendingFlow, error) {
			if flow.Purpose != model.FlowPurposeSignup || verifyURL != "https://example.com/verify" {
				t.Fatalf("flow=%#v verifyURL=%q", flow, verifyURL)
			}
			next := flow
			next.Secret = "NEWSECRET"
			return next, nil
		},
	}, mgr)

	withPendingFlowSession(t, mgr, model.PendingFlow{
		Purpose:     model.FlowPurposeSignup,
		Email:       "ada@example.com",
		DisplayName: "Ada",
		Role:        model.RoleUser,
		Secret:      "OLDSECRET",
		IssuedAt:    time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt:   time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	c, rec := newRequestContext(http.MethodPost, "/verify/resend", nil)
	setCSRFToken(c)

	if err := h.ResendVerify(c); err != nil {
		t.Fatalf("ResendVerify() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != "/verify" {
		t.Fatalf("location = %q", got)
	}

	flow, ok := getPendingFlow(mgr.state)
	if !ok || flow.Secret != "NEWSECRET" {
		t.Fatalf("flow=%#v ok=%v", flow, ok)
	}
}

func TestVerifyRedirectsOnSuccess(t *testing.T) {
	mgr := newFakeHandlerSessionManager()
	h := newTestHandler(fakeAuthService{
		verifyPendingFn: func(_ context.Context, flow model.PendingFlow, input model.VerifyInput) (model.VerifyResult, error) {
			if flow.Purpose != model.FlowPurposeSignin || input.Code != "123456" {
				t.Fatalf("flow=%#v input=%#v", flow, input)
			}
			return model.VerifyResult{
				Purpose: model.FlowPurposeSignin,
				User:    handlerTestUser,
				Method:  "email_totp",
			}, nil
		},
	}, mgr)

	withPendingFlowSession(t, mgr, model.PendingFlow{
		Purpose:   model.FlowPurposeSignin,
		Email:     handlerTestUser.Email,
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	c, rec := newFormContext(http.MethodPost, "/verify", url.Values{"code": {"123456"}})
	setCSRFToken(c)

	if err := h.Verify(c); err != nil {
		t.Fatalf("Verify() error = %v", err)
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

func TestEmailChangeRedirectsToVerify(t *testing.T) {
	mgr := newFakeHandlerSessionManager()
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
	}, mgr)

	withAuthSession(t, mgr, handlerTestUser.ID.String(), handlerTestUser.DisplayName)

	c, rec := newFormContext(http.MethodPost, "/account/email", url.Values{"email": {"new@example.com"}})
	setCSRFToken(c)

	if err := h.BeginEmailChange(c); err != nil {
		t.Fatalf("BeginEmailChange() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != "/verify" {
		t.Fatalf("location = %q", got)
	}
}

func TestSignoutClearsSession(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, nil)
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
	setCookie := rec.Header().Get(echo.HeaderSetCookie)
	if !strings.Contains(setCookie, "Max-Age=0") && !strings.Contains(setCookie, "Max-Age=-1") {
		t.Fatalf("set-cookie = %q", setCookie)
	}
}

func TestVerifyRejectsExpiredCode(t *testing.T) {
	mgr := newFakeHandlerSessionManager()
	h := newTestHandler(fakeAuthService{
		verifyPendingFn: func(context.Context, model.PendingFlow, model.VerifyInput) (model.VerifyResult, error) {
			return model.VerifyResult{}, model.ErrVerificationExpired
		},
	}, mgr)

	withPendingFlowSession(t, mgr, model.PendingFlow{
		Purpose:   model.FlowPurposeSignin,
		Email:     "ada@example.com",
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	c, rec := newFormContext(http.MethodPost, "/verify", url.Values{"code": {"123456"}})
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
	mgr := newFakeHandlerSessionManager()
	h := newTestHandler(fakeAuthService{
		verifyPendingFn: func(context.Context, model.PendingFlow, model.VerifyInput) (model.VerifyResult, error) {
			return model.VerifyResult{}, model.ErrInvalidVerificationCode
		},
	}, mgr)

	withPendingFlowSession(t, mgr, model.PendingFlow{
		Purpose:   model.FlowPurposeSignin,
		Email:     "ada@example.com",
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	c, rec := newFormContext(http.MethodPost, "/verify", url.Values{"code": {"000000"}})
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
	h := newTestHandler(fakeAuthService{}, nil)
	c, _ := newFormContext(http.MethodPost, "/verify", url.Values{"code": {"123456"}})
	setCSRFToken(c)

	err := h.Verify(c)
	assertHTTPErrorStatus(t, err, http.StatusBadRequest)
}

func TestEmailChangeRejectsConflictingEmail(t *testing.T) {
	mgr := newFakeHandlerSessionManager()
	h := newTestHandler(fakeAuthService{
		beginEmailChangeFn: func(context.Context, uuid.UUID, model.BeginEmailChangeInput, string) (model.PendingFlow, error) {
			return model.PendingFlow{}, model.ErrEmailTaken
		},
	}, mgr)

	withAuthSession(t, mgr, handlerTestUser.ID.String(), handlerTestUser.DisplayName)

	c, rec := newFormContext(http.MethodPost, "/account/email", url.Values{"email": {"taken@example.com"}})
	setCSRFToken(c)

	if err := h.BeginEmailChange(c); err != nil {
		t.Fatalf("BeginEmailChange() error = %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Email already in use.") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestVerifyCallsBindSession(t *testing.T) {
	mgr := newFakeHandlerSessionManager()
	handler := newTestHandler(fakeAuthService{
		verifyPendingFn: func(_ context.Context, flow model.PendingFlow, input model.VerifyInput) (model.VerifyResult, error) {
			return model.VerifyResult{
				Purpose: model.FlowPurposeSignin,
				User:    handlerTestUser,
				Method:  "email_totp",
			}, nil
		},
	}, mgr)

	withPendingFlowSession(t, mgr, model.PendingFlow{
		Purpose:   model.FlowPurposeSignin,
		Email:     handlerTestUser.Email,
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	c, _ := newFormContext(http.MethodPost, "/verify", url.Values{"code": {"123456"}})
	setCSRFToken(c)

	if err := handler.Verify(c); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !mgr.bindCalled {
		t.Fatal("BindSession was not called")
	}
	if mgr.bindUserID != handlerTestUser.ID.String() {
		t.Fatalf("bindUserID = %q, want %q", mgr.bindUserID, handlerTestUser.ID.String())
	}
	if mgr.bindMeta.AuthMethod != "email_totp" {
		t.Fatalf("bindMeta.AuthMethod = %q, want %q", mgr.bindMeta.AuthMethod, "email_totp")
	}
}

func TestVerifyBindSessionFailureNonFatal(t *testing.T) {
	mgr := newFakeHandlerSessionManager()
	mgr.bindErr = errors.New("kv down")
	handler := newTestHandler(fakeAuthService{
		verifyPendingFn: func(_ context.Context, flow model.PendingFlow, input model.VerifyInput) (model.VerifyResult, error) {
			return model.VerifyResult{
				Purpose: model.FlowPurposeSignin,
				User:    handlerTestUser,
				Method:  "email_totp",
			}, nil
		},
	}, mgr)

	withPendingFlowSession(t, mgr, model.PendingFlow{
		Purpose:   model.FlowPurposeSignin,
		Email:     handlerTestUser.Email,
		Secret:    "SECRET",
		IssuedAt:  time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 28, 12, 5, 0, 0, time.UTC),
	})

	c, rec := newFormContext(http.MethodPost, "/verify", url.Values{"code": {"123456"}})
	setCSRFToken(c)

	if err := handler.Verify(c); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != "/" {
		t.Fatalf("location = %q, want /", got)
	}
}

func TestSignoutCallsUnbindSession(t *testing.T) {
	mgr := newFakeHandlerSessionManager()
	handler := newTestHandler(fakeAuthService{}, mgr)

	withAuthSession(t, mgr, handlerTestUser.ID.String(), handlerTestUser.DisplayName)

	c, _ := newRequestContext(http.MethodPost, "/signout", nil)

	if err := handler.Signout(c); err != nil {
		t.Fatalf("Signout() error = %v", err)
	}
	if !mgr.unbindCalled {
		t.Fatal("UnbindSession was not called")
	}
	if mgr.unbindUserID != handlerTestUser.ID.String() {
		t.Fatalf("unbindUserID = %q, want %q", mgr.unbindUserID, handlerTestUser.ID.String())
	}
}

func TestSignoutWithFailedLoadSkipsUnbind(t *testing.T) {
	mgr := newFakeHandlerSessionManager()
	mgr.loadErr = errors.New("bad")
	handler := newTestHandler(fakeAuthService{}, mgr)

	c, rec := newRequestContext(http.MethodPost, "/signout", nil)

	if err := handler.Signout(c); err != nil {
		t.Fatalf("Signout() error = %v", err)
	}
	if mgr.unbindCalled {
		t.Fatal("UnbindSession should not have been called when Load fails")
	}
	setCookie := rec.Header().Get(echo.HeaderSetCookie)
	if !strings.Contains(setCookie, "Max-Age=0") && !strings.Contains(setCookie, "Max-Age=-1") {
		t.Fatalf("set-cookie = %q, expected cookie to be cleared", setCookie)
	}
	if got := rec.Header().Get(echo.HeaderLocation); got != "/signin" {
		t.Fatalf("location = %q, want /signin", got)
	}
}
