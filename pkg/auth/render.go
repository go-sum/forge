package auth

import (
	"net/http"

	"github.com/labstack/echo/v5"
	g "maragu.dev/gomponents"
)

func renderNode(c *echo.Context, status int, node g.Node) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return node.Render(c.Response())
}

func renderOK(c *echo.Context, node g.Node) error {
	return renderNode(c, http.StatusOK, node)
}
