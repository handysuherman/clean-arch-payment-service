package repository

import (
	"context"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestRepoCreatePaymentMethod(t *testing.T) {
	createRandomPaymentMethod(t)
}

func TestRepoGetPaymentMethodByPaymentMethodID(t *testing.T) {
	pm := createRandomPaymentMethod(t)

	res, err := testStore.GetPaymentMethodByPaymentMethodID(context.TODO(), pm.PaymentMethodID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, pm.Uid)
	require.Equal(t, res.PaymentMethodID, pm.PaymentMethodID)
	require.Equal(t, res.PaymentRequestID, pm.PaymentRequestID)
	require.Equal(t, res.PaymentReferenceID, pm.PaymentReferenceID)
	require.Equal(t, res.PaymentCustomerID, pm.PaymentCustomerID)
	require.Equal(t, res.PaymentType, pm.PaymentType)
	require.Equal(t, res.PaymentStatus, pm.PaymentStatus)
	require.Equal(t, res.PaymentReusability, pm.PaymentReusability)
	require.Equal(t, res.PaymentChannel, pm.PaymentChannel)
	require.Equal(t, res.PaymentAmount.String(), pm.PaymentAmount.String())
	require.Equal(t, res.PaymentQrCode, pm.PaymentQrCode)
	require.Equal(t, res.PaymentVirtualAccountNumber, pm.PaymentVirtualAccountNumber)
	require.Equal(t, res.PaymentUrl, pm.PaymentUrl)
}

func TestRepoGetPaymentMethodCustomer(t *testing.T) {
	pm := createRandomPaymentMethod(t)

	arg := GetPaymentMethodCustomerParams{
		PaymentMethodID:   pm.PaymentMethodID,
		PaymentCustomerID: pm.PaymentCustomerID,
	}

	res, err := testStore.GetPaymentMethodCustomer(context.TODO(), &arg)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, pm.Uid)
	require.Equal(t, res.PaymentMethodID, pm.PaymentMethodID)
	require.Equal(t, res.PaymentRequestID, pm.PaymentRequestID)
	require.Equal(t, res.PaymentReferenceID, pm.PaymentReferenceID)
	require.Equal(t, res.PaymentCustomerID, pm.PaymentCustomerID)
	require.Equal(t, res.PaymentType, pm.PaymentType)
	require.Equal(t, res.PaymentStatus, pm.PaymentStatus)
	require.Equal(t, res.PaymentReusability, pm.PaymentReusability)
	require.Equal(t, res.PaymentChannel, pm.PaymentChannel)
	require.Equal(t, res.PaymentAmount.String(), pm.PaymentAmount.String())
	require.Equal(t, res.PaymentQrCode, pm.PaymentQrCode)
	require.Equal(t, res.PaymentVirtualAccountNumber, pm.PaymentVirtualAccountNumber)
	require.Equal(t, res.PaymentUrl, pm.PaymentUrl)
}

func TestREpoGetPaymentMethodByReferenceID(t *testing.T) {
	pm := createRandomPaymentMethod(t)

	res, err := testStore.GetPaymentMethodByReferenceID(context.TODO(), pm.PaymentReferenceID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, pm.Uid)
	require.Equal(t, res.PaymentMethodID, pm.PaymentMethodID)
	require.Equal(t, res.PaymentRequestID, pm.PaymentRequestID)
	require.Equal(t, res.PaymentReferenceID, pm.PaymentReferenceID)
	require.Equal(t, res.PaymentCustomerID, pm.PaymentCustomerID)
	require.Equal(t, res.PaymentType, pm.PaymentType)
	require.Equal(t, res.PaymentStatus, pm.PaymentStatus)
	require.Equal(t, res.PaymentReusability, pm.PaymentReusability)
	require.Equal(t, res.PaymentChannel, pm.PaymentChannel)
	require.Equal(t, res.PaymentAmount.String(), pm.PaymentAmount.String())
	require.Equal(t, res.PaymentQrCode, pm.PaymentQrCode)
	require.Equal(t, res.PaymentVirtualAccountNumber, pm.PaymentVirtualAccountNumber)
	require.Equal(t, res.PaymentUrl, pm.PaymentUrl)
}

func TestRepoUpdatePaymentMethod(t *testing.T) {
	pm := createRandomPaymentMethod(t)
	paymentStatus := createRandomPaymentStatus(t)

	arg := &UpdatePaymentMethodCustomerParams{
		PaymentStatus: pgtype.Text{
			String: paymentStatus.Psname,
			Valid:  true,
		},
		PaymentMethodID:   pm.PaymentMethodID,
		PaymentCustomerID: pm.PaymentCustomerID,
	}

	res, err := testStore.UpdatePaymentMethodCustomer(context.TODO(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, pm.Uid)
	require.Equal(t, res.PaymentMethodID, pm.PaymentMethodID)
	require.Equal(t, res.PaymentRequestID, pm.PaymentRequestID)
	require.Equal(t, res.PaymentReferenceID, pm.PaymentReferenceID)
	require.Equal(t, res.PaymentCustomerID, pm.PaymentCustomerID)
	require.Equal(t, res.PaymentType, pm.PaymentType)
	require.Equal(t, res.PaymentReusability, pm.PaymentReusability)
	require.Equal(t, res.PaymentChannel, pm.PaymentChannel)
	require.Equal(t, res.PaymentAmount.String(), pm.PaymentAmount.String())
	require.Equal(t, res.PaymentQrCode, pm.PaymentQrCode)
	require.Equal(t, res.PaymentVirtualAccountNumber, pm.PaymentVirtualAccountNumber)
	require.Equal(t, res.PaymentUrl, pm.PaymentUrl)

	require.NotEqual(t, res.PaymentStatus, pm.PaymentStatus)
	require.Equal(t, res.PaymentStatus, paymentStatus.Psname)
}

func createRandomPaymentMethod(t *testing.T) *PaymentMethod {
	ulid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, ulid.String())

	paymentMethodID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, paymentMethodID.String())

	paymentBusinessID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, paymentBusinessID.String())

	paymentRequestID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, paymentRequestID.String())

	paymentReferenceID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, paymentReferenceID.String())

	customer := createRandomCustomer(t)
	paymentType := createRandomPaymentType(t)
	paymentStatus := createRandomPaymentStatus(t)
	paymentReusability := createRandomPaymentReusability(t)
	paymentChannel := createRandomPaymentChannel(t)

	arg := CreatePaymentMethodParams{
		Uid:             ulid.String(),
		PaymentMethodID: paymentMethodID.String(),
		PaymentRequestID: pgtype.Text{
			String: paymentRequestID.String(),
			Valid:  true,
		},
		PaymentBusinessID:  paymentBusinessID.String(),
		PaymentReferenceID: paymentReferenceID.String(),
		PaymentCustomerID:  customer.PaymentCustomerID,
		PaymentType:        paymentType.Ptname,
		PaymentStatus:      paymentStatus.Psname,
		PaymentReusability: paymentReusability.Prname,
		PaymentChannel:     paymentChannel.Pcname,
		PaymentAmount:      decimal.NewFromInt(helper.RandomInt(5000, 500000)),
		PaymentQrCode: pgtype.Text{
			String: helper.RandomString(300),
			Valid:  true,
		},
		PaymentVirtualAccountNumber: pgtype.Text{
			String: helper.RandomString(38),
			Valid:  true,
		},
		PaymentUrl: pgtype.Text{
			String: helper.RandomUrl(),
			Valid:  true,
		},
		PaymentDescription: helper.RandomString(100),
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(24 * 3 * time.Hour),
			Valid: true,
		},
	}

	res, err := testStore.CreatePaymentMethod(context.TODO(), &arg)
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
	// require.Equal(t, res.CreatedAt.Time.string)

	return res
}
