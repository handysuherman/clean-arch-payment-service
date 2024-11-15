// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package repository

import (
	"context"

	"github.com/shopspring/decimal"
)

type Querier interface {
	CreateCustomer(ctx context.Context, arg *CreateCustomerParams) (*Customer, error)
	CreatePaymentChannel(ctx context.Context, arg *CreatePaymentChannelParams) (*PaymentChannel, error)
	CreatePaymentMethod(ctx context.Context, arg *CreatePaymentMethodParams) (*PaymentMethod, error)
	CreatePaymentReusability(ctx context.Context, prname string) (string, error)
	CreatePaymentStatus(ctx context.Context, psname string) (string, error)
	CreatePaymentType(ctx context.Context, ptname string) (string, error)
	GetAvailablePaymentChannel(ctx context.Context, arg *GetAvailablePaymentChannelParams) (*PaymentChannel, error)
	GetAvailablePaymentChannels(ctx context.Context, minAmount decimal.Decimal) ([]*PaymentChannel, error)
	GetCustomerByCustomerAppID(ctx context.Context, customerAppID string) (*Customer, error)
	GetCustomerByPaymentCustomerID(ctx context.Context, paymentCustomerID string) (*Customer, error)
	GetPaymentChannelByID(ctx context.Context, uid string) (*PaymentChannel, error)
	GetPaymentChannelByName(ctx context.Context, pcname string) (*PaymentChannel, error)
	GetPaymentMethodByPaymentMethodID(ctx context.Context, paymentMethodID string) (*PaymentMethod, error)
	GetPaymentMethodByReferenceID(ctx context.Context, paymentReferenceID string) (*PaymentMethod, error)
	GetPaymentMethodCustomer(ctx context.Context, arg *GetPaymentMethodCustomerParams) (*PaymentMethod, error)
	GetPaymentReusabilityByName(ctx context.Context, prname string) (string, error)
	GetPaymentStatusByName(ctx context.Context, psname string) (string, error)
	GetPaymentTypeByName(ctx context.Context, ptname string) (string, error)
	UpdatePaymentMethodCustomer(ctx context.Context, arg *UpdatePaymentMethodCustomerParams) (*PaymentMethod, error)
}

var _ Querier = (*Queries)(nil)
