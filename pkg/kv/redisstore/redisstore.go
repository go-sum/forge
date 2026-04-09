// Package redisstore implements kv.Store and kv.Scanner using go-redis.
// It is compatible with Redis, Dragonfly, KeyDB, and any Redis-protocol server.
package redisstore

import (
	"cmp"
	"context"
	"errors"
	"time"

	"github.com/go-sum/kv"
	"github.com/redis/go-redis/v9"
)

// Config holds Redis/Dragonfly connection parameters.
type Config struct {
	Addr         string        // host:port (required)
	Password     string        // empty for no auth
	DB           int           // database number (default 0)
	PoolSize     int           // maximum connections (0 = default 10)
	MinIdleConns int           // minimum idle connections
	DialTimeout  time.Duration // 0 = default 5s
	ReadTimeout  time.Duration // 0 = default 3s
	WriteTimeout time.Duration // 0 = default 3s
}

// defaultConfig holds the zero-omitted defaults applied by New.
// Edit here to change package-wide defaults.
var defaultConfig = Config{
	PoolSize:     10,
	DialTimeout:  5 * time.Second,
	ReadTimeout:  3 * time.Second,
	WriteTimeout: 3 * time.Second,
}

// RedisStore implements kv.Store and kv.Scanner backed by a Redis-protocol server.
type RedisStore struct {
	client *redis.Client
}

// Compile-time interface checks.
var (
	_ kv.Store   = (*RedisStore)(nil)
	_ kv.Scanner = (*RedisStore)(nil)
)

// New creates a RedisStore connected to the given address.
func New(cfg Config) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cmp.Or(cfg.PoolSize, defaultConfig.PoolSize),
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cmp.Or(cfg.DialTimeout, defaultConfig.DialTimeout),
		ReadTimeout:  cmp.Or(cfg.ReadTimeout, defaultConfig.ReadTimeout),
		WriteTimeout: cmp.Or(cfg.WriteTimeout, defaultConfig.WriteTimeout),
	})

	return &RedisStore{client: client}, nil
}

// Ping verifies the store is reachable.
func (s *RedisStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Get retrieves the value for key. Returns kv.ErrNotFound if the key does not exist.
func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, kv.ErrNotFound
		}
		return nil, err
	}
	return val, nil
}

// Set stores value under key with optional TTL.
func (s *RedisStore) Set(ctx context.Context, key string, value []byte, opts kv.SetOptions) error {
	return s.client.Set(ctx, key, value, opts.TTL).Err()
}

// Delete removes one or more keys. Non-existent keys are ignored.
func (s *RedisStore) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}

// Exists returns the count of keys that exist.
func (s *RedisStore) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	return s.client.Exists(ctx, keys...).Result()
}

// Close releases the underlying connection pool.
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// Scan iterates over keys matching the given pattern, calling fn for each key.
func (s *RedisStore) Scan(ctx context.Context, pattern string, fn func(key string) error) error {
	var cursor uint64
	for {
		keys, next, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			if err := fn(key); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
