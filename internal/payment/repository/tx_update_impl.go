package repository

import (
	"context"
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/payment"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/opentracing/opentracing-go"
)

type UpdateTxParams struct {
	UpdateParams UpdatePaymentMethodCustomerParams
}

type UpdateTxResult struct {
	Payment *PaymentMethod
}

func (r *Store) UpdateTx(ctx context.Context, arg *UpdateTxParams) (UpdateTxResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Store.UpdateTx")
	defer span.Finish()

	var result UpdateTxResult

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := r.execTx(ctx, func(q *Queries) error {
		var err error

		getPaymentArg := GetPaymentMethodCustomerParams{
			PaymentMethodID:   arg.UpdateParams.PaymentMethodID,
			PaymentCustomerID: arg.UpdateParams.PaymentCustomerID,
		}

		result.Payment, err = q.GetPaymentMethodCustomer(ctx, &getPaymentArg)
		if err != nil {
			return tracing.TraceWithError(span, fmt.Errorf("q.GetPaymentMethodCustomer.err: %v", err))
		}

		if result.Payment.PaymentStatus == payment.STATUS_SUCCEEDED || result.Payment.PaymentStatus == payment.STATUS_FAILED {
			return nil
		}

		result.Payment, err = q.UpdatePaymentMethodCustomer(ctx, &arg.UpdateParams)
		if err != nil {
			return tracing.TraceWithError(span, fmt.Errorf("q.UpdatePaymentMethodCustomer.err: %v", err))
		}

		return err
	})

	return result, err
}
