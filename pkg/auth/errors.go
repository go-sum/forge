package auth

import (
	"fmt"
	"net/http"
)

// HTTPError is returned by auth handlers and middleware when the error
// maps to a specific HTTP status. The host application's error handler
// recognizes it via StatusCode() and PublicMessage() interface methods.
type HTTPError struct {
	status  int
	message string
	cause   error
}

func (e *HTTPError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *HTTPError) Unwrap() error        { return e.cause }
func (e *HTTPError) StatusCode() int      { return e.status }
func (e *HTTPError) PublicMessage() string { return e.message }

func errInternal(cause error) *HTTPError {
	return &HTTPError{status: http.StatusInternalServerError, message: "Something went wrong on our side. Please try again.", cause: cause}
}

func errBadRequest(msg string) *HTTPError {
	return &HTTPError{status: http.StatusBadRequest, message: msg}
}

func errUnauthorized(msg string) *HTTPError {
	return &HTTPError{status: http.StatusUnauthorized, message: msg}
}

func errForbidden(msg string) *HTTPError {
	return &HTTPError{status: http.StatusForbidden, message: msg}
}

func errNotFound(msg string) *HTTPError {
	return &HTTPError{status: http.StatusNotFound, message: msg}
}

func errUnavailable(msg string, cause error) *HTTPError {
	return &HTTPError{status: http.StatusServiceUnavailable, message: msg, cause: cause}
}
