package repository

import (
	"context"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
	"github.com/xendit/xendit-go/v5/payment_method"
)

func TestRepoCreatePaymentQrcodeQRIS(t *testing.T) {
	createRandomPaymentQrCode(t, payment_method.QRCODECHANNELCODE_QRIS)
}

func TestRepoGetPaymentQrCodeByID(t *testing.T) {
	qrcode := createRandomPaymentQrCode(t, payment_method.QRCODECHANNELCODE_QRIS)

	res, err := testStore.GetQrCodePaymentByID(context.TODO(), qrcode.Id)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.NotEmpty(t, res.Id)
	require.Equal(t, string(*res.GetQrCode().ChannelCode.Get()), string(*qrcode.GetQrCode().ChannelCode.Get()))
	require.Equal(t, res.GetQrCode().Amount.Get(), qrcode.GetQrCode().Amount.Get())
	require.Equal(t, res.GetCustomerId(), qrcode.GetCustomerId())
	require.Equal(t, string(res.GetReusability()), string(payment_method.PAYMENTMETHODREUSABILITY_ONE_TIME_USE))
	require.Equal(t, string(res.GetType()), string(payment_method.PAYMENTMETHODTYPE_QR_CODE))
}

func createRandomPaymentQrCode(t *testing.T, channelCode payment_method.QRCodeChannelCode) *payment_method.PaymentMethod {
	customer := createRandomPaymentProviderCustomer(t)

	referenceID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, referenceID)

	arg := CreateQrCodePaymentParams{
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

	res, err := testStore.CreateQrCodePayment(context.TODO(), &arg)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.NotEmpty(t, res.Id)
	require.NotEmpty(t, res.GetQrCode().ChannelProperties.Get().GetQrString())
	require.Equal(t, string(*res.GetQrCode().ChannelCode.Get()), string(channelCode))
	require.Equal(t, res.GetQrCode().Amount.Get(), &arg.Amount)
	require.Equal(t, res.GetCustomerId(), customer.Id)
	require.Equal(t, string(res.GetReusability()), string(payment_method.PAYMENTMETHODREUSABILITY_ONE_TIME_USE))
	require.Equal(t, string(res.GetType()), string(payment_method.PAYMENTMETHODTYPE_QR_CODE))
	require.Equal(t, res.GetReferenceId(), arg.ReferenceID)

	return res
}
