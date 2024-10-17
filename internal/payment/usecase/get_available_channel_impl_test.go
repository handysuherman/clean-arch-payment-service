package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository/mock"
	wkmock "github.com/handysuherman/clean-arch-payment-service/internal/payment/worker/mock"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/payment"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_MOCK_GET_AVAILABLE_CHANNEL(t *testing.T) {
	okParams, okResp := createRandomPaymentChannel(t)
	typeNotAvaiableParams, _ := createRandomPaymentChannel(t)
	typeNotAvaiableParams.PcType = "AKSDAOISDJG0IAJDGIJAWRGIAJRG"

	testCases := []struct {
		tname         string
		body          *models.GetPaymentChannelRequest
		stubs         func(store *mock.MockRepository)
		checkResponse func(t *testing.T, res *pb.GetPaymentChannelResponse, err error)
	}{
		{
			tname: "OK",
			body: &models.GetPaymentChannelRequest{
				Amount:             okParams.MinAmount.InexactFloat64(),
				PaymentChannelName: okParams.Pcname,
				PaymentChannelType: okParams.PcType,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetAvailablePaymentChannel(gomock.Any(), EqGetAvailablePaymentChannelParamsMatcher(okParams)).Times(1).Return(okResp, nil)
			},
			checkResponse: func(t *testing.T, res *pb.GetPaymentChannelResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.Equal(t, res.GetPaymentChannel().GetUid(), okResp.Uid)
				require.Equal(t, res.GetPaymentChannel().GetPaymentChannelType(), okResp.PcType)
				require.Equal(t, res.GetPaymentChannel().GetPaymentChannelName(), okResp.Pcname)
				require.Equal(t, res.GetPaymentChannel().GetTax(), okResp.Tax.InexactFloat64())
				require.Equal(t, res.GetPaymentChannel().GetMinAmount(), okResp.MinAmount.InexactFloat64())
				require.Equal(t, res.GetPaymentChannel().GetMaxAmount(), okResp.MaxAmount.InexactFloat64())
				require.Equal(t, res.GetPaymentChannel().GetIsTaxPercentage(), okResp.IsTaxPercentage)
				require.Equal(t, res.GetPaymentChannel().GetIsActive(), okResp.IsActive)
				require.Equal(t, res.GetPaymentChannel().GetIsAvailable(), okResp.IsAvailable)
			},
		},
		{
			tname: "ERR_TYPE_NOT_AVAILABLE",
			body: &models.GetPaymentChannelRequest{
				Amount:             typeNotAvaiableParams.MinAmount.InexactFloat64(),
				PaymentChannelName: typeNotAvaiableParams.Pcname,
				PaymentChannelType: typeNotAvaiableParams.PcType,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetAvailablePaymentChannel(gomock.Any(), EqGetAvailablePaymentChannelParamsMatcher(typeNotAvaiableParams)).Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.GetPaymentChannelResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_NOT_FOUND",
			body: &models.GetPaymentChannelRequest{
				Amount:             okParams.MinAmount.InexactFloat64(),
				PaymentChannelName: okParams.Pcname,
				PaymentChannelType: okParams.PcType,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetAvailablePaymentChannel(gomock.Any(), EqGetAvailablePaymentChannelParamsMatcher(okParams)).Times(1).Return(nil, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, res *pb.GetPaymentChannelResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_INTERNAL_SERVER_ERROR",
			body: &models.GetPaymentChannelRequest{
				Amount:             okParams.MinAmount.InexactFloat64(),
				PaymentChannelName: okParams.Pcname,
				PaymentChannelType: okParams.PcType,
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetAvailablePaymentChannel(gomock.Any(), EqGetAvailablePaymentChannelParamsMatcher(okParams)).Times(1).Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.GetPaymentChannelResponse, err error) {
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

			actualBody, actualError := u.GetAvailableChannel(context.TODO(), tc.body)
			tc.checkResponse(t, actualBody, actualError)
		})
	}
}

func createRandomPaymentChannel(t *testing.T) (*repository.GetAvailablePaymentChannelParams, *repository.PaymentChannel) {
	ulid, err := helper.GenerateULID()
	require.NoError(t, err)
	require.NotEmpty(t, ulid)

	paymentChannel := &repository.PaymentChannel{
		Uid:             ulid.String(),
		Pcname:          helper.RandomString(32),
		PcType:          payment.METHODE_TYPE_VIRTUAL_ACCOUNT,
		MinAmount:       decimal.NewFromInt(helper.RandomInt(10, 100)),
		MaxAmount:       decimal.NewFromInt(helper.RandomInt(100000, 200000)),
		Tax:             decimal.NewFromInt(helper.RandomInt(1, 10)),
		IsTaxPercentage: true,
		IsActive:        true,
		IsAvailable:     true,
	}

	params := &repository.GetAvailablePaymentChannelParams{
		PcType:    paymentChannel.PcType,
		Pcname:    paymentChannel.Pcname,
		MinAmount: decimal.NewFromInt(helper.RandomInt(int64(paymentChannel.MinAmount.IntPart()), int64(paymentChannel.MaxAmount.IntPart()))),
	}

	return params, paymentChannel
}

type eqGetAvailablePaymentChannelParamsMatcher struct {
	arg *repository.GetAvailablePaymentChannelParams
}

func EqGetAvailablePaymentChannelParamsMatcher(arg *repository.GetAvailablePaymentChannelParams) gomock.Matcher {
	return &eqGetAvailablePaymentChannelParamsMatcher{arg: arg}
}

func (ex *eqGetAvailablePaymentChannelParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(*repository.GetAvailablePaymentChannelParams)
	if !ok {
		return false
	}

	ex.arg.Pcname = arg.Pcname
	ex.arg.PcType = arg.PcType
	ex.arg.MinAmount = arg.MinAmount

	return reflect.DeepEqual(ex.arg, arg)
}

func (ex *eqGetAvailablePaymentChannelParamsMatcher) String() string {
	var errMsg string

	return errMsg + fmt.Sprintf("matches arg: %v", ex.arg)
}
