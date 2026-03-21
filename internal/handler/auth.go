package handler

import (
	"errors"
	"net/http"

	"starter/internal/apperr"
	"starter/internal/model"
	"starter/internal/routes"
	"starter/internal/view"
	"starter/internal/view/page"
	"starter/pkg/components/patterns/flash"
	pkgform "starter/pkg/components/patterns/form"
	"starter/pkg/components/patterns/redirect"

	"github.com/labstack/echo/v5"
)

// LoginPage renders the login form.
func (h *Handler) LoginPage(c *echo.Context) error {
	req := h.request(c)
	return view.Render(c, req, page.LoginPage(req, nil, model.LoginInput{}), nil)
}

// Login processes a login form submission.
// On success it establishes a session and redirects to /.
func (h *Handler) Login(c *echo.Context) error {
	req := h.request(c)
	var input model.LoginInput
	sub := pkgform.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		return view.RenderWithStatus(c, req, http.StatusUnprocessableEntity, page.LoginPage(req, sub, input), nil)
	}

	user, err := h.services.Auth.Login(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrInvalidCredentials) {
			sub.SetFormError("Invalid email or password.")
			return view.RenderWithStatus(c, req, http.StatusUnauthorized, page.LoginPage(req, sub, input), nil)
		}
		return apperr.Internal(err)
	}

	if err := h.sessions.SetUserID(c.Response(), c.Request(), user.ID.String()); err != nil {
		return apperr.Internal(err)
	}

	return redirect.New(c.Response(), c.Request()).To(routes.Home).Go()
}

// RegisterPage renders the account registration form.
func (h *Handler) RegisterPage(c *echo.Context) error {
	req := h.request(c)
	return view.Render(c, req, page.RegisterPage(req, nil, model.CreateUserInput{}), nil)
}

// Register processes a registration form submission.
// On success it sets a flash message and redirects to /login.
func (h *Handler) Register(c *echo.Context) error {
	req := h.request(c)
	var input model.CreateUserInput
	sub := pkgform.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		return view.RenderWithStatus(c, req, http.StatusUnprocessableEntity, page.RegisterPage(req, sub, input), nil)
	}

	_, err := h.services.Auth.Register(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			sub.SetFieldError("Email", "Email already in use.")
			return view.RenderWithStatus(c, req, http.StatusConflict, page.RegisterPage(req, sub, input), nil)
		}
		return apperr.Internal(err)
	}

	if err := flash.Success(c.Response(), "Account created. Please sign in."); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(routes.Login).Go()
}

// Logout clears the session and redirects to /login.
func (h *Handler) Logout(c *echo.Context) error {
	if err := h.sessions.Clear(c.Response(), c.Request()); err != nil {
		return apperr.Internal(err)
	}
	return redirect.New(c.Response(), c.Request()).To(routes.Login).Go()
}
