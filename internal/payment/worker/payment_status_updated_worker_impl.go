package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	messages "github.com/handysuherman/clean-arch-payment-service/internal/proto/kafka"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (w *Worker) PaymentStatusUpdated(ctx context.Context, task *models.PaymentStatusUpdatedTask) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Worker.PaymentStatusUpdated")
	defer span.Finish()

	amount, ok := task.PaymentMethod.PaymentAmount.Float64()
	if !ok {
		return w.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "task.PaymentMethod.PaymentAmount.Float64.err", task.PaymentMethod.PaymentAmount),
			errors.New("rounded demical error"),
		)
	}

	arg := messages.KafkaPaymentStatusUpdated{
		Uid:                         task.PaymentMethod.Uid,
		PaymentMethodId:             task.PaymentMethod.PaymentMethodID,
		PaymentRequestId:            &task.PaymentMethod.PaymentRequestID.String,
		PaymentReferenceId:          task.PaymentMethod.PaymentReferenceID,
		PaymentBusinessId:           task.PaymentMethod.PaymentBusinessID,
		PaymentCustomerId:           task.PaymentMethod.PaymentCustomerID,
		PaymentType:                 task.PaymentMethod.PaymentType,
		PaymentStatus:               task.PaymentMethod.PaymentStatus,
		PaymentReusability:          task.PaymentMethod.PaymentReusability,
		PaymentChannel:              task.PaymentMethod.PaymentChannel,
		PaymentAmount:               amount,
		PaymentQrCode:               &task.PaymentMethod.PaymentQrCode.String,
		PaymentVirtualAccountNumber: &task.PaymentMethod.PaymentVirtualAccountNumber.String,
		PaymentUrl:                  &task.PaymentMethod.PaymentUrl.String,
		PaymentDescription:          task.PaymentMethod.PaymentDescription,
		CreatedAt:                   timestamppb.New(task.PaymentMethod.CreatedAt.Time),
		UpdatedAt:                   timestamppb.New(task.PaymentMethod.UpdatedAt.Time),
		ExpiresAt:                   timestamppb.New(task.PaymentMethod.ExpiresAt.Time),
		PaidAt:                      timestamppb.New(task.PaymentMethod.PaidAt.Time),
	}

	protoMsg, err := proto.Marshal(&arg)
	if err != nil {
		return w.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "proto.Marshal.err", err),
			err,
		)
	}

	message := kafka.Message{
		Topic:   helper.StringBuilder(w.cfg.Services.Internal.ID, "_", w.cfg.Brokers.Kafka.Topics.PaymentStatusUpdated.TopicName),
		Value:   protoMsg,
		Time:    time.Now().UTC(),
		Headers: tracing.GetKafkaTracingHeadersFromSpanCtx(span.Context()),
	}

	err = w.distributor.PublishMessage(ctx, message)
	if err != nil {
		return w.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "w.distributor.PublishMessage.err", err),
			err,
		)
	}

	return nil
}
