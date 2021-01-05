package sync

import (
	"github.com/prometheus/client_golang/prometheus"

	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

type XDSSyncMetrics struct {
	XdsGenerations       prometheus.Summary
	XdsGenerationsErrors prometheus.Counter
}

func NewXDSSyncMetrics(metrics core_metrics.Metrics) (*XDSSyncMetrics, error) {
	xdsGenerations := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "xds_generation",
		Help:       "Summary of XDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	})
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

	return &XDSSyncMetrics{
		XdsGenerations:       xdsGenerations,
		XdsGenerationsErrors: xdsGenerationsErrors,
	}, nil
}
