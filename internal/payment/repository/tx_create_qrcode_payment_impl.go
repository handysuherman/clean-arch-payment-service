package repository

import (
	"context"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
	"github.com/shopspring/decimal"
)

type CreateQrCodePaymentTxParams struct {
	CreateQrCodePayment CreateQrCodePaymentParams
}

type CreateQrCodePaymentTxResult struct {
	Payment *PaymentMethod
}

func (r *Store) CreateQrCodePaymentTx(ctx context.Context, arg *CreateQrCodePaymentTxParams) (CreateQrCodePaymentTxResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Store.CreateQrCodePaymentTx")
	defer span.Finish()

	var result CreateQrCodePaymentTxResult

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := r.execTx(ctx, func(q *Queries) error {
		var err error

		paymentReqResult, err := r.CreateQrCodePayment(ctx, &arg.CreateQrCodePayment)
		if err != nil {
			return tracing.TraceWithError(span, err)
		}

		paymentMethodInternalID, err := helper.GenerateULID()
		if err != nil {
			return tracing.TraceWithError(span, err)
		}

		createPaymentMethodArg := CreatePaymentMethodParams{
			Uid:                paymentMethodInternalID.String(),
			PaymentMethodID:    paymentReqResult.Id,
			PaymentBusinessID:  paymentReqResult.GetBusinessId(),
			PaymentReferenceID: paymentReqResult.GetReferenceId(),
			PaymentType:        paymentReqResult.GetType().String(),
			PaymentStatus:      paymentReqResult.GetStatus().String(),
			PaymentReusability: paymentReqResult.GetReusability().String(),
			PaymentCustomerID:  paymentReqResult.GetCustomerId(),
			PaymentDescription: paymentReqResult.GetDescription(),
			PaymentChannel:     paymentReqResult.GetQrCode().ChannelCode.Get().String(),
			PaymentAmount:      decimal.NewFromFloat(*paymentReqResult.GetQrCode().Amount.Get()),
			PaymentQrCode: pgtype.Text{
				String: paymentReqResult.GetQrCode().ChannelProperties.Get().GetQrString(),
				Valid:  true,
			},
			CreatedAt: pgtype.Timestamptz{
				Time:  paymentReqResult.GetCreated(),
				Valid: true,
			},
			ExpiresAt: pgtype.Timestamptz{
				Time:  arg.CreateQrCodePayment.Expiry,
				Valid: true,
			},
		}

		result.Payment, err = q.CreatePaymentMethod(ctx, &createPaymentMethodArg)
		if err != nil {
			return tracing.TraceWithError(span, err)
		}

		return err
	})

	return result, err
}
