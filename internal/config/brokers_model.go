package config

type BrokerKey string

const (
	BROKER       BrokerKey = "brokers"
	BROKER_KAFKA BrokerKey = "kafka"
)

type Brokers struct {
	Kafka *Kafka `mapstructure:"kafka"`
}
