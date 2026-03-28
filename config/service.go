package config

import (
	auth "github.com/go-sum/auth"
	"github.com/go-sum/send"
)

// ServiceConfig holds service-level configuration loaded from config/service.yaml.
type ServiceConfig struct {
	Send SendConfig  `koanf:"send"`
	Auth auth.Config `koanf:"auth"`
}

// SendConfig holds app-specific email workflow config plus provider delivery config.
type SendConfig struct {
	SendTo   string      `koanf:"send_to"`
	Delivery send.Config `koanf:"delivery"`
}
