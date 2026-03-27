package app

import (
	"net/http"

	authadapter "github.com/go-sum/auth/adapters/echocomponentry"
	"github.com/go-sum/forge/internal/adapters"
	"github.com/go-sum/forge/internal/handler"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/server/route"
	sitehandlers "github.com/go-sum/site/handlers"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes binds the application's concrete handlers to their URL paths.
// This is the single source of truth for HTTP route registration.
func RegisterRoutes(c *Container, h *handler.Handler, authH *authadapter.Handler) error {
	users := adapters.NewAuthUserReader(c.Repos.User)

	authKeys := authadapter.ContextKeys{
		UserID:      c.Config.App.Keys.UserID,
		UserRole:    c.Config.App.Keys.UserRole,
		DisplayName: c.Config.App.Keys.DisplayName,
	}
	c.Web.Use(authadapter.LoadSession(c.Sessions, authKeys))
	c.Web.Use(authadapter.LoadUserContext(users, authKeys))

	siteH := sitehandlers.New(sitehandlers.Config{
		Origin:  c.Config.App.Security.ExternalOrigin,
		Robots:  c.Config.Site.Robots,
		Sitemap: c.Config.Site.Sitemap,
	}, func() echo.Routes { return c.Web.Router().Routes() })

	publicPost := c.Web.Group("") // (cross-origin-guarded public POST)
	publicPost.Use(
		appserver.CrossOriginGuard(c.Config),
		appserver.RateLimitMiddleware(c.Config, "auth"),
	)

	authGuarded := c.Web.Group("") // (requires session)
	authGuarded.Use(
		authadapter.RequireAuthPath(func() string {
			return route.Reverse(c.Web.Router().Routes(), "signin.get")
		}, authKeys),
	)

	authGuardedPost := authGuarded.Group("") // (session + cross-origin-guarded POST)
	authGuardedPost.Use(appserver.CrossOriginGuard(c.Config))

	adminGuarded := authGuarded.Group("") // (admin + requires session)
	adminGuarded.Use(authadapter.RequireAdmin(authKeys))

	adminGuardedPost := adminGuarded.Group("") // (admin + session + cross-origin-guarded POST)
	adminGuardedPost.Use(appserver.CrossOriginGuard(c.Config))

	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: h.Home})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/contact", Name: "contact.show", Handler: h.ContactForm})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/contact", Name: "contact.submit", Handler: h.ContactSubmit})
	// auth
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: authH.SigninPage})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/signup", Name: "signup.get", Handler: authH.SignupPage})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/signin", Name: "signin.post", Handler: authH.Signin})
	route.Add(publicPost, echo.Route{Method: http.MethodPost, Path: "/signup", Name: "signup.post", Handler: authH.Signup})
	route.Add(authGuardedPost, echo.Route{Method: http.MethodPost, Path: "/signout", Name: "signout.post", Handler: authH.Signout})
	// users (admin only)
	usersGroup := adminGuarded.Group("/users")
	usersPost := adminGuardedPost.Group("/users")
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "", Name: "user.list", Handler: h.UserList})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "user.edit", Handler: h.UserEditForm})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "user.row", Handler: h.UserRow})
	route.Add(usersPost, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "user.update", Handler: h.UserUpdate})
	route.Add(usersPost, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "user.delete", Handler: h.UserDelete})
	// site
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/robots.txt", Name: "robots.show", Handler: siteH.RobotsTxt})
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/sitemap.xml", Name: "sitemap.show", Handler: siteH.SitemapXML})
	// extras
	route.Add(c.Web, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: h.HealthCheck})
	route.Add(authGuarded, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: h.ComponentExamples})

	return nil
}
