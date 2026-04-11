package auth

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

// ── PasskeyMethodConfig.Validate() tests (T3-4) ────────────────────────────────

// TestApplyDefaults verifies that ApplyDefaults fills zero-valued fields from
// the package default and does not overwrite caller-supplied values.
func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name  string
		input Config
		check func(t *testing.T, got Config)
	}{
		{
			name:  "empty config gets package defaults",
			input: Config{},
			check: func(t *testing.T, got Config) {
				t.Helper()
				if got.Preferred != MethodEmailTOTP {
					t.Errorf("Preferred = %q, want %q", got.Preferred, MethodEmailTOTP)
				}
				if got.Methods.EmailTOTP.PeriodSeconds != 300 {
					t.Errorf("PeriodSeconds = %d, want 300", got.Methods.EmailTOTP.PeriodSeconds)
				}
			},
		},
		{
			name:  "explicit Preferred survives",
			input: Config{Preferred: MethodPasskey},
			check: func(t *testing.T, got Config) {
				t.Helper()
				if got.Preferred != MethodPasskey {
					t.Errorf("Preferred = %q, want %q", got.Preferred, MethodPasskey)
				}
			},
		},
		{
			name: "explicit PeriodSeconds survives",
			input: Config{
				Methods: MethodsConfig{
					EmailTOTP: EmailTOTPMethodConfig{PeriodSeconds: 60},
				},
			},
			check: func(t *testing.T, got Config) {
				t.Helper()
				if got.Methods.EmailTOTP.PeriodSeconds != 60 {
					t.Errorf("PeriodSeconds = %d, want 60", got.Methods.EmailTOTP.PeriodSeconds)
				}
			},
		},
		{
			name: "Issuer is not touched by ApplyDefaults",
			input: Config{
				Methods: MethodsConfig{
					EmailTOTP: EmailTOTPMethodConfig{Issuer: "MyApp"},
				},
			},
			check: func(t *testing.T, got Config) {
				t.Helper()
				if got.Methods.EmailTOTP.Issuer != "MyApp" {
					t.Errorf("Issuer = %q, want %q", got.Methods.EmailTOTP.Issuer, "MyApp")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyDefaults(tt.input)
			tt.check(t, got)
		})
	}
}

// TestPreferredMethod verifies that PreferredMethod falls back to email_totp.
func TestPreferredMethod(t *testing.T) {
	tests := []struct {
		name      string
		preferred MethodName
		want      MethodName
	}{
		{"empty uses default", "", MethodEmailTOTP},
		{"email_totp explicit", MethodEmailTOTP, MethodEmailTOTP},
		{"passkey explicit", MethodPasskey, MethodPasskey},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Preferred: tt.preferred}
			if got := cfg.PreferredMethod(); got != tt.want {
				t.Errorf("PreferredMethod() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestAuthConfigRules exercises the cross-field validator.
func TestAuthConfigRules(t *testing.T) {
	v := validator.New()

	tests := []struct {
		name      string
		cfg       Config
		wantErr   bool
		wantTag   string
		wantField string
	}{
		{
			name:    "TOTP disabled → totp_must_be_enabled",
			cfg:     Config{},
			wantErr: true, wantTag: "totp_must_be_enabled", wantField: "EmailTOTP",
		},
		{
			name: "passkey only (TOTP disabled) → totp_must_be_enabled",
			cfg: Config{
				Methods: MethodsConfig{
					EmailTOTP: EmailTOTPMethodConfig{Enabled: false},
					Passkey: PasskeyMethodConfig{
						Enabled:       true,
						RPDisplayName: "Test",
						RPID:          "example.com",
						RPOrigins:     []string{"https://example.com"},
					},
				},
			},
			wantErr: true, wantTag: "totp_must_be_enabled", wantField: "EmailTOTP",
		},
		{
			name: "Preferred=passkey but passkey disabled → preferred_method_disabled",
			cfg: Config{
				Preferred: MethodPasskey,
				Methods: MethodsConfig{
					EmailTOTP: EmailTOTPMethodConfig{Enabled: true},
					Passkey:   PasskeyMethodConfig{Enabled: false},
				},
			},
			wantErr: true, wantTag: "preferred_method_disabled", wantField: "Preferred",
		},
		{
			name: "empty Preferred + TOTP enabled → passes (default resolves to email_totp)",
			cfg: Config{
				Methods: MethodsConfig{
					EmailTOTP: EmailTOTPMethodConfig{Enabled: true},
				},
			},
			wantErr: false,
		},
		{
			name: "both enabled, no Preferred → passes",
			cfg: Config{
				Methods: MethodsConfig{
					EmailTOTP: EmailTOTPMethodConfig{Enabled: true},
					Passkey: PasskeyMethodConfig{
						Enabled:       true,
						RPDisplayName: "Test",
						RPID:          "example.com",
						RPOrigins:     []string{"https://example.com"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Preferred=passkey, passkey enabled → passes",
			cfg: Config{
				Preferred: MethodPasskey,
				Methods: MethodsConfig{
					EmailTOTP: EmailTOTPMethodConfig{Enabled: true},
					Passkey: PasskeyMethodConfig{
						Enabled:       true,
						RPDisplayName: "Test",
						RPID:          "example.com",
						RPOrigins:     []string{"https://example.com"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.RegisterValidationRules(v)
			err := v.Struct(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected validation error, got nil")
				}
				for _, fe := range err.(validator.ValidationErrors) {
					if fe.Tag() == tt.wantTag && fe.Field() == tt.wantField {
						return // found the expected error
					}
				}
				t.Errorf("expected error with tag=%q field=%q; got: %v", tt.wantTag, tt.wantField, err)
			} else if err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// TestPasskeyConfig_ValidateRejectsHTTPOriginInProd verifies that a non-localhost HTTP
// origin is rejected by Validate().
func TestPasskeyConfig_ValidateRejectsHTTPOriginInProd(t *testing.T) {
	cfg := PasskeyMethodConfig{
		Enabled:       true,
		RPDisplayName: "Prod App",
		RPID:          "example.com",
		RPOrigins:     []string{"http://example.com"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() returned nil, want error for http:// non-localhost origin")
	}
}

// TestPasskeyConfig_ValidateAllowsHTTPLocalhost verifies that localhost may use HTTP.
func TestPasskeyConfig_ValidateAllowsHTTPLocalhost(t *testing.T) {
	cfg := PasskeyMethodConfig{
		Enabled:       true,
		RPDisplayName: "Dev App",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil (localhost http is allowed)", err)
	}
}

// TestPasskeyConfig_ValidateRejectsOriginHostMismatchRPID verifies that an origin whose
// hostname does not match or is not a subdomain of RPID is rejected.
func TestPasskeyConfig_ValidateRejectsOriginHostMismatchRPID(t *testing.T) {
	cfg := PasskeyMethodConfig{
		Enabled:       true,
		RPDisplayName: "App",
		RPID:          "example.com",
		RPOrigins:     []string{"https://other.com"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() returned nil, want error for origin hostname mismatch with RPID")
	}
}

// TestPasskeyConfig_ValidateAcceptsSubdomainOrigin verifies that a subdomain of RPID is accepted.
func TestPasskeyConfig_ValidateAcceptsSubdomainOrigin(t *testing.T) {
	cfg := PasskeyMethodConfig{
		Enabled:       true,
		RPDisplayName: "App",
		RPID:          "example.com",
		RPOrigins:     []string{"https://app.example.com"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil for subdomain origin", err)
	}
}
