package config

type Etcd struct {
	Host   []string    `mapstructure:"hosts"`
	Keys   *PrefixKeys `mapstructure:"keys"`
	Prefix string      `mapstructure:"prefix"`
	Nonce  string      `mapstructure:"nonce"`
	TLS    *EtcdTLS    `mapstructure:"tls"`
}

type EtcdTLS struct {
	Ca      string `mapstructure:"ca"`
	Cert    string `mapstructure:"cert"`
	Key     string `mapstructure:"key"`
	Enabled bool   `mapstructure:"enabled"`
}

type PrefixKeys struct {
	Configurations *ConfigPrefixes `mapstructure:"configurations"`
	TLS            *TLSPrefixes    `mapstructure:"tls"`
}

type ConfigPrefixes struct {
	Brokers          string `mapstructure:"brokers"`
	Postgresql       string `mapstructure:"postgresql"`
	ServiceDiscovery string `mapstructure:"service_discovery"`
	Services         string `mapstructure:"services"`
	Redis            string `mapstructure:"redis"`
	Storages         string `mapstructure:"storages"`
	Encryption       string `mapstructure:"encryption"`
	Monitoring       string `mapstructure:"monitoring"`
}

type TLSPrefixes struct {
	App        string `mapstructure:"app"`
	Consul     string `mapstructure:"consul"`
	Redis      string `mapstructure:"redis"`
	Postgresql string `mapstructure:"postgresql"`
	Kafka      string `mapstructure:"kafka"`
	Paseto     string `mapstructure:"paseto"`
}
