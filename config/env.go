package config

func (c *Config) IsDevelopment() bool { return c.App.Env == "development" }

func (c *Config) IsProduction() bool { return c.App.Env == "production" }

func (c *Config) IsTest() bool { return c.App.Env == "test" }
