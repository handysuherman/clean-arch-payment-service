package redis

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host      string      `json:"host"`
	Password  string      `json:"password"`
	DB        int         `json:"db"`
	PoolSize  int         `json:"poolSize"`
	TLsEnable bool        `json:"tlsEnable"`
	TLs       *tls.Config `json:"tlsConfig"`
}

const (
	maxRetries      = 5
	minRetryBackoff = 300 * time.Millisecond
	maxRetryBackoff = 500 * time.Millisecond
	dialTimeout     = 5 * time.Second
	readTimeout     = 5 * time.Second
	writeTimeout    = 3 * time.Second
	minIdleConns    = 20
	poolTimeout     = 6 * time.Second
	idleTimeout     = 12 * time.Second
)

func NewUniversalRedisClient(ctx context.Context, cfg *Config) (redis.UniversalClient, error) {
	opts := &redis.UniversalOptions{
		Addrs:           []string{cfg.Host},
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      maxRetries,
		MinRetryBackoff: minRetryBackoff,
		MaxRetryBackoff: maxRetryBackoff,
		DialTimeout:     dialTimeout,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    minIdleConns,
		PoolTimeout:     poolTimeout,
	}

	if cfg.TLsEnable {
		opts.TLSConfig = cfg.TLs
	}

	client := redis.NewUniversalClient(opts)

	_, err := client.Ping(ctx).Result()
	if err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
