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

func TestRepoCreatePaymentQrCodeTxQRIS(t *testing.T) {
	createRandomQrCodePaymentTx(t, payment_method.QRCODECHANNELCODE_QRIS)
}

func createRandomQrCodePaymentTx(t *testing.T, channelCode payment_method.QRCodeChannelCode) *PaymentMethod {
	customer := createRandomPaymentProviderCustomer(t)

	arg := CreateQrCodePaymentParams{
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

	res, err := testStore.CreateQrCodePaymentTx(context.TODO(), &CreateQrCodePaymentTxParams{CreateQrCodePayment: arg})
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.NotEmpty(t, res.Payment.PaymentMethodID)
	require.NotEmpty(t, res.Payment.PaymentCustomerID)
	require.NotEmpty(t, res.Payment.PaymentQrCode.String)
	require.Equal(t, res.Payment.PaymentType, string(payment_method.PAYMENTMETHODTYPE_QR_CODE))
	require.Equal(t, res.Payment.PaymentReusability, string(payment_method.PAYMENTMETHODREUSABILITY_ONE_TIME_USE))
	require.Equal(t, res.Payment.PaymentChannel, string(arg.ChannelCode))
	require.Equal(t, res.Payment.PaymentDescription, arg.Description)

	resAmount, err := strconv.ParseFloat(res.Payment.PaymentAmount.String(), 64)
	require.NoError(t, err)
	require.NotEmpty(t, resAmount)

	require.Equal(t, resAmount, arg.Amount)

	return res.Payment
}
