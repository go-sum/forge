package authadapter

import (
	"net/http"

	auth "github.com/go-sum/auth"
	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/componentry/patterns/form"
	"github.com/go-sum/componentry/patterns/redirect"
	"github.com/go-sum/session"
	"github.com/labstack/echo/v5"
)

// SessionManagerAdapter wraps session.Manager to satisfy auth.SessionManager.
type SessionManagerAdapter struct {
	Mgr session.Manager
}

func (a *SessionManagerAdapter) Load(r *http.Request) (auth.SessionState, error) {
	return a.Mgr.Load(r)
}

func (a *SessionManagerAdapter) Commit(w http.ResponseWriter, r *http.Request, s auth.SessionState) error {
	return a.Mgr.Commit(w, r, s.(*session.State))
}

func (a *SessionManagerAdapter) Destroy(w http.ResponseWriter, r *http.Request) error {
	return a.Mgr.Destroy(w, r)
}

func (a *SessionManagerAdapter) RotateID(w http.ResponseWriter, r *http.Request, s auth.SessionState) error {
	return a.Mgr.RotateID(w, r, s.(*session.State))
}

// FormParserAdapter wraps form.New(validator).Submit to satisfy auth.FormParser.
type FormParserAdapter struct {
	V form.StructValidator
}

func (a *FormParserAdapter) Parse(c *echo.Context, dest any) auth.FormSubmission {
	sub := form.New(a.V)
	sub.Submit(c, dest)
	return sub
}

// FlashAdapter wraps componentry flash functions to satisfy auth.Flasher.
type FlashAdapter struct{}

func (a *FlashAdapter) Success(w http.ResponseWriter, text string) error {
	return flash.Success(w, text)
}

func (a *FlashAdapter) Error(w http.ResponseWriter, text string) error {
	return flash.Error(w, text)
}

// RedirectAdapter wraps componentry redirect to satisfy auth.Redirector.
type RedirectAdapter struct{}

func (a *RedirectAdapter) Redirect(w http.ResponseWriter, r *http.Request, url string) error {
	return redirect.New(w, r).To(url).Go()
}
