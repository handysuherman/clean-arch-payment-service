package repository

import (
	"context"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
	"github.com/shopspring/decimal"
)

type CreateVirtualAccountBankPaymentTxParams struct {
	CreateVirtualAccountBankPayment CreateVirtualAccountBankPaymentParams
}

type CreateVirtualAccountBankPaymentTxResult struct {
	Payment *PaymentMethod
}

func (r *Store) CreateVirtualAccountBankPaymentTx(ctx context.Context, arg *CreateVirtualAccountBankPaymentTxParams) (CreateVirtualAccountBankPaymentTxResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Store.CreateVirtualAccountBankPaymentTx")
	defer span.Finish()

	var result CreateVirtualAccountBankPaymentTxResult

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := r.execTx(ctx, func(q *Queries) error {
		var err error

		paymentReqResult, err := r.CreateVirtualAccountBankPayment(ctx, &arg.CreateVirtualAccountBankPayment)
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
			PaymentReferenceID: paymentReqResult.GetReferenceId(),
			PaymentCustomerID:  paymentReqResult.GetCustomerId(),
			PaymentBusinessID:  paymentReqResult.GetBusinessId(),
			PaymentType:        paymentReqResult.GetType().String(),
			PaymentStatus:      paymentReqResult.GetStatus().String(),
			PaymentReusability: paymentReqResult.GetReusability().String(),
			PaymentChannel:     paymentReqResult.GetVirtualAccount().ChannelCode.String(),
			PaymentDescription: paymentReqResult.GetDescription(),
			CreatedAt: pgtype.Timestamptz{
				Time:  paymentReqResult.GetCreated(),
				Valid: true,
			},
			ExpiresAt: pgtype.Timestamptz{
				Time:  *paymentReqResult.GetVirtualAccount().ChannelProperties.ExpiresAt,
				Valid: true,
			},
			PaymentAmount: decimal.NewFromFloat(*paymentReqResult.GetVirtualAccount().Amount.Get()),
			PaymentVirtualAccountNumber: pgtype.Text{
				String: *paymentReqResult.GetVirtualAccount().ChannelProperties.VirtualAccountNumber,
				Valid:  true,
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
