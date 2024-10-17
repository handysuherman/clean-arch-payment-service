package metrics

import (
	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	SuccessGrpcRequest prometheus.Counter
	ErrorGrpcRequest   prometheus.Counter

	CreatePaymentGrpcRequests              prometheus.Counter
	GetPaymentByIDGrpcRequests             prometheus.Counter
	GetPaymentChannelGrpcRequests          prometheus.Counter
	GetAvailablePaymentChannelsGrpcRequest prometheus.Counter

	SuccessKafkaRequest prometheus.Counter
	ErrorKafkaRequest   prometheus.Counter

	PaymentStatusUpdateKafkaMessages prometheus.Counter
}

func New(cfg *config.App) *Metrics {
	return &Metrics{
		SuccessGrpcRequest: NewCounter(cfg, "success_grpc", constants.GRPC),
		ErrorGrpcRequest:   NewCounter(cfg, "error_grpc", constants.GRPC),

		CreatePaymentGrpcRequests:              NewCounter(cfg, "create_payment_grpc", constants.GRPC),
		GetPaymentByIDGrpcRequests:             NewCounter(cfg, "get_payment_by_id_grpc", constants.GRPC),
		GetPaymentChannelGrpcRequests:          NewCounter(cfg, "get_payment_channel_grpc", constants.GRPC),
		GetAvailablePaymentChannelsGrpcRequest: NewCounter(cfg, "get_available_payment_channels_grpc", constants.GRPC),

		SuccessKafkaRequest: NewCounter(cfg, "success_kafka", constants.Kafka),
		ErrorKafkaRequest:   NewCounter(cfg, "error_kafka", constants.Kafka),

		PaymentStatusUpdateKafkaMessages: NewCounter(cfg, "payment_status_update_kafka", constants.Kafka),
	}
}
