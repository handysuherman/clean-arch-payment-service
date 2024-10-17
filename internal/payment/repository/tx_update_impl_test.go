package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func Test_REPO_UPDATE_TX(t *testing.T) {
	pm := createRandomPaymentMethod(t)
	paymentStatus := createRandomPaymentStatus(t)

	arg := UpdatePaymentMethodCustomerParams{
		PaymentStatus: pgtype.Text{
			String: paymentStatus.Psname,
			Valid:  true,
		},
		PaymentMethodID:   pm.PaymentMethodID,
		PaymentCustomerID: pm.PaymentCustomerID,
	}

	res, err := testStore.UpdateTx(context.TODO(), &UpdateTxParams{UpdateParams: arg})
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Payment.Uid, pm.Uid)
	require.Equal(t, res.Payment.PaymentMethodID, pm.PaymentMethodID)
	require.Equal(t, res.Payment.PaymentRequestID, pm.PaymentRequestID)
	require.Equal(t, res.Payment.PaymentReferenceID, pm.PaymentReferenceID)
	require.Equal(t, res.Payment.PaymentCustomerID, pm.PaymentCustomerID)
	require.Equal(t, res.Payment.PaymentType, pm.PaymentType)
	require.Equal(t, res.Payment.PaymentReusability, pm.PaymentReusability)
	require.Equal(t, res.Payment.PaymentChannel, pm.PaymentChannel)
	require.Equal(t, res.Payment.PaymentAmount.String(), pm.PaymentAmount.String())
	require.Equal(t, res.Payment.PaymentQrCode, pm.PaymentQrCode)
	require.Equal(t, res.Payment.PaymentVirtualAccountNumber, pm.PaymentVirtualAccountNumber)
	require.Equal(t, res.Payment.PaymentUrl, pm.PaymentUrl)

	require.NotEqual(t, res.Payment.PaymentStatus, pm.PaymentStatus)
	require.Equal(t, res.Payment.PaymentStatus, paymentStatus.Psname)
}
