package app

import (
	auth "github.com/go-sum/auth"
	authadapter "github.com/go-sum/forge/internal/adapters/auth"
	"github.com/go-sum/forge/internal/features/availability"
	"github.com/go-sum/forge/internal/features/contact"
	"github.com/go-sum/docs"
	"github.com/go-sum/forge/internal/features/examples"
	"github.com/go-sum/forge/internal/features/public"
	"github.com/go-sum/forge/internal/features/sessions"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/send"
	"github.com/go-sum/server/headers"
	smw "github.com/go-sum/server/middleware"
	"github.com/go-sum/server/middleware/etag"
	"github.com/go-sum/server/route"
	sitehandlers "github.com/go-sum/site/handlers"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

// RegisterRoutes binds the application's concrete handlers to their URL paths.
func RegisterRoutes(r *Runtime, availHandler *availability.Handler, authHandler *auth.Handler, adminHandler *auth.AdminHandler, passkeyHandler *auth.PasskeyHandler) error {
	registerStaticRoutes(r)
	resolve := route.NewResolver(func() echo.Routes { return r.Web.Router().Routes() })
	r.Web.Use(auth.LoadSession(&authadapter.SessionManagerAdapter{Mgr: r.Sessions}))

	siteHandler := sitehandlers.New(sitehandlers.Config{
		Origin:  r.Config.Security.ExternalOrigin,
		Robots:  r.Config.Site.Robots,
		Sitemap: r.Config.Site.Sitemap,
	}, resolve.Routes())

	publicHandler := public.NewModule(r.Config, siteHandler)
	docsHandler := docs.NewHandler(r.PublicDir)
	contactHandler := contact.NewHandler(r.Config, contact.NewService(r.Queue, contact.Config{
		SendTo:   r.Config.Service.Send.SendTo,
		SendFrom: send.DefaultRegistry.SendFrom(r.Config.Service.Send.Delivery),
	}), r.Validator)
	examplesHandler := examples.NewModule(r.Config)
	sessionsModule := sessions.NewModule(r.Config, r.Sessions, r.Validator)

	crossOriginGuard := appserver.CrossOriginGuard(r.Config)

	route.Register(r.Web,
		// Public GET — no middleware
		route.GET("/health", "health.show", availHandler.Health),
		route.GET("/", "home.show", publicHandler.Home),
		route.GET("/robots.txt", "robots.show", siteHandler.RobotsTxt),
		route.GET("/sitemap.xml", "sitemap.show", siteHandler.SitemapXML),
		route.GET("/docs", "docs.index", docsHandler.Handle),
		route.GET("/docs/*", "docs.show", docsHandler.Handle),
		route.GET("/contact", "contact.show", contactHandler.Form),

		// Authenticated — server rate limit + auth required
		route.Layout(
			route.Use(r.RateLimiters.Middleware(r.Config, "server"), auth.RequireAuthPath(resolve.Path("signin.get"))),
			route.GET("/_components", "components.list", examplesHandler.Handle),

			// Admin — user management
			route.Group("/admin",
				route.GET("/elevate", "admin.elevate", adminHandler.AdminElevatePage),
				// Admin elevation write — adds cross-origin guard
				route.Layout(
					route.Use(crossOriginGuard),
					route.POST("/elevate", "admin.elevate.post", adminHandler.AdminElevate),
				),
				// Admin user management — adds role load + admin check
				route.Layout(
					route.Use(auth.LoadUserRole(r.AuthStore), auth.RequireAdmin()),
					route.Group("/users",
						route.GET("", "admin.user.list", adminHandler.UserList),
						route.GET("/:id/edit", "admin.user.edit", adminHandler.UserEditForm),
						// Cached fragments — private cache + ETag
						route.Layout(
							route.Use(
								smw.CacheHeaders(headers.NewCacheControl().Private().MustRevalidate().String(), "Cookie"),
								etag.Middleware(),
							),
							route.GET("/:id/row", "admin.user.row", adminHandler.UserRow),
						),
						// Admin writes — adds cross-origin guard
						route.Layout(
							route.Use(crossOriginGuard),
							route.PUT("/:id", "admin.user.update", adminHandler.UserUpdate),
							route.DELETE("/:id", "admin.user.delete", adminHandler.UserDelete),
						),
					),
				),
			),
		),
	)

	route.Register(r.Web,
		// TOTP public GET
		route.GET("/signin", "signin.get", authHandler.SigninPage),
		route.GET("/signup", "signup.get", authHandler.SignupPage),
		route.GET("/verify", "verify.get", authHandler.VerifyPage),

		// TOTP public POST — cross-origin guard + auth rate limit
		route.Layout(
			route.Use(crossOriginGuard, r.RateLimiters.Middleware(r.Config, "auth")),
			route.POST("/signin", "signin.post", authHandler.Signin),
			route.POST("/signup", "signup.post", authHandler.Signup),
			route.POST("/verify", "verify.post", authHandler.Verify),
			route.POST("/verify/resend", "verify.resend.post", authHandler.ResendVerify),
			route.POST("/contact", "contact.submit", contactHandler.Submit),
		),

		// TOTP authenticated profile routes
		route.Layout(
			route.Use(r.RateLimiters.Middleware(r.Config, "server"), auth.RequireAuthPath(resolve.Path("signin.get"))),
			route.Group("/profile",
				route.GET("/email", "profile.email.get", authHandler.EmailChangePage),
				route.GET("/sessions", "profile.session.list", sessionsModule.Handler().List),
				// Profile writes — adds cross-origin guard
				route.Layout(
					route.Use(crossOriginGuard),
					route.POST("/signout", "profile.signout.post", authHandler.Signout),
					route.POST("/email", "profile.email.post", authHandler.BeginEmailChange),
					route.DELETE("/sessions/:id", "profile.session.revoke", sessionsModule.Handler().Revoke),
					route.DELETE("/sessions", "profile.session.revoke.all", sessionsModule.Handler().RevokeAll),
				),
			),
		),
	)

	if passkeyHandler != nil {
		passkeyBodyLimit := echomw.BodyLimit(64 * 1024)
		route.Register(r.Web,
			// Public authenticate ceremony — crossOriginGuard + auth rate limit + body limit
			route.Layout(
				route.Use(crossOriginGuard, r.RateLimiters.Middleware(r.Config, "auth"), passkeyBodyLimit),
				route.Group("/auth/passkeys/authenticate",
					route.POST("/begin", "passkey.authenticate.begin", passkeyHandler.AuthenticateBegin),
					route.POST("/finish", "passkey.authenticate.finish", passkeyHandler.AuthenticateFinish),
				),
			),
			// Authenticated passkey routes — server rate limit + auth required
			route.Layout(
				route.Use(r.RateLimiters.Middleware(r.Config, "server"), auth.RequireAuthPath(resolve.Path("signin.get"))),
				// Registration ceremony — JSON + crossOriginGuard + body limit
				route.Layout(
					route.Use(crossOriginGuard, passkeyBodyLimit),
					route.Group("/auth/passkeys/register",
						route.POST("/begin", "passkey.register.begin", passkeyHandler.RegisterBegin),
						route.POST("/finish", "passkey.register.finish", passkeyHandler.RegisterFinish),
					),
				),
				// Management — HTMX
				route.Group("/account/passkeys",
					route.GET("", "passkey.list", passkeyHandler.ListPasskeys),
					route.GET("/:id/row", "passkey.row", passkeyHandler.GetPasskeyRow),
					route.GET("/:id/rename", "passkey.rename.form", passkeyHandler.GetRenameForm),
					route.Layout(
						route.Use(crossOriginGuard),
						route.POST("/:id/rename", "passkey.rename", passkeyHandler.RenamePasskey),
						route.DELETE("/:id", "passkey.delete", passkeyHandler.DeletePasskey),
					),
				),
			),
		)
	}

	return nil
}

// RegisterStartupRoutes binds a degraded route set used when startup fails
// before the full application can be wired.
func RegisterStartupRoutes(r *Runtime, availHandler *availability.Handler) error {
	registerStaticRoutes(r)
	availability.RegisterStartupRoutes(r.Web, availHandler)
	return nil
}

func registerStaticRoutes(r *Runtime) {
	staticGroup := r.Web.Group(r.PublicPrefix)
	staticGroup.Use(smw.StaticCache(smw.StaticCacheConfig{}))
	staticGroup.Static("", r.PublicDir)
}
