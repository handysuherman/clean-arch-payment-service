package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoCreatePaymentCache(t *testing.T) {
	createRandomPaymentCache(t)
}

func TestRepoGetPaymentCache(t *testing.T) {
	arg := createRandomPaymentCache(t)

	res, err := testStore.GetCache(context.TODO(), arg.PaymentCustomerID, arg.PaymentMethodID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, arg.Uid)
	require.Equal(t, res.PaymentMethodID, arg.PaymentMethodID)
	require.Equal(t, res.PaymentRequestID, arg.PaymentRequestID)
	require.Equal(t, res.PaymentReferenceID, arg.PaymentReferenceID)
	require.Equal(t, res.PaymentCustomerID, arg.PaymentCustomerID)
	require.Equal(t, res.PaymentType, arg.PaymentType)
	require.Equal(t, res.PaymentStatus, arg.PaymentStatus)
	require.Equal(t, res.PaymentReusability, arg.PaymentReusability)
	require.Equal(t, res.PaymentChannel, arg.PaymentChannel)
	require.Equal(t, res.PaymentAmount.String(), arg.PaymentAmount.String())
	require.Equal(t, res.PaymentQrCode, arg.PaymentQrCode)
	require.Equal(t, res.PaymentVirtualAccountNumber, arg.PaymentVirtualAccountNumber)
	require.Equal(t, res.PaymentUrl, arg.PaymentUrl)

}

func TestRepoDeletePaymentCache(t *testing.T) {
	arg := createRandomPaymentCache(t)

	res, err := testStore.GetCache(context.TODO(), arg.PaymentCustomerID, arg.PaymentMethodID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, arg.Uid)
	require.Equal(t, res.PaymentMethodID, arg.PaymentMethodID)
	require.Equal(t, res.PaymentRequestID, arg.PaymentRequestID)
	require.Equal(t, res.PaymentReferenceID, arg.PaymentReferenceID)
	require.Equal(t, res.PaymentCustomerID, arg.PaymentCustomerID)
	require.Equal(t, res.PaymentType, arg.PaymentType)
	require.Equal(t, res.PaymentStatus, arg.PaymentStatus)
	require.Equal(t, res.PaymentReusability, arg.PaymentReusability)
	require.Equal(t, res.PaymentChannel, arg.PaymentChannel)
	require.Equal(t, res.PaymentAmount.String(), arg.PaymentAmount.String())
	require.Equal(t, res.PaymentQrCode, arg.PaymentQrCode)
	require.Equal(t, res.PaymentVirtualAccountNumber, arg.PaymentVirtualAccountNumber)
	require.Equal(t, res.PaymentUrl, arg.PaymentUrl)

	testStore.DeleteCache(context.TODO(), arg.PaymentCustomerID, arg.PaymentMethodID)

	res2, err := testStore.GetCache(context.TODO(), arg.PaymentCustomerID, arg.PaymentMethodID)
	require.Error(t, err)
	require.Empty(t, res2)
}

func createRandomPaymentCache(t *testing.T) *PaymentMethod {
	payment := createRandomPaymentMethod(t)

	testStore.PutCache(context.TODO(), payment)

	return payment
}
