package echocomponentry

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/go-sum/auth/model"
	"github.com/go-sum/auth/session"
	authvalidate "github.com/go-sum/auth/validate"
	pkgflash "github.com/go-sum/componentry/patterns/flash"
	pkgform "github.com/go-sum/componentry/patterns/form"
	"github.com/go-sum/componentry/patterns/redirect"
	renderecho "github.com/go-sum/componentry/render/echo"
	"github.com/go-sum/server/apperr"
)

type authService interface {
	Login(context.Context, model.LoginInput) (model.User, error)
	Register(context.Context, model.CreateUserInput) (model.User, error)
}

// Handler holds auth transport dependencies for the Echo + componentry adapter.
type Handler struct {
	service      authService
	sessions     *session.SessionManager
	validator    authvalidate.Validator
	loginPath    string
	registerPath string
	homePath     string
	csrfField    string
	requestFn    func(c *echo.Context) Request
}

// Config parameterises the adapter with application-specific route paths and layout wiring.
type Config struct {
	LoginPath    string
	RegisterPath string
	HomePath     string
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
		service:      svc,
		sessions:     sessions,
		validator:    validator,
		loginPath:    cfg.LoginPath,
		registerPath: cfg.RegisterPath,
		homePath:     cfg.HomePath,
		csrfField:    csrfField,
		requestFn:    cfg.RequestFn,
	}
}

func (h *Handler) req(c *echo.Context) Request {
	if h.requestFn != nil {
		return h.requestFn(c)
	}
	return Request{}
}

// LoginPage renders the login form.
func (h *Handler) LoginPage(c *echo.Context) error {
	req := h.req(c)
	node := LoginPage(req, nil, model.LoginInput{}, h.loginPath, h.registerPath, h.csrfField)
	return renderecho.Component(c, node)
}

// Login processes a login form submission.
func (h *Handler) Login(c *echo.Context) error {
	req := h.req(c)
	var input model.LoginInput
	sub := pkgform.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		node := LoginPage(req, sub, input, h.loginPath, h.registerPath, h.csrfField)
		return renderecho.ComponentWithStatus(c, http.StatusUnprocessableEntity, node)
	}

	user, err := h.service.Login(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrInvalidCredentials) {
			sub.SetFormError("Invalid email or password.")
			node := LoginPage(req, sub, input, h.loginPath, h.registerPath, h.csrfField)
			return renderecho.ComponentWithStatus(c, http.StatusUnauthorized, node)
		}
		return apperr.Internal(err)
	}

	if err := h.sessions.SetUserID(c.Response(), c.Request(), user.ID.String()); err != nil {
		return apperr.Internal(err)
	}

	return redirect.New(c.Response(), c.Request()).To(h.homePath).Go()
}

// RegisterPage renders the account registration form.
func (h *Handler) RegisterPage(c *echo.Context) error {
	req := h.req(c)
	node := RegisterPage(req, nil, model.CreateUserInput{}, h.loginPath, h.registerPath, h.csrfField)
	return renderecho.Component(c, node)
}

// Register processes a registration form submission.
func (h *Handler) Register(c *echo.Context) error {
	req := h.req(c)
	var input model.CreateUserInput
	sub := pkgform.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		node := RegisterPage(req, sub, input, h.loginPath, h.registerPath, h.csrfField)
		return renderecho.ComponentWithStatus(c, http.StatusUnprocessableEntity, node)
	}

	_, err := h.service.Register(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			sub.SetFieldError("Email", "Email already in use.")
			node := RegisterPage(req, sub, input, h.loginPath, h.registerPath, h.csrfField)
			return renderecho.ComponentWithStatus(c, http.StatusConflict, node)
		}
		return apperr.Internal(err)
	}

	if err := pkgflash.Success(c.Response(), "Account created. Please sign in."); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(h.loginPath).Go()
}

// Logout clears the session and redirects to the login page.
func (h *Handler) Logout(c *echo.Context) error {
	if err := h.sessions.Clear(c.Response(), c.Request()); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(h.loginPath).Go()
}
