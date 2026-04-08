// Package auth defines package-owned configuration for selectable auth methods.
package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
)

// Service is the provider-agnostic interface for auth operations.
// *service.AuthService satisfies this interface; future providers must as well.
type Service interface {
	BeginSignin(context.Context, model.BeginSigninInput, string) (model.PendingFlow, error)
	BeginSignup(context.Context, model.BeginSignupInput, string) (model.PendingFlow, error)
	BeginEmailChange(context.Context, uuid.UUID, model.BeginEmailChangeInput, string) (model.PendingFlow, error)
	ResendPendingFlow(context.Context, model.PendingFlow, string) (model.PendingFlow, error)
	VerifyPendingFlow(context.Context, model.PendingFlow, model.VerifyInput) (model.VerifyResult, error)
	VerifyToken(context.Context, string, model.VerifyInput) (model.VerifyResult, error)
	VerifyPageState(string) (model.VerifyPageState, error)
}

// ValidateConfig checks that cfg selects a supported, enabled method.
// Returns an error suitable for wrapping in a startup panic.
func ValidateConfig(cfg Config) error {
	if cfg.SelectedMethod() != MethodEmailTOTP {
		return fmt.Errorf("unsupported method %q", cfg.SelectedMethod())
	}
	if !cfg.Methods.EmailTOTP.Enabled {
		return errors.New("email_totp method must be enabled")
	}
	return nil
}

// MethodName names an auth method implementation.
type MethodName string

const (
	MethodEmailTOTP MethodName = "email_totp"
)

// Config holds auth-method configuration for the host application.
type Config struct {
	Selected MethodName    `validate:"omitempty,oneof=email_totp"`
	Methods  MethodsConfig
}

// MethodsConfig groups method-specific configuration.
type MethodsConfig struct {
	EmailTOTP EmailTOTPMethodConfig
}

// SelectedMethod resolves the configured auth method, defaulting to email TOTP.
func (c Config) SelectedMethod() MethodName {
	if c.Selected == "" {
		return MethodEmailTOTP
	}
	return c.Selected
}

// EmailTOTPMethodConfig configures the email-backed TOTP auth method.
type EmailTOTPMethodConfig struct {
	Enabled       bool
	Issuer        string
	PeriodSeconds int
}
