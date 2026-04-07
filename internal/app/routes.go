package app

import (
	"net/http"

	auth "github.com/go-sum/auth"
	authadapter "github.com/go-sum/forge/internal/adapters/auth"
	"github.com/go-sum/forge/internal/handler"
	appserver "github.com/go-sum/forge/internal/server"
	smw "github.com/go-sum/server/middleware"
	"github.com/go-sum/server/route"
	sitehandlers "github.com/go-sum/site/handlers"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes binds the application's concrete handlers to their URL paths.
// This is the single source of truth for HTTP route registration.
//
// Route Group Policy:
//
//	public:            No auth required. Read-only pages (global Echo instance).
//	publicPost:        Cross-origin guard + rate limit ("auth"). Public mutations.
//	authGuarded:       Rate limit ("server") + RequireAuthPath. Authenticated read pages.
//	authGuardedPost:   authGuarded + CrossOriginGuard. Authenticated mutations.
//	adminGuarded:      authGuarded + LoadUserRole + RequireAdmin. Admin read pages.
//	adminGuardedPost:  adminGuarded + CrossOriginGuard. Admin mutations.
//
// Future groups (not yet registered):
//
//	apiV1:             Bearer token auth, JSON responses, rate-limited.
//	webhookInbound:    Signature verification, no session, rate-limited.
func RegisterRoutes(c *Container, h *handler.Handler, availabilityH *handler.AvailabilityHandler, authH *auth.Handler) error {
	staticGroup := c.Web.Group(c.PublicPrefix)
	staticGroup.Use(smw.StaticCache(smw.StaticCacheConfig{}))
	staticGroup.Static("", c.PublicDir)

	c.Web.Use(auth.LoadSession(&authadapter.SessionManagerAdapter{Mgr: c.Sessions}))

	siteH := sitehandlers.New(sitehandlers.Config{
		Origin:  c.Config.App.Security.ExternalOrigin,
		Robots:  c.Config.Site.Robots,
		Sitemap: c.Config.Site.Sitemap,
	}, func() echo.Routes { return c.Web.Router().Routes() })

	docsH := docsHandler(c.PublicDir)

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

	registerPublicRoutes(c.Web, authGuarded, h, availabilityH, siteH, docsH)
	registerAuthRoutes(c.Web, publicPost, authGuarded, authGuardedPost, h, authH)
	registerAdminRoutes(adminGuarded, adminGuardedPost, h)

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
