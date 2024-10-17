package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/serializer"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/xendit/xendit-go/v5/customer"
)

type CreateCustomerPaymentParams struct {
	CustomerName   string `json:"customerName"`
	CustomerNumber string `json:"customerNumber"`
	CustomerUID    string `json:"customerUID"`
	CustomerEmail  string `json:"customerEmail"`
	ReferenceID    string `json:"referenceID"`
}

func (p *PaymentProviderImpl) CreateCustomerPayment(ctx context.Context, arg *CreateCustomerPaymentParams) (*customer.Customer, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PaymentProviderImpl.CreateCustomerPayment")
	defer span.Finish()
	number := "+" + arg.CustomerNumber

	customerRequest := *customer.NewCustomerRequest(arg.CustomerUID)
	customerRequest.MobileNumber = &number
	customerRequest.PhoneNumber = &number
	customerRequest.ReferenceId = arg.ReferenceID
	customerRequest.IndividualDetail = *customer.NewNullableIndividualDetail(&customer.IndividualDetail{
		GivenNames: &arg.CustomerName,
	})

	requestKey, err := helper.GenerateULID()
	if err != nil {
		return nil, tracing.TraceWithError(span, err)
	}

	idempotencyKey := fmt.Sprintf("customer-%s", requestKey.String())
	resp, _, errs := p.xenditClient.CustomerApi.CreateCustomer(ctx).
		IdempotencyKey(idempotencyKey).
		CustomerRequest(customerRequest).
		Execute()
	if errs != nil {
		fullErr, _ := serializer.Marshal(errs.FullError())
		return nil, errorResponse(
			span,
			errors.New(errs.Error()),
			errors.New(string(fullErr)),
			"unable to create customer",
			"p.xenditClient.CustomerApi.CreateCustomer.err",
		)
	}

	return resp, nil
}

func (p *PaymentProviderImpl) GetCustomerPaymentByID(ctx context.Context, arg string) (*customer.Customer, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PaymentProviderImpl.GetCustomerPaymentByID")
	defer span.Finish()

	resp, _, err := p.xenditClient.CustomerApi.GetCustomer(ctx, arg).Execute()
	if err != nil {
		fullErr, _ := serializer.Marshal(err.FullError())
		return nil, errorResponse(
			span,
			errors.New(err.Error()),
			errors.New(string(fullErr)),
			"unable to get customer payment by id",
			"p.xenditClient.CustomerApi.GetCustomer.err",
		)
	}

	return resp, nil
}
