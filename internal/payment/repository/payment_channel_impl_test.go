package repository

import (
	"context"
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestRepoCreatePaymentChannel(t *testing.T) {
	createRandomPaymentChannel(t)
}

func TestRepoGetPaymentChannelByID(t *testing.T) {
	pc := createRandomPaymentChannel(t)

	res, err := testStore.GetPaymentChannelByID(context.TODO(), pc.Uid)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, pc.Uid, res.Uid)
	require.Equal(t, pc.Pcname, res.Pcname)
	require.Equal(t, pc.MinAmount.String(), res.MinAmount.String())
	require.Equal(t, pc.MaxAmount.String(), res.MaxAmount.String())
	require.Equal(t, pc.Tax.String(), res.Tax.String())
	require.Equal(t, pc.IsTaxPercentage, res.IsTaxPercentage)
}

func TestRepoGetPaymentChannelByName(t *testing.T) {
	pc := createRandomPaymentChannel(t)

	res, err := testStore.GetPaymentChannelByName(context.TODO(), pc.Pcname)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, pc.Uid, res.Uid)
	require.Equal(t, pc.Pcname, res.Pcname)
	require.Equal(t, pc.MinAmount.String(), res.MinAmount.String())
	require.Equal(t, pc.MaxAmount.String(), res.MaxAmount.String())
	require.Equal(t, pc.Tax.String(), res.Tax.String())
	require.Equal(t, pc.IsTaxPercentage, res.IsTaxPercentage)
}

func TestRepoGetAvailableChannels(t *testing.T) {
	pc := createRandomPaymentChannel(t)

	res, err := testStore.GetAvailablePaymentChannels(context.TODO(), decimal.NewFromInt(helper.RandomInt(int64(pc.MinAmount.Round(0).IntPart()), int64(pc.MaxAmount.Round(0).IntPart()))))
	require.NoError(t, err)
	require.NotEmpty(t, res)
}

func TestRepoGetAvailableChannel(t *testing.T) {
	pc, err := testStore.GetPaymentChannelByName(context.TODO(), "BCA")
	require.NoError(t, err)
	require.NotEmpty(t, pc)

	arg := GetAvailablePaymentChannelParams{
		PcType:    pc.PcType,
		Pcname:    pc.Pcname,
		MinAmount: decimal.NewFromInt(helper.RandomInt(int64(pc.MinAmount.Round(0).IntPart()), int64(pc.MaxAmount.Round(0).IntPart()))),
	}

	res, err := testStore.GetAvailablePaymentChannel(context.TODO(), &arg)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, pc.Uid, res.Uid)
	require.Equal(t, pc.Pcname, res.Pcname)
	require.Equal(t, pc.MinAmount.String(), res.MinAmount.String())
	require.Equal(t, pc.MaxAmount.String(), res.MaxAmount.String())
	require.Equal(t, pc.Tax.String(), res.Tax.String())
	require.Equal(t, pc.IsTaxPercentage, res.IsTaxPercentage)
}

func createRandomPaymentChannel(t *testing.T) *PaymentChannel {
	ulid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, ulid.String())

	paymentType := createRandomPaymentType(t)

	arg := CreatePaymentChannelParams{
		Uid:             ulid.String(),
		Pcname:          helper.RandomString(32),
		PcType:          paymentType.Ptname,
		MinAmount:       decimal.NewFromInt(helper.RandomInt(1000, 100000)),
		MaxAmount:       decimal.NewFromInt(helper.RandomInt(100001, 200000)),
		Tax:             decimal.NewFromInt(helper.RandomInt(1000, 4000)),
		IsTaxPercentage: false,
	}

	res, err := testStore.CreatePaymentChannel(context.TODO(), &arg)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, arg.Uid, res.Uid)
	require.Equal(t, arg.Pcname, res.Pcname)
	require.Equal(t, arg.MinAmount.String(), res.MinAmount.String())
	require.Equal(t, arg.MaxAmount.String(), res.MaxAmount.String())
	require.Equal(t, arg.Tax.String(), res.Tax.String())
	require.Equal(t, arg.IsTaxPercentage, res.IsTaxPercentage)
	require.False(t, res.IsActive)

	return res
}
