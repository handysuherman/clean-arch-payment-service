package repository

import (
	"context"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestRepoCreateCustomer(t *testing.T) {
	createRandomCustomer(t)
}

func TestRepoGetCustomerByCustomerAppID(t *testing.T) {
	customer := createRandomCustomer(t)

	res, err := testStore.GetCustomerByCustomerAppID(context.TODO(), customer.CustomerAppID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, customer.Uid)
	require.Equal(t, res.CustomerAppID, customer.CustomerAppID)
	require.Equal(t, res.PaymentCustomerID, customer.PaymentCustomerID)
	require.Equal(t, res.CustomerName, customer.CustomerName)
	require.Equal(t, res.CreatedAt, customer.CreatedAt)
	require.Equal(t, res.Email, customer.Email)
	require.Equal(t, res.PhoneNumber, customer.PhoneNumber)
}

func TestRepoGetCustomerByPaymentCustomerID(t *testing.T) {
	customer := createRandomCustomer(t)

	res, err := testStore.GetCustomerByPaymentCustomerID(context.TODO(), customer.PaymentCustomerID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, customer.Uid)
	require.Equal(t, res.CustomerAppID, customer.CustomerAppID)
	require.Equal(t, res.PaymentCustomerID, customer.PaymentCustomerID)
	require.Equal(t, res.CustomerName, customer.CustomerName)
	require.Equal(t, res.CreatedAt, customer.CreatedAt)
	require.Equal(t, res.Email, customer.Email)
	require.Equal(t, res.PhoneNumber, customer.PhoneNumber)
}

func createRandomCustomer(t *testing.T) *Customer {
	ulid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, ulid.String())

	customerUlid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, customerUlid.String())

	paymentCustomerUlid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, paymentCustomerUlid.String())

	arg := CreateCustomerParams{
		Uid:               ulid.String(),
		CustomerAppID:     customerUlid.String(),
		PaymentCustomerID: paymentCustomerUlid.String(),
		CustomerName:      helper.RandomString(30),
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Email: pgtype.Text{
			String: helper.RandomEmail(),
			Valid:  true,
		},
		PhoneNumber: pgtype.Text{
			String: helper.RandomStringInt(12),
			Valid:  true,
		},
	}

	res, err := testStore.CreateCustomer(context.TODO(), &arg)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Uid, arg.Uid)
	require.Equal(t, res.CustomerAppID, arg.CustomerAppID)
	require.Equal(t, res.PaymentCustomerID, arg.PaymentCustomerID)
	require.Equal(t, res.CustomerName, arg.CustomerName)
	require.Equal(t, res.Email, arg.Email)
	require.Equal(t, res.PhoneNumber, arg.PhoneNumber)

	return res
}
