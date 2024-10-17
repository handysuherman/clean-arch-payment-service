package kafka

import (
	"context"

	"github.com/avast/retry-go"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	_kafkaMessage "github.com/handysuherman/clean-arch-payment-service/internal/proto/kafka"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

func (m *messageProcessor) processUpdatePaymentStatus(ctx context.Context, msg kafka.Message) {
	m.metrics.PaymentStatusUpdateKafkaMessages.Inc()

	ctx, span := tracing.StartKafkaConsumerTracerSpan(ctx, msg.Headers, "messageProcessor.processUpdatePaymentStatus")
	defer span.Finish()

	var _msg _kafkaMessage.KafkaPaymentStatusUpdate
	if err := proto.Unmarshal(msg.Value, &_msg); err != nil {
		m.log.Warnf("proto.Unmarshal: %v", err)
		m.commitErrorMessage(ctx, msg)
		return
	}

	dto := &pb.UpdatePaymentRequest{
		PaymentEvent:       _msg.GetPaymentEvent(),
		PaymentType:        _msg.GetPaymentType(),
		PaymentCustomerId:  _msg.GetPaymentCustomerId(),
		PaymentMethodId:    _msg.GetPaymentMethodId(),
		PaymentBusinessId:  _msg.GetPaymentBusinessId(),
		PaymentChannel:     _msg.GetPaymentChannel(),
		UpdatedAt:          _msg.GetUpdatedAt(),
		PaymentStatus:      _msg.GetPaymentStatus(),
		PaymentFailureCode: _msg.PaymentFailureCode,
	}

	params := models.NewUpdatePaymentRequestParams(dto)
	if err := m.v.StructCtx(ctx, params); err != nil {
		m.log.Warnf("validate", err)
		m.commitErrorMessage(ctx, msg)
		return
	}

	if err := retry.Do(func() error {
		return m.usecase.Update(ctx, params)
	}, append(retryOption, retry.Context(ctx))...); err != nil {
		m.log.Warnf("m.usecase.Update.err: %v", err)
		m.commitErrorMessage(ctx, msg)
		return
	}

	m.commitMessage(ctx, msg)
}
