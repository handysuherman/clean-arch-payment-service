package tracing

import (
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
)

type Config struct {
	ServiceName string `json:"service_name"`
	HostPort    string `json:"host_port"`
	Enable      bool   `json:"enable"`
	LogSpans    bool   `json:"log_spans"`
}

func (c *Config) String() string {
	return fmt.Sprintf("ServiceName: %s, HostPort: %s, Enable: %v, LogSpans: %v", c.ServiceName, c.HostPort, c.Enable, c.LogSpans)
}

func NewJaegerMetrics(conf *Config) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		ServiceName: conf.ServiceName,

		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},

		Reporter: &config.ReporterConfig{
			LogSpans:           conf.LogSpans,
			LocalAgentHostPort: conf.HostPort,
		},
	}

	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	return cfg.NewTracer(
		config.Logger(jaeger.StdLogger),
		config.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		config.Injector(opentracing.TextMap, zipkinPropagator),
		config.Injector(opentracing.Binary, zipkinPropagator),
		config.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		config.Extractor(opentracing.TextMap, zipkinPropagator),
		config.Extractor(opentracing.Binary, zipkinPropagator),
		config.ZipkinSharedRPCSpan(false),
	)
}
