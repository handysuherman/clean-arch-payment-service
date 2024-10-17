package config

import "sort"

// NotifyObservers noties all registered observers about the updated configuration.
// It sorts the observers based on their priority levels in ascending order.
func (m *Manager) NotifyObservers(key string) {
	sort.Sort(&sortObservers{m.observers})
	for _, observer := range m.observers {
		observer.Observer.OnConfigUpdate(key, m.app)
	}
}

func (m *Manager) NotifyConsulServerObservers(key string) {
	for _, consulServerObserver := range m.consulServerObservers {
		consulServerObserver.OnConsulUpdate(key, m.consulClientConnection)
	}
}

// NotifyPqsqlObservers notifies all registered Repository observers about the updated configuration.
func (m *Manager) NotifyPqsqlObservers(key string) {
	for _, pqsqlObserver := range m.pqsqlObservers {
		pqsqlObserver.OnPqsqlUpdate(key, m.pqsqlConnection)
	}
}

func (m *Manager) NotifyRedisObservers(key string) {
	for _, redisObserver := range m.redisObservers {
		redisObserver.OnRedisUpdate(key, m.redisConnection)
	}
}

func (m *Manager) NotifyProducerWorkerObservers(key string) {
	for _, producerWorkerObserver := range m.workerProducerObservers {
		producerWorkerObserver.OnProducerWorkerUpdate(key, m.workerProducerConnection)
	}
}

func (m *Manager) NotifyConsumerWorkerObservers(key string) {
	for _, consumerWorkerObserver := range m.workerConsumerObservers {
		consumerWorkerObserver.OnConsumerWorkerUpdate(key, m.workerConsumerConnection)
	}
}

func (m *Manager) NotifyTokenMakerObservers(key string) {
	for _, tokenObserver := range m.tokenMakerObservers {
		tokenObserver.OnTokenUpdate(key, m.tokenMaker)
	}
}
