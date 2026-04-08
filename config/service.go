package config

import (
	auth "github.com/go-sum/auth"
	"github.com/go-sum/send"
	cfgs "github.com/go-sum/server/config"
)

// ServiceConfig holds service-level configuration.
type ServiceConfig struct {
	Send SendConfig
	Auth auth.Config
}

// SendConfig holds app-specific email workflow config plus provider delivery config.
type SendConfig struct {
	SendTo   string `validate:"required"`
	Delivery send.Config
}

func defaultService() ServiceConfig {
	emailAPIKey := cfgs.ExpandEnv("${EMAIL_API_KEY}")
	emailSendFrom := cfgs.ExpandEnv("${EMAIL_SEND_FROM}")
	emailSendTo := cfgs.ExpandEnv("${EMAIL_SEND_TO}")
	return ServiceConfig{
		Send: SendConfig{
			SendTo: emailSendTo,
			Delivery: send.Config{
				Selected: "log",
				Providers: send.ProvidersConfig{
					Resend: send.HTTPProviderConfig{
						APIKey:   emailAPIKey,
						SendFrom: emailSendFrom,
					},
					Mailchannels: send.HTTPProviderConfig{
						APIKey:   emailAPIKey,
						SendFrom: emailSendFrom,
					},
				},
			},
		},
		Auth: auth.Config{
			Selected: "email_totp",
			Methods: auth.MethodsConfig{
				EmailTOTP: auth.EmailTOTPMethodConfig{
					Enabled:       true,
					Issuer:        "Forge",
					PeriodSeconds: 300,
				},
			},
		},
	}
}
