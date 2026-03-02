package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
)

type Metrics struct {
	VipGenerations       prometheus.Histogram
	VipGenerationsErrors prometheus.Counter
}

func NewMetrics(metrics core_metrics.Metrics) (*Metrics, error) {
	vipGenerations := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "vip_generation",
		Help: "Summary of VIP generation",
	})
	vipGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "vip_generation_errors",
		Help: "Counter of errors during VIP generation",
	})
	if err := metrics.BulkRegister(vipGenerations, vipGenerationsErrors); err != nil {
		return nil, err
	}

	return &Metrics{
		VipGenerations:       vipGenerations,
		VipGenerationsErrors: vipGenerationsErrors,
	}, nil
}
