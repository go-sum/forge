package echocomponentry

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-sum/auth/model"
	"github.com/go-sum/auth/session"
	authvalidate "github.com/go-sum/auth/validate"
	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/componentry/patterns/form"
	"github.com/go-sum/componentry/patterns/redirect"
	render "github.com/go-sum/componentry/render/echo"
	"github.com/go-sum/server/apperr"
	"github.com/labstack/echo/v5"
)

type authService interface {
	Signin(context.Context, model.SigninInput) (model.User, error)
	Signup(context.Context, model.SignupInput) (model.User, error)
}

// Handler holds auth transport dependencies for the Echo + componentry adapter.
type Handler struct {
	service    authService
	sessions   *session.SessionManager
	validator  authvalidate.Validator
	signinPath func() string
	signupPath func() string
	homePath   func() string
	csrfField  string
	requestFn  func(c *echo.Context) Request
}

// Config parameterises the adapter with application-specific route paths and layout wiring.
type Config struct {
	SigninPath   string
	SignupPath   string
	HomePath     string
	SigninPathFn func() string
	SignupPathFn func() string
	HomePathFn   func() string
	CSRFField    string
	RequestFn    func(c *echo.Context) Request
}

// New constructs a Handler.
func New(
	svc authService,
	sessions *session.SessionManager,
	validator authvalidate.Validator,
	cfg Config,
) *Handler {
	csrfField := cfg.CSRFField
	if csrfField == "" {
		csrfField = "_csrf"
	}
	return &Handler{
		service:    svc,
		sessions:   sessions,
		validator:  validator,
		signinPath: resolvePath(cfg.SigninPath, cfg.SigninPathFn),
		signupPath: resolvePath(cfg.SignupPath, cfg.SignupPathFn),
		homePath:   resolvePath(cfg.HomePath, cfg.HomePathFn),
		csrfField:  csrfField,
		requestFn:  cfg.RequestFn,
	}
}

func resolvePath(path string, pathFn func() string) func() string {
	if pathFn != nil {
		return pathFn
	}
	return func() string { return path }
}

func (h *Handler) req(c *echo.Context) Request {
	if h.requestFn != nil {
		return h.requestFn(c)
	}
	return Request{}
}

// SigninPage renders the signin form.
func (h *Handler) SigninPage(c *echo.Context) error {
	req := h.req(c)
	node := SigninPage(req, nil, model.SigninInput{}, h.signinPath(), h.signupPath(), h.csrfField)
	return render.Component(c, node)
}

// Signin processes a signin form submission.
func (h *Handler) Signin(c *echo.Context) error {
	req := h.req(c)
	var input model.SigninInput
	sub := form.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		node := SigninPage(req, sub, input, h.signinPath(), h.signupPath(), h.csrfField)
		return render.ComponentWithStatus(c, http.StatusUnprocessableEntity, node)
	}

	user, err := h.service.Signin(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrInvalidCredentials) {
			sub.SetFormError("Invalid email or password.")
			node := SigninPage(req, sub, input, h.signinPath(), h.signupPath(), h.csrfField)
			return render.ComponentWithStatus(c, http.StatusUnauthorized, node)
		}
		return apperr.Internal(err)
	}

	if err := h.sessions.SetUserID(c.Response(), c.Request(), user.ID.String()); err != nil {
		return apperr.Internal(err)
	}

	return redirect.New(c.Response(), c.Request()).To(h.homePath()).Go()
}

// SignupPage renders the account signup form.
func (h *Handler) SignupPage(c *echo.Context) error {
	req := h.req(c)
	node := SignupPage(req, nil, model.SignupInput{}, h.signinPath(), h.signupPath(), h.csrfField)
	return render.Component(c, node)
}

// Signup processes a signup form submission.
func (h *Handler) Signup(c *echo.Context) error {
	req := h.req(c)
	var input model.SignupInput
	sub := form.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		node := SignupPage(req, sub, input, h.signinPath(), h.signupPath(), h.csrfField)
		return render.ComponentWithStatus(c, http.StatusUnprocessableEntity, node)
	}

	_, err := h.service.Signup(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			sub.SetFieldError("Email", "Email already in use.")
			node := SignupPage(req, sub, input, h.signinPath(), h.signupPath(), h.csrfField)
			return render.ComponentWithStatus(c, http.StatusConflict, node)
		}
		return apperr.Internal(err)
	}

	if err := flash.Success(c.Response(), "Account created. Please sign in."); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(h.signinPath()).Go()
}

// Signout clears the session and redirects to the signin page.
func (h *Handler) Signout(c *echo.Context) error {
	if err := h.sessions.Clear(c.Response(), c.Request()); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(h.signinPath()).Go()
}
