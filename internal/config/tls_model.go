package config

type TLSKey string

const (
	TLS_APP    TLSKey = "app"
	TLS_PQSQL  TLSKey = "postgresql"
	TLS_REDIS  TLSKey = "redis"
	TLS_KAFKA  TLSKey = "kafka"
	TLS_PASETO TLSKey = "paseto"
	TLS_CONSUL TLSKey = "consul"
)

type TLS struct {
	App        *Certs           `mapstructure:"app"`
	Consul     *Certs           `mapstructure:"consul"`
	PostgreSQL *Certs           `mapstructure:"postgresql"`
	Redis      *Certs           `mapstructure:"redis"`
	Kafka      *Certs           `mapstructure:"kafka"`
	Paseto     *AssymetricCerts `mapstructure:"paseto"`
}

type Certs struct {
	Ca   string `mapstructure:"ca"`
	Cert string `mapstructure:"cert"`
	Key  string `mapstructure:"key"`
}

type AssymetricCerts struct {
	PrivateKey string `mapstructure:"privateKey"`
	PublicKey  string `mapstructure:"publicKey"`
}
