package worker

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	producerMock "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka/mock"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/payment"
	messages "github.com/handysuherman/clean-arch-payment-service/internal/proto/kafka"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/xendit/xendit-go/v5/payment_method"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_MOCK_PAYMENT_METHOD_UPDATED(t *testing.T) {
	_, paymentVirtualAccountBankRespOK := createRandomVirtualAccountBankPayment(t)
	okTopic := helper.StringBuilder(conf.Services.Internal.ID, "_", conf.Brokers.Kafka.Topics.PaymentStatusUpdated.TopicName)

	okParams := createKafkaMessageParams(t, okTopic, paymentVirtualAccountBankRespOK)

	testCases := []struct {
		tname         string
		body          *models.PaymentStatusUpdatedTask
		stub          func(producerStore *producerMock.MockProducer)
		checkResponse func(t *testing.T, err error)
	}{
		{
			tname: "OK",
			body: &models.PaymentStatusUpdatedTask{
				PaymentMethod: paymentVirtualAccountBankRespOK,
			},
			stub: func(producerStore *producerMock.MockProducer) {
				producerStore.EXPECT().PublishMessage(gomock.Any(), EqKafkaMessageParamsMatcher(okParams, okTopic)).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			tname: "NOT_OK",
			body: &models.PaymentStatusUpdatedTask{
				PaymentMethod: paymentVirtualAccountBankRespOK,
			},
			stub: func(producerStore *producerMock.MockProducer) {
				producerStore.EXPECT().PublishMessage(gomock.Any(), EqKafkaMessageParamsMatcher(okParams, okTopic)).Times(1).Return(errors.New("any err"))
			},
			checkResponse: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.tname, func(t *testing.T) {
			producerStoreCtrl := gomock.NewController(t)
			defer producerStoreCtrl.Finish()
			producerStore := producerMock.NewMockProducer(producerStoreCtrl)

			u := New(tlog, conf, producerStore)
			tc.stub(producerStore)

			actualError := u.PaymentStatusUpdated(context.TODO(), tc.body)
			tc.checkResponse(t, actualError)
		})
	}
}

func createKafkaMessageParams(t *testing.T, topic string, task *repository.PaymentMethod) kafka.Message {
	amount, _ := task.PaymentAmount.Float64()

	arg := messages.KafkaPaymentStatusUpdated{
		Uid:                         task.Uid,
		PaymentMethodId:             task.PaymentMethodID,
		PaymentRequestId:            &task.PaymentRequestID.String,
		PaymentReferenceId:          task.PaymentReferenceID,
		PaymentBusinessId:           task.PaymentBusinessID,
		PaymentCustomerId:           task.PaymentCustomerID,
		PaymentType:                 task.PaymentType,
		PaymentStatus:               task.PaymentStatus,
		PaymentReusability:          task.PaymentReusability,
		PaymentChannel:              task.PaymentChannel,
		PaymentAmount:               amount,
		PaymentQrCode:               &task.PaymentQrCode.String,
		PaymentVirtualAccountNumber: &task.PaymentVirtualAccountNumber.String,
		PaymentUrl:                  &task.PaymentUrl.String,
		PaymentDescription:          task.PaymentDescription,
		CreatedAt:                   timestamppb.New(task.CreatedAt.Time),
		UpdatedAt:                   timestamppb.New(task.UpdatedAt.Time),
		ExpiresAt:                   timestamppb.New(task.ExpiresAt.Time),
	}

	protoMsg, err := proto.Marshal(&arg)
	require.NoError(t, err)
	require.NotEmpty(t, &arg)

	return kafka.Message{
		Topic: topic,
		Value: protoMsg,
		Time:  time.Now().UTC(),
	}
}

func EqKafkaMessageParamsMatcher(arg kafka.Message, expectedTopic string) gomock.Matcher {
	return &eqKafkaMessageParamsMatcher{arg: arg, expectedTopic: expectedTopic}
}

type eqKafkaMessageParamsMatcher struct {
	arg           kafka.Message
	expectedTopic string
}

func (ex *eqKafkaMessageParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(kafka.Message)
	if !ok {
		return false
	}

	ex.arg.Time = arg.Time
	ex.arg.Topic = arg.Topic
	ex.arg.Value = arg.Value
	ex.arg.Headers = arg.Headers

	if ex.arg.Value == nil {
		return false
	}

	if ex.arg.Time.IsZero() {
		return false
	}

	if ex.arg.Topic == "" {
		return false
	}

	if ex.arg.Topic != ex.expectedTopic {
		return false
	}

	if !reflect.DeepEqual(ex.arg, arg) {
		return false
	}

	return true
}

func (ex *eqKafkaMessageParamsMatcher) String() string {
	var errMsg string

	if ex.arg.Value == nil {
		errMsg += "value should not be empty"
	}

	if ex.arg.Time.IsZero() {
		errMsg += "time should not be zero"
	}

	if ex.arg.Topic == "" {
		errMsg += "topic should not be empty"
	}

	if ex.arg.Topic != ex.expectedTopic {
		errMsg += "topic not equal to expected topic"
	}

	return errMsg + fmt.Sprintf("matches arg: %v", ex.arg)
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
