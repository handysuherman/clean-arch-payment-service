package consul

import (
	"encoding/base64"
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	_consul "github.com/handysuherman/clean-arch-payment-service/internal/pkg/consul"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

func NewConsulGRPCServiceResolver(
	cfg *config.App,
	cm *config.Manager,
	log logger.Logger,
	scheme string,
	dnsResolver string,
	queryOptions *api.QueryOptions,
) error {
	if cfg.ServiceDiscovery.Consul.External.Payment.ServiceName == "" {
		return fmt.Errorf("service name not specified")
	}

	certs, err := loadConsulCerts(cfg, log)
	if err != nil {
		return fmt.Errorf("unable to retrive consul certs: %v", err)
	}

	consulClient, err := _consul.New(cfg.ServiceDiscovery.Consul.External.Payment.Host, cfg.ServiceDiscovery.Consul.External.Payment.EnableTLS, certs)
	if err != nil {
		return fmt.Errorf("unable to retrieve client from consul: %v", err)
	}

	if queryOptions == nil {
		queryOptions = _consul.DefaultQueryOptions()
	}

	resolverBuilder := &consulGrpcClientResolverBuilder{
		consulClient: consulClient,
		serviceName:  cfg.ServiceDiscovery.Consul.External.Payment.ServiceName,
		scheme:       scheme,
		dnsResolver:  dnsResolver,
		queryOptions: queryOptions,
	}
	cm.RegisterConsulServerObserver(resolverBuilder)

	resolver.Register(resolverBuilder)

	return nil
}

func loadConsulCerts(cfg *config.App, log logger.Logger) (*_consul.ConsulClientCerts, error) {
	ca, err := base64.StdEncoding.DecodeString(cfg.TLS.Consul.Ca)
	if err != nil {
		log.Errorf("error while decoding consulConn ca: %v", err)
		return nil, err
	}

	cert, err := base64.StdEncoding.DecodeString(cfg.TLS.Consul.Cert)
	if err != nil {
		log.Errorf("error while decoding consulConn cert: %v", err)
		return nil, err
	}

	key, err := base64.StdEncoding.DecodeString(cfg.TLS.Consul.Key)
	if err != nil {
		log.Errorf("error while ecoding consulConn key: %v", err)
		return nil, err
	}

	return &_consul.ConsulClientCerts{
		Ca:   ca,
		Cert: cert,
		Key:  key,
	}, nil
}
