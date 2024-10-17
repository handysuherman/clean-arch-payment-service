package kafka

import (
	"context"
	"crypto/tls"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/segmentio/kafka-go"
)

// Producer is an interface for publishing messages to Kafka.
type Producer interface {
	PublishMessage(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// ProducerImpl is an implementation of the Producer interface.
type ProducerImpl struct {
	log     logger.Logger
	brokers []string
	w       *kafka.Writer
}

// NewProducerImpl creates a new instance of the Kafka producer.
func NewProducerImpl(log logger.Logger, brokers []string, enableTLS bool, tlsConf *tls.Config) *ProducerImpl {
	// Create a new Kafka writer with specified configurations.
	writer := NewWriter(brokers, kafka.LoggerFunc(log.Errorf), enableTLS, tlsConf)

	return &ProducerImpl{
		log:     log,
		brokers: brokers,
		w:       writer,
	}
}

// PublishMessage publishes Kafka messages using the underlying Kafka writer.
func (p *ProducerImpl) PublishMessage(ctx context.Context, msgs ...kafka.Message) error {
	return p.w.WriteMessages(ctx, msgs...)
}

// Close closes the underlying Kafka writer.
func (p *ProducerImpl) Close() error {
	return p.w.Close()
}
