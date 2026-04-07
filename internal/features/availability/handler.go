package availability

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/server/apperr"

	"github.com/labstack/echo/v5"
)

// Handler serves the app's health endpoint and degraded startup routes.
type Handler struct {
	checkHealth func(context.Context) error
	message     string
	cause       error
	version     string
}

func NewHandler(checkHealth func(context.Context) error, cause error, version string) *Handler {
	return &Handler{
		checkHealth: checkHealth,
		message:     startupPublicMessage(cause),
		cause:       cause,
		version:     version,
	}
}

func (h *Handler) Health(c *echo.Context) error {
	status, code := "ok", http.StatusOK
	if err := h.checkHealth(c.Request().Context()); err != nil {
		status, code = "error", http.StatusServiceUnavailable
	}
	resp := map[string]string{"status": status}
	if h.version != "" {
		resp["version"] = h.version
	}
	return c.JSON(code, resp)
}

func (h *Handler) Unavailable(*echo.Context) error {
	return apperr.Unavailable(h.message, h.cause)
}

func startupPublicMessage(err error) string {
	if errors.Is(err, model.ErrRequiredRelationsMissing) {
		return "The app is starting, but some services are not ready yet. Setup needs to be completed before proceeding."
	}
	return "Waiting for services to start."
}
