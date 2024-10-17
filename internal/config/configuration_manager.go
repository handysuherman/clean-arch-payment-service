package config

import (
	"context"
	"fmt"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	kafkaClient "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/token"
	"github.com/hashicorp/consul/api"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ConfigObserver is an interface representing an entity that observes and reacts to updates
// in the configuration managed by the etcd-backend configuration manager.
type ConfigObserver interface {
	//OnConfigUpdate is called by the configuration manager when a configuration kv pair is updated
	// into the etcd store. Implementing this method allows an object to react to changes in the configuration.
	OnConfigUpdate(key string, config *App)
}

type ObserverWithPriority struct {
	Observer ConfigObserver
	Priority int
}

type KeyPriorities map[string]int

type ConsulServerObserver interface {
	OnConsulUpdate(key string, consulClientConnection *api.Client)
}

type PqsqlObserver interface {
	OnPqsqlUpdate(key string, pqsqlConnection *pgxpool.Pool)
}

type RedisObserver interface {
	OnRedisUpdate(key string, redisConnection redis.UniversalClient)
}

type WorkerConsumerObserver interface {
	OnConsumerWorkerUpdate(key string, workerConsumerObserver *kafka.Reader)
}

type WorkerProducerObserver interface {
	OnProducerWorkerUpdate(key string, workerProducerConnection *kafkaClient.ProducerImpl)
}

type TokenMakerObserver interface {
	OnTokenUpdate(key string, maker token.Maker)
}

type consulAppSrvRegistry struct {
	appUniqueID      string
	appPublishedPort int
	appHostname      string
	healthCheckAddr  string
}

// Manager represents a configuration manager that facilitates communications
// and updates related to configuration settings.
type Manager struct {
	observers               []ObserverWithPriority
	consulServerObservers   []ConsulServerObserver
	pqsqlObservers          []PqsqlObserver
	redisObservers          []RedisObserver
	tokenMakerObservers     []TokenMakerObserver
	workerConsumerObservers []WorkerConsumerObserver
	workerProducerObservers []WorkerProducerObserver
	consulAppSrvCfg         *consulAppSrvRegistry
	client                  *clientv3.Client
	log                     logger.Logger
	updateCh                chan *App
	app                     *App
	etcdTimeout             time.Duration

	pqsqlConnection          *pgxpool.Pool
	consulClientConnection   *api.Client
	redisConnection          redis.UniversalClient
	workerProducerConnection *kafkaClient.ProducerImpl
	workerConsumerConnection *kafka.Reader
	tokenMaker               token.Maker
}

func NewManager(
	log logger.Logger,
	etcdTimeout time.Duration,
) *Manager {
	return &Manager{
		log:         log.WithPrefix("configuration-manager"),
		updateCh:    make(chan *App),
		etcdTimeout: etcdTimeout,
	}
}

func (m *Manager) Bootstrap(path string) (*App, error) {
	viper.SetConfigType(constants.Yaml)
	viper.SetConfigFile(path)

	m.app = &App{}
	m.app.Services = &Services{}
	m.app.ServiceDiscovery = &ServiceDiscovery{}
	m.app.Brokers = &Brokers{}
	m.app.Encryption = &Encryption{}
	m.app.Databases = &Databases{}
	m.app.TLS = &TLS{}
	m.app.Monitoring = &Monitoring{}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper.ReadInConfig.err: %v", err)
	}

	if err := viper.UnmarshalKey(constants.ETCD, &m.app.Etcd); err != nil {
		return nil, fmt.Errorf("viper.Unmarshal.err: %v", err)
	}

	if err := m.setEtcd(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.etcdTimeout)
	defer cancel()

	if err := m.loadConfig(ctx); err != nil {
		return nil, err
	}

	return m.app, nil
}

func (m *Manager) SetConsulAppSrvConfig(
	appUniqueID string,
	appPublishedPort int,
	appHostname string,
	healthCheckAddr string,
	healthCheckPort int,
) {
	m.consulAppSrvCfg = &consulAppSrvRegistry{
		appUniqueID:      appUniqueID,
		appPublishedPort: appPublishedPort,
		appHostname:      appHostname,
		healthCheckAddr:  fmt.Sprintf("%s:%d", healthCheckAddr, healthCheckPort),
	}
}

func (m *Manager) SetAppConsulRegisterConfig() *api.AgentServiceRegistration {
	return &api.AgentServiceRegistration{
		ID:      m.consulAppSrvCfg.appUniqueID,
		Name:    m.app.ServiceDiscovery.Consul.Internal.ServiceName,
		Tags:    []string{constants.GRPC},
		Port:    m.consulAppSrvCfg.appPublishedPort,
		Address: m.consulAppSrvCfg.appHostname,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s/consul", m.consulAppSrvCfg.healthCheckAddr),
			Timeout:                        m.app.ServiceDiscovery.Consul.Internal.Timeout,
			Interval:                       m.app.ServiceDiscovery.Consul.Internal.Interval,
			DeregisterCriticalServiceAfter: m.app.ServiceDiscovery.Consul.Internal.DeregisterCriticalServiceAfter,
		},
	}
}

func (m *Manager) WithConsulConnection(consulClientConnection *api.Client) {
	m.consulClientConnection = consulClientConnection
}

func (m *Manager) ConsulConnection() *api.Client {
	return m.consulClientConnection
}

func (m *Manager) PqsqlConnection() *pgxpool.Pool {
	return m.pqsqlConnection
}

func (m *Manager) WithPqsqlConnection(pqsqlConnection *pgxpool.Pool) {
	m.pqsqlConnection = pqsqlConnection
}

func (m *Manager) RedisConnection() redis.UniversalClient {
	return m.redisConnection
}

func (m *Manager) WithRedisConnection(redisConnection redis.UniversalClient) {
	m.redisConnection = redisConnection
}

func (m *Manager) ProducerWorker() *kafkaClient.ProducerImpl {
	return m.workerProducerConnection
}

func (m *Manager) WithProducerWorker(producerConnection *kafkaClient.ProducerImpl) {
	m.workerProducerConnection = producerConnection
}

func (m *Manager) WithConsumerWorker(workerConsumerConnection *kafka.Reader) {
	m.workerConsumerConnection = workerConsumerConnection
}

func (m *Manager) TokenMaker() token.Maker {
	return m.tokenMaker
}

func (m *Manager) WithTokenMaker(tokenMaker token.Maker) {
	m.tokenMaker = tokenMaker
}

func (m *Manager) Close() error {
	return m.client.Close()
}

func (m *Manager) ClosePqsql() {
	m.pqsqlConnection.Close()
}

func (m *Manager) CloseRedis() error {
	return m.redisConnection.Close()
}

func (m *Manager) CloseConsumerWorker() error {
	return m.workerConsumerConnection.Close()
}

func (m *Manager) CloseProducerWorker() error {
	return m.workerProducerConnection.Close()
}
