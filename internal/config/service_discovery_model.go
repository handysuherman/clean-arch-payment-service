package config

import "time"

type ServiceDiscoveryKey string

const (
	SERVICE_DISCOVERY ServiceDiscoveryKey = "service_discovery"
	CONSUL            ServiceDiscoveryKey = "consul"
)

type ServiceDiscovery struct {
	Consul *Consul `mapstructure:"consul"`
}

type Consul struct {
	Internal *InternalService `mapstructure:"internal"`
	External *ExternalService `mapstructure:"external"`
}

type InternalService struct {
	Scheme                         string `mapstructure:"scheme"`
	Host                           string `mapstructure:"host"`
	Register                       bool   `mapstructure:"register"`
	EnableTLS                      bool   `mapstructure:"enableTLS"`
	ServiceName                    string `mapstructure:"serviceName"`
	Timeout                        string `mapstructure:"timeout"`
	Interval                       string `mapstructure:"interval"`
	DeregisterCriticalServiceAfter string `mapstructure:"deregisterCriticalServiceAfter"`
}

type ExternalService struct {
	Payment *ExternalConsulService `mapstructure:"payment"`
}

type ExternalConsulService struct {
	Scheme            string        `mapstructure:"scheme"`
	Host              string        `mapstructure:"host"`
	Register          bool          `mapstructure:"register"`
	EnableTLS         bool          `mapstructure:"enableTLS"`
	ServiceName       string        `mapstructure:"serviceName"`
	UseCache          bool          `mapstructure:"useCache"`
	MaxAge            time.Duration `mapstructure:"maxAge"`
	StaleIfError      time.Duration `mapstructure:"staleIfError"`
	WaitTime          time.Duration `mapstructure:"waitTime"`
	RequireConsistent bool          `mapstructure:"requireConsistent"`
}
