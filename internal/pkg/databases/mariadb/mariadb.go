package MariaDB

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	// _ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Driver    string      `json:"driver"`
	Source    string      `json:"source"`
	TLStype   string      `json:"tlsType"`
	TLSEnable bool        `json:"tlsEnable"`
	TLS       *tls.Config `json:"tls"`
}

const (
	maxConn         = 50
	maxConnIdleTime = 1 * time.Minute
	maxConnLifeTime = 3 * time.Minute
)

func NewMariaDBConn(config *Config) (*sql.DB, error) {
	db, err := sql.Open(config.Driver, config.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	// Common settings for both TLS and non-TLS connections
	db.SetConnMaxIdleTime(maxConnIdleTime)
	db.SetConnMaxLifetime(maxConnLifeTime)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(maxConn)

	if config.TLSEnable {
		mysql.RegisterTLSConfig(config.TLStype, config.TLS)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}
