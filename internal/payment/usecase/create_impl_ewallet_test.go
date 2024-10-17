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

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/mapper"
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

func Test_MOCK_CREATE_EWALLET_PAYMENT(t *testing.T) {
	custParamsOK, custRespOK := createRandomCustomer(t)

	paymentEwalletParamsOK, paymentEwalletRespOK := createRandomEwalletPayment(t)
	paymentQrCodeParamsOK, paymentQrCodeRespOK := createRandomQrCodePayment(t)
	paymentVirtualAccountBankParamsOK, paymentVirtualAccountBankRespOK := createRandomVirtualAccountBankPayment(t)

	idempotentKey, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, idempotentKey)

	mockRes := createMockRes(t, custRespOK, paymentEwalletRespOK)

	testCases := []struct {
		tname         string
		body          *models.CreatePaymentRequest
		stubs         func(store *mock.MockRepository)
		checkResponse func(t *testing.T, res *pb.CreatePaymentResponse, err error)
	}{
		{
			tname: "OK_EWALLET_IDEMPOTENT_REQUEST",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentEwalletRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentEwalletRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentEwalletRespOK.PaymentReusability)
				amount, _ := paymentEwalletRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentEwalletRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_EWALLET_CUSTOMER_EXISTS_REDIS_OK",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(custRespOK, nil)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(1).Return(repository.CreateEwalletPaymentTxResult{Payment: paymentEwalletRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(1)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentEwalletRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentEwalletRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentEwalletRespOK.PaymentReusability)
				amount, _ := paymentEwalletRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentEwalletRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_EWALLET_CUSTOMER_EXISTS_REDIS_NOT_OK",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(custRespOK, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(1).Return(repository.CreateEwalletPaymentTxResult{Payment: paymentEwalletRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(1)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentEwalletRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentEwalletRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentEwalletRespOK.PaymentReusability)
				amount, _ := paymentEwalletRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentEwalletRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_EWALLET_CUSTOMER_EXISTS_REDIS_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, redis.ErrClosed)
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(custRespOK, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(0)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(1).Return(repository.CreateEwalletPaymentTxResult{Payment: paymentEwalletRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(1)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentEwalletRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentEwalletRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentEwalletRespOK.PaymentReusability)
				amount, _ := paymentEwalletRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentEwalletRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_EWALLET_CUSTOMER_NOT_EXISTS_OK",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{Customer: custRespOK}, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(1).Return(repository.CreateEwalletPaymentTxResult{Payment: paymentEwalletRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentEwalletRespOK)).Times(1)

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(0)

				store.EXPECT().CreateVirtualAccountBankPaymentTx(gomock.Any(), EqCreateVirtualAccountBankPaymentTxParamsMatcher(paymentVirtualAccountBankParamsOK)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)

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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentEwalletRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentEwalletRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentEwalletRespOK.PaymentReusability)
				amount, _ := paymentEwalletRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentEwalletRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "ERR_EWALLET_PAYMENT_TYPE_NOT_EXISTS",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             "ASDKOGNSAIOJFGHSAOFHGSAKDFHGOSADFHG",
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_PAYMENT_TYPE_NOT_SUPPORTED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             "CRYPTOCURRENCY",
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_PHONE_NUMBER_NOT_VALID_DIGITS",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     "Q3RGNQ29U3GRNVGW",
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_AMOUNT_LESS_THAN_MINIMUM",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           float64(99),
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_REFERENCE_ID_SHOULD_NOT_BE_EMPTY",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      "",
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           float64(99),
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_INVALID_SUCCESS_RETURN_URL",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: "gsdfljglskdfjsgldkfjgsldkfj",
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_INVALID_FAILURE_RETURN_URL",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
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
			tname: "ERR_EWALLET_EXPIRY_PAYMENT_LESS_THAN_72_HOURS",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 1,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_GET_CUSTOMER_EXISTS_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_CREATE_CUSTOMER_NOT_EXISTS_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_CREATE_CUSTOMER_NOT_EXISTS_TX_CLOSED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_CREATE_CUSTOMER_NOT_EXISTS_TX_COMMIT_ROLLBACKED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_CREATE_PAYMENT_PAYMENT_CHANNEL_NOT_EXIST",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          "paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String()",
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
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
			tname: "ERR_EWALLET_CREATE_PAYMENT_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{Customer: custRespOK}, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(1).Return(repository.CreateEwalletPaymentTxResult{}, sql.ErrConnDone)
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
			tname: "ERR_EWALLET_CREATE_PAYMENT_TX_CLOSED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{Customer: custRespOK}, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(1).Return(repository.CreateEwalletPaymentTxResult{}, pgx.ErrTxClosed)
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
			tname: "ERR_EWALLET_CREATE_PAYMENT_TX_ROLLBACKED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				PaymentReferenceId:      paymentEwalletRespOK.Uid,
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentEwalletParamsOK.CreateEwalletPayment.Description,
				PaymentAmount:           paymentEwalletParamsOK.CreateEwalletPayment.Amount,
				PaymentType:             paymentEwalletRespOK.PaymentType,
				PaymentChannel:          paymentEwalletParamsOK.CreateEwalletPayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentEwalletParamsOK.CreateEwalletPayment.FailureReturnURL,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCreatePaymentIdempotencyKey(gomock.Any(), gomock.Eq(idempotentKey.String())).Times(1).Return(nil, redis.ErrClosed)

				store.EXPECT().GetCustomerCache(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, errors.New("not-found"))
				store.EXPECT().GetCustomerByCustomerAppID(gomock.Any(), gomock.Eq(custRespOK.Uid)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(0)

				store.EXPECT().CreateCustomerTx(gomock.Any(), EqCreateCustomerTxParamsMatcher(custParamsOK)).Times(1).Return(repository.CreateCustomerTxResult{Customer: custRespOK}, nil)
				store.EXPECT().PutCustomerCache(gomock.Any(), gomock.Eq(custRespOK)).Times(1)

				store.EXPECT().CreateEwalletPaymentTx(gomock.Any(), EqCreateEwalletPaymentTxParamsMatcher(paymentEwalletParamsOK)).Times(1).Return(repository.CreateEwalletPaymentTxResult{}, pgx.ErrTxCommitRollback)
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

func createRandomCustomer(t *testing.T) (*repository.CreateCustomerTxParams, *repository.Customer) {
	ulid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, ulid.String())

	paymentCustomerID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, paymentCustomerID.String())

	createParams := &repository.CreateCustomerTxParams{
		CreateCustomer: repository.CreateCustomerParams{
			Uid:               ulid.String(),
			CustomerAppID:     ulid.String(),
			PaymentCustomerID: paymentCustomerID.String(),
			CustomerName:      helper.RandomString(30),
			CreatedAt: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
			},
			PhoneNumber: pgtype.Text{
				String: "628" + helper.RandomStringInt(6),
				Valid:  true,
			},
		},
	}

	customerParams := &repository.Customer{
		Uid:               createParams.CreateCustomer.Uid,
		CustomerAppID:     createParams.CreateCustomer.CustomerAppID,
		PaymentCustomerID: createParams.CreateCustomer.PaymentCustomerID,
		CustomerName:      createParams.CreateCustomer.CustomerName,
		CreatedAt:         createParams.CreateCustomer.CreatedAt,
		Email:             createParams.CreateCustomer.Email,
		PhoneNumber:       createParams.CreateCustomer.PhoneNumber,
	}

	return createParams, customerParams
}

func createRandomEwalletPayment(t *testing.T) (*repository.CreateEwalletPaymentTxParams, *repository.PaymentMethod) {
	ulid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, ulid)

	customerPaymentID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, customerPaymentID)

	customerPaymentMethodID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, customerPaymentMethodID)

	createParams := &repository.CreateEwalletPaymentTxParams{
		CreateEwalletPayment: repository.CreateEwalletPaymentParams{
			CustomerName:      helper.RandomString(30),
			CustomerPaymentID: customerPaymentID.String(),
			CustomerNumber:    "628" + helper.RandomStringInt(9),
			Description:       helper.RandomString(100),
			Amount:            float64(helper.RandomInt(100, 200000)),
			Expiry:            time.Now().Add(24 * 3 * time.Hour),
			ChannelCode:       payment_method.EWALLETCHANNELCODE_OVO,
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
		PaymentCustomerID:  createParams.CreateEwalletPayment.CustomerPaymentID,
		PaymentType:        payment.METHODE_TYPE_EWALLET,
		PaymentStatus:      payment.STATUS_ACTIVE,
		PaymentReusability: payment.USAGE_TYPE_ONE_TIME_USE,
		PaymentChannel:     string(createParams.CreateEwalletPayment.ChannelCode),
		PaymentAmount:      decimal.NewFromFloat(createParams.CreateEwalletPayment.Amount),
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
		PaymentDescription: createParams.CreateEwalletPayment.Description,
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  createParams.CreateEwalletPayment.Expiry,
			Valid: true,
		},
	}

	return createParams, respParams
}

func createMockRes(t *testing.T, customer *repository.Customer, paymentMethod *repository.PaymentMethod) *pb.CreatePaymentResponse {
	return &pb.CreatePaymentResponse{
		Customer:      mapper.CustomerToDto(customer),
		PaymentMethod: mapper.PaymentToDto(paymentMethod),
	}
}

type eqCreateCustomerTxParamsMatcher struct {
	arg *repository.CreateCustomerTxParams
}

func EqCreateCustomerTxParamsMatcher(arg *repository.CreateCustomerTxParams) gomock.Matcher {
	return &eqCreateCustomerTxParamsMatcher{arg: arg}
}

func (ex *eqCreateCustomerTxParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(*repository.CreateCustomerTxParams)
	if !ok {
		return false
	}

	ex.arg.CreateCustomer.CustomerName = arg.CreateCustomer.CustomerName
	ex.arg.CreateCustomer.CustomerAppID = arg.CreateCustomer.CustomerAppID
	ex.arg.CreateCustomer.PaymentCustomerID = arg.CreateCustomer.PaymentCustomerID
	ex.arg.CreateCustomer.Uid = arg.CreateCustomer.Uid
	ex.arg.CreateCustomer.CustomerName = arg.CreateCustomer.CustomerName
	ex.arg.CreateCustomer.Email = arg.CreateCustomer.Email
	ex.arg.CreateCustomer.CreatedAt = arg.CreateCustomer.CreatedAt
	ex.arg.CreateCustomer.PhoneNumber = arg.CreateCustomer.PhoneNumber

	return reflect.DeepEqual(ex.arg, arg)
}

func (ex *eqCreateCustomerTxParamsMatcher) String() string {
	var errMsg string

	return errMsg + fmt.Sprintf("matches arg: %v", ex.arg)
}

type eqCreateEwalletPaymentTxParamsMatcher struct {
	arg *repository.CreateEwalletPaymentTxParams
}

func EqCreateEwalletPaymentTxParamsMatcher(arg *repository.CreateEwalletPaymentTxParams) gomock.Matcher {
	return &eqCreateEwalletPaymentTxParamsMatcher{arg: arg}
}

func (ex *eqCreateEwalletPaymentTxParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(*repository.CreateEwalletPaymentTxParams)
	if !ok {
		return false
	}

	ex.arg.CreateEwalletPayment = arg.CreateEwalletPayment
	ex.arg.CreateEwalletPayment.CustomerName = arg.CreateEwalletPayment.CustomerName
	ex.arg.CreateEwalletPayment.ReferenceID = arg.CreateEwalletPayment.ReferenceID
	ex.arg.CreateEwalletPayment.Amount = arg.CreateEwalletPayment.Amount
	ex.arg.CreateEwalletPayment.CustomerNumber = arg.CreateEwalletPayment.CustomerNumber
	ex.arg.CreateEwalletPayment.CustomerPaymentID = arg.CreateEwalletPayment.CustomerPaymentID
	ex.arg.CreateEwalletPayment.Description = arg.CreateEwalletPayment.Description
	ex.arg.CreateEwalletPayment.Expiry = arg.CreateEwalletPayment.Expiry
	ex.arg.CreateEwalletPayment.SuccessReturnURL = arg.CreateEwalletPayment.SuccessReturnURL
	ex.arg.CreateEwalletPayment.FailureReturnURL = arg.CreateEwalletPayment.FailureReturnURL

	if ex.arg.CreateEwalletPayment.CustomerPaymentID == "" {
		return false
	}

	if len(ex.arg.CreateEwalletPayment.CustomerName) > 30 {
		return false
	}

	if len(ex.arg.CreateEwalletPayment.Description) > 100 {
		return false
	}

	if !reflect.DeepEqual(ex.arg, arg) {
		return false
	}

	return true
}

func (ex *eqCreateEwalletPaymentTxParamsMatcher) String() string {
	var errMsg string

	if ex.arg.CreateEwalletPayment.CustomerPaymentID == "" {
		errMsg += "CustomerPaymentID should not be empty\n"
	}

	if len(ex.arg.CreateEwalletPayment.CustomerName) > 30 {
		errMsg += "CustomerName should not exceeding 30 characters"
	}

	if len(ex.arg.CreateEwalletPayment.Description) > 100 {
		errMsg += "PaymentDescription should not exceeding 100 characters"
	}

	return errMsg + fmt.Sprintf("matches arg: %v", ex.arg)
}
