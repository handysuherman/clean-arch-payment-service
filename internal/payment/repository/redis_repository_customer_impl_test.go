package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoCreateCustomerCache(t *testing.T) {
	createRandomCustomerCache(t)
}

func TestRepoGetCustomerCache(t *testing.T) {
	arg := createRandomCustomerCache(t)

	res, err := testStore.GetCustomerCache(context.TODO(), arg.CustomerAppID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, arg.Uid)
	require.Equal(t, res.CustomerAppID, arg.CustomerAppID)
	require.Equal(t, res.PaymentCustomerID, arg.PaymentCustomerID)
	require.Equal(t, res.CustomerName, arg.CustomerName)
	require.Equal(t, res.Email, arg.Email)
	require.Equal(t, res.PhoneNumber, arg.PhoneNumber)
}

func TestRepoDeleteCustomerCache(t *testing.T) {
	arg := createRandomCustomerCache(t)

	res, err := testStore.GetCustomerCache(context.TODO(), arg.CustomerAppID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, arg.Uid)
	require.Equal(t, res.CustomerAppID, arg.CustomerAppID)
	require.Equal(t, res.PaymentCustomerID, arg.PaymentCustomerID)
	require.Equal(t, res.CustomerName, arg.CustomerName)
	require.Equal(t, res.Email, arg.Email)
	require.Equal(t, res.PhoneNumber, arg.PhoneNumber)

	testStore.DeleteCustomerCache(context.TODO(), arg.CustomerAppID)

	res2, err := testStore.GetCustomerCache(context.TODO(), arg.CustomerAppID)
	require.Error(t, err)
	require.Empty(t, res2)
}

func createRandomCustomerCache(t *testing.T) *Customer {
	customer := createRandomCustomerTx(t)

	testStore.PutCustomerCache(context.TODO(), customer)

	return customer
}
