package infra

import (
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/repository"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/usecase"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/worker"
	"github.com/xendit/xendit-go/v5"
)

func (a *app) handlers() {
	a.xenditPaymentGateway = xendit.NewClient(a.cfg.Services.Internal.PaymentGatewayKeys.Development)

	if a.cfg.Services.Internal.Environment == "production" {
		a.xenditPaymentGateway = xendit.NewClient(a.cfg.Services.Internal.PaymentGatewayKeys.Production)
	}

	repo := repository.NewStore(
		a.log,
		a.cfg,
		a.cfgManager.PqsqlConnection(),
		a.cfgManager.RedisConnection(),
		a.xenditPaymentGateway,
	)
	worker := worker.New(a.log, a.cfg, a.cfgManager.ProducerWorker())
	a.usecase = usecase.New(a.log, a.cfg, repo, worker)

	a.cfgManager.RegisterPqsqlObserver(repo)
	a.cfgManager.RegisterRedisObserver(repo)
	a.cfgManager.RegisterProducerWorkerObserver(worker)

	a.cfgManager.RegisterObserver(repo, 1)
	a.cfgManager.RegisterObserver(worker, 2)
	a.cfgManager.RegisterObserver(a.usecase, 3)
}
