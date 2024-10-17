package usecase

import (
	"context"
	"database/sql"
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository/mock"
	wkmock "github.com/handysuherman/clean-arch-payment-service/internal/payment/worker/mock"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_MOCK_GET_PAYMENT_BY_ID(t *testing.T) {
	_, paymentVirtualAccountBankRespOK := createRandomVirtualAccountBankPayment(t)

	arg := repository.GetPaymentMethodCustomerParams{
		PaymentCustomerID: paymentVirtualAccountBankRespOK.PaymentCustomerID,
		PaymentMethodID:   paymentVirtualAccountBankRespOK.PaymentMethodID,
	}

	testCases := []struct {
		tname         string
		body          *models.GetByIDPaymentRequest
		stubs         func(store *mock.MockRepository)
		checkResponse func(t *testing.T, res *pb.GetByIDPaymentResponse, err error)
	}{
		{
			tname: "OK_REDIS",
			body: &models.GetByIDPaymentRequest{
				PaymentCustomerId: paymentVirtualAccountBankRespOK.PaymentCustomerID,
				PaymentMethodId:   paymentVirtualAccountBankRespOK.PaymentMethodID,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK.PaymentCustomerID), gomock.Eq(paymentVirtualAccountBankRespOK.PaymentMethodID)).Times(1).Return(paymentVirtualAccountBankRespOK, nil)
				store.EXPECT().GetPaymentMethodCustomer(gomock.Any(), gomock.Eq(&arg)).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.GetByIDPaymentResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentMethodId())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentChannel())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentAmount())

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentVirtualAccountBankRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentVirtualAccountBankRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentVirtualAccountBankRespOK.PaymentReusability)
				amount, _ := paymentVirtualAccountBankRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentVirtualAccountBankRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "OK_NOT_REDIS",
			body: &models.GetByIDPaymentRequest{
				PaymentCustomerId: paymentVirtualAccountBankRespOK.PaymentCustomerID,
				PaymentMethodId:   paymentVirtualAccountBankRespOK.PaymentMethodID,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK.PaymentCustomerID), gomock.Eq(paymentVirtualAccountBankRespOK.PaymentMethodID)).Times(1).Return(nil, redis.ErrClosed)
				store.EXPECT().GetPaymentMethodCustomer(gomock.Any(), gomock.Eq(&arg)).Times(1).Return(paymentVirtualAccountBankRespOK, nil)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(1)
			},
			checkResponse: func(t *testing.T, res *pb.GetByIDPaymentResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentMethodId())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentChannel())
				require.NotEmpty(t, res.GetPaymentMethod().GetPaymentAmount())

				require.Equal(t, res.GetPaymentMethod().GetPaymentChannel(), paymentVirtualAccountBankRespOK.PaymentChannel)
				require.Equal(t, res.GetPaymentMethod().GetPaymentType(), paymentVirtualAccountBankRespOK.PaymentType)
				require.Equal(t, res.GetPaymentMethod().GetPaymentReusability(), paymentVirtualAccountBankRespOK.PaymentReusability)
				amount, _ := paymentVirtualAccountBankRespOK.PaymentAmount.Float64()
				require.Equal(t, res.GetPaymentMethod().GetPaymentAmount(), amount)
				require.Equal(t, res.GetPaymentMethod().GetPaymentUrl(), paymentVirtualAccountBankRespOK.PaymentUrl.String)
			},
		},
		{
			tname: "ERR_NOT_FOUND",
			body: &models.GetByIDPaymentRequest{
				PaymentCustomerId: paymentVirtualAccountBankRespOK.PaymentCustomerID,
				PaymentMethodId:   paymentVirtualAccountBankRespOK.PaymentMethodID,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK.PaymentCustomerID), gomock.Eq(paymentVirtualAccountBankRespOK.PaymentMethodID)).Times(1).Return(nil, redis.ErrClosed)
				store.EXPECT().GetPaymentMethodCustomer(gomock.Any(), gomock.Eq(&arg)).Times(1).Return(nil, pgx.ErrNoRows)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.GetByIDPaymentResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_INTERNAL_SERVER_ERROR",
			body: &models.GetByIDPaymentRequest{
				PaymentCustomerId: paymentVirtualAccountBankRespOK.PaymentCustomerID,
				PaymentMethodId:   paymentVirtualAccountBankRespOK.PaymentMethodID,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK.PaymentCustomerID), gomock.Eq(paymentVirtualAccountBankRespOK.PaymentMethodID)).Times(1).Return(nil, redis.ErrClosed)
				store.EXPECT().GetPaymentMethodCustomer(gomock.Any(), gomock.Eq(&arg)).Times(1).Return(nil, sql.ErrConnDone)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.GetByIDPaymentResponse, err error) {
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

			actualBody, actualError := u.GetByID(context.TODO(), tc.body)
			tc.checkResponse(t, actualBody, actualError)
		})
	}
}
