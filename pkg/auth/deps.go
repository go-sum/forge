package auth

import (
	"context"
	"net/http"

	"github.com/go-sum/auth/model"
	"github.com/labstack/echo/v5"
	g "maragu.dev/gomponents"
)

// SessionState abstracts session key-value operations.
// *session.State from pkg/session satisfies this implicitly.
type SessionState interface {
	ID() string
	Get(key string, dst any) (bool, error)
	Put(key string, v any) error
	Delete(key string)
}

// SessionMeta holds metadata about a session for binding purposes.
type SessionMeta struct {
	AuthMethod string
	IPAddress  string
	UserAgent  string
}

// SessionManager abstracts the HTTP session lifecycle.
type SessionManager interface {
	Load(r *http.Request) (SessionState, error)
	Commit(w http.ResponseWriter, r *http.Request, s SessionState) error
	Destroy(w http.ResponseWriter, r *http.Request) error
	RotateID(w http.ResponseWriter, r *http.Request, s SessionState) error
	BindSession(ctx context.Context, sessionID, userID string, meta SessionMeta) error
	UnbindSession(ctx context.Context, sessionID, userID string) error
	TouchSession(ctx context.Context, sessionID, userID string) error
}

// FormSubmission represents a validated form submission result.
// *form.Submission from pkg/componentry satisfies this implicitly.
type FormSubmission interface {
	IsValid() bool
	GetFieldErrors(field string) []string
	SetFieldError(field, msg string)
	GetFormErrors() []string
	SetFormError(msg string)
	GetErrors() map[string][]string
}

// FormParser binds and validates a form submission from an Echo context.
type FormParser interface {
	Parse(c *echo.Context, dest any) FormSubmission
}

// Flasher writes one-time flash messages to the HTTP response.
type Flasher interface {
	Success(w http.ResponseWriter, text string) error
	Error(w http.ResponseWriter, text string) error
}

// Redirector issues HTTP redirects with HTMX awareness.
type Redirector interface {
	Redirect(w http.ResponseWriter, r *http.Request, url string) error
}

// PageRenderer renders auth UI pages as gomponents nodes.
// The host application implements this using its component library.
type PageRenderer interface {
	SigninPage(req Request, sub FormSubmission, input model.BeginSigninInput, signinPath, signupPath, csrfField string) g.Node
	SignupPage(req Request, sub FormSubmission, input model.BeginSignupInput, signinPath, signupPath, csrfField string) g.Node
	VerifyPage(req Request, sub FormSubmission, input model.VerifyInput, state model.VerifyPageState, stateErrors []string, verifyPath, resendPath, csrfField string) g.Node
	EmailChangePage(req Request, sub FormSubmission, input model.BeginEmailChangeInput, actionPath, csrfField string) g.Node
}
