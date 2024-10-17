package infra

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"

	kafkaConsumer "github.com/handysuherman/clean-arch-payment-service/internal/payment/delivery/kafka"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	kafkaClient "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka"
	"github.com/segmentio/kafka-go"
)

func (a *app) kafkaConsumer(ctx context.Context) error {
	brokerClient := kafkaClient.NewConsumerGroup(a.cfg.Brokers.Kafka.Config.Brokers, a.cfg.Brokers.Kafka.Config.GroupID, a.log)
	consumer := kafkaConsumer.New(a.log, a.cfg, a.v, a.usecase, a.metrics, a.getConsumerGroupTopics(), a.cfgManager)
	kafkaTls := &tls.Config{}

	if a.cfg.Brokers.Kafka.Config.EnableTLS {
		tlsCfg, err := helper.Base64EncodedTLS(a.cfg.TLS.Kafka.Ca, a.cfg.TLS.Kafka.Cert, a.cfg.TLS.Kafka.Key)
		if err != nil {
			a.log.Warnf("connectKafkaBrokers.helper.base64encodedtls.err: %v", err)
			return err
		}

		kafkaTls = tlsCfg
	}

	if err := a.connectKafkaBrokers(ctx); err != nil {
		return err
	}

	kafkaReader := brokerClient.GetNewKafkaReader(
		a.cfg.Brokers.Kafka.Config.Brokers,
		a.getConsumerGroupTopics(),
		a.cfg.Brokers.Kafka.Config.GroupID,
		a.cfg.Brokers.Kafka.Config.EnableTLS,
		kafkaTls,
	)

	a.cfgManager.WithConsumerWorker(kafkaReader)

	go brokerClient.ConsumeTopic(
		ctx,
		a.getConsumerGroupTopics(),
		kafkaConsumer.PoolSize,
		consumer.ProcessMessages,
		kafkaReader,
	)

	return nil
}

func (a *app) kafkaProducer() error {
	kafkaTls := &tls.Config{}

	if a.cfg.Brokers.Kafka.Config.EnableTLS {
		tlsCfg, err := helper.Base64EncodedTLS(a.cfg.TLS.Kafka.Ca, a.cfg.TLS.Kafka.Cert, a.cfg.TLS.Kafka.Key)
		if err != nil {
			a.log.Warnf("connectKafkaBrokers.helper.base64encodedtls.err: %v", err)
			return err
		}

		kafkaTls = tlsCfg
	}

	kafkaProducer := kafkaClient.NewProducerImpl(a.log, a.cfg.Brokers.Kafka.Config.Brokers, a.cfg.Brokers.Kafka.Config.EnableTLS, kafkaTls)
	a.cfgManager.WithProducerWorker(kafkaProducer)

	return nil
}

func (a *app) connectKafkaBrokers(ctx context.Context) error {
	kafkaOpt := kafkaClient.Config{
		Brokers:   a.cfg.Brokers.Kafka.Config.Brokers,
		GroupID:   a.cfg.Brokers.Kafka.Config.GroupID,
		InitTopic: true,
		EnableTLS: a.cfg.Brokers.Kafka.Config.EnableTLS,
	}

	kafkaTls := &tls.Config{}

	if a.cfg.Brokers.Kafka.Config.EnableTLS {
		tlsCfg, err := helper.Base64EncodedTLS(a.cfg.TLS.Kafka.Ca, a.cfg.TLS.Kafka.Cert, a.cfg.TLS.Kafka.Key)
		if err != nil {
			a.log.Warnf("connectKafkaBrokers.helper.base64encodedtls.err: %v", err)
			return err
		}

		kafkaTls = tlsCfg
	}

	kafkaConn, err := kafkaClient.NewKafkaConnection(ctx, &kafkaOpt, kafkaTls)
	if err != nil {
		a.log.Warnf("connectKafkaBrokers.kafkaClient.newkafkaconnection.err: %v", err)
		return err
	}

	a.kafkaConn = kafkaConn

	brokers, err := a.kafkaConn.Brokers()
	if err != nil {
		a.log.Warnf("connectKafkaBrokers.kafkaConn.brokers: %v", err)
		return err
	}

	a.log.Infof("connected to kafka brokers: %+v", brokers)

	return nil
}

func (a *app) initKafkaTopic(ctx context.Context) error {
	controller, err := a.kafkaConn.Controller()
	if err != nil {
		a.log.Warnf("initKafkaTopic.kafkaConn.Controller.err: %v", err)
		return err
	}

	controllerURI := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	a.log.Infof("kafka controller uri: %v", controllerURI)

	paymentStatusUpdate := kafka.TopicConfig{
		Topic:             helper.StringBuilder(a.cfg.Services.External.PaymentGateway.ID, "_", a.cfg.Brokers.Kafka.Topics.PaymentStatusUpdate.TopicName),
		NumPartitions:     int(a.cfg.Brokers.Kafka.Topics.PaymentStatusUpdate.Partitions),
		ReplicationFactor: int(a.cfg.Brokers.Kafka.Topics.PaymentStatusUpdate.ReplicationFactor),
	}

	paymentStatusUpdated := kafka.TopicConfig{
		Topic:             helper.StringBuilder(a.cfg.Services.Internal.ID, "_", a.cfg.Brokers.Kafka.Topics.PaymentStatusUpdated.TopicName),
		NumPartitions:     int(a.cfg.Brokers.Kafka.Topics.PaymentStatusUpdated.Partitions),
		ReplicationFactor: int(a.cfg.Brokers.Kafka.Topics.PaymentStatusUpdated.ReplicationFactor),
	}

	if err := a.kafkaConn.CreateTopics(paymentStatusUpdate, paymentStatusUpdated); err != nil {
		a.log.Warnf("initKafkaTopic.kafkaConn.CreateTopics.err: %v", err)
		return err
	}

	a.log.Infof("kafka topics created or already exists: %+v", []kafka.TopicConfig{paymentStatusUpdate, paymentStatusUpdated})
	return nil
}

func (a *app) getConsumerGroupTopics() []string {
	return []string{
		helper.StringBuilder(a.cfg.Services.External.PaymentGateway.ID, "_", a.cfg.Brokers.Kafka.Topics.PaymentStatusUpdate.TopicName),
	}
}
