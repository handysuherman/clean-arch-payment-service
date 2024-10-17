package repository

import (
	"context"
	"errors"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/serializer"
	"github.com/opentracing/opentracing-go"
	"github.com/xendit/xendit-go/v5/payment_method"
)

type CreateVirtualAccountBankPaymentParams struct {
	CustomerName      string    `json:"customerName"`
	CustomerPaymentID string    `json:"customerPaymentID"`
	CustomerNumber    string    `json:"customerNumber"`
	Description       string    `json:"description"`
	ReferenceID       string    `json:"referenceID"`
	Amount            float64   `json:"amount"`
	Expiry            time.Time `json:"expiry"`
	ChannelCode       payment_method.VirtualAccountChannelCode
	SuccessReturnURL  string `json:"successReturnURL"`
	FailureReturnURL  string `json:"failureReturnURL"`
}

func (p *PaymentProviderImpl) CreateVirtualAccountBankPayment(ctx context.Context, arg *CreateVirtualAccountBankPaymentParams) (*payment_method.PaymentMethod, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PaymentProviderImpl.CreateVirtualAccountBankPayment")
	defer span.Finish()

	paymentMethodParameters := *payment_method.NewPaymentMethodParameters(
		payment_method.PaymentMethodType(payment_method.PAYMENTMETHODTYPE_VIRTUAL_ACCOUNT),
		payment_method.PAYMENTMETHODREUSABILITY_ONE_TIME_USE,
	)
	paymentMethodParameters.CustomerId = *payment_method.NewNullableString(&arg.CustomerPaymentID)
	paymentMethodParameters.Description = *payment_method.NewNullableString(&arg.Description)
	paymentMethodParameters.ReferenceId = &arg.ReferenceID

	method := payment_method.VirtualAccountParameters{
		ChannelCode:       arg.ChannelCode,
		ChannelProperties: *payment_method.NewVirtualAccountChannelPropertiesWithDefaults(),
	}
	method.ChannelProperties.CustomerName = &arg.CustomerName
	method.ChannelProperties.ExpiresAt = &arg.Expiry
	method.Amount = *payment_method.NewNullableFloat64(&arg.Amount)

	paymentMethodParameters.SetVirtualAccount(method)

	resp, _, err := p.xenditClient.PaymentMethodApi.CreatePaymentMethod(ctx).
		PaymentMethodParameters(paymentMethodParameters).
		Execute()
	if err != nil {
		fullErr, _ := serializer.Marshal(err.FullError())
		return nil, errorResponse(
			span,
			errors.New(err.Error()),
			errors.New(string(fullErr)),
			"unable to create payment",
			"p.xenditClient.PaymentMethodApi.CreatePaymentMethod.err",
		)
	}

	return resp, nil
}

func (p *PaymentProviderImpl) GetVirtualAccountBankPaymentByID(ctx context.Context, arg string) (*payment_method.PaymentMethod, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PaymentProviderImpl.CreateVirtualAccountBankPayment")
	defer span.Finish()

	resp, _, err := p.xenditClient.PaymentMethodApi.GetPaymentMethodByID(ctx, arg).Execute()
	if err != nil {
		fullErr, _ := serializer.Marshal(err.FullError())
		return nil, errorResponse(
			span,
			err,
			errors.New(string(fullErr)),
			"unable to get payment method",
			"p.xenditClient.PaymentMethodApi.GetPaymentMethodByID",
		)
	}

	return resp, nil
}
