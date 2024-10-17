package kafka

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/segmentio/kafka-go"
)

var (
	PoolSize      = 20
	retryAttempts = 3
	retryDelay    = 300 * time.Millisecond
	retryOption   = []retry.Option{
		retry.Attempts(uint(retryAttempts)),
		retry.Delay(retryDelay),
		retry.DelayType(retry.BackOffDelay),
	}
)

func (m *messageProcessor) OnConfigUpdate(key string, cfg *config.App) {
	m.cfg = cfg
}

func (m *messageProcessor) OnConsumerWorkerUpdate(key string, workerConnection *kafka.Reader) {
	switch key {
	case m.cfg.Etcd.Keys.Configurations.Brokers:
		m.log.Info("closing previous reader connection due to changes...")

		m.r = workerConnection

		m.log.Infof("worker re-connected to brokers: %v", m.r.Config().Brokers[0])
		m.log.Infof("worker re-connected with id: %v", m.r.Config().GroupID)
		m.log.Infof("worker re-subscribe to topics: %v", m.r.Config().GroupTopics)

		m.log.Info("reader connection successfully updated...")
	case m.cfg.Etcd.Keys.TLS.Kafka:
		m.log.Info("closing previous reader connection due to changes...")

		m.r = workerConnection

		m.log.Infof("worker re-connected to brokers: %v", m.r.Config().Brokers[0])
		m.log.Infof("worker re-connected with id: %v", m.r.Config().GroupID)
		m.log.Infof("worker re-subscribe to topics: %v", m.r.Config().GroupTopics)

		m.log.Info("reader connection successfully updated...")
	}
}

func (m *messageProcessor) commitMessage(ctx context.Context, msg kafka.Message) {
	m.metrics.SuccessKafkaRequest.Inc()
	m.log.KafkaLogCommitedMessage(msg.Topic, msg.Partition, msg.Offset)
	if err := m.r.CommitMessages(ctx, msg); err != nil {
		m.log.Infof("commitMessages.err: %v", err)
	}
}

func (m *messageProcessor) commitErrorMessage(ctx context.Context, msg kafka.Message) {
	m.metrics.ErrorKafkaRequest.Inc()
	m.log.KafkaLogCommitedMessage(msg.Topic, msg.Partition, msg.Offset)
	if err := m.r.CommitMessages(ctx, msg); err != nil {
		m.log.Infof("commitErrorMessages.err: %v", err)
	}
}

func (m *messageProcessor) logProcessMessage(msg kafka.Message, workerId int) {
	m.log.KafkaProcessMessage(msg.Topic, msg.Partition, string(msg.Value), workerId, msg.Offset, msg.Time)
}
