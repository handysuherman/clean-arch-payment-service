package config

import "time"

type RedisKey string

const (
	REDIS RedisKey = "redis"
)

type Redis struct {
	Servers   []string       `mapstructure:"servers"`
	DB        int            `mapstructure:"db"`
	Password  string         `mapstructure:"password"`
	AppID     string         `mapstructure:"appID"`
	PoolSize  int            `mapstructure:"poolSize"`
	EnableTLS bool           `mapstructure:"enableTLS"`
	Prefixes  *RedisPrefixes `mapstructure:"prefixes"`
}

type RedisPrefixes struct {
	CreatePayment *Prefixes `mapstructure:"create_payment"`
	Customer      *Prefixes `mapstructure:"customer"`
	Payment       *Prefixes `mapstructure:"payment"`
}

type Prefixes struct {
	Prefix             string        `mapstructure:"prefix"`
	ExpirationDuration time.Duration `mapstructure:"expirationDuration"`
}
