package usecase

import (
	"context"
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/mapper"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/opentracing/opentracing-go"
	"github.com/shopspring/decimal"
	"github.com/xendit/xendit-go/v5/payment_method"
)

func (u *usecaseImpl) GetAvailableChannel(ctx context.Context, arg *models.GetPaymentChannelRequest) (*pb.GetPaymentChannelResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UsecaseImpl.GetAvailableChannel")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(ctx, u.cfg.Services.Internal.OperationTimeout)
	defer cancel()

	_, err := payment_method.NewPaymentMethodTypeFromValue(arg.PaymentChannelType)
	if err != nil {
		return nil, u.errorResponse(
			span,
			fmt.Sprintf("%s: %v", "payment_method.NewPaymentMethodTypeFromValue.err", arg.PaymentChannelType),
			err,
		)
	}

	getArg := &repository.GetAvailablePaymentChannelParams{
		PcType:    arg.PaymentChannelType,
		Pcname:    arg.PaymentChannelName,
		MinAmount: decimal.NewFromFloat(arg.Amount),
	}

	channel, err := u.repo.GetAvailablePaymentChannel(ctx, getArg)
	if err != nil {
		return nil, u.errorResponse(
			span,
			"u.repo.GetAvailablePaymentChannel.err",
			err,
		)
	}

	return &pb.GetPaymentChannelResponse{
		PaymentChannel: mapper.PaymentChannelToDto(channel),
	}, nil
}
