package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestRepoCreateCustomerTx(t *testing.T) {
	createRandomCustomerTx(t)
}

func createRandomCustomerTx(t *testing.T) *Customer {
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

	res, err := testStore.CreateCustomerTx(context.TODO(), &CreateCustomerTxParams{CreateCustomer: arg})
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Customer.Uid, arg.Uid)
	require.Equal(t, res.Customer.CustomerAppID, arg.CustomerAppID)
	require.Equal(t, res.Customer.CustomerName, arg.CustomerName)
	require.Equal(t, res.Customer.Email, arg.Email)
	require.Equal(t, strings.ReplaceAll(res.Customer.PhoneNumber.String, "+", ""), arg.PhoneNumber.String)

	return res.Customer
}
