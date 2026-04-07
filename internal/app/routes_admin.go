package app

import (
	"net/http"

	"github.com/go-sum/forge/internal/handler"
	"github.com/go-sum/server/headers"
	smw "github.com/go-sum/server/middleware"
	"github.com/go-sum/server/middleware/etag"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

// registerAdminRoutes registers admin-only user management routes including the
// ETag-cached user.row fragment endpoint.
func registerAdminRoutes(
	adminGuarded *echo.Group,
	adminGuardedPost *echo.Group,
	h *handler.Handler,
) {
	usersGroup := adminGuarded.Group("/users")
	usersPost := adminGuardedPost.Group("/users")

	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "", Name: "user.list", Handler: h.UserList})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "user.edit", Handler: h.UserEditForm})

	// user.row is a read-only HTMX fragment — short-circuit repeat requests with
	// 304 when the rendered output is unchanged.
	cachedFragments := usersGroup.Group("")
	cachedFragments.Use(smw.CacheHeaders(headers.NewCacheControl().Private().MustRevalidate().String(), "Cookie"))
	cachedFragments.Use(etag.Middleware())
	route.Add(cachedFragments, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "user.row", Handler: h.UserRow})

	route.Add(usersPost, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "user.update", Handler: h.UserUpdate})
	route.Add(usersPost, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "user.delete", Handler: h.UserDelete})
}
