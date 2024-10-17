package infra

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	stackSize    = 1 << 10
	bodyLimit    = 4 * 1024 * 1024
	readTimeout  = 15 * time.Second
	writeTimeout = 15 * time.Second
)

func (a *app) runMetrics(cancel context.CancelFunc) {
	a.metricsServer = echo.New()
	go func() {
		a.metricsServer.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
			StackSize:         stackSize,
			DisableStackAll:   true,
			DisablePrintStack: true,
		}))
		a.metricsServer.GET(a.cfg.Monitoring.Probes.Prometheus.Path, echo.WrapHandler(promhttp.Handler()))
		a.log.Infof("metrics server is running on port: %v", a.cfg.Monitoring.Probes.Prometheus.Port)
		if err := a.metricsServer.Start(a.cfg.Monitoring.Probes.Prometheus.Port); err != nil {
			a.log.Errorf("a.runMetrics.Start: %v", err)
			cancel()
		}
	}()
}
