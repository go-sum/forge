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
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type authService interface {
	BeginSignin(context.Context, model.BeginSigninInput, string) (model.PendingFlow, error)
	BeginSignup(context.Context, model.BeginSignupInput, string) (model.PendingFlow, error)
	BeginEmailChange(context.Context, uuid.UUID, model.BeginEmailChangeInput, string) (model.PendingFlow, error)
	ResendPendingFlow(context.Context, model.PendingFlow, string) (model.PendingFlow, error)
	VerifyPendingFlow(context.Context, model.PendingFlow, model.VerifyInput) (model.VerifyResult, error)
	VerifyToken(context.Context, string, model.VerifyInput) (model.VerifyResult, error)
	VerifyPageState(string) (model.VerifyPageState, error)
}

// Handler holds auth transport dependencies for the Echo + componentry adapter.
type Handler struct {
	service          authService
	sessions         *session.SessionManager
	validator        authvalidate.Validator
	signinPath       func() string
	signupPath       func() string
	verifyPath       func() string
	verifyResendPath func() string
	verifyURL        func() string
	emailChangePath  func() string
	homePath         func() string
	csrfField        string
	requestFn        func(c *echo.Context) Request
}

// Config parameterises the adapter with application-specific route paths and layout wiring.
type Config struct {
	SigninPath         string
	SignupPath         string
	VerifyPath         string
	VerifyResendPath   string
	VerifyURL          string
	EmailChangePath    string
	HomePath           string
	SigninPathFn       func() string
	SignupPathFn       func() string
	VerifyPathFn       func() string
	VerifyResendPathFn func() string
	VerifyURLFn        func() string
	EmailChangeFn      func() string
	HomePathFn         func() string
	CSRFField          string
	RequestFn          func(c *echo.Context) Request
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
		service:          svc,
		sessions:         sessions,
		validator:        validator,
		signinPath:       resolvePath(cfg.SigninPath, cfg.SigninPathFn),
		signupPath:       resolvePath(cfg.SignupPath, cfg.SignupPathFn),
		verifyPath:       resolvePath(cfg.VerifyPath, cfg.VerifyPathFn),
		verifyResendPath: resolvePath(cfg.VerifyResendPath, cfg.VerifyResendPathFn),
		verifyURL:        resolvePath(cfg.VerifyURL, cfg.VerifyURLFn),
		emailChangePath:  resolvePath(cfg.EmailChangePath, cfg.EmailChangeFn),
		homePath:         resolvePath(cfg.HomePath, cfg.HomePathFn),
		csrfField:        csrfField,
		requestFn:        cfg.RequestFn,
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
	node := SigninPage(req, nil, model.BeginSigninInput{}, h.signinPath(), h.signupPath(), h.csrfField)
	return render.Component(c, node)
}

// Signin starts a passwordless signin flow.
func (h *Handler) Signin(c *echo.Context) error {
	req := h.req(c)
	var input model.BeginSigninInput
	sub := form.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		node := SigninPage(req, sub, input, h.signinPath(), h.signupPath(), h.csrfField)
		return render.ComponentWithStatus(c, http.StatusUnprocessableEntity, node)
	}

	flow, err := h.service.BeginSignin(c.Request().Context(), input, h.verifyURL())
	if err != nil {
		return apperr.Internal(err)
	}
	if err := h.sessions.SetPendingFlow(c.Response(), c.Request(), flow); err != nil {
		return apperr.Internal(err)
	}
	if err := flash.Success(c.Response(), "Check your email for the verification code."); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(h.verifyPath()).Go()
}

// SignupPage renders the signup form.
func (h *Handler) SignupPage(c *echo.Context) error {
	req := h.req(c)
	node := SignupPage(req, nil, model.BeginSignupInput{}, h.signinPath(), h.signupPath(), h.csrfField)
	return render.Component(c, node)
}

// Signup starts a signup verification flow.
func (h *Handler) Signup(c *echo.Context) error {
	req := h.req(c)
	var input model.BeginSignupInput
	sub := form.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		node := SignupPage(req, sub, input, h.signinPath(), h.signupPath(), h.csrfField)
		return render.ComponentWithStatus(c, http.StatusUnprocessableEntity, node)
	}

	flow, err := h.service.BeginSignup(c.Request().Context(), input, h.verifyURL())
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			sub.SetFieldError("Email", "Email already in use.")
			node := SignupPage(req, sub, input, h.signinPath(), h.signupPath(), h.csrfField)
			return render.ComponentWithStatus(c, http.StatusConflict, node)
		}
		return apperr.Internal(err)
	}

	if err := h.sessions.SetPendingFlow(c.Response(), c.Request(), flow); err != nil {
		return apperr.Internal(err)
	}
	if err := flash.Success(c.Response(), "Check your email for the verification code."); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(h.verifyPath()).Go()
}

// VerifyPage renders the shared verification screen.
func (h *Handler) VerifyPage(c *echo.Context) error {
	req := h.req(c)
	input := model.VerifyInput{Token: c.QueryParam("token")}
	state, formErrs := h.verifyStateFromRequest(c)
	node := VerifyPage(req, nil, input, state, formErrs, h.verifyPath(), h.verifyResendPath(), h.csrfField)
	return render.Component(c, node)
}

// Verify completes either a pending browser flow or a token-based flow.
func (h *Handler) Verify(c *echo.Context) error {
	req := h.req(c)
	var input model.VerifyInput
	sub := form.New(h.validator.Validate())
	sub.Submit(c, &input)

	state, stateErrs := h.verifyStateFromPost(c, input.Token)
	if !sub.IsValid() {
		node := VerifyPage(req, sub, input, state, stateErrs, h.verifyPath(), h.verifyResendPath(), h.csrfField)
		return render.ComponentWithStatus(c, http.StatusUnprocessableEntity, node)
	}

	var (
		result model.VerifyResult
		err    error
	)
	if input.Token != "" {
		result, err = h.service.VerifyToken(c.Request().Context(), input.Token, input)
	} else {
		flow, flowErr := h.sessions.GetPendingFlow(c.Request())
		if flowErr != nil {
			return apperr.BadRequest("Your verification session is missing. Start again.")
		}
		result, err = h.service.VerifyPendingFlow(c.Request().Context(), flow, input)
	}
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidVerificationCode):
			sub.SetFormError("The verification code is invalid.")
			node := VerifyPage(req, sub, input, state, stateErrs, h.verifyPath(), h.verifyResendPath(), h.csrfField)
			return render.ComponentWithStatus(c, http.StatusUnauthorized, node)
		case errors.Is(err, model.ErrVerificationExpired):
			sub.SetFormError("The verification code has expired. Start again.")
			node := VerifyPage(req, sub, input, state, stateErrs, h.verifyPath(), h.verifyResendPath(), h.csrfField)
			return render.ComponentWithStatus(c, http.StatusGone, node)
		case errors.Is(err, model.ErrInvalidCredentials):
			sub.SetFormError("The verification session is no longer valid. Start again.")
			node := VerifyPage(req, sub, input, state, stateErrs, h.verifyPath(), h.verifyResendPath(), h.csrfField)
			return render.ComponentWithStatus(c, http.StatusUnauthorized, node)
		case errors.Is(err, model.ErrEmailTaken):
			sub.SetFormError("The target email is already in use.")
			node := VerifyPage(req, sub, input, state, stateErrs, h.verifyPath(), h.verifyResendPath(), h.csrfField)
			return render.ComponentWithStatus(c, http.StatusConflict, node)
		default:
			return apperr.Internal(err)
		}
	}

	if err := h.sessions.SetUserID(c.Response(), c.Request(), result.User.ID.String()); err != nil {
		return apperr.Internal(err)
	}
	if err := h.sessions.ClearPendingFlow(c.Response(), c.Request()); err != nil {
		return apperr.Internal(err)
	}
	if result.Purpose == model.FlowPurposeEmailChange {
		if err := flash.Success(c.Response(), "Your email address has been updated."); err != nil {
			return apperr.Internal(err)
		}
	}
	return redirect.New(c.Response(), c.Request()).To(h.homePath()).Go()
}

// ResendVerify starts a fresh verification cycle from the current pending flow.
func (h *Handler) ResendVerify(c *echo.Context) error {
	flow, err := h.sessions.GetPendingFlow(c.Request())
	if err != nil {
		return apperr.BadRequest("Your verification session is missing. Start again.")
	}

	nextFlow, err := h.service.ResendPendingFlow(c.Request().Context(), flow, h.verifyURL())
	if err != nil {
		if flashErr := flash.Error(c.Response(), "Unable to resend that verification code. Start again."); flashErr != nil {
			return apperr.Internal(flashErr)
		}
		return redirect.New(c.Response(), c.Request()).To(h.startPathForPurpose(flow.Purpose)).Go()
	}
	if err := h.sessions.SetPendingFlow(c.Response(), c.Request(), nextFlow); err != nil {
		return apperr.Internal(err)
	}
	if err := flash.Success(c.Response(), "A new verification code has been sent."); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(h.verifyPath()).Go()
}

// EmailChangePage renders the self-service email-change form.
func (h *Handler) EmailChangePage(c *echo.Context) error {
	req := h.req(c)
	node := EmailChangePage(req, nil, model.BeginEmailChangeInput{}, h.emailChangePath(), h.csrfField)
	return render.Component(c, node)
}

// BeginEmailChange starts an email-change verification flow.
func (h *Handler) BeginEmailChange(c *echo.Context) error {
	req := h.req(c)
	userIDRaw, err := h.sessions.GetUserID(c.Request())
	if err != nil {
		return apperr.Unauthorized("Please sign in again.")
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return apperr.Unauthorized("Please sign in again.")
	}

	var input model.BeginEmailChangeInput
	sub := form.New(h.validator.Validate())
	sub.Submit(c, &input)
	if !sub.IsValid() {
		node := EmailChangePage(req, sub, input, h.emailChangePath(), h.csrfField)
		return render.ComponentWithStatus(c, http.StatusUnprocessableEntity, node)
	}

	flow, err := h.service.BeginEmailChange(c.Request().Context(), userID, input, h.verifyURL())
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			sub.SetFieldError("Email", "Email already in use.")
			node := EmailChangePage(req, sub, input, h.emailChangePath(), h.csrfField)
			return render.ComponentWithStatus(c, http.StatusConflict, node)
		}
		return apperr.Internal(err)
	}

	if err := h.sessions.SetPendingFlow(c.Response(), c.Request(), flow); err != nil {
		return apperr.Internal(err)
	}
	if err := flash.Success(c.Response(), "Check your new email address for the verification code."); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(h.verifyPath()).Go()
}

// Signout clears the session and redirects to the signin page.
func (h *Handler) Signout(c *echo.Context) error {
	if err := h.sessions.Clear(c.Response(), c.Request()); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(h.signinPath()).Go()
}

func (h *Handler) verifyStateFromRequest(c *echo.Context) (model.VerifyPageState, []string) {
	token := c.QueryParam("token")
	if token != "" {
		state, err := h.service.VerifyPageState(token)
		if err != nil {
			return model.VerifyPageState{Token: token}, []string{"The verification link is invalid or expired."}
		}
		return state, nil
	}

	flow, err := h.sessions.GetPendingFlow(c.Request())
	if err != nil {
		return model.VerifyPageState{}, []string{"Start a signup, signin, or email-change flow first."}
	}
	return model.VerifyPageState{
		Purpose:   flow.Purpose,
		Email:     flow.Email,
		CanResend: true,
	}, nil
}

func (h *Handler) verifyStateFromPost(c *echo.Context, token string) (model.VerifyPageState, []string) {
	if token != "" {
		state, err := h.service.VerifyPageState(token)
		if err != nil {
			return model.VerifyPageState{Token: token}, []string{"The verification link is invalid or expired."}
		}
		return state, nil
	}
	flow, err := h.sessions.GetPendingFlow(c.Request())
	if err != nil {
		return model.VerifyPageState{}, []string{"Your verification session is missing. Start again."}
	}
	return model.VerifyPageState{
		Purpose:   flow.Purpose,
		Email:     flow.Email,
		CanResend: true,
	}, nil
}

func (h *Handler) startPathForPurpose(purpose model.FlowPurpose) string {
	switch purpose {
	case model.FlowPurposeSignup:
		return h.signupPath()
	case model.FlowPurposeEmailChange:
		return h.emailChangePath()
	default:
		return h.signinPath()
	}
}
