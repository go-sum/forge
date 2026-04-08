package config

import cfgs "github.com/go-sum/server/config"

// StoreConfig groups all persistence backend configuration.
type StoreConfig struct {
	Database DatabaseConfig
	Queue    QueueConfig
	KV       KVConfig
}

type DatabaseConfig struct {
	URL           string
	AutoMigrate   bool
	MigrationsDir string `validate:"required"`
	MaxConns      int32  // pool_max_conns; 0 → package default (10)
}

// QueueConfig holds the background job queue configuration.
type QueueConfig struct {
	Enabled       bool
	Store         string `validate:"omitempty,oneof=postgres"`
	PollInterval  int    // seconds between queue polls
	ShutdownWait  int    // seconds to wait for in-flight jobs on shutdown
	ReapThreshold int    // seconds; 0 → store default (300s / 5 min)
	Queues        []QueueEntryConfig
}

// QueueEntryConfig defines a single named queue and its behavior.
type QueueEntryConfig struct {
	Name        string `validate:"required"`
	Priority    int
	Workers     int `validate:"min=0"`
	MaxAttempts int `validate:"min=0"`
	Timeout     int // seconds; maximum job execution time
	Backoff     int // seconds; delay between retry attempts
}

// KVConfig holds the key-value store configuration.
type KVConfig struct {
	Enabled bool
	Store   string `validate:"omitempty,oneof=redis"`
	Redis   RedisKVConfig
}

// RedisKVConfig holds Redis/Dragonfly connection parameters.
type RedisKVConfig struct {
	Addr         string `validate:"required_if=Enabled true"`
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  int // seconds
	ReadTimeout  int // seconds
	WriteTimeout int // seconds
}

func defaultStore() StoreConfig {
	return StoreConfig{
		Database: DatabaseConfig{
			URL:           cfgs.ExpandEnv("${DATABASE_URL}"),
			AutoMigrate:   false,
			MigrationsDir: "db/migrations",
			MaxConns:      10,
		},

		Queue: QueueConfig{
			Enabled:       false,
			Store:         "postgres",
			PollInterval:  1,
			ShutdownWait:  30,
			ReapThreshold: 300,
			Queues: []QueueEntryConfig{
				{Name: "email", Priority: 10, Workers: 2, MaxAttempts: 5, Timeout: 30, Backoff: 10},
				{Name: "default", Priority: 20, Workers: 1, MaxAttempts: 3, Timeout: 60, Backoff: 5},
			},
		},

		KV: KVConfig{
			Enabled: false,
			Store:   "redis",
			Redis: RedisKVConfig{
				Addr:         cfgs.ExpandEnv("${KV_HOST:-localhost}:${KV_PORT:-6379}"),
				Password:     cfgs.ExpandEnv("${KV_PASSWORD}"),
				DB:           0,
				PoolSize:     10,
				DialTimeout:  5,
				ReadTimeout:  3,
				WriteTimeout: 3,
			},
		},
	}
}
