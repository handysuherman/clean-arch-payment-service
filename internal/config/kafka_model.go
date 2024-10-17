package config

import "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka"

type Kafka struct {
	Config *kafka.Config `mapstructure:"config"`
	Topics *KafkaTopics  `mapstructure:"topics"`
}

type KafkaTopics struct {
	PaymentStatusUpdate  *kafka.Topic `mapstructure:"payment_status_update"`
	PaymentStatusUpdated *kafka.Topic `mapstructure:"payment_status_updated"`
}
