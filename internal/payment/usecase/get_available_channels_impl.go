package usecase

import (
	"context"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/mapper"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/opentracing/opentracing-go"
	"github.com/shopspring/decimal"
)

func (u *usecaseImpl) GetAvailableChannels(ctx context.Context, arg *models.GetPaymentChannelsRequest) (*pb.GetPaymentChannelsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UsecaseImpl.GetAvailablePaymentChannels")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(ctx, u.cfg.Services.Internal.OperationTimeout)
	defer cancel()

	res, err := u.repo.GetAvailablePaymentChannels(ctx, decimal.NewFromFloat(arg.Amount))
	if err != nil {
		return nil, u.errorResponse(
			span,
			"u.repo.GetAvailablePaymentChannels.err",
			err,
		)
	}

	return &pb.GetPaymentChannelsResponse{
		List: mapper.PaymentChannelsToDto(res),
	}, nil
}
