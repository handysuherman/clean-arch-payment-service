package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository/mock"
	wkmock "github.com/handysuherman/clean-arch-payment-service/internal/payment/worker/mock"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/payment"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/xendit/xendit-go/v5/payment_method"
	"go.uber.org/mock/gomock"
)

func Test_MOCK_CREATE_VIRTUAL_ACCOUNT_BANK_PAYMENT(t *testing.T) {
	custParamsOK, custRespOK := createRandomCustomer(t)

	paymentEwalletParamsOK, paymentEwalletRespOK := createRandomEwalletPayment(t)
	paymentQrCodeParamsOK, paymentQrCodeRespOK := createRandomQrCodePayment(t)
	paymentVirtualAccountBankParamsOK, paymentVirtualAccountBankRespOK := createRandomVirtualAccountBankPayment(t)

	idempotentKey, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, idempotentKey)

	mockRes := createMockRes(t, custRespOK, paymentVirtualAccountBankRespOK)

	testCases := []struct {
		tname         string
		body          *models.CreatePaymentRequest
		stubs         func(store *mock.MockRepository)
		checkResponse func(t *testing.T, res *pb.CreatePaymentResponse, err error)
	}{
		{
			tname: "OK_VIRTUAL_ACCOUNT_BANK_IDEMPOTENT_REQUEST",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(mockRes, nil)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.NotEmpty(t, res.GetCustomer().GetUid())
				require.NotEmpty(t, res.GetCustomer().GetPaymentCustomerId())
				require.NotEmpty(t, res.GetCustomer().GetPhoneNumber())
				_, err = ulid.Parse(res.GetCustomer().GetUid())
				require.NoError(t, err)

				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentMethodId())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentChannel())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentAmount())

				require.Equal(t, res.GetCustomer().GetPhoneNumber(), custRespOK.PhoneNumber.String)

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentVirtualAccountBankRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentVirtualAccountBankRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentVirtualAccountBankRespOK.PaymentReusability)
				amount, _ := paymentVirtualAccountBankRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentVirtualAccountBankRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_VIRTUAL_ACCOUNT_BANK_CUSTOMER_EXISTS_REDIS_OK",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(custRespOK, nil)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(1).Return(repository.CreateVirtualAccountBankPaymentTxResult{Payment: paymentVirtualAccountBankRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(1)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(1)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.NotEmpty(t, res.GetCustomer().GetUid())
				require.NotEmpty(t, res.GetCustomer().GetPaymentCustomerId())
				require.NotEmpty(t, res.GetCustomer().GetPhoneNumber())
				_, err = ulid.Parse(res.GetCustomer().GetUid())
				require.NoError(t, err)

				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentMethodId())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentChannel())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentAmount())

				require.Equal(t, res.GetCustomer().GetPhoneNumber(), custRespOK.PhoneNumber.String)

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentVirtualAccountBankRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentVirtualAccountBankRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentVirtualAccountBankRespOK.PaymentReusability)
				amount, _ := paymentVirtualAccountBankRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentVirtualAccountBankRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_VIRTUAL_ACCOUNT_BANK_CUSTOMER_EXISTS_REDIS_NOT_OK",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(custRespOK, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(1).Return(repository.CreateVirtualAccountBankPaymentTxResult{Payment: paymentVirtualAccountBankRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(1)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(1)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.NotEmpty(t, res.GetCustomer().GetUid())
				require.NotEmpty(t, res.GetCustomer().GetPaymentCustomerId())
				require.NotEmpty(t, res.GetCustomer().GetPhoneNumber())
				_, err = ulid.Parse(res.GetCustomer().GetUid())
				require.NoError(t, err)

				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentMethodId())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentChannel())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentAmount())

				require.Equal(t, res.GetCustomer().GetPhoneNumber(), custRespOK.PhoneNumber.String)

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentVirtualAccountBankRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentVirtualAccountBankRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentVirtualAccountBankRespOK.PaymentReusability)
				amount, _ := paymentVirtualAccountBankRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentVirtualAccountBankRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_VIRTUAL_ACCOUNT_BANK_CUSTOMER_EXISTS_REDIS_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, redis.ErrClosed)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(custRespOK, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(1).Return(repository.CreateVirtualAccountBankPaymentTxResult{Payment: paymentVirtualAccountBankRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(1)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(1)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.NotEmpty(t, res.GetCustomer().GetUid())
				require.NotEmpty(t, res.GetCustomer().GetPaymentCustomerId())
				require.NotEmpty(t, res.GetCustomer().GetPhoneNumber())
				_, err = ulid.Parse(res.GetCustomer().GetUid())
				require.NoError(t, err)

				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentMethodId())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentChannel())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentAmount())

				require.Equal(t, res.GetCustomer().GetPhoneNumber(), custRespOK.PhoneNumber.String)

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentVirtualAccountBankRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentVirtualAccountBankRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentVirtualAccountBankRespOK.PaymentReusability)
				amount, _ := paymentVirtualAccountBankRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentVirtualAccountBankRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_VIRTUAL_ACCOUNT_BANK_CUSTOMER_NOT_EXISTS_OK",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{Customer: custRespOK}, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(1).Return(repository.CreateVirtualAccountBankPaymentTxResult{Payment: paymentVirtualAccountBankRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(1)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(1)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.NotEmpty(t, res.GetCustomer().GetUid())
				require.NotEmpty(t, res.GetCustomer().GetPaymentCustomerId())
				require.NotEmpty(t, res.GetCustomer().GetPhoneNumber())
				_, err = ulid.Parse(res.GetCustomer().GetUid())
				require.NoError(t, err)

				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentMethodId())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentChannel())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentAmount())

				require.Equal(t, res.GetCustomer().GetPhoneNumber(), custRespOK.PhoneNumber.String)

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentVirtualAccountBankRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentVirtualAccountBankRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentVirtualAccountBankRespOK.PaymentReusability)
				amount, _ := paymentVirtualAccountBankRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentVirtualAccountBankRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_PAYMENT_TYPE_NOT_EXISTS",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             "ASDKOGNSAIOJFGHSAOFHGSAKDFHGOSADFHG",
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_PAYMENT_TYPE_NOT_SUPPORTED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             "CRYPTOCURRENCY",
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.True(t, strings.Contains(strings.ToLower(err.Error()), strings.ToLower(unierror.ErrUnsupportedPaymentType.Error())))
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_PHONE_NUMBER_NOT_VALID_DIGITS",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     "Q3RGNQ29U3GRNVGW",
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.True(t, strings.Contains(strings.ToLower(err.Error()), strings.ToLower(unierror.ErrInvalidCustomerPhoneNumberInput.Error())))
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_AMOUNT_LESS_THAN_MINIMUM",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           float64(99),
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.True(t, strings.Contains(strings.ToLower(err.Error()), strings.ToLower(unierror.ErrInvalidAmount.Error())))
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_REFERENCE_ID_SHOULD_NOT_BE_EMPTY",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      "",
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           float64(99),
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.True(t, strings.Contains(strings.ToLower(err.Error()), strings.ToLower(unierror.ErrReferenceIDShouldNotBeEmpty.Error())))
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_INVALID_SUCCESS_RETURN_URL",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: "gsdfljglskdfjsgldkfjgsldkfj",
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.True(t, strings.Contains(strings.ToLower(err.Error()), strings.ToLower(unierror.ErrInvalidSuccessURL.Error())))
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_INVALID_FAILURE_RETURN_URL",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: "asdfgijasdfigjaisdfgjaofdigsadifhg",
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.True(t, strings.Contains(strings.ToLower(err.Error()), strings.ToLower(unierror.ErrInvalidFailureURL.Error())))
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_EXPIRY_PAYMENT_LESS_THAN_72_HOURS",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 1,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.True(t, strings.Contains(strings.ToLower(err.Error()), strings.ToLower(unierror.ErrExpiryLessThan3Days.Error())))
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_GET_CUSTOMER_EXISTS_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, sql.ErrConnDone)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_CREATE_CUSTOMER_NOT_EXISTS_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{}, sql.ErrConnDone)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_CREATE_CUSTOMER_NOT_EXISTS_TX_CLOSED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{}, pgx.ErrTxClosed)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_CREATE_CUSTOMER_NOT_EXISTS_TX_COMMIT_ROLLBACKED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{}, pgx.ErrTxCommitRollback)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_CREATE_PAYMENT_PAYMENT_CHANNEL_NOT_EXIST",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          "paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String()",
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{Customer: custRespOK}, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_CREATE_PAYMENT_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{Customer: custRespOK}, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(1).Return(repository.CreateVirtualAccountBankPaymentTxResult{}, sql.ErrConnDone)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_CREATE_PAYMENT_TX_CLOSED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{Customer: custRespOK}, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(1).Return(repository.CreateVirtualAccountBankPaymentTxResult{}, pgx.ErrTxClosed)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_VIRTUAL_ACCOUNT_BANK_CREATE_PAYMENT_TX_ROLLBACKED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentVirtualAccountBankRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Description,
				PaymentAmount:           paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.Amount,
				PaymentType:             paymentVirtualAccountBankRespOK.PaymentType,
				PaymentChannel:          paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentVirtualAccountBankParamsOK.CreateVirtualAccountBankPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{Customer: custRespOK}, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(0)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(1).Return(repository.CreateVirtualAccountBankPaymentTxResult{}, pgx.ErrTxCommitRollback)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

				store.EXPECT().PutCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String()), gomock.Eq(mockRes)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreatePaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.tname, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mock.NewMockRepository(storeCtrl)

			workerCtrl := gomock.NewController(t)
			defer workerCtrl.Finish()
			wkstore := wkmock.NewMockProducerWorker(workerCtrl)

			u := New(tlog, conf, store, wkstore)
			tc.stubs(store)

			actualBody, actualError := u.Create(context.TODO(), tc.body)
			tc.checkResponse(t, actualBody, actualError)
		})
	}
}

func createRandomVirtualAccountBankPayment(t *testing.T) (*repository.CreateVirtualAccountBankPaymentTxParams, *repository.PaymentMethod) {
	ulid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, ulid)

	customerPaymentID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, customerPaymentID)

	customerPaymentMethodID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, customerPaymentMethodID)

	createParams := &repository.CreateVirtualAccountBankPaymentTxParams{
		CreateVirtualAccountBankPayment: repository.CreateVirtualAccountBankPaymentParams{
			CustomerName:      helper.RandomString(30),
			CustomerPaymentID: customerPaymentID.String(),
			CustomerNumber:    "+628" + helper.RandomString(9),
			Description:       helper.RandomString(100),
			Amount:            float64(helper.RandomInt(100, 200000)),
			Expiry:            time.Now().Add(24 * 3 * time.Hour),
			ChannelCode:       payment_method.VIRTUALACCOUNTCHANNELCODE_BCA,
			SuccessReturnURL:  helper.RandomUrl(),
			FailureReturnURL:  helper.RandomUrl(),
		},
	}

	respParams := &repository.PaymentMethod{
		Uid:             ulid.String(),
		PaymentMethodID: customerPaymentMethodID.String(),
		PaymentRequestID: pgtype.Text{
			String: customerPaymentMethodID.String(),
			Valid:  true,
		},
		PaymentReferenceID: customerPaymentMethodID.String(),
		PaymentBusinessID:  customerPaymentMethodID.String(),
		PaymentCustomerID:  createParams.CreateVirtualAccountBankPayment.CustomerPaymentID,
		PaymentType:        payment.METHODE_TYPE_VIRTUAL_ACCOUNT,
		PaymentStatus:      payment.STATUS_ACTIVE,
		PaymentReusability: payment.USAGE_TYPE_ONE_TIME_USE,
		PaymentChannel:     string(createParams.CreateVirtualAccountBankPayment.ChannelCode),
		PaymentAmount:      decimal.NewFromFloat(createParams.CreateVirtualAccountBankPayment.Amount),
		PaymentQrCode: pgtype.Text{
			String: helper.RandomString(16),
			Valid:  true,
		},
		PaymentVirtualAccountNumber: pgtype.Text{
			String: helper.RandomString(16),
			Valid:  true,
		},
		PaymentUrl: pgtype.Text{
			String: helper.RandomUrl(),
			Valid:  true,
		},
		PaymentDescription: createParams.CreateVirtualAccountBankPayment.Description,
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  createParams.CreateVirtualAccountBankPayment.Expiry,
			Valid: true,
		},
	}

	return createParams, respParams
}

type eqCreateVirtualAccountBankPaymentTxParamsMatcher struct {
	arg *repository.CreateVirtualAccountBankPaymentTxParams
}

func EqCreateVirtualAccountBankPaymentTxParamsMatcher(arg *repository.CreateVirtualAccountBankPaymentTxParams) gomock.Matcher {
	return &eqCreateVirtualAccountBankPaymentTxParamsMatcher{arg: arg}
}

func (ex *eqCreateVirtualAccountBankPaymentTxParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(*repository.CreateVirtualAccountBankPaymentTxParams)
	if !ok {
		return false
	}

	ex.arg.CreateVirtualAccountBankPayment = arg.CreateVirtualAccountBankPayment
	ex.arg.CreateVirtualAccountBankPayment.CustomerName = arg.CreateVirtualAccountBankPayment.CustomerName
	ex.arg.CreateVirtualAccountBankPayment.ReferenceID = arg.CreateVirtualAccountBankPayment.ReferenceID
	ex.arg.CreateVirtualAccountBankPayment.Amount = arg.CreateVirtualAccountBankPayment.Amount
	ex.arg.CreateVirtualAccountBankPayment.CustomerNumber = arg.CreateVirtualAccountBankPayment.CustomerNumber
	ex.arg.CreateVirtualAccountBankPayment.CustomerPaymentID = arg.CreateVirtualAccountBankPayment.CustomerPaymentID
	ex.arg.CreateVirtualAccountBankPayment.Description = arg.CreateVirtualAccountBankPayment.Description
	ex.arg.CreateVirtualAccountBankPayment.Expiry = arg.CreateVirtualAccountBankPayment.Expiry
	ex.arg.CreateVirtualAccountBankPayment.SuccessReturnURL = arg.CreateVirtualAccountBankPayment.SuccessReturnURL
	ex.arg.CreateVirtualAccountBankPayment.FailureReturnURL = arg.CreateVirtualAccountBankPayment.FailureReturnURL

	if ex.arg.CreateVirtualAccountBankPayment.CustomerPaymentID == "" {
		return false
	}

	if len(ex.arg.CreateVirtualAccountBankPayment.CustomerName) > 30 {
		return false
	}

	if len(ex.arg.CreateVirtualAccountBankPayment.Description) > 100 {
		return false
	}

	if !reflect.DeepEqual(ex.arg, arg) {
		return false
	}

	return true
}

func (ex *eqCreateVirtualAccountBankPaymentTxParamsMatcher) String() string {
	var errMsg string

	if ex.arg.CreateVirtualAccountBankPayment.CustomerPaymentID == "" {
		errMsg += "CustomerPaymentID should not be empty\n"
	}

	if len(ex.arg.CreateVirtualAccountBankPayment.CustomerName) > 30 {
		errMsg += "CustomerName should not exceeding 30 characters"
	}

	if len(ex.arg.CreateVirtualAccountBankPayment.Description) > 100 {
		errMsg += "PaymentDescription should not exceeding 100 characters"
	}

	return errMsg + fmt.Sprintf("matches arg: %v", ex.arg)
}
