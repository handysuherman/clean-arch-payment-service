package consul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

type ConsulClientCerts struct {
	Ca   []byte
	Cert []byte
	Key  []byte
}

type ConsulServiceConfig struct {
	ServiceName  string
	ConsulAddr   string
	TLSEnabled   bool
	ConsulCerts  *ConsulClientCerts
	QueryOptions *api.QueryOptions
}

func New(consulAddr string, tlsEnabled bool, certs *ConsulClientCerts) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr

	if tlsEnabled {
		if certs == nil {
			return nil, fmt.Errorf("certs should not be nil when tls communication enabled")
		}

		config.Scheme = "https"
		config.TLSConfig = api.TLSConfig{
			CAPem:   certs.Ca,
			CertPEM: certs.Cert,
			KeyPEM:  certs.Key,
		}
	}

	return api.NewClient(config)
}

func DefaultQueryOptions() *api.QueryOptions {
	return &api.QueryOptions{
		UseCache:          true,
		MaxAge:            30 * time.Second,
		StaleIfError:      15 * time.Second,
		WaitTime:          15 * time.Second,
		RequireConsistent: true,
	}
}
