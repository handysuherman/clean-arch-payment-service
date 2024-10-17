package models

import (
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
)

type PaymentStatusUpdatedTask struct {
	PaymentMethod *repository.PaymentMethod `json:"payment_method"`
}

type CreatePaymentRequest struct {
	CustomerUid             *string `json:"customer_uid,omitempty"`
	CustomerName            string  `json:"customer_name" validate:"required,gt=0"`
	CustomerPhoneNumber     string  `json:"customer_phone_number" validate:"required,lte=17"`
	PaymentDescription      string  `json:"payment_description" validate:"required,gt=0"`
	PaymentReferenceId      string  `json:"payment_reference_id" validate:"required,gt=0"`
	PaymentAmount           float64 `json:"payment_amount" validate:"required,gte=100"`
	PaymentType             string  `json:"payment_type" validate:"required,gt=0"`
	PaymentChannel          string  `json:"payment_channel" validate:"required,gt=0"`
	ExpiryHour              int64   `json:"expiry_hour" validate:"required,gte=72"`
	PaymentSuccessReturnUrl string  `json:"payment_success_return_url" validate:"required,gt=0"`
	PaymentFailureReturnUrl string  `json:"payment_failure_return_url" validate:"required,gt=0"`
	XIdempotencyKey         string  `json:"x_idempotency_key" validate:"required,gt=0"`
}

func NewCreatePaymentRequestParams(arg *pb.CreatePaymentRequest) *CreatePaymentRequest {
	return &CreatePaymentRequest{
		CustomerUid:             arg.CustomerUid,
		CustomerName:            arg.GetCustomerName(),
		CustomerPhoneNumber:     arg.GetCustomerPhoneNumber(),
		PaymentDescription:      arg.GetPaymentDescription(),
		PaymentReferenceId:      arg.GetPaymentReferenceId(),
		PaymentAmount:           arg.GetPaymentAmount(),
		PaymentType:             arg.GetPaymentType(),
		PaymentChannel:          arg.GetPaymentChannel(),
		ExpiryHour:              arg.GetExpiryHour(),
		PaymentSuccessReturnUrl: arg.GetPaymentSuccessReturnUrl(),
		PaymentFailureReturnUrl: arg.GetPaymentFailureReturnUrl(),
		XIdempotencyKey:         arg.GetXIdempotencyKey(),
	}
}

type UpdatePaymentRequest struct {
	PaymentEvent       string     `json:"payment_event,omitempty"`
	PaymentType        string     `json:"payment_type,omitempty"`
	PaymentCustomerId  string     `json:"payment_customer_id" validate:"required,gt=0"`
	PaymentMethodId    string     `json:"payment_method_id" validate:"required,gt=0"`
	PaymentBusinessId  string     `json:"payment_business_id"`
	PaymentChannel     string     `json:"payment_channel,omitempty"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty"`
	PaymentStatus      string     `json:"payment_status" validate:"required,gt=0"`
	PaymentFailureCode *string    `json:"payment_failure_code,omitempty"`
}

func NewUpdatePaymentRequestParams(arg *pb.UpdatePaymentRequest) *UpdatePaymentRequest {
	updatedAt := arg.GetUpdatedAt().AsTime()
	return &UpdatePaymentRequest{
		PaymentEvent:       arg.GetPaymentEvent(),
		PaymentType:        arg.GetPaymentType(),
		PaymentCustomerId:  arg.GetPaymentCustomerId(),
		PaymentMethodId:    arg.GetPaymentMethodId(),
		PaymentBusinessId:  arg.GetPaymentBusinessId(),
		PaymentChannel:     arg.GetPaymentChannel(),
		UpdatedAt:          &updatedAt,
		PaymentStatus:      arg.GetPaymentStatus(),
		PaymentFailureCode: arg.PaymentFailureCode,
	}
}

type GetByIDPaymentRequest struct {
	PaymentCustomerId string `json:"payment_customer_id" validate:"required,gt=0"`
	PaymentMethodId   string `json:"payment_method_id" validate:"required,gt=0"`
}

func NewGetByIDPaymentRequestParams(arg *pb.GetByIDPaymentRequest) *GetByIDPaymentRequest {
	return &GetByIDPaymentRequest{
		PaymentCustomerId: arg.GetPaymentCustomerId(),
		PaymentMethodId:   arg.GetPaymentMethodId(),
	}
}

type GetPaymentChannelRequest struct {
	Amount             float64 `json:"amount" validate:"required,gte=100"`
	PaymentChannelName string  `json:"payment_channel_name" validate:"required,gt=0"`
	PaymentChannelType string  `json:"payment_channel_type" validate:"required,gt=0"`
}

func NewGetPaymentChannelRequestParams(arg *pb.GetPaymentChannelRequest) *GetPaymentChannelRequest {
	return &GetPaymentChannelRequest{
		Amount:             arg.GetAmount(),
		PaymentChannelName: arg.GetPaymentChannelName(),
		PaymentChannelType: arg.GetPaymentChannelType(),
	}
}

type GetPaymentChannelsRequest struct {
	Amount float64 `json:"amount" validate:"required,gte=100"`
}

func NewGetPaymentChannelsRequestParams(arg *pb.GetPaymentChannelsRequest) *GetPaymentChannelsRequest {
	return &GetPaymentChannelsRequest{
		Amount: arg.GetAmount(),
	}
}
