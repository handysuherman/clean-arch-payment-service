package repository

import (
	"context"
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/xendit/xendit-go/v5"
	"github.com/xendit/xendit-go/v5/customer"
	"github.com/xendit/xendit-go/v5/payment_method"
	"github.com/xendit/xendit-go/v5/payment_request"
)

type PaymentProvider interface {
	CreateCustomerPayment(ctx context.Context, arg *CreateCustomerPaymentParams) (*customer.Customer, error)
	GetCustomerPaymentByID(ctx context.Context, arg string) (*customer.Customer, error)

	CreateEwalletPayment(ctx context.Context, arg *CreateEwalletPaymentParams) (*payment_request.PaymentRequest, error)
	GetEwalletPaymentRequestByID(ctx context.Context, arg string) (*payment_request.PaymentRequest, error)

	CreateQrCodePayment(ctx context.Context, arg *CreateQrCodePaymentParams) (*payment_method.PaymentMethod, error)
	GetQrCodePaymentByID(ctx context.Context, arg string) (*payment_method.PaymentMethod, error)

	CreateVirtualAccountBankPayment(ctx context.Context, arg *CreateVirtualAccountBankPaymentParams) (*payment_method.PaymentMethod, error)
	GetVirtualAccountBankPaymentByID(ctx context.Context, arg string) (*payment_method.PaymentMethod, error)
}

type PaymentProviderImpl struct {
	log          logger.Logger
	cfg          *config.App
	xenditClient *xendit.APIClient
}

func NewPaymentProviderImpl(log logger.Logger, cfg *config.App, xenditClient *xendit.APIClient) *PaymentProviderImpl {
	return &PaymentProviderImpl{
		log:          log.WithPrefix(fmt.Sprintf("%s-%s", "payment-gateway-provider", constants.Repository)),
		cfg:          cfg,
		xenditClient: xenditClient,
	}
}
