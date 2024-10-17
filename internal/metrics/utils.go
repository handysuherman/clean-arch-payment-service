package metrics

import (
	"fmt"
	"strings"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func NewCounter(cfg *config.App, name string, protocol string) prometheus.Counter {
	promCounter := prometheus.CounterOpts{}

	switch protocol {
	case constants.Kafka:
		promCounter.Name = fmt.Sprintf("%s_%s_%s_messages_total", cfg.Services.Internal.Name, protocol, name)
		promCounter.Help = fmt.Sprintf("The total number of %s kafka messages", strings.ReplaceAll(name, "_", " "))
	default:
		promCounter.Name = fmt.Sprintf("%s_%s_%s_requests_total", cfg.Services.Internal.Name, protocol, name)
		promCounter.Help = fmt.Sprintf("The total number of %s %s requests", strings.ReplaceAll(name, "_", " "), protocol)
	}

	return promauto.NewCounter(promCounter)
}
