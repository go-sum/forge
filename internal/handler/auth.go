package handler

import (
	"errors"
	"net/http"

	"starter/internal/model"
	"starter/internal/view/page"
	pkgform "starter/pkg/components/patterns/form"
	"starter/pkg/components/patterns/flash"
	"starter/pkg/components/patterns/redirect"
	"starter/pkg/render"

	"github.com/labstack/echo/v5"
)

// LoginPage renders the login form.
func (h *Handler) LoginPage(c *echo.Context) error {
	return render.Component(c, page.LoginPage(page.LoginProps{
		CSRFToken: h.csrfToken(c),
	}))
}

// Login processes a login form submission.
// On success it establishes a session and redirects to /.
func (h *Handler) Login(c *echo.Context) error {
	var input model.LoginInput
	sub := pkgform.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		return render.ComponentWithStatus(c, http.StatusUnprocessableEntity, page.LoginPage(page.LoginProps{
			Form:      sub,
			CSRFToken: h.csrfToken(c),
		}))
	}

	user, _, err := h.services.Auth.Login(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrInvalidCredentials) {
			return render.ComponentWithStatus(c, http.StatusUnauthorized, page.LoginPage(page.LoginProps{
				Form:      sub,
				CSRFToken: h.csrfToken(c),
				ErrorMsg:  "Invalid email or password.",
			}))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := h.sessions.SetUserID(c.Response(), c.Request(), user.ID.String()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "session error")
	}

	return redirect.New(c.Response(), c.Request()).To("/").Go()
}

// RegisterPage renders the account registration form.
func (h *Handler) RegisterPage(c *echo.Context) error {
	return render.Component(c, page.RegisterPage(page.RegisterProps{
		CSRFToken: h.csrfToken(c),
	}))
}

// Register processes a registration form submission.
// On success it sets a flash message and redirects to /login.
func (h *Handler) Register(c *echo.Context) error {
	var input model.CreateUserInput
	sub := pkgform.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		return render.ComponentWithStatus(c, http.StatusUnprocessableEntity, page.RegisterPage(page.RegisterProps{
			Form:      sub,
			CSRFToken: h.csrfToken(c),
		}))
	}

	_, err := h.services.Auth.Register(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			sub.SetFieldError("Email", "Email already in use.")
			return render.ComponentWithStatus(c, http.StatusConflict, page.RegisterPage(page.RegisterProps{
				Form:      sub,
				CSRFToken: h.csrfToken(c),
			}))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	flash.Success(c.Response(), "Account created. Please sign in.") //nolint:errcheck
	return redirect.New(c.Response(), c.Request()).To("/login").Go()
}

// Logout clears the session and redirects to /login.
func (h *Handler) Logout(c *echo.Context) error {
	h.sessions.Clear(c.Response(), c.Request()) //nolint:errcheck
	return redirect.New(c.Response(), c.Request()).To("/login").Go()
}
