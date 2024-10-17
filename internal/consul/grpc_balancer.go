package consul

import (
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

type consulGrpcClientResolverBuilder struct {
	consulClient *api.Client
	log          logger.Logger
	serviceName  string
	scheme       string
	dnsResolver  string
	queryOptions *api.QueryOptions
}

type consulGrpcClientResolver struct {
	target    resolver.Target
	cc        resolver.ClientConn
	addrStore map[string][]string
}

func (b *consulGrpcClientResolverBuilder) OnConsulUpdate(key string, consulClientConnection *api.Client) {
	b.log.Infof("received update from '%s' key", key)

	b.consulClient = consulClientConnection

	b.log.Infof("updated configuration from '%s' key successfully applied", key)
}

func (b *consulGrpcClientResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var addrs []string

	services, _, err := b.consulClient.Health().Service(b.serviceName, "", true, b.queryOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to retrive services from consul: %v", err)
	}

	for _, service := range services {
		addrs = append(addrs, fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port))
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("no available connection retrieved, please try again later")
	}

	r := &consulGrpcClientResolver{
		target: target,
		cc:     cc,
		addrStore: map[string][]string{
			b.dnsResolver: addrs,
		},
	}

	r.start()

	return r, nil
}

func (b *consulGrpcClientResolverBuilder) Scheme() string { return b.scheme }

func (r *consulGrpcClientResolver) start() {
	addrStrs := r.addrStore[r.target.Endpoint()]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}

	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (r *consulGrpcClientResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (r *consulGrpcClientResolver) Close()                                  {}
