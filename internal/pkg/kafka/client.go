package kafka

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func NewKafkaConnection(ctx context.Context, kafkaCfg *Config, tlsCfg *tls.Config) (*kafka.Conn, error) {
	if kafkaCfg.EnableTLS {
		if tlsCfg == nil {
			return nil, fmt.Errorf("tls should not be nil when tls enabled")
		}
		dial := &kafka.Dialer{
			TLS: tlsCfg,
		}
		return dial.DialContext(ctx, "tcp", kafkaCfg.Brokers[0])
	}

	return kafka.DialContext(ctx, "tcp", kafkaCfg.Brokers[0])
}
