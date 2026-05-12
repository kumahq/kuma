package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
)

type Metrics struct {
	VipGenerations          prometheus.Histogram
	VipGenerationsErrors    prometheus.Counter
	VipAllocationExhaustion *prometheus.CounterVec
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
	vipAllocationExhaustion := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "vip_allocation_exhaustion_total",
		Help: "Total VIP allocation failures due to IPAM pool exhaustion, by mesh.",
	}, []string{"mesh"})
	if err := metrics.BulkRegister(vipGenerations, vipGenerationsErrors, vipAllocationExhaustion); err != nil {
		return nil, err
	}

	return &Metrics{
		VipGenerations:          vipGenerations,
		VipGenerationsErrors:    vipGenerationsErrors,
		VipAllocationExhaustion: vipAllocationExhaustion,
	}, nil
}
