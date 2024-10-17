package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
	"github.com/xendit/xendit-go/v5/payment_method"
)

func TestRepoCreateEwalletPaymentTxOVO(t *testing.T) {
	createRandomEwalletPaymentTx(t, payment_method.EWALLETCHANNELCODE_OVO)
}

func TestRepoCreateEwalletPaymentTxASTRAPAY(t *testing.T) {
	createRandomEwalletPaymentTx(t, payment_method.EWALLETCHANNELCODE_ASTRAPAY)
}

func TestRepoCreateEwalletPaymentTxLINKAJA(t *testing.T) {
	createRandomEwalletPaymentTx(t, payment_method.EWALLETCHANNELCODE_LINKAJA)
}

func TestRepoCreateEwalletPaymentTxDANA(t *testing.T) {
	createRandomEwalletPaymentTx(t, payment_method.EWALLETCHANNELCODE_DANA)
}

func TestRepoCreateEwalletPaymentTxSHOPEEPAY(t *testing.T) {
	createRandomEwalletPaymentTx(t, payment_method.EWALLETCHANNELCODE_SHOPEEPAY)
}

func createRandomEwalletPaymentTx(t *testing.T, channelCode payment_method.EWalletChannelCode) *PaymentMethod {
	customer := createRandomPaymentProviderCustomer(t)

	arg := CreateEwalletPaymentParams{
		CustomerName:      *customer.GetIndividualDetail().GivenNames,
		CustomerPaymentID: customer.Id,
		CustomerNumber:    customer.GetMobileNumber(),
		Description:       helper.RandomString(100),
		Amount:            float64(helper.RandomInt(100000, 200000)),
		Expiry:            time.Now().Add(24 * 3 * time.Hour),
		ChannelCode:       channelCode,
		SuccessReturnURL:  helper.RandomUrl(),
		FailureReturnURL:  helper.RandomUrl(),
	}

	res, err := testStore.CreateEwalletPaymentTx(context.TODO(), &CreateEwalletPaymentTxParams{CreateEwalletPayment: arg})
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.NotEmpty(t, res.Payment.PaymentMethodID)
	require.NotEmpty(t, res.Payment.PaymentCustomerID)

	if arg.ChannelCode != payment_method.EWALLETCHANNELCODE_OVO {
		require.NotEmpty(t, res.Payment.PaymentUrl.String)
	}

	require.Equal(t, res.Payment.PaymentType, string(payment_method.PAYMENTMETHODTYPE_EWALLET))
	require.Equal(t, res.Payment.PaymentReusability, string(payment_method.PAYMENTMETHODREUSABILITY_ONE_TIME_USE))
	require.Equal(t, res.Payment.PaymentChannel, string(arg.ChannelCode))
	require.Equal(t, res.Payment.PaymentDescription, arg.Description)

	resAmount, err := strconv.ParseFloat(res.Payment.PaymentAmount.String(), 64)
	require.NoError(t, err)
	require.NotEmpty(t, resAmount)

	require.Equal(t, resAmount, arg.Amount)

	return res.Payment
}
