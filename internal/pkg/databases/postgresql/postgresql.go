package postgresql

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Config struct {
	Host      string      `json:"host"`
	Port      string      `json:"port"`
	User      string      `json:"user"`
	DBName    string      `json:"dbName"`
	SSLMode   string      `json:"sslMode"`
	Password  string      `json:"password"`
	EnableTls bool        `json:"enableTls"`
	TLs       *tls.Config `json:"tlsConfig"`
}

const (
	maxConn           = 50
	healthCheckPeriod = 1 * time.Minute
	maxConnIdleTime   = 1 * time.Minute
	maxConnLifeTime   = 3 * time.Minute
	minConns          = 10
	lazyConnect       = false
)

func NewPgxConn(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.DBName,
		cfg.Password,
	)

	if cfg.EnableTls {
		dataSourceName = fmt.Sprintf("%s sslmode=verify-full", dataSourceName)
	}

	poolCfg, err := pgxpool.ParseConfig(dataSourceName)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = maxConn
	poolCfg.HealthCheckPeriod = healthCheckPeriod
	poolCfg.MaxConnIdleTime = maxConnIdleTime
	poolCfg.MaxConnLifetime = maxConnLifeTime
	poolCfg.MinConns = minConns

	if cfg.EnableTls {
		poolCfg.ConnConfig.TLSConfig = cfg.TLs
	}

	connPool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, errors.Wrap(err, "pgx.ConnecConfig")
	}

	if err := connPool.Ping(ctx); err != nil {
		connPool.Close()
		return nil, errors.Wrap(err, "conPool.Ping")
	}

	return connPool, nil
}
