package config

type App struct {
	Services         *Services         `mapstructure:"services"`
	ServiceDiscovery *ServiceDiscovery `mapstructure:"service_discovery"`
	Databases        *Databases        `mapstructure:"databases"`
	Encryption       *Encryption       `mapstructure:"encryption"`
	TLS              *TLS              `mapstructure:"tls"`
	Etcd             *Etcd             `mapstructure:"etcd"`
	Brokers          *Brokers          `mapstructure:"brokers"`
	Monitoring       *Monitoring       `mapstructure:"monitoring"`
}
