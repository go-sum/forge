package config

import cfgs "github.com/go-sum/server/config"

// SessionsConfig groups all session namespace configurations.
type SessionsConfig struct {
	Auth SessionConfig
}

type SessionConfig struct {
	Store      string `validate:"omitempty,oneof=cookie server"`
	Name       string
	AuthKey    string `validate:"required,min=32"`
	EncryptKey string `validate:"required,min=16,max=32"`
	MaxAge     int    // seconds; session cookie TTL
	Secure     bool
	SameSite   string `validate:"omitempty,oneof=strict lax none"`
}

func defaultSession() SessionsConfig {
	return SessionsConfig{
		Auth: SessionConfig{
			Store:      "cookie",
			Name:       "_session",
			AuthKey:    cfgs.ExpandEnv("${AUTH_SESSION_AUTH_KEY}"),
			EncryptKey: cfgs.ExpandEnv("${AUTH_SESSION_ENCRYPT_KEY}"),
			MaxAge:     86400,
			Secure:     true,
			SameSite:   "strict",
		},
	}
}
