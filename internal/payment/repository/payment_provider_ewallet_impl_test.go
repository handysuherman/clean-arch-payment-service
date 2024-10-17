package repository

import (
	"context"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
	"github.com/xendit/xendit-go/v5/payment_method"
	"github.com/xendit/xendit-go/v5/payment_request"
)

func TestRepoCreatePaymentEwalletOVO(t *testing.T) {
	createRandomPaymentEwallet(t, payment_method.EWALLETCHANNELCODE_OVO)
}

func TestRepoCreatePaymentEwalletAstraPay(t *testing.T) {
	createRandomPaymentEwallet(t, payment_method.EWALLETCHANNELCODE_ASTRAPAY)
}

func TestRepoCreatePaymentEwalletLinkAja(t *testing.T) {
	createRandomPaymentEwallet(t, payment_method.EWALLETCHANNELCODE_LINKAJA)
}

func TestRepoCreatePaymentEwalletShopeePay(t *testing.T) {
	createRandomPaymentEwallet(t, payment_method.EWALLETCHANNELCODE_SHOPEEPAY)
}

func TestRepoCreatePaymentEwalletDANA(t *testing.T) {
	createRandomPaymentEwallet(t, payment_method.EWALLETCHANNELCODE_DANA)
}

func createRandomPaymentEwallet(t *testing.T, channelCode payment_method.EWalletChannelCode) *payment_request.PaymentRequest {
	customer := createRandomPaymentProviderCustomer(t)

	referenceID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, referenceID)

	arg := CreateEwalletPaymentParams{
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

	res, err := testStore.CreateEwalletPayment(context.TODO(), &arg)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.NotEmpty(t, res.Id)

	url, err := getURl(arg.ChannelCode, res.GetActions())
	require.NoError(t, err)

	if arg.ChannelCode != payment_method.EWALLETCHANNELCODE_OVO {
		require.NotEmpty(t, url)
	}

	require.Equal(t, string(*res.PaymentMethod.GetEwallet().ChannelCode), string(arg.ChannelCode))
	require.Equal(t, res.GetReferenceId(), referenceID.String())
	require.Equal(t, res.GetAmount(), arg.Amount)
	require.Equal(t, res.GetCustomerId(), customer.Id)
	require.Equal(t, string(res.PaymentMethod.GetReusability()), string(payment_method.PAYMENTMETHODREUSABILITY_ONE_TIME_USE))
	require.Equal(t, string(res.PaymentMethod.GetType()), string(payment_method.PAYMENTMETHODTYPE_EWALLET))

	return res
}
