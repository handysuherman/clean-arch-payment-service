package infra

import "github.com/handysuherman/clean-arch-payment-service/internal/pkg/consul"

func (a *app) consul() error {
	consulCerts, err := a.loadConsulCerts()
	if err != nil {
		a.log.Warnf("unable to load consul certs: %v", err)
		return err
	}

	consulClient, err := consul.New(a.cfg.ServiceDiscovery.Consul.Internal.Host, a.cfg.ServiceDiscovery.Consul.Internal.EnableTLS, consulCerts)
	if err != nil {
		a.log.Warnf("unable to initiate new connection of consul: %v", err)
		return err
	}

	a.cfgManager.WithConsulConnection(consulClient)

	return nil
}
