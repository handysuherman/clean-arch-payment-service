package kafka

type Config struct {
	Brokers   []string `mapstructure:"brokers"`
	GroupID   string   `mapstructure:"groupID"`
	InitTopic bool     `mapstructure:"initTopics"`
	EnableTLS bool     `mapstructure:"enableTLS"`
}

type Topic struct {
	TopicName         string `mapstructure:"topicName"`
	Partitions        uint8  `mapstructure:"partitions"`
	ReplicationFactor uint8  `mapstructure:"replicationFactor"`
}
