package worker

import (
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	messageClient "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka"
	producerMock "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka/mock"
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
			producerCtrl := gomock.NewController(t)
			defer producerCtrl.Finish()
			producerStore := producerMock.NewMockProducer(producerCtrl)

			u := New(tlog, conf, producerStore)

			u.OnConfigUpdate(tc.key, tc.config)
		})
	}
}

func TestTokenMakerUpdate(t *testing.T) {
	testCases := []struct {
		tname    string
		key      string
		producer *messageClient.ProducerImpl
	}{
		{
			tname:    "OK",
			key:      helper.RandomString(12),
			producer: producer,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.tname, func(t *testing.T) {

			producerCtrl := gomock.NewController(t)
			defer producerCtrl.Finish()
			producerStore := producerMock.NewMockProducer(producerCtrl)

			u := New(tlog, conf, producerStore)

			u.OnProducerWorkerUpdate(tc.key, tc.producer)
		})
	}
}
