package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/serializer"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/xendit/xendit-go/v5/payment_method"
	"github.com/xendit/xendit-go/v5/payment_request"
)

var (
	ID_COUNTRY = "ID"
)

type CreateEwalletPaymentParams struct {
	CustomerName      string    `json:"customerName"`
	CustomerPaymentID string    `json:"customerPaymentID"`
	CustomerNumber    string    `json:"customerNumber"`
	Description       string    `json:"description"`
	ReferenceID       string    `json:"referenceID"`
	Amount            float64   `json:"amount"`
	Expiry            time.Time `json:"expiry"`
	ChannelCode       payment_method.EWalletChannelCode
	SuccessReturnURL  string `json:"successReturnURL"`
	FailureReturnURL  string `json:"failureReturnURL"`
}

func (p *PaymentProviderImpl) CreateEwalletPayment(ctx context.Context, arg *CreateEwalletPaymentParams) (*payment_request.PaymentRequest, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PaymentProviderImpl.CreateEwalletPayment")
	defer span.Finish()

	paymentMethod, err := p.createEwalletMethod(ctx, span, arg)
	if err != nil {
		return nil, err
	}

	paymentRequestParameters := *payment_request.NewPaymentRequestParameters(payment_request.PAYMENTREQUESTCURRENCY_IDR)
	paymentRequestParameters.CustomerId = *payment_request.NewNullableString(&arg.CustomerPaymentID)
	paymentRequestParameters.PaymentMethodId = &paymentMethod.Id
	paymentRequestParameters.Description = *payment_request.NewNullableString(&arg.Description)
	paymentRequestParameters.ReferenceId = &arg.ReferenceID
	paymentRequestParameters.Amount = &arg.Amount

	requestKey, err := helper.GenerateULID()
	if err != nil {
		return nil, tracing.TraceWithError(span, err)
	}

	idempotencyKey := fmt.Sprintf("pr-%s", requestKey.String())
	resp, _, errs := p.xenditClient.PaymentRequestApi.CreatePaymentRequest(ctx).
		IdempotencyKey(idempotencyKey).
		PaymentRequestParameters(paymentRequestParameters).
		Execute()
	if errs != nil {
		fullErr, _ := serializer.Marshal(errs.FullError())
		return nil, errorResponse(
			span,
			errors.New(err.Error()),
			errors.New(string(fullErr)),
			"unable to create payment request",
			"p.xenditClient.PaymentRequestApi.CreatePaymentRequest.err",
		)
	}

	return resp, nil
}

func (p *PaymentProviderImpl) GetEwalletPaymentRequestByID(ctx context.Context, arg string) (*payment_request.PaymentRequest, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PaymentProviderImpl.GetEwalletPaymentRequestByID")
	defer span.Finish()

	resp, _, err := p.xenditClient.PaymentRequestApi.GetPaymentRequestByID(ctx, arg).Execute()
	if err != nil {
		fullErr, _ := serializer.Marshal(err.FullError())
		return nil, errorResponse(
			span,
			errors.New(err.Error()),
			errors.New(string(fullErr)),
			"unable to get payment request id",
			"p.xenditClient.PaymentRequestApi.GetPaymentRequestByID.err",
		)
	}

	return resp, nil
}

func (p *PaymentProviderImpl) createEwalletMethod(ctx context.Context, span opentracing.Span, arg *CreateEwalletPaymentParams) (*payment_method.PaymentMethod, error) {
	paymentMethodParameters := *payment_method.NewPaymentMethodParameters(
		payment_method.PaymentMethodType(payment_method.PAYMENTMETHODTYPE_EWALLET),
		payment_method.PAYMENTMETHODREUSABILITY_ONE_TIME_USE,
	)
	paymentMethodParameters.CustomerId = *payment_method.NewNullableString(&arg.CustomerPaymentID)
	paymentMethodParameters.Country = *payment_method.NewNullableString(&ID_COUNTRY)
	paymentMethodParameters.ReferenceId = &arg.ReferenceID

	eWalletArg := *payment_method.NewEWalletParameters(arg.ChannelCode)
	eWalletArg.Account = payment_method.NewEWalletAccountWithDefaults()
	eWalletArg.Account.Name = *payment_method.NewNullableString(&arg.CustomerName)

	eWalletArg.ChannelProperties = &payment_method.EWalletChannelProperties{
		SuccessReturnUrl: &arg.SuccessReturnURL,
		FailureReturnUrl: &arg.FailureReturnURL,
	}
	eWalletArg.ChannelProperties.MobileNumber = &arg.CustomerNumber

	paymentMethodParameters.SetEwallet(eWalletArg)

	resp, _, err := p.xenditClient.PaymentMethodApi.CreatePaymentMethod(ctx).
		PaymentMethodParameters(paymentMethodParameters).
		Execute()
	if err != nil {
		fullErr, _ := serializer.Marshal(err.FullError())
		return nil, errorResponse(
			span,
			errors.New(err.Error()),
			errors.New(string(fullErr)),
			"unable to create payment method",
			"p.xenditClient.PaymentMethodApi.CreatePaymentMethod.err",
		)
	}
	span.LogFields(log.Object("create_payment_resp", resp))

	return resp, nil
}
