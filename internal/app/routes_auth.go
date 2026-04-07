package app

import (
	"net/http"

	auth "github.com/go-sum/auth"
	"github.com/go-sum/forge/internal/handler"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

// registerAuthRoutes registers signin/signup/verify flows, session management
// (signout), authenticated account routes (email change), admin elevation, and
// the contact form submission.
func registerAuthRoutes(
	web *echo.Echo,
	publicPost *echo.Group,
	authGuarded *echo.Group,
	authGuardedPost *echo.Group,
	h *handler.Handler,
	authH *auth.Handler,
) {
	// Auth flow — read pages (public, no session required)
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: authH.SigninPage})
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/signup", Name: "signup.get", Handler: authH.SignupPage})
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/verify", Name: "verify.get", Handler: authH.VerifyPage})

	// Auth flow — mutations (cross-origin-guarded public POST)
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/signin", Name: "signin.post", Handler: authH.Signin})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/signup", Name: "signup.post", Handler: authH.Signup})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/verify", Name: "verify.post", Handler: authH.Verify})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/verify/resend", Name: "verify.resend.post", Handler: authH.ResendVerify})

	// Contact form submission (cross-origin-guarded public POST)
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/contact", Name: "contact.submit", Handler: h.ContactSubmit})

	// Authenticated session management
	route.Add(authGuardedPost, echo.Route{Method: http.MethodPost, Path: "/signout", Name: "signout.post", Handler: authH.Signout})

	// Authenticated account routes
	route.Add(authGuarded, echo.Route{Method: http.MethodGet, Path: "/account/email", Name: "account.email.get", Handler: authH.EmailChangePage})
	route.Add(authGuardedPost, echo.Route{Method: http.MethodPost, Path: "/account/email", Name: "account.email.post", Handler: authH.BeginEmailChange})

	// Admin elevation (only works when no admin exists)
	route.Add(authGuarded, echo.Route{Method: http.MethodGet, Path: "/account/admin", Name: "account.admin", Handler: h.AdminElevateForm})
	route.Add(authGuardedPost, echo.Route{Method: http.MethodPost, Path: "/account/admin", Name: "account.admin.post", Handler: h.AdminElevate})
}
