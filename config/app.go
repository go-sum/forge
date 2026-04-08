package config

import cfgs "github.com/go-sum/server/config"

// AppConfig holds the full application configuration.
type AppConfig struct {
	Env     string `validate:"required,oneof=development production test"`
	Name    string `validate:"required"`
	Assets  AssetsConfig
	Log     LogConfig
	Server  ServerConfig
	Version string // Version is the build-time app version via -ldflags - never from env
}

type AssetsConfig struct {
	PublicDir    string `validate:"required"`
	PublicPrefix string `validate:"required"`
}

type ServerConfig struct {
	Host            string `validate:"required"`
	Port            int    `validate:"required,min=1,max=65535"`
	GracefulTimeout int
	IdleTimeout     int
	ReadTimeout     int
	WriteTimeout    int
	MaxHeaderBytes  int
	TrustProxy      string `validate:"omitempty,oneof=direct xff"`
	TrustedProxies  []string
}

type LogConfig struct {
	Level string `validate:"required,oneof=debug info warn error"`
}

// defaultApp returns the production defaults for AppConfig.
func defaultApp() AppConfig {
	return AppConfig{
		Env:  cfgs.ExpandEnv("${APP_ENV:-production}"),
		Name: "starter",

		Assets: AssetsConfig{
			PublicDir:    cfgs.ExpandEnv("${PUBLIC_DIR:-public}"),
			PublicPrefix: cfgs.ExpandEnv("${PUBLIC_PREFIX:-/public}"),
		},

		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			GracefulTimeout: 10,
			IdleTimeout:     120,
			ReadTimeout:     5,
			WriteTimeout:    10,
			MaxHeaderBytes:  1 << 20, // 1 MB
			TrustProxy:      "xff",
			TrustedProxies: []string{
				"172.16.0.0/12",
				"127.0.0.0/8",
				"::1/128",
			},
		},

		Log: LogConfig{
			Level: "info",
		},
	}
}
