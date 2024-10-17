package config

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"sort"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/consul"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/databases/postgresql"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/databases/redis"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	kafkaClient "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/token"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func (m *Manager) setRedisConn(ctx context.Context) error {
	redisOpt := redis.Config{
		Host:     m.app.Databases.Redis.Servers[0],
		Password: m.app.Databases.Redis.Password,
		DB:       m.app.Databases.Redis.DB,
		PoolSize: m.app.Databases.Redis.PoolSize,
	}

	if m.app.Databases.Redis.EnableTLS {
		tlsCfg, err := helper.Base64EncodedTLS(
			m.app.TLS.Redis.Ca,
			m.app.TLS.Redis.Cert,
			m.app.TLS.Redis.Key,
		)
		if err != nil {
			return err
		}

		redisOpt.TLsEnable = m.app.Databases.Redis.EnableTLS
		redisOpt.TLs = tlsCfg
	}

	redisConn, err := redis.NewUniversalRedisClient(ctx, &redisOpt)
	if err != nil {
		return err
	}

	m.redisConnection = redisConn

	m.log.Info("establishing updated configuration of REDIS repository")
	return nil
}

func (m *Manager) setPqsqlConn(ctx context.Context) error {
	psqlOpt := &postgresql.Config{
		Host:      m.app.Databases.PostgreSQL.Host,
		Port:      m.app.Databases.PostgreSQL.Port,
		DBName:    m.app.Databases.PostgreSQL.DBName,
		User:      m.app.Databases.PostgreSQL.Username,
		Password:  m.app.Databases.PostgreSQL.Password,
		EnableTls: m.app.Databases.PostgreSQL.EnableTLS,
	}

	if m.app.Databases.PostgreSQL.EnableTLS {
		tlsCfg, err := helper.Base64EncodedTLS(
			m.app.TLS.PostgreSQL.Ca,
			m.app.TLS.PostgreSQL.Cert,
			m.app.TLS.PostgreSQL.Key,
		)
		if err != nil {
			return err
		}

		tlsCfg.ServerName = m.app.Databases.PostgreSQL.Host
		psqlOpt.TLs = tlsCfg
	}

	m.pqsqlConnection.Close()

	pgxConn, err := postgresql.NewPgxConn(ctx, psqlOpt)
	if err != nil {
		return errors.Wrap(err, "postgresql.NewPgxConnection")
	}

	m.pqsqlConnection = pgxConn

	m.log.Infof("establishing updated configuration of PQSQL repository: %v", m.pqsqlConnection.Stat().TotalConns())
	return nil
}

func (m *Manager) setEtcd() error {
	etcdConfig := clientv3.Config{
		DialTimeout: m.etcdTimeout,
	}

	if m.app.Etcd.TLS.Enabled {
		tlsCfg, err := helper.Base64EncodedTLS(
			m.app.Etcd.TLS.Ca,
			m.app.Etcd.TLS.Cert,
			m.app.Etcd.TLS.Key,
		)
		if err != nil {
			return err
		}

		etcdConfig.TLS = tlsCfg
	}

	etcdConfig.Endpoints = m.app.Etcd.Host

	etcdClient, err := clientv3.New(etcdConfig)
	if err != nil {
		return err
	}

	m.client = etcdClient

	return nil
}

func (m *Manager) setKafkaConsumerConn() error {
	m.workerConsumerConnection.Close()

	kc := kafkaClient.NewConsumerGroup(m.app.Brokers.Kafka.Config.Brokers, m.app.Brokers.Kafka.Config.GroupID, m.log)

	kafkaTLS, err := helper.Base64EncodedTLS(m.app.TLS.Kafka.Ca, m.app.TLS.Kafka.Cert, m.app.TLS.Kafka.Key)
	if err != nil {
		return err
	}

	m.workerConsumerConnection = kc.GetNewKafkaReader(
		m.app.Brokers.Kafka.Config.Brokers,
		m.getConsumerTopics(),
		m.app.Brokers.Kafka.Config.GroupID,
		m.app.Etcd.TLS.Enabled,
		kafkaTLS,
	)

	m.log.Info("establishing updated configuration of KAFKA Consumer repository")
	return nil
}

func (m *Manager) setKafkaProducerConn() error {
	m.workerProducerConnection.Close()

	kafkaTLS, err := helper.Base64EncodedTLS(m.app.TLS.Kafka.Ca, m.app.TLS.Kafka.Cert, m.app.TLS.Kafka.Key)
	if err != nil {
		return err
	}

	m.workerProducerConnection = kafkaClient.NewProducerImpl(m.log, m.app.Brokers.Kafka.Config.Brokers, m.app.Brokers.Kafka.Config.EnableTLS, kafkaTLS)

	m.log.Info("establishing updated configuration of KAFKA PRODUCER repository")
	return nil
}

func (m *Manager) setTokenMaker() error {
	privateKey, err := base64.StdEncoding.DecodeString(m.app.TLS.Paseto.PrivateKey)
	if err != nil {
		m.log.Errorf("error while decoding tokenMaker privateKey: %v", err)
		return err
	}

	publicKey, err := base64.StdEncoding.DecodeString(m.app.TLS.Paseto.PublicKey)
	if err != nil {
		m.log.Errorf("error while ecoding tokenMaker publicKey: %v", err)
		return err
	}

	decodePrivateKey, _ := pem.Decode(privateKey)
	decodeBlockPrvKey, err := x509.ParsePKCS8PrivateKey(decodePrivateKey.Bytes)
	if err != nil {
		m.log.Errorf("error decoding block private key: %v", err)
		return err
	}
	ed25519privateKey, _ := decodeBlockPrvKey.(ed25519.PrivateKey)

	decodePublicKey, _ := pem.Decode(publicKey)
	decodeBlockPubKey, err := x509.ParsePKIXPublicKey(decodePublicKey.Bytes)
	if err != nil {
		m.log.Errorf("error decoding block public key: %v", err)
		return err
	}
	ed25519publicKey, _ := decodeBlockPubKey.(ed25519.PublicKey)

	tokenCreator, err := token.NewPaseto(ed25519privateKey, ed25519publicKey, "", m.app.Etcd.Nonce)
	if err != nil {
		m.log.Errorf("token.NewPaseto.err: %v", err)
		return err
	}

	m.tokenMaker = tokenCreator

	return nil
}

func (m *Manager) setConsulConn() error {
	ca, err := base64.StdEncoding.DecodeString(m.app.TLS.Consul.Ca)
	if err != nil {
		m.log.Errorf("error while decoding consulConn ca: %v", err)
		return err
	}

	cert, err := base64.StdEncoding.DecodeString(m.app.TLS.Consul.Cert)
	if err != nil {
		m.log.Errorf("error while decoding consulConn cert: %v", err)
		return err
	}

	key, err := base64.StdEncoding.DecodeString(m.app.TLS.Consul.Key)
	if err != nil {
		m.log.Errorf("error while ecoding consulConn key: %v", err)
		return err
	}

	consulCerts := &consul.ConsulClientCerts{
		Ca:   ca,
		Cert: cert,
		Key:  key,
	}

	if m.app.ServiceDiscovery.Consul.Internal.Register {
		m.consulClientConnection.Agent().ServiceDeregister(m.app.ServiceDiscovery.Consul.Internal.ServiceName)
	}

	consulClient, err := consul.New(m.app.ServiceDiscovery.Consul.Internal.Host, m.app.ServiceDiscovery.Consul.Internal.EnableTLS, consulCerts)
	if err != nil {
		m.log.Errorf("error while creating new consul connection: %v", err)
		return err
	}

	m.consulClientConnection = consulClient

	if m.app.ServiceDiscovery.Consul.Internal.Register {
		if err := m.consulClientConnection.Agent().ServiceRegister(m.SetAppConsulRegisterConfig()); err != nil {
			m.log.Errorf("m.consulClientConnection.Agent().ServiceRegister.err: %v", err)
			return err
		}
	}

	m.log.Info("establishing updated configuration of CONSUL repository")
	return nil
}

func (m *Manager) getConsumerTopics() []string {
	return []string{
		helper.StringBuilder(m.app.Services.External.PaymentGateway.ID, "_", m.app.Brokers.Kafka.Topics.PaymentStatusUpdate.TopicName),
	}
}

func (m *Manager) getSortedPriorities(priorities map[string]int) []int {
	sorted := make([]int, 0, len(priorities))
	for _, priority := range priorities {
		sorted = append(sorted, priority)
	}

	sort.Sort(&sortPriorities{sorted})
	return sorted
}
