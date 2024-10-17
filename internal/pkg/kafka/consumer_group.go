package kafka

import (
	"context"
	"crypto/tls"
	"sync"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// MessageProcessor is an interface for processing Kafka messages.
type MessageProcessor interface {
	ProcessMessages(ctx context.Context, r *kafka.Reader, wg *sync.WaitGroup, workerID int)
}

// Worker is a function signature for processing Kafka messages concurrently.
type Worker func(ctx context.Context, r *kafka.Reader, wg *sync.WaitGroup, workerID int)

// ConsumerGroup is an interface for managing Kafka consumer groups.
type ConsumerGroup interface {
	ConsumeTopic(
		ctx context.Context,
		cancel context.CancelFunc,
		groupID, topic string,
		poolSize int,
		worker Worker,
		tlsConf *tls.Config,
	)
	GetNewKafkaReader(kafkaURL []string, topic, groupID string) *kafka.Reader
	GetNewKafkaWriter() *kafka.Writer
}

// consumerGroup is an implementation of the ConsumerGroup interface.
type consumerGroup struct {
	Brokers []string
	GroupID string
	log     logger.Logger
	r       *kafka.Reader
}

// NewConsumerGroup creates a new instance of the ConsumerGroup.
func NewConsumerGroup(brokers []string, groupID string, log logger.Logger) *consumerGroup {
	return &consumerGroup{Brokers: brokers, GroupID: groupID, log: log}
}

// GetNewKafkaReader creates a new Kafka reader with specified configurations.
func (c *consumerGroup) GetNewKafkaReader(
	kafkaURL []string,
	groupTopics []string,
	groupID string,
	enableTLS bool,
	tlsConf *tls.Config,
) *kafka.Reader {
	config := kafka.ReaderConfig{
		Brokers:                kafkaURL,
		GroupID:                groupID,
		GroupTopics:            groupTopics,
		MinBytes:               minBytes,
		MaxBytes:               maxBytes,
		QueueCapacity:          queueCapacity,
		HeartbeatInterval:      heartbeatInterval,
		CommitInterval:         commitInterval,
		PartitionWatchInterval: partitionWatchInterval,
		MaxAttempts:            maxAttempts,
		MaxWait:                maxWait,
		Dialer: &kafka.Dialer{
			Timeout: dialTimeout,
		},
	}

	// Enable TLS if specified.
	if enableTLS {
		config.Dialer.TLS = tlsConf
	}

	return kafka.NewReader(config)
}

// GetNewKafkaWriter creates a new Kafka writer with specified configurations.
func (c *consumerGroup) GetNewKafkaWriter() *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(c.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: writerRequiredAcks,
		MaxAttempts:  writerMaxAttempts,
		Compression:  compress.Snappy,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
	}

	return w
}

// ConsumeTopic starts consuming messages from a Kafka topic.
func (c *consumerGroup) ConsumeTopic(
	ctx context.Context,
	groupTopics []string,
	poolSize int,
	worker Worker,
	kafkaReader *kafka.Reader,
) {
	c.r = kafkaReader

	defer func() {
		if err := c.r.Close(); err != nil {
			c.log.Warnf("consumerGroup.r.CLose: %v", err)
		}
	}()

	c.log.Infof(
		"startting consumer groupId: %v, topic: %v, poolSize: %v",
		c.GroupID,
		groupTopics,
		poolSize,
	)

	wg := &sync.WaitGroup{}

	for i := 0; i <= poolSize; i++ {
		wg.Add(1)
		go worker(ctx, c.r, wg, i)
	}
	wg.Wait()
}
