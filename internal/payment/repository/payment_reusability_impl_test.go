package repository

import (
	"context"
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
)

func TestRepoCreateRandomPaymentReusability(t *testing.T) {
	createRandomPaymentReusability(t)
}

func TestRepoGetPaymentReusabilityByName(t *testing.T) {
	pr := createRandomPaymentReusability(t)

	res, err := testStore.GetPaymentReusabilityByName(context.TODO(), pr.Prname)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, pr.Prname, res)
}

func createRandomPaymentReusability(t *testing.T) *PaymentReusability {
	name := helper.RandomString(32)

	res, err := testStore.CreatePaymentReusability(context.TODO(), name)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, name, res)

	return &PaymentReusability{
		Prname: res,
	}
}
