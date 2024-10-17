package worker

import (
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	kafkaClient "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror/tracing"
	"github.com/opentracing/opentracing-go"
)

func (w *Worker) errorResponse(span opentracing.Span, details string, err error) error {
	errfmt := fmt.Errorf("%s: %v", details, err)
	w.log.Warn(errfmt)
	tracing.TraceWithError(span, errfmt)

	return err
}

func (w *Worker) OnConfigUpdate(key string, config *config.App) {
	w.log.Infof("received update from '%s' key", key)

	w.cfg = config

	w.log.Infof("updated configuration from '%s' key successfully applied", key)
}

func (w *Worker) OnProducerWorkerUpdate(key string, workerProducerConnection *kafkaClient.ProducerImpl) {
	w.log.Infof("received update from '%s' key", key)

	w.distributor = workerProducerConnection

	w.log.Info("newly updated producer worker connection successfully applied")
}
