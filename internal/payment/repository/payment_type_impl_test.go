package repository

import (
	"context"
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
)

func TestRepoCreatePaymentType(t *testing.T) {
	createRandomPaymentType(t)
}

func TestRepoGetPaymentTypeByName(t *testing.T) {
	pt := createRandomPaymentType(t)

	res, err := testStore.GetPaymentTypeByName(context.TODO(), pt.Ptname)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, pt.Ptname, res)
}

func createRandomPaymentType(t *testing.T) *PaymentType {
	name := helper.RandomString(32)

	res, err := testStore.CreatePaymentType(context.TODO(), name)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, name, res)

	return &PaymentType{
		Ptname: name,
	}
}
