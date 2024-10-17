package config

func (m *Manager) RegisterObserver(observer ConfigObserver, priority int) {
	m.observers = append(m.observers, ObserverWithPriority{Observer: observer, Priority: priority})
}

func (m *Manager) RegisterConsulServerObserver(observer ConsulServerObserver) {
	m.consulServerObservers = append(m.consulServerObservers, observer)
}

func (m *Manager) RegisterPqsqlObserver(observer PqsqlObserver) {
	m.pqsqlObservers = append(m.pqsqlObservers, observer)
}

func (m *Manager) RegisterRedisObserver(observer RedisObserver) {
	m.redisObservers = append(m.redisObservers, observer)
}

// RegisterWorkerObserver registers an broker consumer layer observer to receive update.
func (m *Manager) RegisterConsumerWorkerObserver(observer WorkerConsumerObserver) {
	m.workerConsumerObservers = append(m.workerConsumerObservers, observer)
}

func (m *Manager) RegisterProducerWorkerObserver(observer WorkerProducerObserver) {
	m.workerProducerObservers = append(m.workerProducerObservers, observer)
}

func (m *Manager) RegisterTokenMakerObserver(observer TokenMakerObserver) {
	m.tokenMakerObservers = append(m.tokenMakerObservers, observer)
}
