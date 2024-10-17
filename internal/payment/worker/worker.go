package worker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/domain"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	messageClient "github.com/handysuherman/clean-arch-payment-service/internal/pkg/kafka"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
)

type Worker struct {
	log         logger.Logger
	cfg         *config.App
	distributor messageClient.Producer
}

func New(
	log logger.Logger,
	cfg *config.App,
	distributor messageClient.Producer,
) domain.ProducerWorker {
	log = log.WithPrefix(fmt.Sprintf("%s-%s", "payment", constants.ProducerWorker))
	return &Worker{
		log:         log,
		cfg:         cfg,
		distributor: distributor,
	}
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
