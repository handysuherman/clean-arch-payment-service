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

func Test_MOCK_CREATE_QRCODE_PAYMENT(t *testing.T) {
	custParamsOK, custRespOK := createRandomCustomer(t)

	paymentEwalletParamsOK, paymentEwalletRespOK := createRandomEwalletPayment(t)
	paymentQrCodeParamsOK, paymentQrCodeRespOK := createRandomQrCodePayment(t)
	paymentVirtualAccountBankParamsOK, paymentVirtualAccountBankRespOK := createRandomVirtualAccountBankPayment(t)

	idempotentKey, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, idempotentKey)

	mockRes := createMockRes(t, custRespOK, paymentQrCodeRespOK)

	testCases := []struct {
		tname         string
		body          *models.CreatePaymentRequest
		stubs         func(store *mock.MockRepository)
		checkResponse func(t *testing.T, res *pb.CreatePaymentResponse, err error)
	}{
		{
			tname: "OK_QRCODE_IDEMPOTENT_REQUEST",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentQrCodeRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentQrCodeRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentQrCodeRespOK.PaymentReusability)
				amount, _ := paymentQrCodeRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentQrCodeRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_QRCODE_CUSTOMER_EXISTS_REDIS_OK",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(1).Return(repository.CreateQrCodePaymentTxResult{Payment: paymentQrCodeRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(1)

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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentQrCodeRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentQrCodeRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentQrCodeRespOK.PaymentReusability)
				amount, _ := paymentQrCodeRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentQrCodeRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_QRCODE_CUSTOMER_EXISTS_REDIS_NOT_OK",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(1).Return(repository.CreateQrCodePaymentTxResult{Payment: paymentQrCodeRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(1)

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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentQrCodeRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentQrCodeRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentQrCodeRespOK.PaymentReusability)
				amount, _ := paymentQrCodeRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentQrCodeRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_QRCODE_CUSTOMER_EXISTS_REDIS_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(1).Return(repository.CreateQrCodePaymentTxResult{Payment: paymentQrCodeRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(1)

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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentQrCodeRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentQrCodeRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentQrCodeRespOK.PaymentReusability)
				amount, _ := paymentQrCodeRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentQrCodeRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_QRCODE_CUSTOMER_NOT_EXISTS_OK",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(1).Return(repository.CreateQrCodePaymentTxResult{Payment: paymentQrCodeRespOK}, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentQrCodeRespOK)).Times(1)

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

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentQrCodeRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentQrCodeRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentQrCodeRespOK.PaymentReusability)
				amount, _ := paymentQrCodeRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentQrCodeRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "ERR_QRCODE_PAYMENT_TYPE_NOT_EXISTS",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             "ASDKOGNSAIOJFGHSAOFHGSAKDFHGOSADFHG",
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_PAYMENT_TYPE_NOT_SUPPORTED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             "CRYPTOCURRENCY",
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_PHONE_NUMBER_NOT_VALID_DIGITS",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     "Q3RGNQ29U3GRNVGW",
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_AMOUNT_LESS_THAN_MINIMUM",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           float64(99),
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_REFERENCE_ID_SHOULD_NOT_BE_EMPTY",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           float64(99),
				PaymentReferenceId:      "",
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_INVALID_SUCCESS_RETURN_URL",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: "gsdfljglskdfjsgldkfjgsldkfj",
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_INVALID_FAILURE_RETURN_URL",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
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
			tname: "ERR_QRCODE_EXPIRY_PAYMENT_LESS_THAN_72_HOURS",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 1,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_GET_CUSTOMER_EXISTS_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_CREATE_CUSTOMER_NOT_EXISTS_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_CREATE_CUSTOMER_NOT_EXISTS_TX_CLOSED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_CREATE_CUSTOMER_NOT_EXISTS_TX_COMMIT_ROLLBACKED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_CREATE_PAYMENT_PAYMENT_CHANNEL_NOT_EXIST",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          "paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String()",
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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
			tname: "ERR_QRCODE_CREATE_PAYMENT_INTERNAL_SERVER_ERROR",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(1).Return(repository.CreateQrCodePaymentTxResult{}, sql.ErrConnDone)
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
			tname: "ERR_QRCODE_CREATE_PAYMENT_TX_CLOSED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(1).Return(repository.CreateQrCodePaymentTxResult{}, pgx.ErrTxClosed)
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
			tname: "ERR_QRCODE_CREATE_PAYMENT_TX_ROLLBACKED",
			body: &models.CreatePaymentRequest{
				XIdempotencyKey:         idempotentKey.String(),
				CustomerUid:             &custRespOK.Uid,
				CustomerName:            custRespOK.CustomerName,
				CustomerPhoneNumber:     custParamsOK.CreateCustomer.PhoneNumber.String,
				PaymentDescription:      paymentQrCodeParamsOK.CreateQrCodePayment.Description,
				PaymentAmount:           paymentQrCodeParamsOK.CreateQrCodePayment.Amount,
				PaymentReferenceId:      paymentQrCodeRespOK.Uid,
				PaymentType:             paymentQrCodeRespOK.PaymentType,
				PaymentChannel:          paymentQrCodeParamsOK.CreateQrCodePayment.ChannelCode.String(),
				ExpiryHour:              24 * 3,
				PaymentSuccessReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.SuccessReturnURL,
				PaymentFailureReturnUrl: paymentQrCodeParamsOK.CreateQrCodePayment.FailureReturnURL,
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

				store.EXPECT().CreateQrCodePaymentTx(gomock.Any(), EqCreateQrCodePaymentTxParamsMatcher(paymentQrCodeParamsOK)).Times(1).Return(repository.CreateQrCodePaymentTxResult{}, pgx.ErrTxCommitRollback)
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

func createRandomQrCodePayment(t *testing.T) (*repository.CreateQrCodePaymentTxParams, *repository.PaymentMethod) {
	ulid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, ulid)

	customerPaymentID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, customerPaymentID)

	customerPaymentMethodID, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, customerPaymentMethodID)

	createParams := &repository.CreateQrCodePaymentTxParams{
		CreateQrCodePayment: repository.CreateQrCodePaymentParams{
			CustomerName:      helper.RandomString(30),
			CustomerPaymentID: customerPaymentID.String(),
			CustomerNumber:    "+628" + helper.RandomString(9),
			Description:       helper.RandomString(100),
			Amount:            float64(helper.RandomInt(100, 200000)),
			Expiry:            time.Now().Add(24 * 3 * time.Hour),
			ChannelCode:       payment_method.QRCODECHANNELCODE_QRIS,
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
		PaymentCustomerID:  createParams.CreateQrCodePayment.CustomerPaymentID,
		PaymentType:        payment.METHODE_TYPE_QR_CODE,
		PaymentStatus:      payment.STATUS_ACTIVE,
		PaymentReusability: payment.USAGE_TYPE_ONE_TIME_USE,
		PaymentChannel:     string(createParams.CreateQrCodePayment.ChannelCode),
		PaymentAmount:      decimal.NewFromFloat(createParams.CreateQrCodePayment.Amount),
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
		PaymentDescription: createParams.CreateQrCodePayment.Description,
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  createParams.CreateQrCodePayment.Expiry,
			Valid: true,
		},
	}

	return createParams, respParams
}

type eqCreateQrCodePaymentTxParamsMatcher struct {
	arg *repository.CreateQrCodePaymentTxParams
}

func EqCreateQrCodePaymentTxParamsMatcher(arg *repository.CreateQrCodePaymentTxParams) gomock.Matcher {
	return &eqCreateQrCodePaymentTxParamsMatcher{arg: arg}
}

func (ex *eqCreateQrCodePaymentTxParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(*repository.CreateQrCodePaymentTxParams)
	if !ok {
		return false
	}

	ex.arg.CreateQrCodePayment = arg.CreateQrCodePayment
	ex.arg.CreateQrCodePayment.CustomerName = arg.CreateQrCodePayment.CustomerName
	ex.arg.CreateQrCodePayment.ReferenceID = arg.CreateQrCodePayment.ReferenceID
	ex.arg.CreateQrCodePayment.Amount = arg.CreateQrCodePayment.Amount
	ex.arg.CreateQrCodePayment.CustomerNumber = arg.CreateQrCodePayment.CustomerNumber
	ex.arg.CreateQrCodePayment.CustomerPaymentID = arg.CreateQrCodePayment.CustomerPaymentID
	ex.arg.CreateQrCodePayment.Description = arg.CreateQrCodePayment.Description
	ex.arg.CreateQrCodePayment.Expiry = arg.CreateQrCodePayment.Expiry
	ex.arg.CreateQrCodePayment.SuccessReturnURL = arg.CreateQrCodePayment.SuccessReturnURL
	ex.arg.CreateQrCodePayment.FailureReturnURL = arg.CreateQrCodePayment.FailureReturnURL

	if ex.arg.CreateQrCodePayment.CustomerPaymentID == "" {
		return false
	}

	if len(ex.arg.CreateQrCodePayment.CustomerName) > 30 {
		return false
	}

	if len(ex.arg.CreateQrCodePayment.Description) > 100 {
		return false
	}

	if !reflect.DeepEqual(ex.arg, arg) {
		return false
	}

	return true
}

func (ex *eqCreateQrCodePaymentTxParamsMatcher) String() string {
	var errMsg string

	if ex.arg.CreateQrCodePayment.CustomerPaymentID == "" {
		errMsg += "CustomerPaymentID should not be empty\n"
	}

	if len(ex.arg.CreateQrCodePayment.CustomerName) > 30 {
		errMsg += "CustomerName should not exceeding 30 characters"
	}

	if len(ex.arg.CreateQrCodePayment.Description) > 100 {
		errMsg += "PaymentDescription should not exceeding 100 characters"
	}

	return errMsg + fmt.Sprintf("matches arg: %v", ex.arg)
}
