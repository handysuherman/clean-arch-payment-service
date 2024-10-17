package config

import (
	"context"
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// loadConfig fetches configuration from the etcd store base on predefines keys and populates
// the application's confioguration struct field accordingly. It leverages the ViperReader method
// to parse and apply the configuration to appropiate sections of the application's configuration.
// The method iterates trhough kv pairs retrieved from etcd store and applies the corresponding
// configuration based on predefined keys. Each configuration is directed to the relevant section within
// the application's configuration struct, ensuring a well-organized and structured configuration setup.
func (m *Manager) loadConfig(ctx context.Context) error {
	response, err := m.client.Get(ctx, m.app.Etcd.Prefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("m.client.Get.err: %v", err)
	}

	for _, kv := range response.Kvs {
		value := string(kv.Value)

		switch string(kv.Key) {
		case m.app.Etcd.Keys.Configurations.Services:
			if err := helper.ViperReader(constants.Services, value, &m.app.Services); err != nil {
				m.log.Errorf("error while reading service configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.Configurations.ServiceDiscovery:
			if err := helper.ViperReader(string(SERVICE_DISCOVERY), value, &m.app.ServiceDiscovery); err != nil {
				m.log.Errorf("error while reading service_discovery configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.Configurations.Brokers:
			if err := helper.ViperReader(constants.Brokers, value, &m.app.Brokers); err != nil {
				m.log.Errorf("error while reading brokers configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.Configurations.Postgresql:
			if err := helper.ViperReader(constants.PostgreSQL, value, &m.app.Databases.PostgreSQL); err != nil {
				m.log.Errorf("error while reading pqsql configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.Configurations.Redis:
			if err := helper.ViperReader(constants.Redis, value, &m.app.Databases.Redis); err != nil {
				m.log.Errorf("error while reading redis configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.Configurations.Encryption:
			if err := helper.ViperReader(constants.Encryption, value, &m.app.Encryption); err != nil {
				m.log.Errorf("error while reading encryption configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.Configurations.Monitoring:
			if err := helper.ViperReader(constants.Monitoring, value, &m.app.Monitoring); err != nil {
				m.log.Errorf("error while reading monitoring configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.TLS.App:
			if err := helper.ViperReader(constants.AppTLS, value, &m.app.TLS.App); err != nil {
				m.log.Errorf("error while reading tls.app configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.TLS.Consul:
			if err := helper.ViperReader(string(TLS_CONSUL), value, &m.app.TLS.Consul); err != nil {
				m.log.Errorf("error while reading tls.consul configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.TLS.Postgresql:
			if err := helper.ViperReader(constants.PostgreSQL, value, &m.app.TLS.PostgreSQL); err != nil {
				m.log.Errorf("error while reading tls.pqsql configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.TLS.Kafka:
			if err := helper.ViperReader(constants.KafkaTLS, value, &m.app.TLS.Kafka); err != nil {
				m.log.Errorf("error while reading tls.kafka configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.TLS.Redis:
			if err := helper.ViperReader(constants.RedisTLS, value, &m.app.TLS.Redis); err != nil {
				m.log.Errorf("error while reading tls.redis configuration block: %v", err)
				return err
			}
		case m.app.Etcd.Keys.TLS.Paseto:
			if err := helper.ViperReader(constants.PasetoTLS, value, &m.app.TLS.Paseto); err != nil {
				m.log.Errorf("error while reading tls.paseto configuration block: %v", err)
				return err
			}
		}
	}

	return nil
}
