package repository

import (
	"context"
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/stretchr/testify/require"
	"github.com/xendit/xendit-go/v5/customer"
)

func TestRepoCreateCustomerPayment(t *testing.T) {
	createRandomPaymentProviderCustomer(t)
}

func TestRepoGetCustomerPaymentByID(t *testing.T) {
	customer := createRandomPaymentProviderCustomer(t)

	res, err := testStore.GetCustomerPaymentByID(context.TODO(), customer.Id)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, customer.Id, res.Id)
	require.Equal(t, customer.GetIndividualDetail().GivenNames, res.GetIndividualDetail().GivenNames)
	require.Equal(t, customer.GetMobileNumber(), res.GetMobileNumber())
	require.Equal(t, customer.GetPhoneNumber(), res.GetPhoneNumber())
}

func createRandomPaymentProviderCustomer(t *testing.T) *customer.Customer {
	ulid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, ulid)

	referenceID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, referenceID)

	arg := CreateCustomerPaymentParams{
		CustomerName:   helper.RandomString(28),
		CustomerNumber: helper.RandomStringInt(12),
		CustomerUID:    ulid.String(),
		ReferenceID:    referenceID.String(),
	}

	res, err := testStore.CreateCustomerPayment(context.TODO(), &arg)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	individualDetail := res.GetIndividualDetail()
	require.Equal(t, arg.CustomerName, *individualDetail.GivenNames)
	_, mobileNumOk := res.GetMobileNumberOk()
	require.True(t, mobileNumOk)

	require.Equal(t, arg.ReferenceID, res.GetReferenceId())

	_, phoneNumberOK := res.GetPhoneNumberOk()
	require.True(t, phoneNumberOK)

	return res
}
