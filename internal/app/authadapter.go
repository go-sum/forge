package app

import (
	"net/http"

	auth "github.com/go-sum/auth"
	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/componentry/patterns/form"
	"github.com/go-sum/componentry/patterns/redirect"
	"github.com/go-sum/session"
	"github.com/labstack/echo/v5"
)

// sessionManagerAdapter wraps session.Manager to satisfy auth.SessionManager.
type sessionManagerAdapter struct {
	mgr session.Manager
}

func (a *sessionManagerAdapter) Load(r *http.Request) (auth.SessionState, error) {
	return a.mgr.Load(r)
}

func (a *sessionManagerAdapter) Commit(w http.ResponseWriter, r *http.Request, s auth.SessionState) error {
	return a.mgr.Commit(w, r, s.(*session.State))
}

func (a *sessionManagerAdapter) Destroy(w http.ResponseWriter, r *http.Request) error {
	return a.mgr.Destroy(w, r)
}

func (a *sessionManagerAdapter) RotateID(w http.ResponseWriter, r *http.Request, s auth.SessionState) error {
	return a.mgr.RotateID(w, r, s.(*session.State))
}

// formParserAdapter wraps form.New(validator).Submit to satisfy auth.FormParser.
type formParserAdapter struct {
	v form.StructValidator
}

func (a *formParserAdapter) Parse(c *echo.Context, dest any) auth.FormSubmission {
	sub := form.New(a.v)
	sub.Submit(c, dest)
	return sub
}

// flashAdapter wraps componentry flash functions to satisfy auth.Flasher.
type flashAdapter struct{}

func (a *flashAdapter) Success(w http.ResponseWriter, text string) error {
	return flash.Success(w, text)
}

func (a *flashAdapter) Error(w http.ResponseWriter, text string) error {
	return flash.Error(w, text)
}

// redirectAdapter wraps componentry redirect to satisfy auth.Redirector.
type redirectAdapter struct{}

func (a *redirectAdapter) Redirect(w http.ResponseWriter, r *http.Request, url string) error {
	return redirect.New(w, r).To(url).Go()
}
