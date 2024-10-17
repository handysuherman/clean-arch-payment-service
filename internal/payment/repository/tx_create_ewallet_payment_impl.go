package repository

import (
	"context"
	"errors"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opentracing/opentracing-go"
	"github.com/shopspring/decimal"
	"github.com/xendit/xendit-go/v5/payment_method"
	"github.com/xendit/xendit-go/v5/payment_request"
)

type CreateEwalletPaymentTxParams struct {
	CreateEwalletPayment CreateEwalletPaymentParams
}

type CreateEwalletPaymentTxResult struct {
	Payment *PaymentMethod
}

func (r *Store) CreateEwalletPaymentTx(ctx context.Context, arg *CreateEwalletPaymentTxParams) (CreateEwalletPaymentTxResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Store.CreateEwalletPaymentTx")
	defer span.Finish()

	var result CreateEwalletPaymentTxResult

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := r.execTx(ctx, func(q *Queries) error {
		var err error

		paymentReqResult, err := r.CreateEwalletPayment(ctx, &arg.CreateEwalletPayment)
		if err != nil {
			return tracing.TraceWithError(span, err)
		}

		paymentMethodInternalID, err := helper.GenerateULID()
		if err != nil {
			return tracing.TraceWithError(span, err)
		}

		eWalletURLPayment, err := getURl(arg.CreateEwalletPayment.ChannelCode, paymentReqResult.GetActions())
		if err != nil {
			return tracing.TraceWithError(span, err)
		}

		paymentCreated, err := time.Parse(constants.TZ, paymentReqResult.GetCreated())
		if err != nil {
			return tracing.TraceWithError(span, err)
		}

		createPaymentMethodArg := CreatePaymentMethodParams{
			Uid:             paymentMethodInternalID.String(),
			PaymentMethodID: paymentReqResult.PaymentMethod.Id,
			PaymentRequestID: pgtype.Text{
				String: paymentReqResult.Id,
				Valid:  true,
			},
			PaymentReferenceID: *paymentReqResult.PaymentMethod.ReferenceId,
			PaymentCustomerID:  paymentReqResult.GetCustomerId(),
			PaymentBusinessID:  paymentReqResult.GetBusinessId(),
			PaymentType:        paymentReqResult.PaymentMethod.Type.String(),
			PaymentStatus:      paymentReqResult.PaymentMethod.Status.String(),
			PaymentReusability: paymentReqResult.PaymentMethod.Reusability.String(),
			PaymentChannel:     string(paymentReqResult.PaymentMethod.Ewallet.Get().GetChannelCode()),
			PaymentAmount:      decimal.NewFromFloat(paymentReqResult.GetAmount()),
			PaymentDescription: paymentReqResult.GetDescription(),
			PaymentUrl: pgtype.Text{
				String: eWalletURLPayment,
				Valid:  true,
			},
			CreatedAt: pgtype.Timestamptz{
				Time:  paymentCreated,
				Valid: true,
			},
			ExpiresAt: pgtype.Timestamptz{
				Time:  arg.CreateEwalletPayment.Expiry,
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

func getURl(channelType payment_method.EWalletChannelCode, actions []payment_request.PaymentRequestAction) (string, error) {
	switch channelType {
	case payment_method.EWALLETCHANNELCODE_SHOPEEPAY:
		action, err := findUrlType(actions, "DEEPLINK")
		if err != nil {
			return "", err
		}

		return action.GetUrl(), nil
	case payment_method.EWALLETCHANNELCODE_OVO:
		return "", nil
	default:
		action, err := findUrlType(actions, "MOBILE")
		if err != nil {
			return "", err
		}

		return action.GetUrl(), nil
	}
}

func findUrlType(actions []payment_request.PaymentRequestAction, urlType string) (*payment_request.PaymentRequestAction, error) {
	for _, action := range actions {
		if action.UrlType == urlType {
			return &action, nil
		}
	}
	return nil, errors.New("no type found in ewallet action response")
}
