package config

import (
	"context"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// WatchForUpdates sets up a watcher for configuration updates, allowing dynamic handling of differen configuration keys.
// It listens for updates in teh specified etcd prefix and processes configuratio nchanges base on their priority.
// Prioritizes configuration keys based on the provided keyPriorities map, allowing dynamic adjustments to key priorities.
// Assigns a default priority of 3 to keys not explicitly defined in the KeyPriorities map.
func (m *Manager) Watch(
	ctx context.Context,
	kp KeyPriorities,
) {
	watchCh := m.client.Watch(ctx, m.app.Etcd.Prefix, clientv3.WithPrefix())
	m.log.Info("watching for updates")
	m.log.Infof("watch config: %t", m.app != nil)
	m.log.Infof("watch pqsql: %t", m.pqsqlConnection != nil)
	m.log.Infof("watch token creator: %t", m.tokenMaker != nil)
	m.log.Infof("watch producer worker: %v", m.workerProducerConnection != nil)
	m.log.Infof("watch consumer worker: %v", m.workerConsumerConnection != nil)
	m.log.Infof("watch redis: %t", m.redisConnection != nil)
	m.log.Infof("watch consul: %v", m.consulClientConnection != nil)
	m.log.Info("start watching configurations...")

	for watchResp := range watchCh {
		eventsByPriority := make(map[int][]*clientv3.Event)

		for _, ev := range watchResp.Events {
			m.log.Info("received updated configuration from: ", string(ev.Kv.Key))
			priority, exists := kp[string(ev.Kv.Key)]
			if !exists {
				kp[string(ev.Kv.Key)] = 3
			}
			eventsByPriority[priority] = append(eventsByPriority[priority], ev)
		}

		for _, priority := range m.getSortedPriorities(kp) {
			events := eventsByPriority[priority]
			for _, ev := range events {
				key := string(ev.Kv.Key)
				value := string(ev.Kv.Value)

				switch key {
				case m.app.Etcd.Keys.Configurations.Services:
					if err := helper.ViperReader(string(SERVICES), value, &m.app.Services); err != nil {
						m.log.Errorf("%v", err)
					}

					m.NotifyObservers(key)
				case m.app.Etcd.Keys.Configurations.ServiceDiscovery:
					if m.consulClientConnection != nil {
						if err := helper.ViperReader(string(SERVICE_DISCOVERY), value, &m.app.ServiceDiscovery); err != nil {
							m.log.Errorf("%v", err)
						}

						if err := m.setConsulConn(); err != nil {
							m.log.Errorf("%v", err)
						}

						m.NotifyConsulServerObservers(key)
					}
				case m.app.Etcd.Keys.Configurations.Postgresql:
					if m.pqsqlConnection != nil {
						if err := helper.ViperReader(string(PQSQL), value, &m.app.Databases.PostgreSQL); err != nil {
							m.log.Errorf("%v", err)

						}

						if err := m.setPqsqlConn(ctx); err != nil {
							m.log.Errorf("%v", err)
						}

						m.NotifyPqsqlObservers(key)
					}
				case m.app.Etcd.Keys.Configurations.Redis:
					if m.redisConnection != nil {
						if err := helper.ViperReader(string(REDIS), value, &m.app.Databases.Redis); err != nil {
							m.log.Errorf("%v", err)

						}

						if err := m.setRedisConn(ctx); err != nil {
							m.log.Errorf("%v", err)

						}

						m.NotifyRedisObservers(key)
					}
				case m.app.Etcd.Keys.Configurations.Brokers:
					if m.workerProducerConnection != nil {
						if err := helper.ViperReader(string(BROKER), value, &m.app.Brokers); err != nil {
							m.log.Errorf("%v", err)

						}

						if err := m.setKafkaProducerConn(); err != nil {
							m.log.Errorf("%v", err)
						}

						m.NotifyProducerWorkerObservers(key)
					}
					if m.workerConsumerConnection != nil {
						if err := helper.ViperReader(string(BROKER), value, &m.app.Brokers); err != nil {
							m.log.Errorf("%v", err)

						}

						if err := m.setKafkaConsumerConn(); err != nil {
							m.log.Errorf("%v", err)
						}

						m.NotifyConsumerWorkerObservers(key)
					}
				case m.app.Etcd.Keys.TLS.Consul:
					if m.consulClientConnection != nil {
						if err := helper.ViperReader(string(TLS_CONSUL), value, &m.app.TLS.Consul); err != nil {
							m.log.Errorf("%v", err)
						}

						if err := m.setConsulConn(); err != nil {
							m.log.Errorf("%v", err)
						}

						m.NotifyConsulServerObservers(key)
					}
				case m.app.Etcd.Keys.TLS.Postgresql:
					if m.pqsqlConnection != nil {
						if err := helper.ViperReader(string(TLS_PQSQL), value, &m.app.TLS.PostgreSQL); err != nil {
							m.log.Errorf("%v", err)
						}

						if err := m.setPqsqlConn(ctx); err != nil {
							m.log.Errorf("%v", err)
						}

						m.NotifyPqsqlObservers(key)
					}
				case m.app.Etcd.Keys.TLS.Redis:
					if m.redisConnection != nil {
						if err := helper.ViperReader(string(TLS_REDIS), value, &m.app.TLS.Redis); err != nil {
							m.log.Errorf("%v", err)

						}

						if err := m.setRedisConn(ctx); err != nil {
							m.log.Errorf("%v", err)
						}

						m.NotifyRedisObservers(key)
					}
				case m.app.Etcd.Keys.TLS.Kafka:
					if m.workerProducerConnection != nil {
						if err := helper.ViperReader(string(TLS_KAFKA), value, &m.app.Brokers); err != nil {
							m.log.Errorf("%v", err)

						}

						if err := m.setKafkaProducerConn(); err != nil {
							m.log.Errorf("%v", err)

						}

						m.NotifyProducerWorkerObservers(key)
					}
					if m.workerConsumerConnection != nil {
						if err := helper.ViperReader(string(BROKER), value, &m.app.Brokers); err != nil {
							m.log.Errorf("%v", err)

						}

						if err := m.setKafkaConsumerConn(); err != nil {
							m.log.Errorf("%v", err)
						}

						m.NotifyConsumerWorkerObservers(key)
					}
				case m.app.Etcd.Keys.TLS.Paseto:
					if m.tokenMaker != nil {
						if err := helper.ViperReader(string(TLS_PASETO), value, &m.app.TLS.Paseto); err != nil {
							m.log.Errorf("%v", err)

						}

						if err := m.setTokenMaker(); err != nil {
							m.log.Errorf("%v", err)

						}

						m.NotifyTokenMakerObservers(key)
					}
				case m.app.Etcd.Keys.Configurations.Encryption:
					if err := helper.ViperReader(string(ENCRYPTION), value, &m.app.Encryption); err != nil {
						m.log.Errorf("%v", err)

					}

					m.NotifyObservers(key)
				}
			}
		}
	}
}
