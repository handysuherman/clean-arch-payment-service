package repository

import (
	"context"
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRepoCreateCreatePaymentIdempotencyKey(t *testing.T) {
	createRandomCreatePaymentIdempotencyKey(t)
}

func TestRepoGetCreatePaymentIdempotencyKey(t *testing.T) {
	key, arg := createRandomCreatePaymentIdempotencyKey(t)

	res, err := testStore.GetCreatePaymentIdempotencyKey(context.TODO(), key)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.GetCustomer().GetUid(), arg.GetCustomer().GetUid())
	require.Equal(t, res.GetCustomer().GetCustomerAppId(), arg.GetCustomer().GetCustomerAppId())
	require.Equal(t, res.GetCustomer().GetPaymentCustomerId(), arg.GetCustomer().GetPaymentCustomerId())
	require.Equal(t, res.GetCustomer().GetCustomerName(), arg.GetCustomer().GetCustomerName())
	require.Equal(t, res.GetCustomer().GetEmail(), arg.GetCustomer().GetEmail())
	require.Equal(t, res.GetCustomer().GetPhoneNumber(), arg.GetCustomer().GetPhoneNumber())

	require.Equal(t, res.GetPaymentMethod().GetUid(), arg.GetPaymentMethod().GetUid())
	require.Equal(t, res.GetPaymentMethod().GetPaymentMethodId(), arg.GetPaymentMethod().GetPaymentMethodId())
	require.Equal(t, res.GetPaymentMethod().GetPaymentRequestId(), arg.GetPaymentMethod().GetPaymentRequestId())
	require.Equal(t, res.GetPaymentMethod().GetPaymentReferenceId(), arg.GetPaymentMethod().GetPaymentReferenceId())
	require.Equal(t, res.GetPaymentMethod().GetPaymentCustomerId(), arg.GetPaymentMethod().GetPaymentCustomerId())
	require.Equal(t, res.GetPaymentMethod().GetPaymentType(), arg.GetPaymentMethod().GetPaymentType())
	require.Equal(t, res.GetPaymentMethod().GetPaymentStatus(), arg.GetPaymentMethod().GetPaymentStatus())
	require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), arg.GetPaymentMethod().GetPaymentReusability())
	require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), arg.GetPaymentMethod().GetPaymentChannel())
	require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), arg.GetPaymentMethod().GetPaymentAmount())
	require.Equal(t, res.GetPaymentMethod().GetPaymentQrCode(), arg.GetPaymentMethod().GetPaymentQrCode())
	require.Equal(t, res.GetPaymentMethod().GetPaymentVirtualAccountNumber(), arg.GetPaymentMethod().GetPaymentVirtualAccountNumber())
	require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), arg.GetPaymentMethod().GetPaymentUrl())
}

func TestRepoDeleteCreatePaymentIdempotencyKey(t *testing.T) {
	key, arg := createRandomCreatePaymentIdempotencyKey(t)

	res, err := testStore.GetCreatePaymentIdempotencyKey(context.TODO(), key)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.GetCustomer().GetUid(), arg.GetCustomer().GetUid())
	require.Equal(t, res.GetCustomer().GetCustomerAppId(), arg.GetCustomer().GetCustomerAppId())
	require.Equal(t, res.GetCustomer().GetPaymentCustomerId(), arg.GetCustomer().GetPaymentCustomerId())
	require.Equal(t, res.GetCustomer().GetCustomerName(), arg.GetCustomer().GetCustomerName())
	require.Equal(t, res.GetCustomer().GetEmail(), arg.GetCustomer().GetEmail())
	require.Equal(t, res.GetCustomer().GetPhoneNumber(), arg.GetCustomer().GetPhoneNumber())

	require.Equal(t, res.GetPaymentMethod().GetUid(), arg.GetPaymentMethod().GetUid())
	require.Equal(t, res.GetPaymentMethod().GetPaymentMethodId(), arg.GetPaymentMethod().GetPaymentMethodId())
	require.Equal(t, res.GetPaymentMethod().GetPaymentRequestId(), arg.GetPaymentMethod().GetPaymentRequestId())
	require.Equal(t, res.GetPaymentMethod().GetPaymentReferenceId(), arg.GetPaymentMethod().GetPaymentReferenceId())
	require.Equal(t, res.GetPaymentMethod().GetPaymentCustomerId(), arg.GetPaymentMethod().GetPaymentCustomerId())
	require.Equal(t, res.GetPaymentMethod().GetPaymentType(), arg.GetPaymentMethod().GetPaymentType())
	require.Equal(t, res.GetPaymentMethod().GetPaymentStatus(), arg.GetPaymentMethod().GetPaymentStatus())
	require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), arg.GetPaymentMethod().GetPaymentReusability())
	require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), arg.GetPaymentMethod().GetPaymentChannel())
	require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), arg.GetPaymentMethod().GetPaymentAmount())
	require.Equal(t, res.GetPaymentMethod().GetPaymentQrCode(), arg.GetPaymentMethod().GetPaymentQrCode())
	require.Equal(t, res.GetPaymentMethod().GetPaymentVirtualAccountNumber(), arg.GetPaymentMethod().GetPaymentVirtualAccountNumber())
	require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), arg.GetPaymentMethod().GetPaymentUrl())

	testStore.DeleteCreatePaymentIdempotencyKey(context.TODO(), key)

	res2, err := testStore.GetCreatePaymentIdempotencyKey(context.TODO(), key)
	require.Error(t, err)
	require.Empty(t, res2)
}

func createRandomCreatePaymentIdempotencyKey(t *testing.T) (string, *pb.CreatePaymentResponse) {
	customer := createRandomCustomer(t)
	paymentMethod := createRandomPaymentMethod(t)

	idmptKey, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, idmptKey)

	amount, _ := paymentMethod.PaymentAmount.Float64()
	resp := &pb.CreatePaymentResponse{
		Customer: &pb.Customer{
			Uid:               customer.Uid,
			CustomerAppId:     customer.CustomerAppID,
			PaymentCustomerId: customer.PaymentCustomerID,
			CustomerName:      customer.CustomerName,
			CreatedAt:         timestamppb.New(customer.CreatedAt.Time),
			Email:             &customer.Email.String,
			PhoneNumber:       &customer.PhoneNumber.String,
		},
		PaymentMethod: &pb.PaymentMethod{
			Uid:                         paymentMethod.Uid,
			PaymentMethodId:             paymentMethod.PaymentMethodID,
			PaymentRequestId:            &paymentMethod.PaymentRequestID.String,
			PaymentReferenceId:          paymentMethod.PaymentReferenceID,
			PaymentBusinessId:           paymentMethod.PaymentBusinessID,
			PaymentCustomerId:           paymentMethod.PaymentCustomerID,
			PaymentType:                 paymentMethod.PaymentType,
			PaymentStatus:               paymentMethod.PaymentStatus,
			PaymentReusability:          paymentMethod.PaymentReusability,
			PaymentChannel:              paymentMethod.PaymentChannel,
			PaymentAmount:               amount,
			PaymentQrCode:               &paymentMethod.PaymentQrCode.String,
			PaymentVirtualAccountNumber: &paymentMethod.PaymentVirtualAccountNumber.String,
			PaymentUrl:                  &paymentMethod.PaymentUrl.String,
			PaymentDescription:          paymentMethod.PaymentDescription,
			CreatedAt:                   timestamppb.New(paymentMethod.CreatedAt.Time),
			UpdatedAt:                   timestamppb.New(paymentMethod.UpdatedAt.Time),
			ExpiresAt:                   timestamppb.New(paymentMethod.ExpiresAt.Time),
		},
	}

	testStore.PutCreatePaymentIdempotencyKey(context.TODO(), idmptKey.String(), resp)

	return idmptKey.String(), resp
}
