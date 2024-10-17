package usecase

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/domain"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/opentracing/opentracing-go"
)

type usecaseImpl struct {
	log    logger.Logger
	cfg    *config.App
	repo   repository.Repository
	worker domain.ProducerWorker
}

func New(
	log logger.Logger,
	cfg *config.App,
	repository repository.Repository,
	worker domain.ProducerWorker,
) domain.Usecase {
	return &usecaseImpl{
		log:    log.WithPrefix(fmt.Sprintf("%s-%s", "payment", constants.Usecase)),
		cfg:    cfg,
		repo:   repository,
		worker: worker,
	}
}

func (u *usecaseImpl) OnConfigUpdate(key string, config *config.App) {
	u.log.Infof("received update from '%s' key", key)

	u.cfg = config

	u.log.Infof("updated configuration from '%s' key successfully applied", key)
}

func (u *usecaseImpl) errorResponse(span opentracing.Span, details string, err error) error {
	errfmt := fmt.Errorf("%s: %v", details, err)
	u.log.Warn(errfmt)
	tracing.TraceWithError(span, errfmt)

	return err
}

func findModuleRoot(dir string) string {
	for {
		_, err := os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}
