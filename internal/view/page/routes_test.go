package page

import (
	"net/http"
	"sync"
	"testing"

	"github.com/go-sum/server/route"
	"github.com/labstack/echo/v5"
)

var (
	pageRoutesOnce sync.Once
	pageRoutes     echo.Routes
)

func mustPageRoutes(t *testing.T) echo.Routes {
	t.Helper()
	pageRoutesOnce.Do(func() {
		e := echo.New()
		noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }

		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signup", Name: "signup.get", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/contact", Name: "contact.show", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodPost, Path: "/contact", Name: "contact.submit", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/admin/elevate", Name: "admin.elevate", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodPost, Path: "/admin/elevate", Name: "admin.elevate.post", Handler: noOp})

		adminUsers := e.Group("/admin/users")
		route.Add(adminUsers, echo.Route{Method: http.MethodGet, Path: "", Name: "admin.user.list", Handler: noOp})
		route.Add(adminUsers, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "admin.user.edit", Handler: noOp})
		route.Add(adminUsers, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "admin.user.row", Handler: noOp})
		route.Add(adminUsers, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "admin.user.update", Handler: noOp})
		route.Add(adminUsers, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "admin.user.delete", Handler: noOp})

		profile := e.Group("/profile")
		route.Add(profile, echo.Route{Method: http.MethodGet, Path: "/sessions", Name: "profile.session.list", Handler: noOp})
		route.Add(profile, echo.Route{Method: http.MethodDelete, Path: "/sessions/:id", Name: "profile.session.revoke", Handler: noOp})
		route.Add(profile, echo.Route{Method: http.MethodDelete, Path: "/sessions", Name: "profile.session.revoke.all", Handler: noOp})

		pageRoutes = e.Router().Routes()
	})
	return pageRoutes
}
