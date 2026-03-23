package userpartial

import (
	"net/http"
	"sync"
	"testing"

	"github.com/go-sum/server/route"
	"github.com/labstack/echo/v5"
)

var (
	partialRoutesOnce sync.Once
	partialRoutes     echo.Routes
)

func mustPartialRoutes(t *testing.T) echo.Routes {
	t.Helper()
	partialRoutesOnce.Do(func() {
		e := echo.New()
		noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }

		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signup", Name: "signup.get", Handler: noOp})

		users := e.Group("/users")
		route.Add(users, echo.Route{Method: http.MethodGet, Path: "", Name: "user.list", Handler: noOp})
		route.Add(users, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "user.edit", Handler: noOp})
		route.Add(users, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "user.row", Handler: noOp})
		route.Add(users, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "user.update", Handler: noOp})
		route.Add(users, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "user.delete", Handler: noOp})

		partialRoutes = e.Router().Routes()
	})
	return partialRoutes
}
