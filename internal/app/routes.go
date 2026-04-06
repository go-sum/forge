package app

import (
	"net/http"

	auth "github.com/go-sum/auth"
	"github.com/go-sum/forge/internal/handler"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/server/headers"
	smw "github.com/go-sum/server/middleware"
	"github.com/go-sum/server/middleware/etag"
	"github.com/go-sum/server/route"
	sitehandlers "github.com/go-sum/site/handlers"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes binds the application's concrete handlers to their URL paths.
// This is the single source of truth for HTTP route registration.
func RegisterRoutes(c *Container, h *handler.Handler, availabilityH *handler.AvailabilityHandler, authH *auth.Handler) error {
	staticGroup := c.Web.Group(c.PublicPrefix)
	staticGroup.Use(smw.StaticCache(smw.StaticCacheConfig{}))
	staticGroup.Static("", c.PublicDir)

	c.Web.Use(auth.LoadSession(&sessionManagerAdapter{mgr: c.Sessions}))

	siteH := sitehandlers.New(sitehandlers.Config{
		Origin:  c.Config.App.Security.ExternalOrigin,
		Robots:  c.Config.Site.Robots,
		Sitemap: c.Config.Site.Sitemap,
	}, func() echo.Routes { return c.Web.Router().Routes() })

	publicPost := c.Web.Group("") // (cross-origin-guarded public POST)
	publicPost.Use(
		appserver.CrossOriginGuard(c.Config),
		c.RateLimiters.Middleware(c.Config, "auth"),
	)

	authGuarded := c.Web.Group("") // (requires session)
	authGuarded.Use(
		c.RateLimiters.Middleware(c.Config, "server"),
		auth.RequireAuthPath(func() string {
			return route.Reverse(c.Web.Router().Routes(), "signin.get")
		}),
	)

	authGuardedPost := authGuarded.Group("") // (session + cross-origin-guarded POST)
	authGuardedPost.Use(appserver.CrossOriginGuard(c.Config))

	adminGuarded := authGuarded.Group("") // (admin + requires session)
	adminGuarded.Use(
		auth.LoadUserRole(c.AuthStore),
		auth.RequireAdmin(),
	)

	adminGuardedPost := adminGuarded.Group("") // (admin + session + cross-origin-guarded POST)
	adminGuardedPost.Use(appserver.CrossOriginGuard(c.Config))

	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: h.Home})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/contact", Name: "contact.show", Handler: h.ContactForm})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/contact", Name: "contact.submit", Handler: h.ContactSubmit})
	// auth
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: authH.SigninPage})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/signup", Name: "signup.get", Handler: authH.SignupPage})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/verify", Name: "verify.get", Handler: authH.VerifyPage})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/signin", Name: "signin.post", Handler: authH.Signin})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/signup", Name: "signup.post", Handler: authH.Signup})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/verify", Name: "verify.post", Handler: authH.Verify})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/verify/resend", Name: "verify.resend.post", Handler: authH.ResendVerify})
	route.Add(authGuardedPost, echo.Route{Method: http.MethodPost, Path: "/signout", Name: "signout.post", Handler: authH.Signout})
	route.Add(authGuarded, echo.Route{Method: http.MethodGet, Path: "/account/email", Name: "account.email.get", Handler: authH.EmailChangePage})
	route.Add(authGuardedPost, echo.Route{Method: http.MethodPost, Path: "/account/email", Name: "account.email.post", Handler: authH.BeginEmailChange})
	// admin elevation (only works when no admin exists)
	route.Add(authGuarded, echo.Route{Method: http.MethodGet, Path: "/account/admin", Name: "account.admin", Handler: h.AdminElevateForm})
	route.Add(authGuardedPost, echo.Route{Method: http.MethodPost, Path: "/account/admin", Name: "account.admin.post", Handler: h.AdminElevate})
	// users (admin only)
	usersGroup := adminGuarded.Group("/users")
	usersPost := adminGuardedPost.Group("/users")
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "", Name: "user.list", Handler: h.UserList})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "user.edit", Handler: h.UserEditForm})
	// short-circuit repeat requests with 304 when the rendered output is unchanged.
	cachedFragments := usersGroup.Group("")
	cachedFragments.Use(smw.CacheHeaders(headers.NewCacheControl().Private().MustRevalidate().String(), "Cookie"))
	cachedFragments.Use(etag.Middleware())
	// user.row is a read-only HTMX fragment — wrap it in ETag middleware
	route.Add(cachedFragments, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "user.row", Handler: h.UserRow})
	route.Add(usersPost, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "user.update", Handler: h.UserUpdate})
	route.Add(usersPost, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "user.delete", Handler: h.UserDelete})
	// site
	docsH := docsHandler(c.PublicDir)
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/docs", Name: "docs.index", Handler: docsH})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/docs/*", Name: "docs.show", Handler: docsH})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/robots.txt", Name: "robots.show", Handler: siteH.RobotsTxt})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/sitemap.xml", Name: "sitemap.show", Handler: siteH.SitemapXML})
	// extras
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: availabilityH.Health})
	route.Add(authGuarded, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: h.ComponentExamples})

	return nil
}

// RegisterStartupRoutes binds a degraded route set used when startup fails
// before the full application can be wired.
func RegisterStartupRoutes(c *Container, availabilityH *handler.AvailabilityHandler) error {
	staticGroup := c.Web.Group(c.PublicPrefix)
	staticGroup.Use(smw.StaticCache(smw.StaticCacheConfig{}))
	staticGroup.Static("", c.PublicDir)

	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: availabilityH.Unavailable})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: availabilityH.Health})
	route.Add(c.Web, echo.Route{Method: echo.RouteAny, Path: "/*", Handler: availabilityH.Unavailable})
	return nil
}
