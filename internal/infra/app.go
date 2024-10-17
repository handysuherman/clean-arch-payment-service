package infra

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/metrics"
	"github.com/handysuherman/clean-arch-payment-service/internal/payment/domain"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/grpc_interceptor"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
	"github.com/xendit/xendit-go/v5"
)

type app struct {
	log        logger.Logger
	cfg        *config.App
	cfgManager *config.Manager
	kafkaConn  *kafka.Conn
	metrics    *metrics.Metrics

	xenditPaymentGateway *xendit.APIClient

	usecase domain.Usecase
	doneCh  chan struct{}

	im                grpc_interceptor.InterceptorManager
	jaegerCloser      io.Closer
	v                 *validator.Validate
	healthCheckServer *http.Server
	metricsServer     *echo.Echo
}

func New(log logger.Logger, cfg *config.App, cfgManager *config.Manager) *app {
	return &app{
		log:        log.WithPrefix("APP"),
		cfg:        cfg,
		doneCh:     make(chan struct{}),
		cfgManager: cfgManager,
		v:          validator.New(),
	}
}

func (a *app) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	a.im = grpc_interceptor.New(a.log)
	a.metrics = metrics.New(a.cfg)

	if err := a.jaeger(); err != nil {
		return err
	}
	defer a.jaegerCloser.Close()

	if err := a.pqsql(ctx); err != nil {
		return err
	}
	defer a.cfgManager.ClosePqsql()

	if err := a.consul(); err != nil {
		return err
	}

	if err := a.redis(ctx); err != nil {
		return err
	}
	defer a.cfgManager.CloseRedis()

	if err := a.kafkaProducer(); err != nil {
		return err
	}
	defer a.cfgManager.CloseProducerWorker()

	a.handlers()

	if err := a.kafkaConsumer(ctx); err != nil {
		return err
	}
	defer a.cfgManager.CloseConsumerWorker()
	defer a.kafkaConn.Close()

	if err := a.initKafkaTopic(ctx); err != nil {
		return err
	}

	if err := a.runDBMigration(); err != nil {
		return err
	}

	go func() {
		a.cfgManager.Watch(ctx, a.configurationPriorities())
	}()

	closeGrpcServer, grpcServer, err := a.newGrpcServer(ctx)
	if err != nil {
		return err
	}
	defer closeGrpcServer()

	go func() {
		a.log.Info("healthCheck server is running on: :%v...", a.cfg.Monitoring.Probes.Port)
		if err := a.runHealthCheck(ctx); err != nil {
			a.log.Errorf("a.runHealthCheck: %v", err)
			cancel()
		}
	}()

	a.runMetrics(cancel)

	a.shutdownProcess(ctx, grpcServer)

	return nil
}
