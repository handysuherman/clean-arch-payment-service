package usecase

import (
	"context"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/mapper"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/opentracing/opentracing-go"
)

func (u *usecaseImpl) GetByID(ctx context.Context, arg *models.GetByIDPaymentRequest) (*pb.GetByIDPaymentResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UsecaseImpl.GetByID")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(ctx, u.cfg.Services.Internal.OperationTimeout)
	defer cancel()

	if payload, err := u.repo.GetCache(ctx, arg.PaymentCustomerId, arg.PaymentMethodId); err == nil && payload != nil {
		return &pb.GetByIDPaymentResponse{
			PaymentMethod: mapper.PaymentToDto(payload),
		}, nil
	}

	getPaymentArg := repository.GetPaymentMethodCustomerParams{
		PaymentMethodID:   arg.PaymentMethodId,
		PaymentCustomerID: arg.PaymentCustomerId,
	}

	payment, err := u.repo.GetPaymentMethodCustomer(ctx, &getPaymentArg)
	if err != nil {
		return nil, u.errorResponse(span, "u.repo.GetPaymentMethodCustomer", err)
	}

	u.repo.PutCache(ctx, payment)
	return &pb.GetByIDPaymentResponse{
		PaymentMethod: mapper.PaymentToDto(payment),
	}, nil
}
