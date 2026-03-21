package apperr

import (
	"errors"
	"net/http"

	"starter/internal/model"
)

// Code identifies an application error class independent of HTTP status text.
type Code string

const (
	CodeBadRequest         Code = "bad_request"
	CodeUnauthorized       Code = "unauthorized"
	CodeForbidden          Code = "forbidden"
	CodeNotFound           Code = "not_found"
	CodeConflict           Code = "conflict"
	CodeValidationFailed   Code = "validation_failed"
	CodeServiceUnavailable Code = "service_unavailable"
	CodeInternal           Code = "internal_error"
)

// Error is a transport-facing application error with a safe public message.
type Error struct {
	Status  int
	Code    Code
	Title   string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	if e.Message != "" {
		return e.Message
	}
	return e.Title
}

func (e *Error) Unwrap() error { return e.Cause }

func (e *Error) StatusCode() int { return e.Status }

// PublicMessage returns the safe message that may be shown to end users.
func (e *Error) PublicMessage() string {
	if e == nil {
		return http.StatusText(http.StatusInternalServerError)
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Title != "" {
		return e.Title
	}
	return http.StatusText(e.Status)
}

func New(status int, code Code, title, message string, cause error) *Error {
	if title == "" {
		title = http.StatusText(status)
	}
	return &Error{
		Status:  status,
		Code:    code,
		Title:   title,
		Message: message,
		Cause:   cause,
	}
}

func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, CodeBadRequest, "Bad Request", message, nil)
}

func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, CodeUnauthorized, "Unauthorized", message, nil)
}

func Forbidden(message string) *Error {
	return New(http.StatusForbidden, CodeForbidden, "Forbidden", message, nil)
}

func NotFound(message string) *Error {
	return New(http.StatusNotFound, CodeNotFound, "Not Found", message, nil)
}

func Conflict(message string) *Error {
	return New(http.StatusConflict, CodeConflict, "Conflict", message, nil)
}

func Validation(message string) *Error {
	return New(http.StatusUnprocessableEntity, CodeValidationFailed, "Validation Failed", message, nil)
}

func Unavailable(message string, cause error) *Error {
	return New(http.StatusServiceUnavailable, CodeServiceUnavailable, "Service Unavailable", message, cause)
}

func Internal(cause error) *Error {
	return New(
		http.StatusInternalServerError,
		CodeInternal,
		"Internal Server Error",
		"Something went wrong on our side. Please try again.",
		cause,
	)
}

// From maps known domain errors onto transport-facing application errors.
func From(err error) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, model.ErrUserNotFound):
		return NotFound("The requested user could not be found.")
	case errors.Is(err, model.ErrEmailTaken):
		return Conflict("That email address is already in use.")
	case errors.Is(err, model.ErrInvalidCredentials):
		return Unauthorized("Invalid email or password.")
	case errors.Is(err, model.ErrForbidden):
		return Forbidden("You are not allowed to perform that action.")
	default:
		return nil
	}
}

// Resolve returns a typed application error for err, falling back to Internal.
func Resolve(err error) *Error {
	if appErr := From(err); appErr != nil {
		return appErr
	}
	return Internal(err)
}
