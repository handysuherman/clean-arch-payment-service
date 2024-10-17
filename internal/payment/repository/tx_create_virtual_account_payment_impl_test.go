package repository

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
	"github.com/xendit/xendit-go/v5/payment_method"
)

func TestRepoCreatePaymentVirtualAccountBankTxBCA(t *testing.T) {
	createRandomPaymentVirtualAccountBankTx(t, payment_method.VIRTUALACCOUNTCHANNELCODE_BCA)
}

func TestRepoCreatePaymentVirtualAccountBankTxBNI(t *testing.T) {
	createRandomPaymentVirtualAccountBankTx(t, payment_method.VIRTUALACCOUNTCHANNELCODE_BNI)
}

func TestRepoCreatePaymentVirtualAccountBankTxBRI(t *testing.T) {
	createRandomPaymentVirtualAccountBankTx(t, payment_method.VIRTUALACCOUNTCHANNELCODE_BRI)
}

func TestRepoCreatePaymentVirtualAccountBankTxBSI(t *testing.T) {
	createRandomPaymentVirtualAccountBankTx(t, payment_method.VIRTUALACCOUNTCHANNELCODE_BSI)
}

func TestRepoCreatePaymentVirtualAccountBankTxCIMB(t *testing.T) {
	createRandomPaymentVirtualAccountBankTx(t, payment_method.VIRTUALACCOUNTCHANNELCODE_CIMB)
}

func TestRepoCreatePaymentVirtualAccountBankTxMANDIRI(t *testing.T) {
	createRandomPaymentVirtualAccountBankTx(t, payment_method.VIRTUALACCOUNTCHANNELCODE_MANDIRI)
}

func TestRepoCreatePaymentVirtualAccountBankTxPERMATA(t *testing.T) {
	createRandomPaymentVirtualAccountBankTx(t, payment_method.VIRTUALACCOUNTCHANNELCODE_PERMATA)
}

func createRandomPaymentVirtualAccountBankTx(t *testing.T, channelCode payment_method.VirtualAccountChannelCode) *PaymentMethod {
	customer := createRandomPaymentProviderCustomer(t)

	arg := CreateVirtualAccountBankPaymentParams{
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

	res, err := testStore.CreateVirtualAccountBankPaymentTx(context.TODO(), &CreateVirtualAccountBankPaymentTxParams{CreateVirtualAccountBankPayment: arg})
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.NotEmpty(t, res.Payment.PaymentMethodID)
	require.NotEmpty(t, res.Payment.PaymentCustomerID)
	require.NotEmpty(t, res.Payment.PaymentVirtualAccountNumber.String)
	require.Equal(t, res.Payment.PaymentType, string(payment_method.PAYMENTMETHODTYPE_VIRTUAL_ACCOUNT))
	require.Equal(t, res.Payment.PaymentReusability, string(payment_method.PAYMENTMETHODREUSABILITY_ONE_TIME_USE))
	require.Equal(t, res.Payment.PaymentChannel, string(arg.ChannelCode))
	require.Equal(t, res.Payment.PaymentDescription, arg.Description)
	fmt.Println(t, res.Payment.ExpiresAt.Time.Unix())
	fmt.Println(t, res.Payment.CreatedAt.Time.Unix())

	resAmount, err := strconv.ParseFloat(res.Payment.PaymentAmount.String(), 64)
	require.NoError(t, err)
	require.NotEmpty(t, resAmount)

	require.Equal(t, resAmount, arg.Amount)

	return res.Payment
}
