package repository

import (
	"context"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
	"github.com/xendit/xendit-go/v5/payment_method"
)

func TestRepoCreatePaymentVritualAccountBankBCA(t *testing.T) {
	createRandomPaymentVirtualAccountBank(t, payment_method.VIRTUALACCOUNTCHANNELCODE_BCA)
}

func TestRepoCreatePaymentVritualAccountBankBNI(t *testing.T) {
	createRandomPaymentVirtualAccountBank(t, payment_method.VIRTUALACCOUNTCHANNELCODE_BNI)
}

func TestRepoCreatePaymentVritualAccountBankBRI(t *testing.T) {
	createRandomPaymentVirtualAccountBank(t, payment_method.VIRTUALACCOUNTCHANNELCODE_BRI)
}

func TestRepoCreatePaymentVritualAccountBankBSI(t *testing.T) {
	createRandomPaymentVirtualAccountBank(t, payment_method.VIRTUALACCOUNTCHANNELCODE_BSI)
}

func TestRepoCreatePaymentVritualAccountBankCIMB(t *testing.T) {
	createRandomPaymentVirtualAccountBank(t, payment_method.VIRTUALACCOUNTCHANNELCODE_CIMB)
}

func TestRepoCreatePaymentVritualAccountBankMANDIRI(t *testing.T) {
	createRandomPaymentVirtualAccountBank(t, payment_method.VIRTUALACCOUNTCHANNELCODE_MANDIRI)
}

func TestRepoCreatePaymentVritualAccountBankPERMATA(t *testing.T) {
	createRandomPaymentVirtualAccountBank(t, payment_method.VIRTUALACCOUNTCHANNELCODE_PERMATA)
}

func createRandomPaymentVirtualAccountBank(t *testing.T, channelCode payment_method.VirtualAccountChannelCode) *payment_method.PaymentMethod {
	customer := createRandomPaymentProviderCustomer(t)

	referenceID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, referenceID)

	arg := CreateVirtualAccountBankPaymentParams{
		CustomerName:      *customer.GetIndividualDetail().GivenNames,
		CustomerPaymentID: customer.Id,
		CustomerNumber:    customer.GetMobileNumber(),
		Description:       helper.RandomString(100),
		ReferenceID:       referenceID.String(),
		Amount:            float64(helper.RandomInt(100000, 200000)),
		Expiry:            time.Now().Add(24 * 3 * time.Hour),
		ChannelCode:       channelCode,
		SuccessReturnURL:  helper.RandomUrl(),
		FailureReturnURL:  helper.RandomUrl(),
	}

	res, err := testStore.CreateVirtualAccountBankPayment(context.TODO(), &arg)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.NotEmpty(t, res.Id)
	require.Equal(t, res.GetVirtualAccount().ChannelCode.String(), string(channelCode))
	require.Equal(t, res.GetVirtualAccount().Amount.Get(), &arg.Amount)
	require.Equal(t, res.GetCustomerId(), customer.Id)
	require.Equal(t, res.GetReferenceId(), arg.ReferenceID)
	require.Equal(t, string(res.GetReusability()), string(payment_method.PAYMENTMETHODREUSABILITY_ONE_TIME_USE))
	require.Equal(t, string(res.GetType()), string(payment_method.PAYMENTMETHODTYPE_VIRTUAL_ACCOUNT))

	return res
}
