// Package etag provides a conditional-response middleware for Echo v5.
//
// The middleware buffers the response body from downstream handlers, computes
// a weak ETag over the buffered content, and short-circuits with a 304 Not
// Modified response when the request's If-None-Match header matches. Only GET
// requests are buffered; all other methods pass through unchanged.
//
// Usage — opt-in on a route group:
//
//	fragmentGroup := e.Group("/fragments")
//	fragmentGroup.Use(etag.Middleware())
package etag

import (
	"bytes"
	"net/http"

	"github.com/go-sum/server/cache"
	"github.com/labstack/echo/v5"
)

// Config defines the ETag middleware configuration.
type Config struct {
	// Skipper defines a function to skip the middleware for a given request.
	// When Skipper returns true the request is passed to the next handler
	// unchanged. Defaults to never skip.
	Skipper func(c *echo.Context) bool
}

// DefaultConfig is the default Config used by Middleware().
var DefaultConfig = Config{}

// Middleware returns an Echo middleware with DefaultConfig.
func Middleware() echo.MiddlewareFunc {
	return NewWithConfig(DefaultConfig)
}

// NewWithConfig returns an Echo middleware with the given config.
// Panics if the config is invalid.
func NewWithConfig(cfg Config) echo.MiddlewareFunc {
	mw, err := cfg.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}

// ToMiddleware converts Config to an echo.MiddlewareFunc or returns an error
// for invalid configuration. Satisfies the echo.MiddlewareConfigurator interface.
func (cfg Config) ToMiddleware() (echo.MiddlewareFunc, error) {
	if cfg.Skipper == nil {
		cfg.Skipper = func(*echo.Context) bool { return false }
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}

			// ETag caching only applies to GET requests.
			if c.Request().Method != http.MethodGet {
				return next(c)
			}

			// 1. Swap the real response writer with a buffer so the handler
			//    writes its output into memory rather than to the client.
			//    c.SetResponse restores the original *echo.Response on the way
			//    back; since the handler never wrote to it, its Committed flag
			//    remains false and we can write the real response through it
			//    without any manual reset.
			buf := newResponseBuffer()
			realWriter := c.Response()
			c.SetResponse(buf)

			// 2. Run the handler chain.
			handlerErr := next(c)

			// 3. Restore the real writer unconditionally so Echo's error
			//    handler (if invoked) can write to the actual client.
			c.SetResponse(realWriter)

			// 4. If the handler returned an error without writing a body, let
			//    Echo's global error handler render the response normally.
			if handlerErr != nil {
				return handlerErr
			}

			body := buf.body.Bytes()

			// 5. Non-200 responses (redirects, 204 No Content, etc.) and
			//    empty bodies are flushed as-is without an ETag.
			if buf.status != http.StatusOK || len(body) == 0 {
				for k, vs := range buf.header {
					for _, v := range vs {
						c.Response().Header().Add(k, v)
					}
				}
				if buf.status != 0 {
					c.Response().WriteHeader(buf.status)
					_, err := c.Response().Write(body)
					return err
				}
				return nil
			}

			// 6. Compute a weak ETag over the buffered body.
			etag := cache.WeakETag(body)

			// 7. Copy all headers the handler set, then attach the ETag.
			for k, vs := range buf.header {
				for _, v := range vs {
					c.Response().Header().Add(k, v)
				}
			}
			cache.SetETag(c.Response().Header(), etag)

			// 8. Short-circuit with 304 when the client already has this ETag.
			if cache.CheckIfNoneMatch(c.Request(), etag) {
				c.Response().WriteHeader(http.StatusNotModified)
				return nil
			}

			// 9. Write the full buffered response.
			c.Response().WriteHeader(buf.status)
			_, err := c.Response().Write(body)
			return err
		}
	}, nil
}

// responseBuffer is an http.ResponseWriter that captures header and body
// writes so the middleware can inspect the complete response before deciding
// whether to send a 304 or the full buffered content.
type responseBuffer struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newResponseBuffer() *responseBuffer {
	return &responseBuffer{header: make(http.Header)}
}

func (b *responseBuffer) Header() http.Header { return b.header }

// WriteHeader captures the status code. Only the first call has effect,
// matching the http.ResponseWriter contract.
func (b *responseBuffer) WriteHeader(code int) {
	if b.status == 0 {
		b.status = code
	}
}

// Write appends data to the buffer. If WriteHeader has not been called yet,
// an implicit 200 OK status is recorded, matching net/http semantics.
func (b *responseBuffer) Write(p []byte) (int, error) {
	if b.status == 0 {
		b.status = http.StatusOK
	}
	return b.body.Write(p)
}
