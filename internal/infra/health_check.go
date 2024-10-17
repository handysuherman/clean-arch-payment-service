package infra

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/heptiolabs/healthcheck"
)

func (a *app) runHealthCheck(ctx context.Context) error {
	health := healthcheck.NewHandler()

	mux := http.NewServeMux()
	mux.HandleFunc(a.cfg.Monitoring.Probes.LivenessPath, health.LiveEndpoint)
	mux.HandleFunc(a.cfg.Monitoring.Probes.ReadinessPath, health.ReadyEndpoint)
	mux.HandleFunc("/consul", a.consulHealthCheckEndpoint)

	a.healthCheckServer = &http.Server{
		Handler:      mux,
		Addr:         a.cfg.Monitoring.Probes.Port,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}

	if err := a.cfgManager.ConsulConnection().Agent().ServiceRegister(a.cfgManager.SetAppConsulRegisterConfig()); err != nil {
		a.log.Warnf("a.cfgManager.ConsulConnection.Agent.ServiceRegister.err: %v", err)
		return err
	}

	a.configureHealthCheckEndpoints(ctx, health)
	return a.healthCheckServer.ListenAndServe()
}

func (a *app) consulHealthCheckEndpoint(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	a.log.Info("health check performed by consul...")
	w.Write([]byte("OK"))
}

func (a *app) configureHealthCheckEndpoints(ctx context.Context, health healthcheck.Handler) {
	health.AddLivenessCheck(a.cfg.Services.Internal.Name, healthcheck.AsyncWithContext(ctx, func() error {
		return nil
	}, time.Duration(a.cfg.Monitoring.Probes.CheckInterval)*time.Second))

	health.AddReadinessCheck(constants.PostgreSQL, healthcheck.AsyncWithContext(ctx, func() error {
		return a.cfgManager.PqsqlConnection().Ping(ctx)
	}, time.Duration(a.cfg.Monitoring.Probes.CheckInterval)*time.Second))
}

func (a *app) gracefulShutDownHealthCheckServer(ctx context.Context) error {
	if err := a.healthCheckServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("a.healthCheckServer.Shutdown.err: %v", err)
	}

	a.cfgManager.ConsulConnection().Agent().ServiceDeregister(a.cfg.ServiceDiscovery.Consul.Internal.ServiceName)
	return nil
}
