package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/server/apperr"
)

func TestAvailabilityHandlerUnavailableReturnsServiceUnavailable(t *testing.T) {
	cause := errors.New("database connect: dial tcp 127.0.0.1:5432: connect: connection refused")
	h := NewAvailability(func(context.Context) error { return cause }, cause)
	c, _ := newRequestContext(http.MethodGet, "/", nil)

	err := h.Unavailable(c)

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %T, want *apperr.Error", err)
	}
	if appErr.Status != http.StatusServiceUnavailable {
		t.Fatalf("appErr.Status = %d, want %d", appErr.Status, http.StatusServiceUnavailable)
	}
	want := "Waiting for services to start."
	if appErr.PublicMessage() != want {
		t.Fatalf("appErr.PublicMessage() = %q, want %q", appErr.PublicMessage(), want)
	}
	if !errors.Is(appErr, cause) {
		t.Fatalf("errors.Is(appErr, cause) = false")
	}
}

func TestAvailabilityHandlerUnavailableUsesSchemaMessage(t *testing.T) {
	cause := fmt.Errorf("database verify: %w", model.ErrRequiredRelationsMissing)
	h := NewAvailability(func(context.Context) error { return cause }, cause)
	c, _ := newRequestContext(http.MethodGet, "/", nil)

	err := h.Unavailable(c)

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %T, want *apperr.Error", err)
	}
	want := "The app is starting, but some services are not ready yet. Setup needs to be completed before proceeding."
	if appErr.PublicMessage() != want {
		t.Fatalf("appErr.PublicMessage() = %q, want %q", appErr.PublicMessage(), want)
	}
}

func TestAvailabilityHandlerHealthReportsStatus(t *testing.T) {
	tests := []struct {
		name        string
		checkHealth func(context.Context) error
		wantStatus  int
		wantBody    string
	}{
		{
			name:        "healthy",
			checkHealth: func(context.Context) error { return nil },
			wantStatus:  http.StatusOK,
			wantBody:    "{\"status\":\"ok\"}\n",
		},
		{
			name:        "unhealthy",
			checkHealth: func(context.Context) error { return errors.New("db down") },
			wantStatus:  http.StatusServiceUnavailable,
			wantBody:    "{\"status\":\"error\"}\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewAvailability(tc.checkHealth, nil)
			c, rec := newRequestContext(http.MethodGet, "/health", nil)

			if err := h.Health(c); err != nil {
				t.Fatalf("Health() error = %v", err)
			}
			if rec.Code != tc.wantStatus {
				t.Fatalf("rec.Code = %d, want %d", rec.Code, tc.wantStatus)
			}
			if rec.Body.String() != tc.wantBody {
				t.Fatalf("rec.Body.String() = %q, want %q", rec.Body.String(), tc.wantBody)
			}
		})
	}
}
