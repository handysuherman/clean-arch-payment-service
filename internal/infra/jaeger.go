package infra

import (
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/opentracing/opentracing-go"
)

func (a *app) jaeger() error {
	jaegerOpt := tracing.Config{
		ServiceName: a.cfg.Services.Internal.Name,
		HostPort:    a.cfg.Monitoring.Jaeger.HostPort,
		Enable:      a.cfg.Monitoring.Jaeger.Enable,
		LogSpans:    a.cfg.Monitoring.Jaeger.Logspan,
	}

	if a.cfg.Monitoring.Jaeger.Enable {
		tracer, closer, err := tracing.NewJaegerMetrics(&jaegerOpt)
		if err != nil {
			return err
		}
		a.jaegerCloser = closer
		opentracing.SetGlobalTracer(tracer)

		a.log.Infof("successfully connected to jaeger: %v", a.cfg.Monitoring.Jaeger.HostPort)
	}

	return nil
}
