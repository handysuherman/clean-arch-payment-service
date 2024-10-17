package usecase

import (
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository/mock"
	wkmock "github.com/handysuherman/clean-arch-payment-service/internal/payment/worker/mock"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"go.uber.org/mock/gomock"
)

func TestOnConfigUpdate(t *testing.T) {
	testCases := []struct {
		tname  string
		key    string
		config *config.App
	}{
		{
			tname:  "OK",
			key:    helper.RandomString(12),
			config: conf,
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

			u.OnConfigUpdate(tc.key, tc.config)
		})
	}
}
