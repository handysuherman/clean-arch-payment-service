package repository

import (
	"context"
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
)

func TestRepoCreatePaymentStatus(t *testing.T) {
	createRandomPaymentStatus(t)
}

func TestRepoGetPaymentStatusByName(t *testing.T) {
	ps := createRandomPaymentStatus(t)

	res, err := testStore.GetPaymentStatusByName(context.TODO(), ps.Psname)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res, ps.Psname)
}

func createRandomPaymentStatus(t *testing.T) *PaymentStatus {
	name := helper.RandomString(32)

	res, err := testStore.CreatePaymentStatus(context.TODO(), name)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, name, res)

	return &PaymentStatus{
		Psname: res,
	}
}
