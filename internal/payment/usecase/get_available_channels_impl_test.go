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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_MOCK_GET_AVAILABLE_CHANNELS(t *testing.T) {
	okParams, okResp := createRandomPaymentChannel(t)
	mockRes := []*repository.PaymentChannel{okResp}
	emptyRes := []*repository.PaymentChannel{}

	testCases := []struct {
		tname         string
		body          *models.GetPaymentChannelsRequest
		stubs         func(store *mock.MockRepository)
		checkResponse func(t *testing.T, res *pb.GetPaymentChannelsResponse, err error)
	}{
		{
			tname: "OK",
			body: &models.GetPaymentChannelsRequest{
				Amount: okParams.MinAmount.InexactFloat64(),
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetAvailablePaymentChannels(gomock.Any(), gomock.Eq(okParams.MinAmount)).Times(1).Return(mockRes, nil)
			},
			checkResponse: func(t *testing.T, res *pb.GetPaymentChannelsResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.Equal(t, len(mockRes), len(res.GetList()))

				for _, v := range res.GetList() {
					require.Equal(t, v.GetUid(), okResp.Uid)
					require.Equal(t, v.GetPaymentChannelType(), okResp.PcType)
					require.Equal(t, v.GetPaymentChannelName(), okResp.Pcname)
					require.Equal(t, v.GetTax(), okResp.Tax.InexactFloat64())
					require.Equal(t, v.GetMinAmount(), okResp.MinAmount.InexactFloat64())
					require.Equal(t, v.GetMaxAmount(), okResp.MaxAmount.InexactFloat64())
					require.Equal(t, v.GetIsTaxPercentage(), okResp.IsTaxPercentage)
					require.Equal(t, v.GetIsActive(), okResp.IsActive)
					require.Equal(t, v.GetIsAvailable(), okResp.IsAvailable)
				}
			},
		},
		{
			tname: "OK_EMPTY_RES",
			body: &models.GetPaymentChannelsRequest{
				Amount: okParams.MinAmount.InexactFloat64(),
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetAvailablePaymentChannels(gomock.Any(), gomock.Eq(okParams.MinAmount)).Times(1).Return(emptyRes, nil)
			},
			checkResponse: func(t *testing.T, res *pb.GetPaymentChannelsResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)

				require.Equal(t, len(emptyRes), len(res.GetList()))
			},
		},
		{
			tname: "ERR_NOT_FOUND",
			body: &models.GetPaymentChannelsRequest{
				Amount: okParams.MinAmount.InexactFloat64(),
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetAvailablePaymentChannels(gomock.Any(), gomock.Eq(okParams.MinAmount)).Times(1).Return(nil, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, res *pb.GetPaymentChannelsResponse, err error) {
				require.Error(t, err)
				require.Empty(t, res)
			},
		},
		{
			tname: "ERR_INTERNAL_SERVER_ERROR",
			body: &models.GetPaymentChannelsRequest{
				Amount: okParams.MinAmount.InexactFloat64(),
			},
			stubs: func(store *mock.MockRepository) {
				store.EXPECT().GetAvailablePaymentChannels(gomock.Any(), gomock.Eq(okParams.MinAmount)).Times(1).Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.GetPaymentChannelsResponse, err error) {
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

			actualBody, actualError := u.GetAvailableChannels(context.TODO(), tc.body)
			tc.checkResponse(t, actualBody, actualError)
		})
	}
}
