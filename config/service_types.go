package config

// ServiceConfig holds service-level configuration loaded from config/service.yaml.
type ServiceConfig struct {
	Send SendConfig `koanf:"send"`
}

// SendConfig controls which email adapter is active and how it is configured.
type SendConfig struct {
	Adapter  string `koanf:"adapter"`
	SendTo   string `koanf:"send_to"`
	SendFrom string `koanf:"send_from"`
	APIKey   string `koanf:"api_key"`
}
