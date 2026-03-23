package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-sum/componentry/patterns/flash"
	componenthtmx "github.com/go-sum/componentry/patterns/htmx"
	render "github.com/go-sum/componentry/render/echo"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/errorpage"
	"github.com/go-sum/server/apperr"

	"github.com/labstack/echo/v5"
)

const problemContentType = "application/problem+json"

type ErrorHandlerConfig struct {
	Debug     bool
	Logger    *slog.Logger
	NavConfig config.NavConfig
}

type problemDetails struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NewErrorHandler returns the application's HTTP error handler.
// HTML requests receive a rendered page, HTMX requests receive out-of-band toasts,
// and API-style requests receive RFC 7807 problem details JSON.
func NewErrorHandler(cfg ErrorHandlerConfig) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		if err == nil {
			return
		}
		if resp, _ := echo.UnwrapResponse(c.Response()); resp != nil && resp.Committed {
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		appErr := classify(err)
		logError(cfg.Logger, c, appErr, err)

		switch {
		case wantsProblemJSON(c.Request()):
			writeProblem(c, appErr, err, cfg.Debug)
		case componenthtmx.NewRequest(c.Request()).IsPartial():
			writeHTMXToast(c, appErr)
		default:
			writeErrorPage(c, appErr, err, cfg.Debug, cfg.NavConfig)
		}
	}
}

func classify(err error) *apperr.Error {
	if err == nil {
		return nil
	}

	appErr := apperr.From(err)
	if appErr != nil {
		return appErr
	}

	var sc interface{ StatusCode() int }
	if errors.As(err, &sc) {
		status := sc.StatusCode()
		title := http.StatusText(status)
		msg := title
		if status >= http.StatusInternalServerError {
			msg = "Something went wrong on our side. Please try again."
		}
		return apperr.New(status, codeForStatus(status), title, msg, err)
	}

	return apperr.Internal(err)
}

func wantsProblemJSON(r *http.Request) bool {
	accept := r.Header.Get(echo.HeaderAccept)
	if accept == "" {
		return false
	}
	for _, part := range strings.Split(accept, ",") {
		mediaType := strings.TrimSpace(strings.Split(part, ";")[0])
		if mediaType == echo.MIMEApplicationJSON || mediaType == problemContentType {
			return true
		}
	}
	return false
}

func codeForStatus(status int) apperr.Code {
	switch status {
	case http.StatusBadRequest:
		return apperr.CodeBadRequest
	case http.StatusUnauthorized:
		return apperr.CodeUnauthorized
	case http.StatusForbidden:
		return apperr.CodeForbidden
	case http.StatusNotFound:
		return apperr.CodeNotFound
	case http.StatusConflict:
		return apperr.CodeConflict
	case http.StatusUnprocessableEntity:
		return apperr.CodeValidationFailed
	case http.StatusServiceUnavailable:
		return apperr.CodeServiceUnavailable
	default:
		return apperr.CodeInternal
	}
}

func logError(logger *slog.Logger, c *echo.Context, appErr *apperr.Error, err error) {
	if logger == nil || appErr == nil {
		return
	}

	attrs := []any{
		"status", appErr.Status,
		"code", appErr.Code,
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
	}
	if reqID := requestID(c); reqID != "" {
		attrs = append(attrs, "request_id", reqID)
	}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}

	switch {
	case appErr.Status >= http.StatusInternalServerError:
		logger.ErrorContext(c.Request().Context(), appErr.Title, attrs...)
	case appErr.Status >= http.StatusBadRequest:
		logger.WarnContext(c.Request().Context(), appErr.Title, attrs...)
	}
}

func writeProblem(c *echo.Context, appErr *apperr.Error, err error, debug bool) {
	pd := problemDetails{
		Type:      "urn:starter:problem:" + string(appErr.Code),
		Title:     appErr.Title,
		Status:    appErr.Status,
		Detail:    appErr.PublicMessage(),
		Instance:  c.Request().URL.Path,
		Code:      string(appErr.Code),
		RequestID: requestID(c),
	}
	if debug && err != nil {
		pd.Error = err.Error()
	}

	writeJSON(c, appErr.Status, problemContentType, pd)
}

func writeHTMXToast(c *echo.Context, appErr *apperr.Error) {
	msg := flash.Message{
		Type: flash.TypeError,
		Text: appErr.PublicMessage(),
	}
	if appErr.Status < http.StatusBadRequest {
		msg.Type = flash.TypeInfo
	}
	_ = render.FragmentWithStatus(c, appErr.Status, flash.RenderOOB([]flash.Message{msg}))
}

func writeErrorPage(c *echo.Context, appErr *apperr.Error, err error, debug bool, navConfig config.NavConfig) {
	technicalDetail := ""
	if debug && err != nil {
		technicalDetail = err.Error()
	}

	req := view.NewRequest(c, navConfig)
	_ = render.ComponentWithStatus(c, appErr.Status, errorpage.Page(req, errorpage.Props{
		Status:          appErr.Status,
		Title:           appErr.Title,
		Message:         appErr.PublicMessage(),
		RequestID:       requestID(c),
		Debug:           debug,
		TechnicalDetail: technicalDetail,
		HomePath:        "/",
	}))
}

func writeJSON(c *echo.Context, status int, contentType string, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		c.NoContent(http.StatusInternalServerError)
		return
	}

	resp := c.Response()
	resp.Header().Set(echo.HeaderContentType, contentType+"; charset=UTF-8")
	resp.WriteHeader(status)
	_, _ = resp.Write(data)
}

func requestID(c *echo.Context) string {
	return c.Response().Header().Get(echo.HeaderXRequestID)
}
