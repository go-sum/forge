package availability

import (
	"net/http"

	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

func RegisterHealth(public *echo.Group, h *Handler) {
	route.Add(public, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: h.Health})
}

func RegisterStartupRoutes(web *echo.Echo, h *Handler) {
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: h.Unavailable})
	route.Add(web, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: h.Health})
	route.Add(web, echo.Route{Method: echo.RouteAny, Path: "/*", Handler: h.Unavailable})
}
