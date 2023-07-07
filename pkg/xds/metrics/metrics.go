package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

type Metrics struct {
	XdsGenerations       *prometheus.SummaryVec
	XdsGenerationsErrors prometheus.Counter
}

func NewMetrics(metrics core_metrics.Metrics) (*Metrics, error) {
	xdsGenerations := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "xds_generation",
		Help:       "Summary of XDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	}, []string{"proxy_type", "result"})
	if err := metrics.Register(xdsGenerations); err != nil {
		return nil, err
	}
	xdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "xds_generation_errors",
		Help: "Counter of errors during XDS generation",
	})
	if err := metrics.Register(xdsGenerationsErrors); err != nil {
		return nil, err
	}

	return &Metrics{
		XdsGenerations:       xdsGenerations,
		XdsGenerationsErrors: xdsGenerationsErrors,
	}, nil
}
