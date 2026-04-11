// Package auth defines package-owned configuration for selectable auth methods.
package auth

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
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

// PasskeyService is the interface for WebAuthn passkey operations.
// Begin* methods return ceremony data for the client; Finish* methods
// complete the ceremony using data from the client.
// Session data lifecycle (storage and retrieval) is the caller's responsibility.
type PasskeyService interface {
	BeginRegistration(ctx context.Context, userID uuid.UUID) (model.PasskeyCreationOptions, model.PasskeyCeremony, error)
	FinishRegistration(ctx context.Context, userID uuid.UUID, name string, ceremony model.PasskeyCeremony, r *http.Request) (model.PasskeyCredential, error)
	BeginAuthentication(ctx context.Context) (model.PasskeyRequestOptions, model.PasskeyCeremony, error)
	FinishAuthentication(ctx context.Context, ceremony model.PasskeyCeremony, r *http.Request) (model.VerifyResult, error)
	GetPasskey(ctx context.Context, userID, passkeyID uuid.UUID) (model.PasskeyCredential, error)
	ListPasskeys(ctx context.Context, userID uuid.UUID) ([]model.PasskeyCredential, error)
	DeletePasskey(ctx context.Context, userID, passkeyID uuid.UUID) error
	RenamePasskey(ctx context.Context, userID, passkeyID uuid.UUID, name string) (model.PasskeyCredential, error)
}

// MethodName names an auth method implementation.
type MethodName string

const (
	MethodEmailTOTP MethodName = "email_totp"
	MethodPasskey   MethodName = "passkey"
)

// Config holds auth-method configuration for the host application.
type Config struct {
	Preferred MethodName `validate:"omitempty,oneof=email_totp passkey"`
	Methods   MethodsConfig
}

// MethodsConfig groups method-specific configuration.
type MethodsConfig struct {
	EmailTOTP EmailTOTPMethodConfig
	Passkey   PasskeyMethodConfig
}

// defaultConfig holds the zero-omitted defaults applied by ApplyDefaults.
// Edit here to change package-wide defaults.
var defaultConfig = Config{
	Preferred: MethodEmailTOTP,
	Methods: MethodsConfig{
		EmailTOTP: EmailTOTPMethodConfig{
			PeriodSeconds: 300, // 5-minute TOTP window
		},
		Passkey: PasskeyMethodConfig{
			ResidentKey:           "required",
			UserVerification:      "required",
			RegistrationTimeout:   5 * time.Minute,
			AuthenticationTimeout: 2 * time.Minute,
		},
	},
}

// ApplyDefaults returns cfg with zero-valued fields filled from defaultConfig.
// Call this once at the composition-root boundary before any consumer reads
// the values (handlers, renderers, services).
func ApplyDefaults(cfg Config) Config {
	cfg.Preferred = cmp.Or(cfg.Preferred, defaultConfig.Preferred)
	cfg.Methods.EmailTOTP.PeriodSeconds = cmp.Or(
		cfg.Methods.EmailTOTP.PeriodSeconds,
		defaultConfig.Methods.EmailTOTP.PeriodSeconds,
	)
	cfg.Methods.Passkey.ResidentKey = cmp.Or(
		cfg.Methods.Passkey.ResidentKey,
		defaultConfig.Methods.Passkey.ResidentKey,
	)
	cfg.Methods.Passkey.UserVerification = cmp.Or(
		cfg.Methods.Passkey.UserVerification,
		defaultConfig.Methods.Passkey.UserVerification,
	)
	if cfg.Methods.Passkey.RegistrationTimeout == 0 {
		cfg.Methods.Passkey.RegistrationTimeout = defaultConfig.Methods.Passkey.RegistrationTimeout
	}
	if cfg.Methods.Passkey.AuthenticationTimeout == 0 {
		cfg.Methods.Passkey.AuthenticationTimeout = defaultConfig.Methods.Passkey.AuthenticationTimeout
	}
	return cfg
}

// PreferredMethod resolves the preferred (top-of-page) auth method for rendering.
// Falls back to defaultConfig.Preferred when unset.
func (c Config) PreferredMethod() MethodName {
	return cmp.Or(c.Preferred, defaultConfig.Preferred)
}

// RegisterValidationRules registers cross-field validation rules for Config on v.
func (c Config) RegisterValidationRules(v *validator.Validate) {
	v.RegisterStructValidation(authConfigRules, Config{})
}

func authConfigRules(sl validator.StructLevel) {
	cfg := sl.Current().Interface().(Config)

	// TOTP is the baseline account recovery method and must always be enabled.
	// Passkey is additive; it cannot be the only enabled method.
	if !cfg.Methods.EmailTOTP.Enabled {
		sl.ReportError(cfg.Methods.EmailTOTP, "EmailTOTP", "EmailTOTP", "totp_must_be_enabled", "")
		return
	}

	if cfg.PreferredMethod() == MethodPasskey && !cfg.Methods.Passkey.Enabled {
		sl.ReportError(cfg.Preferred, "Preferred", "Preferred", "preferred_method_disabled", "")
	}
}

// EmailTOTPMethodConfig configures the email-backed TOTP auth method.
type EmailTOTPMethodConfig struct {
	Enabled       bool
	Issuer        string
	PeriodSeconds int
}

// PasskeyMethodConfig configures the WebAuthn passkey auth method.
type PasskeyMethodConfig struct {
	Enabled       bool
	RPDisplayName string   `validate:"required_if=Enabled true"` // Relying Party display name
	RPID          string   `validate:"required_if=Enabled true"` // e.g. "example.com"
	RPOrigins     []string `validate:"required_if=Enabled true"` // e.g. ["https://example.com"]

	// ResidentKey controls whether the authenticator must store a discoverable
	// credential. Valid values: "required", "preferred", "discouraged".
	// Defaults to "preferred" — creates a discoverable passkey when possible but
	// does not reject authenticators that cannot guarantee it.
	// Set to "required" once broad authenticator support is confirmed.
	ResidentKey string `validate:"omitempty,oneof=required preferred discouraged"`

	// UserVerification controls whether the authenticator must verify the user
	// (biometric, PIN, etc.). Valid values: "required", "preferred", "discouraged".
	// Defaults to "preferred" — verifies when available but does not reject
	// authenticators that lack a local verification mechanism.
	UserVerification string `validate:"omitempty,oneof=required preferred discouraged"`

	// RegistrationTimeout is the server-side timeout for registration ceremonies.
	// Zero means no server-side enforcement. Defaults to 5 minutes.
	RegistrationTimeout time.Duration `validate:"min=0"`

	// AuthenticationTimeout is the server-side timeout for authentication ceremonies.
	// Zero means no server-side enforcement. Defaults to 2 minutes.
	AuthenticationTimeout time.Duration `validate:"min=0"`
}

// Validate checks cross-field constraints for PasskeyMethodConfig.
// Each RPOrigin must be a valid HTTPS URL (except localhost/127.0.0.1 which may use HTTP)
// whose hostname matches RPID or is a subdomain of RPID.
func (c PasskeyMethodConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	for _, origin := range c.RPOrigins {
		u, err := url.Parse(origin)
		if err != nil {
			return fmt.Errorf("passkey config: invalid RPOrigin %q: %w", origin, err)
		}
		host := u.Hostname()
		isLocalhost := host == "localhost" || host == "127.0.0.1"
		if u.Scheme == "http" && !isLocalhost {
			return fmt.Errorf("passkey config: RPOrigin %q uses http for non-localhost host", origin)
		}
		if host != c.RPID && !strings.HasSuffix(host, "."+c.RPID) {
			return fmt.Errorf("passkey config: RPOrigin %q hostname %q does not match RPID %q", origin, host, c.RPID)
		}
	}
	return nil
}
