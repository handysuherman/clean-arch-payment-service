package domain

import (
	"context"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/models"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	kafkaClient "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka"
)

type ProducerWorker interface {
	OnProducerWorkerUpdate(key string, workerProducerConnection *kafkaClient.ProducerImpl)
	OnConfigUpdate(key string, config *config.App)

	PaymentStatusUpdated(ctx context.Context, task *models.PaymentStatusUpdatedTask) error
}

type Usecase interface {
	OnConfigUpdate(key string, config *config.App)

	Create(ctx context.Context, arg *models.CreatePaymentRequest) (*pb.CreatePaymentResponse, error)
	Update(ctx context.Context, arg *models.UpdatePaymentRequest) error
	GetByID(ctx context.Context, arg *models.GetByIDPaymentRequest) (*pb.GetByIDPaymentResponse, error)

	GetAvailableChannel(ctx context.Context, arg *models.GetPaymentChannelRequest) (*pb.GetPaymentChannelResponse, error)
	GetAvailableChannels(ctx context.Context, arg *models.GetPaymentChannelsRequest) (*pb.GetPaymentChannelsResponse, error)
}

type Worker interface {
	OnConfigUpdate(key string, config *config.App)
}
