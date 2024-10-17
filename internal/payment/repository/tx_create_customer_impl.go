package repository

import (
	"context"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
)

type CreateCustomerTxParams struct {
	CreateCustomer CreateCustomerParams
}

type CreateCustomerTxResult struct {
	Customer *Customer
}

func (r *Store) CreateCustomerTx(ctx context.Context, arg *CreateCustomerTxParams) (CreateCustomerTxResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Store.CreateCustomerTx")
	defer span.Finish()

	var result CreateCustomerTxResult

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := r.execTx(ctx, func(q *Queries) error {
		var err error

		createCustomerPaymentArg := CreateCustomerPaymentParams{
			CustomerName:   arg.CreateCustomer.CustomerName,
			CustomerNumber: arg.CreateCustomer.PhoneNumber.String,
			CustomerUID:    arg.CreateCustomer.Uid,
			CustomerEmail:  arg.CreateCustomer.Email.String,
			ReferenceID:    arg.CreateCustomer.Uid,
		}

		customerRes, err := r.CreateCustomerPayment(ctx, &createCustomerPaymentArg)
		if err != nil {
			return tracing.TraceWithError(span, err)
		}

		if customerRes != nil {
			arg.CreateCustomer.PaymentCustomerID = customerRes.Id
		}

		if customerRes.PhoneNumber.Get() != nil {
			arg.CreateCustomer.PhoneNumber = pgtype.Text{
				String: customerRes.GetPhoneNumber(),
				Valid:  true,
			}
		}

		if customerRes.Email.Get() != nil {
			arg.CreateCustomer.Email = pgtype.Text{
				String: customerRes.GetEmail(),
				Valid:  true,
			}
		}

		result.Customer, err = q.CreateCustomer(ctx, &arg.CreateCustomer)
		if err != nil {
			return tracing.TraceWithError(span, err)
		}

		return err
	})

	return result, err
}
