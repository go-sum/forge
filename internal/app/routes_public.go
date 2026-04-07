package app

import (
	"net/http"

	"github.com/go-sum/forge/internal/handler"
	"github.com/go-sum/server/route"
	sitehandlers "github.com/go-sum/site/handlers"

	"github.com/labstack/echo/v5"
)

// registerPublicRoutes registers unauthenticated read-only routes and public
// infrastructure endpoints (health, docs, site files). The component gallery
// requires authentication and is registered on authGuarded.
func registerPublicRoutes(
	web *echo.Echo,
	authGuarded *echo.Group,
	h *handler.Handler,
	availabilityH *handler.AvailabilityHandler,
	siteH *sitehandlers.Handler,
	docsH echo.HandlerFunc,
) {
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: h.Home})
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/contact", Name: "contact.show", Handler: h.ContactForm})
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/docs", Name: "docs.index", Handler: docsH})
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/docs/*", Name: "docs.show", Handler: docsH})
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/robots.txt", Name: "robots.show", Handler: siteH.RobotsTxt})
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/sitemap.xml", Name: "sitemap.show", Handler: siteH.SitemapXML})
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: availabilityH.Health})
	route.Add(authGuarded, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: h.ComponentExamples})
}
