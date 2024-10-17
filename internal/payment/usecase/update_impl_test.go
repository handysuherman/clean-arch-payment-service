package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository/mock"
	wkmock "github.com/handysuherman/clean-arch-payment-service/internal/payment/worker/mock"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/payment"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_MOCK_UPDATE_PAYMENT(t *testing.T) {
	_, paymentVirtualAccountBankRespOK := createRandomVirtualAccountBankPayment(t)
	okArg := createRandomUpdateArg(t, paymentVirtualAccountBankRespOK)

	_, paymentVirtualAccountBankRespStatusSucceeded := createRandomVirtualAccountBankPayment(t)
	paymentVirtualAccountBankRespStatusSucceeded.PaymentStatus = payment.STATUS_SUCCEEDED

	_, paymentVirtualAccountBankRespStatusFailed := createRandomVirtualAccountBankPayment(t)
	paymentVirtualAccountBankRespStatusFailed.PaymentStatus = payment.STATUS_FAILED

	testCases := []struct {
		tname         string
		body          *models.UpdatePaymentRequest
		stub          func(store *mock.MockRepository, wkstore *wkmock.MockProducerWorker)
		checkResponse func(t *testing.T, err error)
	}{
		{
			tname: "OK",
			body: &models.UpdatePaymentRequest{
				PaymentEvent:       helper.RandomString(32),
				PaymentType:        helper.RandomString(32),
				PaymentCustomerId:  paymentVirtualAccountBankRespOK.PaymentCustomerID,
				PaymentMethodId:    paymentVirtualAccountBankRespOK.PaymentMethodID,
				PaymentBusinessId:  helper.RandomString(32),
				PaymentChannel:     helper.RandomString(32),
				UpdatedAt:          &paymentVirtualAccountBankRespOK.UpdatedAt.Time,
				PaymentStatus:      okArg.UpdateParams.PaymentStatus.String,
				PaymentFailureCode: &paymentVirtualAccountBankRespOK.PaymentDescription,
			},
			stub: func(store *mock.MockRepository, wkstore *wkmock.MockProducerWorker) {
				store.EXPECT().UpdateTx(gomock.Any(), EqUpdateTxParamsMatcher(okArg)).Times(1).Return(repository.UpdateTxResult{Payment: paymentVirtualAccountBankRespOK}, nil)
				wkstore.EXPECT().PaymentStatusUpdated(gomock.Any(), EqPaymentStatusUpdatedParams(&models.PaymentStatusUpdatedTask{PaymentMethod: paymentVirtualAccountBankRespOK})).Times(1)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(1)
			},
			checkResponse: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			tname: "ERR_UPDATE_PAYMENT_METHOD_NOT_FOUND",
			body: &models.UpdatePaymentRequest{
				PaymentEvent:       helper.RandomString(32),
				PaymentType:        helper.RandomString(32),
				PaymentCustomerId:  paymentVirtualAccountBankRespOK.PaymentCustomerID,
				PaymentMethodId:    paymentVirtualAccountBankRespOK.PaymentMethodID,
				PaymentBusinessId:  helper.RandomString(32),
				PaymentChannel:     helper.RandomString(32),
				UpdatedAt:          &paymentVirtualAccountBankRespOK.UpdatedAt.Time,
				PaymentStatus:      okArg.UpdateParams.PaymentStatus.String,
				PaymentFailureCode: &paymentVirtualAccountBankRespOK.PaymentDescription,
			},
			stub: func(store *mock.MockRepository, wkstore *wkmock.MockProducerWorker) {
				store.EXPECT().UpdateTx(gomock.Any(), EqUpdateTxParamsMatcher(okArg)).Times(1).Return(repository.UpdateTxResult{}, pgx.ErrNoRows)
				wkstore.EXPECT().PaymentStatusUpdated(gomock.Any(), EqPaymentStatusUpdatedParams(&models.PaymentStatusUpdatedTask{PaymentMethod: paymentVirtualAccountBankRespOK})).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)
			},
			checkResponse: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			tname: "ERR_UPDATE_PAYMENT_METHOD_INTERNAL_SERVER_ERROR",
			body: &models.UpdatePaymentRequest{
				PaymentEvent:       helper.RandomString(32),
				PaymentType:        helper.RandomString(32),
				PaymentCustomerId:  paymentVirtualAccountBankRespOK.PaymentCustomerID,
				PaymentMethodId:    paymentVirtualAccountBankRespOK.PaymentMethodID,
				PaymentBusinessId:  helper.RandomString(32),
				PaymentChannel:     helper.RandomString(32),
				UpdatedAt:          &paymentVirtualAccountBankRespOK.UpdatedAt.Time,
				PaymentStatus:      okArg.UpdateParams.PaymentStatus.String,
				PaymentFailureCode: &paymentVirtualAccountBankRespOK.PaymentDescription,
			},
			stub: func(store *mock.MockRepository, wkstore *wkmock.MockProducerWorker) {

				store.EXPECT().UpdateTx(gomock.Any(), EqUpdateTxParamsMatcher(okArg)).Times(1).Return(repository.UpdateTxResult{}, sql.ErrConnDone)
				wkstore.EXPECT().PaymentStatusUpdated(gomock.Any(), EqPaymentStatusUpdatedParams(&models.PaymentStatusUpdatedTask{PaymentMethod: paymentVirtualAccountBankRespOK})).Times(0)
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)
			},
			checkResponse: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			tname: "ERR_UPDATE_PAYMENT_METHOD_WORKER_ERROR",
			body: &models.UpdatePaymentRequest{
				PaymentEvent:       helper.RandomString(32),
				PaymentType:        helper.RandomString(32),
				PaymentCustomerId:  paymentVirtualAccountBankRespOK.PaymentCustomerID,
				PaymentMethodId:    paymentVirtualAccountBankRespOK.PaymentMethodID,
				PaymentBusinessId:  helper.RandomString(32),
				PaymentChannel:     helper.RandomString(32),
				UpdatedAt:          &paymentVirtualAccountBankRespOK.UpdatedAt.Time,
				PaymentStatus:      okArg.UpdateParams.PaymentStatus.String,
				PaymentFailureCode: &paymentVirtualAccountBankRespOK.PaymentDescription,
			},
			stub: func(store *mock.MockRepository, wkstore *wkmock.MockProducerWorker) {

				store.EXPECT().UpdateTx(gomock.Any(), EqUpdateTxParamsMatcher(okArg)).Times(1).Return(repository.UpdateTxResult{Payment: paymentVirtualAccountBankRespOK}, nil)
				wkstore.EXPECT().PaymentStatusUpdated(gomock.Any(), EqPaymentStatusUpdatedParams(&models.PaymentStatusUpdatedTask{PaymentMethod: paymentVirtualAccountBankRespOK})).Times(1).Return(errors.New("error"))
				store.EXPECT().PutCache(gomock.Any(), gomock.Eq(paymentVirtualAccountBankRespOK)).Times(0)
			},
			checkResponse: func(t *testing.T, err error) {
				require.Error(t, err)
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
			tc.stub(store, wkstore)

			actualError := u.Update(context.TODO(), tc.body)
			tc.checkResponse(t, actualError)
		})
	}
}

type eqUpdateTxParamsMatcher struct {
	arg *repository.UpdateTxParams
}

func EqUpdateTxParamsMatcher(arg *repository.UpdateTxParams) gomock.Matcher {
	return &eqUpdateTxParamsMatcher{arg: arg}
}

func (ex *eqUpdateTxParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(*repository.UpdateTxParams)
	if !ok {
		return false
	}

	emptyTime := pgtype.Timestamptz{}
	ex.arg.UpdateParams.PaymentCustomerID = arg.UpdateParams.PaymentCustomerID
	ex.arg.UpdateParams.PaymentFailureCode = arg.UpdateParams.PaymentFailureCode
	ex.arg.UpdateParams.PaymentMethodID = arg.UpdateParams.PaymentMethodID
	ex.arg.UpdateParams.PaymentStatus = arg.UpdateParams.PaymentStatus
	ex.arg.UpdateParams.UpdatedAt.Time = arg.UpdateParams.UpdatedAt.Time

	if ex.arg.UpdateParams.PaymentCustomerID == "" {
		return false
	}

	if ex.arg.UpdateParams.PaymentMethodID == "" {
		return false
	}

	if ex.arg.UpdateParams.PaymentStatus.String == payment.STATUS_SUCCEEDED {
		if ex.arg.UpdateParams.PaidAt == emptyTime {
			return false
		}
	}

	return reflect.DeepEqual(ex.arg, arg)
}

func (ex *eqUpdateTxParamsMatcher) String() string {
	var errMsg string
	emptyTime := pgtype.Timestamptz{}
	if ex.arg.UpdateParams.PaymentCustomerID == "" {
		errMsg += fmt.Sprintf("PaymentCustomerID should not be empty %v\n", ex.arg.UpdateParams.PaymentCustomerID)
	}

	if ex.arg.UpdateParams.PaymentMethodID == "" {
		errMsg += fmt.Sprintf("PaymentCustomerID should not be empty%v\n", ex.arg.UpdateParams.PaymentMethodID)
	}

	if ex.arg.UpdateParams.PaymentStatus.String == payment.STATUS_SUCCEEDED {
		if ex.arg.UpdateParams.PaidAt == emptyTime {
			errMsg += fmt.Sprintf("PaidAt should not be empty if payment status was succeeded %v\n", ex.arg.UpdateParams.PaymentStatus.String)
		}
	}

	return errMsg + fmt.Sprintf("matches arg: %v", ex.arg)
}

type eqPaymentStatusUpdatedParamsMatcher struct {
	arg *models.PaymentStatusUpdatedTask
}

func EqPaymentStatusUpdatedParams(arg *models.PaymentStatusUpdatedTask) gomock.Matcher {
	return eqPaymentStatusUpdatedParamsMatcher{arg}
}

func (ex eqPaymentStatusUpdatedParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(*models.PaymentStatusUpdatedTask)
	if !ok {
		return false
	}

	ex.arg.PaymentMethod = arg.PaymentMethod

	if arg.PaymentMethod == nil {
		return false
	}

	if !reflect.DeepEqual(ex.arg.PaymentMethod, arg.PaymentMethod) {
		return false
	}

	return true
}

func (ex eqPaymentStatusUpdatedParamsMatcher) String() string {
	var errMsg string
	if ex.arg.PaymentMethod == nil {
		errMsg += "PaymentMethod should not be empty/nil"
	}

	return errMsg + fmt.Sprintf("matches arg: %v", ex.arg)
}

func createRandomUpdateArg(t *testing.T, paymentMethod *repository.PaymentMethod) *repository.UpdateTxParams {
	return &repository.UpdateTxParams{
		UpdateParams: repository.UpdatePaymentMethodCustomerParams{
			PaymentMethodID:   paymentMethod.PaymentMethodID,
			PaymentCustomerID: paymentMethod.PaymentCustomerID,
			PaymentStatus: pgtype.Text{
				String: paymentMethod.PaymentStatus,
				Valid:  true,
			},
			PaymentFailureCode: pgtype.Text{
				String: helper.RandomString(32),
				Valid:  true,
			},
			UpdatedAt: pgtype.Timestamptz{
				Time:  paymentMethod.ExpiresAt.Time,
				Valid: true,
			},
		},
	}
}
