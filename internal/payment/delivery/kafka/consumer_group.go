package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/metrics"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/domain"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type messageProcessor struct {
	metrics             *metrics.Metrics
	log                 logger.Logger
	v                   *validator.Validate
	cfg                 *config.App
	cfgManager          *config.Manager
	usecase             domain.Usecase
	r                   *kafka.Reader
	consumerGroupTopics []string
}

func New(
	log logger.Logger,
	cfg *config.App,
	v *validator.Validate,
	usecase domain.Usecase,
	metrics *metrics.Metrics,
	topics []string,
	cfgManager *config.Manager,
) *messageProcessor {
	return &messageProcessor{
		log:                 log.WithPrefix(fmt.Sprintf("%s-%s", "payment", constants.Worker)),
		cfg:                 cfg,
		v:                   v,
		usecase:             usecase,
		metrics:             metrics,
		consumerGroupTopics: topics,
		cfgManager:          cfgManager,
	}
}

func (m *messageProcessor) ProcessMessages(
	ctx context.Context,
	r *kafka.Reader,
	wg *sync.WaitGroup,
	workerId int,
) {
	defer wg.Done()

	m.r = r

	m.cfgManager.RegisterObserver(m, 3)
	m.cfgManager.RegisterConsumerWorkerObserver(m)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := m.r.FetchMessage(ctx)
		if err != nil {
			m.log.Warnf("workerId: %v, err: %v", workerId, err)
			continue
		}

		m.logProcessMessage(msg, workerId)

		switch msg.Topic {
		case helper.StringBuilder(m.cfg.Services.External.PaymentGateway.ID, "_", m.cfg.Brokers.Kafka.Topics.PaymentStatusUpdate.TopicName):
			m.processUpdatePaymentStatus(ctx, msg)
		}
	}
}
